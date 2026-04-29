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

// SendRawTransaction takes already-serialized, fully-signed transaction
// bytes — useful when a wallet or external service hands you the bytes
// to relay. This example builds and signs a transaction here only to
// produce those bytes; in real code you would receive them.
//
// If you have a *solana.Transaction, use SendTransaction instead
// (see the sendTransaction example). If you have base64 instead of
// raw bytes, use SendEncodedTransaction (see sendEncodedTransaction).
func main() {
	ctx := context.Background()
	client := rpc.New(rpc.DevNet_RPC)

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
				solana.LAMPORTS_PER_SOL/1000,
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

	// Serialize the fully-signed transaction to its wire bytes.
	raw, err := tx.MarshalBinary()
	if err != nil {
		panic(fmt.Errorf("marshal: %w", err))
	}

	sig, err := client.SendRawTransaction(ctx, raw)
	if err != nil {
		panic(fmt.Errorf("send raw: %w", err))
	}

	fmt.Println("submitted tx signature:", sig.String())
}
