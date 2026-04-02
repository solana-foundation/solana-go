package token2022

import (
	"bytes"
	"encoding/binary"

	"github.com/gagliardetto/solana-go"
)

const DEFAULT_SCALED_UI_AMOUNT_MINT_LEN = 226

func NewScaledUiAmountInstruction(
	mint solana.PublicKey,
	authority solana.PublicKey,
	multiplier float64,
) solana.Instruction {
	programID := ProgramID

	pointerData := createScaledUiAmountInstructionData{
		Instruction:                    ScaledUiAmountExtension,
		DefaultAccountStateInstruction: initialize,
		Authority:                      authority,
		Multiplier:                     multiplier,
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

type createScaledUiAmountInstructionData struct {
	Instruction                    TokenInstruction
	DefaultAccountStateInstruction programInstruction
	Authority                      solana.PublicKey
	Multiplier                     float64
}

func (data *createScaledUiAmountInstructionData) encode() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, data.Instruction)
	binary.Write(&buf, binary.LittleEndian, data.DefaultAccountStateInstruction)
	buf.Write(data.Authority.Bytes())
	binary.Write(&buf, binary.LittleEndian, data.Multiplier)
	return buf.Bytes()
}
