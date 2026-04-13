package token2022

import (
	"errors"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Initialize the close account authority on a new mint.
type InitializeMintCloseAuthority struct {
	// Authority that can close the mint, or nil to not set this authority.
	CloseAuthority *ag_solanago.PublicKey `bin:"optional"`

	// [0] = [WRITE] mint
	// ··········· The mint to initialize.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewInitializeMintCloseAuthorityInstructionBuilder creates a new `InitializeMintCloseAuthority` instruction builder.
func NewInitializeMintCloseAuthorityInstructionBuilder() *InitializeMintCloseAuthority {
	nd := &InitializeMintCloseAuthority{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

// SetCloseAuthority sets the "close_authority" parameter.
// Authority that can close the mint, or nil to not set this authority.
func (inst *InitializeMintCloseAuthority) SetCloseAuthority(closeAuthority ag_solanago.PublicKey) *InitializeMintCloseAuthority {
	inst.CloseAuthority = &closeAuthority
	return inst
}

// SetMintAccount sets the "mint" account.
// The mint to initialize.
func (inst *InitializeMintCloseAuthority) SetMintAccount(mint ag_solanago.PublicKey) *InitializeMintCloseAuthority {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// GetMintAccount gets the "mint" account.
// The mint to initialize.
func (inst *InitializeMintCloseAuthority) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst InitializeMintCloseAuthority) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_InitializeMintCloseAuthority),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst InitializeMintCloseAuthority) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeMintCloseAuthority) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Mint is not set")
		}
	}
	return nil
}

func (inst *InitializeMintCloseAuthority) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("InitializeMintCloseAuthority")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("CloseAuthority (OPT)", inst.CloseAuthority))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj InitializeMintCloseAuthority) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `CloseAuthority` param (optional):
	{
		if obj.CloseAuthority == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.WriteBytes(obj.CloseAuthority[:], false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (obj *InitializeMintCloseAuthority) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `CloseAuthority` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.CloseAuthority)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NewInitializeMintCloseAuthorityInstruction declares a new InitializeMintCloseAuthority instruction with the provided parameters and accounts.
func NewInitializeMintCloseAuthorityInstruction(
	// Parameters:
	closeAuthority ag_solanago.PublicKey,
	// Accounts:
	mint ag_solanago.PublicKey,
) *InitializeMintCloseAuthority {
	return NewInitializeMintCloseAuthorityInstructionBuilder().
		SetCloseAuthority(closeAuthority).
		SetMintAccount(mint)
}
