use solana_zk_sdk::encryption::elgamal::ElGamalKeypair;

use crate::{
    constants::{ELGAMAL_KEYPAIR_LEN, ERR_DECRYPTION},
    memory::{stash, stash_u64},
    parsing::{ciphertext, opening, pubkey, secret_key, try_status},
};

#[no_mangle]
pub extern "C" fn elgamal_keypair_new_rand() -> i64 {
    stash(&keypair_bytes(&ElGamalKeypair::new_rand()))
}

/// Derive the keypair (pubkey || secret, 64 bytes) from a 32-byte secret key.
#[no_mangle]
pub unsafe extern "C" fn elgamal_keypair_from_secret(secret_ptr: u32) -> i64 {
    let secret = try_status!(secret_key(secret_ptr));
    stash(&keypair_bytes(&ElGamalKeypair::new(secret)))
}

#[no_mangle]
pub unsafe extern "C" fn elgamal_encrypt(pubkey_ptr: u32, amount: u64) -> i64 {
    let pk = try_status!(pubkey(pubkey_ptr));
    stash(&pk.encrypt_u64(amount).to_bytes())
}

#[no_mangle]
pub unsafe extern "C" fn elgamal_encrypt_with(
    pubkey_ptr: u32,
    amount: u64,
    opening_ptr: u32,
) -> i64 {
    let pk = try_status!(pubkey(pubkey_ptr));
    let opening = try_status!(opening(opening_ptr));
    stash(&pk.encrypt_with_u64(amount, &opening).to_bytes())
}

/// Decrypt a ciphertext whose plaintext is known to fit in 32 bits.
/// Result is the amount as 8 little-endian bytes.
#[no_mangle]
pub unsafe extern "C" fn elgamal_decrypt_u32(secret_ptr: u32, ciphertext_ptr: u32) -> i64 {
    let secret = try_status!(secret_key(secret_ptr));
    let ct = try_status!(ciphertext(ciphertext_ptr));
    match ct.decrypt_u32(&secret) {
        Some(amount) => stash_u64(amount),
        None => ERR_DECRYPTION.into(),
    }
}

#[no_mangle]
pub unsafe extern "C" fn elgamal_add_ciphertexts(left_ptr: u32, right_ptr: u32) -> i64 {
    let left = try_status!(ciphertext(left_ptr));
    let right = try_status!(ciphertext(right_ptr));
    stash(&(left + right).to_bytes())
}

#[no_mangle]
pub unsafe extern "C" fn elgamal_sub_ciphertexts(left_ptr: u32, right_ptr: u32) -> i64 {
    let left = try_status!(ciphertext(left_ptr));
    let right = try_status!(ciphertext(right_ptr));
    stash(&(left - right).to_bytes())
}

#[no_mangle]
pub unsafe extern "C" fn elgamal_add_amount(ciphertext_ptr: u32, amount: u64) -> i64 {
    let ct = try_status!(ciphertext(ciphertext_ptr));
    stash(&ct.add_amount(amount).to_bytes())
}

#[no_mangle]
pub unsafe extern "C" fn elgamal_sub_amount(ciphertext_ptr: u32, amount: u64) -> i64 {
    let ct = try_status!(ciphertext(ciphertext_ptr));
    stash(&ct.subtract_amount(amount).to_bytes())
}

fn keypair_bytes(kp: &ElGamalKeypair) -> Vec<u8> {
    Into::<[u8; ELGAMAL_KEYPAIR_LEN]>::into(kp).to_vec()
}
