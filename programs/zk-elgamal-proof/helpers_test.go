package zk

import (
	"errors"
	"math/rand/v2"
	"testing"
)

func genKeyPair(t *testing.T) *ElGamalKeypair {
	t.Helper()
	kp, err := NewElGamalKeypair()
	if err != nil {
		t.Fatal(err)
	}
	return kp
}

// genAmount returns a random amount in [1, max].
func genAmount(t *testing.T, max uint64) uint64 {
	t.Helper()
	n := rand.Uint64N(max) + 1
	t.Logf("amount=%d", n)
	return n
}

// expectStatusError asserts err matches the prover status code.
func expectStatusError(t *testing.T, err error, status int32) {
	t.Helper()
	if !errors.Is(err, Error(status)) {
		t.Fatalf("expected %v, got %v", Error(status), err)
	}
}
