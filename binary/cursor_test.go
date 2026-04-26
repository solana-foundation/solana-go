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
	"encoding/binary"
	"testing"
)

// TestCursor_PrimitivesParity confirms each Cursor primitive emits bytes
// identical to the Encoder's equivalent WriteXxx method. If these ever
// diverge the Cursor silently desynchronizes from the reference encoder.
func TestCursor_PrimitivesParity(t *testing.T) {
	cases := []struct {
		name     string
		bufSize  int
		viaCur   func(c *Cursor) *Cursor
		viaEnc   func(e *Encoder) error
		encoding Encoding
	}{
		{
			name: "u8", bufSize: 1, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteU8(0xab) },
			viaEnc: func(e *Encoder) error { return e.WriteUint8(0xab) },
		},
		{
			name: "bool-true", bufSize: 1, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteBool(true) },
			viaEnc: func(e *Encoder) error { return e.WriteBool(true) },
		},
		{
			name: "bool-false", bufSize: 1, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteBool(false) },
			viaEnc: func(e *Encoder) error { return e.WriteBool(false) },
		},
		{
			name: "u16-le", bufSize: 2, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteU16LE(0xbeef) },
			viaEnc: func(e *Encoder) error { return e.WriteUint16(0xbeef, binary.LittleEndian) },
		},
		{
			name: "u16-be", bufSize: 2, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteU16BE(0xbeef) },
			viaEnc: func(e *Encoder) error { return e.WriteUint16(0xbeef, binary.BigEndian) },
		},
		{
			name: "u32-le", bufSize: 4, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteU32LE(0xdeadbeef) },
			viaEnc: func(e *Encoder) error { return e.WriteUint32(0xdeadbeef, binary.LittleEndian) },
		},
		{
			name: "u64-le", bufSize: 8, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteU64LE(0x0123456789abcdef) },
			viaEnc: func(e *Encoder) error { return e.WriteUint64(0x0123456789abcdef, binary.LittleEndian) },
		},
		{
			name: "f32-le", bufSize: 4, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteF32LE(1.5) },
			viaEnc: func(e *Encoder) error { return e.WriteFloat32(1.5, binary.LittleEndian) },
		},
		{
			name: "f64-le", bufSize: 8, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteF64LE(1.5) },
			viaEnc: func(e *Encoder) error { return e.WriteFloat64(1.5, binary.LittleEndian) },
		},
		{
			name: "bytes", bufSize: 5, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteBytes([]byte{1, 2, 3, 4, 5}) },
			viaEnc: func(e *Encoder) error { return e.WriteBytes([]byte{1, 2, 3, 4, 5}, false) },
		},
		{
			name: "uvarint", bufSize: 10, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteUvarint(1_000_000) },
			viaEnc: func(e *Encoder) error { return e.WriteUVarInt(1_000_000) },
		},
		{
			name: "varint", bufSize: 10, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteVarint(-42) },
			viaEnc: func(e *Encoder) error { return e.WriteVarInt(-42) },
		},
		{
			name: "len-bin", bufSize: 10, encoding: EncodingBin,
			viaCur: func(c *Cursor) *Cursor { return c.WriteLenBin(1234) },
			viaEnc: func(e *Encoder) error { return e.WriteLength(1234) },
		},
		{
			name: "len-borsh", bufSize: 4, encoding: EncodingBorsh,
			viaCur: func(c *Cursor) *Cursor { return c.WriteLenBorsh(1234) },
			viaEnc: func(e *Encoder) error { return e.WriteLength(1234) },
		},
		{
			name: "len-compact-u16-1b", bufSize: 3, encoding: EncodingCompactU16,
			viaCur: func(c *Cursor) *Cursor { return c.WriteLenCompactU16(127) },
			viaEnc: func(e *Encoder) error { return e.WriteLength(127) },
		},
		{
			name: "len-compact-u16-2b", bufSize: 3, encoding: EncodingCompactU16,
			viaCur: func(c *Cursor) *Cursor { return c.WriteLenCompactU16(16383) },
			viaEnc: func(e *Encoder) error { return e.WriteLength(16383) },
		},
		{
			name: "len-compact-u16-3b", bufSize: 3, encoding: EncodingCompactU16,
			viaCur: func(c *Cursor) *Cursor { return c.WriteLenCompactU16(0xFFFF) },
			viaEnc: func(e *Encoder) error { return e.WriteLength(0xFFFF) },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Encoder reference.
			encBuf := make([]byte, tc.bufSize)
			enc := NewEncoderIntoWithEncoding(encBuf, tc.encoding)
			if err := tc.viaEnc(enc); err != nil {
				t.Fatalf("encoder write failed: %v", err)
			}
			want := enc.Bytes()

			// Cursor.
			curBuf := make([]byte, tc.bufSize)
			got := tc.viaCur(NewCursor(curBuf)).Written()

			if !bytes.Equal(got, want) {
				t.Fatalf("mismatch\n cursor: %x\nencoder: %x", got, want)
			}
		})
	}
}

// TestCursor_Chaining confirms that chained and imperative call styles
// produce identical output.
func TestCursor_Chaining(t *testing.T) {
	// Chained.
	cBuf := make([]byte, 16)
	chained := NewCursor(cBuf).
		WriteU8(0xaa).
		WriteU16LE(0xbeef).
		WriteU32LE(0xdeadbeef).
		WriteU64LE(0x0123456789abcdef).
		WriteU8(0xbb).
		Written()

	// Imperative.
	iBuf := make([]byte, 16)
	c := NewCursor(iBuf)
	c.WriteU8(0xaa)
	c.WriteU16LE(0xbeef)
	c.WriteU32LE(0xdeadbeef)
	c.WriteU64LE(0x0123456789abcdef)
	c.WriteU8(0xbb)
	imperative := c.Written()

	if !bytes.Equal(chained, imperative) {
		t.Fatalf("chained vs imperative mismatch:\n chained:    %x\n imperative: %x", chained, imperative)
	}
	if len(chained) != 16 {
		t.Fatalf("expected 16 bytes written, got %d", len(chained))
	}
}

// TestCursor_BackPatch exercises the pattern where a fixed-size header
// is reserved up front (Skip), the body is written, and the header
// fields are filled in later using a second cursor at offset 0 (or
// SetPos).
func TestCursor_BackPatch(t *testing.T) {
	buf := make([]byte, 32)
	c := NewCursor(buf)

	// Reserve 4 bytes for a length prefix.
	c.Skip(4)
	// Write body.
	body := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	c.WriteBytes(body)
	bodyEnd := c.Pos()

	// Back-patch the header: go to offset 0, write length as u32-LE.
	c.SetPos(0).WriteU32LE(uint32(len(body)))
	// Restore pos so Written() reflects the whole payload.
	c.SetPos(bodyEnd)

	got := c.Written()
	want := []byte{
		8, 0, 0, 0, // u32-le length
		1, 2, 3, 4, 5, 6, 7, 8, // body
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("back-patch mismatch\n got: %x\nwant: %x", got, want)
	}
}

// TestCursor_BoundsPanic confirms that writing past Cap() triggers a
// Go slice-bounds panic (not a silent overwrite or a grow).
func TestCursor_BoundsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on out-of-bounds write")
		}
	}()
	c := NewCursor(make([]byte, 2))
	c.WriteU8(1).WriteU8(2).WriteU8(3) // third write overruns.
}

// TestCursor_ZeroAllocs confirms that a typical chained sequence of
// primitive writes does not allocate. The Cursor and its buffer are
// stack-ish (buf escapes because make() allocates, but the writes
// themselves should not trigger any extra allocations).
func TestCursor_ZeroAllocs(t *testing.T) {
	if testing.Short() {
		t.Skip("allocation test disabled under -short")
	}
	buf := make([]byte, 64)
	c := NewCursor(buf)

	allocs := testing.AllocsPerRun(200, func() {
		c.Reset().
			WriteU8(1).
			WriteU16LE(2).
			WriteU32LE(3).
			WriteU64LE(4).
			WriteBytes([]byte{5, 6, 7, 8}).
			WriteLenCompactU16(127)
	})
	if allocs != 0 {
		t.Fatalf("cursor allocs = %g, want 0", allocs)
	}
}

// TestCursor_ResetTo confirms the cursor can be rebound to a new buffer
// without reallocating the Cursor struct.
func TestCursor_ResetTo(t *testing.T) {
	c := NewCursor(nil)
	buf1 := make([]byte, 4)
	buf2 := make([]byte, 4)
	c.ResetTo(buf1).WriteU32LE(0x11223344)
	c.ResetTo(buf2).WriteU32LE(0xaabbccdd)

	wantBuf1 := []byte{0x44, 0x33, 0x22, 0x11}
	wantBuf2 := []byte{0xdd, 0xcc, 0xbb, 0xaa}
	if !bytes.Equal(buf1, wantBuf1) {
		t.Errorf("buf1 %x, want %x", buf1, wantBuf1)
	}
	if !bytes.Equal(buf2, wantBuf2) {
		t.Errorf("buf2 %x, want %x", buf2, wantBuf2)
	}
}

// TestCursor_WriteZero confirms WriteZero clears the specified range
// even if it previously held non-zero bytes.
func TestCursor_WriteZero(t *testing.T) {
	buf := []byte{0xff, 0xff, 0xff, 0xff, 0xff}
	c := NewCursor(buf)
	c.WriteZero(3)
	c.WriteU8(0xaa)
	want := []byte{0, 0, 0, 0xaa, 0xff}
	if !bytes.Equal(buf, want) {
		t.Fatalf("got %x, want %x", buf, want)
	}
}
