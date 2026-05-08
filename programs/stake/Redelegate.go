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

// Deprecated: Redelegate was deprecated in Solana v2.1.0.
type Redelegate struct {

	// [0] = [WRITE] Stake Account
	// ··········· Delegated stake account
	//
	// [1] = [WRITE] Uninitialized Stake Account
	// ··········· Uninitialized stake account
	//
	// [2] = [] Vote Account
	// ··········· Vote account to redelegate to
	//
	// [3] = [] Unused Account
	// ··········· Unused (formerly stake config)
	//
	// [4] = [SIGNER] Stake Authority
	// ··········· Stake authority
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *Redelegate) Validate() error {
	// Check whether all accounts are set:
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}

func (inst *Redelegate) SetStakeAccount(stakeAccount solana.PublicKey) *Redelegate {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE()
	return inst
}
func (inst *Redelegate) SetUninitializedStakeAccount(uninitStakeAccount solana.PublicKey) *Redelegate {
	inst.AccountMetaSlice[1] = solana.Meta(uninitStakeAccount).WRITE()
	return inst
}
func (inst *Redelegate) SetVoteAccount(voteAccount solana.PublicKey) *Redelegate {
	inst.AccountMetaSlice[2] = solana.Meta(voteAccount)
	return inst
}
func (inst *Redelegate) SetUnusedAccount(unused solana.PublicKey) *Redelegate {
	inst.AccountMetaSlice[3] = solana.Meta(unused)
	return inst
}
func (inst *Redelegate) SetStakeAuthority(stakeAuthority solana.PublicKey) *Redelegate {
	inst.AccountMetaSlice[4] = solana.Meta(stakeAuthority).SIGNER()
	return inst
}

func (inst *Redelegate) GetStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}
func (inst *Redelegate) GetUninitializedStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}
func (inst *Redelegate) GetVoteAccount() *solana.AccountMeta   { return inst.AccountMetaSlice[2] }
func (inst *Redelegate) GetUnusedAccount() *solana.AccountMeta { return inst.AccountMetaSlice[3] }
func (inst *Redelegate) GetStakeAuthority() *solana.AccountMeta {
	return inst.AccountMetaSlice[4]
}

func (inst Redelegate) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_Redelegate, bin.LE),
	}}
}

func (inst *Redelegate) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("Redelegate")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("              StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("UninitializedStakeAccount", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("               VoteAccount", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(format.Meta("             UnusedAccount", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(format.Meta("            StakeAuthority", inst.AccountMetaSlice.Get(4)))
					})
				})
		})
}

// NewRedelegateInstructionBuilder creates a new `Redelegate` instruction builder.
//
// Deprecated: Redelegate was deprecated in Solana v2.1.0.
func NewRedelegateInstructionBuilder() *Redelegate {
	nd := &Redelegate{
		AccountMetaSlice: make(solana.AccountMetaSlice, 5),
	}
	return nd
}

// NewRedelegateInstruction declares a new Redelegate instruction with the provided parameters and accounts.
//
// Deprecated: Redelegate was deprecated in Solana v2.1.0.
func NewRedelegateInstruction(
	// Accounts:
	stakeAccount solana.PublicKey,
	uninitializedStakeAccount solana.PublicKey,
	voteAccount solana.PublicKey,
	stakeAuthority solana.PublicKey,
) *Redelegate {
	return NewRedelegateInstructionBuilder().
		SetStakeAccount(stakeAccount).
		SetUninitializedStakeAccount(uninitializedStakeAccount).
		SetVoteAccount(voteAccount).
		SetUnusedAccount(solana.SysVarStakeConfigPubkey).
		SetStakeAuthority(stakeAuthority)
}
