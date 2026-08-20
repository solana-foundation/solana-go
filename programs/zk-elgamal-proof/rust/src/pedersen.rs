use solana_zk_sdk::encryption::pedersen::{Pedersen, PedersenOpening};

use crate::{
    constants::{PEDERSEN_COMMITMENT_LEN, PEDERSEN_OPENING_LEN},
    memory::stash,
    parsing::{commitment, opening, try_status, two_power},
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

/// lo + hi·2^bit_length for commitments
#[no_mangle]
pub unsafe extern "C" fn pedersen_combine_lo_hi_commitments(
    lo_ptr: u32,
    hi_ptr: u32,
    bit_length: u64,
) -> i64 {
    let lo = try_status!(commitment(lo_ptr));
    let hi = try_status!(commitment(hi_ptr));
    let two_power = try_status!(two_power(bit_length));
    stash(&(lo + (hi * two_power)).to_bytes())
}

/// lo + hi·2^bit_length for openings
#[no_mangle]
pub unsafe extern "C" fn pedersen_combine_lo_hi_openings(
    lo_ptr: u32,
    hi_ptr: u32,
    bit_length: u64,
) -> i64 {
    let lo = try_status!(opening(lo_ptr));
    let hi = try_status!(opening(hi_ptr));
    let two_power = try_status!(two_power(bit_length));
    stash(&(lo + (hi * two_power)).to_bytes())
}

#[no_mangle]
pub unsafe extern "C" fn pedersen_sub_commitments(left_ptr: u32, right_ptr: u32) -> i64 {
    let left = try_status!(commitment(left_ptr));
    let right = try_status!(commitment(right_ptr));
    stash(&(left - right).to_bytes())
}

#[no_mangle]
pub unsafe extern "C" fn pedersen_sub_openings(left_ptr: u32, right_ptr: u32) -> i64 {
    let left = try_status!(opening(left_ptr));
    let right = try_status!(opening(right_ptr));
    stash(&(&left - &right).to_bytes())
}

/// Delta commitment and opening for the transfer-fee sigma proof:
/// fee·10000 − combined·fee_rate_basis_points, mirroring
/// `compute_delta_commitment_and_opening` in spl-token's
/// confidential-transfer proof-generation crate.
///
/// Result is commitment (32) || opening (32).
#[no_mangle]
pub unsafe extern "C" fn pedersen_fee_delta(
    combined_commitment_ptr: u32,
    combined_opening_ptr: u32,
    fee_commitment_ptr: u32,
    fee_opening_ptr: u32,
    fee_rate_basis_points: u64,
) -> i64 {
    const MAX_FEE_BASIS_POINTS: u64 = 10_000;
    let combined_commitment = try_status!(commitment(combined_commitment_ptr));
    let combined_opening = try_status!(opening(combined_opening_ptr));
    let fee_commitment = try_status!(commitment(fee_commitment_ptr));
    let fee_opening = try_status!(opening(fee_opening_ptr));

    let delta_commitment =
        (fee_commitment * MAX_FEE_BASIS_POINTS) - (combined_commitment * fee_rate_basis_points);
    let delta_opening =
        (fee_opening * MAX_FEE_BASIS_POINTS) - (combined_opening * fee_rate_basis_points);

    let mut out = Vec::with_capacity(PEDERSEN_COMMITMENT_LEN + PEDERSEN_OPENING_LEN);
    out.extend_from_slice(&delta_commitment.to_bytes());
    out.extend_from_slice(&delta_opening.to_bytes());
    stash(&out)
}
