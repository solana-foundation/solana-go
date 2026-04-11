// Copyright 2024 github.com/cordialsys
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stake

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

type Deactivate struct {

	// [0] = [WRITE] Stake Account
	// ··········· Delegated stake account to be deactivated
	//
	// [1] = [] Clock Sysvar
	// ··········· The Clock Sysvar Account
	//
	// [2] = [SIGNER] Stake Authority
	// ··········· Stake authority
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *Deactivate) Validate() error {

	// Check whether all accounts are set:
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}
func (inst *Deactivate) SetStakeAccount(stakeAccount solana.PublicKey) *Deactivate {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE()
	return inst
}
func (inst *Deactivate) SetClockSysvar(clockSysvar solana.PublicKey) *Deactivate {
	inst.AccountMetaSlice[1] = solana.Meta(clockSysvar)
	return inst
}
func (inst *Deactivate) SetStakeAuthority(stakeAuthority solana.PublicKey) *Deactivate {
	inst.AccountMetaSlice[2] = solana.Meta(stakeAuthority).SIGNER()
	return inst
}

func (inst *Deactivate) GetStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}
func (inst *Deactivate) GetClockSysvar() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}
func (inst *Deactivate) GetStakeAuthority() *solana.AccountMeta {
	return inst.AccountMetaSlice[2]
}

func (inst Deactivate) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_Deactivate, bin.LE),
	}}
}

func (inst *Deactivate) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("Deactivate")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("            StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("             ClockSysvar", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("           StakeAuthoriy", inst.AccountMetaSlice.Get(2)))
					})
				})
		})
}

// NewDeactivateInstructionBuilder creates a new `Deactivate` instruction builder.
func NewDeactivateInstructionBuilder() *Deactivate {
	nd := &Deactivate{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
	}
	return nd
}

// NewDeactivateInstruction declares a new Deactivate instruction with the provided parameters and accounts.
func NewDeactivateInstruction(
	// Params:
	// Accounts:
	stakeAccount solana.PublicKey,
	stakeAuthority solana.PublicKey,
) *Deactivate {
	return NewDeactivateInstructionBuilder().
		SetStakeAccount(stakeAccount).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetStakeAuthority(stakeAuthority)
}
