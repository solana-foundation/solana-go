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

package loaderv4

import (
	"testing"

	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// TestLayout mirrors upstream's `test_layout`:
//
//	slot @ 0x00, authority @ 0x08, status @ 0x28, program_data_offset == 0x30.
func TestLayout(t *testing.T) {
	require.Equal(t, 48, ProgramDataOffset)

	auth := ag_solanago.MustPublicKeyFromBase58("5fjG93skfVqRnF5M3n8h3fLvj1Fq1VmHnT1tqYfQPVpF")
	state := &LoaderV4State{
		Slot:                          0x11_22_33_44_55_66_77_88,
		AuthorityAddressOrNextVersion: auth,
		Status:                        StatusDeployed,
	}
	packed := state.Pack()
	require.Len(t, packed, 48)

	// slot @ 0x00..0x08
	require.Equal(t, byte(0x88), packed[0])
	require.Equal(t, byte(0x11), packed[7])

	// authority @ 0x08..0x28
	require.Equal(t, auth[:], packed[8:40])

	// status @ 0x28..0x30 (u64 LE, value 1 = Deployed)
	require.Equal(t, byte(0x01), packed[40])
	require.Equal(t, byte(0x00), packed[47])

	// Round-trip.
	got, err := UnpackLoaderV4State(packed)
	require.NoError(t, err)
	require.Equal(t, state.Slot, got.Slot)
	require.Equal(t, state.AuthorityAddressOrNextVersion, got.AuthorityAddressOrNextVersion)
	require.Equal(t, state.Status, got.Status)
}

func TestUnpack_ShortData(t *testing.T) {
	_, err := UnpackLoaderV4State(make([]byte, 47))
	require.Error(t, err)
}

func TestUnpack_WithProgramBytes(t *testing.T) {
	state := &LoaderV4State{Status: StatusFinalized}
	payload := append(state.Pack(), []byte{0xAB, 0xCD, 0xEF}...)
	got, err := UnpackLoaderV4State(payload)
	require.NoError(t, err)
	require.Equal(t, StatusFinalized, got.Status)
}

func TestStatusString(t *testing.T) {
	require.Equal(t, "Retracted", StatusRetracted.String())
	require.Equal(t, "Deployed", StatusDeployed.String())
	require.Equal(t, "Finalized", StatusFinalized.String())
	require.Equal(t, "Unknown(42)", LoaderV4Status(42).String())
}
