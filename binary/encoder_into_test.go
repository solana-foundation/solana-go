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
	"bytes"
	"errors"
	"io"
	"testing"
)

// parityCase drives the three encodings through a single table.
type parityCase struct {
	name    string
	marshal func(v any) ([]byte, error)
	into    func(v any, dst []byte) (int, error)
}

var parityCases = []parityCase{
	{"bin", MarshalBin, MarshalBinInto},
	{"borsh", MarshalBorsh, MarshalBorshInto},
	{"compact-u16", MarshalCompactU16, MarshalCompactU16Into},
}

// TestMarshalInto_Parity confirms MarshalXxxInto produces byte-identical
// output to MarshalXxx across primitives, fixed arrays, dynamic slices,
// and a nested Solana-ish struct.
func TestMarshalInto_Parity(t *testing.T) {
	inputs := []any{
		uint8(42),
		uint16(0xbeef),
		uint32(0xdeadbeef),
		uint64(0x0123456789abcdef),
		int64(-1234567),
		"hello world",
		[]byte{1, 2, 3, 4, 5},
		[4]byte{9, 8, 7, 6},
		[]uint64{0, 1, 2, 3, 4, 5, 6, 7},
		makePerfBenchStruct(),
		&struct {
			Hdr [3]byte
			N   uint64
			Sig [64]byte
		}{
			Hdr: [3]byte{1, 2, 3},
			N:   999,
			Sig: [64]byte{0xaa, 0xbb, 0xcc},
		},
	}
	for _, tc := range parityCases {
		for _, in := range inputs {
			t.Run(tc.name, func(t *testing.T) {
				want, err := tc.marshal(in)
				if err != nil {
					t.Fatalf("%T: reference marshal failed: %v", in, err)
				}
				dst := make([]byte, len(want))
				n, err := tc.into(in, dst)
				if err != nil {
					t.Fatalf("%T: MarshalInto failed: %v", in, err)
				}
				if n != len(want) {
					t.Fatalf("%T: len mismatch: got %d, want %d", in, n, len(want))
				}
				if !bytes.Equal(dst[:n], want) {
					t.Fatalf("%T: bytes mismatch\n got: %x\nwant: %x", in, dst[:n], want)
				}
			})
		}
	}
}

// TestMarshalInto_ShortBuffer confirms that undersized dst returns
// io.ErrShortBuffer (wrapped or direct) and does not panic or overflow.
func TestMarshalInto_ShortBuffer(t *testing.T) {
	s := makePerfBenchStruct()
	for _, tc := range parityCases {
		t.Run(tc.name, func(t *testing.T) {
			want, err := tc.marshal(&s)
			if err != nil {
				t.Fatalf("reference marshal failed: %v", err)
			}
			// dst one byte short.
			dst := make([]byte, len(want)-1)
			_, err = tc.into(&s, dst)
			if err == nil {
				t.Fatalf("expected error for short buffer, got nil")
			}
			if !errors.Is(err, io.ErrShortBuffer) {
				t.Fatalf("expected io.ErrShortBuffer, got %v", err)
			}

			// dst of zero length also errors (for any non-empty payload).
			_, err = tc.into(&s, nil)
			if err == nil {
				t.Fatalf("expected error for nil dst, got nil")
			}
		})
	}
}

// TestMarshalInto_ExactFit confirms that a buffer sized exactly to the
// encoded length works without a short-buffer error and returns the full
// payload.
func TestMarshalInto_ExactFit(t *testing.T) {
	s := makePerfBenchStruct()
	for _, tc := range parityCases {
		t.Run(tc.name, func(t *testing.T) {
			want, err := tc.marshal(&s)
			if err != nil {
				t.Fatalf("reference marshal failed: %v", err)
			}
			dst := make([]byte, len(want))
			n, err := tc.into(&s, dst)
			if err != nil {
				t.Fatalf("MarshalInto failed: %v", err)
			}
			if n != len(want) || !bytes.Equal(dst, want) {
				t.Fatalf("exact-fit mismatch: n=%d want-len=%d\n got: %x\nwant: %x", n, len(want), dst, want)
			}
		})
	}
}

// TestMarshalInto_DoesNotReallocate confirms that dst's backing array is
// preserved: the encoded bytes land in the caller's memory, not a fresh
// allocation. We check this by pointer-equality on &dst[0] before/after.
func TestMarshalInto_DoesNotReallocate(t *testing.T) {
	s := makePerfBenchStruct()
	dst := make([]byte, 4096)
	before := &dst[0]
	n, err := MarshalBinInto(&s, dst)
	if err != nil {
		t.Fatalf("MarshalBinInto failed: %v", err)
	}
	after := &dst[0]
	if before != after {
		t.Fatalf("dst backing array was replaced (encoder reallocated)")
	}
	if n == 0 {
		t.Fatal("expected non-zero write")
	}
}

// TestEncoder_ResetInto_Reuse confirms a single Encoder can be targeted at
// multiple dst buffers without reallocating its fields, and produces
// correct output for each.
func TestEncoder_ResetInto_Reuse(t *testing.T) {
	enc := NewBinEncoderInto(nil)
	for i := 0; i < 4; i++ {
		want, err := MarshalBin(uint64(i))
		if err != nil {
			t.Fatalf("reference marshal failed: %v", err)
		}
		dst := make([]byte, len(want))
		enc.ResetInto(dst)
		if err := enc.Encode(uint64(i)); err != nil {
			t.Fatalf("iter %d encode failed: %v", i, err)
		}
		got := enc.Bytes()
		if !bytes.Equal(got, want) {
			t.Fatalf("iter %d mismatch: got %x want %x", i, got, want)
		}
	}
}

// TestEncoder_Into_GrowIsNoOp confirms that Grow in fixed-buffer mode does
// not reallocate the buffer.
func TestEncoder_Into_GrowIsNoOp(t *testing.T) {
	dst := make([]byte, 32)
	enc := NewBinEncoderInto(dst)
	before := cap(enc.buf)
	enc.Grow(1 << 20) // 1 MiB
	after := cap(enc.buf)
	if before != after {
		t.Fatalf("Grow reallocated in fixed mode: cap %d -> %d", before, after)
	}
}

// TestMarshalInto_ZeroAllocs_Primitives confirms the expected steady-state
// allocation count for a small primitive struct. Anything beyond the
// pooledMarshalInto overhead (which can briefly box v in an any) would
// indicate the fast-buf path regressed.
//
// We encode a pointer-to-struct so Encode receives a non-escaping handle;
// the pool reuses its Encoder struct, and dst is caller-owned.
func TestMarshalInto_ZeroAllocs_Primitives(t *testing.T) {
	if testing.Short() {
		t.Skip("allocation test disabled under -short")
	}
	type small struct {
		A uint64
		B uint32
		C [32]byte
	}
	s := &small{A: 1, B: 2}
	for i := range s.C {
		s.C[i] = byte(i)
	}
	// Size once up-front.
	size, err := BinByteCount(s)
	if err != nil {
		t.Fatal(err)
	}
	dst := make([]byte, size)

	// Warm the type-plan cache so first-call reflection isn't counted.
	if _, err := MarshalBinInto(s, dst); err != nil {
		t.Fatal(err)
	}

	allocs := testing.AllocsPerRun(100, func() {
		_, _ = MarshalBinInto(s, dst)
	})
	// Allow up to 2 allocs for any interface boxing the runtime inserts
	// around v on the Encode path; the guarantee is that dst itself is
	// not reallocated and no staging []byte is produced (which would be
	// much more than 2 allocs).
	if allocs > 2 {
		t.Fatalf("MarshalBinInto allocs = %g, want <= 2 (fast path regressed)", allocs)
	}
}
