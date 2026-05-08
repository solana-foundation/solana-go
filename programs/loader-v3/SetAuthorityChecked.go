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
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// SetAuthorityChecked is the successor to SetAuthority: the new authority
// must co-sign the transaction, preventing typos that would lock a program.
//
// Account references:
//
//	[0] = [WRITE]   Target account (buffer or programdata)
//	[1] = [SIGNER]  Current authority
//	[2] = [SIGNER]  New authority
type SetAuthorityChecked struct {
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewSetAuthorityCheckedInstructionBuilder() *SetAuthorityChecked {
	return &SetAuthorityChecked{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 3),
	}
}

func (inst SetAuthorityChecked) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_SetAuthorityChecked, binary.LittleEndian),
	}}
}

func (inst SetAuthorityChecked) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *SetAuthorityChecked) Validate() error {
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *SetAuthorityChecked) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("SetAuthorityChecked")).
				ParentFunc(func(ib ag_treeout.Branches) {
					ib.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						a.Child(ag_format.Meta("         Target", inst.AccountMetaSlice[0]))
						a.Child(ag_format.Meta("CurrentAuthority", inst.AccountMetaSlice[1]))
						a.Child(ag_format.Meta("    NewAuthority", inst.AccountMetaSlice[2]))
					})
				})
		})
}

func (inst SetAuthorityChecked) MarshalWithEncoder(_ *ag_binary.Encoder) error    { return nil }
func (inst *SetAuthorityChecked) UnmarshalWithDecoder(_ *ag_binary.Decoder) error { return nil }

// NewSetBufferAuthorityCheckedInstruction builds a SetAuthorityChecked that
// transfers a buffer's authority with new-authority co-signing.
func NewSetBufferAuthorityCheckedInstruction(
	buffer, currentAuthority, newAuthority ag_solanago.PublicKey,
) *SetAuthorityChecked {
	inst := NewSetAuthorityCheckedInstructionBuilder()
	inst.AccountMetaSlice[0] = ag_solanago.Meta(buffer).WRITE()
	inst.AccountMetaSlice[1] = ag_solanago.Meta(currentAuthority).SIGNER()
	inst.AccountMetaSlice[2] = ag_solanago.Meta(newAuthority).SIGNER()
	return inst
}

// NewSetUpgradeAuthorityCheckedInstruction builds a SetAuthorityChecked that
// transfers a program's upgrade authority with new-authority co-signing.
func NewSetUpgradeAuthorityCheckedInstruction(
	program, currentAuthority, newAuthority ag_solanago.PublicKey,
) *SetAuthorityChecked {
	programDataPDA := MustGetProgramDataAddress(program)
	inst := NewSetAuthorityCheckedInstructionBuilder()
	inst.AccountMetaSlice[0] = ag_solanago.Meta(programDataPDA).WRITE()
	inst.AccountMetaSlice[1] = ag_solanago.Meta(currentAuthority).SIGNER()
	inst.AccountMetaSlice[2] = ag_solanago.Meta(newAuthority).SIGNER()
	return inst
}
