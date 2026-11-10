package confirmation

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	stdjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/require"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// The cases follow kit's confirmation-strategy-nonce-test.ts.

var (
	testNonceAccount = solana.MustPublicKeyFromBase58("9999999999999999999999999999999999999999999")
	testNonce        = solana.HashFromBytes(bytes.Repeat([]byte{4}, 32))
	otherNonce       = solana.HashFromBytes(bytes.Repeat([]byte{5}, 32))
)

// nonceAccountData returns the data of a nonce account that holds nonce. Only
// the nonce is filled in.
func nonceAccountData(nonce solana.Hash) []byte {
	data := make([]byte, 80)
	copy(data[nonceValueOffset:], nonce[:])
	return data
}

func accountJSON(data []byte) string {
	return fmt.Sprintf(`{"lamports":1447680,"owner":"11111111111111111111111111111111","data":[%q,"base64"],"executable":false,"rentEpoch":0,"space":%d}`,
		base64.StdEncoding.EncodeToString(data), len(data))
}

// accountInfo answers getAccountInfo with data, which is the requested slice.
func accountInfo(data []byte) rpcHandler {
	return result(`{"context":{"slot":1},"value":` + accountJSON(data) + `}`)
}

func (h *harness) startNonce(t *testing.T, commitment rpc.CommitmentType) (context.CancelFunc, <-chan error) {
	t.Helper()
	strategy := NewNonceInvalidation(h.rpcClient, h.wsClient)
	return run(t, func(ctx context.Context) error {
		return strategy(ctx, testNonceAccount, testNonce, commitment)
	})
}

func (h *harness) accountNotification(t *testing.T, sub subscribeRequest, data []byte) {
	t.Helper()
	h.validator.notify(t, "accountNotification", sub.SubID,
		stdjson.RawMessage(`{"context":{"slot":5},"value":`+accountJSON(data)+`}`))
}

func requireInvalidNonce(t *testing.T, err error, actual solana.Hash) {
	t.Helper()
	require.ErrorIs(t, err, ErrInvalidNonce)
	var invalid *InvalidNonceError
	require.ErrorAs(t, err, &invalid)
	require.Equal(t, InvalidNonceError{NonceAccount: testNonceAccount, Expected: testNonce, Actual: actual}, *invalid)
}

func TestNonce_CancelStopsAccountInfoRequest(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getAccountInfo", blockUntilCanceled(h.events, "getAccountInfo"))

	cancel, done := h.startNonce(t, "")
	h.events.waitFor(t, "rpc getAccountInfo")

	requireCanceled(t, cancel, done)
	h.events.waitFor(t, "rpc getAccountInfo canceled")
}

func TestNonce_CancelUnsubscribes(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getAccountInfo", blockUntilCanceled(h.events, "getAccountInfo"))

	cancel, done := h.startNonce(t, "")
	sub := h.validator.nextSubscribe(t)

	requireCanceled(t, cancel, done)
	h.events.waitFor(t, fmt.Sprintf("unsubscribe accountUnsubscribe [%d]", sub.SubID))
}

func TestNonce_SubscribesToNonceAccount(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getAccountInfo", blockUntilCanceled(h.events, "getAccountInfo"))

	cancel, done := h.startNonce(t, rpc.CommitmentConfirmed)
	sub := h.validator.nextSubscribe(t)

	require.Equal(t, "accountSubscribe", sub.Method)
	require.JSONEq(t, fmt.Sprintf(`[%q,{"commitment":"confirmed","encoding":"base64"}]`, testNonceAccount), sub.Params)
	requireCanceled(t, cancel, done)
}

func TestNonce_ReadsNonceOnlyAfterSubscribed(t *testing.T) {
	h := newHarness(t)
	h.validator.hold("accountSubscribe")
	h.rpc.on("getAccountInfo", blockUntilCanceled(h.events, "getAccountInfo"))

	cancel, done := h.startNonce(t, "")
	sub := h.validator.nextSubscribe(t)
	require.Never(t, func() bool { return h.events.count("rpc getAccountInfo") > 0 }, 100*time.Millisecond, testTick,
		"the nonce must not be read before the subscription is confirmed")

	h.validator.ack(sub)
	h.events.waitFor(t, "rpc getAccountInfo")
	h.events.requireBefore(t, "ack accountSubscribe", "rpc getAccountInfo")
	requireCanceled(t, cancel, done)
}

func TestNonce_ReadsNonceSlice(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getAccountInfo", blockUntilCanceled(h.events, "getAccountInfo"))

	cancel, done := h.startNonce(t, rpc.CommitmentProcessed)
	h.events.waitFor(t, "rpc getAccountInfo")

	params, err := stdjson.Marshal(h.rpc.callParams("getAccountInfo")[0])
	require.NoError(t, err)
	require.JSONEq(t, fmt.Sprintf(
		`[%q,{"commitment":"processed","dataSlice":{"length":32,"offset":40},"encoding":"base64"}]`, testNonceAccount),
		string(params))
	requireCanceled(t, cancel, done)
}

func TestNonce_PendsWhenReadNonceMatches(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getAccountInfo", accountInfo(testNonce[:]))

	cancel, done := h.startNonce(t, "")
	h.events.waitFor(t, "rpc getAccountInfo")

	requireCanceled(t, cancel, done)
}

func TestNonce_NonceAccountNotFound(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getAccountInfo", result(`{"context":{"slot":1},"value":null}`))

	_, done := h.startNonce(t, "")

	err := requireResult(t, done)
	require.ErrorIs(t, err, ErrNonceAccountNotFound)
	var notFound *NonceAccountNotFoundError
	require.ErrorAs(t, err, &notFound)
	require.Equal(t, testNonceAccount, notFound.NonceAccount)
}

func TestNonce_InvalidWhenReadNonceDiffers(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getAccountInfo", accountInfo(otherNonce[:]))

	_, done := h.startNonce(t, "")

	requireInvalidNonce(t, requireResult(t, done), otherNonce)
}

func TestNonce_PendsWhileNotifiedNonceMatches(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getAccountInfo", blockUntilCanceled(h.events, "getAccountInfo"))

	_, done := h.startNonce(t, "")
	sub := h.validator.nextSubscribe(t)
	h.accountNotification(t, sub, nonceAccountData(testNonce))
	requirePending(t, done)
	h.accountNotification(t, sub, nonceAccountData(otherNonce))

	// Only the second notification ends the wait.
	requireInvalidNonce(t, requireResult(t, done), otherNonce)
}

func TestNonce_InvalidWhenNotifiedNonceDiffers(t *testing.T) {
	h := newHarness(t)
	h.rpc.on("getAccountInfo", blockUntilCanceled(h.events, "getAccountInfo"))

	_, done := h.startNonce(t, "")
	sub := h.validator.nextSubscribe(t)
	h.accountNotification(t, sub, nonceAccountData(otherNonce))

	requireInvalidNonce(t, requireResult(t, done), otherNonce)
}

func TestNonce_ShortDataIsInvalid(t *testing.T) {
	t.Run("read", func(t *testing.T) {
		h := newHarness(t)
		h.rpc.on("getAccountInfo", accountInfo(testNonce[:16]))

		_, done := h.startNonce(t, "")

		requireInvalidNonce(t, requireResult(t, done), solana.Hash{})
	})
	t.Run("closed account notification", func(t *testing.T) {
		h := newHarness(t)
		h.rpc.on("getAccountInfo", blockUntilCanceled(h.events, "getAccountInfo"))

		_, done := h.startNonce(t, "")
		sub := h.validator.nextSubscribe(t)
		h.accountNotification(t, sub, nil)

		requireInvalidNonce(t, requireResult(t, done), solana.Hash{})
	})
}

func TestNonce_RejectsUnknownCommitment(t *testing.T) {
	h := newHarness(t)

	err := NewNonceInvalidation(h.rpcClient, h.wsClient)(context.Background(), testNonceAccount, testNonce, "max")

	require.ErrorContains(t, err, `unsupported commitment "max"`)
	require.Empty(t, h.events.list())
}
