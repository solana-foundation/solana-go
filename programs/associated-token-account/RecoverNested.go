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
	"fmt"

	solana "github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	format "github.com/gagliardetto/solana-go/text/format"
	treeout "github.com/gagliardetto/treeout"
)

// RecoverNested transfers tokens from and closes a nested associated token account:
// an associated token account owned by an associated token account.
type RecoverNested struct {
	// The wallet that owns the owner ATA (owner of the owner ATA's mint).
	Wallet solana.PublicKey `bin:"-" borsh_skip:"true"`
	// The mint of the nested ATA (the tokens to recover).
	NestedMint solana.PublicKey `bin:"-" borsh_skip:"true"`
	// The mint of the owner ATA.
	OwnerMint solana.PublicKey `bin:"-" borsh_skip:"true"`

	// [0] = [WRITE] NestedAssociatedTokenAccount
	// ··········· Nested associated token account, must be owned by [3]
	//
	// [1] = [] NestedTokenMint
	// ··········· Token mint for the nested associated token account
	//
	// [2] = [WRITE] WalletAssociatedTokenAccount
	// ··········· Wallet's associated token account for the nested token mint
	//
	// [3] = [] OwnerAssociatedTokenAccount
	// ··········· Owner associated token account, owned by [5]
	//
	// [4] = [] OwnerTokenMint
	// ··········· Token mint for the owner associated token account
	//
	// [5] = [WRITE, SIGNER] Wallet
	// ··········· Wallet address for the owner associated token account
	//
	// [6] = [] TokenProgram
	// ··········· SPL token program ID
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewRecoverNestedInstructionBuilder creates a new `RecoverNested` instruction builder.
func NewRecoverNestedInstructionBuilder() *RecoverNested {
	return &RecoverNested{}
}

func (inst *RecoverNested) SetWallet(wallet solana.PublicKey) *RecoverNested {
	inst.Wallet = wallet
	return inst
}

func (inst *RecoverNested) SetNestedMint(nestedMint solana.PublicKey) *RecoverNested {
	inst.NestedMint = nestedMint
	return inst
}

func (inst *RecoverNested) SetOwnerMint(ownerMint solana.PublicKey) *RecoverNested {
	inst.OwnerMint = ownerMint
	return inst
}

func (inst *RecoverNested) SetAccounts(accounts []*solana.AccountMeta) error {
	inst.AccountMetaSlice = accounts
	if len(accounts) < 7 {
		return fmt.Errorf("insufficient accounts, RecoverNested requires at-least 7 accounts not %d", len(accounts))
	}
	inst.NestedMint = accounts[1].PublicKey
	inst.OwnerMint = accounts[4].PublicKey
	inst.Wallet = accounts[5].PublicKey
	return nil
}

func (inst RecoverNested) Build() *Instruction {
	// Derive the owner ATA (wallet's ATA for the owner mint).
	ownerATA, _, _ := solana.FindAssociatedTokenAddress(
		inst.Wallet,
		inst.OwnerMint,
	)

	// Derive the wallet's ATA for the nested mint (destination).
	walletATA, _, _ := solana.FindAssociatedTokenAddress(
		inst.Wallet,
		inst.NestedMint,
	)

	// Derive the nested ATA (the ATA owned by the owner ATA).
	nestedATA, _, _ := solana.FindAssociatedTokenAddress(
		ownerATA,
		inst.NestedMint,
	)

	keys := []*solana.AccountMeta{
		{
			PublicKey:  nestedATA,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  inst.NestedMint,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  walletATA,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  ownerATA,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  inst.OwnerMint,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  inst.Wallet,
			IsSigner:   true,
			IsWritable: true,
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
		TypeID: bin.TypeIDFromUint8(Instruction_RecoverNested),
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
	if inst.Wallet.IsZero() {
		return errors.New("wallet not set")
	}
	if inst.NestedMint.IsZero() {
		return errors.New("nested mint not set")
	}
	if inst.OwnerMint.IsZero() {
		return errors.New("owner mint not set")
	}
	_, _, err := solana.FindAssociatedTokenAddress(
		inst.Wallet,
		inst.OwnerMint,
	)
	if err != nil {
		return fmt.Errorf("error while FindAssociatedTokenAddress for owner: %w", err)
	}
	_, _, err = solana.FindAssociatedTokenAddress(
		inst.Wallet,
		inst.NestedMint,
	)
	if err != nil {
		return fmt.Errorf("error while FindAssociatedTokenAddress for nested: %w", err)
	}
	return nil
}

func (inst *RecoverNested) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("RecoverNested")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch treeout.Branches) {})

					instructionBranch.Child("Accounts[len=7]").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("nestedAssociatedTokenAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("             nestedTokenMint", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("walletAssociatedTokenAccount", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(format.Meta(" ownerAssociatedTokenAccount", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(format.Meta("              ownerTokenMint", inst.AccountMetaSlice.Get(4)))
						accountsBranch.Child(format.Meta("                      wallet", inst.AccountMetaSlice.Get(5)))
						accountsBranch.Child(format.Meta("                tokenProgram", inst.AccountMetaSlice.Get(6)))
					})
				})
		})
}

func (inst RecoverNested) MarshalWithEncoder(encoder *bin.Encoder) error {
	return nil
}

func (inst *RecoverNested) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	return nil
}

// NewRecoverNestedInstruction creates a new RecoverNested instruction.
func NewRecoverNestedInstruction(
	wallet solana.PublicKey,
	nestedMint solana.PublicKey,
	ownerMint solana.PublicKey,
) *RecoverNested {
	return NewRecoverNestedInstructionBuilder().
		SetWallet(wallet).
		SetNestedMint(nestedMint).
		SetOwnerMint(ownerMint)
}

func (inst *RecoverNested) GetNestedAssociatedTokenAccountAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

func (inst *RecoverNested) GetNestedTokenMintAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

func (inst *RecoverNested) GetWalletAssociatedTokenAccountAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

func (inst *RecoverNested) GetOwnerAssociatedTokenAccountAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

func (inst *RecoverNested) GetOwnerTokenMintAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice.Get(4)
}

func (inst *RecoverNested) GetWalletAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice.Get(5)
}

func (inst *RecoverNested) GetTokenProgramAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice.Get(6)
}
