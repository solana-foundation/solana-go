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

func TestRecoverNestedInstruction(t *testing.T) {
	nestedAssociatedToken := solana.MustPublicKeyFromBase58("EvN4kgKmCmYzdbd5kL8Q8YgkUW5RoqMTpBczrfLExtx7")
	nestedMint := solana.MustPublicKeyFromBase58("5vxoRv2P12q2YwUWQHQj65wWAQUF4z3K8U4BF7B5Hkgf")
	destinationAssociatedToken := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	ownerAssociatedToken := solana.MustPublicKeyFromBase58("DYw8jMTpGLyAsRk4rH1jJJJYMR3M9aWKB6xVBVVhWxWF")
	ownerMint := solana.MustPublicKeyFromBase58("H8oWBMFpKZWDSGSkEBknEU5qJnCELB3bESmQmA4uVW2s")
	owner := solana.MustPublicKeyFromBase58("B1aLzaNMeFVAyQ6f3XbbUyKcH2YPHu2fqiEagmiF23VR")

	// Create the instruction
	instruction := NewRecoverNestedInstruction(
		nestedAssociatedToken,
		nestedMint,
		destinationAssociatedToken,
		ownerAssociatedToken,
		ownerMint,
		owner,
	)

	// Test validation
	err := instruction.Validate()
	assert.NoError(t, err)

	// Build the instruction
	built := instruction.Build()

	// Verify the accounts
	accounts := built.Accounts()
	assert.Equal(t, 7, len(accounts))

	// Verify the nested associated token account
	assert.False(t, accounts[0].IsSigner)
	assert.True(t, accounts[0].IsWritable)
	assert.Equal(t, nestedAssociatedToken, accounts[0].PublicKey)

	// Verify the nested mint account
	assert.False(t, accounts[1].IsSigner)
	assert.False(t, accounts[1].IsWritable)
	assert.Equal(t, nestedMint, accounts[1].PublicKey)

	// Verify the destination associated token account
	assert.False(t, accounts[2].IsSigner)
	assert.True(t, accounts[2].IsWritable)
	assert.Equal(t, destinationAssociatedToken, accounts[2].PublicKey)

	// Verify the owner associated token account
	assert.False(t, accounts[3].IsSigner)
	assert.False(t, accounts[3].IsWritable)
	assert.Equal(t, ownerAssociatedToken, accounts[3].PublicKey)

	// Verify the owner mint account
	assert.False(t, accounts[4].IsSigner)
	assert.False(t, accounts[4].IsWritable)
	assert.Equal(t, ownerMint, accounts[4].PublicKey)

	// Verify the owner account
	assert.True(t, accounts[5].IsSigner)
	assert.False(t, accounts[5].IsWritable)
	assert.Equal(t, owner, accounts[5].PublicKey)

	// Verify the token program
	assert.False(t, accounts[6].IsSigner)
	assert.False(t, accounts[6].IsWritable)
	assert.Equal(t, solana.TokenProgramID, accounts[6].PublicKey)

	// Test validation with zero values
	invalidInstruction := NewRecoverNestedInstructionBuilder()
	err = invalidInstruction.Validate()
	assert.Error(t, err)
}
