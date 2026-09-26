package token2022

import (
	"encoding"
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

func TestConfidentialTransferFeeDataRoundTrip(t *testing.T) {
	t.Parallel()
	testConfidentialDataRoundTrip(t, []encoding.BinaryMarshaler{
		ConfidentialTransferFeeInitializeConfigData{Authority: ctAuthority, WithdrawWithheldAuthorityElGamalPubkey: *ctAuditorPubkey},
		ConfidentialTransferFeeWithdrawWithheldTokensFromMintData{ProofInstructionOffset: -2, NewDecryptableAvailableBalance: ctDecryptableBalance},
		ConfidentialTransferFeeWithdrawWithheldTokensFromAccountsData{NumTokenAccounts: 2, ProofInstructionOffset: 1, NewDecryptableAvailableBalance: ctDecryptableBalance},
		ConfidentialTransferFeeHarvestWithheldTokensToMintData{},
		ConfidentialTransferFeeEnableHarvestToMintData{},
		ConfidentialTransferFeeDisableHarvestToMintData{},
	})
}

func TestConfidentialTransferFeeTypedDecode(t *testing.T) {
	t.Parallel()
	inner, err := NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromAccountsInstruction(
		ctMint, ctDestination, ctDecryptableBalance, ctAuthority, nil, ctfSources,
		zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCiphertextEqualityProofData](ctContextSingle))
	if err != nil {
		t.Fatalf("builder: %v", err)
	}
	built, err := inner.ValidateAndBuild()
	if err != nil {
		t.Fatalf("ValidateAndBuild: %v", err)
	}
	sub := decodeBuiltConfidentialSubinstruction[ConfidentialTransferFeeSubInstructionData, *ConfidentialTransferFeeExtension](t, built)
	withdrawalData, ok := sub.(*ConfidentialTransferFeeWithdrawWithheldTokensFromAccountsData)
	if !ok {
		t.Fatalf("sub-instruction decoded to %T, want *ConfidentialTransferFeeWithdrawWithheldTokensFromAccountsData", sub)
	}
	if int(withdrawalData.NumTokenAccounts) != len(ctfSources) {
		t.Errorf("decoded NumTokenAccounts = %d, want %d", withdrawalData.NumTokenAccounts, len(ctfSources))
	}
}

func TestConfidentialTransferFeeDecodeRejectsMalformed(t *testing.T) {
	t.Parallel()
	testConfidentialDecodeRejectsMalformed(
		t, &ConfidentialTransferFeeExtension{}, ConfidentialTransferFee_WithdrawWithheldTokensFromMint, ConfidentialTransferFee_HarvestWithheldTokensToMint)
}

// The source accounts trail the multisig signers, which the Signers slice of
// the other sub-instructions cannot express.
func TestConfidentialTransferFeeSourcesFollowSigners(t *testing.T) {
	t.Parallel()
	inner, err := NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromAccountsInstruction(
		ctMint, ctDestination, ctDecryptableBalance, ctAuthority, ctMultisig, ctfSources,
		zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCiphertextEqualityProofData](ctContextSingle))
	if err != nil {
		t.Fatalf("builder: %v", err)
	}
	want := []solana.PublicKey{ctMint, ctDestination, ctContextSingle, ctAuthority}
	want = append(want, ctMultisig...)
	want = append(want, ctfSources...)

	accounts := inner.GetAccounts()
	if len(accounts) != len(want) {
		t.Fatalf("got %d accounts, want %d", len(accounts), len(want))
	}
	for i, wantKey := range want {
		if accounts[i].PublicKey != wantKey {
			t.Errorf("account %d = %s, want %s", i, accounts[i].PublicKey, wantKey)
		}
	}
	// The authority delegates signing to the multisig members.
	if accounts[3].IsSigner {
		t.Error("multisig authority is marked as a signer")
	}
	for i := range ctMultisig {
		if !accounts[4+i].IsSigner {
			t.Errorf("multisig signer %d is not marked as a signer", i)
		}
	}
	for i := range ctfSources {
		if source := accounts[4+len(ctMultisig)+i]; !source.IsWritable || source.IsSigner {
			t.Errorf("source %d = %+v, want writable non-signer", i, source)
		}
	}
}

func TestConfidentialTransferFeeRejectsTooManySources(t *testing.T) {
	t.Parallel()
	sources := make([]solana.PublicKey, 256)
	_, err := NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromAccountsInstruction(
		ctMint, ctDestination, ctDecryptableBalance, ctAuthority, nil, sources,
		zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCiphertextEqualityProofData](ctContextSingle))
	if err == nil || !strings.Contains(err.Error(), "source") {
		t.Errorf("256 sources err = %v, want source account count error", err)
	}
}

// The outer builders append the proof directly after the withdraw instruction,
// so the offset they are handed must be 1.
func TestConfidentialTransferFeeOuterOffsetValidation(t *testing.T) {
	t.Parallel()
	proofAt := func(offset int8) zkprogram.ProofLocation[*proofdata.CiphertextCiphertextEqualityProofData] {
		return zkprogram.ProofLocationInstructionOffset(offset, &proofdata.CiphertextCiphertextEqualityProofData{})
	}
	_, err := NewConfidentialTransferFeeWithdrawWithheldTokensFromMintInstructions(
		ctMint, ctDestination, ctDecryptableBalance, ctAuthority, nil, proofAt(2))
	wantProofOffsetError(t, err, "offset 2")

	_, err = NewConfidentialTransferFeeWithdrawWithheldTokensFromAccountsInstructions(
		ctMint, ctDestination, ctDecryptableBalance, ctAuthority, nil, ctfSources, proofAt(-1))
	wantProofOffsetError(t, err, "offset -1")
}

func TestConfidentialTransferFeeRejectsUnsetProofLocation(t *testing.T) {
	t.Parallel()
	testConfidentialRejectsUnsetProofLocation(t,
		func(location zkprogram.ProofLocation[*proofdata.CiphertextCiphertextEqualityProofData]) error {
			_, err := NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromMintInstruction(
				ctMint, ctDestination, ctDecryptableBalance, ctAuthority, nil, location)
			return err
		})
}

func TestConfidentialTransferFeeRawInstruction(t *testing.T) {
	t.Parallel()
	raw := NewConfidentialTransferFeeInstruction(
		ConfidentialTransferFee_InitializeConfidentialTransferFeeConfig,
		ConfidentialTransferFeeInitializeConfigData{
			Authority:                              ctAuthority,
			WithdrawWithheldAuthorityElGamalPubkey: *ctAuditorPubkey,
		}.bytes(),
		*solana.Meta(ctMint).WRITE(),
	)
	typed := NewConfidentialTransferFeeInitializeConfigInstruction(
		ctMint, &ctAuthority, *ctAuditorPubkey)
	testConfidentialRawMatchesTyped(t, raw, typed)
}
