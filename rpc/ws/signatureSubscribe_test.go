// Copyright 2026 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ws

import (
	"testing"

	stdjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
)

// TestSignatureResultUnmarshalStatus covers the default notification shape:
// `{ "value": { "err": null } }` for a successful transaction.
func TestSignatureResultUnmarshalStatus(t *testing.T) {
	var res SignatureResult
	require.NoError(t, stdjson.Unmarshal([]byte(`{
		"context": {"slot": 42},
		"value":   {"err": null}
	}`), &res))
	require.Equal(t, uint64(42), res.Context.Slot)
	require.Nil(t, res.Value.Err)
	require.False(t, res.ReceivedSignature)
}

// TestSignatureResultUnmarshalStatusWithErr covers the failed-tx branch
// where `err` is a non-null object.
func TestSignatureResultUnmarshalStatusWithErr(t *testing.T) {
	var res SignatureResult
	require.NoError(t, stdjson.Unmarshal([]byte(`{
		"context": {"slot": 42},
		"value":   {"err": {"InstructionError": [0, "InvalidAccountData"]}}
	}`), &res))
	require.NotNil(t, res.Value.Err)
	require.False(t, res.ReceivedSignature)
}

// TestSignatureResultUnmarshalReceived covers the second notification
// shape introduced by EnableReceivedNotification: `"value":
// "receivedSignature"`. Without the custom unmarshaler the default
// decoder would fail because the field is typed as a struct.
func TestSignatureResultUnmarshalReceived(t *testing.T) {
	var res SignatureResult
	require.NoError(t, stdjson.Unmarshal([]byte(`{
		"context": {"slot": 7},
		"value":   "receivedSignature"
	}`), &res))
	require.Equal(t, uint64(7), res.Context.Slot)
	require.True(t, res.ReceivedSignature)
	require.Nil(t, res.Value.Err)
}

// TestSignatureResultUnmarshalUnknownMarker rejects unexpected string
// markers so a future RPC change surfaces as a decode error rather
// than a silent miscategorisation.
func TestSignatureResultUnmarshalUnknownMarker(t *testing.T) {
	var res SignatureResult
	err := res.UnmarshalJSON([]byte(`{"value":"someUnknownMarker"}`))
	require.Error(t, err)
}

// TestSignatureResultUnmarshalNullValue treats a null value as a no-op
// (rather than an error) so the field can be omitted by an upstream RPC
// without breaking notification dispatch.
func TestSignatureResultUnmarshalNullValue(t *testing.T) {
	var res SignatureResult
	require.NoError(t, res.UnmarshalJSON([]byte(`{"context":{"slot":1},"value":null}`)))
	require.Equal(t, uint64(1), res.Context.Slot)
	require.False(t, res.ReceivedSignature)
	require.Nil(t, res.Value.Err)
}
