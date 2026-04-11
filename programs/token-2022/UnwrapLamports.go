package token2022

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// UnwrapLamports unwraps native SOL tokens by burning the wrapped tokens and
// transferring the underlying lamports to a destination account.
// If amount is nil, all tokens are unwrapped.
type UnwrapLamports struct {
	// The amount of tokens to unwrap, or nil to unwrap all.
	Amount *uint64 `bin:"optional"`

	// [0] = [WRITE] account
	// ··········· The token account to unwrap from.
	//
	// [1] = [WRITE] destination
	// ··········· The destination account for the lamports.
	//
	// [2] = [] owner
	// ··········· The account's owner or its multisignature account.
	//
	// [3...] = [SIGNER] signers
	// ··········· M signer accounts.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *UnwrapLamports) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(3)
	return nil
}

func (slice UnwrapLamports) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func NewUnwrapLamportsInstructionBuilder() *UnwrapLamports {
	nd := &UnwrapLamports{
		Accounts: make(ag_solanago.AccountMetaSlice, 3),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

func (inst *UnwrapLamports) SetAmount(amount uint64) *UnwrapLamports {
	inst.Amount = &amount
	return inst
}

func (inst *UnwrapLamports) SetAccountAccount(account ag_solanago.PublicKey) *UnwrapLamports {
	inst.Accounts[0] = ag_solanago.Meta(account).WRITE()
	return inst
}

func (inst *UnwrapLamports) GetAccountAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

func (inst *UnwrapLamports) SetDestinationAccount(destination ag_solanago.PublicKey) *UnwrapLamports {
	inst.Accounts[1] = ag_solanago.Meta(destination).WRITE()
	return inst
}

func (inst *UnwrapLamports) GetDestinationAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

func (inst *UnwrapLamports) SetOwnerAccount(owner ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *UnwrapLamports {
	inst.Accounts[2] = ag_solanago.Meta(owner)
	if len(multisigSigners) == 0 {
		inst.Accounts[2].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

func (inst *UnwrapLamports) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[2]
}

func (inst UnwrapLamports) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_UnwrapLamports),
	}}
}

func (inst UnwrapLamports) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *UnwrapLamports) Validate() error {
	if inst.Accounts[0] == nil {
		return errors.New("accounts.Account is not set")
	}
	if inst.Accounts[1] == nil {
		return errors.New("accounts.Destination is not set")
	}
	if inst.Accounts[2] == nil {
		return errors.New("accounts.Owner is not set")
	}
	if !inst.Accounts[2].IsSigner && len(inst.Signers) == 0 {
		return fmt.Errorf("accounts.Signers is not set")
	}
	if len(inst.Signers) > MAX_SIGNERS {
		return fmt.Errorf("too many signers; got %v, but max is 11", len(inst.Signers))
	}
	return nil
}

func (inst *UnwrapLamports) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("UnwrapLamports")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Amount (OPT)", inst.Amount))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("    account", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("destination", inst.Accounts[1]))
						accountsBranch.Child(ag_format.Meta("      owner", inst.Accounts[2]))
					})
				})
		})
}

func (obj UnwrapLamports) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// COption<u64>: 1 byte discriminator + optional 8 bytes
	if obj.Amount == nil {
		err = encoder.WriteBool(false)
		if err != nil {
			return err
		}
	} else {
		err = encoder.WriteBool(true)
		if err != nil {
			return err
		}
		err = encoder.Encode(obj.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (obj *UnwrapLamports) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	ok, err := decoder.ReadBool()
	if err != nil {
		return err
	}
	if ok {
		err = decoder.Decode(&obj.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewUnwrapLamportsInstruction(
	amount *uint64,
	account ag_solanago.PublicKey,
	destination ag_solanago.PublicKey,
	owner ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *UnwrapLamports {
	inst := NewUnwrapLamportsInstructionBuilder().
		SetAccountAccount(account).
		SetDestinationAccount(destination).
		SetOwnerAccount(owner, multisigSigners...)
	if amount != nil {
		inst.SetAmount(*amount)
	}
	return inst
}
