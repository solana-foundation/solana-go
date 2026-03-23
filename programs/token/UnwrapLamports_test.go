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

package token

import (
	"bytes"
	ag_require "github.com/stretchr/testify/require"
	"testing"
)

func TestEncodeDecode_UnwrapLamports(t *testing.T) {
	t.Run("UnwrapLamports_NoAmount", func(t *testing.T) {
		params := &UnwrapLamports{Amount: nil}
		buf := new(bytes.Buffer)
		err := encodeT(*params, buf)
		ag_require.NoError(t, err)
		got := new(UnwrapLamports)
		err = decodeT(got, buf.Bytes())
		ag_require.NoError(t, err)
		ag_require.Nil(t, got.Amount)
	})

	t.Run("UnwrapLamports_WithAmount", func(t *testing.T) {
		amount := uint64(1000000)
		params := &UnwrapLamports{Amount: &amount}
		buf := new(bytes.Buffer)
		err := encodeT(*params, buf)
		ag_require.NoError(t, err)
		got := new(UnwrapLamports)
		err = decodeT(got, buf.Bytes())
		ag_require.NoError(t, err)
		ag_require.NotNil(t, got.Amount)
		ag_require.Equal(t, amount, *got.Amount)
	})
}
