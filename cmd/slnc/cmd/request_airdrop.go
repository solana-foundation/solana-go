// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
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

package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/rpc"
	"github.com/spf13/cobra"
)

var requestCmd = &cobra.Command{
	Use:   "request-airdrop {address} {lamport}",
	Short: "Request lamport airdrop for an account",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		client := getClient()

		address, err := solana.PublicKeyFromBase58(args[0])
		if err != nil {
			return fmt.Errorf("invalid account address %q: %w", args[0], err)
		}

		lamport, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid lamport value, expected a int value, got : %s", args[1])
		}

		airDrop, err := client.RequestAirdrop(
			context.Background(),
			address,
			uint64(lamport),
			rpc.CommitmentFinalized,
		)
		if err != nil {
			return fmt.Errorf("airdrop request failed: %w", err)
		}

		fmt.Println("Air drop succeeded, transaction hash:", airDrop)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(requestCmd)
}
