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

package ws

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type AccountResult struct {
	Context struct {
		Slot uint64 `json:"slot"`
	} `json:"context"`
	Value *rpc.Account `json:"value"`
}

// AccountSubscribeOpts mirrors the optional configuration object the
// accountSubscribe RPC method accepts. See
// https://solana.com/docs/rpc/websocket/accountsubscribe.
type AccountSubscribeOpts struct {
	// Commitment selects the bank state the validator should query.
	Commitment rpc.CommitmentType

	// Encoding controls how the account data field is encoded. When
	// empty the request defaults to "base64".
	Encoding solana.EncodingType

	// DataSlice asks the validator to return only the requested slice
	// of the account's data field. Only valid for binary encodings
	// (base58, base64, base64+zstd) per the RPC spec.
	DataSlice *rpc.DataSlice
}

// AccountSubscribe subscribes to an account to receive notifications
// when the lamports or data for a given account public key changes.
func (cl *Client) AccountSubscribe(
	account solana.PublicKey,
	commitment rpc.CommitmentType,
) (*AccountSubscription, error) {
	return cl.AccountSubscribeWithOpts(
		account,
		commitment,
		"",
	)
}

// AccountSubscribeWithOpts subscribes to an account with explicit
// commitment and encoding overrides. Kept for backward compatibility;
// new callers should prefer AccountSubscribeWithConfig which exposes
// the full optional configuration object (including DataSlice).
func (cl *Client) AccountSubscribeWithOpts(
	account solana.PublicKey,
	commitment rpc.CommitmentType,
	encoding solana.EncodingType,
) (*AccountSubscription, error) {
	return cl.AccountSubscribeWithConfig(account, &AccountSubscribeOpts{
		Commitment: commitment,
		Encoding:   encoding,
	})
}

// AccountSubscribeWithConfig subscribes to an account and forwards the
// full AccountSubscribeOpts configuration object to the validator,
// including DataSlice.
func (cl *Client) AccountSubscribeWithConfig(
	account solana.PublicKey,
	opts *AccountSubscribeOpts,
) (*AccountSubscription, error) {
	params := []any{account.String()}
	conf := map[string]any{
		"encoding": "base64",
	}
	if opts != nil {
		if opts.Commitment != "" {
			conf["commitment"] = opts.Commitment
		}
		if opts.Encoding != "" {
			conf["encoding"] = opts.Encoding
		}
		if opts.DataSlice != nil {
			conf["dataSlice"] = opts.DataSlice
		}
	}

	genSub, err := cl.subscribe(
		params,
		conf,
		"accountSubscribe",
		"accountUnsubscribe",
		func(msg []byte) (any, error) {
			var res AccountResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &AccountSubscription{
		sub: genSub,
	}, nil
}

type AccountSubscription struct {
	sub *Subscription
}

func (sw *AccountSubscription) Recv(ctx context.Context) (*AccountResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-sw.sub.stream:
		if !ok {
			return nil, ErrSubscriptionClosed
		}
		return d.(*AccountResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *AccountSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *AccountSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
