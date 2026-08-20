use solana_zk_sdk::encryption::{
    auth_encryption::AeKey,
    elgamal::{ElGamalCiphertext, ElGamalKeypair, ElGamalPubkey, ElGamalSecretKey},
    grouped_elgamal::GroupedElGamalCiphertext,
    pedersen::{PedersenCommitment, PedersenOpening},
};

use crate::constants::{
    AE_KEY_LEN, ELGAMAL_CIPHERTEXT_LEN, ELGAMAL_KEYPAIR_LEN, ELGAMAL_PUBKEY_LEN,
    ELGAMAL_SECRET_KEY_LEN, ERR_BAD_INPUT, PEDERSEN_COMMITMENT_LEN, PEDERSEN_OPENING_LEN,
};

pub(crate) unsafe fn keypair(ptr: u32) -> Result<ElGamalKeypair, i32> {
    ElGamalKeypair::try_from(input::<ELGAMAL_KEYPAIR_LEN>(ptr)?.as_slice())
        .map_err(|_| ERR_BAD_INPUT)
}

pub(crate) unsafe fn pubkey(ptr: u32) -> Result<ElGamalPubkey, i32> {
    ElGamalPubkey::try_from(input::<ELGAMAL_PUBKEY_LEN>(ptr)?.as_slice()).map_err(|_| ERR_BAD_INPUT)
}

pub(crate) unsafe fn pubkeys(ptr: u32, count: usize) -> Result<Vec<ElGamalPubkey>, i32> {
    input_slice(ptr, count * ELGAMAL_PUBKEY_LEN)?
        .chunks_exact(ELGAMAL_PUBKEY_LEN)
        .map(|chunk| ElGamalPubkey::try_from(chunk).map_err(|_| ERR_BAD_INPUT))
        .collect()
}

pub(crate) unsafe fn secret_key(ptr: u32) -> Result<ElGamalSecretKey, i32> {
    ElGamalSecretKey::try_from(input::<ELGAMAL_SECRET_KEY_LEN>(ptr)?.as_slice())
        .map_err(|_| ERR_BAD_INPUT)
}

pub(crate) unsafe fn ciphertext(ptr: u32) -> Result<ElGamalCiphertext, i32> {
    ElGamalCiphertext::from_bytes(input::<ELGAMAL_CIPHERTEXT_LEN>(ptr)?.as_slice())
        .ok_or(ERR_BAD_INPUT)
}

pub(crate) unsafe fn commitment(ptr: u32) -> Result<PedersenCommitment, i32> {
    PedersenCommitment::from_bytes(input::<PEDERSEN_COMMITMENT_LEN>(ptr)?.as_slice())
        .ok_or(ERR_BAD_INPUT)
}

pub(crate) unsafe fn opening(ptr: u32) -> Result<PedersenOpening, i32> {
    PedersenOpening::from_bytes(input::<PEDERSEN_OPENING_LEN>(ptr)?.as_slice()).ok_or(ERR_BAD_INPUT)
}

pub(crate) unsafe fn grouped_ct<const N: usize>(
    ptr: u32,
) -> Result<GroupedElGamalCiphertext<N>, i32> {
    GroupedElGamalCiphertext::from_bytes(input_slice(ptr, 32 + 32 * N)?).ok_or(ERR_BAD_INPUT)
}

pub(crate) unsafe fn ae_key(ptr: u32) -> Result<AeKey, i32> {
    AeKey::try_from(input::<AE_KEY_LEN>(ptr)?.as_slice()).map_err(|_| ERR_BAD_INPUT)
}

/// Validate a lo/hi recombination shift and return 2^bit_length.
pub(crate) fn two_power(bit_length: u64) -> Result<u64, i32> {
    if bit_length >= u64::BITS as u64 {
        return Err(ERR_BAD_INPUT);
    }
    Ok(1u64 << bit_length)
}

/// Copy a fixed-size input out of a host buffer into an owned array.
pub(crate) unsafe fn input<const N: usize>(ptr: u32) -> Result<[u8; N], i32> {
    let mut buf = [0u8; N];
    buf.copy_from_slice(input_slice(ptr, N)?);
    Ok(buf)
}

/// Borrow the first `len` bytes of a host buffer.
pub(crate) unsafe fn input_slice<'a>(ptr: u32, len: usize) -> Result<&'a [u8], i32> {
    match crate::memory::buffer_len(ptr) {
        Some(buf_len) if buf_len as usize >= len => {}
        _ => return Err(ERR_BAD_INPUT),
    }
    Ok(std::slice::from_raw_parts(ptr as *const u8, len))
}

macro_rules! try_status {
    ($expr:expr) => {
        match $expr {
            Ok(value) => value,
            Err(code) => return code.into(),
        }
    };
}
pub(crate) use try_status;
