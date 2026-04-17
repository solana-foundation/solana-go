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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Version represents the cmd command version
var Version string

// const defaultRPCURL = "http://localhost:8899"
const defaultRPCURL = "http://api.mainnet-beta.solana.com"

// RootCmd represents the eosc command
var RootCmd = &cobra.Command{
	Use:   "cmd",
	Short: "cmd is a client to Solana clusters - by dfuse.io",
}

// Execute executes the configured RootCmd
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	// Not implemnted
	//RootCmd.PersistentFlags().BoolP("debug", "", false, "Enables verbose API debug messages")
	RootCmd.PersistentFlags().StringP("vault-file", "", "./solana-vault.json", "Wallet file that contains encrypted key material")
	RootCmd.PersistentFlags().StringP("rpc-url", "u", defaultRPCURL, "API endpoint of eos.io blockchain node")
	RootCmd.PersistentFlags().StringSliceP("http-header", "H", []string{}, "HTTP header to add to JSON-RPC requests")
	RootCmd.PersistentFlags().StringP("kms-gcp-keypath", "", "", "Path to the cryptoKeys within a keyRing on GCP")

	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		SetupLogger()
		return nil
	}
}

func initConfig() {
	viper.SetEnvPrefix("SLNC")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	recurseViperCommands(RootCmd, nil)
}

func recurseViperCommands(root *cobra.Command, segments []string) {
	// Stolen from: github.com/abourget/viperbind
	var segmentPrefix string
	if len(segments) > 0 {
		segmentPrefix = strings.Join(segments, "-") + "-"
	}

	root.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		newVar := segmentPrefix + "global-" + f.Name
		viper.BindPFlag(newVar, f)
	})
	root.Flags().VisitAll(func(f *pflag.Flag) {
		newVar := segmentPrefix + "cmd-" + f.Name
		viper.BindPFlag(newVar, f)
	})

	for _, cmd := range root.Commands() {
		recurseViperCommands(cmd, append(segments, cmd.Name()))
	}
}
