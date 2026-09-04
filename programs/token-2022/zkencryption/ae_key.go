package zkencryption

import (
	"crypto/hkdf"
	"crypto/sha512"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/bip39"
)

// AeKeyLen is the byte length of an authenticated-encryption key (AES-128-GCM-SIV).
const AeKeyLen = 16

// aeInfo is the HKDF info string that scopes the AE key expansion, matching
// AE_HKDF_INFO in solana-zk-sdk.
const aeInfo = "ae"

// aeMinSeedLen is the minimum seed length accepted by the AE derivation. It
// mirrors MINIMUM_SEED_LEN in solana-zk-sdk's derive_confidential_keys_from_ikm
// (AE_KEY_LEN = 16), which is independent from the ElGamal minimum of 32.
const aeMinSeedLen = AeKeyLen

// AeKey is a 128-bit authenticated-encryption key used by the Token-2022
// confidential-transfer extension to encrypt u64 amounts under AES-128-GCM-SIV.
type AeKey [AeKeyLen]byte

// deriveAeKey implements the solana-conf-bal/v1 derivation:
// HKDF-SHA512(salt=SigningDomain, ikm).expand(info="ae", 16), matching
// derive_confidential_keys_from_ikm in solana-zk-sdk.
func deriveAeKey(ikm []byte) (AeKey, error) {
	if len(ikm) < aeMinSeedLen {
		return AeKey{}, ErrSeedTooShort
	}
	if len(ikm) > maxSeedLen {
		return AeKey{}, ErrSeedTooLong
	}

	key, err := hkdf.Key(sha512.New, ikm, []byte(SigningDomain), aeInfo, AeKeyLen)
	if err != nil {
		return AeKey{}, fmt.Errorf("zkencryption: HKDF expand ae: %w", err)
	}

	var out AeKey
	copy(out[:], key)
	// key holds expanded secret material; scrub it before it leaves scope.
	clear(key)
	return out, nil
}

// AeKeyFromSeed derives an AeKey from raw input key material, matching
// derive_confidential_keys_from_ikm in solana-zk-sdk.
func AeKeyFromSeed(seed []byte) (AeKey, error) {
	return deriveAeKey(seed)
}

// AeKeyFromSignature derives an AeKey from an ed25519 signature over
// ConfidentialDerivationMessage. Mirrors derive_confidential_keys_from_signature
// in solana-zk-sdk. An all-zero (default) signature is rejected, matching the
// Rust implementation.
func AeKeyFromSignature(sig solana.Signature) (AeKey, error) {
	if sig == (solana.Signature{}) {
		return AeKey{}, ErrDefaultSignature
	}
	return deriveAeKey(sig[:])
}

// AeKeyFromSigner deterministically derives an AeKey from a Solana signer and
// a public seed. The signer signs b"solana-conf-bal/v1" || publicSeed (see
// ConfidentialDerivationMessage); the signature is fed through the HKDF-SHA512
// solana-conf-bal/v1 derivation. The all-zero signature rejection lives in
// AeKeyFromSignature. The standard publicSeed is empty (wallet-only keys that
// match other standard clients); non-empty seeds are supported but
// non-standard.
func AeKeyFromSigner(signer Signer, publicSeed []byte) (AeKey, error) {
	sig, err := signer.Sign(ConfidentialDerivationMessage(publicSeed))
	if err != nil {
		return AeKey{}, fmt.Errorf("zkencryption: sign confidential-balances public seed: %w", err)
	}
	return AeKeyFromSignature(sig)
}

// AeKeyFromSeedPhraseAndPassphrase derives an AeKey from a BIP39 mnemonic and
// an optional passphrase using the standard BIP39 PBKDF2-HMAC-SHA512 seed
// derivation (2048 iterations, 64-byte output). The seed is used directly as
// the HKDF input key material.
func AeKeyFromSeedPhraseAndPassphrase(mnemonic, passphrase string) (AeKey, error) {
	return AeKeyFromSeed(bip39.NewSeed(mnemonic, passphrase))
}
