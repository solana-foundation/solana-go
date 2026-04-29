package vote

import (
	"errors"
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

// UpdateCommission updates the commission of a vote account.
// Data: u8 commission.
type UpdateCommission struct {
	Commission *uint8

	// [0] = [WRITE] VoteAccount
	// [1] = [SIGNER] WithdrawAuthority
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewUpdateCommissionInstructionBuilder() *UpdateCommission {
	return &UpdateCommission{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
	}
}

func NewUpdateCommissionInstruction(
	commission uint8,
	voteAccount solana.PublicKey,
	withdrawAuthority solana.PublicKey,
) *UpdateCommission {
	return NewUpdateCommissionInstructionBuilder().
		SetCommission(commission).
		SetVoteAccount(voteAccount).
		SetWithdrawAuthority(withdrawAuthority)
}

func (inst *UpdateCommission) SetCommission(c uint8) *UpdateCommission {
	inst.Commission = &c
	return inst
}

func (inst *UpdateCommission) SetVoteAccount(pk solana.PublicKey) *UpdateCommission {
	inst.AccountMetaSlice[0] = solana.Meta(pk).WRITE()
	return inst
}

func (inst *UpdateCommission) SetWithdrawAuthority(pk solana.PublicKey) *UpdateCommission {
	inst.AccountMetaSlice[1] = solana.Meta(pk).SIGNER()
	return inst
}

func (inst UpdateCommission) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_UpdateCommission, bin.LE),
	}}
}

func (inst UpdateCommission) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *UpdateCommission) Validate() error {
	if inst.Commission == nil {
		return errors.New("commission parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *UpdateCommission) UnmarshalWithDecoder(dec *bin.Decoder) error {
	v, err := dec.ReadUint8()
	if err != nil {
		return err
	}
	inst.Commission = &v
	return nil
}

func (inst UpdateCommission) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.Commission == nil {
		return errors.New("UpdateCommission.Commission is nil")
	}
	return enc.WriteUint8(*inst.Commission)
}

func (inst *UpdateCommission) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("UpdateCommission")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("Commission", inst.Commission))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("      VoteAccount", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta("WithdrawAuthority", inst.AccountMetaSlice[1]))
					})
				})
		})
}
