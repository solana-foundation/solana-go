package token2022

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// cmbBuilders maps every parity vector to the Go builder call that must encode
// identically to its Rust counterpart.
func cmbBuilders() map[string]func() ([]solana.Instruction, error) {
	return map[string]func() ([]solana.Instruction, error){
		"initialize_mint": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialMintBurnInitializeMintInstruction(
				ctMint, *ctAuditorPubkey, ctDecryptableBalance), nil)
		},
		"update_decryptable_supply": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialMintBurnUpdateDecryptableSupplyInstruction(
				ctMint, ctAuthority, nil, ctDecryptableBalance), nil)
		},
		"update_decryptable_supply_multisig": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialMintBurnUpdateDecryptableSupplyInstruction(
				ctMint, ctAuthority, ctMultisig, ctDecryptableBalance), nil)
		},
		"rotate_supply_elgamal_pubkey_offset": func() ([]solana.Instruction, error) {
			return NewConfidentialMintBurnRotateSupplyElGamalPubkeyInstructions(
				ctMint, ctAuthority, nil, cmbNewSupplyElGamalPubkey,
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.CiphertextCiphertextEqualityProofData{}))
		},
		"rotate_supply_elgamal_pubkey_context": func() ([]solana.Instruction, error) {
			return NewConfidentialMintBurnRotateSupplyElGamalPubkeyInstructions(
				ctMint, ctAuthority, nil, cmbNewSupplyElGamalPubkey,
				zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCiphertextEqualityProofData](ctContextSingle))
		},
		"rotate_supply_elgamal_pubkey_multisig": func() ([]solana.Instruction, error) {
			return NewConfidentialMintBurnRotateSupplyElGamalPubkeyInstructions(
				ctMint, ctAuthority, ctMultisig, cmbNewSupplyElGamalPubkey,
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.CiphertextCiphertextEqualityProofData{}))
		},
		"mint_offset": func() ([]solana.Instruction, error) {
			return NewConfidentialMintBurnMintInstructions(
				ctTokenAccount, ctMint, ctCiphertextLo, ctCiphertextHi, ctAuthority, nil,
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.CiphertextCommitmentEqualityProofData{}),
				zkprogram.ProofLocationInstructionOffset(2, &proofdata.BatchedGroupedCiphertext3HandlesValidityProofData{}),
				zkprogram.ProofLocationInstructionOffset(3, &proofdata.BatchedRangeProofU128Data{}),
				ctDecryptableBalance)
		},
		"mint_context": func() ([]solana.Instruction, error) {
			return NewConfidentialMintBurnMintInstructions(
				ctTokenAccount, ctMint, ctCiphertextLo, ctCiphertextHi, ctAuthority, nil,
				zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCommitmentEqualityProofData](ctContextEquality),
				zkprogram.ProofLocationContextStateAccount[*proofdata.BatchedGroupedCiphertext3HandlesValidityProofData](ctContextValidity),
				zkprogram.ProofLocationContextStateAccount[*proofdata.BatchedRangeProofU128Data](ctContextRange),
				ctDecryptableBalance)
		},
		"mint_equality_context": func() ([]solana.Instruction, error) {
			return NewConfidentialMintBurnMintInstructions(
				ctTokenAccount, ctMint, ctCiphertextLo, ctCiphertextHi, ctAuthority, nil,
				zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCommitmentEqualityProofData](ctContextEquality),
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.BatchedGroupedCiphertext3HandlesValidityProofData{}),
				zkprogram.ProofLocationInstructionOffset(2, &proofdata.BatchedRangeProofU128Data{}),
				ctDecryptableBalance)
		},
		"mint_range_context_multisig": func() ([]solana.Instruction, error) {
			return NewConfidentialMintBurnMintInstructions(
				ctTokenAccount, ctMint, ctCiphertextLo, ctCiphertextHi, ctAuthority, ctMultisig,
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.CiphertextCommitmentEqualityProofData{}),
				zkprogram.ProofLocationInstructionOffset(2, &proofdata.BatchedGroupedCiphertext3HandlesValidityProofData{}),
				zkprogram.ProofLocationContextStateAccount[*proofdata.BatchedRangeProofU128Data](ctContextRange),
				ctDecryptableBalance)
		},
		"burn_offset": func() ([]solana.Instruction, error) {
			return NewConfidentialMintBurnBurnInstructions(
				ctTokenAccount, ctMint, ctDecryptableBalance, ctCiphertextLo, ctCiphertextHi,
				ctAuthority, nil,
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.CiphertextCommitmentEqualityProofData{}),
				zkprogram.ProofLocationInstructionOffset(2, &proofdata.BatchedGroupedCiphertext3HandlesValidityProofData{}),
				zkprogram.ProofLocationInstructionOffset(3, &proofdata.BatchedRangeProofU128Data{}))
		},
		"burn_context": func() ([]solana.Instruction, error) {
			return NewConfidentialMintBurnBurnInstructions(
				ctTokenAccount, ctMint, ctDecryptableBalance, ctCiphertextLo, ctCiphertextHi,
				ctAuthority, nil,
				zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCommitmentEqualityProofData](ctContextEquality),
				zkprogram.ProofLocationContextStateAccount[*proofdata.BatchedGroupedCiphertext3HandlesValidityProofData](ctContextValidity),
				zkprogram.ProofLocationContextStateAccount[*proofdata.BatchedRangeProofU128Data](ctContextRange))
		},
		"burn_equality_context": func() ([]solana.Instruction, error) {
			return NewConfidentialMintBurnBurnInstructions(
				ctTokenAccount, ctMint, ctDecryptableBalance, ctCiphertextLo, ctCiphertextHi,
				ctAuthority, ctMultisig,
				zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCommitmentEqualityProofData](ctContextEquality),
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.BatchedGroupedCiphertext3HandlesValidityProofData{}),
				zkprogram.ProofLocationInstructionOffset(2, &proofdata.BatchedRangeProofU128Data{}))
		},
		"apply_pending_burn": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialMintBurnApplyPendingBurnInstruction(ctMint, ctAuthority, nil), nil)
		},
		"apply_pending_burn_multisig": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialMintBurnApplyPendingBurnInstruction(ctMint, ctAuthority, ctMultisig), nil)
		},
	}
}

func TestConfidentialMintBurnRustParity(t *testing.T) {
	t.Parallel()
	runConfidentialRustParity(t, "testdata/confidential_mint_burn_rust_parity.json", cmbBuilders())
}
