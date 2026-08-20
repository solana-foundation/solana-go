package zk

import (
	"testing"
)

func TestGroupedElGamalEncrypt3RoundTrip(t *testing.T) {
	amount := genAmount(t, 1<<32-1)
	keypairs := [3]*ElGamalKeypair{genKeyPair(t), genKeyPair(t), genKeyPair(t)}
	pubkeys := [3]ElGamalPubkey{keypairs[0].Pubkey, keypairs[1].Pubkey, keypairs[2].Pubkey}

	opening, err := NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	grouped, err := GroupedElGamalEncrypt3(pubkeys, amount, opening)
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
	amount := genAmount(t, 1<<32-1)
	keypairs := [2]*ElGamalKeypair{genKeyPair(t), genKeyPair(t)}
	pubkeys := [2]ElGamalPubkey{keypairs[0].Pubkey, keypairs[1].Pubkey}

	opening, err := NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	grouped, err := GroupedElGamalEncrypt2(pubkeys, amount, opening)
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
