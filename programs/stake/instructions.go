// Copyright 2021 github.com/gagliardetto
// Copyright 2024 github.com/cordialsys
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

package stake

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/treeout"
	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/text"
)

var ProgramID solana.PublicKey = solana.StakeProgramID

func SetProgramID(pubkey solana.PublicKey) error {
	ProgramID = pubkey
	return solana.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "Stake"

func init() {
	solana.MustRegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const (
	// Initializes a new stake account
	Instruction_Initialize uint32 = iota
	// Authorize a key to manage stake or withdrawal
	Instruction_Authorize
	// Delegate a stake account to a validator vote account
	Instruction_DelegateStake
	// Split an active stake account into a new stake account
	Instruction_Split
	// Withdraw unstaked lamports from the stake account
	Instruction_Withdraw
	// Deactivates the stake in the account
	Instruction_Deactivate
	// Set lockup for a stake account
	Instruction_SetLockup
	// Merge two stake accounts
	Instruction_Merge
	// Authorize a key to manage stake or withdrawal with a derived key
	Instruction_AuthorizeWithSeed
	// Initialize a stake account with checked authorities
	Instruction_InitializeChecked
	// Authorize with checked new authority as signer
	Instruction_AuthorizeChecked
	// Authorize with checked new authority as signer using a derived key
	Instruction_AuthorizeCheckedWithSeed
	// Set lockup with checked new lockup authority as signer
	Instruction_SetLockupChecked
	// Get the minimum delegation
	Instruction_GetMinimumDelegation
	// Deactivate delinquent stake
	Instruction_DeactivateDelinquent
	// Redelegate (deprecated)
	Instruction_Redelegate
	// Move stake between accounts with the same authorities
	Instruction_MoveStake
	// Move lamports between accounts with the same authorities
	Instruction_MoveLamports
)

type Instruction struct {
	bin.BaseVariant
}

func (inst *Instruction) EncodeToTree(parent treeout.Branches) {
	if enToTree, ok := inst.Impl.(text.EncodableToTree); ok {
		enToTree.EncodeToTree(parent)
	} else {
		parent.Child(spew.Sdump(inst))
	}
}

var InstructionImplDef = bin.NewVariantDefinition(
	bin.Uint32TypeIDEncoding,
	[]bin.VariantType{
		{Name: "Initialize", Type: (*Initialize)(nil)},
		{Name: "Authorize", Type: (*Authorize)(nil)},
		{Name: "DelegateStake", Type: (*DelegateStake)(nil)},
		{Name: "Split", Type: (*Split)(nil)},
		{Name: "Withdraw", Type: (*Withdraw)(nil)},
		{Name: "Deactivate", Type: (*Deactivate)(nil)},
		{Name: "SetLockup", Type: (*SetLockup)(nil)},
		{Name: "Merge", Type: (*Merge)(nil)},
		{Name: "AuthorizeWithSeed", Type: (*AuthorizeWithSeed)(nil)},
		{Name: "InitializeChecked", Type: (*InitializeChecked)(nil)},
		{Name: "AuthorizeChecked", Type: (*AuthorizeChecked)(nil)},
		{Name: "AuthorizeCheckedWithSeed", Type: (*AuthorizeCheckedWithSeed)(nil)},
		{Name: "SetLockupChecked", Type: (*SetLockupChecked)(nil)},
		{Name: "GetMinimumDelegation", Type: (*GetMinimumDelegation)(nil)},
		{Name: "DeactivateDelinquent", Type: (*DeactivateDelinquent)(nil)},
		{Name: "Redelegate", Type: (*Redelegate)(nil)},
		{Name: "MoveStake", Type: (*MoveStake)(nil)},
		{Name: "MoveLamports", Type: (*MoveLamports)(nil)},
	},
)

func (inst *Instruction) ProgramID() solana.PublicKey {
	return ProgramID
}

func (inst *Instruction) Accounts() (out []*solana.AccountMeta) {
	return inst.Impl.(solana.AccountsGettable).GetAccounts()
}

func (inst *Instruction) Data() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := bin.NewBinEncoder(buf).Encode(inst); err != nil {
		return nil, fmt.Errorf("unable to encode instruction: %w", err)
	}
	return buf.Bytes(), nil
}

func (inst *Instruction) TextEncode(encoder *text.Encoder, option *text.Option) error {
	return encoder.Encode(inst.Impl, option)
}

func (inst *Instruction) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	return inst.BaseVariant.UnmarshalBinaryVariant(decoder, InstructionImplDef)
}

func (inst Instruction) MarshalWithEncoder(encoder *bin.Encoder) error {
	err := encoder.WriteUint32(inst.TypeID.Uint32(), binary.LittleEndian)
	if err != nil {
		return fmt.Errorf("unable to write variant type: %w", err)
	}
	return encoder.Encode(inst.Impl)
}

func registryDecodeInstruction(accounts []*solana.AccountMeta, data []byte) (any, error) {
	inst, err := DecodeInstruction(accounts, data)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func DecodeInstruction(accounts []*solana.AccountMeta, data []byte) (*Instruction, error) {
	inst := new(Instruction)
	if err := bin.NewBinDecoder(data).Decode(inst); err != nil {
		return nil, fmt.Errorf("unable to decode instruction: %w", err)
	}
	if v, ok := inst.Impl.(solana.AccountsSettable); ok {
		err := v.SetAccounts(accounts)
		if err != nil {
			return nil, fmt.Errorf("unable to set accounts for instruction: %w", err)
		}
	}
	return inst, nil
}
