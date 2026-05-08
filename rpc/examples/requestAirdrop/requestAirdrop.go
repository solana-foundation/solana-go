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

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/rpc"
)

func main() {
	ctx := context.Background()
	// The testnet airdrop faucet is frequently dry; devnet is reliable.
	client := rpc.New(rpc.DevNet_RPC)

	wallet := solana.NewWallet()
	fmt.Println("airdropping to:", wallet.PublicKey())

	sig, err := client.RequestAirdrop(
		ctx,
		wallet.PublicKey(),
		solana.LAMPORTS_PER_SOL, // 1 SOL
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(fmt.Errorf("airdrop: %w", err))
	}

	fmt.Println("airdrop signature:", sig)
}
