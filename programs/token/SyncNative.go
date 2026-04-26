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

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Given a wrapped / native token account (a token account containing SOL)
// updates its amount field based on the account's underlying `lamports`.
// This is useful if a non-wrapped SOL account uses `system_instruction::transfer`
// to move lamports to a wrapped token account, and needs to have its token
// `amount` field updated.
type SyncNative struct {

	// [0] = [WRITE] tokenAccount
	// ··········· The native token account to sync with its underlying lamports.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewSyncNativeInstructionBuilder creates a new `SyncNative` instruction builder.
func NewSyncNativeInstructionBuilder() *SyncNative {
	nd := &SyncNative{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

// SetTokenAccount sets the "tokenAccount" account.
// The native token account to sync with its underlying lamports.
func (inst *SyncNative) SetTokenAccount(tokenAccount ag_solanago.PublicKey) *SyncNative {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(tokenAccount).WRITE()
	return inst
}

// GetTokenAccount gets the "tokenAccount" account.
// The native token account to sync with its underlying lamports.
func (inst *SyncNative) GetTokenAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst SyncNative) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_SyncNative),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst SyncNative) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *SyncNative) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.TokenAccount is not set")
		}
	}
	return nil
}

func (inst *SyncNative) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("SyncNative")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("tokenAccount", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj SyncNative) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *SyncNative) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewSyncNativeInstruction declares a new SyncNative instruction with the provided parameters and accounts.
func NewSyncNativeInstruction(
	// Accounts:
	tokenAccount ag_solanago.PublicKey) *SyncNative {
	return NewSyncNativeInstructionBuilder().
		SetTokenAccount(tokenAccount)
}
