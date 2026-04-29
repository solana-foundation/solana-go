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

package loaderv2

import (
	"encoding/binary"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Finalize an account loaded with program data for execution.
//
// Account references:
//
//	[0] = [WRITE, SIGNER] Account to prepare for execution
//	[1] = []              Rent sysvar
type Finalize struct {
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewFinalizeInstructionBuilder() *Finalize {
	return &Finalize{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 2),
	}
}

func (inst *Finalize) SetAccount(account ag_solanago.PublicKey) *Finalize {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(account).WRITE().SIGNER()
	return inst
}

func (inst *Finalize) GetAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetRentSysvar attaches the rent sysvar account. The helper NewFinalizeInstruction
// uses the canonical sysvar ID; this setter is exposed for callers that wish
// to override it.
func (inst *Finalize) SetRentSysvar(rent ag_solanago.PublicKey) *Finalize {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(rent)
	return inst
}

func (inst *Finalize) GetRentSysvar() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
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
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						a.Child(ag_format.Meta("Account", inst.AccountMetaSlice[0]))
						a.Child(ag_format.Meta("   Rent", inst.AccountMetaSlice[1]))
					})
				})
		})
}

// Finalize carries no payload: the discriminant alone is the data.
func (inst Finalize) MarshalWithEncoder(_ *ag_binary.Encoder) error    { return nil }
func (inst *Finalize) UnmarshalWithDecoder(_ *ag_binary.Decoder) error { return nil }

func NewFinalizeInstruction(account ag_solanago.PublicKey) *Finalize {
	return NewFinalizeInstructionBuilder().
		SetAccount(account).
		SetRentSysvar(ag_solanago.SysVarRentPubkey)
}
