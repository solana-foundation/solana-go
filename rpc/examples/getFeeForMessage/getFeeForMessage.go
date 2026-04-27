// Copyright 2022 github.com/gagliardetto
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

	"github.com/davecgh/go-spew/spew"
	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/system"
	"github.com/solana-foundation/solana-go/v2/rpc"
)

// GetFeeForMessage returns the fee for a base64-encoded Solana Message
// (not a full Transaction). Here we construct a realistic Transfer
// message to show how the base64 argument is produced.
func main() {
	ctx := context.Background()
	client := rpc.New(rpc.MainNetBeta_RPC)

	from := solana.NewWallet().PublicKey()
	to := solana.NewWallet().PublicKey()

	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		panic(fmt.Errorf("get blockhash: %w", err))
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				solana.LAMPORTS_PER_SOL/1000, // 0.001 SOL
				from,
				to,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(from),
	)
	if err != nil {
		panic(fmt.Errorf("build tx: %w", err))
	}

	out, err := client.GetFeeForMessage(
		ctx,
		tx.Message.ToBase64(),
		rpc.CommitmentProcessed,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}
