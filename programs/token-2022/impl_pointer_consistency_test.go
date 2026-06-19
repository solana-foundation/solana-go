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
	"reflect"
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

// TestAllBuildersSetPointerImpl pins the fix across every builder, not just
// the two spot-checked above. token-2022 has many more instructions than
// token, so a single missed `&` is easy to overlook; looping over all of them
// and asserting Impl is a pointer of the matching concrete type guards the
// whole surface for almost no extra code.
func TestAllBuildersSetPointerImpl(t *testing.T) {
	cases := []struct {
		name string
		want any
		inst *Instruction
	}{
		{"AmountToUiAmount", (*AmountToUiAmount)(nil), NewAmountToUiAmountInstructionBuilder().Build()},
		{"Approve", (*Approve)(nil), NewApproveInstructionBuilder().Build()},
		{"ApproveChecked", (*ApproveChecked)(nil), NewApproveCheckedInstructionBuilder().Build()},
		{"Burn", (*Burn)(nil), NewBurnInstructionBuilder().Build()},
		{"BurnChecked", (*BurnChecked)(nil), NewBurnCheckedInstructionBuilder().Build()},
		{"CloseAccount", (*CloseAccount)(nil), NewCloseAccountInstructionBuilder().Build()},
		{"CreateNativeMint", (*CreateNativeMint)(nil), NewCreateNativeMintInstructionBuilder().Build()},
		{"FreezeAccount", (*FreezeAccount)(nil), NewFreezeAccountInstructionBuilder().Build()},
		{"GetAccountDataSize", (*GetAccountDataSize)(nil), NewGetAccountDataSizeInstructionBuilder().Build()},
		{"InitializeAccount", (*InitializeAccount)(nil), NewInitializeAccountInstructionBuilder().Build()},
		{"InitializeImmutableOwner", (*InitializeImmutableOwner)(nil), NewInitializeImmutableOwnerInstructionBuilder().Build()},
		{"InitializeMint", (*InitializeMint)(nil), NewInitializeMintInstructionBuilder().Build()},
		{"InitializeMintCloseAuthority", (*InitializeMintCloseAuthority)(nil), NewInitializeMintCloseAuthorityInstructionBuilder().Build()},
		{"InitializeMultisig", (*InitializeMultisig)(nil), NewInitializeMultisigInstructionBuilder().Build()},
		{"InitializeNonTransferableMint", (*InitializeNonTransferableMint)(nil), NewInitializeNonTransferableMintInstructionBuilder().Build()},
		{"InitializePermanentDelegate", (*InitializePermanentDelegate)(nil), NewInitializePermanentDelegateInstructionBuilder().Build()},
		{"MintTo", (*MintTo)(nil), NewMintToInstructionBuilder().Build()},
		{"MintToChecked", (*MintToChecked)(nil), NewMintToCheckedInstructionBuilder().Build()},
		{"PermissionedBurnExtension", (*PermissionedBurnExtension)(nil), NewPermissionedBurnExtensionInstructionBuilder().Build()},
		{"Reallocate", (*Reallocate)(nil), NewReallocateInstructionBuilder().Build()},
		{"Revoke", (*Revoke)(nil), NewRevokeInstructionBuilder().Build()},
		{"SetAuthority", (*SetAuthority)(nil), NewSetAuthorityInstructionBuilder().Build()},
		{"SyncNative", (*SyncNative)(nil), NewSyncNativeInstructionBuilder().Build()},
		{"ThawAccount", (*ThawAccount)(nil), NewThawAccountInstructionBuilder().Build()},
		{"Transfer", (*Transfer)(nil), NewTransferInstructionBuilder().Build()},
		{"TransferChecked", (*TransferChecked)(nil), NewTransferCheckedInstructionBuilder().Build()},
		{"UiAmountToAmount", (*UiAmountToAmount)(nil), NewUiAmountToAmountInstructionBuilder().Build()},
		{"UnwrapLamports", (*UnwrapLamports)(nil), NewUnwrapLamportsInstructionBuilder().Build()},
		{"WithdrawExcessLamports", (*WithdrawExcessLamports)(nil), NewWithdrawExcessLamportsInstructionBuilder().Build()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, reflect.Ptr, reflect.TypeOf(tc.inst.Impl).Kind(),
				"%s Build() must set Impl to a pointer", tc.name)
			require.IsType(t, tc.want, tc.inst.Impl,
				"%s Build() must set Impl to the same *%s type DecodeInstruction returns",
				tc.name, tc.name)
		})
	}
}

// TestDecodeRoundTripCarriesPayload makes the round-trip assert data fidelity,
// not just the pointer type: a Transfer built with a known amount must decode
// back to that same amount, proving the bytes actually carried the payload.
func TestDecodeRoundTripCarriesPayload(t *testing.T) {
	src := solana.MustPublicKeyFromBase58("11111111111111111111111111111112")
	dst := solana.MustPublicKeyFromBase58("11111111111111111111111111111113")
	owner := solana.MustPublicKeyFromBase58("11111111111111111111111111111114")

	data, err := NewTransferInstructionBuilder().
		SetAmount(42).
		SetSourceAccount(src).
		SetDestinationAccount(dst).
		SetOwnerAccount(owner).
		Build().
		Data()
	require.NoError(t, err)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	transfer, ok := decoded.Impl.(*Transfer)
	require.True(t, ok)
	require.NotNil(t, transfer.Amount)
	require.Equal(t, uint64(42), *transfer.Amount)
}
