package confidential

import (
	"errors"
	"fmt"
	"testing"

	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/internal/zktest"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

func TestBurnProofData(t *testing.T) {
	// Test vectors from token-2022's confidential proof-tests.
	for _, tt := range []struct {
		balance, amount uint64
	}{
		{0, 0},
		{77, 55},
		{65535, 65535},                     // 2^16 - 1
		{65536, 65536},                     // 2^16
		{281474976710655, 281474976710655}, // 2^48 - 1
	} {
		t.Run(fmt.Sprintf("balance=%d,amount=%d", tt.balance, tt.amount), func(t *testing.T) {
			testBurnProofValidity(t, tt.balance, tt.amount)
		})
	}
	sourcekp, sourceAesKey := generateSourceAccount(t)
	supplykp := zktest.GenKeyPair(t)
	auditorkp := zktest.GenKeyPair(t)
	const currentBalance = uint64(2_000_000)
	balanceEgCt, balanceAesCt := encryptBalance(t, sourcekp, sourceAesKey, currentBalance)
	if _, err := BurnSplitProofData(balanceEgCt, balanceAesCt, currentBalance+1,
		sourcekp, sourceAesKey, supplykp.Pubkey, &auditorkp.Pubkey); !errors.Is(err, ErrNotEnoughFunds) {
		t.Fatalf("burn exceeding balance: got %v, want ErrNotEnoughFunds", err)
	}
	_, bigAesCt := encryptBalance(t, sourcekp, sourceAesKey, 1<<50)
	if _, err := BurnSplitProofData(balanceEgCt, bigAesCt, 1<<49,
		sourcekp, sourceAesKey, supplykp.Pubkey, &auditorkp.Pubkey); !errors.Is(err, ErrIllegalAmountBitLength) {
		t.Fatalf("burn exceeding 48-bit amount limit: got %v, want ErrIllegalAmountBitLength", err)
	}
}

func testBurnProofValidity(t *testing.T, currentBalance, burnAmount uint64) {
	source, aesKey := generateSourceAccount(t)
	supply := zktest.GenKeyPair(t)
	auditor := zktest.GenKeyPair(t)

	balanceCt, decryptable := encryptBalance(t, source, aesKey, currentBalance)
	proofs, err := BurnSplitProofData(balanceCt, decryptable, burnAmount,
		source, aesKey, supply.Pubkey, &auditor.Pubkey)
	if err != nil {
		t.Fatal(err)
	}
	validity := proofs.CiphertextValidityProofDataWithCiphertext
	verifyAll(t, map[string]proofdata.ProofData{
		"equality": proofs.EqualityProofData,
		"validity": validity.ProofData,
		"range":    proofs.RangeProofData,
	})

	// The source can decrypt the new balance, derived homomorphically the way
	// the token program recomputes it on-chain.
	newBalance := applyToBalance(t, balanceCt, validity, 0, encryption.SubtractCiphertexts)
	decryptEquals(t, source, newBalance, currentBalance-burnAmount, "new balance")

	// The supply and auditor recover the burn amount from their handles
	// (lo + hi<<16) in the proof context's grouped ciphertexts.
	validityContext := validity.ProofData.Context
	for _, holder := range []struct {
		name  string
		kp    *encryption.ElGamalKeypair
		index int
	}{
		{"supply", supply, 1},
		{"auditor", auditor, 2},
	} {
		got := decryptHandle(t, holder.kp, validityContext.GroupedCiphertextLo, validityContext.GroupedCiphertextHi, holder.index, AmountLoBitLength)
		if got != burnAmount {
			t.Fatalf("%s decrypts burn amount %d, want %d", holder.name, got, burnAmount)
		}
	}

	// The extracted auditor ciphertexts decrypt to the burn amount.
	if got := decryptLoHiPair(t, auditor, validity.CiphertextLo, validity.CiphertextHi, AmountLoBitLength); got != burnAmount {
		t.Fatalf("auditor decrypts extracted ciphertexts to %d, want %d", got, burnAmount)
	}
}
