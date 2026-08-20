package zk

import "fmt"

type ElGamalKeypair struct {
	Pubkey ElGamalPubkey
	Secret ElGamalSecretKey
}

type ElGamalPubkey [32]byte

type ElGamalSecretKey [32]byte

func (kp *ElGamalKeypair) bytes() []byte {
	out := make([]byte, 0, 64)
	out = append(out, kp.Pubkey[:]...)
	return append(out, kp.Secret[:]...)
}

// zeroize clears a transient buffer holding secret material.
func zeroize(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func elGamalKeypairFromBytes(b []byte) (*ElGamalKeypair, error) {
	if len(b) != 64 {
		return nil, fmt.Errorf("zk: guest returned %d-byte keypair, want 64", len(b))
	}
	var kp ElGamalKeypair
	copy(kp.Pubkey[:], b[:32])
	copy(kp.Secret[:], b[32:64])
	return &kp, nil
}

// ElGamalCiphertext is of form [Pedersen commitment, decrypt handle].
type ElGamalCiphertext [64]byte

type PedersenCommitment [32]byte

type PedersenOpening [32]byte

// Grouped ElGamal ciphertext with 2 decrypt handles
type GroupedElGamalCiphertext2 [96]byte

// Grouped ElGamal ciphertext with 3 decrypt handles
type GroupedElGamalCiphertext3 [128]byte

// AES-GCM-SIV key
type AeKey [16]byte

// Authenticated encryption of a u64 amount.
type AeCiphertext [36]byte

// ProofType tags a generated proof for verification.
type ProofType uint32

const (
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

type Proof struct {
	Type ProofType
	Data []byte
}

func (p *Proof) Verify() error {
	return invokeStatus("zk_verify_proof", uint64(p.Type), p.Data)
}
