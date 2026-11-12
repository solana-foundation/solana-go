package confirmation_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/confirmation"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func ExampleWaitForRecentTransactionConfirmation() {
	ctx := context.Background()
	rpcClient := rpc.New(rpc.DevNet_RPC)
	wsClient, err := ws.Connect(ctx, rpc.DevNet_WS)
	if err != nil {
		panic(err)
	}
	defer wsClient.Close()

	payer := solana.NewWallet()
	latest, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		panic(err)
	}
	tx, err := solana.NewTransaction(
		[]solana.Instruction{system.NewTransferInstruction(1, payer.PublicKey(), solana.SystemProgramID).Build()},
		latest.Value.Blockhash,
		solana.TransactionPayer(payer.PublicKey()),
	)
	if err != nil {
		panic(err)
	}
	if _, err := tx.Sign(func(solana.PublicKey) *solana.PrivateKey { return &payer.PrivateKey }); err != nil {
		panic(err)
	}
	if _, err := rpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{PreflightCommitment: rpc.CommitmentConfirmed}); err != nil {
		panic(err)
	}

	err = confirmation.WaitForRecentTransactionConfirmation(ctx, confirmation.RecentTransactionConfig{
		Commitment:                  rpc.CommitmentConfirmed,
		Transaction:                 tx,
		LastValidBlockHeight:        latest.Value.LastValidBlockHeight,
		RecentSignatureConfirmation: confirmation.NewRecentSignatureConfirmation(rpcClient, wsClient),
		BlockHeightExceedance:       confirmation.NewBlockHeightExceedance(rpcClient, wsClient),
	})
	var txErr *confirmation.TransactionError
	switch {
	case err == nil:
		fmt.Println("confirmed", tx.Signatures[0])
	case errors.Is(err, confirmation.ErrBlockHeightExceeded):
		fmt.Println("the blockhash expired")
	case errors.As(err, &txErr):
		fmt.Println("the transaction failed:", txErr.Err)
	default:
		panic(err)
	}
}

func ExampleWaitForDurableNonceTransactionConfirmation() {
	ctx := context.Background()
	rpcClient := rpc.New(rpc.DevNet_RPC)
	wsClient, err := ws.Connect(ctx, rpc.DevNet_WS)
	if err != nil {
		panic(err)
	}
	defer wsClient.Close()

	// tx starts with an AdvanceNonceAccount instruction, its recent blockhash
	// is the nonce, and it has been signed and sent.
	var tx *solana.Transaction

	err = confirmation.WaitForDurableNonceTransactionConfirmation(ctx, confirmation.DurableNonceTransactionConfig{
		Commitment:                  rpc.CommitmentConfirmed,
		Transaction:                 tx,
		RecentSignatureConfirmation: confirmation.NewRecentSignatureConfirmation(rpcClient, wsClient),
		NonceInvalidation:           confirmation.NewNonceInvalidation(rpcClient, wsClient),
	})
	switch {
	case err == nil:
		fmt.Println("confirmed")
	case errors.Is(err, confirmation.ErrInvalidNonce):
		// The nonce advanced. The transaction can no longer land, unless it
		// landed and failed (kit #1234): check its status first.
		fmt.Println("the nonce advanced")
	case errors.Is(err, confirmation.ErrNonceAccountNotFound):
		fmt.Println("the nonce account does not exist")
	default:
		panic(err)
	}
}
