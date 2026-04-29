package token2022

import (
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// PermissionedBurn sub-instruction IDs.
const (
	PermissionedBurn_Initialize uint8 = iota
)

// PermissionedBurnExtension is the instruction wrapper for the PermissionedBurn extension (ID 46).
// Sub-instructions: Initialize.
type PermissionedBurnExtension struct {
	SubInstruction uint8

	// [0] = [WRITE] mint
	// ··········· The mint to initialize.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewPermissionedBurnExtensionInstructionBuilder() *PermissionedBurnExtension {
	nd := &PermissionedBurnExtension{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

func (inst *PermissionedBurnExtension) SetMintAccount(mint ag_solanago.PublicKey) *PermissionedBurnExtension {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

func (inst *PermissionedBurnExtension) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst PermissionedBurnExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_PermissionedBurnExtension),
	}}
}

func (inst PermissionedBurnExtension) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *PermissionedBurnExtension) Validate() error {
	if inst.AccountMetaSlice[0] == nil {
		return errors.New("accounts.Mint is not set")
	}
	return nil
}

func (inst *PermissionedBurnExtension) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("PermissionedBurn.Initialize")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj PermissionedBurnExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return encoder.WriteUint8(obj.SubInstruction)
}

func (obj *PermissionedBurnExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	obj.SubInstruction, err = decoder.ReadUint8()
	return err
}

// NewInitializePermissionedBurnInstruction creates an instruction to initialize the permissioned burn extension.
func NewInitializePermissionedBurnInstruction(
	mint ag_solanago.PublicKey,
) *PermissionedBurnExtension {
	inst := NewPermissionedBurnExtensionInstructionBuilder()
	inst.SubInstruction = PermissionedBurn_Initialize
	inst.SetMintAccount(mint)
	return inst
}
