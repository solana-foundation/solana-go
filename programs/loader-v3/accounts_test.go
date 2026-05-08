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
	"testing"

	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// Each TestAccounts_* replicates the per-instruction account-meta assertions
// from upstream's builder-fn unit tests: program ID, count, and the
// (writable, signer) flags on each account in upstream's documented order.

func TestAccounts_InitializeBuffer(t *testing.T) {
	inst := NewInitializeBufferInstruction(testBuffer, testAuthority).Build()
	require.Equal(t, ProgramID, inst.ProgramID())
	a := inst.Accounts()
	require.Len(t, a, 2)
	require.Equal(t, testBuffer, a[0].PublicKey)
	require.True(t, a[0].IsWritable)
	require.False(t, a[0].IsSigner)
	require.Equal(t, testAuthority, a[1].PublicKey)
	require.False(t, a[1].IsWritable)
	require.False(t, a[1].IsSigner)
}

func TestAccounts_Write(t *testing.T) {
	inst := NewWriteInstruction(testBuffer, testAuthority, 0, nil).Build()
	a := inst.Accounts()
	require.Len(t, a, 2)
	require.Equal(t, testBuffer, a[0].PublicKey)
	require.True(t, a[0].IsWritable)
	require.False(t, a[0].IsSigner)
	require.Equal(t, testAuthority, a[1].PublicKey)
	require.False(t, a[1].IsWritable)
	require.True(t, a[1].IsSigner)
}

func TestAccounts_DeployWithMaxDataLen(t *testing.T) {
	pda := MustGetProgramDataAddress(testProgram)
	inst := NewDeployWithMaxDataLenInstruction(
		testPayer, pda, testProgram, testBuffer, testAuthority, 100, true,
	).Build()
	a := inst.Accounts()
	require.Len(t, a, 8)

	require.Equal(t, testPayer, a[0].PublicKey)
	require.True(t, a[0].IsWritable)
	require.True(t, a[0].IsSigner)

	require.Equal(t, pda, a[1].PublicKey)
	require.True(t, a[1].IsWritable)
	require.False(t, a[1].IsSigner)

	require.Equal(t, testProgram, a[2].PublicKey)
	require.True(t, a[2].IsWritable)
	require.False(t, a[2].IsSigner)

	require.Equal(t, testBuffer, a[3].PublicKey)
	require.True(t, a[3].IsWritable)
	require.False(t, a[3].IsSigner)

	require.Equal(t, ag_solanago.SysVarRentPubkey, a[4].PublicKey)
	require.False(t, a[4].IsWritable)

	require.Equal(t, ag_solanago.SysVarClockPubkey, a[5].PublicKey)
	require.False(t, a[5].IsWritable)

	require.Equal(t, ag_solanago.SystemProgramID, a[6].PublicKey)
	require.False(t, a[6].IsWritable)

	require.Equal(t, testAuthority, a[7].PublicKey)
	require.False(t, a[7].IsWritable)
	require.True(t, a[7].IsSigner)
}

func TestAccounts_Upgrade(t *testing.T) {
	pda := MustGetProgramDataAddress(testProgram)
	inst := NewUpgradeInstruction(testProgram, testBuffer, testAuthority, testSpill, true).Build()
	a := inst.Accounts()
	require.Len(t, a, 7)
	require.Equal(t, pda, a[0].PublicKey)
	require.True(t, a[0].IsWritable)
	require.Equal(t, testProgram, a[1].PublicKey)
	require.True(t, a[1].IsWritable)
	require.Equal(t, testBuffer, a[2].PublicKey)
	require.True(t, a[2].IsWritable)
	require.Equal(t, testSpill, a[3].PublicKey)
	require.True(t, a[3].IsWritable)
	require.Equal(t, ag_solanago.SysVarRentPubkey, a[4].PublicKey)
	require.Equal(t, ag_solanago.SysVarClockPubkey, a[5].PublicKey)
	require.Equal(t, testAuthority, a[6].PublicKey)
	require.True(t, a[6].IsSigner)
}

func TestAccounts_SetBufferAuthority(t *testing.T) {
	inst := NewSetBufferAuthorityInstruction(testBuffer, testAuthority, testNewAuth).Build()
	a := inst.Accounts()
	require.Len(t, a, 3)
	require.Equal(t, testBuffer, a[0].PublicKey)
	require.True(t, a[0].IsWritable)
	require.Equal(t, testAuthority, a[1].PublicKey)
	require.True(t, a[1].IsSigner)
	require.Equal(t, testNewAuth, a[2].PublicKey)
	require.False(t, a[2].IsSigner)
	require.False(t, a[2].IsWritable)
}

func TestAccounts_SetBufferAuthorityChecked(t *testing.T) {
	inst := NewSetBufferAuthorityCheckedInstruction(testBuffer, testAuthority, testNewAuth).Build()
	a := inst.Accounts()
	require.Len(t, a, 3)
	require.Equal(t, testBuffer, a[0].PublicKey)
	require.True(t, a[0].IsWritable)
	require.Equal(t, testAuthority, a[1].PublicKey)
	require.True(t, a[1].IsSigner)
	require.Equal(t, testNewAuth, a[2].PublicKey)
	require.True(t, a[2].IsSigner)
	require.False(t, a[2].IsWritable)
}

func TestAccounts_SetUpgradeAuthority_Drop(t *testing.T) {
	// Passing nil for newAuthority makes the program immutable: only 2 accounts.
	inst := NewSetUpgradeAuthorityInstruction(testProgram, testAuthority, nil).Build()
	a := inst.Accounts()
	require.Len(t, a, 2)
	pda := MustGetProgramDataAddress(testProgram)
	require.Equal(t, pda, a[0].PublicKey)
	require.True(t, a[0].IsWritable)
	require.Equal(t, testAuthority, a[1].PublicKey)
	require.True(t, a[1].IsSigner)
}

func TestAccounts_SetUpgradeAuthority_Transfer(t *testing.T) {
	inst := NewSetUpgradeAuthorityInstruction(testProgram, testAuthority, &testNewAuth).Build()
	a := inst.Accounts()
	require.Len(t, a, 3)
	require.Equal(t, testNewAuth, a[2].PublicKey)
	require.False(t, a[2].IsSigner)
}

func TestAccounts_SetUpgradeAuthorityChecked(t *testing.T) {
	inst := NewSetUpgradeAuthorityCheckedInstruction(testProgram, testAuthority, testNewAuth).Build()
	a := inst.Accounts()
	require.Len(t, a, 3)
	require.Equal(t, testNewAuth, a[2].PublicKey)
	require.True(t, a[2].IsSigner)
}

func TestAccounts_Close_Three(t *testing.T) {
	inst := NewCloseInstruction(testBuffer, testPayer, testAuthority, false).Build()
	a := inst.Accounts()
	require.Len(t, a, 3)
	require.Equal(t, testBuffer, a[0].PublicKey)
	require.True(t, a[0].IsWritable)
	require.Equal(t, testPayer, a[1].PublicKey)
	require.True(t, a[1].IsWritable)
	require.Equal(t, testAuthority, a[2].PublicKey)
	require.True(t, a[2].IsSigner)
}

func TestAccounts_CloseAny_Uninitialized(t *testing.T) {
	// Uninitialized close: no authority, no program.
	inst := NewCloseAnyInstruction(testBuffer, testPayer, nil, nil, false).Build()
	a := inst.Accounts()
	require.Len(t, a, 2)
}

func TestAccounts_CloseAny_ProgramData(t *testing.T) {
	inst := NewCloseAnyInstruction(testBuffer, testPayer, &testAuthority, &testProgram, true).Build()
	a := inst.Accounts()
	require.Len(t, a, 4)
	require.Equal(t, testProgram, a[3].PublicKey)
	require.True(t, a[3].IsWritable)
}

func TestAccounts_ExtendProgram_NoPayer(t *testing.T) {
	inst := NewExtendProgramInstruction(testProgram, nil, 10_240).Build()
	a := inst.Accounts()
	require.Len(t, a, 2)
	pda := MustGetProgramDataAddress(testProgram)
	require.Equal(t, pda, a[0].PublicKey)
	require.True(t, a[0].IsWritable)
	require.Equal(t, testProgram, a[1].PublicKey)
	require.True(t, a[1].IsWritable)
}

func TestAccounts_ExtendProgram_WithPayer(t *testing.T) {
	inst := NewExtendProgramInstruction(testProgram, &testPayer, 10_240).Build()
	a := inst.Accounts()
	require.Len(t, a, 4)
	require.Equal(t, ag_solanago.SystemProgramID, a[2].PublicKey)
	require.False(t, a[2].IsWritable)
	require.False(t, a[2].IsSigner)
	require.Equal(t, testPayer, a[3].PublicKey)
	require.True(t, a[3].IsWritable)
	require.True(t, a[3].IsSigner)
}

func TestCreateBufferInstructions(t *testing.T) {
	insts := NewCreateBufferInstructions(testPayer, testBuffer, testAuthority, 1000, 128)
	require.Len(t, insts, 2)
	// ix1 is the InitializeBuffer; its program ID must be the loader.
	require.Equal(t, ProgramID, insts[1].ProgramID())
}

func TestDeployWithMaxProgramLenInstructions(t *testing.T) {
	insts := NewDeployWithMaxProgramLenInstructions(
		testPayer, testProgram, testBuffer, testAuthority,
		1_000, 2_000, false,
	)
	require.Len(t, insts, 2)
	require.Equal(t, ProgramID, insts[1].ProgramID())
}

func TestPDA_GetProgramDataAddress(t *testing.T) {
	// Spot-check: derivation is stable for the same inputs and distinct
	// from the program address itself.
	pda, _, err := GetProgramDataAddress(testProgram)
	require.NoError(t, err)
	require.NotEqual(t, testProgram, pda)

	pda2, _, err := GetProgramDataAddress(testProgram)
	require.NoError(t, err)
	require.Equal(t, pda, pda2)
}
