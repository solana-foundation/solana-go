// Package zkencryption ports the deterministic key-derivation functions from
// solana-zk-sdk (zk-sdk/src/encryption) to Go. It produces byte-for-byte
// identical ElGamal secret keys and authenticated-encryption (AeKey) keys to
// the Rust and JS/WASM reference implementations, so the same signer and
// public seed derive the same key material across all three SDKs.
//
// The standard derivation is DeriveConfidentialKeys(signer): one signature
// over the constant StandardDerivationMessage ("solana-conf-bal/v1"), keys
// bound to the main wallet only, so a signer derives one key pair for all of
// its confidential balances and matches what other standard clients derive
// for the same wallet. There is no seed to pass on the standard path, so two
// standard clients cannot accidentally derive different keys.
//
// The *WithSeed variants (DeriveConfidentialKeysWithSeed,
// ElGamalSecretKeyFromSignerWithSeed, AeKeyFromSignerWithSeed) scope keys
// with a caller-chosen public seed for schemes that need finer granularity
// (for example single-signer PDA wallets). Keys derived with a non-empty seed
// are non-standard and will not match other clients.
//
// Wallets should expose derivation through a dedicated API and refuse generic
// signMessage requests whose message starts with "solana-conf-bal/v1": a
// signature over the derivation message is equivalent to handing out the
// wallet's confidential-balance decryption keys.
//
// Scope: key derivation only. Encryption, decryption, Pedersen commitments,
// and zero-knowledge proof generation are not in this package; callers that
// need a full confidential-transfer flow must still produce proofs via an
// external source (Rust solana-zk-sdk or JS @solana/zk-sdk WASM).
//
// Reference: https://github.com/solana-program/zk-elgamal-proof
package zkencryption
