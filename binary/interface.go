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
	"fmt"
	"sync"
)

// marshalBufInitialCap is the initial capacity allocated for a pooled
// Encoder's internal buffer. Chosen to cover the typical Solana
// instruction payload size (~256 bytes fits the vast majority of
// variant-wrapped instructions and most small state structs) without
// requiring a grow step on the hot path. Larger payloads still grow
// via append.
const marshalBufInitialCap = 256

// Per-encoding Encoder pools. Pooled encoders are always in buffered
// mode (output == nil, writes append to e.buf). Marshal helpers Get a
// pooled encoder, encode into it, copy the result out (the pooled
// buffer can be reused by the next caller), Reset, and Put back.
var (
	binEncoderPool = sync.Pool{
		New: func() any {
			return &Encoder{
				encoding: EncodingBin,
				buf:      make([]byte, 0, marshalBufInitialCap),
			}
		},
	}
	borshEncoderPool = sync.Pool{
		New: func() any {
			return &Encoder{
				encoding: EncodingBorsh,
				buf:      make([]byte, 0, marshalBufInitialCap),
			}
		},
	}
	compactU16EncoderPool = sync.Pool{
		New: func() any {
			return &Encoder{
				encoding: EncodingCompactU16,
				buf:      make([]byte, 0, marshalBufInitialCap),
			}
		},
	}

	// Per-encoding Decoder pools. Pooled decoders have their data/pos
	// reset between uses via Decoder.Reset. The encoding field is set
	// at pool-New time and preserved by Reset.
	binDecoderPool = sync.Pool{
		New: func() any {
			return &Decoder{encoding: EncodingBin}
		},
	}
	borshDecoderPool = sync.Pool{
		New: func() any {
			return &Decoder{encoding: EncodingBorsh}
		},
	}
	compactU16DecoderPool = sync.Pool{
		New: func() any {
			return &Decoder{encoding: EncodingCompactU16}
		},
	}
)

// pooledMarshal runs enc.Encode(v) on a pooled encoder, copies the
// resulting bytes out so the pooled buffer can be reused safely, and
// returns the encoder to the pool. It is the shared implementation
// behind MarshalBin / MarshalBorsh / MarshalCompactU16.
func pooledMarshal(pool *sync.Pool, v any) ([]byte, error) {
	enc := pool.Get().(*Encoder)
	err := enc.Encode(v)
	if err != nil {
		enc.Reset()
		pool.Put(enc)
		return nil, err
	}
	// Copy the bytes out: the pooled encoder's underlying slice will
	// be reused by future callers, so returning enc.buf directly would
	// let the next Marshal call silently stomp on the caller's result.
	out := make([]byte, len(enc.buf))
	copy(out, enc.buf)
	enc.Reset()
	pool.Put(enc)
	return out, nil
}

// pooledMarshalInto runs enc.Encode(v) on a pooled encoder whose internal
// buffer is temporarily swapped for the caller's dst. On return the pool's
// original buffer is restored so the encoder stays pool-safe. If the
// caller's dst is too small, the error surfaces as io.ErrShortBuffer and
// the partial write in dst[:written] is still valid (same convention as
// io.Writer short writes).
func pooledMarshalInto(pool *sync.Pool, v any, dst []byte) (int, error) {
	enc := pool.Get().(*Encoder)
	savedBuf := enc.buf
	enc.buf = dst[:0:cap(dst)]
	enc.fixedBuf = true
	err := enc.Encode(v)
	n := len(enc.buf)
	enc.buf = savedBuf
	enc.fixedBuf = false
	enc.Reset()
	pool.Put(enc)
	return n, err
}

// pooledUnmarshal runs dec.Decode(v) on a pooled decoder over the
// provided bytes, then returns the decoder to the pool. Shared
// implementation behind UnmarshalBin / UnmarshalBorsh /
// UnmarshalCompactU16.
func pooledUnmarshal(pool *sync.Pool, v any, b []byte) error {
	dec := pool.Get().(*Decoder)
	dec.Reset(b)
	err := dec.Decode(v)
	// Clear the data reference before returning to pool so we don't
	// pin the caller's input buffer in the pool.
	dec.Reset(nil)
	pool.Put(dec)
	return err
}

type BinaryMarshaler interface {
	MarshalWithEncoder(encoder *Encoder) error
}

type BinaryUnmarshaler interface {
	UnmarshalWithDecoder(decoder *Decoder) error
}

type EncoderDecoder interface {
	BinaryMarshaler
	BinaryUnmarshaler
}

func MarshalBin(v any) ([]byte, error) {
	return pooledMarshal(&binEncoderPool, v)
}

func MarshalBorsh(v any) ([]byte, error) {
	return pooledMarshal(&borshEncoderPool, v)
}

func MarshalCompactU16(v any) ([]byte, error) {
	return pooledMarshal(&compactU16EncoderPool, v)
}

// MarshalBinInto encodes v into the caller-supplied dst buffer using the
// bincode-like "bin" encoding and returns the number of bytes written. If
// dst is too small to hold the payload, the returned error is
// io.ErrShortBuffer; dst[:n] contains the successful prefix.
//
// This is the zero-allocation encode path: dst is never reallocated and
// no intermediate buffer is produced. Pre-size dst via BinByteCount when
// the size is not statically known:
//
//	size, _ := BinByteCount(v)
//	buf := make([]byte, size)
//	n, err := MarshalBinInto(v, buf)
func MarshalBinInto(v any, dst []byte) (int, error) {
	return pooledMarshalInto(&binEncoderPool, v, dst)
}

// MarshalBorshInto is the Borsh-encoding counterpart to MarshalBinInto.
func MarshalBorshInto(v any, dst []byte) (int, error) {
	return pooledMarshalInto(&borshEncoderPool, v, dst)
}

// MarshalCompactU16Into is the CompactU16-encoding counterpart to
// MarshalBinInto.
func MarshalCompactU16Into(v any, dst []byte) (int, error) {
	return pooledMarshalInto(&compactU16EncoderPool, v, dst)
}

func UnmarshalBin(v any, b []byte) error {
	return pooledUnmarshal(&binDecoderPool, v, b)
}

func UnmarshalBorsh(v any, b []byte) error {
	return pooledUnmarshal(&borshDecoderPool, v, b)
}

func UnmarshalCompactU16(v any, b []byte) error {
	return pooledUnmarshal(&compactU16DecoderPool, v, b)
}

type byteCounter struct {
	count uint64
}

func (c *byteCounter) Write(p []byte) (n int, err error) {
	c.count += uint64(len(p))
	return len(p), nil
}

// BinByteCount computes the byte count size for the received populated structure. The reported size
// is the one for the populated structure received in arguments. Depending on how serialization of
// your fields is performed, size could vary for different structure.
func BinByteCount(v any) (uint64, error) {
	counter := byteCounter{}
	err := NewBinEncoder(&counter).Encode(v)
	if err != nil {
		return 0, fmt.Errorf("encode %T: %w", v, err)
	}
	return counter.count, nil
}

// BorshByteCount computes the byte count size for the received populated structure. The reported size
// is the one for the populated structure received in arguments. Depending on how serialization of
// your fields is performed, size could vary for different structure.
func BorshByteCount(v any) (uint64, error) {
	counter := byteCounter{}
	err := NewBorshEncoder(&counter).Encode(v)
	if err != nil {
		return 0, fmt.Errorf("encode %T: %w", v, err)
	}
	return counter.count, nil
}

// CompactU16ByteCount computes the byte count size for the received populated structure. The reported size
// is the one for the populated structure received in arguments. Depending on how serialization of
// your fields is performed, size could vary for different structure.
func CompactU16ByteCount(v any) (uint64, error) {
	counter := byteCounter{}
	err := NewCompactU16Encoder(&counter).Encode(v)
	if err != nil {
		return 0, fmt.Errorf("encode %T: %w", v, err)
	}
	return counter.count, nil
}

// MustBinByteCount acts just like BinByteCount but panics if it encounters any encoding errors.
func MustBinByteCount(v any) uint64 {
	count, err := BinByteCount(v)
	if err != nil {
		panic(err)
	}
	return count
}

// MustBorshByteCount acts just like BorshByteCount but panics if it encounters any encoding errors.
func MustBorshByteCount(v any) uint64 {
	count, err := BorshByteCount(v)
	if err != nil {
		panic(err)
	}
	return count
}

// MustCompactU16ByteCount acts just like CompactU16ByteCount but panics if it encounters any encoding errors.
func MustCompactU16ByteCount(v any) uint64 {
	count, err := CompactU16ByteCount(v)
	if err != nil {
		panic(err)
	}
	return count
}
