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

type ProgramResult struct {
	Context RPCResponseContext `json:"context"`
	Value   rpc.KeyedAccount   `json:"value"`
}

// ProgramSubscribeConfig matches Agave's RpcProgramAccountsConfig. All
// encoding types supported by getProgramAccounts are supported here too
// (base58, base64, base64+zstd, jsonParsed) — the per-account data
// decodes via rpc.KeyedAccount.Account.Data (a DataBytesOrJSON union).
type ProgramSubscribeConfig struct {
	Commitment     rpc.CommitmentType
	Encoding       solana.EncodingType
	DataSlice      *rpc.DataSlice
	Filters        []rpc.RPCFilter
	MinContextSlot *uint64
	WithContext    *bool
	SortResults    *bool
}

// params converts the config to the JSON-RPC params object the
// programSubscribe method expects. Missing options are omitted so the
// wire format matches Agave's serde(skip_serializing_if) behavior.
func (c *ProgramSubscribeConfig) params() map[string]any {
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
	if len(c.Filters) > 0 {
		conf["filters"] = c.Filters
	}
	if c.MinContextSlot != nil {
		conf["minContextSlot"] = *c.MinContextSlot
	}
	if c.WithContext != nil {
		conf["withContext"] = *c.WithContext
	}
	if c.SortResults != nil {
		conf["sortResults"] = *c.SortResults
	}
	return conf
}

// ProgramSubscribe subscribes to a program to receive notifications
// when the lamports or data for any account owned by the program change.
func (cl *Client) ProgramSubscribe(
	programID solana.PublicKey,
	commitment rpc.CommitmentType,
) (*ProgramSubscription, error) {
	return cl.ProgramSubscribeWithOpts(programID, commitment, "", nil)
}

// ProgramSubscribeWithOpts is the simple variant that accepts bare
// commitment, encoding, and filters.
//
// Deprecated: use ProgramSubscribeWithConfig for the full option set
// (dataSlice, minContextSlot, withContext, sortResults) exposed by
// Agave's RpcProgramAccountsConfig.
func (cl *Client) ProgramSubscribeWithOpts(
	programID solana.PublicKey,
	commitment rpc.CommitmentType,
	encoding solana.EncodingType,
	filters []rpc.RPCFilter,
) (*ProgramSubscription, error) {
	return cl.ProgramSubscribeWithConfig(programID, &ProgramSubscribeConfig{
		Commitment: commitment,
		Encoding:   encoding,
		Filters:    filters,
	})
}

// ProgramSubscribeWithConfig mirrors the full RpcProgramAccountsConfig
// option surface of the underlying programSubscribe RPC.
func (cl *Client) ProgramSubscribeWithConfig(
	programID solana.PublicKey,
	config *ProgramSubscribeConfig,
) (*ProgramSubscription, error) {
	genSub, err := cl.subscribe(
		[]any{programID.String()},
		config.params(),
		"programSubscribe",
		"programUnsubscribe",
		func(msg []byte) (any, error) {
			var res ProgramResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &ProgramSubscription{sub: genSub}, nil
}

type ProgramSubscription struct {
	sub *Subscription
}

func (sw *ProgramSubscription) Recv(ctx context.Context) (*ProgramResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-sw.sub.stream:
		if !ok {
			return nil, ErrSubscriptionClosed
		}
		return d.(*ProgramResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *ProgramSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *ProgramSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
