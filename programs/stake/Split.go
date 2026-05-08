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
	"errors"
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/treeout"
	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/text/format"
)

type Split struct {
	// Amount to split to new stake account
	Lamports *uint64
	// [0] = [WRITE] Stake Account
	// ··········· Stake account to be split; must be in the Initialized or Stake state
	//
	// [1] = [WRITE] New Stake Account
	// ··········· Uninitialized stake account that will take the split-off amount
	//
	// [2] = [SIGNER] Stake Authority
	// ··········· Stake authority
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *Split) Validate() error {
	{
		if inst.Lamports == nil {
			return errors.New("lamports parameter is not set")
		}
	}
	// Check whether all accounts are set:
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}
func (inst *Split) SetStakeAccount(stakeAccount solana.PublicKey) *Split {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE()
	return inst
}
func (inst *Split) SetNewStakeAccount(voteAcc solana.PublicKey) *Split {
	inst.AccountMetaSlice[1] = solana.Meta(voteAcc).WRITE()
	return inst
}

func (inst *Split) SetStakeAuthority(stakeAuthority solana.PublicKey) *Split {
	inst.AccountMetaSlice[2] = solana.Meta(stakeAuthority).SIGNER()
	return inst
}

func (inst *Split) GetStakeAccount() *solana.AccountMeta    { return inst.AccountMetaSlice[0] }
func (inst *Split) GetNewStakeAccount() *solana.AccountMeta { return inst.AccountMetaSlice[1] }
func (inst *Split) GetStakeAuthority() *solana.AccountMeta  { return inst.AccountMetaSlice[2] }

func (inst *Split) SetLamports(lamports uint64) *Split {
	inst.Lamports = &lamports
	return inst
}

func (inst *Split) UnmarshalWithDecoder(dec *bin.Decoder) error {
	{
		err := dec.Decode(&inst.Lamports)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst *Split) MarshalWithEncoder(encoder *bin.Encoder) error {
	{
		err := encoder.Encode(*inst.Lamports)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst Split) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_Split, bin.LE),
	}}
}

func (inst *Split) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("Split")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("Lamports", inst.Lamports))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("              StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("           NewStakeAccount", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("             StakeAuthoriy", inst.AccountMetaSlice.Get(2)))
					})
				})
		})
}

// NewSplitInstructionBuilder creates a new `Split` instruction builder.
func NewSplitInstructionBuilder() *Split {
	nd := &Split{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
	}
	return nd
}

// NewSplitInstruction declares a new Split instruction with the provided parameters and accounts.
func NewSplitInstruction(
	// Params:
	lamports uint64,
	// Accounts:
	stakeAccount solana.PublicKey,
	newStakeAccount solana.PublicKey,
	stakeAuthority solana.PublicKey,
) *Split {
	return NewSplitInstructionBuilder().
		SetLamports(lamports).
		SetStakeAccount(stakeAccount).
		SetNewStakeAccount(newStakeAccount).
		SetStakeAuthority(stakeAuthority)
}
