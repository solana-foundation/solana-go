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

// A Token program on the Solana blockchain.
// This program defines a common implementation for Fungible and Non Fungible tokens.

package addresslookuptable

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

// Maximum number of multisignature signers (max N)
const MAX_SIGNERS = 11

var ProgramID ag_solanago.PublicKey = ag_solanago.AddressLookupTableProgramID

func SetProgramID(pubkey ag_solanago.PublicKey) {
	ProgramID = pubkey
	ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "AddressLookupTable"

func init() {
	if !ProgramID.IsZero() {
		ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
	}
}

const (
	// Initializes a new mint and optionally deposits all the newly minted
	// tokens in an account.
	//
	// The `InitializeMint` instruction requires no signers and MUST be
	// included within the same Transaction as the system program's
	// `CreateAccount` instruction that creates the account being initialized.
	// Otherwise another party can acquire ownership of the uninitialized
	// account.
	Instruction_CreateAddressLookupTable uint32 = iota

	// Initializes a new account to hold tokens.  If this account is associated
	// with the native mint then the token balance of the initialized account
	// will be equal to the amount of SOL in the account. If this account is
	// associated with another mint, that mint must be initialized before this
	// command can succeed.
	//
	// The `InitializeAccount` instruction requires no signers and MUST be
	// included within the same Transaction as the system program's
	// `CreateAccount` instruction that creates the account being initialized.
	// Otherwise another party can acquire ownership of the uninitialized
	// account.
	Instruction_FreezeAddressLookupTable

	// Initializes a multisignature account with N provided signers.
	//
	// Multisignature accounts can used in place of any single owner/delegate
	// accounts in any token instruction that require an owner/delegate to be
	// present.  The variant field represents the number of signers (M)
	// required to validate this multisignature account.
	//
	// The `InitializeMultisig` instruction requires no signers and MUST be
	// included within the same Transaction as the system program's
	// `CreateAccount` instruction that creates the account being initialized.
	// Otherwise another party can acquire ownership of the uninitialized
	// account.
	Instruction_ExtendAddressLookupTable

	// Transfers tokens from one account to another either directly or via a
	// delegate.  If this account is associated with the native mint then equal
	// amounts of SOL and Tokens will be transferred to the destination
	// account.
	Instruction_DeactivateAddressLookupTable

	// Approves a delegate.  A delegate is given the authority over tokens on
	// behalf of the source account's owner.
	Instruction_CloseAddressLookupTable
)

// InstructionIDToName returns the name of the instruction given its ID.
func InstructionIDToName(id uint32) string {
	switch id {
	case Instruction_CreateAddressLookupTable:
		return "CreateAddressLookupTable"
	case Instruction_FreezeAddressLookupTable:
		return "FreezeAddressLookupTable"
	case Instruction_ExtendAddressLookupTable:
		return "ExtendAddressLookupTable"
	case Instruction_DeactivateAddressLookupTable:
		return "DeactivateAddressLookupTable"
	case Instruction_CloseAddressLookupTable:
		return "CloseAddressLookupTable"
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
		{
			"CreateAddressLookupTable", (*CreateAddressLookupTable)(nil),
		},
		{
			"FreezeAddressLookupTable", (*FreezeAddressLookupTable)(nil),
		},
		{
			"ExtendAddressLookupTable", (*ExtendAddressLookupTable)(nil),
		},
		{
			"DeactivateAddressLookupTable", (*DeactivateAddressLookupTable)(nil),
		},
		{
			"CloseAddressLookupTable", (*CloseAddressLookupTable)(nil),
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
	err := encoder.WriteUint32(inst.TypeID.Uint32(), binary.LittleEndian)
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
