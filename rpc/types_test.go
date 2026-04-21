// Copyright 2021 github.com/gagliardetto
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
	stdjson "encoding/json"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
)

func TestData_base64_zstd(t *testing.T) {
	val := "KLUv/QQAWQAAaGVsbG8td29ybGTcLcaB"
	in := `["` + val + `", "base64+zstd"]`

	var data DataBytesOrJSON
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		[]byte("hello-world"),
		data.GetBinary(),
	)
	assert.Equal(t,
		solana.EncodingBase64Zstd,
		data.asDecodedBinary.Encoding,
	)
	assert.Equal(t,
		[]any{
			val,
			"base64+zstd",
		},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

func TestData_base64_zstd_empty(t *testing.T) {
	in := `["", "base64+zstd"]`

	var data DataBytesOrJSON
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		[]byte(""),
		data.GetBinary(),
	)
	assert.Equal(t,
		solana.EncodingBase64Zstd,
		data.asDecodedBinary.Encoding,
	)
	assert.Equal(t,
		[]any{
			"",
			"base64+zstd",
		},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

func TestData_jsonParsed(t *testing.T) {
	in := `{"hello":"world"}`

	var data DataBytesOrJSON
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		stdjson.RawMessage(in),
		data.GetRawJSON(),
	)
	assert.Equal(t,
		map[string]any{
			"hello": "world",
		},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

func TestData_jsonParsed_empty(t *testing.T) {
	in := `{}`

	var data DataBytesOrJSON
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		stdjson.RawMessage(in),
		data.GetRawJSON(),
	)
	assert.Equal(t,
		map[string]any{},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

func TestData_DataBytesOrJSONFromBytes(t *testing.T) {
	in := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	dataBytesOrJSON := DataBytesOrJSONFromBytes(in)
	out := dataBytesOrJSON.GetBinary()
	assert.Equal(t, in, out)
}

// TestParsedTransactionMeta_Decode guards issue #284: the jsonParsed
// and binary encodings return the same UiTransactionStatusMeta shape
// on the wire, so every field on TransactionMeta must also decode
// cleanly onto ParsedTransactionMeta. The fixture carries every trailing
// field that was historically dropped (status, rewards, loadedAddresses,
// returnData, computeUnitsConsumed) plus the parsed inner-instructions
// shape that distinguishes this path from the binary one.
func TestParsedTransactionMeta_Decode(t *testing.T) {
	in := []byte(`{
      "err": null,
      "fee": 5000,
      "preBalances": [1000000, 0],
      "postBalances": [994000, 1000],
      "innerInstructions": [{
        "index": 0,
        "instructions": [{
          "program": "system",
          "programId": "11111111111111111111111111111111",
          "parsed": {"type":"transfer","info":{"lamports":1}},
          "stackHeight": 1
        }]
      }],
      "preTokenBalances": [],
      "postTokenBalances": [],
      "logMessages": ["Program 11111111111111111111111111111111 invoke [1]"],
      "status": {"Ok": null},
      "rewards": [{
        "pubkey": "4ejjNYBbaETZyqaiK8aDj2BWER8LKHgDcCnRrPC22YGg",
        "lamports": 10,
        "postBalance": 1000010,
        "rewardType": "Fee"
      }],
      "loadedAddresses": {
        "writable": ["4ejjNYBbaETZyqaiK8aDj2BWER8LKHgDcCnRrPC22YGg"],
        "readonly": ["11111111111111111111111111111111"]
      },
      "returnData": {
        "programId": "11111111111111111111111111111111",
        "data": ["", "base64"]
      },
      "computeUnitsConsumed": 150
    }`)

	var got ParsedTransactionMeta
	assert.NoError(t, stdjson.Unmarshal(in, &got))

	assert.Equal(t, uint64(5000), got.Fee)
	assert.Len(t, got.InnerInstructions, 1)
	assert.Equal(t, "system", got.InnerInstructions[0].Instructions[0].Program)

	// Fields that were missing before the #284 fix — regression guards.
	assert.Len(t, got.Rewards, 1)
	assert.Equal(t, int64(10), got.Rewards[0].Lamports)
	assert.Len(t, got.LoadedAddresses.Writable, 1)
	assert.Len(t, got.LoadedAddresses.ReadOnly, 1)
	assert.Equal(t, solana.MustPublicKeyFromBase58("11111111111111111111111111111111"), got.ReturnData.ProgramId)
	if assert.NotNil(t, got.ComputeUnitsConsumed) {
		assert.Equal(t, uint64(150), *got.ComputeUnitsConsumed)
	}
}
