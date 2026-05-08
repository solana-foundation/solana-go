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

// BlockResult is the unified notification payload for blockSubscribe.
// Exactly one of Value.Block or Value.ParsedBlock is populated per
// frame, based on the Encoding the caller passed at subscribe time.
type BlockResult struct {
	Context RPCResponseContext `json:"context"`
	Value   BlockResultValue   `json:"value"`
}

// BlockResultValue mirrors Agave's RpcBlockUpdate with a Go-style union
// over the two block shapes.
//
//	Encoding                              Field populated
//	----------------------------------------------------
//	base58 / base64 / base64+zstd         Block        (*rpc.GetBlockResult)
//	jsonParsed                            ParsedBlock  (*rpc.GetParsedBlockResult)
//
// Err / Slot are always decoded regardless of which block field is set.
type BlockResultValue struct {
	Slot uint64 `json:"slot"`
	Err  any    `json:"err,omitempty"`

	// Block is populated for binary encodings.
	Block *rpc.GetBlockResult `json:"block,omitempty"`

	// ParsedBlock is populated when Encoding=jsonParsed was requested.
	// BlockSubscribe sets this at decode time; it is not read by the
	// default json.Unmarshal path.
	ParsedBlock *rpc.GetParsedBlockResult `json:"-"`
}

type BlockSubscribeFilter interface {
	isBlockSubscribeFilter()
}

var _ BlockSubscribeFilter = BlockSubscribeFilterAll("")

type BlockSubscribeFilterAll string

func (BlockSubscribeFilterAll) isBlockSubscribeFilter() {}

type BlockSubscribeFilterMentionsAccountOrProgram struct {
	Pubkey solana.PublicKey `json:"pubkey"`
}

func (BlockSubscribeFilterMentionsAccountOrProgram) isBlockSubscribeFilter() {}

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
// BlockSubscribe subscribes to new blocks. Supports every encoding the
// RPC does — base58, base64, base64+zstd, and jsonParsed. The
// BlockResult.Value field is a union: the binary encodings populate
// Block (*rpc.GetBlockResult); jsonParsed populates ParsedBlock
// (*rpc.GetParsedBlockResult).
//
// **This subscription is unstable and only available if the validator was started
// with the `--rpc-pubsub-enable-block-subscription` flag. The format of this
// subscription may change in the future**
func (cl *Client) BlockSubscribe(
	filter BlockSubscribeFilter,
	opts *BlockSubscribeOpts,
) (*BlockSubscription, error) {
	isParsed := opts != nil && opts.Encoding == solana.EncodingJSONParsed

	var params []any
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
				solana.EncodingBase58,
				solana.EncodingBase64,
				solana.EncodingBase64Zstd,
				solana.EncodingJSONParsed,
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

	genSub, err := cl.subscribe(
		params,
		nil,
		"blockSubscribe",
		"blockUnsubscribe",
		decodeBlockNotification(isParsed),
	)
	if err != nil {
		return nil, err
	}
	return &BlockSubscription{
		sub: genSub,
	}, nil
}

// decodeBlockNotification returns a frame decoder that lands the block
// either in BlockResultValue.Block (binary) or .ParsedBlock (jsonParsed).
// Unexported but package-local so tests can exercise both branches
// without a live server.
func decodeBlockNotification(isParsed bool) func([]byte) (any, error) {
	return func(msg []byte) (any, error) {
		if isParsed {
			var tmp struct {
				Context RPCResponseContext `json:"context"`
				Value   struct {
					Slot  uint64                    `json:"slot"`
					Err   any                       `json:"err,omitempty"`
					Block *rpc.GetParsedBlockResult `json:"block,omitempty"`
				} `json:"value"`
			}
			if err := decodeResponseFromMessage(msg, &tmp); err != nil {
				return nil, err
			}
			return &BlockResult{
				Context: tmp.Context,
				Value: BlockResultValue{
					Slot:        tmp.Value.Slot,
					Err:         tmp.Value.Err,
					ParsedBlock: tmp.Value.Block,
				},
			}, nil
		}

		var res BlockResult
		if err := decodeResponseFromMessage(msg, &res); err != nil {
			return nil, err
		}
		return &res, nil
	}
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

func (sw *BlockSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
