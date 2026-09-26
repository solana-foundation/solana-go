package token2022

import (
	"encoding"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

func TestConfidentialMintBurnDataRoundTrip(t *testing.T) {
	t.Parallel()
	testConfidentialDataRoundTrip(t, []encoding.BinaryMarshaler{
		ConfidentialMintBurnInitializeMintData{SupplyElGamalPubkey: *ctAuditorPubkey, DecryptableSupply: ctDecryptableBalance},
		ConfidentialMintBurnRotateSupplyElGamalPubkeyData{NewSupplyElGamalPubkey: cmbNewSupplyElGamalPubkey, ProofInstructionOffset: -1},
		ConfidentialMintBurnUpdateDecryptableSupplyData{NewDecryptableSupply: ctDecryptableBalance},
		ConfidentialMintBurnMintData{NewDecryptableSupply: ctDecryptableBalance, MintAmountAuditorCiphertextLo: ctCiphertextLo, MintAmountAuditorCiphertextHi: ctCiphertextHi, EqualityProofInstructionOffset: 1, CiphertextValidityProofInstructionOffset: 2, RangeProofInstructionOffset: 3},
		ConfidentialMintBurnBurnData{NewDecryptableAvailableBalance: ctDecryptableBalance, BurnAmountAuditorCiphertextLo: ctCiphertextLo, BurnAmountAuditorCiphertextHi: ctCiphertextHi, EqualityProofInstructionOffset: 1, CiphertextValidityProofInstructionOffset: 2, RangeProofInstructionOffset: 3},
		ConfidentialMintBurnApplyPendingBurnData{},
	})
}

func TestConfidentialMintBurnTypedDecode(t *testing.T) {
	t.Parallel()
	built, err := NewConfidentialMintBurnUpdateDecryptableSupplyInstruction(
		ctMint, ctAuthority, nil, ctDecryptableBalance).ValidateAndBuild()
	if err != nil {
		t.Fatalf("ValidateAndBuild: %v", err)
	}
	sub := decodeBuiltConfidentialSubinstruction[ConfidentialMintBurnSubInstructionData, *ConfidentialMintBurnExtension](t, built)
	supply, ok := sub.(*ConfidentialMintBurnUpdateDecryptableSupplyData)
	if !ok {
		t.Fatalf("sub-instruction decoded to %T, want *ConfidentialMintBurnUpdateDecryptableSupplyData", sub)
	}
	if supply.NewDecryptableSupply != ctDecryptableBalance {
		t.Errorf("decoded supply = %x, want %x", supply.NewDecryptableSupply, ctDecryptableBalance)
	}
}

func TestConfidentialMintBurnDecodeRejectsMalformed(t *testing.T) {
	t.Parallel()
	testConfidentialDecodeRejectsMalformed(
		t, &ConfidentialMintBurnExtension{}, ConfidentialMintBurn_Mint, ConfidentialMintBurn_ApplyPendingBurn)
}

// The split proofs of Mint and Burn are verified in the order they are
// processed, so the sibling instructions they name must run 1, 2, 3.
func TestConfidentialMintBurnSequentialProofOffsets(t *testing.T) {
	t.Parallel()
	equality := func(offset int8) zkprogram.ProofLocation[*proofdata.CiphertextCommitmentEqualityProofData] {
		return zkprogram.ProofLocationInstructionOffset(offset, &proofdata.CiphertextCommitmentEqualityProofData{})
	}
	validity := func(offset int8) zkprogram.ProofLocation[*proofdata.BatchedGroupedCiphertext3HandlesValidityProofData] {
		return zkprogram.ProofLocationInstructionOffset(offset, &proofdata.BatchedGroupedCiphertext3HandlesValidityProofData{})
	}
	rangeProof := func(offset int8) zkprogram.ProofLocation[*proofdata.BatchedRangeProofU128Data] {
		return zkprogram.ProofLocationInstructionOffset(offset, &proofdata.BatchedRangeProofU128Data{})
	}

	_, err := NewConfidentialMintBurnMintInstructions(
		ctTokenAccount, ctMint, ctCiphertextLo, ctCiphertextHi, ctAuthority, nil,
		equality(2), validity(2), rangeProof(3), ctDecryptableBalance)
	wantProofOffsetError(t, err, "equality offset 2")

	_, err = NewConfidentialMintBurnBurnInstructions(
		ctTokenAccount, ctMint, ctDecryptableBalance, ctCiphertextLo, ctCiphertextHi,
		ctAuthority, nil, equality(1), validity(3), rangeProof(2))
	wantProofOffsetError(t, err, "offsets 1,3,2")

	// A context state proof takes no sibling slot, so the offsets of the
	// remaining proofs shift down to 1 and 2.
	instructions, err := NewConfidentialMintBurnMintInstructions(
		ctTokenAccount, ctMint, ctCiphertextLo, ctCiphertextHi, ctAuthority, nil,
		zkprogram.ProofLocationContextStateAccount[*proofdata.CiphertextCommitmentEqualityProofData](ctContextEquality),
		validity(1), rangeProof(2), ctDecryptableBalance)
	if err != nil {
		t.Fatalf("context equality + offsets 1,2: %v", err)
	}
	if len(instructions) != 3 {
		t.Errorf("built %d instructions, want 3", len(instructions))
	}
}

func TestConfidentialMintBurnRejectsUnsetProofLocation(t *testing.T) {
	t.Parallel()
	testConfidentialRejectsUnsetProofLocation(t,
		func(location zkprogram.ProofLocation[*proofdata.CiphertextCiphertextEqualityProofData]) error {
			_, err := NewConfidentialMintBurnRotateSupplyElGamalPubkeyInstructions(
				ctMint, ctAuthority, nil, cmbNewSupplyElGamalPubkey, location)
			return err
		})
}

func TestConfidentialMintBurnRawInstruction(t *testing.T) {
	t.Parallel()
	raw := NewConfidentialMintBurnInstruction(
		ConfidentialMintBurn_UpdateDecryptableSupply,
		ConfidentialMintBurnUpdateDecryptableSupplyData{NewDecryptableSupply: ctDecryptableBalance}.bytes(),
		*solana.Meta(ctMint).WRITE(),
		*solana.Meta(ctAuthority).SIGNER(),
	)
	typed := NewConfidentialMintBurnUpdateDecryptableSupplyInstruction(
		ctMint, ctAuthority, nil, ctDecryptableBalance)
	testConfidentialRawMatchesTyped(t, raw, typed)
}
