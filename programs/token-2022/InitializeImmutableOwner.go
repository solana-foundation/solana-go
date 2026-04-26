package token2022

import (
	"errors"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Initialize the Immutable Owner extension for the given token account.
type InitializeImmutableOwner struct {

	// [0] = [WRITE] account
	// ··········· The account to initialize.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewInitializeImmutableOwnerInstructionBuilder creates a new `InitializeImmutableOwner` instruction builder.
func NewInitializeImmutableOwnerInstructionBuilder() *InitializeImmutableOwner {
	nd := &InitializeImmutableOwner{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

// SetAccount sets the "account" account.
// The account to initialize.
func (inst *InitializeImmutableOwner) SetAccount(account ag_solanago.PublicKey) *InitializeImmutableOwner {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(account).WRITE()
	return inst
}

// GetAccount gets the "account" account.
// The account to initialize.
func (inst *InitializeImmutableOwner) GetAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst InitializeImmutableOwner) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_InitializeImmutableOwner),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst InitializeImmutableOwner) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeImmutableOwner) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Account is not set")
		}
	}
	return nil
}

func (inst *InitializeImmutableOwner) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("InitializeImmutableOwner")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("account", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj InitializeImmutableOwner) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *InitializeImmutableOwner) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewInitializeImmutableOwnerInstruction declares a new InitializeImmutableOwner instruction with the provided parameters and accounts.
func NewInitializeImmutableOwnerInstruction(
	// Accounts:
	account ag_solanago.PublicKey,
) *InitializeImmutableOwner {
	return NewInitializeImmutableOwnerInstructionBuilder().
		SetAccount(account)
}
