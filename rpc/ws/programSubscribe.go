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

// ProgramSubscribe subscribes to a program to receive notifications
// when the lamports or data for a given account owned by the program changes.
func (c *Client) ProgramSubscribe(
	programID solana.PublicKey,
	commitment rpc.CommitmentType,
) (*ProgramSubscription, error) {
	return c.ProgramSubscribeWithOpts(
		programID,
		commitment,
		"",
		nil,
	)
}

// ProgramSubscribe subscribes to a program to receive notifications
// when the lamports or data for a given account owned by the program changes.
func (c *Client) ProgramSubscribeWithOpts(
	programID solana.PublicKey,
	commitment rpc.CommitmentType,
	encoding solana.EncodingType,
	filters []rpc.RPCFilter,
) (*ProgramSubscription, error) {

	params := []interface{}{programID.String()}
	conf := map[string]interface{}{
		"encoding": "base64",
	}
	if commitment != "" {
		conf["commitment"] = commitment
	}
	if encoding != "" {
		conf["encoding"] = encoding
	}
	if filters != nil && len(filters) > 0 {
		conf["filters"] = filters
	}

	genSub, err := c.subscribe(
		params,
		conf,
		"programSubscribe",
		"programUnsubscribe",
		func(msg []byte) (interface{}, error) {
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

func (sw *ProgramSubscription) Response() <-chan *ProgramResult {
	typedChan := make(chan *ProgramResult, 1)
	go func(ch chan *ProgramResult) {
		// TODO: will this subscription yield more than one result?
		d, ok := <-sw.sub.stream
		if !ok {
			return
		}
		ch <- d.(*ProgramResult)
	}(typedChan)
	return typedChan
}

func (sw *ProgramSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
