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

	"github.com/davecgh/go-spew/spew"
	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/rpc"
)

func main() {
	ctx := context.Background()
	client := rpc.New(rpc.MainNetBeta_RPC)

	slot, err := client.GetSlot(ctx, rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	// Mainnet blocks contain v0 (versioned) transactions; GetBlock
	// requires MaxSupportedTransactionVersion or it errors out.
	maxVersion := uint64(0)

	{
		out, err := client.GetBlockWithOpts(ctx, slot, &rpc.GetBlockOpts{
			MaxSupportedTransactionVersion: &maxVersion,
		})
		if err != nil {
			panic(err)
		}
		// Full block is large; just show the tx count.
		fmt.Println("transactions in block:", len(out.Transactions))
	}

	{
		includeRewards := false
		out, err := client.GetBlockWithOpts(
			ctx,
			slot,
			&rpc.GetBlockOpts{
				Encoding:                       solana.EncodingBase64,
				Commitment:                     rpc.CommitmentFinalized,
				TransactionDetails:             rpc.TransactionDetailsSignatures,
				Rewards:                        &includeRewards,
				MaxSupportedTransactionVersion: &maxVersion,
			},
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(out)
	}
}
