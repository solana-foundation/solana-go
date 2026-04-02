package token2022

import (
	"bytes"
	"encoding/binary"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
)

const DEFAULT_ACCOUNT_STATE_MINT_LEN = 171

// Instruction to initialize the DefaultAccountState Extension
func NewInitializeDefaultAccountStateInstruction(
	mint solana.PublicKey, // Mint Account address
	accountState token.AccountState, // Default AccountState
) solana.Instruction {
	programID := ProgramID

	pointerData := createInitializeDefaultAccountStateInstructionData{
		Instruction:                    DefaultAccountStateExtension,
		DefaultAccountStateInstruction: initialize,
		AccountState:                   accountState,
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

func CreateUpdateDefaultAccountStateInstruction(
	mint solana.PublicKey, // Mint Account address
	accountState token.AccountState, // Default AccountState
	freezeAuthority solana.PublicKey, // Freeze Authority
) solana.Instruction {
	programID := ProgramID

	pointerData := createInitializeDefaultAccountStateInstructionData{
		Instruction:                    DefaultAccountStateExtension,
		DefaultAccountStateInstruction: update,
		AccountState:                   accountState,
	}

	ix := &instruction{
		programID: programID,
		accounts: []*solana.AccountMeta{
			solana.Meta(mint).WRITE(),
			solana.Meta(freezeAuthority).SIGNER(),
		},
		data: pointerData.encode(),
	}

	return ix
}

type createInitializeDefaultAccountStateInstructionData struct {
	Instruction                    TokenInstruction
	DefaultAccountStateInstruction programInstruction
	AccountState                   token.AccountState
}

func (data *createInitializeDefaultAccountStateInstructionData) encode() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, data.Instruction)
	binary.Write(&buf, binary.LittleEndian, data.DefaultAccountStateInstruction)
	buf.Write([]byte{byte(data.AccountState)})
	return buf.Bytes()
}
