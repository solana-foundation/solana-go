package token2022

import (
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_treeout "github.com/gagliardetto/treeout"
	ag_solanago "github.com/solana-foundation/solana-go/v2"
	ag_format "github.com/solana-foundation/solana-go/v2/text/format"
)

// InitializePermanentDelegate initializes the permanent delegate extension for a mint.
// The permanent delegate can transfer or burn any tokens from any account for this mint.
type InitializePermanentDelegate struct {
	// The permanent delegate for the mint.
	Delegate *ag_solanago.PublicKey

	// [0] = [WRITE] mint
	// ··········· The mint to initialize.
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewInitializePermanentDelegateInstructionBuilder() *InitializePermanentDelegate {
	nd := &InitializePermanentDelegate{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

func (inst *InitializePermanentDelegate) SetDelegate(delegate ag_solanago.PublicKey) *InitializePermanentDelegate {
	inst.Delegate = &delegate
	return inst
}

func (inst *InitializePermanentDelegate) SetMintAccount(mint ag_solanago.PublicKey) *InitializePermanentDelegate {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

func (inst *InitializePermanentDelegate) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst InitializePermanentDelegate) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_InitializePermanentDelegate),
	}}
}

func (inst InitializePermanentDelegate) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializePermanentDelegate) Validate() error {
	if inst.Delegate == nil {
		return errors.New("delegate parameter is not set")
	}
	if inst.AccountMetaSlice[0] == nil {
		return errors.New("accounts.Mint is not set")
	}
	return nil
}

func (inst *InitializePermanentDelegate) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("InitializePermanentDelegate")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Delegate", inst.Delegate))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj InitializePermanentDelegate) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	err = encoder.Encode(obj.Delegate)
	if err != nil {
		return err
	}
	return nil
}

func (obj *InitializePermanentDelegate) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	err = decoder.Decode(&obj.Delegate)
	if err != nil {
		return err
	}
	return nil
}

func NewInitializePermanentDelegateInstruction(
	delegate ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
) *InitializePermanentDelegate {
	return NewInitializePermanentDelegateInstructionBuilder().
		SetDelegate(delegate).
		SetMintAccount(mint)
}
