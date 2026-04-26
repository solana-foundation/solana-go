package vote

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

// UpdateCommissionCollector updates the collector account for a commission bucket.
// Data: CommissionKind (u8)
type UpdateCommissionCollector struct {
	Kind *CommissionKind

	// [0] = [WRITE] VoteAccount
	// [1] = [WRITE] NewCollector
	// [2] = [SIGNER] WithdrawAuthority
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewUpdateCommissionCollectorInstructionBuilder() *UpdateCommissionCollector {
	return &UpdateCommissionCollector{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
	}
}

func NewUpdateCommissionCollectorInstruction(
	kind CommissionKind,
	voteAccount solana.PublicKey,
	newCollector solana.PublicKey,
	withdrawAuthority solana.PublicKey,
) *UpdateCommissionCollector {
	inst := NewUpdateCommissionCollectorInstructionBuilder()
	inst.Kind = &kind
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	inst.AccountMetaSlice[1] = solana.Meta(newCollector).WRITE()
	inst.AccountMetaSlice[2] = solana.Meta(withdrawAuthority).SIGNER()
	return inst
}

func (inst UpdateCommissionCollector) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_UpdateCommissionCollector, bin.LE),
	}}
}

func (inst UpdateCommissionCollector) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *UpdateCommissionCollector) Validate() error {
	if inst.Kind == nil {
		return errors.New("kind parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *UpdateCommissionCollector) UnmarshalWithDecoder(dec *bin.Decoder) error {
	v, err := dec.ReadUint8()
	if err != nil {
		return err
	}
	k := CommissionKind(v)
	inst.Kind = &k
	return nil
}

func (inst UpdateCommissionCollector) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.Kind == nil {
		return errors.New("UpdateCommissionCollector.Kind is nil")
	}
	return enc.WriteUint8(uint8(*inst.Kind))
}

func (inst *UpdateCommissionCollector) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("UpdateCommissionCollector")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("Kind", inst.Kind))
					})
				})
		})
}

// UpdateCommissionBps updates the commission rate (in basis points) for a bucket.
// Data: { commission_bps: u16, kind: CommissionKind }
type UpdateCommissionBps struct {
	CommissionBps *uint16
	Kind          *CommissionKind

	// [0] = [WRITE] VoteAccount
	// [1] = [SIGNER] WithdrawAuthority
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewUpdateCommissionBpsInstructionBuilder() *UpdateCommissionBps {
	return &UpdateCommissionBps{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
	}
}

func NewUpdateCommissionBpsInstruction(
	commissionBps uint16,
	kind CommissionKind,
	voteAccount solana.PublicKey,
	withdrawAuthority solana.PublicKey,
) *UpdateCommissionBps {
	inst := NewUpdateCommissionBpsInstructionBuilder()
	inst.CommissionBps = &commissionBps
	inst.Kind = &kind
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	inst.AccountMetaSlice[1] = solana.Meta(withdrawAuthority).SIGNER()
	return inst
}

func (inst UpdateCommissionBps) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_UpdateCommissionBps, bin.LE),
	}}
}

func (inst UpdateCommissionBps) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *UpdateCommissionBps) Validate() error {
	if inst.CommissionBps == nil {
		return errors.New("CommissionBps parameter is not set")
	}
	if inst.Kind == nil {
		return errors.New("kind parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *UpdateCommissionBps) UnmarshalWithDecoder(dec *bin.Decoder) error {
	bps, err := dec.ReadUint16(binary.LittleEndian)
	if err != nil {
		return err
	}
	inst.CommissionBps = &bps
	v, err := dec.ReadUint8()
	if err != nil {
		return err
	}
	k := CommissionKind(v)
	inst.Kind = &k
	return nil
}

func (inst UpdateCommissionBps) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.CommissionBps == nil {
		return errors.New("UpdateCommissionBps.CommissionBps is nil")
	}
	if inst.Kind == nil {
		return errors.New("UpdateCommissionBps.Kind is nil")
	}
	if err := enc.WriteUint16(*inst.CommissionBps, binary.LittleEndian); err != nil {
		return err
	}
	return enc.WriteUint8(uint8(*inst.Kind))
}

func (inst *UpdateCommissionBps) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("UpdateCommissionBps")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("CommissionBps", inst.CommissionBps))
						paramsBranch.Child(format.Param("         Kind", inst.Kind))
					})
				})
		})
}

// DepositDelegatorRewards deposits delegator rewards into a vote account.
// Data: { deposit: u64 }
type DepositDelegatorRewards struct {
	Deposit *uint64

	// [0] = [WRITE] VoteAccount
	// [1] = [WRITE, SIGNER] Depositor
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewDepositDelegatorRewardsInstructionBuilder() *DepositDelegatorRewards {
	return &DepositDelegatorRewards{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
	}
}

func NewDepositDelegatorRewardsInstruction(
	deposit uint64,
	voteAccount solana.PublicKey,
	depositor solana.PublicKey,
) *DepositDelegatorRewards {
	inst := NewDepositDelegatorRewardsInstructionBuilder()
	inst.Deposit = &deposit
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	inst.AccountMetaSlice[1] = solana.Meta(depositor).WRITE().SIGNER()
	return inst
}

func (inst DepositDelegatorRewards) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_DepositDelegatorRewards, bin.LE),
	}}
}

func (inst DepositDelegatorRewards) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *DepositDelegatorRewards) Validate() error {
	if inst.Deposit == nil {
		return errors.New("deposit parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *DepositDelegatorRewards) UnmarshalWithDecoder(dec *bin.Decoder) error {
	v, err := dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	inst.Deposit = &v
	return nil
}

func (inst DepositDelegatorRewards) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.Deposit == nil {
		return errors.New("DepositDelegatorRewards.Deposit is nil")
	}
	return enc.WriteUint64(*inst.Deposit, binary.LittleEndian)
}

func (inst *DepositDelegatorRewards) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("DepositDelegatorRewards")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("Deposit", inst.Deposit))
					})
				})
		})
}
