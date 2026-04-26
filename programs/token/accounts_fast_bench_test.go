// Copyright 2024 github.com/gagliardetto
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

package token

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
)

// Four bench payloads:
//
//  mintAllSet       — mint with both authorities present (worst-case
//                     for the traditional decoder: two COption branches)
//  mintNoOptionals  — mint with no authorities (best-case)
//  accountAllSet    — account with delegate, is_native, close_authority
//                     all present
//  accountTypical   — account with no optional fields (most common
//                     shape on-chain; majority of token accounts)
//
// The "typical" shape is what a Geyser subscriber actually processes
// at volume, so its numbers are the ones to care about.

func benchMintAllSet(b *testing.B) []byte {
	b.Helper()
	authMint := benchKey(1)
	authFreeze := benchKey(2)
	m := &Mint{
		MintAuthority:   &authMint,
		Supply:          1_000_000_000,
		Decimals:        9,
		IsInitialized:   true,
		FreezeAuthority: &authFreeze,
	}
	return encodeMintToBytes(b, m)
}

func benchMintNoOptionals(b *testing.B) []byte {
	b.Helper()
	return encodeMintToBytes(b, &Mint{Supply: 42, Decimals: 6, IsInitialized: true})
}

func benchAccountAllSet(b *testing.B) []byte {
	b.Helper()
	mint := benchKey(3)
	owner := benchKey(4)
	delegate := benchKey(5)
	closeAuth := benchKey(6)
	native := uint64(2_039_280)
	a := &Account{
		Mint:            mint,
		Owner:           owner,
		Amount:          999_999,
		Delegate:        &delegate,
		State:           Initialized,
		IsNative:        &native,
		DelegatedAmount: 10_000,
		CloseAuthority:  &closeAuth,
	}
	return encodeAccountToBytes(b, a)
}

func benchAccountTypical(b *testing.B) []byte {
	b.Helper()
	mint := benchKey(7)
	owner := benchKey(8)
	a := &Account{
		Mint:   mint,
		Owner:  owner,
		Amount: 1_000_000,
		State:  Initialized,
	}
	return encodeAccountToBytes(b, a)
}

func benchKey(x byte) solana.PublicKey {
	var k solana.PublicKey
	for i := range k {
		k[i] = x + byte(i)
	}
	return k
}

func encodeMintToBytes(b *testing.B, m *Mint) []byte {
	b.Helper()
	buf := make([]byte, MintAccountSize)
	enc := bin.NewBinEncoderInto(buf)
	if err := m.MarshalWithEncoder(enc); err != nil {
		b.Fatal(err)
	}
	return buf
}

func encodeAccountToBytes(b *testing.B, a *Account) []byte {
	b.Helper()
	buf := make([]byte, TokenAccountSize)
	enc := bin.NewBinEncoderInto(buf)
	if err := a.MarshalWithEncoder(enc); err != nil {
		b.Fatal(err)
	}
	return buf
}

// ---- Mint decode benchmarks ----

// Traditional: uses the hand-written UnmarshalWithDecoder through the
// bin.Decoder. Bounds check per field read, method dispatch per call.
func BenchmarkDecodeMint_Traditional_AllSet(b *testing.B) {
	wire := benchMintAllSet(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m Mint
		if err := bin.NewBinDecoder(wire).Decode(&m); err != nil {
			b.Fatal(err)
		}
	}
}

// Fast raw decode: one memcpy into MintRaw. No Mint struct materialized.
// This is the fastest path if the caller is happy to work with the
// raw layout and call accessors.
func BenchmarkDecodeMint_Raw_AllSet(b *testing.B) {
	wire := benchMintAllSet(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var raw MintRaw
		if err := DecodeMintRaw(&raw, wire); err != nil {
			b.Fatal(err)
		}
	}
}

// Fast view: zero-copy, aliases the wire buffer. Accessor calls remain
// free (single u64/u32 load each). Fastest possible for read-only
// access patterns.
func BenchmarkDecodeMint_View_AllSet(b *testing.B) {
	wire := benchMintAllSet(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		view, err := ViewMintRaw(wire)
		if err != nil {
			b.Fatal(err)
		}
		// Touch a couple of fields so the compiler cannot dead-code
		// eliminate the call.
		_ = view.SupplyU64()
		_ = view.HasMintAuthority()
	}
}

// Fast decode with materialized Mint: POD fast path + ToMint()
// conversion. Drop-in replacement for the traditional decode, still
// pays for *PublicKey allocations on present COption fields.
func BenchmarkDecodeMint_Fast_AllSet(b *testing.B) {
	wire := benchMintAllSet(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DecodeMintFast(wire); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeMint_Traditional_NoOptionals(b *testing.B) {
	wire := benchMintNoOptionals(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m Mint
		if err := bin.NewBinDecoder(wire).Decode(&m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeMint_Raw_NoOptionals(b *testing.B) {
	wire := benchMintNoOptionals(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var raw MintRaw
		if err := DecodeMintRaw(&raw, wire); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeMint_Fast_NoOptionals(b *testing.B) {
	wire := benchMintNoOptionals(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DecodeMintFast(wire); err != nil {
			b.Fatal(err)
		}
	}
}

// ---- Account decode benchmarks ----

func BenchmarkDecodeAccount_Traditional_AllSet(b *testing.B) {
	wire := benchAccountAllSet(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var a Account
		if err := bin.NewBinDecoder(wire).Decode(&a); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeAccount_Raw_AllSet(b *testing.B) {
	wire := benchAccountAllSet(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var raw AccountRaw
		if err := DecodeAccountRaw(&raw, wire); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeAccount_View_AllSet(b *testing.B) {
	wire := benchAccountAllSet(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		view, err := ViewAccountRaw(wire)
		if err != nil {
			b.Fatal(err)
		}
		_ = view.AmountU64()
		_ = view.HasDelegate()
	}
}

func BenchmarkDecodeAccount_Fast_AllSet(b *testing.B) {
	wire := benchAccountAllSet(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DecodeAccountFast(wire); err != nil {
			b.Fatal(err)
		}
	}
}

// Typical: most on-chain token accounts have no optional fields. This
// is the shape that dominates Geyser subscriber workloads.
func BenchmarkDecodeAccount_Traditional_Typical(b *testing.B) {
	wire := benchAccountTypical(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var a Account
		if err := bin.NewBinDecoder(wire).Decode(&a); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeAccount_Raw_Typical(b *testing.B) {
	wire := benchAccountTypical(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var raw AccountRaw
		if err := DecodeAccountRaw(&raw, wire); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeAccount_View_Typical(b *testing.B) {
	wire := benchAccountTypical(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		view, err := ViewAccountRaw(wire)
		if err != nil {
			b.Fatal(err)
		}
		_ = view.AmountU64()
	}
}

func BenchmarkDecodeAccount_Fast_Typical(b *testing.B) {
	wire := benchAccountTypical(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DecodeAccountFast(wire); err != nil {
			b.Fatal(err)
		}
	}
}
