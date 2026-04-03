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

// AccountSubscribe subscribes to an account to receive notifications
// when the lamports or data for a given account public key changes.
func (c *Client) AccountSubscribe(
	account solana.PublicKey,
	commitment rpc.CommitmentType,
) (*AccountSubscription, error) {
	return c.AccountSubscribeWithOpts(
		account,
		commitment,
		"",
	)
}

// AccountSubscribe subscribes to an account to receive notifications
// when the lamports or data for a given account public key changes.
func (c *Client) AccountSubscribeWithOpts(
	account solana.PublicKey,
	commitment rpc.CommitmentType,
	encoding solana.EncodingType,
) (*AccountSubscription, error) {

	params := []interface{}{account.String()}
	conf := map[string]interface{}{
		"encoding": "base64",
	}
	if commitment != "" {
		conf["commitment"] = commitment
	}
	if encoding != "" {
		conf["encoding"] = encoding
	}

	genSub, err := c.subscribe(
		params,
		conf,
		"accountSubscribe",
		"accountUnsubscribe",
		func(msg []byte) (interface{}, error) {
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

func (sw *AccountSubscription) Response() <-chan *AccountResult {
	typedChan := make(chan *AccountResult, 1)
	go func(ch chan *AccountResult) {
		// TODO: will this subscription yield more than one result?
		d, ok := <-sw.sub.stream
		if !ok {
			return
		}
		ch <- d.(*AccountResult)
	}(typedChan)
	return typedChan
}

func (sw *AccountSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
