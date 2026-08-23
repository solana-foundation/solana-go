package zkencryption

import (
	"crypto/sha3"
	"fmt"

	"filippo.io/edwards25519"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/bip39"
)

// The functions in this file preserve the pre-solana-conf-bal/v1 SHA3-512 key
// derivation. Callers who created keys under the old scheme can use them to
// reproduce the same key material and decrypt balances encrypted before the
// migration. New code must use the solana-conf-bal/v1 functions in ae_key.go
// and elgamal_secret.go; these variants are retained for migration only.

// aeLegacySigningDomain is the domain-separation prefix used by the legacy AE
// derivation. It must match b"AeKey" in solana-zk-sdk.
const aeLegacySigningDomain = "AeKey"

// aeLegacyMinSeedLen mirrors the minimum accepted by the legacy SeedDerivable
// implementation for AeKey in solana-zk-sdk.
const aeLegacyMinSeedLen = AeKeyLen

// elgamalLegacySigningDomain is the domain-separation prefix used by the legacy
// ElGamal derivation. It must match b"ElGamalSecretKey" in solana-zk-sdk.
const elgamalLegacySigningDomain = "ElGamalSecretKey"

// elgamalLegacyMinSeedLen mirrors the minimum accepted by the legacy
// ElGamalSecretKey::from_seed implementation in solana-zk-sdk.
const elgamalLegacyMinSeedLen = ElGamalSecretKeyLen

// Deprecated: legacy SHA3-512 AE derivation. Prefer AeKeyFromSeed
// (solana-conf-bal/v1). Use only to reproduce keys derived under the pre-v1
// scheme.
func AeKeyFromSeedLegacy(seed []byte) (AeKey, error) {
	if len(seed) < aeLegacyMinSeedLen {
		return AeKey{}, ErrSeedTooShort
	}
	if len(seed) > maxSeedLen {
		return AeKey{}, ErrSeedTooLong
	}
	h := sha3.Sum512(seed)
	var out AeKey
	copy(out[:], h[:AeKeyLen])
	return out, nil
}

// Deprecated: legacy SHA3-512 AE derivation from a signature. Prefer
// AeKeyFromSignature (solana-conf-bal/v1).
func AeKeyFromSignatureLegacy(sig solana.Signature) (AeKey, error) {
	h := sha3.Sum512(sig[:])
	return AeKeyFromSeedLegacy(h[:])
}

// Deprecated: legacy SHA3-512 AE derivation from a signer. Prefer
// AeKeyFromSigner (solana-conf-bal/v1).
func AeKeyFromSignerLegacy(signer Signer, publicSeed []byte) (AeKey, error) {
	msg := make([]byte, 0, len(aeLegacySigningDomain)+len(publicSeed))
	msg = append(msg, aeLegacySigningDomain...)
	msg = append(msg, publicSeed...)
	sig, err := signer.Sign(msg)
	if err != nil {
		return AeKey{}, fmt.Errorf("zkencryption: sign legacy AeKey public seed: %w", err)
	}
	if sig == (solana.Signature{}) {
		return AeKey{}, ErrDefaultSignature
	}
	return AeKeyFromSignatureLegacy(sig)
}

// Deprecated: legacy SHA3-512 AE derivation from a BIP39 mnemonic. Prefer
// AeKeyFromSeedPhraseAndPassphrase (solana-conf-bal/v1).
func AeKeyFromSeedPhraseAndPassphraseLegacy(mnemonic, passphrase string) (AeKey, error) {
	return AeKeyFromSeedLegacy(bip39.NewSeed(mnemonic, passphrase))
}

// Deprecated: legacy SHA3-512 ElGamal derivation. Prefer
// ElGamalSecretKeyFromSeed (solana-conf-bal/v1).
func ElGamalSecretKeyFromSeedLegacy(seed []byte) (ElGamalSecretKey, error) {
	if len(seed) < elgamalLegacyMinSeedLen {
		return ElGamalSecretKey{}, ErrSeedTooShort
	}
	if len(seed) > maxSeedLen {
		return ElGamalSecretKey{}, ErrSeedTooLong
	}
	h := sha3.Sum512(seed)
	s, err := edwards25519.NewScalar().SetUniformBytes(h[:])
	if err != nil {
		return ElGamalSecretKey{}, ErrInvalidScalarEncoding
	}
	var out ElGamalSecretKey
	copy(out[:], s.Bytes())
	return out, nil
}

// Deprecated: legacy SHA3-512 ElGamal derivation from a signature. Prefer
// ElGamalSecretKeyFromSignature (solana-conf-bal/v1).
func ElGamalSecretKeyFromSignatureLegacy(sig solana.Signature) (ElGamalSecretKey, error) {
	h := sha3.Sum512(sig[:])
	return ElGamalSecretKeyFromSeedLegacy(h[:])
}

// Deprecated: legacy SHA3-512 ElGamal derivation from a signer. Prefer
// ElGamalSecretKeyFromSigner (solana-conf-bal/v1).
func ElGamalSecretKeyFromSignerLegacy(signer Signer, publicSeed []byte) (ElGamalSecretKey, error) {
	msg := make([]byte, 0, len(elgamalLegacySigningDomain)+len(publicSeed))
	msg = append(msg, elgamalLegacySigningDomain...)
	msg = append(msg, publicSeed...)
	sig, err := signer.Sign(msg)
	if err != nil {
		return ElGamalSecretKey{}, fmt.Errorf("zkencryption: sign legacy ElGamalSecretKey public seed: %w", err)
	}
	if sig == (solana.Signature{}) {
		return ElGamalSecretKey{}, ErrDefaultSignature
	}
	return ElGamalSecretKeyFromSignatureLegacy(sig)
}

// Deprecated: legacy SHA3-512 ElGamal derivation from a BIP39 mnemonic. Prefer
// ElGamalSecretKeyFromSeedPhraseAndPassphrase (solana-conf-bal/v1).
func ElGamalSecretKeyFromSeedPhraseAndPassphraseLegacy(mnemonic, passphrase string) (ElGamalSecretKey, error) {
	return ElGamalSecretKeyFromSeedLegacy(bip39.NewSeed(mnemonic, passphrase))
}
