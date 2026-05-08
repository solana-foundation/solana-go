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

package loaderv4

import (
	"encoding/binary"
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// SetProgramLength resizes a program account. Setting the size to zero
// effectively closes the account and refunds its rent to the recipient.
//
// Account references:
//
//	[0] = [WRITE]   Program account
//	[1] = [SIGNER]  Authority
//	[2] = [WRITE]   Recipient (rent refund destination)
type SetProgramLength struct {
	NewSize *uint32

	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewSetProgramLengthInstructionBuilder() *SetProgramLength {
	return &SetProgramLength{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 3),
	}
}

func (inst *SetProgramLength) SetNewSize(v uint32) *SetProgramLength {
	inst.NewSize = &v
	return inst
}

func (inst *SetProgramLength) SetProgramAccount(program ag_solanago.PublicKey) *SetProgramLength {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(program).WRITE()
	return inst
}

func (inst *SetProgramLength) SetAuthority(authority ag_solanago.PublicKey) *SetProgramLength {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(authority).SIGNER()
	return inst
}

func (inst *SetProgramLength) SetRecipient(recipient ag_solanago.PublicKey) *SetProgramLength {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(recipient).WRITE()
	return inst
}

func (inst SetProgramLength) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_SetProgramLength, binary.LittleEndian),
	}}
}

func (inst SetProgramLength) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *SetProgramLength) Validate() error {
	if inst.NewSize == nil {
		return errors.New("NewSize parameter is not set")
	}
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *SetProgramLength) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("SetProgramLength")).
				ParentFunc(func(ib ag_treeout.Branches) {
					ib.Child("Params").ParentFunc(func(p ag_treeout.Branches) {
						p.Child(ag_format.Param("NewSize", *inst.NewSize))
					})
					ib.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						a.Child(ag_format.Meta("  Program", inst.AccountMetaSlice[0]))
						a.Child(ag_format.Meta("Authority", inst.AccountMetaSlice[1]))
						a.Child(ag_format.Meta("Recipient", inst.AccountMetaSlice[2]))
					})
				})
		})
}

func (inst SetProgramLength) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return encoder.WriteUint32(*inst.NewSize, binary.LittleEndian)
}

func (inst *SetProgramLength) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	v, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	inst.NewSize = &v
	return nil
}

func NewSetProgramLengthInstruction(
	program, authority, recipient ag_solanago.PublicKey,
	newSize uint32,
) *SetProgramLength {
	return NewSetProgramLengthInstructionBuilder().
		SetProgramAccount(program).
		SetAuthority(authority).
		SetRecipient(recipient).
		SetNewSize(newSize)
}

// NewCreateBufferInstructions mirrors the upstream `create_buffer` helper.
// It allocates a zero-size program account via the system program and then
// calls SetProgramLength to grow it to `newSize`.
func NewCreateBufferInstructions(
	payer, buffer, authority, recipient ag_solanago.PublicKey,
	lamports uint64,
	newSize uint32,
) []ag_solanago.Instruction {
	create := system.NewCreateAccountInstruction(
		lamports,
		0,
		ProgramID,
		payer,
		buffer,
	).Build()
	spl := NewSetProgramLengthInstruction(buffer, authority, recipient, newSize).Build()
	return []ag_solanago.Instruction{create, spl}
}
