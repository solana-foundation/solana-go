package token

import (
	"context"
	"os"
	"testing"
	"time"

	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	sendandconfirmtransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	ag_require "github.com/stretchr/testify/require"
)

const testnetRPC = "https://api.testnet.solana.com"
const testnetWS = "wss://api.testnet.solana.com"

func loadTestnetWallet(t *testing.T) ag_solanago.PrivateKey {
	t.Helper()
	home, err := os.UserHomeDir()
	ag_require.NoError(t, err)
	key, err := ag_solanago.PrivateKeyFromSolanaKeygenFile(home + "/.config/solana/id.json")
	ag_require.NoError(t, err)
	return key
}

func signAndSend(
	t *testing.T,
	ctx context.Context,
	rpcClient *rpc.Client,
	wsClient *ws.Client,
	payer ag_solanago.PrivateKey,
	extraSigners []ag_solanago.PrivateKey,
	instructions []ag_solanago.Instruction,
) ag_solanago.Signature {
	t.Helper()
	recent, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	ag_require.NoError(t, err)

	tx, err := ag_solanago.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		ag_solanago.TransactionPayer(payer.PublicKey()),
	)
	ag_require.NoError(t, err)

	signerMap := map[ag_solanago.PublicKey]ag_solanago.PrivateKey{
		payer.PublicKey(): payer,
	}
	for _, s := range extraSigners {
		signerMap[s.PublicKey()] = s
	}

	_, err = tx.Sign(func(key ag_solanago.PublicKey) *ag_solanago.PrivateKey {
		if pk, ok := signerMap[key]; ok {
			return &pk
		}
		return nil
	})
	ag_require.NoError(t, err)

	sig, err := sendandconfirmtransaction.SendAndConfirmTransactionWithTimeout(
		ctx, rpcClient, wsClient, tx, 30*time.Second,
	)
	ag_require.NoError(t, err)
	return sig
}

func TestTestnet_SimulateInitializeMint2(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testnet test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := rpc.New(testnetRPC)
	payer := loadTestnetWallet(t)
	mintKeypair := ag_solanago.NewWallet()

	rentExempt, err := client.GetMinimumBalanceForRentExemption(ctx, 82, rpc.CommitmentFinalized)
	ag_require.NoError(t, err)

	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	ag_require.NoError(t, err)

	tx, err := ag_solanago.NewTransaction(
		[]ag_solanago.Instruction{
			system.NewCreateAccountInstruction(
				rentExempt,
				82,
				ag_solanago.TokenProgramID,
				payer.PublicKey(),
				mintKeypair.PublicKey(),
			).Build(),
			NewInitializeMint2Instruction(
				9,
				payer.PublicKey(),
				payer.PublicKey(),
				mintKeypair.PublicKey(),
			).Build(),
		},
		recent.Value.Blockhash,
		ag_solanago.TransactionPayer(payer.PublicKey()),
	)
	ag_require.NoError(t, err)

	_, err = tx.Sign(func(key ag_solanago.PublicKey) *ag_solanago.PrivateKey {
		if key.Equals(payer.PublicKey()) {
			return &payer
		}
		pk := mintKeypair.PrivateKey
		return &pk
	})
	ag_require.NoError(t, err)

	result, err := client.SimulateTransaction(ctx, tx)
	ag_require.NoError(t, err)
	ag_require.Nil(t, result.Value.Err, "simulation failed: %v", result.Value.Err)
	t.Logf("InitializeMint2 simulation succeeded, consumed %d CUs", *result.Value.UnitsConsumed)
}

func TestTestnet_SimulateGetAccountDataSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testnet test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := rpc.New(testnetRPC)
	payer := loadTestnetWallet(t)

	nativeMint := ag_solanago.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")

	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	ag_require.NoError(t, err)

	tx, err := ag_solanago.NewTransaction(
		[]ag_solanago.Instruction{
			NewGetAccountDataSizeInstruction(nativeMint).Build(),
		},
		recent.Value.Blockhash,
		ag_solanago.TransactionPayer(payer.PublicKey()),
	)
	ag_require.NoError(t, err)

	_, err = tx.Sign(func(key ag_solanago.PublicKey) *ag_solanago.PrivateKey {
		if key.Equals(payer.PublicKey()) {
			return &payer
		}
		return nil
	})
	ag_require.NoError(t, err)

	result, err := client.SimulateTransaction(ctx, tx)
	ag_require.NoError(t, err)
	ag_require.Nil(t, result.Value.Err, "simulation failed: %v", result.Value.Err)
	t.Logf("GetAccountDataSize simulation succeeded, consumed %d CUs", *result.Value.UnitsConsumed)
}

func TestTestnet_SimulateAmountToUiAmount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testnet test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := rpc.New(testnetRPC)
	payer := loadTestnetWallet(t)

	nativeMint := ag_solanago.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")

	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	ag_require.NoError(t, err)

	tx, err := ag_solanago.NewTransaction(
		[]ag_solanago.Instruction{
			NewAmountToUiAmountInstruction(1_000_000_000, nativeMint).Build(),
		},
		recent.Value.Blockhash,
		ag_solanago.TransactionPayer(payer.PublicKey()),
	)
	ag_require.NoError(t, err)

	_, err = tx.Sign(func(key ag_solanago.PublicKey) *ag_solanago.PrivateKey {
		if key.Equals(payer.PublicKey()) {
			return &payer
		}
		return nil
	})
	ag_require.NoError(t, err)

	result, err := client.SimulateTransaction(ctx, tx)
	ag_require.NoError(t, err)
	ag_require.Nil(t, result.Value.Err, "simulation failed: %v", result.Value.Err)
	t.Logf("AmountToUiAmount simulation succeeded, consumed %d CUs", *result.Value.UnitsConsumed)
}

func TestTestnet_SimulateUiAmountToAmount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testnet test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := rpc.New(testnetRPC)
	payer := loadTestnetWallet(t)

	nativeMint := ag_solanago.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")

	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	ag_require.NoError(t, err)

	tx, err := ag_solanago.NewTransaction(
		[]ag_solanago.Instruction{
			NewUiAmountToAmountInstruction("1.5", nativeMint).Build(),
		},
		recent.Value.Blockhash,
		ag_solanago.TransactionPayer(payer.PublicKey()),
	)
	ag_require.NoError(t, err)

	_, err = tx.Sign(func(key ag_solanago.PublicKey) *ag_solanago.PrivateKey {
		if key.Equals(payer.PublicKey()) {
			return &payer
		}
		return nil
	})
	ag_require.NoError(t, err)

	result, err := client.SimulateTransaction(ctx, tx)
	ag_require.NoError(t, err)
	ag_require.Nil(t, result.Value.Err, "simulation failed: %v", result.Value.Err)
	t.Logf("UiAmountToAmount simulation succeeded, consumed %d CUs", *result.Value.UnitsConsumed)
}

func TestTestnet_WithdrawExcessLamports(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testnet test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	rpcClient := rpc.New(testnetRPC)
	wsClient, err := ws.Connect(ctx, testnetWS)
	ag_require.NoError(t, err)
	defer wsClient.Close()

	payer := loadTestnetWallet(t)

	mintKeypair := ag_solanago.NewWallet()
	rentExempt, err := rpcClient.GetMinimumBalanceForRentExemption(ctx, 82, rpc.CommitmentFinalized)
	ag_require.NoError(t, err)

	// Fund the mint with extra lamports above rent-exempt minimum.
	extraLamports := uint64(100_000)
	sig := signAndSend(t, ctx, rpcClient, wsClient, payer, []ag_solanago.PrivateKey{mintKeypair.PrivateKey}, []ag_solanago.Instruction{
		system.NewCreateAccountInstruction(
			rentExempt+extraLamports,
			82,
			ag_solanago.TokenProgramID,
			payer.PublicKey(),
			mintKeypair.PublicKey(),
		).Build(),
		NewInitializeMint2Instruction(
			9,
			payer.PublicKey(),
			payer.PublicKey(),
			mintKeypair.PublicKey(),
		).Build(),
	})
	t.Logf("Created mint with excess lamports: %s", sig)

	// Now withdraw excess lamports using the p-token instruction (ID 38).
	sig = signAndSend(t, ctx, rpcClient, wsClient, payer, nil, []ag_solanago.Instruction{
		NewWithdrawExcessLamportsInstruction(
			mintKeypair.PublicKey(),
			payer.PublicKey(),
			payer.PublicKey(),
			nil,
		).Build(),
	})
	t.Logf("WithdrawExcessLamports succeeded: %s", sig)
}

func TestTestnet_UnwrapLamports(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testnet test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	rpcClient := rpc.New(testnetRPC)
	wsClient, err := ws.Connect(ctx, testnetWS)
	ag_require.NoError(t, err)
	defer wsClient.Close()

	payer := loadTestnetWallet(t)
	nativeMint := ag_solanago.SolMint

	tokenAccKeypair := ag_solanago.NewWallet()
	rentExempt, err := rpcClient.GetMinimumBalanceForRentExemption(ctx, 165, rpc.CommitmentFinalized)
	ag_require.NoError(t, err)

	wrapAmount := uint64(10_000_000) // 0.01 SOL
	sig := signAndSend(t, ctx, rpcClient, wsClient, payer, []ag_solanago.PrivateKey{tokenAccKeypair.PrivateKey}, []ag_solanago.Instruction{
		system.NewCreateAccountInstruction(
			rentExempt+wrapAmount,
			165,
			ag_solanago.TokenProgramID,
			payer.PublicKey(),
			tokenAccKeypair.PublicKey(),
		).Build(),
		NewInitializeAccount3Instruction(
			payer.PublicKey(),
			tokenAccKeypair.PublicKey(),
			nativeMint,
		).Build(),
	})
	t.Logf("Created wrapped SOL account and funded: %s", sig)

	// Unwrap partial using UnwrapLamports (ID 45) with a specific amount.
	unwrapAmount := uint64(5_000_000) // 0.005 SOL
	sig = signAndSend(t, ctx, rpcClient, wsClient, payer, nil, []ag_solanago.Instruction{
		NewUnwrapLamportsWithAmountInstruction(
			unwrapAmount,
			tokenAccKeypair.PublicKey(),
			payer.PublicKey(),
			payer.PublicKey(),
			nil,
		).Build(),
	})
	t.Logf("UnwrapLamports (partial) succeeded: %s", sig)

	// Unwrap remaining using UnwrapLamports with nil amount (unwrap all).
	sig = signAndSend(t, ctx, rpcClient, wsClient, payer, nil, []ag_solanago.Instruction{
		NewUnwrapLamportsInstruction(
			tokenAccKeypair.PublicKey(),
			payer.PublicKey(),
			payer.PublicKey(),
			nil,
		).Build(),
	})
	t.Logf("UnwrapLamports (all) succeeded: %s", sig)
}

func TestTestnet_Batch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testnet test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	rpcClient := rpc.New(testnetRPC)
	wsClient, err := ws.Connect(ctx, testnetWS)
	ag_require.NoError(t, err)
	defer wsClient.Close()

	payer := loadTestnetWallet(t)

	mintKeypair := ag_solanago.NewWallet()
	tokenAcc1Keypair := ag_solanago.NewWallet()
	tokenAcc2Keypair := ag_solanago.NewWallet()

	mintRent, err := rpcClient.GetMinimumBalanceForRentExemption(ctx, 82, rpc.CommitmentFinalized)
	ag_require.NoError(t, err)
	accRent, err := rpcClient.GetMinimumBalanceForRentExemption(ctx, 165, rpc.CommitmentFinalized)
	ag_require.NoError(t, err)

	// Create mint + two token accounts + mint tokens to account 1.
	sig := signAndSend(t, ctx, rpcClient, wsClient, payer,
		[]ag_solanago.PrivateKey{mintKeypair.PrivateKey, tokenAcc1Keypair.PrivateKey, tokenAcc2Keypair.PrivateKey},
		[]ag_solanago.Instruction{
			system.NewCreateAccountInstruction(mintRent, 82, ag_solanago.TokenProgramID, payer.PublicKey(), mintKeypair.PublicKey()).Build(),
			NewInitializeMint2Instruction(9, payer.PublicKey(), payer.PublicKey(), mintKeypair.PublicKey()).Build(),
			system.NewCreateAccountInstruction(accRent, 165, ag_solanago.TokenProgramID, payer.PublicKey(), tokenAcc1Keypair.PublicKey()).Build(),
			NewInitializeAccount3Instruction(payer.PublicKey(), tokenAcc1Keypair.PublicKey(), mintKeypair.PublicKey()).Build(),
			system.NewCreateAccountInstruction(accRent, 165, ag_solanago.TokenProgramID, payer.PublicKey(), tokenAcc2Keypair.PublicKey()).Build(),
			NewInitializeAccount3Instruction(payer.PublicKey(), tokenAcc2Keypair.PublicKey(), mintKeypair.PublicKey()).Build(),
			NewMintToInstruction(1000, mintKeypair.PublicKey(), tokenAcc1Keypair.PublicKey(), payer.PublicKey(), nil).Build(),
		},
	)
	t.Logf("Setup (mint + accounts + mintTo): %s", sig)

	// Batch: two transfers from account1 -> account2 in a single instruction.
	transfer1 := NewTransferInstruction(100, tokenAcc1Keypair.PublicKey(), tokenAcc2Keypair.PublicKey(), payer.PublicKey(), nil).Build()
	transfer2 := NewTransferInstruction(200, tokenAcc1Keypair.PublicKey(), tokenAcc2Keypair.PublicKey(), payer.PublicKey(), nil).Build()

	batchIx := NewBatchInstruction(transfer1, transfer2).Build()
	sig = signAndSend(t, ctx, rpcClient, wsClient, payer, nil, []ag_solanago.Instruction{batchIx})
	t.Logf("Batch (2 transfers) succeeded: %s", sig)
}
