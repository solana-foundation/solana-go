package confidential

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/zktest"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
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
	sender, aeKey := genTransferAccount(t)
	recipient := zktest.GenKeyPair(t)
	auditor := zktest.GenKeyPair(t)
	withdrawAuthority := zktest.GenKeyPair(t)
	const (
		currentBalance = uint64(1_000_000)
		feeRate        = uint16(250)
		maximumFee     = uint64(10_000)
	)
	balanceCt, decryptable := encryptBalance(t, sender, aeKey, currentBalance)
	if _, err := TransferWithFeeSplitProofData(
		balanceCt, decryptable, currentBalance+1, sender, aeKey,
		recipient.Pubkey, &auditor.Pubkey, withdrawAuthority.Pubkey,
		feeRate, maximumFee); !errors.Is(err, ErrNotEnoughFunds) {
		t.Fatalf("transfer exceeding balance: got %v, want ErrNotEnoughFunds", err)
	}
	_, bigDecryptable := encryptBalance(t, sender, aeKey, 1<<50)
	if _, err := TransferWithFeeSplitProofData(
		balanceCt, bigDecryptable, 1<<49, sender, aeKey,
		recipient.Pubkey, &auditor.Pubkey, withdrawAuthority.Pubkey,
		feeRate, maximumFee); !errors.Is(err, ErrIllegalAmountBitLength) {
		t.Fatalf("transfer exceeding 48-bit amount limit: got %v, want ErrIllegalAmountBitLength", err)
	}
}

func testTransferWithFeeProofValidity(t *testing.T, currentBalance, transferAmount uint64, feeRate uint16, maximumFee uint64) {
	remaining := currentBalance - transferAmount
	feeAmount := min((transferAmount*uint64(feeRate)+9_999)/10_000, maximumFee)

	sender, aeKey := genTransferAccount(t)
	recipient := zktest.GenKeyPair(t)
	auditor := zktest.GenKeyPair(t)
	withdrawAuthority := zktest.GenKeyPair(t)

	balanceCt, decryptable := encryptBalance(t, sender, aeKey, currentBalance)
	proofs, err := TransferWithFeeSplitProofData(
		balanceCt, decryptable, transferAmount, sender, aeKey,
		recipient.Pubkey, &auditor.Pubkey, withdrawAuthority.Pubkey,
		feeRate, maximumFee)
	if err != nil {
		t.Fatal(err)
	}
	validity := proofs.TransferAmountCiphertextValidityProofDataWithCiphertext
	for name, proof := range map[string]proofdata.ProofData{
		"equality":     proofs.RemainingBalanceProofData,
		"validity":     validity.ProofData,
		"percentage":   proofs.PercentageWithCapProofData,
		"fee validity": proofs.FeeCiphertextValidityProofData,
		"range":        proofs.RangeProofData,
	} {
		if err := proof.Verify(); err != nil {
			t.Fatalf("%s proof rejected: %v", name, err)
		}
	}

	// The sender can decrypt the new balance, derived homomorphically the way
	// the token program recomputes it on-chain.
	senderLo, err := validity.CiphertextLo.ToElGamalCiphertext(0)
	if err != nil {
		t.Fatal(err)
	}
	senderHi, err := validity.CiphertextHi.ToElGamalCiphertext(0)
	if err != nil {
		t.Fatal(err)
	}
	combined, err := encryption.CombineLoHiCiphertexts(senderLo, senderHi, TransferAmountLoBitLength)
	if err != nil {
		t.Fatal(err)
	}
	newBalance, err := encryption.SubtractCiphertexts(balanceCt, combined)
	if err != nil {
		t.Fatal(err)
	}
	got, err := sender.DecryptU32(newBalance)
	if err != nil {
		t.Fatal(err)
	}
	if got != remaining {
		t.Fatalf("new balance decrypts to %d, want %d", got, remaining)
	}

	// The recipient recovers the gross transfer amount from its handles.
	loCt, err := validity.CiphertextLo.ToElGamalCiphertext(1)
	if err != nil {
		t.Fatal(err)
	}
	hiCt, err := validity.CiphertextHi.ToElGamalCiphertext(1)
	if err != nil {
		t.Fatal(err)
	}
	lo, err := recipient.DecryptU32(loCt)
	if err != nil {
		t.Fatal(err)
	}
	hi, err := recipient.DecryptU32(hiCt)
	if err != nil {
		t.Fatal(err)
	}
	if got := lo + hi<<TransferAmountLoBitLength; got != transferAmount {
		t.Fatalf("recipient decrypts transfer amount %d, want %d", got, transferAmount)
	}

	// The withdraw withheld authority recovers the fee from the fee validity
	// proof context's grouped ciphertexts.
	feeContext := proofs.FeeCiphertextValidityProofData.Context
	feeLoCt, err := feeContext.GroupedCiphertextLo.ToElGamalCiphertext(1)
	if err != nil {
		t.Fatal(err)
	}
	feeHiCt, err := feeContext.GroupedCiphertextHi.ToElGamalCiphertext(1)
	if err != nil {
		t.Fatal(err)
	}
	feeLo, err := withdrawAuthority.DecryptU32(feeLoCt)
	if err != nil {
		t.Fatal(err)
	}
	feeHi, err := withdrawAuthority.DecryptU32(feeHiCt)
	if err != nil {
		t.Fatal(err)
	}
	if got := feeLo + feeHi<<FeeAmountLoBitLength; got != feeAmount {
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
		{100, 100, 1, 0},     // exact 1%
		{101, 100, 2, 9_900}, // rounds up
		{1, 1, 1, 9_999},     // minimum nonzero fee
		{MaxTransferAmount, 10_000, MaxTransferAmount, 0}, // 100% of max amount
	} {
		fee, delta := calculateFee(tt.amount, tt.rate)
		if fee != tt.fee || delta != tt.delta {
			t.Errorf("calculateFee(%d, %d) = (%d, %d), want (%d, %d)",
				tt.amount, tt.rate, fee, delta, tt.fee, tt.delta)
		}
	}
}
