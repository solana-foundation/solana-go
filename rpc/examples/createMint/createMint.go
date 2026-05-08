// Copyright 2026 github.com/gagliardetto
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

// createMint builds and submits the two-instruction transaction that
// creates a new SPL Token mint on devnet:
//
//  1. system.CreateAccount — allocates a 82-byte account funded with
//     the rent-exempt minimum, owned by the Token program.
//  2. token.InitializeMint2 — writes the mint state (decimals, mint
//     authority, freeze authority). Prefer Mint2 over Mint: v1.1+ of
//     the Token program dropped the Rent sysvar account that the
//     original instruction required.
//
// Swap token.ProgramID for token2022.ProgramID (and import the sibling
// package) to create a Token-2022 mint. Extensions must be initialized
// between CreateAccount and InitializeMint2, and the account allocated
// larger than MINT_SIZE accordingly.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/system"
	"github.com/solana-foundation/solana-go/v2/programs/token"
	"github.com/solana-foundation/solana-go/v2/rpc"
)

func main() {
	ctx := context.Background()
	client := rpc.New(rpc.DevNet_RPC)

	// In real code, load an existing keypair instead of generating one:
	//   payer, _ := solana.PrivateKeyFromSolanaKeygenFile("/path/to/id.json")
	payer := solana.NewWallet()
	mint := solana.NewWallet()
	fmt.Println("payer:", payer.PublicKey())
	fmt.Println("mint: ", mint.PublicKey())

	airdropSig, err := client.RequestAirdrop(
		ctx,
		payer.PublicKey(),
		solana.LAMPORTS_PER_SOL,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(fmt.Errorf("airdrop: %w", err))
	}
	fmt.Println("airdrop signature:", airdropSig)
	time.Sleep(20 * time.Second) // wait for the airdrop to finalize

	// Rent-exempt minimum lamports for a 82-byte (MINT_SIZE) account.
	rentLamports, err := client.GetMinimumBalanceForRentExemption(
		ctx,
		token.MINT_SIZE,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(fmt.Errorf("get min balance for rent exemption: %w", err))
	}

	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		panic(fmt.Errorf("get blockhash: %w", err))
	}

	createAccountIx := system.NewCreateAccountInstruction(
		rentLamports,
		token.MINT_SIZE,
		token.ProgramID,   // owner: the Token program
		payer.PublicKey(), // funding account
		mint.PublicKey(),  // the new account (the mint itself)
	).Build()

	initMintIx := token.NewInitializeMint2Instruction(
		9,                  // decimals
		payer.PublicKey(),  // mint authority
		solana.PublicKey{}, // freeze authority; pass a zero pubkey to disable
		mint.PublicKey(),
	).Build()

	tx, err := solana.NewTransaction(
		[]solana.Instruction{createAccountIx, initMintIx},
		recent.Value.Blockhash,
		solana.TransactionPayer(payer.PublicKey()),
	)
	if err != nil {
		panic(fmt.Errorf("build tx: %w", err))
	}

	// Both keypairs must sign: the mint keypair authorizes creation at
	// its own pubkey, and the payer funds it.
	if _, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		switch {
		case payer.PublicKey().Equals(key):
			return &payer.PrivateKey
		case mint.PublicKey().Equals(key):
			return &mint.PrivateKey
		}
		return nil
	}); err != nil {
		panic(fmt.Errorf("sign: %w", err))
	}

	sig, err := client.SendTransaction(ctx, tx)
	if err != nil {
		panic(fmt.Errorf("send: %w", err))
	}

	fmt.Println("created mint:", mint.PublicKey())
	fmt.Println("tx signature:", sig)
}
