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

type AuthorizeCheckedWithSeed struct {
	// Authorization arguments
	Args *AuthorizeCheckedWithSeedArgs

	// [0] = [WRITE] Stake Account
	// ··········· Stake account to be updated
	//
	// [1] = [SIGNER] Authority Base
	// ··········· Base key of stake or withdraw authority
	//
	// [2] = [] Clock Sysvar
	// ··········· The Clock Sysvar Account
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

func (inst *AuthorizeCheckedWithSeed) Validate() error {
	{
		if inst.Args == nil {
			return errors.New("args parameter is not set")
		}
		if inst.Args.StakeAuthorize == nil {
			return errors.New("stake authorize parameter is not set")
		}
		if inst.Args.AuthoritySeed == nil {
			return errors.New("authority seed parameter is not set")
		}
		if inst.Args.AuthorityOwner == nil {
			return errors.New("authority owner parameter is not set")
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

func (inst *AuthorizeCheckedWithSeed) SetStakeAccount(stakeAccount solana.PublicKey) *AuthorizeCheckedWithSeed {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE()
	return inst
}
func (inst *AuthorizeCheckedWithSeed) SetAuthorityBase(authorityBase solana.PublicKey) *AuthorizeCheckedWithSeed {
	inst.AccountMetaSlice[1] = solana.Meta(authorityBase).SIGNER()
	return inst
}
func (inst *AuthorizeCheckedWithSeed) SetClockSysvar(clockSysvar solana.PublicKey) *AuthorizeCheckedWithSeed {
	inst.AccountMetaSlice[2] = solana.Meta(clockSysvar)
	return inst
}
func (inst *AuthorizeCheckedWithSeed) SetNewAuthority(newAuthority solana.PublicKey) *AuthorizeCheckedWithSeed {
	inst.AccountMetaSlice[3] = solana.Meta(newAuthority).SIGNER()
	return inst
}
func (inst *AuthorizeCheckedWithSeed) SetLockupAuthority(lockupAuthority solana.PublicKey) *AuthorizeCheckedWithSeed {
	inst.AccountMetaSlice = append(inst.AccountMetaSlice, solana.Meta(lockupAuthority).SIGNER())
	return inst
}

func (inst *AuthorizeCheckedWithSeed) GetStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}
func (inst *AuthorizeCheckedWithSeed) GetAuthorityBase() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}
func (inst *AuthorizeCheckedWithSeed) GetClockSysvar() *solana.AccountMeta {
	return inst.AccountMetaSlice[2]
}
func (inst *AuthorizeCheckedWithSeed) GetNewAuthority() *solana.AccountMeta {
	return inst.AccountMetaSlice[3]
}
func (inst *AuthorizeCheckedWithSeed) GetLockupAuthority() *solana.AccountMeta {
	if len(inst.AccountMetaSlice) > 4 {
		return inst.AccountMetaSlice[4]
	}
	return nil
}

func (inst *AuthorizeCheckedWithSeed) SetArgs(args AuthorizeCheckedWithSeedArgs) *AuthorizeCheckedWithSeed {
	inst.Args = &args
	return inst
}

func (inst *AuthorizeCheckedWithSeed) SetStakeAuthorize(stakeAuthorize StakeAuthorize) *AuthorizeCheckedWithSeed {
	inst.Args.StakeAuthorize = &stakeAuthorize
	return inst
}

func (inst *AuthorizeCheckedWithSeed) SetAuthoritySeed(seed string) *AuthorizeCheckedWithSeed {
	inst.Args.AuthoritySeed = &seed
	return inst
}

func (inst *AuthorizeCheckedWithSeed) SetAuthorityOwner(owner solana.PublicKey) *AuthorizeCheckedWithSeed {
	inst.Args.AuthorityOwner = &owner
	return inst
}

func (inst AuthorizeCheckedWithSeed) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_AuthorizeCheckedWithSeed, bin.LE),
	}}
}

func (inst *AuthorizeCheckedWithSeed) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("AuthorizeCheckedWithSeed")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child("AuthorizeCheckedWithSeedArgs").ParentFunc(func(argsBranch treeout.Branches) {
							argsBranch.Child(format.Param("StakeAuthorize", inst.Args.StakeAuthorize))
							argsBranch.Child(format.Param(" AuthoritySeed", inst.Args.AuthoritySeed))
							argsBranch.Child(format.Account("AuthorityOwner", *inst.Args.AuthorityOwner))
						})
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("       StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("     AuthorityBase", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("        ClockSysvar", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(format.Meta("       NewAuthority", inst.AccountMetaSlice.Get(3)))
						if len(inst.AccountMetaSlice) > 4 {
							accountsBranch.Child(format.Meta("   LockupAuthority", inst.AccountMetaSlice.Get(4)))
						}
					})
				})
		})
}

// NewAuthorizeCheckedWithSeedInstructionBuilder creates a new `AuthorizeCheckedWithSeed` instruction builder.
func NewAuthorizeCheckedWithSeedInstructionBuilder() *AuthorizeCheckedWithSeed {
	nd := &AuthorizeCheckedWithSeed{
		AccountMetaSlice: make(solana.AccountMetaSlice, 4),
		Args:             &AuthorizeCheckedWithSeedArgs{},
	}
	return nd
}

// NewAuthorizeCheckedWithSeedInstruction declares a new AuthorizeCheckedWithSeed instruction with the provided parameters and accounts.
func NewAuthorizeCheckedWithSeedInstruction(
	// Params:
	stakeAuthorize StakeAuthorize,
	authoritySeed string,
	authorityOwner solana.PublicKey,
	// Accounts:
	stakeAccount solana.PublicKey,
	authorityBase solana.PublicKey,
	newAuthority solana.PublicKey,
) *AuthorizeCheckedWithSeed {
	return NewAuthorizeCheckedWithSeedInstructionBuilder().
		SetStakeAuthorize(stakeAuthorize).
		SetAuthoritySeed(authoritySeed).
		SetAuthorityOwner(authorityOwner).
		SetStakeAccount(stakeAccount).
		SetAuthorityBase(authorityBase).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetNewAuthority(newAuthority)
}
