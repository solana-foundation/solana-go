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
	"os"

	json "github.com/goccy/go-json"
	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/text"
	"github.com/spf13/cobra"
)

var getProgramAccountsCmd = &cobra.Command{
	Use:   "program-accounts {program_addr}",
	Short: "Retrieve info about an account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()

		resp, err := client.GetProgramAccounts(
			cmd.Context(),
			solana.MustPublicKeyFromBase58(args[0]),
		)
		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("program account not found")
		}

		for _, keyedAcct := range resp {
			acct := keyedAcct.Account

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
				return nil
			}

			if err := text.NewEncoder(os.Stdout).Encode(acct, nil); err != nil {
				return fmt.Errorf("unable to text encode account: %w", err)
			}
		}

		return nil
	},
}

func init() {
	getCmd.AddCommand(getProgramAccountsCmd)
}
