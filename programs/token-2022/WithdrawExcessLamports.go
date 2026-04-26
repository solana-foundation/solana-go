package token2022

import (
	"errors"
	"fmt"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// WithdrawExcessLamports withdraws excess lamports from a Token-2022 account.
type WithdrawExcessLamports struct {
	// [0] = [WRITE] source
	// ··········· The source account to withdraw from.
	//
	// [1] = [WRITE] destination
	// ··········· The destination account for the excess lamports.
	//
	// [2] = [] authority
	// ··········· The source account's owner or multisignature account.
	//
	// [3...] = [SIGNER] signers
	// ··········· M signer accounts.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *WithdrawExcessLamports) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(3)
	return nil
}

func (slice WithdrawExcessLamports) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func NewWithdrawExcessLamportsInstructionBuilder() *WithdrawExcessLamports {
	nd := &WithdrawExcessLamports{
		Accounts: make(ag_solanago.AccountMetaSlice, 3),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

func (inst *WithdrawExcessLamports) SetSourceAccount(source ag_solanago.PublicKey) *WithdrawExcessLamports {
	inst.Accounts[0] = ag_solanago.Meta(source).WRITE()
	return inst
}

func (inst *WithdrawExcessLamports) GetSourceAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

func (inst *WithdrawExcessLamports) SetDestinationAccount(destination ag_solanago.PublicKey) *WithdrawExcessLamports {
	inst.Accounts[1] = ag_solanago.Meta(destination).WRITE()
	return inst
}

func (inst *WithdrawExcessLamports) GetDestinationAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

func (inst *WithdrawExcessLamports) SetAuthorityAccount(authority ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *WithdrawExcessLamports {
	inst.Accounts[2] = ag_solanago.Meta(authority)
	if len(multisigSigners) == 0 {
		inst.Accounts[2].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

func (inst *WithdrawExcessLamports) GetAuthorityAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[2]
}

func (inst WithdrawExcessLamports) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_WithdrawExcessLamports),
	}}
}

func (inst WithdrawExcessLamports) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *WithdrawExcessLamports) Validate() error {
	if inst.Accounts[0] == nil {
		return errors.New("accounts.Source is not set")
	}
	if inst.Accounts[1] == nil {
		return errors.New("accounts.Destination is not set")
	}
	if inst.Accounts[2] == nil {
		return errors.New("accounts.Authority is not set")
	}
	if !inst.Accounts[2].IsSigner && len(inst.Signers) == 0 {
		return fmt.Errorf("accounts.Signers is not set")
	}
	if len(inst.Signers) > MAX_SIGNERS {
		return fmt.Errorf("too many signers; got %v, but max is 11", len(inst.Signers))
	}
	return nil
}

func (inst *WithdrawExcessLamports) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("WithdrawExcessLamports")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("     source", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("destination", inst.Accounts[1]))
						accountsBranch.Child(ag_format.Meta("  authority", inst.Accounts[2]))
					})
				})
		})
}

func (obj WithdrawExcessLamports) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *WithdrawExcessLamports) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func NewWithdrawExcessLamportsInstruction(
	source ag_solanago.PublicKey,
	destination ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *WithdrawExcessLamports {
	return NewWithdrawExcessLamportsInstructionBuilder().
		SetSourceAccount(source).
		SetDestinationAccount(destination).
		SetAuthorityAccount(authority, multisigSigners...)
}
