package proofdata

import (
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/bridge"
)

// NewPubkeyValidityProofData proves knowledge of the secret key for the
// keypair's public key.
func NewPubkeyValidityProofData(kp *encryption.ElGamalKeypair) (*PubkeyValidityProofData, error) {
	kb := kp.Bytes()
	defer bridge.Zeroize(kb)
	return invokeProof[PubkeyValidityProofData]("proof_pubkey_validity", kb)
}

// NewZeroCiphertextProofData proves that ct encrypts zero under the keypair's
// public key.
func NewZeroCiphertextProofData(kp *encryption.ElGamalKeypair, ct encryption.ElGamalCiphertext) (*ZeroCiphertextProofData, error) {
	kb := kp.Bytes()
	defer bridge.Zeroize(kb)
	return invokeProof[ZeroCiphertextProofData]("proof_zero_ciphertext", kb, ct[:])
}

// NewCiphertextCommitmentEqualityProofData proves that ct and commitment
// encode the same amount.
func NewCiphertextCommitmentEqualityProofData(
	kp *encryption.ElGamalKeypair,
	ct encryption.ElGamalCiphertext,
	commitment encryption.PedersenCommitment,
	opening encryption.PedersenOpening,
	amount uint64,
) (*CiphertextCommitmentEqualityProofData, error) {
	kb := kp.Bytes()
	defer bridge.Zeroize(kb)
	return invokeProof[CiphertextCommitmentEqualityProofData]("proof_ciphertext_commitment_equality",
		kb, ct[:], commitment[:], opening[:], amount)
}

// NewCiphertextCiphertextEqualityProofData proves that firstCt (under the
// first keypair) and secondCt (under secondPubkey, with secondOpening)
// encrypt the same amount.
func NewCiphertextCiphertextEqualityProofData(
	firstKeypair *encryption.ElGamalKeypair,
	secondPubkey encryption.ElGamalPubkey,
	firstCt encryption.ElGamalCiphertext,
	secondCt encryption.ElGamalCiphertext,
	secondOpening encryption.PedersenOpening,
	amount uint64,
) (*CiphertextCiphertextEqualityProofData, error) {
	kb := firstKeypair.Bytes()
	defer bridge.Zeroize(kb)
	return invokeProof[CiphertextCiphertextEqualityProofData]("proof_ciphertext_ciphertext_equality",
		kb, secondPubkey[:], firstCt[:], secondCt[:], secondOpening[:], amount)
}

// NewPercentageWithCapProofData proves a fee computation.
func NewPercentageWithCapProofData(
	percentageCommitment encryption.PedersenCommitment, percentageOpening encryption.PedersenOpening, percentageAmount uint64,
	deltaCommitment encryption.PedersenCommitment, deltaOpening encryption.PedersenOpening, deltaAmount uint64,
	claimedCommitment encryption.PedersenCommitment, claimedOpening encryption.PedersenOpening, maxValue uint64,
) (*PercentageWithCapProofData, error) {
	return invokeProof[PercentageWithCapProofData]("proof_percentage_with_cap",
		percentageCommitment[:], percentageOpening[:], percentageAmount,
		deltaCommitment[:], deltaOpening[:], deltaAmount,
		claimedCommitment[:], claimedOpening[:], maxValue)
}

// NewBatchedRangeProofU64Data proves that each committed amount fits in its bit length; the bit lengths must sum to 64.
func NewBatchedRangeProofU64Data(commitments []encryption.PedersenCommitment, amounts []uint64, bitLengths []uint8, openings []encryption.PedersenOpening) (*BatchedRangeProofU64Data, error) {
	return batchedRangeProof[BatchedRangeProofU64Data]("proof_batched_range_u64", 64, commitments, amounts, bitLengths, openings)
}

// NewBatchedRangeProofU128Data proves that each committed amount fits in its bit length; the bit lengths must sum to 128.
func NewBatchedRangeProofU128Data(commitments []encryption.PedersenCommitment, amounts []uint64, bitLengths []uint8, openings []encryption.PedersenOpening) (*BatchedRangeProofU128Data, error) {
	return batchedRangeProof[BatchedRangeProofU128Data]("proof_batched_range_u128", 128, commitments, amounts, bitLengths, openings)
}

// NewBatchedRangeProofU256Data proves that each committed amount fits in its bit length; the bit lengths must sum to 256.
func NewBatchedRangeProofU256Data(commitments []encryption.PedersenCommitment, amounts []uint64, bitLengths []uint8, openings []encryption.PedersenOpening) (*BatchedRangeProofU256Data, error) {
	return batchedRangeProof[BatchedRangeProofU256Data]("proof_batched_range_u256", 256, commitments, amounts, bitLengths, openings)
}

func batchedRangeProof[T any, P interface {
	*T
	ProofData
}](export string, totalBits int, commitments []encryption.PedersenCommitment, amounts []uint64, bitLengths []uint8, openings []encryption.PedersenOpening) (*T, error) {
	n := len(commitments)
	if n == 0 || len(amounts) != n || len(bitLengths) != n || len(openings) != n {
		return nil, fmt.Errorf("zk: batched range proof requires equal-length non-empty inputs, got %d commitments, %d amounts, %d bit lengths, %d openings",
			n, len(amounts), len(bitLengths), len(openings))
	}
	if n > MaxRangeProofCommitments {
		return nil, fmt.Errorf("zk: batched range proof supports at most %d commitments, got %d", MaxRangeProofCommitments, n)
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
	for i := range n {
		commitmentBytes = append(commitmentBytes, commitments[i][:]...)
		openingBytes = append(openingBytes, openings[i][:]...)
		binary.LittleEndian.PutUint64(amountBytes[i*8:], amounts[i])
	}
	return invokeProof[T, P](export,
		uint64(n), commitmentBytes, amountBytes, []byte(bitLengths), openingBytes)
}

// NewGroupedCiphertext2HandlesValidityProofData proves that the grouped ciphertext is a valid encryption of amount under both public keys.
func NewGroupedCiphertext2HandlesValidityProofData(
	pubkeys [2]encryption.ElGamalPubkey,
	grouped encryption.GroupedElGamalCiphertext2,
	amount uint64,
	opening encryption.PedersenOpening,
) (*GroupedCiphertext2HandlesValidityProofData, error) {
	return invokeProof[GroupedCiphertext2HandlesValidityProofData]("proof_grouped_ciphertext_2_handles_validity",
		concatPubkeys2(pubkeys), grouped[:], amount, opening[:])
}

// NewGroupedCiphertext3HandlesValidityProofData proves that the grouped
// ciphertext is a valid encryption of amount under all three public keys.
func NewGroupedCiphertext3HandlesValidityProofData(
	pubkeys [3]encryption.ElGamalPubkey,
	grouped encryption.GroupedElGamalCiphertext3,
	amount uint64,
	opening encryption.PedersenOpening,
) (*GroupedCiphertext3HandlesValidityProofData, error) {
	return invokeProof[GroupedCiphertext3HandlesValidityProofData]("proof_grouped_ciphertext_3_handles_validity",
		concatPubkeys3(pubkeys), grouped[:], amount, opening[:])
}

// NewBatchedGroupedCiphertext2HandlesValidityProofData proves validity of a
// pair of grouped ciphertexts (the lo/hi split of an amount) under two public
// keys.
func NewBatchedGroupedCiphertext2HandlesValidityProofData(
	pubkeys [2]encryption.ElGamalPubkey,
	groupedLo, groupedHi encryption.GroupedElGamalCiphertext2,
	amountLo, amountHi uint64,
	openingLo, openingHi encryption.PedersenOpening,
) (*BatchedGroupedCiphertext2HandlesValidityProofData, error) {
	return invokeProof[BatchedGroupedCiphertext2HandlesValidityProofData]("proof_batched_grouped_ciphertext_2_handles_validity",
		concatPubkeys2(pubkeys), groupedLo[:], groupedHi[:], amountLo, amountHi, openingLo[:], openingHi[:])
}

// NewBatchedGroupedCiphertext3HandlesValidityProofData proves validity of a
// pair of grouped ciphertexts (the lo/hi split of a transfer amount) under
// three public keys (sender, recipient, auditor). Used by confidential
// Transfer.
func NewBatchedGroupedCiphertext3HandlesValidityProofData(
	pubkeys [3]encryption.ElGamalPubkey,
	groupedLo, groupedHi encryption.GroupedElGamalCiphertext3,
	amountLo, amountHi uint64,
	openingLo, openingHi encryption.PedersenOpening,
) (*BatchedGroupedCiphertext3HandlesValidityProofData, error) {
	return invokeProof[BatchedGroupedCiphertext3HandlesValidityProofData]("proof_batched_grouped_ciphertext_3_handles_validity",
		concatPubkeys3(pubkeys), groupedLo[:], groupedHi[:], amountLo, amountHi, openingLo[:], openingHi[:])
}

// invokeProof calls a proof-generation export and parses the pod bytes it
// returns into a fresh proof data value.
func invokeProof[T any, P interface {
	*T
	ProofData
}](name string, parts ...any) (*T, error) {
	out, err := bridge.InvokeWith(name, parts...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	d := new(T)
	if err := readFields(out, P(d).fields()...); err != nil {
		return nil, err
	}
	return d, nil
}

func concatPubkeys2(pubkeys [2]encryption.ElGamalPubkey) []byte {
	out := make([]byte, 0, 64)
	for i := range pubkeys {
		out = append(out, pubkeys[i][:]...)
	}
	return out
}

func concatPubkeys3(pubkeys [3]encryption.ElGamalPubkey) []byte {
	out := make([]byte, 0, 96)
	for i := range pubkeys {
		out = append(out, pubkeys[i][:]...)
	}
	return out
}
