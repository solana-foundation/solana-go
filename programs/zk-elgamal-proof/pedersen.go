package zk

import "fmt"

// NewPedersenCommitment commits to amount with a fresh random opening.
func NewPedersenCommitment(amount uint64) (PedersenCommitment, PedersenOpening, error) {
	var commitment PedersenCommitment
	var opening PedersenOpening
	out, err := invokeWith("pedersen_commit", amount)
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
	out, err := invokeWith("pedersen_commit_with", amount, opening[:])
	err = copyOut(commitment[:], out, err)
	return commitment, err
}

// NewPedersenOpening samples a random Pedersen opening.
func NewPedersenOpening() (PedersenOpening, error) {
	var opening PedersenOpening
	out, err := invokeWith("pedersen_opening_new_rand")
	err = copyOut(opening[:], out, err)
	return opening, err
}
