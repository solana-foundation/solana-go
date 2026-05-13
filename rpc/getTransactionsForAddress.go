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
	"fmt"

	"github.com/gagliardetto/solana-go"
)

type TransactionsForAddressSortOrder string

const (
	TransactionsForAddressSortOrderAscending  TransactionsForAddressSortOrder = "asc"
	TransactionsForAddressSortOrderDescending TransactionsForAddressSortOrder = "desc"
)

type TransactionsForAddressStatus string

const (
	TransactionsForAddressStatusSucceeded TransactionsForAddressStatus = "succeeded"
	TransactionsForAddressStatusFailed    TransactionsForAddressStatus = "failed"
	TransactionsForAddressStatusAny       TransactionsForAddressStatus = "any"
)

type TransactionsForAddressTokenAccounts string

const (
	TransactionsForAddressTokenAccountsNone           TransactionsForAddressTokenAccounts = "none"
	TransactionsForAddressTokenAccountsBalanceChanged TransactionsForAddressTokenAccounts = "balanceChanged"
	TransactionsForAddressTokenAccountsAll            TransactionsForAddressTokenAccounts = "all"
)

type GetTransactionsForAddressUint64Filter struct {
	Gte *uint64 `json:"gte,omitempty"`
	Gt  *uint64 `json:"gt,omitempty"`
	Lte *uint64 `json:"lte,omitempty"`
	Lt  *uint64 `json:"lt,omitempty"`
}

type GetTransactionsForAddressInt64Filter struct {
	Gte *int64 `json:"gte,omitempty"`
	Gt  *int64 `json:"gt,omitempty"`
	Lte *int64 `json:"lte,omitempty"`
	Lt  *int64 `json:"lt,omitempty"`
	Eq  *int64 `json:"eq,omitempty"`
}

type GetTransactionsForAddressSignatureFilter struct {
	Gte *solana.Signature `json:"gte,omitempty"`
	Gt  *solana.Signature `json:"gt,omitempty"`
	Lte *solana.Signature `json:"lte,omitempty"`
	Lt  *solana.Signature `json:"lt,omitempty"`
}

type GetTransactionsForAddressFilters struct {
	Slot          *GetTransactionsForAddressUint64Filter    `json:"slot,omitempty"`
	BlockTime     *GetTransactionsForAddressInt64Filter     `json:"blockTime,omitempty"`
	Signature     *GetTransactionsForAddressSignatureFilter `json:"signature,omitempty"`
	Status        TransactionsForAddressStatus              `json:"status,omitempty"`
	TokenAccounts TransactionsForAddressTokenAccounts       `json:"tokenAccounts,omitempty"`
}

type GetTransactionsForAddressOpts struct {
	// Level of transaction detail to return:
	// - "signatures": basic signature info
	// - "full": complete transaction data
	TransactionDetails TransactionDetailsType `json:"transactionDetails,omitempty"`

	// Sort order for results:
	// - "desc": newest first
	// - "asc": oldest first
	SortOrder TransactionsForAddressSortOrder `json:"sortOrder,omitempty"`

	// Maximum transactions to return.
	Limit *int `json:"limit,omitempty"`

	// Pagination token from a previous response.
	PaginationToken string `json:"paginationToken,omitempty"`

	// Desired commitment. "processed" is not supported. If omitted, the default is "finalized".
	Commitment CommitmentType `json:"commitment,omitempty"`

	// Advanced filtering options.
	Filters *GetTransactionsForAddressFilters `json:"filters,omitempty"`

	// Encoding format for transaction data when TransactionDetails is "full".
	Encoding solana.EncodingType `json:"encoding,omitempty"`

	// Max transaction version to return in responses.
	MaxSupportedTransactionVersion *uint64

	// The minimum slot that the request can be evaluated at.
	MinContextSlot *uint64
}

type GetTransactionsForAddressResult struct {
	Data            []TransactionsForAddressSignature `json:"data"`
	PaginationToken *string                           `json:"paginationToken"`
}

type TransactionsForAddressSignature struct {
	TransactionSignature
	TransactionIndex uint64 `json:"transactionIndex"`
}

type GetTransactionsForAddressFullResult struct {
	Data            []TransactionsForAddressTransaction `json:"data"`
	PaginationToken *string                             `json:"paginationToken"`
}

type TransactionsForAddressTransaction struct {
	Slot             uint64                  `json:"slot"`
	TransactionIndex uint64                  `json:"transactionIndex"`
	BlockTime        *solana.UnixTimeSeconds `json:"blockTime,omitempty"`
	Transaction      *DataBytesOrJSON        `json:"transaction,omitempty"`
	Meta             *TransactionMeta        `json:"meta,omitempty"`
	Version          *TransactionVersion     `json:"version,omitempty"`
}

// GetTransactionsForAddress returns signature-level transaction history for an address.
func (cl *Client) GetTransactionsForAddress(
	ctx context.Context,
	account solana.PublicKey,
) (out *GetTransactionsForAddressResult, err error) {
	return cl.GetTransactionsForAddressWithOpts(ctx, account, nil)
}

// GetTransactionsForAddressWithOpts returns signature-level transaction history for an address.
func (cl *Client) GetTransactionsForAddressWithOpts(
	ctx context.Context,
	account solana.PublicKey,
	opts *GetTransactionsForAddressOpts,
) (out *GetTransactionsForAddressResult, err error) {
	params, err := getTransactionsForAddressParams(account, opts, "")
	if err != nil {
		return nil, err
	}

	err = cl.rpcClient.CallForInto(ctx, &out, "getTransactionsForAddress", params)
	return
}

// GetTransactionsForAddressFull returns full transaction history for an address.
func (cl *Client) GetTransactionsForAddressFull(
	ctx context.Context,
	account solana.PublicKey,
) (out *GetTransactionsForAddressFullResult, err error) {
	return cl.GetTransactionsForAddressFullWithOpts(ctx, account, nil)
}

// GetTransactionsForAddressFullWithOpts returns full transaction history for an address.
func (cl *Client) GetTransactionsForAddressFullWithOpts(
	ctx context.Context,
	account solana.PublicKey,
	opts *GetTransactionsForAddressOpts,
) (out *GetTransactionsForAddressFullResult, err error) {
	params, err := getTransactionsForAddressParams(account, opts, TransactionDetailsFull)
	if err != nil {
		return nil, err
	}

	err = cl.rpcClient.CallForInto(ctx, &out, "getTransactionsForAddress", params)
	return
}

func getTransactionsForAddressParams(
	account solana.PublicKey,
	opts *GetTransactionsForAddressOpts,
	forcedTransactionDetails TransactionDetailsType,
) ([]any, error) {
	params := []any{account}
	if opts == nil && forcedTransactionDetails == "" {
		return params, nil
	}

	obj := M{}
	if opts != nil {
		if opts.TransactionDetails != "" {
			obj["transactionDetails"] = opts.TransactionDetails
		}
		if opts.SortOrder != "" {
			obj["sortOrder"] = opts.SortOrder
		}
		if opts.Limit != nil {
			obj["limit"] = *opts.Limit
		}
		if opts.PaginationToken != "" {
			obj["paginationToken"] = opts.PaginationToken
		}
		if opts.Commitment != "" {
			obj["commitment"] = opts.Commitment
		}
		if opts.Filters != nil {
			obj["filters"] = opts.Filters
		}
		if opts.Encoding != "" {
			if !solana.IsAnyOfEncodingType(
				opts.Encoding,
				solana.EncodingJSON,
				solana.EncodingJSONParsed,
				solana.EncodingBase58,
				solana.EncodingBase64,
			) {
				return nil, fmt.Errorf("provided encoding is not supported: %s", opts.Encoding)
			}
			obj["encoding"] = opts.Encoding
		}
		if opts.MaxSupportedTransactionVersion != nil {
			obj["maxSupportedTransactionVersion"] = *opts.MaxSupportedTransactionVersion
		}
		if opts.MinContextSlot != nil {
			obj["minContextSlot"] = *opts.MinContextSlot
		}
	}
	if forcedTransactionDetails != "" {
		obj["transactionDetails"] = forcedTransactionDetails
	}
	if len(obj) > 0 {
		params = append(params, obj)
	}
	return params, nil
}
