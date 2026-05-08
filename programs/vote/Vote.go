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
	"fmt"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/treeout"
	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/text/format"
)

type Vote struct {
	Slots     []uint64
	Hash      solana.Hash
	Timestamp *int64

	// [0] = [WRITE] VoteAccount
	// ··········· Vote account to vote with
	//
	// [1] = [] SysVarSlotHashes
	// ··········· Slot hashes sysvar
	//
	// [2] = [] SysVarClock
	// ··········· Clock sysvar
	//
	// [3] = [SIGNER] VoteAuthority
	// ··········· Vote authority
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewVoteInstructionBuilder() *Vote {
	return &Vote{
		AccountMetaSlice: make(solana.AccountMetaSlice, 4),
	}
}

func NewVoteInstruction(
	slots []uint64,
	hash solana.Hash,
	timestamp *int64,
	voteAccount solana.PublicKey,
	voteAuthority solana.PublicKey,
) *Vote {
	v := NewVoteInstructionBuilder().
		SetSlots(slots).
		SetHash(hash).
		SetVoteAccount(voteAccount).
		SetSlotHashesSysvar(solana.SysVarSlotHashesPubkey).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetVoteAuthority(voteAuthority)
	if timestamp != nil {
		v.SetTimestamp(*timestamp)
	}
	return v
}

func (v *Vote) SetSlots(slots []uint64) *Vote {
	v.Slots = slots
	return v
}

func (v *Vote) SetHash(hash solana.Hash) *Vote {
	v.Hash = hash
	return v
}

func (v *Vote) SetTimestamp(ts int64) *Vote {
	v.Timestamp = &ts
	return v
}

func (v *Vote) SetVoteAccount(pk solana.PublicKey) *Vote {
	v.AccountMetaSlice[0] = solana.Meta(pk).WRITE()
	return v
}

func (v *Vote) SetSlotHashesSysvar(pk solana.PublicKey) *Vote {
	v.AccountMetaSlice[1] = solana.Meta(pk)
	return v
}

func (v *Vote) SetClockSysvar(pk solana.PublicKey) *Vote {
	v.AccountMetaSlice[2] = solana.Meta(pk)
	return v
}

func (v *Vote) SetVoteAuthority(pk solana.PublicKey) *Vote {
	v.AccountMetaSlice[3] = solana.Meta(pk).SIGNER()
	return v
}

func (v *Vote) GetVoteAccount() *solana.AccountMeta      { return v.AccountMetaSlice[0] }
func (v *Vote) GetSlotHashesSysvar() *solana.AccountMeta { return v.AccountMetaSlice[1] }
func (v *Vote) GetClockSysvar() *solana.AccountMeta      { return v.AccountMetaSlice[2] }
func (v *Vote) GetVoteAuthority() *solana.AccountMeta    { return v.AccountMetaSlice[3] }

func (v Vote) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   v,
		TypeID: bin.TypeIDFromUint32(Instruction_Vote, bin.LE),
	}}
}

func (v Vote) ValidateAndBuild() (*Instruction, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return v.Build(), nil
}

func (v *Vote) UnmarshalWithDecoder(dec *bin.Decoder) error {
	v.Slots = nil
	numSlots, err := dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	for i := uint64(0); i < numSlots; i++ {
		slot, err := dec.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
		v.Slots = append(v.Slots, slot)
	}
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(v.Hash[:], b)
	timestampVariant, err := dec.ReadUint8()
	if err != nil {
		return err
	}
	switch timestampVariant {
	case 0:
	case 1:
		ts, err := dec.ReadInt64(binary.LittleEndian)
		if err != nil {
			return err
		}
		v.Timestamp = &ts
	default:
		return fmt.Errorf("invalid vote timestamp variant %#x", timestampVariant)
	}
	return nil
}

func (v Vote) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteUint64(uint64(len(v.Slots)), binary.LittleEndian); err != nil {
		return err
	}
	for _, s := range v.Slots {
		if err := enc.WriteUint64(s, binary.LittleEndian); err != nil {
			return err
		}
	}
	if err := enc.WriteBytes(v.Hash[:], false); err != nil {
		return err
	}
	if v.Timestamp == nil {
		return enc.WriteUint8(0)
	}
	if err := enc.WriteUint8(1); err != nil {
		return err
	}
	return enc.WriteInt64(*v.Timestamp, binary.LittleEndian)
}

func (inst *Vote) Validate() error {
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("accounts[%d] is not set", accIndex)
		}
	}
	return nil
}

func (inst *Vote) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("Vote")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("Slots", inst.Slots))
						paramsBranch.Child(format.Param(" Hash", inst.Hash))
						var ts time.Time
						if inst.Timestamp != nil {
							ts = time.Unix(*inst.Timestamp, 0).UTC()
						}
						paramsBranch.Child(format.Param("Timestamp", ts))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("     VoteAccount", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta("SlotHashesSysvar", inst.AccountMetaSlice[1]))
						accountsBranch.Child(format.Meta("     ClockSysvar", inst.AccountMetaSlice[2]))
						accountsBranch.Child(format.Meta("   VoteAuthority", inst.AccountMetaSlice[3]))
					})
				})
		})
}
