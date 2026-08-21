use bytemuck::Pod;
use solana_zk_elgamal_proof_interface::proof_data::{
    BatchedGroupedCiphertext2HandlesValidityProofData,
    BatchedGroupedCiphertext3HandlesValidityProofData, BatchedRangeProofU128Data,
    BatchedRangeProofU256Data, BatchedRangeProofU64Data, CiphertextCiphertextEqualityProofData,
    CiphertextCommitmentEqualityProofData, GroupedCiphertext2HandlesValidityProofData,
    GroupedCiphertext3HandlesValidityProofData, PercentageWithCapProofData, PodProofType,
    ProofType, PubkeyValidityProofData, ZeroCiphertextProofData,
};
use solana_zk_sdk::zk_elgamal_proof_program::VerifyZkProof;

use crate::{
    constants::{ERR_BAD_INPUT, ERR_PROOF_VERIFICATION, ERR_UNKNOWN_PROOF_TYPE, OK},
    memory::buffer_len,
};

#[no_mangle]
pub unsafe extern "C" fn zk_verify_proof(proof_type: u32, data_ptr: u32) -> i32 {
    let Ok(byte) = u8::try_from(proof_type) else {
        return ERR_UNKNOWN_PROOF_TYPE;
    };
    let Ok(proof_type) = ProofType::try_from(PodProofType(byte)) else {
        return ERR_UNKNOWN_PROOF_TYPE;
    };
    match proof_type {
        ProofType::PubkeyValidity => verify_as::<PubkeyValidityProofData>(data_ptr),
        ProofType::ZeroCiphertext => verify_as::<ZeroCiphertextProofData>(data_ptr),
        ProofType::CiphertextCommitmentEquality => {
            verify_as::<CiphertextCommitmentEqualityProofData>(data_ptr)
        }
        ProofType::CiphertextCiphertextEquality => {
            verify_as::<CiphertextCiphertextEqualityProofData>(data_ptr)
        }
        ProofType::PercentageWithCap => verify_as::<PercentageWithCapProofData>(data_ptr),
        ProofType::BatchedRangeProofU64 => verify_as::<BatchedRangeProofU64Data>(data_ptr),
        ProofType::BatchedRangeProofU128 => verify_as::<BatchedRangeProofU128Data>(data_ptr),
        ProofType::BatchedRangeProofU256 => verify_as::<BatchedRangeProofU256Data>(data_ptr),
        ProofType::GroupedCiphertext2HandlesValidity => {
            verify_as::<GroupedCiphertext2HandlesValidityProofData>(data_ptr)
        }
        ProofType::GroupedCiphertext3HandlesValidity => {
            verify_as::<GroupedCiphertext3HandlesValidityProofData>(data_ptr)
        }
        ProofType::BatchedGroupedCiphertext2HandlesValidity => {
            verify_as::<BatchedGroupedCiphertext2HandlesValidityProofData>(data_ptr)
        }
        ProofType::BatchedGroupedCiphertext3HandlesValidity => {
            verify_as::<BatchedGroupedCiphertext3HandlesValidityProofData>(data_ptr)
        }
        ProofType::Uninitialized => ERR_UNKNOWN_PROOF_TYPE,
    }
}

unsafe fn verify_as<T: Pod + VerifyZkProof>(ptr: u32) -> i32 {
    let Some(len) = buffer_len(ptr) else {
        return ERR_BAD_INPUT;
    };
    let bytes = std::slice::from_raw_parts(ptr as *const u8, len as usize);
    let data: &T = match bytemuck::try_from_bytes(bytes) {
        Ok(data) => data,
        Err(_) => return ERR_BAD_INPUT,
    };
    match data.verify_proof() {
        Ok(()) => OK,
        Err(_) => ERR_PROOF_VERIFICATION,
    }
}
