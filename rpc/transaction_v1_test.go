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

package rpc

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/gagliardetto/solana-go"
	stdjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// V1 (SIMD-0385) transaction produced by the Rust SDK; same vector as
// `v1GoldenTxFeeHeap` in the root package.
const v1GoldenTxFeeHeapHex = "8101000113000000070707070707070707070707070707070707070707070707070707070707070701028a88e3dd7409f195fd52db2d3cba5d72ca6709bf1d94121bf3748801b40f6f5c1010101010101010101010101010101010101010101010101010101010101010010000000000000000800000010100000070083fb1bf5a9e2c90c6903b7328529cdf59e04e2fcf29c718bea8c76fd325f4b67e18feb1f95f9aab171140b81b8c8238c9ffea33d4133bae92af68c6453800"

// A base64 getTransaction envelope holding a v1 tx decodes as v1.
func TestTransactionResultEnvelope_V1Base64(t *testing.T) {
	raw, err := hex.DecodeString(v1GoldenTxFeeHeapHex)
	require.NoError(t, err)
	in := `["` + base64.StdEncoding.EncodeToString(raw) + `","base64"]`

	var env TransactionResultEnvelope
	require.NoError(t, env.UnmarshalJSON([]byte(in)))
	assert.Equal(t, raw, env.GetBinary())

	tx, err := env.GetTransaction()
	require.NoError(t, err)
	require.NotNil(t, tx)
	assert.Equal(t, solana.MessageVersionV1, tx.Message.GetVersion())
	require.NotNil(t, tx.Message.TransactionConfig.PriorityFee)
	assert.Equal(t, uint64(1), *tx.Message.TransactionConfig.PriorityFee)
	require.NotNil(t, tx.Message.TransactionConfig.HeapSize)
	assert.Equal(t, uint32(32768), *tx.Message.TransactionConfig.HeapSize)
	require.NoError(t, tx.VerifySignatures())

	// TransactionWithMeta binary path.
	twm := TransactionWithMeta{}
	require.NoError(t, stdjson.Unmarshal([]byte(`{"transaction":`+in+`,"version":1}`), &twm))
	assert.Equal(t, TransactionVersion(1), twm.Version)
	tx2, err := twm.GetTransaction()
	require.NoError(t, err)
	assert.Equal(t, solana.MessageVersionV1, tx2.Message.GetVersion())
}

// The `json` encoding of a V1 transaction (agave UiRawMessage) carries
// `transactionConfig` and no `addressTableLookups`.
func TestTransactionResultEnvelope_V1JSON(t *testing.T) {
	in := `{
		"signatures": ["3Euy3SsYzpQ4EvwX4qvVui1YcNAwdDGinuTVhMxvy92NzWDYnNMeacmv4wEmPJndGz6tsX7VSLHFJFawFzQh12b9"],
		"message": {
			"header": {"numRequiredSignatures":1,"numReadonlySignedAccounts":0,"numReadonlyUnsignedAccounts":1},
			"accountKeys": ["AKnL4NNf3DGWZJS6cPknBuEGnVsV4A4m5tgebLHaRSZ9","25hjHpTATmkdET17ynDhf1MCuYNDn1z7wXfVw5iaxLAK"],
			"recentBlockhash": "US517G5965aydkZ46HS38QLi7UQiSojurfbQfKCELFx",
			"instructions": [{"programIdIndex":1,"accounts":[0],"data":"","stackHeight":null}],
			"transactionConfig": {"priorityFee":1,"computeUnitLimit":null,"loadedAccountsDataSizeLimit":null,"heapSize":32768}
		}
	}`
	var env TransactionResultEnvelope
	require.NoError(t, env.UnmarshalJSON([]byte(in)))
	tx, err := env.GetTransaction()
	require.NoError(t, err)
	require.NotNil(t, tx)
	assert.Equal(t, solana.MessageVersionV1, tx.Message.GetVersion())
	assert.Nil(t, tx.Message.AddressTableLookups)
	require.NotNil(t, tx.Message.TransactionConfig.PriorityFee)
	assert.Equal(t, uint64(1), *tx.Message.TransactionConfig.PriorityFee)
	assert.Nil(t, tx.Message.TransactionConfig.ComputeUnitLimit)
	require.NotNil(t, tx.Message.TransactionConfig.HeapSize)
	assert.Equal(t, uint32(32768), *tx.Message.TransactionConfig.HeapSize)

	// Same tx as the binary golden vector: re-encodes identically, signature verifies.
	raw, err := hex.DecodeString(v1GoldenTxFeeHeapHex)
	require.NoError(t, err)
	again, err := tx.MarshalBinary()
	require.NoError(t, err)
	assert.Equal(t, raw, again)
	require.NoError(t, tx.VerifySignatures())
}

// jsonParsed: ParsedMessage exposes transactionConfig for V1 transactions
// and leaves it nil otherwise.
func TestParsedMessage_TransactionConfig(t *testing.T) {
	v1 := `{
		"accountKeys": [{"pubkey":"AKnL4NNf3DGWZJS6cPknBuEGnVsV4A4m5tgebLHaRSZ9","signer":true,"writable":true,"source":"transaction"}],
		"recentBlockhash": "US517G5965aydkZ46HS38QLi7UQiSojurfbQfKCELFx",
		"instructions": [],
		"transactionConfig": {"priorityFee":5000,"computeUnitLimit":200000,"loadedAccountsDataSizeLimit":null,"heapSize":null}
	}`
	var pm ParsedMessage
	require.NoError(t, stdjson.Unmarshal([]byte(v1), &pm))
	require.NotNil(t, pm.TransactionConfig)
	assert.Equal(t, uint64(5000), *pm.TransactionConfig.PriorityFee)
	assert.Equal(t, uint32(200000), *pm.TransactionConfig.ComputeUnitLimit)
	assert.Nil(t, pm.TransactionConfig.LoadedAccountsDataSizeLimit)
	assert.Nil(t, pm.TransactionConfig.HeapSize)

	legacy := `{"accountKeys":[],"recentBlockhash":"US517G5965aydkZ46HS38QLi7UQiSojurfbQfKCELFx","instructions":[]}`
	var pm2 ParsedMessage
	require.NoError(t, stdjson.Unmarshal([]byte(legacy), &pm2))
	assert.Nil(t, pm2.TransactionConfig)

	// Round-trip keeps the config, and omits it when nil.
	out, err := stdjson.Marshal(pm)
	require.NoError(t, err)
	assert.Contains(t, string(out), `"transactionConfig":{"priorityFee":5000`)
	out2, err := stdjson.Marshal(pm2)
	require.NoError(t, err)
	assert.NotContains(t, string(out2), "transactionConfig")
}
