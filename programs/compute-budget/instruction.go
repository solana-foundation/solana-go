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

package computebudget

import (
	"bytes"
	"fmt"

	ag_spew "github.com/davecgh/go-spew/spew"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_text "github.com/gagliardetto/solana-go/text"
	ag_treeout "github.com/gagliardetto/treeout"
)

var ProgramID ag_solanago.PublicKey = ag_solanago.ComputeBudget

func SetProgramID(pubkey ag_solanago.PublicKey) error {
	ProgramID = pubkey
	return ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "ComputeBudget"

func init() {
	if !ProgramID.IsZero() {
		ag_solanago.MustRegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
	}
}

const (
	// Deprecated
	// after feature remove_deprecated_request_unit_ix::id() is activated
	Instruction_RequestUnitsDeprecated uint8 = iota

	// Request a specific transaction-wide program heap region size in bytes.
	// The value requested must be a multiple of 1024. This new heap region
	// size applies to each program executed in the transaction, including all
	// calls to CPIs.
	Instruction_RequestHeapFrame

	// Set a specific compute unit limit that the transaction is allowed to consume.
	Instruction_SetComputeUnitLimit

	// Set a compute unit price in "micro-lamports" to pay a higher transaction
	// fee for higher transaction prioritization.
	Instruction_SetComputeUnitPrice

	// Set a specific transaction-wide account data size limit, in bytes, is allowed to load.
	Instruction_SetLoadedAccountsDataSizeLimit
)

// InstructionIDToName returns the name of the instruction given its ID.
func InstructionIDToName(id uint8) string {
	switch id {
	case Instruction_RequestUnitsDeprecated:
		return "RequestUnitsDeprecated"
	case Instruction_RequestHeapFrame:
		return "RequestHeapFrame"
	case Instruction_SetComputeUnitLimit:
		return "SetComputeUnitLimit"
	case Instruction_SetComputeUnitPrice:
		return "SetComputeUnitPrice"
	case Instruction_SetLoadedAccountsDataSizeLimit:
		return "SetLoadedAccountsDataSizeLimit"
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
	ag_binary.Uint8TypeIDEncoding,
	[]ag_binary.VariantType{
		{
			Name: "RequestUnitsDeprecated", Type: (*RequestUnitsDeprecated)(nil),
		},
		{
			Name: "RequestHeapFrame", Type: (*RequestHeapFrame)(nil),
		},
		{
			Name: "SetComputeUnitLimit", Type: (*SetComputeUnitLimit)(nil),
		},
		{
			Name: "SetComputeUnitPrice", Type: (*SetComputeUnitPrice)(nil),
		},
		{
			Name: "SetLoadedAccountsDataSizeLimit", Type: (*SetLoadedAccountsDataSizeLimit)(nil),
		},
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
	err := encoder.WriteUint8(inst.TypeID.Uint8())
	if err != nil {
		return fmt.Errorf("unable to write variant type: %w", err)
	}
	return encoder.Encode(inst.Impl)
}

func registryDecodeInstruction(accounts []*ag_solanago.AccountMeta, data []byte) (any, error) {
	inst, err := DecodeInstruction(accounts, data)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func DecodeInstruction(accounts []*ag_solanago.AccountMeta, data []byte) (*Instruction, error) {
	inst := new(Instruction)
	if err := ag_binary.NewBinDecoder(data).Decode(inst); err != nil {
		return nil, fmt.Errorf("unable to decode instruction: %w", err)
	}
	if v, ok := inst.Impl.(ag_solanago.AccountsSettable); ok {
		err := v.SetAccounts(accounts)
		if err != nil {
			return nil, fmt.Errorf("unable to set accounts for instruction: %w", err)
		}
	}
	return inst, nil
}
