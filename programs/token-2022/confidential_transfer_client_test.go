package token2022

import (
	"encoding/binary"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
	"github.com/stretchr/testify/require"
)

func TestConfidentialTransferClient(t *testing.T) {
	t.Parallel()
	forEachProofLocation(t, "ConfigureAccount", testClientConfigureAccount)
	forEachProofLocation(t, "EmptyAccount", testClientEmptyAccount)
	forEachProofLocation(t, "Withdraw", testClientWithdraw)
	forEachProofLocation(t, "Transfer", testClientTransfer)
	forEachProofLocation(t, "TransferWithFee", testClientTransferWithFee)
}

func testClientConfigureAccount(t *testing.T, areProofsInline bool) {
	_, kp, aesKey := makeAccountState(t, 0, 0, 0, 0)
	instructions, err := ConfidentialTransferConfigureTokenAccount(
		ctTokenAccount, ctMint, ctAuthority, nil, orNil(areProofsInline, &ctContextSingle), nil, kp, aesKey)
	require.NoError(t, err)
	configure := clientData[*ConfidentialTransferConfigureAccountData](t, instructions)
	checkDecryptable(t, aesKey, configure.DecryptableZeroBalance, 0)
	checkClientProofs(t, instructions, areProofsInline, pubkeyValidityProofs,
		[]int8{configure.ProofInstructionOffset})
}

func testClientEmptyAccount(t *testing.T, areProofsInline bool) {
	empty, kp, _ := makeAccountState(t, 0, 0, 0, 0)
	instructions, err := ConfidentialTransferEmptyAccount(
		ctTokenAccount, ctAuthority, nil, orNil(areProofsInline, &ctContextSingle), NewEmptyAccountInfo(empty), kp)
	require.NoError(t, err)
	emptyData := clientData[*ConfidentialTransferEmptyAccountData](t, instructions)
	checkClientProofs(t, instructions, areProofsInline, zeroCiphertextProofs,
		[]int8{emptyData.ProofInstructionOffset})
}

func testClientWithdraw(t *testing.T, areProofsInline bool) {
	spender, kp, aesKey := makeAccountState(t, 0, 0, clientBalance, 0)
	instructions, err := ConfidentialTransferWithdraw(
		ctTokenAccount, ctMint, ctAuthority, nil,
		orNil(areProofsInline, &ctContextEquality), orNil(areProofsInline, &ctContextRange),
		clientAmount, ctDecimals, NewWithdrawAccountInfo(spender), kp, aesKey)
	require.NoError(t, err)
	withdraw := clientData[*ConfidentialTransferWithdrawData](t, instructions)
	if withdraw.Amount != clientAmount || withdraw.Decimals != ctDecimals {
		t.Errorf("amount/decimals = (%d, %d), want (%d, %d)", withdraw.Amount, withdraw.Decimals, clientAmount, ctDecimals)
	}
	checkDecryptable(t, aesKey, withdraw.NewDecryptableAvailableBalance,
		clientBalance-clientAmount)
	checkClientProofs(t, instructions, areProofsInline, withdrawProofs,
		[]int8{withdraw.EqualityProofInstructionOffset, withdraw.RangeProofInstructionOffset})
}

func testClientTransfer(t *testing.T, areProofsInline bool) {
	spender, kp, aesKey := makeAccountState(t, 0, 0, clientBalance, 0)
	destination, auditor := genKeyPair(t), genKeyPair(t)
	instructions, err := ConfidentialTransferTransfer(
		ctTokenAccount, ctDestination, ctMint, ctAuthority, nil,
		orNil(areProofsInline, &ctContextEquality), orNil(areProofsInline, validityProofAccount), orNil(areProofsInline, &ctContextRange),
		clientAmount, NewTransferAccountInfo(spender), kp, aesKey, destination.Pubkey, &auditor.Pubkey)
	require.NoError(t, err)
	transfer := clientData[*ConfidentialTransferTransferData](t, instructions)
	checkDecryptable(t, aesKey, transfer.NewSourceDecryptableAvailableBalance,
		clientBalance-clientAmount)
	checkCiphertexts(t, areProofsInline, auditor, clientAmount,
		transfer.TransferAmountAuditorCiphertextLo, transfer.TransferAmountAuditorCiphertextHi)
	checkClientProofs(t, instructions, areProofsInline, splitProofs, []int8{
		transfer.EqualityProofInstructionOffset,
		transfer.CiphertextValidityProofInstructionOffset,
		transfer.RangeProofInstructionOffset,
	})
}

func testClientTransferWithFee(t *testing.T, areProofsInline bool) {
	spender, kp, aesKey := makeAccountState(t, 0, 0, clientBalance, 0)
	destination, auditor, withheld := genKeyPair(t), genKeyPair(t), genKeyPair(t)
	instructions, err := ConfidentialTransferTransferWithFee(
		ctTokenAccount, ctDestination, ctMint, ctAuthority, nil,
		orNil(areProofsInline, &ctContextEquality), orNil(areProofsInline, validityProofAccount),
		orNil(areProofsInline, &ctContextFeeSigma), orNil(areProofsInline, &ctContextFeeValidity), orNil(areProofsInline, &ctContextRange),
		clientAmount, NewTransferAccountInfo(spender), kp, aesKey,
		destination.Pubkey, &auditor.Pubkey, withheld.Pubkey, 250, 50)
	require.NoError(t, err)
	transfer := clientData[*ConfidentialTransferTransferWithFeeData](t, instructions)
	checkDecryptable(t, aesKey, transfer.NewSourceDecryptableAvailableBalance,
		clientBalance-clientAmount)
	checkCiphertexts(t, areProofsInline, auditor, clientAmount,
		transfer.TransferAmountAuditorCiphertextLo, transfer.TransferAmountAuditorCiphertextHi)
	checkClientProofs(t, instructions, areProofsInline, transferWithFeeProofs, []int8{
		transfer.EqualityProofInstructionOffset,
		transfer.TransferAmountCiphertextValidityProofInstructionOffset,
		transfer.FeeSigmaProofInstructionOffset,
		transfer.FeeCiphertextValidityProofInstructionOffset,
		transfer.RangeProofInstructionOffset,
	})
}

// TestConfidentialTransferClientMixedProofLocations checks the limitation that an operation cannot
// take one proof from a context state account and inline another.
func TestConfidentialTransferClientMixedProofLocations(t *testing.T) {
	t.Parallel()
	state, kp, aesKey := makeAccountState(t, 0, 0, 1000, 0)
	_, err := ConfidentialTransferWithdraw(
		ctTokenAccount, ctMint, ctAuthority, nil, &ctContextEquality, nil,
		10, ctDecimals, NewWithdrawAccountInfo(state), kp, aesKey)
	wantProofOffsetError(t, err, "ConfidentialTransferWithdraw")
}

// TestConfidentialTransferCreateContextStateAccountRejectsMissingProof checks
// a proof is required, rather than panicking on the nil.
func TestConfidentialTransferCreateContextStateAccountRejectsMissingProof(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name  string
		proof proofdata.ProofData
	}{
		{"nil interface", nil},
		{"typed nil", (*proofdata.PubkeyValidityProofData)(nil)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if _, err := ConfidentialTransferCreateContextStateAccount(
				ctContextSingle, ctAuthority, ctPayer, testCase.proof, 1); err == nil {
				t.Fatal("got nil error, want a missing proof data error")
			}
		})
	}
}

// TestConfidentialTransferCreateContextStateAccountFromRecord checks the
// variant reading the proof out of a record program account
func TestConfidentialTransferCreateContextStateAccountFromRecord(t *testing.T) {
	t.Parallel()
	const rentLamports = uint64(7_654_321)

	instructions, err := ConfidentialTransferCreateContextStateAccountFromRecord(
		ctContextRange, ctAuthority, ctPayer, ctRecord,
		proofdata.ProofTypeBatchedRangeProofU128, rentLamports)
	require.NoError(t, err)
	if len(instructions) != 2 {
		t.Fatalf("got %d instructions, want 2", len(instructions))
	}

	accountCreationInstructionData, err := instructions[0].Data()
	require.NoError(t, err)
	decodedAccountCreationInstructionData, err := system.DecodeInstruction(instructions[0].Accounts(), accountCreationInstructionData)
	require.NoError(t, err)
	accountCreationData, ok := decodedAccountCreationInstructionData.Impl.(*system.CreateAccount)
	if !ok {
		t.Fatalf("first instruction is %T, want system.CreateAccount", decodedAccountCreationInstructionData.Impl)
	}
	space, err := zkprogram.ProofContextSize(proofdata.ProofTypeBatchedRangeProofU128)
	require.NoError(t, err)
	if *accountCreationData.Lamports != rentLamports || *accountCreationData.Space != space || !accountCreationData.Owner.Equals(zkprogram.ProgramID) {
		t.Errorf("create account lamports/space/owner = (%d, %d, %s), want (%d, %d, %s)",
			*accountCreationData.Lamports, *accountCreationData.Space, accountCreationData.Owner, rentLamports, space, zkprogram.ProgramID)
	}
	if decodedPayer := accountCreationData.GetFundingAccount().PublicKey; !decodedPayer.Equals(ctPayer) {
		t.Errorf("funding account = %s, want %s", decodedPayer, ctPayer)
	}
	if decodedContextAccount := accountCreationData.GetNewAccount().PublicKey; !decodedContextAccount.Equals(ctContextRange) {
		t.Errorf("new account = %s, want %s", decodedContextAccount, ctContextRange)
	}

	if decodedProgramID := instructions[1].ProgramID(); !decodedProgramID.Equals(zkprogram.ProgramID) {
		t.Fatalf("verify program ID = %s, want %s", decodedProgramID, zkprogram.ProgramID)
	}
	verifyInstructionData, err := instructions[1].Data()
	require.NoError(t, err)
	if got := zkprogram.ProofInstruction(verifyInstructionData[0]); got != zkprogram.VerifyBatchedRangeProofU128 {
		t.Errorf("proof instruction = %s, want %s", got, zkprogram.VerifyBatchedRangeProofU128)
	}
	wantAccounts := []solana.PublicKey{ctRecord, ctContextRange, ctAuthority}
	verifyInstructionAccounts := instructions[1].Accounts()
	if len(verifyInstructionAccounts) != len(wantAccounts) {
		t.Fatalf("verify accounts = %v, want %v", verifyInstructionAccounts, wantAccounts)
	}
	for i, account := range wantAccounts {
		if !verifyInstructionAccounts[i].PublicKey.Equals(account) {
			t.Errorf("verify account %d = %s, want %s", i, verifyInstructionAccounts[i].PublicKey, account)
		}
	}
	if !verifyInstructionAccounts[1].IsWritable {
		t.Error("context state account is not writable")
	}
	if len(verifyInstructionData) != 5 || binary.LittleEndian.Uint32(verifyInstructionData[1:]) != recordAccountDataOffset {
		t.Errorf("verify instruction data = % x, want a %d proof account offset", verifyInstructionData, recordAccountDataOffset)
	}
}
