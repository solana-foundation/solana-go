// Copyright 2021 github.com/gagliardetto
// Copyright 2023 github.com/jumpcrypto
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

package vote

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
	// Number of lamports to withdraw from the vote account
	Lamports *uint64

	// [0] = [WRITE] VoteAccount
	// ··········· Vote account to withdraw from
	//
	// [1] = [WRITE] ToAccount
	// ··········· Account to receive the funds
	//
	// [2] = [WRITE SIGNER] AuthorizedWithdrawerPubkey
	// ··········· Account authorized to do the witdraw
	//
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (v *Withdraw) UnmarshalWithDecoder(dec *bin.Decoder) error {
	// Deserialize `Lamports` param:
	{
		err := dec.Decode(&v.Lamports)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst *Withdraw) MarshalWithEncoder(encoder *bin.Encoder) error {
	// Serialize `Lamports` param:
	{
		err := encoder.WriteUint64(*inst.Lamports, binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst *Withdraw) Validate() error {
	// Check whether all (required) parameters are set:
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

// Vote account
func (inst *Withdraw) SetVoteAccount(voteAccount solana.PublicKey) *Withdraw {
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	return inst
}

// Recipient account
func (inst *Withdraw) SetRecipientAccount(recipientAccount solana.PublicKey) *Withdraw {
	inst.AccountMetaSlice[1] = solana.Meta(recipientAccount).WRITE()
	return inst
}

// Withdraw authority account
func (inst *Withdraw) SetWithdrawAuthorityAccount(withdrawAccount solana.PublicKey) *Withdraw {
	inst.AccountMetaSlice[2] = solana.Meta(withdrawAccount).WRITE().SIGNER()
	return inst
}

func (inst *Withdraw) GetVoteAccount() *solana.AccountMeta      { return inst.AccountMetaSlice[0] }
func (inst *Withdraw) GetRecipientAccount() *solana.AccountMeta { return inst.AccountMetaSlice[1] }
func (inst *Withdraw) GetWithdrawAuthorityAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[2]
}

// Number of lamports to transfer to the recipient account
func (inst *Withdraw) SetLamports(lamports uint64) *Withdraw {
	inst.Lamports = &lamports
	return inst
}

func (inst Withdraw) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_Withdraw, bin.LE),
	}}
}

func (inst Withdraw) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
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
						accountsBranch.Child(format.Meta("                Vote", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(format.Meta("           Recipient", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(format.Meta("AuthorizedWithdrawer", inst.AccountMetaSlice.Get(2)))
					})
				})
		})
}

// NewWithdrawInstructionBuilder creates a new `Withdraw` instruction builder.
func NewWithdrawInstructionBuilder() *Withdraw {
	nd := &Withdraw{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
	}
	return nd
}

// NewWithdrawInstruction declares a new Withdraw instruction with the provided parameters and accounts.
func NewWithdrawInstruction(
	// Parameters:
	lamports uint64,
	// Accounts:
	voteAccount solana.PublicKey,
	recipientAccount solana.PublicKey,
	withdrawAuthAccount solana.PublicKey,
) *Withdraw {
	return NewWithdrawInstructionBuilder().
		SetLamports(lamports).
		SetVoteAccount(voteAccount).
		SetRecipientAccount(recipientAccount).
		SetWithdrawAuthorityAccount(withdrawAuthAccount)
}
