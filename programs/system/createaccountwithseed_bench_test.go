package system

import (
	"testing"

	bin "github.com/gagliardetto/solana-go/binary"

	"github.com/gagliardetto/solana-go"
)

// makeBenchCreateAccountWithSeed builds a fully-populated
// CreateAccountWithSeed instruction. With 5 marshalable parameter fields
// (*PublicKey × 2, *uint64 × 2, *string), it is the largest instruction
// struct in the repo and a representative target for benchmarking the
// instruction-build hot path: a length-prefixed Rust string + two 32-byte
// foreign-package pointer types + two primitive pointers all in a single
// Marshal call.
func makeBenchCreateAccountWithSeed() *CreateAccountWithSeed {
	base := solana.PublicKey{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	owner := solana.PublicKey{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	funding := solana.PublicKey{0xaa, 0xbb, 0xcc}
	created := solana.PublicKey{0xdd, 0xee, 0xff}
	return NewCreateAccountWithSeedInstruction(
		base,
		"benchmark-seed-string",
		1_000_000_000, // lamports
		8192,          // space
		owner,
		funding,
		created,
		base, // baseAccount (same key as base for the test)
	)
}

// BenchmarkEncode_CreateAccountWithSeed exercises the convenience
// MarshalBin path on the largest instruction struct in the repo. Goes
// through the pooled *Encoder + sync.Pool path inside MarshalBin.
func BenchmarkEncode_CreateAccountWithSeed(b *testing.B) {
	inst := makeBenchCreateAccountWithSeed()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err := bin.MarshalBin(inst)
		if err != nil {
			b.Fatal(err)
		}
		_ = buf
	}
}

// BenchmarkDecode_CreateAccountWithSeed exercises the direct
// NewBinDecoder path; does not benefit from the Decoder pool. Compare
// to BenchmarkDecode_CreateAccountWithSeed_UnmarshalBin which does.
func BenchmarkDecode_CreateAccountWithSeed(b *testing.B) {
	inst := makeBenchCreateAccountWithSeed()
	data, err := bin.MarshalBin(inst)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out CreateAccountWithSeed
		dec := bin.NewBinDecoder(data)
		if err := dec.Decode(&out); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecode_CreateAccountWithSeed_UnmarshalBin exercises the
// pooled UnmarshalBin convenience helper.
func BenchmarkDecode_CreateAccountWithSeed_UnmarshalBin(b *testing.B) {
	inst := makeBenchCreateAccountWithSeed()
	data, err := bin.MarshalBin(inst)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out CreateAccountWithSeed
		if err := bin.UnmarshalBin(&out, data); err != nil {
			b.Fatal(err)
		}
	}
}
