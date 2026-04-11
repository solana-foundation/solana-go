package vote

import (
	"errors"
	"fmt"

	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

// VoteSwitch is a Vote with an additional proof hash for fork switching.
// Data: (Vote, Hash)
type VoteSwitch struct {
	Vote *Vote
	Hash solana.Hash

	// Same accounts as Vote.
	// [0] = [WRITE] VoteAccount
	// [1] = [] SysVarSlotHashes
	// [2] = [] SysVarClock
	// [3] = [SIGNER] VoteAuthority
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewVoteSwitchInstructionBuilder() *VoteSwitch {
	return &VoteSwitch{
		AccountMetaSlice: make(solana.AccountMetaSlice, 4),
	}
}

func NewVoteSwitchInstruction(
	slots []uint64,
	voteHash solana.Hash,
	timestamp *int64,
	proofHash solana.Hash,
	voteAccount solana.PublicKey,
	voteAuthority solana.PublicKey,
) *VoteSwitch {
	v := &Vote{Slots: slots, Hash: voteHash, Timestamp: timestamp}
	inst := NewVoteSwitchInstructionBuilder().
		SetVote(v).
		SetProofHash(proofHash).
		SetVoteAccount(voteAccount).
		SetSlotHashesSysvar(solana.SysVarSlotHashesPubkey).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetVoteAuthority(voteAuthority)
	return inst
}

func (inst *VoteSwitch) SetVote(v *Vote) *VoteSwitch {
	inst.Vote = v
	return inst
}

func (inst *VoteSwitch) SetProofHash(h solana.Hash) *VoteSwitch {
	inst.Hash = h
	return inst
}

func (inst *VoteSwitch) SetVoteAccount(pk solana.PublicKey) *VoteSwitch {
	inst.AccountMetaSlice[0] = solana.Meta(pk).WRITE()
	return inst
}
func (inst *VoteSwitch) SetSlotHashesSysvar(pk solana.PublicKey) *VoteSwitch {
	inst.AccountMetaSlice[1] = solana.Meta(pk)
	return inst
}
func (inst *VoteSwitch) SetClockSysvar(pk solana.PublicKey) *VoteSwitch {
	inst.AccountMetaSlice[2] = solana.Meta(pk)
	return inst
}
func (inst *VoteSwitch) SetVoteAuthority(pk solana.PublicKey) *VoteSwitch {
	inst.AccountMetaSlice[3] = solana.Meta(pk).SIGNER()
	return inst
}

func (inst VoteSwitch) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_VoteSwitch, bin.LE),
	}}
}

func (inst VoteSwitch) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *VoteSwitch) Validate() error {
	if inst.Vote == nil {
		return errors.New("Vote parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *VoteSwitch) UnmarshalWithDecoder(dec *bin.Decoder) error {
	inst.Vote = new(Vote)
	if err := inst.Vote.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(inst.Hash[:], b)
	return nil
}

func (inst VoteSwitch) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.Vote == nil {
		return errors.New("VoteSwitch.Vote is nil")
	}
	if err := inst.Vote.MarshalWithEncoder(enc); err != nil {
		return err
	}
	return enc.WriteBytes(inst.Hash[:], false)
}

func (inst *VoteSwitch) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("VoteSwitch")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						if inst.Vote != nil {
							paramsBranch.Child(format.Param("Slots", inst.Vote.Slots))
							paramsBranch.Child(format.Param(" Hash", inst.Vote.Hash))
						}
						paramsBranch.Child(format.Param("ProofHash", inst.Hash))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("     VoteAccount", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta("SlotHashesSysvar", inst.AccountMetaSlice[1]))
						accountsBranch.Child(format.Meta("     ClockSysvar", inst.AccountMetaSlice[2]))
						accountsBranch.Child(format.Meta("   VoteAuthority", inst.AccountMetaSlice[3]))
					})
				})
		})
}
