package encryption

import (
	"fmt"

	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/internal/bridge"
)

// GroupedElGamalEncrypt2 encrypts amount under two public keys with the given opening.
func GroupedElGamalEncrypt2(pubkeys [2]ElGamalPubkey, amount uint64, opening PedersenOpening) (GroupedElGamalCiphertext2, error) {
	return bridge.InvokeWith[GroupedElGamalCiphertext2]("grouped_elgamal_2_encrypt_with", bridge.Slice[ElGamalPubkey](pubkeys[:]), bridge.Scalar(amount), opening)
}

// GroupedElGamalEncrypt3 encrypts amount under three public keys with the given opening.
func GroupedElGamalEncrypt3(pubkeys [3]ElGamalPubkey, amount uint64, opening PedersenOpening) (GroupedElGamalCiphertext3, error) {
	return bridge.InvokeWith[GroupedElGamalCiphertext3]("grouped_elgamal_3_encrypt_with", bridge.Slice[ElGamalPubkey](pubkeys[:]), bridge.Scalar(amount), opening)
}

// ToElGamalCiphertext extracts the regular ElGamal ciphertext for the key at
// the given handle index (the key's position at encryption time).
func (g GroupedElGamalCiphertext2) ToElGamalCiphertext(index int) (ElGamalCiphertext, error) {
	if index < 0 || index > 1 {
		return ElGamalCiphertext{}, fmt.Errorf("zk: handle index %d out of range [0, 1]", index)
	}
	return bridge.InvokeWith[ElGamalCiphertext]("grouped_ciphertext_2_to_elgamal", g, bridge.Scalar(index))
}

// ToElGamalCiphertext extracts the regular ElGamal ciphertext for the key at
// the given handle index (the key's position at encryption time).
func (g GroupedElGamalCiphertext3) ToElGamalCiphertext(index int) (ElGamalCiphertext, error) {
	if index < 0 || index > 2 {
		return ElGamalCiphertext{}, fmt.Errorf("zk: handle index %d out of range [0, 2]", index)
	}
	return bridge.InvokeWith[ElGamalCiphertext]("grouped_ciphertext_3_to_elgamal", g, bridge.Scalar(index))
}
