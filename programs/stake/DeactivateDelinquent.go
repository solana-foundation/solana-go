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

type DeactivateDelinquent struct {

	// [0] = [WRITE] Stake Account
	// ··········· Delegated stake account
	//
	// [1] = [] Delinquent Vote Account
	// ··········· Delinquent vote account for the delegated stake account
	//
	// [2] = [] Reference Vote Account
	// ··········· Reference vote account that has voted at least once in the last
	// ··········· Epoch::MINIMUM_SLOTS_PER_EPOCH slots
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *DeactivateDelinquent) Validate() error {
	// Check whether all accounts are set:
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}

func (inst *DeactivateDelinquent) SetStakeAccount(stakeAccount solana.PublicKey) *DeactivateDelinquent {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE()
	return inst
}
func (inst *DeactivateDelinquent) SetDelinquentVoteAccount(voteAccount solana.PublicKey) *DeactivateDelinquent {
	inst.AccountMetaSlice[1] = solana.Meta(voteAccount)
	return inst
}
func (inst *DeactivateDelinquent) SetReferenceVoteAccount(voteAccount solana.PublicKey) *DeactivateDelinquent {
	inst.AccountMetaSlice[2] = solana.Meta(voteAccount)
	return inst
}

func (inst *DeactivateDelinquent) GetStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}
func (inst *DeactivateDelinquent) GetDelinquentVoteAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}
func (inst *DeactivateDelinquent) GetReferenceVoteAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[2]
}

func (inst DeactivateDelinquent) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_DeactivateDelinquent, bin.LE),
	}}
}

func (inst *DeactivateDelinquent) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("DeactivateDelinquent")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("           StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta(" DelinquentVoteAccount", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("  ReferenceVoteAccount", inst.AccountMetaSlice.Get(2)))
					})
				})
		})
}

// NewDeactivateDelinquentInstructionBuilder creates a new `DeactivateDelinquent` instruction builder.
func NewDeactivateDelinquentInstructionBuilder() *DeactivateDelinquent {
	nd := &DeactivateDelinquent{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
	}
	return nd
}

// NewDeactivateDelinquentInstruction declares a new DeactivateDelinquent instruction with the provided parameters and accounts.
func NewDeactivateDelinquentInstruction(
	// Accounts:
	stakeAccount solana.PublicKey,
	delinquentVoteAccount solana.PublicKey,
	referenceVoteAccount solana.PublicKey,
) *DeactivateDelinquent {
	return NewDeactivateDelinquentInstructionBuilder().
		SetStakeAccount(stakeAccount).
		SetDelinquentVoteAccount(delinquentVoteAccount).
		SetReferenceVoteAccount(referenceVoteAccount)
}
