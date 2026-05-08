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

package cmd

import (
	"fmt"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/spf13/cobra"
)

var isBlockhashValidCmd = &cobra.Command{
	Use:   "isblockhashvalid {blockhash}",
	Short: "Checks if a given blockhash is valid",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()

		resp, err := client.IsBlockhashValid(
			cmd.Context(),
			solana.MustHashFromBase58(args[0]),
			"",
		)
		if err != nil {
			return err
		}

		fmt.Println(resp.Value)

		return nil
	},
}

func init() {
	getCmd.AddCommand(isBlockhashValidCmd)
}
