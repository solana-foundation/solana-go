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

type LargestAccountsFilterType string

const (
	LargestAccountsFilterCirculating    LargestAccountsFilterType = "circulating"
	LargestAccountsFilterNonCirculating LargestAccountsFilterType = "nonCirculating"
)

type GetLargestAccountsOpts struct {
	Commitment  CommitmentType
	Filter      LargestAccountsFilterType
	SortResults *bool
}

// GetLargestAccounts returns the 20 largest accounts,
// by lamport balance (results may be cached up to two hours).
func (cl *Client) GetLargestAccounts(
	ctx context.Context,
	commitment CommitmentType,
	filter LargestAccountsFilterType, // filter results by account type; currently supported: circulating|nonCirculating
) (out *GetLargestAccountsResult, err error) {
	return cl.GetLargestAccountsWithOpts(ctx, &GetLargestAccountsOpts{
		Commitment: commitment,
		Filter:     filter,
	})
}

// GetLargestAccountsWithOpts returns the 20 largest accounts,
// by lamport balance (results may be cached up to two hours).
func (cl *Client) GetLargestAccountsWithOpts(
	ctx context.Context,
	opts *GetLargestAccountsOpts,
) (out *GetLargestAccountsResult, err error) {
	params := []any{}
	obj := M{}
	if opts != nil {
		if opts.Commitment != "" {
			obj["commitment"] = opts.Commitment
		}
		if opts.Filter != "" {
			obj["filter"] = opts.Filter
		}
		if opts.SortResults != nil {
			obj["sortResults"] = *opts.SortResults
		}
	}
	if len(obj) > 0 {
		params = append(params, obj)
	}
	err = cl.rpcClient.CallForInto(ctx, &out, "getLargestAccounts", params)
	return
}

type GetLargestAccountsResult struct {
	RPCContext
	Value []LargestAccountsResult `json:"value"`
}

type LargestAccountsResult struct {
	// Address of the account.
	Address solana.PublicKey `json:"address"`

	// Number of lamports in the account.
	Lamports uint64 `json:"lamports"`
}
