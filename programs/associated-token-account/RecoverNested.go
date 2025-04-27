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

package associatedtokenaccount

import (
	"errors"
	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
	format "github.com/gagliardetto/solana-go/text/format"
	treeout "github.com/gagliardetto/treeout"
)

type RecoverNested struct {
	// The nested associated token account (must be owned by OwnerAssociatedToken)
	NestedAssociatedToken solana.PublicKey `bin:"-" borsh_skip:"true"`

	// Token mint for the nested associated token account
	NestedMint solana.PublicKey `bin:"-" borsh_skip:"true"`

	// Wallet's associated token account where the tokens will be transferred
	DestinationAssociatedToken solana.PublicKey `bin:"-" borsh_skip:"true"`

	// Owner associated token account address (must be owned by Owner)
	OwnerAssociatedToken solana.PublicKey `bin:"-" borsh_skip:"true"`

	// Token mint for the owner associated token account
	OwnerMint solana.PublicKey `bin:"-" borsh_skip:"true"`

	// Wallet address for the owner associated token account
	Owner solana.PublicKey `bin:"-" borsh_skip:"true"`

	// [0] = [WRITE] NestedAssociatedToken
	// ··········· Nested associated token account (must be owned by OwnerAssociatedToken)
	//
	// [1] = [] NestedMint
	// ··········· Token mint for the nested associated token account
	//
	// [2] = [WRITE] DestinationAssociatedToken
	// ··········· Wallet's associated token account where the tokens will be transferred
	//
	// [3] = [] OwnerAssociatedToken
	// ··········· Owner associated token account address (must be owned by Owner)
	//
	// [4] = [] OwnerMint
	// ··········· Token mint for the owner associated token account
	//
	// [5] = [SIGNER] Owner
	// ··········· Wallet address for the owner associated token account
	//
	// [6] = [] TokenProgram
	// ··········· SPL token program ID
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewRecoverNestedInstructionBuilder creates a new `RecoverNested` instruction builder.
func NewRecoverNestedInstructionBuilder() *RecoverNested {
	nd := &RecoverNested{}
	return nd
}

func (inst *RecoverNested) SetNestedAssociatedToken(nestedAssociatedToken solana.PublicKey) *RecoverNested {
	inst.NestedAssociatedToken = nestedAssociatedToken
	return inst
}

func (inst *RecoverNested) SetNestedMint(nestedMint solana.PublicKey) *RecoverNested {
	inst.NestedMint = nestedMint
	return inst
}

func (inst *RecoverNested) SetDestinationAssociatedToken(destinationAssociatedToken solana.PublicKey) *RecoverNested {
	inst.DestinationAssociatedToken = destinationAssociatedToken
	return inst
}

func (inst *RecoverNested) SetOwnerAssociatedToken(ownerAssociatedToken solana.PublicKey) *RecoverNested {
	inst.OwnerAssociatedToken = ownerAssociatedToken
	return inst
}

func (inst *RecoverNested) SetOwnerMint(ownerMint solana.PublicKey) *RecoverNested {
	inst.OwnerMint = ownerMint
	return inst
}

func (inst *RecoverNested) SetOwner(owner solana.PublicKey) *RecoverNested {
	inst.Owner = owner
	return inst
}

func (inst RecoverNested) Build() *Instruction {
	keys := []*solana.AccountMeta{
		{
			PublicKey:  inst.NestedAssociatedToken,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  inst.NestedMint,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  inst.DestinationAssociatedToken,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  inst.OwnerAssociatedToken,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  inst.OwnerMint,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  inst.Owner,
			IsSigner:   true,
			IsWritable: false,
		},
		{
			PublicKey:  solana.TokenProgramID,
			IsSigner:   false,
			IsWritable: false,
		},
	}

	inst.AccountMetaSlice = keys

	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint8(2), // Changed from bin.NoTypeIDDefaultID to match the index in InstructionImplDef
	}}
}

// ValidateAndBuild validates the instruction accounts.
// If there is a validation error, return the error.
// Otherwise, build and return the instruction.
func (inst RecoverNested) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *RecoverNested) Validate() error {
	if inst.NestedAssociatedToken.IsZero() {
		return errors.New("NestedAssociatedToken not set")
	}
	if inst.NestedMint.IsZero() {
		return errors.New("NestedMint not set")
	}
	if inst.DestinationAssociatedToken.IsZero() {
		return errors.New("DestinationAssociatedToken not set")
	}
	if inst.OwnerAssociatedToken.IsZero() {
		return errors.New("OwnerAssociatedToken not set")
	}
	if inst.OwnerMint.IsZero() {
		return errors.New("OwnerMint not set")
	}
	if inst.Owner.IsZero() {
		return errors.New("Owner not set")
	}
	return nil
}

func (inst *RecoverNested) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("RecoverNested")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=7").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("  nestedAssociatedToken", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("            nestedMint", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("destinationAssociatedToken", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(format.Meta("    ownerAssociatedToken", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(format.Meta("             ownerMint", inst.AccountMetaSlice.Get(4)))
						accountsBranch.Child(format.Meta("                 owner", inst.AccountMetaSlice.Get(5)))
						accountsBranch.Child(format.Meta("          tokenProgram", inst.AccountMetaSlice.Get(6)))
					})
				})
		})
}

func (inst RecoverNested) MarshalWithEncoder(encoder *bin.Encoder) error {
	// The RecoverNested instruction uses [2] as the instruction data
	return encoder.WriteBytes([]byte{2}, false)
}

func (inst *RecoverNested) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	return nil
}

func NewRecoverNestedInstruction(
	nestedAssociatedToken solana.PublicKey,
	nestedMint solana.PublicKey,
	destinationAssociatedToken solana.PublicKey,
	ownerAssociatedToken solana.PublicKey,
	ownerMint solana.PublicKey,
	owner solana.PublicKey,
) *RecoverNested {
	return NewRecoverNestedInstructionBuilder().
		SetNestedAssociatedToken(nestedAssociatedToken).
		SetNestedMint(nestedMint).
		SetDestinationAssociatedToken(destinationAssociatedToken).
		SetOwnerAssociatedToken(ownerAssociatedToken).
		SetOwnerMint(ownerMint).
		SetOwner(owner)
}
