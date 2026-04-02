package token2022

import (
	"bytes"
	"encoding/binary"

	"github.com/gagliardetto/solana-go"
)

const DEFAULT_PERMANENT_DELEGATE_MINT_LEN = 202

func NewInitializePermanentDelegateInstruction(
	mint solana.PublicKey,
	permanentDelegate solana.PublicKey,
) solana.Instruction {
	programID := ProgramID

	pointerData := createInitializePermanentDelegateInstructionData{
		Instruction: InitializePermanentDelegate,
		Delegate:    permanentDelegate,
	}

	ix := &instruction{
		programID: programID,
		accounts: []*solana.AccountMeta{
			solana.Meta(mint).WRITE(),
		},
		data: pointerData.encode(),
	}

	return ix
}

type createInitializePermanentDelegateInstructionData struct {
	Instruction TokenInstruction
	Delegate    solana.PublicKey
}

func (data *createInitializePermanentDelegateInstructionData) encode() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, data.Instruction)
	buf.Write(data.Delegate.Bytes())
	return buf.Bytes()
}
