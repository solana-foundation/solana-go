package vote

import (
	"errors"
	"fmt"

	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

// InitializeAccountV2 initializes a vote account with V2 (Alpenglow) parameters.
// Data: VoteInitV2
type InitializeAccountV2 struct {
	VoteInit *VoteInitV2

	// [0] = [WRITE] VoteAccount
	// [1] = [] SysVarRent
	// [2] = [] SysVarClock
	// [3] = [SIGNER] NodePubkey
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewInitializeAccountV2InstructionBuilder() *InitializeAccountV2 {
	return &InitializeAccountV2{
		AccountMetaSlice: make(solana.AccountMetaSlice, 4),
	}
}

func NewInitializeAccountV2Instruction(
	voteInit VoteInitV2,
	voteAccount solana.PublicKey,
) *InitializeAccountV2 {
	inst := NewInitializeAccountV2InstructionBuilder()
	inst.VoteInit = &voteInit
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	inst.AccountMetaSlice[1] = solana.Meta(solana.SysVarRentPubkey)
	inst.AccountMetaSlice[2] = solana.Meta(solana.SysVarClockPubkey)
	inst.AccountMetaSlice[3] = solana.Meta(voteInit.NodePubkey).SIGNER()
	return inst
}

func (inst InitializeAccountV2) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_InitializeAccountV2, bin.LE),
	}}
}

func (inst InitializeAccountV2) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeAccountV2) Validate() error {
	if inst.VoteInit == nil {
		return errors.New("VoteInit parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *InitializeAccountV2) UnmarshalWithDecoder(dec *bin.Decoder) error {
	inst.VoteInit = new(VoteInitV2)
	return inst.VoteInit.UnmarshalWithDecoder(dec)
}

func (inst InitializeAccountV2) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.VoteInit == nil {
		return errors.New("InitializeAccountV2.VoteInit is nil")
	}
	return inst.VoteInit.MarshalWithEncoder(enc)
}

func (inst *InitializeAccountV2) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("InitializeAccountV2")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						if inst.VoteInit != nil {
							paramsBranch.Child(format.Param("NodePubkey", inst.VoteInit.NodePubkey))
							paramsBranch.Child(format.Param("InflationRewardsBps", inst.VoteInit.InflationRewardsCommissionBps))
							paramsBranch.Child(format.Param("BlockRevenueBps", inst.VoteInit.BlockRevenueCommissionBps))
						}
					})
				})
		})
}
