package confirmation

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	stdjson "github.com/goccy/go-json"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

const (
	testWait = 3 * time.Second
	testTick = time.Millisecond
)

// eventLog records what the fakes saw, in order, so tests can assert the
// order of RPC calls and WebSocket messages.
type eventLog struct {
	mu     sync.Mutex
	events []string
}

func (l *eventLog) add(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, fmt.Sprintf(format, args...))
}

func (l *eventLog) list() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.events)
}

func (l *eventLog) count(event string) int {
	n := 0
	for _, e := range l.list() {
		if e == event {
			n++
		}
	}
	return n
}

// waitFor waits until event has been recorded.
func (l *eventLog) waitFor(t *testing.T, event string) {
	t.Helper()
	require.Eventually(t, func() bool { return l.count(event) > 0 }, 3*time.Second, time.Millisecond,
		"event %q not recorded; got %q", event, l.list())
}

// requireBefore asserts that the first occurrence of first comes before the
// first occurrence of second.
func (l *eventLog) requireBefore(t *testing.T, first, second string) {
	t.Helper()
	events := l.list()
	i, j := slices.Index(events, first), slices.Index(events, second)
	require.True(t, i >= 0 && j >= 0 && i < j, "want %q before %q, got %q", first, second, events)
}

// rpcHandler answers one JSON-RPC call. The result is encoded to JSON and
// decoded into the caller's result value, so the real decoders run. A string
// result is used as raw JSON.
type rpcHandler func(ctx context.Context, params []any) (any, error)

// fakeRPC implements rpc.JSONRPCClient. Each method answers with its queued
// handlers in order, and then with its default handler.
type fakeRPC struct {
	events *eventLog

	mu       sync.Mutex
	queued   map[string][]rpcHandler
	defaults map[string]rpcHandler
	calls    map[string][][]any
}

var _ rpc.JSONRPCClient = (*fakeRPC)(nil)

func newFakeRPC(events *eventLog) *fakeRPC {
	return &fakeRPC{
		events:   events,
		queued:   map[string][]rpcHandler{},
		defaults: map[string]rpcHandler{},
		calls:    map[string][][]any{},
	}
}

// on sets the handler for every call of method that has no queued handler.
func (f *fakeRPC) on(method string, h rpcHandler) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defaults[method] = h
}

// queue adds handlers that answer the next calls of method, one each.
func (f *fakeRPC) queue(method string, hs ...rpcHandler) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.queued[method] = append(f.queued[method], hs...)
}

func (f *fakeRPC) callParams(method string) [][]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.calls[method])
}

func (f *fakeRPC) CallForInto(ctx context.Context, out any, method string, params []any) error {
	f.mu.Lock()
	f.calls[method] = append(f.calls[method], params)
	var h rpcHandler
	if q := f.queued[method]; len(q) > 0 {
		h, f.queued[method] = q[0], q[1:]
	} else {
		h = f.defaults[method]
	}
	f.mu.Unlock()

	f.events.add("rpc %s", method)
	if h == nil {
		return fmt.Errorf("fakeRPC: unexpected call to %s", method)
	}
	res, err := h(ctx, params)
	if err != nil {
		return err
	}
	raw, ok := res.(string)
	if !ok {
		data, err := stdjson.Marshal(res)
		if err != nil {
			return err
		}
		raw = string(data)
	}
	return stdjson.Unmarshal([]byte(raw), out)
}

func (f *fakeRPC) CallWithCallback(context.Context, string, []any, func(*http.Request, *http.Response) error) error {
	return errors.New("fakeRPC: CallWithCallback is not supported")
}

func (f *fakeRPC) CallBatch(context.Context, jsonrpc.RPCRequests) (jsonrpc.RPCResponses, error) {
	return nil, errors.New("fakeRPC: CallBatch is not supported")
}

// result answers with the raw JSON result.
func result(raw string) rpcHandler {
	return func(context.Context, []any) (any, error) { return raw, nil }
}

// failWith answers with err.
func failWith(err error) rpcHandler {
	return func(context.Context, []any) (any, error) { return nil, err }
}

// blockUntilCanceled answers once ctx is done, which it records, like a kit
// mock that never settles.
func blockUntilCanceled(events *eventLog, method string) rpcHandler {
	return func(ctx context.Context, _ []any) (any, error) {
		<-ctx.Done()
		events.add("rpc %s canceled", method)
		return nil, ctx.Err()
	}
}

// subscribeRequest is a subscribe request the fake validator received.
type subscribeRequest struct {
	ReqID  uint64
	Method string
	Params string
	SubID  uint64
}

// fakeValidator is a WebSocket server that answers subscribe requests like
// Agave: identical subscriptions share an id, and ids start at 0.
type fakeValidator struct {
	t      *testing.T
	server *httptest.Server
	events *eventLog

	mu       sync.Mutex
	conn     *websocket.Conn
	ids      map[string]uint64
	nextID   uint64
	held     map[string]bool
	rejected map[string]string

	writeMu    sync.Mutex
	subscribes chan subscribeRequest
}

func newFakeValidator(t *testing.T, events *eventLog) *fakeValidator {
	t.Helper()
	v := &fakeValidator{
		t:          t,
		events:     events,
		ids:        map[string]uint64{},
		held:       map[string]bool{},
		rejected:   map[string]string{},
		subscribes: make(chan subscribeRequest, 100),
	}
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	v.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		v.mu.Lock()
		v.conn = conn
		v.mu.Unlock()
		defer conn.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			v.handle(msg)
		}
	}))
	t.Cleanup(v.server.Close)
	return v
}

func (v *fakeValidator) url() string {
	return "ws" + strings.TrimPrefix(v.server.URL, "http")
}

// hold keeps the confirmation of subscribe requests for method until ack is
// called.
func (v *fakeValidator) hold(method string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.held[method] = true
}

// reject answers subscribe requests for method with a JSON-RPC error.
func (v *fakeValidator) reject(method, message string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.rejected[method] = message
}

func (v *fakeValidator) handle(msg []byte) {
	var req struct {
		ID     uint64             `json:"id"`
		Method string             `json:"method"`
		Params stdjson.RawMessage `json:"params"`
	}
	if err := stdjson.Unmarshal(msg, &req); err != nil {
		v.events.add("bad request %s", msg)
		return
	}

	if strings.HasSuffix(req.Method, "Unsubscribe") {
		var params []uint64
		_ = stdjson.Unmarshal(req.Params, &params)
		v.mu.Lock()
		for key, id := range v.ids {
			if len(params) == 1 && id == params[0] {
				delete(v.ids, key)
			}
		}
		v.mu.Unlock()
		v.events.add("unsubscribe %s %v", req.Method, params)
		v.write(fmt.Sprintf(`{"jsonrpc":"2.0","result":true,"id":%d}`, req.ID))
		return
	}

	v.mu.Lock()
	if message, ok := v.rejected[req.Method]; ok {
		v.mu.Unlock()
		v.events.add("reject %s", req.Method)
		v.write(fmt.Sprintf(`{"jsonrpc":"2.0","error":{"code":-32602,"message":%q},"id":%d}`, message, req.ID))
		return
	}
	key := req.Method + string(req.Params)
	id, ok := v.ids[key]
	if !ok {
		id = v.nextID
		v.nextID++
		v.ids[key] = id
	}
	held := v.held[req.Method]
	v.mu.Unlock()

	sub := subscribeRequest{ReqID: req.ID, Method: req.Method, Params: string(req.Params), SubID: id}
	v.events.add("subscribe %s", req.Method)
	// Confirm before the test sees the request, so that a notification the
	// test sends next follows the confirmation on the wire.
	if !held {
		v.ack(sub)
	}
	v.subscribes <- sub
}

// ack confirms a subscribe request.
func (v *fakeValidator) ack(sub subscribeRequest) {
	v.events.add("ack %s", sub.Method)
	v.write(fmt.Sprintf(`{"jsonrpc":"2.0","result":%d,"id":%d}`, sub.SubID, sub.ReqID))
}

func (v *fakeValidator) write(msg string) {
	v.mu.Lock()
	conn := v.conn
	v.mu.Unlock()
	if conn == nil {
		return
	}
	v.writeMu.Lock()
	defer v.writeMu.Unlock()
	_ = conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

// nextSubscribe returns the next subscribe request. Unless its method is held,
// the confirmation has been written by then.
func (v *fakeValidator) nextSubscribe(t *testing.T) subscribeRequest {
	t.Helper()
	select {
	case sub := <-v.subscribes:
		return sub
	case <-time.After(3 * time.Second):
		t.Fatalf("no subscribe request; events %q", v.events.list())
		return subscribeRequest{}
	}
}

// notify sends a notification for the server id subID.
func (v *fakeValidator) notify(t *testing.T, method string, subID uint64, result any) {
	t.Helper()
	data, err := stdjson.Marshal(result)
	require.NoError(t, err)
	v.write(fmt.Sprintf(`{"jsonrpc":"2.0","method":%q,"params":{"result":%s,"subscription":%d}}`, method, data, subID))
}

// harness connects an RPC client and a WebSocket client to the fakes.
type harness struct {
	events    *eventLog
	rpc       *fakeRPC
	validator *fakeValidator
	rpcClient *rpc.Client
	wsClient  *ws.Client
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	events := &eventLog{}
	h := &harness{
		events:    events,
		rpc:       newFakeRPC(events),
		validator: newFakeValidator(t, events),
	}
	h.rpcClient = rpc.NewWithCustomRPCClient(h.rpc)
	wsClient, err := ws.Connect(context.Background(), h.validator.url())
	require.NoError(t, err)
	h.wsClient = wsClient
	t.Cleanup(wsClient.Close)
	return h
}

// run starts fn in a goroutine and returns a cancel func and its result.
func run(t *testing.T, fn func(ctx context.Context) error) (context.CancelFunc, <-chan error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	done := make(chan error, 1)
	go func() { done <- fn(ctx) }()
	return cancel, done
}

// requireResult waits for the result of a run.
func requireResult(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(3 * time.Second):
		t.Fatal("no result")
		return nil
	}
}

// requirePending asserts that a run has no result yet. It can only catch a
// result that arrives within a short time, so tests follow it with an event
// that settles the run.
func requirePending(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		t.Fatalf("returned %v, want it to keep waiting", err)
	case <-time.After(50 * time.Millisecond):
	}
}

// requireCanceled cancels a pending run and asserts that it reports the
// cancellation.
func requireCanceled(t *testing.T, cancel context.CancelFunc, done <-chan error) {
	t.Helper()
	requirePending(t, done)
	cancel()
	require.ErrorIs(t, requireResult(t, done), context.Canceled)
}
