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

	"github.com/solana-foundation/solana-go/v2"
	"github.com/spf13/cobra"
)

var getBalanceCmd = &cobra.Command{
	Use:   "balance {account_addr}",
	Short: "Retrieve an account's balance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()

		resp, err := client.GetBalance(
			cmd.Context(),
			solana.MustPublicKeyFromBase58(args[0]),
			"",
		)
		if err != nil {
			return err
		}

		if resp.Value == 0 {
			return fmt.Errorf("account not found")
		}

		fmt.Println(resp.Value, "lamports")

		return nil
	},
}

func init() {
	getCmd.AddCommand(getBalanceCmd)
}
