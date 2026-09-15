package encryption

import (
	"github.com/solana-foundation/solana-go/v2/programs/token-2022/zkencryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/internal/bridge"
)

// NewElGamalKeypair generates a random ElGamal keypair.
//
// For wallet-bound keys, derive the secret deterministically from a signature using ElGamalKeypairFromSecret.
func NewElGamalKeypair() (*ElGamalKeypair, error) {
	kp, err := bridge.InvokeWith[ElGamalKeypair]("elgamal_keypair_new_rand")
	if err != nil {
		return nil, err
	}
	return &kp, nil
}

// ElGamalKeypairFromSecret derives the public key for secret
func ElGamalKeypairFromSecret(secret zkencryption.ElGamalSecretKey) (*ElGamalKeypair, error) {
	kp, err := bridge.InvokeWith[ElGamalKeypair]("elgamal_keypair_from_secret", secretKeyArg(secret))
	if err != nil {
		return nil, err
	}
	return &kp, nil
}

// Encrypt encrypts amount under the public key with a random Pedersen opening.
func (pk ElGamalPubkey) Encrypt(amount uint64) (ElGamalCiphertext, error) {
	return bridge.InvokeWith[ElGamalCiphertext]("elgamal_encrypt", pk, bridge.Scalar(amount))
}

// EncryptWith encrypts amount under the public key, given a Pedersen opening
func (pk ElGamalPubkey) EncryptWith(amount uint64, opening PedersenOpening) (ElGamalCiphertext, error) {
	return bridge.InvokeWith[ElGamalCiphertext]("elgamal_encrypt_with", pk, bridge.Scalar(amount), opening)
}

// DecryptU32 decrypts a ciphertext whose plaintext is known to fit in 32 bits.
func (kp *ElGamalKeypair) DecryptU32(ct ElGamalCiphertext) (uint64, error) {
	amount, err := bridge.InvokeWith[bridge.Scalar]("elgamal_decrypt_u32", secretKeyArg(kp.Secret), ct)
	return uint64(amount), err
}

// CombineLoHiCiphertexts computes lo + 2^bitLength·hi.
func CombineLoHiCiphertexts(lo, hi ElGamalCiphertext, bitLength uint8) (ElGamalCiphertext, error) {
	return bridge.InvokeWith[ElGamalCiphertext]("elgamal_combine_lo_hi_ciphertexts", lo, hi, bridge.Scalar(bitLength))
}

// AddCiphertexts homomorphically adds two ciphertexts encrypted under the same public key.
func AddCiphertexts(a, b ElGamalCiphertext) (ElGamalCiphertext, error) {
	return bridge.InvokeWith[ElGamalCiphertext]("elgamal_add_ciphertexts", a, b)
}

// SubtractCiphertexts homomorphically subtracts ciphertext b from ciphertext a.
func SubtractCiphertexts(a, b ElGamalCiphertext) (ElGamalCiphertext, error) {
	return bridge.InvokeWith[ElGamalCiphertext]("elgamal_sub_ciphertexts", a, b)
}

// AddAmount adds a plaintext amount to the ciphertext.
func (ct ElGamalCiphertext) AddAmount(amount uint64) (ElGamalCiphertext, error) {
	return bridge.InvokeWith[ElGamalCiphertext]("elgamal_add_amount", ct, bridge.Scalar(amount))
}

// SubtractAmount subtracts a plaintext amount from the ciphertext.
func (ct ElGamalCiphertext) SubtractAmount(amount uint64) (ElGamalCiphertext, error) {
	return bridge.InvokeWith[ElGamalCiphertext]("elgamal_sub_amount", ct, bridge.Scalar(amount))
}
