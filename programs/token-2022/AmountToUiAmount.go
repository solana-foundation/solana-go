package token2022

import (
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_treeout "github.com/gagliardetto/treeout"
	ag_solanago "github.com/solana-foundation/solana-go/v2"
	ag_format "github.com/solana-foundation/solana-go/v2/text/format"
)

// Convert an Amount of tokens to a UiAmount string, using the given mint's decimals.
type AmountToUiAmount struct {
	// The amount of tokens to convert.
	Amount *uint64

	// [0] = [] mint
	// ··········· The mint to calculate for.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewAmountToUiAmountInstructionBuilder creates a new `AmountToUiAmount` instruction builder.
func NewAmountToUiAmountInstructionBuilder() *AmountToUiAmount {
	nd := &AmountToUiAmount{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

// SetAmount sets the "amount" parameter.
// The amount of tokens to convert.
func (inst *AmountToUiAmount) SetAmount(amount uint64) *AmountToUiAmount {
	inst.Amount = &amount
	return inst
}

// SetMintAccount sets the "mint" account.
// The mint to calculate for.
func (inst *AmountToUiAmount) SetMintAccount(mint ag_solanago.PublicKey) *AmountToUiAmount {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(mint)
	return inst
}

// GetMintAccount gets the "mint" account.
// The mint to calculate for.
func (inst *AmountToUiAmount) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst AmountToUiAmount) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_AmountToUiAmount),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst AmountToUiAmount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *AmountToUiAmount) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Amount == nil {
			return errors.New("amount parameter is not set")
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

func (inst *AmountToUiAmount) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("AmountToUiAmount")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Amount", *inst.Amount))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj AmountToUiAmount) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Amount` param:
	err = encoder.Encode(obj.Amount)
	if err != nil {
		return err
	}
	return nil
}

func (obj *AmountToUiAmount) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Amount`:
	err = decoder.Decode(&obj.Amount)
	if err != nil {
		return err
	}
	return nil
}

// NewAmountToUiAmountInstruction declares a new AmountToUiAmount instruction with the provided parameters and accounts.
func NewAmountToUiAmountInstruction(
	// Parameters:
	amount uint64,
	// Accounts:
	mint ag_solanago.PublicKey,
) *AmountToUiAmount {
	return NewAmountToUiAmountInstructionBuilder().
		SetAmount(amount).
		SetMintAccount(mint)
}
