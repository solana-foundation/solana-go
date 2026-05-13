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
