package rpc

import (
	"context"
	"io"
	"net/http"

	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"golang.org/x/time/rate"
)

type OptionFuncs []func(*RPCClientOpts)

type RPCClientOpts struct {
	*jsonrpc.RPCClientOpts
	limiter *rate.Limiter
}

// New creates a new Solana JSON RPC client.
// Client is safe for concurrent use by multiple goroutines.
//
// Use WithHTTPClient to set a custom HTTP client on the client.
// Use WithCustomHeaders to set custom headers on the client.
// Use WithLimiter to set a rate limiter on the client.
//
// Examples:
//
// client := rpc.New2("https://api.mainnet-beta.solana.com")
// client := rpc.New2("https://api.mainnet-beta.solana.com", rpc.WithHTTPClient(http.DefaultClient))
// client := rpc.New2("https://api.mainnet-beta.solana.com", rpc.WithHTTPClient(http.DefaultClient), rpc.WithLimiter(rate.Every(time.Second), 1))
func New2(rpcEndpoint string, fns ...func(*RPCClientOpts)) JSONRPCClient {
	opts := &RPCClientOpts{
		RPCClientOpts: &jsonrpc.RPCClientOpts{
			HTTPClient: newHTTP(),
		},
	}

	for _, f := range fns {
		f(opts)
	}

	rpcClient := jsonrpc.NewClientWithOpts(rpcEndpoint, opts.RPCClientOpts)

	return &client2{
		rpcClient: rpcClient,
		limiter:   opts.limiter,
	}
}

func WithHTTPClient(
	client *http.Client,
) func(*RPCClientOpts) {
	return func(opts *RPCClientOpts) {
		opts.HTTPClient = client
	}
}

func WithCustomHeaders(
	headers map[string]string,
) func(*RPCClientOpts) {
	return func(opts *RPCClientOpts) {
		opts.CustomHeaders = headers
	}
}

func WithLimiter(
	every rate.Limit, // time frame
	b int, // number of requests per time frame
) func(*RPCClientOpts) {
	return func(opts *RPCClientOpts) {
		opts.limiter = rate.NewLimiter(every, b)
	}
}

type client2 struct {
	rpcClient jsonrpc.RPCClient
	limiter   *rate.Limiter
}

func (cl *client2) CallForInto(
	ctx context.Context,
	out any,
	method string,
	params []any,
) error {
	if cl.limiter == nil {
		return cl.rpcClient.CallForInto(ctx, out, method, params)
	}
	if err := cl.limiter.Wait(ctx); err != nil {
		return err
	}
	return cl.rpcClient.CallForInto(ctx, &out, method, params)
}

func (cl *client2) CallWithCallback(
	ctx context.Context,
	method string,
	params []any,
	callback func(*http.Request, *http.Response) error,
) error {
	if cl.limiter == nil {
		return cl.rpcClient.CallWithCallback(ctx, method, params, callback)
	}
	if err := cl.limiter.Wait(ctx); err != nil {
		return err
	}
	return cl.rpcClient.CallWithCallback(ctx, method, params, callback)
}

func (cl *client2) CallBatch(
	ctx context.Context,
	requests jsonrpc.RPCRequests,
) (jsonrpc.RPCResponses, error) {
	if cl.limiter == nil {
		return cl.rpcClient.CallBatch(ctx, requests)
	}

	if err := cl.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return cl.rpcClient.CallBatch(ctx, requests)
}

// Close closes.
func (cl *client2) Close() error {
	if c, ok := cl.rpcClient.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
