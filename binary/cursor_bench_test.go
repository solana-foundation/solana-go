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
	"encoding/binary"
	"testing"
)

// Solana-shaped transaction header: three 1-byte counters, a compact-u16
// length prefix, N 32-byte pubkeys, a 32-byte blockhash. A common shape
// for hand-rolled Solana encoders.
type txHeaderBench struct {
	NumReqSigs      uint8
	NumROSigned     uint8
	NumROUnsigned   uint8
	AccountKeys     [][32]byte // e.g. 8 keys
	RecentBlockhash [32]byte
}

func makeTxHeaderBench() *txHeaderBench {
	h := &txHeaderBench{
		NumReqSigs:    1,
		NumROSigned:   0,
		NumROUnsigned: 1,
		AccountKeys:   make([][32]byte, 8),
	}
	for i := range h.AccountKeys {
		for j := range h.AccountKeys[i] {
			h.AccountKeys[i][j] = byte(i*32 + j)
		}
	}
	for i := range h.RecentBlockhash {
		h.RecentBlockhash[i] = byte(0xaa + i)
	}
	return h
}

func txHeaderSize(h *txHeaderBench) int {
	// 3 counter bytes + compact-u16 (1-3) + N*32 + 32 blockhash.
	return 3 + compactU16Size(len(h.AccountKeys)) + len(h.AccountKeys)*32 + 32
}

func compactU16Size(n int) int {
	switch {
	case n <= 0x7f:
		return 1
	case n <= 0x3fff:
		return 2
	default:
		return 3
	}
}

// Hand-rolled encode via Cursor — one direct memory poke per primitive.
func encodeTxHeaderCursor(dst []byte, h *txHeaderBench) int {
	c := NewCursor(dst).
		WriteU8(h.NumReqSigs).
		WriteU8(h.NumROSigned).
		WriteU8(h.NumROUnsigned).
		WriteLenCompactU16(len(h.AccountKeys))
	for i := range h.AccountKeys {
		c.WriteBytes(h.AccountKeys[i][:])
	}
	c.WriteBytes(h.RecentBlockhash[:])
	return c.Len()
}

// Hand-rolled encode via Encoder in fixed-buffer mode — each primitive
// goes through WriteXxx → scratch buffer → toWriter → append.
func encodeTxHeaderEncoder(dst []byte, h *txHeaderBench) (int, error) {
	e := NewCompactU16EncoderInto(dst)
	if err := e.WriteUint8(h.NumReqSigs); err != nil {
		return 0, err
	}
	if err := e.WriteUint8(h.NumROSigned); err != nil {
		return 0, err
	}
	if err := e.WriteUint8(h.NumROUnsigned); err != nil {
		return 0, err
	}
	if err := e.WriteCompactU16(len(h.AccountKeys)); err != nil {
		return 0, err
	}
	for i := range h.AccountKeys {
		if err := e.WriteBytes(h.AccountKeys[i][:], false); err != nil {
			return 0, err
		}
	}
	if err := e.WriteBytes(h.RecentBlockhash[:], false); err != nil {
		return 0, err
	}
	return e.Written(), nil
}

// Raw encode: direct binary.LittleEndian calls with no wrapper. The
// hard floor for what the Cursor should approach after inlining.
func encodeTxHeaderRaw(dst []byte, h *txHeaderBench) int {
	pos := 0
	dst[pos] = h.NumReqSigs
	pos++
	dst[pos] = h.NumROSigned
	pos++
	dst[pos] = h.NumROUnsigned
	pos++
	// Inline compact-u16 for len <= 127 (fast path — our bench key count
	// fits).
	n := len(h.AccountKeys)
	if n <= 0x7f {
		dst[pos] = byte(n)
		pos++
	} else if n <= 0x3fff {
		dst[pos] = byte(n&0x7f) | 0x80
		dst[pos+1] = byte(n >> 7)
		pos += 2
	} else {
		dst[pos] = byte(n&0x7f) | 0x80
		dst[pos+1] = byte((n>>7)&0x7f) | 0x80
		dst[pos+2] = byte(n >> 14)
		pos += 3
	}
	for i := range h.AccountKeys {
		pos += copy(dst[pos:], h.AccountKeys[i][:])
	}
	pos += copy(dst[pos:], h.RecentBlockhash[:])
	return pos
}

// BenchmarkTxHeader_Cursor measures the hand-rolled Cursor encode path.
func BenchmarkTxHeader_Cursor(b *testing.B) {
	h := makeTxHeaderBench()
	dst := make([]byte, txHeaderSize(h))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeTxHeaderCursor(dst, h)
	}
}

// BenchmarkTxHeader_Encoder measures the same payload via Encoder in
// fixed-buffer mode — the #2 path.
func BenchmarkTxHeader_Encoder(b *testing.B) {
	h := makeTxHeaderBench()
	dst := make([]byte, txHeaderSize(h))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encodeTxHeaderEncoder(dst, h)
	}
}

// BenchmarkTxHeader_Raw is the lower bound: direct encoding/binary calls
// with no wrapper. The Cursor should approach this after inlining.
func BenchmarkTxHeader_Raw(b *testing.B) {
	h := makeTxHeaderBench()
	dst := make([]byte, txHeaderSize(h))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeTxHeaderRaw(dst, h)
	}
}

// ---- Primitive-heavy micro-benchmark (8 u64 writes) ----

func BenchmarkCursor_8xU64LE(b *testing.B) {
	dst := make([]byte, 64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewCursor(dst).
			WriteU64LE(1).
			WriteU64LE(2).
			WriteU64LE(3).
			WriteU64LE(4).
			WriteU64LE(5).
			WriteU64LE(6).
			WriteU64LE(7).
			WriteU64LE(8)
	}
}

func BenchmarkEncoder_8xU64LE(b *testing.B) {
	dst := make([]byte, 64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := NewBinEncoderInto(dst)
		_ = e.WriteUint64(1, binary.LittleEndian)
		_ = e.WriteUint64(2, binary.LittleEndian)
		_ = e.WriteUint64(3, binary.LittleEndian)
		_ = e.WriteUint64(4, binary.LittleEndian)
		_ = e.WriteUint64(5, binary.LittleEndian)
		_ = e.WriteUint64(6, binary.LittleEndian)
		_ = e.WriteUint64(7, binary.LittleEndian)
		_ = e.WriteUint64(8, binary.LittleEndian)
	}
}
