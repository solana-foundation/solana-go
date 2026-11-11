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

// Package confirmation waits until a transaction reaches a commitment level,
// or until its lifetime ends. It is a port of @solana/transaction-confirmation
// from the Solana Kit (https://github.com/anza-xyz/kit).
//
// A transaction can only land while its lifetime lasts:
//
//   - A transaction that uses a recent blockhash expires once the block height
//     passes its last valid block height, as returned by getLatestBlockhash.
//   - A transaction that uses a durable nonce stays valid until the nonce
//     account holds a different nonce.
//
// The waiters race a signature confirmation against the lifetime of the
// transaction, so they end with a definite answer instead of a guess based on
// time:
//
//	sig := tx.Signatures[0]
//	err := confirmation.WaitForRecentTransactionConfirmation(ctx, confirmation.RecentTransactionConfig{
//		Commitment:                  rpc.CommitmentConfirmed,
//		Transaction:                 tx,
//		LastValidBlockHeight:        latest.Value.LastValidBlockHeight,
//		RecentSignatureConfirmation: confirmation.NewRecentSignatureConfirmation(rpcClient, wsClient),
//		BlockHeightExceedance:       confirmation.NewBlockHeightExceedance(rpcClient, wsClient),
//	})
//	var txErr *confirmation.TransactionError
//	switch {
//	case err == nil:
//		// sig reached the confirmed commitment level.
//	case errors.Is(err, confirmation.ErrBlockHeightExceeded):
//		// The blockhash expired. See WaitForRecentTransactionConfirmation
//		// before signing the transaction again.
//	case errors.As(err, &txErr):
//		// The transaction landed, but it failed.
//	}
//
// Most programs call the helpers of the rpc/sendAndConfirmTransaction package
// instead, which send the transaction and then wait with these strategies.
//
// # Strategies
//
// Each strategy follows the same steps as in kit. It subscribes over
// WebSocket, waits until the server confirms the subscription, and then checks
// the current state once over RPC, so that no change can fall between the
// check and the subscription:
//
//   - NewRecentSignatureConfirmation: signatureSubscribe, then
//     getSignatureStatuses.
//   - NewBlockHeightExceedance: slotSubscribe and getEpochInfo. The block
//     height is estimated from slot notifications and checked again before
//     it reports the expiry.
//   - NewNonceInvalidation: accountSubscribe on the nonce account, then
//     getAccountInfo.
//   - Timeout: 30 seconds for the processed commitment, 60 seconds otherwise.
//
// One ws.Client can serve any number of confirmations at the same time.
//
// # Errors
//
// The errors map to the SolanaError codes of kit and work with errors.Is and
// errors.As:
//
//   - ErrBlockHeightExceeded, as *BlockHeightExceededError
//   - ErrInvalidNonce, as *InvalidNonceError
//   - ErrNonceAccountNotFound, as *NonceAccountNotFoundError
//   - ErrTransactionFailed, as *TransactionError: the transaction landed,
//     but it failed
//   - ErrFeePayerSignatureMissing
//
// When ctx is canceled, every function returns ctx.Err(). Returning can take
// up to 10 seconds when the WebSocket connection stalls, because the
// unsubscribe requests are written without a context.
package confirmation
