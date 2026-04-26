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
	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go/rpc"
)

func main() {
	ctx := context.Background()
	client := rpc.New(rpc.MainNetBeta_RPC)

	// Fetch a fresh signature so the example keeps working as the
	// cluster advances. The SPL Token program always has recent activity
	// on mainnet-beta.
	limit := 1
	sigs, err := client.GetSignaturesForAddressWithOpts(
		ctx,
		solana.TokenProgramID,
		&rpc.GetSignaturesForAddressOpts{Limit: &limit},
	)
	if err != nil {
		panic(fmt.Errorf("getSignaturesForAddress: %w", err))
	}
	if len(sigs) == 0 {
		panic("no recent signatures found")
	}
	txSig := sigs[0].Signature
	fmt.Println("fetching tx:", txSig)

	maxVersion := uint64(0)
	{
		out, err := client.GetTransaction(
			ctx,
			txSig,
			&rpc.GetTransactionOpts{
				MaxSupportedTransactionVersion: &maxVersion,
			},
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(out)
		spew.Dump(out.Transaction.GetTransaction())
	}
	{
		out, err := client.GetTransaction(
			ctx,
			txSig,
			&rpc.GetTransactionOpts{
				Encoding:                       solana.EncodingJSON,
				MaxSupportedTransactionVersion: &maxVersion,
			},
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(out)
		spew.Dump(out.Transaction.GetTransaction())
	}
	{
		out, err := client.GetTransaction(
			ctx,
			txSig,
			&rpc.GetTransactionOpts{
				Encoding:                       solana.EncodingBase64,
				MaxSupportedTransactionVersion: &maxVersion,
			},
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(out)
		spew.Dump(out.Transaction.GetBinary())

		decodedTx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
		if err != nil {
			panic(err)
		}
		spew.Dump(decodedTx)
	}
}
