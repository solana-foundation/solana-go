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
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoder_Size(t *testing.T) {
	{
		buf := new(bytes.Buffer)

		enc := NewBinEncoder(buf)
		assert.Equal(t, enc.Written(), 0)
		enc.Encode(SafeString("hello"))

		assert.Equal(t, enc.Written(), 6)
		enc.WriteBool(true)
		assert.Equal(t, enc.Written(), 7)
	}
	{
		buf := new(bytes.Buffer)

		enc := NewBorshEncoder(buf)
		assert.Equal(t, enc.Written(), 0)
		enc.WriteByte(123)

		assert.Equal(t, enc.Written(), 1)
		enc.WriteBool(true)
		assert.Equal(t, enc.Written(), 2)
	}
}

func TestEncoder_AliastTestType(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewBinEncoder(buf)
	enc.Encode(aliasTestType(23))

	assert.Equal(t, []byte{
		0x17, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0, 0x0,
	}, buf.Bytes())
}

func TestEncoder_safeString(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.Encode(SafeString("hello"))

	assert.Equal(t, []byte{
		0x5, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
	}, buf.Bytes())
}

func TestEncoder_int8(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	v := int8(-99)
	enc.WriteByte(byte(v))
	enc.WriteByte(byte(int8(100)))

	assert.Equal(t, []byte{
		0x9d, // -99
		0x64, // 100
	}, buf.Bytes())
}

func TestEncoder_int16(t *testing.T) {
	// little endian
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteInt16(int16(-82), LE)
	enc.WriteInt16(int16(73), LE)

	assert.Equal(t, []byte{
		0xae, 0xff, // -82
		0x49, 0x00, // 73
	}, buf.Bytes())

	// big endian
	buf = new(bytes.Buffer)

	enc = NewBinEncoder(buf)
	enc.WriteInt16(int16(-82), BE)
	enc.WriteInt16(int16(73), BE)

	assert.Equal(t, []byte{
		0xff, 0xae, // -82
		0x00, 0x49, // 73
	}, buf.Bytes())
}

func TestEncoder_int32(t *testing.T) {
	// little endian
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteInt32(int32(-276132392), LE)
	enc.WriteInt32(int32(237391), LE)

	assert.Equal(t, []byte{
		0xd8, 0x8d, 0x8a, 0xef,
		0x4f, 0x9f, 0x3, 0x00,
	}, buf.Bytes())

	// big endian
	buf = new(bytes.Buffer)

	enc = NewBinEncoder(buf)
	enc.WriteInt32(int32(-276132392), BE)
	enc.WriteInt32(int32(237391), BE)

	assert.Equal(t, []byte{
		0xef, 0x8a, 0x8d, 0xd8,
		0x00, 0x3, 0x9f, 0x4f,
	}, buf.Bytes())
}

func TestEncoder_int64(t *testing.T) {
	// little endian
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteInt64(int64(-819823), LE)
	enc.WriteInt64(int64(72931), LE)

	assert.Equal(t, []byte{
		0x91, 0x7d, 0xf3, 0xff, 0xff, 0xff, 0xff, 0xff, //-819823
		0xe3, 0x1c, 0x1, 0x00, 0x00, 0x00, 0x00, 0x00, // 72931
	}, buf.Bytes())

	// big endian
	buf = new(bytes.Buffer)

	enc = NewBinEncoder(buf)
	enc.WriteInt64(int64(-819823), BE)
	enc.WriteInt64(int64(72931), BE)

	assert.Equal(t, []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xf3, 0x7d, 0x91, //-819823
		0x00, 0x00, 0x00, 0x00, 0x00, 0x1, 0x1c, 0xe3, // 72931
	}, buf.Bytes())
}

func TestEncoder_uint8(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteByte(uint8(99))
	enc.WriteByte(uint8(100))

	assert.Equal(t, []byte{
		0x63, // 99
		0x64, // 100
	}, buf.Bytes())
}

func TestEncoder_uint16(t *testing.T) {
	// little endian
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteUint16(uint16(82), LE)
	enc.WriteUint16(uint16(73), LE)

	assert.Equal(t, []byte{
		0x52, 0x00, // 82
		0x49, 0x00, // 73
	}, buf.Bytes())

	// big endian
	buf = new(bytes.Buffer)

	enc = NewBinEncoder(buf)
	enc.WriteUint16(uint16(82), BE)
	enc.WriteUint16(uint16(73), BE)

	assert.Equal(t, []byte{
		0x00, 0x52, // 82
		0x00, 0x49, // 73
	}, buf.Bytes())
}

func TestEncoder_uint32(t *testing.T) {
	// little endian
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteUint32(uint32(276132392), LE)
	enc.WriteUint32(uint32(237391), LE)

	assert.Equal(t, []byte{
		0x28, 0x72, 0x75, 0x10, // 276132392 as LE
		0x4f, 0x9f, 0x03, 0x00, // 237391 as LE
	}, buf.Bytes())

	// big endian
	buf = new(bytes.Buffer)

	enc = NewBinEncoder(buf)
	enc.WriteUint32(uint32(276132392), BE)
	enc.WriteUint32(uint32(237391), BE)

	assert.Equal(t, []byte{
		0x10, 0x75, 0x72, 0x28, // 276132392 as LE
		0x00, 0x03, 0x9f, 0x4f, // 237391 as LE
	}, buf.Bytes())
}

func TestEncoder_uint64(t *testing.T) {
	// little endian
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteUint64(uint64(819823), LE)
	enc.WriteUint64(uint64(72931), LE)

	assert.Equal(t, []byte{
		0x6f, 0x82, 0x0c, 0x00, 0x00, 0x00, 0x00, 0x00, // 819823
		0xe3, 0x1c, 0x1, 0x00, 0x00, 0x00, 0x00, 0x00, // 72931
	}, buf.Bytes())

	// big endian
	buf = new(bytes.Buffer)

	enc = NewBinEncoder(buf)
	enc.WriteUint64(uint64(819823), BE)
	enc.WriteUint64(uint64(72931), BE)

	assert.Equal(t, []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x0c, 0x82, 0x6f, // 819823
		0x00, 0x00, 0x00, 0x00, 0x00, 0x1, 0x1c, 0xe3, // 72931
	}, buf.Bytes())
}

func TestEncoder_float32(t *testing.T) {
	// little endian
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteFloat32(float32(1.32), LE)
	enc.WriteFloat32(float32(-3.21), LE)

	assert.Equal(t, []byte{
		0xc3, 0xf5, 0xa8, 0x3f,
		0xa4, 0x70, 0x4d, 0xc0,
	}, buf.Bytes())

	// big endian
	buf = new(bytes.Buffer)

	enc = NewBinEncoder(buf)
	enc.WriteFloat32(float32(1.32), BE)
	enc.WriteFloat32(float32(-3.21), BE)
	assert.Equal(t, []byte{
		0x3f, 0xa8, 0xf5, 0xc3,
		0xc0, 0x4d, 0x70, 0xa4,
	}, buf.Bytes())
}

func TestEncoder_float64(t *testing.T) {
	// little endian
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteFloat64(float64(-62.23), LE)
	enc.WriteFloat64(float64(23.239), LE)
	enc.WriteFloat64(float64(math.Inf(1)), LE)
	enc.WriteFloat64(float64(math.Inf(-1)), LE)

	assert.Equal(t, []byte{
		0x3d, 0x0a, 0xd7, 0xa3, 0x70, 0x1d, 0x4f, 0xc0,
		0x77, 0xbe, 0x9f, 0x1a, 0x2f, 0x3d, 0x37, 0x40,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x7f,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0xff,
	}, buf.Bytes())

	// big endian
	buf = new(bytes.Buffer)

	enc = NewBinEncoder(buf)
	enc.WriteFloat64(float64(-62.23), BE)
	enc.WriteFloat64(float64(23.239), BE)
	enc.WriteFloat64(float64(math.Inf(1)), BE)
	enc.WriteFloat64(float64(math.Inf(-1)), BE)

	assert.Equal(t, []byte{
		0xc0, 0x4f, 0x1d, 0x70, 0xa3, 0xd7, 0x0a, 0x3d,
		0x40, 0x37, 0x3d, 0x2f, 0x1a, 0x9f, 0xbe, 0x77,
		0x7f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xff, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}, buf.Bytes())
}

func TestEncoder_string(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteString("123")
	enc.WriteString("")
	enc.WriteString("abc")

	assert.Equal(t, []byte{
		0x03, 0x31, 0x32, 0x33, // "123"
		0x00,                   // ""
		0x03, 0x61, 0x62, 0x63, // "abc
	}, buf.Bytes())
}

func TestEncoder_byte(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteByte(0)
	enc.WriteByte(1)

	assert.Equal(t, []byte{
		0x00, 0x01,
	}, buf.Bytes())
}

func TestEncoder_bool(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteBool(true)
	enc.WriteBool(false)

	assert.Equal(t, []byte{
		0x01, 0x00,
	}, buf.Bytes())
}

func TestEncoder_ByteArray(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteBytes([]byte{1, 2, 3}, true)
	enc.WriteBytes([]byte{4, 5, 6}, true)
	enc.WriteBytes([]byte{7, 8}, false)

	assert.Equal(t, []byte{
		0x03, 0x01, 0x02, 0x03,
		0x03, 0x04, 0x05, 0x06,
		0x07, 0x08,
	}, buf.Bytes())

	bufB := new(bytes.Buffer)

	enc = NewBinEncoder(bufB)
	enc.Encode([]byte{1, 2, 3})

	assert.Equal(t, []byte{
		0x03, 0x01, 0x02, 0x03,
	}, bufB.Bytes())
}

func TestEncode_Array(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.Encode([3]byte{1, 2, 4})

	assert.Equal(t,
		[]byte{1, 2, 4},
		buf.Bytes(),
	)
}

func Test_OptionalPointerToPrimitiveType(t *testing.T) {
	type test struct {
		ID *Uint64 `bin:"optional"`
	}

	expect := []byte{0x00, 0x00, 0x00, 0x00}

	out, err := MarshalBin(test{ID: nil})
	require.NoError(t, err)
	assert.Equal(t, expect, out)

	id := Uint64(0)
	out, err = MarshalBin(test{ID: &id})
	require.NoError(t, err)
	assert.Equal(t, []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, out)

	id = Uint64(10)
	out, err = MarshalBin(test{ID: &id})
	require.NoError(t, err)

	assert.Equal(t, []byte{0x1, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, out)
}

func TestEncoder_Uint128(t *testing.T) {
	// little endian
	u := Uint128{
		Lo: 7,
		Hi: 9,
	}

	buf := new(bytes.Buffer)

	enc := NewBinEncoder(buf)
	enc.WriteUint128(u, LE)

	assert.Equal(t, []byte{
		0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}, buf.Bytes())

	// big endian
	buf = new(bytes.Buffer)

	enc = NewBinEncoder(buf)
	enc.WriteUint128(u, BE)

	assert.Equal(t, []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x07,
	}, buf.Bytes())
}

func TestEncoder_BinaryStruct(t *testing.T) {
	s := &binaryTestStruct{
		F1:  "abc",
		F2:  -75,
		F3:  99,
		F4:  -231,
		F5:  999,
		F6:  -13231,
		F7:  99999,
		F8:  -23.13,
		F9:  3.92,
		F10: []string{"def", "789"},
		F11: [2]string{"foo", "bar"},
		F12: 0xff,
		F13: []byte{1, 2, 3, 4, 5},
		F14: true,
		F15: Int64(-23),
		F16: Uint64(23),
		F17: JSONFloat64(3.14),
		F18: Uint128{
			Lo: 10,
			Hi: 82,
		},
		F19: Int128{
			Lo: 7,
			Hi: 3,
		},
		F20: Float128{
			Lo: 10,
			Hi: 82,
		},
		F21: Varuint32(999),
		F22: Varint32(-999),
		F23: Bool(true),
		F24: HexBytes([]byte{1, 2, 3, 4, 5}),
	}

	buf := new(bytes.Buffer)
	enc := NewBinEncoder(buf)
	err := enc.Encode(s)
	assert.NoError(t, err)

	assert.Equal(t,
		"0300000000000000616263b5ff630019ffffffe703000051ccffffffffffff9f860100000000003d0ab9c15c8fc2f5285c0f4002030000000000000064656603000000000000003738390300000000000000666f6f0300000000000000626172ff05010203040501e9ffffffffffffff17000000000000001f85eb51b81e09400a000000000000005200000000000000070000000000000003000000000000000a000000000000005200000000000000e707cd0f01050102030405",
		hex.EncodeToString(buf.Bytes()),
	)
}

func TestEncoder_BinaryTestStructWithTags(t *testing.T) {
	s := &binaryTestStructWithTags{
		F1:  "abc",
		F2:  -75,
		F3:  99,
		F4:  -231,
		F5:  999,
		F6:  -13231,
		F7:  99999,
		F8:  -23.13,
		F9:  3.92,
		F10: true,
		F12: (*[]int64)(&[]int64{99, 33}),
	}

	expected := []byte{
		255, 181, // F2
		0, 99, // F3
		255, 255, 255, 25, // F4
		0, 0, 3, 231, // F5
		255, 255, 255, 255, 255, 255, 204, 81, // F6
		0, 0, 0, 0, 0, 1, 134, 159, // F7
		193, 185, 10, 61, // F8
		64, 15, 92, 40, 245, 194, 143, 92, // F9
		1, // F10

		0, 0, 0, 0, // F11 is optional, and NOT SET (meaning uint32(0))

		1, 0, 0, 0, // F12 is optional, and IS SET (meaning uint32(1))
		2, // F12 is a slice, and the len is encoded as WriteUVarInt)
		99, 0, 0, 0, 0, 0, 0, 0,
		33, 0, 0, 0, 0, 0, 0, 0,
	}
	{
		buf := new(bytes.Buffer)
		enc := NewBinEncoder(buf)
		{
			err := enc.WriteInt16(s.F2, binary.BigEndian) // [255, 181](len=2)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteUint16(s.F3, binary.BigEndian) // [0, 99](len=2)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteInt32(s.F4, binary.BigEndian) // [255, 255, 255, 25](len=4)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteUint32(s.F5, binary.BigEndian) // [0, 0, 3, 231](len=4)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteInt64(s.F6, binary.BigEndian) // [255, 255, 255, 255, 255, 255, 204, 81](len=8)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteUint64(s.F7, binary.BigEndian) // [0, 0, 0, 0, 0, 1, 134, 159](len=8)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteFloat32(s.F8, binary.BigEndian) // [193, 185, 10, 61](len=4)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteFloat64(s.F9, binary.BigEndian) // [64, 15, 92, 40, 245, 194, 143, 92](len=8)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteBool(s.F10) // [1](len=1)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteUint32(0, binary.LittleEndian) // [0, 0, 0, 0](len=4)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteUint32(1, binary.LittleEndian) // [1, 0, 0, 0](len=4)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteUVarInt(2) // [2](len=1)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteInt64((*s.F12)[0], binary.LittleEndian) // [99, 0, 0, 0, 0, 0, 0, 0](len=8)
			if err != nil {
				panic(err)
			}
		}
		{
			err := enc.WriteInt64((*s.F12)[1], binary.LittleEndian) // [33, 0, 0, 0, 0, 0, 0, 0](len=8)
			if err != nil {
				panic(err)
			}
		}

		assert.Equal(t,
			expected,
			buf.Bytes(),
			FormatByteSlice(buf.Bytes()),
		)
	}

	buf := new(bytes.Buffer)
	enc := NewBinEncoder(buf)
	err := enc.Encode(s)
	assert.NoError(t, err)

	assert.Equal(t,
		expected,
		buf.Bytes(),
		FormatByteSlice(buf.Bytes()),
	)
}

func TestEncoder_InterfaceNil(t *testing.T) {
	var foo interface{}
	foo = nil
	buf := new(bytes.Buffer)
	enc := NewBinEncoder(buf)
	err := enc.Encode(foo)
	assert.NoError(t, err)
}

func TestByteArrays(t *testing.T) {
	{
		var buf bytes.Buffer
		enc := NewBinEncoder(&buf)
		err := enc.Encode([3]byte{1, 2, 3})
		assert.NoError(t, err)
		assert.Equal(t, []byte{1, 2, 3}, buf.Bytes())
	}
	{
		var buf bytes.Buffer
		enc := NewBorshEncoder(&buf)
		err := enc.Encode([3]byte{1, 2, 3})
		assert.NoError(t, err)
		assert.Equal(t, []byte{1, 2, 3}, buf.Bytes())
	}
}

func TestUintArrays(t *testing.T) {
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)
			err := enc.Encode([3]uint8{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t, []byte{1, 2, 3}, buf.Bytes())
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)
			err := enc.Encode([3]uint8{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t, []byte{1, 2, 3}, buf.Bytes())
		}
	}
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)
			err := enc.Encode([3]uint16{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t, concatByteSlices(
				[]byte{1, 0, 2, 0, 3, 0},
			), buf.Bytes())
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)
			err := enc.Encode([3]uint16{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t, concatByteSlices(
				[]byte{1, 0, 2, 0, 3, 0},
			), buf.Bytes())
		}
	}
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)
			err := enc.Encode([3]uint32{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					[]byte{1, 0, 0, 0},
					[]byte{2, 0, 0, 0},
					[]byte{3, 0, 0, 0},
				), buf.Bytes())
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)
			err := enc.Encode([3]uint32{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					[]byte{1, 0, 0, 0},
					[]byte{2, 0, 0, 0},
					[]byte{3, 0, 0, 0},
				), buf.Bytes())
		}
	}
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)
			err := enc.Encode([3]uint64{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					[]byte{1, 0, 0, 0, 0, 0, 0, 0},
					[]byte{2, 0, 0, 0, 0, 0, 0, 0},
					[]byte{3, 0, 0, 0, 0, 0, 0, 0},
				), buf.Bytes())
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)
			err := enc.Encode([3]uint64{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					[]byte{1, 0, 0, 0, 0, 0, 0, 0},
					[]byte{2, 0, 0, 0, 0, 0, 0, 0},
					[]byte{3, 0, 0, 0, 0, 0, 0, 0},
				), buf.Bytes())
		}
	}
}

func TestUintSlices(t *testing.T) {
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)
			err := enc.Encode([]uint8{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t, concatByteSlices(
				// length:
				[]byte{3},
				// data:
				[]byte{1, 2, 3},
			), buf.Bytes())
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)
			err := enc.Encode([]uint8{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					// length:
					[]byte{3, 0, 0, 0},
					// data:
					[]byte{1, 2, 3},
				), buf.Bytes())
		}
	}
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)
			err := enc.Encode([]uint16{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t, concatByteSlices(
				// length:
				[]byte{3},
				// data:
				[]byte{1, 0, 2, 0, 3, 0},
			), buf.Bytes())
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)
			err := enc.Encode([]uint16{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t, concatByteSlices(
				// length:
				[]byte{3, 0, 0, 0},
				// data:
				[]byte{1, 0, 2, 0, 3, 0},
			), buf.Bytes())
		}
	}
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)
			err := enc.Encode([]uint32{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					// length:
					[]byte{3},
					// data:
					[]byte{1, 0, 0, 0},
					[]byte{2, 0, 0, 0},
					[]byte{3, 0, 0, 0},
				), buf.Bytes())
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)
			err := enc.Encode([]uint32{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					// length:
					[]byte{3, 0, 0, 0},
					// data:
					[]byte{1, 0, 0, 0},
					[]byte{2, 0, 0, 0},
					[]byte{3, 0, 0, 0},
				), buf.Bytes())
		}
	}
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)
			err := enc.Encode([]uint64{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					// length:
					[]byte{3},
					// data:
					[]byte{1, 0, 0, 0, 0, 0, 0, 0},
					[]byte{2, 0, 0, 0, 0, 0, 0, 0},
					[]byte{3, 0, 0, 0, 0, 0, 0, 0},
				), buf.Bytes())
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)
			err := enc.Encode([]uint64{1, 2, 3})
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					// length:
					[]byte{3, 0, 0, 0},
					// data:
					[]byte{1, 0, 0, 0, 0, 0, 0, 0},
					[]byte{2, 0, 0, 0, 0, 0, 0, 0},
					[]byte{3, 0, 0, 0, 0, 0, 0, 0},
				), buf.Bytes())
		}
	}
}

func Test_writeArrayOfBytes(t *testing.T) {
	{
		var buf bytes.Buffer
		enc := NewBinEncoder(&buf)

		arr := [3]byte{1, 2, 3}
		l := len(arr)

		err := reflect_writeArrayOfBytes(enc, l, reflect.ValueOf(arr))
		assert.NoError(t, err)
		assert.Equal(t, arr[:], buf.Bytes())
	}
	{
		var buf bytes.Buffer
		enc := NewBinEncoder(&buf)

		arr := [10]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		l := len(arr)

		err := reflect_writeArrayOfBytes(enc, l, reflect.ValueOf(arr))
		assert.NoError(t, err)
		assert.Equal(t, arr[:], buf.Bytes())
	}
	{
		var buf bytes.Buffer
		enc := NewBinEncoder(&buf)

		arr := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
		l := len(arr)

		err := reflect_writeArrayOfBytes(enc, l, reflect.ValueOf(arr))
		assert.NoError(t, err)
		assert.Equal(t, arr[:], buf.Bytes())
	}
}

func Test_writeArrayOfUint16(t *testing.T) {
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)

			arr := [3]uint16{1, 2, 3}
			l := len(arr)

			err := reflect_writeArrayOfUint16(enc, l, reflect.ValueOf(arr), LE)
			assert.NoError(t, err)
			assert.Equal(t, concatByteSlices(
				[]byte{1, 0, 2, 0, 3, 0},
			), buf.Bytes())
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)

			arr := [3]uint16{1, 2, 3}
			l := len(arr)

			err := reflect_writeArrayOfUint16(enc, l, reflect.ValueOf(arr), LE)
			assert.NoError(t, err)
			assert.Equal(t, concatByteSlices(
				[]byte{1, 0, 2, 0, 3, 0},
			), buf.Bytes())
		}
	}
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)

			arr := []uint16{1, 2, 3}
			l := len(arr)

			err := reflect_writeArrayOfUint16(enc, l, reflect.ValueOf(arr), LE)
			assert.NoError(t, err)
			assert.Equal(t, concatByteSlices(
				[]byte{1, 0, 2, 0, 3, 0},
			), buf.Bytes())
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)

			arr := []uint16{1, 2, 3}
			l := len(arr)

			err := reflect_writeArrayOfUint16(enc, l, reflect.ValueOf(arr), LE)
			assert.NoError(t, err)
			assert.Equal(t, concatByteSlices(
				[]byte{1, 0, 2, 0, 3, 0},
			), buf.Bytes())
		}
	}
}

func Test_writeArrayOfUint32(t *testing.T) {
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)

			arr := [3]uint32{1, 2, 3}
			l := len(arr)

			err := reflect_writeArrayOfUint32(enc, l, reflect.ValueOf(arr), LE)
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					[]byte{1, 0, 0, 0},
					[]byte{2, 0, 0, 0},
					[]byte{3, 0, 0, 0},
				),
				buf.Bytes(),
			)
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)

			arr := [3]uint32{1, 2, 3}
			l := len(arr)

			err := reflect_writeArrayOfUint32(enc, l, reflect.ValueOf(arr), LE)
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					[]byte{1, 0, 0, 0},
					[]byte{2, 0, 0, 0},
					[]byte{3, 0, 0, 0},
				),
				buf.Bytes(),
			)
		}
	}
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)

			arr := []uint32{1, 2, 3}
			l := len(arr)

			err := reflect_writeArrayOfUint32(enc, l, reflect.ValueOf(arr), LE)
			assert.NoError(t, err)

			assert.Equal(t,
				concatByteSlices(
					[]byte{1, 0, 0, 0},
					[]byte{2, 0, 0, 0},
					[]byte{3, 0, 0, 0},
				),
				buf.Bytes(),
			)
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)

			arr := []uint32{1, 2, 3}
			l := len(arr)

			err := reflect_writeArrayOfUint32(enc, l, reflect.ValueOf(arr), LE)
			assert.NoError(t, err)

			assert.Equal(t,
				concatByteSlices(
					[]byte{1, 0, 0, 0},
					[]byte{2, 0, 0, 0},
					[]byte{3, 0, 0, 0},
				),
				buf.Bytes(),
			)
		}
	}
	{
		var buf bytes.Buffer
		enc := NewBinEncoder(&buf)

		arr := [10]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		l := len(arr)

		err := reflect_writeArrayOfUint32(enc, l, reflect.ValueOf(arr), LE)
		assert.NoError(t, err)
		assert.Equal(t,
			concatByteSlices(
				[]byte{1, 0, 0, 0},
				[]byte{2, 0, 0, 0},
				[]byte{3, 0, 0, 0},
				[]byte{4, 0, 0, 0},
				[]byte{5, 0, 0, 0},
				[]byte{6, 0, 0, 0},
				[]byte{7, 0, 0, 0},
				[]byte{8, 0, 0, 0},
				[]byte{9, 0, 0, 0},
				[]byte{10, 0, 0, 0},
			),
			buf.Bytes())
	}
	{
		var buf bytes.Buffer
		enc := NewBinEncoder(&buf)

		arr := [32]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
		l := len(arr)

		err := reflect_writeArrayOfUint32(enc, l, reflect.ValueOf(arr), LE)
		assert.NoError(t, err)
		assert.Equal(t,
			concatByteSlices(
				[]byte{1, 0, 0, 0},
				[]byte{2, 0, 0, 0},
				[]byte{3, 0, 0, 0},
				[]byte{4, 0, 0, 0},
				[]byte{5, 0, 0, 0},
				[]byte{6, 0, 0, 0},
				[]byte{7, 0, 0, 0},
				[]byte{8, 0, 0, 0},
				[]byte{9, 0, 0, 0},
				[]byte{10, 0, 0, 0},
				[]byte{11, 0, 0, 0},
				[]byte{12, 0, 0, 0},
				[]byte{13, 0, 0, 0},
				[]byte{14, 0, 0, 0},
				[]byte{15, 0, 0, 0},
				[]byte{16, 0, 0, 0},
				[]byte{17, 0, 0, 0},
				[]byte{18, 0, 0, 0},
				[]byte{19, 0, 0, 0},
				[]byte{20, 0, 0, 0},
				[]byte{21, 0, 0, 0},
				[]byte{22, 0, 0, 0},
				[]byte{23, 0, 0, 0},
				[]byte{24, 0, 0, 0},
				[]byte{25, 0, 0, 0},
				[]byte{26, 0, 0, 0},
				[]byte{27, 0, 0, 0},
				[]byte{28, 0, 0, 0},
				[]byte{29, 0, 0, 0},
				[]byte{30, 0, 0, 0},
				[]byte{31, 0, 0, 0},
				[]byte{32, 0, 0, 0},
			),
			buf.Bytes())
	}
}

func Test_writeArrayOfUint64(t *testing.T) {
	{
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)

			arr := [3]uint64{1, 2, 3}
			l := len(arr)

			err := reflect_writeArrayOfUint64(enc, l, reflect.ValueOf(arr), LE)
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					[]byte{1, 0, 0, 0, 0, 0, 0, 0},
					[]byte{2, 0, 0, 0, 0, 0, 0, 0},
					[]byte{3, 0, 0, 0, 0, 0, 0, 0},
				),
				buf.Bytes(),
			)
		}
		{
			var buf bytes.Buffer
			enc := NewBinEncoder(&buf)

			arr := []uint64{1, 2, 3}
			l := len(arr)

			err := reflect_writeArrayOfUint64(enc, l, reflect.ValueOf(arr), LE)
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					[]byte{1, 0, 0, 0, 0, 0, 0, 0},
					[]byte{2, 0, 0, 0, 0, 0, 0, 0},
					[]byte{3, 0, 0, 0, 0, 0, 0, 0},
				),
				buf.Bytes(),
			)
		}
	}
	{
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)

			arr := [3]uint64{1, 2, 3}
			l := len(arr)

			err := reflect_writeArrayOfUint64(enc, l, reflect.ValueOf(arr), LE)
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					[]byte{1, 0, 0, 0, 0, 0, 0, 0},
					[]byte{2, 0, 0, 0, 0, 0, 0, 0},
					[]byte{3, 0, 0, 0, 0, 0, 0, 0},
				),
				buf.Bytes(),
			)
		}
		{
			var buf bytes.Buffer
			enc := NewBorshEncoder(&buf)

			arr := []uint64{1, 2, 3}
			l := len(arr)

			err := reflect_writeArrayOfUint64(enc, l, reflect.ValueOf(arr), LE)
			assert.NoError(t, err)
			assert.Equal(t,
				concatByteSlices(
					[]byte{1, 0, 0, 0, 0, 0, 0, 0},
					[]byte{2, 0, 0, 0, 0, 0, 0, 0},
					[]byte{3, 0, 0, 0, 0, 0, 0, 0},
				),
				buf.Bytes(),
			)
		}
	}
	{
		var buf bytes.Buffer

		enc := NewBinEncoder(&buf)
		arr := [64]uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64}
		l := len(arr)

		err := reflect_writeArrayOfUint64(enc, l, reflect.ValueOf(arr), LE)
		assert.NoError(t, err)

		assert.Equal(t,
			concatByteSlices(
				[]byte{1, 0, 0, 0, 0, 0, 0, 0},
				[]byte{2, 0, 0, 0, 0, 0, 0, 0},
				[]byte{3, 0, 0, 0, 0, 0, 0, 0},
				[]byte{4, 0, 0, 0, 0, 0, 0, 0},
				[]byte{5, 0, 0, 0, 0, 0, 0, 0},
				[]byte{6, 0, 0, 0, 0, 0, 0, 0},
				[]byte{7, 0, 0, 0, 0, 0, 0, 0},
				[]byte{8, 0, 0, 0, 0, 0, 0, 0},
				[]byte{9, 0, 0, 0, 0, 0, 0, 0},
				[]byte{10, 0, 0, 0, 0, 0, 0, 0},
				[]byte{11, 0, 0, 0, 0, 0, 0, 0},
				[]byte{12, 0, 0, 0, 0, 0, 0, 0},
				[]byte{13, 0, 0, 0, 0, 0, 0, 0},
				[]byte{14, 0, 0, 0, 0, 0, 0, 0},
				[]byte{15, 0, 0, 0, 0, 0, 0, 0},
				[]byte{16, 0, 0, 0, 0, 0, 0, 0},
				[]byte{17, 0, 0, 0, 0, 0, 0, 0},
				[]byte{18, 0, 0, 0, 0, 0, 0, 0},
				[]byte{19, 0, 0, 0, 0, 0, 0, 0},
				[]byte{20, 0, 0, 0, 0, 0, 0, 0},
				[]byte{21, 0, 0, 0, 0, 0, 0, 0},
				[]byte{22, 0, 0, 0, 0, 0, 0, 0},
				[]byte{23, 0, 0, 0, 0, 0, 0, 0},
				[]byte{24, 0, 0, 0, 0, 0, 0, 0},
				[]byte{25, 0, 0, 0, 0, 0, 0, 0},
				[]byte{26, 0, 0, 0, 0, 0, 0, 0},
				[]byte{27, 0, 0, 0, 0, 0, 0, 0},
				[]byte{28, 0, 0, 0, 0, 0, 0, 0},
				[]byte{29, 0, 0, 0, 0, 0, 0, 0},
				[]byte{30, 0, 0, 0, 0, 0, 0, 0},
				[]byte{31, 0, 0, 0, 0, 0, 0, 0},
				[]byte{32, 0, 0, 0, 0, 0, 0, 0},
				[]byte{33, 0, 0, 0, 0, 0, 0, 0},
				[]byte{34, 0, 0, 0, 0, 0, 0, 0},
				[]byte{35, 0, 0, 0, 0, 0, 0, 0},
				[]byte{36, 0, 0, 0, 0, 0, 0, 0},
				[]byte{37, 0, 0, 0, 0, 0, 0, 0},
				[]byte{38, 0, 0, 0, 0, 0, 0, 0},
				[]byte{39, 0, 0, 0, 0, 0, 0, 0},
				[]byte{40, 0, 0, 0, 0, 0, 0, 0},
				[]byte{41, 0, 0, 0, 0, 0, 0, 0},
				[]byte{42, 0, 0, 0, 0, 0, 0, 0},
				[]byte{43, 0, 0, 0, 0, 0, 0, 0},
				[]byte{44, 0, 0, 0, 0, 0, 0, 0},
				[]byte{45, 0, 0, 0, 0, 0, 0, 0},
				[]byte{46, 0, 0, 0, 0, 0, 0, 0},
				[]byte{47, 0, 0, 0, 0, 0, 0, 0},
				[]byte{48, 0, 0, 0, 0, 0, 0, 0},
				[]byte{49, 0, 0, 0, 0, 0, 0, 0},
				[]byte{50, 0, 0, 0, 0, 0, 0, 0},
				[]byte{51, 0, 0, 0, 0, 0, 0, 0},
				[]byte{52, 0, 0, 0, 0, 0, 0, 0},
				[]byte{53, 0, 0, 0, 0, 0, 0, 0},
				[]byte{54, 0, 0, 0, 0, 0, 0, 0},
				[]byte{55, 0, 0, 0, 0, 0, 0, 0},
				[]byte{56, 0, 0, 0, 0, 0, 0, 0},
				[]byte{57, 0, 0, 0, 0, 0, 0, 0},
				[]byte{58, 0, 0, 0, 0, 0, 0, 0},
				[]byte{59, 0, 0, 0, 0, 0, 0, 0},
				[]byte{60, 0, 0, 0, 0, 0, 0, 0},
				[]byte{61, 0, 0, 0, 0, 0, 0, 0},
				[]byte{62, 0, 0, 0, 0, 0, 0, 0},
				[]byte{63, 0, 0, 0, 0, 0, 0, 0},
				[]byte{64, 0, 0, 0, 0, 0, 0, 0},
			),
			buf.Bytes())
	}
}
