// Result/Error codes
pub(crate) const OK: i32 = 0;
/// An input buffer failed to deserialise.
pub(crate) const ERR_BAD_INPUT: i32 = -1;
/// Proof generation failed.
pub(crate) const ERR_PROOF_GENERATION: i32 = -2;
/// Proof verification failed.
pub(crate) const ERR_PROOF_VERIFICATION: i32 = -3;
/// Decryption failed.
pub(crate) const ERR_DECRYPTION: i32 = -4;
/// Unknown proof type tag passed to `zk_verify_proof`.
pub(crate) const ERR_UNKNOWN_PROOF_TYPE: i32 = -5;
/// Out-of-memory: a guest allocation for a result buffer failed.
pub(crate) const ERR_OOM: i32 = -6;

// Buffer lengths
pub(crate) const ELGAMAL_KEYPAIR_LEN: usize = 64;
pub(crate) const ELGAMAL_PUBKEY_LEN: usize = 32;
pub(crate) const ELGAMAL_SECRET_KEY_LEN: usize = 32;
pub(crate) const ELGAMAL_CIPHERTEXT_LEN: usize = 64;
pub(crate) const PEDERSEN_COMMITMENT_LEN: usize = 32;
pub(crate) const PEDERSEN_OPENING_LEN: usize = 32;
pub(crate) const AE_KEY_LEN: usize = 16;
pub(crate) const AE_CIPHERTEXT_LEN: usize = 36;
