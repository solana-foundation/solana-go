package token2022

import (
	"errors"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// CreateNativeMint creates the native mint (SOL) for the Token-2022 program.
type CreateNativeMint struct {
	// [0] = [WRITE, SIGNER] payer
	// ··········· The payer for the native mint creation.
	//
	// [1] = [WRITE] nativeMint
	// ··········· The native mint address.
	//
	// [2] = [] systemProgram
	// ··········· System program.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewCreateNativeMintInstructionBuilder() *CreateNativeMint {
	nd := &CreateNativeMint{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 3),
	}
	return nd
}

func (inst *CreateNativeMint) SetPayerAccount(payer ag_solanago.PublicKey) *CreateNativeMint {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(payer).WRITE().SIGNER()
	return inst
}

func (inst *CreateNativeMint) GetPayerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst *CreateNativeMint) SetNativeMintAccount(nativeMint ag_solanago.PublicKey) *CreateNativeMint {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(nativeMint).WRITE()
	return inst
}

func (inst *CreateNativeMint) GetNativeMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

func (inst *CreateNativeMint) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *CreateNativeMint {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(systemProgram)
	return inst
}

func (inst *CreateNativeMint) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[2]
}

func (inst CreateNativeMint) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_CreateNativeMint),
	}}
}

func (inst CreateNativeMint) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CreateNativeMint) Validate() error {
	if inst.AccountMetaSlice[0] == nil {
		return errors.New("accounts.Payer is not set")
	}
	if inst.AccountMetaSlice[1] == nil {
		return errors.New("accounts.NativeMint is not set")
	}
	if inst.AccountMetaSlice[2] == nil {
		return errors.New("accounts.SystemProgram is not set")
	}
	return nil
}

func (inst *CreateNativeMint) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CreateNativeMint")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("        payer", inst.AccountMetaSlice[0]))
						accountsBranch.Child(ag_format.Meta("   nativeMint", inst.AccountMetaSlice[1]))
						accountsBranch.Child(ag_format.Meta("systemProgram", inst.AccountMetaSlice[2]))
					})
				})
		})
}

func (obj CreateNativeMint) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *CreateNativeMint) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func NewCreateNativeMintInstruction(
	payer ag_solanago.PublicKey,
	nativeMint ag_solanago.PublicKey,
) *CreateNativeMint {
	return NewCreateNativeMintInstructionBuilder().
		SetPayerAccount(payer).
		SetNativeMintAccount(nativeMint).
		SetSystemProgramAccount(ag_solanago.SystemProgramID)
}
