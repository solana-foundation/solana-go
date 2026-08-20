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

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

// Builds, signs and submits a v1 (SIMD-0385) transaction on devnet.
// v1: 4096-byte cap, inline compute budget, no address lookup tables.
// The cluster must have SIMD-0385 activated. See also `sendTransaction`.
func main() {
	ctx := context.Background()
	client := rpc.New(rpc.DevNet_RPC)

	// Fresh funded sender; real code loads a keypair (PrivateKeyFromSolanaKeygenFile).
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

	// Inline compute budget: unset CU limit means 0; PriorityFee is total lamports.
	config := solana.TransactionConfig{}.
		WithComputeUnitLimit(20_000).
		WithPriorityFee(1_000)

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
		// Selects the V1 message format and embeds the compute budget.
		solana.TransactionV1Config(config),
	)
	if err != nil {
		panic(fmt.Errorf("build tx: %w", err))
	}
	fmt.Println("message version is v1:", tx.Message.GetVersion() == solana.MessageVersionV1)

	if _, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if sender.PublicKey().Equals(key) {
			return &sender.PrivateKey
		}
		return nil
	}); err != nil {
		panic(fmt.Errorf("sign: %w", err))
	}

	wire, err := tx.MarshalBinary()
	if err != nil {
		panic(fmt.Errorf("serialize: %w", err))
	}
	fmt.Printf("serialized size: %d bytes (max %d)\n", len(wire), solana.MaxTransactionSizeV1)

	sig, err := client.SendTransaction(ctx, tx)
	if err != nil {
		panic(fmt.Errorf("send: %w", err))
	}

	fmt.Println("submitted tx signature:", sig.String())

	// Fetch it back with MaxSupportedTransactionVersion: &rpc.MaxSupportedTransactionVersion1.
}
