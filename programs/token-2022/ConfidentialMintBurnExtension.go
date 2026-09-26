package token2022

import (
	"encoding"
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_treeout "github.com/gagliardetto/treeout"
)

// ConfidentialMintBurn sub-instruction IDs.
const (
	ConfidentialMintBurn_InitializeMint uint8 = iota
	ConfidentialMintBurn_RotateSupplyElGamalPubkey
	ConfidentialMintBurn_UpdateDecryptableSupply
	ConfidentialMintBurn_Mint
	ConfidentialMintBurn_Burn
	ConfidentialMintBurn_ApplyPendingBurn
)

// ConfidentialMintBurnSubInstructionData is the data of a ConfidentialMintBurn
// sub-instruction, implemented by the ConfidentialMintBurn*Data structs.
type ConfidentialMintBurnSubInstructionData interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	bytes() []byte
}

// ConfidentialMintBurnExtension is the instruction wrapper for the
// ConfidentialMintBurn extension (ID 42).
// This is a complex extension involving encrypted supply and ZK proofs.
type ConfidentialMintBurnExtension struct {
	SubInstruction uint8
	// Raw data for the sub-instruction.
	RawData []byte

	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *ConfidentialMintBurnExtension) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	return nil
}

func (slice ConfidentialMintBurnExtension) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func (obj ConfidentialMintBurnExtension) getName() string { return confidentialMintBurnName }

// getSubInstructions describes every ConfidentialMintBurn sub-instruction,
// indexed by sub-instruction ID: the index is the on-chain discriminator.
func (obj ConfidentialMintBurnExtension) getSubInstructions() []confidentialSubInstruction[ConfidentialMintBurnSubInstructionData] {
	return []confidentialSubInstruction[ConfidentialMintBurnSubInstructionData]{
		ConfidentialMintBurn_InitializeMint: {"InitializeMint", func() ConfidentialMintBurnSubInstructionData {
			return &ConfidentialMintBurnInitializeMintData{}
		}},
		ConfidentialMintBurn_RotateSupplyElGamalPubkey: {"RotateSupplyElGamalPubkey", func() ConfidentialMintBurnSubInstructionData {
			return &ConfidentialMintBurnRotateSupplyElGamalPubkeyData{}
		}},
		ConfidentialMintBurn_UpdateDecryptableSupply: {"UpdateDecryptableSupply", func() ConfidentialMintBurnSubInstructionData {
			return &ConfidentialMintBurnUpdateDecryptableSupplyData{}
		}},
		ConfidentialMintBurn_Mint: {"Mint", func() ConfidentialMintBurnSubInstructionData {
			return &ConfidentialMintBurnMintData{}
		}},
		ConfidentialMintBurn_Burn: {"Burn", func() ConfidentialMintBurnSubInstructionData {
			return &ConfidentialMintBurnBurnData{}
		}},
		ConfidentialMintBurn_ApplyPendingBurn: {"ApplyPendingBurn", func() ConfidentialMintBurnSubInstructionData {
			return &ConfidentialMintBurnApplyPendingBurnData{}
		}},
	}
}

func (obj ConfidentialMintBurnExtension) getSubInstruction() uint8 { return obj.SubInstruction }

func (obj ConfidentialMintBurnExtension) getRawData() []byte { return obj.RawData }

// DecodeSubInstructionData outputs the typed instruction data for SubInstruction.
func (obj ConfidentialMintBurnExtension) DecodeSubInstructionData() (ConfidentialMintBurnSubInstructionData, error) {
	return decodeConfidentialSubInstruction(obj)
}

func (inst ConfidentialMintBurnExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   &inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_ConfidentialMintBurnExtension),
	}}
}

func (inst ConfidentialMintBurnExtension) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *ConfidentialMintBurnExtension) Validate() error {
	if len(inst.Accounts) == 0 {
		return errors.New("accounts is empty")
	}
	return nil
}

func (inst *ConfidentialMintBurnExtension) EncodeToTree(parent ag_treeout.Branches) {
	encodeConfidentialExtensionToTree(*inst, parent)
}

func (obj ConfidentialMintBurnExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
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

func (obj *ConfidentialMintBurnExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
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

// NewConfidentialMintBurnInstruction creates a confidential mint/burn extension
// instruction from a raw sub-instruction payload.
//
// Prefer the typed NewConfidentialMintBurn*Instruction builders.
func NewConfidentialMintBurnInstruction(
	subInstruction uint8,
	rawData []byte,
	accounts ...ag_solanago.AccountMeta,
) *ConfidentialMintBurnExtension {
	inst := &ConfidentialMintBurnExtension{
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
