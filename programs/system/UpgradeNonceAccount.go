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

package system

import (
	"encoding/binary"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// One-time idempotent upgrade of legacy nonce versions in order to bump
// them out of chain blockhash domain.
type UpgradeNonceAccount struct {

	// [0] = [WRITE] NonceAccount
	// ··········· Nonce account
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewUpgradeNonceAccountInstructionBuilder creates a new `UpgradeNonceAccount` instruction builder.
func NewUpgradeNonceAccountInstructionBuilder() *UpgradeNonceAccount {
	nd := &UpgradeNonceAccount{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

// Nonce account
func (inst *UpgradeNonceAccount) SetNonceAccount(nonceAccount ag_solanago.PublicKey) *UpgradeNonceAccount {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(nonceAccount).WRITE()
	return inst
}

func (inst *UpgradeNonceAccount) GetNonceAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst UpgradeNonceAccount) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_UpgradeNonceAccount, binary.LittleEndian),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst UpgradeNonceAccount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *UpgradeNonceAccount) Validate() error {
	// Check whether all accounts are set:
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}

func (inst *UpgradeNonceAccount) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("UpgradeNonceAccount")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("Nonce", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (inst UpgradeNonceAccount) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return nil
}

func (inst *UpgradeNonceAccount) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	return nil
}

// NewUpgradeNonceAccountInstruction declares a new UpgradeNonceAccount instruction with the provided parameters and accounts.
func NewUpgradeNonceAccountInstruction(
	// Accounts:
	nonceAccount ag_solanago.PublicKey) *UpgradeNonceAccount {
	return NewUpgradeNonceAccountInstructionBuilder().
		SetNonceAccount(nonceAccount)
}
