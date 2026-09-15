package encryption

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/solana-foundation/solana-go/v2/programs/token-2022/zkencryption"
)

var ErrZeroSecretKey = errors.New("zk: ElGamal secret key is zero")

type ElGamalKeypair struct {
	Pubkey ElGamalPubkey
	Secret zkencryption.ElGamalSecretKey
}

func (kp *ElGamalKeypair) MarshalBinary() ([]byte, error) {
	if kp.Secret == (zkencryption.ElGamalSecretKey{}) {
		return nil, ErrZeroSecretKey
	}
	out := make([]byte, 0, 64)
	out = append(out, kp.Pubkey[:]...)
	return append(out, kp.Secret[:]...), nil
}

func (kp *ElGamalKeypair) UnmarshalBinary(b []byte) error {
	if len(b) != 64 {
		return fmt.Errorf("zk: guest returned %d-byte keypair, want 64", len(b))
	}
	copy(kp.Pubkey[:], b[:32])
	copy(kp.Secret[:], b[32:64])
	return nil
}

type ElGamalPubkey [32]byte

func (pk ElGamalPubkey) MarshalBinary() ([]byte, error) { return bytes.Clone(pk[:]), nil }

// secretKeyArg marshals a zkencryption.ElGamalSecretKey into a wasm call
// argument, rejecting the zero key.
type secretKeyArg zkencryption.ElGamalSecretKey

func (sk secretKeyArg) MarshalBinary() ([]byte, error) {
	if sk == (secretKeyArg{}) {
		return nil, ErrZeroSecretKey
	}
	return bytes.Clone(sk[:]), nil
}

// ElGamalCiphertext is of form [Pedersen commitment, decrypt handle].
type ElGamalCiphertext [64]byte

func (ct *ElGamalCiphertext) UnmarshalBinary(b []byte) error { return copyExact(ct[:], b) }

func (ct ElGamalCiphertext) MarshalBinary() ([]byte, error) { return bytes.Clone(ct[:]), nil }

type PedersenCommitment [32]byte

func (c *PedersenCommitment) UnmarshalBinary(b []byte) error { return copyExact(c[:], b) }

func (c PedersenCommitment) MarshalBinary() ([]byte, error) { return bytes.Clone(c[:]), nil }

type PedersenOpening [32]byte

func (o *PedersenOpening) UnmarshalBinary(b []byte) error { return copyExact(o[:], b) }

func (o PedersenOpening) MarshalBinary() ([]byte, error) { return bytes.Clone(o[:]), nil }

type PedersenCommitmentOpening struct {
	Commitment PedersenCommitment
	Opening    PedersenOpening
}

func (co *PedersenCommitmentOpening) UnmarshalBinary(b []byte) error {
	if len(b) != 64 {
		return fmt.Errorf("zk: guest returned %d bytes, want 64", len(b))
	}
	copy(co.Commitment[:], b[:32])
	copy(co.Opening[:], b[32:64])
	return nil
}

// Grouped ElGamal ciphertext with 2 decrypt handles
type GroupedElGamalCiphertext2 [96]byte

func (g *GroupedElGamalCiphertext2) UnmarshalBinary(b []byte) error { return copyExact(g[:], b) }

func (g GroupedElGamalCiphertext2) MarshalBinary() ([]byte, error) { return bytes.Clone(g[:]), nil }

// Grouped ElGamal ciphertext with 3 decrypt handles
type GroupedElGamalCiphertext3 [128]byte

func (g *GroupedElGamalCiphertext3) UnmarshalBinary(b []byte) error { return copyExact(g[:], b) }

func (g GroupedElGamalCiphertext3) MarshalBinary() ([]byte, error) { return bytes.Clone(g[:]), nil }

// aeKeyArg is secretKeyArg for the AES-GCM-SIV key zkencryption.AeKey.
type aeKeyArg zkencryption.AeKey

func (k *aeKeyArg) UnmarshalBinary(b []byte) error { return copyExact(k[:], b) }

func (k aeKeyArg) MarshalBinary() ([]byte, error) { return bytes.Clone(k[:]), nil }

// Authenticated encryption of a u64 amount.
type AeCiphertext [36]byte

func (ct *AeCiphertext) UnmarshalBinary(b []byte) error { return copyExact(ct[:], b) }

func (ct AeCiphertext) MarshalBinary() ([]byte, error) { return bytes.Clone(ct[:]), nil }

func copyExact(dst, b []byte) error {
	if len(b) != len(dst) {
		return fmt.Errorf("zk: guest returned %d bytes, want %d", len(b), len(dst))
	}
	copy(dst, b)
	return nil
}
