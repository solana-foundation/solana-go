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

package addresslookuptable

import (
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

type DeactivateAddressLookupTable struct {
	// [0] = [WRITE] address
	// ···········
	//
	// [1] = [SIGNER] authority
	// ···········
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewDeactivateAddressLookupTableInstructionBuilder creates a new `DeactivateAddressLookupTable` instruction builder.
func NewDeactivateAddressLookupTableInstructionBuilder() *DeactivateAddressLookupTable {
	nd := &DeactivateAddressLookupTable{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 2),
	}
	return nd
}

// SetAddress sets the "address" account.
func (inst *DeactivateAddressLookupTable) SetAddress(address ag_solanago.PublicKey) *DeactivateAddressLookupTable {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(address).WRITE()
	return inst
}

// GetAddress gets the "address" account.
func (inst *DeactivateAddressLookupTable) GetAddress() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetAuthority sets the "authority" account.
func (inst *DeactivateAddressLookupTable) SetAuthority(authority ag_solanago.PublicKey) *DeactivateAddressLookupTable {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(authority).SIGNER()
	return inst
}

// GetAuthority gets the "authority" account.
func (inst *DeactivateAddressLookupTable) GetAuthority() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

func (inst DeactivateAddressLookupTable) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_DeactivateAddressLookupTable, ag_binary.LE),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst DeactivateAddressLookupTable) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *DeactivateAddressLookupTable) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.address is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.authority is not set")
		}
	}
	return nil
}

func (inst *DeactivateAddressLookupTable) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("DeactivateAddressLookupTable")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("  address", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("authority", inst.AccountMetaSlice.Get(1)))
					})
				})
		})
}

func (obj DeactivateAddressLookupTable) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *DeactivateAddressLookupTable) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewDeactivateAddressLookupTableInstruction declares a new DeactivateAddressLookupTable instruction with the provided parameters and accounts.
func NewDeactivateAddressLookupTableInstruction(
	// Accounts:
	address ag_solanago.PublicKey,
	authority ag_solanago.PublicKey) *DeactivateAddressLookupTable {
	return NewDeactivateAddressLookupTableInstructionBuilder().
		SetAddress(address).
		SetAuthority(authority)
}
