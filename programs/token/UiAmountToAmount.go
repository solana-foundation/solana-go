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

// Convert a UiAmount of tokens to a little-endian u64 raw Amount, using the given mint.
// In this version of the program, the mint can only specify the number of decimals.
//
// Return data can be fetched using sol_get_return_data and deserializing
// the return data as a little-endian u64.
type UiAmountToAmount struct {
	// The ui_amount of tokens to reformat.
	UiAmount *string

	// [0] = [] mint
	// ··········· The mint to calculate for.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewUiAmountToAmountInstructionBuilder() *UiAmountToAmount {
	nd := &UiAmountToAmount{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

func (inst *UiAmountToAmount) SetUiAmount(uiAmount string) *UiAmountToAmount {
	inst.UiAmount = &uiAmount
	return inst
}

func (inst *UiAmountToAmount) SetMintAccount(mint ag_solanago.PublicKey) *UiAmountToAmount {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(mint)
	return inst
}

func (inst *UiAmountToAmount) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst UiAmountToAmount) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_UiAmountToAmount),
	}}
}

func (inst UiAmountToAmount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *UiAmountToAmount) Validate() error {
	if inst.UiAmount == nil {
		return errors.New("UiAmount parameter is not set")
	}
	if inst.AccountMetaSlice[0] == nil {
		return errors.New("accounts.Mint is not set")
	}
	return nil
}

func (inst *UiAmountToAmount) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("UiAmountToAmount")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("UiAmount", *inst.UiAmount))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj UiAmountToAmount) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	_, err = encoder.Write([]byte(*obj.UiAmount))
	if err != nil {
		return err
	}
	return nil
}

func (obj *UiAmountToAmount) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	data, err := decoder.ReadNBytes(decoder.Remaining())
	if err != nil {
		return err
	}
	s := string(data)
	obj.UiAmount = &s
	return nil
}

func NewUiAmountToAmountInstruction(
	uiAmount string,
	mint ag_solanago.PublicKey,
) *UiAmountToAmount {
	return NewUiAmountToAmountInstructionBuilder().
		SetUiAmount(uiAmount).
		SetMintAccount(mint)
}
