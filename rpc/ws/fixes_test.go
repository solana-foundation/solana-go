package ws

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWSServer creates a test WebSocket server that can be controlled from tests.
type mockWSServer struct {
	server    *httptest.Server
	connMu    sync.Mutex
	conn      *websocket.Conn
	incoming  chan []byte
	closeOnce sync.Once
	closed    chan struct{}
}

func newMockWSServer(t *testing.T) *mockWSServer {
	t.Helper()
	m := &mockWSServer{
		incoming: make(chan []byte, 100),
		closed:   make(chan struct{}),
	}

	upgrader := websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade error: %v", err)
			return
		}
		m.connMu.Lock()
		m.conn = conn
		m.connMu.Unlock()

		defer conn.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				m.closeOnce.Do(func() { close(m.closed) })
				return
			}
			m.incoming <- msg
		}
	}))
	return m
}

func (m *mockWSServer) wsURL() string {
	return "ws" + strings.TrimPrefix(m.server.URL, "http")
}

// trySend sends a message to the connected client. Returns false if the
// connection is nil or the write fails — safe to call from non-test goroutines.
func (m *mockWSServer) trySend(msg string) bool {
	m.connMu.Lock()
	defer m.connMu.Unlock()
	if m.conn == nil {
		return false
	}
	return m.conn.WriteMessage(websocket.TextMessage, []byte(msg)) == nil
}

// send sends a message and fails the test on error. Only call from the test
// goroutine (not from spawned goroutines).
func (m *mockWSServer) send(t *testing.T, msg string) {
	t.Helper()
	m.connMu.Lock()
	defer m.connMu.Unlock()
	require.NotNil(t, m.conn, "no client connected")
	err := m.conn.WriteMessage(websocket.TextMessage, []byte(msg))
	require.NoError(t, err)
}

func (m *mockWSServer) closeConn() {
	m.connMu.Lock()
	defer m.connMu.Unlock()
	if m.conn != nil {
		m.conn.Close()
	}
}

func (m *mockWSServer) stop() {
	m.closeConn()
	m.server.Close()
}

// waitForSubscriptionAndRespond reads the subscribe request from the mock server's
// incoming channel and sends back a subscription confirmation with the given subID.
func (m *mockWSServer) waitForSubscriptionAndRespond(t *testing.T, subID uint64) {
	t.Helper()
	select {
	case msg := <-m.incoming:
		reqID, ok := getUint64WithOk(msg, "id")
		require.True(t, ok, "could not parse request ID from: %s", string(msg))
		resp := fmt.Sprintf(`{"jsonrpc":"2.0","result":%d,"id":%d}`, subID, reqID)
		m.send(t, resp)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for subscription request")
	}
}

// connectClient creates a ws.Client connected to the mock server.
func connectClient(t *testing.T, m *mockWSServer) *Client {
	t.Helper()
	c, err := Connect(context.Background(), m.wsURL())
	require.NoError(t, err)
	return c
}

// subscribeWithMock creates a subscription on the client, and confirms it on
// the mock server side, returning the Subscription.
func subscribeWithMock(t *testing.T, c *Client, m *mockWSServer, wsSubID uint64) *Subscription {
	t.Helper()

	type subResult struct {
		sub *Subscription
		err error
	}
	ch := make(chan subResult, 1)

	go func() {
		sub, err := c.subscribe(
			[]any{"test"},
			nil,
			"testSubscribe",
			"testUnsubscribe",
			func(msg []byte) (any, error) {
				var res SlotResult
				err := decodeResponseFromMessage(msg, &res)
				return &res, err
			},
		)
		ch <- subResult{sub, err}
	}()

	m.waitForSubscriptionAndRespond(t, wsSubID)

	r := <-ch
	require.NoError(t, r.err)
	return r.sub
}

// sendSubscriptionMessage sends a slot notification for the given wsSubID.
// Only call from the test goroutine.
func sendSubscriptionMessage(t *testing.T, m *mockWSServer, wsSubID uint64, slot, parent, root uint64) {
	t.Helper()
	msg := fmt.Sprintf(
		`{"jsonrpc":"2.0","method":"testNotification","params":{"subscription":%d,"result":{"parent":%d,"root":%d,"slot":%d}}}`,
		wsSubID, parent, root, slot,
	)
	m.send(t, msg)
}

// trySendSubscriptionMessage is safe to call from non-test goroutines.
func trySendSubscriptionMessage(m *mockWSServer, wsSubID uint64, slot uint64) {
	msg := fmt.Sprintf(
		`{"jsonrpc":"2.0","method":"testNotification","params":{"subscription":%d,"result":{"parent":0,"root":0,"slot":%d}}}`,
		wsSubID, slot,
	)
	m.trySend(msg)
}

// --- Tests ---

func TestClose_WaitsForGoroutines(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)

	done := make(chan struct{})
	go func() {
		c.Close()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Close() did not return in time — goroutines likely leaked")
	}
}

func TestCloseAllSubscription_ClosesChannelsOnConnectionDrop(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)
	defer c.Close()

	sub := subscribeWithMock(t, c, m, 42)

	// Drop the server-side connection to trigger receiveMessages exit.
	m.closeConn()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := sub.Recv(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSubscriptionClosed)
}

func TestRecv_ReturnsClosedAfterUnsubscribe(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)
	defer c.Close()

	sub := subscribeWithMock(t, c, m, 42)

	sub.Unsubscribe()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := sub.Recv(ctx)
	require.Error(t, err)
}

func TestRecv_DeliversMessages(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)
	defer c.Close()

	sub := subscribeWithMock(t, c, m, 42)

	sendSubscriptionMessage(t, m, 42, 100, 99, 98)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	got, err := sub.Recv(ctx)
	require.NoError(t, err)
	slotResult, ok := got.(*SlotResult)
	require.True(t, ok)
	assert.Equal(t, uint64(100), slotResult.Slot)
	assert.Equal(t, uint64(99), slotResult.Parent)
	assert.Equal(t, uint64(98), slotResult.Root)
}

func TestRecv_ContextCancellation(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)
	defer c.Close()

	sub := subscribeWithMock(t, c, m, 42)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := sub.Recv(ctx)
	require.ErrorIs(t, err, context.Canceled)

	sub.Unsubscribe()
}

func TestConcurrentUnsubscribeDuringMessageDelivery_NoPanic(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)
	defer c.Close()

	sub := subscribeWithMock(t, c, m, 42)

	var wg sync.WaitGroup

	// Sender goroutine: sends messages as fast as possible.
	// Uses trySend to avoid calling t.Fatal from a non-test goroutine.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range 50 {
			trySendSubscriptionMessage(m, 42, uint64(i))
			time.Sleep(time.Millisecond)
		}
	}()

	// Give some messages time to arrive before unsubscribing.
	time.Sleep(10 * time.Millisecond)

	// Unsubscribe from another goroutine — this used to panic with
	// "send on closed channel" before the mutex fix.
	wg.Add(1)
	go func() {
		defer wg.Done()
		sub.Unsubscribe()
	}()

	wg.Wait()
}

func TestDoubleUnsubscribe_NoPanic(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)
	defer c.Close()

	sub := subscribeWithMock(t, c, m, 42)

	sub.Unsubscribe()
	sub.Unsubscribe()
}

func TestCloseSubscription_Idempotent(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)
	defer c.Close()

	sub := subscribeWithMock(t, c, m, 42)
	reqID := sub.req.ID

	c.closeSubscription(reqID, fmt.Errorf("test error"))
	c.closeSubscription(reqID, fmt.Errorf("test error again"))
}

func TestStreamCapacityOverflow_ClosesSubscription(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)
	defer c.Close()

	sub := subscribeWithMock(t, c, m, 42)

	// Fill the stream channel to capacity.
	for range cap(sub.stream) {
		sub.stream <- &SlotResult{Slot: 0}
	}

	// Next message delivery should trigger a close due to capacity overflow.
	sendSubscriptionMessage(t, m, 42, 999, 0, 0)

	require.Eventually(t, func() bool {
		sub.mu.Lock()
		defer sub.mu.Unlock()
		return sub.closed
	}, 2*time.Second, 10*time.Millisecond, "subscription should be closed after capacity overflow")
}

func TestMultipleSubscriptions_IndependentLifecycles(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)
	defer c.Close()

	sub1 := subscribeWithMock(t, c, m, 10)
	sub2 := subscribeWithMock(t, c, m, 20)

	sendSubscriptionMessage(t, m, 20, 200, 0, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	got, err := sub2.Recv(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(200), got.(*SlotResult).Slot)

	// Unsubscribe sub1, sub2 should still work.
	sub1.Unsubscribe()

	sendSubscriptionMessage(t, m, 20, 300, 0, 0)
	got, err = sub2.Recv(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(300), got.(*SlotResult).Slot)

	sub2.Unsubscribe()
}

func TestConnectionDrop_AllSubscriptionsNotified(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)
	defer c.Close()

	sub1 := subscribeWithMock(t, c, m, 10)
	sub2 := subscribeWithMock(t, c, m, 20)
	sub3 := subscribeWithMock(t, c, m, 30)

	m.closeConn()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for _, sub := range []*Subscription{sub1, sub2, sub3} {
		_, err := sub.Recv(ctx)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSubscriptionClosed)
	}
}

func TestSubscribe_ErrorResponseClosesSubscription(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)
	defer c.Close()

	type subResult struct {
		sub *Subscription
		err error
	}
	ch := make(chan subResult, 1)
	go func() {
		sub, err := c.subscribe(
			[]any{"test"},
			nil,
			"testSubscribe",
			"testUnsubscribe",
			func(msg []byte) (any, error) {
				var res SlotResult
				err := decodeResponseFromMessage(msg, &res)
				return &res, err
			},
		)
		ch <- subResult{sub, err}
	}()

	var reqID uint64
	select {
	case msg := <-m.incoming:
		id, ok := getUint64WithOk(msg, "id")
		require.True(t, ok, "could not parse request ID from %s", string(msg))
		reqID = id
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for subscription request")
	}

	resp := fmt.Sprintf(`{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"},"id":%d}`, reqID)
	m.send(t, resp)

	r := <-ch
	require.NoError(t, r.err)
	require.NotNil(t, r.sub)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := r.sub.Recv(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Method not found")
}

func TestConcurrentUnsubscribeAndConnectionDrop_NoPanic(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()

	c := connectClient(t, m)

	sub1 := subscribeWithMock(t, c, m, 10)
	sub2 := subscribeWithMock(t, c, m, 20)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		sub1.Unsubscribe()
	}()
	go func() {
		defer wg.Done()
		m.closeConn()
	}()
	wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := sub2.Recv(ctx)
	if err != nil {
		assert.True(t, err == ErrSubscriptionClosed || err == context.DeadlineExceeded,
			"unexpected error: %v", err)
	}

	c.Close()
}

func TestSubscription_RecvAfterChannelsClosed(t *testing.T) {
	sub := newSubscription(
		&request{ID: 1},
		func(err error) {},
		"testUnsubscribe",
		func(msg []byte) (any, error) { return nil, nil },
	)

	// Manually close channels as closeAllSubscription would.
	sub.mu.Lock()
	sub.err <- ErrSubscriptionClosed
	sub.closed = true
	close(sub.stream)
	close(sub.err)
	sub.mu.Unlock()

	ctx := context.Background()
	_, err := sub.Recv(ctx)
	assert.ErrorIs(t, err, ErrSubscriptionClosed)

	// Subsequent Recv returns ErrSubscriptionClosed from closed channels.
	_, err = sub.Recv(ctx)
	assert.ErrorIs(t, err, ErrSubscriptionClosed)
}

func TestSubscription_BufferSizes(t *testing.T) {
	sub := newSubscription(
		&request{ID: 1},
		func(err error) {},
		"testUnsubscribe",
		func(msg []byte) (any, error) { return nil, nil },
	)

	assert.Equal(t, 200, cap(sub.stream), "stream buffer should be 200")
	assert.Equal(t, 1, cap(sub.err), "err buffer should be 1")
}
