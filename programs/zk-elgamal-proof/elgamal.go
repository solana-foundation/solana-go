package zk

// NewElGamalKeypair generates a random ElGamal keypair.
//
// For wallet-bound keys, derive the secret deterministically from a signature using ElGamalKeypairFromSecret.
func NewElGamalKeypair() (*ElGamalKeypair, error) {
	out, err := invokeWith("elgamal_keypair_new_rand")
	if err != nil {
		return nil, err
	}
	return elGamalKeypairFromBytes(out)
}

// ElGamalKeypairFromSecret derives the public key for secret
func ElGamalKeypairFromSecret(secret ElGamalSecretKey) (*ElGamalKeypair, error) {
	out, err := invokeWith("elgamal_keypair_from_secret", secret[:])
	if err != nil {
		return nil, err
	}
	return elGamalKeypairFromBytes(out)
}

// Encrypt encrypts amount under the public key with a random Pedersen opening.
func (pk ElGamalPubkey) Encrypt(amount uint64) (ElGamalCiphertext, error) {
	out, err := invokeWith("elgamal_encrypt", pk[:], amount)
	return toCiphertext(out, err)
}

// EncryptWith encrypts amount under the public key, given a Pedersen opening
func (pk ElGamalPubkey) EncryptWith(amount uint64, opening PedersenOpening) (ElGamalCiphertext, error) {
	out, err := invokeWith("elgamal_encrypt_with", pk[:], amount, opening[:])
	return toCiphertext(out, err)
}

// DecryptU32 decrypts a ciphertext whose plaintext is known to fit in 32 bits.
func (kp *ElGamalKeypair) DecryptU32(ct ElGamalCiphertext) (uint64, error) {
	out, err := invokeWith("elgamal_decrypt_u32", kp.Secret[:], ct[:])
	return toAmount(out, err)
}

// CombineLoHiCiphertexts computes lo + 2^bitLength·hi.
func CombineLoHiCiphertexts(lo, hi ElGamalCiphertext, bitLength uint8) (ElGamalCiphertext, error) {
	out, err := invokeWith("elgamal_combine_lo_hi_ciphertexts", lo[:], hi[:], uint64(bitLength))
	return toCiphertext(out, err)
}

// AddCiphertexts homomorphically adds two ciphertexts encrypted under the same public key.
func AddCiphertexts(a, b ElGamalCiphertext) (ElGamalCiphertext, error) {
	out, err := invokeWith("elgamal_add_ciphertexts", a[:], b[:])
	return toCiphertext(out, err)
}

// SubtractCiphertexts homomorphically subtracts ciphertext b from ciphertext a.
func SubtractCiphertexts(a, b ElGamalCiphertext) (ElGamalCiphertext, error) {
	out, err := invokeWith("elgamal_sub_ciphertexts", a[:], b[:])
	return toCiphertext(out, err)
}

// AddAmount adds a plaintext amount to the ciphertext.
func (ct ElGamalCiphertext) AddAmount(amount uint64) (ElGamalCiphertext, error) {
	out, err := invokeWith("elgamal_add_amount", ct[:], amount)
	return toCiphertext(out, err)
}

// SubtractAmount subtracts a plaintext amount from the ciphertext.
func (ct ElGamalCiphertext) SubtractAmount(amount uint64) (ElGamalCiphertext, error) {
	out, err := invokeWith("elgamal_sub_amount", ct[:], amount)
	return toCiphertext(out, err)
}

func toCiphertext(out []byte, err error) (ElGamalCiphertext, error) {
	var ct ElGamalCiphertext
	err = copyOut(ct[:], out, err)
	return ct, err
}
