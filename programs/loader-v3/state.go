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

package loaderv3

import (
	"encoding/binary"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
)

// State discriminants for UpgradeableLoaderState (u32 LE).
const (
	State_Uninitialized uint32 = iota
	State_Buffer
	State_Program
	State_ProgramData
)

// Size constants for the serialized UpgradeableLoaderState header, matching
// upstream's `UpgradeableLoaderState::size_of_*` helpers.
const (
	SizeOfUninitialized       = 4  // u32 discriminant
	SizeOfBufferMetadata      = 37 // disc(4) + Option<Pubkey> tag(1) + 32
	SizeOfProgram             = 36 // disc(4) + Pubkey(32)
	SizeOfProgramDataMetadata = 45 // disc(4) + slot(8) + Option<Pubkey> tag(1) + 32
)

// SizeOfBuffer returns the account size required to hold a Buffer of the
// given program length.
func SizeOfBuffer(programLen int) int { return SizeOfBufferMetadata + programLen }

// SizeOfProgramData returns the account size required to hold ProgramData of
// the given program length.
func SizeOfProgramData(programLen int) int { return SizeOfProgramDataMetadata + programLen }

// UpgradeableLoaderState is the tagged-union account payload for buffer,
// program, and programdata accounts owned by BPFLoaderUpgradeab1e.
// Exactly one of the Buffer/Program/ProgramData pointer fields is populated
// after a successful decode. Uninitialized is represented by all fields nil.
type UpgradeableLoaderState struct {
	Buffer      *StateBuffer
	Program     *StateProgram
	ProgramData *StateProgramData
}

type StateBuffer struct {
	// Authority address for the buffer; nil means None.
	AuthorityAddress *ag_solanago.PublicKey
}

type StateProgram struct {
	ProgramDataAddress ag_solanago.PublicKey
}

type StateProgramData struct {
	Slot uint64
	// nil means None (final/immutable ProgramData).
	UpgradeAuthorityAddress *ag_solanago.PublicKey
}

// IsUninitialized reports whether the state is the Uninitialized variant.
func (s *UpgradeableLoaderState) IsUninitialized() bool {
	return s.Buffer == nil && s.Program == nil && s.ProgramData == nil
}

// MarshalWithEncoder serializes the state with upstream's bincode layout:
// u32 LE discriminant followed by the variant payload.
func (s UpgradeableLoaderState) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	set := 0
	if s.Buffer != nil {
		set++
	}
	if s.Program != nil {
		set++
	}
	if s.ProgramData != nil {
		set++
	}
	if set > 1 {
		return fmt.Errorf("UpgradeableLoaderState has multiple variants set")
	}

	switch {
	case s.Buffer != nil:
		if err := encoder.WriteUint32(State_Buffer, binary.LittleEndian); err != nil {
			return err
		}
		return writeOptionPubkey(encoder, s.Buffer.AuthorityAddress)
	case s.Program != nil:
		if err := encoder.WriteUint32(State_Program, binary.LittleEndian); err != nil {
			return err
		}
		return encoder.Encode(s.Program.ProgramDataAddress)
	case s.ProgramData != nil:
		if err := encoder.WriteUint32(State_ProgramData, binary.LittleEndian); err != nil {
			return err
		}
		if err := encoder.WriteUint64(s.ProgramData.Slot, binary.LittleEndian); err != nil {
			return err
		}
		return writeOptionPubkey(encoder, s.ProgramData.UpgradeAuthorityAddress)
	default:
		return encoder.WriteUint32(State_Uninitialized, binary.LittleEndian)
	}
}

func (s *UpgradeableLoaderState) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	disc, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	*s = UpgradeableLoaderState{}
	switch disc {
	case State_Uninitialized:
		return nil
	case State_Buffer:
		auth, err := readOptionPubkey(decoder)
		if err != nil {
			return err
		}
		s.Buffer = &StateBuffer{AuthorityAddress: auth}
		return nil
	case State_Program:
		var pda ag_solanago.PublicKey
		if err := decoder.Decode(&pda); err != nil {
			return err
		}
		s.Program = &StateProgram{ProgramDataAddress: pda}
		return nil
	case State_ProgramData:
		slot, err := decoder.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
		auth, err := readOptionPubkey(decoder)
		if err != nil {
			return err
		}
		s.ProgramData = &StateProgramData{Slot: slot, UpgradeAuthorityAddress: auth}
		return nil
	default:
		return fmt.Errorf("unknown UpgradeableLoaderState discriminant: %d", disc)
	}
}

func writeOptionPubkey(encoder *ag_binary.Encoder, pk *ag_solanago.PublicKey) error {
	if pk == nil {
		return encoder.WriteUint8(0)
	}
	if err := encoder.WriteUint8(1); err != nil {
		return err
	}
	return encoder.Encode(*pk)
}

func readOptionPubkey(decoder *ag_binary.Decoder) (*ag_solanago.PublicKey, error) {
	tag, err := decoder.ReadUint8()
	if err != nil {
		return nil, err
	}
	switch tag {
	case 0:
		return nil, nil
	case 1:
		var pk ag_solanago.PublicKey
		if err := decoder.Decode(&pk); err != nil {
			return nil, err
		}
		return &pk, nil
	default:
		return nil, fmt.Errorf("invalid Option tag: %d", tag)
	}
}
