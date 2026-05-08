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

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/treeout"
	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/text/format"
)

type DelegateStake struct {
	// [0] = [WRITE SIGNER] StakeAccount
	// ··········· Stake account getting initialized
	//
	// [1] = [] Vote Account
	// ··········· The validator vote account being delegated to
	//
	// [2] = [] Clock Sysvar
	// ··········· The Clock Sysvar Account
	//
	// [3] = [] Stake History Sysvar
	// ··········· The Stake History Sysvar Account
	//
	// [4] = [] Stake Config Account
	// ··········· The Stake Config Account
	//
	// [5] = [WRITE SIGNER] Stake Authoriy
	// ··········· The Stake Authority
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *DelegateStake) Validate() error {
	// Check whether all accounts are set:
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}
func (inst *DelegateStake) SetStakeAccount(stakeAccount solana.PublicKey) *DelegateStake {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE().SIGNER()
	return inst
}
func (inst *DelegateStake) SetVoteAccount(voteAcc solana.PublicKey) *DelegateStake {
	inst.AccountMetaSlice[1] = solana.Meta(voteAcc)
	return inst
}
func (inst *DelegateStake) SetClockSysvar(clockSysVarAcc solana.PublicKey) *DelegateStake {
	inst.AccountMetaSlice[2] = solana.Meta(clockSysVarAcc)
	return inst
}
func (inst *DelegateStake) SetStakeHistorySysvar(stakeHistorySysVarAcc solana.PublicKey) *DelegateStake {
	inst.AccountMetaSlice[3] = solana.Meta(stakeHistorySysVarAcc)
	return inst
}
func (inst *DelegateStake) SetConfigAccount(stakeConfigAcc solana.PublicKey) *DelegateStake {
	inst.AccountMetaSlice[4] = solana.Meta(stakeConfigAcc)
	return inst
}
func (inst *DelegateStake) SetStakeAuthority(stakeAuthority solana.PublicKey) *DelegateStake {
	inst.AccountMetaSlice[5] = solana.Meta(stakeAuthority).WRITE().SIGNER()
	return inst
}
func (inst *DelegateStake) GetStakeAccount() *solana.AccountMeta { return inst.AccountMetaSlice[0] }
func (inst *DelegateStake) GetVoteAccount() *solana.AccountMeta  { return inst.AccountMetaSlice[1] }
func (inst *DelegateStake) GetClockSysvar() *solana.AccountMeta  { return inst.AccountMetaSlice[2] }
func (inst *DelegateStake) GetStakeHistorySysvar() *solana.AccountMeta {
	return inst.AccountMetaSlice[3]
}
func (inst *DelegateStake) GetConfigAccount() *solana.AccountMeta  { return inst.AccountMetaSlice[4] }
func (inst *DelegateStake) GetStakeAuthority() *solana.AccountMeta { return inst.AccountMetaSlice[5] }

func (inst DelegateStake) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_DelegateStake, bin.LE),
	}}
}

func (inst *DelegateStake) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("DelegateStake")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("           StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("           VoteAccount", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("           ClockSysvar", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(format.Meta("           StakeHistorySysvar", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(format.Meta("           StakeConfigAccount", inst.AccountMetaSlice.Get(4)))
						accountsBranch.Child(format.Meta("           StakeAuthoriy", inst.AccountMetaSlice.Get(5)))
					})
				})
		})
}

// NewDelegateStakeInstructionBuilder creates a new `DelegateStake` instruction builder.
func NewDelegateStakeInstructionBuilder() *DelegateStake {
	nd := &DelegateStake{
		AccountMetaSlice: make(solana.AccountMetaSlice, 6),
	}
	return nd
}

// NewDelegateStakeInstruction declares a new DelegateStake instruction with the provided parameters and accounts.
func NewDelegateStakeInstruction(
	// Accounts:
	validatorVoteAccount solana.PublicKey,
	stakeAuthority solana.PublicKey,
	stakeAccount solana.PublicKey,
) *DelegateStake {
	return NewDelegateStakeInstructionBuilder().
		SetStakeAccount(stakeAccount).
		SetVoteAccount(validatorVoteAccount).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetStakeHistorySysvar(solana.SysVarStakeHistoryPubkey).
		SetConfigAccount(solana.SysVarStakeConfigPubkey).
		SetStakeAuthority(stakeAuthority)
}
