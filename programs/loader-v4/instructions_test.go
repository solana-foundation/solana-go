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

var (
	testProgram    = ag_solanago.MustPublicKeyFromBase58("4Nd1mZjsqhMzGPhFHZ4mJ4nZxkFwVoQCpT8xcVFt5Kfr")
	testAuthority  = ag_solanago.MustPublicKeyFromBase58("5fjG93skfVqRnF5M3n8h3fLvj1Fq1VmHnT1tqYfQPVpF")
	testSource     = ag_solanago.MustPublicKeyFromBase58("7QcXLBB23bJ4q5QUXpxLkQBr37g8mNEPSSPyVvU22qUS")
	testRecipient  = ag_solanago.MustPublicKeyFromBase58("6WQZ5tL94vnB2G89fjCH6WVDSSz8Q6PF8U9X3vBt8aJH")
	testNewAuth    = ag_solanago.MustPublicKeyFromBase58("8U1JpQ4Z6GMg5Xdz5r78KjWqytBxEQNkTP2Xb6RRjXgH")
	testNextVer    = ag_solanago.MustPublicKeyFromBase58("3UVYmECPPMZSCqWKfENfuoTv51fTDTWicX9xmBD2euKe")
	testPayer      = ag_solanago.MustPublicKeyFromBase58("9VBbBpcPLmJcAJfrj45dWqPuKePYzRnT24yUhGmBe7uB")
	testBufferAcct = ag_solanago.MustPublicKeyFromBase58("2xnCAsCZ9kC2L9QH4Zh8nZPyJkNuHfKqG8f9oNTtK3Vr")
)

// TestWireCompat_Bincode asserts each instruction's Data() matches the
// default bincode encoding: u32 LE discriminant followed by the fixed-width
// payload. Although LoaderV4Instruction is marked `#[repr(u8)]` in Rust,
// bincode ignores that and serializes the discriminant as u32 LE.
func TestWireCompat_Bincode(t *testing.T) {
	tests := []struct {
		name string
		inst *Instruction
		want []byte
	}{
		{
			name: "Write",
			inst: NewWriteInstruction(testProgram, testAuthority, 42, []byte{1, 2, 3}).Build(),
			want: []byte{
				0x00, 0x00, 0x00, 0x00, // disc
				0x2a, 0x00, 0x00, 0x00, // offset = 42
				0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // len = 3
				0x01, 0x02, 0x03,
			},
		},
		{
			name: "Copy",
			inst: NewCopyInstruction(testProgram, testAuthority, testSource, 10, 20, 30).Build(),
			want: []byte{
				0x01, 0x00, 0x00, 0x00,
				0x0a, 0x00, 0x00, 0x00, // dst_offset = 10
				0x14, 0x00, 0x00, 0x00, // src_offset = 20
				0x1e, 0x00, 0x00, 0x00, // length = 30
			},
		},
		{
			name: "SetProgramLength",
			inst: NewSetProgramLengthInstruction(testProgram, testAuthority, testRecipient, 100_000).Build(),
			want: []byte{
				0x02, 0x00, 0x00, 0x00,
				0xa0, 0x86, 0x01, 0x00, // 100_000 u32 LE
			},
		},
		{
			name: "Deploy",
			inst: NewDeployInstruction(testProgram, testAuthority).Build(),
			want: []byte{0x03, 0x00, 0x00, 0x00},
		},
		{
			name: "Retract",
			inst: NewRetractInstruction(testProgram, testAuthority).Build(),
			want: []byte{0x04, 0x00, 0x00, 0x00},
		},
		{
			name: "TransferAuthority",
			inst: NewTransferAuthorityInstruction(testProgram, testAuthority, testNewAuth).Build(),
			want: []byte{0x05, 0x00, 0x00, 0x00},
		},
		{
			name: "Finalize",
			inst: NewFinalizeInstruction(testProgram, testAuthority, testNextVer).Build(),
			want: []byte{0x06, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.inst.Data()
			require.NoError(t, err)
			require.Equal(t, tc.want, got)

			decoded, err := DecodeInstruction(tc.inst.Accounts(), got)
			require.NoError(t, err)
			require.NotNil(t, decoded.Impl)
		})
	}
}

func TestIsXxxInstructionPredicates(t *testing.T) {
	// Each `is_*_instruction` helper checks only the first byte, relying on
	// bincode's u32 LE discriminant putting the low byte at index 0.
	write := buildData(t, NewWriteInstruction(testProgram, testAuthority, 0, nil).Build())
	require.True(t, IsWriteInstruction(write))
	require.False(t, IsCopyInstruction(write))

	cp := buildData(t, NewCopyInstruction(testProgram, testAuthority, testSource, 0, 0, 0).Build())
	require.True(t, IsCopyInstruction(cp))

	spl := buildData(t, NewSetProgramLengthInstruction(testProgram, testAuthority, testRecipient, 0).Build())
	require.True(t, IsSetProgramLengthInstruction(spl))

	dep := buildData(t, NewDeployInstruction(testProgram, testAuthority).Build())
	require.True(t, IsDeployInstruction(dep))

	ret := buildData(t, NewRetractInstruction(testProgram, testAuthority).Build())
	require.True(t, IsRetractInstruction(ret))

	ta := buildData(t, NewTransferAuthorityInstruction(testProgram, testAuthority, testNewAuth).Build())
	require.True(t, IsTransferAuthorityInstruction(ta))

	fin := buildData(t, NewFinalizeInstruction(testProgram, testAuthority, testNextVer).Build())
	require.True(t, IsFinalizeInstruction(fin))

	// Empty data never matches anything.
	require.False(t, IsWriteInstruction(nil))
	require.False(t, IsFinalizeInstruction(nil))
}

func buildData(t *testing.T, inst *Instruction) []byte {
	t.Helper()
	data, err := inst.Data()
	require.NoError(t, err)
	return data
}

func TestInstructionIDToName(t *testing.T) {
	require.Equal(t, "Write", InstructionIDToName(Instruction_Write))
	require.Equal(t, "Copy", InstructionIDToName(Instruction_Copy))
	require.Equal(t, "SetProgramLength", InstructionIDToName(Instruction_SetProgramLength))
	require.Equal(t, "Deploy", InstructionIDToName(Instruction_Deploy))
	require.Equal(t, "Retract", InstructionIDToName(Instruction_Retract))
	require.Equal(t, "TransferAuthority", InstructionIDToName(Instruction_TransferAuthority))
	require.Equal(t, "Finalize", InstructionIDToName(Instruction_Finalize))
	require.Equal(t, "", InstructionIDToName(99))
}
