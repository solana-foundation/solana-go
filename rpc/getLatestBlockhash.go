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

// GetLatestBlockhash returns the latest blockhash.
func (cl *Client) GetLatestBlockhash(
	ctx context.Context,
	commitment CommitmentType, // optional
) (out *GetLatestBlockhashResult, err error) {
	return cl.GetLatestBlockhashWithOpts(ctx, &GetLatestBlockhashOpts{Commitment: commitment})
}

// GetLatestBlockhashOpts is the optional configuration object for
// `getLatestBlockhash`, mirroring the JSON-RPC spec for this method.
//
// See https://solana.com/docs/rpc/http/getlatestblockhash.
type GetLatestBlockhashOpts struct {
	// Commitment level to query the blockhash at.
	Commitment CommitmentType

	// MinContextSlot is the minimum slot at which the RPC node should
	// have processed the request. The validator returns a
	// `MinContextSlotNotReached` error to the caller if the local slot
	// has not yet caught up, instead of silently serving stale state.
	MinContextSlot *uint64
}

// GetLatestBlockhashWithOpts returns the latest blockhash, with the
// full set of optional configuration fields the `getLatestBlockhash`
// JSON-RPC method accepts.
func (cl *Client) GetLatestBlockhashWithOpts(
	ctx context.Context,
	opts *GetLatestBlockhashOpts,
) (out *GetLatestBlockhashResult, err error) {
	params := []any{}
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

	err = cl.rpcClient.CallForInto(ctx, &out, "getLatestBlockhash", params)
	return
}

type GetLatestBlockhashResult struct {
	RPCContext
	Value *LatestBlockhashResult `json:"value"`
}

type LatestBlockhashResult struct {
	Blockhash            solana.Hash `json:"blockhash"`
	LastValidBlockHeight uint64      `json:"lastValidBlockHeight"` // Slot.
}
