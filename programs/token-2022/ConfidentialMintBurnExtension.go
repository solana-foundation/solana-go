package token2022

import (
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// ConfidentialMintBurn sub-instruction IDs.
const (
	ConfidentialMintBurn_InitializeMint uint8 = iota
	ConfidentialMintBurn_UpdateDecryptableSupply
	ConfidentialMintBurn_RotateSupplyElGamalPubkey
	ConfidentialMintBurn_Mint
	ConfidentialMintBurn_Burn
)

// ConfidentialMintBurnExtension is the instruction wrapper for the
// ConfidentialMintBurn extension (ID 42).
// This is a complex extension involving encrypted supply and ZK proofs.
type ConfidentialMintBurnExtension struct {
	SubInstruction uint8
	RawData        []byte

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

func (inst ConfidentialMintBurnExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
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
	names := []string{
		"InitializeMint", "UpdateDecryptableSupply",
		"RotateSupplyElGamalPubkey", "Mint", "Burn",
	}
	name := "Unknown"
	if int(inst.SubInstruction) < len(names) {
		name = names[inst.SubInstruction]
	}
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("ConfidentialMintBurn." + name)).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("RawData (len)", len(inst.RawData)))
					})
				})
		})
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
	remaining := decoder.Remaining()
	if remaining > 0 {
		obj.RawData, err = decoder.ReadNBytes(remaining)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewConfidentialMintBurnInstruction creates a raw confidential mint/burn extension instruction.
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
