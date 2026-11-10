package confirmation

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gagliardetto/solana-go/rpc"
)

// The cases follow kit's confirmation-strategy-blockheight-test.ts.

func epochInfo(absoluteSlot, blockHeight uint64) rpcHandler {
	return result(fmt.Sprintf(`{"absoluteSlot":%d,"blockHeight":%d,"epoch":0,"slotIndex":0,"slotsInEpoch":432000}`,
		absoluteSlot, blockHeight))
}

// startBlockHeight runs the block height strategy and returns its slot
// subscription.
func (h *harness) startBlockHeight(t *testing.T, lastValidBlockHeight uint64, commitment rpc.CommitmentType) (context.CancelFunc, <-chan error, subscribeRequest) {
	t.Helper()
	strategy := NewBlockHeightExceedance(h.rpcClient, h.wsClient)
	cancel, done := run(t, func(ctx context.Context) error {
		return strategy(ctx, lastValidBlockHeight, commitment)
	})
	sub := h.validator.nextSubscribe(t)
	require.Equal(t, "slotSubscribe", sub.Method)
	return cancel, done, sub
}

func (h *harness) slot(t *testing.T, sub subscribeRequest, slot uint64) {
	t.Helper()
	h.validator.notify(t, "slotNotification", sub.SubID, map[string]uint64{"parent": slot - 1, "root": 0, "slot": slot})
}

func requireExceeded(t *testing.T, err error, current, lastValid uint64) {
	t.Helper()
	require.ErrorIs(t, err, ErrBlockHeightExceeded)
	var exceeded *BlockHeightExceededError
	require.ErrorAs(t, err, &exceeded)
	require.Equal(t, BlockHeightExceededError{CurrentBlockHeight: current, LastValidBlockHeight: lastValid}, *exceeded)
}

func TestBlockHeight_ExceededWhenCalled(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getEpochInfo", epochInfo(101, 101))

	_, done, sub := h.startBlockHeight(t, 100, "")

	requireExceeded(t, requireResult(t, done), 101, 100)
	h.events.waitFor(t, fmt.Sprintf("unsubscribe slotUnsubscribe [%d]", sub.SubID))
}

func TestBlockHeight_PendsAtLastValidBlockHeight(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getEpochInfo", epochInfo(100, 100))

	cancel, done, _ := h.startBlockHeight(t, 100, "")

	requireCanceled(t, cancel, done)
}

func TestBlockHeight_ExceededAtEstimatedSlot(t *testing.T) {
	h := newHarness(t)
	// 100 slots were skipped, so slot 201 is expected to be block 101.
	h.rpc.queue("getEpochInfo", epochInfo(198, 98), epochInfo(201, 101))

	_, done, sub := h.startBlockHeight(t, 100, "")
	for slot := uint64(199); slot <= 201; slot++ {
		h.slot(t, sub, slot)
	}

	requireExceeded(t, requireResult(t, done), 101, 100)
	require.Len(t, h.rpc.callParams("getEpochInfo"), 2, "only slot 201 needs a recheck")
}

func TestBlockHeight_PendsBeforeEstimatedSlot(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getEpochInfo", epochInfo(198, 98))

	cancel, done, sub := h.startBlockHeight(t, 100, "")
	h.slot(t, sub, 199)
	h.slot(t, sub, 200)

	requireCanceled(t, cancel, done)
	require.Len(t, h.rpc.callParams("getEpochInfo"), 1, "no slot reached the estimated height")
}

func TestBlockHeight_ExceededAfterRecalibration(t *testing.T) {
	h := newHarness(t)
	// At slot 201 the recheck shows that one more slot was skipped, so the
	// height is only exceeded at slot 202.
	h.rpc.queue("getEpochInfo", epochInfo(198, 98), epochInfo(201, 100), epochInfo(202, 101))

	_, done, sub := h.startBlockHeight(t, 100, "")
	for slot := uint64(199); slot <= 202; slot++ {
		h.slot(t, sub, slot)
	}

	requireExceeded(t, requireResult(t, done), 101, 100)
	require.Len(t, h.rpc.callParams("getEpochInfo"), 3)
}

func TestBlockHeight_PendsWhenSkippedSlotsGrew(t *testing.T) {
	h := newHarness(t)
	h.rpc.queue("getEpochInfo", epochInfo(198, 98), epochInfo(201, 100))

	cancel, done, sub := h.startBlockHeight(t, 100, "")
	for slot := uint64(199); slot <= 201; slot++ {
		h.slot(t, sub, slot)
	}
	require.Eventually(t, func() bool { return h.events.count("rpc getEpochInfo") == 2 }, testWait, testTick)
	// With 101 skipped slots, slot 201 is block 100. Without the
	// recalibration it would be checked again, and the fake has no answer
	// left for a third call.
	h.slot(t, sub, 201)

	requireCanceled(t, cancel, done)
	require.Len(t, h.rpc.callParams("getEpochInfo"), 2)
}

func TestBlockHeight_PassesCommitment(t *testing.T) {
	for _, tt := range []struct {
		commitment, want rpc.CommitmentType
	}{
		{rpc.CommitmentProcessed, rpc.CommitmentProcessed},
		{rpc.CommitmentConfirmed, rpc.CommitmentConfirmed},
		{rpc.CommitmentFinalized, rpc.CommitmentFinalized},
		{"", rpc.CommitmentFinalized},
	} {
		t.Run(string(tt.want)+"/"+string(tt.commitment), func(t *testing.T) {
			h := newHarness(t)
			h.rpc.on("getEpochInfo", epochInfo(100, 100))

			cancel, done, _ := h.startBlockHeight(t, 100, tt.commitment)
			h.events.waitFor(t, "rpc getEpochInfo")

			require.Equal(t, []any{rpc.M{"commitment": tt.want}}, h.rpc.callParams("getEpochInfo")[0])
			requireCanceled(t, cancel, done)
		})
	}
}

func TestBlockHeight_RejectsUnknownCommitment(t *testing.T) {
	h := newHarness(t)

	err := NewBlockHeightExceedance(h.rpcClient, h.wsClient)(context.Background(), 100, "recent")

	require.ErrorContains(t, err, `unsupported commitment "recent"`)
	require.Empty(t, h.events.list())
}

func TestBlockHeight_CancelStopsEpochInfoRequest(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getEpochInfo", blockUntilCanceled(h.events, "getEpochInfo"))

	cancel, done, _ := h.startBlockHeight(t, 100, "")
	h.events.waitFor(t, "rpc getEpochInfo")

	requireCanceled(t, cancel, done)
	h.events.waitFor(t, "rpc getEpochInfo canceled")
}

func TestBlockHeight_CancelUnsubscribes(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getEpochInfo", epochInfo(100, 100))

	cancel, done, sub := h.startBlockHeight(t, 100, "")

	requireCanceled(t, cancel, done)
	h.events.waitFor(t, fmt.Sprintf("unsubscribe slotUnsubscribe [%d]", sub.SubID))
}

func TestBlockHeight_EpochInfoError(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getEpochInfo", failWith(errors.New("o no")))

	_, done, sub := h.startBlockHeight(t, 100, "")

	require.ErrorContains(t, requireResult(t, done), "o no")
	h.events.waitFor(t, fmt.Sprintf("unsubscribe slotUnsubscribe [%d]", sub.SubID))
}

func TestBlockHeight_SubscriptionError(t *testing.T) {
	h := newHarness(t)
	h.validator.reject("slotSubscribe", "o no")
	h.rpc.on("getEpochInfo", epochInfo(100, 100))

	_, done := run(t, func(ctx context.Context) error {
		return NewBlockHeightExceedance(h.rpcClient, h.wsClient)(ctx, 100, "")
	})

	require.ErrorContains(t, requireResult(t, done), "o no")
}

func TestBlockHeight_StartedCanceled(t *testing.T) {
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewBlockHeightExceedance(h.rpcClient, h.wsClient)(ctx, 100, "")

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, h.events.list(), "nothing is requested")
}

// Kit #3353: a result that arrives after cancellation must not be reported.
func TestBlockHeight_CanceledWhileReadingEpochInfo(t *testing.T) {
	h := newHarness(t)
	release := make(chan struct{})
	h.rpc.on("getEpochInfo", func(context.Context, []any) (any, error) {
		<-release
		return `{"absoluteSlot":101,"blockHeight":101}`, nil
	})

	cancel, done, _ := h.startBlockHeight(t, 100, "")
	h.events.waitFor(t, "rpc getEpochInfo")
	cancel()
	close(release)

	require.ErrorIs(t, requireResult(t, done), context.Canceled)
}

func TestBlockHeight_CanceledWhileFollowingSlots(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getEpochInfo", epochInfo(100, 100))

	cancel, done, sub := h.startBlockHeight(t, 100, "")
	requirePending(t, done)
	cancel()
	h.slot(t, sub, 101)

	require.ErrorIs(t, requireResult(t, done), context.Canceled)
}

// Every block height strategy subscribes to slots, so concurrent strategies
// on one ws.Client share one server subscription id.
func TestBlockHeight_ConcurrentStrategiesShareSlotSubscription(t *testing.T) {
	h := newHarness(t)
	var height atomic.Uint64
	height.Store(98)
	h.rpc.on("getEpochInfo", func(context.Context, []any) (any, error) {
		bh := height.Load()
		return fmt.Sprintf(`{"absoluteSlot":%d,"blockHeight":%d}`, bh+100, bh), nil
	})
	strategy := NewBlockHeightExceedance(h.rpcClient, h.wsClient)
	_, first := run(t, func(ctx context.Context) error { return strategy(ctx, 100, "") })
	_, second := run(t, func(ctx context.Context) error { return strategy(ctx, 105, "") })
	a, b := h.validator.nextSubscribe(t), h.validator.nextSubscribe(t)
	require.Equal(t, a.SubID, b.SubID, "identical subscriptions share an id")
	require.Eventually(t, func() bool { return h.events.count("rpc getEpochInfo") == 2 }, testWait, testTick)

	for slot := uint64(199); slot <= 201; slot++ {
		height.Store(slot - 100)
		h.slot(t, a, slot)
	}
	requireExceeded(t, requireResult(t, first), 101, 100)

	// The second strategy still receives slots after the first one closed
	// its subscription.
	requirePending(t, second)
	for slot := uint64(202); slot <= 206; slot++ {
		height.Store(slot - 100)
		h.slot(t, a, slot)
	}
	requireExceeded(t, requireResult(t, second), 106, 105)
	h.events.waitFor(t, fmt.Sprintf("unsubscribe slotUnsubscribe [%d]", a.SubID))
	require.Equal(t, 1, h.events.count(fmt.Sprintf("unsubscribe slotUnsubscribe [%d]", a.SubID)))
}

func TestEpochPosition_EstimateExceeds(t *testing.T) {
	for _, tt := range []struct {
		name                      string
		absoluteSlot, blockHeight uint64
		slot, limit               uint64
		want                      bool
	}{
		{"estimate below limit", 198, 98, 199, 100, false},
		{"estimate at limit", 198, 98, 200, 100, false},
		{"estimate past limit", 198, 98, 201, 100, true},
		{"slot below skipped slots", 198, 98, 50, 0, false},
		{"bogus reply with more blocks than slots", 5, 10, 0, 4, true},
		{"no overflow at the top", math.MaxUint64, math.MaxUint64 - 1, math.MaxUint64, math.MaxUint64, false},
		{"no overflow past the top", 1, math.MaxUint64, math.MaxUint64, math.MaxUint64 - 1, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := epochPosition{absoluteSlot: tt.absoluteSlot, blockHeight: tt.blockHeight}
			require.Equal(t, tt.want, p.estimateExceeds(tt.slot, tt.limit))
		})
	}
}
