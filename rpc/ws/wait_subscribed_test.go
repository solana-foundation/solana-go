package ws

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// waitSubscribedAsync runs WaitSubscribed in a goroutine and returns its
// result channel.
func waitSubscribedAsync(sub interface {
	WaitSubscribed(context.Context) error
}) <-chan error {
	done := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		done <- sub.WaitSubscribed(ctx)
	}()
	return done
}

func requireResult(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("WaitSubscribed did not return")
		return nil
	}
}

func requireNotSubscribed(t *testing.T, sub *Subscription) {
	t.Helper()
	select {
	case <-sub.subscribed:
		t.Fatal("subscription reported as confirmed before the server confirmed it")
	default:
	}
}

func TestWaitSubscribed_ReturnsAfterConfirmation(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	sub := newTestSubscription(t, c)
	req := m.nextRequest(t)
	done := waitSubscribedAsync(sub)
	requireNotSubscribed(t, sub)

	m.ack(t, req.ID, 3)
	require.NoError(t, requireResult(t, done))
}

func TestWaitSubscribed_ReturnsServerError(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	sub := newTestSubscription(t, c)
	req := m.nextRequest(t)
	done := waitSubscribedAsync(sub)

	m.send(t, fmt.Sprintf(`{"jsonrpc":"2.0","error":{"code":-32602,"message":"Invalid params"},"id":%d}`, req.ID))

	err := requireResult(t, done)
	require.ErrorContains(t, err, "rpc error (code -32602): Invalid params")
}

func TestWaitSubscribed_ReturnsClosedBeforeConfirmation(t *testing.T) {
	tests := []struct {
		name  string
		close func(c *Client, m *mockWSServer, sub *Subscription)
	}{
		{"unsubscribe", func(_ *Client, _ *mockWSServer, sub *Subscription) { sub.Unsubscribe() }},
		{"connection drop", func(_ *Client, m *mockWSServer, _ *Subscription) { m.closeConn() }},
		{"client close", func(c *Client, _ *mockWSServer, _ *Subscription) { c.Close() }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockWSServer(t)
			defer m.stop()
			c := connectClient(t, m)
			defer c.Close()

			sub := newTestSubscription(t, c)
			m.nextRequest(t)
			done := waitSubscribedAsync(sub)

			tt.close(c, m, sub)
			require.ErrorIs(t, requireResult(t, done), ErrSubscriptionClosed)
		})
	}
}

func TestWaitSubscribed_ReturnsContextError(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	sub := newTestSubscription(t, c)
	m.nextRequest(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, sub.WaitSubscribed(ctx), context.Canceled)

	ctx, cancel = context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	require.ErrorIs(t, sub.WaitSubscribed(ctx), context.DeadlineExceeded)
}

func TestWaitSubscribed_StaysConfirmedAfterClose(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	sub := confirmedSubscription(t, c, m, 4)
	require.NoError(t, requireResult(t, waitSubscribedAsync(sub)))

	sub.Unsubscribe()
	require.NoError(t, sub.WaitSubscribed(context.Background()))
}

func TestWaitSubscribed_WaitsThroughResend(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	first := confirmedSubscription(t, c, m, 7)
	require.NoError(t, requireResult(t, waitSubscribedAsync(first)))
	second := newTestSubscription(t, c)
	secondReq := m.nextRequest(t)
	done := waitSubscribedAsync(second)

	first.Unsubscribe()
	requireUnsubscribe(t, m.nextRequest(t), 7)
	m.ack(t, secondReq.ID, 7) // id 7 is gone on the server: the client resends

	resend := m.nextRequest(t)
	require.Equal(t, "testSubscribe", resend.Method)
	requireNotSubscribed(t, second)

	m.ack(t, resend.ID, 8)
	require.NoError(t, requireResult(t, done))
}

func TestWaitSubscribed_LateConfirmationOfClosedSubscription(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	sub := newTestSubscription(t, c)
	req := m.nextRequest(t)
	sub.Unsubscribe()
	require.ErrorIs(t, sub.WaitSubscribed(context.Background()), ErrSubscriptionClosed)

	// The late confirmation is answered with an unsubscribe and leaves the
	// closed subscription as it is.
	m.ack(t, req.ID, 6)
	requireUnsubscribe(t, m.nextRequest(t), 6)
	require.ErrorIs(t, sub.WaitSubscribed(context.Background()), ErrSubscriptionClosed)
}

func TestWaitSubscribed_TypedSubscriptions(t *testing.T) {
	type waiter interface {
		WaitSubscribed(context.Context) error
	}
	key := solana.MustPublicKeyFromBase58("SysvarC1ock11111111111111111111111111111111")
	tests := []struct {
		name      string
		subscribe func(c *Client) (waiter, error)
	}{
		{"account", func(c *Client) (waiter, error) { return c.AccountSubscribe(key, "") }},
		{"block", func(c *Client) (waiter, error) { return c.BlockSubscribe(NewBlockSubscribeFilterAll(), nil) }},
		{"logs", func(c *Client) (waiter, error) { return c.LogsSubscribe(LogsSubscribeFilterAll, "") }},
		{"parsed block", func(c *Client) (waiter, error) {
			return c.ParsedBlockSubscribe(NewBlockSubscribeFilterAll(), nil)
		}},
		{"program", func(c *Client) (waiter, error) { return c.ProgramSubscribe(key, "") }},
		{"root", func(c *Client) (waiter, error) { return c.RootSubscribe() }},
		{"signature", func(c *Client) (waiter, error) { return c.SignatureSubscribe(solana.Signature{1}, "") }},
		{"slot", func(c *Client) (waiter, error) { return c.SlotSubscribe() }},
		{"slots updates", func(c *Client) (waiter, error) { return c.SlotsUpdatesSubscribe() }},
		{"vote", func(c *Client) (waiter, error) { return c.VoteSubscribe() }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockWSServer(t)
			defer m.stop()
			c := connectClient(t, m)
			defer c.Close()

			sub, err := tt.subscribe(c)
			require.NoError(t, err)
			req := m.nextRequest(t)
			done := waitSubscribedAsync(sub)

			m.ack(t, req.ID, 11)
			require.NoError(t, requireResult(t, done))
		})
	}
}
