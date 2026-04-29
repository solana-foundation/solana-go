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

	"github.com/gagliardetto/solana-go/rpc"
)

func main() {
	ctx := context.Background()
	client := rpc.New(rpc.MainNetBeta_RPC)

	// Without options, returns every vote account on the cluster. To
	// filter to one validator, pass &GetVoteAccountsOpts{VotePubkey: &...}.
	out, err := client.GetVoteAccounts(ctx, nil)
	if err != nil {
		panic(err)
	}

	// Full response is large on mainnet; print a summary.
	fmt.Println("current vote accounts:", len(out.Current))
	fmt.Println("delinquent vote accounts:", len(out.Delinquent))
	for i, v := range out.Current {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s  stake=%d commission=%d%%\n",
			v.VotePubkey, v.ActivatedStake, v.Commission)
	}
}
