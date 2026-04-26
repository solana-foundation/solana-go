// Copyright 2021 github.com/gagliardetto
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
	"io"
	"math"
)

// EncodeCompactU16Length encodes a "Compact-u16" length into the provided slice pointer.
// See https://docs.solana.com/developing/programming-model/transactions#compact-u16-format
// See https://github.com/solana-labs/solana/blob/2ef2b6daa05a7cff057e9d3ef95134cee3e4045d/web3.js/src/util/shortvec-encoding.ts
func EncodeCompactU16Length(buf *[]byte, ln int) error {
	if ln < 0 || ln > math.MaxUint16 {
		return fmt.Errorf("length %d out of range", ln)
	}
	u := uint(ln)
	switch {
	case u < 0x80:
		*buf = append(*buf, byte(u))
	case u < 0x4000:
		*buf = append(*buf, byte(u)|0x80, byte(u>>7))
	default:
		*buf = append(*buf, byte(u)|0x80, byte(u>>7)|0x80, byte(u>>14))
	}
	return nil
}

// PutCompactU16Length writes a "Compact-u16" length into dst and returns the
// number of bytes written (1, 2, or 3). dst must be at least 3 bytes long.
// This is the allocation-free variant of EncodeCompactU16Length, used by the
// Encoder's scratch-buffer hot path.
func PutCompactU16Length(dst []byte, ln int) (int, error) {
	if ln < 0 || ln > math.MaxUint16 {
		return 0, fmt.Errorf("length %d out of range", ln)
	}
	u := uint(ln)
	switch {
	case u < 0x80:
		dst[0] = byte(u)
		return 1, nil
	case u < 0x4000:
		dst[0] = byte(u) | 0x80
		dst[1] = byte(u >> 7)
		return 2, nil
	default:
		dst[0] = byte(u) | 0x80
		dst[1] = byte(u>>7) | 0x80
		dst[2] = byte(u >> 14)
		return 3, nil
	}
}

const _MAX_COMPACTU16_ENCODING_LENGTH = 3

// DecodeCompactU16 decodes a Solana "Compact-u16" length from bytes and returns
// (value, bytes_consumed, error). Hand-unrolled for the max 3-byte encoding to
// avoid a per-iteration loop overhead.
func DecodeCompactU16(bytes []byte) (int, int, error) {
	if len(bytes) == 0 {
		return 0, 0, io.ErrUnexpectedEOF
	}
	b0 := int(bytes[0])
	if b0&0x80 == 0 {
		return b0, 1, nil
	}
	if len(bytes) < 2 {
		return 0, 0, io.ErrUnexpectedEOF
	}
	b1 := int(bytes[1])
	if b1&0x80 == 0 {
		if b1 == 0 {
			return 0, 0, fmt.Errorf("compact-u16: non-canonical 2-byte encoding (trailing zero byte)")
		}
		return (b0 & 0x7f) | (b1 << 7), 2, nil
	}
	if len(bytes) < 3 {
		return 0, 0, io.ErrUnexpectedEOF
	}
	b2 := int(bytes[2])
	if b2 == 0 {
		return 0, 0, fmt.Errorf("compact-u16: non-canonical 3-byte encoding (trailing zero byte)")
	}
	if b2&0x80 != 0 {
		return 0, 0, fmt.Errorf("byte three continues")
	}
	ln := (b0 & 0x7f) | ((b1 & 0x7f) << 7) | (b2 << 14)
	if ln > math.MaxUint16 {
		return 0, 0, fmt.Errorf("invalid length: %d", ln)
	}
	return ln, 3, nil
}

// DecodeCompactU16LengthFromByteReader decodes a "Compact-u16" length from the provided io.ByteReader.
func DecodeCompactU16LengthFromByteReader(reader io.ByteReader) (int, error) {
	ln := 0
	size := 0
	for nthByte := range _MAX_COMPACTU16_ENCODING_LENGTH {
		elemByte, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}
		elem := int(elemByte)
		if elem == 0 && nthByte != 0 {
			return 0, fmt.Errorf("compact-u16: non-canonical encoding (trailing zero byte at position %d)", nthByte)
		}
		if nthByte == _MAX_COMPACTU16_ENCODING_LENGTH-1 && (elem&0x80) != 0 {
			return 0, fmt.Errorf("compact-u16: byte three has continuation bit set")
		}
		ln |= (elem & 0x7f) << (size * 7)
		size += 1
		if (elem & 0x80) == 0 {
			break
		}
	}
	// check for non-valid sizes
	if size == 0 || size > _MAX_COMPACTU16_ENCODING_LENGTH {
		return 0, fmt.Errorf("compact-u16: invalid size: %d", size)
	}
	// check for non-valid lengths
	if ln < 0 || ln > math.MaxUint16 {
		return 0, fmt.Errorf("compact-u16: invalid length: %d", ln)
	}
	return ln, nil
}
