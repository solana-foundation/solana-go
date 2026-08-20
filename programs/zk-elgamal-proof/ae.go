package zk

// NewAeKey generates a random AES-GCM-SIV key.
func NewAeKey() (AeKey, error) {
	var key AeKey
	out, err := invokeWith("ae_key_new_rand")
	err = copyOut(key[:], out, err)
	return key, err
}

// Encrypt produces the encryption of amount.
func (k AeKey) Encrypt(amount uint64) (AeCiphertext, error) {
	var ct AeCiphertext
	out, err := invokeWith("ae_encrypt", k[:], amount)
	err = copyOut(ct[:], out, err)
	return ct, err
}

// Decrypt recovers the amount from an AE ciphertext.
func (k AeKey) Decrypt(ct AeCiphertext) (uint64, error) {
	out, err := invokeWith("ae_decrypt", k[:], ct[:])
	return toAmount(out, err)
}
