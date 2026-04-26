package token2022

import (
	"encoding/binary"
	"errors"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Gets the required size of an account for the given mint as a little-endian
// u64, including any extensions that are required for the mint.
type GetAccountDataSize struct {
	// The extension types to include in the account size calculation.
	ExtensionTypes []ExtensionType

	// [0] = [] mint
	// ··········· The mint to calculate for.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewGetAccountDataSizeInstructionBuilder creates a new `GetAccountDataSize` instruction builder.
func NewGetAccountDataSizeInstructionBuilder() *GetAccountDataSize {
	nd := &GetAccountDataSize{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

// SetExtensionTypes sets the "extension_types" parameter.
func (inst *GetAccountDataSize) SetExtensionTypes(extensionTypes []ExtensionType) *GetAccountDataSize {
	inst.ExtensionTypes = extensionTypes
	return inst
}

// SetMintAccount sets the "mint" account.
// The mint to calculate for.
func (inst *GetAccountDataSize) SetMintAccount(mint ag_solanago.PublicKey) *GetAccountDataSize {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(mint)
	return inst
}

// GetMintAccount gets the "mint" account.
// The mint to calculate for.
func (inst *GetAccountDataSize) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst GetAccountDataSize) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_GetAccountDataSize),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst GetAccountDataSize) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *GetAccountDataSize) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Mint is not set")
		}
	}
	return nil
}

func (inst *GetAccountDataSize) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("GetAccountDataSize")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("ExtensionTypes", inst.ExtensionTypes))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj GetAccountDataSize) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize extension types as raw u16 values (no length prefix).
	for _, et := range obj.ExtensionTypes {
		err = encoder.WriteUint16(uint16(et), binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

func (obj *GetAccountDataSize) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read extension types until EOF.
	for {
		val, err := decoder.ReadUint16(binary.LittleEndian)
		if err != nil {
			break
		}
		obj.ExtensionTypes = append(obj.ExtensionTypes, ExtensionType(val))
	}
	return nil
}

// NewGetAccountDataSizeInstruction declares a new GetAccountDataSize instruction with the provided parameters and accounts.
func NewGetAccountDataSizeInstruction(
	// Parameters:
	extensionTypes []ExtensionType,
	// Accounts:
	mint ag_solanago.PublicKey,
) *GetAccountDataSize {
	return NewGetAccountDataSizeInstructionBuilder().
		SetExtensionTypes(extensionTypes).
		SetMintAccount(mint)
}
