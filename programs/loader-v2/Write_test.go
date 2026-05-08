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

package loaderv2

import (
	"bytes"
	"testing"

	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// TestWrite_Instruction_Bincode is a golden-byte test reproducing the wire
// format produced by the upstream Rust helper
//
//	loader_v2_interface::write(&account, &program_id, 42, vec![1,2,3,4,5])
//
// via `Instruction::new_with_bincode`. Default bincode encodes:
//
//	[disc: u32 LE][offset: u32 LE][len(bytes): u64 LE][bytes...]
func TestWrite_Instruction_Bincode(t *testing.T) {
	account := ag_solanago.MustPublicKeyFromBase58("7QcXLBB23bJ4q5QUXpxLkQBr37g8mNEPSSPyVvU22qUS")

	inst, err := NewWriteInstruction(42, []byte{1, 2, 3, 4, 5}, account).ValidateAndBuild()
	require.NoError(t, err)

	data, err := inst.Data()
	require.NoError(t, err)

	expected := []byte{
		// discriminant = 0 (u32 LE)
		0x00, 0x00, 0x00, 0x00,
		// offset = 42 (u32 LE)
		0x2a, 0x00, 0x00, 0x00,
		// len(bytes) = 5 (u64 LE)
		0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// bytes
		0x01, 0x02, 0x03, 0x04, 0x05,
	}
	require.Equal(t, expected, data)
}

func TestWrite_Instruction_Bincode_EmptyBytes(t *testing.T) {
	account := ag_solanago.MustPublicKeyFromBase58("7QcXLBB23bJ4q5QUXpxLkQBr37g8mNEPSSPyVvU22qUS")

	inst, err := NewWriteInstruction(0, []byte{}, account).ValidateAndBuild()
	require.NoError(t, err)

	data, err := inst.Data()
	require.NoError(t, err)

	expected := []byte{
		0x00, 0x00, 0x00, 0x00, // disc = 0
		0x00, 0x00, 0x00, 0x00, // offset = 0
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // len = 0
	}
	require.Equal(t, expected, data)
}

func TestWrite_RoundTrip(t *testing.T) {
	account := ag_solanago.MustPublicKeyFromBase58("7QcXLBB23bJ4q5QUXpxLkQBr37g8mNEPSSPyVvU22qUS")

	original := NewWriteInstruction(1234, []byte("hello loader"), account)
	buf := new(bytes.Buffer)
	require.NoError(t, encodeT(*original, buf))

	got := new(Write)
	require.NoError(t, decodeT(got, buf.Bytes()))
	require.Equal(t, *original.Offset, *got.Offset)
	require.Equal(t, original.Bytes, got.Bytes)
}

// TestWrite_Accounts mirrors the Rust account-meta assertions: the target
// account is passed as [WRITE, SIGNER] and no other accounts are set.
func TestWrite_Accounts(t *testing.T) {
	account := ag_solanago.MustPublicKeyFromBase58("7QcXLBB23bJ4q5QUXpxLkQBr37g8mNEPSSPyVvU22qUS")
	inst := NewWriteInstruction(0, nil, account).Build()

	require.Equal(t, ProgramID, inst.ProgramID())
	accounts := inst.Accounts()
	require.Len(t, accounts, 1)
	require.Equal(t, account, accounts[0].PublicKey)
	require.True(t, accounts[0].IsWritable)
	require.True(t, accounts[0].IsSigner)
}

func TestWrite_Validate_MissingOffset(t *testing.T) {
	account := ag_solanago.MustPublicKeyFromBase58("7QcXLBB23bJ4q5QUXpxLkQBr37g8mNEPSSPyVvU22qUS")
	b := NewWriteInstructionBuilder().SetAccount(account).SetBytes([]byte{1})
	_, err := b.ValidateAndBuild()
	require.Error(t, err)
}
