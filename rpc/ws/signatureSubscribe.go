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
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type SignatureResult struct {
	Context struct {
		Slot uint64
	} `json:"context"`
	Value struct {
		Err interface{} `json:"err"`
	} `json:"value"`
}

// SignatureSubscribe subscribes to a transaction signature to receive
// notification when the transaction is confirmed On signatureNotification,
// the subscription is automatically cancelled
func (c *Client) SignatureSubscribe(
	signature solana.Signature, // Transaction Signature.
	commitment rpc.CommitmentType, // (optional)
) (*SignatureSubscription, error) {
	params := []interface{}{signature.String()}
	conf := map[string]interface{}{}
	if commitment != "" {
		conf["commitment"] = commitment
	}

	genSub, err := c.subscribe(
		params,
		conf,
		"signatureSubscribe",
		"signatureUnsubscribe",
		func(msg []byte) (interface{}, error) {
			var res SignatureResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &SignatureSubscription{
		sub: genSub,
	}, nil
}

type SignatureSubscription struct {
	sub *Subscription
}

func (sw *SignatureSubscription) Recv(ctx context.Context) (*SignatureResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-sw.sub.stream:
		if !ok {
			return nil, ErrSubscriptionClosed
		}
		return d.(*SignatureResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *SignatureSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *SignatureSubscription) Response() <-chan *SignatureResult {
	typedChan := make(chan *SignatureResult, 1)
	go func(ch chan *SignatureResult) {
		// TODO: will this subscription yield more than one result?
		d, ok := <-sw.sub.stream
		if !ok {
			return
		}
		ch <- d.(*SignatureResult)
	}(typedChan)
	return typedChan
}

var ErrTimeout = fmt.Errorf("timeout waiting for confirmation")

func (sw *SignatureSubscription) RecvWithTimeout(timeout time.Duration) (*SignatureResult, error) {
	select {
	case <-time.After(timeout):
		return nil, ErrTimeout
	case d := <-sw.sub.stream:
		return d.(*SignatureResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *SignatureSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
