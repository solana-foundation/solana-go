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
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Burns tokens by removing them from an account.  `BurnChecked` does not
// support accounts associated with the native mint, use `CloseAccount`
// instead.
//
// This instruction differs from Burn in that the decimals value is checked
// by the caller. This may be useful when creating transactions offline or
// within a hardware wallet.
type BurnChecked struct {
	// The amount of tokens to burn.
	Amount *uint64

	// Expected number of base 10 digits to the right of the decimal place.
	Decimals *uint8

	// [0] = [WRITE] source
	// ··········· The account to burn from.
	//
	// [1] = [WRITE] mint
	// ··········· The token mint.
	//
	// [2] = [] owner
	// ··········· The account's owner/delegate.
	//
	// [3...] = [SIGNER] signers
	// ··········· M signer accounts.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *BurnChecked) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(3)
	return nil
}

func (slice BurnChecked) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

// NewBurnCheckedInstructionBuilder creates a new `BurnChecked` instruction builder.
func NewBurnCheckedInstructionBuilder() *BurnChecked {
	nd := &BurnChecked{
		Accounts: make(ag_solanago.AccountMetaSlice, 3),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

// SetAmount sets the "amount" parameter.
// The amount of tokens to burn.
func (inst *BurnChecked) SetAmount(amount uint64) *BurnChecked {
	inst.Amount = &amount
	return inst
}

// SetDecimals sets the "decimals" parameter.
// Expected number of base 10 digits to the right of the decimal place.
func (inst *BurnChecked) SetDecimals(decimals uint8) *BurnChecked {
	inst.Decimals = &decimals
	return inst
}

// SetSourceAccount sets the "source" account.
// The account to burn from.
func (inst *BurnChecked) SetSourceAccount(source ag_solanago.PublicKey) *BurnChecked {
	inst.Accounts[0] = ag_solanago.Meta(source).WRITE()
	return inst
}

// GetSourceAccount gets the "source" account.
// The account to burn from.
func (inst *BurnChecked) GetSourceAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

// SetMintAccount sets the "mint" account.
// The token mint.
func (inst *BurnChecked) SetMintAccount(mint ag_solanago.PublicKey) *BurnChecked {
	inst.Accounts[1] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// GetMintAccount gets the "mint" account.
// The token mint.
func (inst *BurnChecked) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

// SetOwnerAccount sets the "owner" account.
// The account's owner/delegate.
func (inst *BurnChecked) SetOwnerAccount(owner ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *BurnChecked {
	inst.Accounts[2] = ag_solanago.Meta(owner)
	if len(multisigSigners) == 0 {
		inst.Accounts[2].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

// GetOwnerAccount gets the "owner" account.
// The account's owner/delegate.
func (inst *BurnChecked) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[2]
}

func (inst BurnChecked) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_BurnChecked),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst BurnChecked) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *BurnChecked) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Amount == nil {
			return errors.New("amount parameter is not set")
		}
		if inst.Decimals == nil {
			return errors.New("decimals parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.Accounts[0] == nil {
			return errors.New("accounts.Source is not set")
		}
		if inst.Accounts[1] == nil {
			return errors.New("accounts.Mint is not set")
		}
		if inst.Accounts[2] == nil {
			return errors.New("accounts.Owner is not set")
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

func (inst *BurnChecked) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("BurnChecked")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("  Amount", *inst.Amount))
						paramsBranch.Child(ag_format.Param("Decimals", *inst.Decimals))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("source", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("  mint", inst.Accounts[1]))
						accountsBranch.Child(ag_format.Meta(" owner", inst.Accounts[2]))

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

func (obj BurnChecked) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Amount` param:
	err = encoder.Encode(obj.Amount)
	if err != nil {
		return err
	}
	// Serialize `Decimals` param:
	err = encoder.Encode(obj.Decimals)
	if err != nil {
		return err
	}
	return nil
}
func (obj *BurnChecked) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Amount`:
	err = decoder.Decode(&obj.Amount)
	if err != nil {
		return err
	}
	// Deserialize `Decimals`:
	err = decoder.Decode(&obj.Decimals)
	if err != nil {
		return err
	}
	return nil
}

// NewBurnCheckedInstruction declares a new BurnChecked instruction with the provided parameters and accounts.
func NewBurnCheckedInstruction(
	// Parameters:
	amount uint64,
	decimals uint8,
	// Accounts:
	source ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
	owner ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *BurnChecked {
	return NewBurnCheckedInstructionBuilder().
		SetAmount(amount).
		SetDecimals(decimals).
		SetSourceAccount(source).
		SetMintAccount(mint).
		SetOwnerAccount(owner, multisigSigners...)
}
