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

type RequestAirdropOpts struct {
	Commitment CommitmentType

	// Must be a recent blockhash as a base-58 encoded string.
	// If not provided, a recent blockhash is used.
	RecentBlockhash *solana.Hash
}

// RequestAirdrop requests an airdrop of lamports to a publicKey.
// Returns transaction signature of airdrop.
func (cl *Client) RequestAirdrop(
	ctx context.Context,
	account solana.PublicKey,
	lamports uint64,
	commitment CommitmentType, // optional; used for retrieving blockhash and verifying airdrop success.
) (signature solana.Signature, err error) {
	return cl.RequestAirdropWithOpts(ctx, account, lamports, &RequestAirdropOpts{
		Commitment: commitment,
	})
}

// RequestAirdropWithOpts requests an airdrop of lamports to a publicKey with additional options.
// Returns transaction signature of airdrop.
func (cl *Client) RequestAirdropWithOpts(
	ctx context.Context,
	account solana.PublicKey,
	lamports uint64,
	opts *RequestAirdropOpts,
) (signature solana.Signature, err error) {
	params := []any{
		account,
		lamports,
	}
	if opts != nil {
		obj := M{}
		if opts.Commitment != "" {
			obj["commitment"] = opts.Commitment
		}
		if opts.RecentBlockhash != nil {
			obj["recentBlockhash"] = opts.RecentBlockhash.String()
		}
		if len(obj) > 0 {
			params = append(params, obj)
		}
	}
	err = cl.rpcClient.CallForInto(ctx, &signature, "requestAirdrop", params)
	return
}
