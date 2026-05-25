// Copyright 2026 github.com/gagliardetto
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

package token2022

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// TestBuilderImplIsPointer_MatchesDecode pins the fix for issue #222:
// the Impl set by a builder's Build() must be the same pointer kind as
// the Impl set by DecodeInstruction, so a caller can use the same type
// assertion in both directions without crashing.
func TestBuilderImplIsPointer_MatchesDecode(t *testing.T) {
	src := solana.MustPublicKeyFromBase58("11111111111111111111111111111112")
	dst := solana.MustPublicKeyFromBase58("11111111111111111111111111111113")
	owner := solana.MustPublicKeyFromBase58("11111111111111111111111111111114")
	mint := solana.MustPublicKeyFromBase58("11111111111111111111111111111115")

	built := NewMintToInstructionBuilder().
		SetAmount(1).
		SetMintAccount(mint).
		SetDestinationAccount(dst).
		SetAuthorityAccount(owner).
		Build()

	_, ok := built.Impl.(*MintTo)
	require.True(t, ok,
		"Build() must place a *MintTo into Impl so the same type "+
			"assertion works whether the instruction was built locally "+
			"or decoded from bytes (issue #222)")

	// Cross-check via DecodeInstruction on a different builder to make
	// sure the decode side also produces *Transfer (not Transfer), so
	// the assertion above is a real contract and not a one-off.
	transferData, err := NewTransferInstructionBuilder().
		SetAmount(1).
		SetSourceAccount(src).
		SetDestinationAccount(dst).
		SetOwnerAccount(owner).
		Build().
		Data()
	require.NoError(t, err)

	decoded, err := DecodeInstruction(nil, transferData)
	require.NoError(t, err)
	_, ok = decoded.Impl.(*Transfer)
	require.True(t, ok, "DecodeInstruction was already a *Transfer; "+
		"Build() now matches that contract")
}
