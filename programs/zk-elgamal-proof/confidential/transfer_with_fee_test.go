package confidential

import (
	"errors"
	"fmt"
	"testing"

	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/internal/zktest"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

func TestTransferWithFeeProofData(t *testing.T) {
	// Test vectors from token-2022's confidential proof-tests.
	for _, tt := range []struct {
		balance, amount uint64
		feeRate         uint16
		maximumFee      uint64
	}{
		{0, 0, 0, 0},
		{0, 0, 0, 1},
		{0, 0, 1, 0},
		{0, 0, 1, 1},
		{1, 0, 0, 0},
		{1, 1, 0, 0},

		{100, 100, 5, 10},
		{100, 100, 5, 1},

		{65535, 65535, 5, 10}, // 2^16 - 1
		{65535, 65535, 5, 1},

		{65536, 65536, 5, 10}, // 2^16
		{65536, 65536, 5, 1},

		{281474976710655, 281474976710655, 5, 10}, // 2^48 - 1
		{281474976710655, 281474976710655, 5, 1},
	} {
		t.Run(fmt.Sprintf("balance=%d,amount=%d,rate=%d,max=%d",
			tt.balance, tt.amount, tt.feeRate, tt.maximumFee), func(t *testing.T) {
			testTransferWithFeeProofValidity(t, tt.balance, tt.amount, tt.feeRate, tt.maximumFee)
		})
	}

	// Structural violations are rejected before any proof work.
	sender, aesKey := generateSourceAccount(t)
	recipient := zktest.GenKeyPair(t)
	auditor := zktest.GenKeyPair(t)
	withdrawAuthority := zktest.GenKeyPair(t)
	const (
		currentBalance = uint64(1_000_000)
		feeRate        = uint16(250)
		maximumFee     = uint64(10_000)
	)
	balanceCt, decryptable := encryptBalance(t, sender, aesKey, currentBalance)
	if _, err := TransferWithFeeSplitProofData(
		balanceCt, decryptable, currentBalance+1, sender, aesKey,
		recipient.Pubkey, &auditor.Pubkey, withdrawAuthority.Pubkey,
		feeRate, maximumFee); !errors.Is(err, ErrNotEnoughFunds) {
		t.Fatalf("transfer exceeding balance: got %v, want ErrNotEnoughFunds", err)
	}
	_, bigDecryptable := encryptBalance(t, sender, aesKey, 1<<50)
	if _, err := TransferWithFeeSplitProofData(
		balanceCt, bigDecryptable, 1<<49, sender, aesKey,
		recipient.Pubkey, &auditor.Pubkey, withdrawAuthority.Pubkey,
		feeRate, maximumFee); !errors.Is(err, ErrIllegalAmountBitLength) {
		t.Fatalf("transfer exceeding 48-bit amount limit: got %v, want ErrIllegalAmountBitLength", err)
	}
}

func testTransferWithFeeProofValidity(t *testing.T, currentBalance, transferAmount uint64, feeRate uint16, maximumFee uint64) {
	remaining := currentBalance - transferAmount
	feeAmount := min((transferAmount*uint64(feeRate)+9_999)/10_000, maximumFee)

	sender, aesKey := generateSourceAccount(t)
	recipient := zktest.GenKeyPair(t)
	auditor := zktest.GenKeyPair(t)
	withdrawAuthority := zktest.GenKeyPair(t)

	balanceCt, decryptable := encryptBalance(t, sender, aesKey, currentBalance)
	proofs, err := TransferWithFeeSplitProofData(
		balanceCt, decryptable, transferAmount, sender, aesKey,
		recipient.Pubkey, &auditor.Pubkey, withdrawAuthority.Pubkey,
		feeRate, maximumFee)
	if err != nil {
		t.Fatal(err)
	}
	validity := proofs.TransferAmountCiphertextValidityProofDataWithCiphertext
	verifyAll(t, map[string]proofdata.ProofData{
		"equality":     proofs.EqualityProofData,
		"validity":     validity.ProofData,
		"percentage":   proofs.PercentageWithCapProofData,
		"fee validity": proofs.FeeCiphertextValidityProofData,
		"range":        proofs.RangeProofData,
	})

	// The sender can decrypt the new balance, derived homomorphically the way
	// the token program recomputes it on-chain.
	newBalance := applyToBalance(t, balanceCt, validity, 0, encryption.SubtractCiphertexts)
	decryptEquals(t, sender, newBalance, remaining, "new balance")

	// The recipient recovers the gross transfer amount from its handles.
	validityContext := validity.ProofData.Context
	if got := decryptHandle(t, recipient, validityContext.GroupedCiphertextLo, validityContext.GroupedCiphertextHi, 1, AmountLoBitLength); got != transferAmount {
		t.Fatalf("recipient decrypts transfer amount %d, want %d", got, transferAmount)
	}

	// The extracted auditor ciphertexts decrypt to the transfer amount.
	if got := decryptLoHiPair(t, auditor, validity.CiphertextLo, validity.CiphertextHi, AmountLoBitLength); got != transferAmount {
		t.Fatalf("auditor decrypts extracted ciphertexts to %d, want %d", got, transferAmount)
	}

	// The withdraw withheld authority recovers the fee from the fee validity
	// proof context's grouped ciphertexts.
	feeContext := proofs.FeeCiphertextValidityProofData.Context
	if got := decryptHandle(t, withdrawAuthority, feeContext.GroupedCiphertextLo, feeContext.GroupedCiphertextHi, 1, FeeAmountLoBitLength); got != feeAmount {
		t.Fatalf("authority decrypts fee %d, want %d", got, feeAmount)
	}
}

func TestCalculateFee(t *testing.T) {
	for _, tt := range []struct {
		amount     uint64
		rate       uint16
		fee, delta uint64
	}{
		{0, 0, 0, 0},
		{100, 100, 1, 0},                  // exact 1%
		{101, 100, 2, 9_900},              // rounds up
		{1, 1, 1, 9_999},                  // minimum nonzero fee
		{MaxAmount, 10_000, MaxAmount, 0}, // 100% of max amount
	} {
		fee, delta := calculateFee(tt.amount, tt.rate)
		if fee != tt.fee || delta != tt.delta {
			t.Errorf("calculateFee(%d, %d) = (%d, %d), want (%d, %d)",
				tt.amount, tt.rate, fee, delta, tt.fee, tt.delta)
		}
	}
}
