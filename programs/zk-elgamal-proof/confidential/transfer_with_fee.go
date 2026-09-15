package confidential

import (
	"github.com/solana-foundation/solana-go/v2/programs/token-2022/zkencryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/internal/bridge"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

// TransferWithFeeProofData is the proof data a confidential Transfer
// instruction carries when the mint is extended for fees.
type TransferWithFeeProofData struct {
	// EqualityProofData proves the new encrypted balance equals a commitment
	EqualityProofData *proofdata.CiphertextCommitmentEqualityProofData
	// TransferAmountCiphertextValidityProofDataWithCiphertext proves the
	// transfer amount lo/hiciphertexts are valid encryptions under the source,
	// destination, and auditor keys, and carries the auditor's extracted lo/hi ciphertexts.
	TransferAmountCiphertextValidityProofDataWithCiphertext CiphertextValidityProofWithAuditorCiphertext
	// PercentageWithCapProofData proves the fee is either the correct
	// percentage of the transfer amount or exactly the maximum fee.
	PercentageWithCapProofData *proofdata.PercentageWithCapProofData
	// FeeCiphertextValidityProofData proves the lo/hi fee ciphertexts are
	// valid encryptions under the destination and withdraw-withheld-authority
	// keys.
	FeeCiphertextValidityProofData *proofdata.BatchedGroupedCiphertext2HandlesValidityProofData
	// RangeProofData proves the remaining balance, transfer lo/hi amounts, fee, fee delta, and net
	// amount are in range (256 bits total).
	RangeProofData *proofdata.BatchedRangeProofU256Data
}

// TransferWithFeeSplitProofData builds the five proofs a confidential
// Transfer on a fee-extended mint requires: remaining-balance equality,
// transfer amount lo/hi ciphertext validity under (source, destination,
// auditor), the percentage-with-cap fee proof, fee lo/hi ciphertext validity
// under (destination, withdraw withheld authority), and the batched range
// proof.
func TransferWithFeeSplitProofData(
	currentAvailableBalanceEGCiphertext encryption.ElGamalCiphertext,
	currentAvailableBalanceAESCiphertext encryption.AeCiphertext,
	transferAmountPlaintext uint64,
	sourceKeypair *encryption.ElGamalKeypair,
	aesKey zkencryption.AeKey,
	destinationPubkey encryption.ElGamalPubkey,
	auditorPubkey *encryption.ElGamalPubkey,
	withdrawWithheldAuthorityPubkey encryption.ElGamalPubkey,
	feeRateBasisPoints uint16,
	maximumFee uint64,
) (*TransferWithFeeProofData, error) {
	currentBalanceAmountPlaintext, err := encryption.AeDecrypt(aesKey, currentAvailableBalanceAESCiphertext)
	if err != nil {
		return nil, err
	}
	if err := checkSpend(transferAmountPlaintext, currentBalanceAmountPlaintext); err != nil {
		return nil, err
	}

	change, err := proveBalanceChange(sourceKeypair, currentAvailableBalanceEGCiphertext,
		[3]encryption.ElGamalPubkey{sourceKeypair.Pubkey, destinationPubkey, orIdentity(auditorPubkey)},
		0, transferAmountPlaintext, currentBalanceAmountPlaintext-transferAmountPlaintext, proveCiphertextDifference)
	if err != nil {
		return nil, err
	}

	// Calculate the fee, capping it at the maximum fee. When the cap is hit
	// the claimed rounding delta is zero for simplicity.
	feeAmountPlaintext, claimedDeltaPlaintext := calculateFee(transferAmountPlaintext, feeRateBasisPoints)
	if maximumFee < feeAmountPlaintext {
		feeAmountPlaintext, claimedDeltaPlaintext = maximumFee, 0
	}
	if feeAmountPlaintext > transferAmountPlaintext {
		return nil, ErrFeeCalculation
	}
	netAmountPlaintext := transferAmountPlaintext - feeAmountPlaintext

	// Encrypt the fee split under the destination and withdraw-withheld-
	// authority keys, and prove validity.
	feeLoPlaintext, feeHiPlaintext := splitAmount(feeAmountPlaintext, FeeAmountLoBitLength)
	feePubkeys := [2]encryption.ElGamalPubkey{destinationPubkey, withdrawWithheldAuthorityPubkey}
	feeLoHiCiphertextValidity, feeLoOpening, feeHiOpening, err := encryptAndProveFeeAmount(feePubkeys, feeLoPlaintext, feeHiPlaintext)
	if err != nil {
		return nil, err
	}

	// Combined commitments and openings to the full transfer amount and fee.
	transferAmountLoCommitment, transferAmountHiCommitment, transferAmountCommitment, transferAmountOpening, err := combineHiLoOpeningsCommitments(change.changeAmountPlaintextLo, change.changeAmountPlaintextHi, change.changeAmountOpeningLo, change.changeAmountOpeningHi)
	if err != nil {
		return nil, err
	}
	feeLoCommitment, feeHiCommitment, combinedFeeCommitment, combinedFeeOpening, err := combineHiLoOpeningsCommitments(feeLoPlaintext, feeHiPlaintext, feeLoOpening, feeHiOpening)
	if err != nil {
		return nil, err
	}

	// Net transfer amount = transfer amount - fee.
	netCommitment, err := encryption.SubtractCommitments(transferAmountCommitment, combinedFeeCommitment)
	if err != nil {
		return nil, err
	}
	netOpening, err := encryption.SubtractOpenings(transferAmountOpening, combinedFeeOpening)
	if err != nil {
		return nil, err
	}

	// Claimed and real fee rounding delta.
	claimedDeltaCommitment, claimedDeltaOpening, err := encryption.NewPedersenCommitment(claimedDeltaPlaintext)
	if err != nil {
		return nil, err
	}

	deltaCommitment, deltaOpening, err := feeDelta(
		transferAmountCommitment, transferAmountOpening, combinedFeeCommitment, combinedFeeOpening, feeRateBasisPoints)
	if err != nil {
		return nil, err
	}

	percentageWithCapProofData, err := proofdata.NewPercentageWithCapProofData(
		combinedFeeCommitment, combinedFeeOpening, feeAmountPlaintext,
		deltaCommitment, deltaOpening, claimedDeltaPlaintext,
		claimedDeltaCommitment, claimedDeltaOpening, maximumFee)
	if err != nil {
		return nil, err
	}

	// The complement claimed delta (9999 - delta) proves the delta itself is
	// at most 9999; its commitment uses the zero opening so the verifier can
	// reconstruct it.
	claimedComplementPlaintext := MaxFeeBasisPoints - 1 - claimedDeltaPlaintext
	var zeroOpening encryption.PedersenOpening
	maxSubOneCommitment, err := encryption.PedersenCommitmentWith(MaxFeeBasisPoints-1, zeroOpening)
	if err != nil {
		return nil, err
	}
	complementCommitment, err := encryption.SubtractCommitments(maxSubOneCommitment, claimedDeltaCommitment)
	if err != nil {
		return nil, err
	}
	complementOpening, err := encryption.SubtractOpenings(zeroOpening, claimedDeltaOpening)
	if err != nil {
		return nil, err
	}

	// Range proof over remaining (64), amount lo (16) and hi (32), claimed
	// delta (16), complement delta (16), fee lo (16) and hi (32), and net
	// amount (64), totalling 256 bits.
	rangeProof, err := proofdata.NewBatchedRangeProofU256Data(
		[]encryption.PedersenCommitment{
			change.finalBalanceCommitment, transferAmountLoCommitment, transferAmountHiCommitment,
			claimedDeltaCommitment, complementCommitment,
			feeLoCommitment, feeHiCommitment, netCommitment,
		},
		[]uint64{change.finalBalance, change.changeAmountPlaintextLo, change.changeAmountPlaintextHi, claimedDeltaPlaintext, claimedComplementPlaintext, feeLoPlaintext, feeHiPlaintext, netAmountPlaintext},
		[]uint8{BalanceBitLength, AmountLoBitLength, AmountHiBitLength, deltaBitLength, deltaBitLength, FeeAmountLoBitLength, FeeAmountHiBitLength, netAmountBitLength},
		[]encryption.PedersenOpening{
			change.finalBalanceOpening, change.changeAmountOpeningLo, change.changeAmountOpeningHi,
			claimedDeltaOpening, complementOpening,
			feeLoOpening, feeHiOpening, netOpening,
		},
	)
	if err != nil {
		return nil, err
	}

	return &TransferWithFeeProofData{
		EqualityProofData: change.finalBalanceEqualityProof,
		TransferAmountCiphertextValidityProofDataWithCiphertext: change.changeAmountCipherTextValidityProof,
		PercentageWithCapProofData:                              percentageWithCapProofData,
		FeeCiphertextValidityProofData:                          feeLoHiCiphertextValidity,
		RangeProofData:                                          rangeProof,
	}, nil
}

func combineHiLoOpeningsCommitments(amountLoPlaintext, amountHiPlaintext uint64, amountLoOpening, amountHiOpening encryption.PedersenOpening) (amountLoCommitment, amountHiCommitment, combinedAmountCommitment encryption.PedersenCommitment, combinedAmountOpening encryption.PedersenOpening, err error) {
	amountLoCommitment, err = encryption.PedersenCommitmentWith(amountLoPlaintext, amountLoOpening)
	if err != nil {
		return
	}
	amountHiCommitment, err = encryption.PedersenCommitmentWith(amountHiPlaintext, amountHiOpening)
	if err != nil {
		return
	}
	combinedAmountCommitment, err = encryption.CombineLoHiCommitments(amountLoCommitment, amountHiCommitment, AmountLoBitLength)
	if err != nil {
		return
	}
	combinedAmountOpening, err = encryption.CombineLoHiOpenings(amountLoOpening, amountHiOpening, AmountLoBitLength)
	if err != nil {
		return
	}
	return
}

// calculateFee returns the fee (transferAmount·feeRateBasisPoints/10000,
// rounded up) and the rounding delta fee·10000 - transferAmount·rate, which
// is always less than 10000. transferAmount must fit in 48 bits so the
// intermediate products cannot overflow.
func calculateFee(transferAmount uint64, feeRateBasisPoints uint16) (fee, delta uint64) {
	numerator := transferAmount * uint64(feeRateBasisPoints)
	fee = (numerator + MaxFeeBasisPoints - 1) / MaxFeeBasisPoints
	delta = fee*MaxFeeBasisPoints - numerator
	return fee, delta
}

// feeDelta computes the delta commitment and opening for the fee sigma proof,
// fee·10000 - combined·feeRateBasisPoints, mirroring
// compute_delta_commitment_and_opening in spl-token's confidential-transfer
// proof-generation crate.
func feeDelta(
	combinedCommitment encryption.PedersenCommitment, combinedOpening encryption.PedersenOpening,
	feeCommitment encryption.PedersenCommitment, feeOpening encryption.PedersenOpening,
	feeRateBasisPoints uint16,
) (encryption.PedersenCommitment, encryption.PedersenOpening, error) {
	pair, err := bridge.InvokeWith[encryption.PedersenCommitmentOpening]("pedersen_fee_delta",
		combinedCommitment, combinedOpening, feeCommitment, feeOpening, bridge.Scalar(feeRateBasisPoints))
	return pair.Commitment, pair.Opening, err
}
