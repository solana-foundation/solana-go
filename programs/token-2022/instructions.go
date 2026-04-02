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

// Token-2022 program (Token Extensions) on the Solana blockchain.
// This program is a superset of the SPL Token program with additional
// extension types and instructions.

package token2022

import (
	"bytes"
	"fmt"

	ag_spew "github.com/davecgh/go-spew/spew"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_text "github.com/gagliardetto/solana-go/text"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Maximum number of multisignature signers (max N)
const MAX_SIGNERS = 11

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
	// Initializes a new mint and optionally deposits all the newly minted
	// tokens in an account.
	Instruction_InitializeMint uint8 = iota

	// Initializes a new account to hold tokens.
	Instruction_InitializeAccount

	// Initializes a multisignature account with N provided signers.
	Instruction_InitializeMultisig

	// Transfers tokens from one account to another either directly or via a delegate.
	Instruction_Transfer

	// Approves a delegate.
	Instruction_Approve

	// Revokes the delegate's authority.
	Instruction_Revoke

	// Sets a new authority of a mint or account.
	Instruction_SetAuthority

	// Mints new tokens to an account.
	Instruction_MintTo

	// Burns tokens by removing them from an account.
	Instruction_Burn

	// Close an account by transferring all its SOL to the destination account.
	Instruction_CloseAccount

	// Freeze an Initialized account using the Mint's freeze_authority (if set).
	Instruction_FreezeAccount

	// Thaw a Frozen account using the Mint's freeze_authority (if set).
	Instruction_ThawAccount

	// Transfers tokens from one account to another either directly or via a
	// delegate. This instruction differs from Transfer in that the token mint
	// and decimals value is checked by the caller.
	Instruction_TransferChecked

	// Approves a delegate. This instruction differs from Approve in that the
	// token mint and decimals value is checked by the caller.
	Instruction_ApproveChecked

	// Mints new tokens to an account. This instruction differs from MintTo in
	// that the decimals value is checked by the caller.
	Instruction_MintToChecked

	// Burns tokens by removing them from an account. This instruction differs
	// from Burn in that the decimals value is checked by the caller.
	Instruction_BurnChecked

	// Like InitializeAccount, but the owner pubkey is passed via instruction data
	// rather than the accounts list.
	Instruction_InitializeAccount2

	// Given a wrapped / native token account updates its amount field based on
	// the account's underlying lamports.
	Instruction_SyncNative

	// Like InitializeAccount2, but does not require the Rent sysvar to be provided.
	Instruction_InitializeAccount3

	// Like InitializeMultisig, but does not require the Rent sysvar to be provided.
	Instruction_InitializeMultisig2

	// Like InitializeMint, but does not require the Rent sysvar to be provided.
	Instruction_InitializeMint2

	// Gets the required size of an account for the given mint as a little-endian
	// u64. In includes any extensions that are required for the mint.
	Instruction_GetAccountDataSize

	// Initialize the Immutable Owner extension for the given token account.
	Instruction_InitializeImmutableOwner

	// Convert an Amount of tokens to a UiAmount string, using the given mint's decimals.
	Instruction_AmountToUiAmount

	// Convert a UiAmount of tokens to a little-endian u64 raw Amount, using the given mint's decimals.
	Instruction_UiAmountToAmount

	// Initialize the close account authority on a new mint.
	Instruction_InitializeMintCloseAuthority
)

// InstructionIDToName returns the name of the instruction given its ID.
func InstructionIDToName(id uint8) string {
	switch id {
	case Instruction_InitializeMint:
		return "InitializeMint"
	case Instruction_InitializeAccount:
		return "InitializeAccount"
	case Instruction_InitializeMultisig:
		return "InitializeMultisig"
	case Instruction_Transfer:
		return "Transfer"
	case Instruction_Approve:
		return "Approve"
	case Instruction_Revoke:
		return "Revoke"
	case Instruction_SetAuthority:
		return "SetAuthority"
	case Instruction_MintTo:
		return "MintTo"
	case Instruction_Burn:
		return "Burn"
	case Instruction_CloseAccount:
		return "CloseAccount"
	case Instruction_FreezeAccount:
		return "FreezeAccount"
	case Instruction_ThawAccount:
		return "ThawAccount"
	case Instruction_TransferChecked:
		return "TransferChecked"
	case Instruction_ApproveChecked:
		return "ApproveChecked"
	case Instruction_MintToChecked:
		return "MintToChecked"
	case Instruction_BurnChecked:
		return "BurnChecked"
	case Instruction_InitializeAccount2:
		return "InitializeAccount2"
	case Instruction_SyncNative:
		return "SyncNative"
	case Instruction_InitializeAccount3:
		return "InitializeAccount3"
	case Instruction_InitializeMultisig2:
		return "InitializeMultisig2"
	case Instruction_InitializeMint2:
		return "InitializeMint2"
	case Instruction_GetAccountDataSize:
		return "GetAccountDataSize"
	case Instruction_InitializeImmutableOwner:
		return "InitializeImmutableOwner"
	case Instruction_AmountToUiAmount:
		return "AmountToUiAmount"
	case Instruction_UiAmountToAmount:
		return "UiAmountToAmount"
	case Instruction_InitializeMintCloseAuthority:
		return "InitializeMintCloseAuthority"
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
			"InitializeMint", (*InitializeMint)(nil),
		},
		{
			"InitializeAccount", (*InitializeAccount)(nil),
		},
		{
			"InitializeMultisig", (*InitializeMultisig)(nil),
		},
		{
			"Transfer", (*Transfer)(nil),
		},
		{
			"Approve", (*Approve)(nil),
		},
		{
			"Revoke", (*Revoke)(nil),
		},
		{
			"SetAuthority", (*SetAuthority)(nil),
		},
		{
			"MintTo", (*MintTo)(nil),
		},
		{
			"Burn", (*Burn)(nil),
		},
		{
			"CloseAccount", (*CloseAccount)(nil),
		},
		{
			"FreezeAccount", (*FreezeAccount)(nil),
		},
		{
			"ThawAccount", (*ThawAccount)(nil),
		},
		{
			"TransferChecked", (*TransferChecked)(nil),
		},
		{
			"ApproveChecked", (*ApproveChecked)(nil),
		},
		{
			"MintToChecked", (*MintToChecked)(nil),
		},
		{
			"BurnChecked", (*BurnChecked)(nil),
		},
		{
			"InitializeAccount2", (*InitializeAccount2)(nil),
		},
		{
			"SyncNative", (*SyncNative)(nil),
		},
		{
			"InitializeAccount3", (*InitializeAccount3)(nil),
		},
		{
			"InitializeMultisig2", (*InitializeMultisig2)(nil),
		},
		{
			"InitializeMint2", (*InitializeMint2)(nil),
		},
		{
			"GetAccountDataSize", (*GetAccountDataSize)(nil),
		},
		{
			"InitializeImmutableOwner", (*InitializeImmutableOwner)(nil),
		},
		{
			"AmountToUiAmount", (*AmountToUiAmount)(nil),
		},
		{
			"UiAmountToAmount", (*UiAmountToAmount)(nil),
		},
		{
			"InitializeMintCloseAuthority", (*InitializeMintCloseAuthority)(nil),
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
