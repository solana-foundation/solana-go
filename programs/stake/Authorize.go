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
	"encoding/binary"
	"errors"
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/treeout"
	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/text/format"
)

type Authorize struct {
	// New authority public key
	NewAuthorized *solana.PublicKey

	// Type of authority to update (Staker or Withdrawer)
	StakeAuthorize *StakeAuthorize

	// [0] = [WRITE] Stake Account
	// ··········· Stake account to be updated
	//
	// [1] = [] Clock Sysvar
	// ··········· The Clock Sysvar Account
	//
	// [2] = [SIGNER] Authority
	// ··········· Current stake or withdraw authority
	//
	// OPTIONAL:
	// [3] = [SIGNER] Lockup Authority
	// ··········· Lockup authority, if updating withdrawer before lockup expiration
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *Authorize) Validate() error {
	{
		if inst.NewAuthorized == nil {
			return errors.New("new authorized parameter is not set")
		}
	}
	{
		if inst.StakeAuthorize == nil {
			return errors.New("stake authorize parameter is not set")
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

func (inst *Authorize) SetStakeAccount(stakeAccount solana.PublicKey) *Authorize {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE()
	return inst
}
func (inst *Authorize) SetClockSysvar(clockSysvar solana.PublicKey) *Authorize {
	inst.AccountMetaSlice[1] = solana.Meta(clockSysvar)
	return inst
}
func (inst *Authorize) SetAuthority(authority solana.PublicKey) *Authorize {
	inst.AccountMetaSlice[2] = solana.Meta(authority).SIGNER()
	return inst
}
func (inst *Authorize) SetLockupAuthority(lockupAuthority solana.PublicKey) *Authorize {
	inst.AccountMetaSlice = append(inst.AccountMetaSlice, solana.Meta(lockupAuthority).SIGNER())
	return inst
}

func (inst *Authorize) GetStakeAccount() *solana.AccountMeta { return inst.AccountMetaSlice[0] }
func (inst *Authorize) GetClockSysvar() *solana.AccountMeta  { return inst.AccountMetaSlice[1] }
func (inst *Authorize) GetAuthority() *solana.AccountMeta    { return inst.AccountMetaSlice[2] }
func (inst *Authorize) GetLockupAuthority() *solana.AccountMeta {
	if len(inst.AccountMetaSlice) > 3 {
		return inst.AccountMetaSlice[3]
	}
	return nil
}

func (inst *Authorize) SetNewAuthorized(newAuthorized solana.PublicKey) *Authorize {
	inst.NewAuthorized = &newAuthorized
	return inst
}

func (inst *Authorize) SetStakeAuthorize(stakeAuthorize StakeAuthorize) *Authorize {
	inst.StakeAuthorize = &stakeAuthorize
	return inst
}

func (inst *Authorize) UnmarshalWithDecoder(dec *bin.Decoder) error {
	{
		err := dec.Decode(&inst.NewAuthorized)
		if err != nil {
			return err
		}
	}
	{
		val, err := dec.ReadUint32(binary.LittleEndian)
		if err != nil {
			return err
		}
		sa := StakeAuthorize(val)
		inst.StakeAuthorize = &sa
	}
	return nil
}

func (inst *Authorize) MarshalWithEncoder(encoder *bin.Encoder) error {
	{
		err := encoder.Encode(*inst.NewAuthorized)
		if err != nil {
			return err
		}
	}
	{
		err := encoder.WriteUint32(uint32(*inst.StakeAuthorize), binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst Authorize) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_Authorize, bin.LE),
	}}
}

func (inst *Authorize) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("Authorize")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Account("NewAuthorized", *inst.NewAuthorized))
						paramsBranch.Child(format.Param("StakeAuthorize", inst.StakeAuthorize))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("          StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("           ClockSysvar", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("            Authority", inst.AccountMetaSlice.Get(2)))
						if len(inst.AccountMetaSlice) > 3 {
							accountsBranch.Child(format.Meta("      LockupAuthority", inst.AccountMetaSlice.Get(3)))
						}
					})
				})
		})
}

// NewAuthorizeInstructionBuilder creates a new `Authorize` instruction builder.
func NewAuthorizeInstructionBuilder() *Authorize {
	nd := &Authorize{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
	}
	return nd
}

// NewAuthorizeInstruction declares a new Authorize instruction with the provided parameters and accounts.
func NewAuthorizeInstruction(
	// Params:
	newAuthorized solana.PublicKey,
	stakeAuthorize StakeAuthorize,
	// Accounts:
	stakeAccount solana.PublicKey,
	authority solana.PublicKey,
) *Authorize {
	return NewAuthorizeInstructionBuilder().
		SetNewAuthorized(newAuthorized).
		SetStakeAuthorize(stakeAuthorize).
		SetStakeAccount(stakeAccount).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetAuthority(authority)
}
