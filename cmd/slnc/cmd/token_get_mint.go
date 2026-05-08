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
	"fmt"

	"github.com/ryanuber/columnize"
	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/token"
	"github.com/spf13/cobra"
)

var tokenGetMintCmd = &cobra.Command{
	Use:   "mint {mint_addr}",
	Short: "Retrieves mint information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		mintAddress, err := solana.PublicKeyFromBase58(args[0])
		if err != nil {
			return fmt.Errorf("decoding mint addr: %w", err)
		}

		client := getClient()

		acct, err := client.GetAccountInfo(ctx, mintAddress)
		if err != nil {
			return fmt.Errorf("couldn't get account data: %w", err)
		}

		mint := &token.Mint{}
		if err := mint.Decode(acct.Value.Data.GetBinary()); err != nil {
			return fmt.Errorf("unable to retrieve int information: %w", err)
		}

		if !mint.IsInitialized {
			fmt.Println("Uninitialized mint. Data length", len(acct.Value.Data.GetBinary()))
			return nil
		}

		var out []string

		out = append(out, fmt.Sprintf("Supply | %d", mint.Supply))
		out = append(out, fmt.Sprintf("Decimals | %d", mint.Decimals))

		if mint.MintAuthority != nil {
			out = append(out, fmt.Sprintf("Token Authority | %s", mint.MintAuthority))
		} else {
			out = append(out, "No mint authority")
		}

		if mint.FreezeAuthority != nil {
			out = append(out, fmt.Sprintf("Freeze Authority | %s", mint.FreezeAuthority))
		} else {
			out = append(out, "No freeze authority")
		}

		fmt.Println(columnize.Format(out, nil))

		return nil
	},
}

func init() {
	tokenGetCmd.AddCommand(tokenGetMintCmd)
}
