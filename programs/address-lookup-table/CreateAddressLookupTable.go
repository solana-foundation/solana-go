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

type CreateAddressLookupTable struct {
	// RecentSlot
	RecentSlot *uint64
	// Bump
	Bump *uint8

	// [0] = [WRITE] address
	// ···········
	//
	// [1] = [SIGNER] authority
	// ···········
	//
	// [2] = [WRITE, SIGNER] payer
	// ···········
	//
	// [3] = [] system
	// ···········
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewCreateAddressLookupTableInstructionBuilder creates a new `CreateAddressLookupTable` instruction builder.
func NewCreateAddressLookupTableInstructionBuilder() *CreateAddressLookupTable {
	nd := &CreateAddressLookupTable{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 4),
	}
	nd.AccountMetaSlice[3] = ag_solanago.Meta(ag_solanago.SystemProgramID)
	return nd
}

// SetRecentSlot sets the "recentSlow" parameter.
func (inst *CreateAddressLookupTable) SetRecentSlot(recentSlot uint64) *CreateAddressLookupTable {
	inst.RecentSlot = &recentSlot
	return inst
}

// SetBump sets the "recentSlow" parameter.
func (inst *CreateAddressLookupTable) SetBump(bump uint8) *CreateAddressLookupTable {
	inst.Bump = &bump
	return inst
}

// SetAddress sets the "address" account.
func (inst *CreateAddressLookupTable) SetAddress(address ag_solanago.PublicKey) *CreateAddressLookupTable {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(address).WRITE()
	return inst
}

// GetAddress gets the "address" account.
func (inst *CreateAddressLookupTable) GetAddress() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetAuthority sets the "authority" account.
func (inst *CreateAddressLookupTable) SetAuthority(authority ag_solanago.PublicKey) *CreateAddressLookupTable {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(authority).SIGNER()
	return inst
}

// GetAuthority gets the "authority" account.
func (inst *CreateAddressLookupTable) GetAuthority() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

// SetPayer sets the "payer" account.
func (inst *CreateAddressLookupTable) SetPayer(payer ag_solanago.PublicKey) *CreateAddressLookupTable {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(payer).WRITE().SIGNER()
	return inst
}

// GetPayer gets the "payer" account.
func (inst *CreateAddressLookupTable) GetPayer() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[2]
}

func (inst CreateAddressLookupTable) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_CreateAddressLookupTable, ag_binary.LE),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CreateAddressLookupTable) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CreateAddressLookupTable) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.RecentSlot == nil {
			return errors.New("RecentSlot parameter is not set")
		}
		if inst.Bump == nil {
			return errors.New("Bump parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.address is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.authority is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.payer is not set")
		}
	}
	return nil
}

func (inst *CreateAddressLookupTable) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CreateAddressLookupTable")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("RecentSlot", *inst.RecentSlot))
						paramsBranch.Child(ag_format.Param("      Bump", *inst.Bump))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("  address", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("authority", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("    payer", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta("   system", inst.AccountMetaSlice.Get(3)))
					})
				})
		})
}

func (obj CreateAddressLookupTable) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `RecentSlot` param:
	err = encoder.Encode(obj.RecentSlot)
	if err != nil {
		return err
	}
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	return nil
}
func (obj *CreateAddressLookupTable) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `RecentSlot`:
	err = decoder.Decode(&obj.RecentSlot)
	if err != nil {
		return err
	}
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	return nil
}

// NewCreateAddressLookupTableInstruction declares a new CreateAddressLookupTable instruction with the provided parameters and accounts.
func NewCreateAddressLookupTableInstruction(
	// Parameters:
	recentSlot uint64,
	bump uint8,
	// Accounts:
	address ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	payer ag_solanago.PublicKey) *CreateAddressLookupTable {
	return NewCreateAddressLookupTableInstructionBuilder().
		SetRecentSlot(recentSlot).
		SetBump(bump).
		SetRecentSlot(recentSlot).
		SetAddress(address).
		SetAuthority(authority).
		SetPayer(payer)
}
