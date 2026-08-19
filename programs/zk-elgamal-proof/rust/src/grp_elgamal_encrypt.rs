use solana_zk_sdk::encryption::{elgamal::ElGamalPubkey, grouped_elgamal::GroupedElGamal};

use crate::{
    constants::ERR_BAD_INPUT,
    memory::stash,
    parsing::{grouped_ct, opening, pubkeys, try_status},
};

/// `pubkeys_ptr` holds 2 concatenated ElGamal pubkeys (64 bytes).
/// Result is a 96-byte grouped ciphertext.
#[no_mangle]
pub unsafe extern "C" fn grouped_elgamal_2_encrypt_with(
    pubkeys_ptr: u32,
    amount: u64,
    opening_ptr: u32,
) -> i64 {
    grouped_elgamal_encrypt_with::<2>(pubkeys_ptr, amount, opening_ptr)
}

/// `pubkeys_ptr` holds 3 concatenated ElGamal pubkeys (96 bytes).
/// Result is a 128-byte grouped ciphertext.
#[no_mangle]
pub unsafe extern "C" fn grouped_elgamal_3_encrypt_with(
    pubkeys_ptr: u32,
    amount: u64,
    opening_ptr: u32,
) -> i64 {
    grouped_elgamal_encrypt_with::<3>(pubkeys_ptr, amount, opening_ptr)
}

/// `pubkeys_ptr` holds N concatenated ElGamal pubkeys (32*N bytes).
/// Result is a 32+32*N-byte grouped ciphertext.
unsafe fn grouped_elgamal_encrypt_with<const N: usize>(
    pubkeys_ptr: u32,
    amount: u64,
    opening_ptr: u32,
) -> i64 {
    let pks = try_status!(pubkeys(pubkeys_ptr, N));
    let opening = try_status!(opening(opening_ptr));
    let pk_refs: [&ElGamalPubkey; N] = std::array::from_fn(|i| &pks[i]);
    let grouped = GroupedElGamal::encrypt_with(pk_refs, amount, &opening);
    stash(&grouped.to_bytes())
}

#[no_mangle]
pub unsafe extern "C" fn grouped_ciphertext_2_to_elgamal(grouped_ptr: u32, index: u32) -> i64 {
    grouped_ciphertext_to_elgamal::<2>(grouped_ptr, index)
}

#[no_mangle]
pub unsafe extern "C" fn grouped_ciphertext_3_to_elgamal(grouped_ptr: u32, index: u32) -> i64 {
    grouped_ciphertext_to_elgamal::<3>(grouped_ptr, index)
}

unsafe fn grouped_ciphertext_to_elgamal<const N: usize>(grouped_ptr: u32, index: u32) -> i64 {
    let grouped = try_status!(grouped_ct::<N>(grouped_ptr));
    match grouped.to_elgamal_ciphertext(index as usize) {
        Ok(ct) => stash(&ct.to_bytes()),
        Err(_) => ERR_BAD_INPUT.into(),
    }
}
