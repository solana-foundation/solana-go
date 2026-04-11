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

type AuthorizeWithSeed struct {
	// Authorization arguments
	Args *AuthorizeWithSeedArgs

	// [0] = [WRITE] Stake Account
	// ··········· Stake account to be updated
	//
	// [1] = [SIGNER] Authority Base
	// ··········· Base key of stake or withdraw authority
	//
	// [2] = [] Clock Sysvar
	// ··········· The Clock Sysvar Account
	//
	// OPTIONAL:
	// [3] = [SIGNER] Lockup Authority
	// ··········· Lockup authority, if updating withdrawer before lockup expiration
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *AuthorizeWithSeed) Validate() error {
	{
		if inst.Args == nil {
			return errors.New("args parameter is not set")
		}
		if inst.Args.NewAuthorizedPubkey == nil {
			return errors.New("new authorized pubkey parameter is not set")
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

func (inst *AuthorizeWithSeed) SetStakeAccount(stakeAccount solana.PublicKey) *AuthorizeWithSeed {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE()
	return inst
}
func (inst *AuthorizeWithSeed) SetAuthorityBase(authorityBase solana.PublicKey) *AuthorizeWithSeed {
	inst.AccountMetaSlice[1] = solana.Meta(authorityBase).SIGNER()
	return inst
}
func (inst *AuthorizeWithSeed) SetClockSysvar(clockSysvar solana.PublicKey) *AuthorizeWithSeed {
	inst.AccountMetaSlice[2] = solana.Meta(clockSysvar)
	return inst
}
func (inst *AuthorizeWithSeed) SetLockupAuthority(lockupAuthority solana.PublicKey) *AuthorizeWithSeed {
	inst.AccountMetaSlice = append(inst.AccountMetaSlice, solana.Meta(lockupAuthority).SIGNER())
	return inst
}

func (inst *AuthorizeWithSeed) GetStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}
func (inst *AuthorizeWithSeed) GetAuthorityBase() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}
func (inst *AuthorizeWithSeed) GetClockSysvar() *solana.AccountMeta {
	return inst.AccountMetaSlice[2]
}
func (inst *AuthorizeWithSeed) GetLockupAuthority() *solana.AccountMeta {
	if len(inst.AccountMetaSlice) > 3 {
		return inst.AccountMetaSlice[3]
	}
	return nil
}

func (inst *AuthorizeWithSeed) SetArgs(args AuthorizeWithSeedArgs) *AuthorizeWithSeed {
	inst.Args = &args
	return inst
}

func (inst *AuthorizeWithSeed) SetNewAuthorizedPubkey(pubkey solana.PublicKey) *AuthorizeWithSeed {
	inst.Args.NewAuthorizedPubkey = &pubkey
	return inst
}

func (inst *AuthorizeWithSeed) SetStakeAuthorize(stakeAuthorize StakeAuthorize) *AuthorizeWithSeed {
	inst.Args.StakeAuthorize = &stakeAuthorize
	return inst
}

func (inst *AuthorizeWithSeed) SetAuthoritySeed(seed string) *AuthorizeWithSeed {
	inst.Args.AuthoritySeed = &seed
	return inst
}

func (inst *AuthorizeWithSeed) SetAuthorityOwner(owner solana.PublicKey) *AuthorizeWithSeed {
	inst.Args.AuthorityOwner = &owner
	return inst
}

func (inst AuthorizeWithSeed) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_AuthorizeWithSeed, bin.LE),
	}}
}

func (inst *AuthorizeWithSeed) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("AuthorizeWithSeed")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child("AuthorizeWithSeedArgs").ParentFunc(func(argsBranch treeout.Branches) {
							argsBranch.Child(format.Account("NewAuthorizedPubkey", *inst.Args.NewAuthorizedPubkey))
							argsBranch.Child(format.Param("    StakeAuthorize", inst.Args.StakeAuthorize))
							argsBranch.Child(format.Param("     AuthoritySeed", inst.Args.AuthoritySeed))
							argsBranch.Child(format.Account("   AuthorityOwner", *inst.Args.AuthorityOwner))
						})
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("       StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("     AuthorityBase", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("        ClockSysvar", inst.AccountMetaSlice.Get(2)))
						if len(inst.AccountMetaSlice) > 3 {
							accountsBranch.Child(format.Meta("   LockupAuthority", inst.AccountMetaSlice.Get(3)))
						}
					})
				})
		})
}

// NewAuthorizeWithSeedInstructionBuilder creates a new `AuthorizeWithSeed` instruction builder.
func NewAuthorizeWithSeedInstructionBuilder() *AuthorizeWithSeed {
	nd := &AuthorizeWithSeed{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
		Args:             &AuthorizeWithSeedArgs{},
	}
	return nd
}

// NewAuthorizeWithSeedInstruction declares a new AuthorizeWithSeed instruction with the provided parameters and accounts.
func NewAuthorizeWithSeedInstruction(
	// Params:
	newAuthorizedPubkey solana.PublicKey,
	stakeAuthorize StakeAuthorize,
	authoritySeed string,
	authorityOwner solana.PublicKey,
	// Accounts:
	stakeAccount solana.PublicKey,
	authorityBase solana.PublicKey,
) *AuthorizeWithSeed {
	return NewAuthorizeWithSeedInstructionBuilder().
		SetNewAuthorizedPubkey(newAuthorizedPubkey).
		SetStakeAuthorize(stakeAuthorize).
		SetAuthoritySeed(authoritySeed).
		SetAuthorityOwner(authorityOwner).
		SetStakeAccount(stakeAccount).
		SetAuthorityBase(authorityBase).
		SetClockSysvar(solana.SysVarClockPubkey)
}
