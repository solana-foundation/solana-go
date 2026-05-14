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
	"bytes"
	"context"
	"fmt"

	stdjson "github.com/goccy/go-json"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// signatureReceivedMarker is the literal value the validator emits
// for the "received" notification when EnableReceivedNotification is
// set on the subscription request.
const signatureReceivedMarker = "receivedSignature"

type SignatureResult struct {
	Context struct {
		Slot uint64
	} `json:"context"`
	Value SignatureValue `json:"value"`
}

// SignatureValue carries the two shapes a signatureNotification can take.
//
// When EnableReceivedNotification is false (the RPC default) only the
// status shape is emitted and `Err` is the field consumers need
// (nil == success).
//
// When EnableReceivedNotification is true the validator may additionally
// emit a "receivedSignature" string before the final status; that one
// sets `ReceivedSignature` to true and `Err` is left at its zero value.
type SignatureValue struct {
	// Err is the transaction execution error reported in a status
	// notification. nil means the transaction succeeded.
	Err any `json:"err,omitempty"`

	// ReceivedSignature is true when the notification is the "received"
	// marker the validator emits once it observes the transaction in
	// its mempool. Only sent when EnableReceivedNotification is set on
	// SignatureSubscribeOpts.
	ReceivedSignature bool `json:"-"`
}

// UnmarshalJSON dispatches on the wire shape of the notification value:
// a JSON string "receivedSignature" → ReceivedSignature=true; a JSON
// object → status, fill Err.
func (v *SignatureValue) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil
	}
	switch trimmed[0] {
	case '"':
		var s string
		if err := stdjson.Unmarshal(trimmed, &s); err != nil {
			return fmt.Errorf("signatureNotification value: %w", err)
		}
		if s != signatureReceivedMarker {
			return fmt.Errorf("signatureNotification value: unexpected marker %q", s)
		}
		v.ReceivedSignature = true
		return nil
	case '{':
		// Alias avoids recursion into this UnmarshalJSON.
		type alias struct {
			Err any `json:"err"`
		}
		var a alias
		if err := stdjson.Unmarshal(trimmed, &a); err != nil {
			return fmt.Errorf("signatureNotification value: %w", err)
		}
		v.Err = a.Err
		return nil
	default:
		return fmt.Errorf("signatureNotification value: unexpected JSON %s", string(trimmed))
	}
}

// SignatureSubscribeOpts mirrors the optional configuration object the
// signatureSubscribe RPC method accepts. See
// https://solana.com/docs/rpc/websocket/signaturesubscribe.
type SignatureSubscribeOpts struct {
	// Commitment selects the bank state the validator should use when
	// deciding the transaction has reached the requested level.
	Commitment rpc.CommitmentType

	// EnableReceivedNotification opts into the additional "received"
	// notification emitted as soon as the validator picks the
	// transaction up in its mempool (in addition to the final status
	// notification). Defaults to false to match the RPC default.
	EnableReceivedNotification bool
}

// SignatureSubscribe subscribes to a transaction signature to receive
// notification when the transaction is confirmed On signatureNotification,
// the subscription is automatically canceled
func (cl *Client) SignatureSubscribe(
	signature solana.Signature, // Transaction Signature.
	commitment rpc.CommitmentType, // (optional)
) (*SignatureSubscription, error) {
	return cl.SignatureSubscribeWithOpts(signature, &SignatureSubscribeOpts{
		Commitment: commitment,
	})
}

// SignatureSubscribeWithOpts subscribes to a transaction signature and
// forwards the full SignatureSubscribeOpts configuration object to the
// validator, including EnableReceivedNotification.
func (cl *Client) SignatureSubscribeWithOpts(
	signature solana.Signature,
	opts *SignatureSubscribeOpts,
) (*SignatureSubscription, error) {
	params := []any{signature.String()}
	conf := map[string]any{}
	if opts != nil {
		if opts.Commitment != "" {
			conf["commitment"] = opts.Commitment
		}
		if opts.EnableReceivedNotification {
			conf["enableReceivedNotification"] = true
		}
	}

	genSub, err := cl.subscribe(
		params,
		conf,
		"signatureSubscribe",
		"signatureUnsubscribe",
		func(msg []byte) (any, error) {
			var res SignatureResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &SignatureSubscription{
		sub: genSub,
	}, nil
}

type SignatureSubscription struct {
	sub *Subscription
}

func (sw *SignatureSubscription) Recv(ctx context.Context) (*SignatureResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-sw.sub.stream:
		if !ok {
			return nil, ErrSubscriptionClosed
		}
		return d.(*SignatureResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *SignatureSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *SignatureSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
