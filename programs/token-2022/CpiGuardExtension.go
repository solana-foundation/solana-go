package token2022

import (
	"errors"
	"fmt"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// CpiGuard sub-instruction IDs.
const (
	CpiGuard_Enable uint8 = iota
	CpiGuard_Disable
)

// CpiGuardExtension is the instruction wrapper for the CpiGuard extension (ID 34).
// Sub-instructions: Enable, Disable.
type CpiGuardExtension struct {
	SubInstruction uint8

	// [0] = [WRITE] account
	// ··········· The token account to update.
	//
	// [1] = [] owner
	// ··········· The account's owner or multisig.
	//
	// [2...] = [SIGNER] signers
	// ··········· M signer accounts.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *CpiGuardExtension) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(2)
	return nil
}

func (slice CpiGuardExtension) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func (inst CpiGuardExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_CpiGuardExtension),
	}}
}

func (inst CpiGuardExtension) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CpiGuardExtension) Validate() error {
	if len(inst.Accounts) < 2 || inst.Accounts[0] == nil {
		return errors.New("accounts.Account is not set")
	}
	if inst.Accounts[1] == nil {
		return errors.New("accounts.Owner is not set")
	}
	if !inst.Accounts[1].IsSigner && len(inst.Signers) == 0 {
		return fmt.Errorf("accounts.Signers is not set")
	}
	return nil
}

func (inst *CpiGuardExtension) EncodeToTree(parent ag_treeout.Branches) {
	name := "Enable"
	if inst.SubInstruction == CpiGuard_Disable {
		name = "Disable"
	}
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CpiGuard." + name)).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("account", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("  owner", inst.Accounts[1]))
					})
				})
		})
}

func (obj CpiGuardExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return encoder.WriteUint8(obj.SubInstruction)
}

func (obj *CpiGuardExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	obj.SubInstruction, err = decoder.ReadUint8()
	return err
}

func newCpiGuardInstruction(
	subInstruction uint8,
	account ag_solanago.PublicKey,
	owner ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *CpiGuardExtension {
	inst := &CpiGuardExtension{
		SubInstruction: subInstruction,
		Accounts:       make(ag_solanago.AccountMetaSlice, 2),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(account).WRITE()
	inst.Accounts[1] = ag_solanago.Meta(owner)
	if len(multisigSigners) == 0 {
		inst.Accounts[1].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

// NewEnableCpiGuardInstruction creates an instruction to enable CPI guard.
func NewEnableCpiGuardInstruction(
	account ag_solanago.PublicKey,
	owner ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *CpiGuardExtension {
	return newCpiGuardInstruction(CpiGuard_Enable, account, owner, multisigSigners)
}

// NewDisableCpiGuardInstruction creates an instruction to disable CPI guard.
func NewDisableCpiGuardInstruction(
	account ag_solanago.PublicKey,
	owner ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *CpiGuardExtension {
	return newCpiGuardInstruction(CpiGuard_Disable, account, owner, multisigSigners)
}
