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
	"math"
)

// Cursor is a zero-overhead write cursor into a caller-provided byte
// slice. Every primitive write is a single memory poke followed by
// position advance — no error return, no scratch buffer, no encoding
// dispatch. The caller pre-sizes the destination slice; writes past the
// end cause a standard Go slice-bounds-out-of-range panic (no cushion).
//
// Cursor is the fastest encode path in this package. For safety-first
// encoding with error returns and grow-on-overflow, use Encoder with
// NewBinEncoderInto / NewBinEncoderBuf instead.
//
// Methods return the receiver so calls can chain. The primitive
// integer methods (WriteU*, WriteI*, WriteF*) are simple enough to
// inline; chained fluent code compiles to the same machine code as
// imperative `c.WriteU8(1); c.WriteU8(2)` statements.
//
// Cursor is not safe for concurrent use.
type Cursor struct {
	buf []byte
	pos int
}

// NewCursor returns a Cursor positioned at offset 0 of dst. Writes will
// advance through dst's backing array; the slice itself is never
// reallocated.
func NewCursor(dst []byte) *Cursor {
	return &Cursor{buf: dst}
}

// NewCursorAt returns a Cursor starting at the specified offset into
// dst. Useful when back-patching a header after knowing the body size:
// allocate, Skip past the header region, write the body, then open a
// second cursor at offset 0 to fill in the header fields.
func NewCursorAt(dst []byte, offset int) *Cursor {
	return &Cursor{buf: dst, pos: offset}
}

// --- State ---

// Len returns the number of bytes written so far.
func (c *Cursor) Len() int { return c.pos }

// Cap returns the cursor's underlying buffer capacity. Writes past
// Cap() panic.
func (c *Cursor) Cap() int { return len(c.buf) }

// Remaining returns the number of bytes available for writing before
// the next poke would panic.
func (c *Cursor) Remaining() int { return len(c.buf) - c.pos }

// Pos returns the current write offset.
func (c *Cursor) Pos() int { return c.pos }

// SetPos repositions the cursor at offset n. No bounds check — pass a
// value in [0, Cap()]. Useful for back-patching after recording a
// position.
func (c *Cursor) SetPos(n int) *Cursor {
	c.pos = n
	return c
}

// Reset repositions the cursor at offset 0. Buffer contents are
// unchanged; subsequent writes overwrite them.
func (c *Cursor) Reset() *Cursor {
	c.pos = 0
	return c
}

// ResetTo repositions the cursor at offset 0 and rebinds it to dst.
// Useful for reusing one Cursor across many messages without
// allocating.
func (c *Cursor) ResetTo(dst []byte) *Cursor {
	c.buf = dst
	c.pos = 0
	return c
}

// Written returns a subslice of the underlying buffer covering the
// bytes written so far (buf[:pos]). Aliases the cursor's backing array.
// Copy the result if you need to retain it across further writes or
// Reset.
func (c *Cursor) Written() []byte { return c.buf[:c.pos] }

// Buffer returns the cursor's full underlying buffer. Aliases the
// backing array.
func (c *Cursor) Buffer() []byte { return c.buf }

// --- Single-byte primitives ---

// WriteU8 writes a uint8 and advances one byte.
func (c *Cursor) WriteU8(v uint8) *Cursor {
	c.buf[c.pos] = v
	c.pos++
	return c
}

// WriteI8 writes an int8 (reinterpreted as uint8) and advances one
// byte.
func (c *Cursor) WriteI8(v int8) *Cursor { return c.WriteU8(uint8(v)) }

// WriteBool writes 0x01 for true, 0x00 for false.
func (c *Cursor) WriteBool(v bool) *Cursor {
	if v {
		return c.WriteU8(1)
	}
	return c.WriteU8(0)
}

// --- Fixed-width integers: little-endian ---

func (c *Cursor) WriteU16LE(v uint16) *Cursor {
	binary.LittleEndian.PutUint16(c.buf[c.pos:], v)
	c.pos += 2
	return c
}

func (c *Cursor) WriteU32LE(v uint32) *Cursor {
	binary.LittleEndian.PutUint32(c.buf[c.pos:], v)
	c.pos += 4
	return c
}

func (c *Cursor) WriteU64LE(v uint64) *Cursor {
	binary.LittleEndian.PutUint64(c.buf[c.pos:], v)
	c.pos += 8
	return c
}

func (c *Cursor) WriteI16LE(v int16) *Cursor { return c.WriteU16LE(uint16(v)) }
func (c *Cursor) WriteI32LE(v int32) *Cursor { return c.WriteU32LE(uint32(v)) }
func (c *Cursor) WriteI64LE(v int64) *Cursor { return c.WriteU64LE(uint64(v)) }

// --- Fixed-width integers: big-endian ---

func (c *Cursor) WriteU16BE(v uint16) *Cursor {
	binary.BigEndian.PutUint16(c.buf[c.pos:], v)
	c.pos += 2
	return c
}

func (c *Cursor) WriteU32BE(v uint32) *Cursor {
	binary.BigEndian.PutUint32(c.buf[c.pos:], v)
	c.pos += 4
	return c
}

func (c *Cursor) WriteU64BE(v uint64) *Cursor {
	binary.BigEndian.PutUint64(c.buf[c.pos:], v)
	c.pos += 8
	return c
}

func (c *Cursor) WriteI16BE(v int16) *Cursor { return c.WriteU16BE(uint16(v)) }
func (c *Cursor) WriteI32BE(v int32) *Cursor { return c.WriteU32BE(uint32(v)) }
func (c *Cursor) WriteI64BE(v int64) *Cursor { return c.WriteU64BE(uint64(v)) }

// --- Floats ---

func (c *Cursor) WriteF32LE(v float32) *Cursor {
	binary.LittleEndian.PutUint32(c.buf[c.pos:], math.Float32bits(v))
	c.pos += 4
	return c
}

func (c *Cursor) WriteF64LE(v float64) *Cursor {
	binary.LittleEndian.PutUint64(c.buf[c.pos:], math.Float64bits(v))
	c.pos += 8
	return c
}

func (c *Cursor) WriteF32BE(v float32) *Cursor {
	binary.BigEndian.PutUint32(c.buf[c.pos:], math.Float32bits(v))
	c.pos += 4
	return c
}

func (c *Cursor) WriteF64BE(v float64) *Cursor {
	binary.BigEndian.PutUint64(c.buf[c.pos:], math.Float64bits(v))
	c.pos += 8
	return c
}

// --- Byte sequences ---

// WriteBytes copies src into the cursor buffer and advances len(src)
// bytes. If src does not fit in Remaining() this panics with the
// standard "index out of range" message from the underlying slice op.
func (c *Cursor) WriteBytes(src []byte) *Cursor {
	n := copy(c.buf[c.pos:c.pos+len(src)], src)
	c.pos += n
	return c
}

// WriteZero writes n zero bytes and advances n positions.
func (c *Cursor) WriteZero(n int) *Cursor {
	end := c.pos + n
	clear(c.buf[c.pos:end])
	c.pos = end
	return c
}

// Skip advances n positions without writing. The skipped bytes keep
// whatever values the underlying buffer already held — callers should
// overwrite them later or zero them via WriteZero if they need the
// payload clean.
func (c *Cursor) Skip(n int) *Cursor {
	c.pos += n
	return c
}

// --- Length-prefix helpers ---
//
// Three variants cover the encoding schemes this package supports:
// uvarint (bin), u32 little-endian (borsh), and Solana's compact-u16.
// They all panic rather than returning errors so they remain
// chainable; use the Encoder for error-returning variants.

// WriteLenBin writes a uvarint-encoded length (1–10 bytes). This matches
// Encoder.WriteLength in EncodingBin mode.
func (c *Cursor) WriteLenBin(l int) *Cursor {
	n := binary.PutUvarint(c.buf[c.pos:], uint64(l))
	c.pos += n
	return c
}

// WriteLenBorsh writes a u32 little-endian length (4 bytes). Matches
// Encoder.WriteLength in EncodingBorsh mode.
func (c *Cursor) WriteLenBorsh(l int) *Cursor { return c.WriteU32LE(uint32(l)) }

// WriteLenCompactU16 writes Solana's compact-u16 length encoding (1–3
// bytes). Panics if l > 0xFFFF.
func (c *Cursor) WriteLenCompactU16(l int) *Cursor {
	n, err := PutCompactU16Length(c.buf[c.pos:c.pos+3], l)
	if err != nil {
		panic(err)
	}
	c.pos += n
	return c
}

// --- UVarint/Varint for standalone values (not just lengths) ---

// WriteUvarint writes v as a uvarint (1–10 bytes).
func (c *Cursor) WriteUvarint(v uint64) *Cursor {
	n := binary.PutUvarint(c.buf[c.pos:], v)
	c.pos += n
	return c
}

// WriteVarint writes v as a zigzag-varint (1–10 bytes).
func (c *Cursor) WriteVarint(v int64) *Cursor {
	n := binary.PutVarint(c.buf[c.pos:], v)
	c.pos += n
	return c
}
