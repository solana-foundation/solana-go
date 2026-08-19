use solana_zk_sdk::encryption::pedersen::{Pedersen, PedersenOpening};

use crate::{
    constants::{PEDERSEN_COMMITMENT_LEN, PEDERSEN_OPENING_LEN},
    memory::stash,
    parsing::{opening, try_status},
};

/// Result is commitment (32) || opening (32).
#[no_mangle]
pub extern "C" fn pedersen_commit(amount: u64) -> i64 {
    let (commitment, opening) = Pedersen::new(amount);
    let mut out = Vec::with_capacity(PEDERSEN_COMMITMENT_LEN + PEDERSEN_OPENING_LEN);
    out.extend_from_slice(&commitment.to_bytes());
    out.extend_from_slice(&opening.to_bytes());
    stash(&out)
}

#[no_mangle]
pub unsafe extern "C" fn pedersen_commit_with(amount: u64, opening_ptr: u32) -> i64 {
    let opening = try_status!(opening(opening_ptr));
    stash(&Pedersen::with(amount, &opening).to_bytes())
}

#[no_mangle]
pub extern "C" fn pedersen_opening_new_rand() -> i64 {
    stash(&PedersenOpening::new_rand().to_bytes())
}
