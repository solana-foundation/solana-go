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

// TransferAuthority hands a program over to a new authority; both the current
// and new authorities must sign (no unchecked variant exists for v4).
//
// Account references:
//
//	[0] = [WRITE]   Program account
//	[1] = [SIGNER]  Current authority
//	[2] = [SIGNER]  New authority
type TransferAuthority struct {
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewTransferAuthorityInstructionBuilder() *TransferAuthority {
	return &TransferAuthority{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 3),
	}
}

func (inst *TransferAuthority) SetProgramAccount(program ag_solanago.PublicKey) *TransferAuthority {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(program).WRITE()
	return inst
}

func (inst *TransferAuthority) SetAuthority(authority ag_solanago.PublicKey) *TransferAuthority {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(authority).SIGNER()
	return inst
}

func (inst *TransferAuthority) SetNewAuthority(newAuthority ag_solanago.PublicKey) *TransferAuthority {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(newAuthority).SIGNER()
	return inst
}

func (inst TransferAuthority) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_TransferAuthority, binary.LittleEndian),
	}}
}

func (inst TransferAuthority) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *TransferAuthority) Validate() error {
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *TransferAuthority) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("TransferAuthority")).
				ParentFunc(func(ib ag_treeout.Branches) {
					ib.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						a.Child(ag_format.Meta("     Program", inst.AccountMetaSlice[0]))
						a.Child(ag_format.Meta("   Authority", inst.AccountMetaSlice[1]))
						a.Child(ag_format.Meta("NewAuthority", inst.AccountMetaSlice[2]))
					})
				})
		})
}

func (inst TransferAuthority) MarshalWithEncoder(_ *ag_binary.Encoder) error    { return nil }
func (inst *TransferAuthority) UnmarshalWithDecoder(_ *ag_binary.Decoder) error { return nil }

func NewTransferAuthorityInstruction(program, authority, newAuthority ag_solanago.PublicKey) *TransferAuthority {
	return NewTransferAuthorityInstructionBuilder().
		SetProgramAccount(program).
		SetAuthority(authority).
		SetNewAuthority(newAuthority)
}
