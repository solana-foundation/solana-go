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

import "encoding/binary"

// option carries per-field decode/encode flags. It is intentionally a small
// value type so it can be passed by value and stack-allocated. Previously it
// was always heap-allocated via &option{...}; now callers should use a literal
// option{} (the zero value is a valid "no special handling" sentinel; the
// decode/encode entry points fill in a default Order if it's nil).
type option struct {
	is_OptionalField  bool
	is_COptionalField bool
	sliceSizeIsSet    bool
	sliceSize         int
	Order             binary.ByteOrder
}

var (
	LE binary.ByteOrder = binary.LittleEndian
	BE binary.ByteOrder = binary.BigEndian
)

var defaultByteOrder = binary.LittleEndian

// defaultOption is the value used when a caller would otherwise pass nil.
var defaultOption = option{Order: defaultByteOrder}

func (o option) is_Optional() bool {
	return o.is_OptionalField
}

func (o option) is_COptional() bool {
	return o.is_COptionalField
}

func (o option) hasSizeOfSlice() bool {
	return o.sliceSizeIsSet
}

func (o option) getSizeOfSlice() int {
	return o.sliceSize
}

type Encoding int

const (
	EncodingBin Encoding = iota
	EncodingCompactU16
	EncodingBorsh
)

func (enc Encoding) String() string {
	switch enc {
	case EncodingBin:
		return "Bin"
	case EncodingCompactU16:
		return "CompactU16"
	case EncodingBorsh:
		return "Borsh"
	default:
		return ""
	}
}

func (en Encoding) IsBorsh() bool {
	return en == EncodingBorsh
}

func (en Encoding) IsBin() bool {
	return en == EncodingBin
}

func (en Encoding) IsCompactU16() bool {
	return en == EncodingCompactU16
}

func isValidEncoding(enc Encoding) bool {
	switch enc {
	case EncodingBin, EncodingCompactU16, EncodingBorsh:
		return true
	default:
		return false
	}
}
