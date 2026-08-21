package encryption

import "fmt"

type ElGamalKeypair struct {
	Pubkey ElGamalPubkey
	Secret ElGamalSecretKey
}

type ElGamalPubkey [32]byte

type ElGamalSecretKey [32]byte

func (kp *ElGamalKeypair) Bytes() []byte {
	out := make([]byte, 0, 64)
	out = append(out, kp.Pubkey[:]...)
	return append(out, kp.Secret[:]...)
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
