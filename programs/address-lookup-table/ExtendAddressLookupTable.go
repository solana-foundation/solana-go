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

type ExtendAddressLookupTable struct {
	// Addresses
	Addresses []ag_solanago.PublicKey

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

// NewExtendAddressLookupTableInstructionBuilder creates a new `ExtendAddressLookupTable` instruction builder.
func NewExtendAddressLookupTableInstructionBuilder() *ExtendAddressLookupTable {
	nd := &ExtendAddressLookupTable{
		Addresses:        make([]ag_solanago.PublicKey, 0),
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 4),
	}
	nd.AccountMetaSlice[3] = ag_solanago.Meta(ag_solanago.SystemProgramID)
	return nd
}

// SetRecentSlot sets the "recentSlow" parameter.
func (inst *ExtendAddressLookupTable) SetAddresses(addresses []ag_solanago.PublicKey) *ExtendAddressLookupTable {
	inst.Addresses = addresses
	return inst
}

// GetAddress gets the "address" account.
func (inst *ExtendAddressLookupTable) GetAddresses() []ag_solanago.PublicKey {
	return inst.Addresses
}

// SetAddress sets the "address" account.
func (inst *ExtendAddressLookupTable) SetAddress(address ag_solanago.PublicKey) *ExtendAddressLookupTable {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(address).WRITE()
	return inst
}

// GetAddress gets the "address" account.
func (inst *ExtendAddressLookupTable) GetAddress() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetAuthority sets the "authority" account.
func (inst *ExtendAddressLookupTable) SetAuthority(authority ag_solanago.PublicKey) *ExtendAddressLookupTable {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(authority).SIGNER()
	return inst
}

// GetAuthority gets the "authority" account.
func (inst *ExtendAddressLookupTable) GetAuthority() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

// SetPayer sets the "payer" account.
func (inst *ExtendAddressLookupTable) SetPayer(payer ag_solanago.PublicKey) *ExtendAddressLookupTable {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(payer).WRITE().SIGNER()
	return inst
}

// GetPayer gets the "payer" account.
func (inst *ExtendAddressLookupTable) GetPayer() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[2]
}

func (inst ExtendAddressLookupTable) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_ExtendAddressLookupTable, ag_binary.LE),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst ExtendAddressLookupTable) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *ExtendAddressLookupTable) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Addresses == nil {
			return errors.New("RecentSlot parameter is not set")
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

func (inst *ExtendAddressLookupTable) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("ExtendAddressLookupTable")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Addresses", inst.Addresses))
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

func (obj ExtendAddressLookupTable) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Addresses` param:
	err = encoder.WriteUint64(uint64(len(obj.Addresses)), ag_binary.LE)
	if err != nil {
		return err
	}
	for _, address := range obj.Addresses {
		err = encoder.WriteBytes(address[:], false)
		if err != nil {
			return err
		}
	}
	return nil
}
func (obj *ExtendAddressLookupTable) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Addresses`:
	l, err := decoder.ReadUint64(ag_binary.LE)
	if err != nil {
		return err
	}
	obj.Addresses = make([]ag_solanago.PublicKey, l)
	for i := 0; i < int(l); i++ {
		v, err := decoder.ReadNBytes(32)
		if err != nil {
			return err
		}
		obj.Addresses[i] = ag_solanago.PublicKeyFromBytes(v)
	}
	return nil
}

// NewExtendAddressLookupTableInstruction declares a new ExtendAddressLookupTable instruction with the provided parameters and accounts.
func NewExtendAddressLookupTableInstruction(
// Parameters:
	addresses []ag_solanago.PublicKey,
// Accounts:
	address ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	payer ag_solanago.PublicKey) *ExtendAddressLookupTable {
	return NewExtendAddressLookupTableInstructionBuilder().
		SetAddresses(addresses).
		SetAddress(address).
		SetAuthority(authority).
		SetPayer(payer)
}
