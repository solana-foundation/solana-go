package token2022

import (
	"encoding"
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_treeout "github.com/gagliardetto/treeout"
)

// ConfidentialTransferFee sub-instruction IDs.
const (
	ConfidentialTransferFee_InitializeConfidentialTransferFeeConfig uint8 = iota
	ConfidentialTransferFee_WithdrawWithheldTokensFromMint
	ConfidentialTransferFee_WithdrawWithheldTokensFromAccounts
	ConfidentialTransferFee_HarvestWithheldTokensToMint
	ConfidentialTransferFee_EnableHarvestToMint
	ConfidentialTransferFee_DisableHarvestToMint
)

// ConfidentialTransferFeeSubInstructionData is the data of a ConfidentialTransferFee sub-instruction,
// implemented by the ConfidentialTransferFee*Data structs.
type ConfidentialTransferFeeSubInstructionData interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	bytes() []byte
}

// ConfidentialTransferFeeExtension is the instruction wrapper for the
// ConfidentialTransferFee extension (ID 37).
// This is a complex extension involving encrypted fee amounts and ZK proofs.
type ConfidentialTransferFeeExtension struct {
	SubInstruction uint8
	// Raw data for the sub-instruction.
	RawData []byte

	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *ConfidentialTransferFeeExtension) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	return nil
}

// GetAccounts emits Signers last; a layout that does not end with the signers must inline them into Accounts and leave Signers empty.
func (slice ConfidentialTransferFeeExtension) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func (obj ConfidentialTransferFeeExtension) getName() string { return confidentialTransferFeeName }

// getSubInstructions describes every ConfidentialTransferFee sub-instruction,
// indexed by sub-instruction ID: the index is the on-chain discriminator.
func (obj ConfidentialTransferFeeExtension) getSubInstructions() []confidentialSubInstruction[ConfidentialTransferFeeSubInstructionData] {
	return []confidentialSubInstruction[ConfidentialTransferFeeSubInstructionData]{
		ConfidentialTransferFee_InitializeConfidentialTransferFeeConfig: {"InitializeConfidentialTransferFeeConfig", func() ConfidentialTransferFeeSubInstructionData {
			return &ConfidentialTransferFeeInitializeConfigData{}
		}},
		ConfidentialTransferFee_WithdrawWithheldTokensFromMint: {"WithdrawWithheldTokensFromMint", func() ConfidentialTransferFeeSubInstructionData {
			return &ConfidentialTransferFeeWithdrawWithheldTokensFromMintData{}
		}},
		ConfidentialTransferFee_WithdrawWithheldTokensFromAccounts: {"WithdrawWithheldTokensFromAccounts", func() ConfidentialTransferFeeSubInstructionData {
			return &ConfidentialTransferFeeWithdrawWithheldTokensFromAccountsData{}
		}},
		ConfidentialTransferFee_HarvestWithheldTokensToMint: {"HarvestWithheldTokensToMint", func() ConfidentialTransferFeeSubInstructionData {
			return &ConfidentialTransferFeeHarvestWithheldTokensToMintData{}
		}},
		ConfidentialTransferFee_EnableHarvestToMint: {"EnableHarvestToMint", func() ConfidentialTransferFeeSubInstructionData {
			return &ConfidentialTransferFeeEnableHarvestToMintData{}
		}},
		ConfidentialTransferFee_DisableHarvestToMint: {"DisableHarvestToMint", func() ConfidentialTransferFeeSubInstructionData {
			return &ConfidentialTransferFeeDisableHarvestToMintData{}
		}},
	}
}

func (obj ConfidentialTransferFeeExtension) getSubInstruction() uint8 { return obj.SubInstruction }

func (obj ConfidentialTransferFeeExtension) getRawData() []byte { return obj.RawData }

// DecodeSubInstructionData outputs the typed instruction data for SubInstruction.
func (obj ConfidentialTransferFeeExtension) DecodeSubInstructionData() (ConfidentialTransferFeeSubInstructionData, error) {
	return decodeConfidentialSubInstruction(obj)
}

func (inst ConfidentialTransferFeeExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   &inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_ConfidentialTransferFeeExtension),
	}}
}

func (inst ConfidentialTransferFeeExtension) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *ConfidentialTransferFeeExtension) Validate() error {
	if len(inst.Accounts) == 0 {
		return errors.New("accounts is empty")
	}
	return nil
}

func (inst *ConfidentialTransferFeeExtension) EncodeToTree(parent ag_treeout.Branches) {
	encodeConfidentialExtensionToTree(*inst, parent)
}

func (obj ConfidentialTransferFeeExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
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

func (obj *ConfidentialTransferFeeExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	obj.SubInstruction, err = decoder.ReadUint8()
	if err != nil {
		return err
	}
	obj.RawData = nil
	if remaining := decoder.Remaining(); remaining > 0 {
		obj.RawData, err = decoder.ReadNBytes(remaining)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewConfidentialTransferFeeInstruction creates a confidential transfer fee
// extension instruction from a raw sub-instruction payload.
//
// Prefer the typed NewConfidentialTransferFee*Instruction builders.
func NewConfidentialTransferFeeInstruction(
	subInstruction uint8,
	rawData []byte,
	accounts ...ag_solanago.AccountMeta,
) *ConfidentialTransferFeeExtension {
	inst := &ConfidentialTransferFeeExtension{
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
