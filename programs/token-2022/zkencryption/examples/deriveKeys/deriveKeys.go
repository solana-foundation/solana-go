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

// Command deriveKeys shows how to deterministically derive the two key
// materials used by the Token-2022 confidential-transfer extension from a
// Solana signer: the ElGamal secret key and the AES (AeKey) key.
//
// This is the primary derivation path used by wallets: the same signer and
// public seed always derive the same keys, so a user can recover their
// confidential-transfer keys from their wallet alone, with nothing stored
// on-chain or off-chain. The keys produced here are byte-for-byte identical
// to those from the Rust solana-zk-sdk and the JS/WASM @solana/zk-sdk, given
// the same signer and seed.
package main

import (
	"encoding/hex"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token-2022/zkencryption"
)

func main() {
	// In a real application this is the user's wallet key (or a hardware
	// wallet / remote signer implementing zkencryption.Signer). solana.PrivateKey
	// already satisfies the Signer interface. We use a fixed seed here only so
	// the example output is reproducible.
	wallet := solana.NewWallet().PrivateKey

	// The standard public seed is empty: keys are bound to the wallet alone,
	// so one signature recovers the keys for every confidential balance the
	// wallet owns, and they match what other standard clients (Rust
	// solana-zk-sdk, JS @solana/zk-sdk, @solana-program/token-2022) derive for
	// the same wallet. Non-empty seeds are supported for finer-grained key
	// scoping, but keys derived with them will not match standard clients.
	publicSeed := []byte{}

	elgamal, err := zkencryption.ElGamalSecretKeyFromSigner(wallet, publicSeed)
	if err != nil {
		panic(err)
	}

	aeKey, err := zkencryption.AeKeyFromSigner(wallet, publicSeed)
	if err != nil {
		panic(err)
	}

	fmt.Println("wallet:            ", wallet.PublicKey())
	fmt.Println("ElGamal secret key:", hex.EncodeToString(elgamal[:]))
	fmt.Println("AeKey:             ", hex.EncodeToString(aeKey[:]))

	// Derivation is deterministic: signer + public seed always yield the same
	// keys, which is what lets a wallet recover them on demand.
	again, err := zkencryption.ElGamalSecretKeyFromSigner(wallet, publicSeed)
	if err != nil {
		panic(err)
	}
	fmt.Println("deterministic:     ", elgamal == again)
}
