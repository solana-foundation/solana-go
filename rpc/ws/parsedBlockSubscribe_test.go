// Copyright 2022 github.com/gagliardetto
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

	"github.com/stretchr/testify/require"
)

// parsedBlockFrame is the shared jsonParsed blockNotification fixture
// exercised by the tests below. It reproduces the shape originally
// reported in issue #291.
const parsedBlockFrame = `{
  "jsonrpc":"2.0",
  "method":"blockNotification",
  "params":{
    "subscription":1,
    "result":{
      "context":{"slot":1234567},
      "value":{
        "slot":1234567,
        "err":null,
        "block":{
          "blockhash":"6TScP1N3f4n23Y5f1cZ1YmLMgwYyb6PTvQQjoNn4XDFz",
          "previousBlockhash":"11111111111111111111111111111111",
          "parentSlot":1234566,
          "blockTime":1700000000,
          "blockHeight":123456,
          "transactions":[{
            "transaction":{
              "signatures":["5j7s1QzqC6rA2FvR2gSzRcbfh4PE9eU7mVnxvKnZtWaAD9vqvR2rB8g4SfQ4ZBqS8PyZt7aX8ybX42kVhbZu8P7w"],
              "message":{
                "accountKeys":[
                  {"pubkey":"4ejjNYBbaETZyqaiK8aDj2BWER8LKHgDcCnRrPC22YGg","signer":true,"writable":true},
                  {"pubkey":"11111111111111111111111111111111","signer":false,"writable":false}
                ],
                "recentBlockhash":"6TScP1N3f4n23Y5f1cZ1YmLMgwYyb6PTvQQjoNn4XDFz",
                "instructions":[{
                  "program":"system",
                  "programId":"11111111111111111111111111111111",
                  "parsed":{"type":"transfer","info":{"source":"4ejjNYBbaETZyqaiK8aDj2BWER8LKHgDcCnRrPC22YGg","destination":"11111111111111111111111111111111","lamports":1000}},
                  "stackHeight":0
                }]
              }
            },
            "meta":{
              "err":null,
              "fee":5000,
              "preBalances":[1000000,0],
              "postBalances":[994000,1000],
              "innerInstructions":[],
              "logMessages":["Program 11111111111111111111111111111111 invoke [1]","Program 11111111111111111111111111111111 success"],
              "preTokenBalances":[],
              "postTokenBalances":[]
            }
          }]
        }
      }
    }
  }
}`

// TestParsedBlockResult_Decode verifies that the (deprecated) dedicated
// ParsedBlockSubscribe path still correctly decodes jsonParsed frames —
// the decode failure regression tracked by issue #291.
func TestParsedBlockResult_Decode(t *testing.T) {
	var got ParsedBlockResult
	err := decodeResponseFromMessage([]byte(parsedBlockFrame), &got)
	require.NoError(t, err, "jsonParsed block must decode into ParsedBlockResult (regression for #291)")

	require.Equal(t, uint64(1234567), got.Value.Slot)
	require.NotNil(t, got.Value.Block)
	require.Equal(t, uint64(1234566), got.Value.Block.ParentSlot)
	require.Len(t, got.Value.Block.Transactions, 1)

	tx := got.Value.Block.Transactions[0]
	require.NotNil(t, tx.Transaction)
	require.Len(t, tx.Transaction.Signatures, 1)
	require.Len(t, tx.Transaction.Message.Instructions, 1)

	ix := tx.Transaction.Message.Instructions[0]
	require.Equal(t, "system", ix.Program)
	require.NotNil(t, ix.Parsed)
	require.NotNil(t, tx.Meta)
	require.Equal(t, uint64(5000), tx.Meta.Fee)
}

// TestBlockSubscribe_ParsedRoute verifies that BlockSubscribe now routes
// jsonParsed frames into BlockResultValue.ParsedBlock instead of
// erroring out. This is the unified-encoding path that closes #291 at
// the API level — callers no longer need to branch between BlockSubscribe
// and ParsedBlockSubscribe.
func TestBlockSubscribe_ParsedRoute(t *testing.T) {
	decoder := decodeBlockNotification(true)
	out, err := decoder([]byte(parsedBlockFrame))
	require.NoError(t, err)
	got := out.(*BlockResult)

	require.Equal(t, uint64(1234567), got.Value.Slot)
	require.Nil(t, got.Value.Block, "binary path must stay nil when jsonParsed was requested")
	require.NotNil(t, got.Value.ParsedBlock)
	require.Equal(t, uint64(1234566), got.Value.ParsedBlock.ParentSlot)
	require.Len(t, got.Value.ParsedBlock.Transactions, 1)
	require.Equal(t, "system", got.Value.ParsedBlock.Transactions[0].Transaction.Message.Instructions[0].Program)
}

// TestBlockSubscribe_BinaryRoute verifies the base64 path decodes into
// BlockResultValue.Block (with ParsedBlock nil) so existing callers
// don't regress. Also covers the "Encoding unset" default path, which
// falls into the same isParsed=false branch. The transaction uses the
// [data, "base64"] 2-array shape — the distinguishing feature of the
// binary encoding — so this test would fail if the decoder were
// accidentally rewired to expect the jsonParsed object shape.
func TestBlockSubscribe_BinaryRoute(t *testing.T) {
	frame := []byte(`{
      "jsonrpc":"2.0",
      "method":"blockNotification",
      "params":{
        "subscription":1,
        "result":{
          "context":{"slot":9},
          "value":{
            "slot":9,
            "err":null,
            "block":{
              "blockhash":"6TScP1N3f4n23Y5f1cZ1YmLMgwYyb6PTvQQjoNn4XDFz",
              "previousBlockhash":"11111111111111111111111111111111",
              "parentSlot":8,
              "blockTime":1700000000,
              "blockHeight":10,
              "transactions":[{
                "transaction":["AQAAAAAAAAA=","base64"],
                "meta":{
                  "err":null,
                  "fee":5000,
                  "preBalances":[1000000,0],
                  "postBalances":[995000,0],
                  "innerInstructions":[],
                  "logMessages":["Program log: noop"],
                  "preTokenBalances":[],
                  "postTokenBalances":[]
                },
                "version":"legacy"
              }]
            }
          }
        }
      }
    }`)
	decoder := decodeBlockNotification(false)
	out, err := decoder(frame)
	require.NoError(t, err)
	got := out.(*BlockResult)
	require.NotNil(t, got.Value.Block)
	require.Nil(t, got.Value.ParsedBlock, "jsonParsed field must stay nil when binary path was requested")
	require.Equal(t, uint64(8), got.Value.Block.ParentSlot)
	require.Len(t, got.Value.Block.Transactions, 1)

	tx := got.Value.Block.Transactions[0]
	require.NotNil(t, tx.Transaction, "transaction is the [blob, 'base64'] 2-array shape")
	require.NotNil(t, tx.Meta)
	require.Equal(t, uint64(5000), tx.Meta.Fee)
}

// TestParsedBlockSubscribe_RejectsWrongEncoding keeps the guard rail on
// the (deprecated) parsed-only entry point: passing a non-jsonParsed
// encoding is a user bug, since ParsedBlockSubscribe hardcodes
// jsonParsed internally.
func TestParsedBlockSubscribe_RejectsWrongEncoding(t *testing.T) {
	cl := &Client{}
	_, err := cl.ParsedBlockSubscribe(
		NewBlockSubscribeFilterAll(),
		&BlockSubscribeOpts{Encoding: "base64"},
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "jsonParsed")
}
