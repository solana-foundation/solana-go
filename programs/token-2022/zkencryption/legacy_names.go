package zkencryption

// ElGamalSecretKeyFromSigner derives an ElGamal secret key from a signer and a
// caller-chosen public seed.
//
// Deprecated: for the standard wallet-level keys use DeriveConfidentialKeys,
// which takes no seed and signs once for both keys; for non-standard seed
// scoping use ElGamalSecretKeyFromSignerWithSeed. This wrapper exists only for
// compatibility with v1.14.0 callers.
func ElGamalSecretKeyFromSigner(signer Signer, publicSeed []byte) (ElGamalSecretKey, error) {
	return ElGamalSecretKeyFromSignerWithSeed(signer, publicSeed)
}

// AeKeyFromSigner derives an AeKey from a signer and a caller-chosen public
// seed.
//
// Deprecated: for the standard wallet-level keys use DeriveConfidentialKeys,
// which takes no seed and signs once for both keys; for non-standard seed
// scoping use AeKeyFromSignerWithSeed. This wrapper exists only for
// compatibility with v1.14.0 callers.
func AeKeyFromSigner(signer Signer, publicSeed []byte) (AeKey, error) {
	return AeKeyFromSignerWithSeed(signer, publicSeed)
}
