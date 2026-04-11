package token2022

import (
	"errors"

	ag_binary "github.com/gagliardetto/solana-go/binary"
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
	ConfidentialTransfer_TransferWithSplitProofs
	ConfidentialTransfer_TransferWithSplitProofsInParallel
)

// ConfidentialTransferExtension is the instruction wrapper for the ConfidentialTransfer extension (ID 27).
// This is a complex extension with many sub-instructions involving zero-knowledge proofs.
// The raw sub-instruction data is preserved for encoding/decoding.
type ConfidentialTransferExtension struct {
	SubInstruction uint8
	// Raw data for the sub-instruction (after the sub-instruction byte).
	RawData []byte

	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

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
		Impl:   inst,
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
		"TransferWithSplitProofs", "TransferWithSplitProofsInParallel",
	}
	name := "Unknown"
	if int(inst.SubInstruction) < len(names) {
		name = names[inst.SubInstruction]
	}
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("ConfidentialTransfer." + name)).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("RawData (len)", len(inst.RawData)))
					})
				})
		})
}

func (obj ConfidentialTransferExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	err = encoder.WriteUint8(obj.SubInstruction)
	if err != nil {
		return err
	}
	if len(obj.RawData) > 0 {
		err = encoder.WriteBytes(obj.RawData, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (obj *ConfidentialTransferExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	obj.SubInstruction, err = decoder.ReadUint8()
	if err != nil {
		return err
	}
	remaining := decoder.Remaining()
	if remaining > 0 {
		obj.RawData, err = decoder.ReadNBytes(remaining)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewConfidentialTransferInstruction creates a raw confidential transfer extension instruction.
// Due to the complexity of ZK proof data, this provides a low-level interface.
func NewConfidentialTransferInstruction(
	subInstruction uint8,
	rawData []byte,
	accounts ...ag_solanago.AccountMeta,
) *ConfidentialTransferExtension {
	inst := &ConfidentialTransferExtension{
		SubInstruction: subInstruction,
		RawData:        rawData,
		Accounts:       make(ag_solanago.AccountMetaSlice, len(accounts)),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	for i := range accounts {
		inst.Accounts[i] = &accounts[i]
	}
	return inst
}
