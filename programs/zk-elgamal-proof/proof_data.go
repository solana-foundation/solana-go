package zk

import (
	"encoding/binary"
	"fmt"
)

// ProofData is implemented by every proof data type.
type ProofData interface {
	// ProofType tags the proof data for the on-chain verifier.
	ProofType() ProofType
	// Bytes is the pod serialization that a VerifyProof instruction carries.
	Bytes() []byte
	// Verify checks the proof against its context data.
	Verify() error

	// fields lists the pod fields in serialization order.
	fields() [][]byte
}

func verifyProofData(p ProofData) error {
	return invokeStatus("zk_verify_proof", uint64(p.ProofType()), p.Bytes())
}

// Pod sigma and range proofs, sized like their Rust counterparts.
type (
	ZeroCiphertextProof                           [96]byte
	CiphertextCiphertextEqualityProof             [224]byte
	CiphertextCommitmentEqualityProof             [192]byte
	PubkeyValidityProof                           [64]byte
	PercentageWithCapProof                        [256]byte
	RangeProofU64                                 [672]byte
	RangeProofU128                                [736]byte
	RangeProofU256                                [800]byte
	GroupedCiphertext2HandlesValidityProof        [160]byte
	BatchedGroupedCiphertext2HandlesValidityProof [160]byte
	GroupedCiphertext3HandlesValidityProof        [192]byte
	BatchedGroupedCiphertext3HandlesValidityProof [192]byte
)

// PodU64 is a little-endian u64.
type PodU64 [8]byte

func (u PodU64) Uint64() uint64 { return binary.LittleEndian.Uint64(u[:]) }

// ZeroCiphertextProofData proves that ciphertext encrypts zero under pubkey.
type ZeroCiphertextProofData struct {
	Context ZeroCiphertextProofContext
	Proof   ZeroCiphertextProof
}

type ZeroCiphertextProofContext struct {
	Pubkey     ElGamalPubkey
	Ciphertext ElGamalCiphertext
}

func (p *ZeroCiphertextProofData) fields() [][]byte {
	return [][]byte{p.Context.Pubkey[:], p.Context.Ciphertext[:], p.Proof[:]}
}
func (p *ZeroCiphertextProofData) ProofType() ProofType { return ProofTypeZeroCiphertext }
func (p *ZeroCiphertextProofData) Bytes() []byte        { return concatFields(p.fields()...) }
func (p *ZeroCiphertextProofData) Verify() error        { return verifyProofData(p) }

// CiphertextCiphertextEqualityProofData proves that two ciphertexts encrypt
// the same amount.
type CiphertextCiphertextEqualityProofData struct {
	Context CiphertextCiphertextEqualityProofContext
	Proof   CiphertextCiphertextEqualityProof
}

type CiphertextCiphertextEqualityProofContext struct {
	FirstPubkey      ElGamalPubkey
	SecondPubkey     ElGamalPubkey
	FirstCiphertext  ElGamalCiphertext
	SecondCiphertext ElGamalCiphertext
}

func (p *CiphertextCiphertextEqualityProofData) fields() [][]byte {
	return [][]byte{p.Context.FirstPubkey[:], p.Context.SecondPubkey[:],
		p.Context.FirstCiphertext[:], p.Context.SecondCiphertext[:], p.Proof[:]}
}
func (p *CiphertextCiphertextEqualityProofData) ProofType() ProofType {
	return ProofTypeCiphertextCiphertextEquality
}
func (p *CiphertextCiphertextEqualityProofData) Bytes() []byte { return concatFields(p.fields()...) }
func (p *CiphertextCiphertextEqualityProofData) Verify() error { return verifyProofData(p) }

// CiphertextCommitmentEqualityProofData proves that ciphertext and commitment
// encode the same amount.
type CiphertextCommitmentEqualityProofData struct {
	Context CiphertextCommitmentEqualityProofContext
	Proof   CiphertextCommitmentEqualityProof
}

type CiphertextCommitmentEqualityProofContext struct {
	Pubkey     ElGamalPubkey
	Ciphertext ElGamalCiphertext
	Commitment PedersenCommitment
}

func (p *CiphertextCommitmentEqualityProofData) fields() [][]byte {
	return [][]byte{p.Context.Pubkey[:], p.Context.Ciphertext[:], p.Context.Commitment[:], p.Proof[:]}
}
func (p *CiphertextCommitmentEqualityProofData) ProofType() ProofType {
	return ProofTypeCiphertextCommitmentEquality
}
func (p *CiphertextCommitmentEqualityProofData) Bytes() []byte { return concatFields(p.fields()...) }
func (p *CiphertextCommitmentEqualityProofData) Verify() error { return verifyProofData(p) }

// PubkeyValidityProofData proves knowledge of the ElGamal secret key for
// pubkey.
type PubkeyValidityProofData struct {
	Context PubkeyValidityProofContext
	Proof   PubkeyValidityProof
}

type PubkeyValidityProofContext struct {
	Pubkey ElGamalPubkey
}

func (p *PubkeyValidityProofData) fields() [][]byte {
	return [][]byte{p.Context.Pubkey[:], p.Proof[:]}
}
func (p *PubkeyValidityProofData) ProofType() ProofType { return ProofTypePubkeyValidity }
func (p *PubkeyValidityProofData) Bytes() []byte        { return concatFields(p.fields()...) }
func (p *PubkeyValidityProofData) Verify() error        { return verifyProofData(p) }

// PercentageWithCapProofData proves a fee computation: the percentage amount
// is either the correct percentage of the delta amount or the cap.
type PercentageWithCapProofData struct {
	Context PercentageWithCapProofContext
	Proof   PercentageWithCapProof
}

type PercentageWithCapProofContext struct {
	PercentageCommitment PedersenCommitment
	DeltaCommitment      PedersenCommitment
	ClaimedCommitment    PedersenCommitment
	MaxValue             PodU64
}

func (p *PercentageWithCapProofData) fields() [][]byte {
	return [][]byte{p.Context.PercentageCommitment[:], p.Context.DeltaCommitment[:],
		p.Context.ClaimedCommitment[:], p.Context.MaxValue[:], p.Proof[:]}
}
func (p *PercentageWithCapProofData) ProofType() ProofType { return ProofTypePercentageWithCap }
func (p *PercentageWithCapProofData) Bytes() []byte        { return concatFields(p.fields()...) }
func (p *PercentageWithCapProofData) Verify() error        { return verifyProofData(p) }

// MaxRangeProofCommitments is the number of commitment slots in a batched
// range proof context; unused slots stay zero.
const MaxRangeProofCommitments = 8

// BatchedRangeProofContext is the context shared by the batched range proofs.
type BatchedRangeProofContext struct {
	Commitments [MaxRangeProofCommitments]PedersenCommitment
	BitLengths  [MaxRangeProofCommitments]uint8
}

func (c *BatchedRangeProofContext) fields() [][]byte {
	fs := make([][]byte, 0, len(c.Commitments)+1)
	for i := range c.Commitments {
		fs = append(fs, c.Commitments[i][:])
	}
	return append(fs, c.BitLengths[:])
}

// BatchedRangeProofU64Data proves that each committed amount fits in its bit
// length, with the bit lengths summing to 64.
type BatchedRangeProofU64Data struct {
	Context BatchedRangeProofContext
	Proof   RangeProofU64
}

func (p *BatchedRangeProofU64Data) fields() [][]byte {
	return append(p.Context.fields(), p.Proof[:])
}
func (p *BatchedRangeProofU64Data) ProofType() ProofType { return ProofTypeBatchedRangeProofU64 }
func (p *BatchedRangeProofU64Data) Bytes() []byte        { return concatFields(p.fields()...) }
func (p *BatchedRangeProofU64Data) Verify() error        { return verifyProofData(p) }

// BatchedRangeProofU128Data proves that each committed amount fits in its bit
// length, with the bit lengths summing to 128.
type BatchedRangeProofU128Data struct {
	Context BatchedRangeProofContext
	Proof   RangeProofU128
}

func (p *BatchedRangeProofU128Data) fields() [][]byte {
	return append(p.Context.fields(), p.Proof[:])
}
func (p *BatchedRangeProofU128Data) ProofType() ProofType { return ProofTypeBatchedRangeProofU128 }
func (p *BatchedRangeProofU128Data) Bytes() []byte        { return concatFields(p.fields()...) }
func (p *BatchedRangeProofU128Data) Verify() error        { return verifyProofData(p) }

// BatchedRangeProofU256Data proves that each committed amount fits in its bit
// length, with the bit lengths summing to 256.
type BatchedRangeProofU256Data struct {
	Context BatchedRangeProofContext
	Proof   RangeProofU256
}

func (p *BatchedRangeProofU256Data) fields() [][]byte {
	return append(p.Context.fields(), p.Proof[:])
}
func (p *BatchedRangeProofU256Data) ProofType() ProofType { return ProofTypeBatchedRangeProofU256 }
func (p *BatchedRangeProofU256Data) Bytes() []byte        { return concatFields(p.fields()...) }
func (p *BatchedRangeProofU256Data) Verify() error        { return verifyProofData(p) }

// GroupedCiphertext2HandlesValidityProofData proves that a 2-handle grouped
// ciphertext is a valid encryption under both public keys.
type GroupedCiphertext2HandlesValidityProofData struct {
	Context GroupedCiphertext2HandlesValidityProofContext
	Proof   GroupedCiphertext2HandlesValidityProof
}

type GroupedCiphertext2HandlesValidityProofContext struct {
	FirstPubkey       ElGamalPubkey
	SecondPubkey      ElGamalPubkey
	GroupedCiphertext GroupedElGamalCiphertext2
}

func (p *GroupedCiphertext2HandlesValidityProofData) fields() [][]byte {
	return [][]byte{p.Context.FirstPubkey[:], p.Context.SecondPubkey[:],
		p.Context.GroupedCiphertext[:], p.Proof[:]}
}
func (p *GroupedCiphertext2HandlesValidityProofData) ProofType() ProofType {
	return ProofTypeGroupedCiphertext2HandlesValidity
}
func (p *GroupedCiphertext2HandlesValidityProofData) Bytes() []byte {
	return concatFields(p.fields()...)
}
func (p *GroupedCiphertext2HandlesValidityProofData) Verify() error { return verifyProofData(p) }

// BatchedGroupedCiphertext2HandlesValidityProofData proves that a lo/hi pair
// of 2-handle grouped ciphertexts is a valid encryption under both public
// keys.
type BatchedGroupedCiphertext2HandlesValidityProofData struct {
	Context BatchedGroupedCiphertext2HandlesValidityProofContext
	Proof   BatchedGroupedCiphertext2HandlesValidityProof
}

type BatchedGroupedCiphertext2HandlesValidityProofContext struct {
	FirstPubkey         ElGamalPubkey
	SecondPubkey        ElGamalPubkey
	GroupedCiphertextLo GroupedElGamalCiphertext2
	GroupedCiphertextHi GroupedElGamalCiphertext2
}

func (p *BatchedGroupedCiphertext2HandlesValidityProofData) fields() [][]byte {
	return [][]byte{p.Context.FirstPubkey[:], p.Context.SecondPubkey[:],
		p.Context.GroupedCiphertextLo[:], p.Context.GroupedCiphertextHi[:], p.Proof[:]}
}
func (p *BatchedGroupedCiphertext2HandlesValidityProofData) ProofType() ProofType {
	return ProofTypeBatchedGroupedCiphertext2HandlesValidity
}
func (p *BatchedGroupedCiphertext2HandlesValidityProofData) Bytes() []byte {
	return concatFields(p.fields()...)
}
func (p *BatchedGroupedCiphertext2HandlesValidityProofData) Verify() error {
	return verifyProofData(p)
}

// GroupedCiphertext3HandlesValidityProofData proves that a 3-handle grouped
// ciphertext is a valid encryption under all three public keys.
type GroupedCiphertext3HandlesValidityProofData struct {
	Context GroupedCiphertext3HandlesValidityProofContext
	Proof   GroupedCiphertext3HandlesValidityProof
}

type GroupedCiphertext3HandlesValidityProofContext struct {
	FirstPubkey       ElGamalPubkey
	SecondPubkey      ElGamalPubkey
	ThirdPubkey       ElGamalPubkey
	GroupedCiphertext GroupedElGamalCiphertext3
}

func (p *GroupedCiphertext3HandlesValidityProofData) fields() [][]byte {
	return [][]byte{p.Context.FirstPubkey[:], p.Context.SecondPubkey[:],
		p.Context.ThirdPubkey[:], p.Context.GroupedCiphertext[:], p.Proof[:]}
}
func (p *GroupedCiphertext3HandlesValidityProofData) ProofType() ProofType {
	return ProofTypeGroupedCiphertext3HandlesValidity
}
func (p *GroupedCiphertext3HandlesValidityProofData) Bytes() []byte {
	return concatFields(p.fields()...)
}
func (p *GroupedCiphertext3HandlesValidityProofData) Verify() error { return verifyProofData(p) }

// BatchedGroupedCiphertext3HandlesValidityProofData proves that a lo/hi pair
// of 3-handle grouped ciphertexts is a valid encryption under all three
// public keys.
type BatchedGroupedCiphertext3HandlesValidityProofData struct {
	Context BatchedGroupedCiphertext3HandlesValidityProofContext
	Proof   BatchedGroupedCiphertext3HandlesValidityProof
}

type BatchedGroupedCiphertext3HandlesValidityProofContext struct {
	FirstPubkey         ElGamalPubkey
	SecondPubkey        ElGamalPubkey
	ThirdPubkey         ElGamalPubkey
	GroupedCiphertextLo GroupedElGamalCiphertext3
	GroupedCiphertextHi GroupedElGamalCiphertext3
}

func (p *BatchedGroupedCiphertext3HandlesValidityProofData) fields() [][]byte {
	return [][]byte{p.Context.FirstPubkey[:], p.Context.SecondPubkey[:], p.Context.ThirdPubkey[:],
		p.Context.GroupedCiphertextLo[:], p.Context.GroupedCiphertextHi[:], p.Proof[:]}
}
func (p *BatchedGroupedCiphertext3HandlesValidityProofData) ProofType() ProofType {
	return ProofTypeBatchedGroupedCiphertext3HandlesValidity
}
func (p *BatchedGroupedCiphertext3HandlesValidityProofData) Bytes() []byte {
	return concatFields(p.fields()...)
}
func (p *BatchedGroupedCiphertext3HandlesValidityProofData) Verify() error {
	return verifyProofData(p)
}

func concatFields(fields ...[]byte) []byte {
	n := 0
	for _, f := range fields {
		n += len(f)
	}
	out := make([]byte, 0, n)
	for _, f := range fields {
		out = append(out, f...)
	}
	return out
}

func readFields(b []byte, fields ...[]byte) error {
	n := 0
	for _, f := range fields {
		n += len(f)
	}
	if len(b) != n {
		return fmt.Errorf("zk: guest returned %d-byte proof data, want %d", len(b), n)
	}
	for _, f := range fields {
		copy(f, b[:len(f)])
		b = b[len(f):]
	}
	return nil
}
