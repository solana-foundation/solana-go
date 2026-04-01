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

type FreezeAddressLookupTable struct {
	// [0] = [WRITE] address
	// ···········
	//
	// [1] = [SIGNER] authority
	// ···········
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewFreezeAddressLookupTableInstructionBuilder creates a new `FreezeAddressLookupTable` instruction builder.
func NewFreezeAddressLookupTableInstructionBuilder() *FreezeAddressLookupTable {
	nd := &FreezeAddressLookupTable{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 2),
	}
	return nd
}

// SetAddress sets the "address" account.
func (inst *FreezeAddressLookupTable) SetAddress(address ag_solanago.PublicKey) *FreezeAddressLookupTable {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(address).WRITE()
	return inst
}

// GetAddress gets the "address" account.
func (inst *FreezeAddressLookupTable) GetAddress() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetAuthority sets the "authority" account.
func (inst *FreezeAddressLookupTable) SetAuthority(authority ag_solanago.PublicKey) *FreezeAddressLookupTable {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(authority).SIGNER()
	return inst
}

// GetAuthority gets the "authority" account.
func (inst *FreezeAddressLookupTable) GetAuthority() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

func (inst FreezeAddressLookupTable) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_FreezeAddressLookupTable, ag_binary.LE),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst FreezeAddressLookupTable) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *FreezeAddressLookupTable) Validate() error {
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

func (inst *FreezeAddressLookupTable) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("FreezeAddressLookupTable")).
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

func (obj FreezeAddressLookupTable) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *FreezeAddressLookupTable) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewFreezeAddressLookupTableInstruction declares a new FreezeAddressLookupTable instruction with the provided parameters and accounts.
func NewFreezeAddressLookupTableInstruction(
	// Accounts:
	address ag_solanago.PublicKey,
	authority ag_solanago.PublicKey) *FreezeAddressLookupTable {
	return NewFreezeAddressLookupTableInstructionBuilder().
		SetAddress(address).
		SetAuthority(authority)
}
