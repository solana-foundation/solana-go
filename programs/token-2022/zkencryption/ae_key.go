package zkencryption

import (
	"crypto/sha3"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/tyler-smith/go-bip39"
)

// AeKeyLen is the byte length of an authenticated-encryption key (AES-128-GCM-SIV).
const AeKeyLen = 16

// aeSigningDomain is the domain-separation prefix prepended to the public
// seed before signing. It must match b"AeKey" in solana-zk-sdk.
const aeSigningDomain = "AeKey"

// minAeSeedLen / maxAeSeedLen mirror the bounds enforced in solana-zk-sdk's
// SeedDerivable::from_seed implementation for AeKey.
const (
	minAeSeedLen = AeKeyLen
	maxAeSeedLen = 65535
)

// AeKey is a 128-bit authenticated-encryption key used by the Token-2022
// confidential-transfer extension to encrypt u64 amounts under AES-128-GCM-SIV.
type AeKey [AeKeyLen]byte

// AeKeyFromSeed derives an AeKey from an entropy seed by hashing the seed with
// SHA3-512 and taking the first 16 bytes, matching SeedDerivable::from_seed in
// solana-zk-sdk.
func AeKeyFromSeed(seed []byte) (AeKey, error) {
	if len(seed) < minAeSeedLen {
		return AeKey{}, ErrSeedTooShort
	}
	if len(seed) > maxAeSeedLen {
		return AeKey{}, ErrSeedTooLong
	}
	h := sha3.Sum512(seed)
	var out AeKey
	copy(out[:], h[:AeKeyLen])
	return out, nil
}

// AeKeyFromSignature derives an AeKey from an ed25519 signature by using
// SHA3-512(signature) as the seed. Mirrors AeKey::seed_from_signature +
// from_seed in solana-zk-sdk. No default-signature check is performed here;
// use AeKeyFromSigner if the signature originates from a local signer.
func AeKeyFromSignature(sig solana.Signature) (AeKey, error) {
	h := sha3.Sum512(sig[:])
	return AeKeyFromSeed(h[:])
}

// AeKeyFromSigner deterministically derives an AeKey from a Solana signer and
// a public seed. The signer signs b"AeKey" || publicSeed; the signature is
// then hashed with SHA3-512 and the result fed into AeKeyFromSeed. An
// all-zero (default) signature is rejected, matching the Rust implementation.
func AeKeyFromSigner(signer Signer, publicSeed []byte) (AeKey, error) {
	msg := make([]byte, 0, len(aeSigningDomain)+len(publicSeed))
	msg = append(msg, aeSigningDomain...)
	msg = append(msg, publicSeed...)

	sig, err := signer.Sign(msg)
	if err != nil {
		return AeKey{}, fmt.Errorf("zkencryption: sign AeKey public seed: %w", err)
	}
	if sig == (solana.Signature{}) {
		return AeKey{}, ErrDefaultSignature
	}
	return AeKeyFromSignature(sig)
}

// AeKeyFromSeedPhraseAndPassphrase derives an AeKey from a BIP39 mnemonic and
// an optional passphrase using the standard BIP39 PBKDF2-HMAC-SHA512 seed
// derivation (2048 iterations, 64-byte output), matching
// solana_seed_phrase::generate_seed_from_seed_phrase_and_passphrase. Solana
// does not validate the mnemonic checksum at this layer, and neither do we.
func AeKeyFromSeedPhraseAndPassphrase(mnemonic, passphrase string) (AeKey, error) {
	return AeKeyFromSeed(bip39.NewSeed(mnemonic, passphrase))
}
