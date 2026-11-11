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

// NonceInvalidationFunc returns an *InvalidNonceError once nonceAccount holds
// a nonce other than currentNonce at commitment, and a
// *NonceAccountNotFoundError if the account does not exist. Until then it
// blocks. An empty commitment means finalized.
type NonceInvalidationFunc func(ctx context.Context, nonceAccount solana.PublicKey, currentNonce solana.Hash, commitment rpc.CommitmentType) error

// The durable nonce follows the version (u32), the state (u32) and the
// authority (32 bytes) in the account data. Kit: NONCE_VALUE_OFFSET.
const (
	nonceValueOffset = 4 + 4 + 32
	nonceValueLength = 32
)

// NewNonceInvalidation returns a NonceInvalidationFunc that subscribes to the
// nonce account, waits until the server confirms the subscription, and then
// reads the nonce once. Kit: createNonceInvalidationPromiseFactory.
//
// A failed durable nonce transaction still advances the nonce. To tell that
// case from a nonce advanced by another transaction, check the signature
// status when this function reports an invalid nonce, as
// sendandconfirmtransaction.SendAndConfirmDurableNonceTransaction does.
func NewNonceInvalidation(rpcClient *rpc.Client, wsClient *ws.Client) NonceInvalidationFunc {
	return func(ctx context.Context, nonceAccount solana.PublicKey, currentNonce solana.Hash, commitment rpc.CommitmentType) error {
		commitment, err := normalizeCommitment(commitment)
		if err != nil {
			return err
		}
		return contextError(ctx, nonceInvalidation(ctx, rpcClient, wsClient, nonceAccount, currentNonce, commitment))
	}
}

func nonceInvalidation(
	ctx context.Context,
	rpcClient *rpc.Client,
	wsClient *ws.Client,
	nonceAccount solana.PublicKey,
	expected solana.Hash,
	commitment rpc.CommitmentType,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Step 1: subscribe to changes of the nonce account.
	sub, err := wsClient.AccountSubscribeWithOpts(nonceAccount, commitment, solana.EncodingBase64)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()
	if err := sub.WaitSubscribed(ctx); err != nil {
		return err
	}

	invalid := func(actual solana.Hash) error {
		return &InvalidNonceError{NonceAccount: nonceAccount, Expected: expected, Actual: actual}
	}
	return firstOf(ctx,
		func(ctx context.Context) error {
			for {
				res, err := sub.Recv(ctx)
				if err != nil {
					return err
				}
				var data []byte
				if res.Value != nil {
					data = res.Value.Data.GetBinary()
				}
				if len(data) < nonceValueOffset+nonceValueLength {
					return invalid(solana.Hash{})
				}
				if actual := solana.HashFromBytes(data[nonceValueOffset : nonceValueOffset+nonceValueLength]); actual != expected {
					return invalid(actual)
				}
			}
		},
		// Step 2: with the subscription in place, read the nonce once to
		// find out whether it advanced already.
		func(ctx context.Context) error {
			offset, length := uint64(nonceValueOffset), uint64(nonceValueLength)
			out, err := rpcClient.GetAccountInfoWithOpts(ctx, nonceAccount, &rpc.GetAccountInfoOpts{
				Commitment: commitment,
				Encoding:   solana.EncodingBase64,
				DataSlice:  &rpc.DataSlice{Offset: &offset, Length: &length},
			})
			if errors.Is(err, rpc.ErrNotFound) {
				return &NonceAccountNotFoundError{NonceAccount: nonceAccount}
			}
			if err != nil {
				return err
			}
			data := out.GetBinary()
			if len(data) < nonceValueLength {
				return invalid(solana.Hash{})
			}
			if actual := solana.HashFromBytes(data[:nonceValueLength]); actual != expected {
				return invalid(actual)
			}
			return waitDone(ctx)
		},
	)
}
