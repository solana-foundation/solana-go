package zk

import (
	"errors"
	"fmt"
)

const OK = 0

// Error codes returned by the wasm prover.
const (
	BAD_INPUT                = -1
	PROOF_GENERATION_ERROR   = -2
	PROOF_VERIFICATION_ERROR = -3
	DECRYPTION_ERROR         = -4
	UNKNOWN_PROOF_TYPE       = -5
	OOM                      = -6
)

var (
	ErrBadInput          = errors.New("zk: invalid input encoding")
	ErrProofGeneration   = errors.New("zk: proof generation failed")
	ErrProofVerification = errors.New("zk: proof verification failed")
	ErrDecryption        = errors.New("zk: decryption failed")
	ErrUnknownProofType  = errors.New("zk: unknown proof type")
	ErrOutOfMemory       = errors.New("zk: out of memory")
)

// Proof-generation errors, mirroring TokenProofGenerationError in the
// spl-token confidential-transfer proof-generation crate.
var (
	ErrNotEnoughFunds         = errors.New("zk: not enough funds in account")
	ErrIllegalAmountBitLength = errors.New("zk: amount has illegal bit length")
	ErrFeeCalculation         = errors.New("zk: fee calculation failed")
)

// Error maps a prover status code to its sentinel error.
func Error(status int32) error {
	switch status {
	case BAD_INPUT:
		return ErrBadInput
	case PROOF_GENERATION_ERROR:
		return ErrProofGeneration
	case PROOF_VERIFICATION_ERROR:
		return ErrProofVerification
	case DECRYPTION_ERROR:
		return ErrDecryption
	case UNKNOWN_PROOF_TYPE:
		return ErrUnknownProofType
	case OOM:
		return ErrOutOfMemory
	default:
		return fmt.Errorf("zk: unknown error %d", status)
	}
}
