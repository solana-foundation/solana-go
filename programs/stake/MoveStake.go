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

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

type MoveStake struct {
	// Amount of active stake to move
	Lamports *uint64

	// [0] = [WRITE] Source Stake Account
	// ··········· Active source stake account
	//
	// [1] = [WRITE] Destination Stake Account
	// ··········· Active or inactive destination stake account
	//
	// [2] = [SIGNER] Stake Authority
	// ··········· Stake authority
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *MoveStake) Validate() error {
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

func (inst *MoveStake) SetSourceStakeAccount(source solana.PublicKey) *MoveStake {
	inst.AccountMetaSlice[0] = solana.Meta(source).WRITE()
	return inst
}
func (inst *MoveStake) SetDestinationStakeAccount(dest solana.PublicKey) *MoveStake {
	inst.AccountMetaSlice[1] = solana.Meta(dest).WRITE()
	return inst
}
func (inst *MoveStake) SetStakeAuthority(stakeAuthority solana.PublicKey) *MoveStake {
	inst.AccountMetaSlice[2] = solana.Meta(stakeAuthority).SIGNER()
	return inst
}

func (inst *MoveStake) GetSourceStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}
func (inst *MoveStake) GetDestinationStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}
func (inst *MoveStake) GetStakeAuthority() *solana.AccountMeta { return inst.AccountMetaSlice[2] }

func (inst *MoveStake) SetLamports(lamports uint64) *MoveStake {
	inst.Lamports = &lamports
	return inst
}

func (inst *MoveStake) UnmarshalWithDecoder(dec *bin.Decoder) error {
	{
		err := dec.Decode(&inst.Lamports)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst *MoveStake) MarshalWithEncoder(encoder *bin.Encoder) error {
	{
		err := encoder.WriteUint64(*inst.Lamports, binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst MoveStake) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_MoveStake, bin.LE),
	}}
}

func (inst *MoveStake) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("MoveStake")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("Lamports", inst.Lamports))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("    SourceStakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("DestinationStakeAccount", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("        StakeAuthority", inst.AccountMetaSlice.Get(2)))
					})
				})
		})
}

// NewMoveStakeInstructionBuilder creates a new `MoveStake` instruction builder.
func NewMoveStakeInstructionBuilder() *MoveStake {
	nd := &MoveStake{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
	}
	return nd
}

// NewMoveStakeInstruction declares a new MoveStake instruction with the provided parameters and accounts.
func NewMoveStakeInstruction(
	// Params:
	lamports uint64,
	// Accounts:
	sourceStakeAccount solana.PublicKey,
	destinationStakeAccount solana.PublicKey,
	stakeAuthority solana.PublicKey,
) *MoveStake {
	return NewMoveStakeInstructionBuilder().
		SetLamports(lamports).
		SetSourceStakeAccount(sourceStakeAccount).
		SetDestinationStakeAccount(destinationStakeAccount).
		SetStakeAuthority(stakeAuthority)
}
