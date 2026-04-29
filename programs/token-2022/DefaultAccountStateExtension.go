package token2022

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// DefaultAccountState sub-instruction IDs.
const (
	DefaultAccountState_Initialize uint8 = iota
	DefaultAccountState_Update
)

// DefaultAccountStateExtension is the instruction wrapper for the DefaultAccountState extension (ID 28).
// Sub-instructions: Initialize, Update.
type DefaultAccountStateExtension struct {
	SubInstruction uint8
	State          *AccountState

	// For Initialize:
	// [0] = [WRITE] mint - The mint to initialize.
	//
	// For Update:
	// [0] = [WRITE] mint - The mint.
	// [1] = [] freezeAuthority - The mint freeze authority or multisig.
	// [2...] = [SIGNER] signers
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *DefaultAccountStateExtension) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	if obj.SubInstruction == DefaultAccountState_Initialize {
		obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	} else {
		obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(2)
	}
	return nil
}

func (slice DefaultAccountStateExtension) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func (inst DefaultAccountStateExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_DefaultAccountStateExtension),
	}}
}

func (inst DefaultAccountStateExtension) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *DefaultAccountStateExtension) Validate() error {
	if inst.State == nil {
		return errors.New("state parameter is not set")
	}
	if len(inst.Accounts) == 0 || inst.Accounts[0] == nil {
		return errors.New("accounts.Mint is not set")
	}
	if inst.SubInstruction == DefaultAccountState_Update {
		if len(inst.Accounts) < 2 || inst.Accounts[1] == nil {
			return errors.New("accounts.FreezeAuthority is not set")
		}
		if !inst.Accounts[1].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
	}
	return nil
}

func (inst *DefaultAccountStateExtension) EncodeToTree(parent ag_treeout.Branches) {
	name := "Initialize"
	if inst.SubInstruction == DefaultAccountState_Update {
		name = "Update"
	}
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("DefaultAccountState." + name)).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("State", *inst.State))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.Accounts[0]))
						if inst.SubInstruction == DefaultAccountState_Update && len(inst.Accounts) > 1 {
							accountsBranch.Child(ag_format.Meta("freezeAuthority", inst.Accounts[1]))
						}
					})
				})
		})
}

func (obj DefaultAccountStateExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	err = encoder.WriteUint8(obj.SubInstruction)
	if err != nil {
		return err
	}
	err = encoder.WriteUint8(uint8(*obj.State))
	if err != nil {
		return err
	}
	return nil
}

func (obj *DefaultAccountStateExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	obj.SubInstruction, err = decoder.ReadUint8()
	if err != nil {
		return err
	}
	v, err := decoder.ReadUint8()
	if err != nil {
		return err
	}
	state := AccountState(v)
	obj.State = &state
	return nil
}

// NewInitializeDefaultAccountStateInstruction creates an instruction to initialize
// the default account state for a mint.
func NewInitializeDefaultAccountStateInstruction(
	state AccountState,
	mint ag_solanago.PublicKey,
) *DefaultAccountStateExtension {
	inst := &DefaultAccountStateExtension{
		SubInstruction: DefaultAccountState_Initialize,
		State:          &state,
		Accounts:       make(ag_solanago.AccountMetaSlice, 1),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// NewUpdateDefaultAccountStateInstruction creates an instruction to update
// the default account state for a mint.
func NewUpdateDefaultAccountStateInstruction(
	state AccountState,
	mint ag_solanago.PublicKey,
	freezeAuthority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *DefaultAccountStateExtension {
	inst := &DefaultAccountStateExtension{
		SubInstruction: DefaultAccountState_Update,
		State:          &state,
		Accounts:       make(ag_solanago.AccountMetaSlice, 2),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	inst.Accounts[1] = ag_solanago.Meta(freezeAuthority)
	if len(multisigSigners) == 0 {
		inst.Accounts[1].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}
