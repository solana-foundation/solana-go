package token2022

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Pausable sub-instruction IDs.
const (
	Pausable_Initialize uint8 = iota
	Pausable_Pause
	Pausable_Resume
)

// PausableExtension is the instruction wrapper for the Pausable extension (ID 44).
// Sub-instructions: Initialize, Pause, Resume.
type PausableExtension struct {
	SubInstruction uint8

	// For Initialize:
	// [0] = [WRITE] mint - The mint to initialize.
	//
	// For Pause/Resume:
	// [0] = [WRITE] mint - The mint.
	// [1] = [] authority - The pause authority or multisig.
	// [2...] = [SIGNER] signers
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *PausableExtension) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	if obj.SubInstruction == Pausable_Initialize {
		obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	} else {
		obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(2)
	}
	return nil
}

func (slice PausableExtension) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func (inst PausableExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_PausableExtension),
	}}
}

func (inst PausableExtension) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *PausableExtension) Validate() error {
	if len(inst.Accounts) == 0 || inst.Accounts[0] == nil {
		return errors.New("accounts.Mint is not set")
	}
	if inst.SubInstruction == Pausable_Pause || inst.SubInstruction == Pausable_Resume {
		if len(inst.Accounts) < 2 || inst.Accounts[1] == nil {
			return errors.New("accounts.Authority is not set")
		}
		if !inst.Accounts[1].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
	}
	return nil
}

func (inst *PausableExtension) EncodeToTree(parent ag_treeout.Branches) {
	names := []string{"Initialize", "Pause", "Resume"}
	name := "Unknown"
	if int(inst.SubInstruction) < len(names) {
		name = names[inst.SubInstruction]
	}
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Pausable." + name)).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("     mint", inst.Accounts[0]))
						if len(inst.Accounts) > 1 && inst.Accounts[1] != nil {
							accountsBranch.Child(ag_format.Meta("authority", inst.Accounts[1]))
						}
					})
				})
		})
}

func (obj PausableExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return encoder.WriteUint8(obj.SubInstruction)
}

func (obj *PausableExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	obj.SubInstruction, err = decoder.ReadUint8()
	return err
}

// NewInitializePausableInstruction creates an instruction to initialize the pausable extension.
func NewInitializePausableInstruction(
	mint ag_solanago.PublicKey,
) *PausableExtension {
	inst := &PausableExtension{
		SubInstruction: Pausable_Initialize,
		Accounts:       make(ag_solanago.AccountMetaSlice, 1),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

func newPausableToggleInstruction(
	subInstruction uint8,
	mint ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *PausableExtension {
	inst := &PausableExtension{
		SubInstruction: subInstruction,
		Accounts:       make(ag_solanago.AccountMetaSlice, 2),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	inst.Accounts[1] = ag_solanago.Meta(authority)
	if len(multisigSigners) == 0 {
		inst.Accounts[1].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

// NewPauseInstruction creates an instruction to pause the mint.
func NewPauseInstruction(
	mint ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *PausableExtension {
	return newPausableToggleInstruction(Pausable_Pause, mint, authority, multisigSigners)
}

// NewResumeInstruction creates an instruction to resume the mint.
func NewResumeInstruction(
	mint ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *PausableExtension {
	return newPausableToggleInstruction(Pausable_Resume, mint, authority, multisigSigners)
}
