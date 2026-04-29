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

	"github.com/gagliardetto/solana-go/rpc"

	"github.com/spf13/viper"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/tokenregistry"
	"github.com/spf13/cobra"
)

var tokenRegistryRegisterCmd = &cobra.Command{
	Use:   "register {token-address} {name} {symbol} {logo} {website}",
	Short: "register meta data for a token",
	Args:  cobra.ExactArgs(5),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		vault := mustGetWallet()
		client := getClient()

		var tokenAddress solana.PublicKey
		if tokenAddress, err = solana.PublicKeyFromBase58(args[0]); err != nil {
			return fmt.Errorf("invalid token address %q: %w", args[0], err)
		}

		var logo tokenregistry.Logo
		var name tokenregistry.Name
		var symbol tokenregistry.Symbol
		var website tokenregistry.Website

		if name, err = tokenregistry.NameFromString(args[1]); err != nil {
			return fmt.Errorf("invalid name %q: %w", args[1], err)
		}

		if symbol, err = tokenregistry.SymbolFromString(args[2]); err != nil {
			return fmt.Errorf("invalid symbol %q: %w", args[2], err)
		}

		if logo, err = tokenregistry.LogoFromString(args[3]); err != nil {
			return fmt.Errorf("invalid logo %q: %w", args[3], err)
		}

		if website, err = tokenregistry.WebsiteFromString(args[4]); err != nil {
			return fmt.Errorf("invalid website %q: %w", args[4], err)
		}

		pkeyStr := viper.GetString("token-registry-register-cmd-registrar")
		if pkeyStr == "" {
			return fmt.Errorf("unable to continue without a specified registrar")
		}

		// TODO: we shoul check on chain if the symbol exists

		registrarPubKey, err := solana.PublicKeyFromBase58(pkeyStr)
		if err != nil {
			return fmt.Errorf("invalid registrar key %q: %w", pkeyStr, err)
		}

		found := false
		for _, privateKey := range vault.KeyBag {
			if privateKey.PublicKey() == registrarPubKey {
				found = true
			}
		}

		if !found {
			return fmt.Errorf("registrar key must be present in the vault to register a token")
		}

		blockHashResult, err := client.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
		if err != nil {
			return fmt.Errorf("unable retrieve latest blockhash: %w", err)
		}

		tokenMetaAccount := solana.NewWallet()

		lamport, err := client.GetMinimumBalanceForRentExemption(
			context.Background(),
			tokenregistry.TOKEN_META_SIZE,
			rpc.CommitmentFinalized,
		)
		if err != nil {
			return fmt.Errorf("unable to retrieve lapoint rent: %w", err)
		}

		tokenRegistryProgramID := tokenregistry.ProgramID()

		createAccountInstruction := system.NewCreateAccountInstruction(
			lamport,
			tokenregistry.TOKEN_META_SIZE,
			tokenRegistryProgramID,
			registrarPubKey,
			tokenMetaAccount.PublicKey(),
		).
			Build()
		registerTokenInstruction := tokenregistry.NewRegisterTokenInstruction(logo, name, symbol, website, tokenMetaAccount.PublicKey(), registrarPubKey, tokenAddress)

		trx, err := solana.NewTransaction(
			[]solana.Instruction{
				createAccountInstruction,
				registerTokenInstruction,
			},
			blockHashResult.Value.Blockhash,
			solana.TransactionPayer(registrarPubKey),
		)
		if err != nil {
			return fmt.Errorf("unable to craft transaction: %w", err)
		}

		_, err = trx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
			// create account need to be signed by the private key of the new account
			// that is not in the vault and will be lost after the execution.
			if key == tokenMetaAccount.PublicKey() {
				return &tokenMetaAccount.PrivateKey
			}

			for _, k := range vault.KeyBag {
				if k.PublicKey() == key {
					return &k
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("unable to sign transaction: %w", err)
		}

		trxHash, err := client.SendTransaction(cmd.Context(), trx)
		if err != nil {
			return fmt.Errorf("unable to send transaction: %w", err)
		}

		fmt.Printf("Token Register successfully, with transaction hash: %s\n", trxHash)
		fmt.Printf("  Mint Address Registered: %s\n", tokenAddress.String())
		fmt.Printf("  Token Registry Meta Address: %s\n", tokenMetaAccount.PublicKey().String())
		return nil
	},
}

func init() {
	tokenRegistryCmd.AddCommand(tokenRegistryRegisterCmd)
	tokenRegistryRegisterCmd.PersistentFlags().String("registrar", "9hFtYBYmBJCVguRYs9pBTWKYAFoKfjYR7zBPpEkVsmD", "The public key that will register the token")
}
