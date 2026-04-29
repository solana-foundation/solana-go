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
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func main() {
	ctx := context.Background()
	client := rpc.New(rpc.MainNetBeta_RPC)

	// Fetch any currently-voting validator so the example keeps working
	// across epochs without a hardcoded pubkey.
	voteAccounts, err := client.GetVoteAccounts(ctx, nil)
	if err != nil {
		panic(err)
	}
	if len(voteAccounts.Current) == 0 {
		panic("no current vote accounts")
	}
	votePubkey := voteAccounts.Current[0].VotePubkey

	out, err := client.GetInflationReward(
		ctx,
		[]solana.PublicKey{votePubkey},
		&rpc.GetInflationRewardOpts{
			Commitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}
