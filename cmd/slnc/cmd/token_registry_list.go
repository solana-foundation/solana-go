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

	"github.com/ryanuber/columnize"

	"github.com/solana-foundation/solana-go/v2/programs/tokenregistry"
	_ "github.com/solana-foundation/solana-go/v2/programs/tokenregistry"
	"github.com/spf13/cobra"
)

var tokenRegistryListCmd = &cobra.Command{
	Use:   "list",
	Short: "Retrieve token register entries",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()

		entries, err := tokenregistry.GetEntries(cmd.Context(), client)
		if err != nil {
			return fmt.Errorf("unable to retrieve entries: %w", err)
		}

		out := []string{"Is Initialized | Mint Address | Registration Authority | Logo | Name | Symbol | Website"}

		for _, e := range entries {
			initialized := "false"
			if e.IsInitialized {
				initialized = "true"
			}

			line := []string{
				initialized,
				e.MintAddress.String(),
				e.RegistrationAuthority.String(),
				e.Logo.String(),
				e.Name.String(),
				e.Symbol.String(),
				e.Website.String(),
			}
			out = append(out, strings.Join(line, " | "))
		}

		fmt.Println(columnize.Format(out, nil))
		return nil
	},
}

func init() {
	tokenRegistryCmd.AddCommand(tokenRegistryListCmd)
}
