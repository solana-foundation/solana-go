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
	"bytes"
	"encoding/binary"
	"testing"

	bin "github.com/gagliardetto/binary"
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
	require.Panics(t,
		func() {
			slice.SplitFrom(-1)
		})
}

// rustAccountFixture builds the expected bincode encoding by hand:
// lamports u64 LE, data u64 LE length + bytes, owner 32 raw bytes,
// executable u8, rent_epoch u64 LE. This pins the wire format against the
// Rust serde encoding of solana_account::Account independent of the codec
// implementation.
func rustAccountFixture(a Account) []byte {
	out := make([]byte, 0, 8+8+len(a.Data)+32+1+8)
	out = binary.LittleEndian.AppendUint64(out, a.Lamports)
	out = binary.LittleEndian.AppendUint64(out, uint64(len(a.Data)))
	out = append(out, a.Data...)
	out = append(out, a.Owner[:]...)
	if a.Executable {
		out = append(out, 1)
	} else {
		out = append(out, 0)
	}
	out = binary.LittleEndian.AppendUint64(out, a.RentEpoch)
	return out
}

func TestAccountBincodeRoundTrip(t *testing.T) {
	owner := MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	cases := []struct {
		name string
		acct Account
	}{
		{
			name: "typical",
			acct: Account{
				Lamports:   1_461_600,
				Data:       []byte{1, 2, 3, 4, 5},
				Owner:      owner,
				Executable: false,
				RentEpoch:  361,
			},
		},
		{
			name: "empty data",
			acct: Account{
				Lamports:  1,
				Owner:     SystemProgramID,
				RentEpoch: ^uint64(0),
			},
		},
		{
			name: "executable",
			acct: Account{
				Lamports:   928_408_320,
				Data:       make([]byte, 36),
				Owner:      MustPublicKeyFromBase58("BPFLoaderUpgradeab1e11111111111111111111111"),
				Executable: true,
			},
		},
		{
			name: "zero value",
			acct: Account{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc, err := tc.acct.MarshalBinary()
			require.NoError(t, err)
			assert.Equal(t, rustAccountFixture(tc.acct), enc, "encoding must match Rust bincode layout")

			dec, err := DecodeAccount(enc)
			require.NoError(t, err)
			assert.Equal(t, tc.acct, *dec)
		})
	}
}

func TestDecodeAccountErrors(t *testing.T) {
	// Truncated header.
	_, err := DecodeAccount([]byte{1, 2, 3})
	require.Error(t, err)

	// Declared data length exceeds the buffer: must error. The explicit
	// bounds check also keeps int(dataLen) from truncating on 32-bit
	// platforms, where ReadNBytes alone could not catch it.
	bad := make([]byte, 16)
	binary.LittleEndian.PutUint64(bad[8:16], 1<<40)
	_, err = DecodeAccount(bad)
	require.Error(t, err)

	// Truncated after data (missing owner/flags/rent epoch).
	acct := Account{Lamports: 5, Data: []byte{9, 9}, RentEpoch: 7}
	enc, err := acct.MarshalBinary()
	require.NoError(t, err)
	_, err = DecodeAccount(enc[:len(enc)-9])
	require.Error(t, err)

	// Executable byte other than 0/1 must be rejected, matching Rust
	// bincode's strict bool decoding.
	enc[len(enc)-9] = 2
	_, err = DecodeAccount(enc)
	require.Error(t, err)
}

func TestDecodeAccountToleratesTrailingBytes(t *testing.T) {
	// Rust's bincode::deserialize free function is declared with
	// allow_trailing_bytes(), and Agave decodes account state through it,
	// so extra bytes after the account (e.g. padding in a fixed-size
	// record) must not fail the decode.
	acct := Account{
		Lamports:  42,
		Data:      []byte{1, 2, 3},
		Owner:     MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		RentEpoch: 9,
	}
	enc, err := acct.MarshalBinary()
	require.NoError(t, err)

	dec, err := DecodeAccount(append(enc, make([]byte, 7)...))
	require.NoError(t, err)
	assert.Equal(t, acct, *dec)
}

func TestAccountStreamRoundTrip(t *testing.T) {
	// Accounts encoded back to back must decode sequentially from a single
	// buffer via the streaming codec, mirroring Rust deserialize_from on a
	// reader carrying multiple values.
	accts := []Account{
		{
			Lamports:  1,
			Data:      []byte{0xAA, 0xBB},
			Owner:     MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
			RentEpoch: 2,
		},
		{
			Lamports:   3,
			Owner:      MustPublicKeyFromBase58("BPFLoaderUpgradeab1e11111111111111111111111"),
			Executable: true,
		},
	}
	buf := new(bytes.Buffer)
	enc := bin.NewBinEncoder(buf)
	for _, a := range accts {
		require.NoError(t, a.MarshalWithEncoder(enc))
	}
	dec := bin.NewBinDecoder(buf.Bytes())
	for _, want := range accts {
		var got Account
		require.NoError(t, got.UnmarshalWithDecoder(dec))
		assert.Equal(t, want, got)
	}
	require.Zero(t, dec.Remaining())
}

func TestDecodeAccountDoesNotAliasInput(t *testing.T) {
	// Owner must be non-zero: clear(enc) zeroes the buffer, so an all-zero
	// owner (e.g. SystemProgramID) would make the owner assertion pass even
	// against an aliased buffer.
	owner := MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	acct := Account{
		Lamports:  42,
		Data:      []byte{1, 2, 3},
		Owner:     owner,
		RentEpoch: 9,
	}
	enc, err := acct.MarshalBinary()
	require.NoError(t, err)

	dec, err := DecodeAccount(enc)
	require.NoError(t, err)

	// Mutating the input buffer after decoding must not change the account.
	clear(enc)
	assert.Equal(t, []byte{1, 2, 3}, dec.Data)
	assert.Equal(t, owner, dec.Owner)
}

func FuzzDecodeAccount(f *testing.F) {
	seedAcct := Account{
		Lamports:  1_461_600,
		Data:      []byte{1, 2, 3, 4, 5},
		Owner:     SystemProgramID,
		RentEpoch: 361,
	}
	seed, err := seedAcct.MarshalBinary()
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte{})
	f.Add(make([]byte, 57))
	f.Fuzz(func(t *testing.T, data []byte) {
		dec, err := DecodeAccount(data)
		if err != nil {
			return
		}
		// Anything that decodes must re-encode to the exact bytes consumed.
		// Trailing input bytes are tolerated, matching Rust's
		// bincode::deserialize free function (declared with
		// allow_trailing_bytes), so compare against the matching prefix.
		enc, err := dec.MarshalBinary()
		require.NoError(t, err)
		require.LessOrEqual(t, len(enc), len(data))
		require.Equal(t, enc, data[:len(enc)])
	})
}
