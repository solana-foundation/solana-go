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

// Recv returns the next notification. After the subscription closes, it
// returns the buffered notifications first, and then the close error.
func (sw *RootSubscription) Recv(ctx context.Context) (*RootResult, error) {
	return recvAs[RootResult](ctx, sw.sub)
}

// WaitSubscribed blocks until the server confirms the subscription. See
// Subscription.WaitSubscribed.
func (sw *RootSubscription) WaitSubscribed(ctx context.Context) error {
	return sw.sub.WaitSubscribed(ctx)
}

func (sw *RootSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *RootSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
