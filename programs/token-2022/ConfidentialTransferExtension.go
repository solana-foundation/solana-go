package token2022

import (
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// ConfidentialTransfer sub-instruction IDs.
const (
	ConfidentialTransfer_InitializeMint uint8 = iota
	ConfidentialTransfer_UpdateMint
	ConfidentialTransfer_ConfigureAccount
	ConfidentialTransfer_ApproveAccount
	ConfidentialTransfer_EmptyAccount
	ConfidentialTransfer_Deposit
	ConfidentialTransfer_Withdraw
	ConfidentialTransfer_Transfer
	ConfidentialTransfer_ApplyPendingBalance
	ConfidentialTransfer_EnableConfidentialCredits
	ConfidentialTransfer_DisableConfidentialCredits
	ConfidentialTransfer_EnableNonConfidentialCredits
	ConfidentialTransfer_DisableNonConfidentialCredits
	ConfidentialTransfer_TransferWithFee
	ConfidentialTransfer_ConfigureAccountWithRegistry
)

// ConfidentialTransferExtension is the instruction wrapper for the ConfidentialTransfer extension (ID 27).
// This is a complex extension with many sub-instructions involving zero-knowledge proofs.
type ConfidentialTransferExtension struct {
	ag_binary.BaseVariant

	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

var ConfidentialTransferImplDef = ag_binary.NewVariantDefinition(
	ag_binary.Uint8TypeIDEncoding,
	[]ag_binary.VariantType{
		{
			Name: "InitializeMint", Type: (*ConfidentialTransferInitializeMintData)(nil),
		},
		{
			Name: "UpdateMint", Type: (*ConfidentialTransferUpdateMintData)(nil),
		},
		{
			Name: "ConfigureAccount", Type: (*ConfidentialTransferConfigureAccountData)(nil),
		},
		{
			Name: "ApproveAccount", Type: (*ConfidentialTransferApproveAccountData)(nil),
		},
		{
			Name: "EmptyAccount", Type: (*ConfidentialTransferEmptyAccountData)(nil),
		},
		{
			Name: "Deposit", Type: (*ConfidentialTransferDepositData)(nil),
		},
		{
			Name: "Withdraw", Type: (*ConfidentialTransferWithdrawData)(nil),
		},
		{
			Name: "Transfer", Type: (*ConfidentialTransferTransferData)(nil),
		},
		{
			Name: "ApplyPendingBalance", Type: (*ConfidentialTransferApplyPendingBalanceData)(nil),
		},
		{
			Name: "EnableConfidentialCredits", Type: (*ConfidentialTransferEnableConfidentialCreditsData)(nil),
		},
		{
			Name: "DisableConfidentialCredits", Type: (*ConfidentialTransferDisableConfidentialCreditsData)(nil),
		},
		{
			Name: "EnableNonConfidentialCredits", Type: (*ConfidentialTransferEnableNonConfidentialCreditsData)(nil),
		},
		{
			Name: "DisableNonConfidentialCredits", Type: (*ConfidentialTransferDisableNonConfidentialCreditsData)(nil),
		},
		{
			Name: "TransferWithFee", Type: (*ConfidentialTransferTransferWithFeeData)(nil),
		},
		{
			Name: "ConfigureAccountWithRegistry", Type: (*ConfidentialTransferConfigureAccountWithRegistryData)(nil),
		},
	},
)

func (obj *ConfidentialTransferExtension) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	return nil
}

func (slice ConfidentialTransferExtension) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func (inst ConfidentialTransferExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   &inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_ConfidentialTransferExtension),
	}}
}

func (inst ConfidentialTransferExtension) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *ConfidentialTransferExtension) Validate() error {
	if inst.Impl == nil {
		return errors.New("sub-instruction data is not set")
	}
	if len(inst.Accounts) == 0 {
		return errors.New("accounts is empty")
	}
	return nil
}

func (inst *ConfidentialTransferExtension) EncodeToTree(parent ag_treeout.Branches) {
	names := []string{
		"InitializeMint", "UpdateMint", "ConfigureAccount", "ApproveAccount",
		"EmptyAccount", "Deposit", "Withdraw", "Transfer",
		"ApplyPendingBalance", "EnableConfidentialCredits", "DisableConfidentialCredits",
		"EnableNonConfidentialCredits", "DisableNonConfidentialCredits",
		"TransferWithFee", "ConfigureAccountWithRegistry",
	}
	name := "Unknown"
	if id := int(inst.TypeID.Uint8()); id < len(names) {
		name = names[id]
	}
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("ConfidentialTransfer." + name)).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Data", inst.Impl))
					})
				})
		})
}

func (obj ConfidentialTransferExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	err := encoder.WriteUint8(obj.TypeID.Uint8())
	if err != nil {
		return err
	}
	return encoder.Encode(obj.Impl)
}

func (obj *ConfidentialTransferExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	return obj.BaseVariant.UnmarshalBinaryVariant(decoder, ConfidentialTransferImplDef)
}
