package token2022

import (
	"errors"
	"math"
	"testing"

	"github.com/gagliardetto/solana-go/programs/token-2022/zkencryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/confidential"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/stretchr/testify/require"
)

func TestSupplyAccountInfo(t *testing.T) {
	const supply, mintAmount = uint64(1_000_000), uint64(4_321)
	state, kp, aesKey := makeSupplyState(t, supply, supply+70_000)
	info := NewSupplyAccountInfo(state)

	if got := info.SupplyElGamalPubkey(); got != kp.Pubkey {
		t.Fatalf("SupplyElGamalPubkey() = %x, want %x", got, kp.Pubkey)
	}
	if got, err := info.DecryptedCurrentSupply(aesKey, kp); err != nil || got != supply {
		t.Fatalf("DecryptedCurrentSupply() = (%d, %v), want (%d, nil)", got, err, supply)
	}

	newKp := genKeyPair(t)
	rotate, err := info.GenerateRotateSupplyElGamalPubkeyProof(kp, newKp.Pubkey, aesKey)
	require.NoError(t, err)
	require.NoError(t, rotate.Verify())
	if rotate.Context.FirstPubkey != kp.Pubkey || rotate.Context.SecondPubkey != newKp.Pubkey ||
		rotate.Context.FirstCiphertext != encryption.ElGamalCiphertext(state.ConfidentialSupply) {
		t.Fatal("rotate proof is not over the current supply and pubkeys")
	}
	if got, err := newKp.DecryptU32(rotate.Context.SecondCiphertext); err != nil || got != supply {
		t.Fatalf("rotated supply ciphertext decrypts to (%d, %v), want (%d, nil)", got, err, supply)
	}

	mint, err := info.GenerateSplitMintProofData(mintAmount, kp, aesKey, genKeyPair(t).Pubkey, &genKeyPair(t).Pubkey)
	require.NoError(t, err)
	verifyAll(t, map[string]proofdata.ProofData{
		"equality": mint.SupplyEqualityProofData,
		"validity": mint.CiphertextValidityProofDataWithCiphertext.ProofData,
		"range":    mint.RangeProofData,
	})

	decryptable, err := info.NewDecryptableSupply(mintAmount, kp, aesKey)
	require.NoError(t, err)
	checkDecryptable(t, aesKey, decryptable, supply+mintAmount)
	if _, err := info.NewDecryptableSupply(math.MaxUint64-supply+1, kp, aesKey); !errors.Is(err, confidential.ErrBalanceOverflow) {
		t.Fatalf("NewDecryptableSupply() overflowing supply: got %v, want ErrBalanceOverflow", err)
	}
}

func TestBurnAccountInfo(t *testing.T) {
	const available, amount = uint64(1000), uint64(400)
	supply, auditor := genKeyPair(t), genKeyPair(t)
	state, source, aesKey := makeAccountState(t, 0, 0, available, 0)
	info := NewBurnAccountInfo(state)

	burn, err := info.GenerateSplitBurnProofData(amount, source, aesKey, supply.Pubkey, &auditor.Pubkey)
	require.NoError(t, err)
	verifyAll(t, map[string]proofdata.ProofData{
		"equality": burn.EqualityProofData,
		"validity": burn.CiphertextValidityProofDataWithCiphertext.ProofData,
		"range":    burn.RangeProofData,
	})

	decryptable, err := info.NewDecryptableBalance(amount, aesKey)
	require.NoError(t, err)
	checkDecryptable(t, aesKey, decryptable, available-amount)

	// A burn exceeding the balance is rejected.
	if _, err := info.GenerateSplitBurnProofData(available+1, source, aesKey, supply.Pubkey, &auditor.Pubkey); !errors.Is(err, confidential.ErrNotEnoughFunds) {
		t.Fatalf("GenerateSplitBurnProofData() exceeding balance: got %v, want ErrNotEnoughFunds", err)
	}
	if _, err := info.NewDecryptableBalance(available+1, aesKey); !errors.Is(err, confidential.ErrNotEnoughFunds) {
		t.Fatalf("NewDecryptableBalance() exceeding balance: got %v, want ErrNotEnoughFunds", err)
	}
}

// makeSupplyState builds a confidential mint-burn mint state holding supply
// as its confidential supply and decryptableSupply as its decryptable supply
// cache, returning it with the fresh keys it is encrypted under.
func makeSupplyState(
	t *testing.T, supply, decryptableSupply uint64,
) (*ConfidentialMintBurnState, *encryption.ElGamalKeypair, zkencryption.AeKey) {
	t.Helper()
	kp := genKeyPair(t)
	aesKey, err := encryption.NewAeKey()
	require.NoError(t, err)
	supplyCiphertext, err := kp.Pubkey.Encrypt(supply)
	require.NoError(t, err)
	decryptable, err := encryption.AeEncrypt(aesKey, decryptableSupply)
	require.NoError(t, err)
	return &ConfidentialMintBurnState{
		ConfidentialSupply:  supplyCiphertext,
		DecryptableSupply:   decryptable,
		SupplyElGamalPubkey: kp.Pubkey,
	}, kp, aesKey
}
