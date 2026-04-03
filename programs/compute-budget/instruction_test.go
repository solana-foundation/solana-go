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

package computebudget

import (
	"bytes"
	"encoding/hex"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodingInstruction(t *testing.T) {
	tests := []struct {
		name              string
		hexData           string
		expectInstruction *Instruction
	}{
		{
			name:    "RequestUnitsDeprecated",
			hexData: "00c05c1500e8030000",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint8(0),
					Impl: &RequestUnitsDeprecated{
						Units:         1400000,
						AdditionalFee: 1000,
					},
				},
			},
		},
		{
			name:    "RequestHeapFrame",
			hexData: "01a00f0000",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint8(1),
					Impl: &RequestHeapFrame{
						HeapSize: 4000,
					},
				},
			},
		},
		{
			name:    "SetComputeUnitLimit",
			hexData: "02c05c1500",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint8(2),
					Impl: &SetComputeUnitLimit{
						Units: 1400000,
					},
				},
			},
		},
		{
			name:    "SetComputeUnitPrice",
			hexData: "03e803000000000000",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint8(3),
					Impl: &SetComputeUnitPrice{
						MicroLamports: 1000,
					},
				},
			},
		},
		{
			name:    "SetLoadedAccountsDataSizeLimit",
			hexData: "0400800000",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint8(4),
					Impl: &SetLoadedAccountsDataSizeLimit{
						Bytes: 32 * 1024,
					},
				},
			},
		},
	}

	t.Run("should encode", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				err := bin.NewBinEncoder(buf).Encode(test.expectInstruction)
				require.NoError(t, err)

				encodedHex := hex.EncodeToString(buf.Bytes())
				require.Equal(t, test.hexData, encodedHex)
			})
		}
	})

	t.Run("should decode", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				data, err := hex.DecodeString(test.hexData)
				require.NoError(t, err)
				var instruction *Instruction
				err = bin.NewBinDecoder(data).Decode(&instruction)
				require.NoError(t, err)
				assert.Equal(t, test.expectInstruction, instruction)
			})
		}
	})

}
