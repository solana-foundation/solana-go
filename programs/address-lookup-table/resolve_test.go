// Copyright 2026 github.com/solana-foundation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package addresslookuptable

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/stretchr/testify/require"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// rpcResponseFunc is invoked per request with the parsed JSON-RPC body and
// returns the result fragment that will be wrapped into a JSON-RPC envelope.
type rpcResponseFunc func(method string, requestBody []byte) string

func newRPCMock(t *testing.T, fn rpcResponseFunc) (*httptest.Server, *rpc.Client) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		method := ""
		if i := bytes.Index(body, []byte(`"method":"`)); i >= 0 {
			rest := body[i+len(`"method":"`):]
			if j := bytes.IndexByte(rest, '"'); j >= 0 {
				method = string(rest[:j])
			}
		}
		result := fn(method, body)
		fmt.Fprintf(rw, `{"jsonrpc":"2.0","result":%s,"id":0}`, result)
	}))
	t.Cleanup(srv.Close)
	return srv, rpc.New(srv.URL)
}

// encodeTableAccountData produces the base64 JSON value the RPC layer
// expects for a getMultipleAccounts entry whose data is an
// AddressLookupTableState.
func encodeTableAccountData(t *testing.T, state AddressLookupTableState) string {
	t.Helper()
	buf := new(bytes.Buffer)
	enc := bin.NewBinEncoder(buf)
	require.NoError(t, state.MarshalWithEncoder(enc))
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func versionedMessageWithLookups(t *testing.T, table solana.PublicKey, writable, readonly []uint8) *solana.Message {
	t.Helper()
	msg, err := (&solana.Message{}).SetVersion(solana.MessageVersionV0)
	require.NoError(t, err)
	msg.AccountKeys = solana.PublicKeySlice{solana.NewWallet().PublicKey()} // placeholder static key
	msg.SetAddressTableLookups([]solana.MessageAddressTableLookup{
		{
			AccountKey:      table,
			WritableIndexes: writable,
			ReadonlyIndexes: readonly,
		},
	})
	// SetVersion is idempotent on V0.
	_, err = msg.SetVersion(solana.MessageVersionV0)
	require.NoError(t, err)
	return msg
}

// TestResolveMessageLookupsFromRPC_NoLookups pins that a legacy (or
// versioned-without-lookups) message is a no-op: ResolveMessageLookupsFromRPC
// returns nil, never calls the RPC client, and never mutates AccountKeys.
func TestResolveMessageLookupsFromRPC_NoLookups(t *testing.T) {
	rpcCalled := false
	srv, client := newRPCMock(t, func(string, []byte) string {
		rpcCalled = true
		return `null`
	})
	_ = srv

	msg := &solana.Message{}
	preLen := len(msg.AccountKeys)

	require.NoError(t, ResolveMessageLookupsFromRPC(context.Background(), client, msg))
	require.False(t, rpcCalled, "no RPC call should be made for a legacy message")
	require.Equal(t, preLen, len(msg.AccountKeys), "AccountKeys must not be mutated")
}

// TestResolveMessageLookupsFromRPC_Happy pins the end-to-end resolution
// path: a V0 message referencing a single ALT, the RPC returns the encoded
// table data, and after the call msg.IsResolved() reports true and the
// table's writable+readonly addresses appear in AccountKeys in the
// canonical order (writable first, then readonly).
func TestResolveMessageLookupsFromRPC_Happy(t *testing.T) {
	tableID := solana.NewWallet().PublicKey()
	writableAddr0 := solana.NewWallet().PublicKey()
	writableAddr1 := solana.NewWallet().PublicKey()
	readonlyAddr0 := solana.NewWallet().PublicKey()

	tableState := AddressLookupTableState{
		TypeIndex:                  1,
		DeactivationSlot:           math.MaxUint64,
		LastExtendedSlot:           42,
		LastExtendedSlotStartIndex: 0,
		Addresses: solana.PublicKeySlice{
			writableAddr0,
			writableAddr1,
			readonlyAddr0,
		},
	}
	dataBase64 := encodeTableAccountData(t, tableState)
	// Solana reports `space` as the decoded byte count, not the base64 length.
	rawData, err := base64.StdEncoding.DecodeString(dataBase64)
	require.NoError(t, err)
	space := len(rawData)

	_, client := newRPCMock(t, func(method string, body []byte) string {
		require.Equal(t, "getMultipleAccounts", method)
		require.Contains(t, string(body), tableID.String())
		return fmt.Sprintf(`{"context":{"slot":1},"value":[{"data":["%s","base64"],"executable":false,"lamports":1,"owner":"AddressLookupTab1e1111111111111111111111111","rentEpoch":0,"space":%d}]}`, dataBase64, space)
	})

	msg := versionedMessageWithLookups(t, tableID, []uint8{0, 1}, []uint8{2})
	require.False(t, msg.IsResolved())
	preStaticLen := len(msg.AccountKeys)

	require.NoError(t, ResolveMessageLookupsFromRPC(context.Background(), client, msg))
	require.True(t, msg.IsResolved())

	// Writable lookups append first, then readonly. Indices reference the
	// table's Addresses slice, so writable={0,1} -> {writableAddr0, writableAddr1}
	// and readonly={2} -> {readonlyAddr0}.
	require.Equal(t, preStaticLen+3, len(msg.AccountKeys))
	require.Equal(t, writableAddr0, msg.AccountKeys[preStaticLen])
	require.Equal(t, writableAddr1, msg.AccountKeys[preStaticLen+1])
	require.Equal(t, readonlyAddr0, msg.AccountKeys[preStaticLen+2])
}

// TestResolveMessageLookupsFromRPC_TableNotFound pins that a nil entry in
// the getMultipleAccounts response surfaces as an error naming the missing
// table, instead of silently producing a half-resolved message.
func TestResolveMessageLookupsFromRPC_TableNotFound(t *testing.T) {
	tableID := solana.NewWallet().PublicKey()
	_, client := newRPCMock(t, func(method string, body []byte) string {
		return `{"context":{"slot":1},"value":[null]}`
	})

	msg := versionedMessageWithLookups(t, tableID, []uint8{0}, nil)
	err := ResolveMessageLookupsFromRPC(context.Background(), client, msg)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), tableID.String()), "error should name the missing table %s, got %v", tableID, err)
	require.False(t, msg.IsResolved(), "message must not be marked resolved on partial fetch")
}

// TestResolveMessageLookupsFromRPC_NilArgs pins explicit argument validation
// so a misuse fails loudly rather than calling .GetAddressTableLookups on a
// nil receiver.
func TestResolveMessageLookupsFromRPC_NilArgs(t *testing.T) {
	_, client := newRPCMock(t, func(string, []byte) string { return `null` })

	require.Error(t, ResolveMessageLookupsFromRPC(context.Background(), client, nil))

	msg := &solana.Message{}
	require.Error(t, ResolveMessageLookupsFromRPC(context.Background(), nil, msg))
}
