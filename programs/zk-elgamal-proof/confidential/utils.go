package confidential

import (
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

// orIdentity resolves an optional auditor pubkey, substituting the identity
// pubkey when nil like the Rust ElGamalPubkey::default().
func orIdentity(pubkey *encryption.ElGamalPubkey) encryption.ElGamalPubkey {
	if pubkey != nil {
		return *pubkey
	}
	return encryption.ElGamalPubkey{}
}

// splitAmount splits amount into its loBits low bits and the remaining high
// bits.
func splitAmount(amount uint64, loBits uint8) (lo, hi uint64) {
	return amount & (1<<loBits - 1), amount >> loBits
}

// checkSpend rejects a spend the encryption cannot represent, or is more than the account balance.
func checkSpend(amount, balance uint64) error {
	if amount > MaxAmount {
		return ErrIllegalAmountBitLength
	}
	if amount > balance {
		return ErrNotEnoughFunds
	}
	return nil
}

// CiphertextValidityProofWithAuditorCiphertext wraps a ciphertext validity proof along with two
// single-key `lo` and `hi` ciphertexts.
//
// ProofData contains grouped ElGamal ciphertexts and a proof containing the validity of these ciphertexts.
type CiphertextValidityProofWithAuditorCiphertext struct {
	ProofData    *proofdata.BatchedGroupedCiphertext3HandlesValidityProofData
	CiphertextLo encryption.ElGamalCiphertext
	CiphertextHi encryption.ElGamalCiphertext
}

const auditorHandleIndex = 2

// encryptAndProveAmount encrypts the lo/hi amount split under the three
// public keys with fresh Pedersen openings and proves the grouped ciphertexts
// are valid encryptions.
func encryptAndProveAmount(
	pubkeys [3]encryption.ElGamalPubkey,
	amountLo, amountHi uint64,
) (validity CiphertextValidityProofWithAuditorCiphertext, openingLo, openingHi encryption.PedersenOpening, err error) {
	openingLo, err = encryption.NewPedersenOpening()
	if err != nil {
		return
	}
	openingHi, err = encryption.NewPedersenOpening()
	if err != nil {
		return
	}
	groupedLo, err := encryption.GroupedElGamalEncrypt3(pubkeys, amountLo, openingLo)
	if err != nil {
		return
	}
	groupedHi, err := encryption.GroupedElGamalEncrypt3(pubkeys, amountHi, openingHi)
	if err != nil {
		return
	}
	proofData, err := proofdata.NewBatchedGroupedCiphertext3HandlesValidityProofData(
		pubkeys, groupedLo, groupedHi, amountLo, amountHi, openingLo, openingHi)
	if err != nil {
		return
	}
	auditorCiphertextLo, err := groupedLo.ToElGamalCiphertext(auditorHandleIndex)
	if err != nil {
		return
	}
	auditorCiphertextHi, err := groupedHi.ToElGamalCiphertext(auditorHandleIndex)
	if err != nil {
		return
	}
	validity = CiphertextValidityProofWithAuditorCiphertext{
		ProofData:    proofData,
		CiphertextLo: auditorCiphertextLo,
		CiphertextHi: auditorCiphertextHi,
	}
	return
}

// encryptAndProveFeeAmount is encryptAndProveAmount for the two-key fee
// split. The fee grouped ciphertexts travel in the proof's context data, so
// only the proof and openings are returned.
func encryptAndProveFeeAmount(
	pubkeys [2]encryption.ElGamalPubkey,
	feeLoPlaintext, feeHiPlaintext uint64,
) (validity *proofdata.BatchedGroupedCiphertext2HandlesValidityProofData, openingLo, openingHi encryption.PedersenOpening, err error) {
	openingLo, err = encryption.NewPedersenOpening()
	if err != nil {
		return
	}
	openingHi, err = encryption.NewPedersenOpening()
	if err != nil {
		return
	}
	groupedLo, err := encryption.GroupedElGamalEncrypt2(pubkeys, feeLoPlaintext, openingLo)
	if err != nil {
		return
	}
	groupedHi, err := encryption.GroupedElGamalEncrypt2(pubkeys, feeHiPlaintext, openingHi)
	if err != nil {
		return
	}
	validity, err = proofdata.NewBatchedGroupedCiphertext2HandlesValidityProofData(
		pubkeys, groupedLo, groupedHi, feeLoPlaintext, feeHiPlaintext, openingLo, openingHi)
	return
}

// combinedCiphertextForHandle extracts the lo and hi ElGamal ciphertexts for
// the key at handle index from the proof's grouped ciphertexts and combines
// them into one single-key ciphertext of the full amount.
func (v CiphertextValidityProofWithAuditorCiphertext) combinedCiphertextForHandle(index int, loBits uint8) (encryption.ElGamalCiphertext, error) {
	lo, err := v.ProofData.Context.GroupedCiphertextLo.ToElGamalCiphertext(index)
	if err != nil {
		return encryption.ElGamalCiphertext{}, err
	}
	hi, err := v.ProofData.Context.GroupedCiphertextHi.ToElGamalCiphertext(index)
	if err != nil {
		return encryption.ElGamalCiphertext{}, err
	}
	return encryption.CombineLoHiCiphertexts(lo, hi, loBits)
}

// balanceOp proves that a homomorophic operation on two ciphertexts yields a fresh commitment to the plaintext result.
type balanceOp func(kp *encryption.ElGamalKeypair, a, b encryption.ElGamalCiphertext, result uint64) (
	*proofdata.CiphertextCommitmentEqualityProofData, encryption.PedersenCommitment, encryption.PedersenOpening, error)

// proveCiphertextDifference subtracts encryptedSubtrahend from encryptedMinuend
// and proves the result matches a fresh commitment to the plaintext difference
func proveCiphertextDifference(kp *encryption.ElGamalKeypair, encryptedMinuend, encryptedSubtrahend encryption.ElGamalCiphertext, difference uint64) (equalityProof *proofdata.CiphertextCommitmentEqualityProofData, differenceCommitment encryption.PedersenCommitment, differenceOpening encryption.PedersenOpening, err error) {
	encryptedDifference, err := encryption.SubtractCiphertexts(encryptedMinuend, encryptedSubtrahend)
	if err != nil {
		return
	}

	differenceCommitment, differenceOpening, err = encryption.NewPedersenCommitment(difference)
	if err != nil {
		return
	}
	equalityProof, err = proofdata.NewCiphertextCommitmentEqualityProofData(
		kp, encryptedDifference, differenceCommitment, differenceOpening, difference)
	return
}

// proveCiphertextSum adds the two ciphertexts home and proves the result
// matches a fresh commitment to the plaintext sum.
func proveCiphertextSum(kp *encryption.ElGamalKeypair, encryptedA, encryptedB encryption.ElGamalCiphertext, sum uint64) (equalityProof *proofdata.CiphertextCommitmentEqualityProofData, sumCommitment encryption.PedersenCommitment, sumOpening encryption.PedersenOpening, err error) {
	encryptedSum, err := encryption.AddCiphertexts(encryptedA, encryptedB)
	if err != nil {
		return
	}

	sumCommitment, sumOpening, err = encryption.NewPedersenCommitment(sum)
	if err != nil {
		return
	}
	equalityProof, err = proofdata.NewCiphertextCommitmentEqualityProofData(
		kp, encryptedSum, sumCommitment, sumOpening, sum)
	return
}

// balanceChangeData is the material the Transfer, Burn, Mint, and
// TransferWithFee builders produce before their range proofs.
type balanceChangeData struct {
	changeAmountCipherTextValidityProof CiphertextValidityProofWithAuditorCiphertext
	finalBalanceEqualityProof           *proofdata.CiphertextCommitmentEqualityProofData

	changeAmountPlaintextLo, changeAmountPlaintextHi uint64
	changeAmountOpeningLo, changeAmountOpeningHi     encryption.PedersenOpening

	// finalBalance is the new balance or supply
	finalBalance           uint64
	finalBalanceCommitment encryption.PedersenCommitment
	finalBalanceOpening    encryption.PedersenOpening
}

// proveBalanceChange encrypts the lo/hi split of amount under pubkeys, then
// recombines the handles at selfIndex and proves that operation yields result.
func proveBalanceChange(
	kp *encryption.ElGamalKeypair,
	currentBalance encryption.ElGamalCiphertext,
	pubkeys [3]encryption.ElGamalPubkey,
	selfIndex int,
	changeAmount, finalBalance uint64,
	op balanceOp,
) (c balanceChangeData, err error) {
	c.changeAmountPlaintextLo, c.changeAmountPlaintextHi = splitAmount(changeAmount, AmountLoBitLength)
	c.finalBalance = finalBalance

	c.changeAmountCipherTextValidityProof, c.changeAmountOpeningLo, c.changeAmountOpeningHi, err = encryptAndProveAmount(pubkeys, c.changeAmountPlaintextLo, c.changeAmountPlaintextHi)
	if err != nil {
		return
	}
	changeAmountCiphertext, err := c.changeAmountCipherTextValidityProof.combinedCiphertextForHandle(selfIndex, AmountLoBitLength)
	if err != nil {
		return
	}
	c.finalBalanceEqualityProof, c.finalBalanceCommitment, c.finalBalanceOpening, err = op(kp, currentBalance, changeAmountCiphertext, finalBalance)
	return
}

// rangeProofU128 proves the 64-bit result and the lo/hi amount split are in
// range, padded out to the 128 bits the batched proof covers.
func (c balanceChangeData) rangeProofU128() (*proofdata.BatchedRangeProofU128Data, error) {
	loCommitment, err := encryption.PedersenCommitmentWith(c.changeAmountPlaintextLo, c.changeAmountOpeningLo)
	if err != nil {
		return nil, err
	}
	hiCommitment, err := encryption.PedersenCommitmentWith(c.changeAmountPlaintextHi, c.changeAmountOpeningHi)
	if err != nil {
		return nil, err
	}
	padCommitment, padOpening, err := encryption.NewPedersenCommitment(0)
	if err != nil {
		return nil, err
	}

	return proofdata.NewBatchedRangeProofU128Data(
		[]encryption.PedersenCommitment{c.finalBalanceCommitment, loCommitment, hiCommitment, padCommitment},
		[]uint64{c.finalBalance, c.changeAmountPlaintextLo, c.changeAmountPlaintextHi, 0},
		[]uint8{BalanceBitLength, AmountLoBitLength, AmountHiBitLength, PadBitLength},
		[]encryption.PedersenOpening{c.finalBalanceOpening, c.changeAmountOpeningLo, c.changeAmountOpeningHi, padOpening},
	)
}
