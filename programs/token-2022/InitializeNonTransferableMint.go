package token2022

import (
	"errors"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// InitializeNonTransferableMint initializes the non-transferable extension for a mint.
// Tokens from this mint cannot be transferred.
type InitializeNonTransferableMint struct {
	// [0] = [WRITE] mint
	// ··········· The mint to initialize.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewInitializeNonTransferableMintInstructionBuilder() *InitializeNonTransferableMint {
	nd := &InitializeNonTransferableMint{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

func (inst *InitializeNonTransferableMint) SetMintAccount(mint ag_solanago.PublicKey) *InitializeNonTransferableMint {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

func (inst *InitializeNonTransferableMint) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst InitializeNonTransferableMint) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_InitializeNonTransferableMint),
	}}
}

func (inst InitializeNonTransferableMint) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeNonTransferableMint) Validate() error {
	if inst.AccountMetaSlice[0] == nil {
		return errors.New("accounts.Mint is not set")
	}
	return nil
}

func (inst *InitializeNonTransferableMint) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("InitializeNonTransferableMint")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj InitializeNonTransferableMint) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *InitializeNonTransferableMint) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func NewInitializeNonTransferableMintInstruction(
	mint ag_solanago.PublicKey,
) *InitializeNonTransferableMint {
	return NewInitializeNonTransferableMintInstructionBuilder().
		SetMintAccount(mint)
}
