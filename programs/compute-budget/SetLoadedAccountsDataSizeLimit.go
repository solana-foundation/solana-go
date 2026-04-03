package computebudget

import (
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

type SetLoadedAccountsDataSizeLimit struct {
	Bytes uint32
}

func (obj *SetLoadedAccountsDataSizeLimit) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	return nil
}

func (slice SetLoadedAccountsDataSizeLimit) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	return
}

// NewSetLoadedAccountsDataSizeLimitInstructionBuilder creates a new
// `SetLoadedAccountsDataSizeLimit` instruction builder.
func NewSetLoadedAccountsDataSizeLimitInstructionBuilder() *SetLoadedAccountsDataSizeLimit {
	nd := &SetLoadedAccountsDataSizeLimit{}
	return nd
}

func (inst *SetLoadedAccountsDataSizeLimit) SetBytes(bytes uint32) *SetLoadedAccountsDataSizeLimit {
	inst.Bytes = bytes
	return inst
}

func (inst SetLoadedAccountsDataSizeLimit) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_SetLoadedAccountsDataSizeLimit),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst SetLoadedAccountsDataSizeLimit) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *SetLoadedAccountsDataSizeLimit) Validate() error {
	if inst.Bytes == 0 {
		return errors.New("bytes parameter is not set")
	}
	return nil
}

func (inst *SetLoadedAccountsDataSizeLimit) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("SetLoadedAccountsDataSizeLimit")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Bytes", inst.Bytes))
					})
				})
		})
}

func (obj SetLoadedAccountsDataSizeLimit) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	err = encoder.Encode(obj.Bytes)
	if err != nil {
		return err
	}
	return nil
}

func (obj *SetLoadedAccountsDataSizeLimit) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	err = decoder.Decode(&obj.Bytes)
	if err != nil {
		return err
	}
	return nil
}

// NewSetLoadedAccountsDataSizeLimitInstruction declares a new
// SetLoadedAccountsDataSizeLimit instruction with the provided parameters and
// accounts.
func NewSetLoadedAccountsDataSizeLimitInstruction(
	// Parameters:
	bytes uint32,
) *SetLoadedAccountsDataSizeLimit {
	return NewSetLoadedAccountsDataSizeLimitInstructionBuilder().SetBytes(bytes)
}
