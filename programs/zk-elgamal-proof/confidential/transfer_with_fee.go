package confidential

import (
	"fmt"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/bridge"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
)

// Transfer fees are split into a 16-bit low and 32-bit high part.
// Fee rates are expressed in basis points of the transfer amount, rounded up.
const (
	FeeAmountLoBitLength = 16
	FeeAmountHiBitLength = 32
	MaxFeeBasisPoints    = 10_000

	// deltaBitLength bounds the fee rounding error certified by the
	// percentage-with-cap proof.
	deltaBitLength = 16
)

// TransferWithFeeProofData is the proof data a confidential Transfer
// instruction carries when the mint is extended for fees.
type TransferWithFeeProofData struct {
	// RemainingBalanceProofData proves the new encrypted balance equals a commitment
	RemainingBalanceProofData *proofdata.CiphertextCommitmentEqualityProofData
	// TransferAmountCiphertextValidityProofDataWithCiphertext proves the
	// transfer amount lo/hiciphertexts are valid encryptions under the source,
	// destination, and auditor keys.
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
	aesKey encryption.AeKey,
	destinationPubkey encryption.ElGamalPubkey,
	auditorPubkey *encryption.ElGamalPubkey,
	withdrawWithheldAuthorityPubkey encryption.ElGamalPubkey,
	feeRateBasisPoints uint16,
	maximumFee uint64,
) (*TransferWithFeeProofData, error) {
	currentBalanceAmountPlaintext, err := aesKey.Decrypt(currentAvailableBalanceAESCiphertext)
	if err != nil {
		return nil, err
	}
	if transferAmountPlaintext > MaxTransferAmount {
		return nil, ErrIllegalAmountBitLength
	}
	if transferAmountPlaintext > currentBalanceAmountPlaintext {
		return nil, ErrNotEnoughFunds
	}
	transferAmountLoPlaintext, transferAmountHiPlaintext := splitAmount(transferAmountPlaintext, TransferAmountLoBitLength)
	remainingBalancePlaintext := currentBalanceAmountPlaintext - transferAmountPlaintext
	pubkeys := [3]encryption.ElGamalPubkey{sourceKeypair.Pubkey, destinationPubkey, orIdentity(auditorPubkey)}

	// Encrypt the transfer amount split under the source, destination, and auditor keys.
	transferAmountLoHiCiphertextValidityProof, transferAmountLoOpening, transferAmountHiOpening, err := encryptAndProveAmount(pubkeys, transferAmountLoPlaintext, transferAmountHiPlaintext)
	if err != nil {
		return nil, err
	}

	// New balance = current balance - (lo + 2^16 * hi), homomorphically, and
	// the equality proof that it matches a commitment to the remaining amount.
	combinedTransferAmountCiphertext, err := transferAmountLoHiCiphertextValidityProof.combinedCiphertextForHandle(0, TransferAmountLoBitLength)
	if err != nil {
		return nil, err
	}
	remainingBalanceEqualityProof, remainingBalanceCommitment, remainingBalanceOpening, err := proveCiphertextDifference(
		sourceKeypair, currentAvailableBalanceEGCiphertext, combinedTransferAmountCiphertext, remainingBalancePlaintext)
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
	transferAmountLoCommitment, transferAmountHiCommitment, transferAmountCommitment, transferAmountOpening, err := combineHiLoOpeningsCommitments(transferAmountLoPlaintext, transferAmountHiPlaintext, transferAmountLoOpening, transferAmountHiOpening)
	feeLoCommitment, feeHiCommitment, combinedFeeCommitment, combinedFeeOpening, err := combineHiLoOpeningsCommitments(feeLoPlaintext, feeHiPlaintext, feeLoOpening, feeHiOpening)

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
			remainingBalanceCommitment, transferAmountLoCommitment, transferAmountHiCommitment,
			claimedDeltaCommitment, complementCommitment,
			feeLoCommitment, feeHiCommitment, netCommitment,
		},
		[]uint64{remainingBalancePlaintext, transferAmountLoPlaintext, transferAmountHiPlaintext, claimedDeltaPlaintext, claimedComplementPlaintext, feeLoPlaintext, feeHiPlaintext, netAmountPlaintext},
		[]uint8{64, TransferAmountLoBitLength, TransferAmountHiBitLength, deltaBitLength, deltaBitLength, FeeAmountLoBitLength, FeeAmountHiBitLength, 64},
		[]encryption.PedersenOpening{
			remainingBalanceOpening, transferAmountLoOpening, transferAmountHiOpening,
			claimedDeltaOpening, complementOpening,
			feeLoOpening, feeHiOpening, netOpening,
		},
	)
	if err != nil {
		return nil, err
	}

	return &TransferWithFeeProofData{
		RemainingBalanceProofData:                               remainingBalanceEqualityProof,
		TransferAmountCiphertextValidityProofDataWithCiphertext: transferAmountLoHiCiphertextValidityProof,
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
	combinedAmountCommitment, err = encryption.CombineLoHiCommitments(amountLoCommitment, amountHiCommitment, TransferAmountLoBitLength)
	if err != nil {
		return
	}
	combinedAmountOpening, err = encryption.CombineLoHiOpenings(amountLoOpening, amountHiOpening, TransferAmountLoBitLength)
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
	var commitment encryption.PedersenCommitment
	var opening encryption.PedersenOpening
	out, err := bridge.InvokeWith("pedersen_fee_delta",
		combinedCommitment[:], combinedOpening[:], feeCommitment[:], feeOpening[:], uint64(feeRateBasisPoints))
	if err != nil {
		return commitment, opening, err
	}
	if len(out) != 64 {
		return commitment, opening, fmt.Errorf("zk: guest returned %d bytes, want 64", len(out))
	}
	copy(commitment[:], out[:32])
	copy(opening[:], out[32:64])
	return commitment, opening, nil
}
