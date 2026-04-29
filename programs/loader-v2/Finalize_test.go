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

// TestFinalize_Instruction_Bincode reproduces:
//
//	loader_v2_interface::finalize(&account, &program_id)
//
// which emits just the u32 LE discriminant (1).
func TestFinalize_Instruction_Bincode(t *testing.T) {
	account := ag_solanago.MustPublicKeyFromBase58("7QcXLBB23bJ4q5QUXpxLkQBr37g8mNEPSSPyVvU22qUS")

	inst, err := NewFinalizeInstruction(account).ValidateAndBuild()
	require.NoError(t, err)

	data, err := inst.Data()
	require.NoError(t, err)

	require.Equal(t, []byte{0x01, 0x00, 0x00, 0x00}, data)
}

// TestFinalize_Accounts mirrors the Rust helper: [target: W+S, rent: R].
func TestFinalize_Accounts(t *testing.T) {
	account := ag_solanago.MustPublicKeyFromBase58("7QcXLBB23bJ4q5QUXpxLkQBr37g8mNEPSSPyVvU22qUS")
	inst := NewFinalizeInstruction(account).Build()

	accounts := inst.Accounts()
	require.Len(t, accounts, 2)

	require.Equal(t, account, accounts[0].PublicKey)
	require.True(t, accounts[0].IsWritable)
	require.True(t, accounts[0].IsSigner)

	require.Equal(t, ag_solanago.SysVarRentPubkey, accounts[1].PublicKey)
	require.False(t, accounts[1].IsWritable)
	require.False(t, accounts[1].IsSigner)
}
