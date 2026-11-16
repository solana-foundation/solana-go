// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
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
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	"github.com/gorilla/rpc/v2/json2"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// ErrSubscriptionClosed reports that a subscription is closed: by
// Unsubscribe, by Client.Close, or because the connection ended. Subscribing
// on a client whose connection ended returns it too, wrapped.
var ErrSubscriptionClosed = errors.New("subscription closed")

type result any

// abandonedRequest describes a subscription that closed before the server
// confirmed it.
type abandonedRequest struct {
	notificationMethod string
	unsubscribeMethod  string
}

// Client is a client of the Solana PubSub WebSocket API. It is safe for
// concurrent use, and one Client can serve many subscriptions at the same
// time, including identical ones that the server gives the same id.
type Client struct {
	rpcURL        string
	conn          *websocket.Conn
	connCtx       context.Context
	connCtxCancel context.CancelFunc
	wg            sync.WaitGroup
	lock          sync.RWMutex
	// subscriptionByRequestID holds every open subscription, confirmed or
	// not, by the id of its subscribe request.
	subscriptionByRequestID map[uint64]*Subscription
	// subscriptionByWSSubID holds the confirmed subscriptions by notification
	// method and server subscription id. Validators such as Agave give
	// identical subscriptions the same id, so one key can serve several
	// subscriptions. The method is part of the key because some servers
	// number their ids per method.
	subscriptionByWSSubID map[subscriptionKey]map[*Subscription]struct{}
	// abandonedRequests holds the methods of each subscription that closed
	// before the server confirmed it, by request id. The late confirmation
	// is answered with an unsubscribe.
	abandonedRequests map[uint64]abandonedRequest
	// readerDone is set once the read loop has exited. No subscription can be
	// confirmed after that, so subscribe fails.
	readerDone bool
	shortID    bool
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Connect creates a new websocket client connecting to the provided endpoint.
func Connect(ctx context.Context, rpcEndpoint string) (c *Client, err error) {
	return ConnectWithOptions(ctx, rpcEndpoint, nil)
}

// ConnectWithOptions creates a new websocket client connecting to the provided
// endpoint with a http header if available The http header can be helpful to
// pass basic authentication params as prescribed
// ref https://github.com/gorilla/websocket/issues/209
func ConnectWithOptions(ctx context.Context, rpcEndpoint string, opt *Options) (c *Client, err error) {
	c = &Client{
		rpcURL:                  rpcEndpoint,
		subscriptionByRequestID: map[uint64]*Subscription{},
		subscriptionByWSSubID:   map[subscriptionKey]map[*Subscription]struct{}{},
		abandonedRequests:       map[uint64]abandonedRequest{},
	}

	dialer := &websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  DefaultHandshakeTimeout,
		EnableCompression: true,
	}

	if opt != nil && opt.ShortID {
		c.shortID = opt.ShortID
	}

	if opt != nil && opt.HandshakeTimeout > 0 {
		dialer.HandshakeTimeout = opt.HandshakeTimeout
	}

	var httpHeader http.Header = nil
	if opt != nil && opt.HttpHeader != nil && len(opt.HttpHeader) > 0 {
		httpHeader = opt.HttpHeader
	}
	var resp *http.Response
	c.conn, resp, err = dialer.DialContext(ctx, rpcEndpoint, httpHeader)
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			err = fmt.Errorf("new ws client: dial: %w, status: %s, body: %q", err, resp.Status, string(body))
		} else {
			err = fmt.Errorf("new ws client: dial: %w", err)
		}
		return nil, err
	}

	c.connCtx, c.connCtxCancel = context.WithCancel(context.Background())
	c.wg.Add(2)
	go func() {
		defer c.wg.Done()
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-c.connCtx.Done():
				return
			case <-ticker.C:
				c.sendPing()
			}
		}
	}()
	go c.receiveMessages()
	return c, nil
}

func (c *Client) sendPing() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
		return
	}
}

// Close cancels the connection context, closes the underlying websocket
// connection, and waits for background goroutines to finish.
func (c *Client) Close() {
	c.lock.Lock()
	c.connCtxCancel()
	c.conn.Close()
	c.lock.Unlock()
	c.wg.Wait()
}

func (c *Client) receiveMessages() {
	defer c.wg.Done()
	defer c.closeAllSubscription(ErrSubscriptionClosed)

	for {
		select {
		case <-c.connCtx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				return
			}
			c.handleMessage(message)
		}
	}
}

// GetUint64 returns the value retrieved by `Get`, cast to a uint64 if possible.
// If key data type do not match, it will return an error.
func getUint64(data []byte, keys ...string) (val uint64, err error) {
	v, t, _, e := jsonparser.Get(data, keys...)
	if e != nil {
		return 0, e
	}
	switch t {
	case jsonparser.Number, jsonparser.String:
		return strconv.ParseUint(string(v), 10, 64)
	default:
		return 0, fmt.Errorf("value is not a number/string: type=%d data=%s", t, string(v))
	}
}

func getUint64WithOk(data []byte, path ...string) (uint64, bool) {
	val, err := getUint64(data, path...)
	if err == nil {
		return val, true
	}
	return 0, false
}

// getRPCErrorWithOk extracts a JSON-RPC error object from a response message.
// A reply such as
//
//	{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"},"id":7534392980301594513}
//
// carries an `id` but no `result`, so it must be detected before the result is
// read. The boolean reports whether an error object was present.
func getRPCErrorWithOk(message []byte) (error, bool) {
	errObj, dataType, _, err := jsonparser.Get(message, "error")
	if err != nil || dataType != jsonparser.Object {
		return nil, false
	}
	code, _ := jsonparser.GetInt(errObj, "code")
	msg, _ := jsonparser.GetString(errObj, "message")
	return fmt.Errorf("rpc error (code %d): %s", code, msg), true
}

func (c *Client) handleMessage(message []byte) {
	// when receiving message with id. the result will be a subscription number.
	// that number will be associated to all future message destine to this request

	requestID, ok := getUint64WithOk(message, "id")
	if ok {
		// A reply carrying an id but an error object is a failed
		// subscribe/unsubscribe request. Surface the error to the caller
		// instead of registering a subscription whose result never arrives,
		// which would otherwise leave Recv blocked forever.
		if rpcErr, hasErr := getRPCErrorWithOk(message); hasErr {
			c.rejectSubscription(requestID, rpcErr)
			return
		}
		subID, _ := getUint64WithOk(message, "result")
		c.handleNewSubscriptionMessage(requestID, subID)
		return
	}

	subID, _ := getUint64WithOk(message, "params", "subscription")
	method, _ := jsonparser.GetString(message, "method")
	c.handleSubscriptionMessage(subscriptionKey{method: method, subID: subID}, message)
}

func (c *Client) handleNewSubscriptionMessage(requestID, subID uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if traceEnabled {
		zlog.Debug("received new subscription message",
			zap.Uint64("message_id", requestID),
			zap.Uint64("subscription_id", subID),
		)
	}

	if abandoned, ok := c.abandonedRequests[requestID]; ok {
		// The subscription closed before this confirmation arrived. Release
		// the server side, unless another subscription shares the id.
		delete(c.abandonedRequests, requestID)
		key := subscriptionKey{method: abandoned.notificationMethod, subID: subID}
		if len(c.subscriptionByWSSubID[key]) == 0 {
			if err := c.unsubscribe(subID, abandoned.unsubscribeMethod); err != nil {
				zlog.Warn("unable to send rpc unsubscribe call", zap.Error(err))
			}
		}
		return
	}

	sub, found := c.subscriptionByRequestID[requestID]
	if !found {
		zlog.Error("cannot find websocket message handler for a new stream.... this should not happen",
			zap.Uint64("request_id", requestID),
			zap.Uint64("subscription_id", subID),
		)
		return
	}

	if _, stale := sub.staleIDs[subID]; stale {
		// The server joined this subscription to an existing one, and an
		// unsubscribe written after the subscribe request has removed that
		// one since. Subscribe again.
		c.resubscribe(sub)
		return
	}

	if !sub.markSubscribed(subID) {
		return
	}
	sub.staleIDs = nil
	holders := c.subscriptionByWSSubID[sub.key()]
	if holders == nil {
		holders = map[*Subscription]struct{}{}
		c.subscriptionByWSSubID[sub.key()] = holders
	}
	holders[sub] = struct{}{}

	zlog.Debug("registered ws subscription",
		zap.Uint64("subscription_id", subID),
		zap.Uint64("request_id", requestID),
		zap.Int("subscription_count", len(c.subscriptionByWSSubID)),
	)
}

// resubscribe sends the subscribe request of sub again under a new request id.
// It must be called with c.lock held, and sub must not be confirmed yet.
func (c *Client) resubscribe(sub *Subscription) {
	delete(c.subscriptionByRequestID, sub.req.ID)
	sub.req = sub.req.withNewID(c.shortID)
	sub.staleIDs = nil
	c.subscriptionByRequestID[sub.req.ID] = sub

	data, err := sub.req.encode()
	if err == nil {
		err = c.write(data)
	}
	if err != nil {
		c.closeSubscriptionLocked(sub, fmt.Errorf("unable to write request: %w", err), true)
	}
}

func (c *Client) handleSubscriptionMessage(key subscriptionKey, message []byte) {
	if traceEnabled {
		zlog.Debug("received subscription message",
			zap.String("method", key.method),
			zap.Uint64("subscription_id", key.subID),
		)
	}

	c.lock.RLock()
	holders := make([]*Subscription, 0, len(c.subscriptionByWSSubID[key]))
	for sub := range c.subscriptionByWSSubID[key] {
		holders = append(holders, sub)
	}
	c.lock.RUnlock()
	if len(holders) == 0 {
		zlog.Warn("unable to find subscription for ws message",
			zap.String("method", key.method),
			zap.Uint64("subscription_id", key.subID),
		)
		return
	}

	for _, sub := range holders {
		c.deliver(sub, key.subID, message)
	}
}

// deliver decodes message for sub and queues the result. Each subscription
// decodes its own copy, so subscriptions that share a server id never share a
// result value.
func (c *Client) deliver(sub *Subscription, subID uint64, message []byte) {
	result, err := sub.decoderFunc(message)
	if err != nil {
		c.closeSubscriptionPtr(sub, fmt.Errorf("unable to decode client response: %w", err))
		return
	}

	// this cannot be blocking or else
	// we  will no read any other message
	if len(sub.stream) >= cap(sub.stream) {
		zlog.Warn("closing ws client subscription... not consuming fast enough",
			zap.Uint64("subscription_id", subID),
		)
		c.closeSubscriptionPtr(sub, fmt.Errorf("reached channel max capacity %d", len(sub.stream)))
		return
	}

	sub.mu.Lock()
	if !sub.closed {
		sub.stream <- result
	}
	sub.mu.Unlock()
}

// closeAllSubscription closes every subscription with err. It runs when the
// read loop exits, so it also marks the client as unable to confirm new
// subscriptions.
func (c *Client) closeAllSubscription(err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.readerDone = true
	for _, sub := range c.subscriptionByRequestID {
		sub.close(err)
	}

	c.subscriptionByRequestID = map[uint64]*Subscription{}
	c.subscriptionByWSSubID = map[subscriptionKey]map[*Subscription]struct{}{}
	c.abandonedRequests = map[uint64]abandonedRequest{}
}

// closeSubscription closes the subscription whose subscribe request has the
// id reqID, if it is still open.
func (c *Client) closeSubscription(reqID uint64, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if sub, found := c.subscriptionByRequestID[reqID]; found {
		c.closeSubscriptionLocked(sub, err, false)
	}
}

// closeSubscriptionPtr closes sub if it is still open.
func (c *Client) closeSubscriptionPtr(sub *Subscription, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.closeSubscriptionLocked(sub, err, false)
}

// rejectSubscription handles an error reply to the request with the id reqID.
// For a subscribe request it closes the subscription with err. The server
// registered nothing, so nothing is unsubscribed.
func (c *Client) rejectSubscription(reqID uint64, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.abandonedRequests, reqID)
	if sub, found := c.subscriptionByRequestID[reqID]; found {
		c.closeSubscriptionLocked(sub, err, true)
	}
}

// closeSubscriptionLocked closes sub with err. It must be called with c.lock
// held. The server side is unsubscribed when the last subscription that shares
// its id closes. A subscription that is not confirmed yet leaves a tombstone,
// so its confirmation can be answered with an unsubscribe, unless rejected
// reports that the server will never confirm it.
func (c *Client) closeSubscriptionLocked(sub *Subscription, err error, rejected bool) {
	if c.subscriptionByRequestID[sub.req.ID] != sub {
		return // already closed
	}
	delete(c.subscriptionByRequestID, sub.req.ID)
	sub.close(err)

	if !sub.acked {
		if !rejected {
			if c.abandonedRequests == nil {
				c.abandonedRequests = map[uint64]abandonedRequest{}
			}
			c.abandonedRequests[sub.req.ID] = abandonedRequest{
				notificationMethod: sub.notificationMethod,
				unsubscribeMethod:  sub.unsubscribeMethod,
			}
		}
		return
	}

	holders := c.subscriptionByWSSubID[sub.key()]
	delete(holders, sub)
	if len(holders) > 0 {
		return
	}
	delete(c.subscriptionByWSSubID, sub.key())
	if e := c.unsubscribe(sub.subID, sub.unsubscribeMethod); e != nil {
		zlog.Warn("unable to send rpc unsubscribe call",
			zap.Error(e),
		)
	}
}

// unsubscribe writes an unsubscribe request for the server id subID. It must
// be called with c.lock held. Subscriptions still waiting for their
// confirmation remember subID: if the server confirms one of them under that
// id, the unsubscribe has already removed it, and it subscribes again.
func (c *Client) unsubscribe(subID uint64, method string) error {
	req := newRequest([]any{subID}, method, nil, c.shortID)
	data, err := req.encode()
	if err != nil {
		return fmt.Errorf("unable to encode unsubscription message for subID %d and method %s: %w", subID, method, err)
	}

	if err := c.write(data); err != nil {
		return fmt.Errorf("unable to send unsubscription message for subID %d and method %s: %w", subID, method, err)
	}
	for _, pending := range c.subscriptionByRequestID {
		// Only a subscription of the same kind can be confirmed under this
		// id: a server that numbers ids per method gives the same id to
		// subscriptions of other kinds that stay live.
		if pending.acked || pending.unsubscribeMethod != method {
			continue
		}
		if pending.staleIDs == nil {
			pending.staleIDs = map[uint64]struct{}{}
		}
		pending.staleIDs[subID] = struct{}{}
	}
	return nil
}

// write sends one text message. It must be called with c.lock held.
func (c *Client) write(data []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *Client) subscribe(
	params []any,
	conf map[string]any,
	subscriptionMethod string,
	unsubscribeMethod string,
	decoderFunc decoderFunc,
) (*Subscription, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.readerDone {
		// Nothing reads the connection any more, so the server's
		// confirmation and notifications would never arrive.
		return nil, fmt.Errorf("subscribe: connection is closed: %w", ErrSubscriptionClosed)
	}

	req := newRequest(params, subscriptionMethod, conf, c.shortID)
	data, err := req.encode()
	if err != nil {
		return nil, fmt.Errorf("subscribe: unable to encode subsciption request: %w", err)
	}

	// closeFunc works on the subscription itself, because a resend changes
	// the request id.
	var sub *Subscription
	sub = newSubscription(
		req,
		func(err error) {
			c.closeSubscriptionPtr(sub, err)
		},
		unsubscribeMethod,
		decoderFunc,
	)

	c.subscriptionByRequestID[req.ID] = sub
	zlog.Info("added new subscription to websocket client", zap.Int("count", len(c.subscriptionByRequestID)))

	zlog.Debug("writing data to conn", zap.String("data", string(data)))
	if err := c.write(data); err != nil {
		delete(c.subscriptionByRequestID, req.ID)
		return nil, fmt.Errorf("unable to write request: %w", err)
	}

	return sub, nil
}

func decodeResponseFromReader(r io.Reader, reply any) (err error) {
	var c *response
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return err
	}

	if c.Error != nil {
		jsonErr := &json2.Error{}
		if err := json.Unmarshal(*c.Error, jsonErr); err != nil {
			return &json2.Error{
				Code:    json2.E_SERVER,
				Message: string(*c.Error),
			}
		}
		return jsonErr
	}

	if c.Params == nil {
		return json2.ErrNullResult
	}

	return json.Unmarshal(*c.Params.Result, &reply)
}

func decodeResponseFromMessage(r []byte, reply any) (err error) {
	var c *response
	if err := json.Unmarshal(r, &c); err != nil {
		return err
	}

	if c.Error != nil {
		jsonErr := &json2.Error{}
		if err := json.Unmarshal(*c.Error, jsonErr); err != nil {
			return &json2.Error{
				Code:    json2.E_SERVER,
				Message: string(*c.Error),
			}
		}
		return jsonErr
	}

	if c.Params == nil {
		return json2.ErrNullResult
	}

	return json.Unmarshal(*c.Params.Result, &reply)
}
