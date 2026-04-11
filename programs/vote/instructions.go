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

package vote

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go/text"
	"github.com/gagliardetto/treeout"
)

var ProgramID solana.PublicKey = solana.VoteProgramID

func SetProgramID(pubkey solana.PublicKey) error {
	ProgramID = pubkey
	return solana.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "Vote"

func init() {
	solana.MustRegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

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
		{Name: "InitializeAccount", Type: (*InitializeAccount)(nil)},
		{Name: "Authorize", Type: (*Authorize)(nil)},
		{Name: "Vote", Type: (*Vote)(nil)},
		{Name: "Withdraw", Type: (*Withdraw)(nil)},
		{Name: "UpdateValidatorIdentity", Type: (*UpdateValidatorIdentity)(nil)},
		{Name: "UpdateCommission", Type: (*UpdateCommission)(nil)},
		{Name: "VoteSwitch", Type: (*VoteSwitch)(nil)},
		{Name: "AuthorizeChecked", Type: (*AuthorizeChecked)(nil)},
		{Name: "UpdateVoteState", Type: (*UpdateVoteState)(nil)},
		{Name: "UpdateVoteStateSwitch", Type: (*UpdateVoteStateSwitch)(nil)},
		{Name: "AuthorizeWithSeed", Type: (*AuthorizeWithSeed)(nil)},
		{Name: "AuthorizeCheckedWithSeed", Type: (*AuthorizeCheckedWithSeed)(nil)},
		{Name: "CompactUpdateVoteState", Type: (*CompactUpdateVoteState)(nil)},
		{Name: "CompactUpdateVoteStateSwitch", Type: (*CompactUpdateVoteStateSwitch)(nil)},
		{Name: "TowerSync", Type: (*TowerSync)(nil)},
		{Name: "TowerSyncSwitch", Type: (*TowerSyncSwitch)(nil)},
		{Name: "InitializeAccountV2", Type: (*InitializeAccountV2)(nil)},
		{Name: "UpdateCommissionCollector", Type: (*UpdateCommissionCollector)(nil)},
		{Name: "UpdateCommissionBps", Type: (*UpdateCommissionBps)(nil)},
		{Name: "DepositDelegatorRewards", Type: (*DepositDelegatorRewards)(nil)},
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
