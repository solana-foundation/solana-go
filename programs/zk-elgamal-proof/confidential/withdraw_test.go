package confidential

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/zktest"
)

func TestWithdrawProofData(t *testing.T) {
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
			testWithdrawProofValidity(t, tt.balance, tt.amount)
		})
	}

	// Structural violations are rejected before any proof work.
	kp := zktest.GenKeyPair(t)
	const currentBalance = uint64(500_000)
	balanceCt, err := kp.Pubkey.Encrypt(currentBalance)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewWithdrawProofData(balanceCt, currentBalance, currentBalance+1, kp); !errors.Is(err, ErrNotEnoughFunds) {
		t.Fatalf("withdrawal exceeding balance: got %v, want ErrNotEnoughFunds", err)
	}
}

func testWithdrawProofValidity(t *testing.T, currentBalance, withdrawAmount uint64) {
	remaining := currentBalance - withdrawAmount
	kp := zktest.GenKeyPair(t)
	balanceCt, err := kp.Pubkey.Encrypt(currentBalance)
	if err != nil {
		t.Fatal(err)
	}
	proofs, err := NewWithdrawProofData(balanceCt, currentBalance, withdrawAmount, kp)
	if err != nil {
		t.Fatal(err)
	}
	if err := proofs.EqualityProofData.Verify(); err != nil {
		t.Fatalf("equality proof rejected: %v", err)
	}
	if err := proofs.RangeProofData.Verify(); err != nil {
		t.Fatalf("range proof rejected: %v", err)
	}

	// The account holder can decrypt the new balance, derived the way the
	// token program recomputes it on-chain.
	newBalance, err := balanceCt.SubtractAmount(withdrawAmount)
	if err != nil {
		t.Fatal(err)
	}
	got, err := kp.DecryptU32(newBalance)
	if err != nil {
		t.Fatal(err)
	}
	if got != remaining {
		t.Fatalf("new balance decrypts to %d, want %d", got, remaining)
	}
}
