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

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

// This example builds, signs, and submits a SOL transfer on devnet.
// It demonstrates the idiomatic flow for `client.SendTransaction`, which
// accepts a *solana.Transaction directly.
//
// For submitting a transaction that is already serialized, see the
// sibling examples `sendRawTransaction` (raw bytes) and
// `sendEncodedTransaction` (base64 string).
func main() {
	ctx := context.Background()
	client := rpc.New(rpc.DevNet_RPC)

	// Generate a fresh sender and fund it via airdrop. In real code, load
	// an existing keypair instead, e.g.:
	//   sender, _ := solana.PrivateKeyFromSolanaKeygenFile("/path/to/id.json")
	sender := solana.NewWallet()
	fmt.Println("sender:", sender.PublicKey())

	airdropSig, err := client.RequestAirdrop(
		ctx,
		sender.PublicKey(),
		solana.LAMPORTS_PER_SOL,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(fmt.Errorf("airdrop: %w", err))
	}
	fmt.Println("airdrop signature:", airdropSig)
	time.Sleep(20 * time.Second) // wait for the airdrop to finalize

	recipient := solana.NewWallet().PublicKey()

	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		panic(fmt.Errorf("get blockhash: %w", err))
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				solana.LAMPORTS_PER_SOL/1000, // 0.001 SOL
				sender.PublicKey(),
				recipient,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(sender.PublicKey()),
	)
	if err != nil {
		panic(fmt.Errorf("build tx: %w", err))
	}

	if _, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if sender.PublicKey().Equals(key) {
			return &sender.PrivateKey
		}
		return nil
	}); err != nil {
		panic(fmt.Errorf("sign: %w", err))
	}

	sig, err := client.SendTransaction(ctx, tx)
	if err != nil {
		panic(fmt.Errorf("send: %w", err))
	}

	fmt.Println("submitted tx signature:", sig.String())
}
