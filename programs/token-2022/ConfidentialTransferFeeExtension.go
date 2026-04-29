package token2022

import (
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
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

// ConfidentialTransferFeeExtension is the instruction wrapper for the
// ConfidentialTransferFee extension (ID 37).
// This is a complex extension involving encrypted fee amounts and ZK proofs.
type ConfidentialTransferFeeExtension struct {
	SubInstruction uint8
	RawData        []byte

	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *ConfidentialTransferFeeExtension) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	return nil
}

func (slice ConfidentialTransferFeeExtension) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func (inst ConfidentialTransferFeeExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
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
	names := []string{
		"InitializeConfidentialTransferFeeConfig",
		"WithdrawWithheldTokensFromMint",
		"WithdrawWithheldTokensFromAccounts",
		"HarvestWithheldTokensToMint",
		"EnableHarvestToMint",
		"DisableHarvestToMint",
	}
	name := "Unknown"
	if int(inst.SubInstruction) < len(names) {
		name = names[inst.SubInstruction]
	}
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("ConfidentialTransferFee." + name)).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("RawData (len)", len(inst.RawData)))
					})
				})
		})
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
	remaining := decoder.Remaining()
	if remaining > 0 {
		obj.RawData, err = decoder.ReadNBytes(remaining)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewConfidentialTransferFeeInstruction creates a raw confidential transfer fee extension instruction.
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
