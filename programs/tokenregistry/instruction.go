// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
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

package tokenregistry

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go/text"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
)

func init() {
	solana.MustRegisterInstructionDecoder(ProgramID(), registryDecodeInstruction)
}

func registryDecodeInstruction(accounts []*solana.AccountMeta, data []byte) (any, error) {
	inst, err := DecodeInstruction(accounts, data)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func DecodeInstruction(accounts []*solana.AccountMeta, data []byte) (*Instruction, error) {
	var inst Instruction
	if err := bin.NewBinDecoder(data).Decode(&inst); err != nil {
		return nil, fmt.Errorf("unable to decode instruction for token registry program: %w", err)
	}

	if v, ok := inst.Impl.(solana.AccountsSettable); ok {
		err := v.SetAccounts(accounts)
		if err != nil {
			return nil, fmt.Errorf("unable to set accounts for instruction: %w", err)
		}
	}

	return &inst, nil
}

func NewRegisterTokenInstruction(logo Logo, name Name, symbol Symbol, website Website, tokenMetaKey, ownerKey, tokenKey solana.PublicKey) *Instruction {
	return &Instruction{
		BaseVariant: bin.BaseVariant{
			TypeID: bin.TypeIDFromUint32(0, bin.LE),
			Impl: &RegisterToken{
				Logo:    logo,
				Name:    name,
				Website: website,
				Symbol:  symbol,
				Accounts: &RegisterTokenAccounts{
					TokenMeta: &solana.AccountMeta{PublicKey: tokenMetaKey, IsWritable: false, IsSigner: true},
					Owner:     &solana.AccountMeta{PublicKey: ownerKey, IsWritable: true, IsSigner: false},
					Token:     &solana.AccountMeta{PublicKey: tokenKey, IsWritable: false, IsSigner: false},
				},
			},
		},
	}
}

type Instruction struct {
	bin.BaseVariant
}

var _ bin.EncoderDecoder = &Instruction{}

func (i *Instruction) Accounts() (out []*solana.AccountMeta) {
	switch i.TypeID {
	case bin.TypeIDFromUint32(0, bin.LE):
		accounts := i.Impl.(*RegisterToken).Accounts
		out = []*solana.AccountMeta{accounts.TokenMeta, accounts.Owner, accounts.Token}
	}
	return
}

func (i *Instruction) ProgramID() solana.PublicKey {
	return ProgramID()
}

func (i *Instruction) Data() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := bin.NewBinEncoder(buf).Encode(i); err != nil {
		return nil, fmt.Errorf("unable to encode instruction: %w", err)
	}
	return buf.Bytes(), nil
}

var InstructionDefVariant = bin.NewVariantDefinition(bin.Uint32TypeIDEncoding, []bin.VariantType{
	{Name: "register_token", Type: (*RegisterToken)(nil)},
})

func (i *Instruction) TextEncode(encoder *text.Encoder, option *text.Option) error {
	return encoder.Encode(i.Impl, option)
}

func (i *Instruction) UnmarshalWithDecoder(decoder *bin.Decoder) (err error) {
	return i.BaseVariant.UnmarshalBinaryVariant(decoder, InstructionDefVariant)
}

func (i Instruction) MarshalWithEncoder(encoder *bin.Encoder) error {
	err := encoder.WriteUint32(i.TypeID.Uint32(), binary.LittleEndian)
	if err != nil {
		return fmt.Errorf("unable to write variant type: %w", err)
	}
	return encoder.Encode(i.Impl)
}

type RegisterTokenAccounts struct {
	TokenMeta *solana.AccountMeta `text:"linear,notype"`
	Owner     *solana.AccountMeta `text:"linear,notype"`
	Token     *solana.AccountMeta `text:"linear,notype"`
}

type RegisterToken struct {
	Logo     Logo
	Name     Name
	Website  Website
	Symbol   Symbol
	Accounts *RegisterTokenAccounts `bin:"-"`
}

func (i *RegisterToken) SetAccounts(accounts []*solana.AccountMeta) error {
	if len(accounts) < 9 {
		return fmt.Errorf("insufficient account")
	}
	i.Accounts = &RegisterTokenAccounts{
		TokenMeta: accounts[0],
		Owner:     accounts[1],
		Token:     accounts[2],
	}

	return nil
}
