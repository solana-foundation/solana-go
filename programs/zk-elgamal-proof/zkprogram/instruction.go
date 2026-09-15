package zkprogram

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

// ProofInstruction is an instruction to be run by the verification program.
type ProofInstruction uint8

const (
	CloseContextState ProofInstruction = iota
	VerifyZeroCiphertext
	VerifyCiphertextCiphertextEquality
	VerifyCiphertextCommitmentEquality
	VerifyPubkeyValidity
	VerifyPercentageWithCap
	VerifyBatchedRangeProofU64
	VerifyBatchedRangeProofU128
	VerifyBatchedRangeProofU256
	VerifyGroupedCiphertext2HandlesValidity
	VerifyBatchedGroupedCiphertext2HandlesValidity
	VerifyGroupedCiphertext3HandlesValidity
	VerifyBatchedGroupedCiphertext3HandlesValidity
)

// isValid reports whether i is one of the instructions the program accepts.
func (i ProofInstruction) isValid() bool {
	return i <= VerifyBatchedGroupedCiphertext3HandlesValidity
}

// verifiesProofs reports whether i is one of the Verify* instructions.
func (i ProofInstruction) verifiesProofs() bool {
	return i.isValid() && i != CloseContextState
}

func (i ProofInstruction) String() string {
	if i == CloseContextState {
		return "CloseContextState"
	}
	if !i.isValid() {
		return fmt.Sprintf("ProofInstruction(%d)", uint8(i))
	}
	return "Verify" + proofdata.ProofType(i).String()
}

// Public keys associated with a context state account
type ContextStateInfo struct {
	ContextStateAccount   solana.PublicKey
	ContextStateAuthority solana.PublicKey
}

// CloseContextStateInstruction builds a CloseContextState instruction.
func CloseContextStateInstruction(
	contextStateInfo ContextStateInfo,
	destination solana.PublicKey,
) *solana.GenericInstruction {
	return solana.NewInstruction(ProgramID, solana.AccountMetaSlice{
		solana.Meta(contextStateInfo.ContextStateAccount).WRITE(),
		solana.Meta(destination).WRITE(),
		solana.Meta(contextStateInfo.ContextStateAuthority).SIGNER(),
	}, []byte{byte(CloseContextState)})
}

// EncodeVerifyProof builds the verification instruction for an inlined proof.
func (i ProofInstruction) EncodeVerifyProof(
	contextStateInfo *ContextStateInfo,
	data proofdata.ProofData,
) (*solana.GenericInstruction, error) {
	if isNilProofData(data) {
		return nil, errors.New("zk: proof data not set")
	}
	if !i.verifiesProofs() || data.ProofType() != proofdata.ProofType(i) {
		return nil, fmt.Errorf("zk: %s instruction cannot carry %s proof data", i, data.ProofType())
	}
	proof := data.Bytes()
	out := make([]byte, 0, 1+len(proof))
	out = append(append(out, byte(i)), proof...)
	var accounts solana.AccountMetaSlice
	if contextStateInfo != nil {
		accounts = solana.AccountMetaSlice{
			solana.Meta(contextStateInfo.ContextStateAccount).WRITE(),
			solana.Meta(contextStateInfo.ContextStateAuthority),
		}
	}
	return solana.NewInstruction(ProgramID, accounts, out), nil
}

// EncodeVerifyProofFromAccount builds the verification instruction for i that
// reads the proof from proofAccount starting at offset.
func (i ProofInstruction) EncodeVerifyProofFromAccount(
	contextStateInfo *ContextStateInfo,
	proofAccount solana.PublicKey,
	offset uint32,
) (*solana.GenericInstruction, error) {
	if !i.verifiesProofs() {
		return nil, fmt.Errorf("zk: %s is not a proof verification instruction", i)
	}
	data := make([]byte, 5)
	data[0] = byte(i)
	binary.LittleEndian.PutUint32(data[1:], offset)

	var accounts solana.AccountMetaSlice
	if contextStateInfo != nil {
		accounts = solana.AccountMetaSlice{
			solana.Meta(proofAccount),
			solana.Meta(contextStateInfo.ContextStateAccount).WRITE(),
			solana.Meta(contextStateInfo.ContextStateAuthority),
		}
	} else {
		accounts = solana.AccountMetaSlice{solana.Meta(proofAccount)}
	}
	return solana.NewInstruction(ProgramID, accounts, data), nil
}

// DecodeInstruction wraps a compiled proof program instruction back into a generic instruction
func DecodeInstruction(accounts []*solana.AccountMeta, data []byte) (*solana.GenericInstruction, error) {
	typ, ok := InstructionType(data)
	if !ok {
		return nil, fmt.Errorf("zk: not a proof program instruction: % x", data)
	}
	if typ == CloseContextState {
		if len(data) != 1 {
			return nil, fmt.Errorf("zk: CloseContextState takes no instruction data, got %d bytes", len(data)-1)
		}
		if len(accounts) != 3 {
			return nil, fmt.Errorf("zk: CloseContextState takes 3 accounts, got %d", len(accounts))
		}
		return solana.NewInstruction(ProgramID, accounts, data), nil
	}
	proofLen := len(proofdata.NewProofData(proofdata.ProofType(typ)).Bytes())
	switch len(data) {
	// The program treats 5-byte data as the proof-from-account form: the
	// discriminant followed by a u32 offset into the proof account.
	case 5:
		if len(accounts) != 1 && len(accounts) != 3 {
			return nil, fmt.Errorf("zk: %s reading a proof account takes 1 or 3 accounts, got %d", typ, len(accounts))
		}
	case 1 + proofLen:
		if len(accounts) != 0 && len(accounts) != 2 {
			return nil, fmt.Errorf("zk: %s with inlined proof data takes 0 or 2 accounts, got %d", typ, len(accounts))
		}
	default:
		return nil, fmt.Errorf("zk: %s takes %d bytes of proof data or a 4-byte proof account offset, got %d bytes",
			typ, proofLen, len(data)-1)
	}
	return solana.NewInstruction(ProgramID, accounts, data), nil
}

func registryDecodeInstruction(accounts []*solana.AccountMeta, data []byte) (any, error) {
	return DecodeInstruction(accounts, data)
}

// InstructionType returns the discriminant leading the instruction data.
func InstructionType(data []byte) (ProofInstruction, bool) {
	if len(data) == 0 || !ProofInstruction(data[0]).isValid() {
		return 0, false
	}
	return ProofInstruction(data[0]), true
}
