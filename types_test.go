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

package solana

import (
	"bytes"
	"testing"

	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompiledInstructions(t *testing.T) {
	traceEnabled = true

	ci := &CompiledInstruction{
		ProgramIDIndex: 5,
		Accounts:       []uint16{2, 5, 8},
		Data:           Base58([]byte{1, 2, 3, 4, 5}),
	}
	buf := &bytes.Buffer{}
	encoder := bin.NewBinEncoder(buf)
	err := encoder.Encode(ci)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x5, 0x0, 0x3, 0x2, 0x0, 0x5, 0x0, 0x8, 0x0, 0x5, 0x1, 0x2, 0x3, 0x4, 0x5}, buf.Bytes())
}
