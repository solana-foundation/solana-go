package vote

import (
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

// UpdateValidatorIdentity updates the validator identity of a vote account.
// No instruction data.
type UpdateValidatorIdentity struct {
	// [0] = [WRITE] VoteAccount
	// [1] = [SIGNER] NewIdentity
	// [2] = [SIGNER] WithdrawAuthority
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewUpdateValidatorIdentityInstructionBuilder() *UpdateValidatorIdentity {
	return &UpdateValidatorIdentity{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
	}
}

func NewUpdateValidatorIdentityInstruction(
	voteAccount solana.PublicKey,
	newIdentity solana.PublicKey,
	withdrawAuthority solana.PublicKey,
) *UpdateValidatorIdentity {
	return NewUpdateValidatorIdentityInstructionBuilder().
		SetVoteAccount(voteAccount).
		SetNewIdentity(newIdentity).
		SetWithdrawAuthority(withdrawAuthority)
}

func (inst *UpdateValidatorIdentity) SetVoteAccount(pk solana.PublicKey) *UpdateValidatorIdentity {
	inst.AccountMetaSlice[0] = solana.Meta(pk).WRITE()
	return inst
}

func (inst *UpdateValidatorIdentity) SetNewIdentity(pk solana.PublicKey) *UpdateValidatorIdentity {
	inst.AccountMetaSlice[1] = solana.Meta(pk).SIGNER()
	return inst
}

func (inst *UpdateValidatorIdentity) SetWithdrawAuthority(pk solana.PublicKey) *UpdateValidatorIdentity {
	inst.AccountMetaSlice[2] = solana.Meta(pk).SIGNER()
	return inst
}

func (inst UpdateValidatorIdentity) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_UpdateValidatorIdentity, bin.LE),
	}}
}

func (inst UpdateValidatorIdentity) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *UpdateValidatorIdentity) Validate() error {
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *UpdateValidatorIdentity) UnmarshalWithDecoder(dec *bin.Decoder) error { return nil }
func (inst UpdateValidatorIdentity) MarshalWithEncoder(enc *bin.Encoder) error    { return nil }

func (inst *UpdateValidatorIdentity) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("UpdateValidatorIdentity")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("      VoteAccount", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta("      NewIdentity", inst.AccountMetaSlice[1]))
						accountsBranch.Child(format.Meta("WithdrawAuthority", inst.AccountMetaSlice[2]))
					})
				})
		})
}
