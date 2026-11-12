package ws

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// These tests pin the client's handling of server subscription ids.
// Validators such as Agave give identical subscriptions the same id, starting
// at 0, keep one entry per id and connection, and handle a connection's
// requests in order. They only use APIs that predate the fixes, so they also
// run against the unfixed client.

// wireRequest is a request the client wrote, as the mock server read it.
type wireRequest struct {
	ID     uint64 `json:"id"`
	Method string `json:"method"`
	Params []any  `json:"params"`
}

// nextRequest returns the next request the client wrote.
func (m *mockWSServer) nextRequest(t *testing.T) wireRequest {
	t.Helper()
	select {
	case msg := <-m.incoming:
		var req wireRequest
		require.NoError(t, json.Unmarshal(msg, &req), "decode client request %s", msg)
		return req
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for a client request")
		return wireRequest{}
	}
}

// endClientReads closes the server side of the connection for writing. The
// next read of the client ends with EOF, while the client can still write to
// the server. Closing the read side on the client instead does not end a
// pending read on Windows.
func (m *mockWSServer) endClientReads(t *testing.T) {
	t.Helper()
	m.connMu.Lock()
	defer m.connMu.Unlock()
	require.NotNil(t, m.conn, "no client connected")
	tcpConn, ok := m.conn.NetConn().(*net.TCPConn)
	require.True(t, ok, "want a TCP connection, got %T", m.conn.NetConn())
	require.NoError(t, tcpConn.CloseWrite())
}

// ack confirms the subscribe request reqID under the server id subID.
func (m *mockWSServer) ack(t *testing.T, reqID, subID uint64) {
	t.Helper()
	m.send(t, fmt.Sprintf(`{"jsonrpc":"2.0","result":%d,"id":%d}`, subID, reqID))
}

// requireUnsubscribe asserts that req unsubscribes the server id subID.
func requireUnsubscribe(t *testing.T, req wireRequest, subID uint64) {
	t.Helper()
	require.Equal(t, "testUnsubscribe", req.Method)
	require.Equal(t, []any{float64(subID)}, req.Params)
}

func decodeSlotResult(msg []byte) (any, error) {
	var res SlotResult
	err := decodeResponseFromMessage(msg, &res)
	return &res, err
}

// newTestSubscription writes a subscribe request and returns the subscription
// without waiting for the server.
func newTestSubscription(t *testing.T, c *Client) *Subscription {
	t.Helper()
	sub, err := c.subscribe([]any{"test"}, nil, "testSubscribe", "testUnsubscribe", decodeSlotResult)
	require.NoError(t, err)
	return sub
}

// confirmedSubscription subscribes and confirms the subscription under subID.
func confirmedSubscription(t *testing.T, c *Client, m *mockWSServer, subID uint64) *Subscription {
	t.Helper()
	sub := newTestSubscription(t, c)
	req := m.nextRequest(t)
	require.Equal(t, "testSubscribe", req.Method)
	m.ack(t, req.ID, subID)
	return sub
}

// recvSlot returns the slot of the next notification of sub.
func recvSlot(t *testing.T, sub *Subscription) uint64 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	got, err := sub.Recv(ctx)
	require.NoError(t, err)
	return got.(*SlotResult).Slot
}

// requireDelivered sends a notification for subID and asserts that every sub
// receives it. The client handles messages in order, so afterwards the
// confirmations sent before it have been processed.
func requireDelivered(t *testing.T, m *mockWSServer, subID, slot uint64, subs ...*Subscription) {
	t.Helper()
	sendSubscriptionMessage(t, m, subID, slot, 0, 0)
	for _, sub := range subs {
		require.Equal(t, slot, recvSlot(t, sub))
	}
}

// requireNextIsSubscribe writes a new subscribe request as a barrier and
// asserts that it is the next request on the wire, so nothing was written in
// between: requests are written in order on one connection.
func requireNextIsSubscribe(t *testing.T, c *Client, m *mockWSServer) {
	t.Helper()
	newTestSubscription(t, c)
	req := m.nextRequest(t)
	require.Equal(t, "testSubscribe", req.Method, "unexpected request before the barrier: %+v", req)
}

func TestSharedSubscriptionID_DeliversToEverySubscription(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	first := confirmedSubscription(t, c, m, 7)
	second := confirmedSubscription(t, c, m, 7)

	requireDelivered(t, m, 7, 100, first, second)
}

func TestSharedSubscriptionID_UnsubscribesWhenLastSubscriptionCloses(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	first := confirmedSubscription(t, c, m, 7)
	second := confirmedSubscription(t, c, m, 7)
	requireDelivered(t, m, 7, 100, first, second)

	// The server keeps one entry for id 7, so closing one subscription must
	// not unsubscribe it.
	first.Unsubscribe()
	requireNextIsSubscribe(t, c, m)
	requireDelivered(t, m, 7, 101, second)

	second.Unsubscribe()
	requireUnsubscribe(t, m.nextRequest(t), 7)
}

func TestSharedSubscriptionID_OverflowClosesOnlyThatSubscription(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	slow := confirmedSubscription(t, c, m, 7)
	fast := confirmedSubscription(t, c, m, 7)
	for range cap(slow.stream) {
		slow.stream <- &SlotResult{}
	}

	sendSubscriptionMessage(t, m, 7, 102, 0, 0)

	require.Equal(t, uint64(102), recvSlot(t, fast))
	require.Eventually(t, func() bool {
		slow.mu.Lock()
		defer slow.mu.Unlock()
		return slow.closed
	}, 2*time.Second, 10*time.Millisecond, "the full subscription should be closed")
}

func TestSubscriptionIDZero_IsUnsubscribed(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	// Agave hands out ids from 0, so 0 is a real subscription id.
	sub := confirmedSubscription(t, c, m, 0)
	requireDelivered(t, m, 0, 100, sub)
	sub.Unsubscribe()

	requireUnsubscribe(t, m.nextRequest(t), 0)
}

func TestUnconfirmedSubscriptionClose_KeepsIDZeroRoute(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	confirmed := confirmedSubscription(t, c, m, 0)
	pending := newTestSubscription(t, c)
	m.nextRequest(t) // never confirmed

	pending.Unsubscribe()

	sendSubscriptionMessage(t, m, 0, 103, 0, 0)
	require.Equal(t, uint64(103), recvSlot(t, confirmed))
}

func TestSubscribe_ResendsWhenConfirmedIDWasUnsubscribed(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	first := confirmedSubscription(t, c, m, 7)
	requireDelivered(t, m, 7, 100, first)
	second := newTestSubscription(t, c)
	secondReq := m.nextRequest(t)

	// The unsubscribe is written after the second subscribe request, so the
	// server joins the second subscription to id 7 and then removes id 7.
	first.Unsubscribe()
	requireUnsubscribe(t, m.nextRequest(t), 7)
	m.ack(t, secondReq.ID, 7)

	resend := m.nextRequest(t)
	require.Equal(t, "testSubscribe", resend.Method)
	require.Equal(t, secondReq.Params, resend.Params)
	require.NotEqual(t, secondReq.ID, resend.ID, "the resend needs a new request id")
	m.ack(t, resend.ID, 8)

	sendSubscriptionMessage(t, m, 8, 104, 0, 0)
	require.Equal(t, uint64(104), recvSlot(t, second))
}

func TestAbandonedSubscription_LateConfirmationIsUnsubscribed(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	sub := newTestSubscription(t, c)
	req := m.nextRequest(t)
	sub.Unsubscribe()

	m.ack(t, req.ID, 9)
	requireUnsubscribe(t, m.nextRequest(t), 9)
}

func TestAbandonedSubscription_LateConfirmationKeepsSharedID(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	confirmed := confirmedSubscription(t, c, m, 9)
	abandoned := newTestSubscription(t, c)
	req := m.nextRequest(t)
	abandoned.Unsubscribe()

	// Id 9 still serves the confirmed subscription, so the late confirmation
	// must not unsubscribe it. The delivery proves the read loop handled the
	// confirmation before the barrier is written.
	m.ack(t, req.ID, 9)
	requireDelivered(t, m, 9, 105, confirmed)
	requireNextIsSubscribe(t, c, m)
}

func TestSubscribe_FailsAfterReadLoopExits(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	sub := confirmedSubscription(t, c, m, 5)

	// A read error ends the read loop while the connection stays open for
	// writing.
	m.endClientReads(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := sub.Recv(ctx)
	require.ErrorIs(t, err, ErrSubscriptionClosed)

	// Nothing would read the confirmation, so subscribing must fail.
	_, err = c.subscribe([]any{"test"}, nil, "testSubscribe", "testUnsubscribe", decodeSlotResult)
	require.ErrorIs(t, err, ErrSubscriptionClosed)
}

func TestTypedRecv_ReturnsBufferedNotificationsBeforeCloseError(t *testing.T) {
	errClosed := errors.New("test close")

	// The old Recv picked at random between a buffered notification and the
	// close error, and between two closed channels afterwards.
	for i := range 100 {
		c := &Client{subscriptionByRequestID: map[uint64]*Subscription{}}
		req := &request{ID: uint64(i + 1)}
		sub := newSubscription(req, func(err error) { c.closeSubscription(req.ID, err) }, "testUnsubscribe", decodeSlotResult)
		c.subscriptionByRequestID[req.ID] = sub
		sub.stream <- &SlotResult{Slot: 42}
		c.closeSubscription(req.ID, errClosed)

		typed := &SlotSubscription{sub: sub}
		got, err := typed.Recv(context.Background())
		require.NoError(t, err, "iteration %d: the buffered notification comes first", i)
		require.Equal(t, uint64(42), got.Slot)

		for range 3 {
			got, err := typed.Recv(context.Background())
			require.ErrorIs(t, err, errClosed, "iteration %d: every later Recv returns the close error", i)
			require.ErrorIs(t, err, ErrSubscriptionClosed, "iteration %d: the close error matches ErrSubscriptionClosed", i)
			require.Nil(t, got)
		}
	}
}

// A loop that stops on ErrSubscriptionClosed and logs other errors must end
// whatever closed the subscription: a decode error, a full buffer or a server
// error reply. Before the fix the close error was returned once and
// ErrSubscriptionClosed afterwards; now the close error is returned on every
// call, so it has to match ErrSubscriptionClosed itself.
func TestRecv_CloseErrorMatchesErrSubscriptionClosed(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	sub := confirmedSubscription(t, c, m, 3)
	// A notification that fails to decode closes the subscription.
	m.send(t, `{"jsonrpc":"2.0","method":"testNotification","params":{"subscription":3,"result":"not a slot"}}`)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for i := range 3 {
		_, err := sub.Recv(ctx)
		require.ErrorIs(t, err, ErrSubscriptionClosed, "call %d", i)
		require.ErrorContains(t, err, "unable to decode client response", "call %d: the cause stays reachable", i)
	}
	require.ErrorIs(t, sub.WaitSubscribed(ctx), nil, "the subscription was confirmed before it closed")
}

func TestWaitSubscribed_ServerErrorMatchesErrSubscriptionClosed(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	sub := newTestSubscription(t, c)
	req := m.nextRequest(t)
	m.send(t, fmt.Sprintf(`{"jsonrpc":"2.0","error":{"code":-32602,"message":"Invalid params"},"id":%d}`, req.ID))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := sub.WaitSubscribed(ctx)
	require.ErrorIs(t, err, ErrSubscriptionClosed)
	require.ErrorContains(t, err, "rpc error (code -32602): Invalid params")
	_, err = sub.Recv(ctx)
	require.ErrorIs(t, err, ErrSubscriptionClosed)
	require.ErrorContains(t, err, "Invalid params")
}

// Agave numbers subscription ids per connection, but a server that numbers
// them per method would give two different subscriptions the same id. The
// notification method tells them apart.
func TestSubscriptionKey_IncludesMethod(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	slots := confirmedSubscription(t, c, m, 7)
	roots, err := c.RootSubscribe()
	require.NoError(t, err)
	rootReq := m.nextRequest(t)
	require.Equal(t, "rootSubscribe", rootReq.Method)
	m.ack(t, rootReq.ID, 7) // the same id as the slot subscription

	m.send(t, `{"jsonrpc":"2.0","method":"rootNotification","params":{"subscription":7,"result":42}}`)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	root, err := roots.Recv(ctx)
	require.NoError(t, err)
	require.Equal(t, RootResult(42), *root)

	// The root notification was not delivered to the slot subscription: the
	// next slot notification is the first thing it receives.
	requireDelivered(t, m, 7, 100, slots)

	// Each subscription unsubscribes its own id, under its own method.
	roots.Unsubscribe()
	req := m.nextRequest(t)
	require.Equal(t, "rootUnsubscribe", req.Method)
	require.Equal(t, []any{float64(7)}, req.Params)
	requireDelivered(t, m, 7, 101, slots)
	slots.Unsubscribe()
	requireUnsubscribe(t, m.nextRequest(t), 7)
}

func TestNotificationMethod(t *testing.T) {
	for subscribe, want := range map[string]string{
		"slotSubscribe":         "slotNotification",
		"accountSubscribe":      "accountNotification",
		"slotsUpdatesSubscribe": "slotsUpdatesNotification",
		"blockSubscribe":        "blockNotification",
		"testSubscribe":         "testNotification",
	} {
		require.Equal(t, want, notificationMethod(subscribe), subscribe)
	}
}
