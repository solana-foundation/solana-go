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

package solana

// The tests in this file are ported 1:1 from the upstream anza-xyz/solana-sdk
// `nonce` and `nonce-account` crates to match the same functionality:
//   - nonce/src/state.rs       -> TestNonceStateDefaultIsUninitialized, TestNonceStateSize
//   - nonce/src/versions.rs    -> TestVerifyRecentBlockhash, TestNonceVersionsUpgrade, TestNonceVersionsAuthorize
//   - nonce-account/src/lib.rs -> TestVerifyBadAccountOwnerFails, TestVerifyNonceAccount,
//                                 TestGetSystemAccountKind* (5 cases)
// A few extra Go-specific tests pin the hashing vector, encode/decode round-trip
// and decode error handling.

import (
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func uniqueNonceKey(t *testing.T) PublicKey {
	t.Helper()
	pk, err := NewRandomPrivateKey()
	require.NoError(t, err)
	return pk.PublicKey()
}

func hashFilled(b byte) Hash {
	var h Hash
	for i := range h {
		h[i] = b
	}
	return h
}

// newNonceAccountData mirrors the upstream test helper new_nonce_account: it
// returns the exact serialized account data for a nonce Versions value (8 bytes
// when uninitialized, NonceStateSize when initialized), the way
// AccountSharedData::new_data allocates it.
func newNonceAccountData(t *testing.T, v NonceVersions) []byte {
	t.Helper()
	data, err := v.MarshalBinary()
	require.NoError(t, err)
	return data
}

// ----------------------------------------------------------------------------
// extra Go-specific coverage
// ----------------------------------------------------------------------------

// TestDurableNonceFromBlockhash pins the exact derivation algorithm
// sha256("DURABLE_NONCE" || blockhash) against a precomputed vector.
func TestDurableNonceFromBlockhash(t *testing.T) {
	blockhash := hashFilled(0xab) // matches the Rust Hash::from([171; 32]) test fixture
	dn := DurableNonceFromBlockhash(blockhash)

	const wantBase58 = "Ab9CuAh7wGinrStS9J5txnEZ5XePj3p2BcrjBQcpYDZz"
	assert.Equal(t, wantBase58, dn.String())
	assert.Equal(t, MustHashFromBase58(wantBase58), dn.AsHash())

	wantHex := []byte{
		0x8e, 0x78, 0x35, 0x1b, 0x9e, 0x96, 0x20, 0x90, 0x14, 0xd3, 0xca, 0x60, 0x56, 0xeb, 0x9d, 0x6b,
		0xea, 0x6c, 0x77, 0x2c, 0x38, 0xe8, 0x4f, 0x34, 0x4a, 0x83, 0xf1, 0x3d, 0x58, 0x1a, 0xd3, 0x91,
	}
	gotHash := dn.AsHash()
	assert.Equal(t, wantHex, gotHash[:])

	// Domain separation: deriving from the nonce itself yields a different value.
	assert.NotEqual(t, dn, DurableNonceFromBlockhash(dn.AsHash()))
	assert.False(t, dn.IsZero())
}

// TestNonceVersionsRoundTrip exercises the bincode encode/decode path both for
// an initialized and an uninitialized (zero-padded) nonce account.
func TestNonceVersionsRoundTrip(t *testing.T) {
	authority := uniqueNonceKey(t)
	durableNonce := DurableNonceFromBlockhash(hashFilled(0x07))

	initialized := NewNonceVersions(NewInitializedNonceState(authority, durableNonce, 5000))

	raw, err := initialized.MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, raw, NonceStateSize)

	data, err := initialized.MarshalAccountData()
	require.NoError(t, err)
	assert.Len(t, data, NonceStateSize)
	assert.Equal(t, raw, data)

	got, err := DecodeNonceVersions(data)
	require.NoError(t, err)
	assert.Equal(t, initialized, *got)
	assert.Equal(t, NonceVersionCurrent, got.Version)
	assert.True(t, got.State.IsInitialized())
	assert.Equal(t, authority, got.State.Data.Authority)
	assert.Equal(t, durableNonce, got.State.Data.DurableNonce)
	assert.Equal(t, uint64(5000), got.State.Data.LamportsPerSignature())

	// Uninitialized: 8 bytes raw, NonceStateSize once padded into account data.
	uninit := NewNonceVersions(NonceState{Kind: NonceStateUninitialized})
	rawUninit, err := uninit.MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, rawUninit, 8)

	dataUninit, err := uninit.MarshalAccountData()
	require.NoError(t, err)
	assert.Len(t, dataUninit, NonceStateSize)

	gotUninit, err := DecodeNonceVersions(dataUninit)
	require.NoError(t, err)
	assert.False(t, gotUninit.State.IsInitialized())
	assert.Equal(t, uninit, *gotUninit)
}

func TestDecodeNonceVersionsInvalid(t *testing.T) {
	_, err := DecodeNonceVersions([]byte{0x02, 0, 0, 0}) // version 2 is invalid
	assert.Error(t, err)

	_, err = DecodeNonceVersions([]byte{0x01, 0, 0, 0, 0x05, 0, 0, 0}) // state discriminant 5
	assert.Error(t, err)

	_, err = DecodeNonceVersions([]byte{0x01, 0, 0, 0, 0x01, 0, 0, 0}) // initialized but truncated
	assert.Error(t, err)
}

// ----------------------------------------------------------------------------
// nonce/src/state.rs
// ----------------------------------------------------------------------------

// TestNonceStateDefaultIsUninitialized ports state.rs::default_is_uninitialized.
func TestNonceStateDefaultIsUninitialized(t *testing.T) {
	var state NonceState
	assert.Equal(t, NonceStateUninitialized, state.Kind)
	assert.False(t, state.IsInitialized())
}

// TestNonceStateSize ports state.rs::test_nonce_state_size: the serialized size
// of Versions::new(State::Initialized(Data::default())) is exactly 80 bytes.
func TestNonceStateSize(t *testing.T) {
	v := NewNonceVersions(NewInitializedNonceState(PublicKey{}, DurableNonce{}, 0))
	raw, err := v.MarshalBinary()
	require.NoError(t, err)
	assert.Equal(t, NonceStateSize, len(raw))
}

// ----------------------------------------------------------------------------
// nonce/src/versions.rs
// ----------------------------------------------------------------------------

// TestVerifyRecentBlockhash ports versions.rs::test_verify_recent_blockhash.
func TestVerifyRecentBlockhash(t *testing.T) {
	blockhash := hashFilled(0xab)

	versions := NonceVersions{Version: NonceVersionLegacy, State: NonceState{Kind: NonceStateUninitialized}}
	assertNoVerify(t, versions, blockhash)
	assertNoVerify(t, versions, Hash{})
	versions = NonceVersions{Version: NonceVersionCurrent, State: NonceState{Kind: NonceStateUninitialized}}
	assertNoVerify(t, versions, blockhash)
	assertNoVerify(t, versions, Hash{})

	durableNonce := DurableNonceFromBlockhash(blockhash)
	data := NonceData{
		Authority:     uniqueNonceKey(t),
		DurableNonce:  durableNonce,
		FeeCalculator: FeeCalculator{LamportsPerSignature: 2718},
	}

	versions = NonceVersions{Version: NonceVersionLegacy, State: NonceState{Kind: NonceStateInitialized, Data: data}}
	assertNoVerify(t, versions, Hash{})
	assertNoVerify(t, versions, blockhash)
	assertNoVerify(t, versions, data.Blockhash())
	assertNoVerify(t, versions, durableNonce.AsHash())

	durableNonce = DurableNonceFromBlockhash(durableNonce.AsHash())
	assert.NotEqual(t, data.DurableNonce, durableNonce)
	data.DurableNonce = durableNonce
	versions = NonceVersions{Version: NonceVersionCurrent, State: NonceState{Kind: NonceStateInitialized, Data: data}}
	assertNoVerify(t, versions, blockhash)
	assertNoVerify(t, versions, Hash{})

	got, ok := versions.VerifyRecentBlockhash(data.Blockhash())
	require.True(t, ok)
	assert.Equal(t, data, *got)
	got, ok = versions.VerifyRecentBlockhash(durableNonce.AsHash())
	require.True(t, ok)
	assert.Equal(t, data, *got)
}

func assertNoVerify(t *testing.T, v NonceVersions, h Hash) {
	t.Helper()
	_, ok := v.VerifyRecentBlockhash(h)
	assert.False(t, ok)
}

// TestNonceVersionsUpgrade ports versions.rs::test_nonce_versions_upgrade.
func TestNonceVersionsUpgrade(t *testing.T) {
	// Uninitialized
	_, ok := NonceVersions{Version: NonceVersionLegacy, State: NonceState{Kind: NonceStateUninitialized}}.Upgrade()
	assert.False(t, ok)

	// Initialized
	blockhash := hashFilled(0xab)
	durableNonce := DurableNonceFromBlockhash(blockhash)
	data := NonceData{
		Authority:     uniqueNonceKey(t),
		DurableNonce:  durableNonce,
		FeeCalculator: FeeCalculator{LamportsPerSignature: 2718},
	}
	versions := NonceVersions{Version: NonceVersionLegacy, State: NonceState{Kind: NonceStateInitialized, Data: data}}

	durableNonce = DurableNonceFromBlockhash(durableNonce.AsHash())
	assert.NotEqual(t, data.DurableNonce, durableNonce)
	data.DurableNonce = durableNonce

	upgraded, ok := versions.Upgrade()
	require.True(t, ok)
	assert.Equal(t, NonceVersions{Version: NonceVersionCurrent, State: NonceState{Kind: NonceStateInitialized, Data: data}}, upgraded)

	_, ok = upgraded.Upgrade()
	assert.False(t, ok)
}

// TestNonceVersionsAuthorize ports versions.rs::test_nonce_versions_authorize.
func TestNonceVersionsAuthorize(t *testing.T) {
	// 16 unique signers, mirroring repeat_with(Pubkey::new_unique).take(16).
	signers := make([]PublicKey, 0, 16)
	for range 16 {
		signers = append(signers, uniqueNonceKey(t))
	}

	// Uninitialized
	_, err := NonceVersions{Version: NonceVersionLegacy, State: NonceState{Kind: NonceStateUninitialized}}.
		Authorize(signers, uniqueNonceKey(t))
	assert.ErrorIs(t, err, ErrNonceUninitialized)
	_, err = NonceVersions{Version: NonceVersionCurrent, State: NonceState{Kind: NonceStateUninitialized}}.
		Authorize(signers, uniqueNonceKey(t))
	assert.ErrorIs(t, err, ErrNonceUninitialized)

	durableNonce := DurableNonceFromBlockhash(hashFilled(0xab))

	// Run the Legacy and Current variants identically; both preserve the version.
	for _, version := range []NonceVersion{NonceVersionLegacy, NonceVersionCurrent} {
		data := NonceData{
			Authority:     uniqueNonceKey(t),
			DurableNonce:  durableNonce,
			FeeCalculator: FeeCalculator{LamportsPerSignature: 2718},
		}
		accountAuthority := data.Authority
		versions := NonceVersions{Version: version, State: NonceState{Kind: NonceStateInitialized, Data: data}}

		authority := uniqueNonceKey(t)
		assert.NotEqual(t, authority, accountAuthority)
		want := data
		want.Authority = authority

		// Without the account authority's signature: MissingRequiredSignature.
		_, err := versions.Authorize(signers, authority)
		var missing *MissingRequiredSignatureError
		require.True(t, errors.As(err, &missing))
		assert.Equal(t, accountAuthority, missing.Authority)

		// With the account authority present, authorize succeeds and the
		// version variant is preserved.
		out, err := versions.Authorize(slices.Concat(signers, []PublicKey{accountAuthority}), authority)
		require.NoError(t, err)
		assert.Equal(t, NonceVersions{Version: version, State: NonceState{Kind: NonceStateInitialized, Data: want}}, out)
	}
}

// ----------------------------------------------------------------------------
// nonce-account/src/lib.rs
// ----------------------------------------------------------------------------

// TestVerifyBadAccountOwnerFails ports
// nonce-account::test_verify_bad_account_owner_fails.
func TestVerifyBadAccountOwnerFails(t *testing.T) {
	programID := uniqueNonceKey(t)
	require.NotEqual(t, programID, SystemProgramID)

	data, err := NewNonceVersions(NonceState{Kind: NonceStateUninitialized}).MarshalAccountData()
	require.NoError(t, err)

	_, ok := VerifyNonceAccount(programID, data, Hash{})
	assert.False(t, ok)
}

// TestVerifyNonceAccount ports nonce-account::test_verify_nonce_account.
func TestVerifyNonceAccount(t *testing.T) {
	blockhash := hashFilled(0xab)

	data := newNonceAccountData(t, NonceVersions{Version: NonceVersionLegacy, State: NonceState{Kind: NonceStateUninitialized}})
	assertNoVerifyAccount(t, data, blockhash)
	assertNoVerifyAccount(t, data, Hash{})

	data = newNonceAccountData(t, NonceVersions{Version: NonceVersionCurrent, State: NonceState{Kind: NonceStateUninitialized}})
	assertNoVerifyAccount(t, data, blockhash)
	assertNoVerifyAccount(t, data, Hash{})

	durableNonce := DurableNonceFromBlockhash(blockhash)
	nonceData := NonceData{
		Authority:     uniqueNonceKey(t),
		DurableNonce:  durableNonce,
		FeeCalculator: FeeCalculator{LamportsPerSignature: 2718},
	}

	data = newNonceAccountData(t, NonceVersions{Version: NonceVersionLegacy, State: NonceState{Kind: NonceStateInitialized, Data: nonceData}})
	assertNoVerifyAccount(t, data, blockhash)
	assertNoVerifyAccount(t, data, Hash{})
	assertNoVerifyAccount(t, data, nonceData.Blockhash())
	assertNoVerifyAccount(t, data, durableNonce.AsHash())

	durableNonce = DurableNonceFromBlockhash(durableNonce.AsHash())
	assert.NotEqual(t, nonceData.DurableNonce, durableNonce)
	nonceData.DurableNonce = durableNonce

	data = newNonceAccountData(t, NonceVersions{Version: NonceVersionCurrent, State: NonceState{Kind: NonceStateInitialized, Data: nonceData}})
	assertNoVerifyAccount(t, data, blockhash)
	assertNoVerifyAccount(t, data, Hash{})

	got, ok := VerifyNonceAccount(SystemProgramID, data, nonceData.Blockhash())
	require.True(t, ok)
	assert.Equal(t, nonceData, *got)
	got, ok = VerifyNonceAccount(SystemProgramID, data, durableNonce.AsHash())
	require.True(t, ok)
	assert.Equal(t, nonceData, *got)
}

func assertNoVerifyAccount(t *testing.T, data []byte, h Hash) {
	t.Helper()
	_, ok := VerifyNonceAccount(SystemProgramID, data, h)
	assert.False(t, ok)
}

// TestGetSystemAccountKindSystemOk ports
// nonce-account::test_get_system_account_kind_system_ok. A default account
// (empty data, owner = all-zeros = system program) is a plain system account.
func TestGetSystemAccountKindSystemOk(t *testing.T) {
	kind, ok := GetSystemAccountKind(SystemProgramID, nil)
	require.True(t, ok)
	assert.Equal(t, SystemAccountKindSystem, kind)
}

// TestGetSystemAccountKindNonceOk ports
// nonce-account::test_get_system_account_kind_nonce_ok.
func TestGetSystemAccountKindNonceOk(t *testing.T) {
	data, err := NewNonceVersions(NewInitializedNonceState(PublicKey{}, DurableNonce{}, 0)).MarshalAccountData()
	require.NoError(t, err)

	kind, ok := GetSystemAccountKind(SystemProgramID, data)
	require.True(t, ok)
	assert.Equal(t, SystemAccountKindNonce, kind)
}

// TestGetSystemAccountKindUninitializedNonceAccountFail ports
// nonce-account::test_get_system_account_kind_uninitialized_nonce_account_fail.
func TestGetSystemAccountKindUninitializedNonceAccountFail(t *testing.T) {
	// create_account allocates State::size() bytes of uninitialized state.
	data, err := NewNonceVersions(NonceState{Kind: NonceStateUninitialized}).MarshalAccountData()
	require.NoError(t, err)

	_, ok := GetSystemAccountKind(SystemProgramID, data)
	assert.False(t, ok)
}

// TestGetSystemAccountKindSystemOwnerNonzeroNonNonceDataFail ports
// nonce-account::test_get_system_account_kind_system_owner_nonzero_nonnonce_data_fail.
func TestGetSystemAccountKindSystemOwnerNonzeroNonNonceDataFail(t *testing.T) {
	_, ok := GetSystemAccountKind(SystemProgramID, []byte("other"))
	assert.False(t, ok)
}

// TestGetSystemAccountKindNonsystemOwnerWithNonceDataFail ports
// nonce-account::test_get_system_account_kind_nonsystem_owner_with_nonce_data_fail.
func TestGetSystemAccountKindNonsystemOwnerWithNonceDataFail(t *testing.T) {
	data, err := NewNonceVersions(NewInitializedNonceState(PublicKey{}, DurableNonce{}, 0)).MarshalAccountData()
	require.NoError(t, err)

	_, ok := GetSystemAccountKind(uniqueNonceKey(t), data)
	assert.False(t, ok)
}

func TestLamportsPerSignatureOf(t *testing.T) {
	initialized := NewNonceVersions(NewInitializedNonceState(uniqueNonceKey(t), DurableNonce{}, 5000))
	data, err := initialized.MarshalAccountData()
	require.NoError(t, err)

	fee, ok := LamportsPerSignatureOf(data)
	require.True(t, ok)
	assert.Equal(t, uint64(5000), fee)

	// Uninitialized has no recorded fee.
	uninit, err := NewNonceVersions(NonceState{Kind: NonceStateUninitialized}).MarshalAccountData()
	require.NoError(t, err)
	_, ok = LamportsPerSignatureOf(uninit)
	assert.False(t, ok)
}

// TestVerifyRejectsTrailingBytes guards the strict-decode behavior: a
// system-owned buffer whose first NonceStateSize bytes form a valid initialized
// nonce but which carries extra trailing bytes is NOT a valid nonce account,
// matching upstream's strict bincode::deserialize (which rejects leftover bytes).
func TestVerifyRejectsTrailingBytes(t *testing.T) {
	durableNonce := DurableNonceFromBlockhash(hashFilled(0x42))
	valid, err := NewNonceVersions(NewInitializedNonceState(uniqueNonceKey(t), durableNonce, 5000)).MarshalAccountData()
	require.NoError(t, err)
	require.Len(t, valid, NonceStateSize)

	// Sanity check: the exact-size buffer verifies.
	_, ok := VerifyNonceAccount(SystemProgramID, valid, durableNonce.AsHash())
	require.True(t, ok)

	// Same bytes plus one trailing byte must be rejected by every helper.
	oversized := append(slices.Clone(valid), 0x00)
	_, ok = VerifyNonceAccount(SystemProgramID, oversized, durableNonce.AsHash())
	assert.False(t, ok)
	_, ok = LamportsPerSignatureOf(oversized)
	assert.False(t, ok)
	_, ok = GetSystemAccountKind(SystemProgramID, oversized) // also fails the len==80 gate
	assert.False(t, ok)
}

// FuzzDecodeNonceVersions ensures the deserializer never panics on arbitrary
// input and that anything it accepts round-trips back to identical bytes.
func FuzzDecodeNonceVersions(f *testing.F) {
	for _, seed := range [][]byte{
		nil,
		{0x01, 0, 0, 0, 0, 0, 0, 0}, // Current, Uninitialized
		mustMarshalAccountData(f, NonceVersionCurrent), // full initialized account
		{0x02, 0, 0, 0}, // invalid version
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		v, err := DecodeNonceVersions(data)
		if err != nil {
			return
		}
		// Re-encoding an accepted value must reproduce the exact prefix it was
		// decoded from (the decoder ignores only trailing padding).
		raw, err := v.MarshalBinary()
		require.NoError(t, err)
		require.LessOrEqual(t, len(raw), len(data))
		assert.Equal(t, data[:len(raw)], raw)
	})
}

func mustMarshalAccountData(f *testing.F, version NonceVersion) []byte {
	f.Helper()
	data, err := NonceVersions{
		Version: version,
		State:   NewInitializedNonceState(PublicKey{}, DurableNonce{}, 1),
	}.MarshalAccountData()
	if err != nil {
		f.Fatal(err)
	}
	return data
}
