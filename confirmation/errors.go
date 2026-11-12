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
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
	stdjson "github.com/goccy/go-json"
)

var (
	// ErrBlockHeightExceeded reports that the block height passed the last
	// valid block height of the transaction's blockhash. The error returned
	// is a *BlockHeightExceededError. Kit: SOLANA_ERROR__BLOCK_HEIGHT_EXCEEDED.
	ErrBlockHeightExceeded = errors.New("block height exceeded")

	// ErrInvalidNonce reports that the nonce account holds a different nonce
	// than the transaction. The error returned is an *InvalidNonceError.
	// Kit: SOLANA_ERROR__INVALID_NONCE.
	ErrInvalidNonce = errors.New("invalid nonce")

	// ErrNonceAccountNotFound reports that the nonce account of the
	// transaction does not exist. The error returned is a
	// *NonceAccountNotFoundError. Kit: SOLANA_ERROR__NONCE_ACCOUNT_NOT_FOUND.
	ErrNonceAccountNotFound = errors.New("nonce account not found")

	// ErrTransactionFailed reports that the transaction landed, but it
	// failed. The error returned is a *TransactionError, which holds the
	// reason. Kit: the SOLANA_ERROR__TRANSACTION_ERROR__* codes.
	ErrTransactionFailed = errors.New("transaction failed")

	// ErrFeePayerSignatureMissing reports that the transaction has no
	// signature of its fee payer, so there is no signature to wait for.
	// Kit: SOLANA_ERROR__TRANSACTION__FEE_PAYER_SIGNATURE_MISSING.
	ErrFeePayerSignatureMissing = errors.New("transaction is not signed by its fee payer")

	// ErrNotDurableNonceTransaction reports that a transaction passed to a
	// durable nonce waiter does not start with an AdvanceNonceAccount
	// instruction.
	ErrNotDurableNonceTransaction = errors.New("transaction does not use a durable nonce")

	// ErrTimeout reports that Timeout elapsed before the transaction was
	// confirmed.
	ErrTimeout = errors.New("confirmation timed out")
)

// BlockHeightExceededError reports that the block height passed the last
// valid block height of the transaction's blockhash. No later block can
// include the transaction.
type BlockHeightExceededError struct {
	// CurrentBlockHeight is the block height at the commitment level that
	// was waited for.
	CurrentBlockHeight uint64
	// LastValidBlockHeight is the last valid block height of the blockhash.
	LastValidBlockHeight uint64
	// Slot is the slot of the bank that showed CurrentBlockHeight. A later
	// check of the signature status is only conclusive if it reads a bank at
	// least this recent. It is 0 when unknown.
	Slot uint64
}

func (e *BlockHeightExceededError) Error() string {
	return fmt.Sprintf("block height exceeded: block height %d is past the last valid block height %d",
		e.CurrentBlockHeight, e.LastValidBlockHeight)
}

// Unwrap returns ErrBlockHeightExceeded.
func (e *BlockHeightExceededError) Unwrap() error { return ErrBlockHeightExceeded }

// InvalidNonceError reports that the nonce account holds a different nonce
// than the transaction, so the transaction can no longer land.
type InvalidNonceError struct {
	NonceAccount solana.PublicKey
	// Expected is the nonce of the transaction.
	Expected solana.Hash
	// Actual is the nonce the account holds now. It is zero when the account
	// data is too short to hold a nonce, for example after the account was
	// closed.
	Actual solana.Hash
	// Slot is the slot at which the account held Actual. A later check of the
	// signature status is only conclusive if it reads a bank at least this
	// recent. It is 0 when unknown.
	Slot uint64
	// ValidSlot is the most recent slot at which the strategy saw the account
	// still hold Expected, or 0 if it never did. The nonce changed after that
	// slot, so a transaction that used it landed after that slot, if it
	// landed. That tells a check of the signature status how far back it has
	// to look.
	ValidSlot uint64
}

func (e *InvalidNonceError) Error() string {
	return fmt.Sprintf("invalid nonce: nonce account %s holds %s, the transaction uses %s",
		e.NonceAccount, e.Actual, e.Expected)
}

// Unwrap returns ErrInvalidNonce.
func (e *InvalidNonceError) Unwrap() error { return ErrInvalidNonce }

// NonceAccountNotFoundError reports that the nonce account of the transaction
// does not exist.
type NonceAccountNotFoundError struct {
	NonceAccount solana.PublicKey
}

func (e *NonceAccountNotFoundError) Error() string {
	return fmt.Sprintf("nonce account not found: %s", e.NonceAccount)
}

// Unwrap returns ErrNonceAccountNotFound.
func (e *NonceAccountNotFoundError) Unwrap() error { return ErrNonceAccountNotFound }

// TransactionError reports that the transaction landed, but it failed.
type TransactionError struct {
	Signature solana.Signature
	// Err is the err field of the signature status or notification as the
	// RPC server returned it, for example
	// map[string]any{"InstructionError": []any{0, map[string]any{"Custom": 1}}}.
	Err any
}

func (e *TransactionError) Error() string {
	reason, err := stdjson.Marshal(e.Err)
	if err != nil {
		reason = fmt.Appendf(nil, "%v", e.Err)
	}
	return fmt.Sprintf("transaction %s failed: %s", e.Signature, reason)
}

// Unwrap returns ErrTransactionFailed.
func (e *TransactionError) Unwrap() error { return ErrTransactionFailed }
