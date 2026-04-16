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

func SetProgramID(pubkey ag_solanago.PublicKey) error {
	ProgramID = pubkey
	return ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "Token"

func init() {
	if !ProgramID.IsZero() {
		ag_solanago.MustRegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
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

	// Gets the required size of an account for the given mint as a little-endian u64.
	Instruction_GetAccountDataSize

	// Initialize the Immutable Owner extension for the given token account.
	// No-ops in this version of the program, but is included for compatibility
	// with the Associated Token Account program.
	Instruction_InitializeImmutableOwner

	// Convert an Amount of tokens to a UiAmount string, using the given mint.
	Instruction_AmountToUiAmount

	// Convert a UiAmount of tokens to a little-endian u64 raw Amount, using the given mint.
	Instruction_UiAmountToAmount
)

const (
	// Withdraw excess lamports from a token account, mint, or multisig.
	// Only available in the p-token (Pinocchio) implementation.
	Instruction_WithdrawExcessLamports uint8 = 38

	// Unwrap lamports from a native SOL token account directly to a destination.
	// Only available in the p-token (Pinocchio) implementation.
	Instruction_UnwrapLamports uint8 = 45

	// Execute multiple token instructions in a single call.
	// Only available in the p-token (Pinocchio) implementation.
	Instruction_Batch uint8 = 255
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
	case Instruction_WithdrawExcessLamports:
		return "WithdrawExcessLamports"
	case Instruction_UnwrapLamports:
		return "UnwrapLamports"
	case Instruction_Batch:
		return "Batch"
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

// instructionImplDefs contains the variant types for instruction IDs 0-24.
// IDs 0-20 are the original SPL Token instructions, IDs 21-24 are additional
// instructions added for Token-2022 compatibility.
var instructionImplDefs = []ag_binary.VariantType{
	{"InitializeMint", (*InitializeMint)(nil)},           // 0
	{"InitializeAccount", (*InitializeAccount)(nil)},     // 1
	{"InitializeMultisig", (*InitializeMultisig)(nil)},   // 2
	{"Transfer", (*Transfer)(nil)},                       // 3
	{"Approve", (*Approve)(nil)},                         // 4
	{"Revoke", (*Revoke)(nil)},                           // 5
	{"SetAuthority", (*SetAuthority)(nil)},               // 6
	{"MintTo", (*MintTo)(nil)},                           // 7
	{"Burn", (*Burn)(nil)},                               // 8
	{"CloseAccount", (*CloseAccount)(nil)},               // 9
	{"FreezeAccount", (*FreezeAccount)(nil)},             // 10
	{"ThawAccount", (*ThawAccount)(nil)},                 // 11
	{"TransferChecked", (*TransferChecked)(nil)},         // 12
	{"ApproveChecked", (*ApproveChecked)(nil)},           // 13
	{"MintToChecked", (*MintToChecked)(nil)},             // 14
	{"BurnChecked", (*BurnChecked)(nil)},                 // 15
	{"InitializeAccount2", (*InitializeAccount2)(nil)},   // 16
	{"SyncNative", (*SyncNative)(nil)},                   // 17
	{"InitializeAccount3", (*InitializeAccount3)(nil)},   // 18
	{"InitializeMultisig2", (*InitializeMultisig2)(nil)}, // 19
	{"InitializeMint2", (*InitializeMint2)(nil)},         // 20
	{"GetAccountDataSize", (*GetAccountDataSize)(nil)},           // 21
	{"InitializeImmutableOwner", (*InitializeImmutableOwner)(nil)}, // 22
	{"AmountToUiAmount", (*AmountToUiAmount)(nil)},       // 23
	{"UiAmountToAmount", (*UiAmountToAmount)(nil)},       // 24
}

// InstructionImplDef is the variant definition for instruction IDs 0-24.
// For p-token instructions with non-contiguous IDs (38, 45, 255),
// use DecodeInstruction which handles them via custom dispatch.
var InstructionImplDef = ag_binary.NewVariantDefinition(
	ag_binary.Uint8TypeIDEncoding,
	instructionImplDefs,
)

// pTokenInstructionMap maps non-contiguous p-token instruction IDs to their types.
var pTokenInstructionMap = map[uint8]ag_binary.VariantType{
	Instruction_WithdrawExcessLamports: {"WithdrawExcessLamports", (*WithdrawExcessLamports)(nil)},
	Instruction_UnwrapLamports:         {"UnwrapLamports", (*UnwrapLamports)(nil)},
	Instruction_Batch:                  {"Batch", (*Batch)(nil)},
}

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
	if len(data) < 1 {
		return nil, fmt.Errorf("instruction data is empty")
	}

	discriminator := data[0]

	if vt, ok := pTokenInstructionMap[discriminator]; ok {
		return decodePTokenInstruction(accounts, data, discriminator, vt)
	}

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

func decodePTokenInstruction(accounts []*ag_solanago.AccountMeta, data []byte, discriminator uint8, vt ag_binary.VariantType) (*Instruction, error) {
	inst := new(Instruction)
	inst.TypeID = ag_binary.TypeIDFromUint8(discriminator)

	switch impl := vt.Type.(type) {
	case *WithdrawExcessLamports:
		_ = impl
		obj := new(WithdrawExcessLamports)
		if err := ag_binary.NewBinDecoder(data[1:]).Decode(obj); err != nil {
			return nil, fmt.Errorf("unable to decode WithdrawExcessLamports: %w", err)
		}
		inst.Impl = obj
	case *UnwrapLamports:
		_ = impl
		obj := new(UnwrapLamports)
		if err := ag_binary.NewBinDecoder(data[1:]).Decode(obj); err != nil {
			return nil, fmt.Errorf("unable to decode UnwrapLamports: %w", err)
		}
		inst.Impl = obj
	case *Batch:
		_ = impl
		obj := new(Batch)
		if err := ag_binary.NewBinDecoder(data[1:]).Decode(obj); err != nil {
			return nil, fmt.Errorf("unable to decode Batch: %w", err)
		}
		inst.Impl = obj
	default:
		return nil, fmt.Errorf("unknown p-token instruction type for discriminator %d", discriminator)
	}

	if v, ok := inst.Impl.(ag_solanago.AccountsSettable); ok {
		if err := v.SetAccounts(accounts); err != nil {
			return nil, fmt.Errorf("unable to set accounts for instruction: %w", err)
		}
	}
	return inst, nil
}
