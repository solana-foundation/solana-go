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

package system

import (
	"bytes"
	"strconv"
	"testing"

	bin "github.com/gagliardetto/binary"
	ag_gofuzz "github.com/gagliardetto/gofuzz"
	"github.com/solana-foundation/solana-go/v2"
	ag_require "github.com/stretchr/testify/require"
)

func TestEncodeDecode_CreateAccountWithSeed(t *testing.T) {
	fu := ag_gofuzz.New().NilChance(0)
	for i := 0; i < 1; i++ {
		t.Run("CreateAccountWithSeed"+strconv.Itoa(i), func(t *testing.T) {
			{
				params := new(CreateAccountWithSeed)
				fu.Fuzz(params)
				params.AccountMetaSlice = nil
				buf := new(bytes.Buffer)
				err := encodeT(*params, buf)
				ag_require.NoError(t, err)
				//
				got := new(CreateAccountWithSeed)
				err = decodeT(got, buf.Bytes())
				got.AccountMetaSlice = nil
				ag_require.NoError(t, err)
				ag_require.Equal(t, params, got)
			}
		})
	}
}

func TestEncDec(t *testing.T) {
	instr := []byte{204, 95, 77, 127, 148, 25, 135, 127, 89, 146, 22, 90, 233, 80, 113, 3, 70, 176, 165, 222, 81, 200, 100, 223, 117, 165, 155, 44, 53, 225, 124, 20, 5, 0, 0, 0, 0, 0, 0, 0, 104, 101, 108, 108, 111, 192, 4, 14, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 57, 111, 59, 111, 183, 248, 249, 251, 128, 174, 206, 0, 81, 22, 3, 173, 244, 104, 15, 249, 239, 112, 33, 255, 66, 169, 29, 66, 7, 106, 231, 230}

	{
		payerPrivateKey := solana.MustPrivateKeyFromBase58("5LRLfrUP22VtiNaPGAEgHPucoJmG8ejmomMVmpn4fkXjexYsT7RQGfGuMePG5PKvecZxMGrqa6EP2RmYcm7TYQvX")
		payerAccount, _ := solana.WalletFromPrivateKeyBase58(payerPrivateKey.String())
		programID := "4sCcZNQR8vfWckyi5L9KdptdaiLxdiMjVgKQay7HxzmK"
		programPubKey := solana.MustPublicKeyFromBase58(programID)

		newSubAccount, err := solana.CreateWithSeed(
			payerAccount.PublicKey(),
			"hello",
			programPubKey,
		)
		ag_require.NoError(t, err)

		instruction := NewCreateAccountWithSeedInstruction(
			payerAccount.PublicKey(),
			"hello",
			918720,
			4,
			programPubKey,
			payerAccount.PublicKey(),
			newSubAccount,
			payerAccount.PublicKey(),
		)

		{
			buffer := new(bytes.Buffer)
			err = bin.NewBinEncoder(buffer).Encode(instruction)
			ag_require.NoError(t, err)
			ag_require.Equal(t, instr, buffer.Bytes())
		}

		{
			got := new(CreateAccountWithSeed)
			err = decodeT(got, instr)
			got.AccountMetaSlice = solana.AccountMetaSlice{
				solana.Meta(payerAccount.PublicKey()).WRITE().SIGNER(),
				solana.Meta(newSubAccount).WRITE(),
				solana.Meta(payerAccount.PublicKey()).SIGNER(),
			}
			ag_require.NoError(t, err)
			ag_require.Equal(t, instruction, got)
		}
	}
}
