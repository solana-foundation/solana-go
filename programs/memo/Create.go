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

package memo

import (
	"errors"
	"fmt"
	"unicode/utf8"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

type Create struct {
	// The memo message
	Message []byte

	// [0..N] = [SIGNER] Signers
	// ··········· Optional signers that approve the memo.
	// ··········· If zero signers are provided, the memo is unsigned.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewMemoInstructionBuilder creates a new `Memo` instruction builder.
func NewMemoInstructionBuilder() *Create {
	nd := &Create{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

// SetMessage sets the memo message
func (inst *Create) SetMessage(message []byte) *Create {
	inst.Message = message
	return inst
}

// SetSigner sets a single signer account (replaces any existing signers).
func (inst *Create) SetSigner(signer ag_solanago.PublicKey) *Create {
	inst.AccountMetaSlice = ag_solanago.AccountMetaSlice{
		ag_solanago.Meta(signer).SIGNER(),
	}
	return inst
}

// AddSigner appends a signer account to the list of signers.
func (inst *Create) AddSigner(signer ag_solanago.PublicKey) *Create {
	inst.AccountMetaSlice = append(inst.AccountMetaSlice, ag_solanago.Meta(signer).SIGNER())
	return inst
}

func (inst *Create) GetSigner() *ag_solanago.AccountMeta {
	if len(inst.AccountMetaSlice) == 0 {
		return nil
	}
	return inst.AccountMetaSlice[0]
}

func (inst Create) Build() *MemoInstruction {
	return &MemoInstruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.NoTypeIDDefaultID,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst Create) ValidateAndBuild() (*MemoInstruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Create) Validate() error {
	if len(inst.Message) == 0 {
		return errors.New("message not set")
	}
	if !utf8.Valid(inst.Message) {
		return errors.New("message is not valid UTF-8")
	}

	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%d] is not set", accIndex)
		}
	}
	return nil
}
func (inst *Create) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program("Memo", ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Create")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Message", inst.Message))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						for i, signer := range inst.AccountMetaSlice {
							accountsBranch.Child(ag_format.Meta(fmt.Sprintf("Signer[%d]", i), signer))
						}
					})
				})
		})
}

func (inst Create) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return encoder.WriteBytes(inst.Message, false)
}

func (inst *Create) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	var err error
	inst.Message, err = decoder.ReadBytes(decoder.Len())
	return err
}

// NewMemoInstruction declares a new Memo instruction with the provided parameters and accounts.
// Accepts zero or more signers. If no signers are provided, the memo is unsigned.
func NewMemoInstruction(
	// Parameters:
	message []byte,
	// Accounts:
	signers ...ag_solanago.PublicKey,
) *Create {
	builder := NewMemoInstructionBuilder().
		SetMessage(message)
	for _, signer := range signers {
		builder.AddSigner(signer)
	}
	return builder
}
