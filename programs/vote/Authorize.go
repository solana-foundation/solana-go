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
	"encoding/binary"
	"errors"
	"fmt"

	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

// Authorize is the Authorize instruction.
// Data: (Pubkey, VoteAuthorize)
type Authorize struct {
	NewAuthority  *solana.PublicKey
	VoteAuthorize *VoteAuthorizeKind

	// [0] = [WRITE] VoteAccount
	//
	// [1] = [] SysVarClock
	//
	// [2] = [SIGNER] Authority
	// ··········· Current vote or withdraw authority.
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewAuthorizeInstructionBuilder() *Authorize {
	return &Authorize{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
	}
}

func NewAuthorizeInstruction(
	newAuthority solana.PublicKey,
	kind VoteAuthorizeKind,
	voteAccount solana.PublicKey,
	currentAuthority solana.PublicKey,
) *Authorize {
	return NewAuthorizeInstructionBuilder().
		SetNewAuthority(newAuthority).
		SetVoteAuthorize(kind).
		SetVoteAccount(voteAccount).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetAuthority(currentAuthority)
}

func (inst *Authorize) SetNewAuthority(pk solana.PublicKey) *Authorize {
	inst.NewAuthority = &pk
	return inst
}

func (inst *Authorize) SetVoteAuthorize(kind VoteAuthorizeKind) *Authorize {
	inst.VoteAuthorize = &kind
	return inst
}

func (inst *Authorize) SetVoteAccount(pk solana.PublicKey) *Authorize {
	inst.AccountMetaSlice[0] = solana.Meta(pk).WRITE()
	return inst
}

func (inst *Authorize) SetClockSysvar(pk solana.PublicKey) *Authorize {
	inst.AccountMetaSlice[1] = solana.Meta(pk)
	return inst
}

func (inst *Authorize) SetAuthority(pk solana.PublicKey) *Authorize {
	inst.AccountMetaSlice[2] = solana.Meta(pk).SIGNER()
	return inst
}

func (inst *Authorize) GetVoteAccount() *solana.AccountMeta { return inst.AccountMetaSlice[0] }
func (inst *Authorize) GetClockSysvar() *solana.AccountMeta { return inst.AccountMetaSlice[1] }
func (inst *Authorize) GetAuthority() *solana.AccountMeta   { return inst.AccountMetaSlice[2] }

func (inst Authorize) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_Authorize, bin.LE),
	}}
}

func (inst Authorize) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Authorize) Validate() error {
	if inst.NewAuthority == nil {
		return errors.New("NewAuthority parameter is not set")
	}
	if inst.VoteAuthorize == nil {
		return errors.New("VoteAuthorize parameter is not set")
	}
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *Authorize) UnmarshalWithDecoder(dec *bin.Decoder) error {
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	pk := solana.PublicKeyFromBytes(b)
	inst.NewAuthority = &pk
	raw, err := dec.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	kind := VoteAuthorizeKind(raw)
	inst.VoteAuthorize = &kind
	return nil
}

func (inst Authorize) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.NewAuthority == nil {
		return errors.New("Authorize.NewAuthority is nil")
	}
	if inst.VoteAuthorize == nil {
		return errors.New("Authorize.VoteAuthorize is nil")
	}
	if err := enc.WriteBytes(inst.NewAuthority[:], false); err != nil {
		return err
	}
	return enc.WriteUint32(uint32(*inst.VoteAuthorize), binary.LittleEndian)
}

func (inst *Authorize) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("Authorize")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param(" NewAuthority", inst.NewAuthority))
						paramsBranch.Child(format.Param("VoteAuthorize", inst.VoteAuthorize))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("VoteAccount", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta("ClockSysvar", inst.AccountMetaSlice[1]))
						accountsBranch.Child(format.Meta("  Authority", inst.AccountMetaSlice[2]))
					})
				})
		})
}
