package confirmation

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gagliardetto/solana-go/rpc"
)

// The cases follow kit's confirmation-strategy-timeout-test.ts. synctest runs
// them on a fake clock and fails if a goroutine is left behind.

func TestTimeout_ElapsesAfterCommitmentTimeout(t *testing.T) {
	for _, tt := range []struct {
		commitment rpc.CommitmentType
		timeout    time.Duration
	}{
		{rpc.CommitmentProcessed, 30 * time.Second},
		{rpc.CommitmentConfirmed, 60 * time.Second},
		{rpc.CommitmentFinalized, 60 * time.Second},
		{"", 60 * time.Second},
	} {
		t.Run(string(tt.commitment), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				done := make(chan error, 1)
				go func() { done <- Timeout(context.Background(), tt.commitment) }()

				time.Sleep(tt.timeout - time.Millisecond)
				synctest.Wait()
				select {
				case err := <-done:
					t.Fatalf("returned %v before the timeout", err)
				default:
				}

				time.Sleep(time.Millisecond)
				synctest.Wait()
				select {
				case err := <-done:
					require.ErrorIs(t, err, ErrTimeout)
					require.EqualError(t, err, "confirmation timed out after "+tt.timeout.String())
				default:
					t.Fatal("no timeout")
				}
			})
		})
	}
}

func TestTimeout_ReturnsWhenCanceled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() { done <- Timeout(ctx, rpc.CommitmentFinalized) }()

		time.Sleep(10 * time.Second)
		start := time.Now()
		cancel()

		require.ErrorIs(t, <-done, context.Canceled)
		require.Zero(t, time.Since(start), "it returns at once")
	})
}

// Kit waits out the timer when the signal is aborted already.
func TestTimeout_StartedCanceled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		start := time.Now()

		require.ErrorIs(t, Timeout(ctx, rpc.CommitmentFinalized), context.Canceled)
		require.Zero(t, time.Since(start))
	})
}

func TestTimeout_RejectsUnknownCommitment(t *testing.T) {
	require.ErrorContains(t, Timeout(context.Background(), "root"), `unsupported commitment "root"`)
}
