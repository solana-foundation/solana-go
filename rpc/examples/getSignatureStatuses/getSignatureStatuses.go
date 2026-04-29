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
	"github.com/gagliardetto/solana-go/rpc"
)

func main() {
	ctx := context.Background()
	client := rpc.New(rpc.MainNetBeta_RPC)

	// Fetch fresh signatures so the example keeps working as the
	// cluster advances.
	limit := 2
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

	toLookup := make([]solana.Signature, 0, len(sigs))
	for _, s := range sigs {
		toLookup = append(toLookup, s.Signature)
	}
	fmt.Println("querying statuses for", len(toLookup), "signatures")

	out, err := client.GetSignatureStatuses(
		ctx,
		true, // searchTransactionHistory
		toLookup...,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}
