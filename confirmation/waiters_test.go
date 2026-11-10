package confirmation

import (
	"context"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

// The cases follow kit's waiters-test.ts, for each of the three waiters.

var (
	testPayer     = solana.MustPublicKeyFromBase58("FnHyam9w4NZoWR6mKN1CuGBritdsEWZQa4Z4oawLZGxa")
	testRecipient = solana.MustPublicKeyFromBase58("5ZWj7a1f8tWkjBESHKgrLmXshuXxqeY9SYcfbshpAqPG")
	testBlockhash = solana.MustHashFromBase58("4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM")
)

func transferInstruction() solana.Instruction {
	return system.NewTransferInstruction(1, testPayer, testRecipient).Build()
}

// signedTransaction returns a blockhash transaction whose fee payer signature
// is testSignature.
func signedTransaction(t *testing.T) *solana.Transaction {
	t.Helper()
	tx, err := solana.NewTransaction([]solana.Instruction{transferInstruction()}, testBlockhash,
		solana.TransactionPayer(testPayer))
	require.NoError(t, err)
	tx.Signatures = []solana.Signature{testSignature}
	return tx
}

// nonceTransaction returns a durable nonce transaction that uses testNonce of
// testNonceAccount.
func nonceTransaction(t *testing.T) *solana.Transaction {
	t.Helper()
	advance := system.NewAdvanceNonceAccountInstruction(testNonceAccount, solana.SysVarRecentBlockHashesPubkey, testPayer).Build()
	tx, err := solana.NewTransaction([]solana.Instruction{advance, transferInstruction()}, testNonce,
		solana.TransactionPayer(testPayer))
	require.NoError(t, err)
	tx.Signatures = []solana.Signature{testSignature}
	return tx
}

// waiterCase runs one waiter with a fake signature strategy and a fake
// lifetime strategy, and reports the inputs each strategy received.
type waiterCase struct {
	name string
	// wait runs the waiter on tx with the given strategies.
	wait func(ctx context.Context, tx *solana.Transaction, signature RecentSignatureConfirmationFunc, lifetime func(context.Context) error) error
	// tx returns a transaction the waiter accepts.
	tx func(t *testing.T) *solana.Transaction
	// lifetimeInput is what the lifetime strategy is expected to receive.
	lifetimeInput string
}

func waiterCases() []waiterCase {
	return []waiterCase{
		{
			name: "recent transaction",
			tx:   signedTransaction,
			wait: func(ctx context.Context, tx *solana.Transaction, signature RecentSignatureConfirmationFunc, lifetime func(context.Context) error) error {
				return WaitForRecentTransactionConfirmation(ctx, RecentTransactionConfig{
					Commitment:                  rpc.CommitmentConfirmed,
					Transaction:                 tx,
					LastValidBlockHeight:        150,
					RecentSignatureConfirmation: signature,
					BlockHeightExceedance: func(ctx context.Context, lastValidBlockHeight uint64, commitment rpc.CommitmentType) error {
						recordInput(ctx, "%d %s", lastValidBlockHeight, commitment)
						return lifetime(ctx)
					},
				})
			},
			lifetimeInput: "150 confirmed",
		},
		{
			name: "durable nonce transaction",
			tx:   nonceTransaction,
			wait: func(ctx context.Context, tx *solana.Transaction, signature RecentSignatureConfirmationFunc, lifetime func(context.Context) error) error {
				return WaitForDurableNonceTransactionConfirmation(ctx, DurableNonceTransactionConfig{
					Commitment:                  rpc.CommitmentConfirmed,
					Transaction:                 tx,
					RecentSignatureConfirmation: signature,
					NonceInvalidation: func(ctx context.Context, nonceAccount solana.PublicKey, currentNonce solana.Hash, commitment rpc.CommitmentType) error {
						recordInput(ctx, "%s %s %s", nonceAccount, currentNonce, commitment)
						return lifetime(ctx)
					},
				})
			},
			lifetimeInput: testNonceAccount.String() + " " + testNonce.String() + " confirmed",
		},
		{
			name: "recent transaction until timeout",
			tx:   signedTransaction,
			wait: func(ctx context.Context, tx *solana.Transaction, signature RecentSignatureConfirmationFunc, lifetime func(context.Context) error) error {
				var sig solana.Signature
				if tx != nil && len(tx.Signatures) > 0 {
					sig = tx.Signatures[0]
				}
				return WaitForRecentTransactionConfirmationUntilTimeout(ctx, RecentTransactionUntilTimeoutConfig{
					Commitment:                  rpc.CommitmentConfirmed,
					Signature:                   sig,
					RecentSignatureConfirmation: signature,
					Timeout: func(ctx context.Context, commitment rpc.CommitmentType) error {
						recordInput(ctx, "%s", commitment)
						return lifetime(ctx)
					},
				})
			},
			lifetimeInput: "confirmed",
		},
	}
}

// inputs collects what the fake strategies were called with. It travels in
// the context, so each test has its own.
type inputs struct{ signature, lifetime []string }

type inputsKey struct{}

func withInputs(ctx context.Context) (context.Context, *inputs) {
	in := &inputs{}
	return context.WithValue(ctx, inputsKey{}, in), in
}

func recordInput(ctx context.Context, format string, args ...any) {
	if in, ok := ctx.Value(inputsKey{}).(*inputs); ok {
		in.lifetime = append(in.lifetime, fmt.Sprintf(format, args...))
	}
}

func signatureStrategy(run func(context.Context) error) RecentSignatureConfirmationFunc {
	return func(ctx context.Context, signature solana.Signature, commitment rpc.CommitmentType) error {
		if in, ok := ctx.Value(inputsKey{}).(*inputs); ok {
			in.signature = append(in.signature, fmt.Sprintf("%s %s", signature, commitment))
		}
		return run(ctx)
	}
}

func TestWaiters(t *testing.T) {
	for _, wc := range waiterCases() {
		t.Run(wc.name, func(t *testing.T) {
			t.Run("throws when the context is already canceled", func(t *testing.T) {
				ctx, in := withInputs(context.Background())
				ctx, cancel := context.WithCancel(ctx)
				cancel()

				err := wc.wait(ctx, wc.tx(t), signatureStrategy(blockOnContext), blockOnContext)

				require.ErrorIs(t, err, context.Canceled)
				require.Empty(t, in.signature)
				require.Empty(t, in.lifetime)
			})

			t.Run("passes the inputs to both strategies", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					ctx, in := withInputs(context.Background())
					ctx, cancel := context.WithCancel(ctx)
					done := make(chan error, 1)
					go func() { done <- wc.wait(ctx, wc.tx(t), signatureStrategy(blockOnContext), blockOnContext) }()

					synctest.Wait()
					require.Equal(t, []string{testSignature.String() + " confirmed"}, in.signature)
					require.Equal(t, []string{wc.lifetimeInput}, in.lifetime)

					// Cancellation reaches both strategies.
					cancel()
					require.ErrorIs(t, <-done, context.Canceled)
				})
			})

			t.Run("throws when the fee payer signature is missing", func(t *testing.T) {
				for name, signatures := range map[string][]solana.Signature{
					"no signatures":  nil,
					"zero signature": {{}},
				} {
					t.Run(name, func(t *testing.T) {
						tx := wc.tx(t)
						tx.Signatures = signatures
						// The missing signature wins over cancellation, as in kit.
						ctx, cancel := context.WithCancel(context.Background())
						cancel()

						err := wc.wait(ctx, tx, signatureStrategy(blockOnContext), blockOnContext)

						require.ErrorIs(t, err, ErrFeePayerSignatureMissing)
					})
				}
			})

			t.Run("resolves when the signature confirms although the lifetime strategy failed", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					failLater := func(context.Context) error {
						time.Sleep(time.Millisecond)
						return errStrategy
					}

					err := wc.wait(context.Background(), wc.tx(t), signatureStrategy(func(context.Context) error { return nil }), failLater)

					require.NoError(t, err)
				})
			})

			t.Run("throws when the lifetime strategy throws", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					err := wc.wait(context.Background(), wc.tx(t), signatureStrategy(blockOnContext),
						func(context.Context) error { return errStrategy })

					require.ErrorIs(t, err, errStrategy)
				})
			})

			t.Run("throws when the signature strategy throws", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					err := wc.wait(context.Background(), wc.tx(t),
						signatureStrategy(func(context.Context) error { return errStrategy }), blockOnContext)

					require.ErrorIs(t, err, errStrategy)
				})
			})
		})
	}
}

func TestWaiters_RejectInvalidConfig(t *testing.T) {
	ctx := context.Background()
	signature := signatureStrategy(blockOnContext)

	t.Run("nil strategy", func(t *testing.T) {
		require.ErrorIs(t, WaitForRecentTransactionConfirmation(ctx, RecentTransactionConfig{
			Transaction: signedTransaction(t), RecentSignatureConfirmation: signature,
		}), errNilStrategy)
		require.ErrorIs(t, WaitForDurableNonceTransactionConfirmation(ctx, DurableNonceTransactionConfig{
			Transaction: nonceTransaction(t), NonceInvalidation: func(context.Context, solana.PublicKey, solana.Hash, rpc.CommitmentType) error { return nil },
		}), errNilStrategy)
		require.ErrorIs(t, WaitForRecentTransactionConfirmationUntilTimeout(ctx, RecentTransactionUntilTimeoutConfig{
			Signature: testSignature, RecentSignatureConfirmation: signature,
		}), errNilStrategy)
	})

	t.Run("unknown commitment", func(t *testing.T) {
		err := WaitForRecentTransactionConfirmation(ctx, RecentTransactionConfig{
			Commitment: "recent", Transaction: signedTransaction(t),
		})
		require.ErrorContains(t, err, `unsupported commitment "recent"`)
	})

	t.Run("nil transaction", func(t *testing.T) {
		require.ErrorContains(t, WaitForRecentTransactionConfirmation(ctx, RecentTransactionConfig{}), "transaction is nil")
		require.ErrorContains(t, WaitForDurableNonceTransactionConfirmation(ctx, DurableNonceTransactionConfig{}), "transaction is nil")
	})

	t.Run("transaction without durable nonce", func(t *testing.T) {
		err := WaitForDurableNonceTransactionConfirmation(ctx, DurableNonceTransactionConfig{
			Transaction: signedTransaction(t),
		})
		require.ErrorIs(t, err, ErrNotDurableNonceTransaction)
	})
}
