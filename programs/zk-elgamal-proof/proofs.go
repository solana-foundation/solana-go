package zk

import (
	"encoding/binary"
	"fmt"
)

// PubkeyValidityProof proves knowledge of the secret key for the keypair's public key.
func PubkeyValidityProof(kp *ElGamalKeypair) (*Proof, error) {
	kb := kp.bytes()
	defer zeroize(kb)
	return invokeProof("proof_pubkey_validity", ProofTypePubkeyValidity, kb)
}

// ZeroCiphertextProof proves that ct encrypts zero under the keypair's public key.
func ZeroCiphertextProof(kp *ElGamalKeypair, ct ElGamalCiphertext) (*Proof, error) {
	kb := kp.bytes()
	defer zeroize(kb)
	return invokeProof("proof_zero_ciphertext", ProofTypeZeroCiphertext, kb, ct[:])
}

// CiphertextCommitmentEqualityProof proves that ct and commitment encode the same amount.
func CiphertextCommitmentEqualityProof(
	kp *ElGamalKeypair,
	ct ElGamalCiphertext,
	commitment PedersenCommitment,
	opening PedersenOpening,
	amount uint64,
) (*Proof, error) {
	kb := kp.bytes()
	defer zeroize(kb)
	return invokeProof("proof_ciphertext_commitment_equality", ProofTypeCiphertextCommitmentEquality,
		kb, ct[:], commitment[:], opening[:], amount)
}

// CiphertextCiphertextEqualityProof proves that firstCt (under the first
// keypair) and secondCt (under secondPubkey, with secondOpening) encrypt the
// same amount.
func CiphertextCiphertextEqualityProof(
	firstKeypair *ElGamalKeypair,
	secondPubkey ElGamalPubkey,
	firstCt ElGamalCiphertext,
	secondCt ElGamalCiphertext,
	secondOpening PedersenOpening,
	amount uint64,
) (*Proof, error) {
	kb := firstKeypair.bytes()
	defer zeroize(kb)
	return invokeProof("proof_ciphertext_ciphertext_equality", ProofTypeCiphertextCiphertextEquality,
		kb, secondPubkey[:], firstCt[:], secondCt[:], secondOpening[:], amount)
}

// PercentageWithCapProof proves a fee computation.
func PercentageWithCapProof(
	percentageCommitment PedersenCommitment, percentageOpening PedersenOpening, percentageAmount uint64,
	deltaCommitment PedersenCommitment, deltaOpening PedersenOpening, deltaAmount uint64,
	claimedCommitment PedersenCommitment, claimedOpening PedersenOpening, maxValue uint64,
) (*Proof, error) {
	return invokeProof("proof_percentage_with_cap", ProofTypePercentageWithCap,
		percentageCommitment[:], percentageOpening[:], percentageAmount,
		deltaCommitment[:], deltaOpening[:], deltaAmount,
		claimedCommitment[:], claimedOpening[:], maxValue)
}

// BatchedRangeProofU64 proves that each committed amount fits in its bit length; the bit lengths must sum to 64.
func BatchedRangeProofU64(commitments []PedersenCommitment, amounts []uint64, bitLengths []uint8, openings []PedersenOpening) (*Proof, error) {
	return batchedRangeProof("proof_batched_range_u64", ProofTypeBatchedRangeProofU64, 64, commitments, amounts, bitLengths, openings)
}

// BatchedRangeProofU128 proves that each committed amount fits in its bit length; the bit lengths must sum to 128.
func BatchedRangeProofU128(commitments []PedersenCommitment, amounts []uint64, bitLengths []uint8, openings []PedersenOpening) (*Proof, error) {
	return batchedRangeProof("proof_batched_range_u128", ProofTypeBatchedRangeProofU128, 128, commitments, amounts, bitLengths, openings)
}

// BatchedRangeProofU256 proves that each committed amount fits in its bit length; the bit lengths must sum to 256.
func BatchedRangeProofU256(commitments []PedersenCommitment, amounts []uint64, bitLengths []uint8, openings []PedersenOpening) (*Proof, error) {
	return batchedRangeProof("proof_batched_range_u256", ProofTypeBatchedRangeProofU256, 256, commitments, amounts, bitLengths, openings)
}

func batchedRangeProof(export string, proofType ProofType, totalBits int, commitments []PedersenCommitment, amounts []uint64, bitLengths []uint8, openings []PedersenOpening) (*Proof, error) {
	n := len(commitments)
	if n == 0 || len(amounts) != n || len(bitLengths) != n || len(openings) != n {
		return nil, fmt.Errorf("zk: batched range proof requires equal-length non-empty inputs, got %d commitments, %d amounts, %d bit lengths, %d openings",
			n, len(amounts), len(bitLengths), len(openings))
	}
	sum := 0
	for _, bits := range bitLengths {
		sum += int(bits)
	}
	if sum != totalBits {
		return nil, fmt.Errorf("zk: batched range proof bit lengths sum to %d, want %d", sum, totalBits)
	}
	commitmentBytes := make([]byte, 0, n*32)
	openingBytes := make([]byte, 0, n*32)
	amountBytes := make([]byte, n*8)
	for i := 0; i < n; i++ {
		commitmentBytes = append(commitmentBytes, commitments[i][:]...)
		openingBytes = append(openingBytes, openings[i][:]...)
		binary.LittleEndian.PutUint64(amountBytes[i*8:], amounts[i])
	}
	return invokeProof(export, proofType,
		uint64(n), commitmentBytes, amountBytes, []byte(bitLengths), openingBytes)
}

// GroupedCiphertext2HandlesValidityProof proves that the grouped ciphertext is a valid encryption of amount under both public keys.
func GroupedCiphertext2HandlesValidityProof(
	pubkeys [2]ElGamalPubkey,
	grouped GroupedElGamalCiphertext2,
	amount uint64,
	opening PedersenOpening,
) (*Proof, error) {
	return invokeProof("proof_grouped_ciphertext_2_handles_validity", ProofTypeGroupedCiphertext2HandlesValidity,
		concatPubkeys2(pubkeys), grouped[:], amount, opening[:])
}

// GroupedCiphertext3HandlesValidityProof proves that the grouped ciphertext
// is a valid encryption of amount under all three public keys.
func GroupedCiphertext3HandlesValidityProof(
	pubkeys [3]ElGamalPubkey,
	grouped GroupedElGamalCiphertext3,
	amount uint64,
	opening PedersenOpening,
) (*Proof, error) {
	return invokeProof("proof_grouped_ciphertext_3_handles_validity", ProofTypeGroupedCiphertext3HandlesValidity,
		concatPubkeys3(pubkeys), grouped[:], amount, opening[:])
}

// BatchedGroupedCiphertext2HandlesValidityProof proves validity of a pair of
// grouped ciphertexts (the lo/hi split of an amount) under two public keys.
func BatchedGroupedCiphertext2HandlesValidityProof(
	pubkeys [2]ElGamalPubkey,
	groupedLo, groupedHi GroupedElGamalCiphertext2,
	amountLo, amountHi uint64,
	openingLo, openingHi PedersenOpening,
) (*Proof, error) {
	return invokeProof("proof_batched_grouped_ciphertext_2_handles_validity", ProofTypeBatchedGroupedCiphertext2HandlesValidity,
		concatPubkeys2(pubkeys), groupedLo[:], groupedHi[:], amountLo, amountHi, openingLo[:], openingHi[:])
}

// BatchedGroupedCiphertext3HandlesValidityProof proves validity of a pair of
// grouped ciphertexts (the lo/hi split of a transfer amount) under three
// public keys (sender, recipient, auditor). Used by confidential Transfer.
func BatchedGroupedCiphertext3HandlesValidityProof(
	pubkeys [3]ElGamalPubkey,
	groupedLo, groupedHi GroupedElGamalCiphertext3,
	amountLo, amountHi uint64,
	openingLo, openingHi PedersenOpening,
) (*Proof, error) {
	return invokeProof("proof_batched_grouped_ciphertext_3_handles_validity", ProofTypeBatchedGroupedCiphertext3HandlesValidity,
		concatPubkeys3(pubkeys), groupedLo[:], groupedHi[:], amountLo, amountHi, openingLo[:], openingHi[:])
}

func invokeProof(name string, proofType ProofType, parts ...any) (*Proof, error) {
	out, err := invokeWith(name, parts...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}

	return &Proof{Type: proofType, Data: out}, nil
}
