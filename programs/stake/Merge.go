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

type Merge struct {

	// [0] = [WRITE] Destination Stake Account
	// ··········· Destination stake account to merge into
	//
	// [1] = [WRITE] Source Stake Account
	// ··········· Source stake account to merge from (will be drained)
	//
	// [2] = [] Clock Sysvar
	// ··········· The Clock Sysvar Account
	//
	// [3] = [] Stake History Sysvar
	// ··········· The Stake History Sysvar Account
	//
	// [4] = [SIGNER] Stake Authority
	// ··········· Stake authority
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *Merge) Validate() error {
	// Check whether all accounts are set:
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}

func (inst *Merge) SetDestinationStakeAccount(dest solana.PublicKey) *Merge {
	inst.AccountMetaSlice[0] = solana.Meta(dest).WRITE()
	return inst
}
func (inst *Merge) SetSourceStakeAccount(source solana.PublicKey) *Merge {
	inst.AccountMetaSlice[1] = solana.Meta(source).WRITE()
	return inst
}
func (inst *Merge) SetClockSysvar(clockSysvar solana.PublicKey) *Merge {
	inst.AccountMetaSlice[2] = solana.Meta(clockSysvar)
	return inst
}
func (inst *Merge) SetStakeHistorySysvar(stakeHistorySysvar solana.PublicKey) *Merge {
	inst.AccountMetaSlice[3] = solana.Meta(stakeHistorySysvar)
	return inst
}
func (inst *Merge) SetStakeAuthority(stakeAuthority solana.PublicKey) *Merge {
	inst.AccountMetaSlice[4] = solana.Meta(stakeAuthority).SIGNER()
	return inst
}

func (inst *Merge) GetDestinationStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}
func (inst *Merge) GetSourceStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}
func (inst *Merge) GetClockSysvar() *solana.AccountMeta        { return inst.AccountMetaSlice[2] }
func (inst *Merge) GetStakeHistorySysvar() *solana.AccountMeta { return inst.AccountMetaSlice[3] }
func (inst *Merge) GetStakeAuthority() *solana.AccountMeta     { return inst.AccountMetaSlice[4] }

func (inst Merge) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_Merge, bin.LE),
	}}
}

func (inst *Merge) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("Merge")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("  DestinationStakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("       SourceStakeAccount", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("              ClockSysvar", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(format.Meta("       StakeHistorySysvar", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(format.Meta("          StakeAuthority", inst.AccountMetaSlice.Get(4)))
					})
				})
		})
}

// NewMergeInstructionBuilder creates a new `Merge` instruction builder.
func NewMergeInstructionBuilder() *Merge {
	nd := &Merge{
		AccountMetaSlice: make(solana.AccountMetaSlice, 5),
	}
	return nd
}

// NewMergeInstruction declares a new Merge instruction with the provided parameters and accounts.
func NewMergeInstruction(
	// Accounts:
	destinationStakeAccount solana.PublicKey,
	sourceStakeAccount solana.PublicKey,
	stakeAuthority solana.PublicKey,
) *Merge {
	return NewMergeInstructionBuilder().
		SetDestinationStakeAccount(destinationStakeAccount).
		SetSourceStakeAccount(sourceStakeAccount).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetStakeHistorySysvar(solana.SysVarStakeHistoryPubkey).
		SetStakeAuthority(stakeAuthority)
}
