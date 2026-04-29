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

	"github.com/gagliardetto/solana-go"
	json "github.com/goccy/go-json"
	"github.com/spf13/cobra"
)

var getAccountCmd = &cobra.Command{
	Use:   "account {account_addr}",
	Short: "Retrieve info about an account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		ctx := context.Background()

		resp, err := client.GetAccountInfo(ctx, solana.MustPublicKeyFromBase58(args[0]))
		if err != nil {
			return err
		}

		acct := resp.Value
		var data []byte
		if data, err = json.MarshalIndent(acct, "", "  "); err != nil {
			return fmt.Errorf("unable to marshall account information: %w", err)
		}

		fmt.Println(string(data))

		obj, err := decode(acct.Owner, acct.Data.GetBinary())
		if err != nil {
			return err
		}

		if obj != nil {
			cnt, err := json.MarshalIndent(obj, "", "  ")
			if err != nil {
				return err
			}
			fmt.Printf("Data %T: %s\n", obj, string(cnt))
		}

		return nil
	},
}

func init() {
	getCmd.AddCommand(getAccountCmd)
}
