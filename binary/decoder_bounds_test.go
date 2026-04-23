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
	stdbinary "encoding/binary"
	"errors"
	"io"
	"testing"
)

// varUint writes a Uvarint of v to dst and returns the bytes written.
func varUint(v uint64) []byte {
	buf := make([]byte, stdbinary.MaxVarintLen64)
	n := stdbinary.PutUvarint(buf, v)
	return buf[:n]
}

// compactU16 writes a compact-u16 length encoding of v and returns bytes.
func compactU16(v uint16) []byte {
	buf := make([]byte, 3)
	n, err := PutCompactU16Length(buf, int(v))
	if err != nil {
		panic(err)
	}
	return buf[:n]
}

// u32LE writes a little-endian uint32 and returns bytes.
func u32LE(v uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf
}

// ---------- Slice: oversize length prefix with no backing data ----------

// The "length overruns Remaining() bytes" case returns io.ErrUnexpectedEOF
// (not ErrSliceLenTooLarge) to preserve backward compatibility with
// historical callers keying off io.ErrUnexpectedEOF.

func TestDecode_Bin_SliceLenOverRemaining(t *testing.T) {
	wire := varUint(1_000_000_000)
	var got []uint32
	err := NewBinDecoder(wire).Decode(&got)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

func TestDecode_Borsh_SliceLenOverRemaining(t *testing.T) {
	wire := u32LE(0x7FFF_FFFF)
	var got []uint32
	err := NewBorshDecoder(wire).Decode(&got)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

func TestDecode_CompactU16_SliceLenOverRemaining(t *testing.T) {
	wire := compactU16(0xFFFF)
	var got []uint32
	err := NewCompactU16Decoder(wire).Decode(&got)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

// ---------- Slice: element-size-aware (PoD elements) ----------

// Even when length * 1 byte <= Remaining, a slice of [32]byte with a claimed
// length that would require 32 * l bytes should be rejected before MakeSlice
// allocates.
// Element-size-aware rejection: a slice of [32]byte elements with a length
// prefix that would require (l * 32) bytes but only partial payload is
// present. Under the old `l > Remaining()` check (which treats every
// element as 1 byte), claiming 50 pubkeys with 80 payload bytes passes
// (50 < 80) — a malicious caller got a 50*32 = 1600-byte allocation from
// 80 wire bytes. The new element-size-aware check rejects it (50*32 > 80)
// and surfaces io.ErrUnexpectedEOF.
func TestDecode_Bin_SliceElemSizeAware(t *testing.T) {
	type Pubkey [32]byte
	payload := make([]byte, 80)
	wire := append(varUint(50), payload...)
	var got []Pubkey
	err := NewBinDecoder(wire).Decode(&got)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

// ---------- Map: oversize length prefix ----------

// Same rationale as for slices: "overruns buffer" surfaces as
// io.ErrUnexpectedEOF (new behavior where previously the decoder would
// allocate billions of entries and run SetMapIndex in a tight loop until
// memory exhaustion).

func TestDecode_Bin_MapLenOverRemaining(t *testing.T) {
	wire := varUint(1_000_000_000)
	got := map[string]uint32{}
	err := NewBinDecoder(wire).Decode(&got)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

func TestDecode_Borsh_MapLenOverRemaining(t *testing.T) {
	wire := u32LE(0x7FFF_FFFF)
	got := map[string]uint32{}
	err := NewBorshDecoder(wire).Decode(&got)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

func TestDecode_CompactU16_MapLenOverRemaining(t *testing.T) {
	wire := compactU16(0xFFFF)
	got := map[string]uint32{}
	err := NewCompactU16Decoder(wire).Decode(&got)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

// ---------- Caller-set caps enforced ----------

func TestDecode_MaxSliceLen_Enforced(t *testing.T) {
	// Encode a legal 20-element slice.
	src := make([]uint32, 20)
	for i := range src {
		src[i] = uint32(i)
	}
	wire, err := MarshalBin(&src)
	if err != nil {
		t.Fatal(err)
	}

	// Without cap: decodes fine.
	var ok []uint32
	if err := NewBinDecoder(wire).Decode(&ok); err != nil {
		t.Fatalf("expected successful decode without cap, got %v", err)
	}
	if len(ok) != 20 {
		t.Fatalf("expected 20 elements, got %d", len(ok))
	}

	// With cap=10 < 20: rejected.
	var capped []uint32
	dec := NewBinDecoder(wire).SetMaxSliceLen(10)
	err = dec.Decode(&capped)
	if err == nil {
		t.Fatal("expected error when len > MaxSliceLen")
	}
	if !errors.Is(err, ErrSliceLenTooLarge) {
		t.Fatalf("expected ErrSliceLenTooLarge, got %v", err)
	}

	// With cap=20: exactly fits.
	var exact []uint32
	if err := NewBinDecoder(wire).SetMaxSliceLen(20).Decode(&exact); err != nil {
		t.Fatalf("expected successful decode at cap, got %v", err)
	}
}

func TestDecode_MaxMapLen_Enforced(t *testing.T) {
	src := map[uint32]uint32{}
	for i := uint32(0); i < 20; i++ {
		src[i] = i * 2
	}
	wire, err := MarshalBin(&src)
	if err != nil {
		t.Fatal(err)
	}

	// Without cap: decodes fine.
	ok := map[uint32]uint32{}
	if err := NewBinDecoder(wire).Decode(&ok); err != nil {
		t.Fatalf("expected successful decode without cap, got %v", err)
	}
	if len(ok) != 20 {
		t.Fatalf("expected 20 entries, got %d", len(ok))
	}

	// With cap=5 < 20: rejected.
	capped := map[uint32]uint32{}
	err = NewBinDecoder(wire).SetMaxMapLen(5).Decode(&capped)
	if err == nil {
		t.Fatal("expected error when len > MaxMapLen")
	}
	if !errors.Is(err, ErrMapLenTooLarge) {
		t.Fatalf("expected ErrMapLenTooLarge, got %v", err)
	}
}

// ---------- Backward compat: default = unlimited ----------

func TestDecode_DefaultUnlimited(t *testing.T) {
	dec := NewBinDecoder(nil)
	if dec.MaxSliceLen() != 0 {
		t.Errorf("default MaxSliceLen = %d, want 0", dec.MaxSliceLen())
	}
	if dec.MaxMapLen() != 0 {
		t.Errorf("default MaxMapLen = %d, want 0", dec.MaxMapLen())
	}
}

// Round-trip a non-trivial value through all three encodings with no cap
// set — confirms no false positives in the new bounds logic.
func TestDecode_Bounds_NoFalsePositives(t *testing.T) {
	s := makePerfBenchStruct()

	for _, tc := range parityCases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := tc.marshal(&s)
			if err != nil {
				t.Fatal(err)
			}
			var got perfBenchStruct
			var dec *Decoder
			switch tc.name {
			case "bin":
				dec = NewBinDecoder(wire)
			case "borsh":
				dec = NewBorshDecoder(wire)
			case "compact-u16":
				dec = NewCompactU16Decoder(wire)
			}
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if got.A != s.A || got.B != s.B || got.C != s.C || len(got.E) != len(s.E) {
				t.Fatalf("round-trip mismatch: %+v vs %+v", got, s)
			}
		})
	}
}

// ---------- Pathological length detection ----------

// A u32 length of 0xFFFF_FFFF should not result in a panic or runaway
// allocation. On 64-bit hosts the value is well above Remaining() — the
// decoder surfaces io.ErrUnexpectedEOF. On 32-bit hosts the int cast
// produces -1, which checkSliceLen rejects with ErrSliceLenTooLarge.
func TestDecode_Borsh_PathologicalLength(t *testing.T) {
	wire := u32LE(0xFFFF_FFFF)
	var got []uint8
	err := NewBorshDecoder(wire).Decode(&got)
	if err == nil {
		t.Fatal("expected error for pathological length")
	}
	if !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, ErrSliceLenTooLarge) {
		t.Fatalf("expected io.ErrUnexpectedEOF or ErrSliceLenTooLarge, got %v", err)
	}
}
