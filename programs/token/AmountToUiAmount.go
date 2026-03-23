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

// Convert an Amount of tokens to a UiAmount string, using the given mint.
// In this version of the program, the mint can only specify the number of decimals.
//
// Return data can be fetched using sol_get_return_data and deserialized
// with String::from_utf8.
type AmountToUiAmount struct {
	// The amount of tokens to reformat.
	Amount *uint64

	// [0] = [] mint
	// ··········· The mint to calculate for.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewAmountToUiAmountInstructionBuilder() *AmountToUiAmount {
	nd := &AmountToUiAmount{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

func (inst *AmountToUiAmount) SetAmount(amount uint64) *AmountToUiAmount {
	inst.Amount = &amount
	return inst
}

func (inst *AmountToUiAmount) SetMintAccount(mint ag_solanago.PublicKey) *AmountToUiAmount {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(mint)
	return inst
}

func (inst *AmountToUiAmount) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst AmountToUiAmount) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_AmountToUiAmount),
	}}
}

func (inst AmountToUiAmount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *AmountToUiAmount) Validate() error {
	if inst.Amount == nil {
		return errors.New("Amount parameter is not set")
	}
	if inst.AccountMetaSlice[0] == nil {
		return errors.New("accounts.Mint is not set")
	}
	return nil
}

func (inst *AmountToUiAmount) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("AmountToUiAmount")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Amount", *inst.Amount))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj AmountToUiAmount) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	err = encoder.Encode(obj.Amount)
	if err != nil {
		return err
	}
	return nil
}

func (obj *AmountToUiAmount) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	err = decoder.Decode(&obj.Amount)
	if err != nil {
		return err
	}
	return nil
}

func NewAmountToUiAmountInstruction(
	amount uint64,
	mint ag_solanago.PublicKey,
) *AmountToUiAmount {
	return NewAmountToUiAmountInstructionBuilder().
		SetAmount(amount).
		SetMintAccount(mint)
}
