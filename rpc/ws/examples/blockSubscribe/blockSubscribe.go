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

// blockSubscribe streams whole blocks. BlockSubscribe accepts every
// encoding the RPC does — base58, base64, base64+zstd, and jsonParsed.
// BlockResult.Value is a union: binary encodings populate Block
// (*rpc.GetBlockResult); jsonParsed populates ParsedBlock
// (*rpc.GetParsedBlockResult).
//
// Public Solana RPC endpoints do NOT expose block subscriptions by
// default — the validator must be started with
// `--rpc-pubsub-enable-block-subscription`. Point this at your own
// endpoint, or at a provider that enables it (e.g. Helius, Triton).
package main

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func main() {
	ctx := context.Background()

	client, err := ws.Connect(ctx, rpc.MainNetBeta_WS)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	maxVersion := uint64(0)
	rewards := false

	// Switch to solana.EncodingJSONParsed to receive human-readable,
	// parsed transactions via BlockResult.Value.ParsedBlock instead.
	sub, err := client.BlockSubscribe(
		ws.NewBlockSubscribeFilterAll(),
		&ws.BlockSubscribeOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			Encoding:                       solana.EncodingBase64,
			TransactionDetails:             rpc.TransactionDetailsFull,
			Rewards:                        &rewards,
			MaxSupportedTransactionVersion: &maxVersion,
		},
	)
	if err != nil {
		panic(fmt.Errorf("subscribe: %w", err))
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv(ctx)
		if err != nil {
			panic(err)
		}

		// Exactly one of Block / ParsedBlock is set, depending on the
		// Encoding requested above.
		switch {
		case got.Value.Block != nil:
			fmt.Printf("slot=%d  txs=%d (binary)\n",
				got.Value.Slot,
				len(got.Value.Block.Transactions),
			)
		case got.Value.ParsedBlock != nil:
			fmt.Printf("slot=%d  txs=%d (parsed)\n",
				got.Value.Slot,
				len(got.Value.ParsedBlock.Transactions),
			)
		default:
			fmt.Printf("slot=%d  no block payload\n", got.Value.Slot)
		}
	}
}
