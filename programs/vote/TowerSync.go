package vote

import (
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

// TowerSync is the TowerSync instruction (the current consensus mechanism,
// replacing UpdateVoteState).
//
// The data field is a TowerSyncUpdate which is serialized using the Solana
// compact serde format (short_vec lengths, varint lockout offsets, delta
// encoding) plus a trailing block_id hash.
//
// Data: TowerSyncUpdate (instruction ID 14).
type TowerSync struct {
	Sync *TowerSyncUpdate

	// [0] = [WRITE] VoteAccount
	// [1] = [SIGNER] VoteAuthority
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewTowerSyncInstructionBuilder creates a new TowerSync instruction builder.
func NewTowerSyncInstructionBuilder() *TowerSync {
	return &TowerSync{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
	}
}

// NewTowerSyncInstruction builds a ready-to-use TowerSync instruction.
func NewTowerSyncInstruction(
	sync TowerSyncUpdate,
	voteAccount solana.PublicKey,
	voteAuthority solana.PublicKey,
) *TowerSync {
	inst := NewTowerSyncInstructionBuilder()
	inst.Sync = &sync
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	inst.AccountMetaSlice[1] = solana.Meta(voteAuthority).SIGNER()
	return inst
}

func (inst TowerSync) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_TowerSync, bin.LE),
	}}
}

func (inst TowerSync) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *TowerSync) Validate() error {
	if inst.Sync == nil {
		return errors.New("sync parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *TowerSync) UnmarshalWithDecoder(dec *bin.Decoder) error {
	inst.Sync = new(TowerSyncUpdate)
	return inst.Sync.UnmarshalWithDecoder(dec)
}

func (inst TowerSync) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.Sync == nil {
		return errors.New("TowerSync.Sync is nil")
	}
	return inst.Sync.MarshalWithEncoder(enc)
}

func (inst *TowerSync) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("TowerSync")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						if inst.Sync != nil {
							paramsBranch.Child(format.Param("#Lockouts", len(inst.Sync.Lockouts)))
							paramsBranch.Child(format.Param("  BlockID", inst.Sync.BlockID))
						}
					})
				})
		})
}

// ---- TowerSyncSwitch ----

// TowerSyncSwitch is TowerSync with an additional proof hash for fork switching.
// Data: (TowerSyncUpdate, Hash) — same compact wire format as TowerSync
// followed by a 32-byte proof hash.
type TowerSyncSwitch struct {
	Sync *TowerSyncUpdate
	Hash solana.Hash

	// [0] = [WRITE] VoteAccount
	// [1] = [SIGNER] VoteAuthority
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewTowerSyncSwitchInstructionBuilder creates a new TowerSyncSwitch instruction builder.
func NewTowerSyncSwitchInstructionBuilder() *TowerSyncSwitch {
	return &TowerSyncSwitch{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
	}
}

// NewTowerSyncSwitchInstruction builds a ready-to-use TowerSyncSwitch instruction.
func NewTowerSyncSwitchInstruction(
	sync TowerSyncUpdate,
	proofHash solana.Hash,
	voteAccount solana.PublicKey,
	voteAuthority solana.PublicKey,
) *TowerSyncSwitch {
	inst := NewTowerSyncSwitchInstructionBuilder()
	inst.Sync = &sync
	inst.Hash = proofHash
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	inst.AccountMetaSlice[1] = solana.Meta(voteAuthority).SIGNER()
	return inst
}

func (inst TowerSyncSwitch) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_TowerSyncSwitch, bin.LE),
	}}
}

func (inst TowerSyncSwitch) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *TowerSyncSwitch) Validate() error {
	if inst.Sync == nil {
		return errors.New("sync parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *TowerSyncSwitch) UnmarshalWithDecoder(dec *bin.Decoder) error {
	inst.Sync = new(TowerSyncUpdate)
	if err := inst.Sync.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(inst.Hash[:], b)
	return nil
}

func (inst TowerSyncSwitch) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.Sync == nil {
		return errors.New("TowerSyncSwitch.Sync is nil")
	}
	if err := inst.Sync.MarshalWithEncoder(enc); err != nil {
		return err
	}
	return enc.WriteBytes(inst.Hash[:], false)
}

func (inst *TowerSyncSwitch) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("TowerSyncSwitch")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("ProofHash", inst.Hash))
					})
				})
		})
}
