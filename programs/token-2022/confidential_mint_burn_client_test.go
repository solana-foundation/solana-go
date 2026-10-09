package token2022

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfidentialMintBurnClient(t *testing.T) {
	t.Parallel()
	forEachProofLocation(t, "RotateSupplyElGamalPubkey", testClientRotateSupplyElGamalPubkey)
	forEachProofLocation(t, "Mint", testClientMint)
	forEachProofLocation(t, "Burn", testClientBurn)
}

func testClientRotateSupplyElGamalPubkey(t *testing.T, inline bool) {
	supplyState, supplyKp, supplyAesKey := makeSupplyState(t, clientSupply, clientSupply+12_345)
	newSupply := genKeyPair(t)
	instructions, err := ConfidentialTransferRotateSupplyElGamalPubkey(
		ctMint, ctAuthority, nil, supplyKp, newSupply.Pubkey, supplyAesKey,
		orNil(inline, &ctContextSingle), NewSupplyAccountInfo(supplyState))
	require.NoError(t, err)
	rotate := clientData[*ConfidentialMintBurnRotateSupplyElGamalPubkeyData](t, instructions)
	if rotate.NewSupplyElGamalPubkey != newSupply.Pubkey {
		t.Error("new supply ElGamal pubkey is not the one given")
	}
	checkClientProofs(t, instructions, inline, ciphertextEqualityProofs, []int8{rotate.ProofInstructionOffset})
}

func testClientMint(t *testing.T, inline bool) {
	supplyState, supplyKp, supplyAesKey := makeSupplyState(t, clientSupply, clientSupply+12_345)
	destination, auditor := genKeyPair(t), genKeyPair(t)
	instructions, err := ConfidentialTransferMint(
		ctMint, ctAuthority, nil, ctDestination,
		orNil(inline, &ctContextEquality), orNil(inline, validityProofAccount), orNil(inline, &ctContextRange),
		clientAmount, supplyKp, destination.Pubkey, &auditor.Pubkey, supplyAesKey, NewSupplyAccountInfo(supplyState))
	require.NoError(t, err)
	mint := clientData[*ConfidentialMintBurnMintData](t, instructions)
	checkDecryptable(t, supplyAesKey, mint.NewDecryptableSupply, clientSupply+clientAmount)
	checkCiphertexts(t, inline, auditor, clientAmount,
		mint.MintAmountAuditorCiphertextLo, mint.MintAmountAuditorCiphertextHi)
	checkClientProofs(t, instructions, inline, splitProofs, []int8{
		mint.EqualityProofInstructionOffset,
		mint.CiphertextValidityProofInstructionOffset,
		mint.RangeProofInstructionOffset,
	})
}

func testClientBurn(t *testing.T, inline bool) {
	holder, holderKp, holderAesKey := makeAccountState(t, 0, 0, clientBalance, 0)
	supply, auditor := genKeyPair(t), genKeyPair(t)
	instructions, err := ConfidentialTransferBurn(
		ctMint, ctAuthority, nil, ctTokenAccount,
		orNil(inline, &ctContextEquality), orNil(inline, validityProofAccount), orNil(inline, &ctContextRange),
		clientAmount, holderKp, supply.Pubkey, &auditor.Pubkey, holderAesKey, NewBurnAccountInfo(holder))
	require.NoError(t, err)
	burn := clientData[*ConfidentialMintBurnBurnData](t, instructions)
	checkDecryptable(t, holderAesKey, burn.NewDecryptableAvailableBalance,
		clientBalance-clientAmount)
	checkCiphertexts(t, inline, auditor, clientAmount,
		burn.BurnAmountAuditorCiphertextLo, burn.BurnAmountAuditorCiphertextHi)
	checkClientProofs(t, instructions, inline, splitProofs, []int8{
		burn.EqualityProofInstructionOffset,
		burn.CiphertextValidityProofInstructionOffset,
		burn.RangeProofInstructionOffset,
	})
}

// TestConfidentialMintBurnClientMixedProofLocations checks the limitation
// that an operation cannot take one proof from a context state account and
// inline another.
func TestConfidentialMintBurnClientMixedProofLocations(t *testing.T) {
	t.Parallel()
	state, supplyKp, aesKey := makeSupplyState(t, 1000, 1000)
	_, err := ConfidentialTransferMint(
		ctMint, ctAuthority, nil, ctDestination, &ctContextEquality, nil, nil,
		10, supplyKp, genKeyPair(t).Pubkey, nil, aesKey, NewSupplyAccountInfo(state))
	wantProofOffsetError(t, err, "ConfidentialTransferMint")
}
