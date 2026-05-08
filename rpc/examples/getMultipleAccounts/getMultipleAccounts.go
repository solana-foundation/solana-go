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
	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/rpc"
)

func main() {
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)

	{
		out, err := client.GetMultipleAccounts(
			context.TODO(),
			solana.MustPublicKeyFromBase58("SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt"),  // serum token
			solana.MustPublicKeyFromBase58("4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R"), // raydium token
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(out)
	}
	{
		out, err := client.GetMultipleAccountsWithOpts(
			context.TODO(),
			[]solana.PublicKey{solana.MustPublicKeyFromBase58("SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt"), // serum token
				solana.MustPublicKeyFromBase58("4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R"), // raydium token
			},
			&rpc.GetMultipleAccountsOpts{
				Encoding:   solana.EncodingBase64Zstd,
				Commitment: rpc.CommitmentFinalized,
				// You can get just a part of the account data by specify a DataSlice:
				// DataSlice: &rpc.DataSlice{
				// 	Offset: pointer.ToUint64(0),
				// 	Length: pointer.ToUint64(1024),
				// },
			},
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(out)
	}
}
