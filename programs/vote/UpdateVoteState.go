package vote

import (
	"errors"
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

// UpdateVoteState is the instruction that updates the vote state tower.
// Data: VoteStateUpdate.
type UpdateVoteState struct {
	Update *VoteStateUpdate

	// [0] = [WRITE] VoteAccount
	// [1] = [SIGNER] VoteAuthority
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewUpdateVoteStateInstructionBuilder() *UpdateVoteState {
	return &UpdateVoteState{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
	}
}

func NewUpdateVoteStateInstruction(
	update VoteStateUpdate,
	voteAccount solana.PublicKey,
	voteAuthority solana.PublicKey,
) *UpdateVoteState {
	return NewUpdateVoteStateInstructionBuilder().
		SetUpdate(update).
		SetVoteAccount(voteAccount).
		SetVoteAuthority(voteAuthority)
}

func (inst *UpdateVoteState) SetUpdate(u VoteStateUpdate) *UpdateVoteState {
	inst.Update = &u
	return inst
}
func (inst *UpdateVoteState) SetVoteAccount(pk solana.PublicKey) *UpdateVoteState {
	inst.AccountMetaSlice[0] = solana.Meta(pk).WRITE()
	return inst
}
func (inst *UpdateVoteState) SetVoteAuthority(pk solana.PublicKey) *UpdateVoteState {
	inst.AccountMetaSlice[1] = solana.Meta(pk).SIGNER()
	return inst
}

func (inst UpdateVoteState) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_UpdateVoteState, bin.LE),
	}}
}

func (inst UpdateVoteState) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *UpdateVoteState) Validate() error {
	if inst.Update == nil {
		return errors.New("update parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *UpdateVoteState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	inst.Update = new(VoteStateUpdate)
	return inst.Update.UnmarshalWithDecoder(dec)
}

func (inst UpdateVoteState) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.Update == nil {
		return errors.New("UpdateVoteState.Update is nil")
	}
	return inst.Update.MarshalWithEncoder(enc)
}

func (inst *UpdateVoteState) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("UpdateVoteState")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						if inst.Update != nil {
							paramsBranch.Child(format.Param("#Lockouts", len(inst.Update.Lockouts)))
							paramsBranch.Child(format.Param("     Hash", inst.Update.Hash))
						}
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("  VoteAccount", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta("VoteAuthority", inst.AccountMetaSlice[1]))
					})
				})
		})
}

// UpdateVoteStateSwitch is UpdateVoteState with a proof hash for fork switching.
// Data: (VoteStateUpdate, Hash)
type UpdateVoteStateSwitch struct {
	Update *VoteStateUpdate
	Hash   solana.Hash

	// [0] = [WRITE] VoteAccount
	// [1] = [SIGNER] VoteAuthority
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewUpdateVoteStateSwitchInstructionBuilder() *UpdateVoteStateSwitch {
	return &UpdateVoteStateSwitch{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
	}
}

func NewUpdateVoteStateSwitchInstruction(
	update VoteStateUpdate,
	proofHash solana.Hash,
	voteAccount solana.PublicKey,
	voteAuthority solana.PublicKey,
) *UpdateVoteStateSwitch {
	inst := NewUpdateVoteStateSwitchInstructionBuilder()
	inst.Update = &update
	inst.Hash = proofHash
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	inst.AccountMetaSlice[1] = solana.Meta(voteAuthority).SIGNER()
	return inst
}

func (inst UpdateVoteStateSwitch) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_UpdateVoteStateSwitch, bin.LE),
	}}
}

func (inst UpdateVoteStateSwitch) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *UpdateVoteStateSwitch) Validate() error {
	if inst.Update == nil {
		return errors.New("update parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *UpdateVoteStateSwitch) UnmarshalWithDecoder(dec *bin.Decoder) error {
	inst.Update = new(VoteStateUpdate)
	if err := inst.Update.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(inst.Hash[:], b)
	return nil
}

func (inst UpdateVoteStateSwitch) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.Update == nil {
		return errors.New("UpdateVoteStateSwitch.Update is nil")
	}
	if err := inst.Update.MarshalWithEncoder(enc); err != nil {
		return err
	}
	return enc.WriteBytes(inst.Hash[:], false)
}

func (inst *UpdateVoteStateSwitch) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("UpdateVoteStateSwitch")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("ProofHash", inst.Hash))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("  VoteAccount", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta("VoteAuthority", inst.AccountMetaSlice[1]))
					})
				})
		})
}

// CompactUpdateVoteState uses the Solana compact serde wire format:
// short_vec lockout offsets, varint deltas, u64 root (u64::MAX = None).
// Absolute lockout slots are delta-encoded at marshal time and reconstructed
// at unmarshal time. Users supply absolute slots in the VoteStateUpdate.
type CompactUpdateVoteState struct {
	Update *VoteStateUpdate

	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewCompactUpdateVoteStateInstructionBuilder() *CompactUpdateVoteState {
	return &CompactUpdateVoteState{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
	}
}

func NewCompactUpdateVoteStateInstruction(
	update VoteStateUpdate,
	voteAccount solana.PublicKey,
	voteAuthority solana.PublicKey,
) *CompactUpdateVoteState {
	inst := NewCompactUpdateVoteStateInstructionBuilder()
	inst.Update = &update
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	inst.AccountMetaSlice[1] = solana.Meta(voteAuthority).SIGNER()
	return inst
}

func (inst CompactUpdateVoteState) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_CompactUpdateVoteState, bin.LE),
	}}
}

func (inst CompactUpdateVoteState) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CompactUpdateVoteState) Validate() error {
	if inst.Update == nil {
		return errors.New("update parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *CompactUpdateVoteState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	inst.Update = new(VoteStateUpdate)
	return unmarshalCompactVoteStateUpdate(dec, inst.Update)
}

func (inst CompactUpdateVoteState) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.Update == nil {
		return errors.New("CompactUpdateVoteState.Update is nil")
	}
	return marshalCompactVoteStateUpdate(enc, inst.Update)
}

func (inst *CompactUpdateVoteState) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("CompactUpdateVoteState")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						if inst.Update != nil {
							paramsBranch.Child(format.Param("#Lockouts", len(inst.Update.Lockouts)))
						}
					})
				})
		})
}

// CompactUpdateVoteStateSwitch is CompactUpdateVoteState with a trailing
// proof hash for fork switching. Uses the same compact wire format as
// CompactUpdateVoteState.
type CompactUpdateVoteStateSwitch struct {
	Update *VoteStateUpdate
	Hash   solana.Hash

	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewCompactUpdateVoteStateSwitchInstructionBuilder() *CompactUpdateVoteStateSwitch {
	return &CompactUpdateVoteStateSwitch{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
	}
}

func NewCompactUpdateVoteStateSwitchInstruction(
	update VoteStateUpdate,
	proofHash solana.Hash,
	voteAccount solana.PublicKey,
	voteAuthority solana.PublicKey,
) *CompactUpdateVoteStateSwitch {
	inst := NewCompactUpdateVoteStateSwitchInstructionBuilder()
	inst.Update = &update
	inst.Hash = proofHash
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	inst.AccountMetaSlice[1] = solana.Meta(voteAuthority).SIGNER()
	return inst
}

func (inst CompactUpdateVoteStateSwitch) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_CompactUpdateVoteStateSwitch, bin.LE),
	}}
}

func (inst CompactUpdateVoteStateSwitch) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CompactUpdateVoteStateSwitch) Validate() error {
	if inst.Update == nil {
		return errors.New("update parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *CompactUpdateVoteStateSwitch) UnmarshalWithDecoder(dec *bin.Decoder) error {
	inst.Update = new(VoteStateUpdate)
	if err := unmarshalCompactVoteStateUpdate(dec, inst.Update); err != nil {
		return err
	}
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(inst.Hash[:], b)
	return nil
}

func (inst CompactUpdateVoteStateSwitch) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.Update == nil {
		return errors.New("CompactUpdateVoteStateSwitch.Update is nil")
	}
	if err := marshalCompactVoteStateUpdate(enc, inst.Update); err != nil {
		return err
	}
	return enc.WriteBytes(inst.Hash[:], false)
}

func (inst *CompactUpdateVoteStateSwitch) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("CompactUpdateVoteStateSwitch")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("ProofHash", inst.Hash))
					})
				})
		})
}
