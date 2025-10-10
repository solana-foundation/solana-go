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

package associatedtokenaccount

import (
	"encoding/hex"
	"testing"

	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodingInstruction(t *testing.T) {
	t.Run("should encode", func(t *testing.T) {
		t.Run("Create", func(t *testing.T) {
			// Build an instruction and ensure current encoding matches implementation
			payer := solana.NewWallet().PublicKey()
			wallet := solana.NewWallet().PublicKey()
			mint := solana.NewWallet().PublicKey()
			ix := NewCreateInstructionBuilder().
				SetPayer(payer).
				SetWallet(wallet).
				SetMint(mint).
				Build()
			data, err := ix.Data()
			require.NoError(t, err)
			encodedHex := hex.EncodeToString(data)
			// Current ATA Create encodes no payload bytes
			require.Equal(t, "", encodedHex)
		})
	})

	tests := []struct {
		name              string
		hexData           string
		expectInstruction *Instruction
	}{
		{
			name:    "Create",
			hexData: "",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint8(0),
					Impl:   &Create{},
				},
			},
		},
	}

	t.Run("should encode", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				data, err := test.expectInstruction.Data()
				require.NoError(t, err)
				encodedHex := hex.EncodeToString(data)
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

func TestDecodeSetsAccountsAndGetters(t *testing.T) {
	payer := solana.NewWallet().PublicKey()
	wallet := solana.NewWallet().PublicKey()
	mint := solana.NewWallet().PublicKey()

	// Build an instruction to obtain correctly ordered accounts and data
	ix := NewCreateInstructionBuilder().
		SetPayer(payer).
		SetWallet(wallet).
		SetMint(mint).
		Build()

	accounts := ix.Accounts()
	data, err := ix.Data()
	require.NoError(t, err)

	decoded, err := DecodeInstruction(accounts, data)
	require.NoError(t, err)

	create, ok := decoded.Impl.(*Create)
	require.True(t, ok)

	// Check decoded fields populated via SetAccounts
	assert.Equal(t, payer, create.Payer)
	assert.Equal(t, wallet, create.Wallet)
	assert.Equal(t, mint, create.Mint)

	// Check getters return expected account metas
	require.NotNil(t, create.GetPayerAccount())
	require.NotNil(t, create.GetAssociatedTokenAddressAccount())
	require.NotNil(t, create.GetWalletAccount())
	require.NotNil(t, create.GetMintAccount())

	assert.True(t, create.GetPayerAccount().IsSigner)
	assert.True(t, create.GetPayerAccount().IsWritable)
	assert.Equal(t, payer, create.GetPayerAccount().PublicKey)
	assert.Equal(t, wallet, create.GetWalletAccount().PublicKey)
	assert.Equal(t, mint, create.GetMintAccount().PublicKey)

	// Verify associated token address is correctly derived and placed at index 1
	ata, _, err := solana.FindAssociatedTokenAddress(wallet, mint)
	require.NoError(t, err)
	assert.Equal(t, ata, create.GetAssociatedTokenAddressAccount().PublicKey)
}
