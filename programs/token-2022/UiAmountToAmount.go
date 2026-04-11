package token2022

import (
	"errors"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Convert a UiAmount of tokens to a little-endian u64 raw Amount, using the
// given mint's decimals.
type UiAmountToAmount struct {
	// The ui_amount of tokens to convert.
	UiAmount *string

	// [0] = [] mint
	// ··········· The mint to calculate for.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewUiAmountToAmountInstructionBuilder creates a new `UiAmountToAmount` instruction builder.
func NewUiAmountToAmountInstructionBuilder() *UiAmountToAmount {
	nd := &UiAmountToAmount{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

// SetUiAmount sets the "ui_amount" parameter.
// The ui_amount of tokens to convert.
func (inst *UiAmountToAmount) SetUiAmount(uiAmount string) *UiAmountToAmount {
	inst.UiAmount = &uiAmount
	return inst
}

// SetMintAccount sets the "mint" account.
// The mint to calculate for.
func (inst *UiAmountToAmount) SetMintAccount(mint ag_solanago.PublicKey) *UiAmountToAmount {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(mint)
	return inst
}

// GetMintAccount gets the "mint" account.
// The mint to calculate for.
func (inst *UiAmountToAmount) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst UiAmountToAmount) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_UiAmountToAmount),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst UiAmountToAmount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *UiAmountToAmount) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.UiAmount == nil {
			return errors.New("UiAmount parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Mint is not set")
		}
	}
	return nil
}

func (inst *UiAmountToAmount) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("UiAmountToAmount")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("UiAmount", *inst.UiAmount))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj UiAmountToAmount) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `UiAmount` param as raw string bytes (no length prefix).
	_, err = encoder.Write([]byte(*obj.UiAmount))
	if err != nil {
		return err
	}
	return nil
}

func (obj *UiAmountToAmount) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read remaining bytes as raw string.
	remaining := decoder.Remaining()
	if remaining > 0 {
		data, err := decoder.ReadNBytes(remaining)
		if err != nil {
			return err
		}
		s := string(data)
		obj.UiAmount = &s
	}
	return nil
}

// NewUiAmountToAmountInstruction declares a new UiAmountToAmount instruction with the provided parameters and accounts.
func NewUiAmountToAmountInstruction(
	// Parameters:
	uiAmount string,
	// Accounts:
	mint ag_solanago.PublicKey,
) *UiAmountToAmount {
	return NewUiAmountToAmountInstructionBuilder().
		SetUiAmount(uiAmount).
		SetMintAccount(mint)
}
