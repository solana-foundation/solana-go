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

package loaderv2

import (
	"encoding/binary"
	"errors"
	"fmt"
	"slices"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Write program data into an account.
//
// Account references:
//
//	[0] = [WRITE, SIGNER] Account to write to
type Write struct {
	Offset *uint32
	Bytes  []byte

	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewWriteInstructionBuilder() *Write {
	return &Write{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
}

func (inst *Write) SetOffset(offset uint32) *Write {
	inst.Offset = &offset
	return inst
}

func (inst *Write) SetBytes(data []byte) *Write {
	inst.Bytes = data
	return inst
}

func (inst *Write) SetAccount(account ag_solanago.PublicKey) *Write {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(account).WRITE().SIGNER()
	return inst
}

func (inst *Write) GetAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst Write) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_Write, binary.LittleEndian),
	}}
}

func (inst Write) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Write) Validate() error {
	if inst.Offset == nil {
		return errors.New("offset parameter is not set")
	}
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *Write) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Write")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(p ag_treeout.Branches) {
						p.Child(ag_format.Param("Offset", *inst.Offset))
						p.Child(ag_format.Param(" Bytes", fmt.Sprintf("%d bytes", len(inst.Bytes))))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						a.Child(ag_format.Meta("Account", inst.AccountMetaSlice[0]))
					})
				})
		})
}

// MarshalWithEncoder emits bincode-compatible bytes:
//
//	[offset: u32 LE][len(bytes): u64 LE][bytes...]
//
// ag_binary's default slice encoder uses UVarInt for the length, which does
// not match bincode, so Vec<u8> is serialized manually.
func (inst Write) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	if err := encoder.WriteUint32(*inst.Offset, binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteUint64(uint64(len(inst.Bytes)), binary.LittleEndian); err != nil {
		return err
	}
	return encoder.WriteBytes(inst.Bytes, false)
}

func (inst *Write) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	offset, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	inst.Offset = &offset
	length, err := decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	bts, err := decoder.ReadNBytes(int(length))
	if err != nil {
		return err
	}
	// Clone: ReadNBytes returns a subslice of the decoder's input, so
	// retaining it here would alias whatever buffer the caller passed in.
	inst.Bytes = slices.Clone(bts)
	return nil
}

func NewWriteInstruction(
	offset uint32,
	data []byte,
	account ag_solanago.PublicKey,
) *Write {
	return NewWriteInstructionBuilder().
		SetOffset(offset).
		SetBytes(data).
		SetAccount(account)
}
