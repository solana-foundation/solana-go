// Copyright 2021 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vote

import (
	"errors"
	"fmt"

	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

// InitializeAccount initializes a vote account.
type InitializeAccount struct {
	VoteInit *VoteInit

	// [0] = [WRITE] VoteAccount
	// ··········· Uninitialized vote account to initialize.
	//
	// [1] = [] SysVarRent
	// ··········· Rent sysvar.
	//
	// [2] = [] SysVarClock
	// ··········· Clock sysvar.
	//
	// [3] = [SIGNER] NodePubkey
	// ··········· New validator identity (node_pubkey).
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewInitializeAccountInstructionBuilder() *InitializeAccount {
	return &InitializeAccount{
		AccountMetaSlice: make(solana.AccountMetaSlice, 4),
		VoteInit:         &VoteInit{},
	}
}

func NewInitializeAccountInstruction(
	nodePubkey solana.PublicKey,
	authorizedVoter solana.PublicKey,
	authorizedWithdrawer solana.PublicKey,
	commission uint8,
	voteAccount solana.PublicKey,
) *InitializeAccount {
	return NewInitializeAccountInstructionBuilder().
		SetVoteInit(VoteInit{
			NodePubkey:           nodePubkey,
			AuthorizedVoter:      authorizedVoter,
			AuthorizedWithdrawer: authorizedWithdrawer,
			Commission:           commission,
		}).
		SetVoteAccount(voteAccount).
		SetRentSysvar(solana.SysVarRentPubkey).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetNodePubkey(nodePubkey)
}

func (inst *InitializeAccount) SetVoteInit(v VoteInit) *InitializeAccount {
	inst.VoteInit = &v
	return inst
}

func (inst *InitializeAccount) SetVoteAccount(voteAccount solana.PublicKey) *InitializeAccount {
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	return inst
}

func (inst *InitializeAccount) SetRentSysvar(rent solana.PublicKey) *InitializeAccount {
	inst.AccountMetaSlice[1] = solana.Meta(rent)
	return inst
}

func (inst *InitializeAccount) SetClockSysvar(clock solana.PublicKey) *InitializeAccount {
	inst.AccountMetaSlice[2] = solana.Meta(clock)
	return inst
}

func (inst *InitializeAccount) SetNodePubkey(node solana.PublicKey) *InitializeAccount {
	inst.AccountMetaSlice[3] = solana.Meta(node).SIGNER()
	return inst
}

func (inst *InitializeAccount) GetVoteAccount() *solana.AccountMeta { return inst.AccountMetaSlice[0] }
func (inst *InitializeAccount) GetRentSysvar() *solana.AccountMeta  { return inst.AccountMetaSlice[1] }
func (inst *InitializeAccount) GetClockSysvar() *solana.AccountMeta { return inst.AccountMetaSlice[2] }
func (inst *InitializeAccount) GetNodePubkey() *solana.AccountMeta  { return inst.AccountMetaSlice[3] }

func (inst InitializeAccount) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_InitializeAccount, bin.LE),
	}}
}

func (inst InitializeAccount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeAccount) Validate() error {
	if inst.VoteInit == nil {
		return errors.New("VoteInit parameter is not set")
	}
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *InitializeAccount) UnmarshalWithDecoder(dec *bin.Decoder) error {
	inst.VoteInit = new(VoteInit)
	return inst.VoteInit.UnmarshalWithDecoder(dec)
}

func (inst InitializeAccount) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.VoteInit == nil {
		return errors.New("InitializeAccount.VoteInit is nil")
	}
	return inst.VoteInit.MarshalWithEncoder(enc)
}

func (inst *InitializeAccount) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("InitializeAccount")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						if inst.VoteInit != nil {
							paramsBranch.Child(format.Param("          NodePubkey", inst.VoteInit.NodePubkey))
							paramsBranch.Child(format.Param("     AuthorizedVoter", inst.VoteInit.AuthorizedVoter))
							paramsBranch.Child(format.Param("AuthorizedWithdrawer", inst.VoteInit.AuthorizedWithdrawer))
							paramsBranch.Child(format.Param("          Commission", inst.VoteInit.Commission))
						}
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("VoteAccount", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta(" RentSysvar", inst.AccountMetaSlice[1]))
						accountsBranch.Child(format.Meta("ClockSysvar", inst.AccountMetaSlice[2]))
						accountsBranch.Child(format.Meta(" NodePubkey", inst.AccountMetaSlice[3]))
					})
				})
		})
}
