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
	Context struct {
		Slot uint64
	} `json:"context"`
	Value rpc.KeyedAccount `json:"value"`
}

// ProgramSubscribeOpts mirrors the optional configuration object the
// programSubscribe RPC method accepts. See
// https://solana.com/docs/rpc/websocket/programsubscribe.
type ProgramSubscribeOpts struct {
	// Commitment selects the bank state the validator should query.
	Commitment rpc.CommitmentType

	// Encoding controls how the account data field is encoded. When
	// empty the request defaults to "base64".
	Encoding solana.EncodingType

	// Filters narrows the stream of account-change notifications to
	// the subset that match every supplied filter (memcmp/dataSize).
	Filters []rpc.RPCFilter

	// DataSlice asks the validator to return only the requested slice
	// of each notified account's data field. Only valid for binary
	// encodings (base58, base64, base64+zstd) per the RPC spec.
	DataSlice *rpc.DataSlice
}

// ProgramSubscribe subscribes to a program to receive notifications
// when the lamports or data for a given account owned by the program changes.
func (cl *Client) ProgramSubscribe(
	programID solana.PublicKey,
	commitment rpc.CommitmentType,
) (*ProgramSubscription, error) {
	return cl.ProgramSubscribeWithOpts(
		programID,
		commitment,
		"",
		nil,
	)
}

// ProgramSubscribeWithOpts subscribes to a program with explicit
// commitment, encoding, and filters. Kept for backward compatibility;
// new callers should prefer ProgramSubscribeWithConfig which exposes
// the full optional configuration object (including DataSlice).
func (cl *Client) ProgramSubscribeWithOpts(
	programID solana.PublicKey,
	commitment rpc.CommitmentType,
	encoding solana.EncodingType,
	filters []rpc.RPCFilter,
) (*ProgramSubscription, error) {
	return cl.ProgramSubscribeWithConfig(programID, &ProgramSubscribeOpts{
		Commitment: commitment,
		Encoding:   encoding,
		Filters:    filters,
	})
}

// ProgramSubscribeWithConfig subscribes to a program and forwards the
// full ProgramSubscribeOpts configuration object to the validator,
// including DataSlice.
func (cl *Client) ProgramSubscribeWithConfig(
	programID solana.PublicKey,
	opts *ProgramSubscribeOpts,
) (*ProgramSubscription, error) {
	params := []any{programID.String()}
	conf := map[string]any{
		"encoding": "base64",
	}
	if opts != nil {
		if opts.Commitment != "" {
			conf["commitment"] = opts.Commitment
		}
		if opts.Encoding != "" {
			conf["encoding"] = opts.Encoding
		}
		if len(opts.Filters) > 0 {
			conf["filters"] = opts.Filters
		}
		if opts.DataSlice != nil {
			conf["dataSlice"] = opts.DataSlice
		}
	}

	genSub, err := cl.subscribe(
		params,
		conf,
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
	return &ProgramSubscription{
		sub: genSub,
	}, nil
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
