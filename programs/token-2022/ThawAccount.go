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

	ag_binary "github.com/gagliardetto/binary"
	ag_treeout "github.com/gagliardetto/treeout"
	ag_solanago "github.com/solana-foundation/solana-go/v2"
	ag_format "github.com/solana-foundation/solana-go/v2/text/format"
)

// Thaw a Frozen account using the Mint's freeze_authority (if set).
type ThawAccount struct {

	// [0] = [WRITE] account
	// ··········· The account to thaw.
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

func (obj *ThawAccount) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(3)
	return nil
}

func (slice ThawAccount) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

// NewThawAccountInstructionBuilder creates a new `ThawAccount` instruction builder.
func NewThawAccountInstructionBuilder() *ThawAccount {
	nd := &ThawAccount{
		Accounts: make(ag_solanago.AccountMetaSlice, 3),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

// SetAccount sets the "account" account.
// The account to thaw.
func (inst *ThawAccount) SetAccount(account ag_solanago.PublicKey) *ThawAccount {
	inst.Accounts[0] = ag_solanago.Meta(account).WRITE()
	return inst
}

// GetAccount gets the "account" account.
// The account to thaw.
func (inst *ThawAccount) GetAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

// SetMintAccount sets the "mint" account.
// The token mint.
func (inst *ThawAccount) SetMintAccount(mint ag_solanago.PublicKey) *ThawAccount {
	inst.Accounts[1] = ag_solanago.Meta(mint)
	return inst
}

// GetMintAccount gets the "mint" account.
// The token mint.
func (inst *ThawAccount) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

// SetAuthorityAccount sets the "authority" account.
// The mint freeze authority.
func (inst *ThawAccount) SetAuthorityAccount(authority ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *ThawAccount {
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
func (inst *ThawAccount) GetAuthorityAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[2]
}

func (inst ThawAccount) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_ThawAccount),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst ThawAccount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *ThawAccount) Validate() error {
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

func (inst *ThawAccount) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("ThawAccount")).
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

func (obj ThawAccount) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *ThawAccount) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewThawAccountInstruction declares a new ThawAccount instruction with the provided parameters and accounts.
func NewThawAccountInstruction(
	// Accounts:
	account ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *ThawAccount {
	return NewThawAccountInstructionBuilder().
		SetAccount(account).
		SetMintAccount(mint).
		SetAuthorityAccount(authority, multisigSigners...)
}
