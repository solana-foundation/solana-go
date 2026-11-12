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
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
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
//     blockhash. At processed, a switch to another fork can still bring it in;
//   - any other error, including ctx.Err(), if the outcome is unknown. Check
//     the signature status before signing the transaction again.
//
// A transaction can still land in the block at lastValidBlockHeight + 1, the
// block at which the expiry is reported. So before it reports the expiry,
// this function reads the signature status until it is conclusive. If the
// transaction landed below commitment, it waits until the transaction reaches
// commitment or a later bank misses it.
//
// The recent status cache of the node decides, which every RPC node has. Only
// a transaction that is sent again long after its blockhash expired, about 40
// seconds or more, may have left that cache. The function then searches the
// transaction history of the node too, and returns an error that wraps
// ErrTransactionHistoryNotAvailable if the node keeps none.
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
//     account does not exist;
//   - an error that wraps ErrTransactionHistoryNotAvailable if the nonce
//     advanced, and only the transaction history could tell whether this
//     transaction did it, but the node keeps none (see below);
//   - any other error, including ctx.Err(), if the outcome is unknown. Check
//     the signature status before signing the transaction again.
//
// A durable nonce transaction advances the nonce even when it fails. So before
// it reports an invalid nonce, this function reads the signature status until
// it is conclusive, as kit does since its fix of this race (#1234). If the
// transaction landed below commitment, it waits until the transaction reaches
// commitment or a later bank misses it.
//
// If the function saw the nonce still in the account less than about 100
// seconds before it changed, the recent status cache of the node decides,
// which every RPC node has. Otherwise the transaction may have landed long
// ago: the account already held another nonce when the wait began, as when a
// transaction is sent again after its result was lost, or the wait was long.
// The function then searches the transaction history of the node too, as far
// back as the node keeps it.
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

// ErrTransactionHistoryNotAvailable reports that the lifetime of a transaction
// ended, and that the RPC node cannot tell whether the transaction landed
// before that: its recent status cache does not know the transaction, the
// transaction may have landed before the window of that cache, and the node
// keeps no transaction history to search. An Agave node that runs without
// --enable-rpc-transaction-history answers such a search with the JSON-RPC
// error -32011, which the returned error wraps too.
//
// The outcome is unknown. Before signing the transaction again, check its
// signature on a node that keeps the transaction history, with
// GetSignatureStatuses and searchTransactionHistory set.
var ErrTransactionHistoryNotAvailable = errors.New("transaction history is not available")

const (
	// statusPollInterval is the pause between two reads of the signature
	// status after the lifetime of a transaction ended.
	statusPollInterval = 500 * time.Millisecond
	// maxUnusableStatuses is how many failed reads, or reads of a bank older
	// than needed, in a row end the checks with an unknown outcome.
	maxUnusableStatuses = 10
	// maxProcessingAge is how many blocks before its last valid block height
	// a transaction can land at the earliest (MAX_PROCESSING_AGE in Agave).
	maxProcessingAge = 150
	// statusCacheWindow is how many slots back a miss in the recent status
	// cache of a node is trusted. A node keeps a signature in that cache for
	// 300 rooted blocks after the transaction landed (MAX_RECENT_BLOCKHASHES
	// in Agave). The rest is a margin.
	statusCacheWindow = 250
	// rpcTransactionHistoryNotAvailable is the code of the JSON-RPC error of
	// a node that is asked to search a transaction history it does not keep
	// (JSON_RPC_SERVER_ERROR_TRANSACTION_HISTORY_NOT_AVAILABLE in Agave).
	rpcTransactionHistoryNotAvailable = -32011
)

// recheckSignature returns err, the result of a lifetime strategy, unless err
// reports that the lifetime ended (it wraps ended) while the transaction may
// have landed. A lifetime can end with the transaction itself: a durable nonce
// transaction advances the nonce even when it fails, and a blockhash
// transaction can land in the block at which its expiry is reported. So the
// signature status decides, as in kit's fix of the durable nonce race (#1234).
// Unlike kit, the status is read again until it is conclusive, and an unknown
// outcome is never reported as the end of the lifetime:
//
//   - At or above commitment: nil, or a *confirmation.TransactionError if the
//     transaction failed.
//   - Below commitment: the transaction landed, maybe on a fork that will be
//     dropped. The status is read again until it reaches commitment, or until
//     a later bank misses the transaction.
//   - Missing from a bank at least as recent as the end of the lifetime, and
//     later than any bank that showed the transaction: err, because the
//     transaction can no longer land. A miss in the recent status cache only
//     proves that while the cache must still hold the transaction if it
//     landed (see statusCacheCovers), which is the usual case. Otherwise the
//     transaction may have landed long ago, for example when it is sent again
//     after its result was lost, and the miss only counts once a search of
//     the transaction history misses the transaction too.
//   - The node keeps no transaction history for that search: an error that
//     wraps ErrTransactionHistoryNotAvailable but not ended.
//   - maxUnusableStatuses failed reads, or reads of older banks, in a row: an
//     error that wraps the last problem but not ended, because the outcome is
//     unknown.
//
// The signature strategy keeps running meanwhile, and the first of the two to
// decide wins.
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
	// A status that misses the transaction only counts if its bank is at
	// least this recent.
	minSlot := lifetimeEndSlot(err)
	// The recent status cache answers first. If it misses the transaction but
	// cannot prove that the transaction did not land, the reads that follow
	// search the transaction history too.
	searchHistory := false
	unusable := 0
	var problem error
	for {
		status, slot, statusErr := readSignatureStatus(ctx, rpcClient, signature, searchHistory)
		switch {
		case statusErr != nil:
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			if searchHistory && isHistoryNotAvailable(statusErr) {
				// The node will never answer this search, so do not retry.
				return fmt.Errorf("sendandconfirmtransaction: %w, so it is unknown whether transaction %s landed before its lifetime ended (%s): %w",
					ErrTransactionHistoryNotAvailable, signature, err.Error(), statusErr)
			}
			unusable, problem = unusable+1, statusErr
		case status != nil && reachedCommitment(status, commitment):
			if status.Err != nil {
				return &confirmation.TransactionError{Signature: signature, Err: status.Err}
			}
			return nil
		case status != nil:
			// Below commitment: only a later bank that misses the
			// transaction shows that its fork was dropped.
			unusable = 0
			minSlot = max(minSlot, slot+1)
		case slot < minSlot:
			unusable, problem = unusable+1, fmt.Errorf("the status is from slot %d, before slot %d", slot, minSlot)
		case searchHistory || statusCacheCovers(err, slot):
			return err
		default:
			// The recent status cache misses the transaction, but the
			// transaction may have landed before the window of that cache.
			// Search the history next.
			searchHistory = true
			continue
		}
		if unusable >= maxUnusableStatuses {
			return fmt.Errorf("sendandconfirmtransaction: the status of transaction %s is unknown after its lifetime ended (%s): %w",
				signature, err.Error(), problem)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(statusPollInterval):
		}
	}
}

// readSignatureStatus returns the status of signature, nil if the node does
// not know it, and the slot of the bank that the status was read from. With
// searchHistory, the node also searches its transaction history, which holds
// signatures that have left the recent status cache.
func readSignatureStatus(ctx context.Context, rpcClient *rpc.Client, signature solana.Signature, searchHistory bool) (*rpc.SignatureStatusesResult, uint64, error) {
	out, err := rpcClient.GetSignatureStatuses(ctx, searchHistory, signature)
	if err != nil {
		return nil, 0, err
	}
	if len(out.Value) != 1 {
		return nil, 0, fmt.Errorf("got %d signature statuses for 1 signature", len(out.Value))
	}
	return out.Value[0], out.Context.Slot, nil
}

// reachedCommitment reports whether status is at or above commitment.
func reachedCommitment(status *rpc.SignatureStatusesResult, commitment rpc.CommitmentType) bool {
	return status.ConfirmationStatus != "" &&
		rpc.CompareCommitment(rpc.CommitmentType(status.ConfirmationStatus), commitment) >= 0
}

// statusCacheCovers reports whether the recent status cache of the bank at
// slot must still hold the transaction if it landed before its lifetime
// ended with end. Only then does a miss in that cache prove that the
// transaction did not land. The cache goes back statusCacheWindow slots, so
// this needs a limit on how long ago the transaction can have landed:
//
//   - A blockhash transaction lands at most maxProcessingAge blocks before
//     its last valid block height. From there, count the blocks to the bank
//     that showed the expiry, and the slots on to the bank at slot.
//   - A durable nonce transaction lands after the last slot at which the
//     strategy saw its nonce in the account. There is no limit if the
//     strategy never saw it, as when the transaction is sent again long
//     after the nonce advanced.
func statusCacheCovers(end error, slot uint64) bool {
	var exceeded *confirmation.BlockHeightExceededError
	if errors.As(end, &exceeded) {
		past := distance(exceeded.LastValidBlockHeight, exceeded.CurrentBlockHeight)
		since := distance(exceeded.Slot, slot)
		// Check each part first, so that the sum cannot overflow.
		return past <= statusCacheWindow && since <= statusCacheWindow &&
			maxProcessingAge+past+since <= statusCacheWindow
	}
	var invalid *confirmation.InvalidNonceError
	if errors.As(end, &invalid) {
		return invalid.ValidSlot != 0 && distance(invalid.ValidSlot, slot) <= statusCacheWindow
	}
	return false
}

// distance returns to - from, or 0 if to is before from.
func distance(from, to uint64) uint64 {
	if to < from {
		return 0
	}
	return to - from
}

// isHistoryNotAvailable reports whether err is the JSON-RPC error of a node
// that keeps no transaction history.
func isHistoryNotAvailable(err error) bool {
	var rpcErr *jsonrpc.RPCError
	return errors.As(err, &rpcErr) && rpcErr.Code == rpcTransactionHistoryNotAvailable
}

// lifetimeEndSlot returns the slot at which a lifetime strategy saw the
// lifetime end, or 0 if err does not tell.
func lifetimeEndSlot(err error) uint64 {
	var exceeded *confirmation.BlockHeightExceededError
	if errors.As(err, &exceeded) {
		return exceeded.Slot
	}
	var invalid *confirmation.InvalidNonceError
	if errors.As(err, &invalid) {
		return invalid.Slot
	}
	return 0
}
