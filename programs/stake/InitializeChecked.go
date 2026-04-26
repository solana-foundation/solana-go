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

type InitializeChecked struct {

	// [0] = [WRITE] Stake Account
	// ··········· Uninitialized stake account
	//
	// [1] = [] Rent Sysvar
	// ··········· Rent sysvar
	//
	// [2] = [] Stake Authority
	// ··········· Stake authority
	//
	// [3] = [SIGNER] Withdraw Authority
	// ··········· Withdraw authority
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *InitializeChecked) Validate() error {
	// Check whether all accounts are set:
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}

func (inst *InitializeChecked) SetStakeAccount(stakeAccount solana.PublicKey) *InitializeChecked {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE()
	return inst
}
func (inst *InitializeChecked) SetRentSysvar(rentSysvar solana.PublicKey) *InitializeChecked {
	inst.AccountMetaSlice[1] = solana.Meta(rentSysvar)
	return inst
}
func (inst *InitializeChecked) SetStakeAuthority(stakeAuthority solana.PublicKey) *InitializeChecked {
	inst.AccountMetaSlice[2] = solana.Meta(stakeAuthority)
	return inst
}
func (inst *InitializeChecked) SetWithdrawAuthority(withdrawAuthority solana.PublicKey) *InitializeChecked {
	inst.AccountMetaSlice[3] = solana.Meta(withdrawAuthority).SIGNER()
	return inst
}

func (inst *InitializeChecked) GetStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}
func (inst *InitializeChecked) GetRentSysvar() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}
func (inst *InitializeChecked) GetStakeAuthority() *solana.AccountMeta {
	return inst.AccountMetaSlice[2]
}
func (inst *InitializeChecked) GetWithdrawAuthority() *solana.AccountMeta {
	return inst.AccountMetaSlice[3]
}

func (inst InitializeChecked) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_InitializeChecked, bin.LE),
	}}
}

func (inst *InitializeChecked) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("InitializeChecked")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("           StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("            RentSysvar", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("        StakeAuthority", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(format.Meta("     WithdrawAuthority", inst.AccountMetaSlice.Get(3)))
					})
				})
		})
}

// NewInitializeCheckedInstructionBuilder creates a new `InitializeChecked` instruction builder.
func NewInitializeCheckedInstructionBuilder() *InitializeChecked {
	nd := &InitializeChecked{
		AccountMetaSlice: make(solana.AccountMetaSlice, 4),
	}
	return nd
}

// NewInitializeCheckedInstruction declares a new InitializeChecked instruction with the provided parameters and accounts.
func NewInitializeCheckedInstruction(
	// Accounts:
	stakeAccount solana.PublicKey,
	stakeAuthority solana.PublicKey,
	withdrawAuthority solana.PublicKey,
) *InitializeChecked {
	return NewInitializeCheckedInstructionBuilder().
		SetStakeAccount(stakeAccount).
		SetRentSysvar(solana.SysVarRentPubkey).
		SetStakeAuthority(stakeAuthority).
		SetWithdrawAuthority(withdrawAuthority)
}
