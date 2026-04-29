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
	"bytes"
	"testing"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// Mirrors upstream's `test_state_size_of_*` and `wire_compat_bincode_vs_wincode`
// (state) tests: asserts both the declared size constants and the actual
// serialized byte length of each variant.
func TestStateSizes(t *testing.T) {
	require.Equal(t, 4, SizeOfUninitialized)
	require.Equal(t, 37, SizeOfBufferMetadata)
	require.Equal(t, 36, SizeOfProgram)
	require.Equal(t, 45, SizeOfProgramDataMetadata)

	require.Equal(t, 37+128, SizeOfBuffer(128))
	require.Equal(t, 45+128, SizeOfProgramData(128))
}

func TestState_Uninitialized(t *testing.T) {
	s := UpgradeableLoaderState{}
	buf := new(bytes.Buffer)
	require.NoError(t, ag_binary.NewBinEncoder(buf).Encode(s))
	require.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, buf.Bytes())

	got := new(UpgradeableLoaderState)
	require.NoError(t, ag_binary.NewBinDecoder(buf.Bytes()).Decode(got))
	require.True(t, got.IsUninitialized())
}

func TestState_Buffer_Some(t *testing.T) {
	var defaultPubkey ag_solanago.PublicKey // zero value
	s := UpgradeableLoaderState{
		Buffer: &StateBuffer{AuthorityAddress: &defaultPubkey},
	}
	buf := new(bytes.Buffer)
	require.NoError(t, ag_binary.NewBinEncoder(buf).Encode(s))
	require.Len(t, buf.Bytes(), SizeOfBufferMetadata)

	// Disc(4) + Option tag(1) + Pubkey(32) = 37
	require.Equal(t, byte(0x01), buf.Bytes()[0])
	require.Equal(t, byte(0x01), buf.Bytes()[4]) // Option = Some

	got := new(UpgradeableLoaderState)
	require.NoError(t, ag_binary.NewBinDecoder(buf.Bytes()).Decode(got))
	require.NotNil(t, got.Buffer)
	require.NotNil(t, got.Buffer.AuthorityAddress)
	require.Equal(t, defaultPubkey, *got.Buffer.AuthorityAddress)
}

func TestState_Buffer_None(t *testing.T) {
	s := UpgradeableLoaderState{Buffer: &StateBuffer{AuthorityAddress: nil}}
	buf := new(bytes.Buffer)
	require.NoError(t, ag_binary.NewBinEncoder(buf).Encode(s))
	// Disc(4) + Option tag(1) = 5
	require.Equal(t, []byte{0x01, 0x00, 0x00, 0x00, 0x00}, buf.Bytes())

	got := new(UpgradeableLoaderState)
	require.NoError(t, ag_binary.NewBinDecoder(buf.Bytes()).Decode(got))
	require.NotNil(t, got.Buffer)
	require.Nil(t, got.Buffer.AuthorityAddress)
}

func TestState_Program(t *testing.T) {
	var defaultPubkey ag_solanago.PublicKey
	s := UpgradeableLoaderState{Program: &StateProgram{ProgramDataAddress: defaultPubkey}}
	buf := new(bytes.Buffer)
	require.NoError(t, ag_binary.NewBinEncoder(buf).Encode(s))
	require.Len(t, buf.Bytes(), SizeOfProgram)
	require.Equal(t, byte(0x02), buf.Bytes()[0])

	got := new(UpgradeableLoaderState)
	require.NoError(t, ag_binary.NewBinDecoder(buf.Bytes()).Decode(got))
	require.NotNil(t, got.Program)
	require.Equal(t, defaultPubkey, got.Program.ProgramDataAddress)
}

func TestState_ProgramData_Some(t *testing.T) {
	var defaultPubkey ag_solanago.PublicKey
	s := UpgradeableLoaderState{
		ProgramData: &StateProgramData{
			Slot:                    123_456_789,
			UpgradeAuthorityAddress: &defaultPubkey,
		},
	}
	buf := new(bytes.Buffer)
	require.NoError(t, ag_binary.NewBinEncoder(buf).Encode(s))
	require.Len(t, buf.Bytes(), SizeOfProgramDataMetadata)
	require.Equal(t, byte(0x03), buf.Bytes()[0])

	got := new(UpgradeableLoaderState)
	require.NoError(t, ag_binary.NewBinDecoder(buf.Bytes()).Decode(got))
	require.NotNil(t, got.ProgramData)
	require.Equal(t, uint64(123_456_789), got.ProgramData.Slot)
	require.NotNil(t, got.ProgramData.UpgradeAuthorityAddress)
}

func TestState_ProgramData_None(t *testing.T) {
	s := UpgradeableLoaderState{
		ProgramData: &StateProgramData{Slot: 0, UpgradeAuthorityAddress: nil},
	}
	buf := new(bytes.Buffer)
	require.NoError(t, ag_binary.NewBinEncoder(buf).Encode(s))
	expected := []byte{
		0x03, 0x00, 0x00, 0x00, // disc
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // slot
		0x00, // Option = None
	}
	require.Equal(t, expected, buf.Bytes())

	got := new(UpgradeableLoaderState)
	require.NoError(t, ag_binary.NewBinDecoder(buf.Bytes()).Decode(got))
	require.NotNil(t, got.ProgramData)
	require.Nil(t, got.ProgramData.UpgradeAuthorityAddress)
}
