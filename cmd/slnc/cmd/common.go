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
	"strings"

	"github.com/solana-foundation/solana-go/v2/rpc"

	"github.com/solana-foundation/solana-go/v2/vault"
	"github.com/spf13/viper"
)

func getClient() *rpc.Client {
	httpHeaders := viper.GetStringSlice("global-http-header")

	for i := 0; i < 25; i++ {
		if val := os.Getenv(fmt.Sprintf("SLNC_GLOBAL_HTTP_HEADER_%d", i)); val != "" {
			httpHeaders = append(httpHeaders, val)
		}
	}

	headers := make(map[string]string)
	for _, header := range httpHeaders {
		headerArray := strings.SplitN(header, ": ", 2)
		if len(headerArray) != 2 || strings.Contains(headerArray[0], " ") {
			errorCheck("validating http headers", fmt.Errorf("invalid HTTP Header format"))
		}
		headers[headerArray[0]] = headerArray[1]
	}
	api := rpc.NewWithHeaders(sanitizeAPIURL(viper.GetString("global-rpc-url")), headers)
	return api
}

func sanitizeAPIURL(input string) string {
	switch input {
	case "devnet":
		return "https://api.devnet.solana.com"
	case "testnet":
		return "https://api.testnet.solana.com"
	case "mainnet":
		return "https://api.mainnet-beta.solana.com"
	}
	return strings.TrimRight(input, "/")
}

func errorCheck(prefix string, err error) {
	if err != nil {
		fmt.Printf("ERROR: %s: %s\n", prefix, err)
		if strings.HasSuffix(err.Error(), "connection refused") && strings.Contains(err.Error(), defaultRPCURL) {
			fmt.Println("Have you selected a valid Solana JSON-RPC endpoint ? You can use the --rpc-url flag or SLNC_GLOBAL_RPC_URL environment variable.")
		}
		os.Exit(1)
	}
}

func mustGetWallet() *vault.Vault {
	vault, err := setupWallet()
	errorCheck("wallet setup", err)
	return vault
}

func setupWallet() (*vault.Vault, error) {
	walletFile := viper.GetString("global-vault-file")
	if _, err := os.Stat(walletFile); err != nil {
		return nil, fmt.Errorf("wallet file %q missing: %w", walletFile, err)
	}

	v, err := vault.NewVaultFromWalletFile(walletFile)
	if err != nil {
		return nil, fmt.Errorf("loading vault: %w", err)
	}

	boxer, err := vault.SecretBoxerForType(v.SecretBoxWrap, viper.GetString("global-kms-gcp-keypath"))
	if err != nil {
		return nil, fmt.Errorf("secret boxer: %w", err)
	}

	if err := v.Open(boxer); err != nil {
		return nil, fmt.Errorf("opening: %w", err)
	}

	return v, nil
}
