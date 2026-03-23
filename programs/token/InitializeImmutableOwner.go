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

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Initialize the Immutable Owner extension for the given token account.
// Fails if the account has already been initialized, so must be called
// before InitializeAccount.
//
// No-ops in this version of the program, but is included for compatibility
// with the Associated Token Account program.
type InitializeImmutableOwner struct {
	// [0] = [WRITE] account
	// ··········· The account to initialize.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewInitializeImmutableOwnerInstructionBuilder() *InitializeImmutableOwner {
	nd := &InitializeImmutableOwner{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

func (inst *InitializeImmutableOwner) SetAccount(account ag_solanago.PublicKey) *InitializeImmutableOwner {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(account).WRITE()
	return inst
}

func (inst *InitializeImmutableOwner) GetAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst InitializeImmutableOwner) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_InitializeImmutableOwner),
	}}
}

func (inst InitializeImmutableOwner) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeImmutableOwner) Validate() error {
	if inst.AccountMetaSlice[0] == nil {
		return errors.New("accounts.Account is not set")
	}
	return nil
}

func (inst *InitializeImmutableOwner) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("InitializeImmutableOwner")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("account", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj InitializeImmutableOwner) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *InitializeImmutableOwner) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func NewInitializeImmutableOwnerInstruction(
	account ag_solanago.PublicKey,
) *InitializeImmutableOwner {
	return NewInitializeImmutableOwnerInstructionBuilder().
		SetAccount(account)
}
