package token2022

import (
	"bytes"
	"encoding/hex"
	stdjson "encoding/json"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

type ctParityAccount struct {
	Pubkey     string `json:"pubkey"`
	IsSigner   bool   `json:"is_signer"`
	IsWritable bool   `json:"is_writable"`
}

type ctParityInstruction struct {
	ProgramID string            `json:"program_id"`
	Data      string            `json:"data"`
	Accounts  []ctParityAccount `json:"accounts"`
}

type ctParityBuilder struct {
	Name         string                `json:"name"`
	Instructions []ctParityInstruction `json:"instructions"`
}

type ctParityData struct {
	ProgramID string            `json:"program_id"`
	Builders  []ctParityBuilder `json:"builders"`
}

// builtOne adapts a single-instruction builder to the shape the parity vectors are checked in.
func builtOne[T interface{ ValidateAndBuild() (*Instruction, error) }](inst T, err error) ([]solana.Instruction, error) {
	if err != nil {
		return nil, err
	}
	built, err := inst.ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{built}, nil
}

// ctBuilders maps every parity vector to the Go builder call that must encode
// identically to its Rust counterpart.
func ctBuilders() map[string]func() ([]solana.Instruction, error) {
	single := builtOne[*ConfidentialTransferExtension]
	noErr := func(inst *ConfidentialTransferExtension) ([]solana.Instruction, error) {
		return builtOne(inst, nil)
	}
	return map[string]func() ([]solana.Instruction, error){
		"initialize_mint": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferInitializeMintInstruction(
				ctMint, &ctAuthority, true, ctAuditorPubkey))
		},
		"initialize_mint_no_optionals": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferInitializeMintInstruction(
				ctMint, nil, false, nil))
		},
		"update_mint": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferUpdateMintInstruction(
				ctMint, ctAuthority, nil, true, ctAuditorPubkey))
		},
		"update_mint_multisig": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferUpdateMintInstruction(
				ctMint, ctAuthority, ctMultisig, false, nil))
		},
		"configure_account_offset": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferConfigureAccountInstructions(
				ctTokenAccount, ctMint, ctDecryptableBalance, ctMaxPendingCounter,
				ctAuthority, nil,
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.PubkeyValidityProofData{}))
		},
		"configure_account_context": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferConfigureAccountInstructions(
				ctTokenAccount, ctMint, ctDecryptableBalance, ctMaxPendingCounter,
				ctAuthority, nil,
				zkprogram.ProofLocationContextStateAccount[*proofdata.PubkeyValidityProofData](ctContextSingle))
		},
		"inner_configure_account_offset_minus_1": func() ([]solana.Instruction, error) {
			return single(NewConfidentialTransferInnerConfigureAccountInstruction(
				ctTokenAccount, ctMint, ctDecryptableBalance, ctMaxPendingCounter,
				ctAuthority, nil,
				zkprogram.ProofLocationInstructionOffset(-1, &proofdata.PubkeyValidityProofData{})))
		},
		"approve_account": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferApproveAccountInstruction(
				ctTokenAccount, ctMint, ctAuthority))
		},
		"empty_account_offset": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferEmptyAccountInstructions(
				ctTokenAccount, ctAuthority, nil,
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.ZeroCiphertextProofData{}))
		},
		"empty_account_context": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferEmptyAccountInstructions(
				ctTokenAccount, ctAuthority, nil,
				zkprogram.ProofLocationContextStateAccount[*proofdata.ZeroCiphertextProofData](ctContextSingle))
		},
		"deposit": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferDepositInstruction(
				ctTokenAccount, ctMint, ctAmount, ctDecimals, ctAuthority, nil))
		},
		"deposit_multisig": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferDepositInstruction(
				ctTokenAccount, ctMint, ctAmount, ctDecimals, ctAuthority, ctMultisig))
		},
		"withdraw_offset": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferWithdrawInstructions(
				ctTokenAccount, ctMint, ctAmount, ctDecimals, ctDecryptableBalance,
				ctAuthority, nil,
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.CiphertextCommitmentEqualityProofData{}),
				zkprogram.ProofLocationInstructionOffset(2, &proofdata.BatchedRangeProofU64Data{}))
		},
		"withdraw_context": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferWithdrawInstructions(
				ctTokenAccount, ctMint, ctAmount, ctDecimals, ctDecryptableBalance,
				ctAuthority, nil,
				zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCommitmentEqualityProofData](ctContextEquality),
				zkprogram.ProofLocationContextStateAccount[*proofdata.BatchedRangeProofU64Data](ctContextRange))
		},
		"withdraw_mixed": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferWithdrawInstructions(
				ctTokenAccount, ctMint, ctAmount, ctDecimals, ctDecryptableBalance,
				ctAuthority, nil,
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.CiphertextCommitmentEqualityProofData{}),
				zkprogram.ProofLocationContextStateAccount[*proofdata.BatchedRangeProofU64Data](ctContextRange))
		},
		"transfer_offset": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferTransferInstructions(
				ctTokenAccount, ctMint, ctDestination, ctDecryptableBalance,
				ctCiphertextLo, ctCiphertextHi, ctAuthority, nil,
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.CiphertextCommitmentEqualityProofData{}),
				zkprogram.ProofLocationInstructionOffset(2, &proofdata.BatchedGroupedCiphertext3HandlesValidityProofData{}),
				zkprogram.ProofLocationInstructionOffset(3, &proofdata.BatchedRangeProofU128Data{}))
		},
		"transfer_context": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferTransferInstructions(
				ctTokenAccount, ctMint, ctDestination, ctDecryptableBalance,
				ctCiphertextLo, ctCiphertextHi, ctAuthority, nil,
				zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCommitmentEqualityProofData](ctContextEquality),
				zkprogram.ProofLocationContextStateAccount[*proofdata.BatchedGroupedCiphertext3HandlesValidityProofData](ctContextValidity),
				zkprogram.ProofLocationContextStateAccount[*proofdata.BatchedRangeProofU128Data](ctContextRange))
		},
		"apply_pending_balance": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferApplyPendingBalanceInstruction(
				ctTokenAccount, ctMaxPendingCounter, ctDecryptableBalance, ctAuthority, nil))
		},
		"enable_confidential_credits": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferEnableConfidentialCreditsInstruction(
				ctTokenAccount, ctAuthority, nil))
		},
		"disable_confidential_credits": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferDisableConfidentialCreditsInstruction(
				ctTokenAccount, ctAuthority, nil))
		},
		"enable_non_confidential_credits": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferEnableNonConfidentialCreditsInstruction(
				ctTokenAccount, ctAuthority, nil))
		},
		"disable_non_confidential_credits": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferDisableNonConfidentialCreditsInstruction(
				ctTokenAccount, ctAuthority, nil))
		},
		"transfer_with_fee_offset": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferTransferWithFeeInstructions(
				ctTokenAccount, ctMint, ctDestination, ctDecryptableBalance,
				ctCiphertextLo, ctCiphertextHi, ctAuthority, nil,
				zkprogram.ProofLocationInstructionOffset(1, &proofdata.CiphertextCommitmentEqualityProofData{}),
				zkprogram.ProofLocationInstructionOffset(2, &proofdata.BatchedGroupedCiphertext3HandlesValidityProofData{}),
				zkprogram.ProofLocationInstructionOffset(3, &proofdata.PercentageWithCapProofData{}),
				zkprogram.ProofLocationInstructionOffset(4, &proofdata.BatchedGroupedCiphertext2HandlesValidityProofData{}),
				zkprogram.ProofLocationInstructionOffset(5, &proofdata.BatchedRangeProofU256Data{}))
		},
		"transfer_with_fee_context": func() ([]solana.Instruction, error) {
			return NewConfidentialTransferTransferWithFeeInstructions(
				ctTokenAccount, ctMint, ctDestination, ctDecryptableBalance,
				ctCiphertextLo, ctCiphertextHi, ctAuthority, nil,
				zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCommitmentEqualityProofData](ctContextEquality),
				zkprogram.ProofLocationContextStateAccount[*proofdata.BatchedGroupedCiphertext3HandlesValidityProofData](ctContextValidity),
				zkprogram.ProofLocationContextStateAccount[*proofdata.PercentageWithCapProofData](ctContextFeeSigma),
				zkprogram.ProofLocationContextStateAccount[*proofdata.BatchedGroupedCiphertext2HandlesValidityProofData](ctContextFeeValidity),
				zkprogram.ProofLocationContextStateAccount[*proofdata.BatchedRangeProofU256Data](ctContextRange))
		},
		"configure_account_with_registry": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferConfigureAccountWithRegistryInstruction(
				ctTokenAccount, ctMint, ctRegistry, &ctPayer))
		},
		"configure_account_with_registry_no_payer": func() ([]solana.Instruction, error) {
			return noErr(NewConfidentialTransferConfigureAccountWithRegistryInstruction(
				ctTokenAccount, ctMint, ctRegistry, nil))
		},
	}
}

func TestConfidentialTransferRustParity(t *testing.T) {
	t.Parallel()
	runConfidentialRustParity(t, "testdata/confidential_transfer_rust_parity.json", ctBuilders())
}

// runConfidentialRustParity checks every Go builder against the Rust encoding
// recorded for it, and holds the two sides in step: a vector without a builder
// and a builder without a vector are both failures.
func runConfidentialRustParity(t *testing.T, testdataPath string, builders map[string]func() ([]solana.Instruction, error)) {
	t.Helper()
	raw, err := os.ReadFile(testdataPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var parity ctParityData
	if err := stdjson.Unmarshal(raw, &parity); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !bytes.Equal(ProgramID.Bytes(), ctUnhex(t, parity.ProgramID)) {
		t.Fatalf("program ID = %x, want %s", ProgramID.Bytes(), parity.ProgramID)
	}

	seen := make(map[string]bool)
	for _, vector := range parity.Builders {
		build, ok := builders[vector.Name]
		if !ok {
			t.Errorf("no Go builder for vector %q", vector.Name)
			continue
		}
		seen[vector.Name] = true
		t.Run(vector.Name, func(t *testing.T) {
			t.Parallel()
			instructions, err := build()
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			if len(instructions) != len(vector.Instructions) {
				t.Fatalf("got %d instructions, want %d", len(instructions), len(vector.Instructions))
			}
			for i, want := range vector.Instructions {
				ctCheckInstruction(t, want, instructions[i])
			}
		})
	}
	for name := range builders {
		if !seen[name] {
			t.Errorf("builder %q has no rust parity vector", name)
		}
	}
}

// ctCheckInstruction compares a built instruction against the Rust encoding.
func ctCheckInstruction(t *testing.T, want ctParityInstruction, got solana.Instruction) {
	t.Helper()

	if !bytes.Equal(got.ProgramID().Bytes(), ctUnhex(t, want.ProgramID)) {
		t.Errorf("program ID = %x, want %s", got.ProgramID().Bytes(), want.ProgramID)
	}

	data, err := got.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	if !bytes.Equal(data, ctUnhex(t, want.Data)) {
		t.Errorf("data = %s, want %s", hex.EncodeToString(data), want.Data)
	}

	accounts := got.Accounts()
	if len(accounts) != len(want.Accounts) {
		t.Fatalf("got %d accounts, want %d", len(accounts), len(want.Accounts))
	}
	for i, wantAccount := range want.Accounts {
		if !bytes.Equal(accounts[i].PublicKey.Bytes(), ctUnhex(t, wantAccount.Pubkey)) {
			t.Errorf("account %d pubkey = %x, want %s", i, accounts[i].PublicKey.Bytes(), wantAccount.Pubkey)
		}
		if accounts[i].IsSigner != wantAccount.IsSigner {
			t.Errorf("account %d signer = %t, want %t", i, accounts[i].IsSigner, wantAccount.IsSigner)
		}
		if accounts[i].IsWritable != wantAccount.IsWritable {
			t.Errorf("account %d writable = %t, want %t", i, accounts[i].IsWritable, wantAccount.IsWritable)
		}
	}
}

func ctUnhex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("hex.DecodeString(%q): %v", s, err)
	}
	return b
}
