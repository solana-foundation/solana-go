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

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go/rpc"
)

func main() {
	ctx := context.Background()
	// Testnet has few validators; mainnet-beta has the real schedule.
	client := rpc.New(rpc.MainNetBeta_RPC)

	out, err := client.GetBlockProductionWithOpts(
		ctx,
		&rpc.GetBlockProductionOpts{
			Commitment: rpc.CommitmentFinalized,
			// Range: &rpc.SlotRangeRequest{ FirstSlot: ..., Identity: ... },
		},
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}
