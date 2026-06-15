package rpc

import (
	"context"
	"testing"

	"github.com/gagliardetto/solana-go"
	stdjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_GetProgramAccountsWithOpts_MinContextSlot pins that the
// minContextSlot opt is forwarded into the params object.
func TestClient_GetProgramAccountsWithOpts_MinContextSlot(t *testing.T) {
	responseBody := `[{"account":{"data":["dGVzdA==","base64"],"executable":true,"lamports":2039280,"owner":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","rentEpoch":206},"pubkey":"7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	minSlot := uint64(123456789)
	_, err := client.GetProgramAccountsWithOpts(
		context.Background(),
		pubKey,
		&GetProgramAccountsOpts{
			MinContextSlot: &minSlot,
		},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getProgramAccounts",
			"params": []any{
				pubkeyString,
				map[string]any{
					"encoding":       "base64",
					"minContextSlot": float64(minSlot),
				},
			},
		},
		reqBody,
	)
}

// TestClient_GetTokenAccountsByOwner_MinContextSlot pins that the
// minContextSlot opt is forwarded into the params object for the owner variant.
func TestClient_GetTokenAccountsByOwner_MinContextSlot(t *testing.T) {
	responseBody := `{"context":{"slot":1},"value":[]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	ownerString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	owner := solana.MustPublicKeyFromBase58(ownerString)
	programIDString := "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	programID := solana.MustPublicKeyFromBase58(programIDString)

	minSlot := uint64(987654321)
	_, err := client.GetTokenAccountsByOwner(
		context.Background(),
		owner,
		&GetTokenAccountsConfig{ProgramId: &programID},
		&GetTokenAccountsOpts{MinContextSlot: &minSlot},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTokenAccountsByOwner",
			"params": []any{
				ownerString,
				map[string]any{"programId": programIDString},
				map[string]any{
					"encoding":       "base64",
					"minContextSlot": float64(minSlot),
				},
			},
		},
		reqBody,
	)
}

// TestClient_GetTokenAccountsByDelegate_MinContextSlot pins the same for the
// delegate variant.
func TestClient_GetTokenAccountsByDelegate_MinContextSlot(t *testing.T) {
	responseBody := `{"context":{"slot":1},"value":[]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	delegateString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	delegate := solana.MustPublicKeyFromBase58(delegateString)
	programIDString := "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	programID := solana.MustPublicKeyFromBase58(programIDString)

	minSlot := uint64(42)
	_, err := client.GetTokenAccountsByDelegate(
		context.Background(),
		delegate,
		&GetTokenAccountsConfig{ProgramId: &programID},
		&GetTokenAccountsOpts{MinContextSlot: &minSlot},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTokenAccountsByDelegate",
			"params": []any{
				delegateString,
				map[string]any{"programId": programIDString},
				map[string]any{
					"encoding":       "base64",
					"minContextSlot": float64(minSlot),
				},
			},
		},
		reqBody,
	)
}

// TestClient_GetBalanceWithOpts_MinContextSlot pins minContextSlot
// forwarding on the new GetBalanceWithOpts variant. Closes the parity
// gap with getAccountInfo / getMultipleAccounts for the single-balance
// query path.
func TestClient_GetBalanceWithOpts_MinContextSlot(t *testing.T) {
	responseBody := `{"context":{"slot":1},"value":12345}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	minSlot := uint64(987654321)
	_, err := client.GetBalanceWithOpts(
		context.Background(),
		pubKey,
		&GetBalanceOpts{
			Commitment:     CommitmentFinalized,
			MinContextSlot: &minSlot,
		},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBalance",
			"params": []any{
				pubkeyString,
				map[string]any{
					"commitment":     "finalized",
					"minContextSlot": float64(minSlot),
				},
			},
		},
		reqBody,
	)
}

// TestClient_GetBalance_BackCompat_NoOpts pins that the original
// `GetBalance(ctx, pubkey, commitment)` signature still produces the
// same wire shape it always did when commitment is empty — the new
// WithOpts variant is purely additive.
func TestClient_GetBalance_BackCompat_NoOpts(t *testing.T) {
	responseBody := `{"context":{"slot":1},"value":12345}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	_, err := client.GetBalance(context.Background(), pubKey, "")
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBalance",
			"params": []any{
				pubkeyString,
			},
		},
		reqBody,
	)
}

// TestClient_GetLatestBlockhashWithOpts_MinContextSlot pins
// minContextSlot forwarding on getLatestBlockhash.
func TestClient_GetLatestBlockhashWithOpts_MinContextSlot(t *testing.T) {
	responseBody := `{"context":{"slot":1},"value":{"blockhash":"EkSnNWid2cvwEVnVx9aBqawnmiCNiDgp3gUdkDPTKN1N","lastValidBlockHeight":42}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	minSlot := uint64(777)
	_, err := client.GetLatestBlockhashWithOpts(
		context.Background(),
		&GetLatestBlockhashOpts{
			Commitment:     CommitmentConfirmed,
			MinContextSlot: &minSlot,
		},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getLatestBlockhash",
			"params": []any{
				map[string]any{
					"commitment":     "confirmed",
					"minContextSlot": float64(minSlot),
				},
			},
		},
		reqBody,
	)
}

// TestClient_GetSlotWithOpts_MinContextSlot pins minContextSlot
// forwarding on getSlot.
func TestClient_GetSlotWithOpts_MinContextSlot(t *testing.T) {
	responseBody := `123`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	minSlot := uint64(555)
	_, err := client.GetSlotWithOpts(
		context.Background(),
		&GetSlotOpts{MinContextSlot: &minSlot},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getSlot",
			"params": []any{
				map[string]any{
					"minContextSlot": float64(minSlot),
				},
			},
		},
		reqBody,
	)
}

// TestClient_GetTokenAccountBalanceWithOpts_MinContextSlot pins
// minContextSlot forwarding on getTokenAccountBalance.
func TestClient_GetTokenAccountBalanceWithOpts_MinContextSlot(t *testing.T) {
	responseBody := `{"context":{"slot":1},"value":{"amount":"100","decimals":6,"uiAmount":0.0001,"uiAmountString":"0.0001"}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	minSlot := uint64(99)
	_, err := client.GetTokenAccountBalanceWithOpts(
		context.Background(),
		pubKey,
		&GetTokenAccountBalanceOpts{MinContextSlot: &minSlot},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTokenAccountBalance",
			"params": []any{
				pubkeyString,
				map[string]any{
					"minContextSlot": float64(minSlot),
				},
			},
		},
		reqBody,
	)
}

// TestClient_GetLatestBlockhash_BackCompat_NoOpts pins that the legacy
// GetLatestBlockhash wrapper emits no params object when called with an
// empty commitment, matching the pre-WithOpts wire shape.
func TestClient_GetLatestBlockhash_BackCompat_NoOpts(t *testing.T) {
	responseBody := `{"context":{"slot":1},"value":{"blockhash":"EkSnNWid2cvwEVnVx9aBqawnmiCNiDgp3gUdkDPTKN1N","lastValidBlockHeight":42}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	_, err := client.GetLatestBlockhash(context.Background(), "")
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getLatestBlockhash",
			"params":  []any{},
		},
		reqBody,
	)
}

// TestClient_GetSlot_BackCompat_NoOpts pins that the legacy GetSlot
// wrapper emits no params object when called with an empty commitment,
// matching the pre-WithOpts wire shape.
func TestClient_GetSlot_BackCompat_NoOpts(t *testing.T) {
	responseBody := `123`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	_, err := client.GetSlot(context.Background(), "")
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getSlot",
			"params":  []any{},
		},
		reqBody,
	)
}

// TestClient_GetTokenAccountBalance_BackCompat_NoOpts pins that the legacy
// GetTokenAccountBalance wrapper sends only the account pubkey when called
// with an empty commitment, matching the pre-WithOpts wire shape.
func TestClient_GetTokenAccountBalance_BackCompat_NoOpts(t *testing.T) {
	responseBody := `{"context":{"slot":1},"value":{"amount":"100","decimals":6,"uiAmount":0.0001,"uiAmountString":"0.0001"}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	_, err := client.GetTokenAccountBalance(context.Background(), pubKey, "")
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTokenAccountBalance",
			"params": []any{
				pubkeyString,
			},
		},
		reqBody,
	)
}
