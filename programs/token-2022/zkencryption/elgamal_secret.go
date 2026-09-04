package zkencryption

import (
	"crypto/hkdf"
	"crypto/sha512"
	"fmt"

	"filippo.io/edwards25519"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/bip39"
)

// ElGamalSecretKeyLen is the canonical length of an ElGamal secret scalar
// encoded in little-endian form (matches curve25519-dalek Scalar::as_bytes).
const ElGamalSecretKeyLen = 32

// SigningDomain is the domain-separation prefix for confidential-balances key
// derivation. It must match HKDF_SALT (b"solana-conf-bal/v1") in solana-zk-sdk
// (derive_confidential_keys_from_ikm): it is both the message prefix the
// signer signs and the HKDF salt.
const SigningDomain = "solana-conf-bal/v1"

// elgamalInfo is the HKDF info string that scopes the ElGamal expansion,
// matching ELGAMAL_HKDF_INFO in solana-zk-sdk.
const elgamalInfo = "elgamal"

// elgamalMinSeedLen is the minimum seed length accepted by the ElGamal
// derivation. It mirrors MINIMUM_SEED_LEN in solana-zk-sdk's
// derive_confidential_keys_from_ikm (32), which is independent from the AE
// minimum of 16.
const elgamalMinSeedLen = ElGamalSecretKeyLen

// maxSeedLen mirrors the maximum bound enforced by solana-zk-sdk's
// derive_confidential_keys_from_ikm (MAXIMUM_IKM_LEN). It applies to both the
// AE and ElGamal derivations.
const maxSeedLen = 65535

// ElGamalSecretKey is a canonical little-endian encoding of a Ristretto/Ed25519
// scalar mod ell. It is the Token-2022 confidential-transfer ElGamal private
// key; byte-for-byte equivalent to ElGamalSecretKey::as_bytes in solana-zk-sdk.
type ElGamalSecretKey [ElGamalSecretKeyLen]byte

// ConfidentialDerivationMessage returns the canonical confidential-balances
// derivation message, b"solana-conf-bal/v1" || publicSeed. Mirrors
// confidential_derivation_message in solana-zk-sdk; this is the exact message
// a Signer must sign to derive confidential-balances keys.
func ConfidentialDerivationMessage(publicSeed []byte) []byte {
	msg := make([]byte, 0, len(SigningDomain)+len(publicSeed))
	msg = append(msg, SigningDomain...)
	msg = append(msg, publicSeed...)
	return msg
}

// deriveElGamalSecretKey implements the solana-conf-bal/v1 derivation:
// HKDF-SHA512(salt=SigningDomain, ikm).expand(info="elgamal", 64) reduced via
// Scalar::from_bytes_mod_order_wide, matching derive_confidential_keys_from_ikm.
func deriveElGamalSecretKey(ikm []byte) (ElGamalSecretKey, error) {
	if len(ikm) < elgamalMinSeedLen {
		return ElGamalSecretKey{}, ErrSeedTooShort
	}
	if len(ikm) > maxSeedLen {
		return ElGamalSecretKey{}, ErrSeedTooLong
	}

	wide, err := hkdf.Key(sha512.New, ikm, []byte(SigningDomain), elgamalInfo, 64)
	if err != nil {
		return ElGamalSecretKey{}, fmt.Errorf("zkencryption: HKDF expand elgamal: %w", err)
	}
	// SetUniformBytes performs Scalar::from_bytes_mod_order_wide on 64 bytes.
	s, err := edwards25519.NewScalar().SetUniformBytes(wide)
	// wide holds expanded secret material; scrub it before it leaves scope.
	clear(wide)
	if err != nil {
		return ElGamalSecretKey{}, ErrInvalidScalarEncoding
	}

	var out ElGamalSecretKey
	copy(out[:], s.Bytes())
	return out, nil
}

// ElGamalSecretKeyFromSeed derives an ElGamal secret key from raw input key
// material, matching derive_confidential_keys_from_ikm in solana-zk-sdk.
func ElGamalSecretKeyFromSeed(seed []byte) (ElGamalSecretKey, error) {
	return deriveElGamalSecretKey(seed)
}

// ElGamalSecretKeyFromSignature derives an ElGamal secret key from an ed25519
// signature over ConfidentialDerivationMessage. Mirrors
// derive_confidential_keys_from_signature in solana-zk-sdk. An all-zero
// (default) signature is rejected, matching the Rust implementation.
func ElGamalSecretKeyFromSignature(sig solana.Signature) (ElGamalSecretKey, error) {
	if sig == (solana.Signature{}) {
		return ElGamalSecretKey{}, ErrDefaultSignature
	}
	return deriveElGamalSecretKey(sig[:])
}

// ElGamalSecretKeyFromSigner deterministically derives an ElGamal secret key
// from a Solana signer and a public seed. The signer signs
// b"solana-conf-bal/v1" || publicSeed (see ConfidentialDerivationMessage); the
// signature is fed through the HKDF-SHA512 solana-conf-bal/v1 derivation. The
// all-zero signature rejection lives in ElGamalSecretKeyFromSignature. The
// standard publicSeed is empty (wallet-only keys that match other standard
// clients); non-empty seeds are supported but non-standard.
func ElGamalSecretKeyFromSigner(signer Signer, publicSeed []byte) (ElGamalSecretKey, error) {
	sig, err := signer.Sign(ConfidentialDerivationMessage(publicSeed))
	if err != nil {
		return ElGamalSecretKey{}, fmt.Errorf("zkencryption: sign confidential-balances public seed: %w", err)
	}
	return ElGamalSecretKeyFromSignature(sig)
}

// ElGamalSecretKeyFromSeedPhraseAndPassphrase derives an ElGamal secret key
// from a BIP39 mnemonic and an optional passphrase using the standard BIP39
// PBKDF2-HMAC-SHA512 seed derivation (2048 iterations, 64-byte output). The
// seed is used directly as the HKDF input key material.
func ElGamalSecretKeyFromSeedPhraseAndPassphrase(mnemonic, passphrase string) (ElGamalSecretKey, error) {
	return ElGamalSecretKeyFromSeed(bip39.NewSeed(mnemonic, passphrase))
}
