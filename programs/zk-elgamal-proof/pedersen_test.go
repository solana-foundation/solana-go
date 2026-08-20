package zk

import (
	"testing"
)

func TestPedersenCommitmentConsistency(t *testing.T) {
	const amount = 987_654
	commitment, opening, err := NewPedersenCommitment(amount)
	if err != nil {
		t.Fatal(err)
	}

	// Recommitting with the returned opening must reproduce the commitment.
	again, err := PedersenCommitmentWith(amount, opening)
	if err != nil {
		t.Fatal(err)
	}
	if again != commitment {
		t.Fatalf("commitment not reproducible from its opening: %x vs %x", again, commitment)
	}

	// A different amount under the same opening must commit differently.
	other, err := PedersenCommitmentWith(amount+1, opening)
	if err != nil {
		t.Fatal(err)
	}
	if other == commitment {
		t.Fatal("different amounts produced identical commitments")
	}
}

func TestNewPedersenOpeningIsRandom(t *testing.T) {
	a, err := NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("two fresh openings are identical")
	}
}
