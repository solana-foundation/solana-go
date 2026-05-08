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
	"errors"
	"testing"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// Placeholder accounts for builder-driven tests. The upstream account-meta
// predicates are validated separately per instruction; these golden tests
// focus on wire bytes.
var (
	testBuffer    = ag_solanago.MustPublicKeyFromBase58("7QcXLBB23bJ4q5QUXpxLkQBr37g8mNEPSSPyVvU22qUS")
	testAuthority = ag_solanago.MustPublicKeyFromBase58("5fjG93skfVqRnF5M3n8h3fLvj1Fq1VmHnT1tqYfQPVpF")
	testPayer     = ag_solanago.MustPublicKeyFromBase58("3UVYmECPPMZSCqWKfENfuoTv51fTDTWicX9xmBD2euKe")
	testProgram   = ag_solanago.MustPublicKeyFromBase58("4Nd1mZjsqhMzGPhFHZ4mJ4nZxkFwVoQCpT8xcVFt5Kfr")
	testSpill     = ag_solanago.MustPublicKeyFromBase58("6WQZ5tL94vnB2G89fjCH6WVDSSz8Q6PF8U9X3vBt8aJH")
	testNewAuth   = ag_solanago.MustPublicKeyFromBase58("8U1JpQ4Z6GMg5Xdz5r78KjWqytBxEQNkTP2Xb6RRjXgH")
)

// buildData is a helper that invokes an instruction's full Data() path,
// including the u32 LE discriminant prefix.
func buildData(t *testing.T, inst *Instruction) []byte {
	t.Helper()
	data, err := inst.Data()
	require.NoError(t, err)
	return data
}

// TestWireCompat_Bincode reproduces the upstream `wire_compat_bincode_vs_wincode`
// test cases (13 variants) by asserting the exact byte layout emitted by each
// instruction's Data() matches the default bincode encoding:
//
//	u32 LE discriminant, u64 LE Vec<u8> length, 1-byte bool, 32-byte Pubkey.
func TestWireCompat_Bincode(t *testing.T) {
	tests := []struct {
		name string
		inst *Instruction
		want []byte
	}{
		{
			name: "InitializeBuffer",
			inst: NewInitializeBufferInstruction(testBuffer, testAuthority).Build(),
			want: []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "Write(offset=42, bytes=[1..5])",
			inst: NewWriteInstruction(testBuffer, testAuthority, 42, []byte{1, 2, 3, 4, 5}).Build(),
			want: []byte{
				0x01, 0x00, 0x00, 0x00, // disc
				0x2a, 0x00, 0x00, 0x00, // offset = 42
				0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // len = 5
				0x01, 0x02, 0x03, 0x04, 0x05,
			},
		},
		{
			name: "Write(offset=0, bytes=[])",
			inst: NewWriteInstruction(testBuffer, testAuthority, 0, []byte{}).Build(),
			want: []byte{
				0x01, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "DeployWithMaxDataLen(1_000_000, close_buffer=true)",
			inst: NewDeployWithMaxDataLenInstruction(
				testPayer, MustGetProgramDataAddress(testProgram), testProgram,
				testBuffer, testAuthority, 1_000_000, true,
			).Build(),
			want: []byte{
				0x02, 0x00, 0x00, 0x00, // disc
				0x40, 0x42, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // 1_000_000 u64 LE
				0x01, // close_buffer
			},
		},
		{
			name: "DeployWithMaxDataLen(0, close_buffer=false)",
			inst: NewDeployWithMaxDataLenInstruction(
				testPayer, MustGetProgramDataAddress(testProgram), testProgram,
				testBuffer, testAuthority, 0, false,
			).Build(),
			want: []byte{
				0x02, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00,
			},
		},
		{
			name: "Upgrade(close_buffer=true)",
			inst: NewUpgradeInstruction(testProgram, testBuffer, testAuthority, testSpill, true).Build(),
			want: []byte{0x03, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name: "Upgrade(close_buffer=false)",
			inst: NewUpgradeInstruction(testProgram, testBuffer, testAuthority, testSpill, false).Build(),
			want: []byte{0x03, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "SetAuthority",
			inst: NewSetBufferAuthorityInstruction(testBuffer, testAuthority, testNewAuth).Build(),
			want: []byte{0x04, 0x00, 0x00, 0x00},
		},
		{
			name: "Close(tombstone=false)",
			inst: NewCloseInstruction(testBuffer, testPayer, testAuthority, false).Build(),
			want: []byte{0x05, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "Close(tombstone=true)",
			inst: NewCloseInstruction(testBuffer, testPayer, testAuthority, true).Build(),
			want: []byte{0x05, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name: "ExtendProgram(10_240)",
			inst: NewExtendProgramInstruction(testProgram, nil, 10_240).Build(),
			want: []byte{
				0x06, 0x00, 0x00, 0x00,
				0x00, 0x28, 0x00, 0x00, // 10240 u32 LE
			},
		},
		{
			name: "ExtendProgram(0)",
			inst: NewExtendProgramInstruction(testProgram, nil, 0).Build(),
			want: []byte{
				0x06, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "SetAuthorityChecked",
			inst: NewSetBufferAuthorityCheckedInstruction(testBuffer, testAuthority, testNewAuth).Build(),
			want: []byte{0x07, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildData(t, tc.inst)
			require.Equal(t, tc.want, got)

			// Round-trip: the produced bytes must decode back to the same variant.
			decoded, err := DecodeInstruction(tc.inst.Accounts(), got)
			require.NoError(t, err)
			require.NotNil(t, decoded.Impl)
		})
	}
}

// TestLegacyDeployDecodesCloseBufferAsTrue verifies the SIMD-0430 decode
// tolerance: a payload lacking the trailing close_buffer byte must decode as
// `close_buffer = true`.
func TestLegacyDeployDecodesCloseBufferAsTrue(t *testing.T) {
	legacy := []byte{
		0x02, 0x00, 0x00, 0x00, // disc = 2
		0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // max_data_len = 42
	}
	decoded := new(Instruction)
	require.NoError(t, ag_binary.NewBinDecoder(legacy).Decode(decoded))

	impl, ok := decoded.Impl.(*DeployWithMaxDataLen)
	require.True(t, ok, "expected *DeployWithMaxDataLen")
	require.Equal(t, uint64(42), *impl.MaxDataLen)
	require.True(t, impl.CloseBuffer, "close_buffer should default to true on exhausted read")
}

func TestLegacyUpgradeDecodesCloseBufferAsTrue(t *testing.T) {
	legacy := []byte{0x03, 0x00, 0x00, 0x00} // bare discriminant
	decoded := new(Instruction)
	require.NoError(t, ag_binary.NewBinDecoder(legacy).Decode(decoded))

	impl, ok := decoded.Impl.(*Upgrade)
	require.True(t, ok)
	require.True(t, impl.CloseBuffer)
}

// TestLegacyCloseDecodesTombstoneAsFalse verifies the SIMD-0432 default.
func TestLegacyCloseDecodesTombstoneAsFalse(t *testing.T) {
	legacy := []byte{0x05, 0x00, 0x00, 0x00}
	decoded := new(Instruction)
	require.NoError(t, ag_binary.NewBinDecoder(legacy).Decode(decoded))

	impl, ok := decoded.Impl.(*Close)
	require.True(t, ok)
	require.False(t, impl.Tombstone)
}

// TestInvalidOptionalTrailingBoolByteErrors mirrors upstream's error-path
// test: any byte other than 0 or 1 in an OptionalTrailingBool slot must
// surface as a decode error. The top-level Instruction path is exercised
// here; errors.Is against ErrInvalidBoolEncoding is covered separately by
// TestInvalidOptionalTrailingBool_Sentinel, which uses the direct-variant
// decode path where the error chain is preserved (ag_binary's variant
// dispatcher wraps with %s, which strips %w context).
func TestInvalidOptionalTrailingBoolByteErrors(t *testing.T) {
	cases := [][]byte{
		{0x02, 0x00, 0x00, 0x00, 0x2a, 0, 0, 0, 0, 0, 0, 0, 0x02}, // DeployWithMaxDataLen
		{0x03, 0x00, 0x00, 0x00, 0x02},                            // Upgrade
		{0x05, 0x00, 0x00, 0x00, 0x02},                            // Close
	}
	for i, data := range cases {
		decoded := new(Instruction)
		err := ag_binary.NewBinDecoder(data).Decode(decoded)
		require.Error(t, err, "case %d should fail", i)
	}
}

// TestInvalidOptionalTrailingBool_Sentinel verifies that direct-variant
// decoders surface ErrInvalidBoolEncoding so callers can match on it via
// errors.Is. This is the path taken when a caller already knows the
// concrete variant (e.g. after inspecting the discriminant themselves).
func TestInvalidOptionalTrailingBool_Sentinel(t *testing.T) {
	// Upgrade's trailing bool byte = 0x02 — invalid.
	dec := ag_binary.NewBinDecoder([]byte{0x02})
	got := new(Upgrade)
	err := got.UnmarshalWithDecoder(dec)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidBoolEncoding),
		"expected ErrInvalidBoolEncoding, got %v", err)
}

func TestInstructionIDToName(t *testing.T) {
	require.Equal(t, "InitializeBuffer", InstructionIDToName(Instruction_InitializeBuffer))
	require.Equal(t, "Write", InstructionIDToName(Instruction_Write))
	require.Equal(t, "DeployWithMaxDataLen", InstructionIDToName(Instruction_DeployWithMaxDataLen))
	require.Equal(t, "Upgrade", InstructionIDToName(Instruction_Upgrade))
	require.Equal(t, "SetAuthority", InstructionIDToName(Instruction_SetAuthority))
	require.Equal(t, "Close", InstructionIDToName(Instruction_Close))
	require.Equal(t, "ExtendProgram", InstructionIDToName(Instruction_ExtendProgram))
	require.Equal(t, "SetAuthorityChecked", InstructionIDToName(Instruction_SetAuthorityChecked))
	require.Equal(t, "", InstructionIDToName(99))
}

func TestMinimumExtendProgramBytes(t *testing.T) {
	require.Equal(t, uint32(10_240), MINIMUM_EXTEND_PROGRAM_BYTES)
}
