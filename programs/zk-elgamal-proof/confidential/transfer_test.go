package confidential

import (
	"errors"
	"fmt"
	"testing"

	zk "github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/internal/zktest"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

func TestTransferProofData(t *testing.T) {
	// Test vectors from token-2022's confidential proof-tests.
	for _, tt := range []struct {
		balance, amount uint64
	}{
		{0, 0},
		{1, 0},
		{1, 1},
		{65535, 65535},                     // 2^16 - 1
		{65536, 65536},                     // 2^16
		{281474976710655, 281474976710655}, // 2^48 - 1
	} {
		t.Run(fmt.Sprintf("balance=%d,amount=%d", tt.balance, tt.amount), func(t *testing.T) {
			testTransferProofValidity(t, tt.balance, tt.amount)
		})
	}

	// A mint without an auditor: the proofs still verify.
	t.Run("no auditor", func(t *testing.T) {
		sender, aesKey := generateSourceAccount(t)
		recipient := zktest.GenKeyPair(t)
		balanceCt, decryptable := encryptBalance(t, sender, aesKey, 1000)
		proofs, err := TransferSplitProofData(balanceCt, decryptable, 500,
			sender, aesKey, recipient.Pubkey, nil)
		if err != nil {
			t.Fatal(err)
		}
		verifyAll(t, map[string]proofdata.ProofData{
			"equality": proofs.EqualityProofData,
			"validity": proofs.CiphertextValidityProofDataWithCiphertext.ProofData,
			"range":    proofs.RangeProofData,
		})
	})

	// Structural violations are rejected before any proof work.
	sender, aesKey := generateSourceAccount(t)
	recipient := zktest.GenKeyPair(t)
	auditor := zktest.GenKeyPair(t)
	const currentBalance = uint64(1_000_000)
	balanceCt, decryptable := encryptBalance(t, sender, aesKey, currentBalance)
	if _, err := TransferSplitProofData(balanceCt, decryptable, currentBalance+1,
		sender, aesKey, recipient.Pubkey, &auditor.Pubkey); !errors.Is(err, ErrNotEnoughFunds) {
		t.Fatalf("transfer exceeding balance: got %v, want ErrNotEnoughFunds", err)
	}
	_, bigDecryptable := encryptBalance(t, sender, aesKey, 1<<50)
	if _, err := TransferSplitProofData(balanceCt, bigDecryptable, 1<<49,
		sender, aesKey, recipient.Pubkey, &auditor.Pubkey); !errors.Is(err, ErrIllegalAmountBitLength) {
		t.Fatalf("transfer exceeding 48-bit amount limit: got %v, want ErrIllegalAmountBitLength", err)
	}

	wrongKey, err := encryption.NewAeKey()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := TransferSplitProofData(balanceCt, decryptable, 1,
		sender, wrongKey, recipient.Pubkey, &auditor.Pubkey); !errors.Is(err, zk.ErrDecryption) {
		t.Fatalf("transfer with undecryptable balance: got %v, want zk.ErrDecryption", err)
	}
}

func testTransferProofValidity(t *testing.T, currentBalance, transferAmount uint64) {
	sender, aesKey := generateSourceAccount(t)
	recipient := zktest.GenKeyPair(t)
	auditor := zktest.GenKeyPair(t)

	balanceCt, decryptable := encryptBalance(t, sender, aesKey, currentBalance)
	proofs, err := TransferSplitProofData(balanceCt, decryptable, transferAmount,
		sender, aesKey, recipient.Pubkey, &auditor.Pubkey)
	if err != nil {
		t.Fatal(err)
	}
	validity := proofs.CiphertextValidityProofDataWithCiphertext
	verifyAll(t, map[string]proofdata.ProofData{
		"equality": proofs.EqualityProofData,
		"validity": validity.ProofData,
		"range":    proofs.RangeProofData,
	})

	// The sender can decrypt the new balance, derived homomorphically the way
	// the token program recomputes it on-chain.
	newBalance := applyToBalance(t, balanceCt, validity, 0, encryption.SubtractCiphertexts)
	decryptEquals(t, sender, newBalance, currentBalance-transferAmount, "new balance")

	// The recipient and auditor recover the transfer amount from their
	// handles (lo + hi<<16) in the proof context's grouped ciphertexts.
	validityContext := validity.ProofData.Context
	for _, holder := range []struct {
		name  string
		kp    *encryption.ElGamalKeypair
		index int
	}{
		{"recipient", recipient, 1},
		{"auditor", auditor, 2},
	} {
		got := decryptHandle(t, holder.kp, validityContext.GroupedCiphertextLo, validityContext.GroupedCiphertextHi, holder.index, AmountLoBitLength)
		if got != transferAmount {
			t.Fatalf("%s decrypts transfer amount %d, want %d", holder.name, got, transferAmount)
		}
	}

	// The extracted auditor ciphertexts decrypt to the transfer amount.
	if got := decryptLoHiPair(t, auditor, validity.CiphertextLo, validity.CiphertextHi, AmountLoBitLength); got != transferAmount {
		t.Fatalf("auditor decrypts extracted ciphertexts to %d, want %d", got, transferAmount)
	}
}
