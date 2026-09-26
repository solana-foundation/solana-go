package token2022

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

func TestConfidentialTransferDataRoundTrip(t *testing.T) {
	t.Parallel()
	testConfidentialDataRoundTrip(t, []encoding.BinaryMarshaler{
		ConfidentialTransferInitializeMintData{Authority: ctAuthority, AutoApproveNewAccounts: true, AuditorElGamalPubkey: *ctAuditorPubkey},
		ConfidentialTransferUpdateMintData{AutoApproveNewAccounts: true, AuditorElGamalPubkey: *ctAuditorPubkey},
		ConfidentialTransferConfigureAccountData{DecryptableZeroBalance: ctDecryptableBalance, MaximumPendingBalanceCreditCounter: ctMaxPendingCounter, ProofInstructionOffset: -3},
		ConfidentialTransferEmptyAccountData{ProofInstructionOffset: 1},
		ConfidentialTransferDepositData{Amount: ctAmount, Decimals: ctDecimals},
		ConfidentialTransferWithdrawData{Amount: ctAmount, Decimals: ctDecimals, NewDecryptableAvailableBalance: ctDecryptableBalance, EqualityProofInstructionOffset: 1, RangeProofInstructionOffset: 2},
		ConfidentialTransferTransferData{NewSourceDecryptableAvailableBalance: ctDecryptableBalance, TransferAmountAuditorCiphertextLo: ctCiphertextLo, TransferAmountAuditorCiphertextHi: ctCiphertextHi, EqualityProofInstructionOffset: 1, CiphertextValidityProofInstructionOffset: 2, RangeProofInstructionOffset: 3},
		ConfidentialTransferApplyPendingBalanceData{ExpectedPendingBalanceCreditCounter: ctMaxPendingCounter, NewDecryptableAvailableBalance: ctDecryptableBalance},
		ConfidentialTransferTransferWithFeeData{NewSourceDecryptableAvailableBalance: ctDecryptableBalance, TransferAmountAuditorCiphertextLo: ctCiphertextLo, TransferAmountAuditorCiphertextHi: ctCiphertextHi, EqualityProofInstructionOffset: 1, TransferAmountCiphertextValidityProofInstructionOffset: 2, FeeSigmaProofInstructionOffset: 3, FeeCiphertextValidityProofInstructionOffset: 4, RangeProofInstructionOffset: 5},
	})
}

func TestConfidentialTransferTypedDecode(t *testing.T) {
	t.Parallel()
	built, err := NewConfidentialTransferDepositInstruction(
		ctTokenAccount, ctMint, ctAmount, ctDecimals, ctAuthority, nil).ValidateAndBuild()
	if err != nil {
		t.Fatalf("ValidateAndBuild: %v", err)
	}
	sub := decodeBuiltConfidentialSubinstruction[ConfidentialTransferSubInstructionData, *ConfidentialTransferExtension](t, built)
	deposit, ok := sub.(*ConfidentialTransferDepositData)
	if !ok {
		t.Fatalf("sub-instruction decoded to %T, want *ConfidentialTransferDepositData", sub)
	}
	if deposit.Amount != ctAmount || deposit.Decimals != ctDecimals {
		t.Errorf("decoded data = %+v, want amount %d decimals %d", deposit, ctAmount, ctDecimals)
	}
}

func TestConfidentialTransferDecodeRejectsMalformed(t *testing.T) {
	t.Parallel()
	testConfidentialDecodeRejectsMalformed[ConfidentialTransferSubInstructionData](
		t, &ConfidentialTransferExtension{}, ConfidentialTransfer_Withdraw, ConfidentialTransfer_ApproveAccount)
}

func TestConfidentialTransferOuterOffsetValidation(t *testing.T) {
	t.Parallel()
	_, err := NewConfidentialTransferConfigureAccountInstructions(
		ctTokenAccount, ctMint, ctDecryptableBalance, ctMaxPendingCounter,
		ctAuthority, nil,
		zkprogram.ProofLocationInstructionOffset(2, &proofdata.PubkeyValidityProofData{}))
	wantProofOffsetError(t, err, "offset 2")

	_, err = NewConfidentialTransferWithdrawInstructions(
		ctTokenAccount, ctMint, ctAmount, ctDecimals, ctDecryptableBalance,
		ctAuthority, nil,
		zkprogram.ProofLocationInstructionOffset(1, &proofdata.CiphertextCommitmentEqualityProofData{}),
		zkprogram.ProofLocationInstructionOffset(3, &proofdata.BatchedRangeProofU64Data{}))
	wantProofOffsetError(t, err, "offsets 1,3")

	// Context state first, then offset: the offset instruction is the only
	// appended one, so it must be 1.
	instructions, err := NewConfidentialTransferWithdrawInstructions(
		ctTokenAccount, ctMint, ctAmount, ctDecimals, ctDecryptableBalance,
		ctAuthority, nil,
		zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCommitmentEqualityProofData](ctContextEquality),
		zkprogram.ProofLocationInstructionOffset(1, &proofdata.BatchedRangeProofU64Data{}))
	if err != nil {
		t.Fatalf("context+offset(1): %v", err)
	}
	if len(instructions) != 2 {
		t.Errorf("context+offset(1) built %d instructions, want 2", len(instructions))
	}
}

func TestConfidentialTransferRejectsUnsetProofLocation(t *testing.T) {
	t.Parallel()
	testConfidentialRejectsUnsetProofLocation(t,
		func(location zkprogram.ProofLocation[*proofdata.ZeroCiphertextProofData]) error {
			_, err := NewConfidentialTransferInnerEmptyAccountInstruction(
				ctTokenAccount, ctAuthority, nil, location)
			return err
		})
}

func TestConfidentialTransferRawInstruction(t *testing.T) {
	t.Parallel()
	rawInstructionData := append(binary.LittleEndian.AppendUint64(nil, ctAmount), ctDecimals)
	instruction, err := NewConfidentialTransferInstruction(
		ConfidentialTransfer_Deposit, rawInstructionData,
		*solana.Meta(ctTokenAccount).WRITE(),
		*solana.Meta(ctMint),
		*solana.Meta(ctAuthority).SIGNER(),
	).ValidateAndBuild()
	if err != nil {
		t.Fatalf("ValidateAndBuild: %v", err)
	}
	encodedInstruction, err := instruction.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	expectedEncodedInstruction := append([]byte{Instruction_ConfidentialTransferExtension, ConfidentialTransfer_Deposit}, rawInstructionData...)
	if !bytes.Equal(encodedInstruction, expectedEncodedInstruction) {
		t.Errorf("encoded data = %x, want %x", encodedInstruction, expectedEncodedInstruction)
	}

	decodedInstruction, err := DecodeInstruction(instruction.Accounts(), encodedInstruction)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	decodedCTInstruction, ok := decodedInstruction.Impl.(*ConfidentialTransferExtension)
	if !ok {
		t.Fatalf("decoded to %T, want *ConfidentialTransferExtension", decodedInstruction.Impl)
	}
	if decodedCTInstruction.SubInstruction != ConfidentialTransfer_Deposit || !bytes.Equal(decodedCTInstruction.RawData, rawInstructionData) {
		t.Errorf("decoded raw form = (%d, %x), want (%d, %x)",
			decodedCTInstruction.SubInstruction, decodedCTInstruction.RawData, ConfidentialTransfer_Deposit, rawInstructionData)
	}
	decodedInstructionData, err := decodedCTInstruction.DecodeSubInstructionData()
	if err != nil {
		t.Fatalf("DecodeSubInstructionData: %v", err)
	}
	ctDepositInstuctionData, ok := decodedInstructionData.(*ConfidentialTransferDepositData)
	if !ok {
		t.Fatalf("sub-instruction decoded to %T, want *ConfidentialTransferDepositData", decodedInstructionData)
	}
	if ctDepositInstuctionData.Amount != ctAmount || ctDepositInstuctionData.Decimals != ctDecimals {
		t.Errorf("decoded data = %+v, want amount %d decimals %d", ctDepositInstuctionData, ctAmount, ctDecimals)
	}
}

// TestConfidentialTransferMultisigSigners checks the authority only signs
// itself when there are no multisig signers.
func TestConfidentialTransferMultisigSigners(t *testing.T) {
	t.Parallel()
	single := NewConfidentialTransferDepositInstruction(
		ctTokenAccount, ctMint, ctAmount, ctDecimals, ctAuthority, nil)
	accounts := single.GetAccounts()
	authority := accounts[len(accounts)-1]
	if !authority.IsSigner {
		t.Error("authority is not a signer without multisig signers")
	}

	multi := NewConfidentialTransferDepositInstruction(
		ctTokenAccount, ctMint, ctAmount, ctDecimals, ctAuthority, ctMultisig)
	accounts = multi.GetAccounts()
	if got, want := len(accounts), 3+len(ctMultisig); got != want {
		t.Fatalf("got %d accounts, want %d", got, want)
	}
	if accounts[2].IsSigner {
		t.Error("multisig authority must not be a signer")
	}
	for i, signer := range accounts[3:] {
		if !signer.IsSigner {
			t.Errorf("multisig signer %d is not a signer", i)
		}
		if signer.PublicKey != ctMultisig[i] {
			t.Errorf("multisig signer %d = %s, want %s", i, signer.PublicKey, ctMultisig[i])
		}
	}
}
