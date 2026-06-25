// Copyright 2022 github.com/gagliardetto
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
)

// GetFeeForMessage returns the fee the network will charge for a particular Message.
func (cl *Client) GetFeeForMessage(
	ctx context.Context,
	message string, // Base-64 encoded Message
	commitment CommitmentType, // optional
) (out *GetFeeForMessageResult, err error) {
	return cl.GetFeeForMessageWithOpts(ctx, message, &GetFeeForMessageOpts{Commitment: commitment})
}

// GetFeeForMessageOpts groups the optional configuration accepted by the
// getFeeForMessage RPC.
type GetFeeForMessageOpts struct {
	Commitment CommitmentType

	// The minimum slot that the request can be evaluated at.
	MinContextSlot *uint64
}

// GetFeeForMessageWithOpts is the variant of GetFeeForMessage that accepts the
// full optional configuration set, including MinContextSlot.
func (cl *Client) GetFeeForMessageWithOpts(
	ctx context.Context,
	message string, // Base-64 encoded Message
	opts *GetFeeForMessageOpts,
) (out *GetFeeForMessageResult, err error) {
	params := []any{message}
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
	err = cl.rpcClient.CallForInto(ctx, &out, "getFeeForMessage", params)
	return
}

type GetFeeForMessageResult struct {
	RPCContext

	// Fee corresponding to the message at the specified blockhash.
	Value *uint64 `json:"value"`
}
