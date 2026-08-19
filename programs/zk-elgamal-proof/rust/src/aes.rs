use solana_zk_sdk::encryption::auth_encryption::{AeCiphertext, AeKey};

use crate::{
    constants::{AE_CIPHERTEXT_LEN, AE_KEY_LEN, ERR_BAD_INPUT, ERR_DECRYPTION},
    memory::{stash, stash_u64},
    parsing::{ae_key, input, try_status},
};

#[no_mangle]
pub extern "C" fn ae_key_new_rand() -> i64 {
    stash(&Into::<[u8; AE_KEY_LEN]>::into(AeKey::new_rand()))
}

#[no_mangle]
pub unsafe extern "C" fn ae_encrypt(key_ptr: u32, amount: u64) -> i64 {
    let key = try_status!(ae_key(key_ptr));
    stash(&key.encrypt(amount).to_bytes())
}

/// Result is the amount as 8 little-endian bytes.
#[no_mangle]
pub unsafe extern "C" fn ae_decrypt(key_ptr: u32, ciphertext_ptr: u32) -> i64 {
    let key = try_status!(ae_key(key_ptr));
    let bytes = try_status!(input::<AE_CIPHERTEXT_LEN>(ciphertext_ptr));
    let ct = match AeCiphertext::from_bytes(bytes.as_slice()) {
        Some(ct) => ct,
        None => return ERR_BAD_INPUT.into(),
    };
    match key.decrypt(&ct) {
        Some(amount) => stash_u64(amount),
        None => ERR_DECRYPTION.into(),
    }
}
