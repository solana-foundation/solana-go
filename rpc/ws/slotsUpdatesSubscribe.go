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

// SlotsUpdatesResult is a Go projection of Agave's SlotUpdate tagged
// union (rename_all="camelCase", tag="type"). All variants carry
// Slot and Timestamp; the rest of the fields are populated only for
// the variants where they apply:
//
//	Type                        Extra fields
//	-----------------------------------------
//	firstShredReceived          (none)
//	completed                   (none)
//	createdBank                 Parent
//	frozen                      Stats
//	dead                        Err
//	optimisticConfirmation      (none)
//	root                        (none)
type SlotsUpdatesResult struct {
	// The update type (discriminator).
	Type SlotsUpdatesType `json:"type"`
	// The newly updated slot.
	Slot uint64 `json:"slot"`
	// The Unix timestamp of the update (milliseconds).
	Timestamp *solana.UnixTimeMilliseconds `json:"timestamp"`
	// The parent slot. Populated only for type=createdBank.
	Parent uint64 `json:"parent,omitempty"`
	// Extra stats. Populated only for type=frozen.
	Stats *BankStats `json:"stats,omitempty"`
	// Error message. Populated only for type=dead.
	Err string `json:"err,omitempty"`
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
func (cl *Client) SlotsUpdatesSubscribe() (*SlotsUpdatesSubscription, error) {
	genSub, err := cl.subscribe(
		nil,
		nil,
		"slotsUpdatesSubscribe",
		"slotsUpdatesUnsubscribe",
		func(msg []byte) (any, error) {
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

func (sw *SlotsUpdatesSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
