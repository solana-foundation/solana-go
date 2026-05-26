// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package rpc

import (
	"context"

	"github.com/gagliardetto/solana-go"
)

// GetBalance returns the balance of the account of provided publicKey.
func (cl *Client) GetBalance(
	ctx context.Context,

	// Pubkey of account to query. Required.
	publicKey solana.PublicKey,

	// Commitment requirement. Optional.
	commitment CommitmentType,
) (out *GetBalanceResult, err error) {
	return cl.GetBalanceWithOpts(ctx, publicKey, &GetBalanceOpts{Commitment: commitment})
}

// GetBalanceOpts is the optional configuration object for `getBalance`,
// mirroring the JSON-RPC spec for this method.
//
// See https://solana.com/docs/rpc/http/getbalance for the canonical field list.
type GetBalanceOpts struct {
	// Commitment level to query the balance at.
	Commitment CommitmentType

	// MinContextSlot is the minimum slot at which the RPC node should
	// have processed the request. The validator returns a
	// `MinContextSlotNotReached` error to the caller if the local slot
	// has not yet caught up, instead of silently serving stale state.
	MinContextSlot *uint64
}

// GetBalanceWithOpts returns the balance of the account of the provided
// publicKey, with the full set of optional configuration fields the
// `getBalance` JSON-RPC method accepts.
//
// Use this over `GetBalance` when you need `MinContextSlot` (closes
// the parity gap with `GetAccountInfo` / `GetMultipleAccounts` after
// #431; see #245 for the broader request).
func (cl *Client) GetBalanceWithOpts(
	ctx context.Context,
	publicKey solana.PublicKey,
	opts *GetBalanceOpts,
) (out *GetBalanceResult, err error) {
	params := []any{publicKey}
	if opts != nil {
		obj := M{}
		if opts.Commitment != "" {
			obj["commitment"] = string(opts.Commitment)
		}
		if opts.MinContextSlot != nil {
			obj["minContextSlot"] = *opts.MinContextSlot
		}
		if len(obj) > 0 {
			params = append(params, obj)
		}
	}

	err = cl.rpcClient.CallForInto(ctx, &out, "getBalance", params)
	return
}
