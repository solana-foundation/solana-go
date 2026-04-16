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

// Gets the required size of an account for the given mint as a little-endian u64.
// Return data can be fetched using sol_get_return_data and deserializing
// the return data as a little-endian u64.
type GetAccountDataSize struct {
	// [0] = [] mint
	// ··········· The mint to calculate for.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewGetAccountDataSizeInstructionBuilder() *GetAccountDataSize {
	nd := &GetAccountDataSize{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

func (inst *GetAccountDataSize) SetMintAccount(mint ag_solanago.PublicKey) *GetAccountDataSize {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(mint)
	return inst
}

func (inst *GetAccountDataSize) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst GetAccountDataSize) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_GetAccountDataSize),
	}}
}

func (inst GetAccountDataSize) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *GetAccountDataSize) Validate() error {
	if inst.AccountMetaSlice[0] == nil {
		return errors.New("accounts.Mint is not set")
	}
	return nil
}

func (inst *GetAccountDataSize) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("GetAccountDataSize")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj GetAccountDataSize) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *GetAccountDataSize) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func NewGetAccountDataSizeInstruction(
	mint ag_solanago.PublicKey,
) *GetAccountDataSize {
	return NewGetAccountDataSizeInstructionBuilder().
		SetMintAccount(mint)
}
