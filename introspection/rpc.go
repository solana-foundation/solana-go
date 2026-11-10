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

package introspection

import (
	"errors"
	"fmt"

	"github.com/buger/jsonparser"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// WalkGetTransactionResult walks the transaction of a getTransaction result
// (see Walk).
//
// Fetch the transaction with the json, base58, base64 or base64+zstd
// encoding, and set MaxSupportedTransactionVersion to receive v0 and v1
// transactions. The jsonParsed encoding is not supported.
func WalkGetTransactionResult(res *rpc.GetTransactionResult) ([]*Instruction, error) {
	if res == nil || res.Transaction == nil {
		return nil, fmt.Errorf("%w: no getTransaction result", ErrUnrecognizedTransaction)
	}
	if err := checkVersion(res.Version); err != nil {
		return nil, err
	}
	tx, err := res.Transaction.GetTransaction()
	if err != nil {
		return nil, fmt.Errorf("introspection: decode transaction: %w", err)
	}
	if tx == nil {
		return nil, fmt.Errorf("%w: empty getTransaction result", ErrUnrecognizedTransaction)
	}
	return Walk(&tx.Message, res.Meta)
}

// WalkTransactionWithMeta walks a transaction of a getBlock result or of a
// blockSubscribe notification (see Walk).
//
// Fetch the block with full transaction details in the json, base58, base64
// or base64+zstd encoding. The jsonParsed encoding returns
// ErrJSONParsedUnsupported, and other transaction details return
// ErrUnrecognizedTransaction.
func WalkTransactionWithMeta(tx *rpc.TransactionWithMeta) ([]*Instruction, error) {
	if tx == nil {
		return nil, fmt.Errorf("%w: no block transaction", ErrUnrecognizedTransaction)
	}
	if err := checkVersion(tx.Version); err != nil {
		return nil, err
	}
	decoded, err := decodeTransaction(tx.Transaction)
	if err != nil {
		return nil, err
	}
	return Walk(&decoded.Message, tx.Meta)
}

// WalkTransactionForAddress walks a transaction of a getTransactionsForAddress
// result (see Walk).
//
// Fetch the transactions with full transaction details in the json, base58,
// base64 or base64+zstd encoding. The jsonParsed encoding returns
// ErrJSONParsedUnsupported, and other transaction details return
// ErrUnrecognizedTransaction.
func WalkTransactionForAddress(tx *rpc.TransactionForAddress) ([]*Instruction, error) {
	if tx == nil {
		return nil, fmt.Errorf("%w: no transaction", ErrUnrecognizedTransaction)
	}
	decoded, err := decodeTransaction(tx.Transaction)
	if err != nil {
		return nil, err
	}
	return Walk(&decoded.Message, tx.Meta)
}

// checkVersion accepts the transaction versions of an RPC response that this
// package supports: legacy, 0 (also the value of a missing version) and 1.
func checkVersion(version rpc.TransactionVersion) error {
	switch version {
	case rpc.LegacyTransactionVersion, 0, 1:
		return nil
	default:
		return fmt.Errorf("%w: %d", ErrUnsupportedVersion, version)
	}
}

// decodeTransaction decodes the transaction of a getBlock, blockSubscribe or
// getTransactionsForAddress item. rpc.DataBytesOrJSON labels every JSON
// transaction as jsonParsed: only the message of the json encoding has a
// header.
func decodeTransaction(data *rpc.DataBytesOrJSON) (*solana.Transaction, error) {
	if data == nil {
		return nil, fmt.Errorf("%w: no transaction, fetch it with full transaction details", ErrUnrecognizedTransaction)
	}
	if raw := data.GetRawJSON(); raw != nil {
		if err := hasKey(raw, "message"); err != nil {
			if errors.Is(err, jsonparser.KeyPathNotFoundError) {
				return nil, fmt.Errorf("%w: no transaction message, fetch it with full transaction details", ErrUnrecognizedTransaction)
			}
			return nil, fmt.Errorf("introspection: decode transaction: %w", err)
		}
		if err := hasKey(raw, "message", "header"); err != nil {
			if errors.Is(err, jsonparser.KeyPathNotFoundError) {
				return nil, ErrJSONParsedUnsupported
			}
			return nil, fmt.Errorf("introspection: decode transaction: %w", err)
		}
	} else if len(data.GetBinary()) == 0 {
		return nil, fmt.Errorf("%w: empty transaction", ErrUnrecognizedTransaction)
	}

	tx, err := rpc.TransactionWithMeta{Transaction: data}.GetTransaction()
	if err != nil {
		return nil, fmt.Errorf("introspection: decode transaction: %w", err)
	}
	return tx, nil
}

func hasKey(raw []byte, keys ...string) error {
	_, _, _, err := jsonparser.Get(raw, keys...)
	return err
}
