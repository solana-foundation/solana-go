package transferfee

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

import (
	"bytes"
	"fmt"

	ag_spew "github.com/davecgh/go-spew/spew"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_text "github.com/gagliardetto/solana-go/text"
	ag_treeout "github.com/gagliardetto/treeout"
)

const MAX_SIGNERS = 11

const SharedInstructionPrefix uint8 = 26

var ProgramID ag_solanago.PublicKey = ag_solanago.Token2022ProgramID

func SetProgramID(pubkey ag_solanago.PublicKey) {
	ProgramID = pubkey
	ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "Token2022"

func init() {
	if !ProgramID.IsZero() {
		ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
	}
}

const (
	// Initialize the transfer fee on a new mint.
	// Fails if the mint has already been initialized, so must be called before InitializeMint.
	// The mint must have exactly enough space allocated for the base mint (82 bytes),
	// plus 83 bytes of padding, 1 byte reserved for the account type, then space required for this extension, plus any others.
	Instruction_InitializeTransferFeeConfig uint8 = iota

	// Transfer, providing expected mint information and fees
	// This instruction succeeds if the mint has no configured transfer fee and the provided fee is 0.
	// This allows applications to use TransferCheckedWithFee with any mint.
	Instruction_TransferCheckedWithFee

	// Transfer all withheld tokens in the mint to an account. Signed by the mint’s withdraw withheld tokens authority.
	Instruction_WithdrawWithheldTokensFromMint

	// Transfer all withheld tokens to an account. Signed by the mint’s withdraw withheld tokens authority.
	Instruction_WithdrawWithheldTokensFromAccounts

	// Permissionless instruction to transfer all withheld tokens to the mint.
	Instruction_HarvestWithheldTokensToMint

	// Set transfer fee. Only supported for mints that include the TransferFeeConfig extension.
	Instruction_SetTransferFee
)

// InstructionIDToName returns the name of the instruction given its ID.
func InstructionIDToName(id uint8) string {
	switch id {
	case Instruction_InitializeTransferFeeConfig:
		return "InitializeTransferFeeConfig"
	case Instruction_TransferCheckedWithFee:
		return "TransferCheckedWithFee"
	case Instruction_WithdrawWithheldTokensFromMint:
		return "WithdrawWithheldTokensFromMint"
	case Instruction_WithdrawWithheldTokensFromAccounts:
		return "WithdrawWithheldTokensFromAccounts"
	case Instruction_HarvestWithheldTokensToMint:
		return "HarvestWithheldTokensToMint"
	case Instruction_SetTransferFee:
		return "SetTransferFee"
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
			"InitializeTransferFeeConfig", (*InitializeTransferFeeConfig)(nil),
		},
		{
			"TransferCheckedWithFee", (*TransferCheckedWithFee)(nil),
		},
		{
			"WithdrawWithheldTokensFromMint", (*WithdrawWithheldTokensFromMint)(nil),
		},
		{
			"WithdrawWithheldTokensFromAccounts", (*WithdrawWithheldTokensFromAccounts)(nil),
		},
		{
			"HarvestWithheldTokensToMint", (*HarvestWithheldTokensToMint)(nil),
		},
		{
			"SetTransferFee", (*SetTransferFee)(nil),
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

func registryDecodeInstruction(accounts []*ag_solanago.AccountMeta, data []byte) (interface{}, error) {
	inst, err := DecodeInstruction(accounts, data)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func DecodeInstruction(accounts []*ag_solanago.AccountMeta, data []byte) (*Instruction, error) {
	inst := new(Instruction)
	if data[0] != SharedInstructionPrefix {
		return nil, fmt.Errorf("unexpected instruction prefix for transfer-fee-extension: %d", data[0])
	}
	data = data[1:]
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
