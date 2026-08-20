package zk

import "fmt"

// GroupedElGamalEncrypt2 encrypts amount under two public keys with the given opening.
func GroupedElGamalEncrypt2(pubkeys [2]ElGamalPubkey, amount uint64, opening PedersenOpening) (GroupedElGamalCiphertext2, error) {
	var grouped GroupedElGamalCiphertext2
	out, err := invokeWith("grouped_elgamal_2_encrypt_with", concatPubkeys2(pubkeys), amount, opening[:])
	err = copyOut(grouped[:], out, err)
	return grouped, err
}

// GroupedElGamalEncrypt3 encrypts amount under three public keys with the given opening.
func GroupedElGamalEncrypt3(pubkeys [3]ElGamalPubkey, amount uint64, opening PedersenOpening) (GroupedElGamalCiphertext3, error) {
	var grouped GroupedElGamalCiphertext3
	out, err := invokeWith("grouped_elgamal_3_encrypt_with", concatPubkeys3(pubkeys), amount, opening[:])
	err = copyOut(grouped[:], out, err)
	return grouped, err
}

// ToElGamalCiphertext extracts the regular ElGamal ciphertext for the key at
// the given handle index (the key's position at encryption time).
func (g GroupedElGamalCiphertext2) ToElGamalCiphertext(index int) (ElGamalCiphertext, error) {
	if index < 0 || index > 1 {
		return ElGamalCiphertext{}, fmt.Errorf("zk: handle index %d out of range [0, 1]", index)
	}
	out, err := invokeWith("grouped_ciphertext_2_to_elgamal", g[:], uint64(index))
	return toCiphertext(out, err)
}

// ToElGamalCiphertext extracts the regular ElGamal ciphertext for the key at
// the given handle index (the key's position at encryption time).
func (g GroupedElGamalCiphertext3) ToElGamalCiphertext(index int) (ElGamalCiphertext, error) {
	if index < 0 || index > 2 {
		return ElGamalCiphertext{}, fmt.Errorf("zk: handle index %d out of range [0, 2]", index)
	}
	out, err := invokeWith("grouped_ciphertext_3_to_elgamal", g[:], uint64(index))
	return toCiphertext(out, err)
}

func concatPubkeys2(pubkeys [2]ElGamalPubkey) []byte {
	out := make([]byte, 0, 64)
	for i := range pubkeys {
		out = append(out, pubkeys[i][:]...)
	}
	return out
}

func concatPubkeys3(pubkeys [3]ElGamalPubkey) []byte {
	out := make([]byte, 0, 96)
	for i := range pubkeys {
		out = append(out, pubkeys[i][:]...)
	}
	return out
}
