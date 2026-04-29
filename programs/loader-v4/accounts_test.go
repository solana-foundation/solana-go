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

	"github.com/stretchr/testify/require"
)

// The following tests mirror upstream's per-builder unit tests
// (`test_*_instruction`): they verify the program ID, account count, and
// the (IsWritable, IsSigner) flags on every AccountMeta.

func TestAccounts_Write(t *testing.T) {
	inst := NewWriteInstruction(testProgram, testAuthority, 0, nil).Build()
	require.Equal(t, ProgramID, inst.ProgramID())
	a := inst.Accounts()
	require.Len(t, a, 2)

	require.Equal(t, testProgram, a[0].PublicKey)
	require.True(t, a[0].IsWritable)
	require.False(t, a[0].IsSigner)

	require.Equal(t, testAuthority, a[1].PublicKey)
	require.False(t, a[1].IsWritable)
	require.True(t, a[1].IsSigner)
}

func TestAccounts_Copy(t *testing.T) {
	inst := NewCopyInstruction(testProgram, testAuthority, testSource, 0, 0, 0).Build()
	a := inst.Accounts()
	require.Len(t, a, 3)
	require.True(t, a[0].IsWritable && !a[0].IsSigner)
	require.True(t, !a[1].IsWritable && a[1].IsSigner)
	require.True(t, !a[2].IsWritable && !a[2].IsSigner)
	require.Equal(t, testSource, a[2].PublicKey)
}

func TestAccounts_SetProgramLength(t *testing.T) {
	inst := NewSetProgramLengthInstruction(testProgram, testAuthority, testRecipient, 100).Build()
	a := inst.Accounts()
	require.Len(t, a, 3)
	require.True(t, a[0].IsWritable && !a[0].IsSigner)
	require.True(t, !a[1].IsWritable && a[1].IsSigner)
	require.True(t, a[2].IsWritable && !a[2].IsSigner)
	require.Equal(t, testRecipient, a[2].PublicKey)
}

func TestAccounts_Deploy(t *testing.T) {
	inst := NewDeployInstruction(testProgram, testAuthority).Build()
	a := inst.Accounts()
	require.Len(t, a, 2)
	require.True(t, a[0].IsWritable && !a[0].IsSigner)
	require.True(t, !a[1].IsWritable && a[1].IsSigner)
}

func TestAccounts_DeployFromSource(t *testing.T) {
	inst := NewDeployFromSourceInstruction(testProgram, testAuthority, testSource).Build()
	a := inst.Accounts()
	require.Len(t, a, 3)
	require.True(t, a[0].IsWritable && !a[0].IsSigner)
	require.True(t, !a[1].IsWritable && a[1].IsSigner)
	require.True(t, a[2].IsWritable && !a[2].IsSigner)
	require.Equal(t, testSource, a[2].PublicKey)
}

func TestAccounts_Retract(t *testing.T) {
	inst := NewRetractInstruction(testProgram, testAuthority).Build()
	a := inst.Accounts()
	require.Len(t, a, 2)
	require.True(t, a[0].IsWritable && !a[0].IsSigner)
	require.True(t, !a[1].IsWritable && a[1].IsSigner)
}

func TestAccounts_TransferAuthority(t *testing.T) {
	inst := NewTransferAuthorityInstruction(testProgram, testAuthority, testNewAuth).Build()
	a := inst.Accounts()
	require.Len(t, a, 3)
	require.True(t, a[0].IsWritable && !a[0].IsSigner)
	require.True(t, !a[1].IsWritable && a[1].IsSigner)
	require.True(t, !a[2].IsWritable && a[2].IsSigner)
	require.Equal(t, testNewAuth, a[2].PublicKey)
}

func TestAccounts_Finalize(t *testing.T) {
	inst := NewFinalizeInstruction(testProgram, testAuthority, testNextVer).Build()
	a := inst.Accounts()
	require.Len(t, a, 3)
	require.True(t, a[0].IsWritable && !a[0].IsSigner)
	require.True(t, !a[1].IsWritable && a[1].IsSigner)
	require.True(t, !a[2].IsWritable && !a[2].IsSigner)
	require.Equal(t, testNextVer, a[2].PublicKey)
}

func TestCreateBufferInstructions(t *testing.T) {
	insts := NewCreateBufferInstructions(testPayer, testBufferAcct, testAuthority, testRecipient, 1_000, 1024)
	require.Len(t, insts, 2)
	// The first instruction is a system create_account; the second is
	// SetProgramLength, whose program ID must match this package.
	require.Equal(t, ProgramID, insts[1].ProgramID())

	spl := insts[1].Accounts()
	require.Len(t, spl, 3)
	require.Equal(t, testBufferAcct, spl[0].PublicKey)
	require.True(t, spl[0].IsWritable)
}
