// Copyright 2024 github.com/cordialsys
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stake

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

type Withdraw struct {
	// Withdraw unstaked lamports from the stake account
	Lamports *uint64
	// [0] = [WRITE] Stake Account
	// ··········· Stake account from which to withdraw
	//
	// [1] = [WRITE] Recipient Account
	// ··········· Recipient account
	//
	// [2] = [] Clock Sysvar
	// ··········· The Clock Sysvar Account
	//
	// [3] = [] Stake History Sysvar
	// ··········· The Stake History Sysvar Account
	//
	// [4] = [SIGNER] Withdraw Authority
	// ··········· Withdraw authority
	//
	// OPTIONAL:
	// [5] = [SIGNER] Lockup authority
	// ··········· If before lockup expiration
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *Withdraw) Validate() error {
	{
		if inst.Lamports == nil {
			return errors.New("lamports parameter is not set")
		}
	}
	// Check whether all accounts are set:
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}
func (inst *Withdraw) SetStakeAccount(stakeAccount solana.PublicKey) *Withdraw {
	inst.AccountMetaSlice[0] = solana.Meta(stakeAccount).WRITE()
	return inst
}
func (inst *Withdraw) SetRecipientAccount(recipient solana.PublicKey) *Withdraw {
	inst.AccountMetaSlice[1] = solana.Meta(recipient).WRITE()
	return inst
}
func (inst *Withdraw) SetClockSysvar(clockSysvar solana.PublicKey) *Withdraw {
	inst.AccountMetaSlice[2] = solana.Meta(clockSysvar)
	return inst
}
func (inst *Withdraw) SetStakeHistorySysvar(historySysvar solana.PublicKey) *Withdraw {
	inst.AccountMetaSlice[3] = solana.Meta(historySysvar)
	return inst
}

func (inst *Withdraw) SetWithdrawAuthority(withdrawAuthority solana.PublicKey) *Withdraw {
	inst.AccountMetaSlice[4] = solana.Meta(withdrawAuthority).SIGNER()
	return inst
}

func (inst *Withdraw) GetStakeAccount() *solana.AccountMeta       { return inst.AccountMetaSlice[0] }
func (inst *Withdraw) GetRecipientAccount() *solana.AccountMeta   { return inst.AccountMetaSlice[1] }
func (inst *Withdraw) GetClockSysvar() *solana.AccountMeta        { return inst.AccountMetaSlice[2] }
func (inst *Withdraw) GetStakeHistorySysvar() *solana.AccountMeta { return inst.AccountMetaSlice[3] }
func (inst *Withdraw) GetWithdrawAuthority() *solana.AccountMeta  { return inst.AccountMetaSlice[4] }

func (inst *Withdraw) SetLamports(lamports uint64) *Withdraw {
	inst.Lamports = &lamports
	return inst
}

func (inst *Withdraw) UnmarshalWithDecoder(dec *bin.Decoder) error {
	{
		err := dec.Decode(&inst.Lamports)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst *Withdraw) MarshalWithEncoder(encoder *bin.Encoder) error {
	{
		err := encoder.WriteUint64(*inst.Lamports, binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst Withdraw) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_Withdraw, bin.LE),
	}}
}

func (inst *Withdraw) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("Withdraw")).
				//
				ParentFunc(func(instructionBranch treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("Lamports", inst.Lamports))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("                   StakeAccount", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("               RecipientAccount", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("                    ClockSysvar", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(format.Meta("             StakeHistorySysvar", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(format.Meta("              WithdrawAuthority", inst.AccountMetaSlice.Get(4)))
					})
				})
		})
}

// NewWithdrawInstructionBuilder creates a new `Withdraw` instruction builder.
func NewWithdrawInstructionBuilder() *Withdraw {
	nd := &Withdraw{
		AccountMetaSlice: make(solana.AccountMetaSlice, 5),
	}
	return nd
}

// NewWithdrawInstruction declares a new Withdraw instruction with the provided parameters and accounts.
func NewWithdrawInstruction(
	// Params:
	lamports uint64,
	// Accounts:
	stakeAccount solana.PublicKey,
	recipient solana.PublicKey,
	withdrawAuthority solana.PublicKey,
) *Withdraw {
	return NewWithdrawInstructionBuilder().
		SetLamports(lamports).
		SetStakeAccount(stakeAccount).
		SetRecipientAccount(recipient).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetStakeHistorySysvar(solana.SysVarStakeHistoryPubkey).
		SetWithdrawAuthority(withdrawAuthority)
}
