package encryption_test

import (
	"testing"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/zktest"
)

func TestGroupedElGamalEncrypt3RoundTrip(t *testing.T) {
	amount := zktest.GenAmount(t, 1<<32-1)
	keypairs := [3]*encryption.ElGamalKeypair{zktest.GenKeyPair(t), zktest.GenKeyPair(t), zktest.GenKeyPair(t)}
	pubkeys := [3]encryption.ElGamalPubkey{keypairs[0].Pubkey, keypairs[1].Pubkey, keypairs[2].Pubkey}

	opening, err := encryption.NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	grouped, err := encryption.GroupedElGamalEncrypt3(pubkeys, amount, opening)
	if err != nil {
		t.Fatal(err)
	}

	// Every key holder recovers the amount from its own handle.
	for i, kp := range keypairs {
		ct, err := grouped.ToElGamalCiphertext(i)
		if err != nil {
			t.Fatal(err)
		}
		got, err := kp.DecryptU32(ct)
		if err != nil {
			t.Fatal(err)
		}
		if got != amount {
			t.Fatalf("handle %d decrypts to %d, want %d", i, got, amount)
		}
	}

	// Out-of-range handle indices are rejected in Go.
	if _, err := grouped.ToElGamalCiphertext(3); err == nil {
		t.Fatal("handle index 3 accepted for a 3-handle ciphertext")
	}
	if _, err := grouped.ToElGamalCiphertext(-1); err == nil {
		t.Fatal("negative handle index accepted")
	}
}

func TestGroupedElGamalEncrypt2RoundTrip(t *testing.T) {
	amount := zktest.GenAmount(t, 1<<32-1)
	keypairs := [2]*encryption.ElGamalKeypair{zktest.GenKeyPair(t), zktest.GenKeyPair(t)}
	pubkeys := [2]encryption.ElGamalPubkey{keypairs[0].Pubkey, keypairs[1].Pubkey}

	opening, err := encryption.NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	grouped, err := encryption.GroupedElGamalEncrypt2(pubkeys, amount, opening)
	if err != nil {
		t.Fatal(err)
	}

	for i, kp := range keypairs {
		ct, err := grouped.ToElGamalCiphertext(i)
		if err != nil {
			t.Fatal(err)
		}
		got, err := kp.DecryptU32(ct)
		if err != nil {
			t.Fatal(err)
		}
		if got != amount {
			t.Fatalf("handle %d decrypts to %d, want %d", i, got, amount)
		}
	}

	if _, err := grouped.ToElGamalCiphertext(2); err == nil {
		t.Fatal("handle index 2 accepted for a 2-handle ciphertext")
	}
}
