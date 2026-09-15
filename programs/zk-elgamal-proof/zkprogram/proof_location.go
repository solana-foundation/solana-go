package zkprogram

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

type ProofLocation[T proofdata.ProofData] struct {
	offset              int8
	proofData           T
	contextStateAccount solana.PublicKey
}

func ConfidentialTransferProofLocation[T proofdata.ProofData](
	contextStateAccount *solana.PublicKey,
	offset int8,
	proofData T,
) ProofLocation[T] {
	if contextStateAccount != nil {
		return ProofLocationContextStateAccount[T](*contextStateAccount)
	}
	return ProofLocationInstructionOffset(offset, proofData)
}

func ProofLocationInstructionOffset[T proofdata.ProofData](offset int8, proofData T) ProofLocation[T] {
	return ProofLocation[T]{offset: offset, proofData: proofData}
}

func ProofLocationContextStateAccount[T proofdata.ProofData](contextStateAccount solana.PublicKey) ProofLocation[T] {
	return ProofLocation[T]{contextStateAccount: contextStateAccount}
}

// IsInstructionOffset reports whether the proof lives in an instruction
func (l ProofLocation[T]) IsInstructionOffset() bool { return l.offset != 0 }

// InstructionOffset is the relative offset of the instruction containing the proof
// to the consuming instruction, zero for the context state account form.
func (l ProofLocation[T]) InstructionOffset() int8 { return l.offset }

// ProofData is the proof carried by the sibling instruction. Only
// set for the instruction offset form.
func (l ProofLocation[T]) ProofData() T { return l.proofData }

// ContextStateAccount is the account holding the verified proof context. Only
// set for the context state account form.
func (l ProofLocation[T]) ContextStateAccount() solana.PublicKey { return l.contextStateAccount }

// Validate rejects the zero value, which names neither a sibling instruction
// nor a context state account, and the instruction offset form with nil proof
// data.
func (l ProofLocation[T]) Validate() error {
	if !l.IsInstructionOffset() {
		if l.contextStateAccount.IsZero() {
			return errors.New("zk: proof location is not set")
		}
		return nil
	}
	if isNilProofData(l.proofData) {
		return errors.New("zk: proof location has no proof data")
	}
	return nil
}

func isNilProofData(data proofdata.ProofData) bool {
	if data == nil {
		return true
	}
	v := reflect.ValueOf(data)
	return v.Kind() == reflect.Pointer && v.IsNil()
}

func ProofContextSize(t proofdata.ProofType) (uint64, error) {
	emptyProof := proofdata.NewProofData(t)
	if emptyProof == nil {
		return 0, fmt.Errorf("zk: proof type %d: %w", t, proofdata.ErrInvalidProofType)
	}
	return ContextStateSize(emptyProof.ContextData()), nil
}
