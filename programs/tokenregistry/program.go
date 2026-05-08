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

package tokenregistry

import (
	"os"

	"github.com/solana-foundation/solana-go/v2"
)

var mainnetProgramID = solana.MustPublicKeyFromBase58("CmPVzy88JSB4S223yCvFmBxTLobLya27KgEDeNPnqEub")
var testnetProgramID = solana.MustPublicKeyFromBase58("99999999999999999999999999999999999999999999")
var devnetProgramID = solana.MustPublicKeyFromBase58("99999999999999999999999999999999999999999999")

func ProgramID() solana.PublicKey {

	if custom := os.Getenv("TOKEN_REGISTRY_PROGRAM_ID"); custom != "" {
		return solana.MustPublicKeyFromBase58(custom)
	}

	network := os.Getenv("SOL_NETWORK")

	switch network {
	case "mainnet":
		return mainnetProgramID
	case "testnet":
		return testnetProgramID
	case "devnet":
		return devnetProgramID
	default:
		return mainnetProgramID
	}
}
