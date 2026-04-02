package token2022

import (
	"bytes"
	"encoding/binary"

	"github.com/gagliardetto/solana-go"
)

type PausableInstruction byte

const (
	InitializePausable PausableInstruction = 0
	PausePausable      PausableInstruction = 1
	ResumePausable     PausableInstruction = 2
)

const DEFAULT_PAUSABLE_MINT_LEN = 203

func NewPausableInstruction(
	mint solana.PublicKey,
	authority solana.PublicKey,
) solana.Instruction {
	programID := ProgramID

	pointerData := createPausableInstructionData{
		Instruction:                    PausableExtension,
		DefaultAccountStateInstruction: initialize,
		Authority:                      authority,
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

type createPausableInstructionData struct {
	Instruction                    TokenInstruction
	DefaultAccountStateInstruction programInstruction
	Authority                      solana.PublicKey
}

func (data *createPausableInstructionData) encode() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, data.Instruction)
	binary.Write(&buf, binary.LittleEndian, data.DefaultAccountStateInstruction)
	buf.Write(data.Authority.Bytes())
	return buf.Bytes()
}
