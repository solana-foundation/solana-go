package encryption

import (
	"github.com/solana-foundation/solana-go/v2/programs/token-2022/zkencryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/internal/bridge"
)

// NewAeKey generates a random AES-GCM-SIV key.
func NewAeKey() (zkencryption.AeKey, error) {
	k, err := bridge.InvokeWith[aeKeyArg]("ae_key_new_rand")
	return zkencryption.AeKey(k), err
}

// AeEncrypt produces the encryption of amount under k.
func AeEncrypt(k zkencryption.AeKey, amount uint64) (AeCiphertext, error) {
	return bridge.InvokeWith[AeCiphertext]("ae_encrypt", aeKeyArg(k), bridge.Scalar(amount))
}

// AeDecrypt recovers the amount from an AE ciphertext.
func AeDecrypt(k zkencryption.AeKey, ct AeCiphertext) (uint64, error) {
	amount, err := bridge.InvokeWith[bridge.Scalar]("ae_decrypt", aeKeyArg(k), ct)
	return uint64(amount), err
}
