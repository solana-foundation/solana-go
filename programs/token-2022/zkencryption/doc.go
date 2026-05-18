// Package zkencryption ports the deterministic key-derivation functions from
// solana-zk-sdk (zk-sdk/src/encryption) to Go. It produces byte-for-byte
// identical ElGamal secret keys and authenticated-encryption (AeKey) keys to
// the Rust and JS/WASM reference implementations, so the same signer and
// public seed derive the same key material across all three SDKs.
//
// Scope: key derivation only. Encryption, decryption, Pedersen commitments,
// and zero-knowledge proof generation are not in this package; callers that
// need a full confidential-transfer flow must still produce proofs via an
// external source (Rust solana-zk-sdk or JS @solana/zk-sdk WASM).
//
// Reference: https://github.com/solana-program/zk-elgamal-proof
package zkencryption
