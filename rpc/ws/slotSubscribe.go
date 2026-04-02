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

type SlotResult struct {
	Parent uint64 `json:"parent"`
	Root   uint64 `json:"root"`
	Slot   uint64 `json:"slot"`
}

// SlotSubscribe subscribes to receive notification anytime a slot is processed by the validator.
func (cl *Client) SlotSubscribe() (*SlotSubscription, error) {
	genSub, err := cl.subscribe(
		nil,
		nil,
		"slotSubscribe",
		"slotUnsubscribe",
		func(msg []byte) (any, error) {
			var res SlotResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &SlotSubscription{
		sub: genSub,
	}, nil
}

type SlotSubscription struct {
	sub *Subscription
}

func (sw *SlotSubscription) Recv(ctx context.Context) (*SlotResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-sw.sub.stream:
		if !ok {
			return nil, ErrSubscriptionClosed
		}
		return d.(*SlotResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *SlotSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *SlotSubscription) Response() <-chan *SlotResult {
	typedChan := make(chan *SlotResult, 1)
	go func(ch chan *SlotResult) {
		// TODO: will this subscription yield more than one result?
		d, ok := <-sw.sub.stream
		if !ok {
			return
		}
		ch <- d.(*SlotResult)
	}(typedChan)
	return typedChan
}

func (sw *SlotSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
