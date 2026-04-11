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
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_text "github.com/gagliardetto/solana-go/text"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Maximum number of multisignature signers (max N)
const MAX_SIGNERS = 11

var ProgramID ag_solanago.PublicKey = ag_solanago.Token2022ProgramID

func SetProgramID(pubkey ag_solanago.PublicKey) error {
	ProgramID = pubkey
	return ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "Token2022"

func init() {
	if !ProgramID.IsZero() {
		ag_solanago.MustRegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
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

	// Transfer fee extension instructions.
	Instruction_TransferFeeExtension

	// Confidential transfer extension instructions.
	Instruction_ConfidentialTransferExtension

	// Default account state extension instructions.
	Instruction_DefaultAccountStateExtension

	// Reallocate an account to hold additional extensions.
	Instruction_Reallocate

	// Memo transfer extension instructions.
	Instruction_MemoTransferExtension

	// Create the native mint for Token-2022.
	Instruction_CreateNativeMint

	// Initialize the non-transferable extension for a mint.
	Instruction_InitializeNonTransferableMint

	// Interest-bearing mint extension instructions.
	Instruction_InterestBearingMintExtension

	// CPI guard extension instructions.
	Instruction_CpiGuardExtension

	// Initialize a permanent delegate for a mint.
	Instruction_InitializePermanentDelegate

	// Transfer hook extension instructions.
	Instruction_TransferHookExtension

	// Confidential transfer fee extension instructions.
	Instruction_ConfidentialTransferFeeExtension

	// Withdraw excess lamports from an account.
	Instruction_WithdrawExcessLamports

	// Metadata pointer extension instructions.
	Instruction_MetadataPointerExtension

	// Group pointer extension instructions.
	Instruction_GroupPointerExtension

	// Group member pointer extension instructions.
	Instruction_GroupMemberPointerExtension

	// Confidential mint/burn extension instructions.
	Instruction_ConfidentialMintBurnExtension

	// Scaled UI amount extension instructions.
	Instruction_ScaledUiAmountExtension

	// Pausable extension instructions.
	Instruction_PausableExtension

	// Unwrap native SOL lamports.
	Instruction_UnwrapLamports

	// Permissioned burn extension instructions.
	Instruction_PermissionedBurnExtension
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
	case Instruction_TransferFeeExtension:
		return "TransferFeeExtension"
	case Instruction_ConfidentialTransferExtension:
		return "ConfidentialTransferExtension"
	case Instruction_DefaultAccountStateExtension:
		return "DefaultAccountStateExtension"
	case Instruction_Reallocate:
		return "Reallocate"
	case Instruction_MemoTransferExtension:
		return "MemoTransferExtension"
	case Instruction_CreateNativeMint:
		return "CreateNativeMint"
	case Instruction_InitializeNonTransferableMint:
		return "InitializeNonTransferableMint"
	case Instruction_InterestBearingMintExtension:
		return "InterestBearingMintExtension"
	case Instruction_CpiGuardExtension:
		return "CpiGuardExtension"
	case Instruction_InitializePermanentDelegate:
		return "InitializePermanentDelegate"
	case Instruction_TransferHookExtension:
		return "TransferHookExtension"
	case Instruction_ConfidentialTransferFeeExtension:
		return "ConfidentialTransferFeeExtension"
	case Instruction_WithdrawExcessLamports:
		return "WithdrawExcessLamports"
	case Instruction_MetadataPointerExtension:
		return "MetadataPointerExtension"
	case Instruction_GroupPointerExtension:
		return "GroupPointerExtension"
	case Instruction_GroupMemberPointerExtension:
		return "GroupMemberPointerExtension"
	case Instruction_ConfidentialMintBurnExtension:
		return "ConfidentialMintBurnExtension"
	case Instruction_ScaledUiAmountExtension:
		return "ScaledUiAmountExtension"
	case Instruction_PausableExtension:
		return "PausableExtension"
	case Instruction_UnwrapLamports:
		return "UnwrapLamports"
	case Instruction_PermissionedBurnExtension:
		return "PermissionedBurnExtension"
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
		{
			Name: "GetAccountDataSize", Type: (*GetAccountDataSize)(nil),
		},
		{
			Name: "InitializeImmutableOwner", Type: (*InitializeImmutableOwner)(nil),
		},
		{
			Name: "AmountToUiAmount", Type: (*AmountToUiAmount)(nil),
		},
		{
			Name: "UiAmountToAmount", Type: (*UiAmountToAmount)(nil),
		},
		{
			Name: "InitializeMintCloseAuthority", Type: (*InitializeMintCloseAuthority)(nil),
		},
		{
			Name: "TransferFeeExtension", Type: (*TransferFeeExtension)(nil),
		},
		{
			Name: "ConfidentialTransferExtension", Type: (*ConfidentialTransferExtension)(nil),
		},
		{
			Name: "DefaultAccountStateExtension", Type: (*DefaultAccountStateExtension)(nil),
		},
		{
			Name: "Reallocate", Type: (*Reallocate)(nil),
		},
		{
			Name: "MemoTransferExtension", Type: (*MemoTransferExtension)(nil),
		},
		{
			Name: "CreateNativeMint", Type: (*CreateNativeMint)(nil),
		},
		{
			Name: "InitializeNonTransferableMint", Type: (*InitializeNonTransferableMint)(nil),
		},
		{
			Name: "InterestBearingMintExtension", Type: (*InterestBearingMintExtension)(nil),
		},
		{
			Name: "CpiGuardExtension", Type: (*CpiGuardExtension)(nil),
		},
		{
			Name: "InitializePermanentDelegate", Type: (*InitializePermanentDelegate)(nil),
		},
		{
			Name: "TransferHookExtension", Type: (*TransferHookExtension)(nil),
		},
		{
			Name: "ConfidentialTransferFeeExtension", Type: (*ConfidentialTransferFeeExtension)(nil),
		},
		{
			Name: "WithdrawExcessLamports", Type: (*WithdrawExcessLamports)(nil),
		},
		{
			Name: "MetadataPointerExtension", Type: (*MetadataPointerExtension)(nil),
		},
		{
			Name: "GroupPointerExtension", Type: (*GroupPointerExtension)(nil),
		},
		{
			Name: "GroupMemberPointerExtension", Type: (*GroupMemberPointerExtension)(nil),
		},
		{
			Name: "ConfidentialMintBurnExtension", Type: (*ConfidentialMintBurnExtension)(nil),
		},
		{
			Name: "ScaledUiAmountExtension", Type: (*ScaledUiAmountExtension)(nil),
		},
		{
			Name: "PausableExtension", Type: (*PausableExtension)(nil),
		},
		{
			Name: "UnwrapLamports", Type: (*UnwrapLamports)(nil),
		},
		{
			Name: "PermissionedBurnExtension", Type: (*PermissionedBurnExtension)(nil),
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
