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

package loaderv4

import (
	"encoding/binary"
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Copy splices bytes from a source program account into the destination
// program account without having to round-trip through `Write`.
//
// Account references:
//
//	[0] = [WRITE]   Destination program
//	[1] = [SIGNER]  Authority
//	[2] = []        Source program
type Copy struct {
	DestinationOffset *uint32
	SourceOffset      *uint32
	Length            *uint32

	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewCopyInstructionBuilder() *Copy {
	return &Copy{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 3),
	}
}

func (inst *Copy) SetDestinationOffset(v uint32) *Copy { inst.DestinationOffset = &v; return inst }
func (inst *Copy) SetSourceOffset(v uint32) *Copy      { inst.SourceOffset = &v; return inst }
func (inst *Copy) SetLength(v uint32) *Copy            { inst.Length = &v; return inst }

func (inst *Copy) SetProgramAccount(program ag_solanago.PublicKey) *Copy {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(program).WRITE()
	return inst
}

func (inst *Copy) SetAuthority(authority ag_solanago.PublicKey) *Copy {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(authority).SIGNER()
	return inst
}

func (inst *Copy) SetSourceAccount(source ag_solanago.PublicKey) *Copy {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(source)
	return inst
}

func (inst Copy) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_Copy, binary.LittleEndian),
	}}
}

func (inst Copy) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Copy) Validate() error {
	if inst.DestinationOffset == nil {
		return errors.New("DestinationOffset parameter is not set")
	}
	if inst.SourceOffset == nil {
		return errors.New("SourceOffset parameter is not set")
	}
	if inst.Length == nil {
		return errors.New("length parameter is not set")
	}
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *Copy) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Copy")).
				ParentFunc(func(ib ag_treeout.Branches) {
					ib.Child("Params").ParentFunc(func(p ag_treeout.Branches) {
						p.Child(ag_format.Param("DestinationOffset", *inst.DestinationOffset))
						p.Child(ag_format.Param("     SourceOffset", *inst.SourceOffset))
						p.Child(ag_format.Param("           Length", *inst.Length))
					})
					ib.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						a.Child(ag_format.Meta("Destination", inst.AccountMetaSlice[0]))
						a.Child(ag_format.Meta("  Authority", inst.AccountMetaSlice[1]))
						a.Child(ag_format.Meta("     Source", inst.AccountMetaSlice[2]))
					})
				})
		})
}

func (inst Copy) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	if err := encoder.WriteUint32(*inst.DestinationOffset, binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteUint32(*inst.SourceOffset, binary.LittleEndian); err != nil {
		return err
	}
	return encoder.WriteUint32(*inst.Length, binary.LittleEndian)
}

func (inst *Copy) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	dst, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	inst.DestinationOffset = &dst
	src, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	inst.SourceOffset = &src
	l, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	inst.Length = &l
	return nil
}

func NewCopyInstruction(
	program, authority, source ag_solanago.PublicKey,
	destinationOffset, sourceOffset, length uint32,
) *Copy {
	return NewCopyInstructionBuilder().
		SetProgramAccount(program).
		SetAuthority(authority).
		SetSourceAccount(source).
		SetDestinationOffset(destinationOffset).
		SetSourceOffset(sourceOffset).
		SetLength(length)
}
