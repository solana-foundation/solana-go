// Copyright 2021 github.com/gagliardetto
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

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type AccountResult struct {
	Context RPCResponseContext `json:"context"`
	Value   *rpc.Account       `json:"value"`
}

// AccountSubscribeConfig matches Agave's RpcAccountInfoConfig. It
// supports every encoding the RPC does — base58, base64, base64+zstd,
// and jsonParsed — via rpc.Account.Data, which is a union type
// (rpc.DataBytesOrJSON) that decodes all of them transparently.
type AccountSubscribeConfig struct {
	Commitment     rpc.CommitmentType
	Encoding       solana.EncodingType
	DataSlice      *rpc.DataSlice
	MinContextSlot *uint64
}

// params converts the config to the JSON-RPC params object the
// accountSubscribe method expects. Missing options are omitted so the
// wire format matches Agave's serde(skip_serializing_if) behavior.
func (c *AccountSubscribeConfig) params() map[string]any {
	conf := map[string]any{"encoding": solana.EncodingBase64}
	if c == nil {
		return conf
	}
	if c.Commitment != "" {
		conf["commitment"] = c.Commitment
	}
	if c.Encoding != "" {
		conf["encoding"] = c.Encoding
	}
	if c.DataSlice != nil {
		conf["dataSlice"] = c.DataSlice
	}
	if c.MinContextSlot != nil {
		conf["minContextSlot"] = *c.MinContextSlot
	}
	return conf
}

// AccountSubscribe subscribes to an account to receive notifications
// when its lamports or data change. Defaults to base64 encoding.
func (cl *Client) AccountSubscribe(
	account solana.PublicKey,
	commitment rpc.CommitmentType,
) (*AccountSubscription, error) {
	return cl.AccountSubscribeWithOpts(account, commitment, "")
}

// AccountSubscribeWithOpts is the simple variant that accepts bare
// commitment and encoding arguments.
//
// Deprecated: use AccountSubscribeWithConfig for the full option set
// (dataSlice, minContextSlot) exposed by Agave's RpcAccountInfoConfig.
func (cl *Client) AccountSubscribeWithOpts(
	account solana.PublicKey,
	commitment rpc.CommitmentType,
	encoding solana.EncodingType,
) (*AccountSubscription, error) {
	return cl.AccountSubscribeWithConfig(account, &AccountSubscribeConfig{
		Commitment: commitment,
		Encoding:   encoding,
	})
}

// AccountSubscribeWithConfig mirrors the full RpcAccountInfoConfig
// option surface of the underlying accountSubscribe RPC.
func (cl *Client) AccountSubscribeWithConfig(
	account solana.PublicKey,
	config *AccountSubscribeConfig,
) (*AccountSubscription, error) {
	genSub, err := cl.subscribe(
		[]any{account.String()},
		config.params(),
		"accountSubscribe",
		"accountUnsubscribe",
		func(msg []byte) (any, error) {
			var res AccountResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &AccountSubscription{sub: genSub}, nil
}

type AccountSubscription struct {
	sub *Subscription
}

func (sw *AccountSubscription) Recv(ctx context.Context) (*AccountResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-sw.sub.stream:
		if !ok {
			return nil, ErrSubscriptionClosed
		}
		return d.(*AccountResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *AccountSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *AccountSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
