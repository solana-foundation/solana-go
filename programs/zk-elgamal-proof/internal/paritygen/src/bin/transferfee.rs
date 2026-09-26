//! Emits the spl-token-2022-interface confidential transfer fee instruction
//! encodings for the go token-2022 package to check itself against.
//!
//! Regenerate the vectors with:
//! cargo run --bin transferfee > ../../../token-2022/testdata/confidential_transfer_fee_rust_parity.json
use {
    bytemuck::Zeroable,
    parity::{addr, emit, offset, pod, print_vectors},
    solana_pubkey::Pubkey,
    solana_zk_elgamal_proof_interface::proof_data::*,
    solana_zk_sdk_pod::encryption::{auth_encryption::PodAeCiphertext, elgamal::PodElGamalPubkey},
    spl_token_2022_interface::extension::confidential_transfer_fee::instruction as ctf,
    spl_token_confidential_transfer_proof_extraction::instruction::ProofLocation,
};

const MINT: u8 = 11;
const DESTINATION: u8 = 12;
const AUTHORITY: u8 = 13;
const SIGNER_1: u8 = 14;
const SIGNER_2: u8 = 15;
const CONTEXT_SINGLE: u8 = 20;
const SOURCE_1: u8 = 40;
const SOURCE_2: u8 = 41;

fn main() {
    let id = spl_token_2022_interface::id();
    let mint = addr(MINT);
    let destination = addr(DESTINATION);
    let authority = addr(AUTHORITY);
    let signer_1 = addr(SIGNER_1);
    let signer_2 = addr(SIGNER_2);
    let multisig: Vec<&Pubkey> = vec![&signer_1, &signer_2];
    let source_1 = addr(SOURCE_1);
    let source_2 = addr(SOURCE_2);
    let sources: Vec<&Pubkey> = vec![&source_1, &source_2];

    let withheld_authority_elgamal_pubkey: PodElGamalPubkey = pod(1);
    let decryptable_balance: PodAeCiphertext = pod(2);

    // Zeroable is enough: proof data only shows up in the appended VerifyProof
    // instructions, whose encoding parity zkprogram already pins with
    // pattern-filled pods.
    let ciphertext_equality = CiphertextCiphertextEqualityProofData::zeroed();

    let vectors: Vec<String> = vec![
        emit(
            "initialize_config",
            vec![ctf::initialize_confidential_transfer_fee_config(
                &id,
                &mint,
                Some(authority),
                &withheld_authority_elgamal_pubkey,
            )
            .unwrap()],
        ),
        emit(
            "initialize_config_no_authority",
            vec![ctf::initialize_confidential_transfer_fee_config(
                &id,
                &mint,
                None,
                &withheld_authority_elgamal_pubkey,
            )
            .unwrap()],
        ),
        emit(
            "withdraw_withheld_tokens_from_mint_offset",
            ctf::withdraw_withheld_tokens_from_mint(
                &id,
                &mint,
                &destination,
                &decryptable_balance,
                &authority,
                &[],
                ProofLocation::InstructionOffset(offset(1), &ciphertext_equality),
            )
            .unwrap(),
        ),
        emit(
            "withdraw_withheld_tokens_from_mint_context",
            ctf::withdraw_withheld_tokens_from_mint(
                &id,
                &mint,
                &destination,
                &decryptable_balance,
                &authority,
                &[],
                ProofLocation::ContextStateAccount(&addr(CONTEXT_SINGLE)),
            )
            .unwrap(),
        ),
        emit(
            "withdraw_withheld_tokens_from_mint_multisig",
            ctf::withdraw_withheld_tokens_from_mint(
                &id,
                &mint,
                &destination,
                &decryptable_balance,
                &authority,
                &multisig,
                ProofLocation::InstructionOffset(offset(1), &ciphertext_equality),
            )
            .unwrap(),
        ),
        emit(
            "inner_withdraw_withheld_tokens_from_mint_offset_minus_1",
            vec![ctf::inner_withdraw_withheld_tokens_from_mint(
                &id,
                &mint,
                &destination,
                &decryptable_balance,
                &authority,
                &[],
                ProofLocation::InstructionOffset(offset(-1), &ciphertext_equality),
            )
            .unwrap()],
        ),
        emit(
            "withdraw_withheld_tokens_from_accounts_offset",
            ctf::withdraw_withheld_tokens_from_accounts(
                &id,
                &mint,
                &destination,
                &decryptable_balance,
                &authority,
                &[],
                &sources,
                ProofLocation::InstructionOffset(offset(1), &ciphertext_equality),
            )
            .unwrap(),
        ),
        emit(
            "withdraw_withheld_tokens_from_accounts_context",
            ctf::withdraw_withheld_tokens_from_accounts(
                &id,
                &mint,
                &destination,
                &decryptable_balance,
                &authority,
                &[],
                &sources,
                ProofLocation::ContextStateAccount(&addr(CONTEXT_SINGLE)),
            )
            .unwrap(),
        ),
        // The source accounts trail the multisig signers.
        emit(
            "withdraw_withheld_tokens_from_accounts_multisig",
            ctf::withdraw_withheld_tokens_from_accounts(
                &id,
                &mint,
                &destination,
                &decryptable_balance,
                &authority,
                &multisig,
                &sources,
                ProofLocation::InstructionOffset(offset(1), &ciphertext_equality),
            )
            .unwrap(),
        ),
        emit(
            "withdraw_withheld_tokens_from_accounts_no_sources",
            ctf::withdraw_withheld_tokens_from_accounts(
                &id,
                &mint,
                &destination,
                &decryptable_balance,
                &authority,
                &[],
                &[],
                ProofLocation::ContextStateAccount(&addr(CONTEXT_SINGLE)),
            )
            .unwrap(),
        ),
        emit(
            "inner_withdraw_withheld_tokens_from_accounts_offset_minus_1",
            vec![ctf::inner_withdraw_withheld_tokens_from_accounts(
                &id,
                &mint,
                &destination,
                &decryptable_balance,
                &authority,
                &[],
                &sources,
                ProofLocation::InstructionOffset(offset(-1), &ciphertext_equality),
            )
            .unwrap()],
        ),
        emit(
            "harvest_withheld_tokens_to_mint",
            vec![ctf::harvest_withheld_tokens_to_mint(&id, &mint, &sources).unwrap()],
        ),
        emit(
            "harvest_withheld_tokens_to_mint_no_sources",
            vec![ctf::harvest_withheld_tokens_to_mint(&id, &mint, &[]).unwrap()],
        ),
        emit(
            "enable_harvest_to_mint",
            vec![ctf::enable_harvest_to_mint(&id, &mint, &authority, &[]).unwrap()],
        ),
        emit(
            "enable_harvest_to_mint_multisig",
            vec![ctf::enable_harvest_to_mint(&id, &mint, &authority, &multisig).unwrap()],
        ),
        emit(
            "disable_harvest_to_mint",
            vec![ctf::disable_harvest_to_mint(&id, &mint, &authority, &[]).unwrap()],
        ),
        emit(
            "disable_harvest_to_mint_multisig",
            vec![ctf::disable_harvest_to_mint(&id, &mint, &authority, &multisig).unwrap()],
        ),
    ];

    print_vectors(&id, vectors);
}
