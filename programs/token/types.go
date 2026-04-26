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

package token

import (
	ag_binary "github.com/gagliardetto/solana-go/binary"
)

type AuthorityType ag_binary.BorshEnum

const (
	// Authority to mint new tokens
	AuthorityMintTokens AuthorityType = iota

	// Authority to freeze any account associated with the Mint
	AuthorityFreezeAccount

	// Owner of a given token account
	AuthorityAccountOwner

	// Authority to close a token account
	AuthorityCloseAccount
)

type AccountState ag_binary.BorshEnum

const (
	// Account is not yet initialized
	Uninitialized AccountState = iota

	// Account is initialized; the account owner and/or delegate may perform permitted operations
	// on this account
	Initialized

	// Account has been frozen by the mint freeze authority. Neither the account owner nor
	// the delegate are able to perform operations on this account.
	Frozen
)
