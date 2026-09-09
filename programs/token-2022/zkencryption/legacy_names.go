package zkencryption

// ElGamalSecretKeyFromSigner derives an ElGamal secret key from a signer and a
// caller-chosen public seed.
//
// Deprecated: use DeriveConfidentialKeys for the standard wallet-level keys
// (no seed, one signature for both keys), or ElGamalSecretKeyFromSignerWithSeed
// for seed-scoped keys.
func ElGamalSecretKeyFromSigner(signer Signer, publicSeed []byte) (ElGamalSecretKey, error) {
	return ElGamalSecretKeyFromSignerWithSeed(signer, publicSeed)
}

// AeKeyFromSigner derives an AeKey from a signer and a caller-chosen public
// seed.
//
// Deprecated: use DeriveConfidentialKeys for the standard wallet-level keys
// (no seed, one signature for both keys), or AeKeyFromSignerWithSeed for
// seed-scoped keys.
func AeKeyFromSigner(signer Signer, publicSeed []byte) (AeKey, error) {
	return AeKeyFromSignerWithSeed(signer, publicSeed)
}
