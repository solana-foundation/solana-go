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

package rpc

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/klauspost/compress/gzhttp"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrNotConfirmed = errors.New("not confirmed")
)

type Client struct {
	rpcURL            string
	rpcClient         JSONRPCClient
	defaultCommitment CommitmentType
}

type JSONRPCClient interface {
	CallForInto(ctx context.Context, out any, method string, params []any) error
	CallWithCallback(ctx context.Context, method string, params []any, callback func(*http.Request, *http.Response) error) error
	CallBatch(ctx context.Context, requests jsonrpc.RPCRequests) (jsonrpc.RPCResponses, error)
}

// New creates a new Solana JSON RPC client.
// Client is safe for concurrent use by multiple goroutines.
func New(rpcEndpoint string) *Client {
	opts := &jsonrpc.RPCClientOpts{
		HTTPClient: newHTTP(),
	}

	rpcClient := jsonrpc.NewClientWithOpts(rpcEndpoint, opts)
	return NewWithCustomRPCClient(rpcClient)
}

// New creates a new Solana JSON RPC client with the provided custom headers.
// The provided headers will be added to each RPC request sent via this RPC client.
func NewWithHeaders(rpcEndpoint string, headers map[string]string) *Client {
	opts := &jsonrpc.RPCClientOpts{
		HTTPClient:    newHTTP(),
		CustomHeaders: headers,
	}
	rpcClient := jsonrpc.NewClientWithOpts(rpcEndpoint, opts)
	return NewWithCustomRPCClient(rpcClient)
}

// NewWithCommitment creates a new Solana JSON RPC client and pins a default
// CommitmentType on the returned Client. Methods that take an explicit
// CommitmentType still receive whatever the caller passes; the stored
// commitment is exposed via Client.DefaultCommitment so callers can fall
// back to it without threading the value through every call site
// themselves. Mirrors the rust-sdk RpcClient::new_with_commitment ergonomics.
func NewWithCommitment(rpcEndpoint string, commitment CommitmentType) *Client {
	cl := New(rpcEndpoint)
	cl.defaultCommitment = commitment
	return cl
}

// NewWithTimeout creates a new Solana JSON RPC client with a custom HTTP
// timeout. The default 5-minute timeout used by New is replaced with the
// supplied value on the underlying *http.Client; the same value is also
// applied to the dialer and idle connection timeout so long-haul reads,
// connect, and pool eviction stay aligned.
func NewWithTimeout(rpcEndpoint string, timeout time.Duration) *Client {
	opts := &jsonrpc.RPCClientOpts{
		HTTPClient: newHTTPWithTimeout(timeout),
	}
	rpcClient := jsonrpc.NewClientWithOpts(rpcEndpoint, opts)
	return NewWithCustomRPCClient(rpcClient)
}

// NewWithTimeoutAndCommitment combines NewWithTimeout and NewWithCommitment.
// Mirrors the rust-sdk RpcClient::new_with_timeout_and_commitment
// constructor.
func NewWithTimeoutAndCommitment(
	rpcEndpoint string,
	timeout time.Duration,
	commitment CommitmentType,
) *Client {
	cl := NewWithTimeout(rpcEndpoint, timeout)
	cl.defaultCommitment = commitment
	return cl
}

// DefaultCommitment returns the CommitmentType pinned on this Client at
// construction time via NewWithCommitment / NewWithTimeoutAndCommitment.
// Returns the empty CommitmentType when no default was configured.
func (cl *Client) DefaultCommitment() CommitmentType {
	return cl.defaultCommitment
}

// Close closes the client.
func (cl *Client) Close() error {
	if cl.rpcClient == nil {
		return nil
	}
	if c, ok := cl.rpcClient.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// NewWithCustomRPCClient creates a new Solana RPC client
// with the provided RPC client.
func NewWithCustomRPCClient(rpcClient JSONRPCClient) *Client {
	return &Client{
		rpcClient: rpcClient,
	}
}

var (
	defaultMaxIdleConnsPerHost = 9
	defaultTimeout             = 5 * time.Minute
	defaultKeepAlive           = 180 * time.Second
)

func newHTTPTransport() *http.Transport {
	return &http.Transport{
		IdleConnTimeout:     defaultTimeout,
		MaxConnsPerHost:     defaultMaxIdleConnsPerHost,
		MaxIdleConnsPerHost: defaultMaxIdleConnsPerHost,
		Proxy:               http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   defaultTimeout,
			KeepAlive: defaultKeepAlive,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2: true,
		// MaxIdleConns:          100,
		TLSHandshakeTimeout: 10 * time.Second,
		// ExpectContinueTimeout: 1 * time.Second,
	}
}

// newHTTP returns a new Client from the provided config.
// Client is safe for concurrent use by multiple goroutines.
func newHTTP() *http.Client {
	return newHTTPWithTimeout(defaultTimeout)
}

// newHTTPWithTimeout returns a new *http.Client whose request timeout, dial
// timeout, and idle connection timeout are all bound to the supplied value.
// Used by NewWithTimeout / NewWithTimeoutAndCommitment so callers can lift
// the hardcoded 5-minute ceiling without dropping into newHTTPTransport.
func newHTTPWithTimeout(timeout time.Duration) *http.Client {
	tr := &http.Transport{
		IdleConnTimeout:     timeout,
		MaxConnsPerHost:     defaultMaxIdleConnsPerHost,
		MaxIdleConnsPerHost: defaultMaxIdleConnsPerHost,
		Proxy:               http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: defaultKeepAlive,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:   true,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: gzhttp.Transport(tr),
	}
}

// RPCCallForInto allows to access the raw RPC client and send custom requests.
func (cl *Client) RPCCallForInto(ctx context.Context, out any, method string, params []any) error {
	return cl.rpcClient.CallForInto(ctx, out, method, params)
}

func (cl *Client) RPCCallWithCallback(
	ctx context.Context,
	method string,
	params []any,
	callback func(*http.Request, *http.Response) error,
) error {
	return cl.rpcClient.CallWithCallback(ctx, method, params, callback)
}

func (cl *Client) RPCCallBatch(
	ctx context.Context,
	requests jsonrpc.RPCRequests,
) (jsonrpc.RPCResponses, error) {
	return cl.rpcClient.CallBatch(ctx, requests)
}

func NewBoolean(b bool) *bool {
	return &b
}

func NewTransactionVersion(v uint64) *uint64 {
	return &v
}
