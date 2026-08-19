use solana_zk_sdk::{
    encryption::pedersen::{PedersenCommitment, PedersenOpening},
    zk_elgamal_proof_program::{
        build_batched_grouped_ciphertext_2_handles_validity_proof_data,
        build_batched_grouped_ciphertext_3_handles_validity_proof_data,
        build_batched_range_proof_u128_data, build_batched_range_proof_u256_data,
        build_batched_range_proof_u64_data, build_ciphertext_ciphertext_equality_proof_data,
        build_ciphertext_commitment_equality_proof_data,
        build_grouped_ciphertext_2_handles_validity_proof_data,
        build_grouped_ciphertext_3_handles_validity_proof_data,
        build_percentage_with_cap_proof_data, build_pubkey_validity_proof_data,
        build_zero_ciphertext_proof_data,
    },
};

use crate::{
    constants::{
        ERR_BAD_INPUT, ERR_PROOF_GENERATION, PEDERSEN_COMMITMENT_LEN, PEDERSEN_OPENING_LEN,
    },
    memory::stash_res,
    parsing::{
        ciphertext, commitment, grouped_ct, input_slice, keypair, opening, pubkey, pubkeys,
        try_status,
    },
};

type RangeProofInputs = (
    Vec<PedersenCommitment>,
    Vec<u64>,
    Vec<usize>,
    Vec<PedersenOpening>,
);

#[no_mangle]
pub unsafe extern "C" fn proof_pubkey_validity(keypair_ptr: u32) -> i64 {
    let kp = try_status!(keypair(keypair_ptr));
    stash_res::<_, _, ERR_PROOF_GENERATION>(build_pubkey_validity_proof_data(&kp))
}

#[no_mangle]
pub unsafe extern "C" fn proof_zero_ciphertext(keypair_ptr: u32, ciphertext_ptr: u32) -> i64 {
    let kp = try_status!(keypair(keypair_ptr));
    let ct = try_status!(ciphertext(ciphertext_ptr));
    stash_res::<_, _, ERR_PROOF_GENERATION>(build_zero_ciphertext_proof_data(&kp, &ct))
}

#[no_mangle]
pub unsafe extern "C" fn proof_ciphertext_commitment_equality(
    keypair_ptr: u32,
    ciphertext_ptr: u32,
    commitment_ptr: u32,
    opening_ptr: u32,
    amount: u64,
) -> i64 {
    let kp = try_status!(keypair(keypair_ptr));
    let ct = try_status!(ciphertext(ciphertext_ptr));
    let comm = try_status!(commitment(commitment_ptr));
    let open = try_status!(opening(opening_ptr));
    stash_res::<_, _, ERR_PROOF_GENERATION>(build_ciphertext_commitment_equality_proof_data(
        &kp, &ct, &comm, &open, amount,
    ))
}

#[no_mangle]
pub unsafe extern "C" fn proof_ciphertext_ciphertext_equality(
    first_keypair_ptr: u32,
    second_pubkey_ptr: u32,
    first_ciphertext_ptr: u32,
    second_ciphertext_ptr: u32,
    second_opening_ptr: u32,
    amount: u64,
) -> i64 {
    let kp = try_status!(keypair(first_keypair_ptr));
    let second_pk = try_status!(pubkey(second_pubkey_ptr));
    let first_ct = try_status!(ciphertext(first_ciphertext_ptr));
    let second_ct = try_status!(ciphertext(second_ciphertext_ptr));
    let second_open = try_status!(opening(second_opening_ptr));
    stash_res::<_, _, ERR_PROOF_GENERATION>(build_ciphertext_ciphertext_equality_proof_data(
        &kp,
        &second_pk,
        &first_ct,
        &second_ct,
        &second_open,
        amount,
    ))
}

#[no_mangle]
pub unsafe extern "C" fn proof_percentage_with_cap(
    percentage_commitment_ptr: u32,
    percentage_opening_ptr: u32,
    percentage_amount: u64,
    delta_commitment_ptr: u32,
    delta_opening_ptr: u32,
    delta_amount: u64,
    claimed_commitment_ptr: u32,
    claimed_opening_ptr: u32,
    max_value: u64,
) -> i64 {
    let percentage_commitment = try_status!(commitment(percentage_commitment_ptr));
    let percentage_opening = try_status!(opening(percentage_opening_ptr));
    let delta_commitment = try_status!(commitment(delta_commitment_ptr));
    let delta_opening = try_status!(opening(delta_opening_ptr));
    let claimed_commitment = try_status!(commitment(claimed_commitment_ptr));
    let claimed_opening = try_status!(opening(claimed_opening_ptr));
    stash_res::<_, _, ERR_PROOF_GENERATION>(build_percentage_with_cap_proof_data(
        &percentage_commitment,
        &percentage_opening,
        percentage_amount,
        &delta_commitment,
        &delta_opening,
        delta_amount,
        &claimed_commitment,
        &claimed_opening,
        max_value,
    ))
}

/// Shared input layout for the batched range proofs:
/// - `commitments_ptr`: batch_len * 32 bytes
/// - `amounts_ptr`: batch_len * 8 bytes, little-endian u64s
/// - `bit_lengths_ptr`: batch_len bytes
/// - `openings_ptr`: batch_len * 32 bytes
unsafe fn range_proof_inputs(
    batch_len: u32,
    commitments_ptr: u32,
    amounts_ptr: u32,
    bit_lengths_ptr: u32,
    openings_ptr: u32,
) -> Result<RangeProofInputs, i32> {
    let n = batch_len as usize;
    let span = |stride: usize| n.checked_mul(stride).ok_or(ERR_BAD_INPUT);
    let commitments = input_slice(commitments_ptr, span(PEDERSEN_COMMITMENT_LEN)?)?
        .chunks_exact(PEDERSEN_COMMITMENT_LEN)
        .map(|chunk| PedersenCommitment::from_bytes(chunk).ok_or(ERR_BAD_INPUT))
        .collect::<Result<Vec<_>, _>>()?;
    let openings = input_slice(openings_ptr, span(PEDERSEN_OPENING_LEN)?)?
        .chunks_exact(PEDERSEN_OPENING_LEN)
        .map(|chunk| PedersenOpening::from_bytes(chunk).ok_or(ERR_BAD_INPUT))
        .collect::<Result<Vec<_>, _>>()?;
    let amounts = input_slice(amounts_ptr, span(8)?)?
        .chunks_exact(8)
        .map(|chunk| u64::from_le_bytes(chunk.try_into().unwrap()))
        .collect();
    let bit_lengths = input_slice(bit_lengths_ptr, n)?
        .iter()
        .map(|&b| b as usize)
        .collect();
    Ok((commitments, amounts, bit_lengths, openings))
}

macro_rules! batched_range_proof {
    ($name:ident, $builder:ident) => {
        #[no_mangle]
        pub unsafe extern "C" fn $name(
            batch_len: u32,
            commitments_ptr: u32,
            amounts_ptr: u32,
            bit_lengths_ptr: u32,
            openings_ptr: u32,
        ) -> i64 {
            let (commitments, amounts, bit_lengths, openings) = try_status!(range_proof_inputs(
                batch_len,
                commitments_ptr,
                amounts_ptr,
                bit_lengths_ptr,
                openings_ptr,
            ));
            stash_res::<_, _, ERR_PROOF_GENERATION>($builder(
                commitments.iter().collect(),
                amounts,
                bit_lengths,
                openings.iter().collect(),
            ))
        }
    };
}

batched_range_proof!(proof_batched_range_u64, build_batched_range_proof_u64_data);
batched_range_proof!(
    proof_batched_range_u128,
    build_batched_range_proof_u128_data
);
batched_range_proof!(
    proof_batched_range_u256,
    build_batched_range_proof_u256_data
);

#[no_mangle]
pub unsafe extern "C" fn proof_grouped_ciphertext_2_handles_validity(
    pubkeys_ptr: u32,
    grouped_ciphertext_ptr: u32,
    amount: u64,
    opening_ptr: u32,
) -> i64 {
    let pks = try_status!(pubkeys(pubkeys_ptr, 2));
    let grouped = try_status!(grouped_ct::<2>(grouped_ciphertext_ptr));
    let open = try_status!(opening(opening_ptr));
    stash_res::<_, _, ERR_PROOF_GENERATION>(build_grouped_ciphertext_2_handles_validity_proof_data(
        &pks[0], &pks[1], &grouped, amount, &open,
    ))
}

#[no_mangle]
pub unsafe extern "C" fn proof_grouped_ciphertext_3_handles_validity(
    pubkeys_ptr: u32,
    grouped_ciphertext_ptr: u32,
    amount: u64,
    opening_ptr: u32,
) -> i64 {
    let pks = try_status!(pubkeys(pubkeys_ptr, 3));
    let grouped = try_status!(grouped_ct::<3>(grouped_ciphertext_ptr));
    let open = try_status!(opening(opening_ptr));
    stash_res::<_, _, ERR_PROOF_GENERATION>(build_grouped_ciphertext_3_handles_validity_proof_data(
        &pks[0], &pks[1], &pks[2], &grouped, amount, &open,
    ))
}

#[no_mangle]
pub unsafe extern "C" fn proof_batched_grouped_ciphertext_2_handles_validity(
    pubkeys_ptr: u32,
    grouped_ciphertext_lo_ptr: u32,
    grouped_ciphertext_hi_ptr: u32,
    amount_lo: u64,
    amount_hi: u64,
    opening_lo_ptr: u32,
    opening_hi_ptr: u32,
) -> i64 {
    let pks = try_status!(pubkeys(pubkeys_ptr, 2));
    let grouped_lo = try_status!(grouped_ct::<2>(grouped_ciphertext_lo_ptr));
    let grouped_hi = try_status!(grouped_ct::<2>(grouped_ciphertext_hi_ptr));
    let opening_lo = try_status!(opening(opening_lo_ptr));
    let opening_hi = try_status!(opening(opening_hi_ptr));
    stash_res::<_, _, ERR_PROOF_GENERATION>(
        build_batched_grouped_ciphertext_2_handles_validity_proof_data(
            &pks[0],
            &pks[1],
            &grouped_lo,
            &grouped_hi,
            amount_lo,
            amount_hi,
            &opening_lo,
            &opening_hi,
        ),
    )
}

#[no_mangle]
pub unsafe extern "C" fn proof_batched_grouped_ciphertext_3_handles_validity(
    pubkeys_ptr: u32,
    grouped_ciphertext_lo_ptr: u32,
    grouped_ciphertext_hi_ptr: u32,
    amount_lo: u64,
    amount_hi: u64,
    opening_lo_ptr: u32,
    opening_hi_ptr: u32,
) -> i64 {
    let pks = try_status!(pubkeys(pubkeys_ptr, 3));
    let grouped_lo = try_status!(grouped_ct::<3>(grouped_ciphertext_lo_ptr));
    let grouped_hi = try_status!(grouped_ct::<3>(grouped_ciphertext_hi_ptr));
    let opening_lo = try_status!(opening(opening_lo_ptr));
    let opening_hi = try_status!(opening(opening_hi_ptr));
    stash_res::<_, _, ERR_PROOF_GENERATION>(
        build_batched_grouped_ciphertext_3_handles_validity_proof_data(
            &pks[0],
            &pks[1],
            &pks[2],
            &grouped_lo,
            &grouped_hi,
            amount_lo,
            amount_hi,
            &opening_lo,
            &opening_hi,
        ),
    )
}
