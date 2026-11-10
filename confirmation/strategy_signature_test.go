package confirmation

import (
	"context"
	"fmt"
	"testing"
	"time"

	stdjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/require"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// The cases follow kit's confirmation-strategy-signature-test.ts.

var testSignature = solana.MustSignatureFromBase58(
	"5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW")

// signatureStatus answers getSignatureStatuses with one status, given as raw
// JSON or null.
func signatureStatus(status string) rpcHandler {
	return result(`{"context":{"slot":1},"value":[` + status + `]}`)
}

func (h *harness) startSignature(t *testing.T, commitment rpc.CommitmentType) (context.CancelFunc, <-chan error) {
	t.Helper()
	strategy := NewRecentSignatureConfirmation(h.rpcClient, h.wsClient)
	return run(t, func(ctx context.Context) error {
		return strategy(ctx, testSignature, commitment)
	})
}

func (h *harness) signatureNotification(t *testing.T, sub subscribeRequest, value string) {
	t.Helper()
	h.validator.notify(t, "signatureNotification", sub.SubID,
		stdjson.RawMessage(`{"context":{"slot":5},"value":`+value+`}`))
}

func requireTransactionError(t *testing.T, err error, want any) {
	t.Helper()
	var txErr *TransactionError
	require.ErrorAs(t, err, &txErr)
	require.Equal(t, testSignature, txErr.Signature)
	require.Equal(t, want, txErr.Err)
}

func TestSignature_SubscribesToSignature(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getSignatureStatuses", blockUntilCanceled(h.events, "getSignatureStatuses"))

	cancel, done := h.startSignature(t, rpc.CommitmentFinalized)
	sub := h.validator.nextSubscribe(t)

	require.Equal(t, "signatureSubscribe", sub.Method)
	require.JSONEq(t, fmt.Sprintf(`[%q,{"commitment":"finalized"}]`, testSignature), sub.Params)
	requireCanceled(t, cancel, done)
}

func TestSignature_ReadsStatusOnlyAfterSubscribed(t *testing.T) {
	h := newHarness(t)
	h.validator.hold("signatureSubscribe")
	h.rpc.on("getSignatureStatuses", blockUntilCanceled(h.events, "getSignatureStatuses"))

	cancel, done := h.startSignature(t, "")
	sub := h.validator.nextSubscribe(t)
	require.Never(t, func() bool { return h.events.count("rpc getSignatureStatuses") > 0 }, 100*time.Millisecond, testTick,
		"the status must not be read before the subscription is confirmed")

	h.validator.ack(sub)
	h.events.waitFor(t, "rpc getSignatureStatuses")
	h.events.requireBefore(t, "ack signatureSubscribe", "rpc getSignatureStatuses")
	require.Equal(t, []any{[]solana.Signature{testSignature}}, h.rpc.callParams("getSignatureStatuses")[0])
	requireCanceled(t, cancel, done)
}

func TestSignature_PendsBelowCommitment(t *testing.T) {
	for _, achieved := range []string{"processed", "confirmed"} {
		t.Run(achieved, func(t *testing.T) {
			h := newHarness(t)
			h.rpc.on("getSignatureStatuses", signatureStatus(
				`{"slot":1,"confirmations":0,"err":null,"confirmationStatus":"`+achieved+`"}`))

			cancel, done := h.startSignature(t, rpc.CommitmentFinalized)
			h.events.waitFor(t, "rpc getSignatureStatuses")

			requireCanceled(t, cancel, done)
		})
	}
}

func TestSignature_PendsWithoutStatus(t *testing.T) {
	for name, reply := range map[string]string{
		"null status":       `{"context":{"slot":1},"value":[null]}`,
		"null value":        `{"context":{"slot":1},"value":null}`,
		"empty value":       `{"context":{"slot":1},"value":[]}`,
		"no status level":   `{"context":{"slot":1},"value":[{"slot":1,"confirmations":null,"err":null}]}`,
		"unknown level":     `{"context":{"slot":1},"value":[{"slot":1,"confirmations":null,"err":null,"confirmationStatus":"max"}]}`,
		"null result":       `null`,
		"status is pending": `{"context":{"slot":1},"value":[{"slot":1,"confirmations":0,"err":null,"confirmationStatus":"processed"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			h.rpc.on("getSignatureStatuses", result(reply))

			cancel, done := h.startSignature(t, rpc.CommitmentConfirmed)
			h.events.waitFor(t, "rpc getSignatureStatuses")

			requireCanceled(t, cancel, done)
		})
	}
}

func TestSignature_ResolvesAtCommitment(t *testing.T) {
	for _, tt := range []struct {
		commitment rpc.CommitmentType
		achieved   string
	}{
		{rpc.CommitmentFinalized, "finalized"},
		{rpc.CommitmentConfirmed, "finalized"},
		{rpc.CommitmentProcessed, "confirmed"},
		{"", "finalized"},
	} {
		t.Run(string(tt.commitment)+"/"+tt.achieved, func(t *testing.T) {
			h := newHarness(t)
			h.rpc.on("getSignatureStatuses", signatureStatus(
				`{"slot":1,"confirmations":null,"err":null,"confirmationStatus":"`+tt.achieved+`"}`))

			_, done := h.startSignature(t, tt.commitment)

			require.NoError(t, requireResult(t, done))
		})
	}
}

func TestSignature_FailsOnStatusError(t *testing.T) {
	// Kit #1003: a failed transaction is reported at any commitment level.
	for _, achieved := range []string{"processed", "finalized"} {
		t.Run(achieved, func(t *testing.T) {
			h := newHarness(t)
			h.rpc.on("getSignatureStatuses", signatureStatus(
				`{"slot":1,"confirmations":null,"err":{"InstructionError":[0,{"Custom":1}]},"confirmationStatus":"`+achieved+`"}`))

			_, done := h.startSignature(t, rpc.CommitmentFinalized)

			requireTransactionError(t, requireResult(t, done),
				map[string]any{"InstructionError": []any{float64(0), map[string]any{"Custom": float64(1)}}})
		})
	}
}

func TestSignature_ResolvesOnNotification(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getSignatureStatuses", blockUntilCanceled(h.events, "getSignatureStatuses"))

	_, done := h.startSignature(t, "")
	sub := h.validator.nextSubscribe(t)
	h.signatureNotification(t, sub, `{"err":null}`)

	require.NoError(t, requireResult(t, done))
}

func TestSignature_FailsOnNotificationError(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getSignatureStatuses", blockUntilCanceled(h.events, "getSignatureStatuses"))

	_, done := h.startSignature(t, "")
	sub := h.validator.nextSubscribe(t)
	h.signatureNotification(t, sub, `{"err":"o no"}`)

	requireTransactionError(t, requireResult(t, done), "o no")
}

// The strategy does not ask for "received" notifications, but one must never
// count as a confirmation.
func TestSignature_IgnoresReceivedNotification(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getSignatureStatuses", blockUntilCanceled(h.events, "getSignatureStatuses"))

	_, done := h.startSignature(t, "")
	sub := h.validator.nextSubscribe(t)
	h.signatureNotification(t, sub, `"receivedSignature"`)
	requirePending(t, done)
	h.signatureNotification(t, sub, `{"err":"o no"}`)

	requireTransactionError(t, requireResult(t, done), "o no")
}

func TestSignature_StatusRequestError(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getSignatureStatuses", failWith(fmt.Errorf("o no")))

	_, done := h.startSignature(t, "")

	require.ErrorContains(t, requireResult(t, done), "o no")
}

func TestSignature_CancelStopsStatusRequest(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getSignatureStatuses", blockUntilCanceled(h.events, "getSignatureStatuses"))

	cancel, done := h.startSignature(t, "")
	h.events.waitFor(t, "rpc getSignatureStatuses")

	requireCanceled(t, cancel, done)
	h.events.waitFor(t, "rpc getSignatureStatuses canceled")
}

func TestSignature_CancelUnsubscribes(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getSignatureStatuses", blockUntilCanceled(h.events, "getSignatureStatuses"))

	cancel, done := h.startSignature(t, "")
	sub := h.validator.nextSubscribe(t)

	requireCanceled(t, cancel, done)
	h.events.waitFor(t, fmt.Sprintf("unsubscribe signatureUnsubscribe [%d]", sub.SubID))
}

func TestSignature_RejectsUnknownCommitment(t *testing.T) {
	h := newHarness(t)

	err := NewRecentSignatureConfirmation(h.rpcClient, h.wsClient)(context.Background(), testSignature, "singleGossip")

	require.ErrorContains(t, err, `unsupported commitment "singleGossip"`)
	require.Empty(t, h.events.list())
}
