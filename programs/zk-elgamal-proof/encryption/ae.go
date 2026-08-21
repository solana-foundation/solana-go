package encryption

import "github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/bridge"

// NewAeKey generates a random AES-GCM-SIV key.
func NewAeKey() (AeKey, error) {
	var key AeKey
	out, err := bridge.InvokeWith("ae_key_new_rand")
	err = bridge.CopyOut(key[:], out, err)
	return key, err
}

// Encrypt produces the encryption of amount.
func (k AeKey) Encrypt(amount uint64) (AeCiphertext, error) {
	var ct AeCiphertext
	out, err := bridge.InvokeWith("ae_encrypt", k[:], amount)
	err = bridge.CopyOut(ct[:], out, err)
	return ct, err
}

// Decrypt recovers the amount from an AE ciphertext.
func (k AeKey) Decrypt(ct AeCiphertext) (uint64, error) {
	out, err := bridge.InvokeWith("ae_decrypt", k[:], ct[:])
	return bridge.ToAmount(out, err)
}
