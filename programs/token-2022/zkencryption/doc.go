// Package zkencryption ports the deterministic key-derivation functions from
// solana-zk-sdk (zk-sdk/src/encryption) to Go. It produces byte-for-byte
// identical ElGamal secret keys and authenticated-encryption (AeKey) keys to
// the Rust and JS/WASM reference implementations, so the same signer and
// public seed derive the same key material across all three SDKs.
//
// The standard public seed is empty (nil or zero-length): keys are bound to
// the main wallet only, so a signer derives one key pair for all of its
// confidential balances, matching what other standard clients derive for the
// same wallet. Non-empty seeds (for example a mint or token-account address)
// are supported for finer-grained scoping, but keys derived with them are
// non-standard and will not match other clients.
//
// Scope: key derivation only. Encryption, decryption, Pedersen commitments,
// and zero-knowledge proof generation are not in this package; callers that
// need a full confidential-transfer flow must still produce proofs via an
// external source (Rust solana-zk-sdk or JS @solana/zk-sdk WASM).
//
// Reference: https://github.com/solana-program/zk-elgamal-proof
package zkencryption
