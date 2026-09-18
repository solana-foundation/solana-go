package ws

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// captureLogs swaps the package logger for an observer for the duration of
// the test and returns the recorded entries.
func captureLogs(t *testing.T) *observer.ObservedLogs {
	t.Helper()
	core, logs := observer.New(zapcore.DebugLevel)
	prevLog, prevTrace := zlog, traceEnabled
	zlog, traceEnabled = zap.New(core), true
	t.Cleanup(func() { zlog, traceEnabled = prevLog, prevTrace })
	return logs
}

// The reply to an unsubscribe request is {"result":true,"id":N}. Its id is
// never registered as a subscription request, and its result is not a
// subscription number, so it must not be routed through the new-subscription
// path, which logged an error for every Unsubscribe call.
func TestHandleMessage_UnsubscribeAckIsNotAnError(t *testing.T) {
	logs := captureLogs(t)
	c := &Client{
		subscriptionByRequestID: map[uint64]*Subscription{},
		subscriptionByWSSubID:   map[uint64]*Subscription{},
	}

	c.handleMessage([]byte(`{"jsonrpc":"2.0","result":true,"id":7}`))

	require.Empty(t, logs.FilterLevelExact(zapcore.ErrorLevel).All(),
		"unsubscribe acknowledgement must not be logged as an error")
	require.Empty(t, c.subscriptionByWSSubID)
}

func TestHandleMessage_RejectedUnsubscribeIsWarned(t *testing.T) {
	logs := captureLogs(t)
	c := &Client{
		subscriptionByRequestID: map[uint64]*Subscription{},
		subscriptionByWSSubID:   map[uint64]*Subscription{},
	}

	c.handleMessage([]byte(`{"jsonrpc":"2.0","result":false,"id":7}`))

	require.Empty(t, logs.FilterLevelExact(zapcore.ErrorLevel).All())
	require.Len(t, logs.FilterLevelExact(zapcore.WarnLevel).All(), 1)
}

// A non-numeric result for a request that IS a pending subscribe is a
// malformed reply, not an unsubscribe acknowledgement: the subscription must
// be failed so a caller blocked in Recv is released.
func TestHandleMessage_NonNumericResultForPendingSubscribeFailsIt(t *testing.T) {
	logs := captureLogs(t)
	c := &Client{
		subscriptionByRequestID: map[uint64]*Subscription{},
		subscriptionByWSSubID:   map[uint64]*Subscription{},
	}
	req := &request{ID: 42}
	sub := newSubscription(
		req,
		func(err error) { c.closeSubscription(req.ID, err) },
		"testUnsubscribe",
		func([]byte) (any, error) { return nil, nil },
	)
	c.subscriptionByRequestID[req.ID] = sub

	c.handleMessage([]byte(`{"jsonrpc":"2.0","result":true,"id":42}`))

	select {
	case err := <-sub.err:
		require.Error(t, err)
		require.Contains(t, err.Error(), "not a subscription id")
	default:
		t.Fatal("malformed subscribe reply was not surfaced to the subscription")
	}
	require.Empty(t, c.subscriptionByRequestID)
	require.Empty(t, logs.FilterLevelExact(zapcore.ErrorLevel).All())
}

// End to end against the mock server: subscribe, unsubscribe, and have the
// server acknowledge the unsubscribe the way an RPC node does.
func TestUnsubscribe_ServerAckDoesNotLogError(t *testing.T) {
	logs := captureLogs(t)
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	sub := subscribeWithMock(t, c, m, 1)
	// subscribe returns before the reader goroutine has registered the
	// server's subscription id; Unsubscribe only sends once it is known.
	require.Eventually(t, func() bool {
		c.lock.RLock()
		defer c.lock.RUnlock()
		return sub.subID != 0
	}, 2*time.Second, 10*time.Millisecond)
	sub.Unsubscribe()

	select {
	case msg := <-m.incoming:
		reqID, ok := getUint64WithOk(msg, "id")
		require.True(t, ok, "could not parse unsubscribe request: %s", string(msg))
		m.send(t, `{"jsonrpc":"2.0","result":true,"id":`+strconv.FormatUint(reqID, 10)+`}`)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for unsubscribe request")
	}

	// Give the reader goroutine a moment to process the acknowledgement.
	require.Eventually(t, func() bool {
		return len(logs.FilterMessageSnippet("unsubscribe acknowledgement").All()) > 0 ||
			len(logs.FilterLevelExact(zapcore.ErrorLevel).All()) > 0
	}, 2*time.Second, 10*time.Millisecond, "acknowledgement was never processed")
	require.Empty(t, logs.FilterLevelExact(zapcore.ErrorLevel).All(),
		"unsubscribe acknowledgement must not be logged as an error")
}
