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

package token

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

var ProgramID ag_solanago.PublicKey = ag_solanago.TokenProgramID

func SetProgramID(pubkey ag_solanago.PublicKey) {
	ProgramID = pubkey
	ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "Token"

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
	Instruction_InitializeMint uint8 = iota

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
	Instruction_InitializeAccount

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
	Instruction_InitializeMultisig

	// Transfers tokens from one account to another either directly or via a
	// delegate.  If this account is associated with the native mint then equal
	// amounts of SOL and Tokens will be transferred to the destination
	// account.
	Instruction_Transfer

	// Approves a delegate.  A delegate is given the authority over tokens on
	// behalf of the source account's owner.
	Instruction_Approve

	// Revokes the delegate's authority.
	Instruction_Revoke

	// Sets a new authority of a mint or account.
	Instruction_SetAuthority

	// Mints new tokens to an account.  The native mint does not support
	// minting.
	Instruction_MintTo

	// Burns tokens by removing them from an account.  `Burn` does not support
	// accounts associated with the native mint, use `CloseAccount` instead.
	Instruction_Burn

	// Close an account by transferring all its SOL to the destination account.
	// Non-native accounts may only be closed if its token amount is zero.
	Instruction_CloseAccount

	// Freeze an Initialized account using the Mint's freeze_authority (if set).
	Instruction_FreezeAccount

	// Thaw a Frozen account using the Mint's freeze_authority (if set).
	Instruction_ThawAccount

	// Transfers tokens from one account to another either directly or via a
	// delegate.  If this account is associated with the native mint then equal
	// amounts of SOL and Tokens will be transferred to the destination
	// account.
	//
	// This instruction differs from Transfer in that the token mint and
	// decimals value is checked by the caller.  This may be useful when
	// creating transactions offline or within a hardware wallet.
	Instruction_TransferChecked

	// Approves a delegate.  A delegate is given the authority over tokens on
	// behalf of the source account's owner.
	//
	// This instruction differs from Approve in that the token mint and
	// decimals value is checked by the caller.  This may be useful when
	// creating transactions offline or within a hardware wallet.
	Instruction_ApproveChecked

	// Mints new tokens to an account.  The native mint does not support minting.
	//
	// This instruction differs from MintTo in that the decimals value is
	// checked by the caller.  This may be useful when creating transactions
	// offline or within a hardware wallet.
	Instruction_MintToChecked

	// Burns tokens by removing them from an account.  `BurnChecked` does not
	// support accounts associated with the native mint, use `CloseAccount`
	// instead.
	//
	// This instruction differs from Burn in that the decimals value is checked
	// by the caller. This may be useful when creating transactions offline or
	// within a hardware wallet.
	Instruction_BurnChecked

	// Like InitializeAccount, but the owner pubkey is passed via instruction data
	// rather than the accounts list. This variant may be preferable when using
	// Cross Program Invocation from an instruction that does not need the owner's
	// `AccountInfo` otherwise.
	Instruction_InitializeAccount2

	// Given a wrapped / native token account (a token account containing SOL)
	// updates its amount field based on the account's underlying `lamports`.
	// This is useful if a non-wrapped SOL account uses `system_instruction::transfer`
	// to move lamports to a wrapped token account, and needs to have its token
	// `amount` field updated.
	Instruction_SyncNative

	// Like InitializeAccount2, but does not require the Rent sysvar to be provided.
	Instruction_InitializeAccount3

	// Like InitializeMultisig, but does not require the Rent sysvar to be provided.
	Instruction_InitializeMultisig2

	// Like InitializeMint, but does not require the Rent sysvar to be provided.
	Instruction_InitializeMint2
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
			Name: "InitializeMint", Type: (*InitializeMint)(nil),
		},
		{
			Name: "InitializeAccount", Type: (*InitializeAccount)(nil),
		},
		{
			Name: "InitializeMultisig", Type: (*InitializeMultisig)(nil),
		},
		{
			Name: "Transfer", Type: (*Transfer)(nil),
		},
		{
			Name: "Approve", Type: (*Approve)(nil),
		},
		{
			Name: "Revoke", Type: (*Revoke)(nil),
		},
		{
			Name: "SetAuthority", Type: (*SetAuthority)(nil),
		},
		{
			Name: "MintTo", Type: (*MintTo)(nil),
		},
		{
			Name: "Burn", Type: (*Burn)(nil),
		},
		{
			Name: "CloseAccount", Type: (*CloseAccount)(nil),
		},
		{
			Name: "FreezeAccount", Type: (*FreezeAccount)(nil),
		},
		{
			Name: "ThawAccount", Type: (*ThawAccount)(nil),
		},
		{
			Name: "TransferChecked", Type: (*TransferChecked)(nil),
		},
		{
			Name: "ApproveChecked", Type: (*ApproveChecked)(nil),
		},
		{
			Name: "MintToChecked", Type: (*MintToChecked)(nil),
		},
		{
			Name: "BurnChecked", Type: (*BurnChecked)(nil),
		},
		{
			Name: "InitializeAccount2", Type: (*InitializeAccount2)(nil),
		},
		{
			Name: "SyncNative", Type: (*SyncNative)(nil),
		},
		{
			Name: "InitializeAccount3", Type: (*InitializeAccount3)(nil),
		},
		{
			Name: "InitializeMultisig2", Type: (*InitializeMultisig2)(nil),
		},
		{
			Name: "InitializeMint2", Type: (*InitializeMint2)(nil),
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
