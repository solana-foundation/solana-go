package encryption

import (
	"fmt"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/bridge"
)

// NewPedersenCommitment commits to amount with a fresh random opening.
func NewPedersenCommitment(amount uint64) (PedersenCommitment, PedersenOpening, error) {
	var commitment PedersenCommitment
	var opening PedersenOpening
	out, err := bridge.InvokeWith("pedersen_commit", amount)
	if err != nil {
		return commitment, opening, err
	}
	if len(out) != 64 {
		return commitment, opening, fmt.Errorf("zk: guest returned %d bytes, want 64", len(out))
	}
	copy(commitment[:], out[:32])
	copy(opening[:], out[32:64])
	return commitment, opening, nil
}

// PedersenCommitmentWith commits to amount with the given opening.
func PedersenCommitmentWith(amount uint64, opening PedersenOpening) (PedersenCommitment, error) {
	var commitment PedersenCommitment
	out, err := bridge.InvokeWith("pedersen_commit_with", amount, opening[:])
	err = bridge.CopyOut(commitment[:], out, err)
	return commitment, err
}

// NewPedersenOpening samples a random Pedersen opening.
func NewPedersenOpening() (PedersenOpening, error) {
	var opening PedersenOpening
	out, err := bridge.InvokeWith("pedersen_opening_new_rand")
	err = bridge.CopyOut(opening[:], out, err)
	return opening, err
}

// CombineLoHiCommitments computes lo + 2^bitLength·hi.
func CombineLoHiCommitments(lo, hi PedersenCommitment, bitLength uint8) (PedersenCommitment, error) {
	var commitment PedersenCommitment
	out, err := bridge.InvokeWith("pedersen_combine_lo_hi_commitments", lo[:], hi[:], uint64(bitLength))
	err = bridge.CopyOut(commitment[:], out, err)
	return commitment, err
}

// CombineLoHiOpenings computes lo + 2^bitLength·hi.
func CombineLoHiOpenings(lo, hi PedersenOpening, bitLength uint8) (PedersenOpening, error) {
	var opening PedersenOpening
	out, err := bridge.InvokeWith("pedersen_combine_lo_hi_openings", lo[:], hi[:], uint64(bitLength))
	err = bridge.CopyOut(opening[:], out, err)
	return opening, err
}

// SubtractCommitments computes a - b.
func SubtractCommitments(a, b PedersenCommitment) (PedersenCommitment, error) {
	var commitment PedersenCommitment
	out, err := bridge.InvokeWith("pedersen_sub_commitments", a[:], b[:])
	err = bridge.CopyOut(commitment[:], out, err)
	return commitment, err
}

// SubtractOpenings computes a - b.
func SubtractOpenings(a, b PedersenOpening) (PedersenOpening, error) {
	var opening PedersenOpening
	out, err := bridge.InvokeWith("pedersen_sub_openings", a[:], b[:])
	err = bridge.CopyOut(opening[:], out, err)
	return opening, err
}
