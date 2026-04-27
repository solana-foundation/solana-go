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
	"log"

	"github.com/solana-foundation/solana-go/v2"

	"github.com/solana-foundation/solana-go/v2/cli"
	"github.com/solana-foundation/solana-go/v2/vault"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var vaultAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add private keys to an existing vault taking input from the shell",
	Run: func(cmd *cobra.Command, args []string) {

		walletFile := viper.GetString("global-vault-file")

		fmt.Println("Loading existing vault from file:", walletFile)
		v, err := vault.NewVaultFromWalletFile(walletFile)
		if err != nil {
			log.Fatalf("unable to load vault file: %s", err)
		}

		boxer, err := vault.SecretBoxerForType(v.SecretBoxWrap, viper.GetString("global-kms-gcp-keypath"))
		if err != nil {
			log.Fatalf("unable to initiate boxer: %s", err)
		}

		err = v.Open(boxer)
		if err != nil {
			log.Fatalf("unable to open vault: %s", err)
		}

		v.PrintPublicKeys()

		privateKeys, err := capturePrivateKeys()
		if err != nil {
			log.Fatalf("failed to enter private keys: %s", err)
		}

		var newKeys []solana.PublicKey
		for _, privateKey := range privateKeys {
			v.AddPrivateKey(privateKey)
			newKeys = append(newKeys, privateKey.PublicKey())
		}

		err = v.Seal(boxer)
		if err != nil {
			log.Fatalf("failed to seal vault: %s", err)
		}

		err = v.WriteToFile(walletFile)
		if err != nil {
			log.Fatalf("failed to write vault file: %s", err)
		}

		vaultWrittenReport(walletFile, newKeys, len(v.KeyBag))
	},
}

func init() {
	vaultCmd.AddCommand(vaultAddCmd)
}

func capturePrivateKeys() (out []solana.PrivateKey, err error) {
	fmt.Println("")
	fmt.Println("PLEASE READ:")
	fmt.Println("We are now going to ask you to paste your private keys, one at a time.")
	fmt.Println("They will not be shown on screen.")
	fmt.Println("Please verify that the public keys printed on screen correspond to what you have noted")
	fmt.Println("")

	first := true
	for {
		privKey, err := capturePrivateKey(first)
		if err != nil {
			return out, fmt.Errorf("capture privkeys: %w", err)
		}
		first = false

		if privKey == nil {
			return out, nil
		}
		out = append(out, privKey)
	}
}

func capturePrivateKey(isFirst bool) (privateKey solana.PrivateKey, err error) {
	prompt := "Paste your first private key: "
	if !isFirst {
		prompt = "Paste your next private key or hit ENTER if you are done: "
	}

	enteredKey, err := cli.GetPassword(prompt)
	if err != nil {
		return nil, fmt.Errorf("get private key: %w", err)
	}

	if enteredKey == "" {
		return nil, nil
	}

	key, err := solana.PrivateKeyFromBase58(enteredKey)
	if err != nil {
		return nil, fmt.Errorf("import private key: %w", err)
	}

	fmt.Printf("- Scanned private key corresponding to %s\n", key.PublicKey().String())

	return key, nil
}

func vaultWrittenReport(walletFile string, newKeys []solana.PublicKey, totalKeys int) {
	fmt.Println("")
	fmt.Printf("Wallet file %q written to disk.\n", walletFile)
	fmt.Println("Here are the keys that were ADDED during this operation (use `list` to see them all):")
	for _, pub := range newKeys {
		fmt.Printf("- %s\n", pub.String())
	}

	fmt.Printf("Total keys stored: %d\n", totalKeys)
}
