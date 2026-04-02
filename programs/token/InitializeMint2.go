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

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Like InitializeMint, but does not require the Rent sysvar to be provided.
type InitializeMint2 struct {
	// Number of base 10 digits to the right of the decimal place.
	Decimals *uint8

	// The authority/multisignature to mint tokens.
	MintAuthority *ag_solanago.PublicKey

	// The freeze authority/multisignature of the mint.
	FreezeAuthority *ag_solanago.PublicKey `bin:"optional"`

	// [0] = [WRITE] mint
	// ··········· The mint to initialize.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewInitializeMint2InstructionBuilder creates a new `InitializeMint2` instruction builder.
func NewInitializeMint2InstructionBuilder() *InitializeMint2 {
	nd := &InitializeMint2{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

// SetDecimals sets the "decimals" parameter.
// Number of base 10 digits to the right of the decimal place.
func (inst *InitializeMint2) SetDecimals(decimals uint8) *InitializeMint2 {
	inst.Decimals = &decimals
	return inst
}

// SetMintAuthority sets the "mint_authority" parameter.
// The authority/multisignature to mint tokens.
func (inst *InitializeMint2) SetMintAuthority(mint_authority ag_solanago.PublicKey) *InitializeMint2 {
	inst.MintAuthority = &mint_authority
	return inst
}

// SetFreezeAuthority sets the "freeze_authority" parameter.
// The freeze authority/multisignature of the mint.
func (inst *InitializeMint2) SetFreezeAuthority(freeze_authority ag_solanago.PublicKey) *InitializeMint2 {
	inst.FreezeAuthority = &freeze_authority
	return inst
}

// SetMintAccount sets the "mint" account.
// The mint to initialize.
func (inst *InitializeMint2) SetMintAccount(mint ag_solanago.PublicKey) *InitializeMint2 {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// GetMintAccount gets the "mint" account.
// The mint to initialize.
func (inst *InitializeMint2) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst InitializeMint2) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_InitializeMint2),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst InitializeMint2) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeMint2) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Decimals == nil {
			return errors.New("decimals parameter is not set")
		}
		if inst.MintAuthority == nil {
			return errors.New("mintAuthority parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Mint is not set")
		}
	}
	return nil
}

func (inst *InitializeMint2) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("InitializeMint2")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("             Decimals", *inst.Decimals))
						paramsBranch.Child(ag_format.Param("        MintAuthority", *inst.MintAuthority))
						paramsBranch.Child(ag_format.Param("FreezeAuthority (OPT)", inst.FreezeAuthority))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj InitializeMint2) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Decimals` param:
	err = encoder.Encode(obj.Decimals)
	if err != nil {
		return err
	}
	// Serialize `MintAuthority` param:
	err = encoder.Encode(obj.MintAuthority)
	if err != nil {
		return err
	}
	// Serialize `FreezeAuthority` param (optional):
	{
		if obj.FreezeAuthority == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.FreezeAuthority)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (obj *InitializeMint2) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Decimals`:
	err = decoder.Decode(&obj.Decimals)
	if err != nil {
		return err
	}
	// Deserialize `MintAuthority`:
	err = decoder.Decode(&obj.MintAuthority)
	if err != nil {
		return err
	}
	// Deserialize `FreezeAuthority` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.FreezeAuthority)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NewInitializeMint2Instruction declares a new InitializeMint2 instruction with the provided parameters and accounts.
func NewInitializeMint2Instruction(
	// Parameters:
	decimals uint8,
	mint_authority ag_solanago.PublicKey,
	freeze_authority ag_solanago.PublicKey,
	// Accounts:
	mint ag_solanago.PublicKey) *InitializeMint2 {
	return NewInitializeMint2InstructionBuilder().
		SetDecimals(decimals).
		SetMintAuthority(mint_authority).
		SetFreezeAuthority(freeze_authority).
		SetMintAccount(mint)
}
