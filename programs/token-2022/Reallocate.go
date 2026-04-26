package token2022

import (
	"encoding/binary"
	"errors"
	"fmt"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Reallocate an account to hold the given list of extensions.
type Reallocate struct {
	// The extension types to reallocate for.
	ExtensionTypes []ExtensionType

	// [0] = [WRITE] account
	// ··········· The account to reallocate.
	//
	// [1] = [WRITE, SIGNER] payer
	// ··········· The payer for the additional rent.
	//
	// [2] = [] systemProgram
	// ··········· System program for reallocation.
	//
	// [3] = [] owner
	// ··········· The account's owner or its multisignature account.
	//
	// [4...] = [SIGNER] signers
	// ··········· M signer accounts.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *Reallocate) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(4)
	return nil
}

func (slice Reallocate) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func NewReallocateInstructionBuilder() *Reallocate {
	nd := &Reallocate{
		Accounts: make(ag_solanago.AccountMetaSlice, 4),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

func (inst *Reallocate) SetExtensionTypes(extensionTypes []ExtensionType) *Reallocate {
	inst.ExtensionTypes = extensionTypes
	return inst
}

func (inst *Reallocate) SetAccountAccount(account ag_solanago.PublicKey) *Reallocate {
	inst.Accounts[0] = ag_solanago.Meta(account).WRITE()
	return inst
}

func (inst *Reallocate) GetAccountAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

func (inst *Reallocate) SetPayerAccount(payer ag_solanago.PublicKey) *Reallocate {
	inst.Accounts[1] = ag_solanago.Meta(payer).WRITE().SIGNER()
	return inst
}

func (inst *Reallocate) GetPayerAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

func (inst *Reallocate) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *Reallocate {
	inst.Accounts[2] = ag_solanago.Meta(systemProgram)
	return inst
}

func (inst *Reallocate) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[2]
}

func (inst *Reallocate) SetOwnerAccount(owner ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *Reallocate {
	inst.Accounts[3] = ag_solanago.Meta(owner)
	if len(multisigSigners) == 0 {
		inst.Accounts[3].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

func (inst *Reallocate) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[3]
}

func (inst Reallocate) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_Reallocate),
	}}
}

func (inst Reallocate) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Reallocate) Validate() error {
	if inst.Accounts[0] == nil {
		return errors.New("accounts.Account is not set")
	}
	if inst.Accounts[1] == nil {
		return errors.New("accounts.Payer is not set")
	}
	if inst.Accounts[2] == nil {
		return errors.New("accounts.SystemProgram is not set")
	}
	if inst.Accounts[3] == nil {
		return errors.New("accounts.Owner is not set")
	}
	if !inst.Accounts[3].IsSigner && len(inst.Signers) == 0 {
		return fmt.Errorf("accounts.Signers is not set")
	}
	if len(inst.Signers) > MAX_SIGNERS {
		return fmt.Errorf("too many signers; got %v, but max is 11", len(inst.Signers))
	}
	return nil
}

func (inst *Reallocate) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Reallocate")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("ExtensionTypes", inst.ExtensionTypes))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("      account", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("        payer", inst.Accounts[1]))
						accountsBranch.Child(ag_format.Meta("systemProgram", inst.Accounts[2]))
						accountsBranch.Child(ag_format.Meta("        owner", inst.Accounts[3]))
					})
				})
		})
}

func (obj Reallocate) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	for _, et := range obj.ExtensionTypes {
		err = encoder.WriteUint16(uint16(et), binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

func (obj *Reallocate) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	for {
		val, err := decoder.ReadUint16(binary.LittleEndian)
		if err != nil {
			break
		}
		obj.ExtensionTypes = append(obj.ExtensionTypes, ExtensionType(val))
	}
	return nil
}

func NewReallocateInstruction(
	extensionTypes []ExtensionType,
	account ag_solanago.PublicKey,
	payer ag_solanago.PublicKey,
	owner ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *Reallocate {
	return NewReallocateInstructionBuilder().
		SetExtensionTypes(extensionTypes).
		SetAccountAccount(account).
		SetPayerAccount(payer).
		SetSystemProgramAccount(ag_solanago.SystemProgramID).
		SetOwnerAccount(owner, multisigSigners...)
}
