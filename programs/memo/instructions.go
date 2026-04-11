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

package memo

import (
	"bytes"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_text "github.com/gagliardetto/solana-go/text"
	"github.com/gagliardetto/treeout"
)

var ProgramID = ag_solanago.MemoProgramID

func SetProgramID(pubkey ag_solanago.PublicKey) error {
	ProgramID = pubkey
	return ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

func init() {
	ag_solanago.MustRegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

type MemoInstruction struct {
	ag_binary.BaseVariant
}

func (inst *MemoInstruction) EncodeToTree(parent treeout.Branches) {
	if enToTree, ok := inst.Impl.(ag_text.EncodableToTree); ok {
		enToTree.EncodeToTree(parent)
	} else {
		parent.Child(spew.Sdump(inst))
	}
}

var InstructionImplDef = ag_binary.NewVariantDefinition(
	ag_binary.NoTypeIDEncoding,
	[]ag_binary.VariantType{
		{
			Name: "Memo", Type: (*Create)(nil),
		},
	},
)

func (inst *MemoInstruction) ProgramID() ag_solanago.PublicKey {
	return ProgramID
}

func (inst *MemoInstruction) Accounts() (out []*ag_solanago.AccountMeta) {
	return inst.Impl.(ag_solanago.AccountsGettable).GetAccounts()
}

func (inst *MemoInstruction) Data() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := ag_binary.NewBinEncoder(buf).Encode(inst.Impl); err != nil {
		return nil, fmt.Errorf("unable to encode instruction: %w", err)
	}
	return buf.Bytes(), nil
}

func (inst *MemoInstruction) TextEncode(encoder *ag_text.Encoder, option *ag_text.Option) error {
	return encoder.Encode(inst.Impl, option)
}

func (inst *MemoInstruction) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	return inst.BaseVariant.UnmarshalBinaryVariant(decoder, InstructionImplDef)
}

func (inst MemoInstruction) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return encoder.Encode(inst.Impl)
}

func registryDecodeInstruction(accounts []*ag_solanago.AccountMeta, data []byte) (any, error) {
	inst, err := DecodeInstruction(accounts, data)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func DecodeInstruction(accounts []*ag_solanago.AccountMeta, data []byte) (*MemoInstruction, error) {
	inst := new(MemoInstruction)
	if err := ag_binary.NewBinDecoder(data).Decode(inst); err != nil {
		return nil, fmt.Errorf("unable to decode instruction: %w", err)
	}
	if v, ok := inst.Impl.(ag_solanago.AccountsSettable); ok {
		err := v.SetAccounts(accounts)
		if err != nil {
			return nil, fmt.Errorf("unable to set accounts for instruction: %w", err)
		}
	}
	return inst, nil
}
