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
)

// GetStakeMinimumDelegation returns the stake minimum delegation, in lamports.
func (cl *Client) GetStakeMinimumDelegation(
	ctx context.Context,
	commitment CommitmentType, // optional
) (out *GetStakeMinimumDelegationResult, err error) {
	return cl.GetStakeMinimumDelegationWithOpts(ctx, &GetStakeMinimumDelegationOpts{Commitment: commitment})
}

// GetStakeMinimumDelegationOpts groups the optional configuration accepted by
// the getStakeMinimumDelegation RPC.
type GetStakeMinimumDelegationOpts struct {
	Commitment CommitmentType

	// The minimum slot that the request can be evaluated at.
	MinContextSlot *uint64
}

// GetStakeMinimumDelegationWithOpts is the variant of GetStakeMinimumDelegation
// that accepts the full optional configuration set, including MinContextSlot.
func (cl *Client) GetStakeMinimumDelegationWithOpts(
	ctx context.Context,
	opts *GetStakeMinimumDelegationOpts,
) (out *GetStakeMinimumDelegationResult, err error) {
	params := []interface{}{}
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
	err = cl.rpcClient.CallForInto(ctx, &out, "getStakeMinimumDelegation", params)
	return
}
