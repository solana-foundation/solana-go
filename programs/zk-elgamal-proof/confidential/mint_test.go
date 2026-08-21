package confidential

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/zktest"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
)

func TestMintProofData(t *testing.T) {
	// Test vectors from token-2022's confidential proof-tests.
	for _, tt := range []struct {
		amount, supply uint64
	}{
		{0, 0},
		{1, 0},
		{65535, 0},
		{65536, 0},
		{281474976710655, 0},

		{0, 65535},
		{1, 65535},
		{65535, 65535},
		{65536, 65535},
		{281474976710655, 65535},

		{0, 281474976710655},
		{1, 281474976710655},
		{65535, 281474976710655},
		{65536, 281474976710655},
		{281474976710655, 281474976710655},
	} {
		t.Run(fmt.Sprintf("amount=%d,supply=%d", tt.amount, tt.supply), func(t *testing.T) {
			testMintProofValidity(t, tt.amount, tt.supply)
		})
	}

	// Structural violations are rejected before any proof work.
	supply := zktest.GenKeyPair(t)
	destination := zktest.GenKeyPair(t)
	auditor := zktest.GenKeyPair(t)
	const currentSupply = uint64(3_000_000)
	supplyCt, err := supply.Pubkey.Encrypt(currentSupply)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := MintSplitProofData(supplyCt, 1<<48, currentSupply,
		supply, destination.Pubkey, &auditor.Pubkey); !errors.Is(err, ErrIllegalAmountBitLength) {
		t.Fatalf("mint exceeding 48-bit amount limit: got %v, want ErrIllegalAmountBitLength", err)
	}
	if _, err := MintSplitProofData(supplyCt, 1, math.MaxUint64,
		supply, destination.Pubkey, &auditor.Pubkey); !errors.Is(err, ErrIllegalAmountBitLength) {
		t.Fatalf("mint overflowing the supply: got %v, want ErrIllegalAmountBitLength", err)
	}
}

func testMintProofValidity(t *testing.T, mintAmount, currentSupply uint64) {
	newSupply := currentSupply + mintAmount
	supply := zktest.GenKeyPair(t)
	destination := zktest.GenKeyPair(t)
	auditor := zktest.GenKeyPair(t)

	supplyCt, err := supply.Pubkey.Encrypt(currentSupply)
	if err != nil {
		t.Fatal(err)
	}
	proofs, err := MintSplitProofData(supplyCt, mintAmount, currentSupply,
		supply, destination.Pubkey, &auditor.Pubkey)
	if err != nil {
		t.Fatal(err)
	}
	validity := proofs.CiphertextValidityProofDataWithCiphertext
	for name, proof := range map[string]proofdata.ProofData{
		"equality": proofs.SupplyEqualityProofData,
		"validity": validity.ProofData,
		"range":    proofs.RangeProofData,
	} {
		if err := proof.Verify(); err != nil {
			t.Fatalf("%s proof rejected: %v", name, err)
		}
	}

	// The supply keypair can decrypt the new supply, derived homomorphically
	// the way the token program recomputes it on-chain (recovery is only
	// feasible for values that fit in 32 bits).
	if newSupply <= math.MaxUint32 {
		supplyLo, err := validity.CiphertextLo.ToElGamalCiphertext(1)
		if err != nil {
			t.Fatal(err)
		}
		supplyHi, err := validity.CiphertextHi.ToElGamalCiphertext(1)
		if err != nil {
			t.Fatal(err)
		}
		combined, err := encryption.CombineLoHiCiphertexts(supplyLo, supplyHi, MintAmountLoBits)
		if err != nil {
			t.Fatal(err)
		}
		newSupplyCt, err := encryption.AddCiphertexts(supplyCt, combined)
		if err != nil {
			t.Fatal(err)
		}
		got, err := supply.DecryptU32(newSupplyCt)
		if err != nil {
			t.Fatal(err)
		}
		if got != newSupply {
			t.Fatalf("new supply decrypts to %d, want %d", got, newSupply)
		}
	}

	// The destination and auditor recover the mint amount from their handles
	// (lo + hi<<16).
	for name, kp := range map[string]*encryption.ElGamalKeypair{"destination": destination, "auditor": auditor} {
		index := 0
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
		if got := lo + hi<<MintAmountLoBits; got != mintAmount {
			t.Fatalf("%s decrypts mint amount %d, want %d", name, got, mintAmount)
		}
	}
}
