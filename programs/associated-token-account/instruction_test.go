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
			require.Equal(t, "00", hex.EncodeToString(data))
		})
		t.Run("CreateIdempotent", func(t *testing.T) {
			payer := solana.NewWallet().PublicKey()
			wallet := solana.NewWallet().PublicKey()
			mint := solana.NewWallet().PublicKey()
			ix := NewCreateIdempotentInstructionBuilder().
				SetPayer(payer).
				SetWallet(wallet).
				SetMint(mint).
				Build()
			data, err := ix.Data()
			require.NoError(t, err)
			require.Equal(t, "01", hex.EncodeToString(data))
		})
		t.Run("RecoverNested", func(t *testing.T) {
			wallet := solana.NewWallet().PublicKey()
			nestedMint := solana.NewWallet().PublicKey()
			ownerMint := solana.NewWallet().PublicKey()
			ix := NewRecoverNestedInstructionBuilder().
				SetWallet(wallet).
				SetNestedMint(nestedMint).
				SetOwnerMint(ownerMint).
				Build()
			data, err := ix.Data()
			require.NoError(t, err)
			require.Equal(t, "02", hex.EncodeToString(data))
		})
	})

	tests := []struct {
		name              string
		hexData           string
		expectInstruction *Instruction
	}{
		{
			name:    "Create",
			hexData: "00",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint8(Instruction_Create),
					Impl:   &Create{},
				},
			},
		},
		{
			name:    "CreateIdempotent",
			hexData: "01",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint8(Instruction_CreateIdempotent),
					Impl:   &CreateIdempotent{},
				},
			},
		},
		{
			name:    "RecoverNested",
			hexData: "02",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint8(Instruction_RecoverNested),
					Impl:   &RecoverNested{},
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

func TestDecodeEmptyDataAsCreate(t *testing.T) {
	// Backward compatibility: empty instruction data should decode as Create.
	// Use a valid set of 6 accounts so SetAccounts doesn't fail.
	payer := solana.NewWallet().PublicKey()
	wallet := solana.NewWallet().PublicKey()
	mint := solana.NewWallet().PublicKey()
	ix := NewCreateInstructionBuilder().
		SetPayer(payer).
		SetWallet(wallet).
		SetMint(mint).
		Build()
	accounts := ix.Accounts()

	inst, err := DecodeInstruction(accounts, []byte{})
	require.NoError(t, err)
	_, ok := inst.Impl.(*Create)
	require.True(t, ok, "empty data should decode as Create")
	assert.Equal(t, bin.TypeIDFromUint8(Instruction_Create), inst.TypeID)
}

func TestDecodeSetsAccountsAndGetters(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		payer := solana.NewWallet().PublicKey()
		wallet := solana.NewWallet().PublicKey()
		mint := solana.NewWallet().PublicKey()

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

		assert.Equal(t, payer, create.Payer)
		assert.Equal(t, wallet, create.Wallet)
		assert.Equal(t, mint, create.Mint)

		require.NotNil(t, create.GetPayerAccount())
		require.NotNil(t, create.GetAssociatedTokenAddressAccount())
		require.NotNil(t, create.GetWalletAccount())
		require.NotNil(t, create.GetMintAccount())

		assert.True(t, create.GetPayerAccount().IsSigner)
		assert.True(t, create.GetPayerAccount().IsWritable)
		assert.Equal(t, payer, create.GetPayerAccount().PublicKey)
		assert.Equal(t, wallet, create.GetWalletAccount().PublicKey)
		assert.Equal(t, mint, create.GetMintAccount().PublicKey)

		ata, _, err := solana.FindAssociatedTokenAddress(wallet, mint)
		require.NoError(t, err)
		assert.Equal(t, ata, create.GetAssociatedTokenAddressAccount().PublicKey)
	})

	t.Run("CreateIdempotent", func(t *testing.T) {
		payer := solana.NewWallet().PublicKey()
		wallet := solana.NewWallet().PublicKey()
		mint := solana.NewWallet().PublicKey()

		ix := NewCreateIdempotentInstructionBuilder().
			SetPayer(payer).
			SetWallet(wallet).
			SetMint(mint).
			Build()

		accounts := ix.Accounts()
		data, err := ix.Data()
		require.NoError(t, err)

		decoded, err := DecodeInstruction(accounts, data)
		require.NoError(t, err)

		ci, ok := decoded.Impl.(*CreateIdempotent)
		require.True(t, ok)

		assert.Equal(t, payer, ci.Payer)
		assert.Equal(t, wallet, ci.Wallet)
		assert.Equal(t, mint, ci.Mint)

		require.NotNil(t, ci.GetPayerAccount())
		require.NotNil(t, ci.GetAssociatedTokenAddressAccount())
		require.NotNil(t, ci.GetWalletAccount())
		require.NotNil(t, ci.GetMintAccount())

		assert.True(t, ci.GetPayerAccount().IsSigner)
		assert.True(t, ci.GetPayerAccount().IsWritable)

		ata, _, err := solana.FindAssociatedTokenAddress(wallet, mint)
		require.NoError(t, err)
		assert.Equal(t, ata, ci.GetAssociatedTokenAddressAccount().PublicKey)
	})

	t.Run("Create token-2022", func(t *testing.T) {
		payer := solana.NewWallet().PublicKey()
		wallet := solana.NewWallet().PublicKey()
		mint := solana.NewWallet().PublicKey()

		ix := NewCreateInstructionBuilder().
			SetPayer(payer).
			SetWallet(wallet).
			SetMint(mint).
			SetTokenProgram(solana.Token2022ProgramID).
			Build()

		accounts := ix.Accounts()
		data, err := ix.Data()
		require.NoError(t, err)

		decoded, err := DecodeInstruction(accounts, data)
		require.NoError(t, err)

		create, ok := decoded.Impl.(*Create)
		require.True(t, ok)
		assert.Equal(t, solana.Token2022ProgramID, create.TokenProgram)

		ata, _, err := solana.FindAssociatedTokenAddressWithProgram(wallet, mint, solana.Token2022ProgramID)
		require.NoError(t, err)
		assert.Equal(t, ata, create.GetAssociatedTokenAddressAccount().PublicKey)
	})

	t.Run("CreateIdempotent token-2022", func(t *testing.T) {
		payer := solana.NewWallet().PublicKey()
		wallet := solana.NewWallet().PublicKey()
		mint := solana.NewWallet().PublicKey()

		ix := NewCreateIdempotentInstructionBuilder().
			SetPayer(payer).
			SetWallet(wallet).
			SetMint(mint).
			SetTokenProgram(solana.Token2022ProgramID).
			Build()

		accounts := ix.Accounts()
		data, err := ix.Data()
		require.NoError(t, err)

		decoded, err := DecodeInstruction(accounts, data)
		require.NoError(t, err)

		ci, ok := decoded.Impl.(*CreateIdempotent)
		require.True(t, ok)
		assert.Equal(t, solana.Token2022ProgramID, ci.TokenProgram)

		ata, _, err := solana.FindAssociatedTokenAddressWithProgram(wallet, mint, solana.Token2022ProgramID)
		require.NoError(t, err)
		assert.Equal(t, ata, ci.GetAssociatedTokenAddressAccount().PublicKey)
	})

	t.Run("RecoverNested", func(t *testing.T) {
		wallet := solana.NewWallet().PublicKey()
		nestedMint := solana.NewWallet().PublicKey()
		ownerMint := solana.NewWallet().PublicKey()

		ix := NewRecoverNestedInstructionBuilder().
			SetWallet(wallet).
			SetNestedMint(nestedMint).
			SetOwnerMint(ownerMint).
			Build()

		accounts := ix.Accounts()
		data, err := ix.Data()
		require.NoError(t, err)

		decoded, err := DecodeInstruction(accounts, data)
		require.NoError(t, err)

		rn, ok := decoded.Impl.(*RecoverNested)
		require.True(t, ok)

		assert.Equal(t, wallet, rn.Wallet)
		assert.Equal(t, nestedMint, rn.NestedMint)
		assert.Equal(t, ownerMint, rn.OwnerMint)

		require.NotNil(t, rn.GetWalletAccount())
		assert.True(t, rn.GetWalletAccount().IsSigner)
		assert.True(t, rn.GetWalletAccount().IsWritable)
		assert.Equal(t, wallet, rn.GetWalletAccount().PublicKey)

		// Verify derived accounts
		ownerATA, _, err := solana.FindAssociatedTokenAddress(wallet, ownerMint)
		require.NoError(t, err)
		assert.Equal(t, ownerATA, rn.GetOwnerAssociatedTokenAccountAccount().PublicKey)

		walletATA, _, err := solana.FindAssociatedTokenAddress(wallet, nestedMint)
		require.NoError(t, err)
		assert.Equal(t, walletATA, rn.GetWalletAssociatedTokenAccountAccount().PublicKey)

		nestedATA, _, err := solana.FindAssociatedTokenAddress(ownerATA, nestedMint)
		require.NoError(t, err)
		assert.Equal(t, nestedATA, rn.GetNestedAssociatedTokenAccountAccount().PublicKey)
	})
}
