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

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

type SetLockup struct {
	// Lockup arguments
	LockupArgs *LockupArgs

	// [0] = [WRITE] Stake Account
	// ··········· Initialized stake account
	//
	// [1] = [SIGNER] Authority
	// ··········· Lockup authority or withdraw authority
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *SetLockup) Validate() error {
	{
		if inst.LockupArgs == nil {
			return errors.New("lockup args parameter is not set")
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

func (inst *SetLockup) SetStakeAccount(stakeAccount solana.PublicKey) *SetLockup {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE()
	return inst
}
func (inst *SetLockup) SetAuthority(authority solana.PublicKey) *SetLockup {
	inst.AccountMetaSlice[1] = solana.Meta(authority).SIGNER()
	return inst
}

func (inst *SetLockup) GetStakeAccount() *solana.AccountMeta { return inst.AccountMetaSlice[0] }
func (inst *SetLockup) GetAuthority() *solana.AccountMeta    { return inst.AccountMetaSlice[1] }

func (inst *SetLockup) SetLockupArgs(args LockupArgs) *SetLockup {
	inst.LockupArgs = &args
	return inst
}

func (inst *SetLockup) SetLockupTimestamp(unixTimestamp int64) *SetLockup {
	inst.LockupArgs.UnixTimestamp = &unixTimestamp
	return inst
}

func (inst *SetLockup) SetLockupEpoch(epoch uint64) *SetLockup {
	inst.LockupArgs.Epoch = &epoch
	return inst
}

func (inst *SetLockup) SetCustodian(custodian solana.PublicKey) *SetLockup {
	inst.LockupArgs.Custodian = &custodian
	return inst
}

func (inst SetLockup) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_SetLockup, bin.LE),
	}}
}

func (inst *SetLockup) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("SetLockup")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child("LockupArgs").ParentFunc(func(lockupBranch treeout.Branches) {
							lockupBranch.Child(format.Param("UnixTimestamp", inst.LockupArgs.UnixTimestamp))
							lockupBranch.Child(format.Param("        Epoch", inst.LockupArgs.Epoch))
							if inst.LockupArgs.Custodian != nil {
								lockupBranch.Child(format.Account("    Custodian", *inst.LockupArgs.Custodian))
							}
						})
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("         StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("            Authority", inst.AccountMetaSlice.Get(1)))
					})
				})
		})
}

// NewSetLockupInstructionBuilder creates a new `SetLockup` instruction builder.
func NewSetLockupInstructionBuilder() *SetLockup {
	nd := &SetLockup{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
		LockupArgs:       &LockupArgs{},
	}
	return nd
}

// NewSetLockupInstruction declares a new SetLockup instruction with the provided parameters and accounts.
func NewSetLockupInstruction(
	// Accounts:
	stakeAccount solana.PublicKey,
	authority solana.PublicKey,
) *SetLockup {
	return NewSetLockupInstructionBuilder().
		SetStakeAccount(stakeAccount).
		SetAuthority(authority)
}
