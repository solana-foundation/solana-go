//! Emits the encodings for the go package to check itself against.
//!
//! Regenerate the vectors with: cargo run --bin zkprogram > ../../zkprogram/testdata/rust_parity.json
use {
    bytemuck::Pod,
    core::mem::size_of,
    num_traits::ToPrimitive,
    solana_address::Address,
    solana_instruction::Instruction,
    solana_zk_elgamal_proof_interface::{
        instruction::{close_context_state, ContextStateInfo, ProofInstruction},
        proof_data::*,
        state::ProofContextState,
    },
};

const CONTEXT_STATE_ACCOUNT: u8 = 1;
const CONTEXT_STATE_AUTHORITY: u8 = 2;
const PROOF_ACCOUNT: u8 = 3;
const DESTINATION: u8 = 4;
const PROOF_DATA_OFFSET: u32 = 64;
const PROOF_DATA_OFFSET_WITH_CONTEXT: u32 = 0xDEADBEEF;

fn hex(b: &[u8]) -> String {
    b.iter().map(|x| format!("{:02x}", x)).collect()
}

fn pattern(n: usize) -> Vec<u8> {
    (0..n).map(|i| (i % 251 + 1) as u8).collect()
}

fn addr(n: u8) -> Address {
    Address::from([n; 32])
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

fn emit<T, U>(name: &str, inst: ProofInstruction) -> String
where
    T: Pod + ZkProofData<U>,
    U: Pod,
{
    let raw = pattern(size_of::<T>());
    let data: &T = bytemuck::from_bytes(&raw);

    let context_state_account = addr(CONTEXT_STATE_ACCOUNT);
    let context_state_authority = addr(CONTEXT_STATE_AUTHORITY);
    let proof_account = addr(PROOF_ACCOUNT);
    let info = ContextStateInfo {
        context_state_account: &context_state_account,
        context_state_authority: &context_state_authority,
    };

    let context = bytemuck::bytes_of(data.context_data());
    let context_state = ProofContextState::<U>::encode(
        &context_state_authority,
        T::PROOF_TYPE,
        data.context_data(),
    );

    format!(
        r#"{{"name":"{}","proof_type":{},"proof_data":"{}","context":"{}","context_state":"{}","verify_no_context":{},"verify_with_context":{},"verify_from_account_no_context":{},"verify_from_account_with_context":{}}}"#,
        name,
        ToPrimitive::to_u8(&T::PROOF_TYPE).unwrap(),
        hex(&raw),
        hex(context),
        hex(&context_state),
        ix_json(&inst.encode_verify_proof(None, data)),
        ix_json(&inst.encode_verify_proof(Some(info), data)),
        ix_json(&inst.encode_verify_proof_from_account(None, &proof_account, PROOF_DATA_OFFSET)),
        ix_json(&inst.encode_verify_proof_from_account(
            Some(info),
            &proof_account,
            PROOF_DATA_OFFSET_WITH_CONTEXT
        )),
    )
}

fn main() {
    let out: Vec<String> = vec![
        emit::<ZeroCiphertextProofData, _>(
            "ZeroCiphertext",
            ProofInstruction::VerifyZeroCiphertext,
        ),
        emit::<CiphertextCiphertextEqualityProofData, _>(
            "CiphertextCiphertextEquality",
            ProofInstruction::VerifyCiphertextCiphertextEquality,
        ),
        emit::<CiphertextCommitmentEqualityProofData, _>(
            "CiphertextCommitmentEquality",
            ProofInstruction::VerifyCiphertextCommitmentEquality,
        ),
        emit::<PubkeyValidityProofData, _>(
            "PubkeyValidity",
            ProofInstruction::VerifyPubkeyValidity,
        ),
        emit::<PercentageWithCapProofData, _>(
            "PercentageWithCap",
            ProofInstruction::VerifyPercentageWithCap,
        ),
        emit::<BatchedRangeProofU64Data, _>(
            "BatchedRangeProofU64",
            ProofInstruction::VerifyBatchedRangeProofU64,
        ),
        emit::<BatchedRangeProofU128Data, _>(
            "BatchedRangeProofU128",
            ProofInstruction::VerifyBatchedRangeProofU128,
        ),
        emit::<BatchedRangeProofU256Data, _>(
            "BatchedRangeProofU256",
            ProofInstruction::VerifyBatchedRangeProofU256,
        ),
        emit::<GroupedCiphertext2HandlesValidityProofData, _>(
            "GroupedCiphertext2HandlesValidity",
            ProofInstruction::VerifyGroupedCiphertext2HandlesValidity,
        ),
        emit::<BatchedGroupedCiphertext2HandlesValidityProofData, _>(
            "BatchedGroupedCiphertext2HandlesValidity",
            ProofInstruction::VerifyBatchedGroupedCiphertext2HandlesValidity,
        ),
        emit::<GroupedCiphertext3HandlesValidityProofData, _>(
            "GroupedCiphertext3HandlesValidity",
            ProofInstruction::VerifyGroupedCiphertext3HandlesValidity,
        ),
        emit::<BatchedGroupedCiphertext3HandlesValidityProofData, _>(
            "BatchedGroupedCiphertext3HandlesValidity",
            ProofInstruction::VerifyBatchedGroupedCiphertext3HandlesValidity,
        ),
    ];

    let context_state_account = addr(CONTEXT_STATE_ACCOUNT);
    let context_state_authority = addr(CONTEXT_STATE_AUTHORITY);
    let close = close_context_state(
        ContextStateInfo {
            context_state_account: &context_state_account,
            context_state_authority: &context_state_authority,
        },
        &addr(DESTINATION),
    );

    println!(
        r#"{{"program_id":"{}","close_context_state":{},"proofs":[{}]}}"#,
        hex(solana_zk_elgamal_proof_interface::ID.as_ref()),
        ix_json(&close),
        out.join(",")
    );
}
