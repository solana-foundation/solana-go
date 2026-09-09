package zkencryption

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
)

// DeriveConfidentialKeys is the standard confidential-balances derivation: the
// signer signs StandardDerivationMessage (the constant "solana-conf-bal/v1")
// exactly once and both keys (ElGamal and AE) are expanded from that one
// signature. The keys are bound to the wallet alone, one key pair across all
// of the wallet's mints and token accounts, byte-identical to what every other
// standard client (solana-zk-sdk, @solana/zk-sdk, the Token-2022 clients)
// derives for the same wallet. Mirrors derive_confidential_keys in
// solana-zk-sdk.
//
// Signing once also means the two keys always belong together, even with
// non-deterministic signers (hardware wallets, hedged ed25519).
func DeriveConfidentialKeys(signer Signer) (ElGamalSecretKey, AeKey, error) {
	return DeriveConfidentialKeysWithSeed(signer, nil)
}

// DeriveConfidentialKeysWithSeed is the non-standard, seed-scoped derivation:
// both keys from a single ed25519 signature over
// ConfidentialDerivationMessage(publicSeed). Use it only for schemes that
// genuinely need keys scoped more finely than the wallet; keys derived with a
// non-empty seed will not match the standard keys other clients derive for the
// same wallet. For standard wallet-level keys use DeriveConfidentialKeys.
func DeriveConfidentialKeysWithSeed(signer Signer, publicSeed []byte) (ElGamalSecretKey, AeKey, error) {
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
