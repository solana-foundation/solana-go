// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
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

package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccount(t *testing.T) {
	a := NewWallet()
	privateKey := a.PrivateKey
	public := a.PublicKey()

	a2, err := WalletFromPrivateKeyBase58(privateKey.String())
	require.NoError(t, err)

	require.Equal(t, privateKey, a2.PrivateKey)
	require.Equal(t, public, a2.PublicKey())
}

func Test_AccountMeta_less(t *testing.T) {
	pkey := MustPublicKeyFromBase58("SysvarS1otHashes111111111111111111111111111")
	tests := []struct {
		name   string
		left   *AccountMeta
		right  *AccountMeta
		expect bool
	}{
		{
			name:   "accounts are equal",
			left:   &AccountMeta{PublicKey: pkey, IsSigner: false, IsWritable: false},
			right:  &AccountMeta{PublicKey: pkey, IsSigner: false, IsWritable: false},
			expect: false,
		},
		{
			name:   "left is a signer, right is not a signer",
			left:   &AccountMeta{PublicKey: pkey, IsSigner: true, IsWritable: false},
			right:  &AccountMeta{PublicKey: pkey, IsSigner: false, IsWritable: false},
			expect: true,
		},
		{
			name:   "left is not a signer, right is a signer",
			left:   &AccountMeta{PublicKey: pkey, IsSigner: false, IsWritable: false},
			right:  &AccountMeta{PublicKey: pkey, IsSigner: true, IsWritable: false},
			expect: false,
		},
		{
			name:   "left is writable, right is not writable",
			left:   &AccountMeta{PublicKey: pkey, IsSigner: false, IsWritable: true},
			right:  &AccountMeta{PublicKey: pkey, IsSigner: false, IsWritable: false},
			expect: true,
		},
		{
			name:   "left is not writable, right is writable",
			left:   &AccountMeta{PublicKey: pkey, IsSigner: false, IsWritable: false},
			right:  &AccountMeta{PublicKey: pkey, IsSigner: false, IsWritable: true},
			expect: false,
		},
		{
			name:   "both are signers and left is writable, right is not writable",
			left:   &AccountMeta{PublicKey: pkey, IsSigner: true, IsWritable: true},
			right:  &AccountMeta{PublicKey: pkey, IsSigner: true, IsWritable: false},
			expect: true,
		},
		{
			name:   "both are signers andleft is not writable, right is writable",
			left:   &AccountMeta{PublicKey: pkey, IsSigner: true, IsWritable: false},
			right:  &AccountMeta{PublicKey: pkey, IsSigner: true, IsWritable: true},
			expect: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, test.left.less(test.right))
		})
	}
}

func TestAccountMetaSlice(t *testing.T) {
	pkey1 := MustPublicKeyFromBase58("SysvarS1otHashes111111111111111111111111111")

	var slice AccountMetaSlice

	setting := []*AccountMeta{
		{PublicKey: pkey1, IsSigner: true, IsWritable: false},
	}
	err := slice.SetAccounts(setting)
	require.NoError(t, err)

	require.Len(t, slice, 1)
	require.Equal(t, setting[0], slice[0])
	require.Equal(t, setting, slice.GetAccounts())

	{
		pkey2 := MustPublicKeyFromBase58("BPFLoaderUpgradeab1e11111111111111111111111")

		meta := NewAccountMeta(pkey2, true, false)
		slice.Append(meta)

		require.Len(t, slice, 2)
		require.Equal(t, meta, slice[1])
		require.Equal(t, meta, slice.GetAccounts()[1])
	}
}

func TestNewAccountMeta(t *testing.T) {
	pkey := MustPublicKeyFromBase58("SysvarS1otHashes111111111111111111111111111")

	isWritable := false
	isSigner := true

	out := NewAccountMeta(pkey, isWritable, isSigner)

	require.NotNil(t, out)

	require.Equal(t, isSigner, out.IsSigner)
	require.Equal(t, isWritable, out.IsWritable)
}

func TestMeta(t *testing.T) {
	pkey := MustPublicKeyFromBase58("SysvarS1otHashes111111111111111111111111111")

	meta := Meta(pkey)
	require.NotNil(t, meta)
	require.Equal(t, pkey, meta.PublicKey)

	require.False(t, meta.IsSigner)
	require.False(t, meta.IsWritable)

	meta.SIGNER()

	require.True(t, meta.IsSigner)
	require.False(t, meta.IsWritable)

	meta.WRITE()

	require.True(t, meta.IsSigner)
	require.True(t, meta.IsWritable)
}

func TestSplitFrom(t *testing.T) {
	slice := make(AccountMetaSlice, 0)
	slice = append(slice, Meta(BPFLoaderDeprecatedProgramID))
	slice = append(slice, Meta(TokenProgramID))
	slice = append(slice, Meta(TokenLendingProgramID))
	slice = append(slice, Meta(SPLAssociatedTokenAccountProgramID))
	slice = append(slice, Meta(MemoProgramID))

	require.Len(t, slice, 5)

	{
		part1, part2 := slice.SplitFrom(0)
		require.Len(t, part1, 0)
		require.Len(t, part2, 5)
	}
	{
		part1, part2 := slice.SplitFrom(1)
		require.Len(t, part1, 1)
		require.Len(t, part2, 4)
		require.Equal(t, Meta(BPFLoaderDeprecatedProgramID), part1[0])
		require.Equal(t, Meta(TokenProgramID), part2[0])
		require.Equal(t, Meta(TokenLendingProgramID), part2[1])
		require.Equal(t, Meta(SPLAssociatedTokenAccountProgramID), part2[2])
		require.Equal(t, Meta(MemoProgramID), part2[3])
	}
	{
		part1, part2 := slice.SplitFrom(2)
		require.Len(t, part1, 2)
		require.Len(t, part2, 3)
	}
	{
		part1, part2 := slice.SplitFrom(3)
		require.Len(t, part1, 3)
		require.Len(t, part2, 2)
	}
	{
		part1, part2 := slice.SplitFrom(4)
		require.Len(t, part1, 4)
		require.Len(t, part2, 1)
	}
	{
		part1, part2 := slice.SplitFrom(5)
		require.Len(t, part1, 5)
		require.Len(t, part2, 0)
	}
	{
		part1, part2 := slice.SplitFrom(6)
		require.Len(t, part1, 5)
		require.Len(t, part2, 0)
	}
	{
		part1, part2 := slice.SplitFrom(10000)
		require.Len(t, part1, 5)
		require.Len(t, part2, 0)
	}
	{
		part1, part2 := slice.SplitFrom(-1)
		require.Len(t, part1, 0)
		require.Len(t, part2, 5)
	}
}
