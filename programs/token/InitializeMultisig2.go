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
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Like InitializeMultisig, but does not require the Rent sysvar to be provided.
type InitializeMultisig2 struct {
	// The number of signers (M) required to validate this multisignature account.
	M *uint8

	// [0] = [WRITE] account
	// ··········· The multisignature account to initialize.
	//
	// [1] = [SIGNER] signers
	// ··········· The signer accounts, must equal to N where 1 <= N <= 11.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *InitializeMultisig2) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(1)
	return nil
}

func (slice InitializeMultisig2) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

// NewInitializeMultisig2InstructionBuilder creates a new `InitializeMultisig2` instruction builder.
func NewInitializeMultisig2InstructionBuilder() *InitializeMultisig2 {
	nd := &InitializeMultisig2{
		Accounts: make(ag_solanago.AccountMetaSlice, 1),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

// SetM sets the "m" parameter.
// The number of signers (M) required to validate this multisignature account.
func (inst *InitializeMultisig2) SetM(m uint8) *InitializeMultisig2 {
	inst.M = &m
	return inst
}

// SetAccount sets the "account" account.
// The multisignature account to initialize.
func (inst *InitializeMultisig2) SetAccount(account ag_solanago.PublicKey) *InitializeMultisig2 {
	inst.Accounts[0] = ag_solanago.Meta(account).WRITE()
	return inst
}

// GetAccount gets the "account" account.
// The multisignature account to initialize.
func (inst *InitializeMultisig2) GetAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

// AddSigners adds the "signers" accounts.
// The signer accounts, must equal to N where 1 <= N <= 11.
func (inst *InitializeMultisig2) AddSigners(signers ...ag_solanago.PublicKey) *InitializeMultisig2 {
	for _, signer := range signers {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

func (inst InitializeMultisig2) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_InitializeMultisig2),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst InitializeMultisig2) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeMultisig2) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.M == nil {
			return errors.New("`M` parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.Accounts[0] == nil {
			return errors.New("accounts.Account is not set")
		}
		if len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
		if len(inst.Signers) > MAX_SIGNERS {
			return fmt.Errorf("too many signers; got %v, but max is 11", len(inst.Signers))
		}
	}
	return nil
}

func (inst *InitializeMultisig2) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("InitializeMultisig2")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("M", *inst.M))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("account", inst.Accounts[0]))

						signersBranch := accountsBranch.Child(fmt.Sprintf("signers[len=%v]", len(inst.Signers)))
						for i, v := range inst.Signers {
							if len(inst.Signers) > 9 && i < 10 {
								signersBranch.Child(ag_format.Meta(fmt.Sprintf(" [%v]", i), v))
							} else {
								signersBranch.Child(ag_format.Meta(fmt.Sprintf("[%v]", i), v))
							}
						}
					})
				})
		})
}

func (obj InitializeMultisig2) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `M` param:
	err = encoder.Encode(obj.M)
	if err != nil {
		return err
	}
	return nil
}
func (obj *InitializeMultisig2) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `M`:
	err = decoder.Decode(&obj.M)
	if err != nil {
		return err
	}
	return nil
}

// NewInitializeMultisig2Instruction declares a new InitializeMultisig2 instruction with the provided parameters and accounts.
func NewInitializeMultisig2Instruction(
	// Parameters:
	m uint8,
	// Accounts:
	account ag_solanago.PublicKey,
	signers []ag_solanago.PublicKey,
) *InitializeMultisig2 {
	return NewInitializeMultisig2InstructionBuilder().
		SetM(m).
		SetAccount(account).
		AddSigners(signers...)
}
