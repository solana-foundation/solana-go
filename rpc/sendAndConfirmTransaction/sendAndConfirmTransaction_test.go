package sendandconfirmtransaction

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	stdjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/require"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/confirmation"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
)

// The cases follow kit's send-transaction-internal-test.ts,
// send-and-confirm-durable-nonce-transaction-test.ts and
// airdrop-internal-test.ts.

var (
	testSignature = solana.MustSignatureFromBase58(
		"5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW")
	testPayer        = solana.MustPublicKeyFromBase58("FnHyam9w4NZoWR6mKN1CuGBritdsEWZQa4Z4oawLZGxa")
	testRecipient    = solana.MustPublicKeyFromBase58("5ZWj7a1f8tWkjBESHKgrLmXshuXxqeY9SYcfbshpAqPG")
	testNonceAccount = solana.MustPublicKeyFromBase58("9999999999999999999999999999999999999999999")
	testBlockhash    = solana.MustHashFromBase58("4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM")
	testNonce        = solana.MustHashFromBase58("8vbqMAqHcrjfQX3nFbpfJmANsm9kj8m2HWQrRfQYa9WF")
	errStrategy      = errors.New("o no")
)

// fakeRPC implements rpc.JSONRPCClient with one handler per method, which
// answers with raw JSON.
type fakeRPC struct {
	mu       sync.Mutex
	handlers map[string]func(ctx context.Context, params []any) (string, error)
	calls    map[string][][]any
}

func newFakeRPC() *fakeRPC {
	f := &fakeRPC{
		handlers: map[string]func(context.Context, []any) (string, error){},
		calls:    map[string][][]any{},
	}
	f.handlers["sendTransaction"] = func(context.Context, []any) (string, error) {
		return fmt.Sprintf("%q", testSignature), nil
	}
	return f
}

func (f *fakeRPC) on(method string, h func(ctx context.Context, params []any) (string, error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.handlers[method] = h
}

func (f *fakeRPC) callParams(method string) [][]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.calls[method])
}

func (f *fakeRPC) CallForInto(ctx context.Context, out any, method string, params []any) error {
	f.mu.Lock()
	f.calls[method] = append(f.calls[method], params)
	h := f.handlers[method]
	f.mu.Unlock()
	if h == nil {
		return fmt.Errorf("fakeRPC: unexpected call to %s", method)
	}
	raw, err := h(ctx, params)
	if err != nil {
		return err
	}
	return stdjson.Unmarshal([]byte(raw), out)
}

func (f *fakeRPC) CallWithCallback(context.Context, string, []any, func(*http.Request, *http.Response) error) error {
	return errors.New("fakeRPC: CallWithCallback is not supported")
}

func (f *fakeRPC) CallBatch(context.Context, jsonrpc.RPCRequests) (jsonrpc.RPCResponses, error) {
	return nil, errors.New("fakeRPC: CallBatch is not supported")
}

func signatureStatus(status string) func(context.Context, []any) (string, error) {
	return func(context.Context, []any) (string, error) {
		return `{"context":{"slot":1},"value":[` + status + `]}`, nil
	}
}

func transfer() solana.Instruction {
	return system.NewTransferInstruction(1, testPayer, testRecipient).Build()
}

func blockhashTransaction(t *testing.T) *solana.Transaction {
	t.Helper()
	tx, err := solana.NewTransaction([]solana.Instruction{transfer()}, testBlockhash, solana.TransactionPayer(testPayer))
	require.NoError(t, err)
	tx.Signatures = []solana.Signature{testSignature}
	return tx
}

func nonceTransaction(t *testing.T) *solana.Transaction {
	t.Helper()
	advance := system.NewAdvanceNonceAccountInstruction(testNonceAccount, solana.SysVarRecentBlockHashesPubkey, testPayer).Build()
	tx, err := solana.NewTransaction([]solana.Instruction{advance, transfer()}, testNonce, solana.TransactionPayer(testPayer))
	require.NoError(t, err)
	tx.Signatures = []solana.Signature{testSignature}
	return tx
}

func blockOnContext(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// strategies are fake strategies that record their inputs.
type strategies struct {
	mu        sync.Mutex
	inputs    []string
	signature func(context.Context) error
	lifetime  func(context.Context) error
}

func (s *strategies) record(format string, args ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inputs = append(s.inputs, fmt.Sprintf(format, args...))
}

func (s *strategies) recorded() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.inputs)
}

func (s *strategies) signatureFunc() confirmation.RecentSignatureConfirmationFunc {
	return func(ctx context.Context, signature solana.Signature, commitment rpc.CommitmentType) error {
		s.record("signature %s %s", signature, commitment)
		return s.signature(ctx)
	}
}

func (s *strategies) blockHeightFunc() confirmation.BlockHeightExceedanceFunc {
	return func(ctx context.Context, lastValidBlockHeight uint64, commitment rpc.CommitmentType) error {
		s.record("block height %d %s", lastValidBlockHeight, commitment)
		return s.lifetime(ctx)
	}
}

func (s *strategies) nonceFunc() confirmation.NonceInvalidationFunc {
	return func(ctx context.Context, nonceAccount solana.PublicKey, currentNonce solana.Hash, commitment rpc.CommitmentType) error {
		s.record("nonce %s %s %s", nonceAccount, currentNonce, commitment)
		return s.lifetime(ctx)
	}
}

func (s *strategies) timeoutFunc() confirmation.TimeoutFunc {
	return func(ctx context.Context, commitment rpc.CommitmentType) error {
		s.record("timeout %s", commitment)
		return s.lifetime(ctx)
	}
}

// lifetime describes one of the two send helpers and how its lifetime ends.
type lifetime struct {
	name string
	tx   func(t *testing.T) *solana.Transaction
	send func(ctx context.Context, rpcClient *rpc.Client, tx *solana.Transaction, commitment rpc.CommitmentType, opts rpc.TransactionOpts, s *strategies) (solana.Signature, error)
	// ended is the error of the lifetime strategy when the lifetime ended.
	ended error
	// endedSentinel is the sentinel that ended wraps.
	endedSentinel error
	// input is what the lifetime strategy receives at the confirmed commitment.
	input string
}

func lifetimes() []lifetime {
	return []lifetime{
		{
			name: "blockhash",
			tx:   blockhashTransaction,
			send: func(ctx context.Context, rpcClient *rpc.Client, tx *solana.Transaction, commitment rpc.CommitmentType, opts rpc.TransactionOpts, s *strategies) (solana.Signature, error) {
				return sendAndConfirmWithBlockhashLifetime(ctx, rpcClient, tx, 150, commitment, opts, s.signatureFunc(), s.blockHeightFunc())
			},
			ended:         &confirmation.BlockHeightExceededError{CurrentBlockHeight: 151, LastValidBlockHeight: 150},
			endedSentinel: confirmation.ErrBlockHeightExceeded,
			input:         "block height 150 confirmed",
		},
		{
			name: "durable nonce",
			tx:   nonceTransaction,
			send: func(ctx context.Context, rpcClient *rpc.Client, tx *solana.Transaction, commitment rpc.CommitmentType, opts rpc.TransactionOpts, s *strategies) (solana.Signature, error) {
				return sendAndConfirmDurableNonce(ctx, rpcClient, tx, commitment, opts, s.signatureFunc(), s.nonceFunc())
			},
			ended:         &confirmation.InvalidNonceError{NonceAccount: testNonceAccount, Expected: testNonce, Actual: testBlockhash},
			endedSentinel: confirmation.ErrInvalidNonce,
			input:         fmt.Sprintf("nonce %s %s confirmed", testNonceAccount, testNonce),
		},
	}
}

func succeed(context.Context) error { return nil }

func TestSend_EncodesTransaction(t *testing.T) {
	for _, lt := range lifetimes() {
		t.Run(lt.name, func(t *testing.T) {
			f := newFakeRPC()
			tx := lt.tx(t)

			_, err := lt.send(context.Background(), rpc.NewWithCustomRPCClient(f), tx, rpc.CommitmentConfirmed,
				rpc.TransactionOpts{Encoding: solana.EncodingBase58},
				&strategies{signature: succeed, lifetime: blockOnContext})
			require.NoError(t, err)

			wire, err := tx.MarshalBinary()
			require.NoError(t, err)
			params := f.callParams("sendTransaction")[0]
			require.Equal(t, base64.StdEncoding.EncodeToString(wire), params[0])
			require.Equal(t, solana.EncodingBase64, params[1].(rpc.M)["encoding"], "always base64, like the encoded data")
		})
	}
}

func TestSend_ForwardsOptions(t *testing.T) {
	for _, lt := range lifetimes() {
		t.Run(lt.name, func(t *testing.T) {
			f := newFakeRPC()
			maxRetries, minContextSlot := uint(42), uint64(123)

			_, err := lt.send(context.Background(), rpc.NewWithCustomRPCClient(f), lt.tx(t), rpc.CommitmentFinalized,
				rpc.TransactionOpts{
					MaxRetries:          &maxRetries,
					MinContextSlot:      &minContextSlot,
					PreflightCommitment: rpc.CommitmentConfirmed,
				},
				&strategies{signature: succeed, lifetime: blockOnContext})
			require.NoError(t, err)

			require.Equal(t, rpc.M{
				"encoding":            solana.EncodingBase64,
				"skipPreflight":       false,
				"preflightCommitment": rpc.CommitmentConfirmed,
				"maxRetries":          maxRetries,
				"minContextSlot":      minContextSlot,
			}, f.callParams("sendTransaction")[0][1])
		})
	}
}

func TestSend_PreflightCommitment(t *testing.T) {
	commitments := []rpc.CommitmentType{rpc.CommitmentProcessed, rpc.CommitmentConfirmed, rpc.CommitmentFinalized}
	type preflightCase struct {
		commitment, preflight, want rpc.CommitmentType
	}
	cases := []preflightCase{
		// Without a preflight commitment, the simulation is capped at the
		// commitment that is waited for.
		{rpc.CommitmentProcessed, "", rpc.CommitmentProcessed},
		{rpc.CommitmentConfirmed, "", rpc.CommitmentConfirmed},
		// At finalized, the server default applies.
		{rpc.CommitmentFinalized, "", ""},
		{"", "", ""},
	}
	// An explicit preflight commitment is always honored.
	for _, commitment := range commitments {
		for _, preflight := range commitments {
			cases = append(cases, preflightCase{commitment, preflight, preflight})
		}
	}
	for _, lt := range lifetimes() {
		for _, tt := range cases {
			t.Run(fmt.Sprintf("%s/%s/%s", lt.name, tt.commitment, tt.preflight), func(t *testing.T) {
				f := newFakeRPC()

				_, err := lt.send(context.Background(), rpc.NewWithCustomRPCClient(f), lt.tx(t), tt.commitment,
					rpc.TransactionOpts{PreflightCommitment: tt.preflight},
					&strategies{signature: succeed, lifetime: blockOnContext})
				require.NoError(t, err)

				got, ok := f.callParams("sendTransaction")[0][1].(rpc.M)["preflightCommitment"]
				if tt.want == "" {
					require.False(t, ok, "no preflight commitment, got %v", got)
				} else {
					require.Equal(t, tt.want, got)
				}
			})
		}
	}
}

func TestSend_PassesInputsToStrategies(t *testing.T) {
	for _, lt := range lifetimes() {
		t.Run(lt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				s := &strategies{signature: blockOnContext, lifetime: blockOnContext}
				ctx, cancel := context.WithCancel(context.Background())
				done := make(chan error, 1)
				go func() {
					_, err := lt.send(ctx, rpc.NewWithCustomRPCClient(newFakeRPC()), lt.tx(t), rpc.CommitmentConfirmed, rpc.TransactionOpts{}, s)
					done <- err
				}()

				synctest.Wait()
				require.ElementsMatch(t, []string{"signature " + testSignature.String() + " confirmed", lt.input}, s.recorded())
				cancel()
				require.ErrorIs(t, <-done, context.Canceled)
			})
		})
	}
}

func TestSend_ReturnsSignature(t *testing.T) {
	for _, lt := range lifetimes() {
		t.Run(lt.name, func(t *testing.T) {
			sig, err := lt.send(context.Background(), rpc.NewWithCustomRPCClient(newFakeRPC()), lt.tx(t), "", rpc.TransactionOpts{},
				&strategies{signature: succeed, lifetime: blockOnContext})

			require.NoError(t, err)
			require.Equal(t, testSignature, sig)
		})
	}
}

func TestSend_ReturnsSignatureWhenWaitingFails(t *testing.T) {
	for _, lt := range lifetimes() {
		t.Run(lt.name, func(t *testing.T) {
			f := newFakeRPC()
			f.on("getSignatureStatuses", signatureStatus("null"))

			sig, err := lt.send(context.Background(), rpc.NewWithCustomRPCClient(f), lt.tx(t), "", rpc.TransactionOpts{},
				&strategies{signature: blockOnContext, lifetime: func(context.Context) error { return lt.ended }})

			require.ErrorIs(t, err, lt.endedSentinel)
			require.Equal(t, testSignature, sig)
		})
	}
}

func TestSend_SendErrorStopsBeforeWaiting(t *testing.T) {
	for _, lt := range lifetimes() {
		t.Run(lt.name, func(t *testing.T) {
			f := newFakeRPC()
			f.on("sendTransaction", func(context.Context, []any) (string, error) {
				return "", &jsonrpc.RPCError{Code: -32002, Message: "Transaction simulation failed"}
			})
			s := &strategies{signature: blockOnContext, lifetime: blockOnContext}

			sig, err := lt.send(context.Background(), rpc.NewWithCustomRPCClient(f), lt.tx(t), "", rpc.TransactionOpts{}, s)

			require.ErrorContains(t, err, "Transaction simulation failed")
			require.Equal(t, testSignature, sig)
			require.Empty(t, s.recorded())
		})
	}
}

func TestSend_ValidatesBeforeSending(t *testing.T) {
	for _, lt := range lifetimes() {
		t.Run(lt.name, func(t *testing.T) {
			for name, tt := range map[string]struct {
				tx         func(t *testing.T) *solana.Transaction
				commitment rpc.CommitmentType
				want       string
				wantErr    error
			}{
				"nil transaction": {tx: func(*testing.T) *solana.Transaction { return nil }, want: "transaction is nil"},
				"unsigned transaction": {tx: func(t *testing.T) *solana.Transaction {
					tx := lt.tx(t)
					tx.Signatures = []solana.Signature{{}}
					return tx
				}, wantErr: confirmation.ErrFeePayerSignatureMissing},
				"unknown commitment": {tx: lt.tx, commitment: "recent", want: `unsupported commitment "recent"`},
			} {
				t.Run(name, func(t *testing.T) {
					f := newFakeRPC()

					_, err := lt.send(context.Background(), rpc.NewWithCustomRPCClient(f), tt.tx(t), tt.commitment,
						rpc.TransactionOpts{}, &strategies{signature: succeed, lifetime: blockOnContext})

					if tt.wantErr != nil {
						require.ErrorIs(t, err, tt.wantErr)
					} else {
						require.ErrorContains(t, err, tt.want)
					}
					require.Empty(t, f.callParams("sendTransaction"), "nothing is sent")
				})
			}
		})
	}

	t.Run("durable nonce helper with a blockhash transaction", func(t *testing.T) {
		f := newFakeRPC()

		_, err := sendAndConfirmDurableNonce(context.Background(), rpc.NewWithCustomRPCClient(f), blockhashTransaction(t), "",
			rpc.TransactionOpts{}, nil, nil)

		require.ErrorIs(t, err, confirmation.ErrNotDurableNonceTransaction)
		require.Empty(t, f.callParams("sendTransaction"))
	})
}

// The lifetime can end with the transaction itself, so the signature status
// decides (kit #1234). These cases run for both lifetimes.
func TestSend_RechecksSignatureWhenLifetimeEnds(t *testing.T) {
	signatureLater := func(context.Context) error {
		time.Sleep(time.Second)
		return nil
	}
	for _, lt := range lifetimes() {
		for _, tt := range []struct {
			name      string
			status    func(context.Context, []any) (string, error)
			signature func(context.Context) error
			check     func(t *testing.T, err error)
		}{
			{
				name:      "resolves when the transaction landed at the commitment",
				status:    signatureStatus(`{"slot":1,"confirmations":null,"err":null,"confirmationStatus":"finalized"}`),
				signature: blockOnContext,
				check:     func(t *testing.T, err error) { require.NoError(t, err) },
			},
			{
				name:      "keeps waiting when the transaction landed below the commitment",
				status:    signatureStatus(`{"slot":1,"confirmations":3,"err":null,"confirmationStatus":"confirmed"}`),
				signature: signatureLater,
				check:     func(t *testing.T, err error) { require.NoError(t, err) },
			},
			{
				name:      "returns the lifetime error when the transaction is unknown",
				status:    signatureStatus("null"),
				signature: blockOnContext,
				check: func(t *testing.T, err error) {
					require.ErrorIs(t, err, lt.endedSentinel)
					require.Equal(t, lt.ended, err)
				},
			},
			{
				name:      "returns the transaction error when it landed and failed",
				status:    signatureStatus(`{"slot":1,"confirmations":null,"err":{"InstructionError":[0,"InvalidAccountData"]},"confirmationStatus":"finalized"}`),
				signature: blockOnContext,
				check: func(t *testing.T, err error) {
					var txErr *confirmation.TransactionError
					require.ErrorAs(t, err, &txErr)
					require.Equal(t, testSignature, txErr.Signature)
					require.Equal(t, map[string]any{"InstructionError": []any{float64(0), "InvalidAccountData"}}, txErr.Err)
				},
			},
			{
				name: "returns the lifetime error when the status request fails",
				status: func(context.Context, []any) (string, error) {
					return "", errors.New("RPC connection failed")
				},
				signature: blockOnContext,
				check:     func(t *testing.T, err error) { require.Equal(t, lt.ended, err) },
			},
			{
				name:      "keeps waiting when the transaction failed below the commitment",
				status:    signatureStatus(`{"slot":1,"confirmations":0,"err":{"InstructionError":[0,"InvalidAccountData"]},"confirmationStatus":"processed"}`),
				signature: signatureLater,
				check:     func(t *testing.T, err error) { require.NoError(t, err) },
			},
		} {
			t.Run(lt.name+"/"+tt.name, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					f := newFakeRPC()
					f.on("getSignatureStatuses", tt.status)
					s := &strategies{signature: tt.signature, lifetime: func(context.Context) error { return lt.ended }}

					_, err := lt.send(context.Background(), rpc.NewWithCustomRPCClient(f), lt.tx(t), rpc.CommitmentFinalized,
						rpc.TransactionOpts{}, s)

					tt.check(t, err)
					require.Equal(t, []any{[]solana.Signature{testSignature}}, f.callParams("getSignatureStatuses")[0])
				})
			})
		}

		t.Run(lt.name+"/passes other lifetime errors through", func(t *testing.T) {
			f := newFakeRPC()
			s := &strategies{signature: blockOnContext, lifetime: func(context.Context) error { return errStrategy }}

			_, err := lt.send(context.Background(), rpc.NewWithCustomRPCClient(f), lt.tx(t), "", rpc.TransactionOpts{}, s)

			require.ErrorIs(t, err, errStrategy)
			require.Empty(t, f.callParams("getSignatureStatuses"))
		})
	}
}

func airdrop(ctx context.Context, f *fakeRPC, commitment rpc.CommitmentType, s *strategies) (solana.Signature, error) {
	return requestAndConfirmAirdrop(ctx, rpc.NewWithCustomRPCClient(f), testRecipient, 1_000_000, commitment,
		s.signatureFunc(), s.timeoutFunc())
}

func newAirdropRPC() *fakeRPC {
	f := newFakeRPC()
	f.on("requestAirdrop", func(context.Context, []any) (string, error) {
		return fmt.Sprintf("%q", testSignature), nil
	})
	return f
}

func TestAirdrop_RequestsAndWaits(t *testing.T) {
	f := newAirdropRPC()
	s := &strategies{signature: succeed, lifetime: blockOnContext}

	sig, err := airdrop(context.Background(), f, rpc.CommitmentConfirmed, s)

	require.NoError(t, err)
	require.Equal(t, testSignature, sig)
	require.Equal(t, []any{testRecipient, uint64(1_000_000), rpc.M{"commitment": rpc.CommitmentConfirmed}},
		f.callParams("requestAirdrop")[0])
	require.Contains(t, s.recorded(), "signature "+testSignature.String()+" confirmed")
}

func TestAirdrop_TimeoutMatchesBothSentinels(t *testing.T) {
	f := newAirdropRPC()
	s := &strategies{signature: blockOnContext, lifetime: func(context.Context) error {
		return fmt.Errorf("%w after 1m0s", confirmation.ErrTimeout)
	}}

	sig, err := airdrop(context.Background(), f, "", s)

	require.ErrorIs(t, err, ErrTimeout)
	require.ErrorIs(t, err, confirmation.ErrTimeout)
	require.Equal(t, testSignature, sig)
	require.Contains(t, s.recorded(), "timeout finalized")
}

func TestAirdrop_RequestErrorStopsBeforeWaiting(t *testing.T) {
	f := newFakeRPC()
	f.on("requestAirdrop", func(context.Context, []any) (string, error) {
		return "", &jsonrpc.RPCError{Code: -32603, Message: "airdrop request failed"}
	})
	s := &strategies{signature: blockOnContext, lifetime: blockOnContext}

	_, err := airdrop(context.Background(), f, "", s)

	require.ErrorContains(t, err, "airdrop request failed")
	require.Empty(t, s.recorded())
}

func TestAirdrop_Canceled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		s := &strategies{signature: blockOnContext, lifetime: blockOnContext}
		done := make(chan error, 1)
		go func() {
			_, err := airdrop(ctx, newAirdropRPC(), rpc.CommitmentProcessed, s)
			done <- err
		}()

		synctest.Wait()
		cancel()

		require.ErrorIs(t, <-done, context.Canceled)
	})
}
