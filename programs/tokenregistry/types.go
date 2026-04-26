// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
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
package tokenregistry

import (
	"fmt"

	bin "github.com/gagliardetto/solana-go/binary"

	"github.com/gagliardetto/solana-go"
)

const TOKEN_META_SIZE = 229

type TokenMeta struct {
	IsInitialized         bool
	Reg                   [3]byte `text:"-"`
	DataType              byte
	MintAddress           *solana.PublicKey
	RegistrationAuthority *solana.PublicKey
	Logo                  Logo
	Name                  Name
	Website               Website
	Symbol                Symbol
}

// MarshalWithEncoder is a hand-written marshaler that bypasses the
// per-field plan-walk overhead of the generic reflect path. The cost
// of the boilerplate is justified for TokenMeta because it is the
// largest reflect-marshaled struct in the repo (9 fields including 4
// nested fixed-size byte arrays and 2 foreign-package pointer types).
//
// We use the explicit Write* methods on the encoder rather than the
// generic encoder.Encode(v any) entry point, because the latter would
// box every [N]byte field into an interface{} on the way in, paying a
// 32-64 byte heap allocation per call. WriteBytes takes a []byte
// directly with no boxing.
func (obj *TokenMeta) MarshalWithEncoder(encoder *bin.Encoder) error {
	if err := encoder.WriteBool(obj.IsInitialized); err != nil {
		return err
	}
	if err := encoder.WriteBytes(obj.Reg[:], false); err != nil {
		return err
	}
	if err := encoder.WriteByte(obj.DataType); err != nil {
		return err
	}
	if obj.MintAddress != nil {
		if err := encoder.WriteBytes(obj.MintAddress[:], false); err != nil {
			return err
		}
	} else {
		var zero solana.PublicKey
		if err := encoder.WriteBytes(zero[:], false); err != nil {
			return err
		}
	}
	if obj.RegistrationAuthority != nil {
		if err := encoder.WriteBytes(obj.RegistrationAuthority[:], false); err != nil {
			return err
		}
	} else {
		var zero solana.PublicKey
		if err := encoder.WriteBytes(zero[:], false); err != nil {
			return err
		}
	}
	if err := encoder.WriteBytes(obj.Logo[:], false); err != nil {
		return err
	}
	if err := encoder.WriteBytes(obj.Name[:], false); err != nil {
		return err
	}
	if err := encoder.WriteBytes(obj.Website[:], false); err != nil {
		return err
	}
	if err := encoder.WriteBytes(obj.Symbol[:], false); err != nil {
		return err
	}
	return nil
}

// UnmarshalWithDecoder reads each field directly into its destination
// with no per-field reflect or option-construction cost. The two
// *solana.PublicKey allocations are unavoidable because the field
// types are pointers.
func (obj *TokenMeta) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	var err error
	if obj.IsInitialized, err = decoder.ReadBool(); err != nil {
		return err
	}
	regBytes, err := decoder.ReadNBytes(3)
	if err != nil {
		return err
	}
	copy(obj.Reg[:], regBytes)
	if obj.DataType, err = decoder.ReadByte(); err != nil {
		return err
	}
	mint := new(solana.PublicKey)
	mintBytes, err := decoder.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(mint[:], mintBytes)
	obj.MintAddress = mint
	auth := new(solana.PublicKey)
	authBytes, err := decoder.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(auth[:], authBytes)
	obj.RegistrationAuthority = auth
	logoBytes, err := decoder.ReadNBytes(64)
	if err != nil {
		return err
	}
	copy(obj.Logo[:], logoBytes)
	nameBytes, err := decoder.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(obj.Name[:], nameBytes)
	siteBytes, err := decoder.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(obj.Website[:], siteBytes)
	symBytes, err := decoder.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(obj.Symbol[:], symBytes)
	return nil
}

func DecodeTokenMeta(in []byte) (*TokenMeta, error) {
	var t *TokenMeta
	decoder := bin.NewBinDecoder(in)
	err := decoder.Decode(&t)
	if err != nil {
		return nil, fmt.Errorf("unpack: %w", err)
	}
	return t, nil
}

type Logo [64]byte

func LogoFromString(logo string) (Logo, error) {
	data := []byte(logo)
	if len(data) > 64 {
		return Logo{}, fmt.Errorf("logo data to long expected 64 got %d", len(data))
	}
	l := Logo{}
	copy(l[:], data)
	return l, nil
}
func (l Logo) String() string {
	return AsciiString(l[:])
}

type Name [32]byte

func NameFromString(name string) (Name, error) {
	data := []byte(name)
	if len(data) > 32 {
		return Name{}, fmt.Errorf("name data to long expected 32 got %d", len(data))
	}
	n := Name{}
	copy(n[:], data)
	return n, nil
}

func (n Name) String() string {
	return AsciiString(n[:])
}

type Symbol [32]byte

func SymbolFromString(symbol string) (Symbol, error) {
	data := []byte(symbol)
	if len(data) > 32 {
		return Symbol{}, fmt.Errorf("symbol data to long expected 12 got %d", len(data))
	}
	s := Symbol{}
	copy(s[:], data)
	return s, nil
}

func (s Symbol) String() string {
	return AsciiString(s[:])
}

type Website [32]byte

func WebsiteFromString(symbol string) (Website, error) {
	data := []byte(symbol)
	if len(data) > 32 {
		return Website{}, fmt.Errorf("website data to long expected 32 got %d", len(data))
	}
	s := Website{}
	copy(s[:], data)
	return s, nil
}

func (s Website) String() string {
	return AsciiString(s[:])
}

func AsciiString(data []byte) string {
	var trimmed []byte
	for _, b := range data {
		if b > 0 {
			trimmed = append(trimmed, b)
		}
	}
	return string(trimmed)
}
