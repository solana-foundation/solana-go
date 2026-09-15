package encryption

import "github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/internal/bridge"

// NewPedersenCommitment commits to amount with a fresh random opening.
func NewPedersenCommitment(amount uint64) (PedersenCommitment, PedersenOpening, error) {
	pair, err := bridge.InvokeWith[PedersenCommitmentOpening]("pedersen_commit", bridge.Scalar(amount))
	return pair.Commitment, pair.Opening, err
}

// PedersenCommitmentWith commits to amount with the given opening.
func PedersenCommitmentWith(amount uint64, opening PedersenOpening) (PedersenCommitment, error) {
	return bridge.InvokeWith[PedersenCommitment]("pedersen_commit_with", bridge.Scalar(amount), opening)
}

// NewPedersenOpening samples a random Pedersen opening.
func NewPedersenOpening() (PedersenOpening, error) {
	return bridge.InvokeWith[PedersenOpening]("pedersen_opening_new_rand")
}

// CombineLoHiCommitments computes lo + 2^bitLength·hi.
func CombineLoHiCommitments(lo, hi PedersenCommitment, bitLength uint8) (PedersenCommitment, error) {
	return bridge.InvokeWith[PedersenCommitment]("pedersen_combine_lo_hi_commitments", lo, hi, bridge.Scalar(bitLength))
}

// CombineLoHiOpenings computes lo + 2^bitLength·hi.
func CombineLoHiOpenings(lo, hi PedersenOpening, bitLength uint8) (PedersenOpening, error) {
	return bridge.InvokeWith[PedersenOpening]("pedersen_combine_lo_hi_openings", lo, hi, bridge.Scalar(bitLength))
}

// SubtractCommitments computes a - b.
func SubtractCommitments(a, b PedersenCommitment) (PedersenCommitment, error) {
	return bridge.InvokeWith[PedersenCommitment]("pedersen_sub_commitments", a, b)
}

// SubtractOpenings computes a - b.
func SubtractOpenings(a, b PedersenOpening) (PedersenOpening, error) {
	return bridge.InvokeWith[PedersenOpening]("pedersen_sub_openings", a, b)
}
