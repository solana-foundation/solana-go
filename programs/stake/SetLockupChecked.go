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

type SetLockupChecked struct {
	// Lockup checked arguments
	LockupCheckedArgs *LockupCheckedArgs

	// [0] = [WRITE] Stake Account
	// ··········· Initialized stake account
	//
	// [1] = [SIGNER] Authority
	// ··········· Lockup authority or withdraw authority
	//
	// OPTIONAL:
	// [2] = [SIGNER] New Lockup Authority
	// ··········· New lockup authority
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *SetLockupChecked) Validate() error {
	{
		if inst.LockupCheckedArgs == nil {
			return errors.New("lockup checked args parameter is not set")
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

func (inst *SetLockupChecked) SetStakeAccount(stakeAccount solana.PublicKey) *SetLockupChecked {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE()
	return inst
}
func (inst *SetLockupChecked) SetAuthority(authority solana.PublicKey) *SetLockupChecked {
	inst.AccountMetaSlice[1] = solana.Meta(authority).SIGNER()
	return inst
}
func (inst *SetLockupChecked) SetNewLockupAuthority(newLockupAuthority solana.PublicKey) *SetLockupChecked {
	inst.AccountMetaSlice = append(inst.AccountMetaSlice, solana.Meta(newLockupAuthority).SIGNER())
	return inst
}

func (inst *SetLockupChecked) GetStakeAccount() *solana.AccountMeta { return inst.AccountMetaSlice[0] }
func (inst *SetLockupChecked) GetAuthority() *solana.AccountMeta    { return inst.AccountMetaSlice[1] }
func (inst *SetLockupChecked) GetNewLockupAuthority() *solana.AccountMeta {
	if len(inst.AccountMetaSlice) > 2 {
		return inst.AccountMetaSlice[2]
	}
	return nil
}

func (inst *SetLockupChecked) SetLockupCheckedArgs(args LockupCheckedArgs) *SetLockupChecked {
	inst.LockupCheckedArgs = &args
	return inst
}

func (inst *SetLockupChecked) SetLockupTimestamp(unixTimestamp int64) *SetLockupChecked {
	inst.LockupCheckedArgs.UnixTimestamp = &unixTimestamp
	return inst
}

func (inst *SetLockupChecked) SetLockupEpoch(epoch uint64) *SetLockupChecked {
	inst.LockupCheckedArgs.Epoch = &epoch
	return inst
}

func (inst *SetLockupChecked) UnmarshalWithDecoder(dec *bin.Decoder) error {
	{
		err := dec.Decode(&inst.LockupCheckedArgs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst *SetLockupChecked) MarshalWithEncoder(encoder *bin.Encoder) error {
	{
		err := encoder.Encode(*inst.LockupCheckedArgs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst SetLockupChecked) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_SetLockupChecked, bin.LE),
	}}
}

func (inst *SetLockupChecked) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("SetLockupChecked")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child("LockupCheckedArgs").ParentFunc(func(lockupBranch treeout.Branches) {
							lockupBranch.Child(format.Param("UnixTimestamp", inst.LockupCheckedArgs.UnixTimestamp))
							lockupBranch.Child(format.Param("        Epoch", inst.LockupCheckedArgs.Epoch))
						})
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("       StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("          Authority", inst.AccountMetaSlice.Get(1)))
						if len(inst.AccountMetaSlice) > 2 {
							accountsBranch.Child(format.Meta("NewLockupAuthority", inst.AccountMetaSlice.Get(2)))
						}
					})
				})
		})
}

// NewSetLockupCheckedInstructionBuilder creates a new `SetLockupChecked` instruction builder.
func NewSetLockupCheckedInstructionBuilder() *SetLockupChecked {
	nd := &SetLockupChecked{
		AccountMetaSlice:  make(solana.AccountMetaSlice, 2),
		LockupCheckedArgs: &LockupCheckedArgs{},
	}
	return nd
}

// NewSetLockupCheckedInstruction declares a new SetLockupChecked instruction with the provided parameters and accounts.
func NewSetLockupCheckedInstruction(
	// Accounts:
	stakeAccount solana.PublicKey,
	authority solana.PublicKey,
) *SetLockupChecked {
	return NewSetLockupCheckedInstructionBuilder().
		SetStakeAccount(stakeAccount).
		SetAuthority(authority)
}
