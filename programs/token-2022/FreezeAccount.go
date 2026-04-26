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

package token2022

import (
	"errors"
	"fmt"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Freeze an Initialized account using the Mint's freeze_authority (if set).
type FreezeAccount struct {

	// [0] = [WRITE] account
	// ··········· The account to freeze.
	//
	// [1] = [] mint
	// ··········· The token mint.
	//
	// [2] = [] authority
	// ··········· The mint freeze authority.
	//
	// [3...] = [SIGNER] signers
	// ··········· M signer accounts.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *FreezeAccount) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(3)
	return nil
}

func (slice FreezeAccount) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

// NewFreezeAccountInstructionBuilder creates a new `FreezeAccount` instruction builder.
func NewFreezeAccountInstructionBuilder() *FreezeAccount {
	nd := &FreezeAccount{
		Accounts: make(ag_solanago.AccountMetaSlice, 3),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

// SetAccount sets the "account" account.
// The account to freeze.
func (inst *FreezeAccount) SetAccount(account ag_solanago.PublicKey) *FreezeAccount {
	inst.Accounts[0] = ag_solanago.Meta(account).WRITE()
	return inst
}

// GetAccount gets the "account" account.
// The account to freeze.
func (inst *FreezeAccount) GetAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

// SetMintAccount sets the "mint" account.
// The token mint.
func (inst *FreezeAccount) SetMintAccount(mint ag_solanago.PublicKey) *FreezeAccount {
	inst.Accounts[1] = ag_solanago.Meta(mint)
	return inst
}

// GetMintAccount gets the "mint" account.
// The token mint.
func (inst *FreezeAccount) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

// SetAuthorityAccount sets the "authority" account.
// The mint freeze authority.
func (inst *FreezeAccount) SetAuthorityAccount(authority ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *FreezeAccount {
	inst.Accounts[2] = ag_solanago.Meta(authority)
	if len(multisigSigners) == 0 {
		inst.Accounts[2].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

// GetAuthorityAccount gets the "authority" account.
// The mint freeze authority.
func (inst *FreezeAccount) GetAuthorityAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[2]
}

func (inst FreezeAccount) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_FreezeAccount),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst FreezeAccount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *FreezeAccount) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.Accounts[0] == nil {
			return errors.New("accounts.Account is not set")
		}
		if inst.Accounts[1] == nil {
			return errors.New("accounts.Mint is not set")
		}
		if inst.Accounts[2] == nil {
			return errors.New("accounts.Authority is not set")
		}
		if !inst.Accounts[2].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
		if len(inst.Signers) > MAX_SIGNERS {
			return fmt.Errorf("too many signers; got %v, but max is 11", len(inst.Signers))
		}
	}
	return nil
}

func (inst *FreezeAccount) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("FreezeAccount")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("  account", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("     mint", inst.Accounts[1]))
						accountsBranch.Child(ag_format.Meta("authority", inst.Accounts[2]))

						signersBranch := accountsBranch.Child(fmt.Sprintf("signers[len=%v]", len(inst.Signers)))
						for i, v := range inst.Signers {
							if len(inst.Signers) > 9 && i < 10 {
								signersBranch.Child(ag_format.Meta(fmt.Sprintf(" [%v]", i), v))
							} else {
								signersBranch.Child(ag_format.Meta(fmt.Sprintf("[%v]", i), v))
							}
						}
					})
				})
		})
}

func (obj FreezeAccount) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *FreezeAccount) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewFreezeAccountInstruction declares a new FreezeAccount instruction with the provided parameters and accounts.
func NewFreezeAccountInstruction(
	// Accounts:
	account ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *FreezeAccount {
	return NewFreezeAccountInstructionBuilder().
		SetAccount(account).
		SetMintAccount(mint).
		SetAuthorityAccount(authority, multisigSigners...)
}
