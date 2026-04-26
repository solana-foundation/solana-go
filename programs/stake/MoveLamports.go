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

type MoveLamports struct {
	// Amount of lamports to move
	Lamports *uint64

	// [0] = [WRITE] Source Stake Account
	// ··········· Active or inactive source stake account
	//
	// [1] = [WRITE] Destination Stake Account
	// ··········· Mergeable destination stake account
	//
	// [2] = [SIGNER] Stake Authority
	// ··········· Stake authority
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *MoveLamports) Validate() error {
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

func (inst *MoveLamports) SetSourceStakeAccount(source solana.PublicKey) *MoveLamports {
	inst.AccountMetaSlice[0] = solana.Meta(source).WRITE()
	return inst
}
func (inst *MoveLamports) SetDestinationStakeAccount(dest solana.PublicKey) *MoveLamports {
	inst.AccountMetaSlice[1] = solana.Meta(dest).WRITE()
	return inst
}
func (inst *MoveLamports) SetStakeAuthority(stakeAuthority solana.PublicKey) *MoveLamports {
	inst.AccountMetaSlice[2] = solana.Meta(stakeAuthority).SIGNER()
	return inst
}

func (inst *MoveLamports) GetSourceStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}
func (inst *MoveLamports) GetDestinationStakeAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}
func (inst *MoveLamports) GetStakeAuthority() *solana.AccountMeta { return inst.AccountMetaSlice[2] }

func (inst *MoveLamports) SetLamports(lamports uint64) *MoveLamports {
	inst.Lamports = &lamports
	return inst
}

func (inst *MoveLamports) UnmarshalWithDecoder(dec *bin.Decoder) error {
	{
		err := dec.Decode(&inst.Lamports)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst *MoveLamports) MarshalWithEncoder(encoder *bin.Encoder) error {
	{
		err := encoder.WriteUint64(*inst.Lamports, binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst MoveLamports) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_MoveLamports, bin.LE),
	}}
}

func (inst *MoveLamports) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("MoveLamports")).
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

// NewMoveLamportsInstructionBuilder creates a new `MoveLamports` instruction builder.
func NewMoveLamportsInstructionBuilder() *MoveLamports {
	nd := &MoveLamports{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
	}
	return nd
}

// NewMoveLamportsInstruction declares a new MoveLamports instruction with the provided parameters and accounts.
func NewMoveLamportsInstruction(
	// Params:
	lamports uint64,
	// Accounts:
	sourceStakeAccount solana.PublicKey,
	destinationStakeAccount solana.PublicKey,
	stakeAuthority solana.PublicKey,
) *MoveLamports {
	return NewMoveLamportsInstructionBuilder().
		SetLamports(lamports).
		SetSourceStakeAccount(sourceStakeAccount).
		SetDestinationStakeAccount(destinationStakeAccount).
		SetStakeAuthority(stakeAuthority)
}
