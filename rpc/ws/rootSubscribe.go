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

import "context"

type RootResult uint64

// SignatureSubscribe subscribes to receive notification
// anytime a new root is set by the validator.
func (cl *Client) RootSubscribe() (*RootSubscription, error) {
	genSub, err := cl.subscribe(
		nil,
		nil,
		"rootSubscribe",
		"rootUnsubscribe",
		func(msg []byte) (any, error) {
			var res RootResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &RootSubscription{
		sub: genSub,
	}, nil
}

type RootSubscription struct {
	sub *Subscription
}

func (sw *RootSubscription) Recv(ctx context.Context) (*RootResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-sw.sub.stream:
		if !ok {
			return nil, ErrSubscriptionClosed
		}
		return d.(*RootResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *RootSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *RootSubscription) Response() <-chan *RootResult {
	typedChan := make(chan *RootResult, 1)
	go func(ch chan *RootResult) {
		// TODO: will this subscription yield more than one result?
		d, ok := <-sw.sub.stream
		if !ok {
			return
		}
		ch <- d.(*RootResult)
	}(typedChan)
	return typedChan
}

func (sw *RootSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
