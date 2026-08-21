package encryption_test

import (
	"testing"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
)

func TestPedersenCommitmentConsistency(t *testing.T) {
	const amount = 987_654
	commitment, opening, err := encryption.NewPedersenCommitment(amount)
	if err != nil {
		t.Fatal(err)
	}

	// Recommitting with the returned opening must reproduce the commitment.
	again, err := encryption.PedersenCommitmentWith(amount, opening)
	if err != nil {
		t.Fatal(err)
	}
	if again != commitment {
		t.Fatalf("commitment not reproducible from its opening: %x vs %x", again, commitment)
	}

	// A different amount under the same opening must commit differently.
	other, err := encryption.PedersenCommitmentWith(amount+1, opening)
	if err != nil {
		t.Fatal(err)
	}
	if other == commitment {
		t.Fatal("different amounts produced identical commitments")
	}
}

func TestNewPedersenOpeningIsRandom(t *testing.T) {
	a, err := encryption.NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	b, err := encryption.NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("two fresh openings are identical")
	}
}
