// Copyright 2024 github.com/cordialsys
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stake

import (
	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

type GetMinimumDelegation struct {
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *GetMinimumDelegation) Validate() error {
	return nil
}

func (inst GetMinimumDelegation) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_GetMinimumDelegation, bin.LE),
	}}
}

func (inst *GetMinimumDelegation) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("GetMinimumDelegation")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
					})
				})
		})
}

// NewGetMinimumDelegationInstructionBuilder creates a new `GetMinimumDelegation` instruction builder.
func NewGetMinimumDelegationInstructionBuilder() *GetMinimumDelegation {
	nd := &GetMinimumDelegation{
		AccountMetaSlice: make(solana.AccountMetaSlice, 0),
	}
	return nd
}

// NewGetMinimumDelegationInstruction declares a new GetMinimumDelegation instruction.
func NewGetMinimumDelegationInstruction() *GetMinimumDelegation {
	return NewGetMinimumDelegationInstructionBuilder()
}
