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

type LogResult struct {
	Context struct {
		Slot uint64
	} `json:"context"`
	Value struct {
		// The transaction signature.
		Signature solana.Signature `json:"signature"`
		// Error if transaction failed, null if transaction succeeded.
		Err interface{} `json:"err"`
		// Array of log messages the transaction instructions output
		// during execution, null if simulation failed before the transaction
		// was able to execute (for example due to an invalid blockhash
		// or signature verification failure)
		Logs []string `json:"logs"`
	} `json:"value"`
}

type LogsSubscribeFilterType string

const (
	// Subscribe to all transactions except for simple vote transactions.
	LogsSubscribeFilterAll LogsSubscribeFilterType = "all"
	// Subscribe to all transactions including simple vote transactions.
	LogsSubscribeFilterAllWithVotes LogsSubscribeFilterType = "allWithVotes"
)

// LogsSubscribe subscribes to transaction logging.
func (c *Client) LogsSubscribe(
	// Filter criteria for the logs to receive results by account type.
	filter LogsSubscribeFilterType,
	commitment rpc.CommitmentType, // (optional)
) (*LogSubscription, error) {
	return c.logsSubscribe(
		filter,
		commitment,
	)
}

// LogsSubscribe subscribes to all transactions that mention the provided Pubkey.
func (c *Client) LogsSubscribeMentions(
	// Subscribe to all transactions that mention the provided Pubkey.
	mentions solana.PublicKey,
	// (optional)
	commitment rpc.CommitmentType,
) (*LogSubscription, error) {
	return c.logsSubscribe(
		rpc.M{
			"mentions": []string{mentions.String()},
		},
		commitment,
	)
}

// LogsSubscribe subscribes to transaction logging.
func (c *Client) logsSubscribe(
	filter interface{},
	commitment rpc.CommitmentType,
) (*LogSubscription, error) {

	params := []interface{}{filter}
	conf := map[string]interface{}{}
	if commitment != "" {
		conf["commitment"] = commitment
	}

	genSub, err := c.subscribe(
		params,
		conf,
		"logsSubscribe",
		"logsUnsubscribe",
		func(msg []byte) (interface{}, error) {
			var res LogResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &LogSubscription{
		sub: genSub,
	}, nil
}

type LogSubscription struct {
	sub *Subscription
}

func (sw *LogSubscription) Recv(ctx context.Context) (*LogResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-sw.sub.stream:
		if !ok {
			return nil, ErrSubscriptionClosed
		}
		return d.(*LogResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *LogSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *LogSubscription) Response() <-chan *LogResult {
	typedChan := make(chan *LogResult, 1)
	go func(ch chan *LogResult) {
		// TODO: will this subscription yield more than one result?
		d, ok := <-sw.sub.stream
		if !ok {
			return
		}
		ch <- d.(*LogResult)
	}(typedChan)
	return typedChan
}

func (sw *LogSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
