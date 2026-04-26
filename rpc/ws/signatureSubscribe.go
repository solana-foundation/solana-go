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
	stdjson "encoding/json"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// SignatureSubscribeConfig matches Agave's RpcSignatureSubscribeConfig.
// When EnableReceivedNotification is true, the subscription also emits
// a "receivedSignature" notification the moment the transaction is first
// received by the node — before confirmation. Callers must check
// SignatureResult.Value.IsReceived / .Processed to disambiguate.
type SignatureSubscribeConfig struct {
	Commitment                 rpc.CommitmentType
	EnableReceivedNotification *bool
}

// params converts the config to the JSON-RPC params object the
// signatureSubscribe method expects. Missing options are omitted so the
// wire format matches Agave's serde(skip_serializing_if) behavior.
func (c *SignatureSubscribeConfig) params() map[string]any {
	conf := make(map[string]any)
	if c == nil {
		return conf
	}
	if c.Commitment != "" {
		conf["commitment"] = c.Commitment
	}
	if c.EnableReceivedNotification != nil {
		conf["enableReceivedNotification"] = *c.EnableReceivedNotification
	}
	return conf
}

// SignatureResult is the notification payload for signatureSubscribe.
// The shape of .Value depends on which subscription mode is active:
//   - standard: {"err": ...}  → Value.Processed is set, IsReceived is false.
//   - received (EnableReceivedNotification=true): the raw string
//     "receivedSignature" → IsReceived is true, Processed is nil.
type SignatureResult struct {
	Context RPCResponseContext   `json:"context"`
	Value   SignatureResultValue `json:"value"`
}

// RPCResponseContext mirrors Agave's RpcResponseContext.
type RPCResponseContext struct {
	Slot       uint64  `json:"slot"`
	APIVersion *string `json:"apiVersion,omitempty"`
}

// SignatureResultValue is a Go-friendly projection of Agave's untagged
// RpcSignatureResult enum (ProcessedSignature | ReceivedSignature).
type SignatureResultValue struct {
	// IsReceived is true when the notification signals that the tx has
	// been received by the node (pre-confirmation). Only populated when
	// the subscription was created with EnableReceivedNotification=true.
	IsReceived bool
	// Processed is non-nil for the standard (post-processing) notification.
	Processed *ProcessedSignatureResult
}

// ProcessedSignatureResult matches Agave's ProcessedSignatureResult.
// Err is nil on success; on failure it mirrors the RPC JSON (typically
// a map or a string identifying the TransactionError).
type ProcessedSignatureResult struct {
	Err any `json:"err"`
}

// receivedSignatureTag is the single string variant emitted by
// Agave's ReceivedSignatureResult enum.
const receivedSignatureTag = "receivedSignature"

// UnmarshalJSON decodes either the object form ({"err":...}) or the
// string form ("receivedSignature") returned by Solana.
func (v *SignatureResultValue) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	switch data[0] {
	case '"':
		var s string
		if err := stdjson.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("SignatureResultValue: %w", err)
		}
		if s != receivedSignatureTag {
			return fmt.Errorf("SignatureResultValue: unknown string variant %q", s)
		}
		v.IsReceived = true
		return nil
	case '{':
		var p ProcessedSignatureResult
		if err := stdjson.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("SignatureResultValue: %w", err)
		}
		v.Processed = &p
		return nil
	}
	return fmt.Errorf("SignatureResultValue: unexpected JSON %s", string(data))
}

// SignatureSubscribe subscribes to confirmation notifications for a
// single transaction signature. The subscription is canceled by the
// server once the notification fires.
func (cl *Client) SignatureSubscribe(
	signature solana.Signature,
	commitment rpc.CommitmentType, // optional
) (*SignatureSubscription, error) {
	return cl.SignatureSubscribeWithOpts(signature, &SignatureSubscribeConfig{Commitment: commitment})
}

// SignatureSubscribeWithOpts mirrors signatureSubscribe + full config.
// Pass EnableReceivedNotification=&true to also receive a
// "receivedSignature" event as soon as the node sees the tx.
func (cl *Client) SignatureSubscribeWithOpts(
	signature solana.Signature,
	config *SignatureSubscribeConfig,
) (*SignatureSubscription, error) {
	genSub, err := cl.subscribe(
		[]any{signature.String()},
		config.params(),
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
	return &SignatureSubscription{sub: genSub}, nil
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
