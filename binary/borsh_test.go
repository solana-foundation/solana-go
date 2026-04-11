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
	"fmt"
	"io"
	"math"
	"reflect"
	strings2 "strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type OptionalPointerFields struct {
	Good uint8
	Arr  *Arr `bin:"optional"`
}

type Arr []string

func TestOptionWithPointer(t *testing.T) {
	// nil (optional not present)
	{
		buf := new(bytes.Buffer)
		enc := NewBorshEncoder(buf)
		val := OptionalPointerFields{
			Good: 9,
			// Will be decoded as nil pointer.
			Arr: nil,
		}
		require.NoError(t, enc.Encode(val))
		require.Equal(t,
			concatByteSlices(
				[]byte{9},
				[]byte{0},
			),
			buf.Bytes())
		{
			dec := NewBorshDecoder(buf.Bytes())
			var got OptionalPointerFields
			require.NoError(t, dec.Decode(&got))
			require.Equal(t, val, got)
		}
	}
	// optional is present but has zero elements
	{
		buf := new(bytes.Buffer)
		enc := NewBorshEncoder(buf)
		val := OptionalPointerFields{
			Good: 9,
			// Will be decoded as pointer to nil Arr.
			Arr: &Arr{},
		}
		require.NoError(t, enc.Encode(val))
		require.Equal(t,
			concatByteSlices(
				[]byte{9},
				[]byte{1},
				[]byte{0, 0, 0, 0},
			),
			buf.Bytes(),
		)
		{
			dec := NewBorshDecoder(buf.Bytes())
			var got OptionalPointerFields
			require.NoError(t, dec.Decode(&got))
			// an empty slice is decoded as nil.
			po := (Arr)(nil)
			val.Arr = &po
			require.Equal(t,
				val, got)
		}
	}
	// optional is present and has elements
	{
		buf := new(bytes.Buffer)
		enc := NewBorshEncoder(buf)
		val := OptionalPointerFields{
			Good: 9,
			Arr:  &Arr{"foo"},
		}
		require.NoError(t, enc.Encode(val))
		require.Equal(t,
			concatByteSlices(
				[]byte{9},
				[]byte{1},
				[]byte{1, 0, 0, 0},

				[]byte{3, 0, 0, 0},
				[]byte("foo"),
			),
			buf.Bytes(),
		)
		{
			dec := NewBorshDecoder(buf.Bytes())
			var got OptionalPointerFields
			require.NoError(t, dec.Decode(&got))
			require.Equal(t, val, got)
		}
	}
}

type StructWithComplexPeculiarEnums struct {
	Complex2NotSet    ComplexEnumPointers
	Complex2PtrNotSet *ComplexEnumPointers

	// Complex2PtrOptionalSet *ComplexEnumPointers `bin:"optional"`
}

func TestBorsh_peculiarEnums(t *testing.T) {
	t.Skip()
	{
		// struct with peculiar complex enums:
		{
			// buf := new(bytes.Buffer)
			buf := NewWriteByWrite("")
			enc := NewBorshEncoder(buf)
			val := StructWithComplexPeculiarEnums{
				// If the enums are left empty, they won't serialize correctly.
			}
			require.NoError(t, enc.Encode(val))
			fmt.Println(buf.String())
			require.Equal(t,
				concatByteSlices(
					[]byte{0},
					[]byte{0, 0, 0, 0},
					[]byte{0, 0, 0, 0},

					[]byte{0},
					[]byte{0, 0, 0, 0},
					[]byte{0, 0, 0, 0},
				),
				buf.Bytes(),
			)

			{
				dec := NewBorshDecoder(buf.Bytes())
				var got StructWithComplexPeculiarEnums
				require.NoError(t, dec.Decode(&got))
				{
				}
				require.Equal(t, val, got)
			}
		}
	}
}

func TestBorsh_Encode(t *testing.T) {
	// ints:
	{
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := int8(33)
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{33}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got int8
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, val, got)
			}
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := int16(44)
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{44, 0}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got int16
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, val, got)
			}
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := int32(55)
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{55, 0, 0, 0}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got int32
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, val, got)
			}
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := int64(556)
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{0x2c, 0x2, 0, 0, 0, 0, 0, 0}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got int64
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, val, got)
			}
		}
		{
			// pointers to a basic type shall be encoded as values.
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := int64(556)
				require.NoError(t, enc.Encode(&val))
				require.Equal(t, []byte{0x2c, 0x2, 0, 0, 0, 0, 0, 0}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got int64
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, val, got)
				}
			}
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := int8(120)
				require.NoError(t, enc.Encode(&val))
				require.Equal(t, []byte{120}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got int8
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, val, got)
				}
			}
		}
		{
			// pointer to a nil value of a basic type shall be encoded as the zero value of that type:
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := new(int64)
				require.NoError(t, enc.Encode(val))
				require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got int64
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, *val, got)
				}
			}
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := new(int8)
				require.NoError(t, enc.Encode(val))
				require.Equal(t, []byte{0}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got int8
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, *val, got)
				}
			}
		}
	}
	// uints:
	{
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := uint8(33)
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{33}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got uint8
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, val, got)
			}
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := uint16(44)
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{44, 0}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got uint16
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, val, got)
			}
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := uint32(55)
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{55, 0, 0, 0}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got uint32
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, val, got)
			}
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := uint64(556)
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{0x2c, 0x2, 0, 0, 0, 0, 0, 0}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got uint64
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, val, got)
			}
		}
		{
			// pouinters to a basic type shall be encoded as values.
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := uint64(556)
				require.NoError(t, enc.Encode(&val))
				require.Equal(t, []byte{0x2c, 0x2, 0, 0, 0, 0, 0, 0}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got uint64
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, val, got)
				}
			}
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := uint8(120)
				require.NoError(t, enc.Encode(&val))
				require.Equal(t, []byte{120}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got uint8
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, val, got)
				}
			}
		}
		{
			// pointer to a nil value of a basic type shall be encoded as the zero value of that type:
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := new(uint64)
				require.NoError(t, enc.Encode(val))
				require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got uint64
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, *val, got)
				}
			}
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := new(uint8)
				require.NoError(t, enc.Encode(val))
				require.Equal(t, []byte{0}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got uint8
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, *val, got)
				}
			}
		}
	}
	{
		// bool
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			require.NoError(t, enc.Encode(true))
			require.Equal(t, []byte{1}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got bool
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, true, got)
			}
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			require.NoError(t, enc.Encode(false))
			require.Equal(t, []byte{0}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got bool
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, false, got)
			}
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := false
			require.NoError(t, enc.Encode(&val))
			require.Equal(t, []byte{0}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got bool
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, false, got)
			}
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := true
			require.NoError(t, enc.Encode(&val))
			require.Equal(t, []byte{1}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got bool
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, true, got)
			}
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := new(bool)
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{0}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got bool
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, false, got)
			}
		}
	}
	{
		// floats
		{
			// float32
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := float32(1.123)
				require.NoError(t, enc.Encode(val))
				require.Equal(t, []byte{0x77, 0xbe, 0x8f, 0x3f}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got float32
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, val, got)
				}
			}
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := float32(1.123)
				require.NoError(t, enc.Encode(&val))
				require.Equal(t, []byte{0x77, 0xbe, 0x8f, 0x3f}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got float32
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, val, got)
				}
			}
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := new(float32)
				require.NoError(t, enc.Encode(val))
				require.Equal(t, []byte{0, 0, 0, 0}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got float32
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, *val, got)
				}
			}
		}
		{
			// float64
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := float64(1.123)
				require.NoError(t, enc.Encode(val))
				require.Equal(t, []byte{0x2b, 0x87, 0x16, 0xd9, 0xce, 0xf7, 0xf1, 0x3f}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got float64
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, val, got)
				}
			}
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := float64(1.123)
				require.NoError(t, enc.Encode(&val))
				require.Equal(t, []byte{0x2b, 0x87, 0x16, 0xd9, 0xce, 0xf7, 0xf1, 0x3f}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got float64
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, val, got)
				}
			}
			{
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := new(float64)
				require.NoError(t, enc.Encode(val))
				require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0}, buf.Bytes())
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got float64
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, *val, got)
				}
			}
		}
	}
	{
		// string
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := string("hello world")
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{0xb, 0x0, 0x0, 0x0, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}, buf.Bytes())
			require.Equal(t, append([]byte{byte(len(val)), 0, 0, 0}, []byte(val)...), buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got string
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, val, got)
			}
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := string("hello world")
			require.NoError(t, enc.Encode(&val))
			require.Equal(t, []byte{0xb, 0x0, 0x0, 0x0, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}, buf.Bytes())
			require.Equal(t, append([]byte{byte(len(val)), 0, 0, 0}, []byte(val)...), buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got string
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, val, got)
			}
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := new(string)
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0}, buf.Bytes())
			require.Equal(t, append([]byte{0, 0, 0, 0}, []byte{}...), buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got string
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, *val, got)
			}
		}
	}
	{
		// interface
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			var val io.Reader
			require.NoError(t, enc.Encode(val))
			require.Equal(t, ([]byte)(nil), buf.Bytes())
		}
		{
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			var val io.Reader
			require.NoError(t, enc.Encode(&val))
			require.Equal(t, ([]byte)(nil), buf.Bytes())
		}
	}
	{
		// type that has `func (e CustomEncoding) MarshalWithEncoder(encoder *Encoder) error` method.
		// NOTE: the `MarshalWithEncoder` method MUST be on value (NOT on pointer).
		{
			// by value:
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := CustomEncoding{
				Prefix: byte('a'),
				Value:  33,
			}
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{33, 0, 0, 0, byte('a')}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got CustomEncoding
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, val, got)
			}
		}
		{
			// by pointer:
			buf := new(bytes.Buffer)
			enc := NewBorshEncoder(buf)
			val := &CustomEncoding{
				Prefix: byte('a'),
				Value:  33,
			}
			require.NoError(t, enc.Encode(val))
			require.Equal(t, []byte{33, 0, 0, 0, byte('a')}, buf.Bytes())
			{
				dec := NewBorshDecoder(buf.Bytes())
				var got CustomEncoding
				require.NoError(t, dec.Decode(&got))
				require.Equal(t, *val, got)
			}
		}
	}
	{
		// struct
		{
			// simple
			{
				// by value:
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := Struct{
					Foo: "hello",
					Bar: 33,
				}
				require.NoError(t, enc.Encode(val))
				require.Equal(t,
					concatByteSlices(
						[]byte{byte(len(val.Foo)), 0, 0, 0},
						[]byte(val.Foo),
						[]byte{33, 0, 0, 0},
					),
					buf.Bytes(),
				)
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got Struct
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, val, got)
				}
			}
			{
				// by pointer:
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := &Struct{
					Foo: "hello",
					Bar: 33,
				}
				require.NoError(t, enc.Encode(val))
				require.Equal(t,
					concatByteSlices(
						[]byte{byte(len(val.Foo)), 0, 0, 0},
						[]byte(val.Foo),
						[]byte{33, 0, 0, 0},
					),
					buf.Bytes(),
				)
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got Struct
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, *val, got)
				}
			}
		}
		{
			// with fields that are pointers
			hello := "hello"
			bar := uint32(33)
			{
				// by value:
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := StructWithPointerFields{
					Foo: &hello,
					Bar: &bar,
				}
				require.NoError(t, enc.Encode(val))
				require.Equal(t,
					concatByteSlices(
						[]byte{byte(len(*val.Foo)), 0, 0, 0},
						[]byte(*val.Foo),
						[]byte{33, 0, 0, 0},
					),
					buf.Bytes(),
				)
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got StructWithPointerFields
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, val, got)
				}
			}
			{
				// by pointer:
				buf := new(bytes.Buffer)
				enc := NewBorshEncoder(buf)
				val := &StructWithPointerFields{
					Foo: &hello,
					Bar: &bar,
				}
				require.NoError(t, enc.Encode(val))
				require.Equal(t,
					concatByteSlices(
						[]byte{byte(len(*val.Foo)), 0, 0, 0},
						[]byte(*val.Foo),
						[]byte{33, 0, 0, 0},
					),
					buf.Bytes(),
				)
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got StructWithPointerFields
					require.NoError(t, dec.Decode(&got))
					require.Equal(t, *val, got)
				}
			}
		}
		{
			// with optional fields
			{
				// buf := new(bytes.Buffer)
				fooRequired := "hello"
				fooOptional := "-world"
				barPointer := uint32(33)
				buf := NewWriteByWrite("")
				enc := NewBorshEncoder(buf)
				val := StructWithOptionalFields{
					FooRequired: &fooRequired,
					FooPointer:  &fooOptional,
					BarPointer:  &barPointer,
					FooValue:    "hi",
				}
				require.NoError(t, enc.Encode(val))
				// fmt.Println(buf.String())
				require.Equal(t,
					concatByteSlices(
						// .FooRequired
						[]byte{5, 0, 0, 0},
						[]byte(*val.FooRequired),

						// .BarRequiredNotSet
						[]byte{0, 0, 0, 0},

						// .FooPointer (optional)
						[]byte{1},
						[]byte{6, 0, 0, 0},
						[]byte(*val.FooPointer),

						// .FooPointerNotSet (optional)
						[]byte{0},

						// .BarPointer (optional)
						[]byte{1},
						[]byte{33, 0, 0, 0},

						// .FooValue (optional)
						[]byte{1},
						[]byte{2, 0, 0, 0},
						[]byte(val.FooValue),

						// .BarValueNotSet (optional)
						[]byte{0},

						// .Hello
						[]byte{0, 0, 0, 0},
					),
					buf.Bytes(),
				)
				// - 0: [5, 0, 0, 0](len=4)
				// - 1: [104, 101, 108, 108, 111](len=5)
				// - 2: [0, 0, 0, 0](len=4)
				// - 3: [1](len=1)
				// - 4: [6, 0, 0, 0](len=4)
				// - 5: [45, 119, 111, 114, 108, 100](len=6)
				// - 6: [0](len=1)
				// - 7: [1](len=1)
				// - 8: [33, 0, 0, 0](len=4)
				// - 9: [1](len=1)
				// - 10: [2, 0, 0, 0](len=4)
				// - 11: [104, 105](len=2)
				// - 12: [0](len=1)
				// - 13: [0, 0, 0, 0](len=4)
				// - 14: [](len=0)
				{
					dec := NewBorshDecoder(buf.Bytes())
					var got StructWithOptionalFields
					require.NoError(t, dec.Decode(&got))
					{
						// .BarRequiredNotSet is NOT an optiona field,
						// which means that it was encoded as zero,
						// and will be decoded as zero.
						zero := uint32(0)
						val.BarRequiredNotSet = &zero
					}
					require.Equal(t, val, got)
				}
			}
		}
		{
			// struct with enums
			{
				buf := NewWriteByWrite("")
				enc := NewBorshEncoder(buf)
				simple := z
				val := StructWithEnum{
					Simple:        y,
					SimplePointer: &simple,

					Complex: ComplexEnum{
						Enum: 1,
						Bar: Bar{
							BarA: 99,
							BarB: "this is bar",
						},
					},
					ComplexPtr: &ComplexEnum{
						Enum: 1,
						Bar: Bar{
							BarA: 22,
							BarB: "this is bar from pointer",
						},
					},
					ComplexEmpty: ComplexEnumEmpty{
						Enum: 0,
						Foo:  EmptyVariant{},
					},

					ComplexPrimitives1: ComplexEnumPrimitives{
						Enum: 0,
						Foo:  20,
					},

					ComplexPrimitives2: ComplexEnumPrimitives{
						Enum: 1,
						Bar:  11,
					},

					Complex2: ComplexEnumPointers{
						Enum: 1,
						Bar: &Bar{
							BarA: 62,
							BarB: "very tested!!!",
						},
					},

					Complex2Ptr: &ComplexEnumPointers{
						Enum: 1,
						Bar: &Bar{
							BarA: 123,
							BarB: "lorem ipsum",
						},
					},

					Complex2PtrOptionalSet: &ComplexEnumPointers{
						Enum: 1,
						Bar: &Bar{
							BarA: 32,
							BarB: "very complex",
						},
					},

					Map: map[string]uint64{
						"foo": 1,
						"bar": 46,
					},

					Slice: []Struct{
						{
							Foo: "this is first foo",
							Bar: 97,
						},
						{
							Foo: "this is second foo",
							Bar: 98,
						},
					},

					Array: [4]Struct{
						{
							Foo: "arr 0",
							Bar: 22,
						},
						{
							Foo: "arr 1",
							Bar: 23,
						},
						{
							Foo: "arr 2",
							Bar: 24,
						},
						{
							Foo: "arr 3",
							Bar: 25,
						},
					},
				}
				require.NoError(t, enc.Encode(val))
				// fmt.Println(buf.String())
				require.Equal(t,
					concatByteSlices(
						// .Simple
						[]byte{1},

						// .SimplePointer
						[]byte{2},

						// .Complex
						[]byte{1},
						[]byte{99, 0, 0, 0, 0, 0, 0, 0},
						[]byte{11, 0, 0, 0},
						[]byte(val.Complex.Bar.BarB),

						// .ComplexNotSet
						[]byte{0},
						[]byte{0, 0, 0, 0},
						[]byte{0, 0, 0, 0},

						// .ComplexPtr
						[]byte{1},
						[]byte{22, 0, 0, 0, 0, 0, 0, 0},
						[]byte{24, 0, 0, 0},
						[]byte(val.ComplexPtr.Bar.BarB),

						// .ComplexPtrNotSet is not set, leaving the index to zero
						// which corresponds to ComplexPtrNotSet.Foo
						[]byte{0},
						[]byte{0, 0, 0, 0},
						[]byte{0, 0, 0, 0},

						// .ComplexEmpty
						[]byte{0},

						// .ComplexPrimitives1
						[]byte{0},
						[]byte{20, 0, 0, 0},

						// .ComplexPrimitives2
						[]byte{1},
						[]byte{11, 0},

						// .Complex2
						[]byte{1},
						[]byte{62, 0, 0, 0, 0, 0, 0, 0},
						[]byte{14, 0, 0, 0},
						[]byte(val.Complex2.Bar.BarB),

						// .Complex2Ptr
						[]byte{1},
						[]byte{123, 0, 0, 0, 0, 0, 0, 0},
						[]byte{11, 0, 0, 0}, // = len(.Complex2Ptr.Bar.BarB)
						[]byte(val.Complex2Ptr.Bar.BarB),

						// .Complex2PtrOptionalSet
						[]byte{1}, // TODO: why is this set? this shouldn't be here.
						[]byte{1},
						[]byte{32, 0, 0, 0, 0, 0, 0, 0},
						[]byte{12, 0, 0, 0}, // = len(.Complex2PtrOptionalSet.Bar.BarB)
						[]byte(val.Complex2PtrOptionalSet.Bar.BarB),

						// .Complex2PtrOptionalNotSet is optional, and is not set.
						[]byte{0},

						// .Map
						[]byte{2, 0, 0, 0}, // len of map
						[]byte{3, 0, 0, 0}, // len of key "bar" (comes in alphabetical order)
						[]byte("bar"),
						[]byte{46, 0, 0, 0, 0, 0, 0, 0},
						[]byte{3, 0, 0, 0}, // len of key "foo" (comes in alphabetical order)
						[]byte("foo"),
						[]byte{1, 0, 0, 0, 0, 0, 0, 0},

						// .Slice
						[]byte{2, 0, 0, 0}, // len of slice
						// .Slice[0]
						[]byte{17, 0, 0, 0}, // len of [0].Foo
						[]byte(val.Slice[0].Foo),
						[]byte{97, 0, 0, 0},
						// .Slice[1]
						[]byte{18, 0, 0, 0}, // len of [1].Foo
						[]byte(val.Slice[1].Foo),
						[]byte{98, 0, 0, 0},

						// .Array
						// .Array[0]
						[]byte{5, 0, 0, 0}, // len of [0].Foo
						[]byte(val.Array[0].Foo),
						[]byte{22, 0, 0, 0},
						// .Array[1]
						[]byte{5, 0, 0, 0}, // len of [1].Foo
						[]byte(val.Array[1].Foo),
						[]byte{23, 0, 0, 0},
						// .Array[2]
						[]byte{5, 0, 0, 0}, // len of [2].Foo
						[]byte(val.Array[2].Foo),
						[]byte{24, 0, 0, 0},
						// .Array[3]
						[]byte{5, 0, 0, 0}, // len of [3].Foo
						[]byte(val.Array[3].Foo),
						[]byte{25, 0, 0, 0},
					),
					buf.Bytes(),
				)

				{
					dec := NewBorshDecoder(buf.Bytes())
					var got StructWithEnum
					require.NoError(t, dec.Decode(&got))
					{
						val.ComplexPtrNotSet = &ComplexEnum{}
					}
					require.Equal(t, val, got)
				}
			}
		}
	}
}

type StructWithEnum struct {
	Simple        Dummy
	SimplePointer *Dummy

	Complex          ComplexEnum
	ComplexNotSet    ComplexEnum
	ComplexPtr       *ComplexEnum
	ComplexPtrNotSet *ComplexEnum

	ComplexEmpty       ComplexEnumEmpty
	ComplexPrimitives1 ComplexEnumPrimitives
	ComplexPrimitives2 ComplexEnumPrimitives

	Complex2    ComplexEnumPointers
	Complex2Ptr *ComplexEnumPointers

	Complex2PtrOptionalSet    *ComplexEnumPointers `bin:"optional"`
	Complex2PtrOptionalNotSet *ComplexEnumPointers `bin:"optional"`

	Map   map[string]uint64
	Slice []Struct
	Array [4]Struct
}

type StructWithOptionalFields struct {
	FooRequired       *string
	BarRequiredNotSet *uint32
	FooPointer        *string `bin:"optional"`
	FooPointerNotSet  *string `bin:"optional"`
	BarPointer        *uint32 `bin:"optional"`
	FooValue          string  `bin:"optional"`
	BarValueNotSet    uint32  `bin:"optional"`
	Hello             string
}

type Struct struct {
	Foo string
	Bar uint32
}
type StructWithPointerFields struct {
	Foo *string
	Bar *uint32
}

type AA struct {
	A int64
	B int32
	C bool
	D *bool   `bin:"optional"`
	E *uint64 `bin:"optional"`
	// NOTE: multilevel pointers are not supported.
	// DoublePointer **uint64

	Map      map[string]string
	EmptyMap map[int64]string
	// // NOTE: pointers to map are not supported.
	// // PointerToMap      *map[string]string
	// // PointerToMapEmpty *map[string]string
	Array [2]int64

	Optional *Struct `bin:"optional"`
	Value    Struct

	InterfaceEncoderDecoderByValue   CustomEncoding
	InterfaceEncoderDecoderByPointer *CustomEncoding

	// InterfaceEncoderDecoderByValueEmpty   CustomEncoding
	// InterfaceEncoderDecoderByPointerEmpty *CustomEncoding `bin:"optional"`

	HighValuesInt64   []int64
	HighValuesUint64  []uint64
	HighValuesFloat64 []float64
}

type CustomEncoding struct {
	Prefix byte
	Value  uint32
}

func (e CustomEncoding) MarshalWithEncoder(encoder *Encoder) error {
	if err := encoder.WriteUint32(e.Value, LE); err != nil {
		return err
	}
	return encoder.WriteByte(e.Prefix)
}

func (e *CustomEncoding) UnmarshalWithDecoder(decoder *Decoder) (err error) {
	if e.Value, err = decoder.ReadUint32(LE); err != nil {
		return err
	}
	if e.Prefix, err = decoder.ReadByte(); err != nil {
		return err
	}
	return nil
}

var _ EncoderDecoder = &CustomEncoding{}

func TestBorsh_kitchenSink(t *testing.T) {
	boolTrue := true
	uint64Num := uint64(25464132585)
	x := AA{
		A:     1,
		B:     32,
		C:     true,
		D:     &boolTrue,
		E:     &uint64Num,
		Map:   map[string]string{"foo": "bar"},
		Array: [2]int64{57, 88},
		Optional: &Struct{
			Foo: "optional foo",
			Bar: 8888886,
		},
		Value: Struct{
			Foo: "value foo",
			Bar: 7777,
		},
		InterfaceEncoderDecoderByValue:   CustomEncoding{Prefix: byte('b'), Value: 72},
		InterfaceEncoderDecoderByPointer: &CustomEncoding{Prefix: byte('c'), Value: 9999},

		HighValuesInt64: []int64{
			math.MaxInt8,
			math.MaxInt16,
			math.MaxInt32,
			math.MaxInt64,

			-math.MaxInt8,
			-math.MaxInt16,
			-math.MaxInt32,
			-math.MaxInt64,

			math.MaxUint8,
			math.MaxUint16,
			math.MaxUint32,
			// math.MaxUint64,

			-math.MaxUint8,
			-math.MaxUint16,
			-math.MaxUint32,
			// -math.MaxUint64,
		},

		HighValuesUint64: []uint64{
			math.MaxInt8,
			math.MaxInt16,
			math.MaxInt32,
			math.MaxInt64,

			math.MaxUint8,
			math.MaxUint16,
			math.MaxUint32,
			math.MaxUint64,
		},

		HighValuesFloat64: []float64{
			math.MaxFloat32,
			math.MaxFloat64,

			-math.MaxFloat32,
			-math.MaxFloat64,
		},
	}
	buf := NewWriteByWrite("")
	borshEnc := NewBorshEncoder(buf)
	err := borshEnc.Encode(x)
	// fmt.Println(buf.String())
	require.NoError(t, err)

	y := new(AA)
	err = UnmarshalBorsh(y, buf.Bytes())
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

type A struct {
	A int64
	B int32
	C bool
	D *bool
	E *uint64
}

func TestSimple(t *testing.T) {
	boolTrue := true
	uint64Num := uint64(25464132585)
	x := A{
		A: 1,
		B: 32,
		C: true,
		D: &boolTrue,
		E: &uint64Num,
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)
	y := new(A)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

type B struct {
	I8         int8
	I16        int16
	I32        int32
	I64        int64
	U8         uint8
	U16        uint16
	U32        uint32
	U64        uint64
	F32        float32
	F64        float64
	unexported int64 // unexported fields are skipped.
	Err        error // nil interfaces must be specified to be skipped.
}

func TestBasic(t *testing.T) {
	x := B{
		I8:         12,
		I16:        -1,
		I32:        124,
		I64:        1243,
		U8:         1,
		U16:        979,
		U32:        123124,
		U64:        1135351135,
		F32:        -231.23,
		F64:        3121221.232,
		unexported: 333,
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)
	y := new(B)

	// expect the unexported field to be zero because
	// it shouldn't have been encoded or be tried to be decoded:
	x.unexported = 0

	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

type C struct {
	A3 [3]int64
	S  []int64
	P  *int64
	M  map[string]string
}

func TestBasicContainer(t *testing.T) {
	ip := new(int64)
	*ip = 213
	x := C{
		A3: [3]int64{234, -123, 123},
		S:  []int64{21442, 421241241, 2424},
		P:  ip,
		M:  map[string]string{"foo": "bar"},
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(C)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

type N struct {
	B B
	C C
}

func TestNested(t *testing.T) {
	ip := new(int64)
	*ip = 213
	x := N{
		B: B{
			I8:  12,
			I16: -1,
			I32: 124,
			I64: 1243,
			U8:  1,
			U16: 979,
			U32: 123124,
			U64: 1135351135,
			F32: -231.23,
			F64: 3121221.232,
		},
		C: C{
			A3: [3]int64{234, -123, 123},
			S:  []int64{21442, 421241241, 2424},
			P:  ip,
			M:  map[string]string{"foo": "bar"},
		},
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(N)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

type Dummy BorshEnum

const (
	x Dummy = iota
	y
	z
)

type D struct {
	D Dummy
}

func TestSimpleEnum(t *testing.T) {
	x := D{
		D: y,
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(D)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)

	require.Equal(t, x, *y)
}

type ComplexEnum struct {
	Enum BorshEnum `borsh_enum:"true"`
	Foo  Foo
	Bar  Bar
}

type ComplexEnumPointers struct {
	Enum BorshEnum `borsh_enum:"true"`
	Foo  *Foo
	Bar  *Bar
}

type ComplexEnumEmpty struct {
	Enum BorshEnum `borsh_enum:"true"`
	Foo  EmptyVariant
	Bar  Bar
}

type ComplexEnumPrimitives struct {
	Enum BorshEnum `borsh_enum:"true"`
	Foo  uint32
	Bar  int16
}

type Foo struct {
	FooA int32
	FooB string
}

type Bar struct {
	BarA int64
	BarB string
}

func TestComplexEnum(t *testing.T) {
	{
		x := ComplexEnum{
			Enum: 1,
			Bar: Bar{
				BarA: 23,
				BarB: "baz",
			},
		}
		data, err := MarshalBorsh(x)
		require.NoError(t, err)

		y := new(ComplexEnum)
		err = UnmarshalBorsh(y, data)
		require.NoError(t, err)

		require.Equal(t, x, *y)
	}
	{
		x := ComplexEnumPointers{
			Enum: 1,
			Bar: &Bar{
				BarA: 99999,
				BarB: "hello world",
			},
		}
		data, err := MarshalBorsh(x)
		require.NoError(t, err)

		y := new(ComplexEnumPointers)
		err = UnmarshalBorsh(y, data)
		require.NoError(t, err)

		require.Equal(t, x, *y)
	}
	{
		x := ComplexEnumEmpty{
			Enum: 1,
			Bar: Bar{
				BarA: 23,
				BarB: "baz",
			},
		}
		data, err := MarshalBorsh(x)
		require.NoError(t, err)

		y := new(ComplexEnumEmpty)
		err = UnmarshalBorsh(y, data)
		require.NoError(t, err)

		require.Equal(t, x, *y)
	}
	{
		x := ComplexEnumPrimitives{
			Enum: 1,
			Bar:  22,
		}
		data, err := MarshalBorsh(x)
		require.NoError(t, err)

		y := new(ComplexEnumPrimitives)
		err = UnmarshalBorsh(y, data)
		require.NoError(t, err)

		require.Equal(t, x, *y)
	}
}

type S struct {
	S map[int64]struct{}
}

func TestSet(t *testing.T) {
	emptyStruct := struct{}{}
	x := S{
		S: map[int64]struct{}{124: emptyStruct, 214: emptyStruct, 24: emptyStruct, 53: emptyStruct},
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(S)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

type Skipped struct {
	A int64
	B int64 `borsh_skip:"true"`
	C int64
}

func TestSkipped(t *testing.T) {
	x := Skipped{
		A: 32,
		B: 535,
		C: 123,
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(Skipped)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)

	require.Equal(t, x.A, y.A)
	require.Equal(t, x.C, y.C)
	require.NotEqual(t, y.B, x.B, "didn't skip field B")
}

type E struct{}

func TestEmpty(t *testing.T) {
	x := E{}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)
	if len(data) != 0 {
		t.Error("not empty")
	}
	y := new(E)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

func testValue(t *testing.T, v interface{}) {
	data, err := MarshalBorsh(v)
	require.NoError(t, err)

	parsed := reflect.New(reflect.TypeOf(v))
	err = UnmarshalBorsh(parsed.Interface(), data)
	require.NoError(t, err)
	require.Equal(t, v, parsed.Elem().Interface())
}

func TestStrings(t *testing.T) {
	tests := []struct {
		in string
	}{
		{""},
		{"a"},
		{"hellow world"},
		{strings2.Repeat("x", 1024)},
		{strings2.Repeat("x", 4096)},
		{strings2.Repeat("x", 65535)},
		{strings2.Repeat("hello world!", 1000)},
		{"🎯"},
	}

	for _, tt := range tests {
		testValue(t, tt.in)
	}
}

func makeInt32Slice(val int32, len int) []int32 {
	s := make([]int32, len)
	for i := 0; i < len; i++ {
		s[i] = val
	}
	return s
}

func TestSlices(t *testing.T) {
	tests := []struct {
		in []int32
	}{
		{nil}, // zero length slice
		{makeInt32Slice(1000000000, 1)},
		{makeInt32Slice(1000000001, 2)},
		{makeInt32Slice(1000000002, 3)},
		{makeInt32Slice(1000000003, 4)},
		{makeInt32Slice(1000000004, 8)},
		{makeInt32Slice(1000000005, 16)},
		{makeInt32Slice(1000000006, 32)},
		{makeInt32Slice(1000000007, 64)},
		{makeInt32Slice(1000000008, 65)},
	}

	for _, tt := range tests {
		testValue(t, tt.in)
	}
}

func TestUint128_old(t *testing.T) {
	tests := []struct {
		in Int128
	}{
		{func() Int128 {
			v := Int128{
				Hi: math.MaxInt16,
				Lo: math.MaxInt16,
			}
			return v
		}()},
	}

	for _, tt := range tests {
		testValue(t, tt.in)
	}
}

type (
	Myu8  uint8
	Myu16 uint16
	Myu32 uint32
	Myu64 uint64
	Myi8  int8
	Myi16 int16
	Myi32 int32
	Myi64 int64
)

type CustomType struct {
	U8  Myu8
	U16 Myu16
	U32 Myu32
	U64 Myu64
	I8  Myi8
	I16 Myi16
	I32 Myi32
	I64 Myi64
}

func TestCustomType(t *testing.T) {
	x := CustomType{
		U8:  1,
		U16: 2,
		U32: 3,
		U64: 4,
		I8:  5,
		I16: 6,
		I32: 7,
		I64: 8,
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(CustomType)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)

	require.Equal(t, x, *y)
}

func TestStringSlice(t *testing.T) {
	{
		// slice:
		x := []string{"a", "b", "c"}
		data, err := MarshalBorsh(x)
		require.NoError(t, err)

		require.Equal(t, concatByteSlices(
			[]byte{0x3, 0x0, 0x0, 0x0}, // length

			[]byte{0x1, 0x0, 0x0, 0x0}, // length of first string
			[]byte("a"),

			[]byte{0x1, 0x0, 0x0, 0x0}, // length of second string
			[]byte("b"),

			[]byte{0x1, 0x0, 0x0, 0x0}, // length of third string
			[]byte("c"),
		), data)

		y := new([]string)
		err = UnmarshalBorsh(y, data)
		require.NoError(t, err)

		require.Equal(t, x, *y)
	}
	{
		// string slice as field:
		type S struct {
			A []string
		}
		x := S{
			A: []string{"a", "b", "c"},
		}
		data, err := MarshalBorsh(x)
		require.NoError(t, err)

		require.Equal(t, concatByteSlices(
			[]byte{0x3, 0x0, 0x0, 0x0}, // length of A

			[]byte{0x1, 0x0, 0x0, 0x0}, // length of A[0]
			[]byte("a"),

			[]byte{0x1, 0x0, 0x0, 0x0}, // length of A[1]
			[]byte("b"),

			[]byte{0x1, 0x0, 0x0, 0x0}, // length of A[2]
			[]byte("c"),
		), data)

		y := new(S)
		err = UnmarshalBorsh(y, data)
		require.NoError(t, err)

		require.Equal(t, x, *y)
	}
	{
		// string slice as optional field (present):
		type S struct {
			A *[]string `bin:"optional"`
		}
		slice := []string{"a", "b", "c"}
		x := S{
			A: &slice,
		}
		data, err := MarshalBorsh(x)
		require.NoError(t, err)

		require.Equal(t, concatByteSlices(
			[]byte{0x01},               // optionality
			[]byte{0x3, 0x0, 0x0, 0x0}, // slice length

			[]byte{0x1, 0x0, 0x0, 0x0}, // slice item length (string)
			[]byte("a"),

			[]byte{0x1, 0x0, 0x0, 0x0}, // slice item length (string)
			[]byte("b"),

			[]byte{0x1, 0x0, 0x0, 0x0}, // slice item length (string)
			[]byte("c"),
		), data)

		y := new(S)
		err = UnmarshalBorsh(y, data)
		require.NoError(t, err)

		require.Equal(t, x, *y)
	}
	{
		// string slice as optional field (absent):
		type S struct {
			A *[]string `bin:"optional"`
		}
		x := S{}
		data, err := MarshalBorsh(x)
		require.NoError(t, err)

		require.Equal(t, concatByteSlices(
			[]byte{0x0}, // optionality
		), data)

		y := new(S)
		err = UnmarshalBorsh(y, data)
		require.NoError(t, err)

		require.Equal(t, x, *y)
	}
}
