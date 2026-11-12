// Copyright 2026 github.com/gagliardetto
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

package confirmation

import (
	"context"
	"errors"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// RecentSignatureConfirmationFunc returns nil once the transaction with the
// given signature reaches commitment. It returns a *TransactionError if the
// transaction landed but failed. An empty commitment means finalized.
type RecentSignatureConfirmationFunc func(ctx context.Context, signature solana.Signature, commitment rpc.CommitmentType) error

// NewRecentSignatureConfirmation returns a RecentSignatureConfirmationFunc
// that subscribes to the signature, waits until the server confirms the
// subscription, and then checks the signature status once. The first of the
// two to report the transaction at commitment, or failed, decides.
//
// The status check only finds signatures in the status cache of the server,
// which covers the last 300 or so rooted slots. Kit:
// createRecentSignatureConfirmationPromiseFactory.
func NewRecentSignatureConfirmation(rpcClient *rpc.Client, wsClient *ws.Client) RecentSignatureConfirmationFunc {
	return func(ctx context.Context, signature solana.Signature, commitment rpc.CommitmentType) error {
		commitment, err := normalizeCommitment(commitment)
		if err != nil {
			return err
		}
		return contextError(ctx, recentSignatureConfirmation(ctx, rpcClient, wsClient, signature, commitment))
	}
}

func recentSignatureConfirmation(
	ctx context.Context,
	rpcClient *rpc.Client,
	wsClient *ws.Client,
	signature solana.Signature,
	commitment rpc.CommitmentType,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Step 1: subscribe to status changes of the signature.
	sub, err := wsClient.SignatureSubscribe(signature, commitment)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()
	if err := sub.WaitSubscribed(ctx); err != nil {
		return err
	}

	return firstOf(ctx,
		func(ctx context.Context) error {
			for {
				res, err := sub.Recv(ctx)
				if err != nil {
					return err
				}
				if res.ReceivedSignature {
					continue
				}
				if res.Value.Err != nil {
					return &TransactionError{Signature: signature, Err: res.Value.Err}
				}
				return nil
			}
		},
		// Step 2: with the subscription in place, check the current status
		// once. It is only found while the signature is in the status cache.
		func(ctx context.Context) error {
			out, err := rpcClient.GetSignatureStatuses(ctx, false, signature)
			if errors.Is(err, rpc.ErrNotFound) {
				return waitDone(ctx)
			}
			if err != nil {
				return err
			}
			if len(out.Value) == 0 || out.Value[0] == nil {
				return waitDone(ctx)
			}
			status := out.Value[0]
			// A failed transaction is reported at any commitment level, as in
			// kit (#1003).
			if status.Err != nil {
				return &TransactionError{Signature: signature, Err: status.Err}
			}
			if status.ConfirmationStatus != "" &&
				rpc.CompareCommitment(rpc.CommitmentType(status.ConfirmationStatus), commitment) >= 0 {
				return nil
			}
			return waitDone(ctx)
		},
	)
}
