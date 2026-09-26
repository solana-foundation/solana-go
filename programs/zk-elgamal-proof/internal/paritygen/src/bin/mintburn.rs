//! Emits the spl-token-2022-interface confidential mint/burn instruction
//! encodings for the go token-2022 package to check itself against.
//!
//! Regenerate the vectors with:
//! cargo run --bin mintburn > ../../../token-2022/testdata/confidential_mint_burn_rust_parity.json
use {
    bytemuck::Zeroable,
    parity::{addr, emit, offset, pod, print_vectors},
    solana_pubkey::Pubkey,
    solana_zk_elgamal_proof_interface::proof_data::*,
    solana_zk_sdk_pod::encryption::{
        auth_encryption::PodAeCiphertext,
        elgamal::{PodElGamalCiphertext, PodElGamalPubkey},
    },
    spl_token_2022_interface::extension::confidential_mint_burn::instruction as cmb,
    spl_token_confidential_transfer_proof_extraction::instruction::ProofLocation,
};

const TOKEN_ACCOUNT: u8 = 10;
const MINT: u8 = 11;
const AUTHORITY: u8 = 13;
const SIGNER_1: u8 = 14;
const SIGNER_2: u8 = 15;
const CONTEXT_SINGLE: u8 = 20;
const CONTEXT_EQUALITY: u8 = 21;
const CONTEXT_VALIDITY: u8 = 22;
const CONTEXT_RANGE: u8 = 25;

fn main() {
    let id = spl_token_2022_interface::id();
    let token_account = addr(TOKEN_ACCOUNT);
    let mint = addr(MINT);
    let authority = addr(AUTHORITY);
    let signer_1 = addr(SIGNER_1);
    let signer_2 = addr(SIGNER_2);
    let multisig: Vec<&Pubkey> = vec![&signer_1, &signer_2];

    let supply_elgamal_pubkey: PodElGamalPubkey = pod(1);
    let decryptable_supply: PodAeCiphertext = pod(2);
    let ciphertext_lo: PodElGamalCiphertext = pod(3);
    let ciphertext_hi: PodElGamalCiphertext = pod(4);
    let new_supply_elgamal_pubkey: PodElGamalPubkey = pod(5);

    // Zeroable is enough: proof data only shows up in the appended VerifyProof
    // instructions, whose encoding parity zkprogram already pins with
    // pattern-filled pods.
    let ciphertext_equality = CiphertextCiphertextEqualityProofData::zeroed();
    let equality = CiphertextCommitmentEqualityProofData::zeroed();
    let validity_3 = BatchedGroupedCiphertext3HandlesValidityProofData::zeroed();
    let range_u128 = BatchedRangeProofU128Data::zeroed();

    let vectors: Vec<String> = vec![
        emit(
            "initialize_mint",
            vec![
                cmb::initialize_mint(&id, &mint, &supply_elgamal_pubkey, &decryptable_supply)
                    .unwrap(),
            ],
        ),
        emit(
            "update_decryptable_supply",
            vec![
                cmb::update_decryptable_supply(&id, &mint, &authority, &[], &decryptable_supply)
                    .unwrap(),
            ],
        ),
        emit(
            "update_decryptable_supply_multisig",
            vec![cmb::update_decryptable_supply(
                &id,
                &mint,
                &authority,
                &multisig,
                &decryptable_supply,
            )
            .unwrap()],
        ),
        emit(
            "rotate_supply_elgamal_pubkey_offset",
            cmb::rotate_supply_elgamal_pubkey(
                &id,
                &mint,
                &authority,
                &[],
                &new_supply_elgamal_pubkey,
                ProofLocation::InstructionOffset(offset(1), &ciphertext_equality),
            )
            .unwrap(),
        ),
        emit(
            "rotate_supply_elgamal_pubkey_context",
            cmb::rotate_supply_elgamal_pubkey(
                &id,
                &mint,
                &authority,
                &[],
                &new_supply_elgamal_pubkey,
                ProofLocation::ContextStateAccount(&addr(CONTEXT_SINGLE)),
            )
            .unwrap(),
        ),
        emit(
            "rotate_supply_elgamal_pubkey_multisig",
            cmb::rotate_supply_elgamal_pubkey(
                &id,
                &mint,
                &authority,
                &multisig,
                &new_supply_elgamal_pubkey,
                ProofLocation::InstructionOffset(offset(1), &ciphertext_equality),
            )
            .unwrap(),
        ),
        emit(
            "mint_offset",
            cmb::confidential_mint_with_split_proofs(
                &id,
                &token_account,
                &mint,
                &ciphertext_lo,
                &ciphertext_hi,
                &authority,
                &[],
                ProofLocation::InstructionOffset(offset(1), &equality),
                ProofLocation::InstructionOffset(offset(2), &validity_3),
                ProofLocation::InstructionOffset(offset(3), &range_u128),
                &decryptable_supply,
            )
            .unwrap(),
        ),
        emit(
            "mint_context",
            cmb::confidential_mint_with_split_proofs(
                &id,
                &token_account,
                &mint,
                &ciphertext_lo,
                &ciphertext_hi,
                &authority,
                &[],
                ProofLocation::ContextStateAccount(&addr(CONTEXT_EQUALITY)),
                ProofLocation::ContextStateAccount(&addr(CONTEXT_VALIDITY)),
                ProofLocation::ContextStateAccount(&addr(CONTEXT_RANGE)),
                &decryptable_supply,
            )
            .unwrap(),
        ),
        emit(
            "mint_equality_context",
            cmb::confidential_mint_with_split_proofs(
                &id,
                &token_account,
                &mint,
                &ciphertext_lo,
                &ciphertext_hi,
                &authority,
                &[],
                ProofLocation::ContextStateAccount(&addr(CONTEXT_EQUALITY)),
                ProofLocation::InstructionOffset(offset(1), &validity_3),
                ProofLocation::InstructionOffset(offset(2), &range_u128),
                &decryptable_supply,
            )
            .unwrap(),
        ),
        emit(
            "mint_range_context_multisig",
            cmb::confidential_mint_with_split_proofs(
                &id,
                &token_account,
                &mint,
                &ciphertext_lo,
                &ciphertext_hi,
                &authority,
                &multisig,
                ProofLocation::InstructionOffset(offset(1), &equality),
                ProofLocation::InstructionOffset(offset(2), &validity_3),
                ProofLocation::ContextStateAccount(&addr(CONTEXT_RANGE)),
                &decryptable_supply,
            )
            .unwrap(),
        ),
        emit(
            "burn_offset",
            cmb::confidential_burn_with_split_proofs(
                &id,
                &token_account,
                &mint,
                &decryptable_supply,
                &ciphertext_lo,
                &ciphertext_hi,
                &authority,
                &[],
                ProofLocation::InstructionOffset(offset(1), &equality),
                ProofLocation::InstructionOffset(offset(2), &validity_3),
                ProofLocation::InstructionOffset(offset(3), &range_u128),
            )
            .unwrap(),
        ),
        emit(
            "burn_context",
            cmb::confidential_burn_with_split_proofs(
                &id,
                &token_account,
                &mint,
                &decryptable_supply,
                &ciphertext_lo,
                &ciphertext_hi,
                &authority,
                &[],
                ProofLocation::ContextStateAccount(&addr(CONTEXT_EQUALITY)),
                ProofLocation::ContextStateAccount(&addr(CONTEXT_VALIDITY)),
                ProofLocation::ContextStateAccount(&addr(CONTEXT_RANGE)),
            )
            .unwrap(),
        ),
        emit(
            "burn_equality_context",
            cmb::confidential_burn_with_split_proofs(
                &id,
                &token_account,
                &mint,
                &decryptable_supply,
                &ciphertext_lo,
                &ciphertext_hi,
                &authority,
                &multisig,
                ProofLocation::ContextStateAccount(&addr(CONTEXT_EQUALITY)),
                ProofLocation::InstructionOffset(offset(1), &validity_3),
                ProofLocation::InstructionOffset(offset(2), &range_u128),
            )
            .unwrap(),
        ),
        emit(
            "apply_pending_burn",
            vec![cmb::apply_pending_burn(&id, &mint, &authority, &[]).unwrap()],
        ),
        emit(
            "apply_pending_burn_multisig",
            vec![cmb::apply_pending_burn(&id, &mint, &authority, &multisig).unwrap()],
        ),
    ];

    print_vectors(&id, vectors);
}
