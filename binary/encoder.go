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

package bin

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"unsafe"

	"go.uber.org/zap"
)

type Encoder struct {
	count int

	// currentFieldOpt is held by value (not pointer) so it doesn't escape.
	// Same role as Decoder.currentFieldOpt — gives nested types like Uint128
	// access to the active byte order.
	currentFieldOpt option
	encoding        Encoding

	// output is the destination io.Writer. May be nil when the Encoder is
	// running in buffered mode (see NewBinEncoderBuf etc.) — in that case all
	// writes accumulate in `buf` and the caller retrieves them via Bytes()
	// or WriteTo.
	output io.Writer
	buf    []byte

	// fixedBuf marks the encoder as bound to a caller-supplied buffer (see
	// NewBinEncoderInto etc.). In this mode writes append to buf without
	// ever reallocating: toWriter bounds-checks every write against
	// cap(buf) and returns io.ErrShortBuffer when the caller's buffer is
	// too small. This is the zero-allocation encode path — the caller
	// pre-sizes dst (typically from a SerializedSize computation), calls
	// EncodeInto, and the encoded payload lives in dst with no copy.
	fixedBuf bool

	// scratch is a per-Encoder staging buffer reused across primitive writes
	// so WriteUint*/WriteFloat*/WriteVarInt/WriteCompactU16/... don't allocate.
	// 16 bytes is enough for any fixed-width primitive (Uint128) and for a
	// Uvarint/Varint header (max 10 bytes) and for a CompactU16 (max 3 bytes).
	// Safe to reuse: io.Writer.Write must not retain p after return.
	scratch [16]byte

	// skipMarshalerCheck tells encodeBin/encodeBorsh/encodeCompactU16 to
	// skip the per-call asBinaryMarshaler() type assertion. encodeStructBin
	// (and friends) sets this to true before encoding a field whose typePlan
	// has already proven that neither the value nor the pointer type
	// implements BinaryMarshaler. The flag is propagated through Ptr.Elem
	// recursion (because *T not implementing means T doesn't either), and
	// reset around array/slice element loops where elements are independent
	// types. Without this flag, the rv.Interface() boxing dominates encode
	// allocations for non-marshaler types like solana.PublicKey.
	skipMarshalerCheck bool
}

func (enc *Encoder) IsBorsh() bool {
	return enc.encoding.IsBorsh()
}

func (enc *Encoder) IsBin() bool {
	return enc.encoding.IsBin()
}

func (enc *Encoder) IsCompactU16() bool {
	return enc.encoding.IsCompactU16()
}

func NewEncoderWithEncoding(writer io.Writer, enc Encoding) *Encoder {
	if !isValidEncoding(enc) {
		panic(fmt.Sprintf("provided encoding is not valid: %s", enc))
	}
	return &Encoder{
		output:   writer,
		count:    0,
		encoding: enc,
	}
}

func NewBinEncoder(writer io.Writer) *Encoder {
	return NewEncoderWithEncoding(writer, EncodingBin)
}

func NewBorshEncoder(writer io.Writer) *Encoder {
	return NewEncoderWithEncoding(writer, EncodingBorsh)
}

func NewCompactU16Encoder(writer io.Writer) *Encoder {
	return NewEncoderWithEncoding(writer, EncodingCompactU16)
}

// NewBufferedEncoder returns an Encoder that writes into an internal []byte
// buffer instead of an io.Writer. Use Bytes() to retrieve the encoded payload
// and Reset()/Bytes() to reuse the encoder across multiple messages.
//
// This is the lowest-overhead encode mode: every primitive write becomes an
// `append(e.buf, ...)` with no interface dispatch and no per-call allocation.
func NewBufferedEncoder(enc Encoding) *Encoder {
	if !isValidEncoding(enc) {
		panic(fmt.Sprintf("provided encoding is not valid: %s", enc))
	}
	return &Encoder{encoding: enc}
}

func NewBinEncoderBuf() *Encoder        { return NewBufferedEncoder(EncodingBin) }
func NewBorshEncoderBuf() *Encoder      { return NewBufferedEncoder(EncodingBorsh) }
func NewCompactU16EncoderBuf() *Encoder { return NewBufferedEncoder(EncodingCompactU16) }

// NewEncoderIntoWithEncoding returns an Encoder that writes directly into
// the caller-supplied dst buffer. The encoded payload will occupy
// dst[:n] where n is the Written() count after Encode returns. If dst is
// too small to hold the payload, writes return io.ErrShortBuffer — dst
// is never reallocated.
//
// Typical use combines this with a pre-computed encoded size:
//
//	buf := make([]byte, EncodedSizeOf(tx))
//	n, err := MarshalBinInto(tx, buf)
//	wire := buf[:n]
//
// Passing dst with non-zero length is allowed — the encoder truncates it
// to dst[:0] internally and uses its capacity. Pass a nil dst to force
// every write to return io.ErrShortBuffer (useful for probing size when
// BinByteCount is not available).
func NewEncoderIntoWithEncoding(dst []byte, enc Encoding) *Encoder {
	if !isValidEncoding(enc) {
		panic(fmt.Sprintf("provided encoding is not valid: %s", enc))
	}
	return &Encoder{
		encoding: enc,
		buf:      dst[:0:cap(dst)],
		fixedBuf: true,
	}
}

func NewBinEncoderInto(dst []byte) *Encoder {
	return NewEncoderIntoWithEncoding(dst, EncodingBin)
}

func NewBorshEncoderInto(dst []byte) *Encoder {
	return NewEncoderIntoWithEncoding(dst, EncodingBorsh)
}

func NewCompactU16EncoderInto(dst []byte) *Encoder {
	return NewEncoderIntoWithEncoding(dst, EncodingCompactU16)
}

// Bytes returns the encoded payload accumulated in buffered mode. The slice
// aliases the encoder's internal buffer; copy it if you need to retain it
// across a Reset() / further writes.
func (e *Encoder) Bytes() []byte {
	return e.buf
}

// Reset clears the encoder's internal state (count, buffer, current option)
// so it can be reused for another message. The output writer is preserved,
// and fixed-buffer mode (see NewEncoderIntoWithEncoding) is preserved so
// repeat encodes into the same dst slice keep the bounds check active.
func (e *Encoder) Reset() {
	e.count = 0
	e.buf = e.buf[:0]
	e.currentFieldOpt = option{}
	e.skipMarshalerCheck = false
}

// ResetInto re-targets the encoder at a new caller-supplied dst buffer and
// clears per-message state. After ResetInto the encoder is in fixed-buffer
// mode: writes go into dst up to cap(dst) and return io.ErrShortBuffer past
// that. Useful for reusing a pooled or long-lived *Encoder across many
// messages without a per-message Encoder allocation.
func (e *Encoder) ResetInto(dst []byte) {
	e.count = 0
	e.buf = dst[:0:cap(dst)]
	e.fixedBuf = true
	e.output = nil
	e.currentFieldOpt = option{}
	e.skipMarshalerCheck = false
}

// Grow ensures the internal buffer has at least n free bytes available.
// Useful in buffered mode to amortize append-driven growth when the encoded
// size is known in advance. No-op in fixed-buffer mode: the caller already
// sized dst and growing would defeat the zero-allocation guarantee.
func (e *Encoder) Grow(n int) {
	if e.fixedBuf {
		return
	}
	if cap(e.buf)-len(e.buf) >= n {
		return
	}
	nb := make([]byte, len(e.buf), len(e.buf)+n)
	copy(nb, e.buf)
	e.buf = nb
}

func (e *Encoder) Encode(v interface{}) (err error) {
	switch e.encoding {
	case EncodingBin:
		return e.encodeBin(reflect.ValueOf(v), defaultOption)
	case EncodingBorsh:
		return e.encodeBorsh(reflect.ValueOf(v), defaultOption)
	case EncodingCompactU16:
		return e.encodeCompactU16(reflect.ValueOf(v), defaultOption)
	default:
		panic(fmt.Errorf("encoding not implemented: %s", e.encoding))
	}
}

func (e *Encoder) toWriter(bytes []byte) (err error) {
	if e.output == nil {
		if e.fixedBuf && len(e.buf)+len(bytes) > cap(e.buf) {
			// Don't advance count — the write failed, so Written() reports
			// the successful prefix, which is useful for debugging
			// short-buffer failures.
			return io.ErrShortBuffer
		}
		e.buf = append(e.buf, bytes...)
		e.count += len(bytes)
		if traceEnabled {
			zlog.Debug("	> encode: appending", zap.Stringer("hex", HexBytes(bytes)), zap.Int("pos", e.count))
		}
		return nil
	}
	e.count += len(bytes)
	if traceEnabled {
		zlog.Debug("	> encode: appending", zap.Stringer("hex", HexBytes(bytes)), zap.Int("pos", e.count))
	}
	_, err = e.output.Write(bytes)
	return
}

// Written returns the count of bytes written.
func (e *Encoder) Written() int {
	return e.count
}

func (e *Encoder) WriteBytes(b []byte, writeLength bool) error {
	if traceEnabled {
		zlog.Debug("encode: write byte array", zap.Int("len", len(b)))
	}
	if writeLength {
		if err := e.WriteLength(len(b)); err != nil {
			return err
		}
	}
	if len(b) == 0 {
		return nil
	}
	return e.toWriter(b)
}

func (e *Encoder) Write(b []byte) (n int, err error) {
	// Route through toWriter so buffered Encoders (output == nil) append to
	// e.buf instead of nil-derefing. Matches WriteBytes semantics.
	if err := e.toWriter(b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (e *Encoder) WriteLength(length int) error {
	if traceEnabled {
		zlog.Debug("encode: write length", zap.Int("len", length))
	}
	switch e.encoding {
	case EncodingBin:
		if err := e.WriteUVarInt(length); err != nil {
			return err
		}
	case EncodingBorsh:
		if err := e.WriteUint32(uint32(length), LE); err != nil {
			return err
		}
	case EncodingCompactU16:
		n, err := PutCompactU16Length(e.scratch[:3], length)
		if err != nil {
			return err
		}
		if err := e.toWriter(e.scratch[:n]); err != nil {
			return err
		}
	default:
		panic(fmt.Errorf("encoding not implemented: %s", e.encoding))
	}
	return nil
}

func (e *Encoder) WriteUVarInt(v int) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write uvarint", zap.Int("val", v))
	}
	l := binary.PutUvarint(e.scratch[:], uint64(v))
	return e.toWriter(e.scratch[:l])
}

func (e *Encoder) WriteVarInt(v int) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write varint", zap.Int("val", v))
	}
	l := binary.PutVarint(e.scratch[:], int64(v))
	return e.toWriter(e.scratch[:l])
}

func (e *Encoder) WriteByte(b byte) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write byte", zap.Uint8("val", b))
	}
	e.scratch[0] = b
	return e.toWriter(e.scratch[:1])
}

func (e *Encoder) WriteOption(b bool) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write option", zap.Bool("val", b))
	}
	return e.WriteBool(b)
}

func (e *Encoder) WriteCOption(b bool) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write c-option", zap.Bool("val", b))
	}
	var num uint32
	if b {
		num = 1
	}
	return e.WriteUint32(num, LE)
}

func (e *Encoder) WriteBool(b bool) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write bool", zap.Bool("val", b))
	}
	var out byte
	if b {
		out = 1
	}
	return e.WriteByte(out)
}

func (e *Encoder) WriteUint8(i uint8) (err error) {
	return e.WriteByte(i)
}

func (e *Encoder) WriteInt8(i int8) (err error) {
	return e.WriteByte(uint8(i))
}

func (e *Encoder) WriteUint16(i uint16, order binary.ByteOrder) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write uint16", zap.Uint16("val", i))
	}
	order.PutUint16(e.scratch[:2], i)
	return e.toWriter(e.scratch[:2])
}

func (e *Encoder) WriteInt16(i int16, order binary.ByteOrder) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write int16", zap.Int16("val", i))
	}
	return e.WriteUint16(uint16(i), order)
}

func (e *Encoder) WriteUint32(i uint32, order binary.ByteOrder) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write uint32", zap.Uint32("val", i))
	}
	order.PutUint32(e.scratch[:4], i)
	return e.toWriter(e.scratch[:4])
}

func (e *Encoder) WriteInt32(i int32, order binary.ByteOrder) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write int32", zap.Int32("val", i))
	}
	return e.WriteUint32(uint32(i), order)
}

func (e *Encoder) WriteUint64(i uint64, order binary.ByteOrder) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write uint64", zap.Uint64("val", i))
	}
	order.PutUint64(e.scratch[:8], i)
	return e.toWriter(e.scratch[:8])
}

func (e *Encoder) WriteInt64(i int64, order binary.ByteOrder) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write int64", zap.Int64("val", i))
	}
	return e.WriteUint64(uint64(i), order)
}

func (e *Encoder) WriteUint128(i Uint128, order binary.ByteOrder) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write uint128", zap.Stringer("hex", i), zap.Uint64("lo", i.Lo), zap.Uint64("hi", i.Hi))
	}
	buf := e.scratch[:16]
	switch order {
	case binary.LittleEndian:
		order.PutUint64(buf[:8], i.Lo)
		order.PutUint64(buf[8:], i.Hi)
	case binary.BigEndian:
		order.PutUint64(buf[:8], i.Hi)
		order.PutUint64(buf[8:], i.Lo)
	default:
		return fmt.Errorf("invalid byte order: %v", order)
	}
	return e.toWriter(buf)
}

func (e *Encoder) WriteInt128(i Int128, order binary.ByteOrder) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write int128", zap.Stringer("hex", i), zap.Uint64("lo", i.Lo), zap.Uint64("hi", i.Hi))
	}
	buf := e.scratch[:16]
	switch order {
	case binary.LittleEndian:
		order.PutUint64(buf[:8], i.Lo)
		order.PutUint64(buf[8:], i.Hi)
	case binary.BigEndian:
		order.PutUint64(buf[:8], i.Hi)
		order.PutUint64(buf[8:], i.Lo)
	default:
		return fmt.Errorf("invalid byte order: %v", order)
	}
	return e.toWriter(buf)
}

func (e *Encoder) WriteFloat32(f float32, order binary.ByteOrder) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write float32", zap.Float32("val", f))
	}

	if e.IsBorsh() {
		if math.IsNaN(float64(f)) {
			return errors.New("NaN float value")
		}
	}

	order.PutUint32(e.scratch[:4], math.Float32bits(f))
	return e.toWriter(e.scratch[:4])
}

func (e *Encoder) WriteFloat64(f float64, order binary.ByteOrder) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write float64", zap.Float64("val", f))
	}

	if e.IsBorsh() {
		if math.IsNaN(float64(f)) {
			return errors.New("NaN float value")
		}
	}
	order.PutUint64(e.scratch[:8], math.Float64bits(f))
	return e.toWriter(e.scratch[:8])
}

func (e *Encoder) WriteString(s string) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write string", zap.String("val", s))
	}
	return e.WriteBytes([]byte(s), true)
}

func (e *Encoder) WriteRustString(s string) (err error) {
	err = e.WriteUint64(uint64(len(s)), binary.LittleEndian)
	if err != nil {
		return err
	}
	if traceEnabled {
		zlog.Debug("encode: write Rust string", zap.String("val", s))
	}
	return e.WriteBytes([]byte(s), false)
}

func (e *Encoder) WriteCompactU16(ln int) (err error) {
	if traceEnabled {
		zlog.Debug("encode: write compact-u16", zap.Int("val", ln))
	}
	n, err := PutCompactU16Length(e.scratch[:3], ln)
	if err != nil {
		return err
	}
	return e.toWriter(e.scratch[:n])
}

func (e *Encoder) WriteCompactU16Length(ln int) (err error) {
	return e.WriteCompactU16(ln)
}

// writePoDSliceBytes is the encoder analog of readPoDSliceBytes. When the
// destination is addressable AND the host and wire byte orders match, it
// sends the destination's backing memory straight through toWriter via a
// single byte-view — no make, no per-element reflect.Index/Uint loop.
//
// When rv is not addressable (caller passed a struct by value), we fall back
// to a per-element reflect.Index path that uses an intermediate scratch
// slice. This is the same behavior as the pre-PoD-fast-path implementation.
//
// elemSize must be 1, 2, 4, or 8.
func writePoDSliceBytes(e *Encoder, rv reflect.Value, l, elemSize int, order binary.ByteOrder) error {
	if l == 0 {
		return nil
	}
	need := l * elemSize

	if rv.CanAddr() && rv.Len() > 0 {
		base := unsafe.Pointer(rv.Index(0).UnsafeAddr())

		if elemSize == 1 || (isHostLittleEndian && order == binary.LittleEndian) {
			// Single memcpy from the slice/array's backing storage. WriteBytes
			// (or the buffered append in toWriter) will copy the bytes onward.
			return e.toWriter(unsafe.Slice((*byte)(base), need))
		}

		// Byte-order mismatch: serialize element-by-element into a temporary
		// scratch slice. We still avoid going through reflect.Index for each
		// element by reading directly from the backing memory.
		tmp := make([]byte, need)
		switch elemSize {
		case 2:
			for i := range l {
				order.PutUint16(tmp[i*2:], *(*uint16)(unsafe.Add(base, i*2)))
			}
		case 4:
			for i := range l {
				order.PutUint32(tmp[i*4:], *(*uint32)(unsafe.Add(base, i*4)))
			}
		case 8:
			for i := range l {
				order.PutUint64(tmp[i*8:], *(*uint64)(unsafe.Add(base, i*8)))
			}
		}
		return e.toWriter(tmp)
	}

	// Fallback: rv is not addressable. Walk via reflect.Index — slower but
	// correct for callers that pass arrays by value.
	tmp := make([]byte, need)
	switch elemSize {
	case 1:
		for i := range l {
			tmp[i] = byte(rv.Index(i).Uint())
		}
	case 2:
		for i := range l {
			order.PutUint16(tmp[i*2:], uint16(rv.Index(i).Uint()))
		}
	case 4:
		for i := range l {
			order.PutUint32(tmp[i*4:], uint32(rv.Index(i).Uint()))
		}
	case 8:
		for i := range l {
			order.PutUint64(tmp[i*8:], rv.Index(i).Uint())
		}
	}
	return e.toWriter(tmp)
}

func reflect_writeArrayOfBytes(e *Encoder, l int, rv reflect.Value) error {
	return writePoDSliceBytes(e, rv, l, 1, binary.LittleEndian)
}

func reflect_writeArrayOfUint16(e *Encoder, l int, rv reflect.Value, order binary.ByteOrder) error {
	return writePoDSliceBytes(e, rv, l, 2, order)
}

func reflect_writeArrayOfUint32(e *Encoder, l int, rv reflect.Value, order binary.ByteOrder) error {
	return writePoDSliceBytes(e, rv, l, 4, order)
}

func reflect_writeArrayOfUint64(e *Encoder, l int, rv reflect.Value, order binary.ByteOrder) error {
	return writePoDSliceBytes(e, rv, l, 8, order)
}

// reflect_writeArrayOfUint_ is used for writing arrays/slices of uints of any size.
func reflect_writeArrayOfUint_(e *Encoder, l int, k reflect.Kind, rv reflect.Value, order binary.ByteOrder) error {
	switch k {
	// case reflect.Uint:
	// 	// switch on system architecture (32 or 64 bit)
	// 	if unsafe.Sizeof(uintptr(0)) == 4 {
	// 		return reflect_writeArrayOfUint32(e, l, rv, order)
	// 	}
	// 	return reflect_writeArrayOfUint64(e, l, rv, order)
	case reflect.Uint8:
		return reflect_writeArrayOfBytes(e, l, rv)
	case reflect.Uint16:
		return reflect_writeArrayOfUint16(e, l, rv, order)
	case reflect.Uint32:
		return reflect_writeArrayOfUint32(e, l, rv, order)
	case reflect.Uint64:
		return reflect_writeArrayOfUint64(e, l, rv, order)
	default:
		return fmt.Errorf("unsupported kind: %v", k)
	}
}
