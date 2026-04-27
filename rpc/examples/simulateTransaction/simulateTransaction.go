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
	"encoding/json"
	"fmt"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/system"
	"github.com/solana-foundation/solana-go/v2/rpc"
)

// This example simulates a SOL transfer on mainnet without submitting it,
// without needing any private key, and without funding any wallet.
//
// The trick: SimulateTransactionWithOpts lets us disable signature
// verification (SigVerify=false) and substitute a recent blockhash
// (ReplaceRecentBlockhash=true), so we only need a payer that already
// has SOL on-chain. Here we use a well-known, publicly documented
// funded address (the Binance-2 hot wallet) purely as the fee payer
// for the simulation.
//
// Swap the payer for any address that has SOL if you want to simulate
// against a different balance.
func main() {
	ctx := context.Background()
	client := rpc.New(rpc.MainNetBeta_RPC)

	payer := solana.MustPublicKeyFromBase58("9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM")
	recipient := solana.NewWallet().PublicKey()

	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		panic(fmt.Errorf("get blockhash: %w", err))
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				solana.LAMPORTS_PER_SOL/1000, // 0.001 SOL
				payer,
				recipient,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(payer),
	)
	if err != nil {
		panic(fmt.Errorf("build tx: %w", err))
	}

	// Placeholder signature slot; SigVerify=false so the contents are not checked.
	tx.Signatures = []solana.Signature{{}}

	replace := true
	out, err := client.SimulateTransactionWithOpts(ctx, tx, &rpc.SimulateTransactionOpts{
		SigVerify:              false,
		ReplaceRecentBlockhash: replace,
		Commitment:             rpc.CommitmentProcessed,
	})
	if err != nil {
		panic(fmt.Errorf("simulate: %w", err))
	}

	pretty, _ := json.MarshalIndent(out.Value, "", "  ")
	fmt.Println(string(pretty))
}
