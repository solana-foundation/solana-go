// Package zktest provides helpers for the zk-elgamal-proof package tests.
package zktest

import (
	"errors"
	"math/rand/v2"
	"testing"

	zk "github.com/gagliardetto/solana-go/programs/zk-elgamal-proof"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
)

// GenKeyPair returns a fresh random ElGamal keypair.
func GenKeyPair(t *testing.T) *encryption.ElGamalKeypair {
	t.Helper()
	kp, err := encryption.NewElGamalKeypair()
	if err != nil {
		t.Fatal(err)
	}
	return kp
}

// GenAmount returns a random amount in [1, max].
func GenAmount(t *testing.T, max uint64) uint64 {
	t.Helper()
	n := rand.Uint64N(max) + 1
	t.Logf("amount=%d", n)
	return n
}

// ExpectStatusError asserts err matches the prover status code.
func ExpectStatusError(t *testing.T, err error, status int32) {
	t.Helper()
	if !errors.Is(err, zk.Error(status)) {
		t.Fatalf("expected %v, got %v", zk.Error(status), err)
	}
}
