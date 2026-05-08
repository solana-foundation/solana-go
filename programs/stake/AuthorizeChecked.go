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

type AuthorizeChecked struct {
	// Type of authority to update (Staker or Withdrawer)
	StakeAuthorize *StakeAuthorize

	// [0] = [WRITE] Stake Account
	// ··········· Stake account to be updated
	//
	// [1] = [] Clock Sysvar
	// ··········· The Clock Sysvar Account
	//
	// [2] = [SIGNER] Current Authority
	// ··········· Current stake or withdraw authority
	//
	// [3] = [SIGNER] New Authority
	// ··········· New stake or withdraw authority
	//
	// OPTIONAL:
	// [4] = [SIGNER] Lockup Authority
	// ··········· Lockup authority, if updating withdrawer before lockup expiration
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *AuthorizeChecked) Validate() error {
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

func (inst *AuthorizeChecked) SetStakeAccount(stakeAccount solana.PublicKey) *AuthorizeChecked {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE()
	return inst
}
func (inst *AuthorizeChecked) SetClockSysvar(clockSysvar solana.PublicKey) *AuthorizeChecked {
	inst.AccountMetaSlice[1] = solana.Meta(clockSysvar)
	return inst
}
func (inst *AuthorizeChecked) SetCurrentAuthority(authority solana.PublicKey) *AuthorizeChecked {
	inst.AccountMetaSlice[2] = solana.Meta(authority).SIGNER()
	return inst
}
func (inst *AuthorizeChecked) SetNewAuthority(newAuthority solana.PublicKey) *AuthorizeChecked {
	inst.AccountMetaSlice[3] = solana.Meta(newAuthority).SIGNER()
	return inst
}
func (inst *AuthorizeChecked) SetLockupAuthority(lockupAuthority solana.PublicKey) *AuthorizeChecked {
	inst.AccountMetaSlice = append(inst.AccountMetaSlice, solana.Meta(lockupAuthority).SIGNER())
	return inst
}

func (inst *AuthorizeChecked) GetStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}
func (inst *AuthorizeChecked) GetClockSysvar() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}
func (inst *AuthorizeChecked) GetCurrentAuthority() *solana.AccountMeta {
	return inst.AccountMetaSlice[2]
}
func (inst *AuthorizeChecked) GetNewAuthority() *solana.AccountMeta {
	return inst.AccountMetaSlice[3]
}
func (inst *AuthorizeChecked) GetLockupAuthority() *solana.AccountMeta {
	if len(inst.AccountMetaSlice) > 4 {
		return inst.AccountMetaSlice[4]
	}
	return nil
}

func (inst *AuthorizeChecked) SetStakeAuthorize(stakeAuthorize StakeAuthorize) *AuthorizeChecked {
	inst.StakeAuthorize = &stakeAuthorize
	return inst
}

func (inst *AuthorizeChecked) UnmarshalWithDecoder(dec *bin.Decoder) error {
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

func (inst *AuthorizeChecked) MarshalWithEncoder(encoder *bin.Encoder) error {
	{
		err := encoder.WriteUint32(uint32(*inst.StakeAuthorize), binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst AuthorizeChecked) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_AuthorizeChecked, bin.LE),
	}}
}

func (inst *AuthorizeChecked) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("AuthorizeChecked")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("StakeAuthorize", inst.StakeAuthorize))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("       StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("        ClockSysvar", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("   CurrentAuthority", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(format.Meta("       NewAuthority", inst.AccountMetaSlice.Get(3)))
						if len(inst.AccountMetaSlice) > 4 {
							accountsBranch.Child(format.Meta("   LockupAuthority", inst.AccountMetaSlice.Get(4)))
						}
					})
				})
		})
}

// NewAuthorizeCheckedInstructionBuilder creates a new `AuthorizeChecked` instruction builder.
func NewAuthorizeCheckedInstructionBuilder() *AuthorizeChecked {
	nd := &AuthorizeChecked{
		AccountMetaSlice: make(solana.AccountMetaSlice, 4),
	}
	return nd
}

// NewAuthorizeCheckedInstruction declares a new AuthorizeChecked instruction with the provided parameters and accounts.
func NewAuthorizeCheckedInstruction(
	// Params:
	stakeAuthorize StakeAuthorize,
	// Accounts:
	stakeAccount solana.PublicKey,
	currentAuthority solana.PublicKey,
	newAuthority solana.PublicKey,
) *AuthorizeChecked {
	return NewAuthorizeCheckedInstructionBuilder().
		SetStakeAuthorize(stakeAuthorize).
		SetStakeAccount(stakeAccount).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetCurrentAuthority(currentAuthority).
		SetNewAuthority(newAuthority)
}
