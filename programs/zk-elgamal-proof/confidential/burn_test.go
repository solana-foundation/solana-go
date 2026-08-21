package confidential

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/zktest"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
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
	sourcekp, sourceAesKey := genTransferAccount(t)
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
	remaining := currentBalance - burnAmount
	source, aeKey := genTransferAccount(t)
	supply := zktest.GenKeyPair(t)
	auditor := zktest.GenKeyPair(t)

	balanceCt, decryptable := encryptBalance(t, source, aeKey, currentBalance)
	proofs, err := BurnSplitProofData(balanceCt, decryptable, burnAmount,
		source, aeKey, supply.Pubkey, &auditor.Pubkey)
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

	// The source can decrypt the new balance, derived homomorphically the way
	// the token program recomputes it on-chain.
	sourceLo, err := validity.CiphertextLo.ToElGamalCiphertext(0)
	if err != nil {
		t.Fatal(err)
	}
	sourceHi, err := validity.CiphertextHi.ToElGamalCiphertext(0)
	if err != nil {
		t.Fatal(err)
	}
	combined, err := encryption.CombineLoHiCiphertexts(sourceLo, sourceHi, BurnAmountLoBitLength)
	if err != nil {
		t.Fatal(err)
	}
	newBalance, err := encryption.SubtractCiphertexts(balanceCt, combined)
	if err != nil {
		t.Fatal(err)
	}
	got, err := source.DecryptU32(newBalance)
	if err != nil {
		t.Fatal(err)
	}
	if got != remaining {
		t.Fatalf("new balance decrypts to %d, want %d", got, remaining)
	}

	// The supply and auditor recover the burn amount from their handles
	// (lo + hi<<16).
	for name, kp := range map[string]*encryption.ElGamalKeypair{"supply": supply, "auditor": auditor} {
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
		if got := lo + hi<<BurnAmountLoBitLength; got != burnAmount {
			t.Fatalf("%s decrypts burn amount %d, want %d", name, got, burnAmount)
		}
	}
}
