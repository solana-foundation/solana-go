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

// Transfers tokens from one account to another either directly or via a
// delegate.  If this account is associated with the native mint then equal
// amounts of SOL and Tokens will be transferred to the destination
// account.
type Transfer struct {
	// The amount of tokens to transfer.
	Amount *uint64

	// [0] = [WRITE] source
	// ··········· The source account.
	//
	// [1] = [WRITE] destination
	// ··········· The destination account.
	//
	// [2] = [] owner
	// ··········· The source account owner/delegate.
	//
	// [3...] = [SIGNER] signers
	// ··········· M signer accounts.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *Transfer) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(3)
	return nil
}

func (slice Transfer) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

// NewTransferInstructionBuilder creates a new `Transfer` instruction builder.
func NewTransferInstructionBuilder() *Transfer {
	nd := &Transfer{
		Accounts: make(ag_solanago.AccountMetaSlice, 3),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

// SetAmount sets the "amount" parameter.
// The amount of tokens to transfer.
func (inst *Transfer) SetAmount(amount uint64) *Transfer {
	inst.Amount = &amount
	return inst
}

// SetSourceAccount sets the "source" account.
// The source account.
func (inst *Transfer) SetSourceAccount(source ag_solanago.PublicKey) *Transfer {
	inst.Accounts[0] = ag_solanago.Meta(source).WRITE()
	return inst
}

// GetSourceAccount gets the "source" account.
// The source account.
func (inst *Transfer) GetSourceAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

// SetDestinationAccount sets the "destination" account.
// The destination account.
func (inst *Transfer) SetDestinationAccount(destination ag_solanago.PublicKey) *Transfer {
	inst.Accounts[1] = ag_solanago.Meta(destination).WRITE()
	return inst
}

// GetDestinationAccount gets the "destination" account.
// The destination account.
func (inst *Transfer) GetDestinationAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

// SetOwnerAccount sets the "owner" account.
// The source account owner/delegate.
func (inst *Transfer) SetOwnerAccount(owner ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *Transfer {
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
// The source account owner/delegate.
func (inst *Transfer) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[2]
}

func (inst Transfer) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_Transfer),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst Transfer) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Transfer) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Amount == nil {
			return errors.New("amount parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.Accounts[0] == nil {
			return fmt.Errorf("accounts.Source is not set")
		}
		if inst.Accounts[1] == nil {
			return fmt.Errorf("accounts.Destination is not set")
		}
		if inst.Accounts[2] == nil {
			return fmt.Errorf("accounts.Owner is not set")
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

func (inst *Transfer) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Transfer")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Amount", *inst.Amount))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("     source", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("destination", inst.Accounts[1]))
						accountsBranch.Child(ag_format.Meta("      owner", inst.Accounts[2]))

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

func (obj Transfer) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Amount` param:
	err = encoder.Encode(obj.Amount)
	if err != nil {
		return err
	}
	return nil
}
func (obj *Transfer) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Amount`:
	err = decoder.Decode(&obj.Amount)
	if err != nil {
		return err
	}
	return nil
}

// NewTransferInstruction declares a new Transfer instruction with the provided parameters and accounts.
func NewTransferInstruction(
	// Parameters:
	amount uint64,
	// Accounts:
	source ag_solanago.PublicKey,
	destination ag_solanago.PublicKey,
	owner ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey) *Transfer {
	return NewTransferInstructionBuilder().
		SetAmount(amount).
		SetSourceAccount(source).
		SetDestinationAccount(destination).
		SetOwnerAccount(owner, multisigSigners...)
}
