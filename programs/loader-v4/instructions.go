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

// Package loaderv4 implements client-side instruction builders for loader-v4
// (program ID LoaderV411...), the forward-looking replacement for the
// upgradeable BPF loader.
//
// The wire format matches upstream's default bincode encoding (u32 LE
// discriminant, u64 LE Vec<u8> length). Although the Rust enum is marked
// `#[repr(u8)]`, bincode ignores that attribute and emits a u32 discriminant.
package loaderv4

import (
	"bytes"
	"encoding/binary"
	"fmt"

	ag_spew "github.com/davecgh/go-spew/spew"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_text "github.com/gagliardetto/solana-go/text"
	ag_treeout "github.com/gagliardetto/treeout"
)

var ProgramID ag_solanago.PublicKey = ag_solanago.LoaderV4ProgramID

func SetProgramID(pubkey ag_solanago.PublicKey) error {
	ProgramID = pubkey
	return ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "LoaderV4"

func init() {
	ag_solanago.MustRegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

// DEPLOYMENT_COOLDOWN_IN_SLOTS is the minimum number of slots that must pass
// between successive deploys of the same program (upstream constant).
const DEPLOYMENT_COOLDOWN_IN_SLOTS uint64 = 1

const (
	Instruction_Write uint32 = iota
	Instruction_Copy
	Instruction_SetProgramLength
	Instruction_Deploy
	Instruction_Retract
	Instruction_TransferAuthority
	Instruction_Finalize
)

func InstructionIDToName(id uint32) string {
	switch id {
	case Instruction_Write:
		return "Write"
	case Instruction_Copy:
		return "Copy"
	case Instruction_SetProgramLength:
		return "SetProgramLength"
	case Instruction_Deploy:
		return "Deploy"
	case Instruction_Retract:
		return "Retract"
	case Instruction_TransferAuthority:
		return "TransferAuthority"
	case Instruction_Finalize:
		return "Finalize"
	default:
		return ""
	}
}

type Instruction struct {
	ag_binary.BaseVariant
}

func (inst *Instruction) EncodeToTree(parent ag_treeout.Branches) {
	if enToTree, ok := inst.Impl.(ag_text.EncodableToTree); ok {
		enToTree.EncodeToTree(parent)
	} else {
		parent.Child(ag_spew.Sdump(inst))
	}
}

var InstructionImplDef = ag_binary.NewVariantDefinition(
	ag_binary.Uint32TypeIDEncoding,
	[]ag_binary.VariantType{
		{Name: "Write", Type: (*Write)(nil)},
		{Name: "Copy", Type: (*Copy)(nil)},
		{Name: "SetProgramLength", Type: (*SetProgramLength)(nil)},
		{Name: "Deploy", Type: (*Deploy)(nil)},
		{Name: "Retract", Type: (*Retract)(nil)},
		{Name: "TransferAuthority", Type: (*TransferAuthority)(nil)},
		{Name: "Finalize", Type: (*Finalize)(nil)},
	},
)

func (inst *Instruction) ProgramID() ag_solanago.PublicKey {
	return ProgramID
}

func (inst *Instruction) Accounts() (out []*ag_solanago.AccountMeta) {
	return inst.Impl.(ag_solanago.AccountsGettable).GetAccounts()
}

func (inst *Instruction) Data() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := ag_binary.NewBinEncoder(buf).Encode(inst); err != nil {
		return nil, fmt.Errorf("unable to encode instruction: %w", err)
	}
	return buf.Bytes(), nil
}

func (inst *Instruction) TextEncode(encoder *ag_text.Encoder, option *ag_text.Option) error {
	return encoder.Encode(inst.Impl, option)
}

func (inst *Instruction) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	return inst.BaseVariant.UnmarshalBinaryVariant(decoder, InstructionImplDef)
}

func (inst Instruction) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	if err := encoder.WriteUint32(inst.TypeID.Uint32(), binary.LittleEndian); err != nil {
		return fmt.Errorf("unable to write variant type: %w", err)
	}
	return encoder.Encode(inst.Impl)
}

func registryDecodeInstruction(accounts []*ag_solanago.AccountMeta, data []byte) (any, error) {
	return DecodeInstruction(accounts, data)
}

func DecodeInstruction(accounts []*ag_solanago.AccountMeta, data []byte) (*Instruction, error) {
	inst := new(Instruction)
	if err := ag_binary.NewBinDecoder(data).Decode(inst); err != nil {
		return nil, fmt.Errorf("unable to decode instruction: %w", err)
	}
	if v, ok := inst.Impl.(ag_solanago.AccountsSettable); ok {
		if err := v.SetAccounts(accounts); err != nil {
			return nil, fmt.Errorf("unable to set accounts for instruction: %w", err)
		}
	}
	return inst, nil
}

// IsWriteInstruction mirrors upstream's `is_write_instruction` helper: given
// a raw instruction-data slice, returns true iff the first byte matches the
// Write discriminant. The check works because all v4 discriminants fit in a
// single byte and bincode's u32 encoding puts the low byte first.
func IsWriteInstruction(data []byte) bool {
	return len(data) > 0 && data[0] == byte(Instruction_Write)
}

func IsCopyInstruction(data []byte) bool {
	return len(data) > 0 && data[0] == byte(Instruction_Copy)
}

func IsSetProgramLengthInstruction(data []byte) bool {
	return len(data) > 0 && data[0] == byte(Instruction_SetProgramLength)
}

func IsDeployInstruction(data []byte) bool {
	return len(data) > 0 && data[0] == byte(Instruction_Deploy)
}

func IsRetractInstruction(data []byte) bool {
	return len(data) > 0 && data[0] == byte(Instruction_Retract)
}

func IsTransferAuthorityInstruction(data []byte) bool {
	return len(data) > 0 && data[0] == byte(Instruction_TransferAuthority)
}

func IsFinalizeInstruction(data []byte) bool {
	return len(data) > 0 && data[0] == byte(Instruction_Finalize)
}
