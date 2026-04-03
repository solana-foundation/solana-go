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

type SlotsUpdatesResult struct {
	// The parent slot.
	Parent uint64 `json:"parent"`
	// The newly updated slot.
	Slot uint64 `json:"slot"`
	// The Unix timestamp of the update.
	Timestamp *solana.UnixTimeMilliseconds `json:"timestamp"`
	// The update type.
	Type SlotsUpdatesType `json:"type"`
	// Extra stats provided when a bank is frozen.
	Stats *BankStats `json:"stats"`
}

type BankStats struct {
	NumTransactionEntries     uint64 `json:"numTransactionEntries"`
	NumSuccessfulTransactions uint64 `json:"numSuccessfulTransactions"`
	NumFailedTransactions     uint64 `json:"numFailedTransactions"`
	MaxTransactionsPerEntry   uint64 `json:"maxTransactionsPerEntry"`
}

type SlotsUpdatesType string

const (
	SlotsUpdatesFirstShredReceived     SlotsUpdatesType = "firstShredReceived"
	SlotsUpdatesCompleted              SlotsUpdatesType = "completed"
	SlotsUpdatesCreatedBank            SlotsUpdatesType = "createdBank"
	SlotsUpdatesFrozen                 SlotsUpdatesType = "frozen"
	SlotsUpdatesDead                   SlotsUpdatesType = "dead"
	SlotsUpdatesOptimisticConfirmation SlotsUpdatesType = "optimisticConfirmation"
	SlotsUpdatesRoot                   SlotsUpdatesType = "root"
)

// SlotsUpdatesSubscribe (UNSTABLE) subscribes to receive a notification
// from the validator on a variety of updates on every slot.
//
// This subscription is unstable; the format of this subscription
// may change in the future and it may not always be supported.
func (c *Client) SlotsUpdatesSubscribe() (*SlotsUpdatesSubscription, error) {
	genSub, err := c.subscribe(
		nil,
		nil,
		"slotsUpdatesSubscribe",
		"slotsUpdatesUnsubscribe",
		func(msg []byte) (interface{}, error) {
			var res SlotsUpdatesResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &SlotsUpdatesSubscription{
		sub: genSub,
	}, nil
}

type SlotsUpdatesSubscription struct {
	sub *Subscription
}

func (sw *SlotsUpdatesSubscription) Recv(ctx context.Context) (*SlotsUpdatesResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-sw.sub.stream:
		if !ok {
			return nil, ErrSubscriptionClosed
		}
		return d.(*SlotsUpdatesResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *SlotsUpdatesSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *SlotsUpdatesSubscription) Response() <-chan *SlotsUpdatesResult {
	typedChan := make(chan *SlotsUpdatesResult, 1)
	go func(ch chan *SlotsUpdatesResult) {
		// TODO: will this subscription yield more than one result?
		d, ok := <-sw.sub.stream
		if !ok {
			return
		}
		ch <- d.(*SlotsUpdatesResult)
	}(typedChan)
	return typedChan
}

func (sw *SlotsUpdatesSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
