package confidential

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/zktest"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
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
		sender, aeKey := genTransferAccount(t)
		recipient := zktest.GenKeyPair(t)
		balanceCt, decryptable := encryptBalance(t, sender, aeKey, 1000)
		proofs, err := TransferSplitProofData(balanceCt, decryptable, 500,
			sender, aeKey, recipient.Pubkey, nil)
		if err != nil {
			t.Fatal(err)
		}
		for name, proof := range map[string]proofdata.ProofData{
			"equality": proofs.EqualityProofData,
			"validity": proofs.CiphertextValidityProofDataWithCiphertext.ProofData,
			"range":    proofs.RangeProofData,
		} {
			if err := proof.Verify(); err != nil {
				t.Fatalf("%s proof rejected: %v", name, err)
			}
		}
	})

	// Structural violations are rejected before any proof work.
	sender, aeKey := genTransferAccount(t)
	recipient := zktest.GenKeyPair(t)
	auditor := zktest.GenKeyPair(t)
	const currentBalance = uint64(1_000_000)
	balanceCt, decryptable := encryptBalance(t, sender, aeKey, currentBalance)
	if _, err := TransferSplitProofData(balanceCt, decryptable, currentBalance+1,
		sender, aeKey, recipient.Pubkey, &auditor.Pubkey); !errors.Is(err, ErrNotEnoughFunds) {
		t.Fatalf("transfer exceeding balance: got %v, want ErrNotEnoughFunds", err)
	}
	_, bigDecryptable := encryptBalance(t, sender, aeKey, 1<<50)
	if _, err := TransferSplitProofData(balanceCt, bigDecryptable, 1<<49,
		sender, aeKey, recipient.Pubkey, &auditor.Pubkey); !errors.Is(err, ErrIllegalAmountBitLength) {
		t.Fatalf("transfer exceeding 48-bit amount limit: got %v, want ErrIllegalAmountBitLength", err)
	}
}

// genTransferAccount returns a source keypair and AE key.
func genTransferAccount(t *testing.T) (*encryption.ElGamalKeypair, encryption.AeKey) {
	t.Helper()
	kp := zktest.GenKeyPair(t)
	aeKey, err := encryption.NewAeKey()
	if err != nil {
		t.Fatal(err)
	}
	return kp, aeKey
}

// encryptBalance encrypts balance under both the ElGamal pubkey and the AE key.
func encryptBalance(t *testing.T, kp *encryption.ElGamalKeypair, aeKey encryption.AeKey, balance uint64) (encryption.ElGamalCiphertext, encryption.AeCiphertext) {
	t.Helper()
	balanceCt, err := kp.Pubkey.Encrypt(balance)
	if err != nil {
		t.Fatal(err)
	}
	decryptable, err := aeKey.Encrypt(balance)
	if err != nil {
		t.Fatal(err)
	}
	return balanceCt, decryptable
}

func testTransferProofValidity(t *testing.T, currentBalance, transferAmount uint64) {
	remaining := currentBalance - transferAmount
	sender, aeKey := genTransferAccount(t)
	recipient := zktest.GenKeyPair(t)
	auditor := zktest.GenKeyPair(t)

	balanceCt, decryptable := encryptBalance(t, sender, aeKey, currentBalance)
	proofs, err := TransferSplitProofData(balanceCt, decryptable, transferAmount,
		sender, aeKey, recipient.Pubkey, &auditor.Pubkey)
	if err != nil {
		t.Fatal(err)
	}
	validity := proofs.CiphertextValidityProofDataWithCiphertext
	for name, proof := range map[string]proofdata.ProofData{
		"equality": proofs.EqualityProofData,
		"validity": validity.ProofData,
		"range":    proofs.RangeProofData,
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

	// The recipient and auditor recover the transfer amount from their
	// handles (lo + hi<<16).
	for name, kp := range map[string]*encryption.ElGamalKeypair{"recipient": recipient, "auditor": auditor} {
		index := 1
		if name == "auditor" {
			index = 2
		}
		loCt, err := validity.CiphertextLo.ToElGamalCiphertext(index)
		if err != nil {
			t.Fatal(err)
		}
		hiCt, err := validity.CiphertextHi.ToElGamalCiphertext(index)
		if err != nil {
			t.Fatal(err)
		}
		lo, err := kp.DecryptU32(loCt)
		if err != nil {
			t.Fatal(err)
		}
		hi, err := kp.DecryptU32(hiCt)
		if err != nil {
			t.Fatal(err)
		}
		if got := lo + hi<<TransferAmountLoBitLength; got != transferAmount {
			t.Fatalf("%s decrypts transfer amount %d, want %d", name, got, transferAmount)
		}
	}
}
