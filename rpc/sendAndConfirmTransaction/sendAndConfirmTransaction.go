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

package sendandconfirmtransaction

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/confirmation"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// Send and wait for confirmation of a transaction.
//
// See also SendAndConfirmTransactionWithBlockhashLifetime and
// SendAndConfirmDurableNonceTransaction, which wait at any commitment level
// and stop when the transaction can no longer land.
func SendAndConfirmTransaction(
	ctx context.Context,
	rpcClient *rpc.Client,
	wsClient *ws.Client,
	transaction *solana.Transaction,
) (signature solana.Signature, err error) {
	opts := rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentFinalized,
	}

	return SendAndConfirmTransactionWithOpts(
		ctx,
		rpcClient,
		wsClient,
		transaction,
		opts,
		nil,
	)
}

func SendAndConfirmTransactionWithTimeout(
	ctx context.Context,
	rpcClient *rpc.Client,
	wsClient *ws.Client,
	transaction *solana.Transaction,
	timeout time.Duration,
) (signature solana.Signature, err error) {
	opts := rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentFinalized,
	}

	return SendAndConfirmTransactionWithOpts(
		ctx,
		rpcClient,
		wsClient,
		transaction,
		opts,
		&timeout,
	)
}

var ErrTimeout = fmt.Errorf("timeout")

// Send and wait for confirmation of a transaction.
func SendAndConfirmTransactionWithOpts(
	ctx context.Context,
	rpcClient *rpc.Client,
	wsClient *ws.Client,
	transaction *solana.Transaction,
	opts rpc.TransactionOpts,
	timeout *time.Duration,
) (sig solana.Signature, err error) {
	sig, err = rpcClient.SendTransactionWithOpts(
		ctx,
		transaction,
		opts,
	)
	if err != nil {
		return sig, err
	}
	_, err = WaitForConfirmation(
		ctx,
		wsClient,
		sig,
		timeout,
	)
	return sig, err
}

// WaitForConfirmation waits for a transaction to be confirmed.
// If the transaction was confirmed, but it failed while executing (one of the instructions failed),
// then this function will return an error (true, error).
// If the transaction was confirmed, and it succeeded, then this function will return nil (true, nil).
//
// See also the waiters of the confirmation package, which stop when the
// transaction can no longer land.
func WaitForConfirmation(
	ctx context.Context,
	wsClient *ws.Client,
	sig solana.Signature,
	timeout *time.Duration,
) (confirmed bool, err error) {
	sub, err := wsClient.SignatureSubscribe(
		sig,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return false, err
	}
	defer sub.Unsubscribe()

	if timeout == nil {
		t := 2 * time.Minute // random default timeout
		timeout = &t
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	got, err := sub.Recv(timeoutCtx)
	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return false, ErrTimeout
		}
		return false, err
	}
	if got.Value.Err != nil {
		return true, fmt.Errorf("confirmed transaction with execution error: %v", got.Value.Err)
	}
	return true, nil
}

// SendAndConfirmTransactionWithBlockhashLifetime sends a transaction that
// uses a recent blockhash and waits until it reaches commitment, or until the
// blockhash expires. Kit: sendAndConfirmTransactionFactory.
//
// lastValidBlockHeight is the value getLatestBlockhash returned with the
// blockhash of the transaction. An empty commitment means finalized. If
// opts.PreflightCommitment is empty and commitment is below finalized, the
// transaction is simulated at commitment, as in kit. The transaction is always
// sent base64 encoded.
//
// It returns the signature of the transaction, also when waiting fails, and:
//
//   - nil once the transaction reached commitment;
//   - a *confirmation.TransactionError if it landed but failed;
//   - an error that wraps confirmation.ErrBlockHeightExceeded if the blockhash
//     expired first. At the confirmed or finalized commitment, the transaction
//     can then no longer land, so it is safe to sign it again with a new
//     blockhash. At processed, a switch to another fork can still bring it in.
//
// A transaction can still land in the block at lastValidBlockHeight + 1, the
// block at which the expiry is reported. So before it reports the expiry,
// this function checks the signature status once more.
func SendAndConfirmTransactionWithBlockhashLifetime(
	ctx context.Context,
	rpcClient *rpc.Client,
	wsClient *ws.Client,
	transaction *solana.Transaction,
	lastValidBlockHeight uint64,
	commitment rpc.CommitmentType,
	opts rpc.TransactionOpts,
) (solana.Signature, error) {
	return sendAndConfirmWithBlockhashLifetime(ctx, rpcClient, transaction, lastValidBlockHeight, commitment, opts,
		confirmation.NewRecentSignatureConfirmation(rpcClient, wsClient),
		confirmation.NewBlockHeightExceedance(rpcClient, wsClient),
	)
}

// SendAndConfirmDurableNonceTransaction sends a transaction that uses a
// durable nonce and waits until it reaches commitment, or until the nonce
// account holds another nonce. Kit: sendAndConfirmDurableNonceTransactionFactory.
//
// The transaction must start with an AdvanceNonceAccount instruction, and its
// recent blockhash is the nonce. An empty commitment means finalized. If
// opts.PreflightCommitment is empty and commitment is below finalized, the
// transaction is simulated at commitment, as in kit. The transaction is always
// sent base64 encoded.
//
// It returns the signature of the transaction, also when waiting fails, and:
//
//   - nil once the transaction reached commitment;
//   - a *confirmation.TransactionError if it landed but failed;
//   - an error that wraps confirmation.ErrInvalidNonce if the nonce advanced
//     without the transaction;
//   - an error that wraps confirmation.ErrNonceAccountNotFound if the nonce
//     account does not exist.
//
// A durable nonce transaction advances the nonce even when it fails. So before
// it reports an invalid nonce, this function checks the signature status, as
// kit does since its fix of this race (#1234).
func SendAndConfirmDurableNonceTransaction(
	ctx context.Context,
	rpcClient *rpc.Client,
	wsClient *ws.Client,
	transaction *solana.Transaction,
	commitment rpc.CommitmentType,
	opts rpc.TransactionOpts,
) (solana.Signature, error) {
	return sendAndConfirmDurableNonce(ctx, rpcClient, transaction, commitment, opts,
		confirmation.NewRecentSignatureConfirmation(rpcClient, wsClient),
		confirmation.NewNonceInvalidation(rpcClient, wsClient),
	)
}

// RequestAndConfirmAirdrop requests an airdrop of lamports to recipient and
// waits until the airdrop transaction reaches commitment. Airdrops only work
// on test clusters. Kit: airdropFactory.
//
// The airdrop transaction is built by the server, so its lifetime is unknown:
// the wait ends with an error that wraps ErrTimeout and
// confirmation.ErrTimeout after confirmation.Timeout, 30 seconds for the
// processed commitment and 60 seconds otherwise. An empty commitment means
// finalized. It returns the signature of the airdrop transaction, also when
// waiting fails.
func RequestAndConfirmAirdrop(
	ctx context.Context,
	rpcClient *rpc.Client,
	wsClient *ws.Client,
	recipient solana.PublicKey,
	lamports uint64,
	commitment rpc.CommitmentType,
) (solana.Signature, error) {
	return requestAndConfirmAirdrop(ctx, rpcClient, recipient, lamports, commitment,
		confirmation.NewRecentSignatureConfirmation(rpcClient, wsClient),
		confirmation.Timeout,
	)
}

// The unexported cores take the strategies, like the
// _INTERNAL_ONLY_DO_NOT_EXPORT functions of kit, so tests can replace them.

func sendAndConfirmWithBlockhashLifetime(
	ctx context.Context,
	rpcClient *rpc.Client,
	transaction *solana.Transaction,
	lastValidBlockHeight uint64,
	commitment rpc.CommitmentType,
	opts rpc.TransactionOpts,
	signatureConfirmation confirmation.RecentSignatureConfirmationFunc,
	blockHeightExceedance confirmation.BlockHeightExceedanceFunc,
) (solana.Signature, error) {
	commitment, err := prepareSend(transaction, commitment)
	if err != nil {
		return solana.Signature{}, err
	}
	signature := transaction.Signatures[0]
	if _, err := send(ctx, rpcClient, transaction, commitment, opts); err != nil {
		return signature, err
	}
	err = confirmation.WaitForRecentTransactionConfirmation(ctx, confirmation.RecentTransactionConfig{
		Commitment:                  commitment,
		Transaction:                 transaction,
		LastValidBlockHeight:        lastValidBlockHeight,
		RecentSignatureConfirmation: signatureConfirmation,
		BlockHeightExceedance: func(ctx context.Context, lastValidBlockHeight uint64, commitment rpc.CommitmentType) error {
			err := blockHeightExceedance(ctx, lastValidBlockHeight, commitment)
			return recheckSignature(ctx, rpcClient, signature, commitment, confirmation.ErrBlockHeightExceeded, err)
		},
	})
	return signature, err
}

func sendAndConfirmDurableNonce(
	ctx context.Context,
	rpcClient *rpc.Client,
	transaction *solana.Transaction,
	commitment rpc.CommitmentType,
	opts rpc.TransactionOpts,
	signatureConfirmation confirmation.RecentSignatureConfirmationFunc,
	nonceInvalidation confirmation.NonceInvalidationFunc,
) (solana.Signature, error) {
	commitment, err := prepareSend(transaction, commitment)
	if err != nil {
		return solana.Signature{}, err
	}
	if _, ok := transaction.GetNonceAccount(); !ok {
		return solana.Signature{}, confirmation.ErrNotDurableNonceTransaction
	}
	signature := transaction.Signatures[0]
	if _, err := send(ctx, rpcClient, transaction, commitment, opts); err != nil {
		return signature, err
	}
	err = confirmation.WaitForDurableNonceTransactionConfirmation(ctx, confirmation.DurableNonceTransactionConfig{
		Commitment:                  commitment,
		Transaction:                 transaction,
		RecentSignatureConfirmation: signatureConfirmation,
		NonceInvalidation: func(ctx context.Context, nonceAccount solana.PublicKey, currentNonce solana.Hash, commitment rpc.CommitmentType) error {
			err := nonceInvalidation(ctx, nonceAccount, currentNonce, commitment)
			return recheckSignature(ctx, rpcClient, signature, commitment, confirmation.ErrInvalidNonce, err)
		},
	})
	return signature, err
}

func requestAndConfirmAirdrop(
	ctx context.Context,
	rpcClient *rpc.Client,
	recipient solana.PublicKey,
	lamports uint64,
	commitment rpc.CommitmentType,
	signatureConfirmation confirmation.RecentSignatureConfirmationFunc,
	timeout confirmation.TimeoutFunc,
) (solana.Signature, error) {
	commitment, err := normalizeCommitment(commitment)
	if err != nil {
		return solana.Signature{}, err
	}
	signature, err := rpcClient.RequestAirdrop(ctx, recipient, lamports, commitment)
	if err != nil {
		return signature, err
	}
	err = confirmation.WaitForRecentTransactionConfirmationUntilTimeout(ctx, confirmation.RecentTransactionUntilTimeoutConfig{
		Commitment:                  commitment,
		Signature:                   signature,
		RecentSignatureConfirmation: signatureConfirmation,
		Timeout:                     timeout,
	})
	if errors.Is(err, confirmation.ErrTimeout) {
		err = fmt.Errorf("%w: %w", ErrTimeout, err)
	}
	return signature, err
}

// prepareSend checks what can be checked before sending: the transaction must
// be signed by its fee payer, and the commitment must be known. It returns the
// commitment to wait for.
func prepareSend(transaction *solana.Transaction, commitment rpc.CommitmentType) (rpc.CommitmentType, error) {
	if transaction == nil {
		return "", errors.New("sendandconfirmtransaction: transaction is nil")
	}
	if len(transaction.Signatures) == 0 || transaction.Signatures[0].IsZero() {
		return "", confirmation.ErrFeePayerSignatureMissing
	}
	return normalizeCommitment(commitment)
}

// normalizeCommitment returns the commitment to wait for. Empty means
// finalized, the default of the RPC server. It follows the same rules as the
// confirmation package, which validates the commitment again, so keep the two
// in sync.
func normalizeCommitment(commitment rpc.CommitmentType) (rpc.CommitmentType, error) {
	switch commitment {
	case "":
		return rpc.CommitmentFinalized, nil
	case rpc.CommitmentProcessed, rpc.CommitmentConfirmed, rpc.CommitmentFinalized:
		return commitment, nil
	default:
		return "", fmt.Errorf("sendandconfirmtransaction: unsupported commitment %q", commitment)
	}
}

// send sends the transaction base64 encoded. When the caller chose no
// preflight commitment and waits for less than finalized, the default of the
// server, the simulation runs at the commitment that is waited for. Kit:
// getSendTransactionConfigWithAdjustedPreflightCommitment.
func send(
	ctx context.Context,
	rpcClient *rpc.Client,
	transaction *solana.Transaction,
	commitment rpc.CommitmentType,
	opts rpc.TransactionOpts,
) (solana.Signature, error) {
	opts.Encoding = solana.EncodingBase64
	if opts.PreflightCommitment == "" && rpc.CompareCommitment(commitment, rpc.CommitmentFinalized) < 0 {
		opts.PreflightCommitment = commitment
	}
	return rpcClient.SendTransactionWithOpts(ctx, transaction, opts)
}

// recheckSignature returns err, the result of a lifetime strategy, unless err
// reports that the lifetime ended (it wraps ended) while the transaction
// landed. A lifetime can end with the transaction itself: a durable nonce
// transaction advances the nonce even when it fails, and a blockhash
// transaction can land in the block at which its expiry is reported. The
// signature status then decides, as in kit's fix of the durable nonce race
// (#1234):
//
//   - not found, or the request failed: err;
//   - at or above commitment: nil, or a *confirmation.TransactionError if the
//     transaction failed;
//   - below commitment: block until ctx is done, so that the signature
//     strategy decides.
func recheckSignature(
	ctx context.Context,
	rpcClient *rpc.Client,
	signature solana.Signature,
	commitment rpc.CommitmentType,
	ended error,
	err error,
) error {
	if !errors.Is(err, ended) {
		return err
	}
	out, statusErr := rpcClient.GetSignatureStatuses(ctx, false, signature)
	if statusErr != nil || len(out.Value) == 0 || out.Value[0] == nil {
		return err
	}
	status := out.Value[0]
	if status.ConfirmationStatus != "" &&
		rpc.CompareCommitment(rpc.CommitmentType(status.ConfirmationStatus), commitment) >= 0 {
		if status.Err != nil {
			return &confirmation.TransactionError{Signature: signature, Err: status.Err}
		}
		return nil
	}
	<-ctx.Done()
	return ctx.Err()
}
