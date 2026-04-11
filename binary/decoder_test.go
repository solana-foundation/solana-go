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
	"encoding/hex"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecoder_Peek(t *testing.T) {
	buf := []byte{
		0x17, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0, 0x0,
	}

	dec := NewBinDecoder(buf)
	{
		peeked, err := dec.Peek(8)
		assert.NoError(t, err)
		assert.Len(t, peeked, 8)
		assert.Equal(t, buf, peeked)
	}
	{
		peeked, err := dec.Peek(8)
		assert.NoError(t, err)
		assert.Len(t, peeked, 8)
		assert.Equal(t, buf, peeked)
	}
	{
		peeked, err := dec.Peek(1)
		assert.NoError(t, err)
		assert.Len(t, peeked, 1)
		assert.Equal(t, buf[0], peeked[0])
	}
	{
		peeked, err := dec.Peek(2)
		assert.NoError(t, err)
		assert.Len(t, peeked, 2)
		assert.Equal(t, buf[:2], peeked)
	}
	{
		read, err := dec.ReadByte()
		assert.Equal(t, buf[0], read)
		assert.NoError(t, err)

		peeked, err := dec.Peek(1)
		assert.NoError(t, err)
		assert.Len(t, peeked, 1)
		assert.Equal(t, buf[1], peeked[0])
	}
}

func TestDecoder_AliastTestType(t *testing.T) {
	buf := []byte{
		0x17, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0, 0x0,
	}

	var s aliasTestType
	err := NewBinDecoder(buf).Decode(&s)
	assert.NoError(t, err)
	assert.Equal(t, uint64(23), uint64(s))
}

func TestDecoder_Remaining(t *testing.T) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint16(b, 1)
	binary.LittleEndian.PutUint16(b[2:], 2)

	d := NewBinDecoder(b)

	n, err := d.ReadUint16(LE)
	assert.NoError(t, err)
	assert.Equal(t, uint16(1), n)
	assert.Equal(t, 2, d.Remaining())

	n, err = d.ReadUint16(LE)
	assert.NoError(t, err)
	assert.Equal(t, uint16(2), n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_int8(t *testing.T) {
	buf := []byte{
		0x9d, // -99
		0x64, // 100
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadInt8()
	assert.NoError(t, err)
	assert.Equal(t, int8(-99), n)
	assert.Equal(t, 1, d.Remaining())

	n, err = d.ReadInt8()
	assert.NoError(t, err)
	assert.Equal(t, int8(100), n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_int16(t *testing.T) {
	// little endian
	buf := []byte{
		0xae, 0xff, // -82
		0x49, 0x00, // 73
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadInt16(LE)
	assert.NoError(t, err)
	assert.Equal(t, int16(-82), n)
	assert.Equal(t, 2, d.Remaining())

	n, err = d.ReadInt16(LE)
	assert.NoError(t, err)
	assert.Equal(t, int16(73), n)
	assert.Equal(t, 0, d.Remaining())

	// big endian
	buf = []byte{
		0xff, 0xae, // -82
		0x00, 0x49, // 73
	}

	d = NewBinDecoder(buf)

	n, err = d.ReadInt16(BE)
	assert.NoError(t, err)
	assert.Equal(t, int16(-82), n)
	assert.Equal(t, 2, d.Remaining())

	n, err = d.ReadInt16(BE)
	assert.NoError(t, err)
	assert.Equal(t, int16(73), n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_int32(t *testing.T) {
	// little endian
	buf := []byte{
		0xd8, 0x8d, 0x8a, 0xef, // -276132392
		0x4f, 0x9f, 0x3, 0x00, // 237391
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadInt32(LE)
	assert.NoError(t, err)
	assert.Equal(t, int32(-276132392), n)
	assert.Equal(t, 4, d.Remaining())

	n, err = d.ReadInt32(LE)
	assert.NoError(t, err)
	assert.Equal(t, int32(237391), n)
	assert.Equal(t, 0, d.Remaining())

	// big endian
	buf = []byte{
		0xef, 0x8a, 0x8d, 0xd8, // -276132392
		0x00, 0x3, 0x9f, 0x4f, // 237391
	}

	d = NewBinDecoder(buf)

	n, err = d.ReadInt32(BE)
	assert.NoError(t, err)
	assert.Equal(t, int32(-276132392), n)
	assert.Equal(t, 4, d.Remaining())

	n, err = d.ReadInt32(BE)
	assert.NoError(t, err)
	assert.Equal(t, int32(237391), n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_int64(t *testing.T) {
	// little endian
	buf := []byte{
		0x91, 0x7d, 0xf3, 0xff, 0xff, 0xff, 0xff, 0xff, //-819823
		0xe3, 0x1c, 0x1, 0x00, 0x00, 0x00, 0x00, 0x00, //72931
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadInt64(LE)
	assert.NoError(t, err)
	assert.Equal(t, int64(-819823), n)
	assert.Equal(t, 8, d.Remaining())

	n, err = d.ReadInt64(LE)
	assert.NoError(t, err)
	assert.Equal(t, int64(72931), n)
	assert.Equal(t, 0, d.Remaining())

	// big endian
	buf = []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xf3, 0x7d, 0x91, //-819823
		0x00, 0x00, 0x00, 0x00, 0x00, 0x1, 0x1c, 0xe3, //72931
	}

	d = NewBinDecoder(buf)

	n, err = d.ReadInt64(BE)
	assert.NoError(t, err)
	assert.Equal(t, int64(-819823), n)
	assert.Equal(t, 8, d.Remaining())

	n, err = d.ReadInt64(BE)
	assert.NoError(t, err)
	assert.Equal(t, int64(72931), n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_uint8(t *testing.T) {
	buf := []byte{
		0x63, // 99
		0x64, // 100
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadUint8()
	assert.NoError(t, err)
	assert.Equal(t, uint8(99), n)
	assert.Equal(t, 1, d.Remaining())

	n, err = d.ReadUint8()
	assert.NoError(t, err)
	assert.Equal(t, uint8(100), n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_uint16(t *testing.T) {
	// little endian
	buf := []byte{
		0x52, 0x00, // 82
		0x49, 0x00, // 73
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadUint16(LE)
	assert.NoError(t, err)
	assert.Equal(t, uint16(82), n)
	assert.Equal(t, 2, d.Remaining())

	n, err = d.ReadUint16(LE)
	assert.NoError(t, err)
	assert.Equal(t, uint16(73), n)
	assert.Equal(t, 0, d.Remaining())

	// big endian
	buf = []byte{
		0x00, 0x52, // 82
		0x00, 0x49, // 73
	}

	d = NewBinDecoder(buf)

	n, err = d.ReadUint16(BE)
	assert.NoError(t, err)
	assert.Equal(t, uint16(82), n)
	assert.Equal(t, 2, d.Remaining())

	n, err = d.ReadUint16(BE)
	assert.NoError(t, err)
	assert.Equal(t, uint16(73), n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_uint32(t *testing.T) {
	// little endian
	buf := []byte{
		0x28, 0x72, 0x75, 0x10, // 276132392 as LE
		0x4f, 0x9f, 0x03, 0x00, // 237391 as LE
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadUint32(LE)
	assert.NoError(t, err)
	assert.Equal(t, uint32(276132392), n)
	assert.Equal(t, 4, d.Remaining())

	n, err = d.ReadUint32(LE)
	assert.NoError(t, err)
	assert.Equal(t, uint32(237391), n)
	assert.Equal(t, 0, d.Remaining())

	// big endian
	buf = []byte{
		0x10, 0x75, 0x72, 0x28, // 276132392 as LE
		0x00, 0x03, 0x9f, 0x4f, // 237391 as LE
	}

	d = NewBinDecoder(buf)

	n, err = d.ReadUint32(BE)
	assert.NoError(t, err)
	assert.Equal(t, uint32(276132392), n)
	assert.Equal(t, 4, d.Remaining())

	n, err = d.ReadUint32(BE)
	assert.NoError(t, err)
	assert.Equal(t, uint32(237391), n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_uint64(t *testing.T) {
	// little endian
	buf := []byte{
		0x6f, 0x82, 0x0c, 0x00, 0x00, 0x00, 0x00, 0x00, //819823
		0xe3, 0x1c, 0x1, 0x00, 0x00, 0x00, 0x00, 0x00, //72931
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadUint64(LE)
	assert.NoError(t, err)
	assert.Equal(t, uint64(819823), n)
	assert.Equal(t, 8, d.Remaining())

	n, err = d.ReadUint64(LE)
	assert.NoError(t, err)
	assert.Equal(t, uint64(72931), n)
	assert.Equal(t, 0, d.Remaining())

	// big endian
	buf = []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x0c, 0x82, 0x6f, //819823
		0x00, 0x00, 0x00, 0x00, 0x00, 0x1, 0x1c, 0xe3, //72931
	}

	d = NewBinDecoder(buf)

	n, err = d.ReadUint64(BE)
	assert.NoError(t, err)
	assert.Equal(t, uint64(819823), n)
	assert.Equal(t, 8, d.Remaining())

	n, err = d.ReadUint64(BE)
	assert.NoError(t, err)
	assert.Equal(t, uint64(72931), n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_float32(t *testing.T) {
	// little endian
	buf := []byte{
		0xc3, 0xf5, 0xa8, 0x3f,
		0xa4, 0x70, 0x4d, 0xc0,
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadFloat32(LE)
	assert.NoError(t, err)
	assert.Equal(t, float32(1.32), n)
	assert.Equal(t, 4, d.Remaining())

	n, err = d.ReadFloat32(LE)
	assert.NoError(t, err)
	assert.Equal(t, float32(-3.21), n)
	assert.Equal(t, 0, d.Remaining())

	// big endian
	buf = []byte{
		0x3f, 0xa8, 0xf5, 0xc3,
		0xc0, 0x4d, 0x70, 0xa4,
	}

	d = NewBinDecoder(buf)

	n, err = d.ReadFloat32(BE)
	assert.NoError(t, err)
	assert.Equal(t, float32(1.32), n)
	assert.Equal(t, 4, d.Remaining())

	n, err = d.ReadFloat32(BE)
	assert.NoError(t, err)
	assert.Equal(t, float32(-3.21), n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_float64(t *testing.T) {
	// little endian
	buf := []byte{
		0x3d, 0x0a, 0xd7, 0xa3, 0x70, 0x1d, 0x4f, 0xc0,
		0x77, 0xbe, 0x9f, 0x1a, 0x2f, 0x3d, 0x37, 0x40,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x7f,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0xff,
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf8, 0x7f,
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadFloat64(LE)
	assert.NoError(t, err)
	assert.Equal(t, float64(-62.23), n)
	assert.Equal(t, 32, d.Remaining())

	n, err = d.ReadFloat64(LE)
	assert.NoError(t, err)
	assert.Equal(t, float64(23.239), n)
	assert.Equal(t, 24, d.Remaining())

	n, err = d.ReadFloat64(LE)
	assert.NoError(t, err)
	assert.Equal(t, math.Inf(1), n)
	assert.Equal(t, 16, d.Remaining())

	n, err = d.ReadFloat64(LE)
	assert.NoError(t, err)
	assert.Equal(t, math.Inf(-1), n)
	assert.Equal(t, 8, d.Remaining())

	n, err = d.ReadFloat64(LE)
	assert.NoError(t, err)
	assert.True(t, math.IsNaN(n))

	// big endian
	buf = []byte{
		0xc0, 0x4f, 0x1d, 0x70, 0xa3, 0xd7, 0x0a, 0x3d,
		0x40, 0x37, 0x3d, 0x2f, 0x1a, 0x9f, 0xbe, 0x77,
		0x7f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xff, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x7f, 0xf8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
	}

	d = NewBinDecoder(buf)

	n, err = d.ReadFloat64(BE)
	assert.NoError(t, err)
	assert.Equal(t, float64(-62.23), n)
	assert.Equal(t, 32, d.Remaining())

	n, err = d.ReadFloat64(BE)
	assert.NoError(t, err)
	assert.Equal(t, float64(23.239), n)
	assert.Equal(t, 24, d.Remaining())

	n, err = d.ReadFloat64(BE)
	assert.NoError(t, err)
	assert.Equal(t, math.Inf(1), n)
	assert.Equal(t, 16, d.Remaining())

	n, err = d.ReadFloat64(BE)
	assert.NoError(t, err)
	assert.Equal(t, math.Inf(-1), n)
	assert.Equal(t, 8, d.Remaining())

	n, err = d.ReadFloat64(BE)
	assert.NoError(t, err)
	assert.True(t, math.IsNaN(n))
}

func TestDecoder_string(t *testing.T) {
	buf := []byte{
		0x03, 0x31, 0x32, 0x33, // "123"
		0x00,                   // ""
		0x03, 0x61, 0x62, 0x63, // "abc
	}

	d := NewBinDecoder(buf)

	s, err := d.ReadString()
	assert.NoError(t, err)
	assert.Equal(t, "123", s)
	assert.Equal(t, 5, d.Remaining())

	s, err = d.ReadString()
	assert.NoError(t, err)
	assert.Equal(t, "", s)
	assert.Equal(t, 4, d.Remaining())

	s, err = d.ReadString()
	assert.NoError(t, err)
	assert.Equal(t, "abc", s)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_Decode_String_Err(t *testing.T) {
	buf := []byte{
		0x01, 0x00, 0x00, 0x00,
		byte('a'),
	}

	decoder := NewBinDecoder(buf)

	var s string
	err := decoder.Decode(&s)
	assert.EqualError(t, err, "decode: uint64 required [8] bytes, remaining [5]")
}

func TestDecoder_Byte(t *testing.T) {
	buf := []byte{
		0x00, 0x01,
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, byte(0), n)
	assert.Equal(t, 1, d.Remaining())

	n, err = d.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, byte(1), n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_Bool(t *testing.T) {
	buf := []byte{
		0x01, 0x00,
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadBool()
	assert.NoError(t, err)
	assert.Equal(t, true, n)
	assert.Equal(t, 1, d.Remaining())

	n, err = d.ReadBool()
	assert.NoError(t, err)
	assert.Equal(t, false, n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_ByteArray(t *testing.T) {
	buf := []byte{
		0x03, 0x01, 0x02, 0x03,
		0x03, 0x04, 0x05, 0x06,
	}

	d := NewBinDecoder(buf)

	data, err := d.ReadByteSlice()
	assert.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3}, data)
	assert.Equal(t, 4, d.Remaining())

	data, err = d.ReadByteSlice()
	assert.Equal(t, []byte{4, 5, 6}, data)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_ByteArray_MissingData(t *testing.T) {
	buf := []byte{
		0x0a,
	}

	d := NewBinDecoder(buf)

	_, err := d.ReadByteSlice()
	assert.EqualError(t, err, "byte array: varlen=10, missing 10 bytes")
}

func TestDecoder_Array(t *testing.T) {
	buf := []byte{1, 2, 4}

	decoder := NewBinDecoder(buf)

	var decoded [3]byte
	decoder.Decode(&decoded)
	assert.Equal(t, [3]byte{1, 2, 4}, decoded)
}

func TestDecoder_Slice_Err(t *testing.T) {
	buf := []byte{}

	decoder := NewBinDecoder(buf)
	var s []string
	err := decoder.Decode(&s)
	assert.Equal(t, ErrVarIntBufferSize, err)

	buf = []byte{0x01}

	decoder = NewBinDecoder(buf)
	err = decoder.Decode(&s)
	assert.EqualError(t, err, "unexpected EOF")
}

func TestDecoder_Slice_InvalidLen(t *testing.T) {
	buf := []byte{0xd7, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}

	decoder := NewBinDecoder(buf)
	var s []string
	err := decoder.Decode(&s)
	assert.EqualError(t, err, "unexpected EOF")
}

func TestDecoder_Int64(t *testing.T) {
	// little endian
	buf := []byte{
		0x91, 0x7d, 0xf3, 0xff, 0xff, 0xff, 0xff, 0xff, //-819823
		0xe3, 0x1c, 0x1, 0x00, 0x00, 0x00, 0x00, 0x00, //72931
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadInt64(LE)
	assert.NoError(t, err)
	assert.Equal(t, int64(-819823), n)
	assert.Equal(t, 8, d.Remaining())

	n, err = d.ReadInt64(LE)
	assert.NoError(t, err)
	assert.Equal(t, int64(72931), n)
	assert.Equal(t, 0, d.Remaining())

	// big endian
	buf = []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xf3, 0x7d, 0x91, //-819823
		0x00, 0x00, 0x00, 0x00, 0x00, 0x1, 0x1c, 0xe3, //72931
	}

	d = NewBinDecoder(buf)

	n, err = d.ReadInt64(BE)
	assert.NoError(t, err)
	assert.Equal(t, int64(-819823), n)
	assert.Equal(t, 8, d.Remaining())

	n, err = d.ReadInt64(BE)
	assert.NoError(t, err)
	assert.Equal(t, int64(72931), n)
	assert.Equal(t, 0, d.Remaining())
}

func TestDecoder_Uint128_2(t *testing.T) {
	// little endian
	buf := []byte{
		0x0d, 0x88, 0xd3, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x6d, 0x0b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	d := NewBinDecoder(buf)

	n, err := d.ReadUint128(LE)
	assert.NoError(t, err)
	assert.Equal(t, Uint128{Hi: 0xb6d, Lo: 0xffffffffffd3880d}, n)

	buf = []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0xbb,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xac, 0xdc, 0xad,
	}

	d = NewBinDecoder(buf)

	n, err = d.ReadUint128(BE)
	assert.NoError(t, err)
	assert.Equal(t, Uint128{Hi: 0x00000000000008bb, Lo: 0xffffffffffacdcad}, n)

}

func TestDecoder_BinaryStruct(t *testing.T) {
	cnt, err := hex.DecodeString("0300000000000000616263b5ff630019ffffffe703000051ccffffffffffff9f860100000000003d0ab9c15c8fc2f5285c0f4002030000000000000064656603000000000000003738390300000000000000666f6f0300000000000000626172ff05010203040501e9ffffffffffffff17000000000000001f85eb51b81e09400a000000000000005200000000000000070000000000000003000000000000000a000000000000005200000000000000e707cd0f01050102030405")
	require.NoError(t, err)

	s := binaryTestStruct{}
	decoder := NewBinDecoder(cnt)
	assert.NoError(t, decoder.Decode(&s))

	assert.Equal(t, "abc", s.F1)
	assert.Equal(t, int16(-75), s.F2)
	assert.Equal(t, uint16(99), s.F3)
	assert.Equal(t, int32(-231), s.F4)
	assert.Equal(t, uint32(999), s.F5)
	assert.Equal(t, int64(-13231), s.F6)
	assert.Equal(t, uint64(99999), s.F7)
	assert.Equal(t, float32(-23.13), s.F8)
	assert.Equal(t, float64(3.92), s.F9)
	assert.Equal(t, []string{"def", "789"}, s.F10)
	assert.Equal(t, [2]string{"foo", "bar"}, s.F11)
	assert.Equal(t, uint8(0xff), s.F12)
	assert.Equal(t, []byte{1, 2, 3, 4, 5}, s.F13)
	assert.Equal(t, true, s.F14)
	assert.Equal(t, Int64(-23), s.F15)
	assert.Equal(t, Uint64(23), s.F16)
	assert.Equal(t, JSONFloat64(3.14), s.F17)
	assert.Equal(t, Uint128{
		Lo: 10,
		Hi: 82,
	}, s.F18)
	assert.Equal(t, Int128{
		Lo: 7,
		Hi: 3,
	}, s.F19)
	assert.Equal(t, Float128{
		Lo: 10,
		Hi: 82,
	}, s.F20)
	assert.Equal(t, Varuint32(999), s.F21)
	assert.Equal(t, Varint32(-999), s.F22)
	assert.Equal(t, Bool(true), s.F23)
	assert.Equal(t, HexBytes([]byte{1, 2, 3, 4, 5}), s.F24)
}

func TestDecoder_Decode_No_Ptr(t *testing.T) {
	decoder := NewBinDecoder([]byte{})
	err := decoder.Decode(1)
	assert.EqualError(t, err, "decoder: Decode(non-pointer int)")
}

func TestDecoder_BinaryTestStructWithTags(t *testing.T) {
	cnt, err := hex.DecodeString("ffb50063ffffff19000003e7ffffffffffffcc51000000000001869fc1b90a3d400f5c28f5c28f5c010000000000000000")
	require.NoError(t, err)

	s := &binaryTestStructWithTags{}
	decoder := NewBinDecoder(cnt)
	assert.NoError(t, decoder.Decode(s))

	assert.Equal(t, "", s.F1)
	assert.Equal(t, int16(-75), s.F2)
	assert.Equal(t, uint16(99), s.F3)
	assert.Equal(t, int32(-231), s.F4)
	assert.Equal(t, uint32(999), s.F5)
	assert.Equal(t, int64(-13231), s.F6)
	assert.Equal(t, uint64(99999), s.F7)
	assert.Equal(t, float32(-23.13), s.F8)
	assert.Equal(t, float64(3.92), s.F9)
	assert.Equal(t, true, s.F10)
	var i *Int64
	assert.Equal(t, i, s.F11)
}

func TestDecoder_SkipBytes(t *testing.T) {
	buf := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	decoder := NewBinDecoder(buf)
	err := decoder.SkipBytes(1)
	require.NoError(t, err)
	require.Equal(t, 7, decoder.Remaining())

	err = decoder.SkipBytes(2)
	require.NoError(t, err)
	require.Equal(t, 5, decoder.Remaining())

	err = decoder.SkipBytes(6)
	require.Error(t, err)

	err = decoder.SkipBytes(5)
	require.NoError(t, err)
	require.Equal(t, 0, decoder.Remaining())
}

func Test_Discard(t *testing.T) {
	buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	decoder := NewBinDecoder(buf)
	err := decoder.Discard(5)
	require.NoError(t, err)
	require.Equal(t, 5, decoder.Remaining())
	remaining, err := decoder.Peek(decoder.Remaining())
	require.NoError(t, err)
	require.Equal(t, []byte{5, 6, 7, 8, 9}, remaining)
}

func Test_reflect_readArrayOfBytes(t *testing.T) {
	{
		{
			buf := []byte{0, 1, 2, 3, 4, 5, 6, 7}
			decoder := NewBinDecoder(buf)

			got := make([]byte, 0)
			err := reflect_readArrayOfBytes(decoder, len(buf), reflect.ValueOf(&got).Elem())
			require.NoError(t, err)
			require.Equal(t, buf, got)
		}
		{
			buf := []byte{0, 1, 2, 3, 4, 5, 6, 7}
			decoder := NewBinDecoder(buf)

			got := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
			err := reflect_readArrayOfBytes(decoder, len(buf), reflect.ValueOf(&got).Elem())
			require.NoError(t, err)
			require.Equal(t, buf, got[:])
		}
	}
	{
		{
			buf := []byte{0, 1, 2, 3, 4, 5, 6, 7}
			decoder := NewBorshDecoder(buf)

			got := make([]byte, 0)
			err := reflect_readArrayOfBytes(decoder, len(buf), reflect.ValueOf(&got).Elem())
			require.NoError(t, err)
			require.Equal(t, buf, got)
		}
		{
			buf := []byte{0, 1, 2, 3, 4, 5, 6, 7}
			decoder := NewBorshDecoder(buf)

			got := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
			err := reflect_readArrayOfBytes(decoder, len(buf), reflect.ValueOf(&got).Elem())
			require.NoError(t, err)
			require.Equal(t, buf, got[:])
		}
	}
}

func Test_reflect_readArrayOfUint16(t *testing.T) {
	{
		{
			buf := concatByteSlices(
				uint16ToBytes(0, LE),
				uint16ToBytes(1, LE),
				uint16ToBytes(2, LE),
				uint16ToBytes(3, LE),
				uint16ToBytes(4, LE),
				uint16ToBytes(5, LE),
				uint16ToBytes(6, LE),
				uint16ToBytes(7, LE),
			)
			decoder := NewBinDecoder(buf)

			got := make([]uint16, 0)
			err := reflect_readArrayOfUint16(decoder, len(buf)/2, reflect.ValueOf(&got).Elem(), LE)
			require.NoError(t, err)
			require.Equal(t, []uint16{0, 1, 2, 3, 4, 5, 6, 7}, got)
		}
		{
			buf := concatByteSlices(
				uint16ToBytes(0, LE),
				uint16ToBytes(1, LE),
				uint16ToBytes(2, LE),
				uint16ToBytes(3, LE),
				uint16ToBytes(4, LE),
				uint16ToBytes(5, LE),
				uint16ToBytes(6, LE),
				uint16ToBytes(7, LE),
			)
			decoder := NewBinDecoder(buf)

			got := [8]uint16{0, 0, 0, 0, 0, 0, 0, 0}
			err := reflect_readArrayOfUint16(decoder, len(buf)/2, reflect.ValueOf(&got).Elem(), LE)
			require.NoError(t, err)
			require.Equal(t, []uint16{0, 1, 2, 3, 4, 5, 6, 7}, got[:])
		}
	}
	{
		{
			buf := concatByteSlices(
				uint16ToBytes(0, LE),
				uint16ToBytes(1, LE),
				uint16ToBytes(2, LE),
				uint16ToBytes(3, LE),
				uint16ToBytes(4, LE),
				uint16ToBytes(5, LE),
				uint16ToBytes(6, LE),
				uint16ToBytes(7, LE),
			)
			decoder := NewBorshDecoder(buf)

			got := make([]uint16, 0)
			err := reflect_readArrayOfUint16(decoder, len(buf)/2, reflect.ValueOf(&got).Elem(), LE)
			require.NoError(t, err)
			require.Equal(t, []uint16{0, 1, 2, 3, 4, 5, 6, 7}, got)
		}
		{
			buf := concatByteSlices(
				uint16ToBytes(0, LE),
				uint16ToBytes(1, LE),
				uint16ToBytes(2, LE),
				uint16ToBytes(3, LE),
				uint16ToBytes(4, LE),
				uint16ToBytes(5, LE),
				uint16ToBytes(6, LE),
				uint16ToBytes(7, LE),
			)
			decoder := NewBorshDecoder(buf)

			got := [8]uint16{0, 0, 0, 0, 0, 0, 0, 0}
			err := reflect_readArrayOfUint16(decoder, len(buf)/2, reflect.ValueOf(&got).Elem(), LE)
			require.NoError(t, err)
			require.Equal(t, []uint16{0, 1, 2, 3, 4, 5, 6, 7}, got[:])
		}
	}
}

func Test_reflect_readArrayOfUint32(t *testing.T) {
	{
		{
			buf := concatByteSlices(
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
				uint32ToBytes(3, LE),
				uint32ToBytes(4, LE),
				uint32ToBytes(5, LE),
				uint32ToBytes(6, LE),
				uint32ToBytes(7, LE),
			)
			decoder := NewBinDecoder(buf)

			got := make([]uint32, 0)
			err := reflect_readArrayOfUint32(decoder, len(buf)/4, reflect.ValueOf(&got).Elem(), LE)
			require.NoError(t, err)
			require.Equal(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7}, got)
		}
		{
			buf := concatByteSlices(
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
				uint32ToBytes(3, LE),
				uint32ToBytes(4, LE),
				uint32ToBytes(5, LE),
				uint32ToBytes(6, LE),
				uint32ToBytes(7, LE),
			)
			decoder := NewBinDecoder(buf)

			got := [8]uint32{0, 0, 0, 0, 0, 0, 0, 0}
			err := reflect_readArrayOfUint32(decoder, len(buf)/4, reflect.ValueOf(&got).Elem(), LE)
			require.NoError(t, err)
			require.Equal(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7}, got[:])
		}
	}
	{
		{
			buf := concatByteSlices(
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
				uint32ToBytes(3, LE),
				uint32ToBytes(4, LE),
				uint32ToBytes(5, LE),
				uint32ToBytes(6, LE),
				uint32ToBytes(7, LE),
			)
			decoder := NewBorshDecoder(buf)

			got := make([]uint32, 0)
			err := reflect_readArrayOfUint32(decoder, len(buf)/4, reflect.ValueOf(&got).Elem(), LE)
			require.NoError(t, err)
			require.Equal(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7}, got)
		}
		{
			buf := concatByteSlices(
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
				uint32ToBytes(3, LE),
				uint32ToBytes(4, LE),
				uint32ToBytes(5, LE),
				uint32ToBytes(6, LE),
				uint32ToBytes(7, LE),
			)
			decoder := NewBorshDecoder(buf)

			got := [8]uint32{0, 0, 0, 0, 0, 0, 0, 0}
			err := reflect_readArrayOfUint32(decoder, len(buf)/4, reflect.ValueOf(&got).Elem(), LE)
			require.NoError(t, err)
			require.Equal(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7}, got[:])
		}
	}
}

func Test_reflect_readArrayOfUint64(t *testing.T) {
	{
		{
			buf := concatByteSlices(
				uint64ToBytes(0, LE),
				uint64ToBytes(1, LE),
				uint64ToBytes(2, LE),
				uint64ToBytes(3, LE),
				uint64ToBytes(4, LE),
				uint64ToBytes(5, LE),
				uint64ToBytes(6, LE),
				uint64ToBytes(7, LE),
			)
			decoder := NewBinDecoder(buf)

			got := make([]uint64, 0)
			err := reflect_readArrayOfUint64(decoder, len(buf)/8, reflect.ValueOf(&got).Elem(), LE)
			require.NoError(t, err)
			require.Equal(t, []uint64{0, 1, 2, 3, 4, 5, 6, 7}, got)
		}
		{
			buf := concatByteSlices(
				uint64ToBytes(0, LE),
				uint64ToBytes(1, LE),
				uint64ToBytes(2, LE),
				uint64ToBytes(3, LE),
				uint64ToBytes(4, LE),
				uint64ToBytes(5, LE),
				uint64ToBytes(6, LE),
				uint64ToBytes(7, LE),
			)
			decoder := NewBinDecoder(buf)
			got := [8]uint64{0, 0, 0, 0, 0, 0, 0, 0}
			err := reflect_readArrayOfUint64(decoder, len(buf)/8, reflect.ValueOf(&got).Elem(), LE)
			require.NoError(t, err)
			require.Equal(t, []uint64{0, 1, 2, 3, 4, 5, 6, 7}, got[:])
		}
	}
	{
		{
			buf := concatByteSlices(
				uint64ToBytes(0, LE),
				uint64ToBytes(1, LE),
				uint64ToBytes(2, LE),
				uint64ToBytes(3, LE),
				uint64ToBytes(4, LE),
				uint64ToBytes(5, LE),
				uint64ToBytes(6, LE),
				uint64ToBytes(7, LE),
			)
			decoder := NewBorshDecoder(buf)

			got := make([]uint64, 0)
			err := reflect_readArrayOfUint64(decoder, len(buf)/8, reflect.ValueOf(&got).Elem(), LE)
			require.NoError(t, err)
			require.Equal(t, []uint64{0, 1, 2, 3, 4, 5, 6, 7}, got)
		}
		{
			buf := concatByteSlices(
				uint64ToBytes(0, LE),
				uint64ToBytes(1, LE),
				uint64ToBytes(2, LE),
				uint64ToBytes(3, LE),
				uint64ToBytes(4, LE),
				uint64ToBytes(5, LE),
				uint64ToBytes(6, LE),
				uint64ToBytes(7, LE),
			)
			decoder := NewBorshDecoder(buf)
			got := [8]uint64{0, 0, 0, 0, 0, 0, 0, 0}
			err := reflect_readArrayOfUint64(decoder, len(buf)/8, reflect.ValueOf(&got).Elem(), LE)
			require.NoError(t, err)
			require.Equal(t, []uint64{0, 1, 2, 3, 4, 5, 6, 7}, got[:])
		}
	}
}

func Test_reflect_readArrayOfUint(t *testing.T) {
	{
		{
			buf := concatByteSlices(
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
				uint32ToBytes(3, LE),
				uint32ToBytes(4, LE),
				uint32ToBytes(5, LE),
				uint32ToBytes(6, LE),
				uint32ToBytes(7, LE),
			)
			decoder := NewBinDecoder(buf)

			got := make([]uint32, 0)
			rv := reflect.ValueOf(&got).Elem()
			k := rv.Type().Elem().Kind()
			err := reflect_readArrayOfUint_(decoder, len(buf)/4, k, rv, LE)
			require.NoError(t, err)
			require.Equal(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7}, got)
		}
		{
			buf := concatByteSlices(
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
				uint32ToBytes(3, LE),
				uint32ToBytes(4, LE),
				uint32ToBytes(5, LE),
				uint32ToBytes(6, LE),
				uint32ToBytes(7, LE),
			)
			decoder := NewBinDecoder(buf)
			got := [8]uint32{0, 0, 0, 0, 0, 0, 0, 0}
			rv := reflect.ValueOf(&got).Elem()
			k := rv.Type().Elem().Kind()
			err := reflect_readArrayOfUint_(decoder, len(buf)/4, k, rv, LE)
			require.NoError(t, err)
			require.Equal(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7}, got[:])
		}
	}
}

func Test_Decode_custom(t *testing.T) {
	t.Run("custom-type-uint32 slice", func(t *testing.T) {
		{
			buf := concatByteSlices(
				// length:
				[]byte{3},
				// data:
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
			)
			decoder := NewBinDecoder(buf)

			type CustomUint32 uint32
			got := make([]CustomUint32, 0)
			err := decoder.Decode(&got)
			require.NoError(t, err)
			require.Equal(t, []CustomUint32{0, 1, 2}, got)
		}
	})
	t.Run("custom-type-uint32 array", func(t *testing.T) {
		{
			buf := concatByteSlices(
				// data:
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
			)
			decoder := NewBinDecoder(buf)

			type CustomUint32 uint32
			got := [3]CustomUint32{0, 0, 0}
			err := decoder.Decode(&got)
			require.NoError(t, err)
			require.Equal(t, [3]CustomUint32{0, 1, 2}, got)
		}
	})
	t.Run("uint32 custom-type-slice", func(t *testing.T) {
		{
			buf := concatByteSlices(
				// length:
				[]byte{3},
				// data:
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
			)
			decoder := NewBinDecoder(buf)

			type CustomSliceUint32 []uint32
			got := make(CustomSliceUint32, 0)
			err := decoder.Decode(&got)
			require.NoError(t, err)
			require.Equal(t, CustomSliceUint32{0, 1, 2}, got)
		}
	})
	t.Run("uint32 custom-type-array", func(t *testing.T) {
		{
			buf := concatByteSlices(
				// data:
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
			)
			decoder := NewBinDecoder(buf)

			type CustomArrayUint32 [3]uint32
			got := CustomArrayUint32{0, 0, 0}
			err := decoder.Decode(&got)
			require.NoError(t, err)
			require.Equal(t, CustomArrayUint32{0, 1, 2}, got)
		}
	})
}

func Test_ReadNBytes(t *testing.T) {
	{
		b1 := []byte{123, 99, 88, 77, 66, 55, 44, 33, 22, 11}
		b2 := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		buf := concatByteSlices(
			b1,
			b2,
		)
		decoder := NewBinDecoder(buf)

		got, err := decoder.ReadNBytes(10)
		require.NoError(t, err)
		require.Equal(t, b1, got)

		got, err = decoder.ReadNBytes(10)
		require.NoError(t, err)
		require.Equal(t, b2, got)
	}
}

func Test_ReadBytes(t *testing.T) {
	{
		b1 := []byte{123, 99, 88, 77, 66, 55, 44, 33, 22, 11}
		b2 := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		buf := concatByteSlices(
			b1,
			b2,
		)
		decoder := NewBinDecoder(buf)

		got, err := decoder.ReadBytes(10)
		require.NoError(t, err)
		require.Equal(t, b1, got)

		got, err = decoder.ReadBytes(10)
		require.NoError(t, err)
		require.Equal(t, b2, got)
	}
}

func Test_Read(t *testing.T) {
	{
		b1 := []byte{123, 99, 88, 77, 66, 55, 44, 33, 22, 11}
		b2 := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		buf := concatByteSlices(
			b1,
			b2,
		)
		decoder := NewBinDecoder(buf)

		{
			got := make([]byte, 10)
			num, err := decoder.Read(got)
			require.NoError(t, err)
			require.Equal(t, b1, got)
			require.Equal(t, 10, num)
		}

		{
			got := make([]byte, 10)
			num, err := decoder.Read(got)
			require.NoError(t, err)
			require.Equal(t, b2, got)
			require.Equal(t, 10, num)
		}
		{
			got := make([]byte, 11)
			_, err := decoder.Read(got)
			require.EqualError(t, err, "short buffer")
		}
		{
			got := make([]byte, 0)
			num, err := decoder.Read(got)
			require.NoError(t, err)
			require.Equal(t, 0, num)
			require.Equal(t, []byte{}, got)
		}
	}
}

func Test_Decode_readArrayOfUint(t *testing.T) {
	{
		{
			buf := concatByteSlices(
				// length:
				[]byte{3},
				// data:
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
			)
			decoder := NewBinDecoder(buf)

			got := make([]uint32, 0)
			err := decoder.Decode(&got)
			require.NoError(t, err)
			require.Equal(t, []uint32{0, 1, 2}, got)
		}
		{
			buf := concatByteSlices(
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
				uint32ToBytes(3, LE),
				uint32ToBytes(4, LE),
				uint32ToBytes(5, LE),
				uint32ToBytes(6, LE),
				uint32ToBytes(7, LE),
			)
			decoder := NewBinDecoder(buf)
			got := [8]uint32{0, 0, 0, 0, 0, 0, 0, 0}
			err := decoder.Decode(&got)
			require.NoError(t, err)
			require.Equal(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7}, got[:])
		}
	}
	{
		{
			buf := concatByteSlices(
				// length:
				uint32ToBytes(8, LE),
				// data:
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
				uint32ToBytes(3, LE),
				uint32ToBytes(4, LE),
				uint32ToBytes(5, LE),
				uint32ToBytes(6, LE),
				uint32ToBytes(7, LE),
			)
			decoder := NewBorshDecoder(buf)

			got := make([]uint32, 0)
			err := decoder.Decode(&got)
			require.NoError(t, err)
			require.Equal(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7}, got)
		}
		{
			buf := concatByteSlices(
				uint32ToBytes(0, LE),
				uint32ToBytes(1, LE),
				uint32ToBytes(2, LE),
				uint32ToBytes(3, LE),
				uint32ToBytes(4, LE),
				uint32ToBytes(5, LE),
				uint32ToBytes(6, LE),
				uint32ToBytes(7, LE),
			)
			decoder := NewBorshDecoder(buf)
			got := [8]uint32{0, 0, 0, 0, 0, 0, 0, 0}
			err := decoder.Decode(&got)
			require.NoError(t, err)
			require.Equal(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7}, got[:])
		}
	}
}

func Test_reflect_readArrayOfUint16_asField(t *testing.T) {
	{
		{
			buf := concatByteSlices(
				// length:
				[]byte{8},
				// data:
				uint16ToBytes(0, LE),
				uint16ToBytes(1, LE),
				uint16ToBytes(2, LE),
				uint16ToBytes(3, LE),
				uint16ToBytes(4, LE),
				uint16ToBytes(5, LE),
				uint16ToBytes(6, LE),
				uint16ToBytes(7, LE),
			)
			decoder := NewBinDecoder(buf)

			type S struct {
				Val []uint16
			}
			var got S
			err := decoder.Decode(&got)
			require.NoError(t, err)
			require.Equal(t, S{[]uint16{0, 1, 2, 3, 4, 5, 6, 7}}, got)
		}
		{
			buf := concatByteSlices(
				// data:
				uint16ToBytes(0, LE),
				uint16ToBytes(1, LE),
				uint16ToBytes(2, LE),
				uint16ToBytes(3, LE),
				uint16ToBytes(4, LE),
				uint16ToBytes(5, LE),
				uint16ToBytes(6, LE),
				uint16ToBytes(7, LE),
			)
			decoder := NewBinDecoder(buf)

			type S struct {
				Val [8]uint16
			}
			var got S
			err := decoder.Decode(&got)
			require.NoError(t, err)
			require.Equal(t, S{[8]uint16{0, 1, 2, 3, 4, 5, 6, 7}}, got)
		}
	}
	{
		{
			buf := concatByteSlices(
				// length:
				uint32ToBytes(8, LE),
				// data:
				uint16ToBytes(0, LE),
				uint16ToBytes(1, LE),
				uint16ToBytes(2, LE),
				uint16ToBytes(3, LE),
				uint16ToBytes(4, LE),
				uint16ToBytes(5, LE),
				uint16ToBytes(6, LE),
				uint16ToBytes(7, LE),
			)
			decoder := NewBorshDecoder(buf)

			type S struct {
				Val []uint16
			}
			var got S
			err := decoder.Decode(&got)
			require.NoError(t, err)
			require.Equal(t, S{[]uint16{0, 1, 2, 3, 4, 5, 6, 7}}, got)
		}
		{
			buf := concatByteSlices(
				uint16ToBytes(0, LE),
				uint16ToBytes(1, LE),
				uint16ToBytes(2, LE),
				uint16ToBytes(3, LE),
				uint16ToBytes(4, LE),
				uint16ToBytes(5, LE),
				uint16ToBytes(6, LE),
				uint16ToBytes(7, LE),
			)
			decoder := NewBorshDecoder(buf)

			type S struct {
				Val [8]uint16
			}
			var got S
			err := decoder.Decode(&got)
			require.NoError(t, err)
			require.Equal(t, S{[8]uint16{0, 1, 2, 3, 4, 5, 6, 7}}, got)
		}
	}
}
