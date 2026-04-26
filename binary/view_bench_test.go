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

package bin

import (
	"testing"
)

// Benchmarks for the in-place patch use case. Target operation: given a
// 100-byte buffer holding a pre-serialized message, replace the 32-byte
// blockhash at a known offset. Three paths are compared:
//
//   1. ViewAs: single unsafe.Pointer cast, one 32-byte memcpy through
//      the returned *blockhashView. Expected cost: ~memcpy.
//
//   2. copy into subslice: `copy(buf[off:off+32], newHash[:])`. The
//      idiomatic safe-Go equivalent. Should be similar to #1 since
//      both ultimately do one memcpy.
//
//   3. Decode full struct, mutate, re-encode: the "no deserialize_mut"
//      baseline — what callers had to do before ViewAs. Involves a
//      full reflect-driven decode and an encode round-trip.
//
// For (3) the struct we Decode is deliberately simple (just the
// blockhash field wrapped); real-world round-trips would be much more
// expensive because the full message is decoded and re-encoded.

type blockhashMessage struct {
	Blockhash [32]byte
}

const blockhashBenchOffset = 3

func makeBlockhashBench() ([]byte, [32]byte) {
	buf := make([]byte, 100)
	for i := range buf {
		buf[i] = 0x5a
	}
	newHash := [32]byte{}
	for i := range newHash {
		newHash[i] = byte(i)
	}
	return buf, newHash
}

func BenchmarkPatchBlockhash_ViewAs(b *testing.B) {
	buf, newHash := makeBlockhashBench()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v, _ := ViewAs[blockhashView](buf[blockhashBenchOffset:])
		*v = newHash
	}
}

func BenchmarkPatchBlockhash_Copy(b *testing.B) {
	buf, newHash := makeBlockhashBench()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf[blockhashBenchOffset:blockhashBenchOffset+32], newHash[:])
	}
}

func BenchmarkPatchBlockhash_DecodeEncode(b *testing.B) {
	// Pre-serialize the struct so we have real wire bytes to decode.
	orig := blockhashMessage{}
	for i := range orig.Blockhash {
		orig.Blockhash[i] = 0xaa
	}
	wire, err := MarshalBorsh(&orig)
	if err != nil {
		b.Fatal(err)
	}

	_, newHash := makeBlockhashBench()
	outBuf := make([]byte, len(wire))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Decode.
		var msg blockhashMessage
		if err := NewBorshDecoder(wire).Decode(&msg); err != nil {
			b.Fatal(err)
		}
		// Mutate.
		msg.Blockhash = newHash
		// Re-encode.
		if _, err := MarshalBorshInto(&msg, outBuf); err != nil {
			b.Fatal(err)
		}
	}
}

// ---- MarshalPOD / UnmarshalPOD vs reflection-driven Borsh ----

// Larger POD struct for a more meaningful memcpy benchmark (64 bytes).
type bigPOD struct {
	A, B, C, D, E, F, G, H uint64
}

func makeBigPOD() *bigPOD {
	return &bigPOD{
		A: 0x1111, B: 0x2222, C: 0x3333, D: 0x4444,
		E: 0x5555, F: 0x6666, G: 0x7777, H: 0x8888,
	}
}

func BenchmarkMarshalPOD_Pubkey(b *testing.B) {
	var k pubkeyView
	for i := range k {
		k[i] = byte(i)
	}
	dst := make([]byte, 32)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalPOD(&k, dst)
	}
}

func BenchmarkMarshalBorshInto_Pubkey(b *testing.B) {
	var k pubkeyView
	for i := range k {
		k[i] = byte(i)
	}
	dst := make([]byte, 32)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalBorshInto(&k, dst)
	}
}

func BenchmarkMarshalPOD_BigStruct(b *testing.B) {
	v := makeBigPOD()
	dst := make([]byte, 64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalPOD(v, dst)
	}
}

func BenchmarkMarshalBorshInto_BigStruct(b *testing.B) {
	v := makeBigPOD()
	dst := make([]byte, 64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalBorshInto(v, dst)
	}
}

func BenchmarkUnmarshalPOD_BigStruct(b *testing.B) {
	v := makeBigPOD()
	src := make([]byte, 64)
	_, _ = MarshalPOD(v, src)
	var out bigPOD
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UnmarshalPOD(&out, src)
	}
}

func BenchmarkUnmarshalBorsh_BigStruct(b *testing.B) {
	v := makeBigPOD()
	src := make([]byte, 64)
	_, _ = MarshalBorshInto(v, src)
	var out bigPOD
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UnmarshalBorsh(&out, src)
	}
}
