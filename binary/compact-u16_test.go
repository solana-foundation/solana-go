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
	"bytes"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompactU16(t *testing.T) {
	candidates := []int{0, 1, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 100, 1000, 10000, math.MaxUint16 - 1, math.MaxUint16}
	for _, val := range candidates {
		if val < 0 || val > math.MaxUint16 {
			panic("value too large")
		}
		buf := make([]byte, 0)
		require.NoError(t, EncodeCompactU16Length(&buf, val))

		buf = append(buf, []byte("hello world")...)
		decoded, _, err := DecodeCompactU16(buf)
		require.NoError(t, err)

		require.Equal(t, val, decoded)
	}
	for _, val := range candidates {
		buf := make([]byte, 0)
		EncodeCompactU16Length(&buf, val)

		buf = append(buf, []byte("hello world")...)
		{
			decoded, err := DecodeCompactU16LengthFromByteReader(bytes.NewReader(buf))
			require.NoError(t, err)
			require.Equal(t, val, decoded)
		}
		{
			decoded, _, err := DecodeCompactU16(buf)
			require.NoError(t, err)
			require.Equal(t, val, decoded)
		}
	}
	{
		// now test all from 0 to 0xffff
		for i := 0; i < math.MaxUint16; i++ {
			buf := make([]byte, 0)
			EncodeCompactU16Length(&buf, i)

			buf = append(buf, []byte("hello world")...)
			{
				decoded, err := DecodeCompactU16LengthFromByteReader(bytes.NewReader(buf))
				require.NoError(t, err)
				require.Equal(t, i, decoded)
			}
			{
				decoded, _, err := DecodeCompactU16(buf)
				require.NoError(t, err)
				require.Equal(t, i, decoded)
			}
		}
	}
}

func BenchmarkCompactU16(b *testing.B) {
	// generate 1000 random values
	candidates := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		candidates[i] = i
	}

	buf := make([]byte, 0)
	EncodeCompactU16Length(&buf, math.MaxUint16)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = DecodeCompactU16(buf)
	}
}

func BenchmarkCompactU16Encode(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0)
		EncodeCompactU16Length(&buf, math.MaxUint16)
	}
}

func BenchmarkCompactU16Reader(b *testing.B) {
	// generate 1000 random values
	candidates := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		candidates[i] = i
	}

	buf := make([]byte, 0)
	EncodeCompactU16Length(&buf, math.MaxUint16)

	reader := NewBorshDecoder(buf)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		out, _ := reader.ReadCompactU16()
		if out != math.MaxUint16 {
			panic("not equal")
		}
		reader.SetPosition(0)
	}
}

func encode_len(len uint16) []byte {
	buf := make([]byte, 0)
	err := EncodeCompactU16Length(&buf, int(len))
	if err != nil {
		panic(err)
	}
	return buf
}

func assert_len_encoding(t *testing.T, len uint16, buf []byte) {
	require.Equal(t, encode_len(len), buf, "unexpected usize encoding")
	decoded, _, err := DecodeCompactU16(buf)
	require.NoError(t, err)
	require.Equal(t, int(len), decoded)
	{
		// now try with a reader
		reader := bytes.NewReader(buf)
		out, _ := DecodeCompactU16LengthFromByteReader(reader)
		require.Equal(t, int(len), out)
	}
}

func TestShortVecEncodeLen(t *testing.T) {
	assert_len_encoding(t, 0x0, []byte{0x0})
	assert_len_encoding(t, 0x7f, []byte{0x7f})
	assert_len_encoding(t, 0x80, []byte{0x80, 0x01})
	assert_len_encoding(t, 0xff, []byte{0xff, 0x01})
	assert_len_encoding(t, 0x100, []byte{0x80, 0x02})
	assert_len_encoding(t, 0x7fff, []byte{0xff, 0xff, 0x01})
	assert_len_encoding(t, 0xffff, []byte{0xff, 0xff, 0x03})
}

func assert_good_deserialized_value(t *testing.T, value uint16, buf []byte) {
	decoded, _, err := DecodeCompactU16(buf)
	require.NoError(t, err)
	require.Equal(t, int(value), decoded)
	{
		// now try with a reader
		reader := bytes.NewReader(buf)
		out, _ := DecodeCompactU16LengthFromByteReader(reader)
		require.Equal(t, int(value), out)
	}
}

func assert_bad_deserialized_value(t *testing.T, buf []byte) {
	_, _, err := DecodeCompactU16(buf)
	require.Error(t, err, "expected an error for bytes: %v", buf)
	{
		// now try with a reader
		reader := bytes.NewReader(buf)
		_, err := DecodeCompactU16LengthFromByteReader(reader)
		require.Error(t, err, "expected an error for bytes: %v", buf)
	}
}

func TestDeserialize(t *testing.T) {
	assert_good_deserialized_value(t, 0x0000, []byte{0x00})
	assert_good_deserialized_value(t, 0x007f, []byte{0x7f})
	assert_good_deserialized_value(t, 0x0080, []byte{0x80, 0x01})
	assert_good_deserialized_value(t, 0x00ff, []byte{0xff, 0x01})
	assert_good_deserialized_value(t, 0x0100, []byte{0x80, 0x02})
	assert_good_deserialized_value(t, 0x07ff, []byte{0xff, 0x0f})
	assert_good_deserialized_value(t, 0x3fff, []byte{0xff, 0x7f})
	assert_good_deserialized_value(t, 0x4000, []byte{0x80, 0x80, 0x01})
	assert_good_deserialized_value(t, 0xffff, []byte{0xff, 0xff, 0x03})

	// aliases
	// 0x0000
	assert_bad_deserialized_value(t, []byte{0x80, 0x00})
	assert_bad_deserialized_value(t, []byte{0x80, 0x80, 0x00})
	// 0x007f
	assert_bad_deserialized_value(t, []byte{0xff, 0x00})
	assert_bad_deserialized_value(t, []byte{0xff, 0x80, 0x00})
	// 0x0080
	assert_bad_deserialized_value(t, []byte{0x80, 0x81, 0x00})
	// 0x00ff
	assert_bad_deserialized_value(t, []byte{0xff, 0x81, 0x00})
	// 0x0100
	assert_bad_deserialized_value(t, []byte{0x80, 0x82, 0x00})
	// 0x07ff
	assert_bad_deserialized_value(t, []byte{0xff, 0x8f, 0x00})
	// 0x3fff
	assert_bad_deserialized_value(t, []byte{0xff, 0xff, 0x00})

	// too short
	assert_bad_deserialized_value(t, []byte{})
	assert_bad_deserialized_value(t, []byte{0x80})

	// too long
	assert_bad_deserialized_value(t, []byte{0x80, 0x80, 0x80, 0x00})

	// too large
	// 0x0001_0000
	assert_bad_deserialized_value(t, []byte{0x80, 0x80, 0x04})
	// 0x0001_8000
	assert_bad_deserialized_value(t, []byte{0x80, 0x80, 0x06})
}
