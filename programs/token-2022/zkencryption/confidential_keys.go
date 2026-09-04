package zkencryption

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
)

// DeriveConfidentialKeys derives both confidential-balances keys (AE and
// ElGamal) from a single ed25519 signature over
// ConfidentialDerivationMessage(publicSeed). The signer signs the message
// exactly once and both keys are expanded from that one signature, so the two
// keys always belong together and are reproducible even with non-deterministic
// signers (hardware wallets, hedged ed25519). Mirrors derive_confidential_keys
// in solana-zk-sdk.
//
// The standard publicSeed is empty (wallet-only keys that match other standard
// clients); non-empty seeds are supported but non-standard.
func DeriveConfidentialKeys(signer Signer, publicSeed []byte) (ElGamalSecretKey, AeKey, error) {
	sig, err := signer.Sign(ConfidentialDerivationMessage(publicSeed))
	if err != nil {
		return ElGamalSecretKey{}, AeKey{}, fmt.Errorf("zkencryption: sign confidential-balances public seed: %w", err)
	}
	return DeriveConfidentialKeysFromSignature(sig)
}

// DeriveConfidentialKeysFromSignature derives both confidential-balances keys
// (AE and ElGamal) from an ed25519 signature over
// ConfidentialDerivationMessage. Mirrors derive_confidential_keys_from_signature
// in solana-zk-sdk. An all-zero (default) signature is rejected, matching the
// Rust implementation.
func DeriveConfidentialKeysFromSignature(sig solana.Signature) (ElGamalSecretKey, AeKey, error) {
	if sig == (solana.Signature{}) {
		return ElGamalSecretKey{}, AeKey{}, ErrDefaultSignature
	}
	ae, err := deriveAeKey(sig[:])
	if err != nil {
		return ElGamalSecretKey{}, AeKey{}, err
	}
	el, err := deriveElGamalSecretKey(sig[:])
	if err != nil {
		return ElGamalSecretKey{}, AeKey{}, err
	}
	return el, ae, nil
}
