//! Emits the spl-token-2022-interface confidential transfer instruction
//! encodings for the go token-2022 package to check itself against.
//!
//! Regenerate the vectors with:
//! cargo run --bin token2022 > ../../../token-2022/testdata/confidential_transfer_rust_parity.json
use {
    bytemuck::Zeroable,
    solana_instruction::Instruction,
    solana_pubkey::Pubkey,
    solana_zk_sdk::encryption::pod::{
        auth_encryption::PodAeCiphertext,
        elgamal::{PodElGamalCiphertext, PodElGamalPubkey},
    },
    solana_zk_sdk::zk_elgamal_proof_program::proof_data::*,
    spl_token_2022_interface::extension::confidential_transfer::instruction as ct,
    spl_token_confidential_transfer_proof_extraction::instruction::ProofLocation,
    std::num::NonZeroI8,
};

const TOKEN_ACCOUNT: u8 = 10;
const MINT: u8 = 11;
const DESTINATION: u8 = 12;
const AUTHORITY: u8 = 13;
const SIGNER_1: u8 = 14;
const SIGNER_2: u8 = 15;
const CONTEXT_SINGLE: u8 = 20;
const CONTEXT_EQUALITY: u8 = 21;
const CONTEXT_VALIDITY: u8 = 22;
const CONTEXT_FEE_SIGMA: u8 = 23;
const CONTEXT_FEE_VALIDITY: u8 = 24;
const CONTEXT_RANGE: u8 = 25;
const REGISTRY: u8 = 30;
const PAYER: u8 = 31;

const AMOUNT: u64 = 0x1122334455667788;
const DECIMALS: u8 = 9;
const MAX_PENDING_COUNTER: u64 = 65536;

fn hex(b: &[u8]) -> String {
    b.iter().map(|x| format!("{:02x}", x)).collect()
}

// pattern fills pods with distinct, non-zero bytes; seed keeps two values of
// the same type distinguishable.
fn pattern(seed: u8, n: usize) -> Vec<u8> {
    (0..n)
        .map(|i| ((i + seed as usize) % 251 + 1) as u8)
        .collect()
}

fn pod<T: bytemuck::Pod>(seed: u8) -> T {
    *bytemuck::from_bytes(&pattern(seed, core::mem::size_of::<T>()))
}

fn addr(n: u8) -> Pubkey {
    Pubkey::from([n; 32])
}

fn ix_json(ix: &Instruction) -> String {
    let accounts: Vec<String> = ix
        .accounts
        .iter()
        .map(|a| {
            format!(
                r#"{{"pubkey":"{}","is_signer":{},"is_writable":{}}}"#,
                hex(a.pubkey.as_ref()),
                a.is_signer,
                a.is_writable
            )
        })
        .collect();
    format!(
        r#"{{"program_id":"{}","data":"{}","accounts":[{}]}}"#,
        hex(ix.program_id.as_ref()),
        hex(&ix.data),
        accounts.join(",")
    )
}

fn emit(name: &str, instructions: Vec<Instruction>) -> String {
    let encoded: Vec<String> = instructions.iter().map(ix_json).collect();
    format!(
        r#"{{"name":"{}","instructions":[{}]}}"#,
        name,
        encoded.join(",")
    )
}

fn offset(n: i8) -> NonZeroI8 {
    NonZeroI8::new(n).unwrap()
}

fn main() {
    let id = spl_token_2022_interface::id();
    let token_account = addr(TOKEN_ACCOUNT);
    let mint = addr(MINT);
    let destination = addr(DESTINATION);
    let authority = addr(AUTHORITY);
    let signer_1 = addr(SIGNER_1);
    let signer_2 = addr(SIGNER_2);
    let multisig: Vec<&Pubkey> = vec![&signer_1, &signer_2];

    let auditor_pubkey: PodElGamalPubkey = pod(1);
    let decryptable_balance: PodAeCiphertext = pod(2);
    let ciphertext_lo: PodElGamalCiphertext = pod(3);
    let ciphertext_hi: PodElGamalCiphertext = pod(4);

    // Zeroable is enough: proof data only shows up in the appended VerifyProof
    // instructions, whose encoding parity zkprogram already pins with
    // pattern-filled pods.
    let pubkey_validity = PubkeyValidityProofData::zeroed();
    let zero_ciphertext = ZeroCiphertextProofData::zeroed();
    let equality = CiphertextCommitmentEqualityProofData::zeroed();
    let range_u64 = BatchedRangeProofU64Data::zeroed();
    let range_u128 = BatchedRangeProofU128Data::zeroed();
    let range_u256 = BatchedRangeProofU256Data::zeroed();
    let validity_3 = BatchedGroupedCiphertext3HandlesValidityProofData::zeroed();
    let validity_2 = BatchedGroupedCiphertext2HandlesValidityProofData::zeroed();
    let fee_sigma = PercentageWithCapProofData::zeroed();

    let vectors: Vec<String> = vec![
        emit(
            "initialize_mint",
            vec![ct::initialize_mint(&id, &mint, Some(authority), true, Some(auditor_pubkey)).unwrap()],
        ),
        emit(
            "initialize_mint_no_optionals",
            vec![ct::initialize_mint(&id, &mint, None, false, None).unwrap()],
        ),
        emit(
            "update_mint",
            vec![ct::update_mint(&id, &mint, &authority, &[], true, Some(auditor_pubkey)).unwrap()],
        ),
        emit(
            "update_mint_multisig",
            vec![ct::update_mint(&id, &mint, &authority, &multisig, false, None).unwrap()],
        ),
        emit(
            "configure_account_offset",
            ct::configure_account(
                &id,
                &token_account,
                &mint,
                &decryptable_balance,
                MAX_PENDING_COUNTER,
                &authority,
                &[],
                ProofLocation::InstructionOffset(offset(1), &pubkey_validity),
            )
            .unwrap(),
        ),
        emit(
            "configure_account_context",
            ct::configure_account(
                &id,
                &token_account,
                &mint,
                &decryptable_balance,
                MAX_PENDING_COUNTER,
                &authority,
                &[],
                ProofLocation::ContextStateAccount(&addr(CONTEXT_SINGLE)),
            )
            .unwrap(),
        ),
        emit(
            "inner_configure_account_offset_minus_1",
            vec![ct::inner_configure_account(
                &id,
                &token_account,
                &mint,
                &decryptable_balance,
                MAX_PENDING_COUNTER,
                &authority,
                &[],
                ProofLocation::InstructionOffset(offset(-1), &pubkey_validity),
            )
            .unwrap()],
        ),
        emit(
            "approve_account",
            vec![ct::approve_account(&id, &token_account, &mint, &authority, &[]).unwrap()],
        ),
        emit(
            "empty_account_offset",
            ct::empty_account(
                &id,
                &token_account,
                &authority,
                &[],
                ProofLocation::InstructionOffset(offset(1), &zero_ciphertext),
            )
            .unwrap(),
        ),
        emit(
            "empty_account_context",
            ct::empty_account(
                &id,
                &token_account,
                &authority,
                &[],
                ProofLocation::ContextStateAccount(&addr(CONTEXT_SINGLE)),
            )
            .unwrap(),
        ),
        emit(
            "deposit",
            vec![ct::deposit(&id, &token_account, &mint, AMOUNT, DECIMALS, &authority, &[]).unwrap()],
        ),
        emit(
            "deposit_multisig",
            vec![ct::deposit(&id, &token_account, &mint, AMOUNT, DECIMALS, &authority, &multisig)
                .unwrap()],
        ),
        emit(
            "withdraw_offset",
            ct::withdraw(
                &id,
                &token_account,
                &mint,
                AMOUNT,
                DECIMALS,
                &decryptable_balance,
                &authority,
                &[],
                ProofLocation::InstructionOffset(offset(1), &equality),
                ProofLocation::InstructionOffset(offset(2), &range_u64),
            )
            .unwrap(),
        ),
        emit(
            "withdraw_context",
            ct::withdraw(
                &id,
                &token_account,
                &mint,
                AMOUNT,
                DECIMALS,
                &decryptable_balance,
                &authority,
                &[],
                ProofLocation::ContextStateAccount(&addr(CONTEXT_EQUALITY)),
                ProofLocation::ContextStateAccount(&addr(CONTEXT_RANGE)),
            )
            .unwrap(),
        ),
        emit(
            "withdraw_mixed",
            ct::withdraw(
                &id,
                &token_account,
                &mint,
                AMOUNT,
                DECIMALS,
                &decryptable_balance,
                &authority,
                &[],
                ProofLocation::InstructionOffset(offset(1), &equality),
                ProofLocation::ContextStateAccount(&addr(CONTEXT_RANGE)),
            )
            .unwrap(),
        ),
        emit(
            "transfer_offset",
            ct::transfer(
                &id,
                &token_account,
                &mint,
                &destination,
                &decryptable_balance,
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
            "transfer_context",
            ct::transfer(
                &id,
                &token_account,
                &mint,
                &destination,
                &decryptable_balance,
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
            "apply_pending_balance",
            vec![ct::apply_pending_balance(
                &id,
                &token_account,
                MAX_PENDING_COUNTER,
                &decryptable_balance,
                &authority,
                &[],
            )
            .unwrap()],
        ),
        emit(
            "enable_confidential_credits",
            vec![ct::enable_confidential_credits(&id, &token_account, &authority, &[]).unwrap()],
        ),
        emit(
            "disable_confidential_credits",
            vec![ct::disable_confidential_credits(&id, &token_account, &authority, &[]).unwrap()],
        ),
        emit(
            "enable_non_confidential_credits",
            vec![ct::enable_non_confidential_credits(&id, &token_account, &authority, &[]).unwrap()],
        ),
        emit(
            "disable_non_confidential_credits",
            vec![ct::disable_non_confidential_credits(&id, &token_account, &authority, &[]).unwrap()],
        ),
        emit(
            "transfer_with_fee_offset",
            ct::transfer_with_fee(
                &id,
                &token_account,
                &mint,
                &destination,
                &decryptable_balance,
                &ciphertext_lo,
                &ciphertext_hi,
                &authority,
                &[],
                ProofLocation::InstructionOffset(offset(1), &equality),
                ProofLocation::InstructionOffset(offset(2), &validity_3),
                ProofLocation::InstructionOffset(offset(3), &fee_sigma),
                ProofLocation::InstructionOffset(offset(4), &validity_2),
                ProofLocation::InstructionOffset(offset(5), &range_u256),
            )
            .unwrap(),
        ),
        emit(
            "transfer_with_fee_context",
            ct::transfer_with_fee(
                &id,
                &token_account,
                &mint,
                &destination,
                &decryptable_balance,
                &ciphertext_lo,
                &ciphertext_hi,
                &authority,
                &[],
                ProofLocation::ContextStateAccount(&addr(CONTEXT_EQUALITY)),
                ProofLocation::ContextStateAccount(&addr(CONTEXT_VALIDITY)),
                ProofLocation::ContextStateAccount(&addr(CONTEXT_FEE_SIGMA)),
                ProofLocation::ContextStateAccount(&addr(CONTEXT_FEE_VALIDITY)),
                ProofLocation::ContextStateAccount(&addr(CONTEXT_RANGE)),
            )
            .unwrap(),
        ),
        emit(
            "configure_account_with_registry",
            vec![ct::configure_account_with_registry(
                &id,
                &token_account,
                &mint,
                &addr(REGISTRY),
                Some(&addr(PAYER)),
            )
            .unwrap()],
        ),
        emit(
            "configure_account_with_registry_no_payer",
            vec![ct::configure_account_with_registry(&id, &token_account, &mint, &addr(REGISTRY), None)
                .unwrap()],
        ),
    ];

    println!(
        r#"{{"program_id":"{}","builders":[{}]}}"#,
        hex(id.as_ref()),
        vectors.join(",")
    );
}
