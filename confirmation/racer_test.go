package confirmation

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gagliardetto/solana-go/rpc/ws"
)

// The cases follow kit's confirmation-strategy-racer-test.ts. synctest fails
// a test that leaves a goroutine behind.

var errStrategy = errors.New("o no")

// capturingStrategy records the context it was called with.
type capturingStrategy struct {
	ctx context.Context
	run func(ctx context.Context) error
}

func (s *capturingStrategy) call(ctx context.Context) error {
	s.ctx = ctx
	return s.run(ctx)
}

func blockOnContext(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func TestPick_SignatureWinsTie(t *testing.T) {
	// select picks at random among ready cases, so try many times.
	for range 1000 {
		signature, lifetime := make(chan error, 1), make(chan error, 1)
		signature <- nil
		lifetime <- errStrategy
		require.NoError(t, pick(signature, lifetime))
	}
}

func TestPick_LifetimeWhenSignatureNotReady(t *testing.T) {
	signature, lifetime := make(chan error, 1), make(chan error, 1)
	lifetime <- errStrategy
	require.ErrorIs(t, pick(signature, lifetime), errStrategy)
}

func TestRaceStrategies_CancelsSignatureWhenFinished(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		signature := &capturingStrategy{run: func(context.Context) error { return nil }}
		lifetime := &capturingStrategy{run: blockOnContext}

		require.NoError(t, raceStrategies(context.Background(), signature.call, lifetime.call))
		require.ErrorIs(t, signature.ctx.Err(), context.Canceled)
		require.ErrorIs(t, lifetime.ctx.Err(), context.Canceled)
	})
}

func TestRaceStrategies_CancelsLifetimeWhenFinished(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		signature := &capturingStrategy{run: blockOnContext}
		lifetime := &capturingStrategy{run: func(context.Context) error { return errStrategy }}

		require.ErrorIs(t, raceStrategies(context.Background(), signature.call, lifetime.call), errStrategy)
		require.ErrorIs(t, signature.ctx.Err(), context.Canceled)
		require.ErrorIs(t, lifetime.ctx.Err(), context.Canceled)
	})
}

func TestRaceStrategies_CancelsBothWhenCallerCancels(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		signature := &capturingStrategy{run: blockOnContext}
		lifetime := &capturingStrategy{run: blockOnContext}
		done := make(chan error, 1)
		go func() { done <- raceStrategies(ctx, signature.call, lifetime.call) }()

		synctest.Wait()
		require.NoError(t, signature.ctx.Err())
		require.NoError(t, lifetime.ctx.Err())
		cancel()

		require.ErrorIs(t, <-done, context.Canceled)
		require.ErrorIs(t, signature.ctx.Err(), context.Canceled)
		require.ErrorIs(t, lifetime.ctx.Err(), context.Canceled)
	})
}

func TestRaceStrategies_ReportsCancellationOverStrategyErrors(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		// A strategy whose subscription closes on cancellation reports
		// ws.ErrSubscriptionClosed.
		closed := func(ctx context.Context) error {
			<-ctx.Done()
			return ws.ErrSubscriptionClosed
		}
		done := make(chan error, 1)
		go func() { done <- raceStrategies(ctx, closed, closed) }()

		time.Sleep(time.Second)
		cancel()

		require.ErrorIs(t, <-done, context.Canceled)
	})
}

func TestRaceStrategies_StartedCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	strategy := func(context.Context) error {
		called = true
		return nil
	}

	require.ErrorIs(t, raceStrategies(ctx, strategy, strategy), context.Canceled)
	require.False(t, called)
}
