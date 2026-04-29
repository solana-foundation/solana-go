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
	"strings"

	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

var tokenListMintsCmd = &cobra.Command{
	Use:   "mints",
	Short: "Lists mints",
	RunE: func(cmd *cobra.Command, _ []string) error {
		rpcCli := getClient()

		mints, err := token.FetchMints(cmd.Context(), rpcCli)
		if err != nil {
			return fmt.Errorf("unable to retrieve mints: %w", err)
		}
		out := []string{"Mint | Decimals | Supply | Token Authority | Freeze Authority"}
		for _, m := range mints {
			line := []string{
				fmt.Sprintf("%d", m.Supply),
				fmt.Sprintf("%d", m.Decimals),
			}
			if m.MintAuthority != nil {
				line = append(line, m.MintAuthority.String())
			} else {
				line = append(line, "No mint authority")
			}
			if m.FreezeAuthority != nil {
				line = append(line, m.FreezeAuthority.String())
			} else {
				line = append(line, "No freeze authority")
			}
			out = append(out, strings.Join(line, " | "))
		}

		fmt.Println(columnize.Format(out, nil))
		return nil
	},
}

func init() {
	tokenListCmd.AddCommand(tokenListMintsCmd)
}
