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

	// accountCounts holds the per-sub-instruction account count taken from the
	// on-chain header. UnmarshalWithDecoder records it so SetAccounts can hand
	// each sub-instruction its own slice of the flat account list; without it
	// the account grouping is lost on decode.
	accountCounts []uint8

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
	counts := make([]uint8, 0, len(inst.Instructions))
	for _, ix := range inst.Instructions {
		ixAccounts := ix.Accounts()
		accounts = append(accounts, ixAccounts...)
		counts = append(counts, uint8(len(ixAccounts)))
	}
	inst.AccountMetaSlice = accounts
	inst.accountCounts = counts

	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   &inst,
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
	for i, ix := range inst.Instructions {
		if ix == nil {
			return fmt.Errorf("batch sub-instruction %d is nil", i)
		}
		// The p-token entrypoint dispatches Batch before anything else
		// precisely so a batch can never contain another batch; nesting is
		// not sound because account ownership is only enforced by the
		// runtime once the outer batch has finished processing.
		if ix.TypeID.Uint8() == Instruction_Batch {
			return fmt.Errorf("batch sub-instruction %d is itself a Batch; nested batches are not supported", i)
		}
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

// SetAccounts stores the flat account list and, when the per-sub-instruction
// account counts are known (i.e. the Batch came from UnmarshalWithDecoder or
// Build), hands each sub-instruction its own slice of that list. Without this
// the sub-instructions of a decoded Batch would come back with no accounts.
func (obj *Batch) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.AccountMetaSlice = accounts

	if len(obj.accountCounts) == 0 {
		// No header information to split on — nothing to distribute.
		return nil
	}
	if len(obj.accountCounts) != len(obj.Instructions) {
		return fmt.Errorf("batch has %d sub-instructions but %d recorded account counts", len(obj.Instructions), len(obj.accountCounts))
	}

	offset := 0
	for i, count := range obj.accountCounts {
		end := offset + int(count)
		if end > len(accounts) {
			return fmt.Errorf("batch sub-instruction %d expects %d accounts, but only %d of %d remain", i, count, len(accounts)-offset, len(accounts))
		}
		settable, ok := obj.Instructions[i].Impl.(ag_solanago.AccountsSettable)
		if !ok {
			return fmt.Errorf("batch sub-instruction %d (%T) cannot receive accounts", i, obj.Instructions[i].Impl)
		}
		if err := settable.SetAccounts(accounts[offset:end]); err != nil {
			return fmt.Errorf("unable to set accounts for batch sub-instruction %d: %w", i, err)
		}
		offset = end
	}
	return nil
}

func (obj Batch) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	for i, ix := range obj.Instructions {
		data, err := ix.Data()
		if err != nil {
			return fmt.Errorf("unable to encode batch sub-instruction: %w", err)
		}
		// The per-sub-instruction header encodes the account count and data
		// length as single bytes, so anything above 255 cannot be represented.
		// Reject it explicitly instead of silently truncating via the uint8 cast.
		numAccounts := len(ix.Accounts())
		if numAccounts > 255 {
			return fmt.Errorf("batch sub-instruction %d has %d accounts, but the per-instruction maximum is 255", i, numAccounts)
		}
		if len(data) > 255 {
			return fmt.Errorf("batch sub-instruction %d has %d bytes of data, but the per-instruction maximum is 255", i, len(data))
		}
		if err = encoder.WriteUint8(uint8(numAccounts)); err != nil {
			return err
		}
		if err = encoder.WriteUint8(uint8(len(data))); err != nil {
			return err
		}
		if _, err = encoder.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func (obj *Batch) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	obj.Instructions = nil
	obj.accountCounts = nil
	for decoder.HasRemaining() {
		accountCount, err := decoder.ReadUint8()
		if err != nil {
			return err
		}
		dataLen, err := decoder.ReadUint8()
		if err != nil {
			return err
		}
		// The header length covers the discriminator, so a zero-length
		// sub-instruction is malformed — the program rejects it too.
		if dataLen == 0 {
			return fmt.Errorf("batch sub-instruction %d has no instruction data", len(obj.Instructions))
		}

		data, err := decoder.ReadNBytes(int(dataLen))
		if err != nil {
			return err
		}
		ix := new(Instruction)
		if err = ag_binary.NewBinDecoder(data).Decode(ix); err != nil {
			return fmt.Errorf("unable to decode batch sub-instruction: %w", err)
		}
		obj.Instructions = append(obj.Instructions, ix)
		obj.accountCounts = append(obj.accountCounts, accountCount)
	}
	return nil
}

// BuildBatchData is a convenience wrapper that returns the fully-encoded batch
// instruction bytes (discriminator + sub-instruction headers + data).
func BuildBatchData(instructions []*Instruction) ([]byte, error) {
	return NewBatchInstruction(instructions...).Build().Data()
}

func NewBatchInstruction(instructions ...*Instruction) *Batch {
	b := NewBatchInstructionBuilder()
	for _, ix := range instructions {
		b.AddInstruction(ix)
	}
	return b
}
