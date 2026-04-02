package rpc

import (
	"context"

	"github.com/gagliardetto/solana-go"
)

// IsBlockhashValid returns whether a blockhash is still valid or not.
func (cl *Client) IsBlockhashValid(
	ctx context.Context,
	// Blockhash to be queried. Required.
	blockHash solana.Hash,

	// Commitment requirement. Optional.
	commitment CommitmentType,
) (out *IsValidBlockhashResult, err error) {
	params := []any{blockHash}
	if commitment != "" {
		params = append(params, M{"commitment": string(commitment)})
	}

	err = cl.rpcClient.CallForInto(ctx, &out, "isBlockhashValid", params)
	return
}
