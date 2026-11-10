// Copyright 2026 github.com/gagliardetto
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

package confirmation

import (
	"context"
	"math/bits"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// BlockHeightExceedanceFunc returns a *BlockHeightExceededError once the
// block height at commitment passes lastValidBlockHeight. Until then it
// blocks. An empty commitment means finalized.
type BlockHeightExceedanceFunc func(ctx context.Context, lastValidBlockHeight uint64, commitment rpc.CommitmentType) error

// NewBlockHeightExceedance returns a BlockHeightExceedanceFunc that follows
// slot notifications and estimates the block height from them, using the
// difference between slot and block height reported by getEpochInfo. When the
// estimate passes the limit, it checks the block height with getEpochInfo
// again before it reports the expiry. Kit:
// createBlockHeightExceedencePromiseFactory.
func NewBlockHeightExceedance(rpcClient *rpc.Client, wsClient *ws.Client) BlockHeightExceedanceFunc {
	return func(ctx context.Context, lastValidBlockHeight uint64, commitment rpc.CommitmentType) error {
		commitment, err := normalizeCommitment(commitment)
		if err != nil {
			return err
		}
		return contextError(ctx, blockHeightExceedance(ctx, rpcClient, wsClient, lastValidBlockHeight, commitment))
	}
}

// epochPosition is the slot and the block height of one getEpochInfo reply.
// Their difference is the number of skipped slots.
type epochPosition struct {
	absoluteSlot uint64
	blockHeight  uint64
}

// estimateExceeds reports whether the block height estimated for slot,
// slot - (absoluteSlot - blockHeight), is past limit. It compares
// slot + blockHeight with limit + absoluteSlot, so it cannot overflow.
func (p epochPosition) estimateExceeds(slot, limit uint64) bool {
	leftLo, leftHi := bits.Add64(slot, p.blockHeight, 0)
	rightLo, rightHi := bits.Add64(limit, p.absoluteSlot, 0)
	return leftHi > rightHi || (leftHi == rightHi && leftLo > rightLo)
}

func getEpochPosition(ctx context.Context, rpcClient *rpc.Client, commitment rpc.CommitmentType) (epochPosition, error) {
	info, err := rpcClient.GetEpochInfo(ctx, commitment)
	if err != nil {
		return epochPosition{}, err
	}
	return epochPosition{absoluteSlot: info.AbsoluteSlot, blockHeight: info.BlockHeight}, nil
}

func blockHeightExceedance(
	ctx context.Context,
	rpcClient *rpc.Client,
	wsClient *ws.Client,
	lastValidBlockHeight uint64,
	commitment rpc.CommitmentType,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	innerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Subscribe to slots and read the current position at the same time.
	type subscribed struct {
		sub *ws.SlotSubscription
		err error
	}
	subResult := make(chan subscribed, 1)
	go func() {
		sub, err := wsClient.SlotSubscribe()
		if err == nil {
			if err = sub.WaitSubscribed(innerCtx); err != nil {
				sub.Unsubscribe()
				sub = nil
			}
		}
		subResult <- subscribed{sub: sub, err: err}
	}()
	position, err := getEpochPosition(innerCtx, rpcClient, commitment)
	if err != nil {
		cancel()
	}
	slots := <-subResult
	if slots.sub != nil {
		defer slots.sub.Unsubscribe()
	}
	if err != nil {
		return err
	}
	if slots.err != nil {
		return slots.err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	currentBlockHeight := position.blockHeight
	if currentBlockHeight <= lastValidBlockHeight {
		for exceeded := false; !exceeded; {
			notification, err := slots.sub.Recv(innerCtx)
			if err != nil {
				return err
			}
			if notification == nil {
				return ws.ErrSubscriptionClosed
			}
			if !position.estimateExceeds(notification.Slot, lastValidBlockHeight) {
				continue
			}
			// Check the block height before deciding.
			rechecked, err := getEpochPosition(innerCtx, rpcClient, commitment)
			if err != nil {
				return err
			}
			currentBlockHeight = rechecked.blockHeight
			if currentBlockHeight > lastValidBlockHeight {
				exceeded = true
			} else {
				// Blocks were skipped since the last check, so the
				// difference between slot and block height grew.
				position = rechecked
			}
		}
	}
	// Report cancellation over the expiry (kit #3353).
	if err := ctx.Err(); err != nil {
		return err
	}
	return &BlockHeightExceededError{
		CurrentBlockHeight:   currentBlockHeight,
		LastValidBlockHeight: lastValidBlockHeight,
	}
}
