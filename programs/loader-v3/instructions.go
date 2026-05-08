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

// Package loaderv3 implements client-side instruction builders for the
// upgradeable BPF loader (program ID BPFLoaderUpgradeab1e...), a.k.a.
// loader-v3. This is the canonical deploy/upgrade path used by
// `solana program deploy` in current Agave releases.
//
// The wire format matches upstream's default bincode encoding: u32 LE
// enum discriminant, u64 LE Vec<u8> length prefix, u8 for bool and
// Option tags, 32 raw bytes for Pubkey. The `OptionalTrailingBool` fields
// introduced by SIMD-0430 (close_buffer on DeployWithMaxDataLen/Upgrade)
// and SIMD-0432 (tombstone on Close) are always emitted on write and
// fall back to the documented defaults on an exhausted read.
package loaderv3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	ag_spew "github.com/davecgh/go-spew/spew"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_text "github.com/gagliardetto/solana-go/text"
	ag_treeout "github.com/gagliardetto/treeout"
)

// ErrInvalidBoolEncoding mirrors upstream's `InvalidBoolEncoding(byte)` error
// variant: any byte other than 0 or 1 in an OptionalTrailingBool slot is
// rejected. Callers can match with errors.Is to distinguish this from
// generic decode failures.
var ErrInvalidBoolEncoding = errors.New("invalid bool encoding")

var ProgramID ag_solanago.PublicKey = ag_solanago.BPFLoaderUpgradeableProgramID

func SetProgramID(pubkey ag_solanago.PublicKey) error {
	ProgramID = pubkey
	return ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "BPFLoaderUpgradeable"

func init() {
	ag_solanago.MustRegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

// MINIMUM_EXTEND_PROGRAM_BYTES is the minimum granularity accepted by the
// ExtendProgram instruction (SIMD-0431).
const MINIMUM_EXTEND_PROGRAM_BYTES uint32 = 10_240

const (
	Instruction_InitializeBuffer uint32 = iota
	Instruction_Write
	Instruction_DeployWithMaxDataLen
	Instruction_Upgrade
	Instruction_SetAuthority
	Instruction_Close
	Instruction_ExtendProgram
	Instruction_SetAuthorityChecked
)

func InstructionIDToName(id uint32) string {
	switch id {
	case Instruction_InitializeBuffer:
		return "InitializeBuffer"
	case Instruction_Write:
		return "Write"
	case Instruction_DeployWithMaxDataLen:
		return "DeployWithMaxDataLen"
	case Instruction_Upgrade:
		return "Upgrade"
	case Instruction_SetAuthority:
		return "SetAuthority"
	case Instruction_Close:
		return "Close"
	case Instruction_ExtendProgram:
		return "ExtendProgram"
	case Instruction_SetAuthorityChecked:
		return "SetAuthorityChecked"
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
		{Name: "InitializeBuffer", Type: (*InitializeBuffer)(nil)},
		{Name: "Write", Type: (*Write)(nil)},
		{Name: "DeployWithMaxDataLen", Type: (*DeployWithMaxDataLen)(nil)},
		{Name: "Upgrade", Type: (*Upgrade)(nil)},
		{Name: "SetAuthority", Type: (*SetAuthority)(nil)},
		{Name: "Close", Type: (*Close)(nil)},
		{Name: "ExtendProgram", Type: (*ExtendProgram)(nil)},
		{Name: "SetAuthorityChecked", Type: (*SetAuthorityChecked)(nil)},
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

// readOptionalTrailingBool implements the OptionalTrailingBool<DEFAULT>
// wincode schema used by SIMD-0430/0432: consume one byte from the decoder
// if any remain, otherwise fall back to the provided default.
func readOptionalTrailingBool(decoder *ag_binary.Decoder, def bool) (bool, error) {
	if !decoder.HasRemaining() {
		return def, nil
	}
	b, err := decoder.ReadByte()
	if err != nil {
		return false, err
	}
	switch b {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("%w: byte %d", ErrInvalidBoolEncoding, b)
	}
}
