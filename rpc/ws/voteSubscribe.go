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
)

type VoteResult struct {
	// The vote hash.
	Hash solana.Hash `json:"hash"`
	// The slots covered by the vote.
	Slots []uint64 `json:"slots"`
	// The timestamp of the vote.
	Timestamp *solana.UnixTimeSeconds `json:"timestamp,omitempty"`
}

// VoteSubscribe (UNSTABLE, disabled by default) subscribes
// to receive notification anytime a new vote is observed in gossip.
// These votes are pre-consensus therefore there is
// no guarantee these votes will enter the ledger.
//
// This subscription is unstable and only available if the validator
// was started with the --rpc-pubsub-enable-vote-subscription flag.
// The format of this subscription may change in the future.
func (c *Client) VoteSubscribe() (*VoteSubscription, error) {
	genSub, err := c.subscribe(
		nil,
		nil,
		"voteSubscribe",
		"voteUnsubscribe",
		func(msg []byte) (interface{}, error) {
			var res VoteResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &VoteSubscription{
		sub: genSub,
	}, nil
}

type VoteSubscription struct {
	sub *Subscription
}

func (sw *VoteSubscription) Recv(ctx context.Context) (*VoteResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-sw.sub.stream:
		if !ok {
			return nil, ErrSubscriptionClosed
		}
		return d.(*VoteResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *VoteSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *VoteSubscription) Response() <-chan *VoteResult {
	typedChan := make(chan *VoteResult, 1)
	go func(ch chan *VoteResult) {
		// TODO: will this subscription yield more than one result?
		d, ok := <-sw.sub.stream
		if !ok {
			return
		}
		ch <- d.(*VoteResult)
	}(typedChan)
	return typedChan
}

func (sw *VoteSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
