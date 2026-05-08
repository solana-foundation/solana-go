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
	"testing"

	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func TestInstructionIDToName(t *testing.T) {
	require.Equal(t, "Write", InstructionIDToName(Instruction_Write))
	require.Equal(t, "Finalize", InstructionIDToName(Instruction_Finalize))
	require.Equal(t, "", InstructionIDToName(99))
}

// TestDecodeInstruction_Write round-trips a built Write through the
// package-level DecodeInstruction, exercising the ag_binary variant
// dispatch path used by the instruction registry.
func TestDecodeInstruction_Write(t *testing.T) {
	account := ag_solanago.MustPublicKeyFromBase58("7QcXLBB23bJ4q5QUXpxLkQBr37g8mNEPSSPyVvU22qUS")
	inst, err := NewWriteInstruction(7, []byte("data"), account).ValidateAndBuild()
	require.NoError(t, err)

	data, err := inst.Data()
	require.NoError(t, err)

	decoded, err := DecodeInstruction(inst.Accounts(), data)
	require.NoError(t, err)

	got, ok := decoded.Impl.(*Write)
	require.True(t, ok, "expected *Write, got %T", decoded.Impl)
	require.Equal(t, uint32(7), *got.Offset)
	require.Equal(t, []byte("data"), got.Bytes)
}

func TestDecodeInstruction_Finalize(t *testing.T) {
	account := ag_solanago.MustPublicKeyFromBase58("7QcXLBB23bJ4q5QUXpxLkQBr37g8mNEPSSPyVvU22qUS")
	inst := NewFinalizeInstruction(account).Build()
	data, err := inst.Data()
	require.NoError(t, err)

	decoded, err := DecodeInstruction(inst.Accounts(), data)
	require.NoError(t, err)
	_, ok := decoded.Impl.(*Finalize)
	require.True(t, ok, "expected *Finalize, got %T", decoded.Impl)
}
