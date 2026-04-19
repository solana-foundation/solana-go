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

package ws

import (
	stdjson "encoding/json"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/require"
)

// notification wraps a notification result in the jsonrpc envelope that
// decodeResponseFromMessage expects (matching Agave's wire format, see
// agave/rpc/src/rpc_subscriptions.rs tests).
func notification(method string, result string) []byte {
	return []byte(`{"jsonrpc":"2.0","method":"` + method +
		`","params":{"subscription":0,"result":` + result + `}}`)
}

// ----- accountSubscribe -----

func TestAccountResult_Decode(t *testing.T) {
	t.Run("base64", func(t *testing.T) {
		frame := notification("accountNotification", `{
          "context":{"slot":42},
          "value":{
            "lamports":1000000,
            "data":["SGVsbG8=","base64"],
            "owner":"11111111111111111111111111111111",
            "executable":false,
            "rentEpoch":18446744073709551615,
            "space":5
          }
        }`)
		var got AccountResult
		require.NoError(t, decodeResponseFromMessage(frame, &got))
		require.Equal(t, uint64(42), got.Context.Slot)
		require.NotNil(t, got.Value)
		require.Equal(t, uint64(1000000), got.Value.Lamports)
		require.Equal(t, solana.SystemProgramID, got.Value.Owner)
		require.False(t, got.Value.Executable)
		require.NotNil(t, got.Value.Data)
		require.Equal(t, []byte("Hello"), got.Value.Data.GetBinary())
	})

	t.Run("jsonParsed", func(t *testing.T) {
		frame := notification("accountNotification", `{
          "context":{"slot":7},
          "value":{
            "lamports":2039280,
            "data":{"program":"spl-token","parsed":{"type":"account","info":{"mint":"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"}},"space":165},
            "owner":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            "executable":false,
            "rentEpoch":0
          }
        }`)
		var got AccountResult
		require.NoError(t, decodeResponseFromMessage(frame, &got))
		require.NotNil(t, got.Value)
		require.NotNil(t, got.Value.Data)
		// Raw JSON path is preserved verbatim via DataBytesOrJSON.
		var asMap map[string]any
		raw, err := got.Value.Data.MarshalJSON()
		require.NoError(t, err)
		require.NoError(t, stdjson.Unmarshal(raw, &asMap))
		require.Equal(t, "spl-token", asMap["program"])
	})
}

// ----- programSubscribe -----

func TestProgramResult_Decode(t *testing.T) {
	frame := notification("programNotification", `{
      "context":{"slot":100},
      "value":{
        "pubkey":"4ejjNYBbaETZyqaiK8aDj2BWER8LKHgDcCnRrPC22YGg",
        "account":{
          "lamports":500,
          "data":["",  "base64"],
          "owner":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
          "executable":false,
          "rentEpoch":0
        }
      }
    }`)
	var got ProgramResult
	require.NoError(t, decodeResponseFromMessage(frame, &got))
	require.Equal(t, uint64(100), got.Context.Slot)
	require.Equal(t, "4ejjNYBbaETZyqaiK8aDj2BWER8LKHgDcCnRrPC22YGg", got.Value.Pubkey.String())
	require.Equal(t, uint64(500), got.Value.Account.Lamports)
}

// ----- signatureSubscribe -----

func TestSignatureResult_Decode_Processed(t *testing.T) {
	frame := notification("signatureNotification", `{"context":{"slot":1},"value":{"err":null}}`)
	var got SignatureResult
	require.NoError(t, decodeResponseFromMessage(frame, &got))
	require.Equal(t, uint64(1), got.Context.Slot)
	require.False(t, got.Value.IsReceived)
	require.NotNil(t, got.Value.Processed)
	require.Nil(t, got.Value.Processed.Err)
}

func TestSignatureResult_Decode_ProcessedWithErr(t *testing.T) {
	frame := notification("signatureNotification",
		`{"context":{"slot":2},"value":{"err":{"InstructionError":[0,"InvalidAccountData"]}}}`)
	var got SignatureResult
	require.NoError(t, decodeResponseFromMessage(frame, &got))
	require.NotNil(t, got.Value.Processed)
	require.NotNil(t, got.Value.Processed.Err)
}

func TestSignatureResult_Decode_Received(t *testing.T) {
	frame := notification("signatureNotification",
		`{"context":{"slot":3},"value":"receivedSignature"}`)
	var got SignatureResult
	require.NoError(t, decodeResponseFromMessage(frame, &got))
	require.True(t, got.Value.IsReceived, "untagged enum variant ReceivedSignature should map to IsReceived=true")
	require.Nil(t, got.Value.Processed)
}

// ----- logsSubscribe -----

func TestLogResult_Decode(t *testing.T) {
	frame := notification("logsNotification", `{
      "context":{"slot":5},
      "value":{
        "signature":"5j7s1QzqC6rA2FvR2gSzRcbfh4PE9eU7mVnxvKnZtWaAD9vqvR2rB8g4SfQ4ZBqS8PyZt7aX8ybX42kVhbZu8P7w",
        "err":null,
        "logs":["Program 11111111111111111111111111111111 invoke [1]","Program 11111111111111111111111111111111 success"]
      }
    }`)
	var got LogResult
	require.NoError(t, decodeResponseFromMessage(frame, &got))
	require.Equal(t, uint64(5), got.Context.Slot)
	require.Len(t, got.Value.Logs, 2)
	require.Nil(t, got.Value.Err)
}

// ----- voteSubscribe -----

func TestVoteResult_Decode(t *testing.T) {
	frame := notification("voteNotification", `{
      "votePubkey":"4ejjNYBbaETZyqaiK8aDj2BWER8LKHgDcCnRrPC22YGg",
      "slots":[100,101,102],
      "hash":"6TScP1N3f4n23Y5f1cZ1YmLMgwYyb6PTvQQjoNn4XDFz",
      "timestamp":1700000000,
      "signature":"5j7s1QzqC6rA2FvR2gSzRcbfh4PE9eU7mVnxvKnZtWaAD9vqvR2rB8g4SfQ4ZBqS8PyZt7aX8ybX42kVhbZu8P7w"
    }`)
	var got VoteResult
	require.NoError(t, decodeResponseFromMessage(frame, &got))
	require.Equal(t, "4ejjNYBbaETZyqaiK8aDj2BWER8LKHgDcCnRrPC22YGg", got.VotePubkey.String())
	require.Equal(t, []uint64{100, 101, 102}, got.Slots)
	require.Equal(t, "6TScP1N3f4n23Y5f1cZ1YmLMgwYyb6PTvQQjoNn4XDFz", got.Hash.String())
	require.Equal(t, "5j7s1QzqC6rA2FvR2gSzRcbfh4PE9eU7mVnxvKnZtWaAD9vqvR2rB8g4SfQ4ZBqS8PyZt7aX8ybX42kVhbZu8P7w", got.Signature.String())
}

// ----- slotSubscribe -----

func TestSlotResult_Decode(t *testing.T) {
	frame := notification("slotNotification", `{"slot":42,"parent":41,"root":40}`)
	var got SlotResult
	require.NoError(t, decodeResponseFromMessage(frame, &got))
	require.Equal(t, uint64(42), got.Slot)
	require.Equal(t, uint64(41), got.Parent)
	require.Equal(t, uint64(40), got.Root)
}

// ----- rootSubscribe -----

func TestRootResult_Decode(t *testing.T) {
	frame := notification("rootNotification", `123456`)
	var got RootResult
	require.NoError(t, decodeResponseFromMessage(frame, &got))
	require.Equal(t, RootResult(123456), got)
}

// ----- slotsUpdatesSubscribe -----

// Covers every SlotUpdate variant in agave/rpc-client-types/src/response.rs.
func TestSlotsUpdatesResult_Decode(t *testing.T) {
	cases := []struct {
		name     string
		payload  string
		assertFn func(*testing.T, *SlotsUpdatesResult)
	}{
		{
			name:    "firstShredReceived",
			payload: `{"type":"firstShredReceived","slot":1,"timestamp":1700000000}`,
			assertFn: func(t *testing.T, r *SlotsUpdatesResult) {
				require.Equal(t, SlotsUpdatesFirstShredReceived, r.Type)
			},
		},
		{
			name:    "createdBank",
			payload: `{"type":"createdBank","slot":10,"parent":9,"timestamp":1700000001}`,
			assertFn: func(t *testing.T, r *SlotsUpdatesResult) {
				require.Equal(t, SlotsUpdatesCreatedBank, r.Type)
				require.Equal(t, uint64(9), r.Parent)
			},
		},
		{
			name:    "frozen",
			payload: `{"type":"frozen","slot":11,"timestamp":1700000002,"stats":{"numTransactionEntries":3,"numSuccessfulTransactions":2,"numFailedTransactions":1,"maxTransactionsPerEntry":64}}`,
			assertFn: func(t *testing.T, r *SlotsUpdatesResult) {
				require.Equal(t, SlotsUpdatesFrozen, r.Type)
				require.NotNil(t, r.Stats)
				require.Equal(t, uint64(3), r.Stats.NumTransactionEntries)
				require.Equal(t, uint64(2), r.Stats.NumSuccessfulTransactions)
				require.Equal(t, uint64(1), r.Stats.NumFailedTransactions)
				require.Equal(t, uint64(64), r.Stats.MaxTransactionsPerEntry)
			},
		},
		{
			name:    "dead",
			payload: `{"type":"dead","slot":12,"timestamp":1700000003,"err":"ProducedSame"}`,
			assertFn: func(t *testing.T, r *SlotsUpdatesResult) {
				require.Equal(t, SlotsUpdatesDead, r.Type)
				require.Equal(t, "ProducedSame", r.Err)
			},
		},
		{
			name:    "optimisticConfirmation",
			payload: `{"type":"optimisticConfirmation","slot":13,"timestamp":1700000004}`,
			assertFn: func(t *testing.T, r *SlotsUpdatesResult) {
				require.Equal(t, SlotsUpdatesOptimisticConfirmation, r.Type)
			},
		},
		{
			name:    "root",
			payload: `{"type":"root","slot":14,"timestamp":1700000005}`,
			assertFn: func(t *testing.T, r *SlotsUpdatesResult) {
				require.Equal(t, SlotsUpdatesRoot, r.Type)
			},
		},
		{
			name:    "completed",
			payload: `{"type":"completed","slot":15,"timestamp":1700000006}`,
			assertFn: func(t *testing.T, r *SlotsUpdatesResult) {
				require.Equal(t, SlotsUpdatesCompleted, r.Type)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			frame := notification("slotsUpdatesNotification", c.payload)
			var got SlotsUpdatesResult
			require.NoError(t, decodeResponseFromMessage(frame, &got))
			c.assertFn(t, &got)
		})
	}
}

// TestConfigSurfaceParity intentionally uses `_ = zero.Field` to assert
// at compile time that each Config struct still carries every option
// from the matching Agave config. A field rename or removal will fail
// to build — do not "simplify" these assignments away.
func TestConfigSurfaceParity(t *testing.T) {
	t.Run("AccountSubscribeConfig", func(t *testing.T) {
		zero := AccountSubscribeConfig{}
		_ = zero.Commitment
		_ = zero.Encoding
		_ = zero.DataSlice
		_ = zero.MinContextSlot
	})
	t.Run("ProgramSubscribeConfig", func(t *testing.T) {
		zero := ProgramSubscribeConfig{}
		_ = zero.Commitment
		_ = zero.Encoding
		_ = zero.DataSlice
		_ = zero.Filters
		_ = zero.MinContextSlot
		_ = zero.WithContext
		_ = zero.SortResults
	})
	t.Run("SignatureSubscribeConfig", func(t *testing.T) {
		zero := SignatureSubscribeConfig{}
		_ = zero.Commitment
		_ = zero.EnableReceivedNotification
	})
	t.Run("BlockSubscribeOpts", func(t *testing.T) {
		zero := BlockSubscribeOpts{}
		_ = zero.Commitment
		_ = zero.Encoding
		_ = zero.TransactionDetails
		_ = zero.Rewards
		_ = zero.MaxSupportedTransactionVersion
	})
}

// TestConfigParams_Wire verifies that each Config's params() emits the
// exact camelCase key set that Agave's RpcXxxConfig expects, with
// unset options omitted (matching serde(skip_serializing_if="Option::is_none")).
func TestConfigParams_Wire(t *testing.T) {
	ptrU64 := func(v uint64) *uint64 { return &v }
	ptrBool := func(v bool) *bool { return &v }

	t.Run("AccountSubscribeConfig nil defaults to base64 encoding", func(t *testing.T) {
		var c *AccountSubscribeConfig
		got := c.params()
		require.Equal(t, solana.EncodingBase64, got["encoding"])
		require.Len(t, got, 1)
	})
	t.Run("AccountSubscribeConfig full", func(t *testing.T) {
		c := &AccountSubscribeConfig{
			Commitment:     rpc.CommitmentConfirmed,
			Encoding:       solana.EncodingJSONParsed,
			DataSlice:      &rpc.DataSlice{Offset: ptrU64(0), Length: ptrU64(64)},
			MinContextSlot: ptrU64(100),
		}
		got := c.params()
		require.Equal(t, rpc.CommitmentConfirmed, got["commitment"])
		require.Equal(t, solana.EncodingJSONParsed, got["encoding"])
		require.NotNil(t, got["dataSlice"])
		require.Equal(t, uint64(100), got["minContextSlot"])
		// MinContextSlot is written as a plain uint64, not as a pointer,
		// so JSON marshalling lands on a number (matching Agave).
		require.IsType(t, uint64(0), got["minContextSlot"])
	})

	t.Run("ProgramSubscribeConfig full", func(t *testing.T) {
		c := &ProgramSubscribeConfig{
			Commitment:     rpc.CommitmentConfirmed,
			Encoding:       solana.EncodingBase64Zstd,
			DataSlice:      &rpc.DataSlice{Offset: ptrU64(0), Length: ptrU64(8)},
			Filters:        []rpc.RPCFilter{{DataSize: 165}},
			MinContextSlot: ptrU64(1),
			WithContext:    ptrBool(true),
			SortResults:    ptrBool(false),
		}
		got := c.params()
		require.Equal(t, rpc.CommitmentConfirmed, got["commitment"])
		require.Equal(t, solana.EncodingBase64Zstd, got["encoding"])
		require.NotNil(t, got["dataSlice"])
		require.NotNil(t, got["filters"])
		require.Equal(t, uint64(1), got["minContextSlot"])
		require.Equal(t, true, got["withContext"])
		require.Equal(t, false, got["sortResults"])
	})
	t.Run("ProgramSubscribeConfig zero omits optional fields", func(t *testing.T) {
		got := (&ProgramSubscribeConfig{}).params()
		require.Equal(t, solana.EncodingBase64, got["encoding"])
		for _, k := range []string{"commitment", "dataSlice", "filters", "minContextSlot", "withContext", "sortResults"} {
			_, present := got[k]
			require.False(t, present, "unset %q must be omitted to match Agave wire format", k)
		}
	})

	t.Run("SignatureSubscribeConfig full", func(t *testing.T) {
		c := &SignatureSubscribeConfig{
			Commitment:                 rpc.CommitmentFinalized,
			EnableReceivedNotification: ptrBool(true),
		}
		got := c.params()
		require.Equal(t, rpc.CommitmentFinalized, got["commitment"])
		require.Equal(t, true, got["enableReceivedNotification"])
	})
	t.Run("SignatureSubscribeConfig zero emits empty map", func(t *testing.T) {
		got := (&SignatureSubscribeConfig{}).params()
		require.Empty(t, got)
	})
}
