package zkprogram

import (
	"errors"
	"reflect"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
)

type ProofLocation[T proofdata.ProofData] struct {
	offset              int8
	proofData           T
	contextStateAccount solana.PublicKey
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
