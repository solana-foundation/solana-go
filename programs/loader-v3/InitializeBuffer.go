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

package loaderv3

import (
	"encoding/binary"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// InitializeBuffer marks a freshly system-allocated account as a loader buffer
// and records the authority that is permitted to write to it.
//
// Account references:
//
//	[0] = [WRITE] Buffer account
//	[1] = []      Authority (not a signer here; only used to record the
//	               authority address)
type InitializeBuffer struct {
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewInitializeBufferInstructionBuilder() *InitializeBuffer {
	return &InitializeBuffer{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 2),
	}
}

func (inst *InitializeBuffer) SetBufferAccount(buffer ag_solanago.PublicKey) *InitializeBuffer {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(buffer).WRITE()
	return inst
}

func (inst *InitializeBuffer) SetAuthority(authority ag_solanago.PublicKey) *InitializeBuffer {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(authority)
	return inst
}

func (inst InitializeBuffer) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_InitializeBuffer, binary.LittleEndian),
	}}
}

func (inst InitializeBuffer) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeBuffer) Validate() error {
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *InitializeBuffer) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("InitializeBuffer")).
				ParentFunc(func(ib ag_treeout.Branches) {
					ib.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						a.Child(ag_format.Meta("   Buffer", inst.AccountMetaSlice[0]))
						a.Child(ag_format.Meta("Authority", inst.AccountMetaSlice[1]))
					})
				})
		})
}

func (inst InitializeBuffer) MarshalWithEncoder(_ *ag_binary.Encoder) error    { return nil }
func (inst *InitializeBuffer) UnmarshalWithDecoder(_ *ag_binary.Decoder) error { return nil }

// NewInitializeBufferInstruction builds the InitializeBuffer instruction
// alone (no system create). Use NewCreateBufferInstructions for the full
// create+initialize pair.
func NewInitializeBufferInstruction(buffer, authority ag_solanago.PublicKey) *InitializeBuffer {
	return NewInitializeBufferInstructionBuilder().
		SetBufferAccount(buffer).
		SetAuthority(authority)
}

// NewCreateBufferInstructions mirrors the upstream `create_buffer` helper:
// it returns the pair of instructions (system CreateAccount followed by
// InitializeBuffer) needed to provision a new buffer account of the given
// program length.
func NewCreateBufferInstructions(
	payer, buffer, authority ag_solanago.PublicKey,
	lamports uint64,
	programLen int,
) []ag_solanago.Instruction {
	create := system.NewCreateAccountInstruction(
		lamports,
		uint64(SizeOfBuffer(programLen)),
		ProgramID,
		payer,
		buffer,
	).Build()
	initBuf := NewInitializeBufferInstruction(buffer, authority).Build()
	return []ag_solanago.Instruction{create, initBuf}
}
