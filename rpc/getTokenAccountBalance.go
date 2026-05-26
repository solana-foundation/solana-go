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

package rpc

import (
	"context"

	"github.com/gagliardetto/solana-go"
)

// GetTokenAccountBalance returns the token balance of an SPL Token account.
func (cl *Client) GetTokenAccountBalance(
	ctx context.Context,
	account solana.PublicKey,
	commitment CommitmentType, // optional
) (out *GetTokenAccountBalanceResult, err error) {
	return cl.GetTokenAccountBalanceWithOpts(ctx, account, &GetTokenAccountBalanceOpts{Commitment: commitment})
}

// GetTokenAccountBalanceOpts is the optional configuration object for
// `getTokenAccountBalance`, mirroring the JSON-RPC spec for this method.
//
// See https://solana.com/docs/rpc/http/gettokenaccountbalance.
type GetTokenAccountBalanceOpts struct {
	// Commitment level to query the balance at.
	Commitment CommitmentType

	// MinContextSlot is the minimum slot at which the RPC node should
	// have processed the request. The validator returns a
	// `MinContextSlotNotReached` error to the caller if the local slot
	// has not yet caught up, instead of silently serving stale state.
	MinContextSlot *uint64
}

// GetTokenAccountBalanceWithOpts returns the token balance of an SPL
// Token account, with the full set of optional configuration fields
// the `getTokenAccountBalance` JSON-RPC method accepts.
func (cl *Client) GetTokenAccountBalanceWithOpts(
	ctx context.Context,
	account solana.PublicKey,
	opts *GetTokenAccountBalanceOpts,
) (out *GetTokenAccountBalanceResult, err error) {
	params := []any{account}
	if opts != nil {
		obj := M{}
		if opts.Commitment != "" {
			obj["commitment"] = opts.Commitment
		}
		if opts.MinContextSlot != nil {
			obj["minContextSlot"] = *opts.MinContextSlot
		}
		if len(obj) > 0 {
			params = append(params, obj)
		}
	}
	err = cl.rpcClient.CallForInto(ctx, &out, "getTokenAccountBalance", params)
	return
}

type GetTokenAccountBalanceResult struct {
	RPCContext
	Value *UiTokenAmount `json:"value"`
}
