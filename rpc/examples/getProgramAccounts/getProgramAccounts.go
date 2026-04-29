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

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// getProgramAccounts returns every account owned by a program. Public
// mainnet RPCs reject it without a restrictive filter because the
// response would be too large. This example filters Token-2022 accounts
// down to just mint accounts (82-byte data) and prints the first few.
func main() {
	ctx := context.Background()
	client := rpc.New(rpc.MainNetBeta_RPC)

	out, err := client.GetProgramAccountsWithOpts(
		ctx,
		solana.Token2022ProgramID,
		&rpc.GetProgramAccountsOpts{
			Commitment: rpc.CommitmentFinalized,
			Filters: []rpc.RPCFilter{
				{DataSize: 82}, // SPL Token mint account size
			},
			// Only fetch the first byte of data so the response stays small.
			DataSlice: &rpc.DataSlice{Offset: ptrU64(0), Length: ptrU64(1)},
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Token-2022 mints found:", len(out))
	for i, acc := range out {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d: %s\n", i, acc.Pubkey)
	}
}

func ptrU64(v uint64) *uint64 { return &v }
