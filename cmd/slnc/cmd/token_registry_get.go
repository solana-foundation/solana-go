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
	"errors"
	"fmt"
	"os"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/tokenregistry"
	"github.com/solana-foundation/solana-go/v2/rpc"
	"github.com/solana-foundation/solana-go/v2/text"
	"github.com/spf13/cobra"
)

var tokenRegistryGetCmd = &cobra.Command{
	Use:   "get {mint-address}",
	Short: "Retrieve token meta for a specific token meta account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()

		address := args[0]
		pubKey, err := solana.PublicKeyFromBase58(address)
		if err != nil {
			return fmt.Errorf("invalid mint address %q: %w", address, err)
		}

		t, err := tokenregistry.GetTokenRegistryEntry(cmd.Context(), client, pubKey)
		if err != nil {
			if errors.Is(err, rpc.ErrNotFound) {
				fmt.Printf("No token registry entry found for given mint %q", pubKey.String())
				return nil
			}
			return fmt.Errorf("unable to retrieve token registry entry for mint %q: %w", pubKey.String(), err)
		}

		err = text.NewEncoder(os.Stdout).Encode(t, nil)
		if err != nil {
			return fmt.Errorf("unable to text encode token registry entry: %w", err)
		}

		return nil
	},
}

func init() {
	tokenRegistryCmd.AddCommand(tokenRegistryGetCmd)
}
