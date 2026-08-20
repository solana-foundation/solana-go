package zk

import (
	"testing"
)

func TestElGamalKeypairFromSecretIsDeterministic(t *testing.T) {
	var secret ElGamalSecretKey
	secret[0] = 7 // canonical scalar

	first, err := ElGamalKeypairFromSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ElGamalKeypairFromSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	if first.Pubkey != second.Pubkey {
		t.Fatalf("pubkey derivation not deterministic: %x vs %x", first.Pubkey, second.Pubkey)
	}
	if first.Secret != secret {
		t.Fatalf("secret roundtrip mismatch")
	}
}

func TestElGamalEncryptDecrypt(t *testing.T) {
	kp := genKeyPair(t)
	amount := genAmount(t, 1<<32-1)

	ct, err := kp.Pubkey.Encrypt(amount)
	if err != nil {
		t.Fatal(err)
	}
	got, err := kp.DecryptU32(ct)
	if err != nil {
		t.Fatal(err)
	}
	if got != amount {
		t.Fatalf("decrypted %d, want %d", got, amount)
	}
}

func TestElGamalHomomorphicAmountOps(t *testing.T) {
	kp := genKeyPair(t)

	// Homomorphic ops: 42 + 8 - 25 = 25
	ct, err := kp.Pubkey.Encrypt(42)
	if err != nil {
		t.Fatal(err)
	}
	ct, err = ct.AddAmount(8)
	if err != nil {
		t.Fatal(err)
	}
	ct, err = ct.SubtractAmount(25)
	if err != nil {
		t.Fatal(err)
	}
	amount, err := kp.DecryptU32(ct)
	if err != nil {
		t.Fatal(err)
	}
	if amount != 25 {
		t.Fatalf("decrypted %d, want 25", amount)
	}
}

func TestElGamalCiphertextArithmetic(t *testing.T) {
	kp := genKeyPair(t)
	a, err := kp.Pubkey.Encrypt(30)
	if err != nil {
		t.Fatal(err)
	}
	b, err := kp.Pubkey.Encrypt(12)
	if err != nil {
		t.Fatal(err)
	}
	sum, err := AddCiphertexts(a, b)
	if err != nil {
		t.Fatal(err)
	}
	amount, err := kp.DecryptU32(sum)
	if err != nil {
		t.Fatal(err)
	}
	if amount != 42 {
		t.Fatalf("decrypted sum %d, want 42", amount)
	}
}
