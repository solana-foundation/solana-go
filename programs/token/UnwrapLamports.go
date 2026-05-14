// Copyright 2021 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package token

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Unwrap lamports from a native SOL token account, transferring them directly
// to a destination account without requiring a temporary associated token account.
//
// If Amount is nil, all lamports (the full token balance) are unwrapped.
// If Amount is set, only the specified amount is unwrapped.
//
// This instruction is only available in the p-token (Pinocchio) implementation.
type UnwrapLamports struct {
	// The amount of lamports to unwrap (optional; nil means unwrap all).
	Amount *uint64

	// [0] = [WRITE] source
	// ··········· The native SOL token account to unwrap from.
	//
	// [1] = [WRITE] destination
	// ··········· The destination account for the lamports.
	//
	// [2] = [] authority
	// ··········· The source account's owner/delegate.
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

func (inst *UnwrapLamports) SetSourceAccount(source ag_solanago.PublicKey) *UnwrapLamports {
	inst.Accounts[0] = ag_solanago.Meta(source).WRITE()
	return inst
}

func (inst *UnwrapLamports) GetSourceAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

func (inst *UnwrapLamports) SetDestinationAccount(destination ag_solanago.PublicKey) *UnwrapLamports {
	inst.Accounts[1] = ag_solanago.Meta(destination).WRITE()
	return inst
}

func (inst *UnwrapLamports) GetDestinationAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

func (inst *UnwrapLamports) SetAuthorityAccount(authority ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *UnwrapLamports {
	inst.Accounts[2] = ag_solanago.Meta(authority)
	if len(multisigSigners) == 0 {
		inst.Accounts[2].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

func (inst *UnwrapLamports) GetAuthorityAccount() *ag_solanago.AccountMeta {
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

func (inst *UnwrapLamports) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("UnwrapLamports")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						if inst.Amount != nil {
							paramsBranch.Child(ag_format.Param("Amount", *inst.Amount))
						} else {
							paramsBranch.Child(ag_format.Param("Amount", "all"))
						}
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("    source", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("destination", inst.Accounts[1]))
						accountsBranch.Child(ag_format.Meta("  authority", inst.Accounts[2]))

						signersBranch := accountsBranch.Child(fmt.Sprintf("signers[len=%v]", len(inst.Signers)))
						for i, v := range inst.Signers {
							if len(inst.Signers) > 9 && i < 10 {
								signersBranch.Child(ag_format.Meta(fmt.Sprintf(" [%v]", i), v))
							} else {
								signersBranch.Child(ag_format.Meta(fmt.Sprintf("[%v]", i), v))
							}
						}
					})
				})
		})
}

// On-chain format: u8(has_amount) + optional u64(amount)
// has_amount=0 means unwrap all, has_amount=1 means unwrap specified amount.
func (obj UnwrapLamports) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	if obj.Amount == nil {
		return encoder.WriteUint8(0)
	}
	if err = encoder.WriteUint8(1); err != nil {
		return err
	}
	return encoder.WriteUint64(*obj.Amount, ag_binary.LE)
}

func (obj *UnwrapLamports) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	hasAmount, err := decoder.ReadUint8()
	if err != nil {
		return err
	}
	if hasAmount == 0 {
		obj.Amount = nil
		return nil
	}
	amount, err := decoder.ReadUint64(ag_binary.LE)
	if err != nil {
		return err
	}
	obj.Amount = &amount
	return nil
}

func NewUnwrapLamportsInstruction(
	source ag_solanago.PublicKey,
	destination ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *UnwrapLamports {
	return NewUnwrapLamportsInstructionBuilder().
		SetSourceAccount(source).
		SetDestinationAccount(destination).
		SetAuthorityAccount(authority, multisigSigners...)
}

func NewUnwrapLamportsWithAmountInstruction(
	amount uint64,
	source ag_solanago.PublicKey,
	destination ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *UnwrapLamports {
	return NewUnwrapLamportsInstructionBuilder().
		SetAmount(amount).
		SetSourceAccount(source).
		SetDestinationAccount(destination).
		SetAuthorityAccount(authority, multisigSigners...)
}
