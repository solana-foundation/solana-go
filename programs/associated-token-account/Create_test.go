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
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
)

func TestCreateInstruction(t *testing.T) {
	payer := solana.MustPublicKeyFromBase58("EvN4kgKmCmYzdbd5kL8Q8YgkUW5RoqMTpBczrfLExtx7")
	wallet := solana.MustPublicKeyFromBase58("5vxoRv2P12q2YwUWQHQj65wWAQUF4z3K8U4BF7B5Hkgf")
	mint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")

	// Create the instruction
	instruction := NewCreateInstruction(payer, wallet, mint)

	// Test validation
	err := instruction.Validate()
	assert.NoError(t, err)

	// Build the instruction
	built := instruction.Build()

	// Verify the accounts
	accounts := built.Accounts()
	assert.Equal(t, 7, len(accounts))

	// Verify the account permissions
	assert.True(t, accounts[0].IsSigner)
	assert.True(t, accounts[0].IsWritable)
	assert.Equal(t, payer, accounts[0].PublicKey)

	// Verify the associated token account is writable
	assert.False(t, accounts[1].IsSigner)
	assert.True(t, accounts[1].IsWritable)

	// Verify the wallet account
	assert.False(t, accounts[2].IsSigner)
	assert.False(t, accounts[2].IsWritable)
	assert.Equal(t, wallet, accounts[2].PublicKey)

	// Verify the mint account
	assert.False(t, accounts[3].IsSigner)
	assert.False(t, accounts[3].IsWritable)
	assert.Equal(t, mint, accounts[3].PublicKey)

	// Verify the system program
	assert.False(t, accounts[4].IsSigner)
	assert.False(t, accounts[4].IsWritable)
	assert.Equal(t, solana.SystemProgramID, accounts[4].PublicKey)

	// Verify the token program
	assert.False(t, accounts[5].IsSigner)
	assert.False(t, accounts[5].IsWritable)
	assert.Equal(t, solana.TokenProgramID, accounts[5].PublicKey)

	// Verify the rent sysvar
	assert.False(t, accounts[6].IsSigner)
	assert.False(t, accounts[6].IsWritable)
	assert.Equal(t, solana.SysVarRentPubkey, accounts[6].PublicKey)
}
