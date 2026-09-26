package token2022

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// ctfBuilders maps every parity vector to the Go builder call that must encode
// identically to its Rust counterpart.
func ctfBuilders() map[string]func() ([]solana.Instruction, error) {
	offsetProof := func(offset int8) zkprogram.ProofLocation[*proofdata.CiphertextCiphertextEqualityProofData] {
		return zkprogram.ProofLocationInstructionOffset(offset, &proofdata.CiphertextCiphertextEqualityProofData{})
	}
	contextProof := zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCiphertextEqualityProofData](ctContextSingle)
	return map[string]func() ([]solana.Instruction, error){
		"initialize_config": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialTransferFeeInitializeConfigInstruction(
				ctMint, &ctAuthority, *ctAuditorPubkey), nil)
		},
		"initialize_config_no_authority": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialTransferFeeInitializeConfigInstruction(
				ctMint, nil, *ctAuditorPubkey), nil)
		},
		"withdraw_withheld_tokens_from_mint_offset": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferFeeWithdrawWithheldTokensFromMintInstructions(
				ctMint, ctDestination, ctDecryptableBalance, ctAuthority, nil, offsetProof(1))
		},
		"withdraw_withheld_tokens_from_mint_context": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferFeeWithdrawWithheldTokensFromMintInstructions(
				ctMint, ctDestination, ctDecryptableBalance, ctAuthority, nil, contextProof)
		},
		"withdraw_withheld_tokens_from_mint_multisig": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferFeeWithdrawWithheldTokensFromMintInstructions(
				ctMint, ctDestination, ctDecryptableBalance, ctAuthority, ctMultisig, offsetProof(1))
		},
		"inner_withdraw_withheld_tokens_from_mint_offset_minus_1": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromMintInstruction(
				ctMint, ctDestination, ctDecryptableBalance, ctAuthority, nil, offsetProof(-1)))
		},
		"withdraw_withheld_tokens_from_accounts_offset": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferFeeWithdrawWithheldTokensFromAccountsInstructions(
				ctMint, ctDestination, ctDecryptableBalance, ctAuthority, nil, ctfSources, offsetProof(1))
		},
		"withdraw_withheld_tokens_from_accounts_context": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferFeeWithdrawWithheldTokensFromAccountsInstructions(
				ctMint, ctDestination, ctDecryptableBalance, ctAuthority, nil, ctfSources, contextProof)
		},
		// The source accounts trail the multisig signers.
		"withdraw_withheld_tokens_from_accounts_multisig": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferFeeWithdrawWithheldTokensFromAccountsInstructions(
				ctMint, ctDestination, ctDecryptableBalance, ctAuthority, ctMultisig, ctfSources, offsetProof(1))
		},
		"withdraw_withheld_tokens_from_accounts_no_sources": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferFeeWithdrawWithheldTokensFromAccountsInstructions(
				ctMint, ctDestination, ctDecryptableBalance, ctAuthority, nil, nil, contextProof)
		},
		"inner_withdraw_withheld_tokens_from_accounts_offset_minus_1": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromAccountsInstruction(
				ctMint, ctDestination, ctDecryptableBalance, ctAuthority, nil, ctfSources, offsetProof(-1)))
		},
		"harvest_withheld_tokens_to_mint": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialTransferFeeHarvestWithheldTokensToMintInstruction(ctMint, ctfSources), nil)
		},
		"harvest_withheld_tokens_to_mint_no_sources": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialTransferFeeHarvestWithheldTokensToMintInstruction(ctMint, nil), nil)
		},
		"enable_harvest_to_mint": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialTransferFeeEnableHarvestToMintInstruction(ctMint, ctAuthority, nil), nil)
		},
		"enable_harvest_to_mint_multisig": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialTransferFeeEnableHarvestToMintInstruction(ctMint, ctAuthority, ctMultisig), nil)
		},
		"disable_harvest_to_mint": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialTransferFeeDisableHarvestToMintInstruction(ctMint, ctAuthority, nil), nil)
		},
		"disable_harvest_to_mint_multisig": func() ([]solana.Instruction, error) {
			return builtOne(NewConfidentialTransferFeeDisableHarvestToMintInstruction(ctMint, ctAuthority, ctMultisig), nil)
		},
	}
}

func TestConfidentialTransferFeeRustParity(t *testing.T) {
	t.Parallel()
	runConfidentialRustParity(t, "testdata/confidential_transfer_fee_rust_parity.json", ctfBuilders())
}
