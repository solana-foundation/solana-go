// Copyright 2021 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package token

import (
	"bytes"
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Batch allows executing multiple token instructions in a single CPI call,
// reducing the overhead of multiple cross-program invocations.
//
// Each sub-instruction in the batch is prefixed with a 2-byte header:
//   - byte 0: number of accounts for this sub-instruction
//   - byte 1: length of instruction data for this sub-instruction
//
// This instruction is only available in the p-token (Pinocchio) implementation.
type Batch struct {
	Instructions []*Instruction

	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewBatchInstructionBuilder() *Batch {
	return &Batch{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 0),
	}
}

func (inst *Batch) AddInstruction(ix *Instruction) *Batch {
	inst.Instructions = append(inst.Instructions, ix)
	return inst
}

func (inst Batch) Build() *Instruction {
	accounts := make(ag_solanago.AccountMetaSlice, 0)
	for _, ix := range inst.Instructions {
		accounts = append(accounts, ix.Accounts()...)
	}
	inst.AccountMetaSlice = accounts

	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_Batch),
	}}
}

func (inst Batch) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Batch) Validate() error {
	if len(inst.Instructions) == 0 {
		return errors.New("batch must contain at least one instruction")
	}
	return nil
}

func (inst *Batch) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Batch")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("InstructionCount", len(inst.Instructions)))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						for i, acc := range inst.AccountMetaSlice {
							accountsBranch.Child(ag_format.Meta(fmt.Sprintf("[%v]", i), acc))
						}
					})
				})
		})
}

func (obj Batch) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	for _, ix := range obj.Instructions {
		accountCount := uint8(len(ix.Accounts()))

		data, err := ix.Data()
		if err != nil {
			return fmt.Errorf("unable to encode batch sub-instruction: %w", err)
		}
		// data includes the discriminator byte from the outer Instruction encoding,
		// but for batch sub-instructions we need the raw inner data (discriminator + params).
		// The ix.Data() already produces [discriminator | params], which is what we need.
		dataLen := uint8(len(data))

		if err = encoder.WriteUint8(accountCount); err != nil {
			return err
		}
		if err = encoder.WriteUint8(dataLen); err != nil {
			return err
		}
		if _, err = encoder.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func (obj *Batch) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	for decoder.HasRemaining() {
		accountCount, err := decoder.ReadUint8()
		if err != nil {
			return err
		}
		dataLen, err := decoder.ReadUint8()
		if err != nil {
			return err
		}
		_ = accountCount

		data, err := decoder.ReadNBytes(int(dataLen))
		if err != nil {
			return err
		}
		ix := new(Instruction)
		if err = ag_binary.NewBinDecoder(data).Decode(ix); err != nil {
			return fmt.Errorf("unable to decode batch sub-instruction: %w", err)
		}
		obj.Instructions = append(obj.Instructions, ix)
	}
	return nil
}

// BuildBatchData constructs the complete instruction data for a batch,
// including the batch discriminator (255) and all sub-instruction data.
func BuildBatchData(instructions []*Instruction) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(Instruction_Batch)
	for _, ix := range instructions {
		accountCount := uint8(len(ix.Accounts()))
		data, err := ix.Data()
		if err != nil {
			return nil, fmt.Errorf("unable to encode batch sub-instruction: %w", err)
		}
		dataLen := uint8(len(data))
		buf.WriteByte(accountCount)
		buf.WriteByte(dataLen)
		buf.Write(data)
	}
	return buf.Bytes(), nil
}

func NewBatchInstruction(instructions ...*Instruction) *Batch {
	b := NewBatchInstructionBuilder()
	for _, ix := range instructions {
		b.AddInstruction(ix)
	}
	return b
}
