// Copyright 2022 github.com/gagliardetto
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
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type BlockResult struct {
	Context struct {
		Slot uint64
	} `json:"context"`
	Value struct {
		Slot  uint64              `json:"slot"`
		Err   interface{}         `json:"err,omitempty"`
		Block *rpc.GetBlockResult `json:"block,omitempty"`
	} `json:"value"`
}

type BlockSubscribeFilter interface {
	isBlockSubscribeFilter()
}

var _ BlockSubscribeFilter = BlockSubscribeFilterAll("")

type BlockSubscribeFilterAll string

func (_ BlockSubscribeFilterAll) isBlockSubscribeFilter() {}

type BlockSubscribeFilterMentionsAccountOrProgram struct {
	Pubkey solana.PublicKey `json:"pubkey"`
}

func (_ BlockSubscribeFilterMentionsAccountOrProgram) isBlockSubscribeFilter() {}

func NewBlockSubscribeFilterAll() BlockSubscribeFilter {
	return BlockSubscribeFilterAll("")
}

func NewBlockSubscribeFilterMentionsAccountOrProgram(pubkey solana.PublicKey) *BlockSubscribeFilterMentionsAccountOrProgram {
	return &BlockSubscribeFilterMentionsAccountOrProgram{
		Pubkey: pubkey,
	}
}

type BlockSubscribeOpts struct {
	Commitment rpc.CommitmentType
	Encoding   solana.EncodingType `json:"encoding,omitempty"`

	// Level of transaction detail to return.
	TransactionDetails rpc.TransactionDetailsType

	// Whether to populate the rewards array. If parameter not provided, the default includes rewards.
	Rewards *bool

	// Max transaction version to return in responses.
	// If the requested block contains a transaction with a higher version, an error will be returned.
	MaxSupportedTransactionVersion *uint64
}

// NOTE: Unstable, disabled by default
//
// Subscribe to receive notification anytime a new block is Confirmed or Finalized.
//
// **This subscription is unstable and only available if the validator was started
// with the `--rpc-pubsub-enable-block-subscription` flag. The format of this
// subscription may change in the future**
func (c *Client) BlockSubscribe(
	filter BlockSubscribeFilter,
	opts *BlockSubscribeOpts,
) (*BlockSubscription, error) {
	var params []interface{}
	if filter != nil {
		switch v := filter.(type) {
		case BlockSubscribeFilterAll:
			params = append(params, "all")
		case *BlockSubscribeFilterMentionsAccountOrProgram:
			params = append(params, rpc.M{"mentionsAccountOrProgram": v.Pubkey})
		}
	}
	if opts != nil {
		obj := make(rpc.M)
		if opts.Commitment != "" {
			obj["commitment"] = opts.Commitment
		}
		if opts.Encoding != "" {
			if !solana.IsAnyOfEncodingType(
				opts.Encoding,
				// Valid encodings:
				// solana.EncodingJSON, // TODO
				solana.EncodingJSONParsed, // TODO
				solana.EncodingBase58,
				solana.EncodingBase64,
				solana.EncodingBase64Zstd,
			) {
				return nil, fmt.Errorf("provided encoding is not supported: %s", opts.Encoding)
			}
			obj["encoding"] = opts.Encoding
		}
		if opts.TransactionDetails != "" {
			obj["transactionDetails"] = opts.TransactionDetails
		}
		if opts.Rewards != nil {
			obj["rewards"] = opts.Rewards
		}
		if opts.MaxSupportedTransactionVersion != nil {
			obj["maxSupportedTransactionVersion"] = *opts.MaxSupportedTransactionVersion
		}
		if len(obj) > 0 {
			params = append(params, obj)
		}
	}
	genSub, err := c.subscribe(
		params,
		nil,
		"blockSubscribe",
		"blockUnsubscribe",
		func(msg []byte) (interface{}, error) {
			var res BlockResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &BlockSubscription{
		sub: genSub,
	}, nil
}

type BlockSubscription struct {
	sub *Subscription
}

func (sw *BlockSubscription) Recv(ctx context.Context) (*BlockResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-sw.sub.stream:
		if !ok {
			return nil, ErrSubscriptionClosed
		}
		return d.(*BlockResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *BlockSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *BlockSubscription) Response() <-chan *BlockResult {
	typedChan := make(chan *BlockResult, 1)
	go func(ch chan *BlockResult) {
		// TODO: will this subscription yield more than one result?
		d, ok := <-sw.sub.stream
		if !ok {
			return
		}
		ch <- d.(*BlockResult)
	}(typedChan)
	return typedChan
}

func (sw *BlockSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
