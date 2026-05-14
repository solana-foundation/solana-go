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
	"fmt"
	ag_gofuzz "github.com/gagliardetto/gofuzz"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_require "github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestEncodeDecode_AmountToUiAmount(t *testing.T) {
	fu := ag_gofuzz.New().NilChance(0)
	for i := 0; i < 1; i++ {
		t.Run("AmountToUiAmount"+strconv.Itoa(i), func(t *testing.T) {
			{
				params := new(AmountToUiAmount)
				fu.Fuzz(params)
				params.AccountMetaSlice = nil
				buf := new(bytes.Buffer)
				err := encodeT(*params, buf)
				ag_require.NoError(t, err)
				got := new(AmountToUiAmount)
				err = decodeT(got, buf.Bytes())
				got.AccountMetaSlice = nil
				ag_require.NoError(t, err)
				ag_require.Equal(t, params, got)
			}
		})
	}
}

func TestAmountToUiAmount_Validate(t *testing.T) {
	mint := ag_solanago.NewWallet().PublicKey()

	t.Run("missing amount returns error", func(t *testing.T) {
		ix := NewAmountToUiAmountInstructionBuilder().SetMintAccount(mint)
		ag_require.Error(t, ix.Validate())
	})

	t.Run("missing mint returns error", func(t *testing.T) {
		ix := NewAmountToUiAmountInstructionBuilder().SetAmount(1000)
		ag_require.Error(t, ix.Validate())
	})

	t.Run("all fields set passes validation", func(t *testing.T) {
		ag_require.NoError(t, NewAmountToUiAmountInstruction(1000, mint).Validate())
	})
}

func TestAmountToUiAmount_EncodeDecode_EdgeCases(t *testing.T) {
	for _, amount := range []uint64{0, 1, 1_000_000_000, ^uint64(0)} {
		amount := amount
		t.Run(fmt.Sprintf("amount=%d", amount), func(t *testing.T) {
			params := &AmountToUiAmount{Amount: &amount}
			buf := new(bytes.Buffer)
			ag_require.NoError(t, encodeT(*params, buf))
			got := new(AmountToUiAmount)
			ag_require.NoError(t, decodeT(got, buf.Bytes()))
			ag_require.Equal(t, amount, *got.Amount)
		})
	}
}
