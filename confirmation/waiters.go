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
)

// RecentTransactionConfig configures WaitForRecentTransactionConfirmation.
type RecentTransactionConfig struct {
	// Commitment is the level to wait for. Empty means finalized.
	Commitment rpc.CommitmentType
	// Transaction is the signed transaction. Its first signature, the one of
	// the fee payer, identifies it.
	Transaction *solana.Transaction
	// LastValidBlockHeight is the last valid block height of the blockhash of
	// the transaction, as returned by getLatestBlockhash.
	LastValidBlockHeight uint64

	// RecentSignatureConfirmation waits for the signature, usually
	// NewRecentSignatureConfirmation(rpcClient, wsClient).
	RecentSignatureConfirmation RecentSignatureConfirmationFunc
	// BlockHeightExceedance reports the expiry of the blockhash, usually
	// NewBlockHeightExceedance(rpcClient, wsClient).
	BlockHeightExceedance BlockHeightExceedanceFunc
}

// DurableNonceTransactionConfig configures
// WaitForDurableNonceTransactionConfirmation.
type DurableNonceTransactionConfig struct {
	// Commitment is the level to wait for. Empty means finalized.
	Commitment rpc.CommitmentType
	// Transaction is the signed transaction. It must start with an
	// AdvanceNonceAccount instruction, and its recent blockhash is the nonce.
	Transaction *solana.Transaction

	// RecentSignatureConfirmation waits for the signature, usually
	// NewRecentSignatureConfirmation(rpcClient, wsClient).
	RecentSignatureConfirmation RecentSignatureConfirmationFunc
	// NonceInvalidation reports that the nonce advanced, usually
	// NewNonceInvalidation(rpcClient, wsClient).
	NonceInvalidation NonceInvalidationFunc
}

// RecentTransactionUntilTimeoutConfig configures
// WaitForRecentTransactionConfirmationUntilTimeout.
type RecentTransactionUntilTimeoutConfig struct {
	// Commitment is the level to wait for. Empty means finalized.
	Commitment rpc.CommitmentType
	// Signature identifies the transaction.
	Signature solana.Signature

	// RecentSignatureConfirmation waits for the signature, usually
	// NewRecentSignatureConfirmation(rpcClient, wsClient).
	RecentSignatureConfirmation RecentSignatureConfirmationFunc
	// Timeout ends the wait, usually Timeout.
	Timeout TimeoutFunc
}

var errNilStrategy = errors.New("confirmation: every strategy func of the config must be set")

// WaitForRecentTransactionConfirmation waits until the transaction reaches
// cfg.Commitment, or until its blockhash expires. Kit:
// waitForRecentTransactionConfirmation.
//
// It returns nil once the transaction reached the commitment, a
// *TransactionError if it landed but failed, and a *BlockHeightExceededError
// if the blockhash expired first.
//
// A transaction can still land in the block at height
// cfg.LastValidBlockHeight + 1, the block at which the expiry is reported.
// If the signature notification of that block loses the race, the expiry is
// reported although the transaction landed. Check the signature status before
// signing the transaction again with a new blockhash, or use
// sendandconfirmtransaction.SendAndConfirmTransactionWithBlockhashLifetime,
// which does that.
func WaitForRecentTransactionConfirmation(ctx context.Context, cfg RecentTransactionConfig) error {
	signature, err := feePayerSignature(cfg.Transaction)
	if err != nil {
		return err
	}
	commitment, err := normalizeCommitment(cfg.Commitment)
	if err != nil {
		return err
	}
	if cfg.RecentSignatureConfirmation == nil || cfg.BlockHeightExceedance == nil {
		return errNilStrategy
	}
	return raceStrategies(ctx,
		func(ctx context.Context) error {
			return cfg.RecentSignatureConfirmation(ctx, signature, commitment)
		},
		func(ctx context.Context) error {
			return cfg.BlockHeightExceedance(ctx, cfg.LastValidBlockHeight, commitment)
		},
	)
}

// WaitForDurableNonceTransactionConfirmation waits until the transaction
// reaches cfg.Commitment, or until its nonce account holds another nonce.
// Kit: waitForDurableNonceTransactionConfirmation.
//
// It returns nil once the transaction reached the commitment, a
// *TransactionError if it landed but failed, an *InvalidNonceError if the
// nonce advanced first, and a *NonceAccountNotFoundError if the nonce account
// does not exist.
//
// A failed durable nonce transaction still advances the nonce. If the account
// notification arrives before the signature notification, the invalid nonce
// is reported although the transaction landed (kit #1234). Check the
// signature status after an *InvalidNonceError, or use
// sendandconfirmtransaction.SendAndConfirmDurableNonceTransaction, which does
// that.
func WaitForDurableNonceTransactionConfirmation(ctx context.Context, cfg DurableNonceTransactionConfig) error {
	signature, err := feePayerSignature(cfg.Transaction)
	if err != nil {
		return err
	}
	nonceAccount, ok := cfg.Transaction.GetNonceAccount()
	if !ok {
		return ErrNotDurableNonceTransaction
	}
	commitment, err := normalizeCommitment(cfg.Commitment)
	if err != nil {
		return err
	}
	if cfg.RecentSignatureConfirmation == nil || cfg.NonceInvalidation == nil {
		return errNilStrategy
	}
	nonce := cfg.Transaction.Message.RecentBlockhash
	return raceStrategies(ctx,
		func(ctx context.Context) error {
			return cfg.RecentSignatureConfirmation(ctx, signature, commitment)
		},
		func(ctx context.Context) error {
			return cfg.NonceInvalidation(ctx, nonceAccount, nonce, commitment)
		},
	)
}

// WaitForRecentTransactionConfirmationUntilTimeout waits until the
// transaction with cfg.Signature reaches cfg.Commitment, or until cfg.Timeout
// returns. Kit: waitForRecentTransactionConfirmationUntilTimeout.
//
// A timeout cannot tell whether the transaction can still land. Prefer the
// waiters that follow the lifetime of the transaction, and use this one only
// when the lifetime is unknown, as for an airdrop.
func WaitForRecentTransactionConfirmationUntilTimeout(ctx context.Context, cfg RecentTransactionUntilTimeoutConfig) error {
	if cfg.Signature.IsZero() {
		return ErrFeePayerSignatureMissing
	}
	commitment, err := normalizeCommitment(cfg.Commitment)
	if err != nil {
		return err
	}
	if cfg.RecentSignatureConfirmation == nil || cfg.Timeout == nil {
		return errNilStrategy
	}
	return raceStrategies(ctx,
		func(ctx context.Context) error {
			return cfg.RecentSignatureConfirmation(ctx, cfg.Signature, commitment)
		},
		func(ctx context.Context) error {
			return cfg.Timeout(ctx, commitment)
		},
	)
}

// feePayerSignature returns the first signature of tx, the one of the fee
// payer. Kit: getSignatureFromTransaction.
func feePayerSignature(tx *solana.Transaction) (solana.Signature, error) {
	if tx == nil {
		return solana.Signature{}, errors.New("confirmation: transaction is nil")
	}
	if len(tx.Signatures) == 0 || tx.Signatures[0].IsZero() {
		return solana.Signature{}, ErrFeePayerSignatureMissing
	}
	return tx.Signatures[0], nil
}
