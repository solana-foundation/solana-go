package zkencryption

import (
	"crypto/sha3"
	"fmt"

	"filippo.io/edwards25519"
	"github.com/gagliardetto/solana-go"
	"github.com/tyler-smith/go-bip39"
)

// ElGamalSecretKeyLen is the canonical length of an ElGamal secret scalar
// encoded in little-endian form (matches curve25519-dalek Scalar::as_bytes).
const ElGamalSecretKeyLen = 32

// elGamalSigningDomain is the domain-separation prefix prepended to the
// public seed before signing. It must match b"ElGamalSecretKey" in
// solana-zk-sdk.
const elGamalSigningDomain = "ElGamalSecretKey"

// minElGamalSeedLen / maxElGamalSeedLen mirror the bounds enforced in
// solana-zk-sdk's ElGamalSecretKey::from_seed implementation.
const (
	minElGamalSeedLen = ElGamalSecretKeyLen
	maxElGamalSeedLen = 65535
)

// ElGamalSecretKey is a canonical little-endian encoding of a Ristretto/Ed25519
// scalar mod ell. It is the Token-2022 confidential-transfer ElGamal private
// key; byte-for-byte equivalent to ElGamalSecretKey::as_bytes in solana-zk-sdk.
type ElGamalSecretKey [ElGamalSecretKeyLen]byte

// ElGamalSecretKeyFromSeed derives an ElGamal secret key from an entropy seed
// by computing Scalar::from_bytes_mod_order_wide(SHA3-512(seed)), matching
// curve25519-dalek's Scalar::hash_from_bytes::<Sha3_512>.
func ElGamalSecretKeyFromSeed(seed []byte) (ElGamalSecretKey, error) {
	if len(seed) < minElGamalSeedLen {
		return ElGamalSecretKey{}, ErrSeedTooShort
	}
	if len(seed) > maxElGamalSeedLen {
		return ElGamalSecretKey{}, ErrSeedTooLong
	}

	h := sha3.Sum512(seed)
	// SetUniformBytes only errors on wrong input length; Sum512 always
	// returns 64 bytes, so this branch is unreachable in practice but kept
	// to avoid an implicit panic if the upstream contract ever changes.
	s, err := edwards25519.NewScalar().SetUniformBytes(h[:])
	if err != nil {
		return ElGamalSecretKey{}, ErrInvalidScalarEncoding
	}

	var out ElGamalSecretKey
	copy(out[:], s.Bytes())
	return out, nil
}

// ElGamalSecretKeyFromSignature derives an ElGamal secret key from an ed25519
// signature by using SHA3-512(signature) as the seed. Mirrors
// ElGamalSecretKey::seed_from_signature + from_seed in solana-zk-sdk.
func ElGamalSecretKeyFromSignature(sig solana.Signature) (ElGamalSecretKey, error) {
	h := sha3.Sum512(sig[:])
	return ElGamalSecretKeyFromSeed(h[:])
}

// ElGamalSecretKeyFromSigner deterministically derives an ElGamal secret key
// from a Solana signer and a public seed. The signer signs
// b"ElGamalSecretKey" || publicSeed; the signature is hashed with SHA3-512
// and fed into ElGamalSecretKeyFromSeed. An all-zero (default) signature is
// rejected to match the Rust implementation.
func ElGamalSecretKeyFromSigner(signer Signer, publicSeed []byte) (ElGamalSecretKey, error) {
	msg := make([]byte, 0, len(elGamalSigningDomain)+len(publicSeed))
	msg = append(msg, elGamalSigningDomain...)
	msg = append(msg, publicSeed...)

	sig, err := signer.Sign(msg)
	if err != nil {
		return ElGamalSecretKey{}, fmt.Errorf("zkencryption: sign ElGamalSecretKey public seed: %w", err)
	}
	if sig == (solana.Signature{}) {
		return ElGamalSecretKey{}, ErrDefaultSignature
	}
	return ElGamalSecretKeyFromSignature(sig)
}

// ElGamalSecretKeyFromSeedPhraseAndPassphrase derives an ElGamal secret key
// from a BIP39 mnemonic and an optional passphrase, matching
// solana_seed_phrase's PBKDF2-HMAC-SHA512 derivation. Solana does not
// validate the mnemonic checksum at this layer, and neither do we.
func ElGamalSecretKeyFromSeedPhraseAndPassphrase(mnemonic, passphrase string) (ElGamalSecretKey, error) {
	return ElGamalSecretKeyFromSeed(bip39.NewSeed(mnemonic, passphrase))
}
