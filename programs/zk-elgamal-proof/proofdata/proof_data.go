package proofdata

import (
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/internal/bridge"
)

// ProofData is implemented by every proof data type.
type ProofData interface {
	// ProofType tags the proof data for the on-chain verifier.
	ProofType() ProofType
	// Bytes is the pod serialization that a VerifyProof instruction carries.
	Bytes() []byte
	// ContextData is the public statement the proof certifies.
	ContextData() ProofContext
	// Verify checks the proof against its context data, by running
	// solana-zk-sdk's verifier inside the embedded wasm.
	Verify() error
	// UnmarshalBinary parses the pod serialization produced by Bytes.
	encoding.BinaryUnmarshaler
}

// ProofContext is the public statement that a proof certifies.
type ProofContext interface {
	Bytes() []byte
	encoding.BinaryUnmarshaler
}

func verifyProofData(p ProofData) error {
	return bridge.InvokeStatus("zk_verify_proof", bridge.Scalar(p.ProofType()), bridge.Bytes(p.Bytes()))
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
	Pubkey     encryption.ElGamalPubkey
	Ciphertext encryption.ElGamalCiphertext
}

func (p *ZeroCiphertextProofData) ProofType() ProofType           { return ProofTypeZeroCiphertext }
func (p *ZeroCiphertextProofData) Bytes() []byte                  { return podBytes(p) }
func (p *ZeroCiphertextProofData) ContextData() ProofContext      { return &p.Context }
func (p *ZeroCiphertextProofData) Verify() error                  { return verifyProofData(p) }
func (p *ZeroCiphertextProofData) UnmarshalBinary(b []byte) error { return podRead(p, b) }

func (c *ZeroCiphertextProofContext) Bytes() []byte                  { return podBytes(c) }
func (c *ZeroCiphertextProofContext) UnmarshalBinary(b []byte) error { return podRead(c, b) }

// CiphertextCiphertextEqualityProofData proves that two ciphertexts encrypt
// the same amount.
type CiphertextCiphertextEqualityProofData struct {
	Context CiphertextCiphertextEqualityProofContext
	Proof   CiphertextCiphertextEqualityProof
}

type CiphertextCiphertextEqualityProofContext struct {
	FirstPubkey      encryption.ElGamalPubkey
	SecondPubkey     encryption.ElGamalPubkey
	FirstCiphertext  encryption.ElGamalCiphertext
	SecondCiphertext encryption.ElGamalCiphertext
}

func (p *CiphertextCiphertextEqualityProofData) ProofType() ProofType {
	return ProofTypeCiphertextCiphertextEquality
}
func (p *CiphertextCiphertextEqualityProofData) Bytes() []byte                  { return podBytes(p) }
func (p *CiphertextCiphertextEqualityProofData) ContextData() ProofContext      { return &p.Context }
func (p *CiphertextCiphertextEqualityProofData) Verify() error                  { return verifyProofData(p) }
func (p *CiphertextCiphertextEqualityProofData) UnmarshalBinary(b []byte) error { return podRead(p, b) }

func (c *CiphertextCiphertextEqualityProofContext) Bytes() []byte { return podBytes(c) }
func (c *CiphertextCiphertextEqualityProofContext) UnmarshalBinary(b []byte) error {
	return podRead(c, b)
}

// CiphertextCommitmentEqualityProofData proves that ciphertext and commitment
// encode the same amount.
type CiphertextCommitmentEqualityProofData struct {
	Context CiphertextCommitmentEqualityProofContext
	Proof   CiphertextCommitmentEqualityProof
}

type CiphertextCommitmentEqualityProofContext struct {
	Pubkey     encryption.ElGamalPubkey
	Ciphertext encryption.ElGamalCiphertext
	Commitment encryption.PedersenCommitment
}

func (p *CiphertextCommitmentEqualityProofData) ProofType() ProofType {
	return ProofTypeCiphertextCommitmentEquality
}
func (p *CiphertextCommitmentEqualityProofData) Bytes() []byte                  { return podBytes(p) }
func (p *CiphertextCommitmentEqualityProofData) ContextData() ProofContext      { return &p.Context }
func (p *CiphertextCommitmentEqualityProofData) Verify() error                  { return verifyProofData(p) }
func (p *CiphertextCommitmentEqualityProofData) UnmarshalBinary(b []byte) error { return podRead(p, b) }

func (c *CiphertextCommitmentEqualityProofContext) Bytes() []byte { return podBytes(c) }
func (c *CiphertextCommitmentEqualityProofContext) UnmarshalBinary(b []byte) error {
	return podRead(c, b)
}

// PubkeyValidityProofData proves knowledge of the ElGamal secret key for
// pubkey.
type PubkeyValidityProofData struct {
	Context PubkeyValidityProofContext
	Proof   PubkeyValidityProof
}

type PubkeyValidityProofContext struct {
	Pubkey encryption.ElGamalPubkey
}

func (p *PubkeyValidityProofData) ProofType() ProofType           { return ProofTypePubkeyValidity }
func (p *PubkeyValidityProofData) Bytes() []byte                  { return podBytes(p) }
func (p *PubkeyValidityProofData) ContextData() ProofContext      { return &p.Context }
func (p *PubkeyValidityProofData) Verify() error                  { return verifyProofData(p) }
func (p *PubkeyValidityProofData) UnmarshalBinary(b []byte) error { return podRead(p, b) }

func (c *PubkeyValidityProofContext) Bytes() []byte                  { return podBytes(c) }
func (c *PubkeyValidityProofContext) UnmarshalBinary(b []byte) error { return podRead(c, b) }

// PercentageWithCapProofData proves a fee computation: the percentage amount
// is either the correct percentage of the transfer amount or the cap.
type PercentageWithCapProofData struct {
	Context PercentageWithCapProofContext
	Proof   PercentageWithCapProof
}

type PercentageWithCapProofContext struct {
	PercentageCommitment encryption.PedersenCommitment
	DeltaCommitment      encryption.PedersenCommitment
	ClaimedCommitment    encryption.PedersenCommitment
	MaxValue             PodU64
}

func (p *PercentageWithCapProofData) ProofType() ProofType           { return ProofTypePercentageWithCap }
func (p *PercentageWithCapProofData) Bytes() []byte                  { return podBytes(p) }
func (p *PercentageWithCapProofData) ContextData() ProofContext      { return &p.Context }
func (p *PercentageWithCapProofData) Verify() error                  { return verifyProofData(p) }
func (p *PercentageWithCapProofData) UnmarshalBinary(b []byte) error { return podRead(p, b) }

func (c *PercentageWithCapProofContext) Bytes() []byte                  { return podBytes(c) }
func (c *PercentageWithCapProofContext) UnmarshalBinary(b []byte) error { return podRead(c, b) }

// MaxRangeProofCommitments is the number of commitment slots in a batched
// range proof context; unused slots stay zero.
const MaxRangeProofCommitments = 8

// BatchedRangeProofContext is the context shared by the batched range proofs.
type BatchedRangeProofContext struct {
	Commitments [MaxRangeProofCommitments]encryption.PedersenCommitment
	BitLengths  [MaxRangeProofCommitments]uint8
}

func (c *BatchedRangeProofContext) Bytes() []byte                  { return podBytes(c) }
func (c *BatchedRangeProofContext) UnmarshalBinary(b []byte) error { return podRead(c, b) }

// BatchedRangeProofU64Data proves that each committed amount fits in its bit
// length, with the bit lengths summing to 64.
type BatchedRangeProofU64Data struct {
	Context BatchedRangeProofContext
	Proof   RangeProofU64
}

func (p *BatchedRangeProofU64Data) ProofType() ProofType           { return ProofTypeBatchedRangeProofU64 }
func (p *BatchedRangeProofU64Data) Bytes() []byte                  { return podBytes(p) }
func (p *BatchedRangeProofU64Data) ContextData() ProofContext      { return &p.Context }
func (p *BatchedRangeProofU64Data) Verify() error                  { return verifyProofData(p) }
func (p *BatchedRangeProofU64Data) UnmarshalBinary(b []byte) error { return podRead(p, b) }

// BatchedRangeProofU128Data proves that each committed amount fits in its bit
// length, with the bit lengths summing to 128.
type BatchedRangeProofU128Data struct {
	Context BatchedRangeProofContext
	Proof   RangeProofU128
}

func (p *BatchedRangeProofU128Data) ProofType() ProofType           { return ProofTypeBatchedRangeProofU128 }
func (p *BatchedRangeProofU128Data) Bytes() []byte                  { return podBytes(p) }
func (p *BatchedRangeProofU128Data) ContextData() ProofContext      { return &p.Context }
func (p *BatchedRangeProofU128Data) Verify() error                  { return verifyProofData(p) }
func (p *BatchedRangeProofU128Data) UnmarshalBinary(b []byte) error { return podRead(p, b) }

// BatchedRangeProofU256Data proves that each committed amount fits in its bit
// length, with the bit lengths summing to 256.
type BatchedRangeProofU256Data struct {
	Context BatchedRangeProofContext
	Proof   RangeProofU256
}

func (p *BatchedRangeProofU256Data) ProofType() ProofType           { return ProofTypeBatchedRangeProofU256 }
func (p *BatchedRangeProofU256Data) Bytes() []byte                  { return podBytes(p) }
func (p *BatchedRangeProofU256Data) ContextData() ProofContext      { return &p.Context }
func (p *BatchedRangeProofU256Data) Verify() error                  { return verifyProofData(p) }
func (p *BatchedRangeProofU256Data) UnmarshalBinary(b []byte) error { return podRead(p, b) }

// GroupedCiphertext2HandlesValidityProofData proves that a 2-handle grouped
// ciphertext is a valid encryption under both public keys.
type GroupedCiphertext2HandlesValidityProofData struct {
	Context GroupedCiphertext2HandlesValidityProofContext
	Proof   GroupedCiphertext2HandlesValidityProof
}

type GroupedCiphertext2HandlesValidityProofContext struct {
	FirstPubkey       encryption.ElGamalPubkey
	SecondPubkey      encryption.ElGamalPubkey
	GroupedCiphertext encryption.GroupedElGamalCiphertext2
}

func (p *GroupedCiphertext2HandlesValidityProofData) ProofType() ProofType {
	return ProofTypeGroupedCiphertext2HandlesValidity
}
func (p *GroupedCiphertext2HandlesValidityProofData) Bytes() []byte             { return podBytes(p) }
func (p *GroupedCiphertext2HandlesValidityProofData) ContextData() ProofContext { return &p.Context }
func (p *GroupedCiphertext2HandlesValidityProofData) Verify() error             { return verifyProofData(p) }
func (p *GroupedCiphertext2HandlesValidityProofData) UnmarshalBinary(b []byte) error {
	return podRead(p, b)
}

func (c *GroupedCiphertext2HandlesValidityProofContext) Bytes() []byte { return podBytes(c) }
func (c *GroupedCiphertext2HandlesValidityProofContext) UnmarshalBinary(b []byte) error {
	return podRead(c, b)
}

// BatchedGroupedCiphertext2HandlesValidityProofData proves that a lo/hi pair
// of 2-handle grouped ciphertexts is a valid encryption under both public
// keys.
type BatchedGroupedCiphertext2HandlesValidityProofData struct {
	Context BatchedGroupedCiphertext2HandlesValidityProofContext
	Proof   BatchedGroupedCiphertext2HandlesValidityProof
}

type BatchedGroupedCiphertext2HandlesValidityProofContext struct {
	FirstPubkey         encryption.ElGamalPubkey
	SecondPubkey        encryption.ElGamalPubkey
	GroupedCiphertextLo encryption.GroupedElGamalCiphertext2
	GroupedCiphertextHi encryption.GroupedElGamalCiphertext2
}

func (p *BatchedGroupedCiphertext2HandlesValidityProofData) ProofType() ProofType {
	return ProofTypeBatchedGroupedCiphertext2HandlesValidity
}
func (p *BatchedGroupedCiphertext2HandlesValidityProofData) Bytes() []byte { return podBytes(p) }
func (p *BatchedGroupedCiphertext2HandlesValidityProofData) ContextData() ProofContext {
	return &p.Context
}
func (p *BatchedGroupedCiphertext2HandlesValidityProofData) Verify() error {
	return verifyProofData(p)
}
func (p *BatchedGroupedCiphertext2HandlesValidityProofData) UnmarshalBinary(b []byte) error {
	return podRead(p, b)
}

func (c *BatchedGroupedCiphertext2HandlesValidityProofContext) Bytes() []byte { return podBytes(c) }
func (c *BatchedGroupedCiphertext2HandlesValidityProofContext) UnmarshalBinary(b []byte) error {
	return podRead(c, b)
}

// GroupedCiphertext3HandlesValidityProofData proves that a 3-handle grouped
// ciphertext is a valid encryption under all three public keys.
type GroupedCiphertext3HandlesValidityProofData struct {
	Context GroupedCiphertext3HandlesValidityProofContext
	Proof   GroupedCiphertext3HandlesValidityProof
}

type GroupedCiphertext3HandlesValidityProofContext struct {
	FirstPubkey       encryption.ElGamalPubkey
	SecondPubkey      encryption.ElGamalPubkey
	ThirdPubkey       encryption.ElGamalPubkey
	GroupedCiphertext encryption.GroupedElGamalCiphertext3
}

func (p *GroupedCiphertext3HandlesValidityProofData) ProofType() ProofType {
	return ProofTypeGroupedCiphertext3HandlesValidity
}
func (p *GroupedCiphertext3HandlesValidityProofData) Bytes() []byte             { return podBytes(p) }
func (p *GroupedCiphertext3HandlesValidityProofData) ContextData() ProofContext { return &p.Context }
func (p *GroupedCiphertext3HandlesValidityProofData) Verify() error             { return verifyProofData(p) }
func (p *GroupedCiphertext3HandlesValidityProofData) UnmarshalBinary(b []byte) error {
	return podRead(p, b)
}

func (c *GroupedCiphertext3HandlesValidityProofContext) Bytes() []byte { return podBytes(c) }
func (c *GroupedCiphertext3HandlesValidityProofContext) UnmarshalBinary(b []byte) error {
	return podRead(c, b)
}

// BatchedGroupedCiphertext3HandlesValidityProofData proves that a lo/hi pair
// of 3-handle grouped ciphertexts is a valid encryption under all three
// public keys.
type BatchedGroupedCiphertext3HandlesValidityProofData struct {
	Context BatchedGroupedCiphertext3HandlesValidityProofContext
	Proof   BatchedGroupedCiphertext3HandlesValidityProof
}

type BatchedGroupedCiphertext3HandlesValidityProofContext struct {
	FirstPubkey         encryption.ElGamalPubkey
	SecondPubkey        encryption.ElGamalPubkey
	ThirdPubkey         encryption.ElGamalPubkey
	GroupedCiphertextLo encryption.GroupedElGamalCiphertext3
	GroupedCiphertextHi encryption.GroupedElGamalCiphertext3
}

func (p *BatchedGroupedCiphertext3HandlesValidityProofData) ProofType() ProofType {
	return ProofTypeBatchedGroupedCiphertext3HandlesValidity
}
func (p *BatchedGroupedCiphertext3HandlesValidityProofData) Bytes() []byte { return podBytes(p) }
func (p *BatchedGroupedCiphertext3HandlesValidityProofData) ContextData() ProofContext {
	return &p.Context
}
func (p *BatchedGroupedCiphertext3HandlesValidityProofData) Verify() error {
	return verifyProofData(p)
}
func (p *BatchedGroupedCiphertext3HandlesValidityProofData) UnmarshalBinary(b []byte) error {
	return podRead(p, b)
}

func (c *BatchedGroupedCiphertext3HandlesValidityProofContext) Bytes() []byte { return podBytes(c) }
func (c *BatchedGroupedCiphertext3HandlesValidityProofContext) UnmarshalBinary(b []byte) error {
	return podRead(c, b)
}

// ProofType tags proof data for verification.
type ProofType uint32

const (
	ProofTypeUninitialized                            ProofType = 0
	ProofTypeZeroCiphertext                           ProofType = 1
	ProofTypeCiphertextCiphertextEquality             ProofType = 2
	ProofTypeCiphertextCommitmentEquality             ProofType = 3
	ProofTypePubkeyValidity                           ProofType = 4
	ProofTypePercentageWithCap                        ProofType = 5
	ProofTypeBatchedRangeProofU64                     ProofType = 6
	ProofTypeBatchedRangeProofU128                    ProofType = 7
	ProofTypeBatchedRangeProofU256                    ProofType = 8
	ProofTypeGroupedCiphertext2HandlesValidity        ProofType = 9
	ProofTypeBatchedGroupedCiphertext2HandlesValidity ProofType = 10
	ProofTypeGroupedCiphertext3HandlesValidity        ProofType = 11
	ProofTypeBatchedGroupedCiphertext3HandlesValidity ProofType = 12
)

// NewProofData returns an empty proof data value of the type t tags.
func NewProofData(t ProofType) ProofData {
	switch t {
	case ProofTypeZeroCiphertext:
		return new(ZeroCiphertextProofData)
	case ProofTypeCiphertextCiphertextEquality:
		return new(CiphertextCiphertextEqualityProofData)
	case ProofTypeCiphertextCommitmentEquality:
		return new(CiphertextCommitmentEqualityProofData)
	case ProofTypePubkeyValidity:
		return new(PubkeyValidityProofData)
	case ProofTypePercentageWithCap:
		return new(PercentageWithCapProofData)
	case ProofTypeBatchedRangeProofU64:
		return new(BatchedRangeProofU64Data)
	case ProofTypeBatchedRangeProofU128:
		return new(BatchedRangeProofU128Data)
	case ProofTypeBatchedRangeProofU256:
		return new(BatchedRangeProofU256Data)
	case ProofTypeGroupedCiphertext2HandlesValidity:
		return new(GroupedCiphertext2HandlesValidityProofData)
	case ProofTypeBatchedGroupedCiphertext2HandlesValidity:
		return new(BatchedGroupedCiphertext2HandlesValidityProofData)
	case ProofTypeGroupedCiphertext3HandlesValidity:
		return new(GroupedCiphertext3HandlesValidityProofData)
	case ProofTypeBatchedGroupedCiphertext3HandlesValidity:
		return new(BatchedGroupedCiphertext3HandlesValidityProofData)
	}
	return nil
}

// ErrInvalidProofType reports a proof type outside the range the program
// defines. It mirrors the Rust `ProofTypeError::InvalidProofType`.
var ErrInvalidProofType = errors.New("zk: invalid proof type")

// IsValid reports whether t tags one of the proof types the program verifies.
// ProofTypeUninitialized is not valid: it tags the absence of a proof.
func (t ProofType) IsValid() bool {
	return t >= ProofTypeZeroCiphertext && t <= ProofTypeBatchedGroupedCiphertext3HandlesValidity
}

func (t ProofType) String() string {
	switch t {
	case ProofTypeUninitialized:
		return "Uninitialized"
	case ProofTypeZeroCiphertext:
		return "ZeroCiphertext"
	case ProofTypeCiphertextCiphertextEquality:
		return "CiphertextCiphertextEquality"
	case ProofTypeCiphertextCommitmentEquality:
		return "CiphertextCommitmentEquality"
	case ProofTypePubkeyValidity:
		return "PubkeyValidity"
	case ProofTypePercentageWithCap:
		return "PercentageWithCap"
	case ProofTypeBatchedRangeProofU64:
		return "BatchedRangeProofU64"
	case ProofTypeBatchedRangeProofU128:
		return "BatchedRangeProofU128"
	case ProofTypeBatchedRangeProofU256:
		return "BatchedRangeProofU256"
	case ProofTypeGroupedCiphertext2HandlesValidity:
		return "GroupedCiphertext2HandlesValidity"
	case ProofTypeBatchedGroupedCiphertext2HandlesValidity:
		return "BatchedGroupedCiphertext2HandlesValidity"
	case ProofTypeGroupedCiphertext3HandlesValidity:
		return "GroupedCiphertext3HandlesValidity"
	case ProofTypeBatchedGroupedCiphertext3HandlesValidity:
		return "BatchedGroupedCiphertext3HandlesValidity"
	default:
		return fmt.Sprintf("ProofType(%d)", uint32(t))
	}
}
