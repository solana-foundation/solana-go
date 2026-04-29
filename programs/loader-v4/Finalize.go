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
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Finalize promotes a Deployed program into the terminal Finalized state,
// making it immutable. The third account records a pointer to the
// next-version program (for upgrade discovery) without enforcing that
// pointer's semantics.
//
// Account references:
//
//	[0] = [WRITE]   Program account
//	[1] = [SIGNER]  Authority
//	[2] = []        Next-version program account
type Finalize struct {
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewFinalizeInstructionBuilder() *Finalize {
	return &Finalize{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 3),
	}
}

func (inst *Finalize) SetProgramAccount(program ag_solanago.PublicKey) *Finalize {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(program).WRITE()
	return inst
}

func (inst *Finalize) SetAuthority(authority ag_solanago.PublicKey) *Finalize {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(authority).SIGNER()
	return inst
}

func (inst *Finalize) SetNextVersionProgram(next ag_solanago.PublicKey) *Finalize {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(next)
	return inst
}

func (inst Finalize) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_Finalize, binary.LittleEndian),
	}}
}

func (inst Finalize) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Finalize) Validate() error {
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *Finalize) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Finalize")).
				ParentFunc(func(ib ag_treeout.Branches) {
					ib.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						a.Child(ag_format.Meta("    Program", inst.AccountMetaSlice[0]))
						a.Child(ag_format.Meta("  Authority", inst.AccountMetaSlice[1]))
						a.Child(ag_format.Meta("NextVersion", inst.AccountMetaSlice[2]))
					})
				})
		})
}

func (inst Finalize) MarshalWithEncoder(_ *ag_binary.Encoder) error    { return nil }
func (inst *Finalize) UnmarshalWithDecoder(_ *ag_binary.Decoder) error { return nil }

func NewFinalizeInstruction(
	program, authority, nextVersionProgram ag_solanago.PublicKey,
) *Finalize {
	return NewFinalizeInstructionBuilder().
		SetProgramAccount(program).
		SetAuthority(authority).
		SetNextVersionProgram(nextVersionProgram)
}
