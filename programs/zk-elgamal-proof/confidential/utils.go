package confidential

import (
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
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
	validity = CiphertextValidityProofWithAuditorCiphertext{
		ProofData:    proofData,
		CiphertextLo: groupedLo,
		CiphertextHi: groupedHi,
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
// the key at handle index and combines them into a single ciphertext of the full amount.
func (v CiphertextValidityProofWithAuditorCiphertext) combinedCiphertextForHandle(index int, loBits uint8) (encryption.ElGamalCiphertext, error) {
	lo, err := v.CiphertextLo.ToElGamalCiphertext(index)
	if err != nil {
		return encryption.ElGamalCiphertext{}, err
	}
	hi, err := v.CiphertextHi.ToElGamalCiphertext(index)
	if err != nil {
		return encryption.ElGamalCiphertext{}, err
	}
	return encryption.CombineLoHiCiphertexts(lo, hi, loBits)
}

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

// proveAmountRangeU128 proves a 64-bit amount and the 16/32-bit lo/hi split
// of a second amount are in range of U128
func proveAmountRangeU128(
	amountCommitment encryption.PedersenCommitment,
	amountOpening encryption.PedersenOpening,
	amount, lo, hi uint64,
	loOpening, hiOpening encryption.PedersenOpening,
) (*proofdata.BatchedRangeProofU128Data, error) {
	loCommitment, err := encryption.PedersenCommitmentWith(lo, loOpening)
	if err != nil {
		return nil, err
	}
	hiCommitment, err := encryption.PedersenCommitmentWith(hi, hiOpening)
	if err != nil {
		return nil, err
	}
	padCommitment, padOpening, err := encryption.NewPedersenCommitment(0)
	if err != nil {
		return nil, err
	}
	const (
		AmountNumBits = 64
		LoNumBits     = 16
		HiNumBits     = 32
		PadNumBits    = 16 // PadNumBits = 128 - AmountNumBits - LoNumBits  - HiNumBits
	)

	return proofdata.NewBatchedRangeProofU128Data(
		[]encryption.PedersenCommitment{amountCommitment, loCommitment, hiCommitment, padCommitment},
		[]uint64{amount, lo, hi, 0},
		[]uint8{AmountNumBits, LoNumBits, HiNumBits, PadNumBits},
		[]encryption.PedersenOpening{amountOpening, loOpening, hiOpening, padOpening},
	)
}
