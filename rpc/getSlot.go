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
)

// GetSlot returns the slot that has reached the given or default commitment level.
func (cl *Client) GetSlot(
	ctx context.Context,
	commitment CommitmentType, // optional
) (out uint64, err error) {
	return cl.GetSlotWithOpts(ctx, &GetSlotOpts{Commitment: commitment})
}

// GetSlotOpts is the optional configuration object for `getSlot`,
// mirroring the JSON-RPC spec for this method.
//
// See https://solana.com/docs/rpc/http/getslot.
type GetSlotOpts struct {
	// Commitment level to query the slot at.
	Commitment CommitmentType

	// MinContextSlot is the minimum slot at which the RPC node should
	// have processed the request. The validator returns a
	// `MinContextSlotNotReached` error to the caller if the local slot
	// has not yet caught up, instead of silently serving stale state.
	MinContextSlot *uint64
}

// GetSlotWithOpts returns the slot that has reached the given or
// default commitment level, with the full set of optional configuration
// fields the `getSlot` JSON-RPC method accepts.
func (cl *Client) GetSlotWithOpts(
	ctx context.Context,
	opts *GetSlotOpts,
) (out uint64, err error) {
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

	err = cl.rpcClient.CallForInto(ctx, &out, "getSlot", params)
	return
}
