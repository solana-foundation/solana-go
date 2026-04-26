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

// Initializes a new account to hold tokens.  If this account is associated
// with the native mint then the token balance of the initialized account
// will be equal to the amount of SOL in the account. If this account is
// associated with another mint, that mint must be initialized before this
// command can succeed.
//
// The `InitializeAccount` instruction requires no signers and MUST be
// included within the same Transaction as the system program's
// `CreateAccount` instruction that creates the account being initialized.
// Otherwise another party can acquire ownership of the uninitialized
// account.
type InitializeAccount struct {

	// [0] = [WRITE] account
	// ··········· The account to initialize.
	//
	// [1] = [] mint
	// ··········· The mint this account will be associated with.
	//
	// [2] = [] owner
	// ··········· The new account's owner/multisignature.
	//
	// [3] = [] $(SysVarRentPubkey)
	// ··········· Rent sysvar.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewInitializeAccountInstructionBuilder creates a new `InitializeAccount` instruction builder.
func NewInitializeAccountInstructionBuilder() *InitializeAccount {
	nd := &InitializeAccount{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 4),
	}
	nd.AccountMetaSlice[3] = ag_solanago.Meta(ag_solanago.SysVarRentPubkey)
	return nd
}

// SetAccount sets the "account" account.
// The account to initialize.
func (inst *InitializeAccount) SetAccount(account ag_solanago.PublicKey) *InitializeAccount {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(account).WRITE()
	return inst
}

// GetAccount gets the "account" account.
// The account to initialize.
func (inst *InitializeAccount) GetAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetMintAccount sets the "mint" account.
// The mint this account will be associated with.
func (inst *InitializeAccount) SetMintAccount(mint ag_solanago.PublicKey) *InitializeAccount {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(mint)
	return inst
}

// GetMintAccount gets the "mint" account.
// The mint this account will be associated with.
func (inst *InitializeAccount) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

// SetOwnerAccount sets the "owner" account.
// The new account's owner/multisignature.
func (inst *InitializeAccount) SetOwnerAccount(owner ag_solanago.PublicKey) *InitializeAccount {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(owner)
	return inst
}

// GetOwnerAccount gets the "owner" account.
// The new account's owner/multisignature.
func (inst *InitializeAccount) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[2]
}

// SetSysVarRentPubkeyAccount sets the "$(SysVarRentPubkey)" account.
// Rent sysvar.
func (inst *InitializeAccount) SetSysVarRentPubkeyAccount(SysVarRentPubkey ag_solanago.PublicKey) *InitializeAccount {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(SysVarRentPubkey)
	return inst
}

// GetSysVarRentPubkeyAccount gets the "$(SysVarRentPubkey)" account.
// Rent sysvar.
func (inst *InitializeAccount) GetSysVarRentPubkeyAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[3]
}

func (inst InitializeAccount) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_InitializeAccount),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst InitializeAccount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeAccount) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Account is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Mint is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.Owner is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.SysVarRentPubkey is not set")
		}
	}
	return nil
}

func (inst *InitializeAccount) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("InitializeAccount")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("   account", inst.AccountMetaSlice[0]))
						accountsBranch.Child(ag_format.Meta("      mint", inst.AccountMetaSlice[1]))
						accountsBranch.Child(ag_format.Meta("     owner", inst.AccountMetaSlice[2]))
						accountsBranch.Child(ag_format.Meta("SysVarRent", inst.AccountMetaSlice[3]))
					})
				})
		})
}

func (obj InitializeAccount) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *InitializeAccount) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewInitializeAccountInstruction declares a new InitializeAccount instruction with the provided parameters and accounts.
func NewInitializeAccountInstruction(
	// Accounts:
	account ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
	owner ag_solanago.PublicKey,
	SysVarRentPubkey ag_solanago.PublicKey) *InitializeAccount {
	return NewInitializeAccountInstructionBuilder().
		SetAccount(account).
		SetMintAccount(mint).
		SetOwnerAccount(owner).
		SetSysVarRentPubkeyAccount(SysVarRentPubkey)
}
