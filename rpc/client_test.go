// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
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
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"

	"github.com/AlekSi/pointer"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	stdjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetAccountInfo(t *testing.T) {
	responseBody := `{"context":{"slot":83986105},"value":{"data":["dGVzdA==","base64"],"executable":true,"lamports":999999,"owner":"11111111111111111111111111111111","rentEpoch":18446744073709551615}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)
	out, err := client.GetAccountInfo(context.Background(), pubKey)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getAccountInfo",
			"params": []any{
				pubkeyString,
				map[string]any{
					"encoding": "base64",
				},
			},
		},
		reqBody,
	)

	rentEpoch, _ := new(big.Int).SetString("18446744073709551615", 10)
	assert.Equal(t,
		&GetAccountInfoResult{
			RPCContext: RPCContext{
				Context{Slot: 83986105},
			},
			Value: &Account{
				Lamports: 999999,
				Owner:    solana.SystemProgramID,
				Data: &DataBytesOrJSON{
					rawDataEncoding: solana.EncodingBase64,
					asDecodedBinary: solana.Data{
						Content:  []byte{0x74, 0x65, 0x73, 0x74},
						Encoding: solana.EncodingBase64,
					},
				},
				Executable: true,
				RentEpoch:  rentEpoch,
			},
		}, out)
}

func TestClient_GetAccountInfoWithOpts(t *testing.T) {
	responseBody := `{"context":{"slot":83986105},"value":{"data":["dGVzdA==","base64"],"executable":true,"lamports":999999,"owner":"11111111111111111111111111111111","rentEpoch":207}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	offset := uint64(22)
	length := uint64(33)
	minContextSlot := uint64(123456)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	opts := &GetAccountInfoOpts{
		Encoding:   solana.EncodingJSON,
		Commitment: CommitmentFinalized,
		DataSlice: &DataSlice{
			Offset: &offset,
			Length: &length,
		},
		MinContextSlot: &minContextSlot,
	}
	_, err := client.GetAccountInfoWithOpts(
		context.Background(),
		pubKey,
		opts,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getAccountInfo",
			"params": []any{
				pubkeyString,
				map[string]any{
					"encoding":   string(solana.EncodingJSON),
					"commitment": string(CommitmentFinalized),
					"dataSlice": map[string]any{
						"offset": float64(offset),
						"length": float64(length),
					},
					"minContextSlot": float64(minContextSlot),
				},
			},
		},
		reqBody,
	)
}

func TestClient_GetTransactionsForAddressWithOpts(t *testing.T) {
	responseBody := `{"data":[{"signature":"5h6xBEauJ3PK6SWCZ1PGjBvj8vDdWG3KpwATGy1ARAXFSDwt8GFXM7W5Ncn16wmqokgpiKRLuS83KUxyZyv2sUYv","slot":1054,"transactionIndex":42,"err":null,"memo":null,"blockTime":1641038400,"confirmationStatus":"finalized"}],"paginationToken":"1055:5"}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	address := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")
	limit := 100
	maxSupportedTransactionVersion := uint64(0)
	minContextSlot := uint64(99)
	slotGte := uint64(1000)
	slotLt := uint64(2000)
	blockTimeGte := int64(1640995200)
	blockTimeLt := int64(1641081600)
	sigLt := solana.MustSignatureFromBase58("5h6xBEauJ3PK6SWCZ1PGjBvj8vDdWG3KpwATGy1ARAXFSDwt8GFXM7W5Ncn16wmqokgpiKRLuS83KUxyZyv2sUYv")

	out, err := client.GetTransactionsForAddressWithOpts(
		context.Background(),
		address,
		&GetTransactionsForAddressOpts{
			TransactionDetails: TransactionDetailsSignatures,
			SortOrder:          TransactionsForAddressSortOrderAscending,
			Limit:              &limit,
			PaginationToken:    "1055:5",
			Commitment:         CommitmentConfirmed,
			Filters: &GetTransactionsForAddressFilters{
				Slot: &GetTransactionsForAddressUint64Filter{
					Gte: &slotGte,
					Lt:  &slotLt,
				},
				BlockTime: &GetTransactionsForAddressInt64Filter{
					Gte: &blockTimeGte,
					Lt:  &blockTimeLt,
				},
				Signature: &GetTransactionsForAddressSignatureFilter{
					Lt: &sigLt,
				},
				Status:        TransactionsForAddressStatusSucceeded,
				TokenAccounts: TransactionsForAddressTokenAccountsBalanceChanged,
			},
			Encoding:                       solana.EncodingJSON,
			MaxSupportedTransactionVersion: &maxSupportedTransactionVersion,
			MinContextSlot:                 &minContextSlot,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, out.Data, 1)
	assert.Equal(t, uint64(42), out.Data[0].TransactionIndex)
	require.NotNil(t, out.PaginationToken)
	assert.Equal(t, "1055:5", *out.PaginationToken)

	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTransactionsForAddress",
			"params": []any{
				address.String(),
				map[string]any{
					"transactionDetails": string(TransactionDetailsSignatures),
					"sortOrder":          string(TransactionsForAddressSortOrderAscending),
					"limit":              float64(limit),
					"paginationToken":    "1055:5",
					"commitment":         string(CommitmentConfirmed),
					"filters": map[string]any{
						"slot": map[string]any{
							"gte": float64(slotGte),
							"lt":  float64(slotLt),
						},
						"blockTime": map[string]any{
							"gte": float64(blockTimeGte),
							"lt":  float64(blockTimeLt),
						},
						"signature": map[string]any{
							"lt": sigLt.String(),
						},
						"status":        string(TransactionsForAddressStatusSucceeded),
						"tokenAccounts": string(TransactionsForAddressTokenAccountsBalanceChanged),
					},
					"encoding":                       string(solana.EncodingJSON),
					"maxSupportedTransactionVersion": float64(maxSupportedTransactionVersion),
					"minContextSlot":                 float64(minContextSlot),
				},
			},
		},
		reqBody,
	)
}

func TestClient_GetTransactionsForAddressFullWithOpts(t *testing.T) {
	responseBody := `{"data":[{"slot":1054,"transactionIndex":42,"blockTime":1641038400,"transaction":["AQID","base64"],"meta":{"err":null,"fee":5000,"preBalances":[1000000],"postBalances":[995000],"innerInstructions":[],"preTokenBalances":[],"postTokenBalances":[],"logMessages":[],"status":{"Ok":null},"rewards":[],"loadedAddresses":{"readonly":[],"writable":[]},"returnData":{"programId":"11111111111111111111111111111111","data":["","base64"]}}}],"paginationToken":"1055:5"}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	address := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")
	limit := 10
	out, err := client.GetTransactionsForAddressFullWithOpts(
		context.Background(),
		address,
		&GetTransactionsForAddressOpts{
			SortOrder: TransactionsForAddressSortOrderAscending,
			Limit:     &limit,
			Encoding:  solana.EncodingBase64,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, out.Data, 1)
	assert.Equal(t, uint64(42), out.Data[0].TransactionIndex)
	require.NotNil(t, out.Data[0].Transaction)
	assert.Equal(t, []byte{1, 2, 3}, out.Data[0].Transaction.GetBinary())

	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTransactionsForAddress",
			"params": []any{
				address.String(),
				map[string]any{
					"sortOrder":          string(TransactionsForAddressSortOrderAscending),
					"limit":              float64(limit),
					"encoding":           string(solana.EncodingBase64),
					"transactionDetails": string(TransactionDetailsFull),
				},
			},
		},
		reqBody,
	)
}

// mustAnyToJSON marshals the provided variable
// to JSON bytes.
func mustAnyToJSON(raw any) []byte {
	out, err := json.Marshal(raw)
	if err != nil {
		panic(err)
	}
	return out
}

// mustJSONToInterface unmarshals the provided JSON bytes
// into an `interface{}` type variable, and returns it.
func mustJSONToInterface(rawJSON []byte) any {
	var out any
	err := json.Unmarshal(rawJSON, &out)
	if err != nil {
		panic(err)
	}
	return out
}

// mustJSONToInterfaceWithUseNumber unmarshals the provided JSON bytes
// into an `interface{}` type variable, and returns it.
// The decoder is configured with `UseNumber()`.
func mustJSONToInterfaceWithUseNumber(rawJSON []byte) any {
	var out any
	dec := json.NewDecoder(bytes.NewReader(rawJSON))
	dec.UseNumber()
	err := dec.Decode(&out)
	if err != nil {
		panic(err)
	}
	return out
}

// wrapIntoRPC wraps the provided string is an RPC payload as result.
func wrapIntoRPC(res string) string {
	return `{"jsonrpc":"2.0","result":` + res + `,"id":0}`
}

func TestClient_GetBalance(t *testing.T) {
	responseBody := `{"context":{"slot":83987501},"value":19039980000}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()

	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)
	out, err := client.GetBalance(
		context.Background(),
		pubKey,
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBalance",
			"params": []any{
				pubkeyString,
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	assert.Equal(t,
		&GetBalanceResult{
			RPCContext: RPCContext{
				Context{Slot: 83987501},
			},
			Value: 19039980000,
		}, out)
}

func TestClient_ContextApiVersion(t *testing.T) {
	responseBody := `{"context":{"slot":83987501,"apiVersion":"2.2.1"},"value":19039980000}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubKey := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")
	out, err := client.GetBalance(context.Background(), pubKey, CommitmentFinalized)
	require.NoError(t, err)

	assert.Equal(t, uint64(83987501), out.Context.Slot)
	require.NotNil(t, out.Context.ApiVersion)
	assert.Equal(t, "2.2.1", *out.Context.ApiVersion)
}

func TestClient_ContextApiVersion_Absent(t *testing.T) {
	responseBody := `{"context":{"slot":83987501},"value":19039980000}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubKey := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")
	out, err := client.GetBalance(context.Background(), pubKey, CommitmentFinalized)
	require.NoError(t, err)

	assert.Equal(t, uint64(83987501), out.Context.Slot)
	assert.Nil(t, out.Context.ApiVersion)
}

func TestClient_GetBlock(t *testing.T) {
	responseBody := `{"blockHeight":69213636,"blockTime":1625227950,"blockhash":"5M77sHdwzH6rckuQwF8HL1w52n7hjrh4GVTFiF6T8QyB","parentSlot":83987983,"previousBlockhash":"Aq9jSXe1jRzfiaBcRFLe4wm7j499vWVEeFQrq5nnXfZN","rewards":[{"lamports":1595000,"postBalance":482032983798,"pubkey":"5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq","rewardType":"Fee"}],"transactions":[{"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[441866063495,40905918933763,1,1,1],"postTokenBalances":[],"preBalances":[441866068495,40905918933763,1,1,1],"preTokenBalances":[],"rewards":[],"status":{"Ok":null}},"transaction":["AQp2TH1spzjBAVM3alvnpaePFx3YEo9dvRglDuSChZUoTMD\/\/2h0HY5+89LJjCdiGJ7Ph3+Fyvbeiz1uJF8gxw0BAAMFyH0KDkXtjL1xebUYflZxYGlpV+LvjazzZCb\/mF2T67xZmkOUM\/A0iDSEkFzD5m4Ol82vsojigvqxrmp7Z1vrQgan1RcZLwqvxvJl4\/t3zHragsUp0L47E24tAFUgAAAABqfVFxjHdMkoVmOYaR1etoteuKObS21cc1VbIQAAAAAHYUgdNXR0u3xNdiTr072z2DVec9EQQ\/wNo1OAAAAAAAMFYbeqrsxJ9\/vZxtOaFi3rT2w9RF5Xi4jsyu61f3t1AQQEAQIDAAR0ZXN0","base64"]},{"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[334759887662,151357332545078,1,1,1],"postTokenBalances":[],"preBalances":[334759892662,151357332545078,1,1,1],"preTokenBalances":[],"rewards":[],"status":{"Ok":null}},"transaction":["ATA7DkBatbe2JB43QV+QRj2yoXSMXXttYFggDxZYOBfsRyYuGtzrbUevivclchxVccRIPlRP9PtS\/9NPXlwmhwwBAAMFSDrhjiNPuNqc4BWwitZz7xJ2NIXtv6XZtwtEOmgLj3n3NQ+OONLFlsu0LoUBSDsp40i9jOjZJBsliMtvTfdV+gan1RcZLwqvxvJl4\/t3zHragsUp0L47E24tAFUgAAAABqfVFxjHdMkoVmOYaR1etoteuKObS21cc1VbIQAAAAAHYUgdNXR0u3xNdiTr072z2DVec9EQQ\/wNo1OAAAAAAAKlcZMqS\/Oh0v+kOq2Ipg73NqbvKBRGQJDK8\/01K+MBAQQEAQIDAAR0ZXN0","base64"]}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()

	client := New(server.URL)

	block := 33
	out, err := client.GetBlock(
		context.Background(),
		uint64(block),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBlock",
			"params": []any{
				float64(block),
				map[string]any{
					"encoding": string(solana.EncodingBase64),
				},
			},
		},
		reqBody,
	)

	// TODO:
	// - test also when requesting only signatures

	tx1 := &solana.Transaction{
		Message: solana.Message{
			AccountKeys: []solana.PublicKey{
				solana.MustPublicKeyFromBase58("EVd8FFVB54svYdZdG6hH4F4hTbqre5mpQ7XyF5rKUmes"),
				solana.MustPublicKeyFromBase58("72miaovmbPqccdbAA861r2uxwB5yL1sMjrgbCnc4JfVT"),
				solana.MustPublicKeyFromBase58("SysvarS1otHashes111111111111111111111111111"),
				solana.MustPublicKeyFromBase58("SysvarC1ock11111111111111111111111111111111"),
				solana.MustPublicKeyFromBase58("Vote111111111111111111111111111111111111111"),
			},
			Header: solana.MessageHeader{
				NumReadonlySignedAccounts:   0,
				NumReadonlyUnsignedAccounts: 3,
				NumRequiredSignatures:       1,
			},
			Instructions: []solana.CompiledInstruction{
				{
					Accounts:       []uint16{1, 2, 3, 0},
					Data:           solana.Base58([]byte{0x74, 0x65, 0x73, 0x74}),
					ProgramIDIndex: 4,
				},
			},
			RecentBlockhash: solana.MustHashFromBase58("CnyzpJmBydX1X2FyXXzsPFc5WPT9UFdLVkEhnvW33at"),
		},
		Signatures: []solana.Signature{
			solana.MustSignatureFromBase58("D8emaP3CaepSGigD3TCrev7j67yPLMi82qfzTb9iZYPxHcCmm6sQBKTU4bzAee4445zbnbWduVAZ87WfbWbXoAU"),
		},
	}
	tx1Data, err := DataBytesOrJSONFromBase64(tx1.MustToBase64())
	require.NoError(t, err)

	tx2 := &solana.Transaction{
		Message: solana.Message{
			AccountKeys: []solana.PublicKey{
				solana.MustPublicKeyFromBase58("5rxRt2GVpSUFJTqQ5E4urqJCDbcBPakb46t6URyxQ5Za"),
				solana.MustPublicKeyFromBase58("HdzdTTjrmRLYVRy3umzZX4NcUmGTHu6hvYLQN2jGJo53"),
				solana.MustPublicKeyFromBase58("SysvarS1otHashes111111111111111111111111111"),
				solana.MustPublicKeyFromBase58("SysvarC1ock11111111111111111111111111111111"),
				solana.MustPublicKeyFromBase58("Vote111111111111111111111111111111111111111"),
			},
			Header: solana.MessageHeader{
				NumReadonlySignedAccounts:   0,
				NumReadonlyUnsignedAccounts: 3,
				NumRequiredSignatures:       1,
			},
			Instructions: []solana.CompiledInstruction{
				{
					Accounts:       []uint16{1, 2, 3, 0},
					Data:           solana.Base58([]byte{0x74, 0x65, 0x73, 0x74}),
					ProgramIDIndex: 4,
				},
			},
			RecentBlockhash: solana.MustHashFromBase58("BL8oo42yoSTKUYpbXR3kdxeV5X1P8JUUZBZaeBL8K6G"),
		},
		Signatures: []solana.Signature{
			solana.MustSignatureFromBase58("xvrkWXwj5h9SsJvboPMtn4jbR6XNmnHYp4MAikKFwdtkpwMxceFZ46QRzeyGUqm5P1kmCagdUubr3aPdxo7vzyq"),
		},
	}
	tx2Data, err := DataBytesOrJSONFromBase64(tx2.MustToBase64())
	require.NoError(t, err)

	blockTime := solana.UnixTimeSeconds(1625227950)
	assert.Equal(t,
		&GetBlockResult{
			BlockHeight:       pointer.ToUint64(69213636),
			BlockTime:         &blockTime,
			Blockhash:         solana.MustHashFromBase58("5M77sHdwzH6rckuQwF8HL1w52n7hjrh4GVTFiF6T8QyB"),
			ParentSlot:        83987983,
			PreviousBlockhash: solana.MustHashFromBase58("Aq9jSXe1jRzfiaBcRFLe4wm7j499vWVEeFQrq5nnXfZN"),
			Rewards: []BlockReward{
				{
					Lamports:    1595000,
					PostBalance: 482032983798,
					Pubkey:      solana.MustPublicKeyFromBase58("5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq"),
					RewardType:  RewardTypeFee,
				},
			},
			Transactions: []TransactionWithMeta{
				{
					Meta: &TransactionMeta{
						Err:               nil,
						Fee:               5000,
						InnerInstructions: []InnerInstruction{},
						LogMessages: []string{
							"Program Vote111111111111111111111111111111111111111 invoke [1]", "Program Vote111111111111111111111111111111111111111 success",
						},
						PostBalances: []uint64{
							441866063495, 40905918933763, 1, 1, 1,
						},
						PostTokenBalances: []TokenBalance{},
						PreBalances: []uint64{
							441866068495, 40905918933763, 1, 1, 1,
						},
						PreTokenBalances: []TokenBalance{},
						Rewards:          []BlockReward{},
						Status: DeprecatedTransactionMetaStatus{
							"Ok": nil,
						},
					},
					Transaction: tx1Data,
				},
				{
					Meta: &TransactionMeta{
						Err:               nil,
						Fee:               5000,
						InnerInstructions: []InnerInstruction{},
						LogMessages: []string{
							"Program Vote111111111111111111111111111111111111111 invoke [1]", "Program Vote111111111111111111111111111111111111111 success",
						},
						PostBalances: []uint64{
							334759887662, 151357332545078, 1, 1, 1,
						},
						PostTokenBalances: []TokenBalance{},
						PreBalances: []uint64{
							334759892662, 151357332545078, 1, 1, 1,
						},
						PreTokenBalances: []TokenBalance{},
						Rewards:          []BlockReward{},
						Status: DeprecatedTransactionMetaStatus{
							"Ok": nil,
						},
					},
					Transaction: tx2Data,
				},
			},
		}, out)
}

func TestClient_GetBlockWithOpts(t *testing.T) {
	responseBody := `{"blockHeight":69213636,"blockTime":1625227950,"blockhash":"5M77sHdwzH6rckuQwF8HL1w52n7hjrh4GVTFiF6T8QyB","parentSlot":83987983,"previousBlockhash":"Aq9jSXe1jRzfiaBcRFLe4wm7j499vWVEeFQrq5nnXfZN","rewards":[{"lamports":1595000,"postBalance":482032983798,"pubkey":"5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq","rewardType":"Fee"}],"transactions":[{"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[441866063495,40905918933763,1,1,1],"postTokenBalances":[],"preBalances":[441866068495,40905918933763,1,1,1],"preTokenBalances":[],"rewards":[],"status":{"Ok":null}},"transaction":{"message":{"accountKeys":["EVd8FFVB54svYdZdG6hH4F4hTbqre5mpQ7XyF5rKUmes","72miaovmbPqccdbAA861r2uxwB5yL1sMjrgbCnc4JfVT","SysvarS1otHashes111111111111111111111111111","SysvarC1ock11111111111111111111111111111111","Vote111111111111111111111111111111111111111"],"header":{"numReadonlySignedAccounts":0,"numReadonlyUnsignedAccounts":3,"numRequiredSignatures":1},"instructions":[{"accounts":[1,2,3,0],"data":"3yZe7d","programIdIndex":4}],"recentBlockhash":"CnyzpJmBydX1X2FyXXzsPFc5WPT9UFdLVkEhnvW33at"},"signatures":["D8emaP3CaepSGigD3TCrev7j67yPLMi82qfzTb9iZYPxHcCmm6sQBKTU4bzAee4445zbnbWduVAZ87WfbWbXoAU"]}},{"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[334759887662,151357332545078,1,1,1],"postTokenBalances":[],"preBalances":[334759892662,151357332545078,1,1,1],"preTokenBalances":[],"rewards":[],"status":{"Ok":null}},"transaction":{"message":{"accountKeys":["5rxRt2GVpSUFJTqQ5E4urqJCDbcBPakb46t6URyxQ5Za","HdzdTTjrmRLYVRy3umzZX4NcUmGTHu6hvYLQN2jGJo53","SysvarS1otHashes111111111111111111111111111","SysvarC1ock11111111111111111111111111111111","Vote111111111111111111111111111111111111111"],"header":{"numReadonlySignedAccounts":0,"numReadonlyUnsignedAccounts":3,"numRequiredSignatures":1},"instructions":[{"accounts":[1,2,3,0],"data":"3yZe7d","programIdIndex":4}],"recentBlockhash":"BL8oo42yoSTKUYpbXR3kdxeV5X1P8JUUZBZaeBL8K6G"},"signatures":["xvrkWXwj5h9SsJvboPMtn4jbR6XNmnHYp4MAikKFwdtkpwMxceFZ46QRzeyGUqm5P1kmCagdUubr3aPdxo7vzyq"]}}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()

	client := New(server.URL)

	block := 33
	rewards := true
	maxSupportedTransactionVersion := uint64(0)
	_, err := client.GetBlockWithOpts(
		context.Background(),
		uint64(block),
		&GetBlockOpts{
			TransactionDetails:             TransactionDetailsSignatures,
			Rewards:                        &rewards,
			Commitment:                     CommitmentFinalized,
			MaxSupportedTransactionVersion: &maxSupportedTransactionVersion,
		},
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBlock",
			"params": []any{
				float64(block),
				map[string]any{
					"encoding":                       string(solana.EncodingBase64),
					"transactionDetails":             string(TransactionDetailsSignatures),
					"rewards":                        rewards,
					"commitment":                     string(CommitmentFinalized),
					"maxSupportedTransactionVersion": float64(maxSupportedTransactionVersion),
				},
			},
		},
		reqBody,
	)

	// TODO:
	// - test also when requesting only signatures
}

func TestClient_GetBlockWithOpts_AccountsMode(t *testing.T) {
	responseBody := `{"blockHeight":69213636,"blockTime":1625227950,"blockhash":"5M77sHdwzH6rckuQwF8HL1w52n7hjrh4GVTFiF6T8QyB","parentSlot":83987983,"previousBlockhash":"Aq9jSXe1jRzfiaBcRFLe4wm7j499vWVEeFQrq5nnXfZN","rewards":[],"transactions":[{"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":[],"postBalances":[441866063495,40905918933763,1],"postTokenBalances":[],"preBalances":[441866068495,40905918933763,1],"preTokenBalances":[],"rewards":[],"status":{"Ok":null}},"transaction":{"signatures":["D8emaP3CaepSGigD3TCrev7j67yPLMi82qfzTb9iZYPxHcCmm6sQBKTU4bzAee4445zbnbWduVAZ87WfbWbXoAU"],"accountKeys":[{"pubkey":"EVd8FFVB54svYdZdG6hH4F4hTbqre5mpQ7XyF5rKUmes","signer":true,"writable":true,"source":"transaction"},{"pubkey":"72miaovmbPqccdbAA861r2uxwB5yL1sMjrgbCnc4JfVT","signer":false,"writable":true,"source":"transaction"},{"pubkey":"Vote111111111111111111111111111111111111111","signer":false,"writable":false,"source":"lookupTable"}]},"version":0}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	maxVersion := uint64(1)
	rewards := false
	out, err := client.GetBlockWithOpts(
		context.Background(),
		389906766,
		&GetBlockOpts{
			Commitment:                     CommitmentFinalized,
			Encoding:                       solana.EncodingBase64,
			TransactionDetails:             TransactionDetailsAccounts,
			Rewards:                        &rewards,
			MaxSupportedTransactionVersion: &maxVersion,
		},
	)
	require.NoError(t, err)

	// Verify the request params
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBlock",
			"params": []any{
				float64(389906766),
				map[string]any{
					"encoding":                       string(solana.EncodingBase64),
					"transactionDetails":             string(TransactionDetailsAccounts),
					"rewards":                        rewards,
					"commitment":                     string(CommitmentFinalized),
					"maxSupportedTransactionVersion": float64(1),
				},
			},
		},
		reqBody,
	)

	// Verify we can extract account keys from transactions in "accounts" mode
	require.Len(t, out.Transactions, 1)
	tx := out.Transactions[0]

	accountKeys, err := tx.GetAccountKeys()
	require.NoError(t, err)

	require.Len(t, accountKeys.Signatures, 1)
	assert.Equal(t,
		solana.MustSignatureFromBase58("D8emaP3CaepSGigD3TCrev7j67yPLMi82qfzTb9iZYPxHcCmm6sQBKTU4bzAee4445zbnbWduVAZ87WfbWbXoAU"),
		accountKeys.Signatures[0],
	)

	require.Len(t, accountKeys.AccountKeys, 3)

	// First account: signer + writable, source=transaction
	assert.Equal(t, solana.MustPublicKeyFromBase58("EVd8FFVB54svYdZdG6hH4F4hTbqre5mpQ7XyF5rKUmes"), accountKeys.AccountKeys[0].Pubkey)
	assert.True(t, accountKeys.AccountKeys[0].Signer)
	assert.True(t, accountKeys.AccountKeys[0].Writable)
	require.NotNil(t, accountKeys.AccountKeys[0].Source)
	assert.Equal(t, AccountKeySourceTransaction, *accountKeys.AccountKeys[0].Source)

	// Second account: not signer, writable, source=transaction
	assert.Equal(t, solana.MustPublicKeyFromBase58("72miaovmbPqccdbAA861r2uxwB5yL1sMjrgbCnc4JfVT"), accountKeys.AccountKeys[1].Pubkey)
	assert.False(t, accountKeys.AccountKeys[1].Signer)
	assert.True(t, accountKeys.AccountKeys[1].Writable)
	require.NotNil(t, accountKeys.AccountKeys[1].Source)
	assert.Equal(t, AccountKeySourceTransaction, *accountKeys.AccountKeys[1].Source)

	// Third account: not signer, not writable, source=lookupTable
	assert.Equal(t, solana.MustPublicKeyFromBase58("Vote111111111111111111111111111111111111111"), accountKeys.AccountKeys[2].Pubkey)
	assert.False(t, accountKeys.AccountKeys[2].Signer)
	assert.False(t, accountKeys.AccountKeys[2].Writable)
	require.NotNil(t, accountKeys.AccountKeys[2].Source)
	assert.Equal(t, AccountKeySourceLookupTable, *accountKeys.AccountKeys[2].Source)
}

func TestClient_GetBlockWithOpts_AccountsMode_GetTransactionFails(t *testing.T) {
	// When using "accounts" mode, calling GetTransaction() should fail
	// because there's no binary transaction data
	responseBody := `{"blockHeight":69213636,"blockTime":1625227950,"blockhash":"5M77sHdwzH6rckuQwF8HL1w52n7hjrh4GVTFiF6T8QyB","parentSlot":83987983,"previousBlockhash":"Aq9jSXe1jRzfiaBcRFLe4wm7j499vWVEeFQrq5nnXfZN","rewards":[],"transactions":[{"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":[],"postBalances":[100],"postTokenBalances":[],"preBalances":[200],"preTokenBalances":[],"rewards":[],"status":{"Ok":null}},"transaction":{"signatures":["D8emaP3CaepSGigD3TCrev7j67yPLMi82qfzTb9iZYPxHcCmm6sQBKTU4bzAee4445zbnbWduVAZ87WfbWbXoAU"],"accountKeys":[{"pubkey":"EVd8FFVB54svYdZdG6hH4F4hTbqre5mpQ7XyF5rKUmes","signer":true,"writable":true,"source":"transaction"}]}}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetBlockWithOpts(context.Background(), 100, nil)
	require.NoError(t, err)

	require.Len(t, out.Transactions, 1)

	// GetTransaction should fail — no binary data
	_, err = out.Transactions[0].GetTransaction()
	assert.Error(t, err)

	// GetAccountKeys should succeed
	accountKeys, err := out.Transactions[0].GetAccountKeys()
	require.NoError(t, err)
	require.Len(t, accountKeys.AccountKeys, 1)
	assert.Equal(t, solana.MustPublicKeyFromBase58("EVd8FFVB54svYdZdG6hH4F4hTbqre5mpQ7XyF5rKUmes"), accountKeys.AccountKeys[0].Pubkey)
}

func TestClient_GetBlockHeight(t *testing.T) {
	responseBody := `69217140`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetBlockHeight(
		context.Background(),
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBlockHeight",
			"params": []any{
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetBlockProduction(t *testing.T) {
	responseBody := `{"context":{"slot":83992896},"value":{"byIdentity":{"121cur1YFVPZSoKQGNyjNr9sZZRa3eX2bSuYjXHtKD6":[44,38],"123vij84ecQEKUvQ7gYMKxKwKF6PbYSzCzzURYA4xULY":[52,49],"12QYHqRxPuTPfkBVLetEuGkLGHD9GhqM5coP67xK7wfG":[64,55],"12Y25eHzGPaK5R5DjQ1kgJWuVd7zrtQ7cmaMPfmacsJV":[40,29],"12uDsSSWyPRGNK3HqcLBRNZiNFJWXCcHvHD7V4RYsKMr":[60,54],"132GXL3pzAjyEKoYLBS3QDTWDLqPnLJdZNcK4cWNDrmb":[64,43],"13nbrL1VjkfTZuaz7rNriYw6fWDFggqEf2g1C4mPETkr":[36,20],"14Z57kkY62p2UZyqeQyoGsXfkKbguAF1G8g2kZdV7Vae":[16,0],"1B4UocjePKwr58Jw4sLBsBHFt9nXGWxi1QDv9g73mrs":[48,36],"1DdfupGTrYUKtRxN9ukGCf3HBquc4buGPPmZr9WEjY4":[36,16],"21Ew2QbeiXprspa96d76RgueZ6HvrQMDTFAHpa71hpoR":[68,53],"234u57PuEif5LkTBwS7rHzu1XF5VWg79ddLLDkYBh44Q":[4,0],"238Fmy2TwU26Fo8XFRu2PzDWNbcn3bitywEPYG6tpztu":[44,36],"23eke7qW4tibp13JfiHKLErVsdmLDTwVsqg52bVQwBCZ":[32,23],"23xMZ9ijUgM9mRVB8sk7ZUR9yCRdP9eWU4ohjgbQbhGV":[64,45],"25UM59KCvciYwhjCq7t1rC8ZuvsxQBC2QRcaRNfq7xML":[56,50],"27JQHoxJi8kpJzwED2eTv8jDPB81gjqNFYBNN6rjM3UM":[44,43],"28LgQ7MeEZVgNJfYRc6UnoAz2SnSjKbyCKM6sntCRotb":[64,49],"28aB6dFf5TPKz3ghnYnu7nNaLsinoAE4xNyid1sy9j9e":[52,49],"295DP6WSsiJ3oLNPhZ3oSZUA6gwbGe7KDoHWRnRBZAHu":[76,59],"29Xwdi4HpBr1u9EAqDz3tBbMuwqBuczLPuVe2gGkg7ZF":[44,33],"29mA6zhspyms8m17FX8ztzz5UU9Fdqbumk1vxEGUkC7H":[52,38],"29tXWWzvGNvE5j8i6FLfHpmanPC9treZsCo1uA4ik1kL":[56,43],"2AY3bKHAMkdj4cCn1UcWCjewrg3ccDnhVmvJ3WrmmkAL":[36,33],"2BT25HZHpyzYmTbqyqzxBK7YjjH4a6aZ783TEgTTGYo5":[48,42],"2C26iHJcU5dqJJQ6NME3Lq583RT1Js9QDtgfmzknRajc":[52,43],"2C9pDcbRQJxbUHivgDdg4LGuMwm5oeVCnHS9w5JktNTo":[68,49],"2CGskjnksG9YwAFMJkPDwsKx1iRAXJSzfpAxyoWzGj6M":[60,36],"2Ceu5z672tACCDB6cmB3n4tih7B8dikv6k6kfYr3Li1e":[44,36],"2CjEU72sCTy1D6GyvpAjtKVGz94jdz8geN2JNWJCzZ6q":[56,0],"2D1oCLRK6geGhV5RyZ52JD9Qzqt311AEH1XrTjZdzbRh":[48,38],"2DvsPbbKrBaJm7SbdVvRjZL1NGCU3MwciGCoCw42fTMu":[4,0],"2Ebt7yP857s1WfdoqNm9FsGeahCvjdXcqhvsVjzNgUfx":[60,38],"2F5vGa1L5f1kKKwfcQvWGQCkJ7aoAxDRg4mZmxq9Ti3i":[40,34],"2FCxeG7mBYy2kpYgvgTZLCYXnaqDc4hGwYunhzHmEUD1":[68,46],"2FTDGeDUAXJokjVSjRXX2WoTc4tW2uMagYB5jk4JCJAK":[36,31],"2GAdxV8QafdRnkTwy9AuX8HvVcNME6JqK2yANaDunhXp":[40,32],"2H6AvmuhZ2yWSN8K8CQTPcAfVaGM63cr3oUeVSw6pUhT":[40,34],"2JPBDCGefojYLyy87VfJkHUVjMhD4H49KPgdCkitRwTi":[52,19],"2KFrkqEeSBKEHiMjUugPxTkBJ2jXepgFBqHu5ZFxtaFg":[56,38],"2LsRRgttA1PKXXeKTZP2QhetgM94Dj5uecmTzyQkTvXK":[32,13],"2MKNRRH59tZXPUas1UcozqZtAHmXJBzpvNHUGySQUaw4":[64,46],"2P3YH9psWAAM6QQgA8NaQnKHQ973cKNqTSFFCNYE4gjk":[64,56],"2PDvmDx6HeKv3wtdwmGQGmz9pGXXDKNFVvGizGGaAqxL":[56,38],"2Pik6jn6yLQVi8jmwvZCibTygPWvhh3pXoGJrGT3eVGf":[44,41],"2PvsR9DM2GZavFQGDsdJwXJvPWsyneyT9Gpu7wXGDkSr":[52,38],"2Pvzm7bGYjCpfjj8iyF724eesh12PejtKgxzv53ctgXk":[4,4],"2RLf3RSy1ScBFL5UzVDw3jYKCAuGA9vHpr9dnbQzJt8V":[52,51],"2RNHZTsFQF7BwgTrrwH4qvibWeguU67BKnPAxysaLUVi":[52,51],"2RpZdDc9ss5VsVUgHox2e6u6yV1SKUrQV6iuzvggdLKK":[48,47],"2TEGxhx2CgHw5fpkrvJWBsRKbJNAT3y6Fco4Lj5DsJjk":[52,39],"2TGWTnUbjfZqvFYTgwUTdA3rXLHshRbeDgVMEMg7icZy":[24,23],"2TcLzmdcpQqbNEBjU2Hxp1MiETSww1HJeWW84F61Z13k":[56,53],"2UCNzcnSVGtkgzpm1guxR6hEDQ5A8gVDkwVTWfZ31bPg":[36,31],"2UJ4q96QBg8dQom8JFgCMASdJnngrxhafgjY3XA5ddso":[40,31],"2Uv6KbG9Smt8PVMiPVGcRzt2GMvUvnvydSsgVRUZZZdS":[56,43],"2VEnBfmR1LW44oemZK29MtBHxTuprLizjPcLycNyGkTt":[56,33],"2VL9dMnHPJG6sD9CuB1BkPCpB6S5EMnvXBECHHnKqz3z":[32,32],"2VzCLy98rzmvKGo23e1eM4LANCt9JFrtVTBBMZzGT4FW":[56,48],"2X2APoUmQcbyVfVNCmivzYRkydxZkfVdXXaSCHQTa8mC":[76,66],"2X5JSTLN9m2wm3ejCxfWRNMieuC2VMtaMWSoqLPbC4Pq":[28,21],"2XP9MfzWQnX3DiAJQFSKXKyqBr4M7GKhHnKc9P7z519H":[4,0],"2YeoCYp1KT5W6S8MEVbu1omSrHVtZPEVcpFFKRdXwfAK":[8,7],"2YgttBBx9Ax6WhF9xu8CSWw5usKFWxaxzJtg8zPajadx":[4,4],"2YtaVYd8fZHXpPezG5gRp9nudXGBEPcqFtJe8ikN1DZ7":[44,38],"2YvDq2K7zBvsqZFVqTGVcKqJqSLZaJH1hny6fah14mqt":[16,11],"2ZETk6Sy3wdrbTRRCFa6u1gzNjg59B5yFJwjiACC6Evc":[56,36],"2ZQbsxdEab52wEFELQnn2wsN4LRhsPrjyCeqfDdtD2f1":[56,29],"2ZZkgKcBfp4tW8qCLj2yjxRYh9CuvEVJWb6e2KKS91Mj":[60,52],"2ZqaTLm1TZfpYdR7d8XhRntPjmrF6q69YZVLb6j4GcWz":[64,38],"2anGa2owPRuQyHyEaWSbrrWws6NiyoakByaXEufuU3hH":[72,51],"2bQyrSEPaQ9BMbu7Ftv7ye1fxtSLW3oZRj4d2U64AJmc":[52,40],"2bZXLja7MqWiTtDfxTm78rvxaxfa34RhF4msQmvCBAWn":[52,37],"2cgXfdfA4EcJjouu5jxruaCMPyc5q3oe4qRMB14EGWyL":[76,67],"2cxgydEWqiVTookrgPWucNuQqmwThyoNAYPX3GqeDdvh":[52,42],"2dm9YbgXtR5yimmgsLkfaMLcNZxhjywW4bLnvChms3tb":[56,38],"2eA3YU5GVKRdFKREMMNmaLjptjvBLrQQZtuRDM8hZWde":[28,23],"2eDDjJSKdxf8qwojH1E2SoZFHqst56GXxtmAnoZtGdtu":[48,35],"2eoKP1tzZkkXWexUY7XHLSNbo9DbuFGBssfhp8zCcdwH":[12,0],"2fCrJDUrArXita8dQDruPjBLMXTWKuMdbQVnVRGZYUtb":[44,32],"2gHxXKYyVCGrTjuFNVi9gTPUF9xN4hRhxZhcHpscS2dQ":[60,35],"2gV5onEfn8KmtZ3Lck39GrNEZyTxJ1RiNV5s7fRdC3gc":[56,42],"2hJpiwrWXbpRaxFdQWnhYHR19bbWfUd5VqNk7hhjALCx":[68,0],"2i5Ms2WfHHWz4P9Bdn2x8ZtUQ3fUoR7GL8d6UjhZodCe":[48,35],"2i5vGJ4RQKFJM8vxbvFaCmPKLprMuKrhVcJZaYXoTmHg":[76,50],"2iAP1WMsKJVje22cgPJNGC7Jgv5DQj37QZwcDNUDd9F3":[52,44],"2ibbdJtxwzzzhK3zc7FR3cfea2ATHwCJ8ybcG7WzKtBd":[60,52],"2iczkZceGZQqimksY8uk6NLrQXoMFZGK1mTWos4QnZ3a":[92,85],"2jypS1SoX6MLEfuNvUH23K7UU3BsRu3vBphcd7BVkEpj":[68,20],"2khFqurxeMKKfhFJ9dfas1L9LsHwt2qHGW8Ztinzoeob":[32,29],"2mXbRwZihk1TfeyS8aQTTm3Lcg2QHz7SgVP1xPavNYho":[56,40],"2n9qy8LiuiNpaeKFw82AaBWfi5F6CFW1rPH1ZXNNSHvo":[64,53],"2ofEZBxkiZoBpxXcXT68RTHfuQQFChSYVXVPGbFfvMTP":[40,30],"2p64GWwGEWtHdwjdeXCMHU5LBstm5BenLdGsJZzDrKHH":[60,48],"2q8YQuJZhoAZkQNaZgQiP58tvr5HE3sfQp5beG9BzerR":[44,44],"2qcfDTvKHzp2RM8ZsAqogZinajULNnDa5a7yQFd3FrPs":[60,46],"2rsXVaKikXHsCuFyYEkoReVZEv4LZBoBBWG3wkNCSWK2":[60,28],"2sAvsH3WPrHJ79P2cbM1RBwuNea9aL8z1Q3u9buFw99g":[48,41],"2sBsdFT58SPfd5LQyE8MhEgJWpaoHUCoN4QFCVqNZpnj":[40,30],"2sCfpKiU1JhtHvveo7SL8mRYD3sPMmHaAMCC69CJ5cwG":[56,39],"2t4ED5vy44LXqRRPsesVdTUE4ScVGL3vamBUx1hi53dQ":[44,18],"2tStw7K6ApvwkgGzZxkLQ263UL76wpeNYgYadTvZe8Vc":[44,39],"2tjfEp1WfgX6n85U9e17aBifn1vNyLNr7jmwPf7SAiSy":[24,8],"2uyMGFNvMwC9tjjKXRZcGFSztnzxxkk2Gbkki1V46L76":[48,42],"2uzuT8dgVVLLgG57n5W9vxMTaNAaLavnt8gtiF7V1FVV":[64,49],"2wAbzXXERjCoRrgEH3sdYfgAaJArrrraXkDzEnDPsuMi":[32,12],"2wqeRoCVwNoDqjXuZJ5iDnP5NR5HsqfU3Jf3DR5oPYMt":[56,0],"2xFjhfxTKGVvGDXLwroqGiKNEF3KCSFaCRVLHfpsiPgd":[56,36],"2xoWe7LGX8Kmnwoc27VF2iYkxfKESjb3b9rU1iM9wHJT":[36,23],"2y857Ss2GgyL9WooNqt6sAgxDVwr9pE6i4BiJ2wrC4g3":[44,28],"2yDwZer11v2TTj86WeHzRDpE4HJVbyJ3fJ8H4AkUtWTc":[68,55],"2yLVWajWK99EZyxMquyJfnMTdUyNG62nfVipVY27ELun":[80,57],"2zHkPFBSxWF4Bc6P7XHaZMJLfBqtSgfDCBqTZ7STXE1a":[60,58],"33FtaV5DrLUPYYQK7QAiD3LBDXD2VoCv6BuCQVoRdq57":[56,43],"33LfdA2yKS6m7E8pSanrKTKYMhpYHEGaSWtNNB5s7xnm":[52,41],"33WStcUFh8kboGRCW2ZiFhhNi9AcSTzeasNsUV1Wqgut":[56,50],"34D4nS1eywoA1wiwcgrBP8Ewj9NXyaZ3dP9DJKfkvpGn":[12,12],"35A7K7f9Nk3YJLLoPvqxLW5ngRvX8fRBJSHhmowWmaSu":[44,40],"35FdizoSUbjCybXeogAZWoXKbnCzvWFx9pgMLohUCRh8":[48,47],"36GzimUeoiBaapYaC1yriTJ9moQK1QvJfexppcZv3PaN":[60,49],"36UkVHPN89MszKNYA7ywFQqyQ4FGKGkpWLZjiBZiDdLr":[64,59],"36gARMU4V3D6hu5EJi7wYFW6cC1tNym9DkjZfAGFQTbk":[44,36],"376e8QLx9qSkjFn7mK2kp3wBwvziKuMqiB3iAbK5Payx":[56,47],"37Gr1zVPr79E3AdPFj8EMyKZYt7Bnz3VWKjdFctQC8fL":[72,63],"37PWrAzWfgn2yptyHBZQ4HDBJu1V4BvhiK4vs81MQzo4":[8,4],"389UtZkyvUzHzQUt2eSEzeTiC32GF5PwsBq4uNZz3cpY":[56,36],"38hgERMK335yrDsyPkc4wbW2FUiXgmuWRght9n7RVAtz":[32,31],"39FH4cnkSawRtr9N2VbUVST4o6ZiixW2K4QCzLqW8tMg":[52,49],"3ANJb42D3pkVtntgT6VtW2cD3icGVyoHi2NGwtXYHQAs":[84,52],"3Ax8WYVrp7prVWixBzNcKzwQPXFZABnaupna1diZQfMK":[32,26],"3BCokPfahX9rLYMh6E6uYTEFuchiKd9wZcXUvwDHFYiH":[40,34],"3DYMXn1LtpPYChqUQFq27oMqnSidjcYzJQAt85jDUowr":[64,57],"3DeoXyzWHzc1puYcNZ8khMRFAXwL7JKEbvkXuUWdNsea":[8,6],"3Df9iVRoSkX3YZ6GmMexeLSX8vyt8BADpHn571KfQSWa":[48,28],"3FDRLuYj1dHoxiVQNbj8Dk16gGv4gmB1fmdVmu66WqPw":[64,41],"3FhfNWGiqDsy4xAXiS74WUb5GLfK7FVnn6kxt3CYLgvr":[72,51],"3Fiu2KFBf3BoT9REvsFbpb7L1vTSs7jnmuDrk4vZ9DNE":[68,58],"3GctwRAZHTAwxx78mU5unwtxSEiDNsF5MgoY1oXXHm6w":[52,45],"3H7BtRE7iGC9Kzxq5eb3Hx3hChNptepF1KRrEufKjNMD":[4,0],"3HM7uGuE3AD9smYFL8uKinAHo4GGtX2PErn66FhGp5mc":[44,31],"3HitRjngqhAgVuNdFwtR1Lp5tQavbJri8MvKUq5Jpw1N":[40,38],"3J2GJs7nWTiF5EcvcFeNYydpZzbL4NjZJetHyKMpxFnE":[52,37],"3JfoYf6wmQxhpry1L61dnDWYJbL7GYi4yt7mybehuhne":[36,25],"3K8BYGTPD9AxqYQDPdU8PPy6AfiSwf4hDmFy1xXGB8Ns":[24,24],"3KVuW6mGLD9Kv1Gjj1A5q685JZLp9hqE29Kbvnrii8gM":[64,52],"3LWv8RrdEyMtePAMCmohBzWAz7fmN7Cf2ctSUxJKEQnS":[64,58],"3LsQLm1hy8Rcxw2neKgFqcJzyXNugJtRxXRpMvSvzWCU":[72,57],"3LtAt3iqmeTgJ3GD8DtCcjkRkJdDKAF42nJytn28syeP":[56,42],"3MdUXXiLWeXQauVSiuGwPjakCv8J5CX5v1fu8eutJ7v1":[40,39],"3NMFamQ5RtVEs5N6KeUnGnwkaoukkp4hduzUPKJr5Y8t":[40,31],"3NchsxHzVUAv6MTGEuAVt8QRdi93uHGNRmS9AEiZkMVh":[64,35],"3NtGCPqA5dTucxitLz5KTxERZ7XdVSZ8c2m97TGupV3S":[56,51],"3PVz8crz85wgqgudf6mxws2psgKc4kr51MhfmU6VekEG":[52,49],"3Pog3tY91JZRv8irJf9sE4JKPn1pWBj9bLB9NHxHgehu":[24,11],"3QK8tbsVSwU6xRzLWhVFJCcnqm9WPxSUdaa7cXzBQZZh":[32,20],"3Qj4rFsMRMsXnYescUVi53kDY4KjNnNy2QE4tc4WpQET":[44,38],"3R82jDjQsrzZgQKiEJbKfdCA9ngYQrjZehYuEFmhhfCP":[96,84],"3SYNAWbZuuWMYKHwL73Gpeo1ySqRSshf5WDoW9vYVz9V":[72,62],"3Teu85ACyEKeqaoFt7ZTfGw256kdYGCcJXkMA5AbMfp4":[72,66],"3UtHK2ZWwmDKxd6QzrKmh9Pey1gWS2SW1MoaZuGZbc7E":[12,10],"3V1cpkuKhJhQuKB3BJvTeezGEDKR78krcJT2esG9Wte6":[32,23],"3W4fe5WTAS4iPzBhjGP8a1LHBTx8vbscqThXT1THqEGC":[52,48],"3WrJpBnPGbmFx8jPpseT9gA5LAtubovpaBjE9waUe7GV":[60,51],"3Wyqj2cgKYK2sSSb3wVv3wJ5yD3yigV8iLLttkZfKn8d":[64,59],"3X6FsQ8awkcU4iXTF82T4RtnTJx9LTY5D3dHK6zDE1Tp":[32,27],"3XE9NQAN6yiKvNedR5msnvdLN6HEdUfqyn4yWoEYYEcW":[4,0],"3Ye1g9E65wj9wtbTLetQbjsQ6SFj4s7RdTJaxjq6duDq":[44,34],"3ZwnSWgQBpphsSzZNA2A2uFMuXZyJKDHi1EHKjbd4ikw":[52,43],"3aN4DVNJFGcHsKHrueVNogCnYFcGQQXe979zXvBLxDhe":[60,44],"3aSjivWpfjcSyqLTf3fAuJfB2R1vkYxDsnNDmByaXQp9":[36,29],"3adqz1JN9sbsjHGxQizz2ibJmyCHtUpP9aPnZYxixB4c":[32,13],"3bQ4s7ynWKjEPrkTfDx1aT2sXejXXYjbfYumBHc5LA83":[52,48],"3cJeH1TCZcNf5gCZnSbfZne9DQCiexkzuH6gwQEeBjqA":[36,27],"3ccsSn54FkE2zYU7ELDyEhB9vQbJ3Fz1wBzuuAHB3KXj":[76,66],"3d8F4eCR9YwvXdmvVzLQ7hHLTBcHAaWpC7jQxcdoBHkk":[56,44],"3dFGZmTgBwsBfYqvJqKdwEQGa7HUdPsSbHmTkm4RjJef":[44,39],"3eLGe4vyZoNK3FvY6C4oxQ3cJMzmUerdVaaTsfN14Ngf":[52,28],"3erfNXKZP8vmLoc5mXzdwhXa8UhjSigAiiL7CczdP2LU":[4,4],"3g7c1Mufk7Bi8Kk4wKTGQ3eLwfHYqc7ySpP46fqEMsEC":[12,8],"3gmNSSQVewvEiBY9Vh4hmekPmtGTiKPvooBW7MSkUXPc":[64,54],"3gnRETF8Tnto3ZCh9yCw5sbqrq8zqgx32zH4y3dzEN7i":[48,33],"3hcdjAmggZJxjaRgMfxqhyCe3Uu2yshLRZPZ3mXAkst3":[56,0],"3ht1z7tMieDiLkukray7AauF214xtsWFFG1E4A1oeAXU":[68,60],"3i7sS5McrJ7EzU8nbdA5rcXT9kNiSxLxhwyfuxbsDvBj":[44,32],"3iPu9xQ3mCFmqME9ZajuZbFHjwagAxhgfTxnc4pWbEBC":[820,641],"3j7SjhZK4P7LrKRwAq5vNKerENDkWJP7xgxJmqsuj2jv":[40,31],"3jjwWrta8PF3paARTXMpKmF8wxDzUWRCcdvRCdscvbr5":[68,56],"3kWT2K2HfxrspLFoJhKUAio3QF85EuTemJKTUcPEjm7m":[60,48],"3kcjg9J1d47mgZkGuqitHd7K6Bz7XBVMpZRB1Sg5dKdN":[64,35],"3kiAniQf6y9ZT3SdE8X7Rq5jM3MX6BUZy5KDT3wt6zAk":[40,32],"3kkVsVzrxiJdYzrFuGwM4PvuHSmVtVFjzEMBKbgjMWvp":[40,33],"3mDhRnsnQdmRyLKx61i3gy2PeNGM2zQxgofPFSDdDLZo":[60,56],"3mx22d1aJLazEutJyHVszdwyLJcrRo26EKB4AWDbRxRc":[68,57],"3nL1oAkcW4M88VG4D78dNxHrqaNdKyJqKW3wbhhBjhig":[48,38],"3ndqwmmqTEFaydt6bgTDohL35WJCjv2cezUcYezcHHcJ":[48,39],"3nvAV4PVG2w1F9GDh3YMnhYNvEEzV3LRMJ5e6bMYcULk":[56,49],"3o3GJr9iAdJ2v2sZhRqiX5nGFJyrpdG7t1jePatMfFkn":[48,39],"3qaaXFYh389e1Ncboc7qbCWxSQdbaiYuTFrJVYuh7jo2":[52,40],"3rFxX6D68YhDpF7c6vDt2yhfp8CXXcjNNga43cCJ8Ww9":[28,26],"3sWB3AMv6Rd96cKTgtZBPCKoxDW74eGWcqKPkhHEzF1K":[64,57],"3si1tYjwb32Mj43LWw87Zy4acgtnxXMYeZuKHmKEYB8B":[36,24],"3si45SHHXsP8C6PVo1Zcpcry7DuivvogscAA63D8AKmR":[44,33],"3tEqZrbb7xwaRwri19Z5TAznrewnM2m2SCkvSmLztWcE":[52,46],"3tHeSnJt7dMSxyFGg2LW7GBXCWMg8KxYQQCKfnx7cFs4":[52,0],"3tSsxpkuuZjTBG9whoPU37kS3NFK48Morvm8ui2vBJLm":[40,35],"3tjCLs5cMKiTgArVujb3S5LhQbBhDppCu5yXd5eysSs7":[4,0],"3v3KN1rtwURN3NVbLJceVSY5zjb7SX9PSXDU8Qgwf9XJ":[64,46],"3viEMMqkPRBiAKXB3Y7yH5GbzqtRn3NmnLPi8JsZmLQw":[20,20],"3vkFbUsjMqkkgNvywvxTPbsGF18NMkwBX5cBeBsrhTRk":[8,0],"3vkog7Kaki74rn7JFWxKyrWfTEUnp4cLpJyvgs233MyM":[52,43],"3w6hQh7Ndx93eqbaEMLyR3BwqtRxT2XVumavvU93mcRk":[84,68],"3wao1rTFLniFiX6vofFyEdyKPuo7coZETkfEQxf1s7mS":[4,4],"3wwYJDVkY1rK5emynSYgbwUy9X3eFcNQiyYxc4Jsd9iL":[48,35],"3wz211BhQAE2n5fjDQSStM2iSizhNRyJDNRkDEc1YwMF":[52,40],"3xKsqGgLMNVazzNBsKa9TPG2Vo5fGLr1xkKrTMVXVVkT":[36,27],"3xUTkgPKNJZ3dkpDMV8zWV34BkmvKanguKipv6M9x2Mt":[44,32],"3xgtKbSXjtZe7hqxHbK2WLYJGPJw1hfvZKzHrTkygiZX":[40,39],"3yEpFZ8Vq3vbbfvLu4r6vkRnV7P2QS6FSqGNuMUXro8J":[56,52],"3zPHhqJE3AR2S7WxJf8YHoVZ6mxPNhvvdjfddRiNY99g":[72,50],"41nRdNqtbMp6xGQYucjjydvRQKjiRyxiqzDHjdaqMxCQ":[44,35],"42DeSPAaef333ZsSzBGADHhAeWTY68t8CTMzJ89Z6s2r":[36,27],"43ZCLRdQgcajUq4WTxtTqkqGtpNnJTmLUs4ef4qGKtAc":[76,61],"43h2uYRTSVhMNXKuxY4Kn6T558u436qy59cV6Sz6rdRi":[60,50],"44J72PpPim1PJHge3TwJWAMnuPhwE7DMLaZmCerYEC61":[60,57],"44Kuawvm2ngsSyvqsMLCTeWXUYxoedthgKAEL3BxCdXP":[44,34],"44yQJPhbBRV6povRiXDc3KkE7SPXUohF9ipqBvCjokhc":[44,42],"45M6om8quE2DnLh3cYnty8kx1D4AYbMUzZMpytku6Gff":[8,0],"45THWNjLaWBh8jbuP3HrcG4iUvenHpSHmFVGnkmuQH4U":[48,35],"45YDFXgHCEbeDs17Amrd851M4gxCSJH3uofCsvdKLhRJ":[56,32],"45aGtJWVx9xbhp11diPithdQS1E9Hzjm5b5HEpAM68Ax":[44,37],"462x4mp5aZ29SetJR3oka3d2ARXVKUcs9f9hZsapf7ML":[52,36],"46GijDorcsduUvWFNWKAV1yB6XwPG699wS2gR4no4zGU":[52,43],"46WCeEExQaEJfatG53qgxMzgPqubbrAvVBeYSyUQt317":[68,50],"46uT7tSZ1US9bJ93ByBxSBCmZhQogn15Pwuxp4fhWXqc":[44,37],"473ToSs8wTyGd2DTmwb1zNkr7TweNC1Wfui2FzKNB1JE":[52,41],"47JuXYUK2UvwBPxq8p4ePvDggkpz49xmw93N3VNGbDm9":[16,12],"486kJEz1XJ95nULg2Ccj9Av9yi1inexzHRVW9UjfR2B6":[68,55],"4958nAd4Gp1MZQEg97b7prdDKAgC5Ab3iQtNzAWyHqEV":[60,43],"49AqLYbpJYc2DrzGUAH1fhWJy62yxBxpLEkfJwjKy2jr":[20,15],"49JYKwBGHPsL5ji9LSzDS6WNNvs2AW2seC2qZDiMWkPk":[48,41],"49Q14TEnx7XTHsFtRs9xhQ12wXRHwaWJ5YSpGhVNhSgy":[68,39],"49YDWPPRQRatsNgUHLbPytGtKgEetBFsq58uGobM8sDz":[48,37],"49gM7gXEJEokKHEoUCNve3uCRMAoRwKUpEiqK2nku6C2":[76,60],"49oW1EjrYFvWJLUK82mhcDqN3hWir2LT8H2Sectvfmr6":[44,29],"4A7XYUpU2Cvj84fBhkcUQPQMJsZywqgjvD65zSRZmquP":[8,0],"4AYWAYndF6EsfgwVTrsHLMviNsvuqh9dAMcJynpJk6YB":[40,31],"4BVYRKYnwWbUYRtxHSNnue8xydUhexegZKohbbtkT7nv":[4,0],"4Bx5bzjmPrU1g74AHfYpTMXvspBt8GnvZVQW3ba9z4Af":[56,32],"4Bzp9fzcdjctbdo23SCwCEkPeQzCeyTb3WtwiK3KNVRc":[32,23],"4CVJ8FMombpnrE7C1a4mdwMMbhJDroAzjG51BuifPmcF":[4,4],"4Cvq4GbYn7jWPpUmdcMSL2tPBV5E6GqAHfFV3u1iG9Zv":[76,51],"4DPoKwdKKWCYMcjSfWVeo3G9dVvcVgc487HRNdAVNMfQ":[44,36],"4EACGRv7miQa6wSw5ymGiV3dnHVZwrsKQoe3aMDbjdEn":[40,36],"4ECpcT3wLE4EzBZ8Th3da1EaALL3pwuKn2jL9rGB1MUH":[84,48],"4Efdqh6SnwMiAcu8fPb7gDo7Eu4vrxMQgMdFb2JtwNLq":[56,32],"4GBSypESidsbB6ACFRUTkwDwcv1G5anashx6UvSypqCF":[32,32],"4Ge8T8WeH1fnv5SijRzPfC38jWnuBhiKe8iE9fsXbqLi":[72,67],"4GhLBaxr1oEHWpoGnWh3mcRXUkBU5EEQZv3L27c7ohoq":[8,4],"4GzmbxmepoggVLYzyXyM2GzYVaisJSuutsxrydoErSeu":[60,58],"4HjA5dBRcMajmaYfwYxqdJBzYbuFxPqjoVjnsTk6Xjqv":[60,58],"4JZsGW4WUSjAjH4joCaAAVnNi5ERfHr93YUDxmHZpDM7":[60,59],"4Jb1YfUUN1xxdYb28wPLT6A52j459uLNBJaetpk3vAKE":[72,50],"4LKx5Rz4NsxnpamAuD3xVcCdt6A5aoN89qaUuFsfBNdW":[52,50],"4MNtUgysSfjwfpgYBJFJQA2Kn5LXPQzgRLnJoCAseKrx":[8,0],"4Nh8T1d4YBZHEuQNRmFbLXPT5HbWicqPxGeKZ5SdAr4i":[68,46],"4P8diDfWD1ra7bF8BXDPUExMg2QAhTxVLTq3tU4QcH8p":[52,43],"4QVu7BnBgYBEkyq9zc6mu9V7HbNUX9dK4EfVo4SMvBwT":[56,46],"4QY21MyFAtXbagGymZuBLu3a6wUkFg5qaUDRwYj4Pnuy":[44,20],"4RwV6detEgRyvVcvhBv8gmjriEHrVmKegeYy1FqRZK6Z":[52,44],"4SqdkosjugZVRdX2kRptUng487Uece5toWHZXVh6cpQV":[16,11],"4SykXpKfFGdy7Yxx1wToYuBhhnTTXMhzewWBMp65wgnP":[12,12],"4U7KuubEDSPR3YY1YjmVjz7CcxVgrdz7sz1svUM4Vx3i":[4,4],"4UNcH9sxWUo6bfZY93gmPiGZNssEgmG9Ho7C9ecjMv5N":[4,0],"4Un8pHPkosqAkRabaxhA48YFbji5sk46ntFAxQxyc4Lf":[4,4],"4WPa1hkBxCBnHmWWgM3yt8TAgA7Rtfow6SHPy4v6yG4z":[40,27],"4WkMVnmyoWuAGifnmqdWNtD3nudHp4hPPqvnyUHLkGWC":[28,28],"4WufhXsUhPc7cdHXYxxDrYZVVLKa9jCDGC4ccfmuBvu2":[52,45],"4XWxphAh1Ji9p3dYMNRNtW3sbmr5Z1cvsGyJXJx5Jvfy":[4,0],"4Xqmh7JpjaFj5wJ6tNGbEY8eoY8U3fPMUKzfQXcGWiDR":[36,29],"4ZD3xAHfPcYacfZEYmAxS7D72UdVFhUUe5XLhEnQfSCD":[12,0],"4ZbygbNLCxdMa3EZLYBuQHF4zzfCtX5V6xJAVSZncnjS":[36,35],"4ZrtLrxqtpQE4juoSAsSmQKgeZLkEdiwi7gXZ8hWsVF2":[52,27],"4ZtE2XX6oQThPpdjwKXVMphTTZctbWwYxmcCV6xR11RT":[88,74],"4Zto93KdBuynSnyyQct6ecMVxGNrjvVHe4CbWJTtvLSq":[60,43],"4ajWybNN1XqaapKEEiz4MPMyCP7Ppuw7FMQwQ57o7gFZ":[56,50],"4bLyjRauEjdJGb86g9V9p2ysveMFZTJiDZZmg8Bj29ss":[56,35],"4baXhu594FEQtZsAmHNjNM8K3NxmPNsYCxyPUZnhwHLm":[64,46],"4bnqGCM2a14j1CiJ31gjJUf9B3kHZXzz3cFB1X1tSGft":[40,35],"4bpkzvzxJkhXCQNufEcybrXsT5vNW5xUiG7mcnxfGRGy":[76,57],"4cLRyEVzhvt1MKqEeVeVfsxfJzZyUwpJGQADBW9qgwks":[64,38],"4dWYFeMhh2Q6bqXdV7CCd4mJC81im2k6CXCBKVPShXjT":[44,38],"4dd19K7UmrUk4aScsqYaXEGcabRVh8opRhLo8uSJAKbZ":[32,15],"4eyn57baA11sgvkQafTcrwJ9qVs6QptXBahf43Li1jKc":[40,40],"4fA2MXsEG1mJfxTouJuFWQoBzsK7jQVXbd5UAfhMZHXk":[40,23],"4fBQr617DmhjekLFckh2JkGWNboKQbpRchNrXwDQdjSv":[40,36],"4fFhfoSezZmrvK5EeFRtMsMhHn3Vfno5iJY3JPXs7F78":[4,4],"4g5gX1mmFGGragqYQ1AsRpB8ZJvwCoUKVT5LtKTDrNSp":[64,51],"4gMboaRFTTxQ6iPoH3NmxLw6Ux3SEAGkQjfrBT1suDZd":[32,24],"4hDEtsHXAf6TMBNJHogmN5noitFzxGxKAs5YwsKZzrDd":[60,49],"4jZMrzWGfMHDRkEBqwnx1cPR6uP3i8v2EaKALzi7bYbc":[64,58],"4jhyvbBHbsRDF6och7pDQ7ahYTUr7wNkAYJTLLuMUtku":[56,42],"4k7N9gtmQeDDYJxbT5NSDkARghkQzEbgvE9mm8gSFicj":[8,4],"4mCp1G9zmqRH53wX7j17wmZimHbn6ep1NvLmsMUwHjDj":[64,27],"4mdxZgQQdkVJvPK8Z8T55sbUXU25ZzjTNs1ydvrzVnYs":[48,43],"4nKuNB7KsFPzfPURvXxpyBZu4Pmm1y9w6jdbHpaAEfTH":[40,36],"4nu5rdaXjhXHniTtVG5ZEZbU3NBZsnbTL6Ug1zcTAfop":[20,15],"4nw9knLrjB893wVF1PwpPofZz19ko5vWcq7dKmriiSdH":[56,42],"4o8VRbGZcmiWm4Zc79LsBgDcqXmmVte3kvCroq2zwLG9":[92,84],"4oJuAMQjsVoQdHybK5JsoKYUoR4akAZHNRa4Qjs83Dgq":[12,8],"4oNUWNoSNnwghHBCGsuAaQEuaB6oZEXE2w4VNhRxoaQc":[28,23],"4pZjWxF6277CRncZjggHdiDN96juPucZHg537d2km4f9":[40,35],"4q1KX2Epud4kS7tYuyndLaon1FskmDqcwh5ubxHiSzdP":[52,48],"4qVaZm5ZhNnNwwBYawGsM4DoSkGXMkxymMYnzCTsH2WY":[60,45],"4rqiq96AtM3V3me5aGKXnSycaVZgoNq8jYD6LwtryNuc":[64,49],"4sRKUyYwqmc38TpPGmkbLfjKkyNBGEBaiYJaMCYfkUBh":[60,49],"4sSihca8PLdP9Q4NBo2LXXBE9o4KUqpp4hSEyXQCS7Qq":[24,19],"4u2qTnf4QVC8PcgNFPBwY2PwdkiMa4jb3KnNZo4zZbtV":[68,45],"4uERagALHEAGx2uwndDoYn9WpJ9D1uia7Z6vuMpvWxuQ":[76,44],"4uVzFAT5ZpJ6cPo9ff7igWCT4MjcTVwETqf6y29YBzgE":[64,52],"4uXyHNPLMpdjs38aorfRUCLarbh5ydhbv2FkZwErBM5j":[48,37],"4uykzcDWW8wnVWMXXgh2RqXaddSVsx8TNvpJV7eACXbz":[36,30],"4v5dEHTVmWTRzP1L2PijNr5B5nDVUk3wNy6eJ7V8qQKQ":[28,24],"4vAu8eDW1YGVSQPMgZqAVjYDVFXJQPtVYQaryCH26yam":[72,63],"4vDoJgjaTyQ9uRLyCfVwb9pyhAZQRcvGNosvYaJS4eux":[56,37],"4vXPjSaZfydRqhnM85uFqDWqYcFyA744R2tjZQN8Nff4":[48,35],"4veSBAABaESW2WpnJzcdNcduopX7X1f63KziC24FhQee":[28,20],"4wjZmBoiwQ2s3fEL1og4gUcgWNtJoEkXNdG1yMW44nzr":[60,50],"4x7HEA12XAiqjsM5FbWkyNnwKfqzSDHWA1XA79uFpzGJ":[52,36],"4xv6aEhBpGsnXStV5GoxEdX22p5uDzVNFKEmHaQUhPnM":[52,46],"4z3zVorKFWWD3ULZ9X1XVcz8rYtxKiNE5AQtSpEuYd84":[48,12],"4z755TDizaUVyRRKw7y8DnTnnon8ksQYsZyU3feF6yFc":[48,43],"4zE9u54ZvrdkbBFS6rWVEx31abdH7GFoRCZQd8mDiiak":[40,37],"511pMfd4oivn6uE7MrcJ21hTvcaCtwPGTLgnQAfopir7":[48,31],"512wm7UysDB8PNwWpjMBmRgYHdQAoj7o6EDJ9CUyK2kb":[36,21],"518q2YT5TjpwZM3sLSTk58VVmdYkF86abh7GGyoUaHZ":[56,0],"51tQJUb76g83KRD1GBdtYq9NCZdrgRrJ2Nva8gdLS41N":[48,38],"525qsEebxz5jk6tEbYoUGJgrw26ttmi4CngP7BhS7vSK":[12,0],"52GEvaeCcEyAUKrfoPcey6vdyw6th588nYPuCkn3Kxes":[72,63],"52JQ5kmWuUN5ZVbWMJjJVpd3raNEtWRJMgxWp8J9mdv9":[88,67],"52rpdXBbJG4ChidZc1BiMU5JucsJQQa98zZUEUaP8Rwy":[56,55],"55nmQ8gdWpNW5tLPoBPsqDkLm1W24cmY5DbMMXZKSP8U":[4,0],"55tZynRDphTaxtH17x87FjcyJjCHCch3SrVxuanUJZmd":[40,31],"56Zc7i6DHP7BKWAy4onLkg2sDqV8U9TtJnWkNx5yQhBW":[8,4],"57DPUrAncC4BUY7KBqRMCQUt4eQeMaJWpmLQwsL35ojZ":[60,51],"57Nqrmi7wnUsvBdrkSpyfJHWic9dJqw1KpfgYtmx7XzR":[56,50],"586jjL8bHmqtNTFaXpajEJxY5mLnQ26e3QTHcx1Z8i5c":[24,16],"58LZCrAp98h2tebZq2Zpzs8zQJ7gFqyJMsBUb2J2CVM2":[44,33],"58M2W8tybgWy6pJVqk7tT7YF7C3rmUxVM4MWN7LG6m7D":[64,46],"59TSbYfnbb4zx4xf54ApjE8fJRhwzTiSjh9vdHfgyg1U":[68,61],"59quUWb5cx7sx669VWzj9umtHBBuRF5rpDrFrHvvtE4T":[64,51],"5APtJpidysCuKZCQQ49D2ba86NPNr1UNkGiDaehmSLSL":[72,0],"5Adaryyuxs39jqDsoke1VgUh3R79nQR53JRKBahuJSA4":[52,45],"5B8dRstrVg4NXw39yswMdr6ETHCsbKaSbWCAxCH6gofs":[44,35],"5BgjKeU8bSLJ5hrTGZXD8NrqY1DYYJFz5dLnD2EQRuEA":[68,61],"5Brx6TNjAkzQ4JjToEdL9sZjcFbNpGQRBgpbNFzXPatk":[8,8],"5Cf18uw63TPsS8XZ2gHiQKzxPh7i5axu6knFfAXFDEUe":[64,43],"5DsrdX4xPok2YNHUEtQsRuyAkDcdSBPXM74ezfRgy8Vm":[36,24],"5EamRRDR1j78iE2Q1TUmoDQRw59m2GTs8QJWtnTZsKf8":[76,71],"5EgBAoCdu6r5BcrntpAW2rm5j77nSxJeD7oMDS927hxq":[4,4],"5FLt2q4cZYcU5tK1zKggbGV6379hhZFKaRWMh5q9Xpc6":[108,81],"5GftYPpZU6r76FCJ7cn9BNGM3gmB38CRC4bfVHNmArA6":[52,39],"5GiX7EzEaooty8he3EJdNsLc5sqbT4iMe388cZfCgqwR":[44,40],"5GrycfarfnDuiKSfrw7XwWFKpfJekANP1RwrfQtmVX5R":[72,62],"5H3sMCaSJdN2k1hyuFTzq2BrZHUq7CinTa82hJS6EDTf":[44,34],"5JEE2MaWy1TMY8Xh7HLK7h4xJQZuAGgTCPTv5Fg6mkUw":[40,38],"5Jg8XRMaQ1FJbyy4YN3t3oCJeXUprHgQTqjysuUMQvbU":[8,0],"5KFXF9DS2ETQVfUTAfMygw6LNbqiXWXYr4y2k1kYr9bA":[76,42],"5KG9uYHFKSmJVgvXys4dKkZ1iVzmsHxDJWP1SsAw9ahj":[8,0],"5KK7GDAws7uYezSUcugdVrWNrKNA9ooP4t57Jq5W1mTa":[80,71],"5Kev1Y8njZLiybgnqTpTnjZ2H6NMtCeSK6J9TeqhyZnL":[16,8],"5KjwhvyQZMbDKfQCSa7L222vxWqJna2sfRKLXTPEyEwg":[72,40],"5LB3ieVKg5tR5w9VYLDfTh5u7DPbvDo5uKjoLQvzHK9T":[44,32],"5LF5MEkfKo74aX9zSz8sqLoKv61rv6bu7YgoLLkwrqJY":[52,29],"5MNLjn4p1bNUMRc7YP3rEWB5BQbzNsHYaqmQLwshAndB":[68,59],"5Mbpdczvb4nSC33AWXmh6wmDxSZpGRANNcZypdPSGv9y":[48,30],"5N2xAEANgsK8pgPrDktrXUGWJ3Cnt8TbPrcWqLWSAbq1":[60,44],"5NH47Zk9NAzfbtqNpUtn8CQgNZeZE88aa2NRpfe7DyTD":[36,29],"5NLjk9HANo3C9kRfxu63h2vZUD1cER2LacWD7idoJtKF":[44,39],"5NY69Bgoaahz6gRVgG4Ub2PBzgscsARgAekLx34Mtcv":[60,49],"5NwYJ83brnYDwQC8Hn9hYXhCiq6HQ4jZNGLZKdZeQHPS":[80,72],"5QAa4WEyAtEi7br4soyHSCHZQmxwrTbBy2JkWJdRPJc":[48,42],"5RhZ1sBj4bxxDF9JVBnxANjYGf1cuGMYbmumwuuuqNe7":[44,31],"5Rjq51GbTVY871gHZsLSknG7a2rqkukBxuanAJYDLVMY":[40,28],"5SAMpCcejTXQMnbrtkNv6nSxqaYgjRbk733QNzc4teJC":[48,0],"5TLhtuxkDdN4Mp2iHeSEZrDzpZ2xZRiEdFxcA9ipbPJV":[8,0],"5TZbMUkDaxxbyhkpgMQHZQCyvHAmsg9ZyDHf4R26qrap":[68,58],"5TkrtJfHoX85sti8xSVvfggVV9SDvhjYjiXe9PqMJVN9":[4,0],"5UCh3FzaGJuX2tBmHvPD6LXNchiGretWwnB9LqVif3iE":[64,35],"5ULHpStmbJSLYhke2WKwWRS2n5dVeqisWZJT4gjkNTec":[24,19],"5WhrU6gqgCwNBW7tkGsAZTB5bno3ymHVrmQb5yyxexBP":[56,36],"5WzxMJDwAwaMPwc9Te8TZJrFu2QmavL1S5SCcPn4VbgB":[72,67],"5Y7Rq8DBLwmDGgAUPKXyqJ57mRC33krMyH9dzMpuwTxF":[48,36],"5Za8eDus559NMWtNxwpWFqW4cNBuuVN6JRSCiRqdXhSn":[60,56],"5ZimkW45n4mWVCqXsqEEJuJWvhoqZFX7iRBz9jtHW3PQ":[48,33],"5aGEHgWCyHNxCcNMHP5TDddUkT5uXGpuwBfonE13jnMB":[56,46],"5aMayQwEmWD5J1anZgavXF7G6ZczrWG1h6UC6GJ11YeV":[44,0],"5buj3kmwRSZSmGPCbPrfbFmYZ3fHX1cjNe7KZSJPR8s5":[12,0],"5c7YoYxtKLC4EyiXBiREB8obfMyVME9zRHVtuMd6KhfV":[40,23],"5cK8WPnW9Q7rfTynaHTGHXHNRyZxHHT1iDH5LyPeaSQe":[12,0],"5cNCJuzzWPmSXyqDEhJL2rD74Xaf649mujFEPm4UzaJp":[44,28],"5cNEV9dx5Puuwp2GSZsUwUchCGfxchcxS6rNuD6yNGEh":[72,27],"5ciLz4FfhhGZnoGX5hgjnKzL5xdc1iNqmczt4moFTQu1":[68,57],"5dB4Ygb8Sf3Sssdxxrpbb4NFX9bMrYnieiz11Vr5xJkJ":[40,32],"5dLMRyPWx6rdPGZpZ7uuZiqry96dUT5yz48u62Gzugi6":[44,40],"5evLgbdZJG6RrZzeW5UqRLEUPjziqpPXmMKPNqdEmxWi":[48,33],"5fNzsJQxTQ93RJUgnGjvCQ8qjtYDGXjVMM8ERJs29YcG":[76,54],"5fnyGEnVu3nyMrUysGQLXz38QH51VNtmYGSA99197xCX":[60,46],"5gYY8gRdTyLP3TyLgfaGBP7x3phoCjUrRRz5JaxCeGEF":[52,41],"5gaASWLJbeYVk2Kd6shQu7JMVfkXHnLNwiSje6XrazyN":[24,21],"5gpRDdBffGa9quGE7hTPVCg9zVnHTS26qvbd12G5kSS2":[52,34],"5isoKqxB8G3CVngTkrHddmvjHhuKBiYZwLfWDufWZtwU":[44,34],"5jAKgxnCLVrb5zdDxjnRotwNirVG26Set4ZZ6BWC6Sx":[52,45],"5jLVeSB8hepuqgReNhcNypntbcD2wi54JZ3pYY3PGrtC":[36,23],"5jLw8DGMmjwaCJWbkT3dksXVEdWrXzQtiBd2TfsF1J1H":[60,20],"5jQqKbCAeYLiKK4WqppHhKBxe4DzDZMRLLaDhDQJ19F9":[92,76],"5ju3ywSUjEfWR6HgMmqhpDedDkcztspcS59T2EiXdxWn":[32,31],"5k1ooGz9ZPjGk8PbmA7Czgk4nS1Ns5CAKXSeJgeWqo2W":[56,52],"5kbpQzj1FEqqbeU2XrmEbC8gX125XQkdt9ZwYdBh2iK":[36,31],"5m28zJcp7CsTrH2szyNQhygvDis3dPwbgrtYsWi3J4jN":[64,41],"5nR5ktqmZufaVuK8N8nNoqVrQqopL6qAnf7YNvsjynhz":[60,48],"5nT7adimwUD2MfMRSrNKDoNrLF4G1mrpsStxzU74v4sZ":[44,27],"5nUy4R3g53WdtS226FdVWaVgXhJkQyaitSn1Duu15mXP":[64,51],"5nVDe1R4QW8XcaWrDUo88tG1V8CgAV2BqWpCX4mF49TE":[36,27],"5nvj4tHGRCRFmTaJfpjx3RUcNPtHv7dDkxMbc3yF8UGP":[48,29],"5o2kjsEZDYnWGfTqBJdrBnRYKvRy7wjrniivKwFqyTsB":[48,43],"5oEY8KESH5k8kB2WXZ1dhRB9YELcMfjs52UBmuL6e5HP":[48,34],"5oHWyQwDW2gfrry8iqyxYsiSrNt3PsREeVyY9RZZg3r":[68,55],"5oR5dh1WTi7ACiq8bdYmQN84kDG4HDQuX6cjyJErgGz4":[44,35],"5oVky3o3pNbZfWndUBJbxH82ZDqaUx7k1CorxfisKWZt":[4,0],"5ogMBk74DTpRaEahTtBrrsFN5mcZ2cfmZfPsJMhJm31t":[4,4],"5p3Y7UV2oZrSTTSLJzJknEzqQpetmk2NB2hQEKPc43dC":[28,28],"5pzJe9dfwsjdSmaeYAg43xyTsQHVP7zpLefhLsFxaktq":[52,47],"5qsTBZQPAPYsCBw9aPC6wCLpyPua7VmK9yFWk8gLQaUP":[20,14],"5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq":[56,40],"5rRT889dQehyRd44HVm87UQ2nTko8QWNKVACbkWBjaZ7":[32,8],"5rxRt2GVpSUFJTqQ5E4urqJCDbcBPakb46t6URyxQ5Za":[44,34],"5saC4V5Kk7Xr5zUUbQZHfXCrzFWyW9Lvjq97N9ydX5nj":[44,35],"5sjVVuHD9wgBgXDEWsPajQrJvdTPh9ed9MydCgmUVsec":[48,38],"5sjXEuFCerACmhdyhSmxGLD7TfvmXcg2XnPQP2o25kYT":[68,58],"5t5yxCvtHxCkDJCCrChBQ2hdcUrK61tr8L2QRHtbnpCY":[32,27],"5tR48Ewee96cx6MNBXZb3jNtTBWmMimYxqZkj6QYHgdt":[52,38],"5ueaf3XmwPAqk92VvUvQfFvwY1XycV4ZFoznxffUz3Hh":[4,4],"5unroM4ZHe4ysnprhGrsHBUMsCbkfAHU1Z4rMtosbL26":[56,44],"5vKzPeQeveU8qnvgaECkdVdBks6MxTWPWe48ZMeC6fdg":[76,56],"5vaCfp7UEpW5qdJYyVH4m93oMzzyzTqXdbr7xLGobY8q":[64,20],"5vcZ2tAziqSyZdJJgknngPW1ngnZ3R8bjSqdeb5mpCzh":[112,90],"5vdpdDS5vvUrPTpGq8zDWmirYheKHq8RWrQfUrbarN29":[52,44],"5vfvM4qv8UERxSU4qjKhcyJYgfvBwxM3zotkbyXg5z4z":[56,38],"5wsN9Q4XLXvxjefK2tszV1z8DRKSXyGo2NxvzrftnDQZ":[40,36],"5xVTCAt58jH6i3Zkr3b3EDyxSP5CanqY2tRGT1MStVUr":[48,44],"5xjnsTJwtYXWNFPwV3DsXmbsi3oz4bZbdSukuEwoYdbT":[48,42],"5yqmdjMVX9F64YuE97neemY6Q1s4MgVaBbJiz9g5qGiC":[40,28],"5z9kpN7JLo7HkmSYLA3dV1yTuwx67WWFg8FiDEk8kJ9o":[48,47],"5zELkZ6RECLDg4gi8sPvnGPAQbfNKUdSPunPh8HMNn5V":[44,38],"5zRUbp1Dtu3qQaRVf36oMDaeH91D2ePnc5DEgnh1ivFg":[52,43],"62RDaC4ARQMyZhdya46zuvYs9L7TrR43KsochCnrMVm2":[60,55],"636LepbxdwpdKNpzuxc9vJhYLkhBQgBJE9yiwrbo4nrc":[4,4],"64TmBCNgzNp9StawrEncNyz1TQWuPmGiGuDeL6fqkAvq":[40,31],"657Tpmj8yRfJuj4Dd1oqdKA1Lo1aTruGeSkNTaSabHAJ":[32,23],"658mqpjmxPwfu7gm3PYPfPsGqMagn5FeiJ7814YWZsNQ":[36,34],"65WF5UBc17DsJ9yMdTejNuU5SpeV6nd35DxNioQDarox":[44,43],"65kBAfNpGpJwShoLbGfzJvWwzeZt8UPvZ9xHbjXC6rJ3":[48,34],"66dX6ZwV4W9etrDRkBvTrgiz2BWxogjELH2T6Pkf35bz":[40,40],"66puVpH5uFvjAHAnAQma4xzmUAUuNCi6nRkWhtRwKY9s":[4,0],"67VDb2iEdx6XjCfBLXhUgKQQjTuLe9X2eLqTq5nBjUTy":[60,44],"688vLxT7Gsb4YX9YotViUauLC5aYbnjm1SQtaEQUKitf":[56,46],"68qujN79HiknCPBbGESncjUDeC8V42DigCGjQjpaVher":[60,31],"69SHUke3phQy5bEEKSVaSp3ytmEJF8h4Yh8FssZDHNjE":[68,55],"69k73WLdHRge7E3vCUiDx7Dkm1DQSBBGAu9FqNj4AeJD":[52,28],"69y24KUYmXFY2N6BMfzL8TfiKjQtBNCCjtnju7bxh4zG":[60,56],"6A2t1aNmY4c9DsQuZgjMBwnigUpCu8vihugqgyAhGrC1":[36,31],"6A7vAYUkn5wKnUqnf2CJxg2kmbtDidhvF3nf1DNhyqfj":[40,36],"6AaA8HJGpYK9RDN5NQjDJfHPcqX63hnw3NXEa9rTXbEs":[40,34],"6BdawAEJvEbgs7UB3VsKSc1WL45ydCR1xqi9pyr9JS4q":[44,40],"6C1mHAPxQACd8NNS1D9KpGxqSRUz5s6itsaJx1uteofx":[48,35],"6D6oRzvSE6cdpKpWgojgHhrc5ef2jqQhRQywtzp5GreP":[52,49],"6D6puBzRwMwVNZUuEipFycFL7xZgL9sPEnj5p68Tn8iP":[60,39],"6Dr57RWT2ctMt2XiQxj9Nec5mBrfjucfAyh8hWQE9cp9":[44,41],"6E5NygCNcfyPHkLbHMckzF25cgQoxN3DfMqH9bwyQRpf":[36,30],"6EPcBVHgAaP1x6rYkcpuufSMoP2jhRYi3N43PUqosTDj":[32,31],"6EfiVm1bAo8yWgZppb5irTqciv5VC2eoNTFToST5c6Mg":[68,45],"6En53HMLJjuqkk54LgiqdEoSDXfNzqwWVLiF7sBWWFsF":[68,40],"6Eph5j55RgzMx5ogbpMg27hDW6yoxE2Tr1EMc4SqpKXp":[52,36],"6F16m2H44H4LHseHDuk67k2zdEXCWQdGnA2BQ4yMQFMX":[76,46],"6FTLATh7CDdqkFyYJuTR7oFyvhVK6UHUK92fELg2mRno":[68,53],"6FWhS2CHjtCf81GMsqHRXQqDUh3UKyyWGF15QGCWWb7Q":[88,61],"6GA16fyWGrr78QWGUBnH469A1dWNv8yTjuVMuDcxjxmg":[68,58],"6H9aeo7woPe5QabarbwHtkrziJii8x6RT72Eroga6o4Z":[4,4],"6Hq3pmDPps9ybX3vR5RLmJyXYa17Sy4ZjREzfRbrzTe4":[60,47],"6HtPhr81VwVaMKwFFzML3RrF6PMcjipbCpPT7JbdsPvE":[32,31],"6JUvAc4NV51SfX8G9zwoRptU6hw1eC3fYz443Mh3Qj7w":[36,35],"6JaKmYstgSj55SwwQDihDq1Q251FrL5ev8XNiqFSP2qe":[40,33],"6KGDh5hSNAeDmtF4tD72Rw5WtZk2efYxqgbVryXmis1J":[44,42],"6KnzXAhpE6ki8GuNQBqpHsVdHhsyg5csChPGLdkTHRvW":[40,37],"6Ku6Cj3Y3FETU6JEuwjpLLu65CnKhi5YGtKdUTRGud7i":[52,21],"6Kwr8fUZPmSFNWaXfRL7e7v38itt276DFVu7RiYn8oW5":[52,0],"6NDU8PkeQhH8DF5Yw1Cn1AexQoQLnqr13GhNuQL1gfuT":[48,45],"6PYMaoJf89uNKjUPyf1eUh6KQ8vGAHt9Fb8EK5SqctKK":[48,20],"6PdBqw4p1iaNd3CYg18THHpzDBuophRUk3qSFy3KNTuD":[48,40],"6Pjg2CBN8tG3u61t1hRQym5aHwNC55DcYV4ypWkpBFTa":[44,37],"6PkpXpQeLMp45TC1PTPUhCpywcCxMcmUXvH3dNPMXQvo":[24,16],"6Q62YX8UpQKABG1FANUsTjJJdrfYZNEAbZyceuaVbx79":[48,44],"6R1745AskhXmKugSRxBPsJPDNggk1nhJZU1fW6beefBQ":[76,54],"6R2SsTxEK89a9m84Z666c7M7wGcmbwNmTCyPFcAcftyX":[60,53],"6SKzCZ2duYHPgDgXKx4ZG8pix59q6WG8gXgyYvgDaNR5":[52,39],"6TxS2SvBtJDVvGADrjYzeTczsCezuknS3gvMH18hpDPF":[44,40],"6UNg56mU9BP1KghPDDgF6v82iJHyLaPW8BEQb5xpFY5c":[88,67],"6UfQVpJKcGT3Cmo5DvUkrDEu9ucHbR3Y3XCs8R1wyNnM":[48,39],"6Un2uhrFfW1q4a4vhQfTxT8tn8xjpi2CF4kXhfaWKHSK":[40,32],"6UrGCcP3H5REdZrPx9X22s8Pj7q2RzUWVT5LFCLBevZ9":[68,58],"6UynSxu2fiY5qU6Ae8cPLxq4jyWVpnr7o4fUyWLxCpcp":[72,30],"6W2xi4iCGU8eTMCGtG3DQXgMGurXFnd5iVXCY5Sq7AbF":[48,38],"6W3xBXKnq4vGHvBjMNSgVviQ6vqDeWiL4LwnSFjvr8Yo":[76,62],"6WHCTDvSa47muoyi5zHoKKPcodkftixsauEfDNB9YSjL":[24,17],"6WJGoPfSX1VNqfiYfRZ9RP6KUm8nLtar9zJxAntdzEWh":[48,44],"6WgbvHvsBkWoUf11RjnStUh13Z2WCcyzs4kTyCooLA8e":[48,39],"6WwCWBHYvNXnDswu6qrHbKoXMqtB1ZwRCD2U3oqWbZmB":[44,29],"6X8sHQkmxRVh7oR94VjsfffmQYPoGZ62Fp8gt4QivszH":[48,36],"6XASEv4VzAyPHmzJEc1wTMZsYLB2ov4ggdzjVY8Q3Uuh":[64,47],"6XjUyrr7fscEiNoQiVFRSoyzzwE4rMYsvQWkZBfrpdH4":[52,44],"6YCRJFeMhPkmpYFaThBktkrWp6qAopiwqHCaj9x6feHs":[68,56],"6ZEbKFxTjEKGC9HUqzy9z4ccJ8Aq3ktPKEzHGDosQJo4":[216,130],"6ZR2mZ3r5oKHCghjMK7J4b61QEHy2b4vqW7k2viGXKLr":[56,51],"6au2pU33RmTdpoZ9WcYrHnTmTByMJbMMPmZPC7Z454hP":[48,44],"6cbkU5eSmvbsDaupLiQxDNSB8YbXfrkzrtqRMLRr9ZLP":[48,29],"6dr7c5k6SsFRFfmoNqADxZQsvPjPjg4meeEHVX8cn6HU":[24,16],"6e4BdfrD42d5FHVbsHrKEqxj5zzn1Fq5bC9e3UwFr4DY":[60,36],"6fdimp8ks17wBAEiX7MuF9CrmJZ7vLGtThXdFsrv7Msj":[56,52],"6g7urUx43pwjUZ9CBD9c76oLQtpHCgCxp9hQhv6RUMB":[52,34],"6giEzjcXWwiodVL48LtoFexax73cBorvq4NM8a2xUkd8":[48,42],"6gjvikJDNh5nqE9T29KaLBWUvc5ouramxzhhNHPwao56":[64,49],"6hS8hZmb9KEoAuF1pXPPdpJBQgsuepYyAhghhAVb8zdE":[28,24],"6hcoeM9dVx6x3QMppHAZwsiXYRsRECxWQJELZ8pYv79n":[72,63],"6iErFcYDqdg8MQcNPZevBJqdGsDQx29FrNDeAH9zH9Ku":[56,38],"6j7DvYDyFTdrK99apFuuT8w2WaeaezfwLDLk8Em8sB2m":[72,60],"6jWgPvH1U6Miu85ow45tHyvEXfjmPuNDFAjovCHRCLid":[36,0],"6jYfBBZderjvyXiqxVWNP46GSkDZmsL12m7D5A9tfaND":[40,30],"6kCRppFT5zDZ8P8L2ex9m6yagTAm9e7682F51ADnZQ7A":[52,42],"6kQGi1Z41F1Kq8Jqn7AT5fUUw5NQQshx8sDwCSznRkpv":[52,6],"6m8LGKXMT5QrRQdQsQAd2VHpYwJZebbrc48WgkPWeRYc":[56,28],"6m8NNtRcrBPvHXfyMQqSYPBpYBNq5Wn8K8VrW6qDv2fB":[56,42],"6n2Ebk85o5BAQSEbkukekWcaeJStyKg9NWWUXD4xyRFQ":[52,36],"6oB8HATu5ApWMWhFpE4Ms5XMNcKmjk83VpcND5U1vHof":[56,52],"6pJAzQhw3MJBbR4BPzhWJk2Hf5p7idivRrepCuh1BrEu":[36,22],"6pU8UoMLbohFiSAoHeNtpRPkifHZpvSVUQngkTuMkdZU":[44,41],"6pUayMw7LVx31eA86LAxomnzqktGX4rTLvDzznRHDuNh":[72,0],"6padcv1em6AzRFRoZ1UVzHEYo4pZ1Quu2UQqxecvZjVM":[8,8],"6pjd2Dfsv7FNkNDCGhzb8vn1DEmvPSVicdQxvGKLVQwQ":[44,36],"6qH22p6RDZFeqszT5fAh4LZEzrp4oiLdNRfHPuetHu6A":[4,4],"6qJPxxgZHCQKBvbGC9zCsuuPtHMMLszVCoiCvEhVULyJ":[72,68],"6qPPKb2zC6U9g8pwAGrYJxy9B9noYiKxwS7NnuRPqpUx":[60,45],"6r51uDXPZf24KrxGU6SnWCXfdickihjQjnTtUeQdRh4A":[52,38],"6rDzQVov7rYNcHSGsVcbQny7VYCinkn4U86Cfz8xYQdC":[48,32],"6sMtZs114UsqAabDyJYjtgD5j92HGVvx2pyA6QnkidWM":[56,43],"6sSBHSuyRRphvkH4GAwccGRB8HdZLWC9VENN3c6S39sd":[44,40],"6smeNG6M7Aers4Ju1drfZPYBS4WFK89EwWphQjKTMQSj":[60,47],"6tknFLCJiEwuYemEmjCSEz6oB5fmQUtH199Q4uuqcTY4":[32,23],"6tptWLfq3o2Q5M74ZJLpGGpA9jCAScHSaun5aQVTtp1h":[60,38],"6tr27Z7WhFGkFFFNBmaf8bHypvajZU8WDJPR7Aji5stF":[52,0],"6v4yeawkVLWpxrb1pTmqtrhYi5ZcQBZtQCvgA6MmRKLg":[36,32],"6vMTzyzXBztMW3Cj2j4SbpR5cLmiG58Nrku4FDmNP7FG":[40,30],"6vZuaLY4n4GP9DVroymfZ4D1oP6xpgF1ExLMqHQbt32L":[84,68],"6vx5vGgqAa9dRaJpbViCNDjzxp6EyGV38YMYbNDqTzLr":[8,4],"6w1jYS7vrmprS1u9cQd9uFo58AZYvJ9JtzihmkRPgSz7":[32,20],"6w8Gxzq1AusnWxrnBH49wkWVemp7MPxXftfyUQy67yJZ":[8,0],"6z5vLtFS6vaV2tY6a6Z29fkw5X84rEqVrtAVhNq4HN7w":[136,0],"72TVuWZN99RZNkEnjPWiXVhrjAirozA7baZo3jEve1zP":[48,35],"731Lnc3mbXpquV8FmFnTL8EE36uoZRiDMUKXehwZq8x2":[68,24],"73YGZpfTSBv7PBLvmgcAwa7K2fAaBoGe2Z9YNWz7J4rB":[60,50],"73YWnZRRabUNGR9NcrEKYdQTonMGWdXHegvo3yphzS23":[36,27],"74TbcQoVmGdmdZdUZTEpMaLonAtKTK29GsZpfn1LWxoo":[52,33],"75A6FVv8hAZn3n4KsTkURtQP7GDU4SDiZxcTzkTHZM3b":[4,0],"77VrLXNcVnpPHcMCcfkNU8SpuPzoCbGfp3Kqc2iV1dgu":[60,0],"77s5a893M4MNWxGepG2GympAjBk9abGkEHZbeGR4pc3V":[36,29],"77uXenX1Y9T2D1pcnHnYsYiwTTHbnzkyrKX5fQFMGVCR":[12,4],"787PK2WaCUZCyYEmuYQSGmoxu7MyqK1usn43FfiVwhcB":[28,8],"78QcjcDqBqvxrjLZVH3Y7vmyCmNdu7VSVnHGMiH4CpcR":[32,28],"79VUA68hwB2XuK654vDgjVP6YU57diuhHiiZrrH5PuCZ":[60,34],"79deTJCsgvqwNZWwXPYxAL1eYE9z8NArP6rkhsVjbM5z":[56,43],"79u2h3dDBUA5xNZwP6GF2cbtzHotobTWuZxZE8DjEUGJ":[40,19],"7A4WJBegWKXVMhVoKshm4GzjW3Pb9od9ECWxF5DbrSZu":[40,26],"7AW5VGSNcaECGKJD2C4rpRuWpcT4kdAHrbahc7KFQM3p":[24,19],"7AzrvKjepC2ohWw5He91UhcwvZwYeHUpXYxma2XkRtE7":[64,47],"7B1wzk2EVeWrorZzFAqX1czbBqB8mV7s5Du8YWCLsQjX":[44,38],"7BS1RfipQ7zwuKAdiUX5CNFCKNEdk82TN2C3CmoXR4ux":[64,51],"7BncGLSgSexiAXz1dRB7cZEDdkKey2sK5xiLpHESDjpf":[44,27],"7C3FrWyhFGc75WgccpnpuuCRSqpZiWpvj6d7U7jScSKU":[8,4],"7CTdZm5CFoWy3gMUhyqG4SMbf1EE3M1qeqLctXVp4ucJ":[56,37],"7DMvXPALnEYhxKqUorHgh13xzBvSFu67dCU2MKF2gsv8":[40,0],"7DyCSDDKvRe1BdxSyN6Q3bvW72VddJbJXG36Ghi8KRcZ":[68,45],"7ED12uoR6C3mr7Apf2x7YnSmEHkApFo4Jfm1bq8i6L4o":[64,44],"7EMmiUb1TxdX6G7B7oAJ3jSsXMgjg7iR3WkqDq67cCKg":[48,34],"7ET3iuFkJDTdWgetabdoZNKgWeRPmAaELMz7LbiHLxT8":[44,29],"7ETjs9tfe3snSKSnKqzxJJHmpNTT474TfcYG8MSQnuet":[56,46],"7Ek9msWDDoe9wSgD9PrPA4Cnm8fVqjvoz3UK1U9LFyL7":[60,45],"7EucomZSKvQdiZLvra8hLszL1kRYiGewxyMJnyyzdbH7":[40,30],"7F2vcJca5ewzdJUNcVMKCLVYneq6CX9JFMH1U7JeVG5":[64,44],"7FVCgatxKrX34VwM4YRhUVdXsJAoB5Kk3EGWW5M2Nqub":[52,36],"7GM9F3HAJeYSdvhfrg6Avq4sw5HBrqutqQqDT7jHMDHf":[28,16],"7Gn1fiPLJp1eb8g8QkyZ6eLvHgZbBnSA1ZTsicWPchcV":[48,34],"7HSAu6Q5LAqrCk7pt649utsDrrEd7yP5NEcqodFd8TTb":[48,33],"7JFfCpPEodnt6SWY41ePBRXR6LUGiKhLSKJNw9ZYjdah":[52,46],"7K32uTNK2zJwp5WTt4t57qJMf1JnHBq2HcSkc4oV5sQb":[36,24],"7KvnzA5iLwJeQ2q84FA3K6ZAKSWwvyBXPGPqcEea45Wu":[56,42],"7LccZ6QXtnSMLqXiTe2tDnfz4WN1Z3B5RPYsEsH8w1XD":[60,23],"7NNpLpbJicSjeHUXiLuQy1cmnHNthnsmosbLes998KqL":[4,4],"7NPcRcHu3jACoQf54nkRBLgdn7zBbUYhnsdC4VHqBQwK":[44,20],"7NfasVzGcyPzcZhiKdn215iyMTY47Gk149TMYApw1Cx":[52,45],"7Q6g52pNXSqmSkty9TzFotJPpPSgnqqEdSYsA63GdogQ":[44,33],"7QqYu69Sh5WB58JjXEmnuKVfvybC4dxWTsDFmNqiY9d4":[44,44],"7QrgfCAeoEhfKSfmrzbYkrHCydCBzEqAgNnGTHZ6vmzY":[4,0],"7RDFicpHYkxPEJAA2FRqtmyCew98m8B6cASzgtSir7mt":[56,28],"7RTTKCWBJ2XwtSHkUfpwBTH7SsdKqHrWfnD9Dv4z2Wyw":[72,49],"7RUobwC33EbHaWWR2sbdaJhT8x8PpgUoAbJQYrQqrSgQ":[48,33],"7RYAwsb91ZL8117TaYuis356M9abKddFtaSga6XpZhjx":[68,48],"7RrKc8sshPPHMkX6FvrxtQDKFytPiJN7NXpmBRXsD62h":[40,29],"7RuUBtHw3j1ncgtPi25uL2ZawHh61iYg6j4KK2BRi9cz":[64,61],"7S1xGwMrB4x5fwhchayjHojKQoCWZsd5HnRHRUJGXekR":[40,12],"7T5ZekSsBSgLNKVzQmCRQ5iqL5ycprREa1tz3GYmb4eT":[60,53],"7TG3LLqWYn8ybpqAJaiop1bVb5mPWJSDhxPaLzUdF3M2":[64,38],"7TcmJn12spW6KQJp4fvvo45d1hpxS8EnLjKMxihtNZ1V":[36,0],"7UUFbQSderHWPqu6BoezL27ymsgrBbXSi3qQHAozwDtP":[68,56],"7UZAaZTjnsFMze3RWtzpxTG1CiJenrvPixvVxW5xSicN":[60,45],"7VV8eZcVAN79xoGL2eEAj5sXVbQEsiqiTCZcbjisjXUx":[44,44],"7VVYonADe1jj2LtKZMfTKNPiK7gjVRDsX7dvA4BXf9sc":[48,30],"7WH6nrQ1MY9i7pLduZEAuUQNFrSDZJTxQWN4XfmD3G3V":[48,38],"7WNkC3cjwngUYWrAjEwXHgWvk3S1adKJd1JK5FZkh1s7":[64,55],"7WgNDqtFHr1hLYo8wcw8X5uCnGwDSYQMz7MMKL6dyLLt":[48,34],"7X3csFXUN2AZph83GC2FZCpkCTZXfVWssaJ72cpwG96w":[4,4],"7ZbzsvjLnAmu7sq7SFzx5BBgrqWvEtcYMqYWrLDH2biA":[4,0],"7aG4CFBwtm6DY5v6bm4ZudBfkjYjiRb5ix6S8ptsxDEg":[24,20],"7aKHeoUDCYbEYdSEj63i9m6vmkXLbiafWxoCvyhcQtPw":[4,4],"7ajm6amGXayr8qe3nPYA6j1bPMLMCxEmhNdzgz1EjnW4":[4,0],"7arfejY2YxX9QrmzHrhu3rG3HofjMqKtfBzQLf8s3Wop":[44,38],"7atMLyqH6yHXfTi2zMXiTrXtXpuJTsqoPrkD4N63ddv1":[56,47],"7bFx5g3sh5CqupFYtch3J1RdZBZs29HtpXAWyPPyptB3":[40,30],"7cY1beonNGzrqUk4pNWErm2vYcyw5yyLqwnrEHr6iKmu":[28,0],"7dEjSFnrm66CJ7Aj5mC1hsYmMzmGgWPr6iZNhcvANZ1w":[40,32],"7dKiopJkRwh6yrsBUnhEX5zTNCDWQwBRsvbD4fTdrUVc":[48,36],"7ecwp7vwPo5b2MNbx75yA7qxKeDBJDpQMzxWKYrM5rB9":[68,51],"7fStZnqGYrsdYdZaNyPEzvMun8AfC2YxFznA3RukvkzD":[52,28],"7fc8chLC68vNnvk1yMuAzFgGmAookHi3E5tie5FaCiuU":[16,16],"7ffVbi9Hsq5scgmBEjiinVWEcKAedeUeqxXMBoBK6JAS":[16,15],"7fv6zGstESoyWYdrfeW1DzN4fabJm3M2mRUVid6bx4EY":[60,45],"7g8Dy1BWrG32N3Hx993PBbpyrf8gTBWUEc9TpeiSmrQE":[80,60],"7hAddyJcvQAS6SsfRKLJzYPuq4h1XykRSJEUmr64p8oF":[60,50],"7ivaGH6xo5sf4hACXVNSqeeEDeiLFLTEyqQQwffDH133":[40,31],"7jdjBzSKJnqHZ676UURpGTWXteU9mr1rpBDW7qanTwsj":[48,32],"7jeyRTfzimBh4PjYbFcJWmvi2J4ieyVaaziMxc61LEdu":[36,19],"7jgcNRZbKcEWGZndJDpsgB6HwYeQ9u2EGnSX5meoyCbw":[32,23],"7kmwiz4wbzf1kUSZmKKzaJRxybGeDMLSqhR9s2FebhoY":[52,38],"7m3rgSgyS4HXnBAc5F8tPY9PTXgB5wLNz4xC8b6XA19z":[40,32],"7mtKMUgM24GPTiR2krRimUiQgXRRmMPmmPkQBzMZak8a":[36,36],"7nuWizYQEpkytGTfgEjrvoFFNUCgyzfvwLr3wbTJotym":[12,12],"7oQ4anNmvJmXUXu6pkXFb5fmovuPSdWzbRkwMBvHi4yf":[48,37],"7odF9faZtwCEATzPFjyFHCYQAgNdBxGzB5znqEVJzbVE":[12,7],"7odNkZmb2sG52sojMv2n2sBXsarYNPvgYPSxV7XJmSei":[24,12],"7otSffb7AdR3zoM2DkrwhhhfLCLhFzwNtn6XxShnwLmL":[36,20],"7pKWpEQLzie4ZwMWaPpcv9Ko66hv8FP7kZEx6w5wgujc":[48,32],"7qzobSMqm6JQns146x94kDeFd8BSTxXtFwCKGAbj3G2c":[56,48],"7sEpbQB3Dryn5JhQVCGWoGgUfYwNEZzjPNa1Tu9mVa5p":[44,32],"7scarR3Z5obfefZr8bPKYoMNipua43K35AJAc1YchQBK":[60,45],"7sdh5QHFPo4ktG9SVTPM7Sek1WLZpwxNNHudojYi8dK6":[44,28],"7taXjCmy78gNRNii8bKngFXYMtPTK3AAGJsrpd2jxe3h":[56,4],"7trWtWjH3cGfSu8z6MgkqEEuCJWN5NhRZBYvbT841Yi5":[56,39],"7urBmScRfdSH9CpQ2SAwfmvGXp59nTDx6Bw16USJVvGa":[64,42],"7uuvTX8C2uys6EcVdoEhESFc6Akd5wZLeAerJXrPcdzH":[68,60],"7v1L5zcmpYyde5s8hvtBSrdW4V5t8Qef9RB2xTRKSqyM":[60,50],"7vu7Q2d4uu9V4xnySHXieeyWvoNh37321kqTd2ATuoj6":[40,0],"7whPeUt57CtDY9uG8WdUaNxiZ9kE4dkxXCYgubpjnbLm":[56,55],"7wsxae1rHhA7x1329kfhGKzukq4Ujhw9D241ziBxdKY7":[44,32],"7x29aMXJ3kxxTXeU7ur7NpLFWCmedz7LFVo2oUqYA7tY":[56,35],"7xAPc3WYomPjiGA5PAwyLMFUNej7ESVsPnXaUqMyrqzE":[68,54],"7yRo8i8dV6MgNnUnzhHvgwExxKo2HqLHJeeyK7AbL6ME":[64,57],"836riBS2E6qfxjnrTkQdzD1JkFAoQDyUjTmzi38Gg84w":[4,0],"83PWQUxkBDrTJmJeFL8VUah6BK4p1JPGdXVuJC9Vf2Pk":[48,42],"84GZWtzfKYX1yfstmjA9eUEp3RnWys8DmsPjsd1ay7wv":[44,31],"84eZv5dPhXSRC7L2yBYW1XBLLZUyJEeBD2Gz5AhQkRFS":[68,67],"84sCKfepG1RZEfAghQfPi4fK33skBx8FuXdAtbPLRDQQ":[28,28],"859Kqd71jDi68MHayXjQJZNy9nWBt4gNs7HomAfdNurH":[60,45],"8641M19beXr6FB4zaf6GPYdLaV695xikBLYFYTVEBZdm":[60,52],"86uo4MtfpLrW9EKd7pxcyKPWBSPQ8jEWYLj5MpVBigFk":[36,32],"87VQhN7dUfS9wacre7vqRm561bNUk5PwUB8xmroc2yEw":[44,39],"88WgTxiDDozTbmzyZXXE1xyv6AKvHUfbJ9Tpw4RmavN6":[48,30],"88ms3Y6Z3pNaMrYY4zdUwHp5K12csjNeffomBnAdyaBr":[108,72],"8AgqfNWYTzmtoxRAvqFB39Z5kdhoW6BV9hYjW8Rs8NvF":[44,34],"8Apz17FY7vts5PUEP28apzqQBVgg6McbetFJqb45ew8F":[40,29],"8BdJzya9vC1PUZrciyMDadChSiEoCLQo8m2yezEzdaXz":[48,46],"8Ce22R38MddAZSpEhLC38BqUEzVAcZh7h9MgfVCWibN3":[44,39],"8E9KWWqX1JMNu1YC3NptLA6M8cGqWRTccrF6T1FDnYRJ":[76,66],"8EUEa9h1NiW2GLrMmqJtDXYer86ixtAgFSrBy3jU71s2":[68,27],"8EqtKHaSgPskksNFSC8oWzSMT2mdSMMtNjGZ7E3KHxSn":[48,39],"8FRFYPcwBan1KBKR6HuPy152L7pr3ePVYVxXXnWzPjEd":[48,24],"8GaMqVpXH7JuEs8D8bdXpe7ztUasAP3wdEXpyZZbUJeb":[44,31],"8GpsptdhGCGybKqEw19pVBZg3gMaopiKtRMVzJFBddfB":[52,41],"8HL5VpqGfTG9SVkHyWU9gjo4xdQYbbdVS7ExTrou3zCE":[56,48],"8HzsgkGhEFP2MKuuPDy5f8qvqR6hmwPqeq7UMY3X2Z6T":[8,4],"8KKQ4QJ7JWAosHwL5pmjKpYWMNSxqtQjJVes2hQezNRQ":[4,0],"8KYQAb2TqCq4Tay6rLTVwWr6BfSMtSmn6E2qosc3xm1f":[48,38],"8L1k1DCCwRoZVEVYZcUzLht9SxUBhkNw9fU5PGnZfw9u":[24,19],"8LkSKTrwqFgw1Knh4gc2BFYb98DnUJvQAxhT1BAFYh2p":[32,31],"8LknwWtMatn1uRenrXYzJS7MxQJXZ245dqTuQiD7wZtq":[40,35],"8Mhs3hgdpL7AKns3wSebNkAfFfqycwd7K651WwSsg57L":[64,50],"8NckKPrPLY4kxcVqD6RV1EVaxyKHkf9DBQe9cz1h6Q6B":[44,32],"8NgvLoYGP7wyramK2gEzS4sj5UKpRVHZeTUSUvMPMna5":[48,35],"8NkJuAPAPTyb5VnUpjdjepHPiD6GR2dT9BxwrmWtzYkf":[48,41],"8NndwQsrH4f6xF6DW1tt7ESMEJpKz346AGqURKMXcNhT":[48,38],"8PTjAikKoAybKXcEPnDSoy8wSNNikUBJ1iKawJKQwXnB":[32,31],"8PUn41CP3VdK15qzrxz1BvENBCjV3LDeSxunxRnRtwGs":[76,71],"8PZNvPTVy3irci4T9HGMFfxx2jiCQov1FWgbWcjps5t6":[12,7],"8PZtnhmPgASnTbefTAFvRPJDR35ivLkEMs4qjfV9LAEa":[4,4],"8QHpiGmDQagfpuBL9jqcAovE1Xihv2nhiyLuwdsemVcK":[52,38],"8Rd6twX9XJQzo8LTshf3Jty7kBQdQsGe9dfLia4vJzfW":[56,48],"8RsYRsi6f3hiK4EhyLS22Cy5KkrNbuidVYmsaYR1Xx78":[1284,1063],"8S4Xb96cH4sNrnKfMDHd4HR2bmjWbeUeo1o6yJC6ZGkY":[40,26],"8SQEcP4FaYQySktNQeyxF3w8pvArx3oMEh7fPrzkN9pu":[64,49],"8SRKNfvMerfA1BdU79CAwU4wNfjnDvFrBo3o5f5TS4uv":[52,46],"8T2ntjCMtcb2zBWmL1BiVD5rho6Eqk41SmMm4AsbuDFi":[72,66],"8VNj7K6ssFcUogRfT6miUzz8HTKu1nX2n8MYr5z49CXb":[36,34],"8Vh9GGKGLQUcHykALzQm5UAb1mFKAWxf6WPbFaxeWWSF":[32,24],"8W8zc3tHfhGbaZgQauBY17TxAQa9mTixmfqbf8PvtYAq":[56,47],"8WqBgoVXkVggLVuvZuF5wP8taQpzTuKGoK6brU5s5Hh8":[80,55],"8WrESd49NkVEUPhnq84ZW3EgmvMWEX6TrNYpjXLmNNHf":[60,39],"8XfjVuniQUTmxiYLJS6AeKDyhajxNqBf9UhhLjaFNcE4":[52,41],"8YhhsDRDEQCi8HJxe7MiXEBhCXeH7LHE9XQYLMUni3k5":[40,10],"8ZZrpXzvuVYPw3HYCPN9GNJhegj6M4pMityCZnCLfVUk":[64,53],"8ZgmpBG5ixt4LVRQEK538hsKTsJBgmFFH5L6X5e9iPTd":[56,51],"8ZjS3d1bQihC3p5voM8by2xi5PoBNxzTJtaQ9rvxUbbB":[52,34],"8Zkx6veTUXdfGcF4VgBJtgZCRnhfhRA7yfpCM3Xty72Z":[4,4],"8a9njgsySJ3LUTvHHyCChKajgZXDoU5cSXkrfn9gf9Um":[52,51],"8aZtHhTNFhVWp4fV3dUfBwsKKBjqzHDwpTZRbpeqo7vo":[52,34],"8bRnkspqHntTjRpnWrCKZpc6pVwChAjUhtZrwUVPo6NN":[64,52],"8butsrHFxUZ75vKbFnyGzvb2DeVZj8uDynWhbv1L6cSF":[40,27],"8bwU8pTjJdC6risWVrmBKvo5gVQcN86djN3CQWYtYrAa":[16,16],"8c475ek3Geh3X5hhCr9Cb61piDKedhMrPo9bkziRqpah":[60,55],"8caQuNVnmywtQnKWv6j8MzzJ8mrLwJkeGcKEtkQkoFZA":[16,0],"8dHEsm9aLBt7q6zu3ESfRXkS2eCwkbbzzynfd2QxDzms":[32,3],"8dYakfEyJqBxobSp6WsSPSe2eEQs9oGHtgJ8xvHbKYiv":[88,67],"8diJdQj3y4QbkjfnXr95SXoktiJ1ad965ZkeFsmutfyz":[40,16],"8eio1idaNjEmeaHUbmJYJDJXAzXm7nNevp96QB9vLA5q":[52,48],"8fELCwf8vTtWJShtMmo7YoySc4CokbsQvm1yptQGwV5G":[48,36],"8fLtWUfZSpAJk7h4XhvM6TqGjXQxiwzWkymxmGtJoGdu":[72,44],"8faCuTioHxq7DYADQwQeAHaKXjqBzELCgUQBieXhmKGb":[8,8],"8guxGZ3yR7L2pBtXgoBnPpq2RE4GM5qvK8UaMG5YXds7":[64,42],"8hpUJeGB6BF1JTZcbiNEgw9w9fdQ8dEi8jF4ohapsq3h":[4,4],"8iZ1Qk38z15xMW5ATSPbb42pC7FJdFj8NtbG7uosNdXF":[64,50],"8igp2RrQ1F4drmXGpV8qNyJL25Aom31jAGJ55avPZLc7":[56,44],"8itTkbGjHRAx3cum5TD7bXaubmEFGxmKxqe6STrVqLdy":[36,0],"8jYnpEZcE9SUYPuaUXA4TMBWn57G1pPecRmT1fLssHqs":[40,0],"8m6aqa6BJj8d7aYA5s7R6mXKUM9wkZABAMYeU2V5u64b":[36,23],"8m9es585UmF8fksY5G8KDuBQZP8sywpGokqpiav6gWSM":[36,16],"8mCsw1jgZ3xKtSs842mCNxVUevrMBdA4oa5Jx5xCQSaG":[64,51],"8nVToBBSCKxiqowzvm7mKGwG8E7mzurHuRcDbf6hGwFw":[12,4],"8noQwzDhpb67yzfREDKvymKWtSdPZtbfjm3pxPYA4bag":[28,12],"8nvD1CUE48WdcmRdvbyWcM5LdJKRTNP3tXT6Qp2CSND5":[60,54],"8oCnS3KEtZGmquSW4khMCuAA8hqewT8wPPE3cxhDR1d9":[44,37],"8oRw7qpj6XgLGXYCDuNoTMCqoJnDd6A8LTpNyqApSfkA":[56,49],"8og65ngX9WuGbkzb5crHCgZdXKmC8AtFaVCPPSWTgxZJ":[8,8],"8omESudy1zEmWPdSc7RWed9jZ8EvbRWqN8A3YxWxgutv":[68,38],"8onJp7KyshoMcxVm5CemhPgGvA1hdSnHbcjLCvJidV8o":[40,27],"8p88nmvQ3uKnZtAi6poYpo28nqzzsRVXmsKEpvqCX9MG":[4,0],"8p89b2h5NHmiDs4tGnK2jGao3ZZWjhmrthyCxjacVFiZ":[56,46],"8pBBcPuSz14SSohf8BiHrBdAxgzrbA1jgtxTkwSjm28j":[68,58],"8pU43qVnsBZfVpFUnFvcNt7RL98mo79VJgGqB8nD4stG":[44,0],"8pWmLkuR3yio1Kcu1CqciTPmPMTiCf72h9n6Z1DmQNgk":[72,60],"8pXEg4dMZYwT2MhyaTgUWr1xrpEey1ArwrXcjXm5Z9wm":[36,0],"8pf1LTFXYNmvB1esMKvqLq92KZaAt2ETe3pGNxqy2pc4":[36,27],"8sJbSYEP7HtR1VGwobWNwrwFkjSMoPZU1hMkPzJoNApb":[52,31],"8sUeES7FdvfW26GirKMbU8SdMrAzQp2bPSFgeMbWMV2o":[52,31],"8t6UUXRkQTBpanRoMjxNxio1baXXkEdeLniCVJGMdzLJ":[48,37],"8tAfWTqBJiBaDfgE4cJCB9AwqcXcfoDeopm1X5QHeq7o":[52,47],"8tSzNoKE2tHYdTpCQB4apHaes2YWhCjbo7J5XCv1ULZ1":[12,4],"8tiCet13nwqVRtG1UbW5Lf4uuj33X16JnHPZssfvPXpE":[60,46],"8tk7QMWkXBbzw9AJJtLkrdf8ZnEQMiWmgXx2prk4DoQv":[52,43],"8uNgUAK2gn8Yc7eWmGFMFvwNfaKZqpj9fV8cCTkahZaX":[72,57],"8uQWXHAr8APT9fJ1bBh2XFXS4kNoWVfqZD9cjo6fUF6R":[36,28],"8uixkd8w27tuBfkRHNw6mqpmChe3QmVgNjpmNFVKKqZe":[68,55],"8uvRcrxAx5e6FRzLzobXupokSLwF31cPEkJV46LxyWuQ":[60,49],"8uxGfWm2g3sV57CbZSxz6GznKrDp4m7nZzThVz3VULmc":[44,35],"8v52QZ9KKj88NJJKMsh3t4kndqWPqkGAUb4NTz6XK2Ts":[44,30],"8whP7n58xbMDDLAAfo4rchFSS7Hu5jU4HLoo63onpzdV":[44,40],"8wzyvMnn6HSeC4EbPCV7XA6LeBddziWXjWKL52wVt7vd":[56,43],"8x3pt3B2RA7er5SCD2UZfhAusHX5UyX2hkHLLgtq5Nrw":[56,47],"8xBzUcv1AsTyjXGyWPZYLcvEe1S2gAcP4ijnaFZosRW7":[56,49],"8xsMN2rQHdmZJ4T829PAYGU6hdUivRU8c8X7jH8zNqmg":[80,71],"8yS3Zc45xptsaay9iaUSpfdb5gaKcQaKAShVvEUFKpeN":[40,29],"8ybtbfJ6rHeU49gtkQUBhAnaXBYGPdMk8dd4VCPmtbGz":[48,35],"8zH1mRkic3WDpUkSgtq1geCXXh4CLVfLrEi2TEqdTgFS":[76,53],"8zYLLHSU8URmRaAyWEY2H7uqUF63uezRCYpzFFkMG1AX":[52,32],"911tr1Hifn3z2opEsEEhxFQuJzp1YNM9QMkBQspJviWz":[52,34],"91K6thzfVGAQJZkdwEdMYDA7sWL3QJ2Bm3PRXHXkq44R":[56,47],"92ZDWNRurKikxrCQcfR9jMMYmqWksgTvSFFJ2Pa5FsMv":[68,50],"92h9nfYrehDW47mwHsjvGZQAsBikFrhhALp1XxzXo7rD":[48,42],"93E7eWXX8pVKLSrbBx13VpvDtvSU5PJs464uPoty9VeK":[52,49],"93g68j8QB4ZWAtEbvL6kfy1X6k2izXosDiuCfPPPYdjx":[36,31],"94HVaECNTwEQ8Q8w599c6BuydY5B32iG3Jem6EWXQvGH":[56,39],"94Pk8zSFvQTvrkwBkMEHzjufx53w3kX6MymDx2ayH45e":[92,78],"94VqMqQBj1WtJfyMRqsxFMWPPaAQQZ2CCUYR4UVyodbL":[44,43],"95VPY8GWPEquURxTXY49Ngv8rkb1LUaYEKyFudrWssUt":[64,53],"97vF6NK1NgmvMunNw9QL6ne9wxzUQ5RLAJqWxmDSkKuH":[36,31],"999vPueFgE7LEjk8awARTr1MVN5MMCAhaMph31EHPwfn":[8,4],"99NHmMDJeSo1AM8dg32nTokVRXByoJuA2gjDUDfiKHem":[32,32],"99PWsEpnfFaBMbW8epmC1pnRp1HrsFxASniofXNxDaQQ":[60,54],"99YEPZoX7Z69961fejw95XdJEAbHx5WVP5t8pmUFdvHC":[40,27],"99uwSf9zhnt8Co6Y1qB1y27dBVJkbWbMSUDU5Sq16XR7":[32,23],"9A7aYiEPK5ymzPjniabXofR7jyEyihGAANakUf5AALT7":[76,59],"9B3b4JvBXkRvy3XZ7xRaKVxy2aQFQtGUo5jfpVZYZcnS":[40,25],"9B4oF52Web2deG8jdbNbBuM8DjiFyLYM6CTid3dmQkQ6":[76,73],"9CCuWxTSZk4aGY8iWcEW89gTzCerQR3RCTBsMbNpbqfN":[4,0],"9CYnw2VNWfipQiDKEjgmZsh36xTmDBcSu93mCfSvMRpc":[8,0],"9CjCwpFfvex43ZrxC8iW26y34PsRbDsF3Y5fnf9iQTdR":[28,7],"9CnXcFUXEGcgHz2SHhy28ShuxYGcfYcRtoNSavUcqdUJ":[48,31],"9CpQtpHJ7UrsT6R27RECtE4dWWBAVnTcCTXj5HkbGJQC":[4,0],"9CqFnQed345m2MWXhLTaP7wzgs3uia3RCdipPeM7WyrJ":[48,1],"9DgTEERummZyV6MVSTmC8A9ZULgnN5Yh7VHjP2PADrws":[72,54],"9ExDakUNsM35KcAqwgmZVny83jqyv3SS55KwRCjt6oTB":[56,35],"9FHjVF9go3LTyZ1TiYjUTjEs9THPjULKzm7BMEB4cSud":[4,0],"9FNnRxn5uU6dVnViJeewy6FKu1AWdnknmLZC1pKqRuwy":[36,23],"9G8zRFACfB3gZAjWkgZb3CTr9KXhvEREHbaSm8Gm2mZy":[52,43],"9GERkwr654jBUn8cvDydFwnTZ6v4MZbyvp9ZKhRep3wU":[28,20],"9GMmVYJBw5Cj58P8QtXtesyQUtA9GyecPb6kCki7QSo5":[36,28],"9Gx6myZBqcVndLT5vf6pEawnqDmjJyB8SfanLTzhWjXU":[68,37],"9JvKbbmSH4T9MuHfpWmb5osoQ59dSnjXzWbS57N9r3bY":[60,44],"9KCFj7pL3hzyCzhgiy1Z9nMxT5mkNBgm2QjfbX4nXBPi":[68,52],"9KJyBBRfCt29mR21aP2NZHuyvZnf1VjSSB55WPExRgSJ":[56,52],"9LyKLKjujwPdaDWNYVuUa2eTFdyhXjp3RsfSRCvhWmxe":[36,23],"9MRUTN19MtA1matBH4ddgpS14mPAdeCoFnsLkaLxFeBQ":[44,27],"9MZY9cHJW9CbYEfVzTVmjsrnFonURh5o1rFHN56q33sn":[12,7],"9P5sFULhNktpQxEST2Wiw6zBH4aJrANCjui8k5FhwcjH":[60,47],"9Paysbs5evoh9BiWiS77NNutMCG9koUK2xyAsJm89Rfh":[76,58],"9PqR63RosK5siiSNvHtQMyEKr3CvJt1jh2qxoVmghhst":[56,41],"9Q8xe8KgzVf2tKwdXgjNaYdJwvChihmjhcHdae7c4jPb":[72,55],"9QUEoFpFLYnRPd3WwPwWkXYeLa24pDUSomoLsu5EGS92":[68,48],"9QVunAXvQbWb7Xo6ZfCWxnGwE4t1xh1dBfPc3qgRBSVV":[48,42],"9Qt4Ja8ArisFLsH2MaoPyYPqzDoiYNzGmYUThVFEKm88":[52,40],"9QxCLckBiJc783jnMvXZubK4wH86Eqqvashtrwvcsgkv":[96,68],"9RLnzRod7LWYb3nemb75vKhEBSsGqS1uHeuqh8Xuz9B2":[4,0],"9RUxQquaeSkuwb2qFqenPw63qXLypEwMwUVNaGHzDifF":[60,57],"9SHZFX3LEuL9dRpCgiETWfakZU1ZaXiw7aaeTzgkDzEJ":[32,26],"9SXpQRC2veMSkTRY1G2vLktNgc3Bbw4Nkg4xK1a1aVjH":[92,62],"9SjdDNazuohEjjoWnhQwyhzQvCnTkq6RQp3fxnezgiSb":[12,11],"9T6WSVQuo2r9b2sLQDEmyn4AaFV8XmJxASbqUiLNUMua":[40,29],"9TA34Aso9JfisCAsdqtpJ6cukxhDdqyE5xSYYvxpCM6H":[68,56],"9U4fqWRd3kcUHEX2jt1kFwF2dSXLnz9RA6B9W656Skbv":[28,15],"9US59KW8j31mxr1opP4fbg2j86b2p88DDKhcSeyDznnA":[56,29],"9UfKWtaruM2whJNqLLcrxKrSuS3VcVssdbTyNvfQCUpg":[72,66],"9V2bqR2Ts54hHnvxuwtG2yaMyTF7uscuoatWzaCUs1RG":[48,28],"9WXjR7Ea8hKt6Z84EGENQvGR3rFsovcxDYu61TJFcWJ":[60,34],"9X1qjnyb5CfMkGfEnuRZS3G58iyzbNZCp27RpiRVAiV7":[68,47],"9XsNWQxnJuFthhQ9peAhf91Am5agfy2y6SdMKvfudwzM":[4,4],"9YHpZqGdwED2uAxZbgixESvavajvuHyVZJKbVBevjitB":[60,54],"9YVpEeZf8uBoUtzCFC6SSFDDqPt16uKFubNhLvGxeUDy":[104,93],"9Yp7sEu3ecy31pKgQkCxrUWMsXiorGsCmxPG8FNwnFuN":[76,33],"9Zqcgqref1GnwwPNWcXaK88qib5hKqRMaoQ4257tvBpG":[48,35],"9aWVG4A2Kutu4tBmg9V1gaLMHSU44iuvLemPNxPVSzWk":[44,39],"9awzdQMQ1GrZJUUymUVm7SXZxfSCUDZMpWcHNGseHW1G":[64,51],"9cZua5prTSEfednQQc9RkEPpbKDCh1AwnTzv3hE1eq3i":[88,77],"9dCMmPpfNuWKyZ2D1iRstMCc1rhr2DbHVFrZ9wFncQjp":[64,55],"9dSTVY7hXEJsqExDcD8vYMAZpJ5mt9HBMPwRe94nBwny":[72,39],"9eeipv4uEyZjweLHQYGjzZqaTQradWjstQ1uW2SyuBPy":[56,50],"9g4NoCzB687Nsp3UhQEvt2Wx8Eov1hqZhnKjdxaY25fu":[56,43],"9gUMvQ8peCVhxU8ut4eyfzyTZZmvBUVDWw3s492yWNYC":[32,27],"9gmnbM2GUVXiTfCg1Pj3ZTJdqyKdS81kjBWwwnZbS4MR":[60,50],"9h2WhxhGjad6vaVc2fGztQViJ3LhYFh2MRvhLE3FgAX":[36,27],"9hedZ9TnXRLHipwYnuD8DdyvAwE7sPs9qdqNwjWvV3YD":[56,46],"9iEPYLQRdJ4FsuXm3JHagDPhYdHBh8o6muP1EM6ddB6C":[44,38],"9jAhC6dhjVqVA184dVczcBAar2GtXT7D7LwtXxLji3Re":[92,77],"9jpddNRkSJTpD5GJFXocmLsP8JUasJzpwgKrHrLtA8a3":[36,32],"9kAi6TF78NfW2gNr6n82dET61fnGU7YyYjhMRRcdEQcR":[8,0],"9kKpZomqGpNYRPa3A9o7s2SKZVeHKFCWGt3GdXxbbymR":[76,66],"9kRMWjDLCSpaWeGntjUp3mNxYApPf5cYJhvfUgJUN1iR":[52,36],"9kkpTAQfndU5SW5iVbG4j1qngoUh59Jwqndd3XpkBzzm":[48,39],"9mN1765LwF5A9iPevcJci5imXHe7kXWqQg1U2xtXP6xc":[76,68],"9me8oFZvWuc9cjBuXiW8YDGGZgkitk9GTKYTNCHaBQaF":[40,32],"9miqenD7FrGa3a4NNP6ygmYbpxtcAmW3AukuTUbAgG59":[80,36],"9mn4o462w5HzyxfnZGo7M84xsqRXL4EfJ4Ggot1Bs3Sd":[76,64],"9nwweyAkLSXutqy13N54c4DMyvgehnkoa72EiwtnBqzB":[32,31],"9oG814Uhivn77HToA3V4M755B6Sthx6aXf6jDG7Bwjh6":[48,36],"9oKMyQpMvPEHawMT1m3ryUZe624onYKhkXZ6S7aKax3Q":[88,84],"9oKrJ9iiEnCC7bewcRFbcdo4LKL2PhUEqcu8gH2eDbVM":[40,34],"9oWDUVn41kNZuVCQBr563sgbLXGvZULKuMr74w7NSkz3":[12,0],"9pHNBdibr5ukpX28foKK3UfCMeaB5GyAuGcHyJ5DmUAJ":[48,38],"9pZZWsvdWsYiWSrt13MrxCuSigDcKfBzmc58HBfoZuwn":[36,26],"9qpfxCUAPyaYPchHgRGXmNDDhPiNmJR39Lfx4A49Uh1P":[76,37],"9qrjiQG33wuqBGd9eWBevemxuw7FkY5osCxwYQt6SmhU":[40,25],"9rGfXDukY86MrUcxZNGq3nTrUaQiE917DMQ2EFW1cbDL":[48,30],"9sQUU9LhZBdYUFd7aG1NiG69sEadf8pXVhutXWL1whgM":[8,4],"9spfoHrvaHHg3VQajESkEboVGDQvk7ycBgsEPGac1RDP":[32,18],"9sttpBHogmgtBFoLZWjFsB2RZp9u6izrTQdUhBn8FHix":[24,24],"9tbzUabDi5D62Kkpd6oQs9r28Ts7TFJHLvx3pFJshZRA":[48,29],"9u3hzeHS8gtxzSEFbto5aQ2nFNuiFLtYj8SAPATiJGwQ":[44,32],"9ud25poQH48x42JefCbjH5po5Uza45MXorEDwdbyd91g":[56,54],"9v7E6oEm1V86hjTubtBon7cRYPvQriWZKHZEX6j92Po4":[64,47],"9wbAKVn7brvRaWuqeWcyBKdce7DUd9FTvjrf99xq63B6":[44,42],"9wgqMFEtHspw67xoNVnmzj8SLSisLeeBoiLEjGqqujb2":[60,50],"9xe4rcxYUe6iADdnvLkWn8K26bvyWgfrp9HYbtwR2sPs":[8,0],"9zkU8suQBdhZVax2DSGNAnyEhEzfEELvA25CJhy5uwnW":[68,60],"A1JevizjSWZwtFe2F9HujwgNZ5AbUoXLApBQcPNLGVEn":[40,28],"A1ieLrRfZyrRQ64RoGVyVQ6zqRhnQKQutm6kRGRPg6ma":[60,43],"A1voPbfnmCq8UBNQTBKnZ3Xbhs2x4cS2Gx2b2wJtqCh1":[56,44],"A1z9q4Vg8fzo5jLhrBDNqz6yE4FTz3ASSWPfcvKjTbp1":[64,46],"A2tBBzjR1zXQE9NXDzTiF4EchmLYdeMdHFKBhCRi3Ki8":[40,30],"A3QTLPL7u3tYyornA5dNkvuFLcypt6TBoB3SmsZFMcDU":[40,35],"A3WaMe47ySMTgS36KyuEWvbBX4SGU2gR5k3pFeYdUMJe":[64,56],"A3iynpHkra3TAVYmHDe3unD8343XJu7ogZn8Mhw6rEcv":[4,0],"A3zoxWHVyqHui8y3Z4rKyqWJTyr78tusgAEpAtr4ZEfg":[40,21],"A42hCCRfgx8ajv9fznoh2vQ67MVHBiYYHUUJPEPzsQaX":[40,31],"A4Bz67GutEFuHpoLLqfvqjU4PgwKkff4uNjEXXUomm6z":[56,34],"A4Kg15NX9i72WeQiH3Gp4u6QceScodz7CrkVdD2xhtws":[60,50],"A4Kh17xLBZmd3Kazu9aa2iEVa2hiMy4VTB7NfMAtmeac":[12,12],"A4xoiWbs1GmkV4p4PXkBZWM5UDfJqXx8z2sDmHP8FmG3":[48,36],"A5TXyVrR7WwfNf2RjoN1W4Dw5CuuMDiLV9e77pWhmwAP":[36,35],"A5hMwgm8QfooAuCMw9Rw2S9vXbBwCknFMhhUwKKHvYeJ":[68,63],"A6RXanjfgm9ivaGUFvjDHeSAe6BXYgJsX58UpiNF7TXe":[52,47],"A6hE8814DXNHWieGcKei3P1FhKL2rwD5YGfR89bKnihK":[36,28],"A7zCq95mtG2enn2zWxNyVDvhU2EsH8T9oWHs6jV3rtCH":[48,37],"A8Lv2ZPKKSBFiAiepFsmCBvWEBSVGzuKxSLVt9z62Bqt":[52,7],"A91g1Y8xXFEvCGg9afjTn222JDuY7iSVmSg4fdbQEB21":[76,68],"A92dCya9ivsYmzeGHzg4chen4or5WfCRcwq3btTV78iQ":[44,28],"A9CwddX4BA8AgPCmcHKAEZU4JDFRzruMFytr9oo5ZzPv":[56,45],"A9XUvhm5yKVs9Z3tYdyiAYRx9mNr2rqnv2VkY8D1N4uZ":[52,23],"A9qeyUzZoNXJQPe3fd3QgDujekiLg9Fd4VLX9UsSzAn6":[48,45],"AAmV6JwejQnHGJdUeke3hiRXch977a1PzTzFacBWximi":[12,1],"ABUhDLm3Y8HyLsmua9Xj9on87RyiEsw5j5eVVZQVw1hT":[64,42],"ACPgwKgncgFAm8goFj4dJ5e5mcH3tRy646f7zYPaWEzc":[44,35],"ACYDnrdasgqavXoMR2pDxeZxTBxG9RS4evfhy9G5PsCe":[84,67],"ACv5dTk7THbmUpHYGhgPzMhWr7oqHSkuPJpa5RfvmG5H":[76,58],"ADVqUcnmGF2Jkm3rVkhDbkNxiPUSSTvZC5GdSza81xSt":[4,4],"ADiT4zpCRryJ6NGtvErT5dtuFzTxwYRv24fj4b4LDQDr":[68,60],"ADriSmPTSeyKwNCo3geTcAY31G94mHmCfRfrJMe3DmbV":[44,36],"AEPNDgaApdcfZEZpww458Az9i2NZrwxsVCdiUih4EaRh":[48,39],"AEWoxb4i4qGP57iJqNyubSA4frWN51oJ7pQ6634skR45":[4,4],"AF3h2gdkGYndVj8W9qQN8jA45kQ5RB2WmoAQN2iBk37c":[28,24],"AFVkDuKCb9V4TGYNYK7H9PT3sj7Ny3DCRumQby5UHBCs":[68,56],"AFkpm8QAMCLbNebefoZsMerbWNAKCkXLFrxeCj2DiRAn":[44,38],"AGCsyz64NLvoDAG7Mi7k3WFbkMjRDCv158Q99WGGvKNM":[36,24],"AJEi7F8fQWAP1xarPkFc6XJWTi7qvPRcZ4JhLSG3CZo9":[72,58],"AJLWyfaLmio6Gm8GA76mdfUtuLT4tmyZsLhbpXsndNab":[28,27],"AJeuXG12CxqTQTnxseejE6PcMapybJejMYHpFc6SErMW":[76,65],"AJyQWpskWfNYu5wdZF8zqmNHgHpi4n6nAEdMNi1SqhYo":[60,45],"AK5hfHFusiS2y5cjqZkiyUyAvH5qfidQgrmCccENnet5":[56,43],"AK7ZZx2sdo39coZN5FsPdae2xNGqVKHX2TWJixmY4ecX":[36,24],"AKdhJ2gVzrMky12QF2j5F79K1F21znpGdFCQdY8M7mi6":[52,39],"AKm41uMcqEUerYPuW5jq1hoW8ZrrvwPWYHJDU2QmrNmp":[52,36],"AKqB1VaJhf2Jsod2ciEjzdTzCUfgh1kUUUaC1sQ7iMnG":[40,29],"ALPXVb1A7C8EkR7NuKy16pXcBRasdeRNmRPnWGQHpe7j":[44,39],"ALzqkbSgVaQz9nn5xh1BtEsey57otKRyGmaSLhwphYSn":[48,39],"AM6BNu2WZibZhYYHNo9ZWxmEAB7PhjNQBGKAhhN2VrFt":[36,26],"AMHwC159us6bok6awd7jdvjFVMY5ewk24JVpZNwK7Dno":[60,55],"AP2ZiF3mdoDsVd9AdzJWRUb5UHHoQjnLLPXn8rrkERGc":[60,51],"APjVTcfzJSzYEkBddGFN1mtFWb8jDzogpgUz99tQW3Ei":[64,48],"AQXUYRhH2meRQtdNiU7bpGSqSPJEbGKC2LaRx9QodhR7":[56,35],"AR5Lgk9sgoz69qGBeeTiMyyxZdhvCi2qkD3XUzre1Uvh":[60,46],"AS3rwVs9WR8HTzN7GA4aLBs3JjWjt1yKHfSzmwoqp2bY":[52,40],"ASDE3uRDLHUQ3rkfeGiThetAyU1bY9UssFv9PeAZDtVi":[8,8],"AT2N17bBBtTAu6ombzhiLNLc8JinjMXmGMzFbxt6AvwC":[72,53],"ATra44iNoKxAxj8zpWfE8oALhdyTZY2AA919CyoQ9bJW":[68,57],"ATtMQ3Aphrc9X2SeTJD9JWbcDJfn5aJs4NHWSknxfNb7":[88,65],"AUA4sQRvzwWiL2DTEk6aCaAqZpKJ2j7RGMsJthP5aa6y":[56,46],"AUcXRTirGAjzepCXnyL4UyuBBxkYMsUThvcei5M7x3fp":[4,4],"AV3nnAu6xyF7GcYRnaaPkWG1EBVtQwGUfFyd73BTrDxK":[72,45],"AVAMqDmPX4qjDZYc71Hdh2ZtjhGrGsT1yv86hAFVNt9u":[72,51],"AW3MLxDTfigYizfigb211N2BZKa2epePVJZ39ChxpEEx":[56,37],"AWTWtog1GtBcjUGuGVY4zpp9xrRm3aepsRi7P8EufjzJ":[52,34],"AXJXq3q6JvQb192nBU2ZQYJBznrc5ucdvhpQRoAfCCkw":[72,55],"AXfAhKbu2urtzCMLgPm4vDWwjK1hEs5ZaqMCbfEEyn2P":[48,33],"AXgJTkDM9AWW62Th6XUD3L1Emdxot2PLRffFSrsSamat":[60,53],"AYEHTBfsPvdGxkCnrMHEu1nTziUJB8Qnhjktph5aQvrw":[40,38],"AYxCoguM1XJXcd5e1bYVQB3Tdtu3vnT5iubjxgwvzNK6":[56,35],"AZBWhbBeVwpAJNFRK6fK72Lap4Yw4vhz6LKpYEQzQCrE":[40,34],"AZFRS7MsjYLd1oYAqeZjxyxXyG53zNYLQtKRk1qVKgG4":[44,31],"AZFkNiUSszcpsTSAmCWFTcLPe7iQf6sGp4ceV72JiCdt":[48,20],"AZY3mmLS39SKps93TrsZoC9nT1nUrpYLUmQWzbtgyF4t":[52,34],"AbF88hkkpZ28VaT3vYn4xu5CeNC8G6Dq9cc8ciRR4fY5":[60,58],"AbnagVJhwwM4wDuZbvoxWeofdpSWoDMhcmZCdCrxtCkN":[56,40],"AcjhWohnu7vYMdu4Yha63XZupqMKVVnrWmt1F57ScXhG":[48,35],"AdVKEVMZSd6VZ53PYbw3PSaj4XzDjsNoEg9LwDnyWRE8":[64,55],"AdrwFufQPWrBRAWPG2ferUA7Bi1EY7SeT62DRkaGmt3J":[36,26],"Aex3fnsTWQF8xf5rEf6bZawaEKsqED79qx7zJKzAR4qb":[44,38],"Af3BY8yRnmGLSX3XZsWoCS4UrCravpZuVY1UofrE28sS":[80,12],"AgFQkQe2Em2GUkDD85qPmHrvybnaXKMa7anSNdCunnM4":[44,39],"AgcvBSS97jBoKY2x1LXrqScziFx1jpCzdE2UpgSiVeQr":[68,57],"AhT4yWiSg7nnEWQokWoGDz9QPURwa9sEHrPkidC2PK26":[12,12],"AiPN5MwTHxRjG4eTQ1nrmxERRj4oXJURHPiTcNpVYcmk":[44,39],"AiWqv1dqsbvkUMec7G4DmM88ka7SaoqDPkn5U2iuvqnn":[76,60],"AicKhNhJmkdqafRDjKLPgVqLzXLzJ8pS6aVrYrRkq1iq":[52,39],"AiiS7TxGeSQcB3MgBzhfzdAXL88feL8ibxNz6t4QXBRr":[32,30],"AjTrfjYY2SiTC5pLJXwNXpcP8q549YQ9VrPAxzjqjUaW":[32,27],"AjpiWM8sXfKy9bH7Ww2qf8stQBY3Hk9AHixaGMYVN7eY":[40,28],"Ak4BNgorzDrbQSUTuxc42hb2rkZt7SY533b1HrA5U3Zq":[32,30],"AkVMbJq8pqKEe87uFaxjmt35tX2cNhUJTJwv13iioHu7":[144,97],"AkkJv1meyo2Ax2XTXEXWpvHTh4F8a68Lja5dx3TaX47K":[64,0],"Akqc2WGCzgLNEvzgTfxrUoZTau1rBgPrs6XDaitHoyR6":[52,44],"AmUY64jsSNtnhz6cXNRD19jmEBYq18B5naJhoKSU41oG":[64,60],"AmZ1nodB6yK5BVbbzVcp7AveuegouTbkfnUvMjeXbDBR":[40,26],"AmqVUpveo37Bt2VAgaQjTj9hw6APN9z7DMHK7X8CgjmL":[60,43],"AmtSHsetLtT12qsPtmTbKURWLSM5A5kipPwe8rLoeQUR":[12,4],"AnCW47d9J5V8W9pGNhagcHiccH7mLYcZR6vXbzTbPn9Q":[64,55],"AnbcuyWTbuDwHK2URBB28kHvhns9S3euEmKtY5gQ8J22":[64,49],"AoKD8QZ5WVvY82UgydBCSDZ5itUmyWeZHmu51KddCvFT":[32,17],"Aom2EwxRjtcCZBDwqvaZiEZDPgmw4AsPQGVrsLa2srCg":[64,54],"Ap7K9JT4WA59s2cukCpaZKZryVUwiFJ2g6793ZgeJqDb":[40,39],"Aq8yWGbM9uA25KDKsU9KPwoPEuquP5vTqrYomh8VK9XL":[76,61],"Aqh2c1x2AA59pek7pz8PymDXzq62qmNiQ5GXhpWq3rNr":[48,42],"AqwRcpXAYMPGSJuZNVKjuduPTGgaesRQ9XJWG4cbzgT3":[24,19],"ArVspBqfajnC23uoQmzgex2ge6NKRzFC6BVDJ9DY5qJX":[44,30],"AsVqSKFD1akXQwL53qiS2JxiQM8xabP1acrWx4SGycoP":[48,30],"AsboE8YchTBGaWdpP6nSe39g39VTpdHKefNWrbPdyLQF":[48,29],"AsdGYpcPVh2BScJYMaNuGaHJWBhJjEn2CPFvj1CpF7Pd":[36,35],"AtUD8QwodrRPdHXhEH8ZUkXkSZwe6eVaQbsfJYCarcLs":[36,28],"AuX8i2wxd4qQpiLcie3eGyWPbVeh28dx2yasJiTbJPNC":[44,0],"AudhVa1DskDfMAFTuF6EC9vDhA4QnSCe9JtJeEg81KXH":[56,50],"AussSM225GLYAGE6wDBoaSnAsSu9tpHqa1c9FG3PQVtL":[44,24],"AvKQ7X3BL4FoBr44VKNtbMtCMfHuwPb5ZUS7PLfYgVbm":[32,31],"AvPbRdiNN5nMJtPgRxihyw1nsL3prEgVdigGRYiUoEGi":[24,19],"AwMRTUyLVeee4eR7ZdPaj4s7FHgTvYSNmD6CacnTAbiB":[36,31],"Ay9TTZQtpbTohcrb3eq6doZAeeLHXmKgMiJeiUumC586":[40,21],"AyXMWbdxpvDoeJmueCBA3B4w9VURpiQu6pbjrwM2z3kR":[12,4],"Ayer8NhVD5xUyWkfKK2bMi9wmhX91QUoJ9kjy3xh9aPy":[52,37],"AyvS9yc8cuHM43EekkAd3kx25iGZvq9axPhHPvzre2Ym":[60,43],"B1Yd287CKFxZnZXMvNjbM9V61kjW1agupihzR2xoMWBt":[4,0],"B2UcYy4WiS1fSYKbMPeAKZoCEzgfQKKt5QBAA8NXLvpZ":[96,60],"B3ZyMGnSX5GjhKtopCtkn2jmmy9g5j3KdwcvXWR4dALk":[44,37],"B3fuLaQ9orHBEkeGL95m2oKcZZQgwm2uRaxVcaAJpcqm":[56,43],"B53tbis1864ZZqg9BBdXFAkDponmVvKHwHySAY2tC8gy":[60,56],"B6ZramCQhQWcq4Vxo3H5y1MYvo37wRZiAwKk88qaYNiF":[68,66],"B6eeWqfF19AGj2HtEk6jzSPEvpnMZjTvbyh3d7HzRBeH":[48,40],"B75YnUyemn7ixtnUtq4cDUVKrFwQmn8J2Er85ypcEJ1c":[44,27],"B7QNbMjAsaZDvNLVaBXo6Z4Wg7tKcESqPY9tQrSefvBy":[60,54],"B8T2dbvYaM4gJfpLLbuBScuWKFhnG6KspRK72v5D5uoK":[52,40],"B8UmpsNTU2RZ49BjTwLv9e2AnzfR3Gz7vBvNBWRWfWnE":[56,30],"B8wQDRb5JLuXjJhAtmY1MAQtLjWQySberTN7wLUHmP2B":[40,31],"B9gJJ4vMLJvnb5geZjU9PqhkyHX4jESMYajfcALQgRry":[48,42],"BBvV71ncxGrMDjmKTkcvcjipzu3bv6ExyVPPuRxAPtrC":[60,49],"BDKAGYg5SLDNRz85TZj4DeDekKVLxMm5kRNJf55WgJcp":[68,40],"BE9viFvKLzRqoc7eKyhTZ57WdTWUys5KekNXq54MYybW":[76,0],"BEFD2nMciBXpk43V8LPQ5D8NAedUzTswMD7JrXjQpQBP":[64,50],"BEq8K2LHtQdNGGURvpUxbbcutFHWQ1YATAvujMz5ju51":[44,34],"BErLhip6XE7Z3XT1p7EACrkwCXPcXKa6pCJo9eFPYh2U":[44,36],"BF1f26A9FdL6uWSajjTfstnLdCpynXGVrAEzqUyKXJKd":[68,62],"BFPdda76UtZM2grLc3NZwxUpP9YAhBmCb2dT9455ei5w":[56,43],"BFZpksditzAhQbBrLXcbBEPNLz8ifVBG9RZfvbTN77sF":[40,30],"BH6QqMa6nqyWP2iTTxqAGawP3E3n9vG8tb3TAPKAs5N6":[56,47],"BHKvJsTQHub7Y8KYuiSfxNN3ztbNJ8LALaRgAHH8Qd5p":[100,71],"BHPjYib5bUmwXAXa1T1UT79eVNmC6QgKkmbBxipKJxkK":[88,49],"BHR2K2tpc1fowNyUf4PfAumc2tfaT2SpvQVqmmpuN6tF":[36,34],"BJBxMx5dEb38YHLooaya4Q3w3s42a8HgHLocDLtcUxhL":[68,43],"BJDPVva3kGqpwRnsPFNw7pJgwdaVwqLTFLgbeuBWiiMa":[28,15],"BJMh3mPmJqwzvXVHgCqGpgJ8o6hAGJThvW2BdcwRbs1g":[64,52],"BJhh7JzBaSagZidw4Fmko6SVqkmMgazfskP3qjeciVFL":[56,50],"BKAgnBWgAMtC5NaP5uS4Vq3WZme6dAbUmEAJmyronddd":[68,67],"BKQq7feS56yp1PvAcBQjb1zV2XYtASm8EGTLy7eGq3bN":[72,57],"BLWxzv9mGX3DLam8z59A55qDpF9KyMEt8krFf88Sacm8":[72,60],"BLZtwHMTMgnZJdhJQxQaksgJgteXFFDBrA13ywWagji4":[64,42],"BLfpk2WoF8RnerCqjxHe7qFJjpEE427riMkf2fS6vUB6":[48,31],"BMeb32DosmmuepncJGK3yrGvpd53FYgsYgZZjz81mM6v":[64,55],"BNa6rYkNdwWpXSzPTL8yf8tiCZ4DqzT5orRh2j7GuWB9":[8,4],"BP7tiUE6JHuR3LRVy6bb79fLL9wfu92rkM6xy64rHu5q":[36,30],"BQ7mx4ScgjetA378LnL1Nm3xiM9bbLuEsX7UKxPseRCU":[52,42],"BQmxWxDnbLEQ3Pr9upNnaeiMV88K77JiXVUNoSHtYjPB":[36,27],"BRBQTuXWHFDvVTEm9XK34MXgp8QpHXxfrq5SLGMNpCnj":[48,41],"BSF2yD9mqzaixDaLEraF1en82EWaXx7wbaCqSuKppqG5":[80,66],"BSRAkuJ2ubtPESMHAyMrhSroNxTSJQTSV7YLdPw3FNau":[40,35],"BSzyD7UGHVbgUsq2yMKyXKysvc94Njudc273pcboJM9W":[72,59],"BT5bANJXEmnacdBHiVWCMGWckJEUBU7VRiVMiLmA65JN":[68,64],"BTFGAGpCMsqi9XbsU4CmCP9pbizVBTzaJxeBVmZAKg8Y":[48,29],"BTPUdVrsgfKFPyBmy5ozHvzMk1QCK9vii79wgxtGhjgn":[68,50],"BTVmKrqHyQU5vSqExPNYeozop6JPMCgCCZkVfCMKEa6w":[48,38],"BU8np8WoqP7cipMKWdLFoAe61yomgpvAEjFpFjoxfJhZ":[52,48],"BUn4mki9CktX2a2SMuxwei7n8ttvUy7uYJadbGDheAZH":[48,33],"BVEX3B7fRUbadEcigqknwc3cM2CUXpyz9vtTccBpwt7r":[52,46],"BVLVnUm2tkgX1sK2f5oVtTMR1opGra72s4qBD9LjMd9Q":[72,62],"BVSbFiadwxwi7RajrGgx4KAuwY7f3s6sB5AaL9pGZ4Hh":[48,40],"BWiftESMUsve87rkjU7HsaA7fkiJRAbv3xZLQrmKZtnz":[28,22],"BXWqrL3ZuDU3fJ2qbBGmmMagLBLgWv65YbXxqEat3Ud6":[44,32],"BYRUnuotppozq3YZp5EcztSSiUbYYg9hLasvSXv3FBir":[48,35],"BYn99WhTKSAuZgsUN1vmnPevVhn5i63cHMcd4gLjCxL4":[48,35],"BZCeZyvroBrSq2SbwjQKU41kcxkutUti1A7rzZqCfaHK":[68,47],"BZNtLRLmFaXi2jjcNGcMkUH35UJVEsv7GnJ4352x2WmH":[56,41],"BZehsFmSj4iUhwxHNqJVzduQQiaLqhzfE5mNT92UZf55":[44,21],"BbAyKMQ2vFmamjNbDREATh4SqStium9q5ZXZ2PMJA4Ha":[48,35],"BbYNpVvXLyDohtsNx52JB6yUsuTtrdC8a3PRALXTpD8t":[48,34],"Bbe9EKucmRtJr2J4dd5Eb5ybQmY7Fm7jYxKXxmmkLFsu":[44,31],"BdVXDJ15krsj829E7Gwou8McVYmVKTsPX3EcoBWPoHmM":[72,35],"Bdd4XhquueXBB7aZXVYUn1XBdJ18G7Wx3LUe6aKkmXEV":[56,39],"BebUNmLyM62d4BgE8N88YsJPygWrCWSNaCeq5s2U8uzC":[40,23],"BerdkMBXBVjUoKUkAuRn4DbadZpxFCB5mSBDadr8GErq":[48,37],"BfFiJgrPfecVSMTn1ma9UWMbqcFMftrxzgVp63TFWvV9":[48,42],"BfjmopwTknigm38Rj2synkw7mNTjgmm6hsCCb1hQetAK":[36,29],"Bg5j1Qfa2SEmio9U3UQKMtERNEjpqvushEmJMAi3SG4W":[56,47],"Bgp1oskwwj8mQe94U1Qn59BLEbqRDk1qRgD54bz2fTcr":[60,42],"Bh8jsbGHJ8WTxdSYqanh2Mv16pHjaCGfH2PKTnGKPMsh":[52,35],"Bha4mjHBAS1yFjvjPWWY7ht3jMneu4Lezq219pdU9dFu":[36,29],"BhbARoxdh2MT3vb4awXraZFPzSwBdmF9pGgURKNsjBqC":[60,51],"Bi5ys8HXXthqy7H5DMvdYfWHXLMWD3Aj3v7EcdXNtEYs":[60,52],"BkDggfkx9Lai8LvA6iu5ke6jTwCkQdwJwafbifTNpQPy":[64,46],"BkTtw74AC3rDKUbFboQaRVnhLEhsUrchotzSvuweaUCH":[52,28],"BmCFZq2tQ3zj3qY1pjK8iegUp2TAHj6cYPM2vJkSA84d":[64,42],"BmytjrTNhPLdJg1TUmtt6JtapA66vWgEm2CQFsG6Q7M5":[64,49],"BnBHhF7VA6vgyGK5goSyc8sdLgWR2DcaNNRr7JMuyGUa":[48,43],"BncRYtvkRJ2TaZ7ud6VgpWY17JcEd6Hk1GoFGf8mua7U":[68,64],"BnqoTc2VgseyBFGVuGbNZnDfrPUCyam2CrTtVEEG7CXp":[44,36],"Bo9T1z62GVKmnttMz4HxPPtRXs2BUkAd7T7yUsKyG4iA":[64,57],"BoZviedWdjsBkTFM7os4RuMHLJuS84qask3Bn7zVDWFK":[40,35],"BpXtvGhGvcfLEFZvTAY7nvHLyKY4YyEoL279KLfYyzAx":[32,24],"Bpq5BM15n4ps9zftpAiJARqVmAfUsmjSKfMQN7yEZARe":[32,24],"BpsPu22aaozdVrZu8kcqs7rrURKPH3fBAPnMF3oLG2Ea":[12,0],"Bq46YaSk3ydeR9y6yvj9887DtRG4iJGyfkZMt4PB6pHB":[72,53],"BqGCBrgYpLv62ebUp7DKfnjvSJ2qBc783kehzKJDERbv":[72,62],"BrgWTUdNh6J9YJyQqWViNSNkJ3wLw2KpVMa4qYmx9XpB":[52,46],"Brjpw4BpNaAw8jvaXmVEBpGYzKfAQi4HdtshD1a1KDLt":[52,32],"BrmBMYWThXPvWmRKt89FsZScadEGt7fy1FRV2QUArpwL":[56,43],"Bs7HtZ3zTNNKtXHXMgpzzkwXQd4ao3jtgbKwjmyEBzyW":[32,23],"BsGcvcCxwNGpcBzY36oCMno6mPVofJK1FuGpv1oLmFTe":[28,12],"Bszp6hDL19ymPZ8efp9venQYb4ae2rRmEtVp4aG6k8nx":[36,32],"BtHbnzt6SDwyKpzQFUS4ReNEGrL7YYwhuRoMdV1zM1gZ":[44,37],"BtNTPnJo2YWQhiaSQtNnCeTz7XDWUARRLSshe1smVmx5":[36,31],"BtfSHF7UneDPFXJdxsiheqEH2q6RX3ZShn7K79fzNr1B":[52,44],"BuFKun14y2uagDQXKou6x4ArWuG46US7X8TEEJVvSMTq":[32,27],"BuiUZY8d14fa5ZTueKiMXRjA8QtWAoP8XDoJXsixUob5":[40,39],"BunxTHgkSEyHHMikCe9ofDB5dsgcXKN6nqKC5WQsd1op":[56,50],"BuvYmqQuvqpNeBrPao4GKQyEVjV7sezT7M2hxb4rRLoX":[64,57],"BvH6ncyZow5NL95LR8MRExfNUib1xKQKSXGC65Ypae1T":[8,0],"BvJSLTtfFz2Qt1MLDJqfEtMw8uGtiB99nthg2aWGXBPP":[64,51],"BvVKJxCQNsFWpB5o1B6To4uZzqAUgNXLcfbS2inyL5XU":[44,39],"BwQPpmy9TF2KqcYmLNbFDGnDLBwdZhKeRnQDu5j9VnYq":[40,31],"BwxUhPBXUAEVEWDk2Co9xbTiiSFAgVieoRiPFNGbjBcY":[48,41],"BxFCg1xrZukbRHhthKXRY1gZVnHvu1xntdg1YmYnqeE7":[44,38],"BxLUkNxbwARzCsroVyFSniLBx9FirpQJviXUiB1ZpBXQ":[44,25],"By8BBocWV2yLsEMfkfHk7JkCmg3wjh8hPyKF4kd5nTrZ":[40,31],"BzoPEBru5un2Weex2nm7zkJQrjFrAUA7AudNBtEDe6hs":[8,4],"C1Ag2mUyZnLcd1o1a5xDWX8XZ8YG4nVXZpehcrxXa42x":[80,47],"C1QUyFjgVeG2mNBHErtmCLz8BUqS38saMUz8KA8r7921":[48,47],"C3XctAkQw2CaBuhV2k7Q2hEcU8ipsak9YsXz5qjAK98s":[8,4],"C3hD8Q7dLoodUm6E6LTWR3XqJgcRrvrVaMscwMBV6vaV":[4,4],"C4N1bMSzbfDwHGMitxyufNZPaAkNYx8vJxHRnHWrptAT":[64,55],"C4eTa4tqvvzpTsp9pa5NKAbeDXJs6sHWS5BfxGB44Xex":[40,32],"C4xEQLjRRKLRMPvZhAakT2D53jaikWsVqQ8dvsqbeZE6":[64,53],"C5EZSW7Mzt5bktvVViVqnY3H7Ujo2ReE4bJEtdqcnrQw":[76,58],"C5HvMeXdHGxi7nVTFPF6KcyK77RSWLLvEEB3ParXoK1F":[16,16],"C5L2t9gSXSSEcGXdbr2LBjb2p2R3HpbrByU9uMUsX1y7":[80,74],"C5VDTdJWA1ck6bPiX7d8CurGTfG45zpWdTU5G2y2deSG":[56,50],"C65X86Rw9BRNGK8TnzcZeozKgH5CwZFmT8aKPe9wyTy3":[48,39],"C66yBCeSs2vyHjuvgjQip8fKZxJtNacYmA1dxmwoWaWY":[36,30],"C7wxxkKVpvbEeHdpNrEa5xN5dAergj4h3jjcXunKJZo4":[40,28],"C8X3QTF52FhTXLg3yArhzgT9FBuEdxZN7aCUnPUprsxT":[40,31],"C8gk6Z2dm2aS8dNQEXDBbuSenoxd8HSWxDSMNM1u4Ebf":[44,26],"C8x8gRPxVQd1rk9VG7fm5MqtbPkTF7C9R7NUvb8HJ6xo":[36,34],"C8yQnRcUMCvP7ZpkbhPUx2TXspyojKNWttdVWj77kAWg":[52,42],"C9UahsjNtQao74K3zYdJdkGrhfcn7Rf1szmMtUP6fSRf":[12,0],"C9pe3PrMgXLaKUeEUjMFt7JFnnJ7ygFXCjzYFenmtqJu":[64,25],"CBnwmAk9KrFy4cvh7dCB8T9KbX1LG1v9ZxQ7nzFR5p8Z":[52,40],"CCM98AN1SENGvAF6mNvYsCvQ8SPbommtAdxSLdYwdt39":[60,53],"CEMuwgTq1TXoTvdFjuMYfTu8Rnvo8HVUbKGquAsiLCXs":[40,26],"CFtGf5wQ7jPgJVSk4GiVxvqVZXfkpxzdnkFJGduUKA88":[4,0],"CG7zvuaN3PTuQN9tFNoE5jERxtYwg8YVuKE5CMYD2jp1":[36,23],"CGYJpRhizVqqEyr3m7Ng9ghVpFRhBdD4sGjXSvvTFeze":[44,36],"CGgiEmA5whBdjKKyJVgFEBe2Z2qDVQQd2rMvAaUJP6Yt":[32,30],"CHFbPXHspWQfRGmxSPh9gkwWviUEdtko39MT6L2Ucy2Y":[76,57],"CHpj822jTX22VcSqzksxkJLB8kBf5gDMCqYbgXv36dvN":[16,10],"CK9yiW9cCVkJGs9qB2SZnXUJ9Q5btmrGWp6KuomttDXo":[20,8],"CKs5FjmJ8qx2o5gzCJukb9Q6Z4TEJ7ogJjuA1Fch4bwA":[64,29],"CLq7ip77pM7Mf69QVh1PorC9gdvzVZ2QXCi7B8oowU5m":[44,31],"CM1c6z3pRNgHFcfZG4z3wE31jaR8c4gCYQBJVEoCUyq8":[56,45],"CMFZtuwCGnXbnARnNx9JrrAXhGHjbGEis9ajFwYPGqCs":[40,0],"CNVw7suEhz3LJFDzN1sjin1MScbjRPWB9yZ3tNT5QrB6":[56,28],"CNdVHTP3ZBganD8qHMvznWeZTXh9KfMB2VHqdfdvYZBL":[24,23],"CNm4mEYYFUqGD1WtvdNZv5iZvVXNcG9rnqodf3w9xkNk":[52,47],"CPPVEbGFbX3XAThetvfveCE1vYLWUwwJGT7DxkPAWb8D":[28,27],"CPZDTXDMmE8pHwxhECEfhWZYNvDNwjyFx2GJ3pSbzBX3":[44,35],"CPi7yFjqm8MiLFJpdyfWowAgQex4DjJxxHcLa2rYZ2XZ":[56,51],"CQtuaJNBz6fh7Anvy4YCqkcadsEyk1y7q9oH6497WXS5":[64,40],"CRCQfYWcQNB3fT23T4wnUiLwneaZ7cySmp3jhPCmmA9Y":[44,42],"CS1Q8yNkw6a8SmY4nJ1jKrqhaDo18Wr4CnNbwsvKoswC":[44,35],"CSX1ZVkdseSUecuAtuTPyPPadWcUuTodE9BygHjMhKmU":[8,8],"CSXcNyNFNL7ymcMowKUByDi6DHCH9oVwdJgbgZGrvJcE":[40,35],"CSY6zb4eK9EVKvWkid5aB5K6uc5L5An7UeY6VwRGXbtC":[60,48],"CSj8o74gANwA4V43QsgpGHhoRT9KarDvQsecBSCkXeVf":[60,40],"CTAqarrLTZ58oAtRVo3jrQYqcuGke1XDkAMqAyK3Yfej":[40,28],"CTinHYqJWMLRstmkfH2mgPRdDnQu5JuFfQucZfMuiisK":[64,57],"CV3F19YAhoW7DpfHQ5W9t2Zomb9h21NRi8k6hCA36Sk6":[68,51],"CV4CmE9CWVp3sRbSVR1m1CUEqQ7pRdK1p3LEhirYF68L":[44,28],"CW6eHayhnTCN486PtdnoPGAxmUKGJ9MooqhpoWjQdhX7":[32,15],"CWL6skWfKLDd6SY7NnkjfMgNR1QxHhxCadyFNL1ssNaS":[64,56],"CWSr3rjCEQQvzNCQqWG4iy1XqCRhCPH3LUHKZpvnpzfo":[20,0],"CXCY8Pzu8rDGrWsp8yuA41z1tLv5ex8NJmRGbfu3gbAz":[68,50],"CXvTFeiDYsmHfUXyYgm5byEwM4T9T1Pcu5qRiTfKdsho":[36,24],"CZEj5m7Q92B2GTYp8wYhz4vXzYu9Vr33yu3y8bsXBey":[52,45],"CZWpCTN4rCWer8fm5ZqFdx82CDiCJjZLKZ5Ti2gdmchQ":[44,39],"CZY1ZJAUyD2ZfHE5ENChUmhqSVFwPnTm6Aq6N5tbBqaP":[32,32],"Ca8DQQagVHeUAhWPWxGCmaMuccr6aGsm9HxeedxUKBC7":[60,0],"CaCaErMi1TtrTLLv9jaM9VVG95Tva2keMe8BZSuji3D8":[60,59],"CaT9dSx37Quj1kcAXEVd6ncM6NLvYUSqhtgEnn1JtNKC":[44,31],"CahiUdyge1w7Z4GrsXNpVYamxj7pJoY7uSThHC1LCBPE":[40,28],"CbaT7xreGMFW7sQV9mtXRNsZwen7SCvK9BwLJjiJgZMf":[64,53],"CbhvvtosVdVwZ8GVrBqgYT3JrXLh8JRqgKpimhnZw31h":[72,59],"CeTirCDrYzkRMBhn4P8WYQen8Ka719PvPnAJqbDPm6Z7":[4,4],"CenFrJAktGiEtJEGZXEtyN96piKvy9EuAmD6AAugTTr7":[68,55],"Cg6ce8stTTtFP41k9FwYUWn2TEv9uq5Cb56Jc1p9XZnU":[8,0],"ChLMXZ4KXsMpa8W1VymMxim1vdPSK5a1jDwfMbm7cycT":[28,14],"ChpV3NPQ4nHdnynwvAUD2PK9X7UGVYRKvUm3wAR8T8BW":[40,34],"CiY5RjWPs1XyegKyBLcG7Ue7YMf98eiEnmvqnSuSKbob":[60,47],"CidJBbVgPwrVFK4Je2Wxz5zuiC5sxcETbRdhBwwxZVvA":[68,42],"CjQzg8D1Z2GNtRoZAFNc8kKyghGvr7NSNKW9hCXHrAMq":[4,4],"Ck3SxoXUShtXfLKfUUXAtCwrFVsEohESJfWGWuSgtTQU":[68,60],"CkEhoG5YKHt6szYHD84r4x9r7VHWN31aXaXYLbse5oYJ":[120,76],"Cmg9ZbuT5pR8o3CBLo4iwHCMxWzd21ZyNAoLH1sAwHxh":[60,52],"Cn5H2oxjXemT13eeFU45gobRYiJrjCrhGaqKTMd66SZM":[40,25],"Cnw2PuZHpJjpLd1ZxxPxetuLiXHniZjjYMGhQpJYqRBU":[72,63],"CoCKdrHVE1bjMZDwP8Z16vd7U3E5tGyt1hLw4tTsysmU":[28,24],"Coh6CJGtbgZDzESkahjCtN7CJsfRG8fVqHn5UxpzFniN":[20,16],"Cowx6w6oyFdnkhVUBsseo3RbGZGMLv13SH6J9Bo3J9VH":[12,12],"CpGXhkQci54bs1vRaJW4EsEbaH9viTZndMFHeoxcP9Sh":[48,40],"CqZVJ4n9EnLR16ed3NTf4UdmKy1gTUBdQUkz4HGiqiiP":[56,45],"CroXW5BHAubsisQRSZVcjdRE7vypucF3GNWMhesipsZu":[72,53],"CsaAGRau3ZvyMQvJ9CWSqbqeVv9zw2Am8FhnL9sr6jTk":[40,38],"CsiPaAm9Zr9EpGojSNjQ92h1kcKQvYubiKmpWb3C8B5w":[44,39],"CtgcuRPSMc1Y6BVyUfj29abjHTD936sMoKFET7sH1qtz":[52,29],"CthYXhfQPZDce1mZ7Jy3PA4WyF9frjqKpvxBgQ4XtPhQ":[48,48],"CuJecFtqRJmuyznjtUwN6cbPm6jiae2Ka92Q2wTsWHiM":[60,51],"CvbC65nNU7WjvgnTd9nPKyy7p6KE5Hd9YwQf41eMN2dH":[64,57],"CxPhLbubp2QN2WRatESeN8YH8r1PkNtSC8mf4x8jirqW":[84,56],"CxWS4R7rdKxdrFNhhWbPkcqbD3LzDXLr1LSvxpDrU6TV":[44,33],"CxkCs6KjydWBdMcpffxrYAiZtXpkjBW7V2Xy7JkgmxVL":[68,56],"D17iHRzwBk5NFzAEiUb5JqhaqDUM269utRqduzMxcTT7":[36,31],"D1SCYXUr2jHsFnZDEw67znD43kXkNU1JrYo6o49NhCnz":[48,39],"D23NCAVxinE53BTemguZCheAqCdMGfNTUzWdoWvq4Xj":[56,34],"D2CEZnBYQ9huwnhCYKr4W6QVdakwQRtWVrz7fZ7HxSWL":[56,39],"D2NjDkcv8Y1dWGdtWAKPT4em2D3sYzM8AzMTpCG1RVf7":[48,30],"D2ULkLgZk1d6RW3Wmd14vFNfkBgi6NMM8CDNsNuNXvfV":[64,47],"D38N1Rjq3Aw6QvhER2CeHCELkCsgcUkyPkP3FbRxp3F8":[40,30],"D3eokjT3EZnEsdqjgRdMv4QVQvjQD9u13ZyqoFX68eMS":[68,39],"D4pLf3e7kDGC4yc156Mb9A7JSAYnFH63jjsvAkfguqZB":[48,43],"D52Q6Ap8RVMw1EvJYTdEABP6M5SPg98aToMcqw7KVLD9":[204,145],"D6Q3V4ZXa8eXuy27WkvumLPN2HVrLhqA9tf1VrRVa1Xm":[48,34],"D6beCFAZeFtXoZKio6JZV1GUmJ99Nz4XhtxMePFvuJWN":[80,58],"D6pwr3woPNRpnDrtzBEvD789wtx1imWvvpD11T99C2eX":[64,53],"D6svmbCCUDFYmw8burYWAJwBq3e3Cdp9wiLdfNZ4SLus":[44,32],"D71JRzjPpHipt8NAWnWb3yZoXezbkGXqSf7TVCir6wvT":[84,76],"D8Bv4FnVhmr118E1HeWYaXNSurvziDYopuxzdernbQ9r":[48,38],"D8P3w7GQ4zTYbJfEGgfdQWQ1vrL6umGYAUrMz4hBJjrN":[32,27],"D8WfQnAbBmoeBj7MPL5qWErd4xonZqJuvbJS6KjkXWj6":[40,35],"D8e7uALahjTMSH1YGUw27reebWx6rE1sgsxwbbCQiW3d":[48,32],"D8goKEZAXaWCfVSFbGKZQtFn3B5XFdLLmJgUSgMjEJf7":[64,49],"D8wUJAcgFSdEqvFrBDFHzSZuf3SE4zM8p6F7HoVe3fRs":[20,19],"D9rCbP5rBrJztzv2EaACNt2LhXVLpPmsNgcWyB6LdfWW":[56,33],"DA7SNDUGAHwcXxHoUhbPqTv2p8GnncMpRYYoT6eJKmSR":[120,103],"DAzVJ9PCv1UMsHg4F9wSGiZ1RrSPBkVS92SK2BPEiL5J":[48,47],"DBbLyHobm2UNyZWQt2FXgZN3ji7hDEx9gJXQ39kHgwr4":[12,8],"DBc9sGfoUaaSqJs6PX1uJwygs2FLwS34YNEhuYf9JRZM":[28,26],"DBd1Se2ugdr6WChDSGzSzkyENnqq8ZWerfqGCfKYUred":[48,47],"DCEHziymZV86Pe64mBBNVR55XPHVMvE7zwQ8in3Uvdid":[44,34],"DD89H8QdPyWGtR5QnrfM734G4qrD775HFGMobyrkHjn9":[76,58],"DERZpWNzHzusMAK29jRQ573LK8bvmENSyeLii7ZJtCZB":[56,43],"DEV5ZXDo73Cok1MvGaGuJaACvra6GHin4wdng3HHEQPH":[76,55],"DEYZsYj5yU86GCgYn22a4w2LZX1jokqA8sA57ez5A41M":[48,46],"DEZAHY54DgLq9Md8CyxBgNCe5hxDQi7fJaSE8jymtazr":[12,4],"DEqpqWRASZVoDVMQc6NpNjJbiY5KxupgiUXWCsc4TUim":[64,54],"DF57amFm9aHKYL9pKLSWindiF8o2RRxtReLb6d8DQc38":[48,29],"DFpi7mgmChYV9whs4uEtioFG1R2WF4TpGd1zcXMjGwF3":[40,19],"DGQ6UDGprq5t6fhfxdykcAcqtx9P8MqkY57ef3sjpGPp":[52,35],"DGf8USMPMty56BWgwFSUz4orb9smQxsWxufKBXSoX97f":[4,4],"DHNSHtEZHwPkkfomi5oMmCQy52ACoDpD5Kc6oNGTJTth":[56,39],"DKNy6YAPt6zq5jVD5S8EFSXpQmqA4NjrQf8t5v3tHo7h":[44,28],"DKZeJyvARVGMouBRpYp31WUj3NQTuPw3rrqpcZSpNDoy":[60,36],"DKmRzfmz1JF1e1hdkjG6pMxnmFuyPcevcdXFG9dVnzBm":[56,44],"DKnZytVA5wKbNPYW1pvPpoE5YeSsxu12KJFa95gBAGm7":[4,0],"DKyon4vSD7mF6uqgEJujABpEdhRbyX9X9EzFjmEz4VBx":[64,44],"DLDshHnnGetLXCyk9o3RpKC4iRATqgw3UyftYF55ffuq":[72,52],"DLbekTqPpWxTsts4a7GqC37zDoCVJ8hDt4VNWfukPnRv":[36,36],"DNCnSkX9eHoTaW2ekwsTBeierxHJPoP7mGTYxKrEdNqg":[48,28],"DPtryjrrMUTTB1uaKkbqngnj2v1adJNuNW4eRScUPSP6":[32,19],"DQAi5Bdgz5aPhMMUMo3Nun8TBewGf4zJo2EqGt1jNNQ":[48,39],"DQSg1PLT3Px71U4LsfBNhg5yT9GgH8FnK7qQAq1aLmk8":[60,44],"DQTiiFVnwD7bSSgkMmwUsBsgnNBza8N6oEGeMC8YaieW":[48,47],"DQZnxmnLdWuGP7wxpw4krfeHzCiu6jNd6m32DSoQZc8Z":[28,21],"DRK6coUhAikJdsNhcKy5ZUGhXJwj6PDrsFXKCo1UJpuT":[48,41],"DRUpYPmGX83wz9d2bv68fN6ruepqqukxUcFQKVB2Kvbi":[36,31],"DRaSSQyKZAbdunz6ZkV7aznM6jBF4GXd7XxMwTmmgERB":[56,32],"DRfnZxdhCF4CLFNGMgvZAcxA17CPCCGarfPcEV1B6vEq":[72,58],"DTPPJWkD94MGE2oA3cyszmCAS3QvxNjXznnRGjpanbRS":[32,29],"DTfJRiAVyiULdo26CQ5Dpzmt8gRCG8MhK5QCRkpngEST":[60,52],"DV78gathrorcpWsWrUkWrWNowLXpizKsPBupStzeAJnL":[4,0],"DVbHAVYnxHTUeXwFvVMRWUPhkeozQTPUu5CtRgW5yD59":[48,34],"DVnKs7XAL9au7cWrTT335gZ3agJVwrqeSVnSWANo1SJG":[44,39],"DWic5rF2jQeAQr771cCgzyLbEMgXaq1AUyhnpmgDze1a":[36,31],"DX6Uuwc6jPTF1nqah3reNYpe28cjAzV5LViwK61bkgBf":[44,35],"DXCzguRGhTGvFm9hdVsDkFi4S3n6W2yrNeUjrFN8tkvL":[68,54],"DXGGtG5B8JuNFRA5RS8X18HbTRot4dqFiY6GVL5WwxxC":[68,37],"DZPn1XMuoBpNQzcXozfMMCUJ8YxgKyJP1Su5oKvvJk6h":[64,49],"DaB3ZwVtGLzSjazk5STQEu3MkJR2nkK3tDdCPAvx9QpM":[48,36],"DadnDZbFH5BHHRHD7TaobaSQ7QATXgvWegHUcZ7ZGzmW":[76,67],"DahDt2bS4EWgf546qQm8PLiRZuZPGeUKD42urQJCBYJS":[68,47],"DatMDyrEByFRiQeLn6gNznSqV3gHK1SN8FVUPW8CDD3J":[64,56],"Db5FG9D5Z1WWDSvQioKkymwRiTTGcTHbryBniRqYE65G":[64,43],"Db8J4gwpgKyQteGE1ZsGgd6kHEAAynqLNNYhwrXNr4Hx":[56,27],"DbzdjE8TFSN1Zb4g3N9NsgFrzJ68G5WKtgSxqVox7Nxr":[84,66],"DcgUD8iGqAbKNAa917SBNwmH5DJZiVC5cNu6BaRhKRk5":[64,48],"De3f8bq9dXXHj61anmFSGnJFCoCzcT1FysLoTDRACBXB":[28,27],"DeysXBwFZpyHzuxygdWo66nRMAohP8P6nGX2ECd914BJ":[56,33],"DfTeDaxk4RufkVbykedVnqa1r9S3z3oKFYL3FFmPdr1o":[60,42],"DfjBTVrgnveaCjUC799e1AxUVg85EtXoY1Zq6qWqLmdE":[44,37],"Dg5E8ktH4GWfKL1vuVTdqZJEkAEgtV8LqmSXyLJuZ3q1":[132,110],"DhMuXF3UqZvi3GhdrAMVyEQ7pW4prM8DkW54scYXo9Ke":[68,55],"DhukytoqRv1H9J7LZiv8FyqTYhBhqb53gMgJT3dtdyk7":[64,40],"DiAZodQBeCoP65gTBfPsghk8No5QH6E1PEwntuxyeQy9":[40,23],"DiM3w5M4ATTQQheYRrizFCSoKCKmefPGnah8cPsrYt17":[52,35],"DihbVNPXN8M3A9TEMBJ55XUX2Bo3w6w1BK8BRsR7A8o3":[40,26],"DjuMPGThkGdyk2vDvDDYjTFSyxzTumdapnDNbvVZbYQE":[40,36],"Dm8Kusyhxmz2NmwF8RivLKembinSL6h7tvh4vrMVNxoR":[44,22],"DoQrvRo2yQqM9BT7UxcjgcdBuXE6b8TNx42a14ucVjBB":[60,46],"DqbeCr74dFGDPvg8BpV6V2Jy2BCTZWmuvChkZVvPC3wP":[56,41],"Drf2oN83THfrUJHA9AGzJZaL9KMKggPoL9HJVttkSCgL":[60,54],"DsaF77cCADh79q7HPfz5TrWPfEmD5Gw1c15zSm4eaFyt":[48,34],"DsnqNtwKA817a2VQypWEzaRXY2soq5Jgc68MgFBMR35p":[52,37],"DtqjPaZAuxaNRsX1u9e5EHFF2JjeTFwF5SZ9YFJ6PyTj":[56,49],"Du5zdm6f23229xviUfyKH9gZmaPDUoxVpSY4TTmzQmj5":[44,31],"Dv9qRs8tB9oCGdA8FBtNtT7SLuttmVnDFT5xoaq9NFTC":[52,39],"Dx4bMuKpGaxAnd53QYDyKhD45PjuFLx16mrgoRK36STf":[28,23],"Dx4qoPTLSRdbd4h5cy2TD5rDyV1d7LxQYdSY1TLStjcp":[68,56],"Dx7UU4my4DNPwRkHnJN657sMCakPWYPcWwBGf7WUiyYE":[52,44],"DxQTxfgNhgzpRR8dZZku9XHPpWyy2Fb7oJSWoLTKwwKt":[68,36],"DyoVzxMFgZfch9L2jDYz3kpEctE7CVPVjvseCnGxqVqt":[52,19],"DzTb7aPtvxo5tbbKxEjiSuBe74tdgswdj3BY8LwstoBg":[64,62],"DzesSjb785mtCs2f2N5UAfnnF8obRbECAe4AQaRVmX4X":[64,56],"DzxNmWD99qvkPdDR94ojXhUrpzT8VdqB5ktYX7fZr4gc":[72,60],"E25BqJUzSjyzeZQR8LYUcUNNVrgRLHhXxhPtZGB7KCCp":[52,39],"E2cy4hqcUpdyMpx3TuHKpdW2cJZ3cTSthk4jfqJryt6B":[48,36],"E2rAcJ1tEzXK1cqDT6UhtyJmHzK3MvnZsLDYD7EkTGJN":[44,34],"E36wj8TugYUB94FtGrs9zPdb9SuT6jP4V2eumMXVy8tk":[60,46],"E43Lu6um98dGLscCuPUobgKC2oLANeByzqdab5KjxV1W":[52,44],"E4YYWrsKv9YkBjLRtVNYn792RavzkXL6NPJ5Z4sHXiG5":[60,42],"E5q213N5LpkpABYBWrHWaiPveFZGCRv49sq96wAZu9Tt":[4,3],"E6MuSSCF5aoBstVcZaD6sk7hkQrxvh7s5ttVt8NzAiNM":[40,16],"E8CDgCsCgLf3gBtQVxhKiL7m8DgUdrDAi4uHMjcPfZVW":[40,27],"E9WbzbArKL5bnW3NTYZPKuU2dxMiznaWAyc4jpNDfdv7":[52,42],"E9bcuniYQhMscfMjE8zaAXQ47TH56gsQoKuzvqXHxnAY":[76,59],"E9bwsVRDSEBnH2CbMRcdfiTjbkbUVcoYDrJMhHbESC1d":[40,23],"E9nLyV1V99DSxbAr9WcvXQCx4sJkJewAw7LF59yJK5io":[60,32],"E9qZxXtwWT5FuwsXHLjA4cjJyyeYb3ixHxBSrJJDzPwx":[56,46],"EBDnuJT5USg5HsQSZtWT1q8y5XjgW21b1ebYSahcX9V5":[24,15],"EBUmXF7Jv2kPk2ANMxu1ggRNtFfUkj6vpnuufqaSXHrs":[68,58],"EBxhSfAWW2Cfouvj1k242W6U8krZVAxJS47SG8UKb4ch":[56,45],"EC1TjttBQaKU1dXuMbv4ZMSFXuPDt7UCMvNXruxCXdA8":[68,49],"EC8cRNmwgbhbs5LtvufLkme1QqedbugySgkofYtPoDKd":[8,8],"ECYTLkfbyQV1piR86TJK38yvPKypUtubb9CoLkGzPFC5":[52,27],"ED4pSxzaemm1KZ2kgSnimKp3nw8GB4qpUSmFCxhqUKRw":[36,28],"EDpQ655JNe3hFVgMdCjz3JJhAVkZpwNguyYxUvQbvbc6":[4,3],"EEBa8frib8zBLxj61NEMAUoEyrHFgV9MUzneHVHFax42":[76,55],"EG8D25QxDJ6nbD3oBpu4tPDvihriy9mFiPq3CxCGFiPF":[56,36],"EGknxV4LZM4DNL1Y68iAPQEdLsMZbL82wQbDmsGw6w8":[44,30],"EJJtA4Uzis7a2eowSGCGgXypScp9aLzjJAiKrBzdT2zJ":[44,36],"EKBBsq7snZXyabwMa7jbyyRTMhaUQqtDHtpVgDnSg19c":[48,47],"EKH17WCLR6oCVbUJR7YcThnTNRNf1upYRvSnha4Xu7cN":[36,28],"EKxYfdfsxAR6BSpYKW3LD63UMn9ZWYUNRJ7GUonWSdrK":[28,20],"EL3RZmhvLAMMoDip59M3oKgqXXzHAPdZ88KQ2h82mCB8":[52,37],"ELJyNgQ1jng6XCTGTJoha7af8srgZdQKdBGrZ988G4Wi":[48,31],"ELWMKHPVZpFTwBSzVPF2q4nmvexLxWycjy8fuoC6egBE":[48,38],"EMeaA1d3kmoBNtZQNgqEZS9y44gMrA7iSuqS4nZ4qxpB":[8,4],"ENtpcfC4cBpkCW3kHza399ZoUcbxYGsYkxuKM1jwei4L":[68,54],"EPVNShGLknd3qtChqrgvDVmBZrpi2yktTk9YfcMbuFqQ":[60,30],"EPdTTdBRGeiXp9yzpPK1RUQeinoCfgq9qdicBJqZcJgg":[56,41],"EPpXthnkXneSTPPxJhuDwzau8aeYddx2yhzX35qVbpR1":[68,39],"EQEEdsGspyFKjFmAsv3gzYXbfqSCWuHrv3iXodb1LyXc":[68,58],"EQEJp5fi1ukX8JcAT4LrNLYqhxhpWW5bp4eGu99RQoZy":[32,28],"EQFsB8CDcLsCYeRJxhZ4fJWnXjCnbxrbhyqjyUJvkDcL":[64,54],"ESSnx4SECiLzhkk6jdQuWUmgxnfmpHTeQFdVTMwGYqE6":[60,51],"ETMbiU7hEt7jkoA8H6ACsfeR7LyGA773k9HA13yJUfex":[44,35],"EUjwiu81EZrQaLfCUthHQsdDPcuny1QfAWQy8Tag3Vay":[60,55],"EUoXy9YP2tAefgW5CHEvMGAu17McAvrXiQ5ucezjNYcd":[68,59],"EUwiTG1ii59qWfgsJwEMjqh34NmShMdP131BWWqJVPaG":[48,42],"EV5QcbuJdafM6tSdz841AUvbQXu4V97R1JPZs5Cq6hJk":[40,23],"EV7arpFrG8SFFAkfYRMEJtuqocMeXpNFD7zDt22qG23T":[76,62],"EVBcvaCjpB7jpT47oX3YGXmx88yZUXaKAhZyZfzjQyJQ":[72,48],"EVP5AqRX9ocjTLTNjzDqeuZEvT6EmmmR9BW4UEezCMRo":[24,19],"EVQxcfApDm4snuJU1XHLcDmwqiLsAwRQ6MatFKmSTWv8":[52,38],"EVd8FFVB54svYdZdG6hH4F4hTbqre5mpQ7XyF5rKUmes":[8,0],"EVtpGGP5SWd5h2WLtGYMGL4e3SeA5Myob5E9PHSRdT7m":[44,39],"EWg9NTC5s7Pa9FktUk6dX8xRYkvJ952peH1z1iznd4nV":[36,31],"EWhSgyD3VM1HuUisMz31xQPuyAXH5skcY5btY5ezu48J":[52,42],"EWi3X56yoWm2WWMA6doetd2S9bSqm2cc8JSBaWHGuJ6n":[64,51],"EXEFh9rPB1VN5NGDQsJiAgR5qaxUzHQDRJqAZdqEYGV6":[32,23],"EXT1KVqLrtmBhLvAtHyyQcjgYS3nuFRRxPsSZacUpXoS":[44,38],"EY1kDqUr6hfV3oevMtbHksUT6tk8f9YG3eSvWzE6DVMy":[48,27],"EYbvBPU9mSPTVJrZgioTt8PGPL9Bjv5342ENBMR5X8r8":[44,18],"EZMocMSpodWGgHthkEwXCqfLXcwi1vspNo57L6ZTCoEW":[48,34],"EaVhV1UzbiAh8BCdTiAbvoGWktfK7fdR4PEXkkN1qG2n":[68,55],"EbVCEn2DQjEspKn1kJYACrx5r5EjNcFKaSL5fxJ2Tois":[44,40],"Ec7d7N5r1xHSzq6nZg7o2nVVaymvU2APpagmBJ6LQ7Zf":[44,39],"Ed2cbqffj85JtCw9fKX6weRN3amAZ6upPMUJ49RZzYru":[88,52],"Edmn4FjZMGSmCjCE2FBLzHNjukEXbzEKiHptMfj87uU9":[44,38],"EeHdr93aBEXELpbDx9p8ScgzstUZuNZmFFHM7oPccXzS":[72,64],"EeebKTaPKaff2WmcaHyMH8mTVyAL8Gku6Z2owg7Aj29Y":[68,57],"Eem6rzePhp56kYBvWNgU89PNjrnWqJs22WcuiSjmkBc5":[52,24],"Ef5gVy3PFRJA7uQ4UkAD6AufNcZNtHN45k5N3L5mYatU":[52,46],"EgQ4ZdobHjVmgdzMvFhk6MrTWcyh6ULD6m2VFNqdCGx7":[48,47],"EhJnXqSV4wjCEA1bH8LeZQZmJnMQXJEMj6Qnya3u68gn":[32,22],"EhpB1NmmfzKXJb6p1w8UFsaftsjxkZXLsAGXh3j68kwt":[52,52],"EjcB41hrq5Ltr9Yvda3jQ8zGkkfFGKadkykTCQnPeCne":[44,36],"EkCbYFvhbY1QmuA4D5XTggghpPWsVWumHkBZYuLBqyMG":[40,36],"EkVaQMGB3cbyKdqBwBagGtURjtoXsP6pS7HGyizwhUs2":[36,30],"EmDwEWAfsjmLJ6kyMKQVJgcM9WCE9LdLxukw5aLG9KjP":[36,30],"EmKohoDf5ofnT8ivyF9RRtQB5YJJLKhHhvy9jRhyj8oF":[76,67],"EmVfgcaPpEth2uRiz8EurhHZREUgSt9bsw14CfmEz7YA":[40,31],"Emqe9upNXhojTRVT24mMxLpNB3Fnaoa1FibRtVELunUL":[32,30],"EnLMAih8NTU7ENb5tiTHqgP8MtTUnY4QDLUKrsRwhjtk":[56,48],"EngVeJ8w7soeVvKwypuSutnXFPSWDLMq3Vw5wuAdSGjf":[52,39],"EnrrqXLwEwgDU9yfa7YVCxmm6uj9vjhCnNUk3qjpFgws":[48,37],"EpfcKyhvVCGXNmf7aRoDwtXBMkE8xCkuLk1YTakwaPdo":[32,28],"Epxg7Ft5s3pwue4NLBRXShZZTPs84Fs2tGaMHMfp3kPT":[12,0],"Er2zvR4xjErXdauF5MVqBWukgRCT5yKBEnYh8W2ZmXTo":[60,55],"ErbvzZx2Ss9GxizKyDviybhZPu8noHv4AM5vuzTh1ij6":[4,0],"ErugkAcX3gz9cw2KwYZ5vPwGuQQxhuobM2tsdoggqB66":[28,24],"ErxAGCPBB5wMWU7mgZRXUoNyYSnMmVR9689hd6CMTfsS":[52,33],"EsXshV7Yva6ZiY5P3u41vjWYNHTNMaS19qcoGAdiPZK4":[48,43],"EtmHTosfXS5zbDbTd6RxWvGLWwmT6fbjR8YENZ6byfQ9":[56,49],"EtpFdJnQ25ZJMheLyURzyCD5ch1SL9smfMcEeKfAkEHq":[36,36],"EvDnGZpca3pWC4tW94U4PUkruVUkq7PUgpVWpDfx6fxV":[12,0],"EvVrzsxoj118sxxSTrcnc9u3fRdQfCc7d4gRzzX6TSqj":[8,4],"ExyBDh4ajq67itKb6HxuWBN7CFtTrg6NbNCF91mFjyx":[44,32],"F12Ah86ymdNPuXya5i3PKG7jeLfSMGpoRTriVTgcXr15":[48,39],"F16J8jYx76jpt2vgTT5SPv8hJGGcrShzCHG9LBV5vQD3":[56,43],"F1MevczAijZ9ZeWNn6T5ZHtiVkn3oSepvmcjCfPvdsSb":[48,28],"F1TuusSghAobmbGAgNrdxRS9nBjwT6J1yyALUvpEA1is":[56,19],"F3S8XEVrUyG8sN4uqyoU5hkyk1zRXwykE8QWmxA8QCBB":[60,56],"F3bXikq6WnjMQjKcvj7U2tasv9Q21xTWVUTp48GmhVas":[44,42],"F3gsehGvHNXtF8mDbGVfB27Lq1paSgTiqe5nzvbFVREK":[52,22],"F3w8NqDxCxS2eL7d2Ucxti3suRkqSg8ox63uHgafUZCW":[20,11],"F42UBCuZo2WxyZRG71p7SKT9iewEE131hyVpr5hG5kBJ":[48,40],"F48TattaDuDLAeEj9nKvUL9aq5vTJTbx2gv8zJJ65hb8":[40,31],"F4R2g7TnRmr88GY9DjhFo5Ssk9Ji3phBRssrZL5rQxWs":[4,0],"F5DXubJ3HqdWAV39GUX3rBUekzuzdjoeJwfKCbNnWKeG":[48,36],"F5f3vcPpfwgouhVVzSW41XZRzch16jr2qx7pYNYyVruW":[52,42],"F5vwBSZUjzdQxwWKdUBSxgsdqXpzyhtr8qiYPL94UqTH":[60,52],"F7FgS6rrWckgC5X4cP5WtRRp3U1u12nnuTRXbWYaKn1u":[4,0],"F7zQemwQJo3bKVAvpcCAfkgXD18kZxYvgMxCP3X3qiK7":[72,49],"F8ZZW4WKUx273i4L2KubqUCVSKLmSgkpxwHRL7gar1F6":[52,44],"F9YANP9X2AeorF1ZCY37GX6NXKLyouWNWELk9PrdPVCy":[56,34],"FAmkcNkVCrHXFcmLwBmv4T4Kwb4wDPmTUmoGc5qCwVeM":[28,15],"FBwS9NWnUfsNWAhPaYm3nzZaWVZkZuefEHAUWHzUue1Q":[60,36],"FEKzY1TLRYWDc7AHTkREpoSHvx9EyNpPmxp9FeojPbJq":[56,43],"FEq9FL3hzRDMtL8DinPAaeJpb28GBZYTpTeRcRyHSrGA":[40,35],"FEzZYThFYwyBn4gCqY8Kb3kfJSUco1XTYdVDPgdgRhWa":[32,27],"FFhtic9yPS8ao7Qg1GKjqyzwhGYK5tsksT9VrLioTgbY":[52,47],"FFrx4NAJginBWNm155TXLgx1annkmdqEwAP79nNRwxjQ":[52,47],"FGzASVsHGu7NdSbrzuTHyGNViE5CdcR2ZAz8MVLY1jx2":[60,55],"FHZbKviw56w6LpFjDaz6MSvwXsJNYS4Gry7Q2EwjDXBJ":[72,60],"FJBA1Q7gUvEkEsSuNDfDgJEKeiJmfF9Z7tF93sx8wNBe":[24,10],"FJMCiuiBEbKum7cLx7pqi1xWeNbpUuPJBKUeiBTkxLdg":[56,39],"FJzX9Cs7zwo5AZKUjf9wkBHULSJr6DaESAbUbBPhx6E4":[68,57],"FK1TQPnYVzg1e925kHurusgkXxxFuBfEko6D8ZirKNeR":[60,47],"FKErtUGcnKvAd4xizrU3jyoe9WWkCSTPZA83J8NB2rFk":[40,25],"FKYKRLCFmh7uUASUQjkL91yXCUuJ9wdbPCZEnv8HkKnf":[28,16],"FKny5Zv2nrLFKfNH4jatujiiNG2c5mq7MXweKAEBBzse":[28,24],"FL117azyKxNeDWGWEoiTj88ygTHFZNs13GKJHa593GW1":[40,30],"FLHB8AGEsED5jAF5sS1kSkAzSXVK23iuT7YDPHGmbcjb":[48,41],"FLcUYvDMd5nh4cyP3oErMHoKnKREmza5rdAZ6XHYU1bd":[48,43],"FLdAnmYGeGmwJY3qECfcZ8pyQ1LoTAeBPm8YKFDxQrMN":[56,40],"FLpMRfbSMkBnFXDdGKdzcGP8JgrNVhaYowmtArNughqt":[48,40],"FLpouJ8TALG2zroSZ2SC8wGdhQwyRjyyPetmJBcmcKQu":[44,42],"FMHjnmeRLszDSDTmHrbqUi7rpXLcrynh9K6jQvjdhqf6":[28,28],"FNBpvn9cNMmMA8GRfGxaD5P5zkG8m3YAJybgJkVi9bbK":[56,43],"FNCnVhef4ZbzxMsKiRsHbjpiciwV1YcKmf79NzwZwDvY":[8,4],"FNH1XmR5WgK7CH7W7YdcfxtdgaKFueKtqaggVr3CnY7M":[56,43],"FNdoUuKVBigMFGpVvSMLXJB4FC7XQL1RjPUqUiwvPiCS":[60,38],"FPF4V5wVGhiM4UxLLFA9LoaMFLs8pfnGrYhsRrjS9xGk":[72,65],"FPncJ2ASP4cVPcmHxibgpqaxMuwjaw3qNQiZvVdoefDg":[44,0],"FQFrdHAhKFP9R5R6JkJPtVJhLDivDia4cNNUcL6Eja6j":[48,38],"FQa4mYpWL7mNEXe8dWbd2FXxpreFtYJkD2S6hMD1oXHH":[28,24],"FRXTSuAxj4RaD2x1kEN6cgdcuMAMMdSEfaXcUhHSu4sM":[32,27],"FRkEFxUbT9e3GCJWbjkajHF2tcie7dDeH8rwDUPThxZk":[68,57],"FSr3fGXzbMPBwWhnWY6oMSAdZqiVXxw4EpwxRAJpu5pB":[8,0],"FTUh2jo7GmxFqLy8c9R9jTPapfGjwcaDdozBjhKJz7UN":[32,31],"FVD3oRjHRyDsGTwisgZrZPV6vbbRvkgx8hm6ctmBG1Ne":[56,36],"FVeAQHyyBjnuPDVHb8aFk8FRejULrB78K4SYyuuj8Q2T":[48,35],"FVnG988wW3uF613QVxmQrkwtdzS8taxjFsuARTzZBwMo":[36,26],"FVqVvYtA4CemZM12MVK1gdow2AELmMG98hXo34eop9p4":[32,27],"FXGBEHAr81PVbD4ExKesR3AwXp7SwP2wX3kLsjijd8eP":[68,57],"FXVTpozaNNzwiaEu5HbS9EK7HMNm7QPLk1UEH55hAkK5":[36,20],"FYH7U2HPhgxQCsHBGDaC135Rsx4tZx7P6ZjxnGqWHBHn":[48,39],"FavsezxBQdxoQYBioVPoAuwH1NwBqCQhpNdtuiHdYyZB":[68,55],"FazMcimRMw64FWqyP7LX4ATBJb8Vm91UCdc5kgmBzqTY":[60,45],"Fb5cEcYNgPXKJoEmvPvsU2ENYRVePQtExqgf77AnVX54":[52,40],"FbWq9mwUQRNVCAUdcECF5yhdwABmcnsZ6a6zpixeKuQE":[80,58],"Fc6NB99bkJQn7JsVSqdJs5fJEzj7KFpe4JHNQCGVCctj":[56,40],"FcTYrxp31zVjTW4qjFKkgRcKXbWcBbiRQqJYpufwcJZN":[4,0],"FcWgrc99RAix3y9th526GnzN23MQSkFmyWaeo9xJ6Jfo":[80,55],"FcWjjkFgHgaARvF3Aa2wu3C81B26CEPRwNRwaHZxcZHx":[80,65],"Fceiwrs9CmFC3ZWSZWYUVDKPH3CR7FSevtQbL3e5XbTY":[48,39],"FdQgwQ38ETKv7x1mWNoAdrLR2YZyp16xFDC9YR8Gseva":[16,12],"Fe8oZecAAGKLDGziBLLpUM1nwD6DyF2pXLMZAkWjNm9q":[12,11],"FfwtopRBJWKEJiCmkNUFyaQ2FMubtzMhAzKgHDF7XrLa":[64,48],"FhUjWiMEFauc8nrypsy4Mp4eGpCme1ZBAJQadmYg9ULr":[64,44],"FiEHvgPh2YMigcYWZtGn2tCgPFefSgTRA3RB3e11GhLE":[64,60],"FiT4sReWDW4hrcPHW4bCmFN4GQQJw5cJMNeHf3VHVrKX":[68,57],"FinkBUX83H7gMds4pBwaTntpdrK6ML28Up6Panpx2Atm":[52,48],"FivGzpupCvU4yr9E3J8RvWtLNWTm6ZRcGS87a51BVHWS":[60,52],"Fj39TNbn7GHy7kbvPvoCEtSFA1M3k8885FaSGc6WxqD7":[8,8],"FjVP2aWyBT1ZWkQWaC8gVcQL6jRrNceV7r7qNjGciqY5":[12,4],"FjXtJ7fvCGt1HRmYrhiqUz3r1DWDqfBwZCSESDPv6bSE":[32,16],"Fjc7dkd5ir2heioaU8eomUgbX2JY49BCqBX3doB8o3H3":[24,0],"FkHabRHvsaNNVQqHEdyNAUnb6mScR7dVCZeL8d4ftEhw":[52,44],"FkMj2LPSWd9LzLZrpZ2L9YL1CpB5eA5W3J1vyvvpp6Jm":[48,36],"FkYhpz7HSGQJvA6apj1BKoUfytQvWseLfSUrE3zjvkQb":[28,19],"Fm4UoSZ4jYnTuLLhav3A5BysiqKani8M6naY5VLFzMsq":[24,15],"FmMWTNcijryGbzDrbB4nsU148YQarwDnGqVnFwW7kexs":[36,26],"FmZc8PpFoPLghneqvw6ZNd6HohL9uBWUZTxh1c6CETrh":[68,55],"FmaWRHAtnhTX3iDhgTzaFHpcuP8TgjWux68zo3kJxttT":[48,44],"FnaAWEDBPSNCUk49EM6JFS2hr9kGqXKkEDXkw5fdR887":[36,28],"FngAas3r7KShgnvmgtx4oQ49tF6q7WpUSnxovVsSyyAX":[40,31],"FnpP7TK6F2hZFVnqSUJagZefwRJ4fmnb1StS1NokpLZM":[4,4],"FoDccJmq4PksAoMpRbygVVocdp4NrC8PSwwDd8nfKYzv":[52,34],"FotUJq3ueDw7HdUj4R73XUJt4YYn5L1QbEADVLgUCc5G":[56,41],"Fou7Du6KtVb8dVMzKMYW39fuSGpMzJGwpkQ45NbxA3Tx":[56,26],"FpX4XYrSFjBcUNgt3x2p2SwcHaKSdZxNDw9KYRjrnKbY":[44,20],"FphypHn7XJ3HisFtdnUX2VdqsWNLa1ATwrA6ytR4VSbG":[48,44],"FqdAcsUQBMibVJVr259uSAnA5FMK2xACaz9vPEtdvkYn":[56,48],"FqheXr2yJSTRGncTqVFFG5sLaTtXZeQbkQAxbL8mcGru":[44,0],"FqrYY7h4wFofqE5DPPfdUUTo7VnZqguhTc6pDx6R9WA3":[44,42],"Fr1BMG4DE17Lu7Jj5kG9gndjKiBvr9kHyqmoFMWEpA4r":[24,18],"Fr3WfD9xCLX83AaktYvcihvqoaJKh7f1AnD8mCN1egHw":[56,41],"Ft4ADhkxMVfgxNDQFqA3ymaNGC39rCHdUj6H8KEWQqXy":[12,11],"Ftms6EEKfgb3FfaoJJk6A42MqQXVcU15RcRa749LLikL":[52,22],"FtnMU4gqaiRgAJ4uFDFARAp6WbCtGcjxqck15L6Q6EEY":[56,45],"Fu93Uz1dJHhp73tRFDHLZ9QrzJMyaDnorgfZFGkrBfoM":[48,40],"Fue9LZxjhk2DNXWxM3rPKr3z2qntChdeth615m7zUo8v":[40,31],"Fv4zJ7RvV8gEYxEtLjnGZAX1qxjqRh56DzBgqvFEVjjM":[60,52],"Fv684z5SvpMvZf4e1aaTwaA8j1kfnkxd3iQNBdyFNWKe":[36,22],"FwhKmbdbaqWqSMPimLFbPGwZqhpbPJEECnhLdURrc362":[28,23],"Fx7dVi2oVpynBKzU2V7nRDbdfBjWrqqjLFxULXCVp2TB":[68,43],"Fxv1ymSwB6tdRCbjBURQK6P68XR2njCGfWbnfzVciJsP":[56,43],"Fxw5NgncJhGbnjH1wuZuavVPDQcowTwG4wKiWAB4WNAw":[60,47],"FyGLvXzJKfNEBFHS4ezGrWZjHfbqdMApYesRjt4yZD35":[53,48],"FzAv1TFpCyR65GrxeqBwnEzNVXEeUMPV5rKZGQhPR7mq":[64,54],"FzeiVQGdYrFWa8p79HGPJdnvy3zTjxmwEk8fuMcDo4U5":[36,27],"FzxDXiFjj1Vu7HvFW2wc2WRRV6swPRvcPnMHNTCd7QVi":[52,37],"G1XAtQaBj7tkwZAHwmXcAKSzMB8C9Kbfge52hZuxBA5B":[12,1],"G1uCu6JrV683QK3kdAzEiiAEBSMk32Ugy56u685cynJ2":[52,43],"G29KDaE6ed3YWzWaNesjgoBu5CFJrcHe9sb8dr7b7LLq":[76,56],"G2bNSPr2Peyi7CmyJ4srX3YdEvu5egpQiorjivMsgnrs":[68,64],"G3XPqL7h6TtKsenCBXi4NftqcVGwKFFVGzxeVbkbAKMt":[44,35],"G3tnUHh4JgDTN8wnesqZqrWN7yA6hvdcj9kRCMNmEyfQ":[60,58],"G5F9RG5hd136BNBtewsChY1E2jnPoABw94mZmeConsFv":[48,45],"G5dS2hFkFQcx3ZWTeLSu5RWmPxjXV1RY6VRwX4fmejQd":[68,21],"G5r4XSC5D4Rw4NaWjbgBKnj6bNDsSGUvE46w9BYAT79r":[64,47],"G5wSAoZRgCw6EMXsqkRJyP38wkyrV6YiyHup3PSnXCXR":[56,0],"G6kUvH1mLRuwhSzv84EwRXT9fpZyUb53Fjm2M2n8oQo1":[28,8],"G7R7QFy95eeELpCgRTdxFSJbGp3EdryFKN8ou41vZJGA":[56,39],"G7cSi3avELxMLTCossnsooLj6UNhnfER6kpnSx8NKHfM":[48,47],"G7gAgJpRHnRvFhrUMA5khWMqHJ3tpWVWdpsBvCq6w6MY":[40,30],"G7mFk3fX4xQmBV5je4926SzLCphWFoww8APYxQKfkNxn":[44,34],"G8QaUmUwzJ9z8vu2XbxctG4eu336mfg1oPbBnWXpSE1H":[68,61],"G9AHwpSz6gRb2PYQTj13oouRmpN1VAdHwXGzb1eoTfCT":[48,43],"GA8fULGnRXvc1K4bnue2toKKBcgqrqoUzzF7yi7TkvHS":[60,48],"GAs5dt2Xjtd4f1mqZob6hKhY7H2J2HEAc8oFQ2fEjAcx":[48,45],"GB1F3KY6GJpiCoNWWZuvMiquQ6kbbtiSFK91mZSHEJ6g":[40,31],"GB8b8Zure8MzwQWPycAcEr9Cx7ARprAUoh8fWG9KiVLY":[52,48],"GBeyvNz19UahKPmJAKVuTUQLNWkYVpCbjDqExmYtgVFP":[36,33],"GBtNh6c3Jbf7xMSA4VCAbjkJpBVJkB1QsZLh7iD6Hpq7":[52,46],"GCBBx5HU1Bidr3xVkc7dox4HRwpBwg9Y81s5h4pmVUrt":[60,52],"GCPW2jinG8pk2KfALJA2FYNhLKwCR42Y8ccQPXz2PYg":[60,42],"GDoZFWJNuiQdP3DMupgBeGr6mQJYCcWuUvcrnr7xhSqj":[52,40],"GDvW9BczzCHnnGLK7wSeZSkGQqP1ie4qcXuoaU3KR1mJ":[40,31],"GEAFsEHsXFJzEUvwHxoy6f57a8GfWq4HMZdaAY8QSHEm":[24,16],"GEBvuMyPAM3Hmsr4UnGMqeeJNPiC6ZqPkCGKW6pADd8h":[44,44],"GEeL5VToy7H2oEFyLW5q4T1HeCDa6AA5YN8r4PiCnyW4":[36,32],"GFFuGhyHAr2fjH1DL42m9EWpAWXXdZ7R6PyzuMzDodLy":[48,35],"GFZ3w5CU4Byjjo6BrQxVsiq9mumADrR58vH2TT2KxmFC":[64,60],"GGQFKUi8FWGSAnNWeoZdpwERRgaW4VMiBdiZbEETq5Qz":[84,52],"GGnmHbA5wzvKcn9kTcy1Q1JgdY8hQoHAYRk9HaCBNJzH":[60,46],"GGyi5gNaZYHpij4uM7UE6EuN93dvyMiKMphNWLfTPBb8":[36,18],"GH1t1LvHefMhw9y7W4LNWWa79HHnB1bQQcXGHaTc18kg":[4,0],"GH383ZjUf2L1MRqbjNVGtvUsJMUk4Wx3avcF1uVDyjLB":[36,4],"GJRLu6i8j4CJukLEQQXe3y3pdk3ynVkt7R7ttcfCZBoA":[56,54],"GJbU1XAJAky6sPVH9BGux11AGyFtkqZwtiRVS1itqcao":[32,26],"GKGba31Zdwu7qRLBYGmPCkk9wwfJ2WH2a4FKxb3xfJuV":[28,27],"GKKznSeuCDB89NqRaVhVPr7ite9bsG1X7ih3uc83wpUH":[64,59],"GKPmKbjhiLqXa1Pp7KDQTF78WfvYxG4oyX2rmAAEJ8FP":[72,70],"GKaK31dY2nb9FdfYxL7jzge5v5BUo8M3APLarCijr4iv":[56,49],"GLKsDBjWBaXHkyMihjpU5ZdKyKWtUpJyE4W7PjEFSEHh":[56,51],"GLPe2gV9zG9kwJmNj9GFBrzEiqridEd5KYgeSiPrubGz":[4,4],"GLQ75DmDNSn3w4RMDQHNVeWjhe8d11kREHJampHUTkfd":[56,41],"GLSFJkyLwsGGNr1MYZwDCFVzkVRdDTdFn9eMk17reeyW":[40,26],"GMhCc4SBnHPHmL4WwRFiXaDkc6qyUpeifCP7jcTim4LX":[56,48],"GNDHwncRV2VpzXQLpGLvQUGkkfDFhA3g36Jg8ZthCNLm":[48,44],"GNPK6pfoaXcz1sKavyYRdAD7EtQy3F1CXBTWVyxEv2xu":[52,47],"GPpFVDjENZafNP8uysYR6TtnXWdN935z9jNDocP5hbzB":[56,27],"GQCgtjErUk8HCgayXyjuCFR15pnssi4saPnxD2Pm9oav":[48,35],"GQnaJJu7h53SVVhVpg2ErkSKhYMtqYrqv1qr13MUobuq":[52,46],"GRqJ1rtEGdJMuraEM5Z3oDbZ5k9N5mW5jkXTeDYTT5rT":[20,15],"GSWzDkiWBaLSHCpyWzFwoLyaCnizLMvXtk1FjQSTimwU":[76,73],"GSZpBMHhS7f8GT2dFT4niF9b8nLNtSkMecXR8thMnkSV":[64,51],"GTHTdVpF9dXusgSWNDQdb6uWRm6NFHYEkMj4XqZGoFEK":[44,31],"GU3NxbUE2GrZRiqzAYuvEr9h8rgCPj4GtqznSqr9yQiA":[60,47],"GUMhwogpDUNSzrFTYzc9soNWQxVKR9XJR3xYttPTVA8v":[48,44],"GUq7RT6MzPK5vswngQpJdJkvtUDWU9cLarM5Rec8VW8C":[56,45],"GVS1B7hSXnUQMQWN78BWjT9Kr1h3k5UW5UTdFsDhtRGk":[32,20],"GWBQoDvmnuANcRCxkBY6YCT6yp2MmGb8NgBF6XNvmdUo":[28,20],"GXYeHHFjnnKdL2un96tLmSduturuWzAi8MEBEYH8WYYm":[64,60],"GXfJaLrWgQbiuutiLyN7ijRgBvAvunJy7bYzaV562VWP":[56,44],"GYNPtzPmbceWBZjTe63zeASYXdW83uh6WK83zghGQqzf":[44,40],"GZeAno1q7JYL493V8aED6XQzqNPjJRZMq84jtKdLhWNB":[52,49],"GZgLom18cFaUKpAhodJ7LPzKCxyPbCzX5UxSFuYbdUdY":[40,24],"GZvmX5ooGidthybAp6NuMuCTyszmGX1aQCs5HxWvDb7i":[28,27],"Ga3U2KVbMn4y6z3u5SC82DZdDYUVAvnpyq83KERD6B9k":[48,36],"GaAb7uwik3bGsMurHJNmbabF8G8k2cYJy4Wrv3tefuWc":[60,44],"GaTDMHvngmoJhuRYrLGcE2GMCofu7K8SLGwkCkDE1mYh":[36,36],"GaVCM6rQVycyYNDVAjAKmnmCMnbe37qUuGmyTKhYPCQM":[56,47],"GaX81Beco7LXBwEpZrguBAHsQsmtsMFVB3EXzc5jZGzV":[40,25],"GajXkJZAas4ZjTKzgXQc8vTMyiwQC8RiFPreqRGBB71b":[60,44],"Gb9j79QFprtbY4sieaLZGjQr4a5ifFxGdnU2qkjPbBJQ":[44,38],"GbY9gVU9wuQKdWXrw4W4i9dBH6feAq3BpPwDKiFZecyY":[44,35],"GbmZkxUgNcxxEFiHd3hMqpBnwSPyRYz5c2Ya6UPFchQj":[72,55],"Gbqh5eq7nVajFNocG1GZGikoiAqPstfQjHxdGBc3SD4M":[48,42],"GbtVg3D6bNFSjem21vrJBJTpUniwwEtmvs8mQkX5XS1V":[64,48],"Gbwg3HCbD9gzma2o6oTqYkDQnojeZ1z7Ygc95DPFA2me":[28,17],"GcibmF4zgb6Vr4bpZZYHGDPZNWiLnBDUHdpJZTsTDvwe":[320,276],"Gcu1fpYPwoyBm7JjLJsY6H2eHRe8J5bjzoXtAwpy6m9F":[44,34],"Gd7zbMcA37tfU6dZYi5GfstUxGowwTd8CQPdDfsAynKT":[68,65],"GdjGkagCgTkVE2rwPdPUy1KXfFFmihD7GGzpZzyRHfz7":[28,20],"Ge2SFnQj7BeJsVaNqSrMz3XFGBjoFMMZ8qThYZRYYNr3":[20,12],"GeS9DzQuMz8PnUnUPPWmr5JrMdrRKDo7RiviNXET8Lak":[60,0],"GevceSyTLxHv55phyp2PirpdsdqNFZRZSYViRCrXmneh":[56,42],"Geza1KeYHg1EmwTxfNWpvZN2MTDxNX5aD6kLXw9ABuDT":[40,32],"GfT8F8MgHwdNkPEhytScMFL8hJcwM52uc1fiuj5e4YKh":[4,0],"Gft346NFxfieeCXCHuwdQ9TN6HyPLr5oyfwGS4DGQWGt":[40,19],"GgVX5vxGgZqMb2M8127Xy6AGA3kc6BCoFUr7Ex7rrK5W":[28,23],"GgfiMJLWSKHr9JPR122BiyEXgCDDmFQuzzkKMNbNkykk":[4,4],"GggvaPf92W5X8u7TxeyL1Aj7Ztn5ync5DjTDfpH63ffW":[44,35],"GgmneSMKWnEcavporN1vPpyTun2QRBzCjFCecQT5km8y":[32,28],"GgqsycDuFrqcM1isQ25SX2X3r8SBRqSwhH5hMn9vLFgo":[48,32],"GhBd6sozvfR9F2YrKTFwEMqTqhfUjxNUYKxye7ZvTr5Q":[60,43],"GikkfYtVZgaUtcmreVpQ1Eamw7mrnf2jnDFGJBnhVQhG":[52,49],"Gj3QmL769joJq8fszxqX2obCfHV3S4ffKPT3rm6xUe92":[40,35],"GkLRAjKa3b9gazDrc2wj6zVNbLWvibEKTuhsRp2i9Yni":[12,0],"GkYsHzF1h4uZ2nGxykLqYJvVK91n9eWHymA5FhTrmRns":[64,58],"GkpQYmJzR81VjuT4Gch8iF3LECvSEfVJomNqjUE1Ef8K":[88,58],"GkuEkzsEKyhgmomjiS9ZxruL9oc8tTWYQHaUv6Xa4Ch3":[56,30],"GmAuzYVHNBDTX3zGRwHqH85hnbBaxQdTbeVpyvAZpufN":[68,58],"GmgV3mnVohRz99rsnMNWNFqzop4oSgNv6Hx1kE7PKvYU":[24,14],"Gmw9GarCUcQNYnqePXNBREuLhcMUwXhQWZMAvxSUf6c2":[52,34],"Go3R344LB8hSYfpZJKs2LHQVRJE4zsm2KaSLXbRygYbd":[64,47],"GokfNYT1GH3c8BQrXoAJBypARuNFWcRG3xa5KQ8hKwPe":[60,42],"Goy1QEMaKzadc7y6hYPbNu5tzQPZMFRLM7oRehiuoeDc":[72,59],"GpXcoJ7jRzCEpoSLERYQCxDi9jQ8oJxCCKqjEMBnMUDQ":[44,38],"GpowxwT8wY9x2uFLWhZtL3ELFdAMnpBxTrpqFgnEukVn":[48,46],"GqFWgFDHj6fgahisFk8TEngmsEdkSxbmk8ZktpgW5LaS":[40,35],"GqsnwvnnwfvevfovAfRu8XrJwGietC8h8t4dwyQerbfq":[68,59],"GrZcGUJ7baE8r9KSmrNJAKtgYAMiD7p2YfxefkbgTng9":[100,84],"GraWXC11stqBmL86aJW8cBEncRhAFX86mX4E5pE2eKg1":[68,56],"GsTKJfSxEXvAvw3Vkw4cLzBFzBUSFSqvx7cW9qnsZFvV":[36,31],"Gsooc16Z2JNRxfcsfGY16pvJ3LGaBaEsFRR1ANDdixfW":[48,32],"GtU7wyz6vwTo7d82qNpFM6zsxWUnN7caxNMaxLwbwCEr":[68,43],"GtgtQLfqKjn3gaHuH7Fw64n49vr2DrYHiJAsSTNNscAE":[48,16],"GuKn8nEJwUPjBfxpwyq2MXU2JNrSpj7gqnKptCZeEk7j":[28,26],"Gv4WPocqzi59G7sbGiVWZAtLwDohpB8xSMG4kg41g3U5":[40,22],"GvWoMZaf8ZbTnKKxrTGV5aGAjofinufohuqT827waAEy":[72,55],"GvZQ6JUcGiw26huYVv2eDTgrgVh3rKtADPYLfBiznVda":[24,20],"GvkYeWeoxH2QzvDrv1Rqr8ZmmMmj6xxZbbBWptCDVX4h":[28,28],"Gwgg3usfEkB2dFZutVWsnCwTDEbTo3AncyYLdpjdNwE7":[44,34],"Gx3a1YzZCLrih1R9FPnqj7yzW2ekFWHVTCM76Zq41D63":[44,24],"Gx6SwGTbYAFrUeBRMgMrgLUKaeGNeCKzkULXdEpwPSwc":[24,19],"GydWayef5RL5qW5zfTjZAqd8c2gPBdctpQRFr9d73FHb":[20,19],"GyoEkrR3Zsfq1FayyXfEcKM2UtTEWztYAEQBZHxqEmTP":[48,46],"GzFTdRZgzMdaTc6nrkgx3eeRbRtsTy78mXNHvzMkYJAu":[40,31],"GzeThbXZzp7iSQBsCNGQU2CsA5YDi3Xr9Gdm4Sjj7gP8":[36,27],"GzjJmb2w3vt788rDHPNo8hCeXvd4fzu4n2jwkoTRBouv":[60,31],"H1xNkt47PQ8HCjUfhoUEtMTtxRRphsqdHXY4B7mP64oG":[52,43],"H28YbKcxkEekLDHmTYnRStkFBhpVdNckNzBQSvRiWHmR":[52,44],"H2qBXkxdFh3XiKbAkJVfRsucdVkf6uCwJe16TA5VZeNt":[60,27],"H4Aq4RPa5coDMtPydkm2WyY8gd3k5R32ieL5fo8QFBY4":[60,52],"H5YdwNcDt6DNeAL816CP3Kn3fYXFoigMw3zaASvZc4rA":[60,49],"H5aQjanQGz5Rrh4s6g4TNHTAXDzdr4czBtEJPdeDJ1jo":[80,68],"H6djbzHAiv46Wxy3iwqD7LA8ART8YbgrWyxQPCNhnLPE":[60,38],"H8MUh74GVNbSqGrYkZviws6xCmdVS3VZF1rbhE3gSESQ":[32,32],"H8kdiUSyvHbxshcFmRqWTB1HZkQHKcQcagQ56TzLe8ib":[48,33],"H9PJujAtZMZJ4tAoWeP4UDFjQtYSCsBtUfpipD4nti4D":[52,46],"H9Th9in92sTsCxiA6oBe49vUW8PPv84HrN2g4KfSVgnM":[36,26],"HCikGbQ6gUseeVTvjwe4GZx145hpd2JR57DcC1DjecrF":[48,43],"HCkbW7BPAcLQ7kFwg89F7swnmnGLF2HeQ4zMNqG1YYoy":[56,41],"HDitfpmCcy8WyJgNCQTrnZ8r71Nn7t7SjmVHnRwumGZi":[68,57],"HE35aDYTJHJ6KA7kLEXvENiRBX8c5UG5xHzgeKiXyQno":[64,41],"HE94g4Qp39SrUtBtLXnjEYPxbjxMb28u7JvrNwcqGryX":[56,55],"HELPwwfg5W9LmXv3axe43EY1YGJjfVf3CcVjA8BZM82P":[44,38],"HENUtcTb7dTxGHDYV9VeLTgHy1DKAMWLwLySym3EziCr":[56,48],"HEYsakKxLuuEWsSdH2cevwXJXdQ8tX75KT2SHWCkerHs":[52,49],"HEbMY624UhDGm1Qhy6neKSyi3bQjQ2RidSTyt7ARK8RW":[72,52],"HEgDHQD4cZsVu3QEjLLE49EQmuDoojxiZZbPywgdX9dE":[76,60],"HGHMEEHCfbVFjqB69Hu9oNW6SviukB8jUheEhYVZJKe2":[52,33],"HGkZ4BCw3mHYoo74ZwwzKJsSjnH8u5BW2poHVVMLTjuR":[40,32],"HGv9NFa8dQeCCZcua4vg4Tqa8FpTK8AWBAQqqAE3z81G":[60,0],"HJhcE1XDYTRoHaDWcfkmfGvJuDWPZLEK8YkuMw9FYpP3":[64,49],"HKpk8f7t6MB5gYb2uSP8R2LZabMRezKqgCAVSb7tmsqQ":[44,18],"HKu753Hd2F1nWLPvcNZHX6RAGSXkg6AtywiVvRqDXxcP":[64,42],"HM2hzFLTd5TAhejGFjaXAm8LLjdmnj7bqQrzpRTaawdo":[48,44],"HMtri4bE9Vs3pztgZRtU21M9CuPhHTanZXtiuiR2azKk":[60,41],"HMzBDqxq5as7JRyd8PfDyvqL1LNF9R7yX3YNUVi8xT9m":[4,3],"HNX6Tba28Y4o4vfU33s5HkYefZcS8xNKY3R8zwxwor4z":[28,24],"HPwNj9cotHgFyt1Z2MhKsQTU1w4LxT2R71GoSdGSrbvP":[60,45],"HQZpRZLSzgDdPc21U7nCxXpK2VjMh8U4PE5G3YE9H2Y4":[44,27],"HRCBqRPZWyxghiRvCG6qQsPEfnnXvDeMGQwXYTvuGkKr":[16,0],"HRddSLZVC1dH1uDvuKKqsJCN9pbihJEmVDAVtyL92JY9":[64,52],"HRjJjGh1T33Rv3TeD9oLLEnkZVVvezP47kXsCbdXgJfo":[40,35],"HRvSdkavd12ZVGnoaDV65ogTPks4pBERY7uuwLG3YxdY":[80,59],"HSPysEmZeXB1VsBnpemb1AdMMjv155m4tZ6LYq2uzWSd":[40,32],"HUoud6qywaWj8kZwdHRTbEPkKmskHa6Md1KNvF1JQFYF":[52,26],"HUtFMtq115zhNf1ecuHHhqhP3fJupC8vt1wkWevHf1Xr":[4,4],"HVD6ZDBgzjqYKyDLNadSkzev3qwSUnYEs6k81JktNuom":[60,51],"HWKwGWgWpnt1HYo7kAbtpuNzi6PovGXz3oWW54GDEQWc":[56,48],"HWvSRgESdWKDccWN91iRVQLN4rRyuCbuAHVWtPR1cJ1C":[28,24],"HXU9vqRYsM6jF8wGwXAAFQW1gFaSu31Ex6Q7d68SmpkW":[72,45],"HYW69eojAvAqiPfebT3S8yUTvTDHnssZbTq1TMCm5LfP":[44,24],"HZCUCLqV3P7QqG1oskLLMJW28zuckhxmRzEQ7UWaH2U4":[52,47],"HZX4MWsSDzRerGuV6kgtj5sGM3dcX9doaiN7qr5y9MAw":[72,61],"HZwbEfKY3TNJ9RPeAYAH6zABK8Jr3CX8xXrDS7fKsc9x":[56,35],"HaDrjvGfvneb4Rs5LyUThKY7PEh9QyBKsZoi7jNcyune":[48,39],"Hahg6pT1qcuRR4iJhuPoCCouwX4rh552ozgNJsHU9HyY":[8,0],"Has5UpRZwnb7TrhyFPDETgnGThLPe38BiMcqkvXBLsra":[60,55],"HcasCt3HSWS1J2YH6qcVy5WNiKuuRqQ4kekN429dbcMm":[8,4],"HeUNAoWmKWrteB8eJB8SnNn345pzM8ymfZARzVrvtKFF":[92,68],"HeeEbBAkuLzqxsFLcbKUfWmeNizywy2uzAfvRg63LFT2":[60,43],"HehQKDuoiNCbzovqzhjzSmNN2VBKYPgKLKyxNYqRsZhi":[60,18],"HfxzFiP19ymtxHagP4Zpga2zVo6ZgivpK6VkBKDowHRr":[48,38],"HgF66KCFTqcs55WAK9co7o9f4ZuPuXmFQTCKvRWz9A3H":[64,51],"HgK19TWcH4FcbtpWNdjU4gRQzBdS9GUokz4XbzTm9WP2":[12,0],"HgTXVg3dA51mGNSh9iPoeC6QLsy2cMEB8WPESRAQMBz7":[8,0],"HgmPwzNcY85HfrN3bYiqaypb6Nmf7ayaZEaivGY37913":[72,70],"Hgp3kh6Vv8iw5wHD86LqkW3H3JApJeW3F5XLaGXkZZW9":[28,20],"HhjxbH3vLpUNShQB34NuMCL1Qc3xoiNDbvALWrAMCCnb":[56,56],"HiVDGAGPSxxydKTY6BkjuLE3CyabGKyEuMMHc1yMw5Qq":[56,47],"HjT9tCUEFWrUXFR37ahB383QTF4u53KWx9J29EWRfzdi":[48,37],"Hjsgy7BuoUFo4WUr1sSvxSYK2oLf7hf2dYq818Fgc5Gh":[28,14],"HkSTpiQR4YTP29yRBSactwZCKh3fp7NoLHLQrMK58xRE":[52,39],"HkXUfo7jkpymbV2LuekirjvJDzXEREB1c8hfy6cPxgLy":[44,31],"Hkj4Y4QxFvyoCd2wzAswsDpwW4vD1vyC5vppVyDDhJ8F":[72,42],"HmEUD98iS9DFkt5dUjtrG3jDixVU53B3YPKPbGwSPrC5":[60,52],"HmSU5YJr4XK2SYdF6dxNXtF9PQRzbXXupUXCVEaJZX37":[40,0],"HngPeCBmKEooRBX4nMYYNvqic6QRDhXAmRAYEGD5WNr1":[8,8],"Hnp981DgpWig4dBQASkdJG8r7KzrgNRvGVUo2Em4ZTAJ":[48,30],"HoMBSLMokd6BUVDT4iGw21Tnxvp2G49MApewzGJr4rfe":[4,0],"HpMNvGvQ2MAQwKvSbgB6mdJvKENdZ4jVXQhM7Ed3zJAm":[64,51],"HqXHSTtrUUraYZ4xcPuze9pX8LbRaV4wQGQ92Y2L26vN":[56,27],"Hr3PCkpdBpotx7CL9P51Xi7Yj9mJVPvmEHKUmUFrNkr4":[44,30],"HrDkZBmgxaV1373agNpFArvae53uga91HYdRHLCEdrkQ":[60,42],"HsWUiXARLPhYitGMapLYyMdV7k27kW2xzy9Z6L77jKBC":[56,45],"Ht4Zd1QjJPkvEUsJUeFgkFNWoqwW49mZZesCVm4xKr94":[40,30],"HtC9DFLScd8PKewDrGzC2dZdMURZgEfNDv2ji43coc2":[40,25],"HtL6WWfAHCEQFHumzYPU3qupZzXth8D5jafz8v7tTgVy":[72,65],"HtNFmWY2Ua4zx2PLKq1uAxmm5iaLXKcV7oGkPk5dFYZ5":[60,53],"Hu7DW7BoXXuKbwaFJaAMEXpBv8pqBJPhfThMD96WHiJR":[44,38],"HuJHVhpsf9nF4vbTWjgqgcCf2h97eFf4DnhAe3txLERo":[44,39],"HvofH3GhdBkVdPctTWhiPmGKzsADqtUpr4tpLgss3NX":[52,50],"HwFvyMbGLkiTUaT66cfL1FnJ26c9VqtpqAg4UbWSXtdq":[48,42],"HxBkCVtiYAymCCv4EYakNDSCPgog3vBJMZx54dCceSyS":[52,0],"HxFyHVXiQMf1M5nFowvW7um7oZx3aR1qkcRJFQGBpo9A":[60,48],"HxnjZ5Qg59nupGGXVo77idUxfsiRXPcBbBt2hw3Nt99c":[44,42],"HxsJQAfgMFVcTqf7hfLBg8UzcCq4rQJdo7g61Z4i6ExG":[44,38],"Hy5Mano3fc6AZceKCCwooL1sb55KaucRTGAMxRxsZ6qL":[64,55],"HyCf5LyHfwnpnvwTQkfPWVdkqJJ2R2A8fBKb52m7cunf":[20,12],"HywMhn8fUgxVFYTnXdCCZGGVKK8QZ4Yz5AC2bgVaEVGQ":[32,24],"HzPFqFKsGRT3Yvd5Wgfng16c8q6e1bDe3W48fZbuuS9Q":[52,20],"HzzApCxMXFzUyeiFkJiVC7sK8De1tX8NfFKPzSU5TZ5N":[56,36],"J16NbAo8MJAthmm4kfrLdyTKFsWRTVn8Vaq8gue6eJMD":[48,38],"J1RpwhRqrLGUpuwazwHn5yVuUUjQXZWC5pVRoU2YqTYx":[72,62],"J1mnigj2PmzRCuLvjqBX3h6Lb5b6PoPt2Cvqu8g2wNG3":[60,55],"J2LPt6Pza5KcNfeou7WrqNKs7wu3v1rSeLnXY5iF7szh":[12,8],"J2NT3pjS4JWMMHb9Ks1rcycAEbLPSA4u7d1eMv87rAWC":[48,39],"J2RzbW77NveGcT2Rse8tRvKZ8f9dnN6bMuT8q6a1LBut":[56,30],"J2VXfywh2oc5eT1LAtSApAqUVB1zJypFTYKTBdJg7BLW":[72,64],"J3wqjuNuduVLKFrZscCwZtoy6F8HSJqbSFQJV3LSqHJ2":[52,42],"J4FGK7xXt6E5pfxBtQhGfEX1djdgsLNhdrei4s5ghSaX":[80,71],"J596ieDeK3eN42jD8Rj3LRwuLJTkGKs3Ju27LPBLhvqz":[16,0],"J5TxmqomwJVQJMSeD6pXo8vmV8nRRRF311D9BFWV8vyi":[60,51],"J5aJV7Pd4SZ2tGA6k5kmdHETXiWEZpNsrXZ8LBtRDUEq":[72,56],"J5dVAuWTHSppRogVgdinqaEHgkkzzKYWdkSRZue5zpvi":[44,33],"J63rQazpR3qLHBz5DQLg5NB5xKDWAe3rLxGjpnJmZvmp":[32,8],"J6YG6AQd1AqZVsvBU3VCuHwzweFYrQL4Wdrqhpur9pUu":[68,62],"J6zWiwZBuMreGBkKQ6eqkbdESgp7BjhjcHBdBfYMZ64a":[48,43],"J7v9ndmcoBuo9to2MnHegLnBkC9x3SAVbQBJo5MMJrN1":[540,450],"J8DQkLZArBMLSdGGfMSdarnUwk7EybVjfBU1wHpZbzG":[56,46],"J8d11HHB1ttEf6wJpFdUhXvpyZefykDpAq2UDcPJsunW":[52,16],"J9Y9xwDkqFgiLypFoFmpC9MtAqkt7B9CTWrLUvyGZfKV":[28,26],"J9wpknjG6QdFZf6KjVkoHcTwsXsfWx7KeisBawMeFHPD":[40,28],"JA5W7X7BxDTt4fQZtAMCFC5Jf8V8beZ1cYNFBDeUvpWo":[96,68],"JAeDQsQhiXVEByzT4eCrXJQGgttf12Gnk2bPYMUt3eBC":[28,18],"JBH2sxUUGfXSDaxrm1Dsh1cLpMJWiDzuyNkDWne4qRHV":[60,39],"JBVWoq5pDYFaQqpP5UEztm54GZVQzrQpWKnWWn56UK6b":[68,60],"JBtk3KGDoQXYedMDiUFDp6VXJr8MxY9tqXg6pZ8EkaT8":[64,39],"JCtrjaF9bagyryB5vC48yY7rMjPyLWEoDMGjkhFqjxBm":[64,46],"JDSSd8fRsY3skSKEE6KbLLz5SxURf7nDkf2caazkMP4G":[40,33],"JgcWmNdwrrmvJiuSo4apJCLx9MozajzKRTF5mQREiXH":[44,33],"JoeKdMCnk9rE2DkLnej7tw5repqdfCDetSLCUedhUVn":[40,38],"JpaDR41DYXTH6FksaSg8tgoXbJxstUb2b5m26sNQx55":[36,27],"KK56APewFFFgM7c2ehKY1Eky6DgdDpmaBH9521roatW":[32,27],"KdNhBD4WCm4Gd1fwi7Uf7Z3JD9KrZcnWWm8nSEZ6NEB":[72,54],"KhBHw8r7pwMqurHv7N8pc4T5CHUaKiFnXKzFdCmkQZV":[44,28],"L9hQqzWE7yv3uB8xT2TpKsB1BqukPg1yYN9e1ygAbM9":[36,32],"LPGATvYWNFLrtAjyfk3Hyfdngp24MFX7S41T238rBDu":[48,32],"M7Pcv3j8KpX8ZAkeSsvJnexgKrZbBAaMEcRTvf6t2Em":[72,56],"NHtR8X7dmwtCagm1FuuC6ngQ3wv52uJYqFvA79G47MX":[64,42],"NNetet8BiymZxMBWLRPCcNGcBPZDBeEcpgtfTSwdFPX":[136,67],"Qfp5wD5TwLiecZZP64cn3SwfqvF7W4fzo2tjEy2c1MR":[88,63],"QxgXRHsBe4vazdrboKaRtXMuf1EhTVvF68tVcfepRr1":[48,40],"RAbGPTmaLVn1HkP38tqKYjGceejMWiyfDPEhenjH1Aw":[40,37],"RmQNKxg8et4ovHsa6Cs6GvwvPb8W1wwspUnyyFw9gs5":[48,39],"SQJYmcjgo1bwJe2YxJwRDAH1JKFdrQM4AfLzuLi5TME":[44,34],"ScN8WkfK7c5nmNvNh7SbFTQcNyw5poXv97h5KRFBRWL":[44,30],"T6CVCqL6Mcea4wjRgvoZRkDSexzNP6fuNcyEdQZH21H":[56,56],"TS7mNjzAkgqKSR3PhSXhrqjVCkyAmUta3skRF1NbRUx":[24,16],"TxChgiaHwnkdT18sBnSepLE5sGk7vsQ4CZnhwiHUMQw":[88,74],"URnkWZGiuB7jXbfCSuNSwir1qkn7sXjiSPeLPaXys7b":[40,34],"UdAZ7oz1WshdwyimF6e2VXiy1eSJ6UdHSRng9yRLtgY":[48,36],"UkQCtmg2gSygRMRJq3wHT8fZahqYzwRp2rHXE2hCTX1":[32,24],"VCRrRTgSjDLHvo6UQKXy8VQbNVG2ioHNUEyS7oB7u3X":[68,54],"VTAZqz5HadKsUWyavErx3hhUeaDPerPVDssjB69hP8b":[8,0],"VisixkGG8H2htLvq2EKiywbHXVjNPQiszUKHeu9r6bd":[56,53],"WhXimQKBLiMUAWBaVVenpweVJNSahrFZdPLd8hr5Tfq":[48,45],"XctiDwBwYbZomH3pfBKca69BKrxjEQtFbdu6TXw9fe3":[48,35],"YKhfczqyMeHPMSJzcs8JiAVCtu3iieampLr3j63yTjk":[56,50],"YYYYW8eKkmwQFhVGUKdBAnDQPuhMTpG7zwm9nikNndC":[4,0],"YpGaosZwUmt5p8gXbdGriya7zKZvU6439CDXQSS5Gcb":[52,52],"YpopmpJ5ryYnLZKD7a2dEbPdPiiSLRARWVj3oAmgWLt":[48,41],"Z36ZMwALNp9sLgBt6nUhb3NAFAJbUp9kCDM3an7Xhb9":[40,0],"ZDCJDkoBMTXpf8zsfQzbLeTfAus1qaxiFHnANseQrmA":[4,4],"aTPi9R1prHHwRf9GCSgK4yLZiCZLs7evMtZP7EPNKw4":[40,23],"bj53eWLx461E3m27qmHGtJE4NZxefhvZUoewioSavqH":[32,32],"cDK4eZakyrZPT4fdhpPUt8q4EekNEwwGEz84LFnQb2S":[56,41],"dyEBiLy8Tty9nMeV6aUYU6qHVobsm4YNQmTGRhyvUB5":[48,18],"eMVkN9G77VnE8QsLtMvArMcsM93cytxowJxbbwwmzWV":[48,38],"eoKpUABi59aT4rR9HGS3LcMecfut9x7zJyodWWP43YQ":[4,4],"eopuRXqXh8HxG5Y7U7oGCN8CpzPudLx3j1CWz9WBDGR":[4,0],"f4fN8n4zEtVnw2fdQD8y8GBwzN5Asc4X6FzuFi1kEdt":[76,53],"fRiGutrC4h4ZdYVE65g3pCeJNDg2g9j21AvMjhDMwW8":[48,37],"fw7Td8AG8Sfha5ZQQnwaKAYFnskbbLBPJrieP4Puxwy":[52,45],"g1jyzFsCaJZmfkCWxcAF8Ay9DFm82cSrM4yfi8aDkgt":[36,29],"gVuUFY5MPjhayfUVTKTuBidEwFxYbQ6sgFMsSDNpRiU":[8,4],"hQBS6cu8RHkXcCzE6N8mQxhgrtbNy4kivoRjTMzF2cA":[28,21],"hTzYQHBSeqW7gjTDsZhUTe3NGGJjTJgwF9sTikSuwiY":[56,51],"iMZEU6pbrLTCHZsaw6HkwABYvMTfn9pND9AqQgaEyuo":[80,62],"ibwMFhkkeMTn9746FERTdb7rGuQwVcRXDbNYXB4QB8q":[76,59],"irmVepLsWGTFjCrqGpkkfjYexrvCgwJ4CTSUQVjsFs7":[60,47],"iwf4VCAm2WYfHMWkGC6p7PScW5vihfVrXs8UTnMuHuc":[44,35],"jw3RWGkKoRDRnvscvmXHPkATtay7oeoTBMvQinprcyc":[40,34],"kffvkDohANNa2rpj8Ti6KWZctCX3Ci6Rj1SnGHx2r63":[32,26],"mETnAkTMdDN41d9wSPYJWDFu7xehfoHyT5py2thcxHB":[52,52],"mFJG277eG7EFS7Zu2UU5BkFZQW7PpAVfjMaFsTqXAUq":[60,47],"markiLNTC3FuWYGXKz8h9XpbwJbVQhzuV5U3bfpPc64":[68,58],"oQcqLhbpxQdDPgFP2d2HBNWqJFKGk9JfHw6kxyLgfK2":[44,41],"rAEVgLDieWcb3N975fNsLbmpNenL2simANdvk35iLeh":[44,35],"tNZMLNhcUGk8pr5LVk2R5eV7jxiuU7SfDNa6uNbHeTF":[72,65],"uknL1QAVVNLT6FEDxc21yN9Q8jykDCJDvCyS1f1qUkJ":[40,30],"vWoxJrxUz4JjxXiRKoMKZ1BfkDFcDcDVB55Lx4isemS":[36,35],"vav8fy4UyYKf91g9uFZybjwZh1VS6hubfaKyFtbYcvT":[48,29],"x31Ldohp254zduxhHMHNR3JbXZYrxgjEaWTvd4vuxZ6":[72,46],"xfCpo4ouRs5BP3WY5BdWhbr41pQxYGcXxz1sFyzPsZr":[36,19],"ygkCkgip2MbboVLSnq6FpEz7N89nqiaTCKBCSexamR8":[52,42],"zKuryCTzgvwoyDZTTh4NuiT9D9bpMHG33tTRyKKZUUT":[60,60],"zjG7sHeExhC7tfLZwTJwHH3zzDSqDcwRVe2LdXg389j":[44,40]},"range":{"firstSlot":83900256,"lastSlot":83992896}}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetBlockProduction(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBlockProduction",
			"params":  []any{},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetBlockProductionWithOpts(t *testing.T) {
	responseBody := `{"context":{"slot":83992896},"value":{"byIdentity":{"121cur1YFVPZSoKQGNyjNr9sZZRa3eX2bSuYjXHtKD6":[44,38],"123vij84ecQEKUvQ7gYMKxKwKF6PbYSzCzzURYA4xULY":[52,49],"12QYHqRxPuTPfkBVLetEuGkLGHD9GhqM5coP67xK7wfG":[64,55],"12Y25eHzGPaK5R5DjQ1kgJWuVd7zrtQ7cmaMPfmacsJV":[40,29],"12uDsSSWyPRGNK3HqcLBRNZiNFJWXCcHvHD7V4RYsKMr":[60,54],"132GXL3pzAjyEKoYLBS3QDTWDLqPnLJdZNcK4cWNDrmb":[64,43],"13nbrL1VjkfTZuaz7rNriYw6fWDFggqEf2g1C4mPETkr":[36,20],"14Z57kkY62p2UZyqeQyoGsXfkKbguAF1G8g2kZdV7Vae":[16,0],"1B4UocjePKwr58Jw4sLBsBHFt9nXGWxi1QDv9g73mrs":[48,36],"1DdfupGTrYUKtRxN9ukGCf3HBquc4buGPPmZr9WEjY4":[36,16],"21Ew2QbeiXprspa96d76RgueZ6HvrQMDTFAHpa71hpoR":[68,53],"234u57PuEif5LkTBwS7rHzu1XF5VWg79ddLLDkYBh44Q":[4,0],"238Fmy2TwU26Fo8XFRu2PzDWNbcn3bitywEPYG6tpztu":[44,36],"23eke7qW4tibp13JfiHKLErVsdmLDTwVsqg52bVQwBCZ":[32,23],"23xMZ9ijUgM9mRVB8sk7ZUR9yCRdP9eWU4ohjgbQbhGV":[64,45],"25UM59KCvciYwhjCq7t1rC8ZuvsxQBC2QRcaRNfq7xML":[56,50],"27JQHoxJi8kpJzwED2eTv8jDPB81gjqNFYBNN6rjM3UM":[44,43],"28LgQ7MeEZVgNJfYRc6UnoAz2SnSjKbyCKM6sntCRotb":[64,49],"28aB6dFf5TPKz3ghnYnu7nNaLsinoAE4xNyid1sy9j9e":[52,49],"295DP6WSsiJ3oLNPhZ3oSZUA6gwbGe7KDoHWRnRBZAHu":[76,59],"29Xwdi4HpBr1u9EAqDz3tBbMuwqBuczLPuVe2gGkg7ZF":[44,33],"29mA6zhspyms8m17FX8ztzz5UU9Fdqbumk1vxEGUkC7H":[52,38],"29tXWWzvGNvE5j8i6FLfHpmanPC9treZsCo1uA4ik1kL":[56,43],"2AY3bKHAMkdj4cCn1UcWCjewrg3ccDnhVmvJ3WrmmkAL":[36,33],"2BT25HZHpyzYmTbqyqzxBK7YjjH4a6aZ783TEgTTGYo5":[48,42],"2C26iHJcU5dqJJQ6NME3Lq583RT1Js9QDtgfmzknRajc":[52,43],"2C9pDcbRQJxbUHivgDdg4LGuMwm5oeVCnHS9w5JktNTo":[68,49],"2CGskjnksG9YwAFMJkPDwsKx1iRAXJSzfpAxyoWzGj6M":[60,36],"2Ceu5z672tACCDB6cmB3n4tih7B8dikv6k6kfYr3Li1e":[44,36],"2CjEU72sCTy1D6GyvpAjtKVGz94jdz8geN2JNWJCzZ6q":[56,0],"2D1oCLRK6geGhV5RyZ52JD9Qzqt311AEH1XrTjZdzbRh":[48,38],"2DvsPbbKrBaJm7SbdVvRjZL1NGCU3MwciGCoCw42fTMu":[4,0],"2Ebt7yP857s1WfdoqNm9FsGeahCvjdXcqhvsVjzNgUfx":[60,38],"2F5vGa1L5f1kKKwfcQvWGQCkJ7aoAxDRg4mZmxq9Ti3i":[40,34],"2FCxeG7mBYy2kpYgvgTZLCYXnaqDc4hGwYunhzHmEUD1":[68,46],"2FTDGeDUAXJokjVSjRXX2WoTc4tW2uMagYB5jk4JCJAK":[36,31],"2GAdxV8QafdRnkTwy9AuX8HvVcNME6JqK2yANaDunhXp":[40,32],"2H6AvmuhZ2yWSN8K8CQTPcAfVaGM63cr3oUeVSw6pUhT":[40,34],"2JPBDCGefojYLyy87VfJkHUVjMhD4H49KPgdCkitRwTi":[52,19],"2KFrkqEeSBKEHiMjUugPxTkBJ2jXepgFBqHu5ZFxtaFg":[56,38],"2LsRRgttA1PKXXeKTZP2QhetgM94Dj5uecmTzyQkTvXK":[32,13],"2MKNRRH59tZXPUas1UcozqZtAHmXJBzpvNHUGySQUaw4":[64,46],"2P3YH9psWAAM6QQgA8NaQnKHQ973cKNqTSFFCNYE4gjk":[64,56],"2PDvmDx6HeKv3wtdwmGQGmz9pGXXDKNFVvGizGGaAqxL":[56,38],"2Pik6jn6yLQVi8jmwvZCibTygPWvhh3pXoGJrGT3eVGf":[44,41],"2PvsR9DM2GZavFQGDsdJwXJvPWsyneyT9Gpu7wXGDkSr":[52,38],"2Pvzm7bGYjCpfjj8iyF724eesh12PejtKgxzv53ctgXk":[4,4],"2RLf3RSy1ScBFL5UzVDw3jYKCAuGA9vHpr9dnbQzJt8V":[52,51],"2RNHZTsFQF7BwgTrrwH4qvibWeguU67BKnPAxysaLUVi":[52,51],"2RpZdDc9ss5VsVUgHox2e6u6yV1SKUrQV6iuzvggdLKK":[48,47],"2TEGxhx2CgHw5fpkrvJWBsRKbJNAT3y6Fco4Lj5DsJjk":[52,39],"2TGWTnUbjfZqvFYTgwUTdA3rXLHshRbeDgVMEMg7icZy":[24,23],"2TcLzmdcpQqbNEBjU2Hxp1MiETSww1HJeWW84F61Z13k":[56,53],"2UCNzcnSVGtkgzpm1guxR6hEDQ5A8gVDkwVTWfZ31bPg":[36,31],"2UJ4q96QBg8dQom8JFgCMASdJnngrxhafgjY3XA5ddso":[40,31],"2Uv6KbG9Smt8PVMiPVGcRzt2GMvUvnvydSsgVRUZZZdS":[56,43],"2VEnBfmR1LW44oemZK29MtBHxTuprLizjPcLycNyGkTt":[56,33],"2VL9dMnHPJG6sD9CuB1BkPCpB6S5EMnvXBECHHnKqz3z":[32,32],"2VzCLy98rzmvKGo23e1eM4LANCt9JFrtVTBBMZzGT4FW":[56,48],"2X2APoUmQcbyVfVNCmivzYRkydxZkfVdXXaSCHQTa8mC":[76,66],"2X5JSTLN9m2wm3ejCxfWRNMieuC2VMtaMWSoqLPbC4Pq":[28,21],"2XP9MfzWQnX3DiAJQFSKXKyqBr4M7GKhHnKc9P7z519H":[4,0],"2YeoCYp1KT5W6S8MEVbu1omSrHVtZPEVcpFFKRdXwfAK":[8,7],"2YgttBBx9Ax6WhF9xu8CSWw5usKFWxaxzJtg8zPajadx":[4,4],"2YtaVYd8fZHXpPezG5gRp9nudXGBEPcqFtJe8ikN1DZ7":[44,38],"2YvDq2K7zBvsqZFVqTGVcKqJqSLZaJH1hny6fah14mqt":[16,11],"2ZETk6Sy3wdrbTRRCFa6u1gzNjg59B5yFJwjiACC6Evc":[56,36],"2ZQbsxdEab52wEFELQnn2wsN4LRhsPrjyCeqfDdtD2f1":[56,29],"2ZZkgKcBfp4tW8qCLj2yjxRYh9CuvEVJWb6e2KKS91Mj":[60,52],"2ZqaTLm1TZfpYdR7d8XhRntPjmrF6q69YZVLb6j4GcWz":[64,38],"2anGa2owPRuQyHyEaWSbrrWws6NiyoakByaXEufuU3hH":[72,51],"2bQyrSEPaQ9BMbu7Ftv7ye1fxtSLW3oZRj4d2U64AJmc":[52,40],"2bZXLja7MqWiTtDfxTm78rvxaxfa34RhF4msQmvCBAWn":[52,37],"2cgXfdfA4EcJjouu5jxruaCMPyc5q3oe4qRMB14EGWyL":[76,67],"2cxgydEWqiVTookrgPWucNuQqmwThyoNAYPX3GqeDdvh":[52,42],"2dm9YbgXtR5yimmgsLkfaMLcNZxhjywW4bLnvChms3tb":[56,38],"2eA3YU5GVKRdFKREMMNmaLjptjvBLrQQZtuRDM8hZWde":[28,23],"2eDDjJSKdxf8qwojH1E2SoZFHqst56GXxtmAnoZtGdtu":[48,35],"2eoKP1tzZkkXWexUY7XHLSNbo9DbuFGBssfhp8zCcdwH":[12,0],"2fCrJDUrArXita8dQDruPjBLMXTWKuMdbQVnVRGZYUtb":[44,32],"2gHxXKYyVCGrTjuFNVi9gTPUF9xN4hRhxZhcHpscS2dQ":[60,35],"2gV5onEfn8KmtZ3Lck39GrNEZyTxJ1RiNV5s7fRdC3gc":[56,42],"2hJpiwrWXbpRaxFdQWnhYHR19bbWfUd5VqNk7hhjALCx":[68,0],"2i5Ms2WfHHWz4P9Bdn2x8ZtUQ3fUoR7GL8d6UjhZodCe":[48,35],"2i5vGJ4RQKFJM8vxbvFaCmPKLprMuKrhVcJZaYXoTmHg":[76,50],"2iAP1WMsKJVje22cgPJNGC7Jgv5DQj37QZwcDNUDd9F3":[52,44],"2ibbdJtxwzzzhK3zc7FR3cfea2ATHwCJ8ybcG7WzKtBd":[60,52],"2iczkZceGZQqimksY8uk6NLrQXoMFZGK1mTWos4QnZ3a":[92,85],"2jypS1SoX6MLEfuNvUH23K7UU3BsRu3vBphcd7BVkEpj":[68,20],"2khFqurxeMKKfhFJ9dfas1L9LsHwt2qHGW8Ztinzoeob":[32,29],"2mXbRwZihk1TfeyS8aQTTm3Lcg2QHz7SgVP1xPavNYho":[56,40],"2n9qy8LiuiNpaeKFw82AaBWfi5F6CFW1rPH1ZXNNSHvo":[64,53],"2ofEZBxkiZoBpxXcXT68RTHfuQQFChSYVXVPGbFfvMTP":[40,30],"2p64GWwGEWtHdwjdeXCMHU5LBstm5BenLdGsJZzDrKHH":[60,48],"2q8YQuJZhoAZkQNaZgQiP58tvr5HE3sfQp5beG9BzerR":[44,44],"2qcfDTvKHzp2RM8ZsAqogZinajULNnDa5a7yQFd3FrPs":[60,46],"2rsXVaKikXHsCuFyYEkoReVZEv4LZBoBBWG3wkNCSWK2":[60,28],"2sAvsH3WPrHJ79P2cbM1RBwuNea9aL8z1Q3u9buFw99g":[48,41],"2sBsdFT58SPfd5LQyE8MhEgJWpaoHUCoN4QFCVqNZpnj":[40,30],"2sCfpKiU1JhtHvveo7SL8mRYD3sPMmHaAMCC69CJ5cwG":[56,39],"2t4ED5vy44LXqRRPsesVdTUE4ScVGL3vamBUx1hi53dQ":[44,18],"2tStw7K6ApvwkgGzZxkLQ263UL76wpeNYgYadTvZe8Vc":[44,39],"2tjfEp1WfgX6n85U9e17aBifn1vNyLNr7jmwPf7SAiSy":[24,8],"2uyMGFNvMwC9tjjKXRZcGFSztnzxxkk2Gbkki1V46L76":[48,42],"2uzuT8dgVVLLgG57n5W9vxMTaNAaLavnt8gtiF7V1FVV":[64,49],"2wAbzXXERjCoRrgEH3sdYfgAaJArrrraXkDzEnDPsuMi":[32,12],"2wqeRoCVwNoDqjXuZJ5iDnP5NR5HsqfU3Jf3DR5oPYMt":[56,0],"2xFjhfxTKGVvGDXLwroqGiKNEF3KCSFaCRVLHfpsiPgd":[56,36],"2xoWe7LGX8Kmnwoc27VF2iYkxfKESjb3b9rU1iM9wHJT":[36,23],"2y857Ss2GgyL9WooNqt6sAgxDVwr9pE6i4BiJ2wrC4g3":[44,28],"2yDwZer11v2TTj86WeHzRDpE4HJVbyJ3fJ8H4AkUtWTc":[68,55],"2yLVWajWK99EZyxMquyJfnMTdUyNG62nfVipVY27ELun":[80,57],"2zHkPFBSxWF4Bc6P7XHaZMJLfBqtSgfDCBqTZ7STXE1a":[60,58],"33FtaV5DrLUPYYQK7QAiD3LBDXD2VoCv6BuCQVoRdq57":[56,43],"33LfdA2yKS6m7E8pSanrKTKYMhpYHEGaSWtNNB5s7xnm":[52,41],"33WStcUFh8kboGRCW2ZiFhhNi9AcSTzeasNsUV1Wqgut":[56,50],"34D4nS1eywoA1wiwcgrBP8Ewj9NXyaZ3dP9DJKfkvpGn":[12,12],"35A7K7f9Nk3YJLLoPvqxLW5ngRvX8fRBJSHhmowWmaSu":[44,40],"35FdizoSUbjCybXeogAZWoXKbnCzvWFx9pgMLohUCRh8":[48,47],"36GzimUeoiBaapYaC1yriTJ9moQK1QvJfexppcZv3PaN":[60,49],"36UkVHPN89MszKNYA7ywFQqyQ4FGKGkpWLZjiBZiDdLr":[64,59],"36gARMU4V3D6hu5EJi7wYFW6cC1tNym9DkjZfAGFQTbk":[44,36],"376e8QLx9qSkjFn7mK2kp3wBwvziKuMqiB3iAbK5Payx":[56,47],"37Gr1zVPr79E3AdPFj8EMyKZYt7Bnz3VWKjdFctQC8fL":[72,63],"37PWrAzWfgn2yptyHBZQ4HDBJu1V4BvhiK4vs81MQzo4":[8,4],"389UtZkyvUzHzQUt2eSEzeTiC32GF5PwsBq4uNZz3cpY":[56,36],"38hgERMK335yrDsyPkc4wbW2FUiXgmuWRght9n7RVAtz":[32,31],"39FH4cnkSawRtr9N2VbUVST4o6ZiixW2K4QCzLqW8tMg":[52,49],"3ANJb42D3pkVtntgT6VtW2cD3icGVyoHi2NGwtXYHQAs":[84,52],"3Ax8WYVrp7prVWixBzNcKzwQPXFZABnaupna1diZQfMK":[32,26],"3BCokPfahX9rLYMh6E6uYTEFuchiKd9wZcXUvwDHFYiH":[40,34],"3DYMXn1LtpPYChqUQFq27oMqnSidjcYzJQAt85jDUowr":[64,57],"3DeoXyzWHzc1puYcNZ8khMRFAXwL7JKEbvkXuUWdNsea":[8,6],"3Df9iVRoSkX3YZ6GmMexeLSX8vyt8BADpHn571KfQSWa":[48,28],"3FDRLuYj1dHoxiVQNbj8Dk16gGv4gmB1fmdVmu66WqPw":[64,41],"3FhfNWGiqDsy4xAXiS74WUb5GLfK7FVnn6kxt3CYLgvr":[72,51],"3Fiu2KFBf3BoT9REvsFbpb7L1vTSs7jnmuDrk4vZ9DNE":[68,58],"3GctwRAZHTAwxx78mU5unwtxSEiDNsF5MgoY1oXXHm6w":[52,45],"3H7BtRE7iGC9Kzxq5eb3Hx3hChNptepF1KRrEufKjNMD":[4,0],"3HM7uGuE3AD9smYFL8uKinAHo4GGtX2PErn66FhGp5mc":[44,31],"3HitRjngqhAgVuNdFwtR1Lp5tQavbJri8MvKUq5Jpw1N":[40,38],"3J2GJs7nWTiF5EcvcFeNYydpZzbL4NjZJetHyKMpxFnE":[52,37],"3JfoYf6wmQxhpry1L61dnDWYJbL7GYi4yt7mybehuhne":[36,25],"3K8BYGTPD9AxqYQDPdU8PPy6AfiSwf4hDmFy1xXGB8Ns":[24,24],"3KVuW6mGLD9Kv1Gjj1A5q685JZLp9hqE29Kbvnrii8gM":[64,52],"3LWv8RrdEyMtePAMCmohBzWAz7fmN7Cf2ctSUxJKEQnS":[64,58],"3LsQLm1hy8Rcxw2neKgFqcJzyXNugJtRxXRpMvSvzWCU":[72,57],"3LtAt3iqmeTgJ3GD8DtCcjkRkJdDKAF42nJytn28syeP":[56,42],"3MdUXXiLWeXQauVSiuGwPjakCv8J5CX5v1fu8eutJ7v1":[40,39],"3NMFamQ5RtVEs5N6KeUnGnwkaoukkp4hduzUPKJr5Y8t":[40,31],"3NchsxHzVUAv6MTGEuAVt8QRdi93uHGNRmS9AEiZkMVh":[64,35],"3NtGCPqA5dTucxitLz5KTxERZ7XdVSZ8c2m97TGupV3S":[56,51],"3PVz8crz85wgqgudf6mxws2psgKc4kr51MhfmU6VekEG":[52,49],"3Pog3tY91JZRv8irJf9sE4JKPn1pWBj9bLB9NHxHgehu":[24,11],"3QK8tbsVSwU6xRzLWhVFJCcnqm9WPxSUdaa7cXzBQZZh":[32,20],"3Qj4rFsMRMsXnYescUVi53kDY4KjNnNy2QE4tc4WpQET":[44,38],"3R82jDjQsrzZgQKiEJbKfdCA9ngYQrjZehYuEFmhhfCP":[96,84],"3SYNAWbZuuWMYKHwL73Gpeo1ySqRSshf5WDoW9vYVz9V":[72,62],"3Teu85ACyEKeqaoFt7ZTfGw256kdYGCcJXkMA5AbMfp4":[72,66],"3UtHK2ZWwmDKxd6QzrKmh9Pey1gWS2SW1MoaZuGZbc7E":[12,10],"3V1cpkuKhJhQuKB3BJvTeezGEDKR78krcJT2esG9Wte6":[32,23],"3W4fe5WTAS4iPzBhjGP8a1LHBTx8vbscqThXT1THqEGC":[52,48],"3WrJpBnPGbmFx8jPpseT9gA5LAtubovpaBjE9waUe7GV":[60,51],"3Wyqj2cgKYK2sSSb3wVv3wJ5yD3yigV8iLLttkZfKn8d":[64,59],"3X6FsQ8awkcU4iXTF82T4RtnTJx9LTY5D3dHK6zDE1Tp":[32,27],"3XE9NQAN6yiKvNedR5msnvdLN6HEdUfqyn4yWoEYYEcW":[4,0],"3Ye1g9E65wj9wtbTLetQbjsQ6SFj4s7RdTJaxjq6duDq":[44,34],"3ZwnSWgQBpphsSzZNA2A2uFMuXZyJKDHi1EHKjbd4ikw":[52,43],"3aN4DVNJFGcHsKHrueVNogCnYFcGQQXe979zXvBLxDhe":[60,44],"3aSjivWpfjcSyqLTf3fAuJfB2R1vkYxDsnNDmByaXQp9":[36,29],"3adqz1JN9sbsjHGxQizz2ibJmyCHtUpP9aPnZYxixB4c":[32,13],"3bQ4s7ynWKjEPrkTfDx1aT2sXejXXYjbfYumBHc5LA83":[52,48],"3cJeH1TCZcNf5gCZnSbfZne9DQCiexkzuH6gwQEeBjqA":[36,27],"3ccsSn54FkE2zYU7ELDyEhB9vQbJ3Fz1wBzuuAHB3KXj":[76,66],"3d8F4eCR9YwvXdmvVzLQ7hHLTBcHAaWpC7jQxcdoBHkk":[56,44],"3dFGZmTgBwsBfYqvJqKdwEQGa7HUdPsSbHmTkm4RjJef":[44,39],"3eLGe4vyZoNK3FvY6C4oxQ3cJMzmUerdVaaTsfN14Ngf":[52,28],"3erfNXKZP8vmLoc5mXzdwhXa8UhjSigAiiL7CczdP2LU":[4,4],"3g7c1Mufk7Bi8Kk4wKTGQ3eLwfHYqc7ySpP46fqEMsEC":[12,8],"3gmNSSQVewvEiBY9Vh4hmekPmtGTiKPvooBW7MSkUXPc":[64,54],"3gnRETF8Tnto3ZCh9yCw5sbqrq8zqgx32zH4y3dzEN7i":[48,33],"3hcdjAmggZJxjaRgMfxqhyCe3Uu2yshLRZPZ3mXAkst3":[56,0],"3ht1z7tMieDiLkukray7AauF214xtsWFFG1E4A1oeAXU":[68,60],"3i7sS5McrJ7EzU8nbdA5rcXT9kNiSxLxhwyfuxbsDvBj":[44,32],"3iPu9xQ3mCFmqME9ZajuZbFHjwagAxhgfTxnc4pWbEBC":[820,641],"3j7SjhZK4P7LrKRwAq5vNKerENDkWJP7xgxJmqsuj2jv":[40,31],"3jjwWrta8PF3paARTXMpKmF8wxDzUWRCcdvRCdscvbr5":[68,56],"3kWT2K2HfxrspLFoJhKUAio3QF85EuTemJKTUcPEjm7m":[60,48],"3kcjg9J1d47mgZkGuqitHd7K6Bz7XBVMpZRB1Sg5dKdN":[64,35],"3kiAniQf6y9ZT3SdE8X7Rq5jM3MX6BUZy5KDT3wt6zAk":[40,32],"3kkVsVzrxiJdYzrFuGwM4PvuHSmVtVFjzEMBKbgjMWvp":[40,33],"3mDhRnsnQdmRyLKx61i3gy2PeNGM2zQxgofPFSDdDLZo":[60,56],"3mx22d1aJLazEutJyHVszdwyLJcrRo26EKB4AWDbRxRc":[68,57],"3nL1oAkcW4M88VG4D78dNxHrqaNdKyJqKW3wbhhBjhig":[48,38],"3ndqwmmqTEFaydt6bgTDohL35WJCjv2cezUcYezcHHcJ":[48,39],"3nvAV4PVG2w1F9GDh3YMnhYNvEEzV3LRMJ5e6bMYcULk":[56,49],"3o3GJr9iAdJ2v2sZhRqiX5nGFJyrpdG7t1jePatMfFkn":[48,39],"3qaaXFYh389e1Ncboc7qbCWxSQdbaiYuTFrJVYuh7jo2":[52,40],"3rFxX6D68YhDpF7c6vDt2yhfp8CXXcjNNga43cCJ8Ww9":[28,26],"3sWB3AMv6Rd96cKTgtZBPCKoxDW74eGWcqKPkhHEzF1K":[64,57],"3si1tYjwb32Mj43LWw87Zy4acgtnxXMYeZuKHmKEYB8B":[36,24],"3si45SHHXsP8C6PVo1Zcpcry7DuivvogscAA63D8AKmR":[44,33],"3tEqZrbb7xwaRwri19Z5TAznrewnM2m2SCkvSmLztWcE":[52,46],"3tHeSnJt7dMSxyFGg2LW7GBXCWMg8KxYQQCKfnx7cFs4":[52,0],"3tSsxpkuuZjTBG9whoPU37kS3NFK48Morvm8ui2vBJLm":[40,35],"3tjCLs5cMKiTgArVujb3S5LhQbBhDppCu5yXd5eysSs7":[4,0],"3v3KN1rtwURN3NVbLJceVSY5zjb7SX9PSXDU8Qgwf9XJ":[64,46],"3viEMMqkPRBiAKXB3Y7yH5GbzqtRn3NmnLPi8JsZmLQw":[20,20],"3vkFbUsjMqkkgNvywvxTPbsGF18NMkwBX5cBeBsrhTRk":[8,0],"3vkog7Kaki74rn7JFWxKyrWfTEUnp4cLpJyvgs233MyM":[52,43],"3w6hQh7Ndx93eqbaEMLyR3BwqtRxT2XVumavvU93mcRk":[84,68],"3wao1rTFLniFiX6vofFyEdyKPuo7coZETkfEQxf1s7mS":[4,4],"3wwYJDVkY1rK5emynSYgbwUy9X3eFcNQiyYxc4Jsd9iL":[48,35],"3wz211BhQAE2n5fjDQSStM2iSizhNRyJDNRkDEc1YwMF":[52,40],"3xKsqGgLMNVazzNBsKa9TPG2Vo5fGLr1xkKrTMVXVVkT":[36,27],"3xUTkgPKNJZ3dkpDMV8zWV34BkmvKanguKipv6M9x2Mt":[44,32],"3xgtKbSXjtZe7hqxHbK2WLYJGPJw1hfvZKzHrTkygiZX":[40,39],"3yEpFZ8Vq3vbbfvLu4r6vkRnV7P2QS6FSqGNuMUXro8J":[56,52],"3zPHhqJE3AR2S7WxJf8YHoVZ6mxPNhvvdjfddRiNY99g":[72,50],"41nRdNqtbMp6xGQYucjjydvRQKjiRyxiqzDHjdaqMxCQ":[44,35],"42DeSPAaef333ZsSzBGADHhAeWTY68t8CTMzJ89Z6s2r":[36,27],"43ZCLRdQgcajUq4WTxtTqkqGtpNnJTmLUs4ef4qGKtAc":[76,61],"43h2uYRTSVhMNXKuxY4Kn6T558u436qy59cV6Sz6rdRi":[60,50],"44J72PpPim1PJHge3TwJWAMnuPhwE7DMLaZmCerYEC61":[60,57],"44Kuawvm2ngsSyvqsMLCTeWXUYxoedthgKAEL3BxCdXP":[44,34],"44yQJPhbBRV6povRiXDc3KkE7SPXUohF9ipqBvCjokhc":[44,42],"45M6om8quE2DnLh3cYnty8kx1D4AYbMUzZMpytku6Gff":[8,0],"45THWNjLaWBh8jbuP3HrcG4iUvenHpSHmFVGnkmuQH4U":[48,35],"45YDFXgHCEbeDs17Amrd851M4gxCSJH3uofCsvdKLhRJ":[56,32],"45aGtJWVx9xbhp11diPithdQS1E9Hzjm5b5HEpAM68Ax":[44,37],"462x4mp5aZ29SetJR3oka3d2ARXVKUcs9f9hZsapf7ML":[52,36],"46GijDorcsduUvWFNWKAV1yB6XwPG699wS2gR4no4zGU":[52,43],"46WCeEExQaEJfatG53qgxMzgPqubbrAvVBeYSyUQt317":[68,50],"46uT7tSZ1US9bJ93ByBxSBCmZhQogn15Pwuxp4fhWXqc":[44,37],"473ToSs8wTyGd2DTmwb1zNkr7TweNC1Wfui2FzKNB1JE":[52,41],"47JuXYUK2UvwBPxq8p4ePvDggkpz49xmw93N3VNGbDm9":[16,12],"486kJEz1XJ95nULg2Ccj9Av9yi1inexzHRVW9UjfR2B6":[68,55],"4958nAd4Gp1MZQEg97b7prdDKAgC5Ab3iQtNzAWyHqEV":[60,43],"49AqLYbpJYc2DrzGUAH1fhWJy62yxBxpLEkfJwjKy2jr":[20,15],"49JYKwBGHPsL5ji9LSzDS6WNNvs2AW2seC2qZDiMWkPk":[48,41],"49Q14TEnx7XTHsFtRs9xhQ12wXRHwaWJ5YSpGhVNhSgy":[68,39],"49YDWPPRQRatsNgUHLbPytGtKgEetBFsq58uGobM8sDz":[48,37],"49gM7gXEJEokKHEoUCNve3uCRMAoRwKUpEiqK2nku6C2":[76,60],"49oW1EjrYFvWJLUK82mhcDqN3hWir2LT8H2Sectvfmr6":[44,29],"4A7XYUpU2Cvj84fBhkcUQPQMJsZywqgjvD65zSRZmquP":[8,0],"4AYWAYndF6EsfgwVTrsHLMviNsvuqh9dAMcJynpJk6YB":[40,31],"4BVYRKYnwWbUYRtxHSNnue8xydUhexegZKohbbtkT7nv":[4,0],"4Bx5bzjmPrU1g74AHfYpTMXvspBt8GnvZVQW3ba9z4Af":[56,32],"4Bzp9fzcdjctbdo23SCwCEkPeQzCeyTb3WtwiK3KNVRc":[32,23],"4CVJ8FMombpnrE7C1a4mdwMMbhJDroAzjG51BuifPmcF":[4,4],"4Cvq4GbYn7jWPpUmdcMSL2tPBV5E6GqAHfFV3u1iG9Zv":[76,51],"4DPoKwdKKWCYMcjSfWVeo3G9dVvcVgc487HRNdAVNMfQ":[44,36],"4EACGRv7miQa6wSw5ymGiV3dnHVZwrsKQoe3aMDbjdEn":[40,36],"4ECpcT3wLE4EzBZ8Th3da1EaALL3pwuKn2jL9rGB1MUH":[84,48],"4Efdqh6SnwMiAcu8fPb7gDo7Eu4vrxMQgMdFb2JtwNLq":[56,32],"4GBSypESidsbB6ACFRUTkwDwcv1G5anashx6UvSypqCF":[32,32],"4Ge8T8WeH1fnv5SijRzPfC38jWnuBhiKe8iE9fsXbqLi":[72,67],"4GhLBaxr1oEHWpoGnWh3mcRXUkBU5EEQZv3L27c7ohoq":[8,4],"4GzmbxmepoggVLYzyXyM2GzYVaisJSuutsxrydoErSeu":[60,58],"4HjA5dBRcMajmaYfwYxqdJBzYbuFxPqjoVjnsTk6Xjqv":[60,58],"4JZsGW4WUSjAjH4joCaAAVnNi5ERfHr93YUDxmHZpDM7":[60,59],"4Jb1YfUUN1xxdYb28wPLT6A52j459uLNBJaetpk3vAKE":[72,50],"4LKx5Rz4NsxnpamAuD3xVcCdt6A5aoN89qaUuFsfBNdW":[52,50],"4MNtUgysSfjwfpgYBJFJQA2Kn5LXPQzgRLnJoCAseKrx":[8,0],"4Nh8T1d4YBZHEuQNRmFbLXPT5HbWicqPxGeKZ5SdAr4i":[68,46],"4P8diDfWD1ra7bF8BXDPUExMg2QAhTxVLTq3tU4QcH8p":[52,43],"4QVu7BnBgYBEkyq9zc6mu9V7HbNUX9dK4EfVo4SMvBwT":[56,46],"4QY21MyFAtXbagGymZuBLu3a6wUkFg5qaUDRwYj4Pnuy":[44,20],"4RwV6detEgRyvVcvhBv8gmjriEHrVmKegeYy1FqRZK6Z":[52,44],"4SqdkosjugZVRdX2kRptUng487Uece5toWHZXVh6cpQV":[16,11],"4SykXpKfFGdy7Yxx1wToYuBhhnTTXMhzewWBMp65wgnP":[12,12],"4U7KuubEDSPR3YY1YjmVjz7CcxVgrdz7sz1svUM4Vx3i":[4,4],"4UNcH9sxWUo6bfZY93gmPiGZNssEgmG9Ho7C9ecjMv5N":[4,0],"4Un8pHPkosqAkRabaxhA48YFbji5sk46ntFAxQxyc4Lf":[4,4],"4WPa1hkBxCBnHmWWgM3yt8TAgA7Rtfow6SHPy4v6yG4z":[40,27],"4WkMVnmyoWuAGifnmqdWNtD3nudHp4hPPqvnyUHLkGWC":[28,28],"4WufhXsUhPc7cdHXYxxDrYZVVLKa9jCDGC4ccfmuBvu2":[52,45],"4XWxphAh1Ji9p3dYMNRNtW3sbmr5Z1cvsGyJXJx5Jvfy":[4,0],"4Xqmh7JpjaFj5wJ6tNGbEY8eoY8U3fPMUKzfQXcGWiDR":[36,29],"4ZD3xAHfPcYacfZEYmAxS7D72UdVFhUUe5XLhEnQfSCD":[12,0],"4ZbygbNLCxdMa3EZLYBuQHF4zzfCtX5V6xJAVSZncnjS":[36,35],"4ZrtLrxqtpQE4juoSAsSmQKgeZLkEdiwi7gXZ8hWsVF2":[52,27],"4ZtE2XX6oQThPpdjwKXVMphTTZctbWwYxmcCV6xR11RT":[88,74],"4Zto93KdBuynSnyyQct6ecMVxGNrjvVHe4CbWJTtvLSq":[60,43],"4ajWybNN1XqaapKEEiz4MPMyCP7Ppuw7FMQwQ57o7gFZ":[56,50],"4bLyjRauEjdJGb86g9V9p2ysveMFZTJiDZZmg8Bj29ss":[56,35],"4baXhu594FEQtZsAmHNjNM8K3NxmPNsYCxyPUZnhwHLm":[64,46],"4bnqGCM2a14j1CiJ31gjJUf9B3kHZXzz3cFB1X1tSGft":[40,35],"4bpkzvzxJkhXCQNufEcybrXsT5vNW5xUiG7mcnxfGRGy":[76,57],"4cLRyEVzhvt1MKqEeVeVfsxfJzZyUwpJGQADBW9qgwks":[64,38],"4dWYFeMhh2Q6bqXdV7CCd4mJC81im2k6CXCBKVPShXjT":[44,38],"4dd19K7UmrUk4aScsqYaXEGcabRVh8opRhLo8uSJAKbZ":[32,15],"4eyn57baA11sgvkQafTcrwJ9qVs6QptXBahf43Li1jKc":[40,40],"4fA2MXsEG1mJfxTouJuFWQoBzsK7jQVXbd5UAfhMZHXk":[40,23],"4fBQr617DmhjekLFckh2JkGWNboKQbpRchNrXwDQdjSv":[40,36],"4fFhfoSezZmrvK5EeFRtMsMhHn3Vfno5iJY3JPXs7F78":[4,4],"4g5gX1mmFGGragqYQ1AsRpB8ZJvwCoUKVT5LtKTDrNSp":[64,51],"4gMboaRFTTxQ6iPoH3NmxLw6Ux3SEAGkQjfrBT1suDZd":[32,24],"4hDEtsHXAf6TMBNJHogmN5noitFzxGxKAs5YwsKZzrDd":[60,49],"4jZMrzWGfMHDRkEBqwnx1cPR6uP3i8v2EaKALzi7bYbc":[64,58],"4jhyvbBHbsRDF6och7pDQ7ahYTUr7wNkAYJTLLuMUtku":[56,42],"4k7N9gtmQeDDYJxbT5NSDkARghkQzEbgvE9mm8gSFicj":[8,4],"4mCp1G9zmqRH53wX7j17wmZimHbn6ep1NvLmsMUwHjDj":[64,27],"4mdxZgQQdkVJvPK8Z8T55sbUXU25ZzjTNs1ydvrzVnYs":[48,43],"4nKuNB7KsFPzfPURvXxpyBZu4Pmm1y9w6jdbHpaAEfTH":[40,36],"4nu5rdaXjhXHniTtVG5ZEZbU3NBZsnbTL6Ug1zcTAfop":[20,15],"4nw9knLrjB893wVF1PwpPofZz19ko5vWcq7dKmriiSdH":[56,42],"4o8VRbGZcmiWm4Zc79LsBgDcqXmmVte3kvCroq2zwLG9":[92,84],"4oJuAMQjsVoQdHybK5JsoKYUoR4akAZHNRa4Qjs83Dgq":[12,8],"4oNUWNoSNnwghHBCGsuAaQEuaB6oZEXE2w4VNhRxoaQc":[28,23],"4pZjWxF6277CRncZjggHdiDN96juPucZHg537d2km4f9":[40,35],"4q1KX2Epud4kS7tYuyndLaon1FskmDqcwh5ubxHiSzdP":[52,48],"4qVaZm5ZhNnNwwBYawGsM4DoSkGXMkxymMYnzCTsH2WY":[60,45],"4rqiq96AtM3V3me5aGKXnSycaVZgoNq8jYD6LwtryNuc":[64,49],"4sRKUyYwqmc38TpPGmkbLfjKkyNBGEBaiYJaMCYfkUBh":[60,49],"4sSihca8PLdP9Q4NBo2LXXBE9o4KUqpp4hSEyXQCS7Qq":[24,19],"4u2qTnf4QVC8PcgNFPBwY2PwdkiMa4jb3KnNZo4zZbtV":[68,45],"4uERagALHEAGx2uwndDoYn9WpJ9D1uia7Z6vuMpvWxuQ":[76,44],"4uVzFAT5ZpJ6cPo9ff7igWCT4MjcTVwETqf6y29YBzgE":[64,52],"4uXyHNPLMpdjs38aorfRUCLarbh5ydhbv2FkZwErBM5j":[48,37],"4uykzcDWW8wnVWMXXgh2RqXaddSVsx8TNvpJV7eACXbz":[36,30],"4v5dEHTVmWTRzP1L2PijNr5B5nDVUk3wNy6eJ7V8qQKQ":[28,24],"4vAu8eDW1YGVSQPMgZqAVjYDVFXJQPtVYQaryCH26yam":[72,63],"4vDoJgjaTyQ9uRLyCfVwb9pyhAZQRcvGNosvYaJS4eux":[56,37],"4vXPjSaZfydRqhnM85uFqDWqYcFyA744R2tjZQN8Nff4":[48,35],"4veSBAABaESW2WpnJzcdNcduopX7X1f63KziC24FhQee":[28,20],"4wjZmBoiwQ2s3fEL1og4gUcgWNtJoEkXNdG1yMW44nzr":[60,50],"4x7HEA12XAiqjsM5FbWkyNnwKfqzSDHWA1XA79uFpzGJ":[52,36],"4xv6aEhBpGsnXStV5GoxEdX22p5uDzVNFKEmHaQUhPnM":[52,46],"4z3zVorKFWWD3ULZ9X1XVcz8rYtxKiNE5AQtSpEuYd84":[48,12],"4z755TDizaUVyRRKw7y8DnTnnon8ksQYsZyU3feF6yFc":[48,43],"4zE9u54ZvrdkbBFS6rWVEx31abdH7GFoRCZQd8mDiiak":[40,37],"511pMfd4oivn6uE7MrcJ21hTvcaCtwPGTLgnQAfopir7":[48,31],"512wm7UysDB8PNwWpjMBmRgYHdQAoj7o6EDJ9CUyK2kb":[36,21],"518q2YT5TjpwZM3sLSTk58VVmdYkF86abh7GGyoUaHZ":[56,0],"51tQJUb76g83KRD1GBdtYq9NCZdrgRrJ2Nva8gdLS41N":[48,38],"525qsEebxz5jk6tEbYoUGJgrw26ttmi4CngP7BhS7vSK":[12,0],"52GEvaeCcEyAUKrfoPcey6vdyw6th588nYPuCkn3Kxes":[72,63],"52JQ5kmWuUN5ZVbWMJjJVpd3raNEtWRJMgxWp8J9mdv9":[88,67],"52rpdXBbJG4ChidZc1BiMU5JucsJQQa98zZUEUaP8Rwy":[56,55],"55nmQ8gdWpNW5tLPoBPsqDkLm1W24cmY5DbMMXZKSP8U":[4,0],"55tZynRDphTaxtH17x87FjcyJjCHCch3SrVxuanUJZmd":[40,31],"56Zc7i6DHP7BKWAy4onLkg2sDqV8U9TtJnWkNx5yQhBW":[8,4],"57DPUrAncC4BUY7KBqRMCQUt4eQeMaJWpmLQwsL35ojZ":[60,51],"57Nqrmi7wnUsvBdrkSpyfJHWic9dJqw1KpfgYtmx7XzR":[56,50],"586jjL8bHmqtNTFaXpajEJxY5mLnQ26e3QTHcx1Z8i5c":[24,16],"58LZCrAp98h2tebZq2Zpzs8zQJ7gFqyJMsBUb2J2CVM2":[44,33],"58M2W8tybgWy6pJVqk7tT7YF7C3rmUxVM4MWN7LG6m7D":[64,46],"59TSbYfnbb4zx4xf54ApjE8fJRhwzTiSjh9vdHfgyg1U":[68,61],"59quUWb5cx7sx669VWzj9umtHBBuRF5rpDrFrHvvtE4T":[64,51],"5APtJpidysCuKZCQQ49D2ba86NPNr1UNkGiDaehmSLSL":[72,0],"5Adaryyuxs39jqDsoke1VgUh3R79nQR53JRKBahuJSA4":[52,45],"5B8dRstrVg4NXw39yswMdr6ETHCsbKaSbWCAxCH6gofs":[44,35],"5BgjKeU8bSLJ5hrTGZXD8NrqY1DYYJFz5dLnD2EQRuEA":[68,61],"5Brx6TNjAkzQ4JjToEdL9sZjcFbNpGQRBgpbNFzXPatk":[8,8],"5Cf18uw63TPsS8XZ2gHiQKzxPh7i5axu6knFfAXFDEUe":[64,43],"5DsrdX4xPok2YNHUEtQsRuyAkDcdSBPXM74ezfRgy8Vm":[36,24],"5EamRRDR1j78iE2Q1TUmoDQRw59m2GTs8QJWtnTZsKf8":[76,71],"5EgBAoCdu6r5BcrntpAW2rm5j77nSxJeD7oMDS927hxq":[4,4],"5FLt2q4cZYcU5tK1zKggbGV6379hhZFKaRWMh5q9Xpc6":[108,81],"5GftYPpZU6r76FCJ7cn9BNGM3gmB38CRC4bfVHNmArA6":[52,39],"5GiX7EzEaooty8he3EJdNsLc5sqbT4iMe388cZfCgqwR":[44,40],"5GrycfarfnDuiKSfrw7XwWFKpfJekANP1RwrfQtmVX5R":[72,62],"5H3sMCaSJdN2k1hyuFTzq2BrZHUq7CinTa82hJS6EDTf":[44,34],"5JEE2MaWy1TMY8Xh7HLK7h4xJQZuAGgTCPTv5Fg6mkUw":[40,38],"5Jg8XRMaQ1FJbyy4YN3t3oCJeXUprHgQTqjysuUMQvbU":[8,0],"5KFXF9DS2ETQVfUTAfMygw6LNbqiXWXYr4y2k1kYr9bA":[76,42],"5KG9uYHFKSmJVgvXys4dKkZ1iVzmsHxDJWP1SsAw9ahj":[8,0],"5KK7GDAws7uYezSUcugdVrWNrKNA9ooP4t57Jq5W1mTa":[80,71],"5Kev1Y8njZLiybgnqTpTnjZ2H6NMtCeSK6J9TeqhyZnL":[16,8],"5KjwhvyQZMbDKfQCSa7L222vxWqJna2sfRKLXTPEyEwg":[72,40],"5LB3ieVKg5tR5w9VYLDfTh5u7DPbvDo5uKjoLQvzHK9T":[44,32],"5LF5MEkfKo74aX9zSz8sqLoKv61rv6bu7YgoLLkwrqJY":[52,29],"5MNLjn4p1bNUMRc7YP3rEWB5BQbzNsHYaqmQLwshAndB":[68,59],"5Mbpdczvb4nSC33AWXmh6wmDxSZpGRANNcZypdPSGv9y":[48,30],"5N2xAEANgsK8pgPrDktrXUGWJ3Cnt8TbPrcWqLWSAbq1":[60,44],"5NH47Zk9NAzfbtqNpUtn8CQgNZeZE88aa2NRpfe7DyTD":[36,29],"5NLjk9HANo3C9kRfxu63h2vZUD1cER2LacWD7idoJtKF":[44,39],"5NY69Bgoaahz6gRVgG4Ub2PBzgscsARgAekLx34Mtcv":[60,49],"5NwYJ83brnYDwQC8Hn9hYXhCiq6HQ4jZNGLZKdZeQHPS":[80,72],"5QAa4WEyAtEi7br4soyHSCHZQmxwrTbBy2JkWJdRPJc":[48,42],"5RhZ1sBj4bxxDF9JVBnxANjYGf1cuGMYbmumwuuuqNe7":[44,31],"5Rjq51GbTVY871gHZsLSknG7a2rqkukBxuanAJYDLVMY":[40,28],"5SAMpCcejTXQMnbrtkNv6nSxqaYgjRbk733QNzc4teJC":[48,0],"5TLhtuxkDdN4Mp2iHeSEZrDzpZ2xZRiEdFxcA9ipbPJV":[8,0],"5TZbMUkDaxxbyhkpgMQHZQCyvHAmsg9ZyDHf4R26qrap":[68,58],"5TkrtJfHoX85sti8xSVvfggVV9SDvhjYjiXe9PqMJVN9":[4,0],"5UCh3FzaGJuX2tBmHvPD6LXNchiGretWwnB9LqVif3iE":[64,35],"5ULHpStmbJSLYhke2WKwWRS2n5dVeqisWZJT4gjkNTec":[24,19],"5WhrU6gqgCwNBW7tkGsAZTB5bno3ymHVrmQb5yyxexBP":[56,36],"5WzxMJDwAwaMPwc9Te8TZJrFu2QmavL1S5SCcPn4VbgB":[72,67],"5Y7Rq8DBLwmDGgAUPKXyqJ57mRC33krMyH9dzMpuwTxF":[48,36],"5Za8eDus559NMWtNxwpWFqW4cNBuuVN6JRSCiRqdXhSn":[60,56],"5ZimkW45n4mWVCqXsqEEJuJWvhoqZFX7iRBz9jtHW3PQ":[48,33],"5aGEHgWCyHNxCcNMHP5TDddUkT5uXGpuwBfonE13jnMB":[56,46],"5aMayQwEmWD5J1anZgavXF7G6ZczrWG1h6UC6GJ11YeV":[44,0],"5buj3kmwRSZSmGPCbPrfbFmYZ3fHX1cjNe7KZSJPR8s5":[12,0],"5c7YoYxtKLC4EyiXBiREB8obfMyVME9zRHVtuMd6KhfV":[40,23],"5cK8WPnW9Q7rfTynaHTGHXHNRyZxHHT1iDH5LyPeaSQe":[12,0],"5cNCJuzzWPmSXyqDEhJL2rD74Xaf649mujFEPm4UzaJp":[44,28],"5cNEV9dx5Puuwp2GSZsUwUchCGfxchcxS6rNuD6yNGEh":[72,27],"5ciLz4FfhhGZnoGX5hgjnKzL5xdc1iNqmczt4moFTQu1":[68,57],"5dB4Ygb8Sf3Sssdxxrpbb4NFX9bMrYnieiz11Vr5xJkJ":[40,32],"5dLMRyPWx6rdPGZpZ7uuZiqry96dUT5yz48u62Gzugi6":[44,40],"5evLgbdZJG6RrZzeW5UqRLEUPjziqpPXmMKPNqdEmxWi":[48,33],"5fNzsJQxTQ93RJUgnGjvCQ8qjtYDGXjVMM8ERJs29YcG":[76,54],"5fnyGEnVu3nyMrUysGQLXz38QH51VNtmYGSA99197xCX":[60,46],"5gYY8gRdTyLP3TyLgfaGBP7x3phoCjUrRRz5JaxCeGEF":[52,41],"5gaASWLJbeYVk2Kd6shQu7JMVfkXHnLNwiSje6XrazyN":[24,21],"5gpRDdBffGa9quGE7hTPVCg9zVnHTS26qvbd12G5kSS2":[52,34],"5isoKqxB8G3CVngTkrHddmvjHhuKBiYZwLfWDufWZtwU":[44,34],"5jAKgxnCLVrb5zdDxjnRotwNirVG26Set4ZZ6BWC6Sx":[52,45],"5jLVeSB8hepuqgReNhcNypntbcD2wi54JZ3pYY3PGrtC":[36,23],"5jLw8DGMmjwaCJWbkT3dksXVEdWrXzQtiBd2TfsF1J1H":[60,20],"5jQqKbCAeYLiKK4WqppHhKBxe4DzDZMRLLaDhDQJ19F9":[92,76],"5ju3ywSUjEfWR6HgMmqhpDedDkcztspcS59T2EiXdxWn":[32,31],"5k1ooGz9ZPjGk8PbmA7Czgk4nS1Ns5CAKXSeJgeWqo2W":[56,52],"5kbpQzj1FEqqbeU2XrmEbC8gX125XQkdt9ZwYdBh2iK":[36,31],"5m28zJcp7CsTrH2szyNQhygvDis3dPwbgrtYsWi3J4jN":[64,41],"5nR5ktqmZufaVuK8N8nNoqVrQqopL6qAnf7YNvsjynhz":[60,48],"5nT7adimwUD2MfMRSrNKDoNrLF4G1mrpsStxzU74v4sZ":[44,27],"5nUy4R3g53WdtS226FdVWaVgXhJkQyaitSn1Duu15mXP":[64,51],"5nVDe1R4QW8XcaWrDUo88tG1V8CgAV2BqWpCX4mF49TE":[36,27],"5nvj4tHGRCRFmTaJfpjx3RUcNPtHv7dDkxMbc3yF8UGP":[48,29],"5o2kjsEZDYnWGfTqBJdrBnRYKvRy7wjrniivKwFqyTsB":[48,43],"5oEY8KESH5k8kB2WXZ1dhRB9YELcMfjs52UBmuL6e5HP":[48,34],"5oHWyQwDW2gfrry8iqyxYsiSrNt3PsREeVyY9RZZg3r":[68,55],"5oR5dh1WTi7ACiq8bdYmQN84kDG4HDQuX6cjyJErgGz4":[44,35],"5oVky3o3pNbZfWndUBJbxH82ZDqaUx7k1CorxfisKWZt":[4,0],"5ogMBk74DTpRaEahTtBrrsFN5mcZ2cfmZfPsJMhJm31t":[4,4],"5p3Y7UV2oZrSTTSLJzJknEzqQpetmk2NB2hQEKPc43dC":[28,28],"5pzJe9dfwsjdSmaeYAg43xyTsQHVP7zpLefhLsFxaktq":[52,47],"5qsTBZQPAPYsCBw9aPC6wCLpyPua7VmK9yFWk8gLQaUP":[20,14],"5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq":[56,40],"5rRT889dQehyRd44HVm87UQ2nTko8QWNKVACbkWBjaZ7":[32,8],"5rxRt2GVpSUFJTqQ5E4urqJCDbcBPakb46t6URyxQ5Za":[44,34],"5saC4V5Kk7Xr5zUUbQZHfXCrzFWyW9Lvjq97N9ydX5nj":[44,35],"5sjVVuHD9wgBgXDEWsPajQrJvdTPh9ed9MydCgmUVsec":[48,38],"5sjXEuFCerACmhdyhSmxGLD7TfvmXcg2XnPQP2o25kYT":[68,58],"5t5yxCvtHxCkDJCCrChBQ2hdcUrK61tr8L2QRHtbnpCY":[32,27],"5tR48Ewee96cx6MNBXZb3jNtTBWmMimYxqZkj6QYHgdt":[52,38],"5ueaf3XmwPAqk92VvUvQfFvwY1XycV4ZFoznxffUz3Hh":[4,4],"5unroM4ZHe4ysnprhGrsHBUMsCbkfAHU1Z4rMtosbL26":[56,44],"5vKzPeQeveU8qnvgaECkdVdBks6MxTWPWe48ZMeC6fdg":[76,56],"5vaCfp7UEpW5qdJYyVH4m93oMzzyzTqXdbr7xLGobY8q":[64,20],"5vcZ2tAziqSyZdJJgknngPW1ngnZ3R8bjSqdeb5mpCzh":[112,90],"5vdpdDS5vvUrPTpGq8zDWmirYheKHq8RWrQfUrbarN29":[52,44],"5vfvM4qv8UERxSU4qjKhcyJYgfvBwxM3zotkbyXg5z4z":[56,38],"5wsN9Q4XLXvxjefK2tszV1z8DRKSXyGo2NxvzrftnDQZ":[40,36],"5xVTCAt58jH6i3Zkr3b3EDyxSP5CanqY2tRGT1MStVUr":[48,44],"5xjnsTJwtYXWNFPwV3DsXmbsi3oz4bZbdSukuEwoYdbT":[48,42],"5yqmdjMVX9F64YuE97neemY6Q1s4MgVaBbJiz9g5qGiC":[40,28],"5z9kpN7JLo7HkmSYLA3dV1yTuwx67WWFg8FiDEk8kJ9o":[48,47],"5zELkZ6RECLDg4gi8sPvnGPAQbfNKUdSPunPh8HMNn5V":[44,38],"5zRUbp1Dtu3qQaRVf36oMDaeH91D2ePnc5DEgnh1ivFg":[52,43],"62RDaC4ARQMyZhdya46zuvYs9L7TrR43KsochCnrMVm2":[60,55],"636LepbxdwpdKNpzuxc9vJhYLkhBQgBJE9yiwrbo4nrc":[4,4],"64TmBCNgzNp9StawrEncNyz1TQWuPmGiGuDeL6fqkAvq":[40,31],"657Tpmj8yRfJuj4Dd1oqdKA1Lo1aTruGeSkNTaSabHAJ":[32,23],"658mqpjmxPwfu7gm3PYPfPsGqMagn5FeiJ7814YWZsNQ":[36,34],"65WF5UBc17DsJ9yMdTejNuU5SpeV6nd35DxNioQDarox":[44,43],"65kBAfNpGpJwShoLbGfzJvWwzeZt8UPvZ9xHbjXC6rJ3":[48,34],"66dX6ZwV4W9etrDRkBvTrgiz2BWxogjELH2T6Pkf35bz":[40,40],"66puVpH5uFvjAHAnAQma4xzmUAUuNCi6nRkWhtRwKY9s":[4,0],"67VDb2iEdx6XjCfBLXhUgKQQjTuLe9X2eLqTq5nBjUTy":[60,44],"688vLxT7Gsb4YX9YotViUauLC5aYbnjm1SQtaEQUKitf":[56,46],"68qujN79HiknCPBbGESncjUDeC8V42DigCGjQjpaVher":[60,31],"69SHUke3phQy5bEEKSVaSp3ytmEJF8h4Yh8FssZDHNjE":[68,55],"69k73WLdHRge7E3vCUiDx7Dkm1DQSBBGAu9FqNj4AeJD":[52,28],"69y24KUYmXFY2N6BMfzL8TfiKjQtBNCCjtnju7bxh4zG":[60,56],"6A2t1aNmY4c9DsQuZgjMBwnigUpCu8vihugqgyAhGrC1":[36,31],"6A7vAYUkn5wKnUqnf2CJxg2kmbtDidhvF3nf1DNhyqfj":[40,36],"6AaA8HJGpYK9RDN5NQjDJfHPcqX63hnw3NXEa9rTXbEs":[40,34],"6BdawAEJvEbgs7UB3VsKSc1WL45ydCR1xqi9pyr9JS4q":[44,40],"6C1mHAPxQACd8NNS1D9KpGxqSRUz5s6itsaJx1uteofx":[48,35],"6D6oRzvSE6cdpKpWgojgHhrc5ef2jqQhRQywtzp5GreP":[52,49],"6D6puBzRwMwVNZUuEipFycFL7xZgL9sPEnj5p68Tn8iP":[60,39],"6Dr57RWT2ctMt2XiQxj9Nec5mBrfjucfAyh8hWQE9cp9":[44,41],"6E5NygCNcfyPHkLbHMckzF25cgQoxN3DfMqH9bwyQRpf":[36,30],"6EPcBVHgAaP1x6rYkcpuufSMoP2jhRYi3N43PUqosTDj":[32,31],"6EfiVm1bAo8yWgZppb5irTqciv5VC2eoNTFToST5c6Mg":[68,45],"6En53HMLJjuqkk54LgiqdEoSDXfNzqwWVLiF7sBWWFsF":[68,40],"6Eph5j55RgzMx5ogbpMg27hDW6yoxE2Tr1EMc4SqpKXp":[52,36],"6F16m2H44H4LHseHDuk67k2zdEXCWQdGnA2BQ4yMQFMX":[76,46],"6FTLATh7CDdqkFyYJuTR7oFyvhVK6UHUK92fELg2mRno":[68,53],"6FWhS2CHjtCf81GMsqHRXQqDUh3UKyyWGF15QGCWWb7Q":[88,61],"6GA16fyWGrr78QWGUBnH469A1dWNv8yTjuVMuDcxjxmg":[68,58],"6H9aeo7woPe5QabarbwHtkrziJii8x6RT72Eroga6o4Z":[4,4],"6Hq3pmDPps9ybX3vR5RLmJyXYa17Sy4ZjREzfRbrzTe4":[60,47],"6HtPhr81VwVaMKwFFzML3RrF6PMcjipbCpPT7JbdsPvE":[32,31],"6JUvAc4NV51SfX8G9zwoRptU6hw1eC3fYz443Mh3Qj7w":[36,35],"6JaKmYstgSj55SwwQDihDq1Q251FrL5ev8XNiqFSP2qe":[40,33],"6KGDh5hSNAeDmtF4tD72Rw5WtZk2efYxqgbVryXmis1J":[44,42],"6KnzXAhpE6ki8GuNQBqpHsVdHhsyg5csChPGLdkTHRvW":[40,37],"6Ku6Cj3Y3FETU6JEuwjpLLu65CnKhi5YGtKdUTRGud7i":[52,21],"6Kwr8fUZPmSFNWaXfRL7e7v38itt276DFVu7RiYn8oW5":[52,0],"6NDU8PkeQhH8DF5Yw1Cn1AexQoQLnqr13GhNuQL1gfuT":[48,45],"6PYMaoJf89uNKjUPyf1eUh6KQ8vGAHt9Fb8EK5SqctKK":[48,20],"6PdBqw4p1iaNd3CYg18THHpzDBuophRUk3qSFy3KNTuD":[48,40],"6Pjg2CBN8tG3u61t1hRQym5aHwNC55DcYV4ypWkpBFTa":[44,37],"6PkpXpQeLMp45TC1PTPUhCpywcCxMcmUXvH3dNPMXQvo":[24,16],"6Q62YX8UpQKABG1FANUsTjJJdrfYZNEAbZyceuaVbx79":[48,44],"6R1745AskhXmKugSRxBPsJPDNggk1nhJZU1fW6beefBQ":[76,54],"6R2SsTxEK89a9m84Z666c7M7wGcmbwNmTCyPFcAcftyX":[60,53],"6SKzCZ2duYHPgDgXKx4ZG8pix59q6WG8gXgyYvgDaNR5":[52,39],"6TxS2SvBtJDVvGADrjYzeTczsCezuknS3gvMH18hpDPF":[44,40],"6UNg56mU9BP1KghPDDgF6v82iJHyLaPW8BEQb5xpFY5c":[88,67],"6UfQVpJKcGT3Cmo5DvUkrDEu9ucHbR3Y3XCs8R1wyNnM":[48,39],"6Un2uhrFfW1q4a4vhQfTxT8tn8xjpi2CF4kXhfaWKHSK":[40,32],"6UrGCcP3H5REdZrPx9X22s8Pj7q2RzUWVT5LFCLBevZ9":[68,58],"6UynSxu2fiY5qU6Ae8cPLxq4jyWVpnr7o4fUyWLxCpcp":[72,30],"6W2xi4iCGU8eTMCGtG3DQXgMGurXFnd5iVXCY5Sq7AbF":[48,38],"6W3xBXKnq4vGHvBjMNSgVviQ6vqDeWiL4LwnSFjvr8Yo":[76,62],"6WHCTDvSa47muoyi5zHoKKPcodkftixsauEfDNB9YSjL":[24,17],"6WJGoPfSX1VNqfiYfRZ9RP6KUm8nLtar9zJxAntdzEWh":[48,44],"6WgbvHvsBkWoUf11RjnStUh13Z2WCcyzs4kTyCooLA8e":[48,39],"6WwCWBHYvNXnDswu6qrHbKoXMqtB1ZwRCD2U3oqWbZmB":[44,29],"6X8sHQkmxRVh7oR94VjsfffmQYPoGZ62Fp8gt4QivszH":[48,36],"6XASEv4VzAyPHmzJEc1wTMZsYLB2ov4ggdzjVY8Q3Uuh":[64,47],"6XjUyrr7fscEiNoQiVFRSoyzzwE4rMYsvQWkZBfrpdH4":[52,44],"6YCRJFeMhPkmpYFaThBktkrWp6qAopiwqHCaj9x6feHs":[68,56],"6ZEbKFxTjEKGC9HUqzy9z4ccJ8Aq3ktPKEzHGDosQJo4":[216,130],"6ZR2mZ3r5oKHCghjMK7J4b61QEHy2b4vqW7k2viGXKLr":[56,51],"6au2pU33RmTdpoZ9WcYrHnTmTByMJbMMPmZPC7Z454hP":[48,44],"6cbkU5eSmvbsDaupLiQxDNSB8YbXfrkzrtqRMLRr9ZLP":[48,29],"6dr7c5k6SsFRFfmoNqADxZQsvPjPjg4meeEHVX8cn6HU":[24,16],"6e4BdfrD42d5FHVbsHrKEqxj5zzn1Fq5bC9e3UwFr4DY":[60,36],"6fdimp8ks17wBAEiX7MuF9CrmJZ7vLGtThXdFsrv7Msj":[56,52],"6g7urUx43pwjUZ9CBD9c76oLQtpHCgCxp9hQhv6RUMB":[52,34],"6giEzjcXWwiodVL48LtoFexax73cBorvq4NM8a2xUkd8":[48,42],"6gjvikJDNh5nqE9T29KaLBWUvc5ouramxzhhNHPwao56":[64,49],"6hS8hZmb9KEoAuF1pXPPdpJBQgsuepYyAhghhAVb8zdE":[28,24],"6hcoeM9dVx6x3QMppHAZwsiXYRsRECxWQJELZ8pYv79n":[72,63],"6iErFcYDqdg8MQcNPZevBJqdGsDQx29FrNDeAH9zH9Ku":[56,38],"6j7DvYDyFTdrK99apFuuT8w2WaeaezfwLDLk8Em8sB2m":[72,60],"6jWgPvH1U6Miu85ow45tHyvEXfjmPuNDFAjovCHRCLid":[36,0],"6jYfBBZderjvyXiqxVWNP46GSkDZmsL12m7D5A9tfaND":[40,30],"6kCRppFT5zDZ8P8L2ex9m6yagTAm9e7682F51ADnZQ7A":[52,42],"6kQGi1Z41F1Kq8Jqn7AT5fUUw5NQQshx8sDwCSznRkpv":[52,6],"6m8LGKXMT5QrRQdQsQAd2VHpYwJZebbrc48WgkPWeRYc":[56,28],"6m8NNtRcrBPvHXfyMQqSYPBpYBNq5Wn8K8VrW6qDv2fB":[56,42],"6n2Ebk85o5BAQSEbkukekWcaeJStyKg9NWWUXD4xyRFQ":[52,36],"6oB8HATu5ApWMWhFpE4Ms5XMNcKmjk83VpcND5U1vHof":[56,52],"6pJAzQhw3MJBbR4BPzhWJk2Hf5p7idivRrepCuh1BrEu":[36,22],"6pU8UoMLbohFiSAoHeNtpRPkifHZpvSVUQngkTuMkdZU":[44,41],"6pUayMw7LVx31eA86LAxomnzqktGX4rTLvDzznRHDuNh":[72,0],"6padcv1em6AzRFRoZ1UVzHEYo4pZ1Quu2UQqxecvZjVM":[8,8],"6pjd2Dfsv7FNkNDCGhzb8vn1DEmvPSVicdQxvGKLVQwQ":[44,36],"6qH22p6RDZFeqszT5fAh4LZEzrp4oiLdNRfHPuetHu6A":[4,4],"6qJPxxgZHCQKBvbGC9zCsuuPtHMMLszVCoiCvEhVULyJ":[72,68],"6qPPKb2zC6U9g8pwAGrYJxy9B9noYiKxwS7NnuRPqpUx":[60,45],"6r51uDXPZf24KrxGU6SnWCXfdickihjQjnTtUeQdRh4A":[52,38],"6rDzQVov7rYNcHSGsVcbQny7VYCinkn4U86Cfz8xYQdC":[48,32],"6sMtZs114UsqAabDyJYjtgD5j92HGVvx2pyA6QnkidWM":[56,43],"6sSBHSuyRRphvkH4GAwccGRB8HdZLWC9VENN3c6S39sd":[44,40],"6smeNG6M7Aers4Ju1drfZPYBS4WFK89EwWphQjKTMQSj":[60,47],"6tknFLCJiEwuYemEmjCSEz6oB5fmQUtH199Q4uuqcTY4":[32,23],"6tptWLfq3o2Q5M74ZJLpGGpA9jCAScHSaun5aQVTtp1h":[60,38],"6tr27Z7WhFGkFFFNBmaf8bHypvajZU8WDJPR7Aji5stF":[52,0],"6v4yeawkVLWpxrb1pTmqtrhYi5ZcQBZtQCvgA6MmRKLg":[36,32],"6vMTzyzXBztMW3Cj2j4SbpR5cLmiG58Nrku4FDmNP7FG":[40,30],"6vZuaLY4n4GP9DVroymfZ4D1oP6xpgF1ExLMqHQbt32L":[84,68],"6vx5vGgqAa9dRaJpbViCNDjzxp6EyGV38YMYbNDqTzLr":[8,4],"6w1jYS7vrmprS1u9cQd9uFo58AZYvJ9JtzihmkRPgSz7":[32,20],"6w8Gxzq1AusnWxrnBH49wkWVemp7MPxXftfyUQy67yJZ":[8,0],"6z5vLtFS6vaV2tY6a6Z29fkw5X84rEqVrtAVhNq4HN7w":[136,0],"72TVuWZN99RZNkEnjPWiXVhrjAirozA7baZo3jEve1zP":[48,35],"731Lnc3mbXpquV8FmFnTL8EE36uoZRiDMUKXehwZq8x2":[68,24],"73YGZpfTSBv7PBLvmgcAwa7K2fAaBoGe2Z9YNWz7J4rB":[60,50],"73YWnZRRabUNGR9NcrEKYdQTonMGWdXHegvo3yphzS23":[36,27],"74TbcQoVmGdmdZdUZTEpMaLonAtKTK29GsZpfn1LWxoo":[52,33],"75A6FVv8hAZn3n4KsTkURtQP7GDU4SDiZxcTzkTHZM3b":[4,0],"77VrLXNcVnpPHcMCcfkNU8SpuPzoCbGfp3Kqc2iV1dgu":[60,0],"77s5a893M4MNWxGepG2GympAjBk9abGkEHZbeGR4pc3V":[36,29],"77uXenX1Y9T2D1pcnHnYsYiwTTHbnzkyrKX5fQFMGVCR":[12,4],"787PK2WaCUZCyYEmuYQSGmoxu7MyqK1usn43FfiVwhcB":[28,8],"78QcjcDqBqvxrjLZVH3Y7vmyCmNdu7VSVnHGMiH4CpcR":[32,28],"79VUA68hwB2XuK654vDgjVP6YU57diuhHiiZrrH5PuCZ":[60,34],"79deTJCsgvqwNZWwXPYxAL1eYE9z8NArP6rkhsVjbM5z":[56,43],"79u2h3dDBUA5xNZwP6GF2cbtzHotobTWuZxZE8DjEUGJ":[40,19],"7A4WJBegWKXVMhVoKshm4GzjW3Pb9od9ECWxF5DbrSZu":[40,26],"7AW5VGSNcaECGKJD2C4rpRuWpcT4kdAHrbahc7KFQM3p":[24,19],"7AzrvKjepC2ohWw5He91UhcwvZwYeHUpXYxma2XkRtE7":[64,47],"7B1wzk2EVeWrorZzFAqX1czbBqB8mV7s5Du8YWCLsQjX":[44,38],"7BS1RfipQ7zwuKAdiUX5CNFCKNEdk82TN2C3CmoXR4ux":[64,51],"7BncGLSgSexiAXz1dRB7cZEDdkKey2sK5xiLpHESDjpf":[44,27],"7C3FrWyhFGc75WgccpnpuuCRSqpZiWpvj6d7U7jScSKU":[8,4],"7CTdZm5CFoWy3gMUhyqG4SMbf1EE3M1qeqLctXVp4ucJ":[56,37],"7DMvXPALnEYhxKqUorHgh13xzBvSFu67dCU2MKF2gsv8":[40,0],"7DyCSDDKvRe1BdxSyN6Q3bvW72VddJbJXG36Ghi8KRcZ":[68,45],"7ED12uoR6C3mr7Apf2x7YnSmEHkApFo4Jfm1bq8i6L4o":[64,44],"7EMmiUb1TxdX6G7B7oAJ3jSsXMgjg7iR3WkqDq67cCKg":[48,34],"7ET3iuFkJDTdWgetabdoZNKgWeRPmAaELMz7LbiHLxT8":[44,29],"7ETjs9tfe3snSKSnKqzxJJHmpNTT474TfcYG8MSQnuet":[56,46],"7Ek9msWDDoe9wSgD9PrPA4Cnm8fVqjvoz3UK1U9LFyL7":[60,45],"7EucomZSKvQdiZLvra8hLszL1kRYiGewxyMJnyyzdbH7":[40,30],"7F2vcJca5ewzdJUNcVMKCLVYneq6CX9JFMH1U7JeVG5":[64,44],"7FVCgatxKrX34VwM4YRhUVdXsJAoB5Kk3EGWW5M2Nqub":[52,36],"7GM9F3HAJeYSdvhfrg6Avq4sw5HBrqutqQqDT7jHMDHf":[28,16],"7Gn1fiPLJp1eb8g8QkyZ6eLvHgZbBnSA1ZTsicWPchcV":[48,34],"7HSAu6Q5LAqrCk7pt649utsDrrEd7yP5NEcqodFd8TTb":[48,33],"7JFfCpPEodnt6SWY41ePBRXR6LUGiKhLSKJNw9ZYjdah":[52,46],"7K32uTNK2zJwp5WTt4t57qJMf1JnHBq2HcSkc4oV5sQb":[36,24],"7KvnzA5iLwJeQ2q84FA3K6ZAKSWwvyBXPGPqcEea45Wu":[56,42],"7LccZ6QXtnSMLqXiTe2tDnfz4WN1Z3B5RPYsEsH8w1XD":[60,23],"7NNpLpbJicSjeHUXiLuQy1cmnHNthnsmosbLes998KqL":[4,4],"7NPcRcHu3jACoQf54nkRBLgdn7zBbUYhnsdC4VHqBQwK":[44,20],"7NfasVzGcyPzcZhiKdn215iyMTY47Gk149TMYApw1Cx":[52,45],"7Q6g52pNXSqmSkty9TzFotJPpPSgnqqEdSYsA63GdogQ":[44,33],"7QqYu69Sh5WB58JjXEmnuKVfvybC4dxWTsDFmNqiY9d4":[44,44],"7QrgfCAeoEhfKSfmrzbYkrHCydCBzEqAgNnGTHZ6vmzY":[4,0],"7RDFicpHYkxPEJAA2FRqtmyCew98m8B6cASzgtSir7mt":[56,28],"7RTTKCWBJ2XwtSHkUfpwBTH7SsdKqHrWfnD9Dv4z2Wyw":[72,49],"7RUobwC33EbHaWWR2sbdaJhT8x8PpgUoAbJQYrQqrSgQ":[48,33],"7RYAwsb91ZL8117TaYuis356M9abKddFtaSga6XpZhjx":[68,48],"7RrKc8sshPPHMkX6FvrxtQDKFytPiJN7NXpmBRXsD62h":[40,29],"7RuUBtHw3j1ncgtPi25uL2ZawHh61iYg6j4KK2BRi9cz":[64,61],"7S1xGwMrB4x5fwhchayjHojKQoCWZsd5HnRHRUJGXekR":[40,12],"7T5ZekSsBSgLNKVzQmCRQ5iqL5ycprREa1tz3GYmb4eT":[60,53],"7TG3LLqWYn8ybpqAJaiop1bVb5mPWJSDhxPaLzUdF3M2":[64,38],"7TcmJn12spW6KQJp4fvvo45d1hpxS8EnLjKMxihtNZ1V":[36,0],"7UUFbQSderHWPqu6BoezL27ymsgrBbXSi3qQHAozwDtP":[68,56],"7UZAaZTjnsFMze3RWtzpxTG1CiJenrvPixvVxW5xSicN":[60,45],"7VV8eZcVAN79xoGL2eEAj5sXVbQEsiqiTCZcbjisjXUx":[44,44],"7VVYonADe1jj2LtKZMfTKNPiK7gjVRDsX7dvA4BXf9sc":[48,30],"7WH6nrQ1MY9i7pLduZEAuUQNFrSDZJTxQWN4XfmD3G3V":[48,38],"7WNkC3cjwngUYWrAjEwXHgWvk3S1adKJd1JK5FZkh1s7":[64,55],"7WgNDqtFHr1hLYo8wcw8X5uCnGwDSYQMz7MMKL6dyLLt":[48,34],"7X3csFXUN2AZph83GC2FZCpkCTZXfVWssaJ72cpwG96w":[4,4],"7ZbzsvjLnAmu7sq7SFzx5BBgrqWvEtcYMqYWrLDH2biA":[4,0],"7aG4CFBwtm6DY5v6bm4ZudBfkjYjiRb5ix6S8ptsxDEg":[24,20],"7aKHeoUDCYbEYdSEj63i9m6vmkXLbiafWxoCvyhcQtPw":[4,4],"7ajm6amGXayr8qe3nPYA6j1bPMLMCxEmhNdzgz1EjnW4":[4,0],"7arfejY2YxX9QrmzHrhu3rG3HofjMqKtfBzQLf8s3Wop":[44,38],"7atMLyqH6yHXfTi2zMXiTrXtXpuJTsqoPrkD4N63ddv1":[56,47],"7bFx5g3sh5CqupFYtch3J1RdZBZs29HtpXAWyPPyptB3":[40,30],"7cY1beonNGzrqUk4pNWErm2vYcyw5yyLqwnrEHr6iKmu":[28,0],"7dEjSFnrm66CJ7Aj5mC1hsYmMzmGgWPr6iZNhcvANZ1w":[40,32],"7dKiopJkRwh6yrsBUnhEX5zTNCDWQwBRsvbD4fTdrUVc":[48,36],"7ecwp7vwPo5b2MNbx75yA7qxKeDBJDpQMzxWKYrM5rB9":[68,51],"7fStZnqGYrsdYdZaNyPEzvMun8AfC2YxFznA3RukvkzD":[52,28],"7fc8chLC68vNnvk1yMuAzFgGmAookHi3E5tie5FaCiuU":[16,16],"7ffVbi9Hsq5scgmBEjiinVWEcKAedeUeqxXMBoBK6JAS":[16,15],"7fv6zGstESoyWYdrfeW1DzN4fabJm3M2mRUVid6bx4EY":[60,45],"7g8Dy1BWrG32N3Hx993PBbpyrf8gTBWUEc9TpeiSmrQE":[80,60],"7hAddyJcvQAS6SsfRKLJzYPuq4h1XykRSJEUmr64p8oF":[60,50],"7ivaGH6xo5sf4hACXVNSqeeEDeiLFLTEyqQQwffDH133":[40,31],"7jdjBzSKJnqHZ676UURpGTWXteU9mr1rpBDW7qanTwsj":[48,32],"7jeyRTfzimBh4PjYbFcJWmvi2J4ieyVaaziMxc61LEdu":[36,19],"7jgcNRZbKcEWGZndJDpsgB6HwYeQ9u2EGnSX5meoyCbw":[32,23],"7kmwiz4wbzf1kUSZmKKzaJRxybGeDMLSqhR9s2FebhoY":[52,38],"7m3rgSgyS4HXnBAc5F8tPY9PTXgB5wLNz4xC8b6XA19z":[40,32],"7mtKMUgM24GPTiR2krRimUiQgXRRmMPmmPkQBzMZak8a":[36,36],"7nuWizYQEpkytGTfgEjrvoFFNUCgyzfvwLr3wbTJotym":[12,12],"7oQ4anNmvJmXUXu6pkXFb5fmovuPSdWzbRkwMBvHi4yf":[48,37],"7odF9faZtwCEATzPFjyFHCYQAgNdBxGzB5znqEVJzbVE":[12,7],"7odNkZmb2sG52sojMv2n2sBXsarYNPvgYPSxV7XJmSei":[24,12],"7otSffb7AdR3zoM2DkrwhhhfLCLhFzwNtn6XxShnwLmL":[36,20],"7pKWpEQLzie4ZwMWaPpcv9Ko66hv8FP7kZEx6w5wgujc":[48,32],"7qzobSMqm6JQns146x94kDeFd8BSTxXtFwCKGAbj3G2c":[56,48],"7sEpbQB3Dryn5JhQVCGWoGgUfYwNEZzjPNa1Tu9mVa5p":[44,32],"7scarR3Z5obfefZr8bPKYoMNipua43K35AJAc1YchQBK":[60,45],"7sdh5QHFPo4ktG9SVTPM7Sek1WLZpwxNNHudojYi8dK6":[44,28],"7taXjCmy78gNRNii8bKngFXYMtPTK3AAGJsrpd2jxe3h":[56,4],"7trWtWjH3cGfSu8z6MgkqEEuCJWN5NhRZBYvbT841Yi5":[56,39],"7urBmScRfdSH9CpQ2SAwfmvGXp59nTDx6Bw16USJVvGa":[64,42],"7uuvTX8C2uys6EcVdoEhESFc6Akd5wZLeAerJXrPcdzH":[68,60],"7v1L5zcmpYyde5s8hvtBSrdW4V5t8Qef9RB2xTRKSqyM":[60,50],"7vu7Q2d4uu9V4xnySHXieeyWvoNh37321kqTd2ATuoj6":[40,0],"7whPeUt57CtDY9uG8WdUaNxiZ9kE4dkxXCYgubpjnbLm":[56,55],"7wsxae1rHhA7x1329kfhGKzukq4Ujhw9D241ziBxdKY7":[44,32],"7x29aMXJ3kxxTXeU7ur7NpLFWCmedz7LFVo2oUqYA7tY":[56,35],"7xAPc3WYomPjiGA5PAwyLMFUNej7ESVsPnXaUqMyrqzE":[68,54],"7yRo8i8dV6MgNnUnzhHvgwExxKo2HqLHJeeyK7AbL6ME":[64,57],"836riBS2E6qfxjnrTkQdzD1JkFAoQDyUjTmzi38Gg84w":[4,0],"83PWQUxkBDrTJmJeFL8VUah6BK4p1JPGdXVuJC9Vf2Pk":[48,42],"84GZWtzfKYX1yfstmjA9eUEp3RnWys8DmsPjsd1ay7wv":[44,31],"84eZv5dPhXSRC7L2yBYW1XBLLZUyJEeBD2Gz5AhQkRFS":[68,67],"84sCKfepG1RZEfAghQfPi4fK33skBx8FuXdAtbPLRDQQ":[28,28],"859Kqd71jDi68MHayXjQJZNy9nWBt4gNs7HomAfdNurH":[60,45],"8641M19beXr6FB4zaf6GPYdLaV695xikBLYFYTVEBZdm":[60,52],"86uo4MtfpLrW9EKd7pxcyKPWBSPQ8jEWYLj5MpVBigFk":[36,32],"87VQhN7dUfS9wacre7vqRm561bNUk5PwUB8xmroc2yEw":[44,39],"88WgTxiDDozTbmzyZXXE1xyv6AKvHUfbJ9Tpw4RmavN6":[48,30],"88ms3Y6Z3pNaMrYY4zdUwHp5K12csjNeffomBnAdyaBr":[108,72],"8AgqfNWYTzmtoxRAvqFB39Z5kdhoW6BV9hYjW8Rs8NvF":[44,34],"8Apz17FY7vts5PUEP28apzqQBVgg6McbetFJqb45ew8F":[40,29],"8BdJzya9vC1PUZrciyMDadChSiEoCLQo8m2yezEzdaXz":[48,46],"8Ce22R38MddAZSpEhLC38BqUEzVAcZh7h9MgfVCWibN3":[44,39],"8E9KWWqX1JMNu1YC3NptLA6M8cGqWRTccrF6T1FDnYRJ":[76,66],"8EUEa9h1NiW2GLrMmqJtDXYer86ixtAgFSrBy3jU71s2":[68,27],"8EqtKHaSgPskksNFSC8oWzSMT2mdSMMtNjGZ7E3KHxSn":[48,39],"8FRFYPcwBan1KBKR6HuPy152L7pr3ePVYVxXXnWzPjEd":[48,24],"8GaMqVpXH7JuEs8D8bdXpe7ztUasAP3wdEXpyZZbUJeb":[44,31],"8GpsptdhGCGybKqEw19pVBZg3gMaopiKtRMVzJFBddfB":[52,41],"8HL5VpqGfTG9SVkHyWU9gjo4xdQYbbdVS7ExTrou3zCE":[56,48],"8HzsgkGhEFP2MKuuPDy5f8qvqR6hmwPqeq7UMY3X2Z6T":[8,4],"8KKQ4QJ7JWAosHwL5pmjKpYWMNSxqtQjJVes2hQezNRQ":[4,0],"8KYQAb2TqCq4Tay6rLTVwWr6BfSMtSmn6E2qosc3xm1f":[48,38],"8L1k1DCCwRoZVEVYZcUzLht9SxUBhkNw9fU5PGnZfw9u":[24,19],"8LkSKTrwqFgw1Knh4gc2BFYb98DnUJvQAxhT1BAFYh2p":[32,31],"8LknwWtMatn1uRenrXYzJS7MxQJXZ245dqTuQiD7wZtq":[40,35],"8Mhs3hgdpL7AKns3wSebNkAfFfqycwd7K651WwSsg57L":[64,50],"8NckKPrPLY4kxcVqD6RV1EVaxyKHkf9DBQe9cz1h6Q6B":[44,32],"8NgvLoYGP7wyramK2gEzS4sj5UKpRVHZeTUSUvMPMna5":[48,35],"8NkJuAPAPTyb5VnUpjdjepHPiD6GR2dT9BxwrmWtzYkf":[48,41],"8NndwQsrH4f6xF6DW1tt7ESMEJpKz346AGqURKMXcNhT":[48,38],"8PTjAikKoAybKXcEPnDSoy8wSNNikUBJ1iKawJKQwXnB":[32,31],"8PUn41CP3VdK15qzrxz1BvENBCjV3LDeSxunxRnRtwGs":[76,71],"8PZNvPTVy3irci4T9HGMFfxx2jiCQov1FWgbWcjps5t6":[12,7],"8PZtnhmPgASnTbefTAFvRPJDR35ivLkEMs4qjfV9LAEa":[4,4],"8QHpiGmDQagfpuBL9jqcAovE1Xihv2nhiyLuwdsemVcK":[52,38],"8Rd6twX9XJQzo8LTshf3Jty7kBQdQsGe9dfLia4vJzfW":[56,48],"8RsYRsi6f3hiK4EhyLS22Cy5KkrNbuidVYmsaYR1Xx78":[1284,1063],"8S4Xb96cH4sNrnKfMDHd4HR2bmjWbeUeo1o6yJC6ZGkY":[40,26],"8SQEcP4FaYQySktNQeyxF3w8pvArx3oMEh7fPrzkN9pu":[64,49],"8SRKNfvMerfA1BdU79CAwU4wNfjnDvFrBo3o5f5TS4uv":[52,46],"8T2ntjCMtcb2zBWmL1BiVD5rho6Eqk41SmMm4AsbuDFi":[72,66],"8VNj7K6ssFcUogRfT6miUzz8HTKu1nX2n8MYr5z49CXb":[36,34],"8Vh9GGKGLQUcHykALzQm5UAb1mFKAWxf6WPbFaxeWWSF":[32,24],"8W8zc3tHfhGbaZgQauBY17TxAQa9mTixmfqbf8PvtYAq":[56,47],"8WqBgoVXkVggLVuvZuF5wP8taQpzTuKGoK6brU5s5Hh8":[80,55],"8WrESd49NkVEUPhnq84ZW3EgmvMWEX6TrNYpjXLmNNHf":[60,39],"8XfjVuniQUTmxiYLJS6AeKDyhajxNqBf9UhhLjaFNcE4":[52,41],"8YhhsDRDEQCi8HJxe7MiXEBhCXeH7LHE9XQYLMUni3k5":[40,10],"8ZZrpXzvuVYPw3HYCPN9GNJhegj6M4pMityCZnCLfVUk":[64,53],"8ZgmpBG5ixt4LVRQEK538hsKTsJBgmFFH5L6X5e9iPTd":[56,51],"8ZjS3d1bQihC3p5voM8by2xi5PoBNxzTJtaQ9rvxUbbB":[52,34],"8Zkx6veTUXdfGcF4VgBJtgZCRnhfhRA7yfpCM3Xty72Z":[4,4],"8a9njgsySJ3LUTvHHyCChKajgZXDoU5cSXkrfn9gf9Um":[52,51],"8aZtHhTNFhVWp4fV3dUfBwsKKBjqzHDwpTZRbpeqo7vo":[52,34],"8bRnkspqHntTjRpnWrCKZpc6pVwChAjUhtZrwUVPo6NN":[64,52],"8butsrHFxUZ75vKbFnyGzvb2DeVZj8uDynWhbv1L6cSF":[40,27],"8bwU8pTjJdC6risWVrmBKvo5gVQcN86djN3CQWYtYrAa":[16,16],"8c475ek3Geh3X5hhCr9Cb61piDKedhMrPo9bkziRqpah":[60,55],"8caQuNVnmywtQnKWv6j8MzzJ8mrLwJkeGcKEtkQkoFZA":[16,0],"8dHEsm9aLBt7q6zu3ESfRXkS2eCwkbbzzynfd2QxDzms":[32,3],"8dYakfEyJqBxobSp6WsSPSe2eEQs9oGHtgJ8xvHbKYiv":[88,67],"8diJdQj3y4QbkjfnXr95SXoktiJ1ad965ZkeFsmutfyz":[40,16],"8eio1idaNjEmeaHUbmJYJDJXAzXm7nNevp96QB9vLA5q":[52,48],"8fELCwf8vTtWJShtMmo7YoySc4CokbsQvm1yptQGwV5G":[48,36],"8fLtWUfZSpAJk7h4XhvM6TqGjXQxiwzWkymxmGtJoGdu":[72,44],"8faCuTioHxq7DYADQwQeAHaKXjqBzELCgUQBieXhmKGb":[8,8],"8guxGZ3yR7L2pBtXgoBnPpq2RE4GM5qvK8UaMG5YXds7":[64,42],"8hpUJeGB6BF1JTZcbiNEgw9w9fdQ8dEi8jF4ohapsq3h":[4,4],"8iZ1Qk38z15xMW5ATSPbb42pC7FJdFj8NtbG7uosNdXF":[64,50],"8igp2RrQ1F4drmXGpV8qNyJL25Aom31jAGJ55avPZLc7":[56,44],"8itTkbGjHRAx3cum5TD7bXaubmEFGxmKxqe6STrVqLdy":[36,0],"8jYnpEZcE9SUYPuaUXA4TMBWn57G1pPecRmT1fLssHqs":[40,0],"8m6aqa6BJj8d7aYA5s7R6mXKUM9wkZABAMYeU2V5u64b":[36,23],"8m9es585UmF8fksY5G8KDuBQZP8sywpGokqpiav6gWSM":[36,16],"8mCsw1jgZ3xKtSs842mCNxVUevrMBdA4oa5Jx5xCQSaG":[64,51],"8nVToBBSCKxiqowzvm7mKGwG8E7mzurHuRcDbf6hGwFw":[12,4],"8noQwzDhpb67yzfREDKvymKWtSdPZtbfjm3pxPYA4bag":[28,12],"8nvD1CUE48WdcmRdvbyWcM5LdJKRTNP3tXT6Qp2CSND5":[60,54],"8oCnS3KEtZGmquSW4khMCuAA8hqewT8wPPE3cxhDR1d9":[44,37],"8oRw7qpj6XgLGXYCDuNoTMCqoJnDd6A8LTpNyqApSfkA":[56,49],"8og65ngX9WuGbkzb5crHCgZdXKmC8AtFaVCPPSWTgxZJ":[8,8],"8omESudy1zEmWPdSc7RWed9jZ8EvbRWqN8A3YxWxgutv":[68,38],"8onJp7KyshoMcxVm5CemhPgGvA1hdSnHbcjLCvJidV8o":[40,27],"8p88nmvQ3uKnZtAi6poYpo28nqzzsRVXmsKEpvqCX9MG":[4,0],"8p89b2h5NHmiDs4tGnK2jGao3ZZWjhmrthyCxjacVFiZ":[56,46],"8pBBcPuSz14SSohf8BiHrBdAxgzrbA1jgtxTkwSjm28j":[68,58],"8pU43qVnsBZfVpFUnFvcNt7RL98mo79VJgGqB8nD4stG":[44,0],"8pWmLkuR3yio1Kcu1CqciTPmPMTiCf72h9n6Z1DmQNgk":[72,60],"8pXEg4dMZYwT2MhyaTgUWr1xrpEey1ArwrXcjXm5Z9wm":[36,0],"8pf1LTFXYNmvB1esMKvqLq92KZaAt2ETe3pGNxqy2pc4":[36,27],"8sJbSYEP7HtR1VGwobWNwrwFkjSMoPZU1hMkPzJoNApb":[52,31],"8sUeES7FdvfW26GirKMbU8SdMrAzQp2bPSFgeMbWMV2o":[52,31],"8t6UUXRkQTBpanRoMjxNxio1baXXkEdeLniCVJGMdzLJ":[48,37],"8tAfWTqBJiBaDfgE4cJCB9AwqcXcfoDeopm1X5QHeq7o":[52,47],"8tSzNoKE2tHYdTpCQB4apHaes2YWhCjbo7J5XCv1ULZ1":[12,4],"8tiCet13nwqVRtG1UbW5Lf4uuj33X16JnHPZssfvPXpE":[60,46],"8tk7QMWkXBbzw9AJJtLkrdf8ZnEQMiWmgXx2prk4DoQv":[52,43],"8uNgUAK2gn8Yc7eWmGFMFvwNfaKZqpj9fV8cCTkahZaX":[72,57],"8uQWXHAr8APT9fJ1bBh2XFXS4kNoWVfqZD9cjo6fUF6R":[36,28],"8uixkd8w27tuBfkRHNw6mqpmChe3QmVgNjpmNFVKKqZe":[68,55],"8uvRcrxAx5e6FRzLzobXupokSLwF31cPEkJV46LxyWuQ":[60,49],"8uxGfWm2g3sV57CbZSxz6GznKrDp4m7nZzThVz3VULmc":[44,35],"8v52QZ9KKj88NJJKMsh3t4kndqWPqkGAUb4NTz6XK2Ts":[44,30],"8whP7n58xbMDDLAAfo4rchFSS7Hu5jU4HLoo63onpzdV":[44,40],"8wzyvMnn6HSeC4EbPCV7XA6LeBddziWXjWKL52wVt7vd":[56,43],"8x3pt3B2RA7er5SCD2UZfhAusHX5UyX2hkHLLgtq5Nrw":[56,47],"8xBzUcv1AsTyjXGyWPZYLcvEe1S2gAcP4ijnaFZosRW7":[56,49],"8xsMN2rQHdmZJ4T829PAYGU6hdUivRU8c8X7jH8zNqmg":[80,71],"8yS3Zc45xptsaay9iaUSpfdb5gaKcQaKAShVvEUFKpeN":[40,29],"8ybtbfJ6rHeU49gtkQUBhAnaXBYGPdMk8dd4VCPmtbGz":[48,35],"8zH1mRkic3WDpUkSgtq1geCXXh4CLVfLrEi2TEqdTgFS":[76,53],"8zYLLHSU8URmRaAyWEY2H7uqUF63uezRCYpzFFkMG1AX":[52,32],"911tr1Hifn3z2opEsEEhxFQuJzp1YNM9QMkBQspJviWz":[52,34],"91K6thzfVGAQJZkdwEdMYDA7sWL3QJ2Bm3PRXHXkq44R":[56,47],"92ZDWNRurKikxrCQcfR9jMMYmqWksgTvSFFJ2Pa5FsMv":[68,50],"92h9nfYrehDW47mwHsjvGZQAsBikFrhhALp1XxzXo7rD":[48,42],"93E7eWXX8pVKLSrbBx13VpvDtvSU5PJs464uPoty9VeK":[52,49],"93g68j8QB4ZWAtEbvL6kfy1X6k2izXosDiuCfPPPYdjx":[36,31],"94HVaECNTwEQ8Q8w599c6BuydY5B32iG3Jem6EWXQvGH":[56,39],"94Pk8zSFvQTvrkwBkMEHzjufx53w3kX6MymDx2ayH45e":[92,78],"94VqMqQBj1WtJfyMRqsxFMWPPaAQQZ2CCUYR4UVyodbL":[44,43],"95VPY8GWPEquURxTXY49Ngv8rkb1LUaYEKyFudrWssUt":[64,53],"97vF6NK1NgmvMunNw9QL6ne9wxzUQ5RLAJqWxmDSkKuH":[36,31],"999vPueFgE7LEjk8awARTr1MVN5MMCAhaMph31EHPwfn":[8,4],"99NHmMDJeSo1AM8dg32nTokVRXByoJuA2gjDUDfiKHem":[32,32],"99PWsEpnfFaBMbW8epmC1pnRp1HrsFxASniofXNxDaQQ":[60,54],"99YEPZoX7Z69961fejw95XdJEAbHx5WVP5t8pmUFdvHC":[40,27],"99uwSf9zhnt8Co6Y1qB1y27dBVJkbWbMSUDU5Sq16XR7":[32,23],"9A7aYiEPK5ymzPjniabXofR7jyEyihGAANakUf5AALT7":[76,59],"9B3b4JvBXkRvy3XZ7xRaKVxy2aQFQtGUo5jfpVZYZcnS":[40,25],"9B4oF52Web2deG8jdbNbBuM8DjiFyLYM6CTid3dmQkQ6":[76,73],"9CCuWxTSZk4aGY8iWcEW89gTzCerQR3RCTBsMbNpbqfN":[4,0],"9CYnw2VNWfipQiDKEjgmZsh36xTmDBcSu93mCfSvMRpc":[8,0],"9CjCwpFfvex43ZrxC8iW26y34PsRbDsF3Y5fnf9iQTdR":[28,7],"9CnXcFUXEGcgHz2SHhy28ShuxYGcfYcRtoNSavUcqdUJ":[48,31],"9CpQtpHJ7UrsT6R27RECtE4dWWBAVnTcCTXj5HkbGJQC":[4,0],"9CqFnQed345m2MWXhLTaP7wzgs3uia3RCdipPeM7WyrJ":[48,1],"9DgTEERummZyV6MVSTmC8A9ZULgnN5Yh7VHjP2PADrws":[72,54],"9ExDakUNsM35KcAqwgmZVny83jqyv3SS55KwRCjt6oTB":[56,35],"9FHjVF9go3LTyZ1TiYjUTjEs9THPjULKzm7BMEB4cSud":[4,0],"9FNnRxn5uU6dVnViJeewy6FKu1AWdnknmLZC1pKqRuwy":[36,23],"9G8zRFACfB3gZAjWkgZb3CTr9KXhvEREHbaSm8Gm2mZy":[52,43],"9GERkwr654jBUn8cvDydFwnTZ6v4MZbyvp9ZKhRep3wU":[28,20],"9GMmVYJBw5Cj58P8QtXtesyQUtA9GyecPb6kCki7QSo5":[36,28],"9Gx6myZBqcVndLT5vf6pEawnqDmjJyB8SfanLTzhWjXU":[68,37],"9JvKbbmSH4T9MuHfpWmb5osoQ59dSnjXzWbS57N9r3bY":[60,44],"9KCFj7pL3hzyCzhgiy1Z9nMxT5mkNBgm2QjfbX4nXBPi":[68,52],"9KJyBBRfCt29mR21aP2NZHuyvZnf1VjSSB55WPExRgSJ":[56,52],"9LyKLKjujwPdaDWNYVuUa2eTFdyhXjp3RsfSRCvhWmxe":[36,23],"9MRUTN19MtA1matBH4ddgpS14mPAdeCoFnsLkaLxFeBQ":[44,27],"9MZY9cHJW9CbYEfVzTVmjsrnFonURh5o1rFHN56q33sn":[12,7],"9P5sFULhNktpQxEST2Wiw6zBH4aJrANCjui8k5FhwcjH":[60,47],"9Paysbs5evoh9BiWiS77NNutMCG9koUK2xyAsJm89Rfh":[76,58],"9PqR63RosK5siiSNvHtQMyEKr3CvJt1jh2qxoVmghhst":[56,41],"9Q8xe8KgzVf2tKwdXgjNaYdJwvChihmjhcHdae7c4jPb":[72,55],"9QUEoFpFLYnRPd3WwPwWkXYeLa24pDUSomoLsu5EGS92":[68,48],"9QVunAXvQbWb7Xo6ZfCWxnGwE4t1xh1dBfPc3qgRBSVV":[48,42],"9Qt4Ja8ArisFLsH2MaoPyYPqzDoiYNzGmYUThVFEKm88":[52,40],"9QxCLckBiJc783jnMvXZubK4wH86Eqqvashtrwvcsgkv":[96,68],"9RLnzRod7LWYb3nemb75vKhEBSsGqS1uHeuqh8Xuz9B2":[4,0],"9RUxQquaeSkuwb2qFqenPw63qXLypEwMwUVNaGHzDifF":[60,57],"9SHZFX3LEuL9dRpCgiETWfakZU1ZaXiw7aaeTzgkDzEJ":[32,26],"9SXpQRC2veMSkTRY1G2vLktNgc3Bbw4Nkg4xK1a1aVjH":[92,62],"9SjdDNazuohEjjoWnhQwyhzQvCnTkq6RQp3fxnezgiSb":[12,11],"9T6WSVQuo2r9b2sLQDEmyn4AaFV8XmJxASbqUiLNUMua":[40,29],"9TA34Aso9JfisCAsdqtpJ6cukxhDdqyE5xSYYvxpCM6H":[68,56],"9U4fqWRd3kcUHEX2jt1kFwF2dSXLnz9RA6B9W656Skbv":[28,15],"9US59KW8j31mxr1opP4fbg2j86b2p88DDKhcSeyDznnA":[56,29],"9UfKWtaruM2whJNqLLcrxKrSuS3VcVssdbTyNvfQCUpg":[72,66],"9V2bqR2Ts54hHnvxuwtG2yaMyTF7uscuoatWzaCUs1RG":[48,28],"9WXjR7Ea8hKt6Z84EGENQvGR3rFsovcxDYu61TJFcWJ":[60,34],"9X1qjnyb5CfMkGfEnuRZS3G58iyzbNZCp27RpiRVAiV7":[68,47],"9XsNWQxnJuFthhQ9peAhf91Am5agfy2y6SdMKvfudwzM":[4,4],"9YHpZqGdwED2uAxZbgixESvavajvuHyVZJKbVBevjitB":[60,54],"9YVpEeZf8uBoUtzCFC6SSFDDqPt16uKFubNhLvGxeUDy":[104,93],"9Yp7sEu3ecy31pKgQkCxrUWMsXiorGsCmxPG8FNwnFuN":[76,33],"9Zqcgqref1GnwwPNWcXaK88qib5hKqRMaoQ4257tvBpG":[48,35],"9aWVG4A2Kutu4tBmg9V1gaLMHSU44iuvLemPNxPVSzWk":[44,39],"9awzdQMQ1GrZJUUymUVm7SXZxfSCUDZMpWcHNGseHW1G":[64,51],"9cZua5prTSEfednQQc9RkEPpbKDCh1AwnTzv3hE1eq3i":[88,77],"9dCMmPpfNuWKyZ2D1iRstMCc1rhr2DbHVFrZ9wFncQjp":[64,55],"9dSTVY7hXEJsqExDcD8vYMAZpJ5mt9HBMPwRe94nBwny":[72,39],"9eeipv4uEyZjweLHQYGjzZqaTQradWjstQ1uW2SyuBPy":[56,50],"9g4NoCzB687Nsp3UhQEvt2Wx8Eov1hqZhnKjdxaY25fu":[56,43],"9gUMvQ8peCVhxU8ut4eyfzyTZZmvBUVDWw3s492yWNYC":[32,27],"9gmnbM2GUVXiTfCg1Pj3ZTJdqyKdS81kjBWwwnZbS4MR":[60,50],"9h2WhxhGjad6vaVc2fGztQViJ3LhYFh2MRvhLE3FgAX":[36,27],"9hedZ9TnXRLHipwYnuD8DdyvAwE7sPs9qdqNwjWvV3YD":[56,46],"9iEPYLQRdJ4FsuXm3JHagDPhYdHBh8o6muP1EM6ddB6C":[44,38],"9jAhC6dhjVqVA184dVczcBAar2GtXT7D7LwtXxLji3Re":[92,77],"9jpddNRkSJTpD5GJFXocmLsP8JUasJzpwgKrHrLtA8a3":[36,32],"9kAi6TF78NfW2gNr6n82dET61fnGU7YyYjhMRRcdEQcR":[8,0],"9kKpZomqGpNYRPa3A9o7s2SKZVeHKFCWGt3GdXxbbymR":[76,66],"9kRMWjDLCSpaWeGntjUp3mNxYApPf5cYJhvfUgJUN1iR":[52,36],"9kkpTAQfndU5SW5iVbG4j1qngoUh59Jwqndd3XpkBzzm":[48,39],"9mN1765LwF5A9iPevcJci5imXHe7kXWqQg1U2xtXP6xc":[76,68],"9me8oFZvWuc9cjBuXiW8YDGGZgkitk9GTKYTNCHaBQaF":[40,32],"9miqenD7FrGa3a4NNP6ygmYbpxtcAmW3AukuTUbAgG59":[80,36],"9mn4o462w5HzyxfnZGo7M84xsqRXL4EfJ4Ggot1Bs3Sd":[76,64],"9nwweyAkLSXutqy13N54c4DMyvgehnkoa72EiwtnBqzB":[32,31],"9oG814Uhivn77HToA3V4M755B6Sthx6aXf6jDG7Bwjh6":[48,36],"9oKMyQpMvPEHawMT1m3ryUZe624onYKhkXZ6S7aKax3Q":[88,84],"9oKrJ9iiEnCC7bewcRFbcdo4LKL2PhUEqcu8gH2eDbVM":[40,34],"9oWDUVn41kNZuVCQBr563sgbLXGvZULKuMr74w7NSkz3":[12,0],"9pHNBdibr5ukpX28foKK3UfCMeaB5GyAuGcHyJ5DmUAJ":[48,38],"9pZZWsvdWsYiWSrt13MrxCuSigDcKfBzmc58HBfoZuwn":[36,26],"9qpfxCUAPyaYPchHgRGXmNDDhPiNmJR39Lfx4A49Uh1P":[76,37],"9qrjiQG33wuqBGd9eWBevemxuw7FkY5osCxwYQt6SmhU":[40,25],"9rGfXDukY86MrUcxZNGq3nTrUaQiE917DMQ2EFW1cbDL":[48,30],"9sQUU9LhZBdYUFd7aG1NiG69sEadf8pXVhutXWL1whgM":[8,4],"9spfoHrvaHHg3VQajESkEboVGDQvk7ycBgsEPGac1RDP":[32,18],"9sttpBHogmgtBFoLZWjFsB2RZp9u6izrTQdUhBn8FHix":[24,24],"9tbzUabDi5D62Kkpd6oQs9r28Ts7TFJHLvx3pFJshZRA":[48,29],"9u3hzeHS8gtxzSEFbto5aQ2nFNuiFLtYj8SAPATiJGwQ":[44,32],"9ud25poQH48x42JefCbjH5po5Uza45MXorEDwdbyd91g":[56,54],"9v7E6oEm1V86hjTubtBon7cRYPvQriWZKHZEX6j92Po4":[64,47],"9wbAKVn7brvRaWuqeWcyBKdce7DUd9FTvjrf99xq63B6":[44,42],"9wgqMFEtHspw67xoNVnmzj8SLSisLeeBoiLEjGqqujb2":[60,50],"9xe4rcxYUe6iADdnvLkWn8K26bvyWgfrp9HYbtwR2sPs":[8,0],"9zkU8suQBdhZVax2DSGNAnyEhEzfEELvA25CJhy5uwnW":[68,60],"A1JevizjSWZwtFe2F9HujwgNZ5AbUoXLApBQcPNLGVEn":[40,28],"A1ieLrRfZyrRQ64RoGVyVQ6zqRhnQKQutm6kRGRPg6ma":[60,43],"A1voPbfnmCq8UBNQTBKnZ3Xbhs2x4cS2Gx2b2wJtqCh1":[56,44],"A1z9q4Vg8fzo5jLhrBDNqz6yE4FTz3ASSWPfcvKjTbp1":[64,46],"A2tBBzjR1zXQE9NXDzTiF4EchmLYdeMdHFKBhCRi3Ki8":[40,30],"A3QTLPL7u3tYyornA5dNkvuFLcypt6TBoB3SmsZFMcDU":[40,35],"A3WaMe47ySMTgS36KyuEWvbBX4SGU2gR5k3pFeYdUMJe":[64,56],"A3iynpHkra3TAVYmHDe3unD8343XJu7ogZn8Mhw6rEcv":[4,0],"A3zoxWHVyqHui8y3Z4rKyqWJTyr78tusgAEpAtr4ZEfg":[40,21],"A42hCCRfgx8ajv9fznoh2vQ67MVHBiYYHUUJPEPzsQaX":[40,31],"A4Bz67GutEFuHpoLLqfvqjU4PgwKkff4uNjEXXUomm6z":[56,34],"A4Kg15NX9i72WeQiH3Gp4u6QceScodz7CrkVdD2xhtws":[60,50],"A4Kh17xLBZmd3Kazu9aa2iEVa2hiMy4VTB7NfMAtmeac":[12,12],"A4xoiWbs1GmkV4p4PXkBZWM5UDfJqXx8z2sDmHP8FmG3":[48,36],"A5TXyVrR7WwfNf2RjoN1W4Dw5CuuMDiLV9e77pWhmwAP":[36,35],"A5hMwgm8QfooAuCMw9Rw2S9vXbBwCknFMhhUwKKHvYeJ":[68,63],"A6RXanjfgm9ivaGUFvjDHeSAe6BXYgJsX58UpiNF7TXe":[52,47],"A6hE8814DXNHWieGcKei3P1FhKL2rwD5YGfR89bKnihK":[36,28],"A7zCq95mtG2enn2zWxNyVDvhU2EsH8T9oWHs6jV3rtCH":[48,37],"A8Lv2ZPKKSBFiAiepFsmCBvWEBSVGzuKxSLVt9z62Bqt":[52,7],"A91g1Y8xXFEvCGg9afjTn222JDuY7iSVmSg4fdbQEB21":[76,68],"A92dCya9ivsYmzeGHzg4chen4or5WfCRcwq3btTV78iQ":[44,28],"A9CwddX4BA8AgPCmcHKAEZU4JDFRzruMFytr9oo5ZzPv":[56,45],"A9XUvhm5yKVs9Z3tYdyiAYRx9mNr2rqnv2VkY8D1N4uZ":[52,23],"A9qeyUzZoNXJQPe3fd3QgDujekiLg9Fd4VLX9UsSzAn6":[48,45],"AAmV6JwejQnHGJdUeke3hiRXch977a1PzTzFacBWximi":[12,1],"ABUhDLm3Y8HyLsmua9Xj9on87RyiEsw5j5eVVZQVw1hT":[64,42],"ACPgwKgncgFAm8goFj4dJ5e5mcH3tRy646f7zYPaWEzc":[44,35],"ACYDnrdasgqavXoMR2pDxeZxTBxG9RS4evfhy9G5PsCe":[84,67],"ACv5dTk7THbmUpHYGhgPzMhWr7oqHSkuPJpa5RfvmG5H":[76,58],"ADVqUcnmGF2Jkm3rVkhDbkNxiPUSSTvZC5GdSza81xSt":[4,4],"ADiT4zpCRryJ6NGtvErT5dtuFzTxwYRv24fj4b4LDQDr":[68,60],"ADriSmPTSeyKwNCo3geTcAY31G94mHmCfRfrJMe3DmbV":[44,36],"AEPNDgaApdcfZEZpww458Az9i2NZrwxsVCdiUih4EaRh":[48,39],"AEWoxb4i4qGP57iJqNyubSA4frWN51oJ7pQ6634skR45":[4,4],"AF3h2gdkGYndVj8W9qQN8jA45kQ5RB2WmoAQN2iBk37c":[28,24],"AFVkDuKCb9V4TGYNYK7H9PT3sj7Ny3DCRumQby5UHBCs":[68,56],"AFkpm8QAMCLbNebefoZsMerbWNAKCkXLFrxeCj2DiRAn":[44,38],"AGCsyz64NLvoDAG7Mi7k3WFbkMjRDCv158Q99WGGvKNM":[36,24],"AJEi7F8fQWAP1xarPkFc6XJWTi7qvPRcZ4JhLSG3CZo9":[72,58],"AJLWyfaLmio6Gm8GA76mdfUtuLT4tmyZsLhbpXsndNab":[28,27],"AJeuXG12CxqTQTnxseejE6PcMapybJejMYHpFc6SErMW":[76,65],"AJyQWpskWfNYu5wdZF8zqmNHgHpi4n6nAEdMNi1SqhYo":[60,45],"AK5hfHFusiS2y5cjqZkiyUyAvH5qfidQgrmCccENnet5":[56,43],"AK7ZZx2sdo39coZN5FsPdae2xNGqVKHX2TWJixmY4ecX":[36,24],"AKdhJ2gVzrMky12QF2j5F79K1F21znpGdFCQdY8M7mi6":[52,39],"AKm41uMcqEUerYPuW5jq1hoW8ZrrvwPWYHJDU2QmrNmp":[52,36],"AKqB1VaJhf2Jsod2ciEjzdTzCUfgh1kUUUaC1sQ7iMnG":[40,29],"ALPXVb1A7C8EkR7NuKy16pXcBRasdeRNmRPnWGQHpe7j":[44,39],"ALzqkbSgVaQz9nn5xh1BtEsey57otKRyGmaSLhwphYSn":[48,39],"AM6BNu2WZibZhYYHNo9ZWxmEAB7PhjNQBGKAhhN2VrFt":[36,26],"AMHwC159us6bok6awd7jdvjFVMY5ewk24JVpZNwK7Dno":[60,55],"AP2ZiF3mdoDsVd9AdzJWRUb5UHHoQjnLLPXn8rrkERGc":[60,51],"APjVTcfzJSzYEkBddGFN1mtFWb8jDzogpgUz99tQW3Ei":[64,48],"AQXUYRhH2meRQtdNiU7bpGSqSPJEbGKC2LaRx9QodhR7":[56,35],"AR5Lgk9sgoz69qGBeeTiMyyxZdhvCi2qkD3XUzre1Uvh":[60,46],"AS3rwVs9WR8HTzN7GA4aLBs3JjWjt1yKHfSzmwoqp2bY":[52,40],"ASDE3uRDLHUQ3rkfeGiThetAyU1bY9UssFv9PeAZDtVi":[8,8],"AT2N17bBBtTAu6ombzhiLNLc8JinjMXmGMzFbxt6AvwC":[72,53],"ATra44iNoKxAxj8zpWfE8oALhdyTZY2AA919CyoQ9bJW":[68,57],"ATtMQ3Aphrc9X2SeTJD9JWbcDJfn5aJs4NHWSknxfNb7":[88,65],"AUA4sQRvzwWiL2DTEk6aCaAqZpKJ2j7RGMsJthP5aa6y":[56,46],"AUcXRTirGAjzepCXnyL4UyuBBxkYMsUThvcei5M7x3fp":[4,4],"AV3nnAu6xyF7GcYRnaaPkWG1EBVtQwGUfFyd73BTrDxK":[72,45],"AVAMqDmPX4qjDZYc71Hdh2ZtjhGrGsT1yv86hAFVNt9u":[72,51],"AW3MLxDTfigYizfigb211N2BZKa2epePVJZ39ChxpEEx":[56,37],"AWTWtog1GtBcjUGuGVY4zpp9xrRm3aepsRi7P8EufjzJ":[52,34],"AXJXq3q6JvQb192nBU2ZQYJBznrc5ucdvhpQRoAfCCkw":[72,55],"AXfAhKbu2urtzCMLgPm4vDWwjK1hEs5ZaqMCbfEEyn2P":[48,33],"AXgJTkDM9AWW62Th6XUD3L1Emdxot2PLRffFSrsSamat":[60,53],"AYEHTBfsPvdGxkCnrMHEu1nTziUJB8Qnhjktph5aQvrw":[40,38],"AYxCoguM1XJXcd5e1bYVQB3Tdtu3vnT5iubjxgwvzNK6":[56,35],"AZBWhbBeVwpAJNFRK6fK72Lap4Yw4vhz6LKpYEQzQCrE":[40,34],"AZFRS7MsjYLd1oYAqeZjxyxXyG53zNYLQtKRk1qVKgG4":[44,31],"AZFkNiUSszcpsTSAmCWFTcLPe7iQf6sGp4ceV72JiCdt":[48,20],"AZY3mmLS39SKps93TrsZoC9nT1nUrpYLUmQWzbtgyF4t":[52,34],"AbF88hkkpZ28VaT3vYn4xu5CeNC8G6Dq9cc8ciRR4fY5":[60,58],"AbnagVJhwwM4wDuZbvoxWeofdpSWoDMhcmZCdCrxtCkN":[56,40],"AcjhWohnu7vYMdu4Yha63XZupqMKVVnrWmt1F57ScXhG":[48,35],"AdVKEVMZSd6VZ53PYbw3PSaj4XzDjsNoEg9LwDnyWRE8":[64,55],"AdrwFufQPWrBRAWPG2ferUA7Bi1EY7SeT62DRkaGmt3J":[36,26],"Aex3fnsTWQF8xf5rEf6bZawaEKsqED79qx7zJKzAR4qb":[44,38],"Af3BY8yRnmGLSX3XZsWoCS4UrCravpZuVY1UofrE28sS":[80,12],"AgFQkQe2Em2GUkDD85qPmHrvybnaXKMa7anSNdCunnM4":[44,39],"AgcvBSS97jBoKY2x1LXrqScziFx1jpCzdE2UpgSiVeQr":[68,57],"AhT4yWiSg7nnEWQokWoGDz9QPURwa9sEHrPkidC2PK26":[12,12],"AiPN5MwTHxRjG4eTQ1nrmxERRj4oXJURHPiTcNpVYcmk":[44,39],"AiWqv1dqsbvkUMec7G4DmM88ka7SaoqDPkn5U2iuvqnn":[76,60],"AicKhNhJmkdqafRDjKLPgVqLzXLzJ8pS6aVrYrRkq1iq":[52,39],"AiiS7TxGeSQcB3MgBzhfzdAXL88feL8ibxNz6t4QXBRr":[32,30],"AjTrfjYY2SiTC5pLJXwNXpcP8q549YQ9VrPAxzjqjUaW":[32,27],"AjpiWM8sXfKy9bH7Ww2qf8stQBY3Hk9AHixaGMYVN7eY":[40,28],"Ak4BNgorzDrbQSUTuxc42hb2rkZt7SY533b1HrA5U3Zq":[32,30],"AkVMbJq8pqKEe87uFaxjmt35tX2cNhUJTJwv13iioHu7":[144,97],"AkkJv1meyo2Ax2XTXEXWpvHTh4F8a68Lja5dx3TaX47K":[64,0],"Akqc2WGCzgLNEvzgTfxrUoZTau1rBgPrs6XDaitHoyR6":[52,44],"AmUY64jsSNtnhz6cXNRD19jmEBYq18B5naJhoKSU41oG":[64,60],"AmZ1nodB6yK5BVbbzVcp7AveuegouTbkfnUvMjeXbDBR":[40,26],"AmqVUpveo37Bt2VAgaQjTj9hw6APN9z7DMHK7X8CgjmL":[60,43],"AmtSHsetLtT12qsPtmTbKURWLSM5A5kipPwe8rLoeQUR":[12,4],"AnCW47d9J5V8W9pGNhagcHiccH7mLYcZR6vXbzTbPn9Q":[64,55],"AnbcuyWTbuDwHK2URBB28kHvhns9S3euEmKtY5gQ8J22":[64,49],"AoKD8QZ5WVvY82UgydBCSDZ5itUmyWeZHmu51KddCvFT":[32,17],"Aom2EwxRjtcCZBDwqvaZiEZDPgmw4AsPQGVrsLa2srCg":[64,54],"Ap7K9JT4WA59s2cukCpaZKZryVUwiFJ2g6793ZgeJqDb":[40,39],"Aq8yWGbM9uA25KDKsU9KPwoPEuquP5vTqrYomh8VK9XL":[76,61],"Aqh2c1x2AA59pek7pz8PymDXzq62qmNiQ5GXhpWq3rNr":[48,42],"AqwRcpXAYMPGSJuZNVKjuduPTGgaesRQ9XJWG4cbzgT3":[24,19],"ArVspBqfajnC23uoQmzgex2ge6NKRzFC6BVDJ9DY5qJX":[44,30],"AsVqSKFD1akXQwL53qiS2JxiQM8xabP1acrWx4SGycoP":[48,30],"AsboE8YchTBGaWdpP6nSe39g39VTpdHKefNWrbPdyLQF":[48,29],"AsdGYpcPVh2BScJYMaNuGaHJWBhJjEn2CPFvj1CpF7Pd":[36,35],"AtUD8QwodrRPdHXhEH8ZUkXkSZwe6eVaQbsfJYCarcLs":[36,28],"AuX8i2wxd4qQpiLcie3eGyWPbVeh28dx2yasJiTbJPNC":[44,0],"AudhVa1DskDfMAFTuF6EC9vDhA4QnSCe9JtJeEg81KXH":[56,50],"AussSM225GLYAGE6wDBoaSnAsSu9tpHqa1c9FG3PQVtL":[44,24],"AvKQ7X3BL4FoBr44VKNtbMtCMfHuwPb5ZUS7PLfYgVbm":[32,31],"AvPbRdiNN5nMJtPgRxihyw1nsL3prEgVdigGRYiUoEGi":[24,19],"AwMRTUyLVeee4eR7ZdPaj4s7FHgTvYSNmD6CacnTAbiB":[36,31],"Ay9TTZQtpbTohcrb3eq6doZAeeLHXmKgMiJeiUumC586":[40,21],"AyXMWbdxpvDoeJmueCBA3B4w9VURpiQu6pbjrwM2z3kR":[12,4],"Ayer8NhVD5xUyWkfKK2bMi9wmhX91QUoJ9kjy3xh9aPy":[52,37],"AyvS9yc8cuHM43EekkAd3kx25iGZvq9axPhHPvzre2Ym":[60,43],"B1Yd287CKFxZnZXMvNjbM9V61kjW1agupihzR2xoMWBt":[4,0],"B2UcYy4WiS1fSYKbMPeAKZoCEzgfQKKt5QBAA8NXLvpZ":[96,60],"B3ZyMGnSX5GjhKtopCtkn2jmmy9g5j3KdwcvXWR4dALk":[44,37],"B3fuLaQ9orHBEkeGL95m2oKcZZQgwm2uRaxVcaAJpcqm":[56,43],"B53tbis1864ZZqg9BBdXFAkDponmVvKHwHySAY2tC8gy":[60,56],"B6ZramCQhQWcq4Vxo3H5y1MYvo37wRZiAwKk88qaYNiF":[68,66],"B6eeWqfF19AGj2HtEk6jzSPEvpnMZjTvbyh3d7HzRBeH":[48,40],"B75YnUyemn7ixtnUtq4cDUVKrFwQmn8J2Er85ypcEJ1c":[44,27],"B7QNbMjAsaZDvNLVaBXo6Z4Wg7tKcESqPY9tQrSefvBy":[60,54],"B8T2dbvYaM4gJfpLLbuBScuWKFhnG6KspRK72v5D5uoK":[52,40],"B8UmpsNTU2RZ49BjTwLv9e2AnzfR3Gz7vBvNBWRWfWnE":[56,30],"B8wQDRb5JLuXjJhAtmY1MAQtLjWQySberTN7wLUHmP2B":[40,31],"B9gJJ4vMLJvnb5geZjU9PqhkyHX4jESMYajfcALQgRry":[48,42],"BBvV71ncxGrMDjmKTkcvcjipzu3bv6ExyVPPuRxAPtrC":[60,49],"BDKAGYg5SLDNRz85TZj4DeDekKVLxMm5kRNJf55WgJcp":[68,40],"BE9viFvKLzRqoc7eKyhTZ57WdTWUys5KekNXq54MYybW":[76,0],"BEFD2nMciBXpk43V8LPQ5D8NAedUzTswMD7JrXjQpQBP":[64,50],"BEq8K2LHtQdNGGURvpUxbbcutFHWQ1YATAvujMz5ju51":[44,34],"BErLhip6XE7Z3XT1p7EACrkwCXPcXKa6pCJo9eFPYh2U":[44,36],"BF1f26A9FdL6uWSajjTfstnLdCpynXGVrAEzqUyKXJKd":[68,62],"BFPdda76UtZM2grLc3NZwxUpP9YAhBmCb2dT9455ei5w":[56,43],"BFZpksditzAhQbBrLXcbBEPNLz8ifVBG9RZfvbTN77sF":[40,30],"BH6QqMa6nqyWP2iTTxqAGawP3E3n9vG8tb3TAPKAs5N6":[56,47],"BHKvJsTQHub7Y8KYuiSfxNN3ztbNJ8LALaRgAHH8Qd5p":[100,71],"BHPjYib5bUmwXAXa1T1UT79eVNmC6QgKkmbBxipKJxkK":[88,49],"BHR2K2tpc1fowNyUf4PfAumc2tfaT2SpvQVqmmpuN6tF":[36,34],"BJBxMx5dEb38YHLooaya4Q3w3s42a8HgHLocDLtcUxhL":[68,43],"BJDPVva3kGqpwRnsPFNw7pJgwdaVwqLTFLgbeuBWiiMa":[28,15],"BJMh3mPmJqwzvXVHgCqGpgJ8o6hAGJThvW2BdcwRbs1g":[64,52],"BJhh7JzBaSagZidw4Fmko6SVqkmMgazfskP3qjeciVFL":[56,50],"BKAgnBWgAMtC5NaP5uS4Vq3WZme6dAbUmEAJmyronddd":[68,67],"BKQq7feS56yp1PvAcBQjb1zV2XYtASm8EGTLy7eGq3bN":[72,57],"BLWxzv9mGX3DLam8z59A55qDpF9KyMEt8krFf88Sacm8":[72,60],"BLZtwHMTMgnZJdhJQxQaksgJgteXFFDBrA13ywWagji4":[64,42],"BLfpk2WoF8RnerCqjxHe7qFJjpEE427riMkf2fS6vUB6":[48,31],"BMeb32DosmmuepncJGK3yrGvpd53FYgsYgZZjz81mM6v":[64,55],"BNa6rYkNdwWpXSzPTL8yf8tiCZ4DqzT5orRh2j7GuWB9":[8,4],"BP7tiUE6JHuR3LRVy6bb79fLL9wfu92rkM6xy64rHu5q":[36,30],"BQ7mx4ScgjetA378LnL1Nm3xiM9bbLuEsX7UKxPseRCU":[52,42],"BQmxWxDnbLEQ3Pr9upNnaeiMV88K77JiXVUNoSHtYjPB":[36,27],"BRBQTuXWHFDvVTEm9XK34MXgp8QpHXxfrq5SLGMNpCnj":[48,41],"BSF2yD9mqzaixDaLEraF1en82EWaXx7wbaCqSuKppqG5":[80,66],"BSRAkuJ2ubtPESMHAyMrhSroNxTSJQTSV7YLdPw3FNau":[40,35],"BSzyD7UGHVbgUsq2yMKyXKysvc94Njudc273pcboJM9W":[72,59],"BT5bANJXEmnacdBHiVWCMGWckJEUBU7VRiVMiLmA65JN":[68,64],"BTFGAGpCMsqi9XbsU4CmCP9pbizVBTzaJxeBVmZAKg8Y":[48,29],"BTPUdVrsgfKFPyBmy5ozHvzMk1QCK9vii79wgxtGhjgn":[68,50],"BTVmKrqHyQU5vSqExPNYeozop6JPMCgCCZkVfCMKEa6w":[48,38],"BU8np8WoqP7cipMKWdLFoAe61yomgpvAEjFpFjoxfJhZ":[52,48],"BUn4mki9CktX2a2SMuxwei7n8ttvUy7uYJadbGDheAZH":[48,33],"BVEX3B7fRUbadEcigqknwc3cM2CUXpyz9vtTccBpwt7r":[52,46],"BVLVnUm2tkgX1sK2f5oVtTMR1opGra72s4qBD9LjMd9Q":[72,62],"BVSbFiadwxwi7RajrGgx4KAuwY7f3s6sB5AaL9pGZ4Hh":[48,40],"BWiftESMUsve87rkjU7HsaA7fkiJRAbv3xZLQrmKZtnz":[28,22],"BXWqrL3ZuDU3fJ2qbBGmmMagLBLgWv65YbXxqEat3Ud6":[44,32],"BYRUnuotppozq3YZp5EcztSSiUbYYg9hLasvSXv3FBir":[48,35],"BYn99WhTKSAuZgsUN1vmnPevVhn5i63cHMcd4gLjCxL4":[48,35],"BZCeZyvroBrSq2SbwjQKU41kcxkutUti1A7rzZqCfaHK":[68,47],"BZNtLRLmFaXi2jjcNGcMkUH35UJVEsv7GnJ4352x2WmH":[56,41],"BZehsFmSj4iUhwxHNqJVzduQQiaLqhzfE5mNT92UZf55":[44,21],"BbAyKMQ2vFmamjNbDREATh4SqStium9q5ZXZ2PMJA4Ha":[48,35],"BbYNpVvXLyDohtsNx52JB6yUsuTtrdC8a3PRALXTpD8t":[48,34],"Bbe9EKucmRtJr2J4dd5Eb5ybQmY7Fm7jYxKXxmmkLFsu":[44,31],"BdVXDJ15krsj829E7Gwou8McVYmVKTsPX3EcoBWPoHmM":[72,35],"Bdd4XhquueXBB7aZXVYUn1XBdJ18G7Wx3LUe6aKkmXEV":[56,39],"BebUNmLyM62d4BgE8N88YsJPygWrCWSNaCeq5s2U8uzC":[40,23],"BerdkMBXBVjUoKUkAuRn4DbadZpxFCB5mSBDadr8GErq":[48,37],"BfFiJgrPfecVSMTn1ma9UWMbqcFMftrxzgVp63TFWvV9":[48,42],"BfjmopwTknigm38Rj2synkw7mNTjgmm6hsCCb1hQetAK":[36,29],"Bg5j1Qfa2SEmio9U3UQKMtERNEjpqvushEmJMAi3SG4W":[56,47],"Bgp1oskwwj8mQe94U1Qn59BLEbqRDk1qRgD54bz2fTcr":[60,42],"Bh8jsbGHJ8WTxdSYqanh2Mv16pHjaCGfH2PKTnGKPMsh":[52,35],"Bha4mjHBAS1yFjvjPWWY7ht3jMneu4Lezq219pdU9dFu":[36,29],"BhbARoxdh2MT3vb4awXraZFPzSwBdmF9pGgURKNsjBqC":[60,51],"Bi5ys8HXXthqy7H5DMvdYfWHXLMWD3Aj3v7EcdXNtEYs":[60,52],"BkDggfkx9Lai8LvA6iu5ke6jTwCkQdwJwafbifTNpQPy":[64,46],"BkTtw74AC3rDKUbFboQaRVnhLEhsUrchotzSvuweaUCH":[52,28],"BmCFZq2tQ3zj3qY1pjK8iegUp2TAHj6cYPM2vJkSA84d":[64,42],"BmytjrTNhPLdJg1TUmtt6JtapA66vWgEm2CQFsG6Q7M5":[64,49],"BnBHhF7VA6vgyGK5goSyc8sdLgWR2DcaNNRr7JMuyGUa":[48,43],"BncRYtvkRJ2TaZ7ud6VgpWY17JcEd6Hk1GoFGf8mua7U":[68,64],"BnqoTc2VgseyBFGVuGbNZnDfrPUCyam2CrTtVEEG7CXp":[44,36],"Bo9T1z62GVKmnttMz4HxPPtRXs2BUkAd7T7yUsKyG4iA":[64,57],"BoZviedWdjsBkTFM7os4RuMHLJuS84qask3Bn7zVDWFK":[40,35],"BpXtvGhGvcfLEFZvTAY7nvHLyKY4YyEoL279KLfYyzAx":[32,24],"Bpq5BM15n4ps9zftpAiJARqVmAfUsmjSKfMQN7yEZARe":[32,24],"BpsPu22aaozdVrZu8kcqs7rrURKPH3fBAPnMF3oLG2Ea":[12,0],"Bq46YaSk3ydeR9y6yvj9887DtRG4iJGyfkZMt4PB6pHB":[72,53],"BqGCBrgYpLv62ebUp7DKfnjvSJ2qBc783kehzKJDERbv":[72,62],"BrgWTUdNh6J9YJyQqWViNSNkJ3wLw2KpVMa4qYmx9XpB":[52,46],"Brjpw4BpNaAw8jvaXmVEBpGYzKfAQi4HdtshD1a1KDLt":[52,32],"BrmBMYWThXPvWmRKt89FsZScadEGt7fy1FRV2QUArpwL":[56,43],"Bs7HtZ3zTNNKtXHXMgpzzkwXQd4ao3jtgbKwjmyEBzyW":[32,23],"BsGcvcCxwNGpcBzY36oCMno6mPVofJK1FuGpv1oLmFTe":[28,12],"Bszp6hDL19ymPZ8efp9venQYb4ae2rRmEtVp4aG6k8nx":[36,32],"BtHbnzt6SDwyKpzQFUS4ReNEGrL7YYwhuRoMdV1zM1gZ":[44,37],"BtNTPnJo2YWQhiaSQtNnCeTz7XDWUARRLSshe1smVmx5":[36,31],"BtfSHF7UneDPFXJdxsiheqEH2q6RX3ZShn7K79fzNr1B":[52,44],"BuFKun14y2uagDQXKou6x4ArWuG46US7X8TEEJVvSMTq":[32,27],"BuiUZY8d14fa5ZTueKiMXRjA8QtWAoP8XDoJXsixUob5":[40,39],"BunxTHgkSEyHHMikCe9ofDB5dsgcXKN6nqKC5WQsd1op":[56,50],"BuvYmqQuvqpNeBrPao4GKQyEVjV7sezT7M2hxb4rRLoX":[64,57],"BvH6ncyZow5NL95LR8MRExfNUib1xKQKSXGC65Ypae1T":[8,0],"BvJSLTtfFz2Qt1MLDJqfEtMw8uGtiB99nthg2aWGXBPP":[64,51],"BvVKJxCQNsFWpB5o1B6To4uZzqAUgNXLcfbS2inyL5XU":[44,39],"BwQPpmy9TF2KqcYmLNbFDGnDLBwdZhKeRnQDu5j9VnYq":[40,31],"BwxUhPBXUAEVEWDk2Co9xbTiiSFAgVieoRiPFNGbjBcY":[48,41],"BxFCg1xrZukbRHhthKXRY1gZVnHvu1xntdg1YmYnqeE7":[44,38],"BxLUkNxbwARzCsroVyFSniLBx9FirpQJviXUiB1ZpBXQ":[44,25],"By8BBocWV2yLsEMfkfHk7JkCmg3wjh8hPyKF4kd5nTrZ":[40,31],"BzoPEBru5un2Weex2nm7zkJQrjFrAUA7AudNBtEDe6hs":[8,4],"C1Ag2mUyZnLcd1o1a5xDWX8XZ8YG4nVXZpehcrxXa42x":[80,47],"C1QUyFjgVeG2mNBHErtmCLz8BUqS38saMUz8KA8r7921":[48,47],"C3XctAkQw2CaBuhV2k7Q2hEcU8ipsak9YsXz5qjAK98s":[8,4],"C3hD8Q7dLoodUm6E6LTWR3XqJgcRrvrVaMscwMBV6vaV":[4,4],"C4N1bMSzbfDwHGMitxyufNZPaAkNYx8vJxHRnHWrptAT":[64,55],"C4eTa4tqvvzpTsp9pa5NKAbeDXJs6sHWS5BfxGB44Xex":[40,32],"C4xEQLjRRKLRMPvZhAakT2D53jaikWsVqQ8dvsqbeZE6":[64,53],"C5EZSW7Mzt5bktvVViVqnY3H7Ujo2ReE4bJEtdqcnrQw":[76,58],"C5HvMeXdHGxi7nVTFPF6KcyK77RSWLLvEEB3ParXoK1F":[16,16],"C5L2t9gSXSSEcGXdbr2LBjb2p2R3HpbrByU9uMUsX1y7":[80,74],"C5VDTdJWA1ck6bPiX7d8CurGTfG45zpWdTU5G2y2deSG":[56,50],"C65X86Rw9BRNGK8TnzcZeozKgH5CwZFmT8aKPe9wyTy3":[48,39],"C66yBCeSs2vyHjuvgjQip8fKZxJtNacYmA1dxmwoWaWY":[36,30],"C7wxxkKVpvbEeHdpNrEa5xN5dAergj4h3jjcXunKJZo4":[40,28],"C8X3QTF52FhTXLg3yArhzgT9FBuEdxZN7aCUnPUprsxT":[40,31],"C8gk6Z2dm2aS8dNQEXDBbuSenoxd8HSWxDSMNM1u4Ebf":[44,26],"C8x8gRPxVQd1rk9VG7fm5MqtbPkTF7C9R7NUvb8HJ6xo":[36,34],"C8yQnRcUMCvP7ZpkbhPUx2TXspyojKNWttdVWj77kAWg":[52,42],"C9UahsjNtQao74K3zYdJdkGrhfcn7Rf1szmMtUP6fSRf":[12,0],"C9pe3PrMgXLaKUeEUjMFt7JFnnJ7ygFXCjzYFenmtqJu":[64,25],"CBnwmAk9KrFy4cvh7dCB8T9KbX1LG1v9ZxQ7nzFR5p8Z":[52,40],"CCM98AN1SENGvAF6mNvYsCvQ8SPbommtAdxSLdYwdt39":[60,53],"CEMuwgTq1TXoTvdFjuMYfTu8Rnvo8HVUbKGquAsiLCXs":[40,26],"CFtGf5wQ7jPgJVSk4GiVxvqVZXfkpxzdnkFJGduUKA88":[4,0],"CG7zvuaN3PTuQN9tFNoE5jERxtYwg8YVuKE5CMYD2jp1":[36,23],"CGYJpRhizVqqEyr3m7Ng9ghVpFRhBdD4sGjXSvvTFeze":[44,36],"CGgiEmA5whBdjKKyJVgFEBe2Z2qDVQQd2rMvAaUJP6Yt":[32,30],"CHFbPXHspWQfRGmxSPh9gkwWviUEdtko39MT6L2Ucy2Y":[76,57],"CHpj822jTX22VcSqzksxkJLB8kBf5gDMCqYbgXv36dvN":[16,10],"CK9yiW9cCVkJGs9qB2SZnXUJ9Q5btmrGWp6KuomttDXo":[20,8],"CKs5FjmJ8qx2o5gzCJukb9Q6Z4TEJ7ogJjuA1Fch4bwA":[64,29],"CLq7ip77pM7Mf69QVh1PorC9gdvzVZ2QXCi7B8oowU5m":[44,31],"CM1c6z3pRNgHFcfZG4z3wE31jaR8c4gCYQBJVEoCUyq8":[56,45],"CMFZtuwCGnXbnARnNx9JrrAXhGHjbGEis9ajFwYPGqCs":[40,0],"CNVw7suEhz3LJFDzN1sjin1MScbjRPWB9yZ3tNT5QrB6":[56,28],"CNdVHTP3ZBganD8qHMvznWeZTXh9KfMB2VHqdfdvYZBL":[24,23],"CNm4mEYYFUqGD1WtvdNZv5iZvVXNcG9rnqodf3w9xkNk":[52,47],"CPPVEbGFbX3XAThetvfveCE1vYLWUwwJGT7DxkPAWb8D":[28,27],"CPZDTXDMmE8pHwxhECEfhWZYNvDNwjyFx2GJ3pSbzBX3":[44,35],"CPi7yFjqm8MiLFJpdyfWowAgQex4DjJxxHcLa2rYZ2XZ":[56,51],"CQtuaJNBz6fh7Anvy4YCqkcadsEyk1y7q9oH6497WXS5":[64,40],"CRCQfYWcQNB3fT23T4wnUiLwneaZ7cySmp3jhPCmmA9Y":[44,42],"CS1Q8yNkw6a8SmY4nJ1jKrqhaDo18Wr4CnNbwsvKoswC":[44,35],"CSX1ZVkdseSUecuAtuTPyPPadWcUuTodE9BygHjMhKmU":[8,8],"CSXcNyNFNL7ymcMowKUByDi6DHCH9oVwdJgbgZGrvJcE":[40,35],"CSY6zb4eK9EVKvWkid5aB5K6uc5L5An7UeY6VwRGXbtC":[60,48],"CSj8o74gANwA4V43QsgpGHhoRT9KarDvQsecBSCkXeVf":[60,40],"CTAqarrLTZ58oAtRVo3jrQYqcuGke1XDkAMqAyK3Yfej":[40,28],"CTinHYqJWMLRstmkfH2mgPRdDnQu5JuFfQucZfMuiisK":[64,57],"CV3F19YAhoW7DpfHQ5W9t2Zomb9h21NRi8k6hCA36Sk6":[68,51],"CV4CmE9CWVp3sRbSVR1m1CUEqQ7pRdK1p3LEhirYF68L":[44,28],"CW6eHayhnTCN486PtdnoPGAxmUKGJ9MooqhpoWjQdhX7":[32,15],"CWL6skWfKLDd6SY7NnkjfMgNR1QxHhxCadyFNL1ssNaS":[64,56],"CWSr3rjCEQQvzNCQqWG4iy1XqCRhCPH3LUHKZpvnpzfo":[20,0],"CXCY8Pzu8rDGrWsp8yuA41z1tLv5ex8NJmRGbfu3gbAz":[68,50],"CXvTFeiDYsmHfUXyYgm5byEwM4T9T1Pcu5qRiTfKdsho":[36,24],"CZEj5m7Q92B2GTYp8wYhz4vXzYu9Vr33yu3y8bsXBey":[52,45],"CZWpCTN4rCWer8fm5ZqFdx82CDiCJjZLKZ5Ti2gdmchQ":[44,39],"CZY1ZJAUyD2ZfHE5ENChUmhqSVFwPnTm6Aq6N5tbBqaP":[32,32],"Ca8DQQagVHeUAhWPWxGCmaMuccr6aGsm9HxeedxUKBC7":[60,0],"CaCaErMi1TtrTLLv9jaM9VVG95Tva2keMe8BZSuji3D8":[60,59],"CaT9dSx37Quj1kcAXEVd6ncM6NLvYUSqhtgEnn1JtNKC":[44,31],"CahiUdyge1w7Z4GrsXNpVYamxj7pJoY7uSThHC1LCBPE":[40,28],"CbaT7xreGMFW7sQV9mtXRNsZwen7SCvK9BwLJjiJgZMf":[64,53],"CbhvvtosVdVwZ8GVrBqgYT3JrXLh8JRqgKpimhnZw31h":[72,59],"CeTirCDrYzkRMBhn4P8WYQen8Ka719PvPnAJqbDPm6Z7":[4,4],"CenFrJAktGiEtJEGZXEtyN96piKvy9EuAmD6AAugTTr7":[68,55],"Cg6ce8stTTtFP41k9FwYUWn2TEv9uq5Cb56Jc1p9XZnU":[8,0],"ChLMXZ4KXsMpa8W1VymMxim1vdPSK5a1jDwfMbm7cycT":[28,14],"ChpV3NPQ4nHdnynwvAUD2PK9X7UGVYRKvUm3wAR8T8BW":[40,34],"CiY5RjWPs1XyegKyBLcG7Ue7YMf98eiEnmvqnSuSKbob":[60,47],"CidJBbVgPwrVFK4Je2Wxz5zuiC5sxcETbRdhBwwxZVvA":[68,42],"CjQzg8D1Z2GNtRoZAFNc8kKyghGvr7NSNKW9hCXHrAMq":[4,4],"Ck3SxoXUShtXfLKfUUXAtCwrFVsEohESJfWGWuSgtTQU":[68,60],"CkEhoG5YKHt6szYHD84r4x9r7VHWN31aXaXYLbse5oYJ":[120,76],"Cmg9ZbuT5pR8o3CBLo4iwHCMxWzd21ZyNAoLH1sAwHxh":[60,52],"Cn5H2oxjXemT13eeFU45gobRYiJrjCrhGaqKTMd66SZM":[40,25],"Cnw2PuZHpJjpLd1ZxxPxetuLiXHniZjjYMGhQpJYqRBU":[72,63],"CoCKdrHVE1bjMZDwP8Z16vd7U3E5tGyt1hLw4tTsysmU":[28,24],"Coh6CJGtbgZDzESkahjCtN7CJsfRG8fVqHn5UxpzFniN":[20,16],"Cowx6w6oyFdnkhVUBsseo3RbGZGMLv13SH6J9Bo3J9VH":[12,12],"CpGXhkQci54bs1vRaJW4EsEbaH9viTZndMFHeoxcP9Sh":[48,40],"CqZVJ4n9EnLR16ed3NTf4UdmKy1gTUBdQUkz4HGiqiiP":[56,45],"CroXW5BHAubsisQRSZVcjdRE7vypucF3GNWMhesipsZu":[72,53],"CsaAGRau3ZvyMQvJ9CWSqbqeVv9zw2Am8FhnL9sr6jTk":[40,38],"CsiPaAm9Zr9EpGojSNjQ92h1kcKQvYubiKmpWb3C8B5w":[44,39],"CtgcuRPSMc1Y6BVyUfj29abjHTD936sMoKFET7sH1qtz":[52,29],"CthYXhfQPZDce1mZ7Jy3PA4WyF9frjqKpvxBgQ4XtPhQ":[48,48],"CuJecFtqRJmuyznjtUwN6cbPm6jiae2Ka92Q2wTsWHiM":[60,51],"CvbC65nNU7WjvgnTd9nPKyy7p6KE5Hd9YwQf41eMN2dH":[64,57],"CxPhLbubp2QN2WRatESeN8YH8r1PkNtSC8mf4x8jirqW":[84,56],"CxWS4R7rdKxdrFNhhWbPkcqbD3LzDXLr1LSvxpDrU6TV":[44,33],"CxkCs6KjydWBdMcpffxrYAiZtXpkjBW7V2Xy7JkgmxVL":[68,56],"D17iHRzwBk5NFzAEiUb5JqhaqDUM269utRqduzMxcTT7":[36,31],"D1SCYXUr2jHsFnZDEw67znD43kXkNU1JrYo6o49NhCnz":[48,39],"D23NCAVxinE53BTemguZCheAqCdMGfNTUzWdoWvq4Xj":[56,34],"D2CEZnBYQ9huwnhCYKr4W6QVdakwQRtWVrz7fZ7HxSWL":[56,39],"D2NjDkcv8Y1dWGdtWAKPT4em2D3sYzM8AzMTpCG1RVf7":[48,30],"D2ULkLgZk1d6RW3Wmd14vFNfkBgi6NMM8CDNsNuNXvfV":[64,47],"D38N1Rjq3Aw6QvhER2CeHCELkCsgcUkyPkP3FbRxp3F8":[40,30],"D3eokjT3EZnEsdqjgRdMv4QVQvjQD9u13ZyqoFX68eMS":[68,39],"D4pLf3e7kDGC4yc156Mb9A7JSAYnFH63jjsvAkfguqZB":[48,43],"D52Q6Ap8RVMw1EvJYTdEABP6M5SPg98aToMcqw7KVLD9":[204,145],"D6Q3V4ZXa8eXuy27WkvumLPN2HVrLhqA9tf1VrRVa1Xm":[48,34],"D6beCFAZeFtXoZKio6JZV1GUmJ99Nz4XhtxMePFvuJWN":[80,58],"D6pwr3woPNRpnDrtzBEvD789wtx1imWvvpD11T99C2eX":[64,53],"D6svmbCCUDFYmw8burYWAJwBq3e3Cdp9wiLdfNZ4SLus":[44,32],"D71JRzjPpHipt8NAWnWb3yZoXezbkGXqSf7TVCir6wvT":[84,76],"D8Bv4FnVhmr118E1HeWYaXNSurvziDYopuxzdernbQ9r":[48,38],"D8P3w7GQ4zTYbJfEGgfdQWQ1vrL6umGYAUrMz4hBJjrN":[32,27],"D8WfQnAbBmoeBj7MPL5qWErd4xonZqJuvbJS6KjkXWj6":[40,35],"D8e7uALahjTMSH1YGUw27reebWx6rE1sgsxwbbCQiW3d":[48,32],"D8goKEZAXaWCfVSFbGKZQtFn3B5XFdLLmJgUSgMjEJf7":[64,49],"D8wUJAcgFSdEqvFrBDFHzSZuf3SE4zM8p6F7HoVe3fRs":[20,19],"D9rCbP5rBrJztzv2EaACNt2LhXVLpPmsNgcWyB6LdfWW":[56,33],"DA7SNDUGAHwcXxHoUhbPqTv2p8GnncMpRYYoT6eJKmSR":[120,103],"DAzVJ9PCv1UMsHg4F9wSGiZ1RrSPBkVS92SK2BPEiL5J":[48,47],"DBbLyHobm2UNyZWQt2FXgZN3ji7hDEx9gJXQ39kHgwr4":[12,8],"DBc9sGfoUaaSqJs6PX1uJwygs2FLwS34YNEhuYf9JRZM":[28,26],"DBd1Se2ugdr6WChDSGzSzkyENnqq8ZWerfqGCfKYUred":[48,47],"DCEHziymZV86Pe64mBBNVR55XPHVMvE7zwQ8in3Uvdid":[44,34],"DD89H8QdPyWGtR5QnrfM734G4qrD775HFGMobyrkHjn9":[76,58],"DERZpWNzHzusMAK29jRQ573LK8bvmENSyeLii7ZJtCZB":[56,43],"DEV5ZXDo73Cok1MvGaGuJaACvra6GHin4wdng3HHEQPH":[76,55],"DEYZsYj5yU86GCgYn22a4w2LZX1jokqA8sA57ez5A41M":[48,46],"DEZAHY54DgLq9Md8CyxBgNCe5hxDQi7fJaSE8jymtazr":[12,4],"DEqpqWRASZVoDVMQc6NpNjJbiY5KxupgiUXWCsc4TUim":[64,54],"DF57amFm9aHKYL9pKLSWindiF8o2RRxtReLb6d8DQc38":[48,29],"DFpi7mgmChYV9whs4uEtioFG1R2WF4TpGd1zcXMjGwF3":[40,19],"DGQ6UDGprq5t6fhfxdykcAcqtx9P8MqkY57ef3sjpGPp":[52,35],"DGf8USMPMty56BWgwFSUz4orb9smQxsWxufKBXSoX97f":[4,4],"DHNSHtEZHwPkkfomi5oMmCQy52ACoDpD5Kc6oNGTJTth":[56,39],"DKNy6YAPt6zq5jVD5S8EFSXpQmqA4NjrQf8t5v3tHo7h":[44,28],"DKZeJyvARVGMouBRpYp31WUj3NQTuPw3rrqpcZSpNDoy":[60,36],"DKmRzfmz1JF1e1hdkjG6pMxnmFuyPcevcdXFG9dVnzBm":[56,44],"DKnZytVA5wKbNPYW1pvPpoE5YeSsxu12KJFa95gBAGm7":[4,0],"DKyon4vSD7mF6uqgEJujABpEdhRbyX9X9EzFjmEz4VBx":[64,44],"DLDshHnnGetLXCyk9o3RpKC4iRATqgw3UyftYF55ffuq":[72,52],"DLbekTqPpWxTsts4a7GqC37zDoCVJ8hDt4VNWfukPnRv":[36,36],"DNCnSkX9eHoTaW2ekwsTBeierxHJPoP7mGTYxKrEdNqg":[48,28],"DPtryjrrMUTTB1uaKkbqngnj2v1adJNuNW4eRScUPSP6":[32,19],"DQAi5Bdgz5aPhMMUMo3Nun8TBewGf4zJo2EqGt1jNNQ":[48,39],"DQSg1PLT3Px71U4LsfBNhg5yT9GgH8FnK7qQAq1aLmk8":[60,44],"DQTiiFVnwD7bSSgkMmwUsBsgnNBza8N6oEGeMC8YaieW":[48,47],"DQZnxmnLdWuGP7wxpw4krfeHzCiu6jNd6m32DSoQZc8Z":[28,21],"DRK6coUhAikJdsNhcKy5ZUGhXJwj6PDrsFXKCo1UJpuT":[48,41],"DRUpYPmGX83wz9d2bv68fN6ruepqqukxUcFQKVB2Kvbi":[36,31],"DRaSSQyKZAbdunz6ZkV7aznM6jBF4GXd7XxMwTmmgERB":[56,32],"DRfnZxdhCF4CLFNGMgvZAcxA17CPCCGarfPcEV1B6vEq":[72,58],"DTPPJWkD94MGE2oA3cyszmCAS3QvxNjXznnRGjpanbRS":[32,29],"DTfJRiAVyiULdo26CQ5Dpzmt8gRCG8MhK5QCRkpngEST":[60,52],"DV78gathrorcpWsWrUkWrWNowLXpizKsPBupStzeAJnL":[4,0],"DVbHAVYnxHTUeXwFvVMRWUPhkeozQTPUu5CtRgW5yD59":[48,34],"DVnKs7XAL9au7cWrTT335gZ3agJVwrqeSVnSWANo1SJG":[44,39],"DWic5rF2jQeAQr771cCgzyLbEMgXaq1AUyhnpmgDze1a":[36,31],"DX6Uuwc6jPTF1nqah3reNYpe28cjAzV5LViwK61bkgBf":[44,35],"DXCzguRGhTGvFm9hdVsDkFi4S3n6W2yrNeUjrFN8tkvL":[68,54],"DXGGtG5B8JuNFRA5RS8X18HbTRot4dqFiY6GVL5WwxxC":[68,37],"DZPn1XMuoBpNQzcXozfMMCUJ8YxgKyJP1Su5oKvvJk6h":[64,49],"DaB3ZwVtGLzSjazk5STQEu3MkJR2nkK3tDdCPAvx9QpM":[48,36],"DadnDZbFH5BHHRHD7TaobaSQ7QATXgvWegHUcZ7ZGzmW":[76,67],"DahDt2bS4EWgf546qQm8PLiRZuZPGeUKD42urQJCBYJS":[68,47],"DatMDyrEByFRiQeLn6gNznSqV3gHK1SN8FVUPW8CDD3J":[64,56],"Db5FG9D5Z1WWDSvQioKkymwRiTTGcTHbryBniRqYE65G":[64,43],"Db8J4gwpgKyQteGE1ZsGgd6kHEAAynqLNNYhwrXNr4Hx":[56,27],"DbzdjE8TFSN1Zb4g3N9NsgFrzJ68G5WKtgSxqVox7Nxr":[84,66],"DcgUD8iGqAbKNAa917SBNwmH5DJZiVC5cNu6BaRhKRk5":[64,48],"De3f8bq9dXXHj61anmFSGnJFCoCzcT1FysLoTDRACBXB":[28,27],"DeysXBwFZpyHzuxygdWo66nRMAohP8P6nGX2ECd914BJ":[56,33],"DfTeDaxk4RufkVbykedVnqa1r9S3z3oKFYL3FFmPdr1o":[60,42],"DfjBTVrgnveaCjUC799e1AxUVg85EtXoY1Zq6qWqLmdE":[44,37],"Dg5E8ktH4GWfKL1vuVTdqZJEkAEgtV8LqmSXyLJuZ3q1":[132,110],"DhMuXF3UqZvi3GhdrAMVyEQ7pW4prM8DkW54scYXo9Ke":[68,55],"DhukytoqRv1H9J7LZiv8FyqTYhBhqb53gMgJT3dtdyk7":[64,40],"DiAZodQBeCoP65gTBfPsghk8No5QH6E1PEwntuxyeQy9":[40,23],"DiM3w5M4ATTQQheYRrizFCSoKCKmefPGnah8cPsrYt17":[52,35],"DihbVNPXN8M3A9TEMBJ55XUX2Bo3w6w1BK8BRsR7A8o3":[40,26],"DjuMPGThkGdyk2vDvDDYjTFSyxzTumdapnDNbvVZbYQE":[40,36],"Dm8Kusyhxmz2NmwF8RivLKembinSL6h7tvh4vrMVNxoR":[44,22],"DoQrvRo2yQqM9BT7UxcjgcdBuXE6b8TNx42a14ucVjBB":[60,46],"DqbeCr74dFGDPvg8BpV6V2Jy2BCTZWmuvChkZVvPC3wP":[56,41],"Drf2oN83THfrUJHA9AGzJZaL9KMKggPoL9HJVttkSCgL":[60,54],"DsaF77cCADh79q7HPfz5TrWPfEmD5Gw1c15zSm4eaFyt":[48,34],"DsnqNtwKA817a2VQypWEzaRXY2soq5Jgc68MgFBMR35p":[52,37],"DtqjPaZAuxaNRsX1u9e5EHFF2JjeTFwF5SZ9YFJ6PyTj":[56,49],"Du5zdm6f23229xviUfyKH9gZmaPDUoxVpSY4TTmzQmj5":[44,31],"Dv9qRs8tB9oCGdA8FBtNtT7SLuttmVnDFT5xoaq9NFTC":[52,39],"Dx4bMuKpGaxAnd53QYDyKhD45PjuFLx16mrgoRK36STf":[28,23],"Dx4qoPTLSRdbd4h5cy2TD5rDyV1d7LxQYdSY1TLStjcp":[68,56],"Dx7UU4my4DNPwRkHnJN657sMCakPWYPcWwBGf7WUiyYE":[52,44],"DxQTxfgNhgzpRR8dZZku9XHPpWyy2Fb7oJSWoLTKwwKt":[68,36],"DyoVzxMFgZfch9L2jDYz3kpEctE7CVPVjvseCnGxqVqt":[52,19],"DzTb7aPtvxo5tbbKxEjiSuBe74tdgswdj3BY8LwstoBg":[64,62],"DzesSjb785mtCs2f2N5UAfnnF8obRbECAe4AQaRVmX4X":[64,56],"DzxNmWD99qvkPdDR94ojXhUrpzT8VdqB5ktYX7fZr4gc":[72,60],"E25BqJUzSjyzeZQR8LYUcUNNVrgRLHhXxhPtZGB7KCCp":[52,39],"E2cy4hqcUpdyMpx3TuHKpdW2cJZ3cTSthk4jfqJryt6B":[48,36],"E2rAcJ1tEzXK1cqDT6UhtyJmHzK3MvnZsLDYD7EkTGJN":[44,34],"E36wj8TugYUB94FtGrs9zPdb9SuT6jP4V2eumMXVy8tk":[60,46],"E43Lu6um98dGLscCuPUobgKC2oLANeByzqdab5KjxV1W":[52,44],"E4YYWrsKv9YkBjLRtVNYn792RavzkXL6NPJ5Z4sHXiG5":[60,42],"E5q213N5LpkpABYBWrHWaiPveFZGCRv49sq96wAZu9Tt":[4,3],"E6MuSSCF5aoBstVcZaD6sk7hkQrxvh7s5ttVt8NzAiNM":[40,16],"E8CDgCsCgLf3gBtQVxhKiL7m8DgUdrDAi4uHMjcPfZVW":[40,27],"E9WbzbArKL5bnW3NTYZPKuU2dxMiznaWAyc4jpNDfdv7":[52,42],"E9bcuniYQhMscfMjE8zaAXQ47TH56gsQoKuzvqXHxnAY":[76,59],"E9bwsVRDSEBnH2CbMRcdfiTjbkbUVcoYDrJMhHbESC1d":[40,23],"E9nLyV1V99DSxbAr9WcvXQCx4sJkJewAw7LF59yJK5io":[60,32],"E9qZxXtwWT5FuwsXHLjA4cjJyyeYb3ixHxBSrJJDzPwx":[56,46],"EBDnuJT5USg5HsQSZtWT1q8y5XjgW21b1ebYSahcX9V5":[24,15],"EBUmXF7Jv2kPk2ANMxu1ggRNtFfUkj6vpnuufqaSXHrs":[68,58],"EBxhSfAWW2Cfouvj1k242W6U8krZVAxJS47SG8UKb4ch":[56,45],"EC1TjttBQaKU1dXuMbv4ZMSFXuPDt7UCMvNXruxCXdA8":[68,49],"EC8cRNmwgbhbs5LtvufLkme1QqedbugySgkofYtPoDKd":[8,8],"ECYTLkfbyQV1piR86TJK38yvPKypUtubb9CoLkGzPFC5":[52,27],"ED4pSxzaemm1KZ2kgSnimKp3nw8GB4qpUSmFCxhqUKRw":[36,28],"EDpQ655JNe3hFVgMdCjz3JJhAVkZpwNguyYxUvQbvbc6":[4,3],"EEBa8frib8zBLxj61NEMAUoEyrHFgV9MUzneHVHFax42":[76,55],"EG8D25QxDJ6nbD3oBpu4tPDvihriy9mFiPq3CxCGFiPF":[56,36],"EGknxV4LZM4DNL1Y68iAPQEdLsMZbL82wQbDmsGw6w8":[44,30],"EJJtA4Uzis7a2eowSGCGgXypScp9aLzjJAiKrBzdT2zJ":[44,36],"EKBBsq7snZXyabwMa7jbyyRTMhaUQqtDHtpVgDnSg19c":[48,47],"EKH17WCLR6oCVbUJR7YcThnTNRNf1upYRvSnha4Xu7cN":[36,28],"EKxYfdfsxAR6BSpYKW3LD63UMn9ZWYUNRJ7GUonWSdrK":[28,20],"EL3RZmhvLAMMoDip59M3oKgqXXzHAPdZ88KQ2h82mCB8":[52,37],"ELJyNgQ1jng6XCTGTJoha7af8srgZdQKdBGrZ988G4Wi":[48,31],"ELWMKHPVZpFTwBSzVPF2q4nmvexLxWycjy8fuoC6egBE":[48,38],"EMeaA1d3kmoBNtZQNgqEZS9y44gMrA7iSuqS4nZ4qxpB":[8,4],"ENtpcfC4cBpkCW3kHza399ZoUcbxYGsYkxuKM1jwei4L":[68,54],"EPVNShGLknd3qtChqrgvDVmBZrpi2yktTk9YfcMbuFqQ":[60,30],"EPdTTdBRGeiXp9yzpPK1RUQeinoCfgq9qdicBJqZcJgg":[56,41],"EPpXthnkXneSTPPxJhuDwzau8aeYddx2yhzX35qVbpR1":[68,39],"EQEEdsGspyFKjFmAsv3gzYXbfqSCWuHrv3iXodb1LyXc":[68,58],"EQEJp5fi1ukX8JcAT4LrNLYqhxhpWW5bp4eGu99RQoZy":[32,28],"EQFsB8CDcLsCYeRJxhZ4fJWnXjCnbxrbhyqjyUJvkDcL":[64,54],"ESSnx4SECiLzhkk6jdQuWUmgxnfmpHTeQFdVTMwGYqE6":[60,51],"ETMbiU7hEt7jkoA8H6ACsfeR7LyGA773k9HA13yJUfex":[44,35],"EUjwiu81EZrQaLfCUthHQsdDPcuny1QfAWQy8Tag3Vay":[60,55],"EUoXy9YP2tAefgW5CHEvMGAu17McAvrXiQ5ucezjNYcd":[68,59],"EUwiTG1ii59qWfgsJwEMjqh34NmShMdP131BWWqJVPaG":[48,42],"EV5QcbuJdafM6tSdz841AUvbQXu4V97R1JPZs5Cq6hJk":[40,23],"EV7arpFrG8SFFAkfYRMEJtuqocMeXpNFD7zDt22qG23T":[76,62],"EVBcvaCjpB7jpT47oX3YGXmx88yZUXaKAhZyZfzjQyJQ":[72,48],"EVP5AqRX9ocjTLTNjzDqeuZEvT6EmmmR9BW4UEezCMRo":[24,19],"EVQxcfApDm4snuJU1XHLcDmwqiLsAwRQ6MatFKmSTWv8":[52,38],"EVd8FFVB54svYdZdG6hH4F4hTbqre5mpQ7XyF5rKUmes":[8,0],"EVtpGGP5SWd5h2WLtGYMGL4e3SeA5Myob5E9PHSRdT7m":[44,39],"EWg9NTC5s7Pa9FktUk6dX8xRYkvJ952peH1z1iznd4nV":[36,31],"EWhSgyD3VM1HuUisMz31xQPuyAXH5skcY5btY5ezu48J":[52,42],"EWi3X56yoWm2WWMA6doetd2S9bSqm2cc8JSBaWHGuJ6n":[64,51],"EXEFh9rPB1VN5NGDQsJiAgR5qaxUzHQDRJqAZdqEYGV6":[32,23],"EXT1KVqLrtmBhLvAtHyyQcjgYS3nuFRRxPsSZacUpXoS":[44,38],"EY1kDqUr6hfV3oevMtbHksUT6tk8f9YG3eSvWzE6DVMy":[48,27],"EYbvBPU9mSPTVJrZgioTt8PGPL9Bjv5342ENBMR5X8r8":[44,18],"EZMocMSpodWGgHthkEwXCqfLXcwi1vspNo57L6ZTCoEW":[48,34],"EaVhV1UzbiAh8BCdTiAbvoGWktfK7fdR4PEXkkN1qG2n":[68,55],"EbVCEn2DQjEspKn1kJYACrx5r5EjNcFKaSL5fxJ2Tois":[44,40],"Ec7d7N5r1xHSzq6nZg7o2nVVaymvU2APpagmBJ6LQ7Zf":[44,39],"Ed2cbqffj85JtCw9fKX6weRN3amAZ6upPMUJ49RZzYru":[88,52],"Edmn4FjZMGSmCjCE2FBLzHNjukEXbzEKiHptMfj87uU9":[44,38],"EeHdr93aBEXELpbDx9p8ScgzstUZuNZmFFHM7oPccXzS":[72,64],"EeebKTaPKaff2WmcaHyMH8mTVyAL8Gku6Z2owg7Aj29Y":[68,57],"Eem6rzePhp56kYBvWNgU89PNjrnWqJs22WcuiSjmkBc5":[52,24],"Ef5gVy3PFRJA7uQ4UkAD6AufNcZNtHN45k5N3L5mYatU":[52,46],"EgQ4ZdobHjVmgdzMvFhk6MrTWcyh6ULD6m2VFNqdCGx7":[48,47],"EhJnXqSV4wjCEA1bH8LeZQZmJnMQXJEMj6Qnya3u68gn":[32,22],"EhpB1NmmfzKXJb6p1w8UFsaftsjxkZXLsAGXh3j68kwt":[52,52],"EjcB41hrq5Ltr9Yvda3jQ8zGkkfFGKadkykTCQnPeCne":[44,36],"EkCbYFvhbY1QmuA4D5XTggghpPWsVWumHkBZYuLBqyMG":[40,36],"EkVaQMGB3cbyKdqBwBagGtURjtoXsP6pS7HGyizwhUs2":[36,30],"EmDwEWAfsjmLJ6kyMKQVJgcM9WCE9LdLxukw5aLG9KjP":[36,30],"EmKohoDf5ofnT8ivyF9RRtQB5YJJLKhHhvy9jRhyj8oF":[76,67],"EmVfgcaPpEth2uRiz8EurhHZREUgSt9bsw14CfmEz7YA":[40,31],"Emqe9upNXhojTRVT24mMxLpNB3Fnaoa1FibRtVELunUL":[32,30],"EnLMAih8NTU7ENb5tiTHqgP8MtTUnY4QDLUKrsRwhjtk":[56,48],"EngVeJ8w7soeVvKwypuSutnXFPSWDLMq3Vw5wuAdSGjf":[52,39],"EnrrqXLwEwgDU9yfa7YVCxmm6uj9vjhCnNUk3qjpFgws":[48,37],"EpfcKyhvVCGXNmf7aRoDwtXBMkE8xCkuLk1YTakwaPdo":[32,28],"Epxg7Ft5s3pwue4NLBRXShZZTPs84Fs2tGaMHMfp3kPT":[12,0],"Er2zvR4xjErXdauF5MVqBWukgRCT5yKBEnYh8W2ZmXTo":[60,55],"ErbvzZx2Ss9GxizKyDviybhZPu8noHv4AM5vuzTh1ij6":[4,0],"ErugkAcX3gz9cw2KwYZ5vPwGuQQxhuobM2tsdoggqB66":[28,24],"ErxAGCPBB5wMWU7mgZRXUoNyYSnMmVR9689hd6CMTfsS":[52,33],"EsXshV7Yva6ZiY5P3u41vjWYNHTNMaS19qcoGAdiPZK4":[48,43],"EtmHTosfXS5zbDbTd6RxWvGLWwmT6fbjR8YENZ6byfQ9":[56,49],"EtpFdJnQ25ZJMheLyURzyCD5ch1SL9smfMcEeKfAkEHq":[36,36],"EvDnGZpca3pWC4tW94U4PUkruVUkq7PUgpVWpDfx6fxV":[12,0],"EvVrzsxoj118sxxSTrcnc9u3fRdQfCc7d4gRzzX6TSqj":[8,4],"ExyBDh4ajq67itKb6HxuWBN7CFtTrg6NbNCF91mFjyx":[44,32],"F12Ah86ymdNPuXya5i3PKG7jeLfSMGpoRTriVTgcXr15":[48,39],"F16J8jYx76jpt2vgTT5SPv8hJGGcrShzCHG9LBV5vQD3":[56,43],"F1MevczAijZ9ZeWNn6T5ZHtiVkn3oSepvmcjCfPvdsSb":[48,28],"F1TuusSghAobmbGAgNrdxRS9nBjwT6J1yyALUvpEA1is":[56,19],"F3S8XEVrUyG8sN4uqyoU5hkyk1zRXwykE8QWmxA8QCBB":[60,56],"F3bXikq6WnjMQjKcvj7U2tasv9Q21xTWVUTp48GmhVas":[44,42],"F3gsehGvHNXtF8mDbGVfB27Lq1paSgTiqe5nzvbFVREK":[52,22],"F3w8NqDxCxS2eL7d2Ucxti3suRkqSg8ox63uHgafUZCW":[20,11],"F42UBCuZo2WxyZRG71p7SKT9iewEE131hyVpr5hG5kBJ":[48,40],"F48TattaDuDLAeEj9nKvUL9aq5vTJTbx2gv8zJJ65hb8":[40,31],"F4R2g7TnRmr88GY9DjhFo5Ssk9Ji3phBRssrZL5rQxWs":[4,0],"F5DXubJ3HqdWAV39GUX3rBUekzuzdjoeJwfKCbNnWKeG":[48,36],"F5f3vcPpfwgouhVVzSW41XZRzch16jr2qx7pYNYyVruW":[52,42],"F5vwBSZUjzdQxwWKdUBSxgsdqXpzyhtr8qiYPL94UqTH":[60,52],"F7FgS6rrWckgC5X4cP5WtRRp3U1u12nnuTRXbWYaKn1u":[4,0],"F7zQemwQJo3bKVAvpcCAfkgXD18kZxYvgMxCP3X3qiK7":[72,49],"F8ZZW4WKUx273i4L2KubqUCVSKLmSgkpxwHRL7gar1F6":[52,44],"F9YANP9X2AeorF1ZCY37GX6NXKLyouWNWELk9PrdPVCy":[56,34],"FAmkcNkVCrHXFcmLwBmv4T4Kwb4wDPmTUmoGc5qCwVeM":[28,15],"FBwS9NWnUfsNWAhPaYm3nzZaWVZkZuefEHAUWHzUue1Q":[60,36],"FEKzY1TLRYWDc7AHTkREpoSHvx9EyNpPmxp9FeojPbJq":[56,43],"FEq9FL3hzRDMtL8DinPAaeJpb28GBZYTpTeRcRyHSrGA":[40,35],"FEzZYThFYwyBn4gCqY8Kb3kfJSUco1XTYdVDPgdgRhWa":[32,27],"FFhtic9yPS8ao7Qg1GKjqyzwhGYK5tsksT9VrLioTgbY":[52,47],"FFrx4NAJginBWNm155TXLgx1annkmdqEwAP79nNRwxjQ":[52,47],"FGzASVsHGu7NdSbrzuTHyGNViE5CdcR2ZAz8MVLY1jx2":[60,55],"FHZbKviw56w6LpFjDaz6MSvwXsJNYS4Gry7Q2EwjDXBJ":[72,60],"FJBA1Q7gUvEkEsSuNDfDgJEKeiJmfF9Z7tF93sx8wNBe":[24,10],"FJMCiuiBEbKum7cLx7pqi1xWeNbpUuPJBKUeiBTkxLdg":[56,39],"FJzX9Cs7zwo5AZKUjf9wkBHULSJr6DaESAbUbBPhx6E4":[68,57],"FK1TQPnYVzg1e925kHurusgkXxxFuBfEko6D8ZirKNeR":[60,47],"FKErtUGcnKvAd4xizrU3jyoe9WWkCSTPZA83J8NB2rFk":[40,25],"FKYKRLCFmh7uUASUQjkL91yXCUuJ9wdbPCZEnv8HkKnf":[28,16],"FKny5Zv2nrLFKfNH4jatujiiNG2c5mq7MXweKAEBBzse":[28,24],"FL117azyKxNeDWGWEoiTj88ygTHFZNs13GKJHa593GW1":[40,30],"FLHB8AGEsED5jAF5sS1kSkAzSXVK23iuT7YDPHGmbcjb":[48,41],"FLcUYvDMd5nh4cyP3oErMHoKnKREmza5rdAZ6XHYU1bd":[48,43],"FLdAnmYGeGmwJY3qECfcZ8pyQ1LoTAeBPm8YKFDxQrMN":[56,40],"FLpMRfbSMkBnFXDdGKdzcGP8JgrNVhaYowmtArNughqt":[48,40],"FLpouJ8TALG2zroSZ2SC8wGdhQwyRjyyPetmJBcmcKQu":[44,42],"FMHjnmeRLszDSDTmHrbqUi7rpXLcrynh9K6jQvjdhqf6":[28,28],"FNBpvn9cNMmMA8GRfGxaD5P5zkG8m3YAJybgJkVi9bbK":[56,43],"FNCnVhef4ZbzxMsKiRsHbjpiciwV1YcKmf79NzwZwDvY":[8,4],"FNH1XmR5WgK7CH7W7YdcfxtdgaKFueKtqaggVr3CnY7M":[56,43],"FNdoUuKVBigMFGpVvSMLXJB4FC7XQL1RjPUqUiwvPiCS":[60,38],"FPF4V5wVGhiM4UxLLFA9LoaMFLs8pfnGrYhsRrjS9xGk":[72,65],"FPncJ2ASP4cVPcmHxibgpqaxMuwjaw3qNQiZvVdoefDg":[44,0],"FQFrdHAhKFP9R5R6JkJPtVJhLDivDia4cNNUcL6Eja6j":[48,38],"FQa4mYpWL7mNEXe8dWbd2FXxpreFtYJkD2S6hMD1oXHH":[28,24],"FRXTSuAxj4RaD2x1kEN6cgdcuMAMMdSEfaXcUhHSu4sM":[32,27],"FRkEFxUbT9e3GCJWbjkajHF2tcie7dDeH8rwDUPThxZk":[68,57],"FSr3fGXzbMPBwWhnWY6oMSAdZqiVXxw4EpwxRAJpu5pB":[8,0],"FTUh2jo7GmxFqLy8c9R9jTPapfGjwcaDdozBjhKJz7UN":[32,31],"FVD3oRjHRyDsGTwisgZrZPV6vbbRvkgx8hm6ctmBG1Ne":[56,36],"FVeAQHyyBjnuPDVHb8aFk8FRejULrB78K4SYyuuj8Q2T":[48,35],"FVnG988wW3uF613QVxmQrkwtdzS8taxjFsuARTzZBwMo":[36,26],"FVqVvYtA4CemZM12MVK1gdow2AELmMG98hXo34eop9p4":[32,27],"FXGBEHAr81PVbD4ExKesR3AwXp7SwP2wX3kLsjijd8eP":[68,57],"FXVTpozaNNzwiaEu5HbS9EK7HMNm7QPLk1UEH55hAkK5":[36,20],"FYH7U2HPhgxQCsHBGDaC135Rsx4tZx7P6ZjxnGqWHBHn":[48,39],"FavsezxBQdxoQYBioVPoAuwH1NwBqCQhpNdtuiHdYyZB":[68,55],"FazMcimRMw64FWqyP7LX4ATBJb8Vm91UCdc5kgmBzqTY":[60,45],"Fb5cEcYNgPXKJoEmvPvsU2ENYRVePQtExqgf77AnVX54":[52,40],"FbWq9mwUQRNVCAUdcECF5yhdwABmcnsZ6a6zpixeKuQE":[80,58],"Fc6NB99bkJQn7JsVSqdJs5fJEzj7KFpe4JHNQCGVCctj":[56,40],"FcTYrxp31zVjTW4qjFKkgRcKXbWcBbiRQqJYpufwcJZN":[4,0],"FcWgrc99RAix3y9th526GnzN23MQSkFmyWaeo9xJ6Jfo":[80,55],"FcWjjkFgHgaARvF3Aa2wu3C81B26CEPRwNRwaHZxcZHx":[80,65],"Fceiwrs9CmFC3ZWSZWYUVDKPH3CR7FSevtQbL3e5XbTY":[48,39],"FdQgwQ38ETKv7x1mWNoAdrLR2YZyp16xFDC9YR8Gseva":[16,12],"Fe8oZecAAGKLDGziBLLpUM1nwD6DyF2pXLMZAkWjNm9q":[12,11],"FfwtopRBJWKEJiCmkNUFyaQ2FMubtzMhAzKgHDF7XrLa":[64,48],"FhUjWiMEFauc8nrypsy4Mp4eGpCme1ZBAJQadmYg9ULr":[64,44],"FiEHvgPh2YMigcYWZtGn2tCgPFefSgTRA3RB3e11GhLE":[64,60],"FiT4sReWDW4hrcPHW4bCmFN4GQQJw5cJMNeHf3VHVrKX":[68,57],"FinkBUX83H7gMds4pBwaTntpdrK6ML28Up6Panpx2Atm":[52,48],"FivGzpupCvU4yr9E3J8RvWtLNWTm6ZRcGS87a51BVHWS":[60,52],"Fj39TNbn7GHy7kbvPvoCEtSFA1M3k8885FaSGc6WxqD7":[8,8],"FjVP2aWyBT1ZWkQWaC8gVcQL6jRrNceV7r7qNjGciqY5":[12,4],"FjXtJ7fvCGt1HRmYrhiqUz3r1DWDqfBwZCSESDPv6bSE":[32,16],"Fjc7dkd5ir2heioaU8eomUgbX2JY49BCqBX3doB8o3H3":[24,0],"FkHabRHvsaNNVQqHEdyNAUnb6mScR7dVCZeL8d4ftEhw":[52,44],"FkMj2LPSWd9LzLZrpZ2L9YL1CpB5eA5W3J1vyvvpp6Jm":[48,36],"FkYhpz7HSGQJvA6apj1BKoUfytQvWseLfSUrE3zjvkQb":[28,19],"Fm4UoSZ4jYnTuLLhav3A5BysiqKani8M6naY5VLFzMsq":[24,15],"FmMWTNcijryGbzDrbB4nsU148YQarwDnGqVnFwW7kexs":[36,26],"FmZc8PpFoPLghneqvw6ZNd6HohL9uBWUZTxh1c6CETrh":[68,55],"FmaWRHAtnhTX3iDhgTzaFHpcuP8TgjWux68zo3kJxttT":[48,44],"FnaAWEDBPSNCUk49EM6JFS2hr9kGqXKkEDXkw5fdR887":[36,28],"FngAas3r7KShgnvmgtx4oQ49tF6q7WpUSnxovVsSyyAX":[40,31],"FnpP7TK6F2hZFVnqSUJagZefwRJ4fmnb1StS1NokpLZM":[4,4],"FoDccJmq4PksAoMpRbygVVocdp4NrC8PSwwDd8nfKYzv":[52,34],"FotUJq3ueDw7HdUj4R73XUJt4YYn5L1QbEADVLgUCc5G":[56,41],"Fou7Du6KtVb8dVMzKMYW39fuSGpMzJGwpkQ45NbxA3Tx":[56,26],"FpX4XYrSFjBcUNgt3x2p2SwcHaKSdZxNDw9KYRjrnKbY":[44,20],"FphypHn7XJ3HisFtdnUX2VdqsWNLa1ATwrA6ytR4VSbG":[48,44],"FqdAcsUQBMibVJVr259uSAnA5FMK2xACaz9vPEtdvkYn":[56,48],"FqheXr2yJSTRGncTqVFFG5sLaTtXZeQbkQAxbL8mcGru":[44,0],"FqrYY7h4wFofqE5DPPfdUUTo7VnZqguhTc6pDx6R9WA3":[44,42],"Fr1BMG4DE17Lu7Jj5kG9gndjKiBvr9kHyqmoFMWEpA4r":[24,18],"Fr3WfD9xCLX83AaktYvcihvqoaJKh7f1AnD8mCN1egHw":[56,41],"Ft4ADhkxMVfgxNDQFqA3ymaNGC39rCHdUj6H8KEWQqXy":[12,11],"Ftms6EEKfgb3FfaoJJk6A42MqQXVcU15RcRa749LLikL":[52,22],"FtnMU4gqaiRgAJ4uFDFARAp6WbCtGcjxqck15L6Q6EEY":[56,45],"Fu93Uz1dJHhp73tRFDHLZ9QrzJMyaDnorgfZFGkrBfoM":[48,40],"Fue9LZxjhk2DNXWxM3rPKr3z2qntChdeth615m7zUo8v":[40,31],"Fv4zJ7RvV8gEYxEtLjnGZAX1qxjqRh56DzBgqvFEVjjM":[60,52],"Fv684z5SvpMvZf4e1aaTwaA8j1kfnkxd3iQNBdyFNWKe":[36,22],"FwhKmbdbaqWqSMPimLFbPGwZqhpbPJEECnhLdURrc362":[28,23],"Fx7dVi2oVpynBKzU2V7nRDbdfBjWrqqjLFxULXCVp2TB":[68,43],"Fxv1ymSwB6tdRCbjBURQK6P68XR2njCGfWbnfzVciJsP":[56,43],"Fxw5NgncJhGbnjH1wuZuavVPDQcowTwG4wKiWAB4WNAw":[60,47],"FyGLvXzJKfNEBFHS4ezGrWZjHfbqdMApYesRjt4yZD35":[53,48],"FzAv1TFpCyR65GrxeqBwnEzNVXEeUMPV5rKZGQhPR7mq":[64,54],"FzeiVQGdYrFWa8p79HGPJdnvy3zTjxmwEk8fuMcDo4U5":[36,27],"FzxDXiFjj1Vu7HvFW2wc2WRRV6swPRvcPnMHNTCd7QVi":[52,37],"G1XAtQaBj7tkwZAHwmXcAKSzMB8C9Kbfge52hZuxBA5B":[12,1],"G1uCu6JrV683QK3kdAzEiiAEBSMk32Ugy56u685cynJ2":[52,43],"G29KDaE6ed3YWzWaNesjgoBu5CFJrcHe9sb8dr7b7LLq":[76,56],"G2bNSPr2Peyi7CmyJ4srX3YdEvu5egpQiorjivMsgnrs":[68,64],"G3XPqL7h6TtKsenCBXi4NftqcVGwKFFVGzxeVbkbAKMt":[44,35],"G3tnUHh4JgDTN8wnesqZqrWN7yA6hvdcj9kRCMNmEyfQ":[60,58],"G5F9RG5hd136BNBtewsChY1E2jnPoABw94mZmeConsFv":[48,45],"G5dS2hFkFQcx3ZWTeLSu5RWmPxjXV1RY6VRwX4fmejQd":[68,21],"G5r4XSC5D4Rw4NaWjbgBKnj6bNDsSGUvE46w9BYAT79r":[64,47],"G5wSAoZRgCw6EMXsqkRJyP38wkyrV6YiyHup3PSnXCXR":[56,0],"G6kUvH1mLRuwhSzv84EwRXT9fpZyUb53Fjm2M2n8oQo1":[28,8],"G7R7QFy95eeELpCgRTdxFSJbGp3EdryFKN8ou41vZJGA":[56,39],"G7cSi3avELxMLTCossnsooLj6UNhnfER6kpnSx8NKHfM":[48,47],"G7gAgJpRHnRvFhrUMA5khWMqHJ3tpWVWdpsBvCq6w6MY":[40,30],"G7mFk3fX4xQmBV5je4926SzLCphWFoww8APYxQKfkNxn":[44,34],"G8QaUmUwzJ9z8vu2XbxctG4eu336mfg1oPbBnWXpSE1H":[68,61],"G9AHwpSz6gRb2PYQTj13oouRmpN1VAdHwXGzb1eoTfCT":[48,43],"GA8fULGnRXvc1K4bnue2toKKBcgqrqoUzzF7yi7TkvHS":[60,48],"GAs5dt2Xjtd4f1mqZob6hKhY7H2J2HEAc8oFQ2fEjAcx":[48,45],"GB1F3KY6GJpiCoNWWZuvMiquQ6kbbtiSFK91mZSHEJ6g":[40,31],"GB8b8Zure8MzwQWPycAcEr9Cx7ARprAUoh8fWG9KiVLY":[52,48],"GBeyvNz19UahKPmJAKVuTUQLNWkYVpCbjDqExmYtgVFP":[36,33],"GBtNh6c3Jbf7xMSA4VCAbjkJpBVJkB1QsZLh7iD6Hpq7":[52,46],"GCBBx5HU1Bidr3xVkc7dox4HRwpBwg9Y81s5h4pmVUrt":[60,52],"GCPW2jinG8pk2KfALJA2FYNhLKwCR42Y8ccQPXz2PYg":[60,42],"GDoZFWJNuiQdP3DMupgBeGr6mQJYCcWuUvcrnr7xhSqj":[52,40],"GDvW9BczzCHnnGLK7wSeZSkGQqP1ie4qcXuoaU3KR1mJ":[40,31],"GEAFsEHsXFJzEUvwHxoy6f57a8GfWq4HMZdaAY8QSHEm":[24,16],"GEBvuMyPAM3Hmsr4UnGMqeeJNPiC6ZqPkCGKW6pADd8h":[44,44],"GEeL5VToy7H2oEFyLW5q4T1HeCDa6AA5YN8r4PiCnyW4":[36,32],"GFFuGhyHAr2fjH1DL42m9EWpAWXXdZ7R6PyzuMzDodLy":[48,35],"GFZ3w5CU4Byjjo6BrQxVsiq9mumADrR58vH2TT2KxmFC":[64,60],"GGQFKUi8FWGSAnNWeoZdpwERRgaW4VMiBdiZbEETq5Qz":[84,52],"GGnmHbA5wzvKcn9kTcy1Q1JgdY8hQoHAYRk9HaCBNJzH":[60,46],"GGyi5gNaZYHpij4uM7UE6EuN93dvyMiKMphNWLfTPBb8":[36,18],"GH1t1LvHefMhw9y7W4LNWWa79HHnB1bQQcXGHaTc18kg":[4,0],"GH383ZjUf2L1MRqbjNVGtvUsJMUk4Wx3avcF1uVDyjLB":[36,4],"GJRLu6i8j4CJukLEQQXe3y3pdk3ynVkt7R7ttcfCZBoA":[56,54],"GJbU1XAJAky6sPVH9BGux11AGyFtkqZwtiRVS1itqcao":[32,26],"GKGba31Zdwu7qRLBYGmPCkk9wwfJ2WH2a4FKxb3xfJuV":[28,27],"GKKznSeuCDB89NqRaVhVPr7ite9bsG1X7ih3uc83wpUH":[64,59],"GKPmKbjhiLqXa1Pp7KDQTF78WfvYxG4oyX2rmAAEJ8FP":[72,70],"GKaK31dY2nb9FdfYxL7jzge5v5BUo8M3APLarCijr4iv":[56,49],"GLKsDBjWBaXHkyMihjpU5ZdKyKWtUpJyE4W7PjEFSEHh":[56,51],"GLPe2gV9zG9kwJmNj9GFBrzEiqridEd5KYgeSiPrubGz":[4,4],"GLQ75DmDNSn3w4RMDQHNVeWjhe8d11kREHJampHUTkfd":[56,41],"GLSFJkyLwsGGNr1MYZwDCFVzkVRdDTdFn9eMk17reeyW":[40,26],"GMhCc4SBnHPHmL4WwRFiXaDkc6qyUpeifCP7jcTim4LX":[56,48],"GNDHwncRV2VpzXQLpGLvQUGkkfDFhA3g36Jg8ZthCNLm":[48,44],"GNPK6pfoaXcz1sKavyYRdAD7EtQy3F1CXBTWVyxEv2xu":[52,47],"GPpFVDjENZafNP8uysYR6TtnXWdN935z9jNDocP5hbzB":[56,27],"GQCgtjErUk8HCgayXyjuCFR15pnssi4saPnxD2Pm9oav":[48,35],"GQnaJJu7h53SVVhVpg2ErkSKhYMtqYrqv1qr13MUobuq":[52,46],"GRqJ1rtEGdJMuraEM5Z3oDbZ5k9N5mW5jkXTeDYTT5rT":[20,15],"GSWzDkiWBaLSHCpyWzFwoLyaCnizLMvXtk1FjQSTimwU":[76,73],"GSZpBMHhS7f8GT2dFT4niF9b8nLNtSkMecXR8thMnkSV":[64,51],"GTHTdVpF9dXusgSWNDQdb6uWRm6NFHYEkMj4XqZGoFEK":[44,31],"GU3NxbUE2GrZRiqzAYuvEr9h8rgCPj4GtqznSqr9yQiA":[60,47],"GUMhwogpDUNSzrFTYzc9soNWQxVKR9XJR3xYttPTVA8v":[48,44],"GUq7RT6MzPK5vswngQpJdJkvtUDWU9cLarM5Rec8VW8C":[56,45],"GVS1B7hSXnUQMQWN78BWjT9Kr1h3k5UW5UTdFsDhtRGk":[32,20],"GWBQoDvmnuANcRCxkBY6YCT6yp2MmGb8NgBF6XNvmdUo":[28,20],"GXYeHHFjnnKdL2un96tLmSduturuWzAi8MEBEYH8WYYm":[64,60],"GXfJaLrWgQbiuutiLyN7ijRgBvAvunJy7bYzaV562VWP":[56,44],"GYNPtzPmbceWBZjTe63zeASYXdW83uh6WK83zghGQqzf":[44,40],"GZeAno1q7JYL493V8aED6XQzqNPjJRZMq84jtKdLhWNB":[52,49],"GZgLom18cFaUKpAhodJ7LPzKCxyPbCzX5UxSFuYbdUdY":[40,24],"GZvmX5ooGidthybAp6NuMuCTyszmGX1aQCs5HxWvDb7i":[28,27],"Ga3U2KVbMn4y6z3u5SC82DZdDYUVAvnpyq83KERD6B9k":[48,36],"GaAb7uwik3bGsMurHJNmbabF8G8k2cYJy4Wrv3tefuWc":[60,44],"GaTDMHvngmoJhuRYrLGcE2GMCofu7K8SLGwkCkDE1mYh":[36,36],"GaVCM6rQVycyYNDVAjAKmnmCMnbe37qUuGmyTKhYPCQM":[56,47],"GaX81Beco7LXBwEpZrguBAHsQsmtsMFVB3EXzc5jZGzV":[40,25],"GajXkJZAas4ZjTKzgXQc8vTMyiwQC8RiFPreqRGBB71b":[60,44],"Gb9j79QFprtbY4sieaLZGjQr4a5ifFxGdnU2qkjPbBJQ":[44,38],"GbY9gVU9wuQKdWXrw4W4i9dBH6feAq3BpPwDKiFZecyY":[44,35],"GbmZkxUgNcxxEFiHd3hMqpBnwSPyRYz5c2Ya6UPFchQj":[72,55],"Gbqh5eq7nVajFNocG1GZGikoiAqPstfQjHxdGBc3SD4M":[48,42],"GbtVg3D6bNFSjem21vrJBJTpUniwwEtmvs8mQkX5XS1V":[64,48],"Gbwg3HCbD9gzma2o6oTqYkDQnojeZ1z7Ygc95DPFA2me":[28,17],"GcibmF4zgb6Vr4bpZZYHGDPZNWiLnBDUHdpJZTsTDvwe":[320,276],"Gcu1fpYPwoyBm7JjLJsY6H2eHRe8J5bjzoXtAwpy6m9F":[44,34],"Gd7zbMcA37tfU6dZYi5GfstUxGowwTd8CQPdDfsAynKT":[68,65],"GdjGkagCgTkVE2rwPdPUy1KXfFFmihD7GGzpZzyRHfz7":[28,20],"Ge2SFnQj7BeJsVaNqSrMz3XFGBjoFMMZ8qThYZRYYNr3":[20,12],"GeS9DzQuMz8PnUnUPPWmr5JrMdrRKDo7RiviNXET8Lak":[60,0],"GevceSyTLxHv55phyp2PirpdsdqNFZRZSYViRCrXmneh":[56,42],"Geza1KeYHg1EmwTxfNWpvZN2MTDxNX5aD6kLXw9ABuDT":[40,32],"GfT8F8MgHwdNkPEhytScMFL8hJcwM52uc1fiuj5e4YKh":[4,0],"Gft346NFxfieeCXCHuwdQ9TN6HyPLr5oyfwGS4DGQWGt":[40,19],"GgVX5vxGgZqMb2M8127Xy6AGA3kc6BCoFUr7Ex7rrK5W":[28,23],"GgfiMJLWSKHr9JPR122BiyEXgCDDmFQuzzkKMNbNkykk":[4,4],"GggvaPf92W5X8u7TxeyL1Aj7Ztn5ync5DjTDfpH63ffW":[44,35],"GgmneSMKWnEcavporN1vPpyTun2QRBzCjFCecQT5km8y":[32,28],"GgqsycDuFrqcM1isQ25SX2X3r8SBRqSwhH5hMn9vLFgo":[48,32],"GhBd6sozvfR9F2YrKTFwEMqTqhfUjxNUYKxye7ZvTr5Q":[60,43],"GikkfYtVZgaUtcmreVpQ1Eamw7mrnf2jnDFGJBnhVQhG":[52,49],"Gj3QmL769joJq8fszxqX2obCfHV3S4ffKPT3rm6xUe92":[40,35],"GkLRAjKa3b9gazDrc2wj6zVNbLWvibEKTuhsRp2i9Yni":[12,0],"GkYsHzF1h4uZ2nGxykLqYJvVK91n9eWHymA5FhTrmRns":[64,58],"GkpQYmJzR81VjuT4Gch8iF3LECvSEfVJomNqjUE1Ef8K":[88,58],"GkuEkzsEKyhgmomjiS9ZxruL9oc8tTWYQHaUv6Xa4Ch3":[56,30],"GmAuzYVHNBDTX3zGRwHqH85hnbBaxQdTbeVpyvAZpufN":[68,58],"GmgV3mnVohRz99rsnMNWNFqzop4oSgNv6Hx1kE7PKvYU":[24,14],"Gmw9GarCUcQNYnqePXNBREuLhcMUwXhQWZMAvxSUf6c2":[52,34],"Go3R344LB8hSYfpZJKs2LHQVRJE4zsm2KaSLXbRygYbd":[64,47],"GokfNYT1GH3c8BQrXoAJBypARuNFWcRG3xa5KQ8hKwPe":[60,42],"Goy1QEMaKzadc7y6hYPbNu5tzQPZMFRLM7oRehiuoeDc":[72,59],"GpXcoJ7jRzCEpoSLERYQCxDi9jQ8oJxCCKqjEMBnMUDQ":[44,38],"GpowxwT8wY9x2uFLWhZtL3ELFdAMnpBxTrpqFgnEukVn":[48,46],"GqFWgFDHj6fgahisFk8TEngmsEdkSxbmk8ZktpgW5LaS":[40,35],"GqsnwvnnwfvevfovAfRu8XrJwGietC8h8t4dwyQerbfq":[68,59],"GrZcGUJ7baE8r9KSmrNJAKtgYAMiD7p2YfxefkbgTng9":[100,84],"GraWXC11stqBmL86aJW8cBEncRhAFX86mX4E5pE2eKg1":[68,56],"GsTKJfSxEXvAvw3Vkw4cLzBFzBUSFSqvx7cW9qnsZFvV":[36,31],"Gsooc16Z2JNRxfcsfGY16pvJ3LGaBaEsFRR1ANDdixfW":[48,32],"GtU7wyz6vwTo7d82qNpFM6zsxWUnN7caxNMaxLwbwCEr":[68,43],"GtgtQLfqKjn3gaHuH7Fw64n49vr2DrYHiJAsSTNNscAE":[48,16],"GuKn8nEJwUPjBfxpwyq2MXU2JNrSpj7gqnKptCZeEk7j":[28,26],"Gv4WPocqzi59G7sbGiVWZAtLwDohpB8xSMG4kg41g3U5":[40,22],"GvWoMZaf8ZbTnKKxrTGV5aGAjofinufohuqT827waAEy":[72,55],"GvZQ6JUcGiw26huYVv2eDTgrgVh3rKtADPYLfBiznVda":[24,20],"GvkYeWeoxH2QzvDrv1Rqr8ZmmMmj6xxZbbBWptCDVX4h":[28,28],"Gwgg3usfEkB2dFZutVWsnCwTDEbTo3AncyYLdpjdNwE7":[44,34],"Gx3a1YzZCLrih1R9FPnqj7yzW2ekFWHVTCM76Zq41D63":[44,24],"Gx6SwGTbYAFrUeBRMgMrgLUKaeGNeCKzkULXdEpwPSwc":[24,19],"GydWayef5RL5qW5zfTjZAqd8c2gPBdctpQRFr9d73FHb":[20,19],"GyoEkrR3Zsfq1FayyXfEcKM2UtTEWztYAEQBZHxqEmTP":[48,46],"GzFTdRZgzMdaTc6nrkgx3eeRbRtsTy78mXNHvzMkYJAu":[40,31],"GzeThbXZzp7iSQBsCNGQU2CsA5YDi3Xr9Gdm4Sjj7gP8":[36,27],"GzjJmb2w3vt788rDHPNo8hCeXvd4fzu4n2jwkoTRBouv":[60,31],"H1xNkt47PQ8HCjUfhoUEtMTtxRRphsqdHXY4B7mP64oG":[52,43],"H28YbKcxkEekLDHmTYnRStkFBhpVdNckNzBQSvRiWHmR":[52,44],"H2qBXkxdFh3XiKbAkJVfRsucdVkf6uCwJe16TA5VZeNt":[60,27],"H4Aq4RPa5coDMtPydkm2WyY8gd3k5R32ieL5fo8QFBY4":[60,52],"H5YdwNcDt6DNeAL816CP3Kn3fYXFoigMw3zaASvZc4rA":[60,49],"H5aQjanQGz5Rrh4s6g4TNHTAXDzdr4czBtEJPdeDJ1jo":[80,68],"H6djbzHAiv46Wxy3iwqD7LA8ART8YbgrWyxQPCNhnLPE":[60,38],"H8MUh74GVNbSqGrYkZviws6xCmdVS3VZF1rbhE3gSESQ":[32,32],"H8kdiUSyvHbxshcFmRqWTB1HZkQHKcQcagQ56TzLe8ib":[48,33],"H9PJujAtZMZJ4tAoWeP4UDFjQtYSCsBtUfpipD4nti4D":[52,46],"H9Th9in92sTsCxiA6oBe49vUW8PPv84HrN2g4KfSVgnM":[36,26],"HCikGbQ6gUseeVTvjwe4GZx145hpd2JR57DcC1DjecrF":[48,43],"HCkbW7BPAcLQ7kFwg89F7swnmnGLF2HeQ4zMNqG1YYoy":[56,41],"HDitfpmCcy8WyJgNCQTrnZ8r71Nn7t7SjmVHnRwumGZi":[68,57],"HE35aDYTJHJ6KA7kLEXvENiRBX8c5UG5xHzgeKiXyQno":[64,41],"HE94g4Qp39SrUtBtLXnjEYPxbjxMb28u7JvrNwcqGryX":[56,55],"HELPwwfg5W9LmXv3axe43EY1YGJjfVf3CcVjA8BZM82P":[44,38],"HENUtcTb7dTxGHDYV9VeLTgHy1DKAMWLwLySym3EziCr":[56,48],"HEYsakKxLuuEWsSdH2cevwXJXdQ8tX75KT2SHWCkerHs":[52,49],"HEbMY624UhDGm1Qhy6neKSyi3bQjQ2RidSTyt7ARK8RW":[72,52],"HEgDHQD4cZsVu3QEjLLE49EQmuDoojxiZZbPywgdX9dE":[76,60],"HGHMEEHCfbVFjqB69Hu9oNW6SviukB8jUheEhYVZJKe2":[52,33],"HGkZ4BCw3mHYoo74ZwwzKJsSjnH8u5BW2poHVVMLTjuR":[40,32],"HGv9NFa8dQeCCZcua4vg4Tqa8FpTK8AWBAQqqAE3z81G":[60,0],"HJhcE1XDYTRoHaDWcfkmfGvJuDWPZLEK8YkuMw9FYpP3":[64,49],"HKpk8f7t6MB5gYb2uSP8R2LZabMRezKqgCAVSb7tmsqQ":[44,18],"HKu753Hd2F1nWLPvcNZHX6RAGSXkg6AtywiVvRqDXxcP":[64,42],"HM2hzFLTd5TAhejGFjaXAm8LLjdmnj7bqQrzpRTaawdo":[48,44],"HMtri4bE9Vs3pztgZRtU21M9CuPhHTanZXtiuiR2azKk":[60,41],"HMzBDqxq5as7JRyd8PfDyvqL1LNF9R7yX3YNUVi8xT9m":[4,3],"HNX6Tba28Y4o4vfU33s5HkYefZcS8xNKY3R8zwxwor4z":[28,24],"HPwNj9cotHgFyt1Z2MhKsQTU1w4LxT2R71GoSdGSrbvP":[60,45],"HQZpRZLSzgDdPc21U7nCxXpK2VjMh8U4PE5G3YE9H2Y4":[44,27],"HRCBqRPZWyxghiRvCG6qQsPEfnnXvDeMGQwXYTvuGkKr":[16,0],"HRddSLZVC1dH1uDvuKKqsJCN9pbihJEmVDAVtyL92JY9":[64,52],"HRjJjGh1T33Rv3TeD9oLLEnkZVVvezP47kXsCbdXgJfo":[40,35],"HRvSdkavd12ZVGnoaDV65ogTPks4pBERY7uuwLG3YxdY":[80,59],"HSPysEmZeXB1VsBnpemb1AdMMjv155m4tZ6LYq2uzWSd":[40,32],"HUoud6qywaWj8kZwdHRTbEPkKmskHa6Md1KNvF1JQFYF":[52,26],"HUtFMtq115zhNf1ecuHHhqhP3fJupC8vt1wkWevHf1Xr":[4,4],"HVD6ZDBgzjqYKyDLNadSkzev3qwSUnYEs6k81JktNuom":[60,51],"HWKwGWgWpnt1HYo7kAbtpuNzi6PovGXz3oWW54GDEQWc":[56,48],"HWvSRgESdWKDccWN91iRVQLN4rRyuCbuAHVWtPR1cJ1C":[28,24],"HXU9vqRYsM6jF8wGwXAAFQW1gFaSu31Ex6Q7d68SmpkW":[72,45],"HYW69eojAvAqiPfebT3S8yUTvTDHnssZbTq1TMCm5LfP":[44,24],"HZCUCLqV3P7QqG1oskLLMJW28zuckhxmRzEQ7UWaH2U4":[52,47],"HZX4MWsSDzRerGuV6kgtj5sGM3dcX9doaiN7qr5y9MAw":[72,61],"HZwbEfKY3TNJ9RPeAYAH6zABK8Jr3CX8xXrDS7fKsc9x":[56,35],"HaDrjvGfvneb4Rs5LyUThKY7PEh9QyBKsZoi7jNcyune":[48,39],"Hahg6pT1qcuRR4iJhuPoCCouwX4rh552ozgNJsHU9HyY":[8,0],"Has5UpRZwnb7TrhyFPDETgnGThLPe38BiMcqkvXBLsra":[60,55],"HcasCt3HSWS1J2YH6qcVy5WNiKuuRqQ4kekN429dbcMm":[8,4],"HeUNAoWmKWrteB8eJB8SnNn345pzM8ymfZARzVrvtKFF":[92,68],"HeeEbBAkuLzqxsFLcbKUfWmeNizywy2uzAfvRg63LFT2":[60,43],"HehQKDuoiNCbzovqzhjzSmNN2VBKYPgKLKyxNYqRsZhi":[60,18],"HfxzFiP19ymtxHagP4Zpga2zVo6ZgivpK6VkBKDowHRr":[48,38],"HgF66KCFTqcs55WAK9co7o9f4ZuPuXmFQTCKvRWz9A3H":[64,51],"HgK19TWcH4FcbtpWNdjU4gRQzBdS9GUokz4XbzTm9WP2":[12,0],"HgTXVg3dA51mGNSh9iPoeC6QLsy2cMEB8WPESRAQMBz7":[8,0],"HgmPwzNcY85HfrN3bYiqaypb6Nmf7ayaZEaivGY37913":[72,70],"Hgp3kh6Vv8iw5wHD86LqkW3H3JApJeW3F5XLaGXkZZW9":[28,20],"HhjxbH3vLpUNShQB34NuMCL1Qc3xoiNDbvALWrAMCCnb":[56,56],"HiVDGAGPSxxydKTY6BkjuLE3CyabGKyEuMMHc1yMw5Qq":[56,47],"HjT9tCUEFWrUXFR37ahB383QTF4u53KWx9J29EWRfzdi":[48,37],"Hjsgy7BuoUFo4WUr1sSvxSYK2oLf7hf2dYq818Fgc5Gh":[28,14],"HkSTpiQR4YTP29yRBSactwZCKh3fp7NoLHLQrMK58xRE":[52,39],"HkXUfo7jkpymbV2LuekirjvJDzXEREB1c8hfy6cPxgLy":[44,31],"Hkj4Y4QxFvyoCd2wzAswsDpwW4vD1vyC5vppVyDDhJ8F":[72,42],"HmEUD98iS9DFkt5dUjtrG3jDixVU53B3YPKPbGwSPrC5":[60,52],"HmSU5YJr4XK2SYdF6dxNXtF9PQRzbXXupUXCVEaJZX37":[40,0],"HngPeCBmKEooRBX4nMYYNvqic6QRDhXAmRAYEGD5WNr1":[8,8],"Hnp981DgpWig4dBQASkdJG8r7KzrgNRvGVUo2Em4ZTAJ":[48,30],"HoMBSLMokd6BUVDT4iGw21Tnxvp2G49MApewzGJr4rfe":[4,0],"HpMNvGvQ2MAQwKvSbgB6mdJvKENdZ4jVXQhM7Ed3zJAm":[64,51],"HqXHSTtrUUraYZ4xcPuze9pX8LbRaV4wQGQ92Y2L26vN":[56,27],"Hr3PCkpdBpotx7CL9P51Xi7Yj9mJVPvmEHKUmUFrNkr4":[44,30],"HrDkZBmgxaV1373agNpFArvae53uga91HYdRHLCEdrkQ":[60,42],"HsWUiXARLPhYitGMapLYyMdV7k27kW2xzy9Z6L77jKBC":[56,45],"Ht4Zd1QjJPkvEUsJUeFgkFNWoqwW49mZZesCVm4xKr94":[40,30],"HtC9DFLScd8PKewDrGzC2dZdMURZgEfNDv2ji43coc2":[40,25],"HtL6WWfAHCEQFHumzYPU3qupZzXth8D5jafz8v7tTgVy":[72,65],"HtNFmWY2Ua4zx2PLKq1uAxmm5iaLXKcV7oGkPk5dFYZ5":[60,53],"Hu7DW7BoXXuKbwaFJaAMEXpBv8pqBJPhfThMD96WHiJR":[44,38],"HuJHVhpsf9nF4vbTWjgqgcCf2h97eFf4DnhAe3txLERo":[44,39],"HvofH3GhdBkVdPctTWhiPmGKzsADqtUpr4tpLgss3NX":[52,50],"HwFvyMbGLkiTUaT66cfL1FnJ26c9VqtpqAg4UbWSXtdq":[48,42],"HxBkCVtiYAymCCv4EYakNDSCPgog3vBJMZx54dCceSyS":[52,0],"HxFyHVXiQMf1M5nFowvW7um7oZx3aR1qkcRJFQGBpo9A":[60,48],"HxnjZ5Qg59nupGGXVo77idUxfsiRXPcBbBt2hw3Nt99c":[44,42],"HxsJQAfgMFVcTqf7hfLBg8UzcCq4rQJdo7g61Z4i6ExG":[44,38],"Hy5Mano3fc6AZceKCCwooL1sb55KaucRTGAMxRxsZ6qL":[64,55],"HyCf5LyHfwnpnvwTQkfPWVdkqJJ2R2A8fBKb52m7cunf":[20,12],"HywMhn8fUgxVFYTnXdCCZGGVKK8QZ4Yz5AC2bgVaEVGQ":[32,24],"HzPFqFKsGRT3Yvd5Wgfng16c8q6e1bDe3W48fZbuuS9Q":[52,20],"HzzApCxMXFzUyeiFkJiVC7sK8De1tX8NfFKPzSU5TZ5N":[56,36],"J16NbAo8MJAthmm4kfrLdyTKFsWRTVn8Vaq8gue6eJMD":[48,38],"J1RpwhRqrLGUpuwazwHn5yVuUUjQXZWC5pVRoU2YqTYx":[72,62],"J1mnigj2PmzRCuLvjqBX3h6Lb5b6PoPt2Cvqu8g2wNG3":[60,55],"J2LPt6Pza5KcNfeou7WrqNKs7wu3v1rSeLnXY5iF7szh":[12,8],"J2NT3pjS4JWMMHb9Ks1rcycAEbLPSA4u7d1eMv87rAWC":[48,39],"J2RzbW77NveGcT2Rse8tRvKZ8f9dnN6bMuT8q6a1LBut":[56,30],"J2VXfywh2oc5eT1LAtSApAqUVB1zJypFTYKTBdJg7BLW":[72,64],"J3wqjuNuduVLKFrZscCwZtoy6F8HSJqbSFQJV3LSqHJ2":[52,42],"J4FGK7xXt6E5pfxBtQhGfEX1djdgsLNhdrei4s5ghSaX":[80,71],"J596ieDeK3eN42jD8Rj3LRwuLJTkGKs3Ju27LPBLhvqz":[16,0],"J5TxmqomwJVQJMSeD6pXo8vmV8nRRRF311D9BFWV8vyi":[60,51],"J5aJV7Pd4SZ2tGA6k5kmdHETXiWEZpNsrXZ8LBtRDUEq":[72,56],"J5dVAuWTHSppRogVgdinqaEHgkkzzKYWdkSRZue5zpvi":[44,33],"J63rQazpR3qLHBz5DQLg5NB5xKDWAe3rLxGjpnJmZvmp":[32,8],"J6YG6AQd1AqZVsvBU3VCuHwzweFYrQL4Wdrqhpur9pUu":[68,62],"J6zWiwZBuMreGBkKQ6eqkbdESgp7BjhjcHBdBfYMZ64a":[48,43],"J7v9ndmcoBuo9to2MnHegLnBkC9x3SAVbQBJo5MMJrN1":[540,450],"J8DQkLZArBMLSdGGfMSdarnUwk7EybVjfBU1wHpZbzG":[56,46],"J8d11HHB1ttEf6wJpFdUhXvpyZefykDpAq2UDcPJsunW":[52,16],"J9Y9xwDkqFgiLypFoFmpC9MtAqkt7B9CTWrLUvyGZfKV":[28,26],"J9wpknjG6QdFZf6KjVkoHcTwsXsfWx7KeisBawMeFHPD":[40,28],"JA5W7X7BxDTt4fQZtAMCFC5Jf8V8beZ1cYNFBDeUvpWo":[96,68],"JAeDQsQhiXVEByzT4eCrXJQGgttf12Gnk2bPYMUt3eBC":[28,18],"JBH2sxUUGfXSDaxrm1Dsh1cLpMJWiDzuyNkDWne4qRHV":[60,39],"JBVWoq5pDYFaQqpP5UEztm54GZVQzrQpWKnWWn56UK6b":[68,60],"JBtk3KGDoQXYedMDiUFDp6VXJr8MxY9tqXg6pZ8EkaT8":[64,39],"JCtrjaF9bagyryB5vC48yY7rMjPyLWEoDMGjkhFqjxBm":[64,46],"JDSSd8fRsY3skSKEE6KbLLz5SxURf7nDkf2caazkMP4G":[40,33],"JgcWmNdwrrmvJiuSo4apJCLx9MozajzKRTF5mQREiXH":[44,33],"JoeKdMCnk9rE2DkLnej7tw5repqdfCDetSLCUedhUVn":[40,38],"JpaDR41DYXTH6FksaSg8tgoXbJxstUb2b5m26sNQx55":[36,27],"KK56APewFFFgM7c2ehKY1Eky6DgdDpmaBH9521roatW":[32,27],"KdNhBD4WCm4Gd1fwi7Uf7Z3JD9KrZcnWWm8nSEZ6NEB":[72,54],"KhBHw8r7pwMqurHv7N8pc4T5CHUaKiFnXKzFdCmkQZV":[44,28],"L9hQqzWE7yv3uB8xT2TpKsB1BqukPg1yYN9e1ygAbM9":[36,32],"LPGATvYWNFLrtAjyfk3Hyfdngp24MFX7S41T238rBDu":[48,32],"M7Pcv3j8KpX8ZAkeSsvJnexgKrZbBAaMEcRTvf6t2Em":[72,56],"NHtR8X7dmwtCagm1FuuC6ngQ3wv52uJYqFvA79G47MX":[64,42],"NNetet8BiymZxMBWLRPCcNGcBPZDBeEcpgtfTSwdFPX":[136,67],"Qfp5wD5TwLiecZZP64cn3SwfqvF7W4fzo2tjEy2c1MR":[88,63],"QxgXRHsBe4vazdrboKaRtXMuf1EhTVvF68tVcfepRr1":[48,40],"RAbGPTmaLVn1HkP38tqKYjGceejMWiyfDPEhenjH1Aw":[40,37],"RmQNKxg8et4ovHsa6Cs6GvwvPb8W1wwspUnyyFw9gs5":[48,39],"SQJYmcjgo1bwJe2YxJwRDAH1JKFdrQM4AfLzuLi5TME":[44,34],"ScN8WkfK7c5nmNvNh7SbFTQcNyw5poXv97h5KRFBRWL":[44,30],"T6CVCqL6Mcea4wjRgvoZRkDSexzNP6fuNcyEdQZH21H":[56,56],"TS7mNjzAkgqKSR3PhSXhrqjVCkyAmUta3skRF1NbRUx":[24,16],"TxChgiaHwnkdT18sBnSepLE5sGk7vsQ4CZnhwiHUMQw":[88,74],"URnkWZGiuB7jXbfCSuNSwir1qkn7sXjiSPeLPaXys7b":[40,34],"UdAZ7oz1WshdwyimF6e2VXiy1eSJ6UdHSRng9yRLtgY":[48,36],"UkQCtmg2gSygRMRJq3wHT8fZahqYzwRp2rHXE2hCTX1":[32,24],"VCRrRTgSjDLHvo6UQKXy8VQbNVG2ioHNUEyS7oB7u3X":[68,54],"VTAZqz5HadKsUWyavErx3hhUeaDPerPVDssjB69hP8b":[8,0],"VisixkGG8H2htLvq2EKiywbHXVjNPQiszUKHeu9r6bd":[56,53],"WhXimQKBLiMUAWBaVVenpweVJNSahrFZdPLd8hr5Tfq":[48,45],"XctiDwBwYbZomH3pfBKca69BKrxjEQtFbdu6TXw9fe3":[48,35],"YKhfczqyMeHPMSJzcs8JiAVCtu3iieampLr3j63yTjk":[56,50],"YYYYW8eKkmwQFhVGUKdBAnDQPuhMTpG7zwm9nikNndC":[4,0],"YpGaosZwUmt5p8gXbdGriya7zKZvU6439CDXQSS5Gcb":[52,52],"YpopmpJ5ryYnLZKD7a2dEbPdPiiSLRARWVj3oAmgWLt":[48,41],"Z36ZMwALNp9sLgBt6nUhb3NAFAJbUp9kCDM3an7Xhb9":[40,0],"ZDCJDkoBMTXpf8zsfQzbLeTfAus1qaxiFHnANseQrmA":[4,4],"aTPi9R1prHHwRf9GCSgK4yLZiCZLs7evMtZP7EPNKw4":[40,23],"bj53eWLx461E3m27qmHGtJE4NZxefhvZUoewioSavqH":[32,32],"cDK4eZakyrZPT4fdhpPUt8q4EekNEwwGEz84LFnQb2S":[56,41],"dyEBiLy8Tty9nMeV6aUYU6qHVobsm4YNQmTGRhyvUB5":[48,18],"eMVkN9G77VnE8QsLtMvArMcsM93cytxowJxbbwwmzWV":[48,38],"eoKpUABi59aT4rR9HGS3LcMecfut9x7zJyodWWP43YQ":[4,4],"eopuRXqXh8HxG5Y7U7oGCN8CpzPudLx3j1CWz9WBDGR":[4,0],"f4fN8n4zEtVnw2fdQD8y8GBwzN5Asc4X6FzuFi1kEdt":[76,53],"fRiGutrC4h4ZdYVE65g3pCeJNDg2g9j21AvMjhDMwW8":[48,37],"fw7Td8AG8Sfha5ZQQnwaKAYFnskbbLBPJrieP4Puxwy":[52,45],"g1jyzFsCaJZmfkCWxcAF8Ay9DFm82cSrM4yfi8aDkgt":[36,29],"gVuUFY5MPjhayfUVTKTuBidEwFxYbQ6sgFMsSDNpRiU":[8,4],"hQBS6cu8RHkXcCzE6N8mQxhgrtbNy4kivoRjTMzF2cA":[28,21],"hTzYQHBSeqW7gjTDsZhUTe3NGGJjTJgwF9sTikSuwiY":[56,51],"iMZEU6pbrLTCHZsaw6HkwABYvMTfn9pND9AqQgaEyuo":[80,62],"ibwMFhkkeMTn9746FERTdb7rGuQwVcRXDbNYXB4QB8q":[76,59],"irmVepLsWGTFjCrqGpkkfjYexrvCgwJ4CTSUQVjsFs7":[60,47],"iwf4VCAm2WYfHMWkGC6p7PScW5vihfVrXs8UTnMuHuc":[44,35],"jw3RWGkKoRDRnvscvmXHPkATtay7oeoTBMvQinprcyc":[40,34],"kffvkDohANNa2rpj8Ti6KWZctCX3Ci6Rj1SnGHx2r63":[32,26],"mETnAkTMdDN41d9wSPYJWDFu7xehfoHyT5py2thcxHB":[52,52],"mFJG277eG7EFS7Zu2UU5BkFZQW7PpAVfjMaFsTqXAUq":[60,47],"markiLNTC3FuWYGXKz8h9XpbwJbVQhzuV5U3bfpPc64":[68,58],"oQcqLhbpxQdDPgFP2d2HBNWqJFKGk9JfHw6kxyLgfK2":[44,41],"rAEVgLDieWcb3N975fNsLbmpNenL2simANdvk35iLeh":[44,35],"tNZMLNhcUGk8pr5LVk2R5eV7jxiuU7SfDNa6uNbHeTF":[72,65],"uknL1QAVVNLT6FEDxc21yN9Q8jykDCJDvCyS1f1qUkJ":[40,30],"vWoxJrxUz4JjxXiRKoMKZ1BfkDFcDcDVB55Lx4isemS":[36,35],"vav8fy4UyYKf91g9uFZybjwZh1VS6hubfaKyFtbYcvT":[48,29],"x31Ldohp254zduxhHMHNR3JbXZYrxgjEaWTvd4vuxZ6":[72,46],"xfCpo4ouRs5BP3WY5BdWhbr41pQxYGcXxz1sFyzPsZr":[36,19],"ygkCkgip2MbboVLSnq6FpEz7N89nqiaTCKBCSexamR8":[52,42],"zKuryCTzgvwoyDZTTh4NuiT9D9bpMHG33tTRyKKZUUT":[60,60],"zjG7sHeExhC7tfLZwTJwHH3zzDSqDcwRVe2LdXg389j":[44,40]},"range":{"firstSlot":83900256,"lastSlot":83992896}}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	firstSlot := uint64(2)
	lastSlot := uint64(3)
	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)
	identity := pubKey
	_, err := client.GetBlockProductionWithOpts(
		context.Background(),
		&GetBlockProductionOpts{
			Commitment: CommitmentFinalized,
			Range: &SlotRangeRequest{
				FirstSlot: firstSlot,
				LastSlot:  &lastSlot,
			},
			Identity: &identity,
		},
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBlockProduction",
			"params": []any{
				map[string]any{
					"commitment": string(CommitmentFinalized),
					"range": map[string]any{
						"firstSlot": float64(firstSlot),
						"lastSlot":  float64(lastSlot),
					},
					"identity": identity.String(),
				},
			},
		},
		reqBody,
	)
}

func TestClient_GetBlockCommitment(t *testing.T) {
	responseBody := `{"commitment":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,44854495719374,0,51599979318189,5070972605440,140323113958535,169550804919131,272061505737107,860587424880950,1374732609383053,2334359721325133,4664454087479672,10122947678661428,52107037802932750],"totalStake":73611541921665680}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	block := 33

	out, err := client.GetBlockCommitment(
		context.Background(),
		uint64(block),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBlockCommitment",
			"params": []any{
				float64(block),
			},
		},
		reqBody,
	)

	expected := map[string]any{
		"commitment": []any{
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("0"),
			stdjson.Number("44854495719374"),
			stdjson.Number("0"),
			stdjson.Number("51599979318189"),
			stdjson.Number("5070972605440"),
			stdjson.Number("140323113958535"),
			stdjson.Number("169550804919131"),
			stdjson.Number("272061505737107"),
			stdjson.Number("860587424880950"),
			stdjson.Number("1374732609383053"),
			stdjson.Number("2334359721325133"),
			stdjson.Number("4664454087479672"),
			stdjson.Number("10122947678661428"),
			stdjson.Number("52107037802932750"),
		},
		"totalStake": stdjson.Number("73611541921665680"),
	}

	got := mustJSONToInterfaceWithUseNumber(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetBlocks(t *testing.T) {
	responseBody := `[83993598,83993599,83993600]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	startSlot := 1
	endSlot := uint64(33)
	out, err := client.GetBlocks(
		context.Background(),
		uint64(startSlot),
		&endSlot,
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBlocks",
			"params": []any{
				float64(startSlot),
				float64(endSlot),
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetBlocksWithLimit(t *testing.T) {
	responseBody := `[83993712,83993713]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	startSlot := 1
	limit := uint64(10)
	out, err := client.GetBlocksWithLimit(
		context.Background(),
		uint64(startSlot),
		limit,
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBlocksWithLimit",
			"params": []any{
				float64(startSlot),
				float64(limit),
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetBlockTime(t *testing.T) {
	responseBody := `1625230849`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	block := 55
	out, err := client.GetBlockTime(
		context.Background(),
		uint64(block),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getBlockTime",
			"params": []any{
				float64(block),
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetClusterNodes(t *testing.T) {
	responseBody := `[{"featureSet":3580551090,"gossip":"34.147.255.155:8000","pubkey":"hyp3Eo67t6FgeuWg5Qxbeme8NPXJPXXdKT4iJ4DsLf2","pubsub":"34.147.255.155:8900","rpc":"34.147.255.155:8899","shredVersion":50093,"tpu":"34.147.255.155:8009","tpuQuic":"34.147.255.155:8015","version":"1.17.22"},{"featureSet":3746964731,"gossip":"162.19.222.39:8001","pubkey":"EvnRmnMrd69kFdbLMxWkTn1icZ7DCceRhvmb2SJXqDo4","pubsub":"162.19.222.39:8900","rpc":"162.19.222.39:8899","shredVersion":50093,"tpu":"208.91.106.87:8005","tpuQuic":"208.91.106.87:8011","version":"1.17.27"},{"featureSet":3746964731,"gossip":"205.209.104.74:8000","pubkey":"J87afqF2bDQQLTQpks4SdF7hXPr96SPTdJ28UJXXWr9N","pubsub":"205.209.104.74:8900","rpc":"205.209.104.74:8899","shredVersion":50093,"tpu":"205.209.104.74:8003","tpuQuic":"205.209.104.74:8009","version":"1.17.27"}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetClusterNodes(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getClusterNodes",
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetClusterNodes_AllFields(t *testing.T) {
	responseBody := `[{"pubkey":"hyp3Eo67t6FgeuWg5Qxbeme8NPXJPXXdKT4iJ4DsLf2","gossip":"127.0.0.1:8000","tpu":null,"tpuQuic":"127.0.0.1:8009","tpuForwards":null,"tpuForwardsQuic":"127.0.0.1:8010","tpuVote":"127.0.0.1:8005","serveRepair":"127.0.0.1:8008","rpc":"127.0.0.1:8899","pubsub":"127.0.0.1:8900","version":"2.2.1","featureSet":3580551090,"shredVersion":50093,"clientId":"Agave"}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetClusterNodes(context.Background())
	require.NoError(t, err)

	require.Len(t, out, 1)
	node := out[0]

	assert.Equal(t, solana.MustPublicKeyFromBase58("hyp3Eo67t6FgeuWg5Qxbeme8NPXJPXXdKT4iJ4DsLf2"), node.Pubkey)
	assert.Equal(t, pointer.ToString("127.0.0.1:8000"), node.Gossip)
	assert.Nil(t, node.TPU)
	assert.Equal(t, pointer.ToString("127.0.0.1:8009"), node.TPUQUIC)
	assert.Nil(t, node.TPUForwards)
	assert.Equal(t, pointer.ToString("127.0.0.1:8010"), node.TPUForwardsQUIC)
	assert.Equal(t, pointer.ToString("127.0.0.1:8005"), node.TPUVote)
	assert.Equal(t, pointer.ToString("127.0.0.1:8008"), node.ServeRepair)
	assert.Equal(t, pointer.ToString("127.0.0.1:8899"), node.RPC)
	assert.Equal(t, pointer.ToString("127.0.0.1:8900"), node.PubSub)
	assert.Equal(t, pointer.ToString("2.2.1"), node.Version)
	require.NotNil(t, node.FeatureSet)
	assert.Equal(t, uint32(3580551090), *node.FeatureSet)
	assert.Equal(t, uint16(50093), node.ShredVersion)
	assert.Equal(t, pointer.ToString("Agave"), node.ClientID)
}

func TestClient_GetClusterNodes_BackwardCompatible(t *testing.T) {
	// Old response without new fields should still parse
	responseBody := `[{"pubkey":"hyp3Eo67t6FgeuWg5Qxbeme8NPXJPXXdKT4iJ4DsLf2","gossip":"127.0.0.1:8000","tpu":"127.0.0.1:8003","tpuQuic":"127.0.0.1:8009","rpc":"127.0.0.1:8899","version":"1.17.22","featureSet":3580551090,"shredVersion":50093}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetClusterNodes(context.Background())
	require.NoError(t, err)

	require.Len(t, out, 1)
	node := out[0]
	assert.Nil(t, node.TPUForwards)
	assert.Nil(t, node.TPUForwardsQUIC)
	assert.Nil(t, node.TPUVote)
	assert.Nil(t, node.ServeRepair)
	assert.Nil(t, node.ClientID)
}

func TestClient_GetEpochInfo(t *testing.T) {
	responseBody := `{"absoluteSlot":83994151,"blockHeight":69218302,"epoch":207,"slotIndex":93895,"slotsInEpoch":432000,"transactionCount":27287000257}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetEpochInfo(
		context.Background(),
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getEpochInfo",
			"params": []any{
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := map[string]any{
		"absoluteSlot":     8.3994151e+07,
		"blockHeight":      6.9218302e+07,
		"epoch":            207.0,
		"slotIndex":        93895.0,
		"slotsInEpoch":     432000.0,
		"transactionCount": 27287000257.0,
	}

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetEpochSchedule(t *testing.T) {
	responseBody := `{"firstNormalEpoch":14,"firstNormalSlot":524256,"leaderScheduleSlotOffset":432000,"slotsPerEpoch":432000,"warmup":true}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetEpochSchedule(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getEpochSchedule",
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetFirstAvailableBlock(t *testing.T) {
	responseBody := `39368303`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetFirstAvailableBlock(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getFirstAvailableBlock",
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetGenesisHash(t *testing.T) {
	responseBody := `"4uhcVJyU9pJkvQyS88uRDiswHXSCkY3zQawwpjk2NsNY"`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetGenesisHash(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getGenesisHash",
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetHealth(t *testing.T) {
	responseBody := `"ok"`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetHealth(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getHealth",
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetIdentity(t *testing.T) {
	responseBody := `{"identity":"DMeohMfD3JzmYZA34jL9iiTXp5N7tpAR3rAoXMygdH3U"}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetIdentity(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getIdentity",
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetInflationGovernor(t *testing.T) {
	responseBody := `{"foundation":0,"foundationTerm":0,"initial":0.15,"taper":0.15,"terminal":0.015}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetInflationGovernor(
		context.Background(),
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getInflationGovernor",
			"params": []any{
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetInflationRate(t *testing.T) {
	responseBody := `{"epoch":207,"foundation":0,"total":0.1403151524615605,"validator":0.1403151524615605}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetInflationRate(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getInflationRate",
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetInflationRate_EpochAsUint64(t *testing.T) {
	responseBody := `{"epoch":207,"foundation":0,"total":0.1403151524615605,"validator":0.1403151524615605}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetInflationRate(context.Background())
	require.NoError(t, err)

	assert.Equal(t, uint64(207), out.Epoch)
	assert.Equal(t, 0.1403151524615605, out.Total)
	assert.Equal(t, 0.1403151524615605, out.Validator)
	assert.Equal(t, float64(0), out.Foundation)
}

func TestClient_GetInflationReward(t *testing.T) {
	// TODO: add test with real value
	responseBody := `[null]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)
	keys := []solana.PublicKey{
		pubKey,
	}
	epoch := uint64(56)
	opts := GetInflationRewardOpts{
		Commitment: CommitmentFinalized,
		Epoch:      &epoch,
	}

	out, err := client.GetInflationReward(
		context.Background(),
		keys,
		&opts,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getInflationReward",
			"params": []any{
				[]any{
					pubkeyString,
				},
				map[string]any{
					"commitment": string(CommitmentFinalized),
					"epoch":      float64(epoch),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetInflationReward_WithCommissionBps(t *testing.T) {
	responseBody := `[{"epoch":100,"effectiveSlot":43200000,"amount":2500000,"postBalance":1002500000,"commission":5,"commissionBps":500}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubKey := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")
	out, err := client.GetInflationReward(context.Background(), []solana.PublicKey{pubKey}, nil)
	require.NoError(t, err)

	require.Len(t, out, 1)
	require.NotNil(t, out[0])
	assert.Equal(t, uint64(100), out[0].Epoch)
	assert.Equal(t, uint64(43200000), out[0].EffectiveSlot)
	assert.Equal(t, uint64(2500000), out[0].Amount)
	assert.Equal(t, uint64(1002500000), out[0].PostBalance)
	require.NotNil(t, out[0].Commission)
	assert.Equal(t, uint8(5), *out[0].Commission)
	require.NotNil(t, out[0].CommissionBps)
	assert.Equal(t, uint16(500), *out[0].CommissionBps)
}

func TestClient_GetInflationReward_WithoutCommissionBps(t *testing.T) {
	responseBody := `[{"epoch":100,"effectiveSlot":43200000,"amount":2500000,"postBalance":1002500000,"commission":5}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubKey := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")
	out, err := client.GetInflationReward(context.Background(), []solana.PublicKey{pubKey}, nil)
	require.NoError(t, err)

	require.Len(t, out, 1)
	require.NotNil(t, out[0])
	require.NotNil(t, out[0].Commission)
	assert.Equal(t, uint8(5), *out[0].Commission)
	assert.Nil(t, out[0].CommissionBps)
}

func TestClient_GetLargestAccounts(t *testing.T) {
	responseBody := `{"context":{"slot":83995022},"value":[{"address":"4Rf9mGD7FeYknun5JczX5nGLTfQuS1GRjNVfkEMKE92b","lamports":398178060209179300},{"address":"KchK7WTjPzq9QL5aCwnV1dLsT8rFjruS1Zfzamxus9G","lamports":215100454508495000},{"address":"8oRw7qpj6XgLGXYCDuNoTMCqoJnDd6A8LTpNyqApSfkA","lamports":99999674507283220},{"address":"9oKrJ9iiEnCC7bewcRFbcdo4LKL2PhUEqcu8gH2eDbVM","lamports":97721650553633650},{"address":"3ANJb42D3pkVtntgT6VtW2cD3icGVyoHi2NGwtXYHQAs","lamports":91160815129021260},{"address":"K7DbiDcRngs4KY3KxSUcMFNEzXW7iQgi3zFzerXYYDZ","lamports":80000000000000000},{"address":"mvines9iiHiQTysrwkJjGf2gb9Ex9jXJX8ns3qwf2kN","lamports":53925298123552904},{"address":"71bhKKL89U3dNHzuZVZ7KarqV6XtHEgjXjvJTsguD11B","lamports":20949230980018784},{"address":"57DPUrAncC4BUY7KBqRMCQUt4eQeMaJWpmLQwsL35ojZ","lamports":18210921605995270},{"address":"hQBS6cu8RHkXcCzE6N8mQxhgrtbNy4kivoRjTMzF2cA","lamports":18191952118880490},{"address":"5vxoRv2P12q4K4cWPCJkvPjg6jYnuCYxzF3juJZJiwba","lamports":14225826149332328},{"address":"2tZoLFgcbeW8Howq8QMRnExvuwHFUeEnx9ZhHq2qX77E","lamports":10099331225079048},{"address":"5NH47Zk9NAzfbtqNpUtn8CQgNZeZE88aa2NRpfe7DyTD","lamports":10000060317056686},{"address":"4xxV5Svt3LPsDv81seuqKB4QXxwhdQiFXzbj9GNYXkEr","lamports":10000000000000000},{"address":"GoCxdowvFindZVAXP3QsKRP3rR2LZBNXWwp3FB1yZznF","lamports":9796480999955000},{"address":"7arfejY2YxX9QrmzHrhu3rG3HofjMqKtfBzQLf8s3Wop","lamports":5465066164230830},{"address":"5TkrtJfHoX85sti8xSVvfggVV9SDvhjYjiXe9PqMJVN9","lamports":5384143441736968},{"address":"123vij84ecQEKUvQ7gYMKxKwKF6PbYSzCzzURYA4xULY","lamports":4350560741967702},{"address":"7vYe2KRUL2sbqSqbCn4UCvn2taaTJWvo3HBsPjZcEogG","lamports":3983999997415000},{"address":"7aeNmoVKnbxUSZGukYz2Gyr3UazXpaxATNszKu8XMW1k","lamports":3324774979081580}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	filter := LargestAccountsFilterCirculating
	out, err := client.GetLargestAccounts(
		context.Background(),
		CommitmentFinalized,
		filter,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getLargestAccounts",
			"params": []any{
				map[string]any{
					"commitment": string(CommitmentFinalized),
					"filter":     string(filter),
				},
			},
		},
		reqBody,
	)

	expected := &GetLargestAccountsResult{
		RPCContext: RPCContext{
			Context: Context{
				Slot: 83995022,
			},
		},
		Value: []LargestAccountsResult{
			{
				Address:  solana.MustPublicKeyFromBase58("4Rf9mGD7FeYknun5JczX5nGLTfQuS1GRjNVfkEMKE92b"),
				Lamports: 398178060209179300,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("KchK7WTjPzq9QL5aCwnV1dLsT8rFjruS1Zfzamxus9G"),
				Lamports: 215100454508495000,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("8oRw7qpj6XgLGXYCDuNoTMCqoJnDd6A8LTpNyqApSfkA"),
				Lamports: 99999674507283220,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("9oKrJ9iiEnCC7bewcRFbcdo4LKL2PhUEqcu8gH2eDbVM"),
				Lamports: 97721650553633650,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("3ANJb42D3pkVtntgT6VtW2cD3icGVyoHi2NGwtXYHQAs"),
				Lamports: 91160815129021260,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("K7DbiDcRngs4KY3KxSUcMFNEzXW7iQgi3zFzerXYYDZ"),
				Lamports: 80000000000000000,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("mvines9iiHiQTysrwkJjGf2gb9Ex9jXJX8ns3qwf2kN"),
				Lamports: 53925298123552904,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("71bhKKL89U3dNHzuZVZ7KarqV6XtHEgjXjvJTsguD11B"),
				Lamports: 20949230980018784,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("57DPUrAncC4BUY7KBqRMCQUt4eQeMaJWpmLQwsL35ojZ"),
				Lamports: 18210921605995270,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("hQBS6cu8RHkXcCzE6N8mQxhgrtbNy4kivoRjTMzF2cA"),
				Lamports: 18191952118880490,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("5vxoRv2P12q4K4cWPCJkvPjg6jYnuCYxzF3juJZJiwba"),
				Lamports: 14225826149332328,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("2tZoLFgcbeW8Howq8QMRnExvuwHFUeEnx9ZhHq2qX77E"),
				Lamports: 10099331225079048,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("5NH47Zk9NAzfbtqNpUtn8CQgNZeZE88aa2NRpfe7DyTD"),
				Lamports: 10000060317056686,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("4xxV5Svt3LPsDv81seuqKB4QXxwhdQiFXzbj9GNYXkEr"),
				Lamports: 10000000000000000,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("GoCxdowvFindZVAXP3QsKRP3rR2LZBNXWwp3FB1yZznF"),
				Lamports: 9796480999955000,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("7arfejY2YxX9QrmzHrhu3rG3HofjMqKtfBzQLf8s3Wop"),
				Lamports: 5465066164230830,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("5TkrtJfHoX85sti8xSVvfggVV9SDvhjYjiXe9PqMJVN9"),
				Lamports: 5384143441736968,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("123vij84ecQEKUvQ7gYMKxKwKF6PbYSzCzzURYA4xULY"),
				Lamports: 4350560741967702,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("7vYe2KRUL2sbqSqbCn4UCvn2taaTJWvo3HBsPjZcEogG"),
				Lamports: 3983999997415000,
			},
			{
				Address:  solana.MustPublicKeyFromBase58("7aeNmoVKnbxUSZGukYz2Gyr3UazXpaxATNszKu8XMW1k"),
				Lamports: 3324774979081580,
			},
		},
	}

	assert.Equal(t, expected, out)
}

func TestClient_GetLargestAccountsWithOpts_SortResults(t *testing.T) {
	responseBody := `{"context":{"slot":83995022},"value":[{"address":"4Rf9mGD7FeYknun5JczX5nGLTfQuS1GRjNVfkEMKE92b","lamports":398178060209179300}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	sortResults := true
	out, err := client.GetLargestAccountsWithOpts(
		context.Background(),
		&GetLargestAccountsOpts{
			Commitment:  CommitmentFinalized,
			Filter:      LargestAccountsFilterCirculating,
			SortResults: &sortResults,
		},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getLargestAccounts",
			"params": []any{
				map[string]any{
					"commitment":  string(CommitmentFinalized),
					"filter":      string(LargestAccountsFilterCirculating),
					"sortResults": true,
				},
			},
		},
		reqBody,
	)

	require.NotNil(t, out)
	require.Len(t, out.Value, 1)
	assert.Equal(t, solana.MustPublicKeyFromBase58("4Rf9mGD7FeYknun5JczX5nGLTfQuS1GRjNVfkEMKE92b"), out.Value[0].Address)
	assert.Equal(t, uint64(398178060209179300), out.Value[0].Lamports)
}

func TestClient_GetLeaderSchedule(t *testing.T) {
	responseBody := `{"DsaF77cCADh79q7HPfz5TrWPfEmD5Gw1c15zSm4eaFyt":[128,129,130,131,9480,9481,9482,9483,9752,9753,9754,9755,16272,16273,16274,16275,19860,19861,19862,19863,19932,19933,19934,19935,26616,26617,26618,26619,28856,28857,28858,28859,36556,36557,36558,36559,37500,37501,37502,37503,47220,47221,47222,47223,58436,58437,58438,58439,79524,79525,79526,79527,90452,90453,90454,90455,90952,90953,90954,90955,91900,91901,91902,91903,102772,102773,102774,102775,103568,103569,103570,103571,111164,111165,111166,111167,117068,117069,117070,117071,123116,123117,123118,123119,136224,136225,136226,136227,145072,145073,145074,145075,146124,146125,146126,146127,148824,148825,148826,148827,158400,158401,158402,158403,158792,158793,158794,158795,161988,161989,161990,161991,163548,163549,163550,163551,167528,167529,167530,167531,174584,174585,174586,174587,176388,176389,176390,176391,184700,184701,184702,184703,186132,186133,186134,186135,199876,199877,199878,199879,201568,201569,201570,201571,205376,205377,205378,205379,207452,207453,207454,207455,223384,223385,223386,223387,225772,225773,225774,225775,255776,255777,255778,255779,256640,256641,256642,256643,262364,262365,262366,262367,269128,269129,269130,269131,272920,272921,272922,272923,274180,274181,274182,274183,293660,293661,293662,293663,303004,303005,303006,303007,317092,317093,317094,317095,323184,323185,323186,323187,323252,323253,323254,323255,328216,328217,328218,328219,333508,333509,333510,333511,336908,336909,336910,336911,337036,337037,337038,337039,341392,341393,341394,341395,341848,341849,341850,341851,351972,351973,351974,351975,363532,363533,363534,363535,397416,397417,397418,397419,398756,398757,398758,398759,414788,414789,414790,414791,428144,428145,428146,428147,428432,428433,428434,428435,430140,430141,430142,430143]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	epoch := uint64(333)
	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)
	identity := pubKey

	out, err := client.GetLeaderScheduleWithOpts(
		context.Background(),
		&GetLeaderScheduleOpts{
			Epoch:      &epoch,
			Commitment: CommitmentFinalized,
			Identity:   &identity,
		},
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getLeaderSchedule",
			"params": []any{
				float64(epoch),
				map[string]any{
					"commitment": string(CommitmentFinalized),
					"identity":   string(identity.String()),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetMaxRetransmitSlot(t *testing.T) {
	responseBody := `83996101`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetMaxRetransmitSlot(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getMaxRetransmitSlot",
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetMaxShredInsertSlot(t *testing.T) {
	responseBody := `83996150`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetMaxShredInsertSlot(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getMaxShredInsertSlot",
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetMinimumBalanceForRentExemption(t *testing.T) {
	responseBody := `70490880`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	dataSize := uint64(1000)
	out, err := client.GetMinimumBalanceForRentExemption(
		context.Background(),
		dataSize,
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getMinimumBalanceForRentExemption",
			"params": []any{
				float64(dataSize),
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetMultipleAccounts(t *testing.T) {
	responseBody := `{"context":{"slot":83996178},"value":[{"data":["","base64"],"executable":true,"lamports":19039980000,"owner":"11111111111111111111111111111111","rentEpoch":207}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)
	out, err := client.GetMultipleAccounts(
		context.Background(),
		pubKey,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getMultipleAccounts",
			"params": []any{
				[]any{pubkeyString},
			},
		},
		reqBody,
	)

	expected := &GetMultipleAccountsResult{
		RPCContext: RPCContext{
			Context: Context{
				Slot: 83996178,
			},
		},
		Value: []*Account{
			{
				Lamports: 19039980000,
				Owner:    solana.SystemProgramID,
				Data: &DataBytesOrJSON{
					asDecodedBinary: solana.Data{
						Content:  []byte{},
						Encoding: solana.EncodingBase64,
					},
					rawDataEncoding: solana.EncodingBase64,
				},
				Executable: true,
				RentEpoch:  big.NewInt(207),
			},
		},
	}

	assert.Equal(t, expected, out)
}

func TestClient_GetProgramAccounts(t *testing.T) {
	responseBody := `[{"account":{"data":["dGVzdA==","base64"],"executable":true,"lamports":2039280,"owner":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","rentEpoch":206},"pubkey":"7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	offset := uint64(13)
	length := uint64(30)
	opts := GetProgramAccountsOpts{
		Commitment: CommitmentFinalized,
		Encoding:   solana.EncodingBase58,
		DataSlice: &DataSlice{
			Offset: &offset,
			Length: &length,
		},
		Filters: []RPCFilter{
			{
				// TODO: make an actual example:
				Memcmp: &RPCFilterMemcmp{
					Offset: offset,
					Bytes:  pubKey[:],
				},
			},
		},
	}
	out, err := client.GetProgramAccountsWithOpts(
		context.Background(),
		pubKey,
		&opts,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getProgramAccounts",
			"params": []any{
				pubkeyString,
				map[string]any{
					"encoding":   string(solana.EncodingBase58),
					"commitment": string(CommitmentFinalized),
					"dataSlice": map[string]any{
						"offset": float64(offset),
						"length": float64(length),
					},
					"filters": []any{
						map[string]any{
							"memcmp": map[string]any{
								"bytes":  "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932",
								"offset": float64(offset),
							},
						},
					},
				},
			},
		},
		reqBody,
	)

	expected := GetProgramAccountsResult{
		{
			Pubkey: solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"),
			Account: &Account{
				Lamports: 2039280,
				Owner:    solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
				Data: &DataBytesOrJSON{
					asDecodedBinary: solana.Data{
						Content:  []byte{0x74, 0x65, 0x73, 0x74},
						Encoding: solana.EncodingBase64,
					},
					rawDataEncoding: solana.EncodingBase64,
				},
				Executable: true,
				RentEpoch:  big.NewInt(206),
			},
		},
	}

	assert.Equal(t, expected, out)
}

func TestClient_GetProgramAccountsWithContext(t *testing.T) {
	responseBody := `{"context":{"slot":83986105,"apiVersion":"2.2.1"},"value":[{"account":{"data":["dGVzdA==","base64"],"executable":true,"lamports":2039280,"owner":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","rentEpoch":206},"pubkey":"7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	out, err := client.GetProgramAccountsWithContext(
		context.Background(),
		pubKey,
		&GetProgramAccountsOpts{
			Commitment: CommitmentFinalized,
		},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getProgramAccounts",
			"params": []any{
				pubkeyString,
				map[string]any{
					"encoding":    "base64",
					"commitment":  string(CommitmentFinalized),
					"withContext": true,
				},
			},
		},
		reqBody,
	)

	// Verify context is returned (not silently dropped)
	assert.Equal(t, uint64(83986105), out.Context.Slot)
	require.NotNil(t, out.Context.ApiVersion)
	assert.Equal(t, "2.2.1", *out.Context.ApiVersion)

	// Verify accounts
	require.Len(t, out.Value, 1)
	assert.Equal(t, solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"), out.Value[0].Pubkey)
	assert.Equal(t, uint64(2039280), out.Value[0].Account.Lamports)
}

func TestClient_GetProgramAccountsWithOpts_SortResults(t *testing.T) {
	responseBody := `[{"account":{"data":["dGVzdA==","base64"],"executable":true,"lamports":2039280,"owner":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","rentEpoch":206},"pubkey":"7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	sortResults := true
	out, err := client.GetProgramAccountsWithOpts(
		context.Background(),
		pubKey,
		&GetProgramAccountsOpts{
			SortResults: &sortResults,
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
					"encoding":    "base64",
					"sortResults": true,
				},
			},
		},
		reqBody,
	)

	require.Len(t, out, 1)
}

func TestClient_GetRecentPerformanceSamples(t *testing.T) {
	responseBody := `[{"numSlots":84,"numTransactions":90402,"samplePeriodSecs":60,"slot":83998844}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	limit := uint(1002)
	out, err := client.GetRecentPerformanceSamples(
		context.Background(),
		&limit,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getRecentPerformanceSamples",
			"params": []any{
				float64(limit),
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetRecentPerformanceSamples_WithNonVoteTx(t *testing.T) {
	// Based on Agave test fixture from response.rs
	responseBody := `[{"slot":1286,"numTransactions":1732,"numNonVoteTransactions":757,"numSlots":393,"samplePeriodSecs":197}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	limit := uint(1)
	out, err := client.GetRecentPerformanceSamples(context.Background(), &limit)
	require.NoError(t, err)

	require.Len(t, out, 1)
	assert.Equal(t, uint64(1286), out[0].Slot)
	assert.Equal(t, uint64(1732), out[0].NumTransactions)
	require.NotNil(t, out[0].NumNonVoteTransactions)
	assert.Equal(t, uint64(757), *out[0].NumNonVoteTransactions)
	assert.Equal(t, uint64(393), out[0].NumSlots)
	assert.Equal(t, uint16(197), out[0].SamplePeriodSecs)
}

func TestClient_GetRecentPerformanceSamples_WithoutNonVoteTx(t *testing.T) {
	// Backward compatibility: old response without numNonVoteTransactions (from Agave response.rs test)
	responseBody := `[{"slot":424,"numTransactions":2597,"numSlots":2783,"samplePeriodSecs":398}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	limit := uint(1)
	out, err := client.GetRecentPerformanceSamples(context.Background(), &limit)
	require.NoError(t, err)

	require.Len(t, out, 1)
	assert.Nil(t, out[0].NumNonVoteTransactions)
	assert.Equal(t, uint64(2597), out[0].NumTransactions)
}

func TestClient_GetSignaturesForAddress(t *testing.T) {
	responseBody := `[{"blockTime":1625231961,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Yig3yd33o2hyZV2qZBJkScDArwVmzurkxhBfKdqJeujTrdKHwrR3U8KR6LrhN5eWNTyugS5rkkYagVXCNnk7pks","slot":83994671},{"blockTime":1625231952,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3oQ7qqpJs5CtH1Xnnn8Ru5MtxkR3SZgshqzXwokuxFRArLihKdvCb9km6gbSiiUaNSHE7zVJqUVUZGfYuEaqWZPV","slot":83994656},{"blockTime":1625231913,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2UyvpGHknssUFJ77vZgUUzhRjMTTttKeMRKvJgmwaW12WLjmhTXJMF7WmVy5DBJtVFbuE25XJH247ma19JFrFb5K","slot":83994591},{"blockTime":1625225568,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2PCayD6PMA5BEkLC5SydWWWong5XPfNZMyH4LwMdRV2cCW7h28hkySmb8Y4RDzjE2YuMHwYMdxnXkvx9mbhGokFt","slot":83984016},{"blockTime":1625225568,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"MqfYQzuJmdCYmwVLSGxvwSEG9kxeB4iudWbUanNrg4DcG8nH267iamAS6dxi4ckYnCPS3H8SANsy5Mo77YbF1ya","slot":83984016},{"blockTime":1625225508,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3RW4i3vVrymSSU32BQGkhDDwBmLnmr4CwFeWBjckddWyGJMLqhjWtY4kCCqbev32cm1WkTX3rvS8Y5mqSN3mWBQe","slot":83983916},{"blockTime":1625225370,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"67kpcDECq6V6VZhwwRnn79XRNW3sS3VvM7UkFA5MmrzEY84wTU8hgSq1Q63UjSn9fprcBYZiNtWsUZepVzVsxGy","slot":83983688},{"blockTime":1625225370,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4n81J2KQjTvjPnS4rbirWJic8D6uoCzoHkjJzVFegApCgdomJ16uLBgGGydZKLsd443ht8iuGCJaDsVaz2pyPXk3","slot":83983686},{"blockTime":1625225365,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"Hi4CKo8kYjXw6sQpT6GncYgbH69TKPTm4rHgF5Y8JaLuUvgcyicoBW2CQiaVdoXUVwXVmCLtgWTE38MkbrgdX1g","slot":83983680},{"blockTime":1625225361,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2oGCdmEv8qSrNPwjLy82s8E4jqejEsWUY9DAYFd9xH4pS7ZjwW8NpTScQZk86eMh4nMHs8YSvLiGB8iKYiLpiyPm","slot":83983671},{"blockTime":1625225352,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5oufLDV3eg1BaiUgMXScfqhtU1JszLYdoJHjmKSdJkWdmkRZnVzHXsPCSXji9haAyXxAmzm3De7peBFPVZoDNi3F","slot":83983658},{"blockTime":1625225344,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3zsq9V5asSpgwQiHTjBA6PUBBGh5Tzwuij2pemzmxCEqB9TiRmgpeP7fVYMdiuTNo9RWsjDYjUyAp4ETtPSrGQSR","slot":83983644},{"blockTime":1625225340,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4RQobEstLtiXMMB7XTeFy7HxAJgnivQZyLwkAg4kgDWd4996XVQb7M6JsbGxqzfohTbWz7EHxvJ9Eet8ip1SLDVu","slot":83983638},{"blockTime":1625225340,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5AMjJzE6qPG4cJ9Vck1z2g5tCwwQjZWGreEqK6CfNM9aiFUzCqtZ6sY6r5vfoFUQ9DkJrF9unroHbUdLrEoZ3b32","slot":83983636},{"blockTime":1625225337,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3ek35DoRGcWxjQksRmVg1EEv7ZHXraHtfJyCvE84jYYt9roDSQYMjaQEkajFPkJWarJMH87wcxQMHuo4H1D6cstm","slot":83983632},{"blockTime":1625225337,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3SUKt8qG1UwMaN2x5HS6hn1jbYkz9uy9qmR9efSb1W6mxAc2kBQPWGCHRPpaNvUHMPMW9M2bnv8mzpnXw3aL5dY3","slot":83983632},{"blockTime":1625225337,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"GBDLedioeZo8JRQ3BDoQyZFCdCexbuMfH4cxD2A691f9kE9Y2BxpvSdiyudNPechNjbsTZteNstykM8titNAXVd","slot":83983632},{"blockTime":1625225334,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5YJKDebNExQ9BSq8aqwRJHDYdCcWz9dP8TsgZg3RMNXzUno6juKFhM4GRdrKDDnoQLrmyXwQ5T7RB9kbkb1hyHsv","slot":83983627},{"blockTime":1625225332,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Zx3j7CB4xGN1QRfy3P6sKLTxAYuJRhHaDs2GqK1AeyWgBVAoffKWFfNoJxBbkrwqtbpiLZNr1PneaooRC8CmUC8","slot":83983625},{"blockTime":1625225329,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"UqiJtFokGzU654Gxx3c4G9Q7hV7Kswz9UjNCT6Zcp6vcG3GUaHEm9wKfZiWSgGBEYKzHFeseEZSwaT2DQNynaP8","slot":83983619},{"blockTime":1625225328,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3TG4PT63Xti5dxu8nPyHJw7fD8Dqt6eVg73yu2wvCBX47RgPBBLQPUkrPeSzGSkhRGyUB7AnK11cXhF4xL9qJ3uJ","slot":83983618},{"blockTime":1625225326,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2rkrVjr62SHC9jv2xz6T7pqvper7nXHEixr8AArTQW2skGpU6BZhrsAhReb74g2UJuGbX6QPtqgGS8YHUnscc1dx","slot":83983615},{"blockTime":1625225325,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3mRNsyMzjWzm5EeU1jy2JyytVrqYMQnXb5iexmAnN5aba6vM6fFvBfMkj3NByLH8oh5MXneQcSfWjBbfVJ34FP9y","slot":83983612},{"blockTime":1625225325,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"kjJdUU2MVrcDM6oKN379KKeVhZwE7pGr5gzjWFVUoKAvLykzteRaz5uQtByB1KgSprCT7Z4HGbKvGLDHLAZfZ6M","slot":83983612},{"blockTime":1625225325,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"BmDRTZSsJrMrdxVVCZdkT3oezn1WArhMEUXU1KrXxN4qhbRB6sfKzJ4onoa1k3TSvUzhvnrSxjn6Bo7yp8HYyUH","slot":83983611},{"blockTime":1625225320,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3NtPCpCENzbSiohdh7zmzSsSrcC7shk6u7hd7Bc2fw5cxRu3ofnKt13pPb4KDxjUR6QBrgqJKhrB34TzEbSF9snA","slot":83983604},{"blockTime":1625225320,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2J2mz26PYFKtrsM8dATQFG2JsTRJainjh8udss3euMaeFGfkAHqa9MHG7mMA7aftWkJtNbLDryK9pxSRTsc5nLMr","slot":83983604},{"blockTime":1625225319,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"7RTpqBx3MyL6yecVEj6dkF5tF3qLSvPhbEXPgGQNMKHTxGgEhoUJeDosKWVgEK8Dk2XnKCqZCU4uWCkunFEnSHK","slot":83983602},{"blockTime":1625225121,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2ZcrngBZGjRJLEsENL9qepSTCNrcpsmP8K62yVZc41AWh7gghrjPzyN8hX5rAEycEy2CG9qeePdSsssYPy3xK45x","slot":83983272},{"blockTime":1625225121,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2J2QwMgqLsbrphHGi6W6r4xJAA92jWtD87se7tFKAGArAyXYCmvAXK7NkTfdebHc72x5wJLFGSTcULy3BWHhkSgo","slot":83983272},{"blockTime":1625225118,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4TY9qpJoVzj3CBECfSYCy6CQGcttRJaAtZKGzdDLJTvNhRBuaV9ut7gDi8nJku2KEre2AJsBzrDEvynY1T78feUM","slot":83983266},{"blockTime":1625225070,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"X4VS3cnBJ1dTQ5qZczuUhtrHRZ58c7kH1h95EUU44dGRswkyi6p42kEbQV4M9K8Syyxghw7AdznLtW7TETkNdLn","slot":83983188},{"blockTime":1625225068,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"Mu5AAPZRDSZe5bwSweE5BCfxxGLyb7vRRbJ68r1tghLtCnj4n7wrxxBALiRVJY62N9GrcoynRdum5VWcUu8jo8w","slot":83983185},{"blockTime":1625225067,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"WSBGXL1NR5A4nwq2cdXoMBasbbZ3f82YNdUfXrZke9qutziw4vcDt4YS9CiB1ZJz1XtpqyBEXrmTf9dmJFLGcW5","slot":83983181},{"blockTime":1625225064,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"243JzPc4HvQzWeHpwkUCBSLzDbxJkpVNfQxVcro52s48Wv2Dr3xMW9nmxWSqu2s4ATAqgfDsPGwHa6o77Gkpkt3X","slot":83983177},{"blockTime":1625225062,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"NpWtDMVy2HNQfHJduywSjztCdwZi9ujy73jniGK8QEwTEA4dexPQe1HwSuo9jz8Xhfwygg66fg8v7KKM724TVy9","slot":83983175},{"blockTime":1625225062,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"27zafwu4iVocp6AE6Mnb4Z87SQjvHFm3YgvG7rCLt4wgL4VwAhiwvNoPbuv6G2QXtqA4QYyJsPgkrEFNkRQxnV4J","slot":83983174},{"blockTime":1625225061,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3a987dZGH8Hac3tHEQ6DWNZMt8SwxcJpGE2dEHGjiC52UFidTmbBkdcaSMedzmpir6F6qpJeAo6zKMakCLXhYbXS","slot":83983172},{"blockTime":1625225058,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4arwtU5RQ2kocBGt6EtUYyJNXszCmsyvDozKSkB3eN4wtzbbEmbVrkDuc5xgmBSDpuh6Jri2e46Lu32Wg2SfuZWW","slot":83983166},{"blockTime":1625225058,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Lq8Z9v6oVr28x1ygHNgjw1PW2pfeendgRpXC5rqv1UWi1nbjaB8QFPXCzo8cT7HKzGSLzJfsD3iLL51QHLUQMxK","slot":83983166},{"blockTime":1625225058,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"31XoJHX5LkA89wsoqh6NEz2ZwR4ib1b9MUcJomrkw5wkocVhsLBvHsDsLiLTARhiavApAyiNqaw4FrQrqZn1fCwt","slot":83983166},{"blockTime":1625225053,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3YGoKoSfxGRZmVSRJSDEMXEqga7o2xvsxLcY1oG1XbDTcT8DxNsoCbTTwMNYeRMSmsLSar2VZn6mPpqkQdZ8svSm","slot":83983160},{"blockTime":1625225053,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"252gzJeEbbvdf1pYrymMFnwewBsTS5GxYdocD5qrFenwrVAk7dmnNv4mDBMuCFCCc4tie4hkZp8DT4Mg8fiooqMH","slot":83983160},{"blockTime":1625225046,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3yGTyh69v35LMp4Wwmrz8hc2wgtRhhfT28v2F1Q3zZCWA31aAhJCbheWNBcHxXj7qcGfqKMtdM4y8FBt6HkekJAr","slot":83983147},{"blockTime":1625225044,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2gxP4vCGF3u8NnVaSXJ1umNmw5SJP3dVV7cDk7RUhP7b8TrH2cT9JZPqLMbF7Soe8EZfMHhddGN9vcMwwPUt5UYA","slot":83983144},{"blockTime":1625225041,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3iAo5e63VCo8b5CCMPMy1dj2kS2b2wTYDRT8gDiYMtzjyEWGxikF4BnWc6bVgNKKemxtXX1WmiRKyfxTdxrF6jcj","slot":83983140},{"blockTime":1625225041,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4NZGvikoRD9FhCj7CXMBz8uLZjCX7WsGbw81irM7qRdfhk7vWUgJ7b5cctrTrt6GPm5vJ3JdECWZxdmZaeuzkPFu","slot":83983139},{"blockTime":1625225038,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"48cHbdw4KsHcXtDmjUTyKftqj7GYvBtLmGQnwUP2kXrQeCb7Gxm5GeQdQQRm16eea7GmePA26EHKLFBNeDiWhnM4","slot":83983134},{"blockTime":1625225035,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"aQ4SqDy8tX1AksDcJsr3yyRrMDzaQa5ZndGwCABVUCTAcc8sk7hSsFfZXwrr1Vr8kxmzSa2yySSQPFUL6SSMhLA","slot":83983130},{"blockTime":1625225034,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"g95qeYrT7wCNpd5eywJuZJTXnLoMWdqkGvkHEBdwXLSkzafsHcFhfCheVP5qrm56wWq2PicFnzyxWXyktxvmHP4","slot":83983128},{"blockTime":1625225031,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"428KFkQVYSNXHiJXzo38Puc5fRTq7m3xFQ3w3JQEUDMtiBYxKCXSDoeU8TjHNN8pJ5PcoDm9voT1KhPBbmWKjPAK","slot":83983123},{"blockTime":1625225031,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"26aXCPbNZQBsPZSLqgF142UPp7uVyDUgYgQF1bw2aWzamdZVTnHuERfrSQzvj3ePv7S2wZYx3fzweQKm9x68PA9s","slot":83983122},{"blockTime":1625225028,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"9oCgCt1XepgpTnm2VD4bgzqnezgYzWhMh1scKjA8UoavKZDK94xAz62DP4TybeCS2sQwig1N4kFgyEfYkVfEhnM","slot":83983118},{"blockTime":1625225025,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5SdibueP5swQU35Kk5M4vYLBbjYi1Bkk4L3S7CGHqToW9nC31KDSci85zGNt4XSTRwSe3fJtRoKkeFEtbSV768zs","slot":83983113},{"blockTime":1625225025,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"LD8BoRfUoxuuLH8BgYpY6UFci2qNUjtW3uWooiN87By5rVzr8TJxB1FFYWoRwEKZPDFf7M8EVFGTQD6PZVojWqX","slot":83983113},{"blockTime":1625224924,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5mx9RvddYkTJfNVS4i4rSLUkL2D6diJGoSgfsPRkg1DyYeRUT1JpN9WxDLvd4RJqYmtJwRbbwtZq2324UsTjqbq1","slot":83982945},{"blockTime":1625224918,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5Cpf7bEg1g4sL3WHDFyU4gq9ts7nz8vtxPpNYrkzEUKjKjKjXKA6jmMSs8MjSNR7w74AnCUf9ZACiJuZ7QhgnpPy","slot":83982935},{"blockTime":1625224917,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3UsiPVApcKJ72z3739yXQY1SsjFSoT1YNYLzPX8MtRDRMjNj2BHxx4V9VAtpzv7rwDqaw5yMqjSuNXGtDSsPXb8R","slot":83982933},{"blockTime":1625224917,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3dHSCVpNLySs4E1DMfit1g9zNJhTUFAMHYy5Rw7dJrE9pDMBmfdKxzegqaV51iJ6qiNmFnkcxRL5Br3AGpcmLo8c","slot":83982931},{"blockTime":1625224914,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5St97pkc4Tmzvfudnn1dc817Y1e5evwTkzmmXWYxHZevjRAkmEuxizkS7tzjRsESQCZ8DoLP3KXnp9h7AGjUUZAQ","slot":83982928},{"blockTime":1625224912,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5aMyctNqjodGVqPPV9ZaTC4R2ytSehFoRSNP96xeuAFw1UWGHG3qfVpRPtyzCH7Bj9ZXMaCApWJSaadE3caUsZjb","slot":83982925},{"blockTime":1625224912,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5s3779heozyv5MtWr7ZYJUcGTktfvA99ojChfTvduCyBs8scG2NQJ2vpyk8MWXBsxQxfsseLgZdQBCaQofcRdGG9","slot":83982924},{"blockTime":1625224911,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"26b4Gcew38T2DBk67MQPaVcBxnasJ9x2R8P3keVPBa56DbA1R8DJtv3TvGnYR7fwFgtYJeYNwxSikD96c9cBRJbn","slot":83982921},{"blockTime":1625224905,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2ZxprAEtUBGn31hQFFvHSv1xPu4LZEcHqEUMxCDcyzqTDwYXBU3nXYvFAKk5gVaGGwQb6Bt3zo8iKcz7E5m2N1WY","slot":83982913},{"blockTime":1625224905,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2VKECsjbH3Jo4vvkAGc4udc2YuzH9BxFH6bSgXaYCBP2WXjfSGvcjNwiC4nn2amW4eMmjadMmHSkMg1EDLkxFnG9","slot":83982913},{"blockTime":1625224905,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4p8wS6YU1xNBYdVgkfbCjsviT87JmKsVnw8S7a28TE2LTRP1eAoByrmHrPWkTpQKWyPHQnZGRzr7tpcpVcY4CT6v","slot":83982912},{"blockTime":1625224900,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2Ydg7uPkLVkoQBU3wqurbD5aBuPqUvrAFXRvvmCRLsDde4PyLNT9sy8A8WY5aBkB7f49SrELQV1BrzdE5TWYAoCT","slot":83982905},{"blockTime":1625224899,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"1LEp2rhVryBpv8UQg44GaJjGWUK6Z9dkTZaCcDxDsSazRDJovifgi84MnrdqNduD5vvfZQhMAP9hobTzcUGbzam","slot":83982901},{"blockTime":1625224896,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Zd6Vpv56N88j2otiJicJa3w5aUbs7kt1B3yh7mAUGfZzgXpNBHyELgtM92ve9sE6vZiksV3piuDgqZ9dtpseuWc","slot":83982898},{"blockTime":1625224894,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3CE3gGZh3srsTb46DsuJ3kKQRmQdCHKipJPTN1KJQcAfVSiTuvnQ2dpSvcpjzxq8PbpUwD1Z11dXm1LGz3KmT8qU","slot":83982894},{"blockTime":1625224893,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5jn5ScLLrtNp8sz8Rkv7MMQQisR2Rv5sk5YtRqKTA32tNSimmyFEaod4y87D4ZX7Jwd4vY9fwiPDSa1dQtecvJX5","slot":83982893},{"blockTime":1625224893,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Q2FfCSLDF3aFw7zDsDRsSKFuB8XcyrM5nW3gXVX2RcQKGUTPghSdDhhsWgxfKZrsmTwVpfBx9TTCRyDMnPFUD1z","slot":83982893},{"blockTime":1625224890,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3RuArwBU4M13pKhenJZLE1cD26Z9kWKC9zMJkbMc8AtWQoZFQDEVBqWRxLmvC9DWFMvcFHQ1soH7NspczJoYDkf9","slot":83982887},{"blockTime":1625224888,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5ve44WDQiyJ9ng7qk6AgcPLSmXYunqk85VCdToyc4vvLJP84PhwxNtmHj7kcrxjF2WZ8Fahxh3honuUs8sg79nSE","slot":83982884},{"blockTime":1625224888,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4QiXj7bWo45dyjSQZAUwaVbzyWWar9PkDPTT6vRPC4mGhGaqFEVZKv9XCyjg6SEWRZVavH2rzdBA28MrEYVpJPaT","slot":83982884},{"blockTime":1625224885,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2jXU95Ap5h2bya8BJyXB5wFGsTL2UMWQiRJmvdvFFGB77joxqZRDB5VbwL9q4zWEiUej3VGXou1B7JGa5zK1tkGS","slot":83982879},{"blockTime":1625224884,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"61J3FCDFRnkDY8Q9bPZniFv39h4UyyAxaUXmttpozTkxs4z2jJ3Q52dJLXLSQsEdpMBUv7Mrz4SFoM11Xht4jbaX","slot":83982877},{"blockTime":1625224884,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5kJy75rpejnt39WL2aTmenNzdgZZVkyHEQPj2cPvZcBNq9H9qt3WPdHqkyyzf2cW9iUvmp5FsvVnTzagMNbMFVKD","slot":83982876},{"blockTime":1625224881,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4phuqBTj7uEx2XNFVFTcuRb8R6CWgj2nDxpkMtkdsYZ4VfJQsGSGxava6ZJ4wfaYGLXfNkNbVbXg3z7pnmEFXBsX","slot":83982872},{"blockTime":1625224881,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5gKb3Z1uWAmyR3kmqYQsTnWfRVhHCMQkGbqKpchydKKHDVm7cemdhs4diEXu9sChmSEQiAPJQ3yLBgsxD1rjRNA6","slot":83982871},{"blockTime":1625224780,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2j3GwHVZ9WCpEwpF2jamAcrQ2nPNJ9zyEMtjrisvW42XokSMbzmsJT2jYVcnMUUkYU6mLFyC2T375xYBnUyjUsXN","slot":83982705},{"blockTime":1625224780,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"37hwAGAyj9Vg6RxWhMUjUnY3ySHDBkTjP8H87uWYs7VvKVV3TYKoNu2hsMGgvn3kYbC3VteqUQs99Sfy9sCAuQjx","slot":83982704},{"blockTime":1625224777,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4m9nR2KfjrjdHYv1qwDJgP2a1PzhhQAm7pE7FaqAX9GfhAuQkNHcdzxghvYNARZGoXxCKwr717CvU93BZBYPxekz","slot":83982700},{"blockTime":1625224777,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"44FykQbp8WSyYi1sesFBFAYMTtrH4vZRmee9j9wmARmKPUi2iDATx1qN8ML6kwWTMsfxuf5srMPeXpiEX1kTmXsZ","slot":83982700},{"blockTime":1625224777,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3nCcUa26Ptw5Tr24vxFCn5W5jxHjnnTSZ4FcoHxFbX32xFparaTsTL2JpySdCk1SpmrzCv3mqUKnn3FNTySHx1fm","slot":83982700},{"blockTime":1625224773,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2SQJdn4ucnE9rYFgWqU9RTeUVVRrRrgS8M9BoiWVNpzKoHqDjhBWV68hoAvdDssz4nb9enPSWXaT7U9yunYUSWfS","slot":83982693},{"blockTime":1625224771,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2cSPQqsrAi7wY3NnMDbBggiupBB2kR5Upns7sxPd4t7TmRV8gruNqKSYF9hKkwDTyd6fPWEo7eNgYP7YEunFLDuS","slot":83982690},{"blockTime":1625224770,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5v7CFR1xyPSxbyKrtzpQonRJ8Bcrnj4KEWqnx9yEPRHCteLDenfJNfcecr7zB6YPPXvnr2y8pdiDYxAD3ZrdcKUE","slot":83982688},{"blockTime":1625224768,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3goynxgFqwoyFyVxAxCNcBdrKASbY8XpeFCHXc6Z94cSC6dD6io1tK4rW8HuNX2Z7pPVdpQSgTvha35biqYaDVKq","slot":83982684},{"blockTime":1625224767,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"PCsRJLujZJjpEwY3hfjFE7vpxTgoEae9hkEZngZ8DAAyMt1gzqF2igMvP9UijtBJoQMHcQzsx4BJPxEGrGXfnN3","slot":83982682},{"blockTime":1625224764,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2i5vuLs9LGMkZvLrRhEBCowNG2DETZUkj6HpBdzGabAHqoHbbZ4qqXWDposfSv61PHKtRtQ8qqNp6t8YaeJq4CMT","slot":83982678},{"blockTime":1625224764,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"mGB3kr8GCwLTcTn6ZEpZNTmkcNx6zswGzh82UwBevzt5ANSRQDqqFQN1TZWMgkm2qh4B1axGQTrBmpvbq4MaZrk","slot":83982677},{"blockTime":1625224764,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2UgPUFenPv1Gv9dYQRPiHdxCMK26nyJyvBRd1TaT5SjPVaKGdk2KAKy7YEPGvCwvsHWmVGhpepFQWfCjzpA3GKrH","slot":83982676},{"blockTime":1625224741,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"NhAXiW37o3DQrP8h6HuYGLDWB84iJ39VGoJygdPYD7ut2cDxz4n6NZEnbZbkqzTK9BxfuWSD4uNVTCrtSU2pX4W","slot":83982640},{"blockTime":1625224584,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5ttknQTY7Y27iFCP2Bjxd7YfV1VuGW5BPMpKojEkrdvnqaQkXUZF9q2CEEoVeW1cR9M4wSZYzQpzjFzCQ69KNh2B","slot":83982377},{"blockTime":1625224584,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5173zGU78Bx3Axa4kV8yzcuvSTxPUd8yFWTQ7q3WqzHq27K49bHLcZz3aDgnLvffxHcB9y7umRC3itsiWbwZkMHe","slot":83982377},{"blockTime":1625224584,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4yWxmXbkyP2hF5FaZun7qGq8UD8rX8x1ujJepL8JU8qCN752voHpSvHX5Q2rwptaw39kjYVbVnYB29PCtyzogoVf","slot":83982377},{"blockTime":1625224584,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"27zaiSpXJLRM1Fm19BS3W2CNJF97DNyJQvayUJr2XAnB5YmiqAn6NSdx27mH7LsFMv7tQJCVUTpkY1RW7A9HtoBK","slot":83982376},{"blockTime":1625224578,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4H8p1k1hwyGRwVuEGbbU83sWQuX2jSDg1VHxurxTAN74LsgLvcs5eBiidwzmWjJ7vJ2LM1zko5G7C2iJBsTi6Bsz","slot":83982367},{"blockTime":1625224578,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4NLbY3P41ycuypqEop5iba4k7PtLA9o9b4NQJGVGyRrAAaZcHyfj4DoC4di1VvkCGDZETSMstznjEqi7U7ZScMe","slot":83982366},{"blockTime":1625224576,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3bwnZZHwGijsv1TYCSLcnpBDRkyMcv9gadwpaK5rMDZzPrDajFCGPE2NmDbWv8rSNFxzEFEMEFNocqg1uRM1o5Sd","slot":83982365},{"blockTime":1625224572,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"rGgC3jksLWSx8BteeiRwmFBbm4SmdjaPuEBj3ms2jd2C4kAEphfFvFPvLpxwdMrXvr6Y678VX8xLEpFXEsJzfYz","slot":83982358},{"blockTime":1625224572,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"25wvFKYfopRvm3dtpY7PMQoTszzsMSLqn8E7HDmPB4ePt6td4ZVJMBaoPM9pPs4PJahsNCN1ZthpJrHSLfzY5pdS","slot":83982357},{"blockTime":1625224566,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"55PFw37VJTwvhZRZicoEXGzB2zU9zZ5dGrkJgwrYme2RhZH9dv41A5pmpnPPrkVk7i1Qv8Ae2SQyEXXu1bDGduU8","slot":83982348},{"blockTime":1625224566,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5SWeCmzU8qsBG2nqSJgYzwHFhwjWmFtVkwvSeeibZqT7ziU4D8TT5jQ3CUrBf4AB4oJYKzv3UbvqqtEaYZC5UvWn","slot":83982347},{"blockTime":1625224564,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5kDYmxMA6MyvywDpsjNPLBj5WRYxT9RZEaqWxTumSyqBUgmVC71x44bKoSfxEFa7BTeGqKMaVPv6kknKAPwY5618","slot":83982345},{"blockTime":1625224563,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2jdBiDJr3Pw8K4iVaxKiESxxMpnEkojNDbMADX2y32ce5FFQKwFvpVtb1V4bJit5oGHfREFj1KB6dn4izGTTWKHt","slot":83982342},{"blockTime":1625224561,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3oMS9SspiZQ6sGqid7MdfhG6LNbAyD6SmKmZoySbDqnSqzKWtLcWZwxnhAx1PwWE7P7wbLdXZHvCfnnRxaFBJk6K","slot":83982340},{"blockTime":1625224561,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"sr9zgdcQ6e6xAi48pg1rUPh2o7UNox2adjx4wHMS9DTcirZV3djbt2X5LHWCLY9Ld26BvdNfS2bSUw7A1Fyq287","slot":83982339},{"blockTime":1625224557,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"qrzeGtL2WH3JaRBCwLnQsQPXVbGBoiNTdn9yWLVL3WSJWhgB6GUsy1jjEqFRFjvHS9sV7weD7nkyXnZjZj8aXqJ","slot":83982333},{"blockTime":1625224557,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2Ci9AmMTFSB5dJu5Q7LHDYgiVXHv2GDdFfp9ArzKApKpHSKCwUMEboni4pywT7cDTJDuu3SbC1KdUb3D6ikByTQf","slot":83982332},{"blockTime":1625224552,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5dhnUiRbtXtzXec5RTiJMFmEn4FCzr8fJg1fc377cqazq7pt4iAYve395ojrh5e7Stwhxnv9ortzcStonTaofndS","slot":83982324},{"blockTime":1625224552,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"E7RcXT8oiwMboPyUd7xr73WCL6ZBcnNbh6W9aEtVhs2F38zSKDcRGacVzfpMfkDZKUzWpQNiwZhrEDspnCdVWxx","slot":83982324},{"blockTime":1625224548,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2eegxiv9cNifX6o6nSmT2BpJqzNPaKaVUT75WkRTsToKVQgxhVtnrV1QgZdvhmPrhc33Z9ksbgEycprfCLNfHgnQ","slot":83982317},{"blockTime":1625224548,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5FNyZ5ZJEECYXetSBex9xDi3ehkeoPLgf4bNYaQugW8tjRXiyYyjY5VLBkV4bP5VBKn6poXMGWfEFZXULBRgo7xU","slot":83982316},{"blockTime":1625224548,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4qjjQwn7aLJtvr361oTC4aBQsVbMuC26gTS5RctHS9k1RB3vboX34k2vhDYcv5AGKyim784uR9ZkSLmN7Kb55FG4","slot":83982316},{"blockTime":1625224546,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5oQajqeJnVU4qxPnkrSMXaRsP8m9VFsY8C3eBW8mebDLdp9jpByAyGWQWemnej5VTJdbXQst2nfebF4MuyDi4XwB","slot":83982315},{"blockTime":1625224545,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"SuipYCikT1GJTMpCsSs3uuM295WYZSgJQz3vFgnRWudFKzarRdCdCZut54dkGV7xSns9qPFpNtEFVvHrGPfPnNi","slot":83982313},{"blockTime":1625224542,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"avqS2HrkVdzmuVmqWT3hdyucTB1i6HuLMo1nAYkmK8G2uAVRqb6dkE2N2UXKMjaSpP7oURPZf6MM7uDK26EgKre","slot":83982306},{"blockTime":1625224386,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"7m58EXRwyGhcRJrGXASMb9ppRBoqK6PEFWfLJa7vXYbMfSedHkPGEnnnGEcMJJzdeK87sykfzhh5LJBym5bPirN","slot":83982048},{"blockTime":1625224383,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4BFxQoLVZ2576pHVwoGraD3oVqYkn9oyzEXNi7UU4tV4hxHMkfmgoSMdUbQrAeS5KXhBah2DHbGfVYvaEJVUh4eE","slot":83982043},{"blockTime":1625224339,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2w9BKPoiSccMPTJXYQFEii47VMLmQ9j3WtZcFZxuLPmPjYfYoLTKt1fp1f7hxU8gGkWCRN79r6Rm7gLBvwpeEXAA","slot":83981970},{"blockTime":1625224329,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"57fAnMvyEVmRe18vDnLmTAZckf2TcByxUqzoLZc4qKSdE8vGbnELGeBz1kvMj3hpkW2819n3TGAEAQ9oNLdAmV4w","slot":83981953},{"blockTime":1625224329,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4FochZtZSfrtER13yZko6H313vU4WHdmbm3ziWd9PuoXmBQ7VQwVPkr3Zn1EjAbLBdgvZBSh1CGRLTy35N9KnNje","slot":83981953},{"blockTime":1625224324,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4yzFgtwiqxD7WaDwTCDkG7G6vkMuJXvFGLNiH4Hzo93M2yA5o6c6753VsJ2wUEsTYpab586RBienKmkfpyUdfWwZ","slot":83981945},{"blockTime":1625224317,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"ievMuB3qdEv6UaHctpRsPQmDq2rWefpw5MGPET6Ne9YdpyhMZUA8fFkCiAbCEBKXKg7FQYziZ3JKhTsefQierG8","slot":83981932},{"blockTime":1625224315,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4mbt2mns4WCoVY9ga26CU8pKPRovTX43xLPyiQBB2TmvQnoCLRzWX3JJrQYsmNj9vbkU9JFB3efpKnYynpZQRPRQ","slot":83981929},{"blockTime":1625224314,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5nVtg4H9vYsn1gAN29nGvAxrsYnumGAhUHP2jDe73FL6hRCXfLsWNFHGTW66W6SCzfaJZAecdHMQ5oi2DPuZHVMD","slot":83981928},{"blockTime":1625224314,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5x7difuwKJKc8pAyBZVdBrPkNDcyhU1g5Ci6kszJh3pZsxJKbpkGckUs4xSjz3cpe4mWyJg7k8kgF5Wu13rYcTuJ","slot":83981927},{"blockTime":1625224311,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4m5kZ4aYtBMgMp5r9NwWb2phopKo83UVJuefwfNe4jPHkHjWpydgGC6kDrp9mcUodvxz7rawDZJXoUkGii9VcHkN","slot":83981921},{"blockTime":1625224311,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2NCLYFBjhm6bumeVQ1CkecdJqMrGocv8d3XnY9A2ryk2Jz9DUod1MQNFsbyRXBTLz9L82JBEXm2p1ZpGNdmCUBD8","slot":83981921},{"blockTime":1625224309,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3FdD5EUrjkFbPB3YvwpovF7t7NSBXu33b9jajUuSoE8U4HdbX2wnDKroi5CwioxPSjFHxpgedaZRNcrmt3WjNbQ5","slot":83981919},{"blockTime":1625224306,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2WjPxanmrC3eiUcoxcqefPjUBJ2wa1pwFqSmhpR6e1Humv6TJkEqBnhgtNLaiTKqzDFYG5hvL6wUdDfwXw2XeoDN","slot":83981915},{"blockTime":1625224305,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"28yezno52Y1CvmkBsQPP8m6hpR5sK7xCeLh4V4LtrEGmPmQgMWu8vaoEwHFVmXNUrAmSr1UXG5Qj9novjJTwC2Un","slot":83981911},{"blockTime":1625224303,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"59FytoYXkcm5xietGztuMpE13og2fHoAup3SMFQNFp6MeHByM4PNdYR1Ai2X9S32mzoRX1pu4U1bvEXM12N5jcq6","slot":83981909},{"blockTime":1625224300,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"pbgwskcx6mKoNzyCfZdZNoiEVNheBpNfz1dc4iA2KWBaQorRHBXmKkxVcVmcu9iiayJtmYosf51EjKPBAjm5nSk","slot":83981905},{"blockTime":1625224300,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5qdpxPagUM5AVtz2MT2NWvYARjwgXeLMDyQ1WDgeM9c2hCUJR5GRg3mkiggHaNeNy9wYQRqtcuXXEe58XkAECQke","slot":83981904},{"blockTime":1625224300,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"ktvrVrdVqhwjw9F6wu1HBovScdYKNwbpdoPGoQwr8JkjpBkL4RnEV86KJh2qn7yG9YXJFu6t5QMNoEmg8Xj75JA","slot":83981904},{"blockTime":1625224297,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"DA2auvWg2H6E4aJmFrT85TgFCnASMFoLieKeCGAoxwgkJtKkYohCfovGGdPcTMQdaAgZSNqW4ayffRSaVD47irC","slot":83981899},{"blockTime":1625224080,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"kSoykkKXra1EEHNVCzmCbwj8zjsXza9uJ8d15bXcGfuqJJzUatKQEcXyjMLipnTi6rKbtrBxaAXqRdtnbnZ7xdS","slot":83981538},{"blockTime":1625224080,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4DmZezH178ms7874UEUYNFpSSu6JD4t1z1ghQFdLizpNgmMfok6N9G71pe84MVU71scYJWYTeijjnKBN3W2bT8v1","slot":83981537},{"blockTime":1625224072,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2JEhXJCXYHhNDE3ANb3y5Xm43Kw7ndN6bUuthvMYLtio7J6JxsPUzxPXV1Lc3s6N61BCasEXcDPtg2WpS4waTcXq","slot":83981525},{"blockTime":1625224008,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4PSqGiSmzTSN5EkfaQCe5iTkGgFqfFduBAYVFyDByYei6mFnNbVdvs4eEFf5QTDbZjkow9HwJdZFLqjLzVx4KVFK","slot":83981418},{"blockTime":1625224008,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"62MtVqp4wPv4ZsFwAqALtuejHqejwSdByuz7ESqupUuKZoJZJrGpbdwzFNfhcoQs2Jcfvnv1cBeCgDqob9W7oD3s","slot":83981416},{"blockTime":1625224008,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4pgp1BrWT9DKEiM6NG9zLNN9zW3b4tmghrbHyg7V4YmdLCpBZJiXzVPd35GQzXscE8ba2HfgsXV9wyW7iigRrYnt","slot":83981416},{"blockTime":1625224005,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2trTxgB55G4Em48SJHwLW6AxP1Q7i2fsukSXKr5ayR1J6ThHE1Fd3Tvn3rzJjz3thfChdW3Rfoydkmh8Vvbs6aqQ","slot":83981413},{"blockTime":1625224000,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5T4evEsHQed19cBqDrvP5k9XK2gFEFpchCVJeRBvnzeY6PkpxFtiJc4yovhLyTPDLgF9gM31BKDmA45A8gYogRTj","slot":83981404},{"blockTime":1625223999,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3sSixgbRhA4jinSnCwtud75cYWVfSxZWzcPW8heLkGASDrUR4RNDo4QZW3Ei49DBNXVU3UF4XH5SxCoxJwXztTvU","slot":83981403},{"blockTime":1625223999,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"43kA8oMU7TidLPLLGCCk53a8gpSUbQFadG7Hgn13cUCAV74ZJFekrc169E2u95LTj7RmKxiyu3DobuYRBXej2pm7","slot":83981401},{"blockTime":1625223997,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5RjB9LmFTAwbhpdMXprKkHzMVE7x8dwYYSg9NWgX9aVcFLB9bNMYw9Dqv6GeL5v3WnBZdUiiDucuXJoLGgSPnZ26","slot":83981400},{"blockTime":1625223993,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"25yeohM9xxKWYi5X9TqSBmZKKvoVuJ6nxERjJje1CcbCi1ig6woeYdax6MibFZWQQmZ15VioJCexDKgoACZ75Lch","slot":83981393},{"blockTime":1625223993,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4hwEf8miSQ6MBEaT7RN2c4m2h7GXw3sxZSqqAMJFvT3Ynz7PEj1ETFAn71KDYnF3tYmnWbtPxbu627mKKbQ3cbHp","slot":83981391},{"blockTime":1625223990,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5QZB7ck6aSF91NHgu5MQgrDzrSvv3WzBjPcKUSqgX1bc97DPD43o1gLpaTx96bTqmuCCTtJbpXEm2sLxBfk6FAys","slot":83981388},{"blockTime":1625223990,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3auL6eK3KJwtt32yZGsVoH7Vp8spb4ZfuU3u7NNjU73eYgESeAhAwYdtwzBRzwfqdswDWAPSCxYvFTcJmJVeA7WY","slot":83981388},{"blockTime":1625223990,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"rHKYyLacPgeguC6KaZgqMTFYtCBNHnFUg1n4DFv5fbLP84sY8jphGioeC2N2HyBTUExgwMh6di77xDF4cfrVGxA","slot":83981388},{"blockTime":1625223987,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5fMAoXFFUjB35Bm7eFYY7XuHKKopx11Xg18Rrzx12VuBwwyinEFF9PViGNCwkviMSB7E2MwrJLVjD4w27ASyfrbT","slot":83981381},{"blockTime":1625223984,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"nkDRX4f7qYER97bKRgjdDXvarovHM4JXB2B132xUHrP8K8Xvg9mUe6UNmuo68D6baafpaPZhd6hAGdua5C74sbM","slot":83981378},{"blockTime":1625223981,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"HV6zrsM9xB78j6cGBD5YRNQP3zmaRASeoLp7CxkUA8E4tL5HUqDP5veEazYwEXBBs2NQDTRXzrzkfoceUQ4d3L6","slot":83981373},{"blockTime":1625223981,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5dAHvXd2uJYeLvQxeMP8Z114FoBdUyTcrRx8zvKBjpi7DUtqNFFUzDhMSd1Bwq6Ein7mMBjLNK1p7GBQY3QgG6Pv","slot":83981372},{"blockTime":1625223981,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4sTf4LhxpTkgy8TKXosiej9iH1hjXmwEERQwAZVAhqdzHEEnh1BXDfqFHLUYFsDyndB23asXKTBhYXmSF43BaAm5","slot":83981372},{"blockTime":1625223981,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3pN9g5krgA9BfQFxLjFLPvbXLmerj2DFdgcGmJdbkhNt5TugCApjVFjBr8DKLZxV1kdPiYEq9Ng9jzWeKrgBuScr","slot":83981372},{"blockTime":1625223978,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3pceFJkCs9B1uLi5Q3wRRnHqAXLomQTfv3cJvPN2Fd5th22VRgDimpWt2PVAnrGJskhQsuzGu9oNt6uWqnCJoqTB","slot":83981367},{"blockTime":1625223975,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"yioySKmC3Ry1FaWgvbjX462w7PnHUcggwZtKGtxE4Ddb3bpkMcDqPMPcvHuLzU3qp9KQpgmCQpdU5qDk8bQot8j","slot":83981363},{"blockTime":1625223909,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3YbdrRRSdL3Bmmi9N2EPWYsHzmU6ardNAqD2GMJpkuKaSSq5ALPjtGW18DrhpP3Rq4KMAbkjiKF1BpyAcHteVV16","slot":83981252},{"blockTime":1625223906,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"CoRBfVkiUAu3trHjLzxhHpC9E7b22WAiCxrEdqGfZTV1ih1rQWQp9bAeCwFKkzjRyvpj6YtzB24Ck8TiX7JWG7z","slot":83981247},{"blockTime":1625223904,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3VpLHDuLRYF2ajESptDpXbYEbFGW6UBJNV3shutmrpgTp339u7AfeMwBD5huugHxTLNepiyPrEjhwzoJV5ZGgW9b","slot":83981244},{"blockTime":1625223904,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3BhmriCcPjAtDBYLPuhXKQmcsg8ygVVw1DzbvzYGyWjcGYdazvB6PQr8rFkhM8NDrie7TBj2NCofXVZH7mRT5XxT","slot":83981244},{"blockTime":1625223904,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2HdGU7NYKav95oCcgbye4L3AgoY9Cdo5RpfmJbBkXJvQf1PpLeFFqEmezRhXkUoBx2JpSPfz3NBLaEVqe9pgCzgZ","slot":83981244},{"blockTime":1625223904,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"R155XszKti2tSL95Dcw6nt3ZZkoVZkh9BoAZrw46MVXW9YiGycjMFN9NAQex9npKcR3emR5XCXxc8T3TZUryj32","slot":83981244},{"blockTime":1625223900,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2sa4wppQ9f8M3kBwLGExhHn25jzuFumjbKRNC4QXnyo7hwYcdicLbooVrzE8MkdNPA5M8iFdLgjM93UrCw1Xa8gq","slot":83981238},{"blockTime":1625223898,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2SaXSeeQMjaT568ioj6kiKhHhYgcxiqQMHXz8Rd6qF7oTYErrPqxAMRRCivXgQT9X6R1jBayxbGzAY3ngB4zgL97","slot":83981235},{"blockTime":1625223895,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2w2uoSUG6ufaLNPqGUnRPNqq2CNuzrkqsZmsCyS2J9cpYdRGjxX7hHE4EU2bt9M9359oRN2X99h3QyNApaj6D3MD","slot":83981229},{"blockTime":1625223894,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4sALKdUEGU8HVbLPgSnf2qj3SVPTUazjSGMyCJuLBYVWejy5ywLEpZRekq8dU5wwJp1QVYhpw494n7h4u7atmVT3","slot":83981228},{"blockTime":1625223894,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4kh8SoZBX33vXuhLB4h4VgCWfbG33xQDNFH5s6EWt4tcJY7tfXJKFw1PJZrbzvwv5nPdD7F6cPDejXnKHWyCT7hd","slot":83981228},{"blockTime":1625223894,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2YdL5uYW5L8aW2mANSkuovKX1Eo8Q2oZ8npLWmjfWwYjBPVYLdDZx3KNJrfWqtJV19LQsUZHdKdUYiQoUDxDTB56","slot":83981228},{"blockTime":1625223894,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"HsQsRZG29zwLSWFt3hjfurnK69AQHZUiEhumdHKzyAE3jkN28fPcRb4ezUNoJEgsGM7fxtbzo4B747vWpxz9kC8","slot":83981228},{"blockTime":1625223891,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"HsGMj3KmTWiyAmTyJW3nhBMCCn5BqMfvEdjdBFMUBgc2RxBUTx7oYPrsJ6K8SXKfM2KtuQaxhbZJPokQeQxwYfE","slot":83981223},{"blockTime":1625223891,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"M2d1dWLts72xJ4WMf3B65tCXitbsBMAj2RLrFXMyDy8LcSgjUgi5gtbpfGa5uchmWakqaqPuoLW2ZmxJhZ8CUjL","slot":83981221},{"blockTime":1625223889,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2zpVFwnGJNmbmyVFQ7FB4FywUM7ozg7HzszS4Acts6f4mZhpJrcRPhGYJoSK2uvvJug3Ug9mtWxHpuwXjUjuW67q","slot":83981220},{"blockTime":1625223888,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Jf5EfQa1bgqYzFcLwDpza2Bneo3p6yUMQMjQSbcLHYhkJCvYTk5zjhUv7tCT2csyLH7JT7S7B5j5UECJcGFmJQH","slot":83981218},{"blockTime":1625223886,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"SPBSJrLVixQ2LgP51dWL196gAj2FeBfCCh3m8ThTbV5haWZCizVVARkXEmYLDuSHtAWeGBU6RZzSuofgXdyTHSK","slot":83981214},{"blockTime":1625223841,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4iYzpTPeEoWgLDCCc4EL5qiZRmZ5tDCTxZQ8VC5nmqt5Yyfe1BKrM5snKa78PaZAC9qNA7cz5zXseR7rn2S2g6Ht","slot":83981140},{"blockTime":1625223841,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4EMbhk1NYySHRygXt2FLeXf1n4FdSAyWtMbHc1wZASz3nWCyuyzoJjQ4R2sMrWJEGjdt3MCuTmmXdacgfYNTqJMb","slot":83981140},{"blockTime":1625223841,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"v4eA1dT6r9WGLdCeSA5JhzBZ2xsk7hu8xPVgYWXvPkjLUmsywEr5Tc1r1NfjWzcWD8B3sHrEL9d7okakHHZGJbU","slot":83981140},{"blockTime":1625223838,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3zzjXyfk7bowKos3Ex63Phw4bYofQH8eycsVxhficgbrYNvm9xL2UrFbgCtXatPbhSfx8mBSjbWFASDXU8VkoeFD","slot":83981134},{"blockTime":1625223837,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5VX5TBC7S4VStmWAotNigptQXTskNhHAUpDKXFvqvLNgZ6jpB7uVNGBw1XKXFUESSTTmCDFM5HYRo6YHqcm6cmDV","slot":83981132},{"blockTime":1625223837,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4zdgfgtqSmmg6UpiicTSYHoDLvDbRfsnVPmVFgtA1WZLmoWCSrHS8PdistmDNhdGHP2BG5ZPhZN1FFGw2YJvBKWo","slot":83981132},{"blockTime":1625223835,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3EpUazUP8G81nwZARXTPuvXvsH1nrDwDTtYa18ysbwNe8o9uE9jUChjUNzkafwiyEvFowzgmN8U3Vvt4dQE39QxH","slot":83981130},{"blockTime":1625223832,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3rJWs7Av6GVS683e5pYvTfUkmkkmccWQsuYaRnKQWdJhi8o9zptt2NwPzc3wgd9c9bG6whhByAbAinSYanpmzHqc","slot":83981125},{"blockTime":1625223832,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"ajZG3Sqd6zpzMkWQBwv25NJSVZNm4bbkis9bope38jt4Domnd586yL1ULAtkiAzcA8gaT9u31SnQ5cGRFhjzCsW","slot":83981125},{"blockTime":1625223831,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3nxtxPo6J1zGkkfDz1h63RBzP962BFmru6t3tWRtFgbBQ8zfSmdpPKgixHKP3nbCUb9sm2Rvy4GxZs6s39vD98XX","slot":83981123},{"blockTime":1625223829,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5mMbYyFTnmkaCadBAAb2qMwtnfAS1Ta2at1yPe5HsByuHj1pHrouCjfeyYdtQ3UgdqdkUpu1hLJkBbog5KPL6b12","slot":83981120},{"blockTime":1625223829,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2EMXLaPUqNbwoDbBZsymeDuQt2nLJfSHwcG72qXP8hyNdrVMu3JRYpscuuM2k1biRVjxvHXfh5oCv4m9EbxanRvk","slot":83981120},{"blockTime":1625223828,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"22HwrrAua7McsBb44Seqznvw8EaentMig6qcrx64CbwgUUTvxD2av4eQhCNYirwgQ7tQfgaAJLebZGecHQjwbpSL","slot":83981117},{"blockTime":1625223828,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"21taaFWuuYxPX1me9byaiFQZ4Yh2fR2HVvSbhfxtUEN3GCF9Q8Wbtm6KujE74ZgNwVzDj7gBZCfXkcERSYKNz9hY","slot":83981116},{"blockTime":1625223826,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2H7RDJeZmP7ac54p6GZ2acoup5p8VdZ8mbjGvArepSai64AM37RqQQXjPBZYe5amZFsCg4F2536CG5TLceKTcQtc","slot":83981114},{"blockTime":1625223825,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5mYQDVbS7wUX2cWBvRyNoQN1MpCEoxbu84ZXu4Z7L7Pac27weBYLjtAxC2yEkgTJxN1bKEigkzUgvQiYthfo2zk1","slot":83981112},{"blockTime":1625223825,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"RTWQbc8a9Crw656xWz1ZpjTZKzWGnGE2TQMguePv5b93CGKRb6JKqPS94W7Qa9w3LWXgNrq1LopCVpXGLqvPyEx","slot":83981112},{"blockTime":1625223822,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3rP87jupiXp6NLemY1Wv1CSkNDpidoEhMDEKyyvcQ6HR4NdLYZdY5b5AohpjoN4bSASp9XPKuRJ9rqEXP3t1hziV","slot":83981107},{"blockTime":1625223820,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"46wvDnL8k7CntNF8gWzpqS7tBuseHJev6gfK25PbcVQA2MUDbv4PNENhJK6uVojo7DmBwsBHeRa7HhqkUnfJPRfH","slot":83981105},{"blockTime":1625223819,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"39myVXcd5QfgkCSuPu5LrMWtrfZdXF89xneWyLucsb2kVbL28uLaiKHsRXYZnXkWie4GKL9HXovCoFQDnyDPuAw6","slot":83981103},{"blockTime":1625223571,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5jqHkrBgQs2BNyqaw1dPxmFLvkSXHGeK5PX6oDrSYR8ULHL4N1PEN47nq1CX8etU85kRjXma8rgkCLy1Aj7i3Kbi","slot":83980690},{"blockTime":1625223571,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5bVxw2m4G9AQhpRwFqWDThscTK8CyQXH1pwio4MLr9azXQTtBmCJ1VvXVkZ4mAS8FZmQgKWwj6vtAPZP7j5TAVgt","slot":83980689},{"blockTime":1625223567,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3DVFQeUB4Q4JjQLtvcq31Cp7q1G2uoHFiGeVbqMANoLo8hcrzp7TWsxUNcKrWYWj2eor32EdFZdGjHQF36RqvtBN","slot":83980683},{"blockTime":1625223567,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5cAXRiAfznbRMhYyDGCBn12AseFcPhbSukYAH1XKTKVGgQ6eS3Qnm7JRFzDZpmBojY3WZxjKRjrfQAgoMXUDHbCY","slot":83980681},{"blockTime":1625223565,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"6FmrADBYLU4sASWUKFK5AkmpJwSumvmrXwGJp5ybTLUwzffy8128V5wbRvUiz6ionuj6TWvPBGiEyRxDrxLpXCf","slot":83980679},{"blockTime":1625223564,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"pyR4vqwj5qSKuRj9Q47s5V5QL7FZigPA9WjsMJYZq6YFVd8Hw3sq2swLrkMpNEUz1VVLHjHt9S7i7xqpCWFKfCg","slot":83980677},{"blockTime":1625223562,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5Rqu1Z9MpNJNhdhyZAdLRiL3inQ7rKinbLg5RLoByEs4fDSzt245vD4c55PGSRkjyWni56s4e4fiE49637zTvF8","slot":83980674},{"blockTime":1625223561,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"42FE8K95qkEDEfWUGBiLPBszKtpz5vxFM6goW58fMKWuqKo29idMe964pivdmppch4L1UsqTqDsnfegGm5Kjv84u","slot":83980673},{"blockTime":1625223561,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2kfWuVZNbkK3X3d2W6wbthnSFZvzCYiKK1a5HmuTGhiJn6KezuGz8n9h1Xtu3HX6aizEqKbZbQGeim7vZp2YQmA","slot":83980672},{"blockTime":1625223558,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"26vQXCbquonBWJ5bhcxsnGHaDx81XvksfmS4kse5bX5UhpyYQNhHbY7q9EVq4uP4D8JVXWSnYCt8JNJRkTkev5v1","slot":83980666},{"blockTime":1625223556,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3TXDJ8dDSeW3YcvcG5CxHGXTXy1rYnfm9dVYALGbHxwCtqb8cSJrjXwVJ1jR4Ci6R6KyEFWjjiZmfdwEfv8ZQBws","slot":83980665},{"blockTime":1625223556,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"FGcUFRVG6uJpToSnNApoPpiP51VS8xjMuDdDBRrtKeQZwX8ZASjNWiFmkWoXppz85SH7CLnYhX6KS5zGS5ekdfF","slot":83980665},{"blockTime":1625223556,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"39zwAQsy4JAhnkeDoC5YtA6yf6BVb6YLYz2dsMPj4Sj1ErBsuLsGmBTn6nsLEAVZFHa8kJkNpUwUwQBmeuD5CrYk","slot":83980664},{"blockTime":1625223555,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"26TcMCKY1mLkoLQcqYZYM6nfcnHVu4u3e6Dj1GAb8bNRevWZ1A8gJqcLgm8G8Z8v8pzeqsGaXZMqsWYhwXXGRsvN","slot":83980662},{"blockTime":1625223555,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3brcoQUDgwxxP67AV5rFrmrEtfzWcHaJdrajJhnKFDLBRVcuXw8mr6W7y1ravJLruZFj14z1ySfxZHyz3pq3kLMv","slot":83980661},{"blockTime":1625223471,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5CXt4cDacfKaJBZE9jimh8SMLwgv3n6DGio6PmPuKurAbRpK491kx7LuK44aPHW8CfX62QRJiTY7cZ6N6wkcjPUk","slot":83980522},{"blockTime":1625223471,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4u9fV3DPNVAt8RNDcLdB9sLWg6SmR3iHCnTyfwFbj7Gr1DwjAbvpaj8xLGHR7u3NJgtbg21ECT9tYVJtKyg51JT","slot":83980522},{"blockTime":1625223469,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4dtbgvuJYDeCWFYmTfF3ffJHa4kLULKbh9AAQ4cqKDjL2m2zdCxh2thkszHiyppdaXvfJDRpoHXCA9TYbU8JyQyj","slot":83980520},{"blockTime":1625223469,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3tSnhFwtzELkegeStQ7piiHKbLbcyFFntPd1nTbjndoaczP52bXLKMNYQHB7RbEdJVJ7mbHHtWxJaTeoojB6aQjN","slot":83980520},{"blockTime":1625223469,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2Fz9eemrfZoa2NHUFtcATmL7snQ37ckMEAaTUjCiJGERcLuY28PW9xiznbaGkEJzYWMVRXxP9dXkUkUXe921aspN","slot":83980520},{"blockTime":1625223466,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2tH4NDyLSmAGg9uQRnNbrtLTbMQK6srrVrb2Tk6eZqv1ZrCKRU2AaNszWaudNTDidPCZ2WBzpMxVgoKK3JUzvyAb","slot":83980514},{"blockTime":1625223465,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5QYhxuuj3CQCbYiVPw5KASQicCWUUJsigpzXBhksJ8PFJAzoyKme2dgey7JgkvHfAXYpD7hxpNP9SmhZWUXA1aj","slot":83980513},{"blockTime":1625223465,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"23kD7i4LsSPoX34jWdCQJafYYJXdzUs4gAJqziQKuwsrYWxBcK7oypMUkeBdcmopAkhkstRj7c3cRAsmWfvJPWBF","slot":83980512},{"blockTime":1625223465,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"dChJmip17XUdupNrNG4Cx4v9SpJqrXGyoxxJbXsLpnoq1fXGtds3m5DNBzKpTuSPnr5s8oUxF75vDYz6GqEkfbp","slot":83980512},{"blockTime":1625223463,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3C5tcg6wqA11JJFWnF92GBj6SRK58iXDQMxib3VCyD4Em87pif3GxKaxrvFyXmKybhg3CWn8pn1XvDJuYeRKkt6W","slot":83980510},{"blockTime":1625223462,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3A34bdLchxDjuSCSmWqvvtCTdumwdf41yU7BpeXKcEPqC6TVqPQCjwyu1hJLZbtgTP8P1efDpwKY63HWwwFFVGVT","slot":83980506},{"blockTime":1625223459,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2h68HfVwAKsSEGzE6ddvXympMnrba6mhz8AZHCDw7w3PsH1BzhDXoPut88moPfR4WRMZNmBFNuGPVntVAeveLW4X","slot":83980503},{"blockTime":1625223457,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"vPrLEXcW51QHqj7XaXdi83JgEPBsvgr6gobbTdkTYBiwotR6kkikniocTAvNmv72zrf4jDaNHhCAdD672qd7pNs","slot":83980499},{"blockTime":1625223456,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4TmfnmUG6toVPM9kSm5aepcaocR1nRtGFPSxT1BboF1QytZHq8HVcz3oKQWKf46aHEqsMKaGTQn1Yo8cXZSUc9AL","slot":83980497},{"blockTime":1625223456,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"323ucXJA5qbsmc83KSpDMhmYCjn5Mzjt5vCdKLPQ42PGwy7iRntJeYy9or4U6xoGzuPR5nL8FXdUTCm7pZk7SsUz","slot":83980497},{"blockTime":1625218707,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4iyycoYxC6adBn3TUVYjaXi69gKrvek8jE8X6mqtR2iuptyoNuqmeEHk2wBHciBBQiJXtLTDhh1p2KfG2rvApfby","slot":83972581},{"blockTime":1625218707,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3ZdjbjJMVXG9Y39gr7Ly4GGB1z3rEaYEgprxPmx375TMYCjqaTHXB5gvffuQtcYY9zSbdyJ85jVNvyHexJwMvycx","slot":83972581},{"blockTime":1625218699,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"45burhLt9ocU71T6N5EPigY85TL6xDVjCugQrcZGGYWgs1EVU5HuC7MTvSg3cYfH6z3ruVq91nHsbLCotj7UPaxN","slot":83972569},{"blockTime":1625218699,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"vWNd8teWPJWuVUkYXKrfsRDEBTwsLjQmTxkVZbV58qXga3DJdLGX8ex9EQShcqWmdTVtGFK3HBXZFfxt9jT5F5j","slot":83972569},{"blockTime":1625218698,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3YmWTKdb8A2uczrhc4sdQRSsnpzodpwEKGevQA5b31NgZRPqrprriXwJVAiD64FryMY21CTyoD6ZJ5YiDAoyP6wj","slot":83972568},{"blockTime":1625218695,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"48zRwzKVocYRxJXFXzw3dvsQmVky87gge3DSuGYccSG3AKjhDwrnszPWg3D2KFiyyuaiQbcdYZkkRccWYN9X7cj9","slot":83972561},{"blockTime":1625218693,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5gqy32q1MA4J8m7d9cG3hnXMJADmgBZbWd9tbtmxdkvT7QXzNK6JEEXxVLfpxM9BncLBBh7zEDr2XTj7XAoeR3jd","slot":83972560},{"blockTime":1625218693,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3qVKzGSWmrzTy5LpGXFJYRH9rCEEJftLK9uhNiV8NLSUBbFNLAKLTcDdCTXYhbHCBvFN7yrB1kJx5fxHt78hAX3H","slot":83972560},{"blockTime":1625218693,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2wMKb1PZYti4CJbGr946VSXsxU48ptSqxDaLQAiRCzoFJbtGhLXc1W7Gr3ysbi7SBzoLEyYWFAdvLQCfwztWRkFR","slot":83972560},{"blockTime":1625218690,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3ps8CcyutqgLjtyXnAoCTjPWyCohqNT1se8z6QPFcT79YxdbzFb7vP7wZoaQUvKCytLDgNJbDi7hUFaCQuSae15Z","slot":83972555},{"blockTime":1625218689,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3GBhXyLajxTQXam9Yiww7hTzVbYecv8BT8RReXSkbAwVF4ZiatmrqQwXtRAgYiU4RYDo2jUhB1WKdwqzkXphWS1Z","slot":83972553},{"blockTime":1625218686,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5y8KbVmNrGyVMtQVxQoXbPYovnmnyimX9jPKctei8Lvmk1CTzAWm1SRetr4xEpnJ4WrskKR9bjqfbTTG5qs9WBng","slot":83972548},{"blockTime":1625218686,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5m972ZujDLWyvaMeDXiQtFiPjvTv7c2PYygRdF55fasqxFNizyE5juQpvodgaMN4z5ymxfFaxjE9m71iYebsvsoo","slot":83972548},{"blockTime":1625218686,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2azhXrkkQfQG1RV4TJTBkjPSNKTw2uo5Tsk2Lj5n9GXT8oUeyR7nNCtPCrse8R43hTFitLSRGpZA7VepQwSe115x","slot":83972548},{"blockTime":1625218686,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5CPefL5q1RT8Q7S3MfJbdBmJeWsaXxaj6wdkV5caMxuqvJUtmVDQnJh1mfXmUhjdeAN7jK92dZMwECdz2K1h7kGm","slot":83972546},{"blockTime":1625218638,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"48w1K6xRVu3wyMFz3FeicezEAu2FYfqnQz6Ce8u2bMv7niyiDGsLzb1qeTJm797ke2Ww5X1RhufSQ8iyYjT3VYJM","slot":83972468},{"blockTime":1625218638,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2HjrxyhBGca5rECbncRaxiVaoVjo8hAQMg7JeQev4GNXhA3EE3FfsneUdLPeYPBzuX2nSR8zsTwwz4nzVY2ccU6E","slot":83972467},{"blockTime":1625218635,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4mUybRz93e5Da7tgxtjPejTjPpfzaPT5vHAy9BvRvaPjZaEU45Af57U1ZU5Aergn1tMrRmY9QZbD7PYa7JT7Hgcw","slot":83972463},{"blockTime":1625218635,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3YsqsKoQXsr4KeEB2YjUVQjuut9oQsy7uRfF7pgAsBCa5RSy3V3DmYkUybKELbSbwjDrVNKmzypw2Vhe3mrfo65u","slot":83972461},{"blockTime":1625218633,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5msYgx34hW9P6PH6PG26t4JErg3vVtK4ui2kqWPJpatBGWAYdqrJz2H3HYVwW2t9yA6eE3iv1BhCXLVnKAUGT5LG","slot":83972459},{"blockTime":1625218524,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4ncyTsuPu2LMJ88Hv3UrtdtAFaXBvMvzBUhr27LqHTTX7vSbe8zYPLEYfBdfBYHAwVbgfS16zBwVCNuQbV5LisT2","slot":83972277},{"blockTime":1625218522,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4fwmdwuSm4AGYgyqtWpePjBDBZGAzZ53XSVsKdvDq67ntq4GAH78HRDS5ggLABUCoM33HhTkzHrsd4vW8CEhbiK6","slot":83972275},{"blockTime":1625218522,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"Y4pn7QnjJ8MMqftJ6a1XkkDdDCfAAmbyDMFD6fruYW2ij3GtRNwa3tuq1vSnXN9FWZBJPZmAkUuvkEKHPqA6N6Z","slot":83972275},{"blockTime":1625218521,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"7mKkvCJjmdTnPFyEZynLRDjGG4PaHxYSxtDKteVpCkjpo1TJZKxrrc9ogiUahTok6CvoTYzoy8XdeJbuEMapxYR","slot":83972272},{"blockTime":1625218518,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3aTi6gQiGwjYFfq5fGUDKLZT873qskeRgY9DQZQQD9x4tsaKhUjUA48cSb5jB8yw3tTEYDCAsQrsQjfywG9iwBDb","slot":83972266},{"blockTime":1625218516,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"54stjXbCLbartJfbkfmcnJowgoT2ztTmpNu3FPZ1HLn4LPgYRmmvCooZL2SdHzbZcbSokkoT1bDUupWRhvJQEykP","slot":83972265},{"blockTime":1625218516,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4HusDeRxPggsoPL56gEyzonWSjeRDpB385R5zfWENo1S9Nt2DiNdjFuUrgytjosFxxK7RamHR851DtY3EqkontaX","slot":83972264},{"blockTime":1625218516,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"27KQHSytQde2ELoHVo2DBKXp9EasUiiLRznS7xAiscwLNzV39BwXvmhn9Eb1y8Gqp1gjxYS17pkzmMaMvu2N66KD","slot":83972264},{"blockTime":1625218515,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5V2fXqzpiiwSeFsU2vGjNkrqveXoLDEDjK33Dm29Sxnrbzbr7VvGoiY5HkuYvUMamoKggQG6J3zzu84GPbug4WA7","slot":83972262},{"blockTime":1625218515,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5DGai6EB9wwnf43WJzYCwjs6QReKGFwcmCWHVYhZH3PXwvR6iZHB2rDxXW8A1YhPXKtJcwNXJtp36boLWiahCRvM","slot":83972261},{"blockTime":1625218513,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5PnHDQmUdWwSGAPWHLRAwJowoxZ1WxwSiJw5hmcLNm5i5AC3YFncicGDZWc8ji5hubQDT88cqaEpLiLqQNXh5dUx","slot":83972259},{"blockTime":1625218512,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3knxZ69KiArAskUMEpFpqeRKAxAmrbTjjHrhytm6F7n5ipUQm5uLEr4E2DiXdUqscBTK6smJZv47Xw58qyEshbbE","slot":83972256},{"blockTime":1625218509,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4jqiyvAkLy8gWMMUoeHk641Kb43vr1oUpTKBQRumpEcr3wBeqtigk9HktoiacxMTY6yiBdHbfZmDMFBMX21RirqN","slot":83972252},{"blockTime":1625218504,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5Rmrds41TgPrpLfFBggGJVN9aEh2cqK1mtv5PuELrjMyQtzTdhFi4BpTc58UvyyYnnVQsb53UnZddSz65b3tQPGA","slot":83972245},{"blockTime":1625218504,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4ihtWc3zAQAgjuC8gwsK5F1wp6GfzVWAbD9tT3acjxgohe1JaA1Eh99kyKzCYL4bUccdTrh9nGxLwNioXLFXxcUE","slot":83972244},{"blockTime":1625218450,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"59pRw2aFS217yZxtwHBvL774cYoFguLnfzs3Ewe8281Vt1GeQTxhUuuyMthpGD39maR6JMEmWkFDQr7H8DUMoDeZ","slot":83972154},{"blockTime":1625218437,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3kyt3aJYKA4ALhWNsMgHH9SvvMcJKp8FrUDgzmxunvjE2RrkUNRJqN2jNXA7pTKxmDi3BWXEHwsYFUrTvc2HurrJ","slot":83972132},{"blockTime":1625218395,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3iPE7t6frdmQR1NShbh97zEwB8M2S1kc2vkiCF6xMNzErP1CpWdvkBEU99VXm5CLXtQSZqMVh9pdGDZ3jWHtVQkw","slot":83972062},{"blockTime":1625218107,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"Kjj8AiLpqFDqdEdDKMJsjqadtzzB5rFwerk2MvFq1EN4sNS4v7mcG4ZurnPRz2PZaPtAtvmHFskVtPQ6pUamM1m","slot":83971582},{"blockTime":1625218105,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3GTy56tRpR1s4CzZftqCtWSRqKtUjCpCcWk49HfMjc5Jhqn8PtBoVKbuPHMhUJa13N5Ha7UbMPFEeywm6pguQGsc","slot":83971580},{"blockTime":1625218105,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"29xCBKguVCf6Yzaghu2hNB8FMHcW2amodsd8tKQrYSe7zU7C4Qc8spKSCFLs7XnWAZWN2sJ6YxsUHvyYvthixzox","slot":83971580},{"blockTime":1625218105,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"Xgdrb4uCKpdoX7uThaBypgedMDLdLw1bNXCqijRQnyrcprFkDRWYKNoohrULcqE2K5G3H7Rd6JTR5nRGvS97oHa","slot":83971580},{"blockTime":1625218105,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4d1khQHL8fFku4FuRsiXy8H3uSaaQgbnMukBcUQNbYtiWChC5UiybTy566pec387kJfLCeBNhdBDCmsHFrFp1Mf","slot":83971580},{"blockTime":1625218101,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5LhKCSy5N63ExBH39PW5XTXXFQGV3QQ4xQCB3KvVz34o1eHPBBJgq5tQwiWFfxAZGimw47SEdeLd5SxhkLDjtJBG","slot":83971573},{"blockTime":1625218101,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4ty5menKDXpqaHvic9Sc3WGCdffDpnCzfMzEaVvjAtMiDNnKYCdGBf9MGBvfAa4LMPhx6uWRgfZwDN8qVJ3Jd4Qz","slot":83971573},{"blockTime":1625218101,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"558FPCkku1E2v16TDAVcRikoxsWqZaAC1Sp51HQrCVoPSZktRsA226qzAp67k7HQG5fpS2Z9fKpnn2T9rY5guxc9","slot":83971572},{"blockTime":1625218101,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"43aiYixNnR2KENtNDiqiTWVHQGmGFWK6RbUcsWvXxkPLPQGSnpFD6QocvUEVguoLY5XFRqN3Df7QDsxiv8NyoEhN","slot":83971572},{"blockTime":1625218101,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"59RSi19h6dxyfmJKjYoCiU93i59EiSEXQorWnoHsxXLfpTR9hhZuk9vh4zSHBPPqMACHaAZNpNp5eEH8UYhNMxow","slot":83971571},{"blockTime":1625218051,"confirmationStatus":"finalized","err":{"InstructionError":[0,{"Custom":21}]},"memo":null,"signature":"56GwQaUZWAH3HhxKgUE5NtAgXYA7KRZTjoBSMfEo16fKiYM6yWtcH9EaHY61ccvfZX6b7qjtYwxoGSsL5UWjUiVr","slot":83971490},{"blockTime":1625218050,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5aBWZhu6KPZXq7UTgDQFa6x4sFqEMvJgbGaCovkLUDxkZ2n2WMjur6wsrB8SFTpDKm5p1nin1tvBzBZgKz1fNKaY","slot":83971486},{"blockTime":1625218047,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2HHwRUV8XmhpApqW3xf8NXWwQv9CdDJS3WGKHH3FeqW1howDrzAFGMhbnFHEqBNrEgSSs5cKrCbdEr7oft7bxMXt","slot":83971483},{"blockTime":1625218047,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5MWCeLRWLVYJtxHcQ8TrXZpJVqiUa731Tvmmb1PkxoEmCiuAQc8Fo6xquvvQNKoe8nCcwPJM8qmmyrnb216xgt8d","slot":83971481},{"blockTime":1625218045,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3eKThS1GAXbFuJAZNhp9n4612aeNFNv5dor9utR2pvDy5VRR5jNcUPcLk3yeAv1cdsVAE8YdAJkuoGLxmxuWCHJf","slot":83971480},{"blockTime":1625218045,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2N2QC6MgsqD3B5daaBnvX6sFZFuFAo8DWZzW5hoqbDGb8BMPCa3NkF9eqyHB72AqkCCKqfhBj1d9tjdgghjG3yYZ","slot":83971480},{"blockTime":1625218042,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2Tk7BumroQVyULa9YEEGFSJi31U6SJLNvnBU4UFutZDkWAU2hCzryjDs9B85hLnLJjTXwnzf1jE4UEhTEvRiekpz","slot":83971474},{"blockTime":1625218041,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"25e4cS3TvPeKFQBCNMjUqbssdx65EHNuSNjk7KuPkFWbmZLrVjch6uWj8sZ29KzyKcfSsaA7sfELj59sjdDVAwkC","slot":83971473},{"blockTime":1625218041,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"26H78B9j2GaM1g6DTxD9YqPqrf7ewKZicJQj5oXrMmPSmBPKfJqYEM8EVokkqxP8UYrNCU8dDD5otLRjGqYn7ncj","slot":83971472},{"blockTime":1625218041,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"SbuqUvfzdBWfT34dB6paGRjKXQDjrPzQCpC4wtdRUpMEFmgW1Yj2Tmy6rz2nFpsVvxWySAKi4vPJSU9ngs9UnBz","slot":83971472},{"blockTime":1625218021,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5wfeHpxNyjZb5UdZa4rDnWCFmUfH6yNWQ7NwMLRn4Beo9abNjKuRzWLLNYV9xg2gLzvFPBL9ynkdXLwPDHo9KSLX","slot":83971440},{"blockTime":1625217466,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"66nHfoYh8hQuM5Qd49A8PsNVvRRxLpgaHx2TekpisW99sFNX3KE27osBGBsSrgJ5r2QajHs7BQgNPBPehzBSB5yg","slot":83970514},{"blockTime":1625217465,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2LVe1aHXkacNVuj1CPvFSyDHfV17fEYrzyJDG6BE3hpGjyadfBDLsLTxCNq5YuHk6Pf4i345XzQYXy1Eoj39micu","slot":83970512},{"blockTime":1625217462,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5BmhTGJPtu8VpFnDywpet9jnRJE6idq9cHCrDhpyr9HsgMJ6XLsx8zLJtS32RQHFfp7bfonTrEsGwwSfycYN9ogA","slot":83970508},{"blockTime":1625217460,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2Yd7zquS3FRuGmjiCZJbvapBQpKsQDZGnmrRN9E4t3fRCTBJKCaBVRko4s7XRX8i7EVj35ngQWrYZ8eZc7CSZ7bG","slot":83970505},{"blockTime":1625217459,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"fX2DeLhg3wzFrTjaRzgmUqczMtLJPecZFDdazk6rwFTzssqXRaQeE5uQSirdfN3EWx2Y16B7kLCMHPPeFAFJc9h","slot":83970502},{"blockTime":1625217457,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3AwVWoT27hTwLqR85tZAbZHz2zt38rnPkvTvWwVgzuneiyVHXq6j6gxCJaUQTC6BFmW1L8UChpXbBtceD6NXeUrT","slot":83970500},{"blockTime":1625217414,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Dds39LroA6pHLfSqeXJWhA5uVisyWgiu8nDSZKKgPMRKuAxKzSZpyTjJH5MLzpkqHyANxXubKajAtGs75dgYkYo","slot":83970427},{"blockTime":1625217414,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2wCMbAhSCDQ7eTNd9gHB1EsB4HWpeL9EDGpY4P15MZvqY33JuEpaoqLRUBFqDNZTcawUGaoFZeA2gcUxPvQYCcDa","slot":83970427},{"blockTime":1625217412,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Q9Aid6EBG9NXA8qLEUhSLFCf9iRPKn9V3W6tGMJHum9TXHtaiHsifL9QxNHtpvu2SRCF6Qh48jaDRoegbbcQQpU","slot":83970424},{"blockTime":1625217412,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2xgwBwRUDU8uq8SZJMHUUaPucBNL1RXYMwCa2SCmkbPWbRcnB7pqKQNFbeZ5WfxXJAY3QkrA5QAjNhtwWpfi59SA","slot":83970424},{"blockTime":1625217412,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2WUSUkszEQEcASdmBcaXdEniFk9nFVaBrtgnpQepx5rUecE9uomV4mGiYP99ZCg3mqRfX9NNwCoeLELDFE3MivQx","slot":83970424},{"blockTime":1625217220,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2iQmrfoDZvyuopzr22rYocCfPK7eYnZTDpd1vugkbhCWDZtSXQ6svLH18Wr3GkSJqF35FYwBCjbUeYNPrb2He8H2","slot":83970105},{"blockTime":1625217220,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5PCDQA8GYqwiQi6A2aZhc5dT44ixjMgyJ9Tbt6JeGcHa3Y7R52pPydrwuDHY6C4TG6Hk6ZLKhJAppk7B2SUJmb3v","slot":83970104},{"blockTime":1625217220,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3A5sgUXHjeLPSPkNRumjG2mWqpB39X1jXbjsxyTFQW3zdW8QbUa5KDiq6Rx4UQXPzRZmkHBHEYfkaotdsikzLevm","slot":83970104},{"blockTime":1625217220,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2MeRc5dUfrcsme86xRrjcD5bBVvt7A48636yZKYmyLXcYfYwhwPFPFmdKNNm26euNdfLcQtobJeF5sFpiJWdKHXG","slot":83970104},{"blockTime":1625217220,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2C8vq6Vg9tSavEcvGXFmSEUVCHVtHMcLVHLcXFnfiZShuhapwHYuJmrSX7Q5hvUGtgVA9VHB8UyX6ssPAq6A7157","slot":83970104},{"blockTime":1625217220,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"c2RW1FGtsomdsTBdqSLJD2LuXjvjJ2QWFfa3PWYLxm1TWwukLpa1dqbzjztc6Cbj64w2Nm7UggPzuib79numZrf","slot":83970104},{"blockTime":1625217216,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4qWDR5WA59kE7H1LseQFC7LyPSwrSYnYvfgrhaA6HUXJzF7A7Lgc7dh5VvSaPoEkiksjGhY6yZjzrpBazbEddsE8","slot":83970097},{"blockTime":1625217130,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"25kMB85xirNRiaWwMF8QXiDLUGx5ibW8MHNd6f5BBAEuWtQodbfpLwoVLEjs1VgPFMKVrQeXTXV5WpvFFUSUQBNp","slot":83969954},{"blockTime":1625217129,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5ADL1W4rS7h9bceuM81XLMourHnXErDDDc8ZjYm4uuashaydZSfc5PxVePd37wxYYHx8GynVPQcVBBJuv8kpV6J8","slot":83969952},{"blockTime":1625217129,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5G87xdjm2vnBc5MZHqN86j8KLwdp8XfnHVXRooAu7PiQdYFLqC4bcGqTzwJPAQkv5Z3VqEppBoyMRivNr4spxMFB","slot":83969951},{"blockTime":1625217127,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5McBozwGTnhTmqaEU9n6zeTYHgoB4snycHSEEV23bCXe4E3DAgdmMkDwNqCXSstwsfXkYW25dx6BpUDeino9vprJ","slot":83969950},{"blockTime":1625217127,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5xoZWCZWe4r1jASVVuDr8CPbJ4qYdc2Z96LDE2W2JsJnYXK1MSEXwac1fBtkMgGv5sYbb8Ax2wnuBd5gKdYncvT7","slot":83969949},{"blockTime":1625217127,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5EE1EcS2STLT6coaW8keQWaBPFxCDiC3rYxeAZ17ESA9KpYRWSTj5XJRLRfyZ58HGyPBFTk64nfDbm56zN6LmoCZ","slot":83969949},{"blockTime":1625217006,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"67V1vkGNDzDyWUURMYbL7V98x1eDvC36xFwHgo9AGw1n27iCibBB2Lb1v6wrdJU1wYM4hsZVf6MshhWVNg8L4t3b","slot":83969748},{"blockTime":1625217006,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5HmHmNy7vSitMnYtnz3LPQwkcJH1wEMsSZSsTEUrBx31cJrepeGBo8FfgT23y5EZWHGTZvRVsyc3HJfKrVAbpqCk","slot":83969748},{"blockTime":1625217003,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2sXoGuxVRn8RQie41Nvr1MEGCR6McezHuBgLm4Y3yfH3JL1vgG3d8wjcDt7j8osnnoKF5EbXuhvUUDzsJmRrjd4e","slot":83969742},{"blockTime":1625217003,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5nWLTrysJRF1iSvUSskBZo9CsjERaPz7n88Y3nBQ9xM9fvXRcTLz3T3an6s7LWcJMMnxRin5Xrz4oJWaF5p1Q9gR","slot":83969741},{"blockTime":1625217003,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2y2P57DLTPjUyZtEAPgmY3V6dgftpnKbkkvnUvPYPSB39s5x5PtPJBzuVBeBRR12PQwov9y6DzFkMYUACb1FJuPn","slot":83969741},{"blockTime":1625217000,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3WnpYW4g4qxR5SVdd1XVXmC5ZiXq4BKxtfHUtNbWX61Ebk4BHXEJnDxdBNTsPecMr5axNv6r2JwUUZ7n3BEpudB6","slot":83969737},{"blockTime":1625216820,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2GC3ZR3oLtgBpeBToXWfma1jBgv14WtnAh8hDvS3estbzfzzJ4TjnNSViUibzqX1wG8jpsriyuBTVBbmGFZy69Rq","slot":83969436},{"blockTime":1625216571,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2CorogicrKt5pSLpGWMULiBtEioffmoP2hwEkm7BypJPv3jXnb9LAR45TPo95r4kEzSctAuTJFUCChZ894JC7zEQ","slot":83969022},{"blockTime":1625216569,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"66RDkEUpquzZkkqPd2AGLjCs2vkDiy1AEqEdDwn1KUHvCBXjTFHMb4DfZdVBJPQTTwWDGwswrgs3TmobyoXJ8buK","slot":83969020},{"blockTime":1625216569,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5eikTj5v5ggPZcX5WHKBk7bApdCwB3CB7euo8cbJAmKnS44hFx5R1VpT6yUFnk8BMhnLF8eBhH4WoUWhrYeMAHWo","slot":83969020},{"blockTime":1625216569,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5RPJiURCdiRHwky2dHyvRRD2YcPZWonhA1tLBEpUHRWbaMa6tpWMTiHzR4BochtrdjpYqwRFa9A61Y4CwZ9Ngfrq","slot":83969020},{"blockTime":1625216569,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"32bJypaLiXJmiuqDreSN4XiFaeVYJfuvA8HgbbYfVJpuHbaVEJBoRwDGbVrmBbtcTF8zmBKHUEAyDsf6XXZdLuu","slot":83969020},{"blockTime":1625216565,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5DpzefixEQ3uwuFKZazmqFcjE88zvsGpDAbuHnYAjY96UpM26xf93e2j5qN9KbGgCXP1tEdrhAyY3wnnNX9uyMAu","slot":83969013},{"blockTime":1625216509,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"62hiVNLLGFtUehUXhaPPHES6xavNTeLuu5gc4qAxHRJdDvD7PQe1HN31CChTNSuiDQtfS8Aq2viNAhTuG77C5inA","slot":83968920},{"blockTime":1625216500,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2ohHE1dposikby5cyeu3gu4WKLv93pPJnXYVxmejKb1MV8B3cdrew6JWDec6Cuer96mhL1mu9jJY3bNXAkEE6mJ6","slot":83968904},{"blockTime":1625216499,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3RCQ1bVV8DrVpyeYsXWZZ2SCqugVdp61yP3EdJ5AQyUMWaMmJuPkbrKFA3AnoPUYzN58sowzyDLDn8vdzLXuKHCu","slot":83968902},{"blockTime":1625216499,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4ZF2danbeDKuva1KudNkqt9N9jvK9Tgt8dGG5iL252tXobUUHFxj2R8DHGp9pth7HfVnQk4g8DkRKQEvYZbsk1Zw","slot":83968901},{"blockTime":1625216499,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4SxnTiSesLstuL3Vfjafu9W89eJ4S3keawNPHDPXCY2ijpo6tS38iuf14Z33LMj3RBfnJeV5Y8zWAEir1ZJoemtA","slot":83968901},{"blockTime":1625216406,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"22e7fgp1roQr54DnRkUad3yZCisTYfZYy4CbEcp5bkCvM2XooLaZ3nPHbPbQwk1w5pnyPZ1ZzFpir21reeCoRTr1","slot":83968748},{"blockTime":1625216401,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5xmW1D8vgDwRtHf9VYA8ZTt2DofX9LtsxBNwmkHZU8PY8SdDi4p65i7sPmF629Aoz9GXA9x3gQSingfCU8Z6FBSs","slot":83968739},{"blockTime":1625216401,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4NUM2SA6FDLBMmLPk7XekhxBvBamjRgx1mK2Qg5zra2C9Z7G8iyr19qYsT4BvW51KVW3ULuSwbBDBq9VJuDEQhm8","slot":83968739},{"blockTime":1625216400,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2fJdFQK1NJMLUXmVhXBzwQzQ3ea2NGid7oeUCuJCsmmcXnvPVNyxnfrZocFZQhUb9dXEq2ucErxEjGKLVyvBuSCk","slot":83968738},{"blockTime":1625216400,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3shhCyWWwB3EK9xMCY9VGDSYtAYQTKCBjk4vVRyC5F7newxgEWam23BkLqhBEaoC78JKaTEPokJUFBvd2MkQPPtZ","slot":83968737},{"blockTime":1625216400,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"57A9CmcmwjuQThH4GpE5ctS3etbD7bJyHkFhUX69uGq8ywpXYFaWoMtLCkSJdkiR9WnL6qeBW3baXvGd1YEnQuVT","slot":83968736},{"blockTime":1625216400,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3mjvMB4j6xtRvZFYLhXoeffV35HMdMJMzrURLFFaXQ7huKyA6tZineVPPABGMXTh5tpumi1C8y9Wq7c2PpdzC2dn","slot":83968736},{"blockTime":1625216352,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3CXveasYjVry7ubR4XBH7eBD9UUybUCkbc2VjkJWWTJm38ezcDYkuJxGPosi8t73e4ovm7tvPHNGQqYs5VCNN7jr","slot":83968656},{"blockTime":1625216284,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"39YcNh9md58vH4kCppTQ5xdfbK4UAZBuGFaTVRTBBX92ZTpuXiJN22FZjV2h3gjC2zvJ5xLwRk5Zfecpychz3uYC","slot":83968545},{"blockTime":1625216271,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"8j6dMMx4TXb3xFCAXpRQsxmPuBs7A4KmTV48nvx9SfT9XxMnXBTCT89QAfaT7e8XSRScd6Mzhc5J46kTCQXRFZA","slot":83968522},{"blockTime":1625216259,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3nmFWWDeUmRbvxxX3TDZTaPyz6wHYeMbc9UY8wERwEcD9VSNCZr1jhLT8cydPBCYYmb4rAoWXCC5yJgYD3ywVK2A","slot":83968501},{"blockTime":1625216178,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5WaJuPgirKjJHijPGW5KZBRdpUXAqFWAJ7iR1SC2WNJ78qaq1jUoWPsvb4jzvJjZrkNRPHSs4o94Ba43fkpjuB8r","slot":83968368},{"blockTime":1625216178,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"243CSuXZvo4H2hcnpF9onFnx51BBztw3m3rQYs6HkrBDrMzkpd8n2b24JnxHaNXSnPkFBa6AETrAzA5NGuRRU7up","slot":83968368},{"blockTime":1625216175,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5ESYzPdoEHZ6jkaULE4zXojPoLYZDLRQLfTbBeYddZgDUA2WAbhPhoK5NM6SPZVhvEMcmqM7BB2HH2NZEzVUEpsW","slot":83968363},{"blockTime":1625216175,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5vsaYabQz6k7TktGVMZqrBRjYHaNNVVdjGb4pMDm9fJ4yVGCBCLP73R85jtzrK7jL8EuU2gz6RpWzuyPrDwUqWM4","slot":83968362},{"blockTime":1625216175,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4kxj7CsLaj7g1cabDWa3JgGXf2qsW2vVE1XrywssANzcvHG1eQQmJgMrSW8XaEdZtoxXitVD4nuuX8Y7Fy1UKuKg","slot":83968361},{"blockTime":1625216173,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"ehTVPDNq1dvcfttfcwyUetpwkuUsKpPt5r9LotoqmYv4yTQBv4D4BGVZni4mLm4hiNiGHLtwAxyStmgRTtKN2ci","slot":83968360},{"blockTime":1625216104,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4rfAAp7JZmc6Czi3HLxAZ22Mkqmt85jZrQt3Qq7VHJaWqnW9Fz9AdS8onz5tVxP9vVNJb4QBS96hYVn13F84Fa1a","slot":83968244},{"blockTime":1625216104,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4nGeSZhTkAVWwX68eRGZXU85hELhDAeugj91AfQKqhGxxcjkYLHVhEYakFcrqjfHWiAXWbB3vgFTRigQeiS86Fwv","slot":83968244},{"blockTime":1625216104,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3nm3uASLXkjkgbNxiPQryfwWSQWzAengZ4hXY7gYYFrtwiYJKEz6x7DDxf4SxZKGS7hJseajEjjUJX32xXZ11Gbz","slot":83968244},{"blockTime":1625216101,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2rGPs7nTVda3ubGkho6SebdHBZ8QSBA1dHrGyzofB5gV4mqJok54CwStUKD46Bt89Mi83cvrJwWzCe1qgPqQNZfn","slot":83968239},{"blockTime":1625216100,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5L2THVDxJvbD5ECoAHvVKAhBs1YJSsqqK3UPVk6sHZ3KagUvfehPakRzSGhr4KMwxhs9P3RSahoYWVwvjwEmx4H6","slot":83968237},{"blockTime":1625216100,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5QzghMKKQSMdxviZ5JAB2cgEdbhjAGAZfW3E9oMidTQEU2F1dgotpQvarXLAfM6kpsbkWAZNAKdCaGst7fZDVTy9","slot":83968236},{"blockTime":1625215365,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"7aWJGbcCzVaWTs5gcqH2HUYYGS3Mi13f4Wzun1gzG6exatJ9HfcYAyRfhtMQaS5fdjEYbsn8urw2pgHFtw9JbGo","slot":83967011},{"blockTime":1625215356,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4eRC2kH8wVQU56qTBLgJeKG995gf3GdNJjPDy955VVgvJZSpaXnpHdemfhWxeVxW5LoM19RpjDdwjKBB39f9tDU6","slot":83966998},{"blockTime":1625215354,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4ehewfKvRibQzdwnvAC4W5Ks6VAbxtBmxUssbPk3oLaNsPNeGHaKq7iqR7ULRtoHGFRMpyKFnUDCtQCzuHrryCcZ","slot":83966995},{"blockTime":1625215354,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"33en855JB9trZfrLf7nnXU3f2WYzPaRueaNxNmMpy3nn8VhoXtLdjMtwtmtHT1sCvSPwwQELWZjRsgUyBNXsN1an","slot":83966995},{"blockTime":1625215354,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2BtAXbGQ8DJwD743yyAR8Qghax1Yt8z82i9a6JxnoEEsFZwoF316mnGpfrhhHdV2QtgxqwhrAeJLJ8WowHhQKCn9","slot":83966995},{"blockTime":1625215354,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2a4D5Qg71wrvkqyjeVKiwGn1bXWvmgCL9nm7wSc1t1WPG3vbNi6ZyJfbRfkQT4XCLf2SWfRypATZt7AUK4BeGKDT","slot":83966994},{"blockTime":1625215320,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5H51FrAdLJrFHFHZsCPFLYTWSno1V45uE9KR9XKi1Lvbo84EC8rG3y5WHZfEFEFioiaepBhf7rJmnDFyhKLvHbcc","slot":83966936},{"blockTime":1625215320,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4mJfjVoLtHNxgMCmH92RU9rkSbnz2VcRkdQkZWN86Ja6Li6o8Gmbc1XPUu3XRjnx33pcQ9hH9wx91B5fW4PPNAcj","slot":83966936},{"blockTime":1625215320,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4T1NPWWdxsct9quXCPrwuNayRz33bgKNudpAxSdj1eevzXh7cCJUpWwPXofS42GUPffALT2k8E3useZLLprrnwiu","slot":83966936},{"blockTime":1625215320,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5rV9nW2Ep9cYLH8VQAr2fYY5jnWRzJXeKvASY3wZEj7BcZZTxA9wpERGBGwBUC7t5fEsYh8Ms4XSg3A7MUfu8Mi","slot":83966936},{"blockTime":1625215168,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Py22YLHuxTYPVa3MdPwo9USTXJDt2ugGbtVBoF9zuTFKUYBhr46atcD8witjJgXs1eFgEiWJnYzLu84VtnQDYsd","slot":83966685},{"blockTime":1625215164,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5cBnPQV8VrLgHsJPWiCndtU4kChBVmxXtgc4wh6WmtJZe8FQrnLhQ1WDBbTGTohMscsNpd4LsRsnUX5eECgCN5Sp","slot":83966678},{"blockTime":1625215164,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4QXkYfS8wrewyuZiJzXuuRdKm22MfhTMj2q6iLDdSSKkizobMen5Q34ze1tLAqi5wzgLsUzjvc1QsaKdbQ6gogJw","slot":83966678},{"blockTime":1625215162,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5J1zuKiPeRZt8S8DjqXZWd3aL4XoD5nPS13jMLCUqsBeTfrGT82KfXPWG7TdFNrC9sF8VKFmeBZGuWhDQMNfxrRY","slot":83966674},{"blockTime":1625215162,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4w5JSgXCSUakRFympmroX8turQVJGJRvcpvp1VoyG6D1c7P54bBvmhobr8AFtdFa8hHnLHUofFrUZ3RpvhbLnqrx","slot":83966674},{"blockTime":1625215162,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Y6ydzukBiikb3mxt3BDVdutM6dfUSLg9wEWtRGgS1xqJDpUPVEDWbwxM7SnY9TF1PMXDKtvjhfuunA8LswHZrwg","slot":83966674},{"blockTime":1625215065,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3bueKdHoNYFZ1RimqyBGdnYvjs7VmiYqgyZ62d53hMTbtozz2eseNqWS7VR9gsgTV8CAu1HwQWhnZsWXcQrEiuaY","slot":83966512},{"blockTime":1625215063,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2kBZMi41fymwDGLMqiV3ew6U8U6SJfx5oFnyFry8Dkxc4yfqiaELYjiY6aMXob175xUbdf1qH4MyDQta5m3XG8jN","slot":83966510},{"blockTime":1625215062,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"36qMy6Pn6ehyeNYSS9eMvF45t8s2SdAgCnHeyTVhv8TziwUKWkSGVMcPR8fkVPp9YxLwqpm71zWYAuNt7c5z6p4n","slot":83966506},{"blockTime":1625215060,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"55Y3jzS4CsMzw8a7A75NUUBBzvm4o71VC3cPAtSLU2vbHEAkXw11kD4SoNRUGLqV4MJ7k9HdYTP1obh2kyZJtLkF","slot":83966504},{"blockTime":1625215057,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"rSQrJNBj6dBDcuyaKW6qXwNxwwG7FZ16TenYQuDj7fh5g61KvDnDnLhk3zpB6bcm8RB2C6eRGsYDc8PSM3W4fMa","slot":83966500},{"blockTime":1625215057,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"QbZqk4g94xzKz8GEXypiWrJSEC5vP9CNETksZGBS5B2Pp6zT8L5pBAc3vCZ9VCFa8cDeWwVZdy75t4u8Td3Ks3a","slot":83966500},{"blockTime":1625214924,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"35SHGEBfGmBcoBpcmNwzTUKStLiF86wVY8hFBxDoXndVf3aDs6bS8VkBsXL2Ut2RVnSxhzGi2emAt4HTT31s5YsK","slot":83966277},{"blockTime":1625214922,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Q89zvxRWWNnaM3f6VoK6MdKk3hbCmZssjGmdov5F7hPpBoM54XcjiQC2MFXZeyvtpvgC4T11WLsAVfNcHZFYu3E","slot":83966274},{"blockTime":1625214918,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"a3r2GJqGCBhhnB2vZfkVxuReoACzSMMLrRsu647vWsvJBA2wpgW3kqJowUZp6TGGCocpN5gx7UegaioLpFCoA4r","slot":83966268},{"blockTime":1625214916,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4zqSJzbdr2yiiSvk2sbnTKAAfoi5UFci2iu5XRZGjsbmsvwm3uRwUYZm7DJ8rLF3vHpsytfw8C2BXVVTNio91WJT","slot":83966264},{"blockTime":1625214915,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3FtL6i2AC3PDAcGmZxU1tJuGN6Qzh2hLd1NPL8FNsEZ69cxiV6NKbAjbJ2JbCrPbzdSx6p9aJN4TmMKQ2FSY7Pwn","slot":83966262},{"blockTime":1625214652,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2QDbR5WiyVREuxRTWVh2NMbtv48cGau9Bn7nS11s9K8DUnz4wjFh6VuQrK17rPscYm7dV7iv1RSTdPWwE8eakrXk","slot":83965824},{"blockTime":1625214645,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"jVnafw1FRXGzjSQhaqryW29HgaQgBSd64peMs4P43MtEWDakmBdhYChFifeCJ918WR6FKYMWK4gPTE1WccrAdFr","slot":83965813},{"blockTime":1625214642,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"GPSwpvvGyFBTwpG1vpD4eZeXCF52HcCWuCmrTqbDSTm4Mnye3MuofZ3LShxr1X85SgnpPC3uhn5Pn9QRKUkCmPE","slot":83965807},{"blockTime":1625214637,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"89yWCfdz3W9WeERHDs63Jr8o9Gp6y8fwvBb24DYp7JQbfZGtTtaXdFnbxsA5eiXhzSnPYKESteuebbq337DxuPe","slot":83965800},{"blockTime":1625214633,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"m3LFjZMSmUm8n6vzAhWEGN6b2XvsZyVT44b8x3TV47KdeYnNQcyPWZxgZn2gsZNk1tDQXN65KwHmfkkVicJF9GX","slot":83965793},{"blockTime":1625214207,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2tFpzsb1kkgysedYbAcgkArhAYZSaHvtbJ3pLZwV2ibTP4QYCkeEKvskRq7KH2NAy1ER35HoURshnSvt1JRjkzwt","slot":83965081},{"blockTime":1625214207,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2rMTu36gA3RgCp1gdxREPB6hkXN6FpRN6fhTzheuVEu692gKZjkDnM3tELkWDXsHrPv8e6ZThRXrTxEYTq1T4v7X","slot":83965081},{"blockTime":1625214202,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4sz7udq84QHcNDMDEfwXabdGFbgw9DHJQZ4H7LZJwjrPecfEFdt44prRrfkRyZ3KgfpYzwsEzuSFhKvn52WQztr9","slot":83965075},{"blockTime":1625214201,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3doAfW4XjD3cWvXwfgkPWm5dmjrpYirCt8rtDHnTZCo2EEy6toAsyZmNT2XxCmbs9h7a44qSTxzXVroLyjCc13Pn","slot":83965073},{"blockTime":1625214082,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4DV9Tv2hcd5ZMA4Uz6Joxi4WJxqvAXLeS1wwjVaGMevHE3zsoBJx8DPx5BtzvurKTjBQ2HNSXJwwSkX4GXTy6Aho","slot":83964874},{"blockTime":1625214081,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3W8U8TCYtZZTabFZKiTvhYt9WEMzEq1zDBi2CByebsqkiMvB8b3g8HM1TApEANz8JbePDU49U3sgMCEhkhswLYVv","slot":83964873},{"blockTime":1625214081,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2yfLdpy1yEZLdszAZKcAfcKU5BW22ebuT949KGYPZMSH2fp2wWPskpK9pmn98JFRwj1U4u1aQJH8CtqL6vJ5Dnhg","slot":83964872},{"blockTime":1625214003,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2V4MoSg89KM7aBuVfenTShVcVD4QPcCnDkVJXw4iA7yHn7sVHgPRZq5RsugxkxVZ4cJDDhFCxt6U1EoVQUAn4PHh","slot":83964742},{"blockTime":1625213983,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5tZUAvs1FGaqVgtJvNCkeU1nAfnb2pmUp1wnBZsRVrBJCmUcGTJAFwfmoegYstwwLmDucMobgZEWEQjxfSetGbs2","slot":83964709},{"blockTime":1625213982,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4qZdZkMg1ZFnCs5kzxjXZXVeFwUankf6GBNPnZ1rgyAyRHyzxYqHpscPR9CZwsZtFKsDZvo4CHvEyCx7xhtopKDn","slot":83964707},{"blockTime":1625213979,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5xDZpNtdLtfJp3WT43psyt7Y79QtHW3NtixLQkZZqWuZCQkFxoD29NCiaWDzDp1jsrVyyQ7FWdK2SGkxe68PjfjV","slot":83964702},{"blockTime":1625213737,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Hz3ENce8S3R67YqszmQXhsQFJvKSGunvhWH6BuyRTYNtdFG5n2V3UH9G7VeDwPnfGwME2gFovuXqDAyU9DBQqJz","slot":83964299},{"blockTime":1625213731,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5PG34aZUb1joQh3GmkzPv5p3UcN9HvYJRkd7dbSQ3yK4aViHe6LmpgqPoai91tMmgpAmyqJECsrsMp91hkNEGMFC","slot":83964289},{"blockTime":1625213448,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4cknry4zDwCJvtHuBWLeFrRXCpuWT9hra5cwn2yGRYebDd4Y9PvATKnrTVX4Gn2B3yFsdDhyPhQzvQXNLeFNdfYv","slot":83963816},{"blockTime":1625213442,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4bS6tsgmvGmtot8NBsAACYPWeGKA48aUuL87zE7betdAYCxmV6nm5nvqJP75Z92y44KrZajpvEBW6F54DgSR361k","slot":83963808},{"blockTime":1625213434,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4zmJA3dhG3aVGsx6qd6mhEGtbBqZDdX414s8Ja9cP1cCBqdRNhnDQLzesE7oMPWKRsFH2pp186jBzdLnfnU7Wd5D","slot":83963795},{"blockTime":1625213356,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"L6pncHwho6LTWSYFwvvd3MskGhe3558r8otfq7BnBG5T6mEmoBj64muA2jAG2uxeboAcynXj5aUbDGd8838uq6F","slot":83963664},{"blockTime":1625213349,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"573PzQgF7WHCf9pGJiPHoHqdmCQKuZvFtWG5utEMNEhJrndPs1NAEkHTCEdmjmoYzGuH7WEfABooQi1SaF6koqwB","slot":83963652},{"blockTime":1625213316,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"gSzGoyaywbsW4EZ4XLTKzGrfQoDPq8Yk1W5eGwAhW6Vmq9pQhAPwnGp57GtwhtMLkhR5LhabTDNDdkE5yDFCSVJ","slot":83963596},{"blockTime":1625213310,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2oKnGN6d4bX9ny4QMLuqsSx9x5MArjh9tWVZkBD5t8mVqWgGRi5JoTdZSRw3kTFpAopUKCUV16ykgPXrySBt93bZ","slot":83963588},{"blockTime":1625213211,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4kutykd89sWqSkGkRerXR2TLv9busK6hfrh4nTSwAatz51gqe8up6xKMQxaKZgXHxh3jRUuoMFuRf9xJ5MfF2hWK","slot":83963422},{"blockTime":1625213163,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5T9PvY7fQuFiHaYij3tR9mU9XRHnDx8xSP8avdGdVmAdCmpYZYJrTpB4YUC8cN44cYzL2ULnnz2VquFxTp1vtNY8","slot":83963341},{"blockTime":1625213067,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"24K1Pnxj1iz6L1kHHTdRZTha2quNRi3QRtRs3tiFaQ3CjcqCP6CxSATB7hsEffvrswMaFmBsEDYH2vyZxWWZPpTv","slot":83963183},{"blockTime":1625213061,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4AtUDQ1AJsszvPTtcyMLoqX3KEJNXtYecVtNWWwfv7KBkpK5cYqHhdyKWQMPSHAWPE5BauS3Qi7REraoQBMhGLJq","slot":83963172},{"blockTime":1625213058,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2bhzRka714MonCmYaVAVRdCde73pVTVveY1X9r3r9hjMfiqRUfb1EH8voXWseRhANYisVgR35C3hhKAUV4GNm8yK","slot":83963168},{"blockTime":1625213007,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"57kTp1kqZhCB4XfbRLFTcpTUqtXtjmn9kUCa3KinXxXQMDh5jpgU3uSjezvzqrMUcGhAYSdt3X7N1SZZtsdkR95C","slot":83963083},{"blockTime":1625212830,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3LmX8CTULidiRwTGj1irQ1XEThqEKMbkymhHynpbjRwT39u7Me7Ri5pbxVYEZMxHjhbjpUoBnWhhxnfWniVCnSVn","slot":83962788},{"blockTime":1625211786,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3DKewFqYN7JfxN7YM8VK11ZCKzXLkrpo7N9fD17AEiCwMiyXWLFpkd3xWAeWsqewdi4EjXTNKM3DFNWCE4hRe1C7","slot":83961048},{"blockTime":1625211312,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4j3Fa84HvvWwhC27GMyauwXj5VteDJRMxC6cPQu1jr66Fg96sKTJavUqccijjGfp1ybR5hf9fW2UoaY7zT88tJyr","slot":83960256},{"blockTime":1625211286,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3cmH7MVVacEJGPJ9nzYNxLtPMTX18hmQce9eiiRkCohQDNHUfVMAFzM6rcbxfybrJt2kAopbqs1oVafaq5CJPSYc","slot":83960214},{"blockTime":1625211244,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4dBs9Fyn1h7BCJcbWiswv6uehy9um6rBvUQnBgdDFeTzhqiNb8WssrUZAsfLonHwCv1BFkeARdWCpTbTkrDYHz2T","slot":83960144},{"blockTime":1625211081,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2C8r4r8hzdrmJbCr6ryz7WGYv9JHneik1rCEVrQhPrsSFJvKEbfP33Ja28DHPDkPShYwBAbqj4vvLFLwts4QS8bd","slot":83959871},{"blockTime":1625211069,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"37q49R6o7X7Tj9zbeY3hLschKJJ2xrmdJQJdwbVybiDM8a4UxfNkdGVun5iwouNXhMmpPw4UuS8hq2rzkCQEsfVW","slot":83959853},{"blockTime":1625211024,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4eNns5trom3pDT9wEZfJddU2QHijkdQkS6FymS115Tx2N73TtovaKbWpxdcvTG66Hm762UJRxkRkCBkcdFfNWNJZ","slot":83959776},{"blockTime":1625210506,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5zdqw1Kq2yv95TGBF6MA1AWQqjpaZZe1hkGzHTwdpt1s7nxDiMrWnkMDNncEJwBH12GErNP7c2dQrMVZRSEvtatt","slot":83958915},{"blockTime":1625210506,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"D4kYTxtu2BGqEqihmoo26dF8yQmf9zjJBEFJBWqcUfGN2N878bYSNSSuKciMmXDdyxJ11eKexbF8KpYNsgH19oW","slot":83958915},{"blockTime":1625210298,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3q4eQk9Ao1zHvAidemSjE6njCNXdy4eF17qsHko9cQ3aehTrdYCcZ2CygKqy5ksBHgSCq25a4VDaLgb8nX3Xqs72","slot":83958568},{"blockTime":1625206813,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2pL51GghQqM14gqRTRauSPMNnTvN8BZEK9o6qvBTwqnTKE9uS7T8Vn8m8uPL88mTJsFZjHdPHfbtEgmpxNdc92rL","slot":83952760},{"blockTime":1625206809,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3PcAUt392biA7G5d7zRcKYWjyTc1qi7SL3x2NktYRBDWWRdwUCQzB5TepMA5HkSByQqqhi6ThkKGLdoowFa44pac","slot":83952752},{"blockTime":1625206804,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3QqvTbhDTjx5nHwW97JkWkHnXR8Sr7vFcbCboNjwZvBsgtkPdMcm6b3Kr5LQJANpJrbXkbA9T5us7ogNvSyyweU7","slot":83952745},{"blockTime":1625206542,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4S21Tf8nuhNSm7TiEe8hAxq4Byb2FFwWA8qukqKVXeeT3DSgEmwXnh6GGECmLKveT1AcHFFovTummxpztGrp5iuG","slot":83952308},{"blockTime":1625206542,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5Sw3L1Uq1VuvHTZT3JMqBS34rVURwBzQps6jFub5Dhrd3to2cARe3cSeduV6LmnqzqdsR8KK6f2bAZnVjHfzokD","slot":83952308},{"blockTime":1625206533,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"67H68yGURAXMPkxeGSAeUhvgaGw36ktE9srYtVc6QxbQp2bd3RNkbXrCFC1NEbHt4BTTuYcndDPHWBrjuBSDC6fd","slot":83952293},{"blockTime":1625206290,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3qgXDz9rArSVXwVqU4ty3mDSiT8wLHBjeY9HvP73DR12z7RASCjnoNEeDox2FHA6ajTLqzn2Xri1VZGn4cFYQe2J","slot":83951887},{"blockTime":1625206285,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3z7jqpP6iYKfUJjbNf6PagHgnDhLwWZz4KzFxTGhLEdB245mrsvqJXgg7M2SxQ3fFTZggY1PZ39yXjsXQ3t1eEGc","slot":83951880},{"blockTime":1625206281,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"43g27VkaBn39qQNBg8XnZzfCgVH8mAZvWeWj7H5Xz3BGBPzoTwHRKfvzjMHjYQzkD8hGwn44oj14LJbVavi5fBET","slot":83951871},{"blockTime":1625206161,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3JcuYp9W82cp7zMYnCqmseU1ZatHrxPPNF7EbhchHJSMhgZJXMBqq1vBB1FWkc5qFmqCScpCQVCMkANfDsC1uwT5","slot":83951672},{"blockTime":1625206153,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"Q31xvmuJvyEBXWrHK1cnGJbFFqm1ngXDFEPxacyte8MJn5QSU3zEnf7FNRqnAjaw9NMxzGdVZDkpbGP8eEHCLWn","slot":83951660},{"blockTime":1625206149,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2RG27drtSHiFPr5Pm1p7pYvQwDBiroMQMrepRP8xacx9m27TDYwM5nZEwmRAtAJKMMGcQhbF42oLQwdm3tih5v1R","slot":83951652},{"blockTime":1625205801,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"FtsZ9wmHJnPX3Tic2TM8eDgPC7Cxdj5pMU7ZhFLoMo9gNpNGPpfTVArVpS83vw14tu2HQSmxjPsyCtf5ESpgBaz","slot":83951073},{"blockTime":1625205745,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3WQxePfaQmPcJK2UuJQFALwDyFaA5C39vxEFGfBCZwETfyQHK8J9GJeX15YLakAqwHWZ9NkAx4pAbtcdNULxdcDz","slot":83950980},{"blockTime":1625205742,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"83JL4awzRS1hf1CEFAxUMwPGyAWBWrDoU2a6MDvGtWRnutWg6qVM2DF6t3wxPgZrMAxX9z8M94tvKMBvVEYVcKM","slot":83950974},{"blockTime":1625205736,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"7LbU2BMLmcSdjM1hXUcvYLDcgVeehu6sDef7Pfd3Xs6PE9vmxhsMD4cpv22hQixaSB7WTDUk6YAHPf1ePQPAx1Y","slot":83950964},{"blockTime":1625205505,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5ARhX6JSicYgfbs6N3nqrDafGypDaUwZUX2SVFTvG5XqYjSWaFKkBqBwr68Q4FzuWyKvyf4e7cuppkvkWQnM518p","slot":83950580},{"blockTime":1625205427,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Vn6J1o7doFQAo1hQ8r16HbwSKWJTErPLAbtEaUYxfuYEREXgL6NqKV7LuAGfg87NDYfC7j728huQF2pNmcVNxJ3","slot":83950450},{"blockTime":1625205412,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"d5tyGeeckhWPQidDHzo7Am8rQvjFBKskD8N7MnDb9dMCRkqPjSLG3xXGfs2ACqqkpBwo212Q3GmA7TwVSozQ2yD","slot":83950424},{"blockTime":1625205342,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4HaeVEWPWMTMsaF18rKRJwbTeVSKdpp1s92aYWz1GNThY5vw4XCbbBmeCyh3SDLz8Sk3Wh48TVzbC9GCKNjZJVbw","slot":83950308},{"blockTime":1625205339,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2TPTNjZ3GnUN9xbXwQCytdEGRpAKSXrEmrrVP2brVtMVDdyyDc5eeTSXrNHejr6B71QgXvnZrn52eUBi1bo4KmGo","slot":83950302},{"blockTime":1625205339,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2b8SMTwV4w9wohNLstnwLmScUL9FPx3zuhH8KMtksktjNgTATrkgnvqUZUYudsCBh3fcnzsLWoQDW3bsJ24u8oxh","slot":83950301},{"blockTime":1625205330,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3hKvwv62yXV61dUwPjknnJbkBxoFBtfr75AzChXc2bMDDMWLcNN2b35PH3RXwCCJdSV9dKn1W2CxxFTVRqxz6ddY","slot":83950286},{"blockTime":1625205327,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2zNVA9pEYNdXC5FNK6ntKGgeAp9Rqnb5DoCeUsLmeZBNYiKyfSnpuSDzeRxK8PGUYJuyNkHQNX1kenX1jajSvoFC","slot":83950281},{"blockTime":1625205321,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2Cftm4L9XRnuCuEcMM2KZ8TW7YoJbyKe3stcVgx2HLv16Ld3JxXGDhakRHMhZXR5YbkBov62YMDaKTbaj4uzWGeu","slot":83950273},{"blockTime":1625205318,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"CmisddAyHd2br4aENQbymagBxxyUyRK65PmSPTz9cioMig41Abpi9cFLJy7xtoiNs4dZfFqkeFDpuT479PR4b7h","slot":83950268},{"blockTime":1625205315,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"25e2itV4FEBrQP5oH3ZduJFDR9EZqBgVYhBBTuhSkiPXhu3RKY2qA6mYbgTqqk2gMmZBujztQwUYqfBDhcy3sfm3","slot":83950263},{"blockTime":1625205310,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2qHsKE333eNZzAQSWZtneyK731CWwnvR5JNybeGXhNNNzMuvmqi6aXdL2s9sLZEmYK2mfQs1qgQJEjo1r1eHrWHK","slot":83950255},{"blockTime":1625205306,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Qsw1ntUV16Lpdji1AbLdvNQ91MCypRbSMQzAUDwNKaLzuQABvCsan4tR5QLYATn8n8G2u7fEcdj1EnPvCjj5FQW","slot":83950247},{"blockTime":1625205304,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5bCXeoJkSAYQ1m9xT3sMgGkwmPEpLmdckPfZdYtiRs4XBAFmP9x37fmH5MxpxbP6t8rELms4qYX6bmWKjYpTxv6K","slot":83950244},{"blockTime":1625205300,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2oLXKxHBKLv5spTF8TtCpisSjAg8KURcZo9zcqxVsSmTn5DtYQA1JWwTxmND18NoFBQNMWcXUYtMasfGkaehndW5","slot":83950236},{"blockTime":1625205282,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3v2zF7Hygfjy37jc7HaTM6PtczeNMEQPJ2BRbaCTVR1fRKJvNBEx3h4EggNTig7B6DyYKoYD18drGVe4wfJaHu5c","slot":83950208},{"blockTime":1625204880,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5E8gMHourXU7CPcSVz4K34JxqsFrWAB611rUmGNAX1S9oyYgQNc59BRuGUUNBbhXPTg6iXdxagMekSgzUVagGfpV","slot":83949536},{"blockTime":1625204863,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5CVc8gYwQjHvChz3CNhJ2kyo9dnKDFTn13bm3SzNKmnrquPPn6LV95QP4AP6XJzhRJjbtHsJb5U2RajtkxPHn29E","slot":83949509},{"blockTime":1625204859,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4snRQv6EDJE5eK7zF5HEGj83JsrUGoDHeFXySRfFj59pH9CnTLrhsH3VnnNWVKz5AbLEDDHYQX15ARTwFdgEbdLq","slot":83949501},{"blockTime":1625204841,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2B9TeyVUy4G4JFfndbFBJfEoAiQ4RnmjeFWZ5zG56E9YGtADd3uHK2NxUDQuev4EUAxM29RFdCF3uXCsV5dvDreG","slot":83949473},{"blockTime":1625204733,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3h3yYuPNjfYCaV6Pp2J4eL1pjw7oBqdvADHATz5GCRB5oDarKPx4JTiJG6ATYCMw9dzS7S73btfSdTKKDkDS2eNR","slot":83949292},{"blockTime":1625204715,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"Pw7c8iDWivYpVJmfd91t9xGr1f6GniYTmgG5Cy1AfB96kw3H2519efAJcGk9tLxZ99EoPBrD55A9f3xUdsdBA6R","slot":83949262},{"blockTime":1625204673,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3XcWSADa5LGFfsj6N77S1EVU1TLqcDT9QWa3rLmE31yt5vm2bb6zzcSwsGGwXENxmpD5jmze2y8TQx3roV6CAGHF","slot":83949191},{"blockTime":1625204668,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3EprfNN1uZyhC3su2rvxAjvngSqHkgR8aYbMAZ2Q5fRRSyxAQxeHXrpb53qSZdGWFXUxBdMq6u97H4Z1Hcq2yJh5","slot":83949184},{"blockTime":1625204649,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"j2eJYtpehksNcVCiZXzu8Jmx7DxYSstw5R6LBEqCX2K6M4Tk5pok1CCJ45Joix2QPq1noxUQiTHLgtL7PiL7LAK","slot":83949153},{"blockTime":1625204632,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2YLziPmReMXsMukYz2yTeBgTPHfqtTKAvtjkDfeR4dh7Hayovzf9ahs1UathizyehSjcpEr7mbYhCxaVFuXzcnDs","slot":83949124},{"blockTime":1625204595,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3sTiUg44RzngH1481jDegsg2AoYpXqAXvFBEtcXKgAnJFxW867TDz6wT8XSo6guzUgtbEpfvxZ8CYwwLJ6Yku7dS","slot":83949063},{"blockTime":1625204487,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2ZfkgCVct8ku3tY4PvoFge5f6ZyT6wx89iWscthmbcbJMbxR5cMyW6WpxZTU8w6J1iRmd3E6Qt7AyMRNiGQq1RvA","slot":83948881},{"blockTime":1625204415,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5M8X9WDDoj3JMgVSo3Eo5Yg5bnKr8b8jKudjaX1XxcT5C76T9QCXY2UwZJogGDdzL6BTfGdGph8G659FgywsjrrP","slot":83948761},{"blockTime":1625204353,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3qm8bo18SDoYQyJ1JuydPoWqcXAXqApCT6iVLDTxuLMPFJQzZ53BnR138bnnoBoW53iuk7Tc3LTcwDu7cWj1Qx8S","slot":83948660},{"blockTime":1625204316,"confirmationStatus":"finalized","err":{"InstructionError":[0,{"Custom":1}]},"memo":null,"signature":"4YDstsaJA7qK3FXaEfznrBEcdhTc7Af6QEH765JFTzNRhPWobrf2ZFMhCypqPFdjUTUgwjfXmyDcMN7pZKaFQpyV","slot":83948597},{"blockTime":1625204316,"confirmationStatus":"finalized","err":{"InstructionError":[0,{"Custom":1}]},"memo":null,"signature":"f7DhSY23PSx6q2uLAuSWP56dJKTaqSiu1rxeXjbHKjXWWDssju4B8bSy6ApZrVJyQVpVivNbyBtBAppEpqqCdXx","slot":83948597},{"blockTime":1625204313,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3ygyBMRa9QhW7YYvaghxXwZvSy4V6ECjjmMiphzvoYqvEZeaKj4oGjFCumr6xLdbuGT1hg7LsrN9XqHLUMseCC42","slot":83948592},{"blockTime":1625204308,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"24XkFDdPpBsCZMuShJ1oKZnqCLKiQDDA5bhDJzWZBgeif8DrrBK4uFD8qP57iKQSyd8fujoWHJnm7G95xVM2E992","slot":83948584},{"blockTime":1625204304,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"KWbzRucd1ZPbMVnsG8btpchZsGuZtYsXsRQZPXwT87sJFjajezak7vrH9nWKZGhDVAyJKkdw1qgvmJFYoqjmoPM","slot":83948578},{"blockTime":1625204296,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2A5zuLEY5G4t4K71ryUFQDuLHDJx8AivvUKrDzTVpkbFhcjgzX63xB75Br4xtnPGMueDz8SzYwmnLRWFBhRyJRVj","slot":83948565},{"blockTime":1625204293,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5xkimbvW5GmgBvN38qKqQEhLkEYaqom4HUxVsfThUgesN3dKPCeMFZk89d8eXfYJV21bPgS3diE9e7WGRKxZfmXv","slot":83948560},{"blockTime":1625204289,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4VPwruPSsVFeNbEdaTSzPxYGaHHqTXU19tfHVa2msKQrppiVhaDfCtuWh1qbN6SzNCTXjqN1sdSHY5YT9xB7ts8N","slot":83948552},{"blockTime":1625204286,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3yYiXHkHGb52z45RZhfioAfdph5R5xvnA6WLj7SMdLJdDG6iZ5PhbiEtybFSb7KAEHqgNjrDVdWDfxHNvo2UUhSz","slot":83948547},{"blockTime":1625204284,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2gVz8pxED1xBrKdWv6KL4YQas1VZvvEgwcEvx5PqhvKtpptzCXj9m7F4MANsXQiX6fik8j4c2FUBTETEeNw95MjD","slot":83948545},{"blockTime":1625204280,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4frJPvmyzrBBoUfQWmvEeWovAxFLVPeBj3g1S5am2RUnC9ugQQkMBukxcDVBajZV3CbwRTrXt6XBxYjX4emkLBHP","slot":83948537},{"blockTime":1625204272,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3GfrE5umxCAAiR2aQ9TbarAyNtJhmB3YxF9FRiT7yhHR62R5hf6XjZd5SNnqHoGqYNRCnudc6dCLKwjM8E7eu7Kr","slot":83948524},{"blockTime":1625204266,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"LbW1gT5T9UkrdLG5ruHjMG2hgaYCgubJvFfwgmNaQmmfzYxKA2xfoXstifZtnmaZif8Fs7Z3MToG4UER6JnmzZq","slot":83948514},{"blockTime":1625204257,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2iR1Aj1iRAQS6SsE489ytKvVCr3Ygq5gjvfsdaqexV4ULvhM9sX3eF8xpmcUimWtgGque3kJuZ1xJpTQFrxWEqfq","slot":83948500},{"blockTime":1625204257,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2FRbRcTR4ZTq6DUDfb8DA26WcgYPcunmqDHJo1tzUzPEczxjcACvXBCPuTrFP3oAhAF38eKxDAxTDnTU8KwbJ2dm","slot":83948500},{"blockTime":1625204257,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"eFXtu8PRkrp46nVmYkBf5eoXobzMcavD1s4eDqtZVeACpyvDg17gusyrhEp4kKVtbEz5nsG53fbtbuoYKuBYq6k","slot":83948500},{"blockTime":1625204247,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3QbWUctCyRtEfDnXTmzWqnGjEySXcd2CRNLxaa81oeAKr1zfci6GBmsdo21YMwhUJ236HdzWPmgYG3qX6YUyt7jG","slot":83948481},{"blockTime":1625204232,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2WDaf5Tq6AAVB14qEcLLgvoFhsJywQ5jeiMwNf8Dfma4aQzVvuqG2NALtGbkW9VmGp9NgBq9peznFvp15ee6JWes","slot":83948456},{"blockTime":1625204220,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"27rjZV8XvCCcgGfztJzVsUVnwFjx4ipDbCeqeWYfhJtuBxo8mpAfEGB3oufnwuD18nrkoTJDDaaGRT8nmBAcG6oz","slot":83948437},{"blockTime":1625204218,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4LSA6MjihgT3xgcDxEsaAa82RxABVTXtZE2gs5moccgebbN8UBy3TeWZNWFAKLQC3nzpUwQiHVM8h1PUFtr4fyHA","slot":83948435},{"blockTime":1625204217,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3bU47Q7fXmpyErYKWwc89mCJUU1B7SAPq8vje3fc2Ye6nErcxNcuvZwAwezkzA11dwaRGQxpNP1HvsEFXZ1reuT3","slot":83948432},{"blockTime":1625204211,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4uWjjZ6RPLKfmxDWyGQWTrKZsm6yL9wQbFCKzyWFzMtDszGihPjoWpB4MENi8VnhTMGgbig9JB7A9ycF9kfowiVN","slot":83948421},{"blockTime":1625204206,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2569w4tNGYbrkeKytuh1kNGqPbNVtGb88tt4sgzX4y9ixX5keRELHqnX8bNUdoLXxcxSzeg8bFtNR7zvcLuq1qCE","slot":83948414},{"blockTime":1625204203,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4ssne9RgRR75PiVJ12KQJ8uftYzXeHjAE4BoYxymner8WDCC9wFsMfHSKSafvvmwGk7NFTUa53vuzsaNHyjYo8jj","slot":83948409},{"blockTime":1625204188,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4gkgURc1UeSk6aPf3jRtYiwf89ggvG6ctEzDJ7TRrrzF4cSK4nKAeLRxNWkKPeyjb6TnjxaqsU7D97BwdESsbo1z","slot":83948384},{"blockTime":1625204157,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5RpS5VW2qk3HYkJRCxoqXZHiM31AUV6sQWYv2GgtuGEMbeakcGSUXgZEJmZfNJ5fnSWkMN9nXTRzmHSA1b3evhET","slot":83948332},{"blockTime":1625204157,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3MSVmMDfL3MjLzmPbsuApsKMmBmp72sWZhDj9Jkd4C3BENtMHjX2PdULq7Pv3w5SXH7SVokLTFBJzH3XStJkE5g5","slot":83948332},{"blockTime":1625204149,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Y1qMYyvZPjP1Kbx3M1f6PXBaisY8dZ6nCFwLLGcXHRaPJEuCYVmEHSeorCuYxMUidZvtMSKoxhwZYKKzVkVhcpJ","slot":83948320},{"blockTime":1625204148,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4k63QDCqGDcEaMa6NDG6CVX939zJD8sPoUJtydymutfcMZ3qyJCcx6xzs5i5vZNezBtKyGxYutcrySjUVPQgaKTa","slot":83948316},{"blockTime":1625204143,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3EYGsKfNKipWFoWab3na3TV8kb94XAyrjGvQyoX81YySHsyYCyvg83v7rAQfmsE37jLCS1PJPSWPsmTLvnnn4YGU","slot":83948309},{"blockTime":1625204142,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"47YbDyZAYPkZxLUmqZrdUgy5y8KEoVKQnnpxGVR8BsUsypYNuui1dSEN1gzRKMDYqP6AzU7FW81y4owH2RSJU9By","slot":83948308},{"blockTime":1625204134,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4eXAKiTnP7FoeFjr21vJuS4po4zTxuY1ZD8qqbuxZE5kPGbb2dKDeCir5H3udh3d9byqsUGotxvVmvpDURdKwWVJ","slot":83948294},{"blockTime":1625204125,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3UuFM7gt9wwC1DFuqobrSqbpZTetgUXf179iVWjCxbSzJVeS6vXXdrZSe8ioQT6gB19JM7L7nYeU2hAhyFSpdhjn","slot":83948280},{"blockTime":1625204125,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2CD33zWJbGcfipmmpXowuCrJHke38qf4bMFPkrx1Nu6awaWJHTpKYc8WGLZArKZx3Lg3w4BnMQykB9zdfXSjXjWP","slot":83948279},{"blockTime":1625204124,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"63rJNFhVMgjmEE1cZceNs1s8axwnpja6cA3Ztn2A9STJywDJYQxRv8kgkVkTAmkywEBSoKq4CNw1TAV1rLmDkX84","slot":83948277},{"blockTime":1625204112,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4PYMp4aMmXL1HN9VUEQotMaDXnhejfqpC6NHvTh5juWYCBMomKHjzRNuH5nCuBoKdaWxhYRwW4pLAZ3be93oZS4q","slot":83948257},{"blockTime":1625204112,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2CjJrYu6xYxoZvK141nHjL6aTNZJkf1V4yLKFfDt56rHNTrSP1Ka5viHHdzJXF6BMENXfYmGFg7KgRXLA3L4pDqK","slot":83948256},{"blockTime":1625204109,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"43N9HJ1FQtRV7K7MLA4WBd1F51EtUwx2PBkeLodK8zq4fjaL25Qt7e7GYdzuri5BDfnfQgqiESoMp61XbeLWhNSv","slot":83948252},{"blockTime":1625204106,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"46Liy6LYx7URZYHgMfZTDV7pddjHmCYVVcsfo7j2dkezwMsNBvvp47xtApGZaJHMxABKbyxAeRTxXmoVm67kKpT3","slot":83948247},{"blockTime":1625204101,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"8q5YCrg8B9gavpREBYBat4bZ5zxiajhYyDQpnsCkqecNBU42zPih6hSjcWwZZqSMUiC3LFvctMzie1dXrvXgQx2","slot":83948239},{"blockTime":1625204100,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2hgA6zU4PVqkUCe4DoVFhFqr7jJUmtJErqKZCKBiL6TmDEK5yyHPqaXAbzphwRKPE44taaKwu7dmTuS8FqKfVzmE","slot":83948236},{"blockTime":1625204097,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5JhPMXm3wv7kRP5i3meGbaD9TisDWkoGU1NX7ZCMZqrozNFEcdTH6br47ZzkLC32yEYyBa11CKbUnB9DPkCr1AAG","slot":83948231},{"blockTime":1625204094,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5eHySeXF3AHjNgvqSExzfiz8mTeSH1WG1HJrsqkFcyxSKw9JgYDJd7h49fbKFBujjdutn77ZYwGjrVMqyJqJ368Y","slot":83948226},{"blockTime":1625204092,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"43KNxgkivnDyJ9gdXKZQu2Rgton4xT3opMjRcXkFiUgTqY71eaVjs1EWutmRZ4w1rNkbY2WNZ3iNAYUZvMCYP6Ri","slot":83948224},{"blockTime":1625204088,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3tVqcwCMw8hGR3cScqtSs3wBHKt5jVnyg3kpHCcto1inkSrnGHNnVDyzV4mYkEZiyCGc8y4NKUhHbZQuU9jZvnfN","slot":83948216},{"blockTime":1625204073,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"EFwrfKvANGaRDKw2pkVrE5ZQUhnhtTWZMbRWeZsTmM5wVS4xyHfFpSUGGSaSyLBYWh1FD5M3mPPuU5fbb5rY1Pk","slot":83948191},{"blockTime":1625204070,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4WZsdnpHtFkqeecLKRZVJ5ChicEpckTLY1jmDT3oD6pAwxP4dCQ9MMdk2MWwMdf43iVp2pF8GjYq87zSTdEgy3eH","slot":83948188},{"blockTime":1625204067,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5aafWFyyg77Zr6jebpZiBkbgxbpEj7E2zZoeRB5tiqZJ6KWkQHELnQLrd9mrqhRYSecKX3hnz2XcNCByH2EUvdWF","slot":83948182},{"blockTime":1625203561,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4t9egLK4e44wHzB8UL6sdbsvxBqah8qc2mWiCzrpXDzK9bP8fKZMq1ZXb5VEegEYQAjNriptaBnFehh8LpY6Fg7F","slot":83947339},{"blockTime":1625203557,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3uYxbH6GqjtxQGyMFYs3iq4JbttVvn7cAEAfyWognf1qAGtCksmujU9DHF6s9Kc5tZk94tz7FhNxHCf3EM17Rr7o","slot":83947332},{"blockTime":1625203551,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4D1wTQpWXCEeYrGGfp52uexFDNFvsnVdUzGDLHdDJkSZoYbFbw5iTQQoZ8pWJhfXTQBAH4kaP66SMStEN3PBDDKC","slot":83947321},{"blockTime":1625203470,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"35NW2dQq2A652v6LcPKCzN8LK71gWuZzZUhZ5xuD84JPpjgp5vPbhWSFtbWqwze69BjxyQkB29ApGw3iwc9LsVsC","slot":83947188},{"blockTime":1625203470,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"pKkX1eqrcyYZNCNPqFTLbwcPEr7EmV6dmGoJErTQn9VRr1gPDyFWT3hS3NGrLf3ffndc8CuYtceFYYRhwSP1jRr","slot":83947188},{"blockTime":1625203464,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4jy4NpYtHpsuTVUndtxhtsUYBpGwMsQrCgxXgmKTJdrNnkKinmie5YbBVE2MTQzvaCPT7E4DDr6r77qxNTH1ah2D","slot":83947176},{"blockTime":1625203117,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"61uT6dHKReCKW5UXdXgMdTtmGv3BirTkArbRtJjxT7NCL7ARyFnrbxZEiqTVZMvS4wnChi2RK6kpyWfzvQtn2hQV","slot":83946600},{"blockTime":1625203111,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4X4dirrsM8EXoi6kMka9Ar6YYHNkVx391fmTDMvJm4BdhLW4LbW84LxtyZUnC4KFBod3VKsCQYwyCog5PmCuEsDe","slot":83946590},{"blockTime":1625203107,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"55qJVLQ94om4v7tpiocw7VZwLNHP7mW8AhKrnzEyJAurdUauwhVrSj3vRSSqqZJQqRLsUU84oTjBV8Kb7jvMEaoC","slot":83946581},{"blockTime":1625202459,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4WydXkKDYDydCKQEWo3SWmdZPZxTxXLBZJwLvDX62k9hfcrimWWEZyHmJcLa821mbjRPGwpWrPH7MAprfX3Pf576","slot":83945501},{"blockTime":1625202438,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4BfCgjiphrCJUUunAtPtJ7vJfXnFoadCAkA95sZiEjgSJrKDrxCfwcNuAizQCwYEqpPKWnLWn9jZ8PTasGgf9oDY","slot":83945467},{"blockTime":1625202436,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5FD3tYL2mQEDaKkosRwNdY3Dpk4L8NwjM8SQjdGKE3AgmbZT1ZsxPXg3pffwBwJWNepr7uT7sp1k7BygA8dEwP9e","slot":83945464},{"blockTime":1625201830,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"37LkeWVyGi4Eo87UmmnuqBE6XpHa1pkkAmnmG9acP2Xg8Kdixsz9TckVPFa7GxsHCoXkiUG63RwfPQjAkCd4UA3Z","slot":83944455},{"blockTime":1625201811,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5612tDdS11BXKBcmLNAWsYZg3zcktyU8yitEjPyPU8zw6KWMabwSQx5TGv8Nur4wwXUt8nZebw7rNqgPJZSd4m9i","slot":83944423},{"blockTime":1625201613,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5ce8RmDfmStxJW6tSicuN2GE87FnAkzpEySohqpBDecZRdhadx2yVLvV8fwczetYemNEioTJmp62eukgpAaJDbJU","slot":83944092},{"blockTime":1625200045,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"51LKeeSa71XBRwxGgarTfmUWdRGxaHbKvp2edYPWcchKpkCXF3dKPM5CYvK9CJ3fGdqXF59v8VQmaFiJ6HgJWipa","slot":83941480},{"blockTime":1625200020,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"z1MQDQFUQss9XGzf2t9cwnd6KkyAhTTXf15mpWkN1SzNnKS5eppZpbM7Vfzx9jmhmarBSopC3YWtCv76LkJhPse","slot":83941437},{"blockTime":1625199859,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3We6Av28Tq37bNv628tcAwuGwfLSnAC9JqWAmRsgibwtD5KSwBdFjEndFXyZ3fUPnQ8neVwx81w5CcQMccbz19Hq","slot":83941170},{"blockTime":1625198449,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4U7PfzPnMN6XA1jkMGaEJWfecc8xmL4c8KsoSBY7csvLBKxgxX1wHHG57idBpRB2hqUnv1xJ6WEaLMfStMFSDXmi","slot":83938819},{"blockTime":1625198433,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"tAPEBdxdwDs9k386EMmCDAktMrJttDQTsWbxTAUdzBFd8a8kNLzQkBKqhrGfhio7Lui2UgCNHNen6mTMb4GmHor","slot":83938791},{"blockTime":1625198398,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5mr6uoHQH1aAo1LoTkdYDL4zLmemrVRF2iDfYTLwB6zVbPDW9Vg5DKCMoL7XUzhYqzJggZFtnZWJnenb3WeAY2Eg","slot":83938735},{"blockTime":1625198199,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2v4BT3jffzdhs5mYmpscwspHZCM381Z4MwCdit5MAvgtCt6ok5F71mVWd5qPcvirevT4onUbhRESWuYmVgY7gj7h","slot":83938402},{"blockTime":1625198019,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"aMK6ERjHfYCs9cZKWf8C42T5WjD3twbJQGFaLP7sYzLG57agFEipe33qDfLxGUpvMuFhnZr66kLHv9iAwJQ4UD2","slot":83938101},{"blockTime":1625198013,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"hdu9ZBfHfDVf4pEi4fcz5amRLNzgEafd7YrmQ52j3w5Y1J8kkyj8HZyWp8c2TmA1CzVAkDaCqg7UbfmfZVzYAX9","slot":83938093},{"blockTime":1625198010,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"YUK4DwCazm4T7JuhtUbESUA2xxN4ryDqQ3Ki645eHJ3p7Q4gUuji737nFicqm2dProhyrgp3iQdXXdfYNbeWVYx","slot":83938088},{"blockTime":1625197885,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3XBjm16RTHamQtbvrNKPVRxXseCHssVipSfDr48SRqvTYy14bMnWyV1pz82dbbFDe6De6N624DK43kqq5vN84CKb","slot":83937880},{"blockTime":1625197299,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4siJqk8h1o5fsTDR29j2Dp8PKTUAgr9ho5CZvjxbzknwWnfzjhnAr838iW9tFgEXfrU6Bm9N7fE2EgR8xcSjtxDT","slot":83936901},{"blockTime":1625197299,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"VwoMK4rfoMqwU7crFQL21oxN5AdrMsGFYcYbjwE79ckxBLuyJgBiS87f9JVgwKAofsX63pVca6u4w8HEEibB2Hw","slot":83936901},{"blockTime":1625197290,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3cWwz5to55oZgy7RYN7PqqiA1yrEnFmRi263kbZveLQ2D7AU4AKhbgyMHqNYSDGuSXZCoSRm91F81QVtpK8svNB8","slot":83936888},{"blockTime":1625197204,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"63DQsgcpVPz4xpWckC8jw2nYtKrPSDFfJtESJuAqrjs9h8Tn2C2sXfWc55qzMEgvkB16pDo1sNy6ChDtfABPzD3S","slot":83936745},{"blockTime":1625197204,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3vUCwDSdtHKb3rTYVpGRYuA1Fs7CEgFSBfQwk5UzBeAmvM11hzKDJx36uLWxq5bs1tsPrwaABLYTxTFVLBW9hyUj","slot":83936745},{"blockTime":1625197204,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Kazmzt82Vzz9YfbgnJwHfdUnopRjgcWaFFxgG3obPZz3sk526xZNJbtFrftPnZe7suPbgDLyxVZF4EoJxeaKdjs","slot":83936744},{"blockTime":1625196654,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2SPudDwu2vyzVcpRrngANWWqsG2ar2VMxJmmbJFPumkk2o96c1Hns8yBoHs1PywxfkXcY8VJ4BzD5XGX1GcUqKjn","slot":83935827},{"blockTime":1625196652,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5ccpqsA55wM2tbARpgKWmuenWad4GSbJwwHzYptnzJKXkzbDEKduHSQ8KGGoJaJNTTY5womT43dLzt7bsx7ZPJHR","slot":83935824},{"blockTime":1625196652,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"21uzYNCFQTJtZpeGV3a5FxNU4DFuxrE514rRkMimYziix3HotvtiH8EdZqpZqLwgtoWpnuhtJkwozi6wTNDszN6b","slot":83935824},{"blockTime":1625196184,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5HhetzNvxZBVJRRL9xGpdZoJ7NnR74EmNWBhUqkC19q8yNYJNewnDZ36LvsGdpqFkFVuhBAmE4VXL7dbVKABeKqR","slot":83935044},{"blockTime":1625196184,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4bnmGmBmagY4mHLkscofoTgmdf4PnaYp2pvpYjb8uFkCR42BCgpAtPAGNLjW1wZF7tHCc4m4UUU1ruxHuHNUJAi7","slot":83935044},{"blockTime":1625196184,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4bg4JAYHkv4WFjWGk5XronEgnFUpFpW4tDFTcc6BFHdZ8qMdGUr9BNiUKWSYtm5TComv9RNDXct5XoEiVMimDMzB","slot":83935044},{"blockTime":1625196091,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"32kxGH4iXQUVEUZGE1zyJPRKnY686HZW1B7oeGS6STbDqaGmN4Fa4pHxPU6QkYAUUXdFJBwB4pnRuiR8jWGVLXBz","slot":83934889},{"blockTime":1625196090,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2rpG3qnLKLgRHhXxYgCBGJjWrjs3dUNQtmfsNFZBoZwr7ykhoZ2NxLaPpJBbdd9anPEJ7xsUaxWkG23zqksAB4oP","slot":83934886},{"blockTime":1625196087,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4vsPV5Vdsw96kGMToHqgpyZ6kX2eLTBJoEjaiS7XVGAHgmrSvbBSQ6tPQvYpxLfEHyS5VrrSnmSojxUVL3EGcf1h","slot":83934881},{"blockTime":1625195973,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5SpUHcpyoFubmroLpLhUPyqL7Jux4QGiAGD1hwDyg361N3qBAu3RL9wX45ZegfgatYewSqaopUVmDdMBctxBkmNP","slot":83934691},{"blockTime":1625195925,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"8kAYA3f4BDwBgFgUu9Z3ysfHgAeDyge2DphqP9Tjk1biujUq7bgo8mRsS1X7ckqEzGX43ngDxmsCgzYZj5ygzG2","slot":83934613},{"blockTime":1625195467,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5Ke2rKtyr9cryhc2NuJZVa7YM2GFaMq9rEXvzGE8JGCWBK55ajzYT4muxgJ2MHbGxiQ9asBNtYxxgrhtYNAEGTya","slot":83933850},{"blockTime":1625195337,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5qBHq98GTjoDrEkRgbUTi31BKYLKYd6M573bnnKUGZFEoxuYekPhcnivbsvBbFtAd4xvSmzcE6pzeWJBTDu2gdQS","slot":83933633},{"blockTime":1625195331,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"GYj4XPDZAV7DJkjGV8QfpLND6uUvPKPuXJneoo2sYgb2XUJTRv4TkZRRLuqM24Zbo7jk6oZXFrBHGfU3cwUwyaD","slot":83933623},{"blockTime":1625195329,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"TQ9Y9x3BWBTCZVnsAxZWm8xNxxaZ7Bp1q9mfU277KPDwirjF1QR7byThUwj8sCJbCcU3CxY6uvPQmmKnBzdfbVM","slot":83933620},{"blockTime":1625195163,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Jqd73haKM2FPzCaxpvZGDxs4gnCHsUiGiRSG9vk35xyiCxBmRZcyQeJtYQN5cx8iTTfs7Q1RgCxCy1Fonm5QQ6E","slot":83933341},{"blockTime":1625195076,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"61NQAyjUiw9YrEbfA3STiNPz1zmMp9V9NNepJBDxs5xAP5v1aP3Gm5HJmy9A5jP7vDEq8P2imRuvKSW6M7N4qQGG","slot":83933197},{"blockTime":1625195076,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"r775jz7vWcfLY5nMP4BEfPqBbpMRh3Ch3gnwGQfBssw6oayWeEaSBQ6mobh7JNLLXGorpyEtSvrbfAZFVYEWG6B","slot":83933197},{"blockTime":1625195068,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4FMaBb3fcb2VbRV3XJVuPHhxWJ5RR4X1bdJQbJiWuBsJNat2KRQeP3pihXk5X8PGVRS3oGZbU2Ws46jn4x9vau5R","slot":83933184},{"blockTime":1625194951,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5arWtk11DwJbfM7XbtBbhsvTWtJdapiGZqHqap7cCpPFQLoLqCZiyFNxFB6sXYgeow6bTJs1Sdp4x2xueALCbPi4","slot":83932989},{"blockTime":1625194899,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"58aEFW2HufRWKaKvnMNMhrCmeYAub8AQ3eoTrXFS2VpMDfSnYBPJta74n6HMuYyvhLLJthQ9ckhNYEDjFH1NSbER","slot":83932903},{"blockTime":1625194845,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4MFR13tzHee9maXRBKWbKseiu7G8tPSK7Md3BPUWSze1mFRKJ175cF3H9BwCo9qTX3D6gkaMBDyuQWoMKwbi6b6B","slot":83932812},{"blockTime":1625194434,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5uVzQ8MzsC914fMznZ1qHgTtb8WC6ug3gFfXm3eDbvx7HPBnWALHU7TsrU9KaVGixHntBbD2TcYsUyUHvjai8jNN","slot":83932128},{"blockTime":1625194374,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3M3T3jxc5Ht6TH9Xk1M33AUbzDWWggy81KbSWNArUjVNbyrwHGbDEBN7b6RXiYLqSrVLgmWvPXDt3Eje9mNKJzEp","slot":83932028},{"blockTime":1625193526,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"YDmBK1UcPfDUX8JCcA4tiG7qBY28SMfbVcyU33Ak7A2gKJ7y332xcj2nUg9Cascfb1QFDMJ2kSmPyWSQq7Wgnst","slot":83930614},{"blockTime":1625193462,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5QUULUgTRECPcb8UNUbEsGiLdFj5VbsKnsLrfiwRLf32aArYd5vdqiCcVQi6WrHXYGoNoMrJV9ghuJ8Fk8i8f8fh","slot":83930507},{"blockTime":1625192206,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3eN3JNDhnJASRs3LXqXiidW7aa4QX5eAGMNfvkXiQQXGMKPpNo9q6u4WGjvuetZdDqsXdr7o76cjp9M8cZqRRaxs","slot":83928414},{"blockTime":1625192080,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2HKf1Zhbr3hd7GSdaNeBvSDPD4jNrQ2ysr8JLnVaRtzkrbv9pkHwX9mJx1fu5nA3DXdHahBxUtfTva59X9hysph8","slot":83928204},{"blockTime":1625191656,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5HeD8mT8AoQsVCmqNgsJeqZLrJ5zfmEdcEDV4AbaDtNQw1FfX6hSRZ2gxyu1WjuFnLWnxpsWEG5qMvUU8F2pResk","slot":83927496},{"blockTime":1625191642,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3DDexgum5WLTxNQdSVAPqpH723N21VT6QbZcYgPFjRrLtNiQqs6vUAsqHhH6abfg8StX8vaNfb3okKfie4pxohAU","slot":83927474},{"blockTime":1625191626,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4QYYgzcbEGd7ZkjJHfSLsnHESbemQdYmpMxUYbW9BCGH2wmYj4SvZpUBsWHXHtDhhDe18LVdLymjdFQCuqXshmSD","slot":83927447},{"blockTime":1625191158,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2az4bnaB45E8Q3Yrji7mdWvyQX1obZ9gpR26kZzVsPaWLegLUyD2xeD1RCgG8BugrUokNFRsCvgTi37yEVkSbdfc","slot":83926668},{"blockTime":1625191113,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Gd3pU9qhrq4HXaef9vERn65tYGSiGFMKQ9yohSww1tABPDd2Xe8ABiZtsJeNHwjEh7e2XK9bQ5ksaFW2BPi85cV","slot":83926591},{"blockTime":1625191110,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"56GU9ff5N368KDL9t1Y1FfEiXv1nNgCtsDaZaJpU2ufo2wHDciKZqZhVSrb2jUdwYBubJYqddyfxvP3KLBR3fPuT","slot":83926588},{"blockTime":1625191086,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"J1xXUMyqyVTdwkZmy4orZzk5f3gJ3zi2rVmZJWz29J2uw9h23p5w6TEqeGTPZFs69FXSnZtYMvFovtpXm7SMQFt","slot":83926547},{"blockTime":1625190784,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5TL1Q6ZTr1AZk5iFhkjsGA9jGFaQQaGQcn7gCtm7T97WM9xG4Lw4Y6X85Ta6PypLDoK6H6gx42WMZhrpXRGpw2Pn","slot":83926044},{"blockTime":1625190777,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4dQmhtEY7rRE14QLxhVc9jr7o4wUobZTN8gZssApfEgHXJmFea4fUDNbf7yCsP5saY3pFBEL2uLt2uHDrjY1ciZi","slot":83926032},{"blockTime":1625190777,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"qF59GT8SwMrWMSvK9bEE6pE54TFtmsPYus3KeyB3x88Rz5rWwhSs9CyLS7quCaX2JowqmtVpoXmJpk5XT71G49y","slot":83926032},{"blockTime":1625190771,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4iLSkgM734UiCG2uXEUFDp48KVeBY9XRY9B3iM5NdFGyyZe5J5DuRb4CnhTEruhKM8MkuiYCLwR9bnhxUsspN4vX","slot":83926023},{"blockTime":1625190768,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"9euhpAEnch2d9KsPjZsfTP3BPn9m1JwGkY4TAzrb4Ap7C4HBnb66rpeNGtTAJ8wbTvNE5FH3P8e23yXEkPY1e54","slot":83926016},{"blockTime":1625190754,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"66gWju5oPa4ovKuMkuFzaDrNzTwxJTZYLXmfveFYsqyRR7iFoUts7LgzQSMhbQaA6rGDxdNBqBL4gmy1iU8xWvcD","slot":83925995},{"blockTime":1625190747,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2EepsMpaj58XQkoSXoc9k3QnZNyFujHsCZTfBJaYzqk8L3FxkzTtqcgmPXT9ocSLLiPMcS32ftKHHL9rkFe9JLJu","slot":83925983},{"blockTime":1625190735,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"22PsToyMp7efQG2CXY8bHowmeVnYkW8mbE91RjYzXoJT1YeTXmp4CQrrUyVGLxDPE7N7jMKXp4y3LThsLuwznrFU","slot":83925961},{"blockTime":1625190732,"confirmationStatus":"finalized","err":{"InstructionError":[0,{"Custom":0}]},"memo":null,"signature":"2SXapfWu5DZSRZVBNyK9pR48NpMdEr5ecsXexHnpycvJSeN2nG22uFujUnPHyP78SD6FXFgDqTCaDP7wcazsAgMa","slot":83925957},{"blockTime":1625190726,"confirmationStatus":"finalized","err":{"InstructionError":[0,{"Custom":0}]},"memo":null,"signature":"4ZevW1kssABvVt3n78FvphHGeG9Wggb3j8866HqG8qmfXfUujKsivJGdRmUXPMbvqEhiYz3R4bWjHGBmGGXDxeGJ","slot":83925948},{"blockTime":1625190717,"confirmationStatus":"finalized","err":{"InstructionError":[0,{"Custom":0}]},"memo":null,"signature":"3sTDJeHzEQGUPgbE4AxNeW5cNUodWnNKRd7HgH2oW1avigBAdUw9hVBWUN3QH8KcaKd3TBijkBmdHwqjjJHZxqQt","slot":83925931},{"blockTime":1625190715,"confirmationStatus":"finalized","err":{"InstructionError":[0,{"Custom":0}]},"memo":null,"signature":"2tD7on9UWL6BYx4QFfoWfJDdptiLTti3zVisPkUzrzQdxL2QCQr52h1rbc5U6G189tyXtuR8wbhqX8v9q1NDKYEC","slot":83925929},{"blockTime":1625190715,"confirmationStatus":"finalized","err":{"InstructionError":[0,{"Custom":0}]},"memo":null,"signature":"rhK1nkGuFdx3SXomKwi66Nyf8K7Jqim81L6ADdUfyqJC6RGHXgn87saxp7XcXGBEnLTznsZTMRodDduELSDn1LH","slot":83925929},{"blockTime":1625190705,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4mRnXB64Fws4jFaagx95Vg5KYQESEcaPPumJg1UKhP1hSs7VzqcLPU8ZfY4BFPpsU6jJup5TWodErCXadwZ5CXwS","slot":83925912},{"blockTime":1625190649,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5iv9NGa8GMpDpgMwBeKAJXsGSa4Lart2HP7PWcPrHHzJLJZuRc5axACZimASqNqsUx6oNft1rRYzAoi73F8yQ7Yq","slot":83925820},{"blockTime":1625184906,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"zPwqGXQ1WdYCrRExA3r6s4hNMtaskJC71DvwT5K3CHagXqgJreLyKF4rChJeqVmbpfAnCY4LZQdDFEZQ8hsx2JP","slot":83916248},{"blockTime":1625184876,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4q91pekdzgDoxd3DSxvejs7fhhm8Z6QHxa2Tufcgfayy2EkxjktfNmuk1ZghXc9dCvufo7KZGwkiPwAJRVa1rZ26","slot":83916196},{"blockTime":1625184859,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4npxr1UcFQSQD8CKp1Nmjx1maGYWLuyFqqEp3dhhxHmZkDM1rmDYTGyxD4922tg8zGUSK41jUnm1Qvj5sddqLzeg","slot":83916170},{"blockTime":1625178234,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"ArnWnAigVij7UAnrTkaUannWNQDbgjoDEB1ovUEAj86nfyDiUHibbbMn1rCtaSFmxsUrdCjYF5oAuYVu98NMkAZ","slot":83905128},{"blockTime":1625178220,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5LvmzEqSERcyosvfugzuEp9K51PE2rjmG6sREpbn9qgcAisHzAgNzyZf1xQJXAahCSnLPoAn5TopS98AHXBFej8g","slot":83905104},{"blockTime":1625178175,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"L3dFPntNEQFnwHA3uQG7pGGuTthxrakexY7dzXztEfhxQ6u3XWsUq76egUaob4P42qJZcF62wLDUzMYogsbrqoN","slot":83905029},{"blockTime":1625173068,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"49gJWqytrqURFVKKp7XgH5ZamXUVMtBCcSKumVme1xwTkrjrYSFTyaTf4abVV4Kh4poWAt87QTo2shJSYyWwELZQ","slot":83896517},{"blockTime":1625173051,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"46hajvdjZRuhcXDkTmXAE47zoSG8cAgdtWen2kTbgZ4rzMyu6MGFbCgUhVf92gHfuV9KRvmmYF9rWMxw8SreVd9K","slot":83896490},{"blockTime":1625171497,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"DUxhxKAvcvtxncPMB9J2r32Yg59es8t9d4RZJtVd13iu4MASjJ9RyY1nW1tK7m65Le25tEBh1wyvJc8KyemoN5W","slot":83893900},{"blockTime":1625171461,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"SWujFNsW3pKbVL2V2kvBshfJPwVALWD2JYmzqX3hEzQvXZsZ1ZQ7vMHvBUuo1MVP85aaX24P4Z6jy7Jj723xb2k","slot":83893840},{"blockTime":1625171457,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3qLDK8H8cnCZSnQNHQoMxve3PuNaExpWCSTURSxBAJA9pZE3jB8KLnFJsc6pNSgYdTUuHgd6weyHM35pK4PidCbT","slot":83893832},{"blockTime":1625169264,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"551s4YqVXP8Vq8UzDkoXRS2vXQJYC97BwJsKiTXdjejh1NtEpuNVQsUg7vNUZCnbetaa6Wn6LdJHhVVeecqAnXqR","slot":83890176},{"blockTime":1625165809,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Nctjdqz98WbPwUNkiPBYnmQYFeNfFs2h6AiVKfWKeu14iomPL3gq8S5yybNt6y8834nqfxnPpPA4YAHDrxpJRoB","slot":83884419},{"blockTime":1625165785,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3BpqbQbvDuyZ4CadA9FhQmqajXtdYgfoF6Dt2WSHkmXj3D4hUF9S5iESUoyC5Pa5v5jx3wLyPzA1YFMW24GsLv4u","slot":83884380},{"blockTime":1625165754,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"249XuKhQhwegGDnP6F5xyAJHp24zjQnRKQXTu52pfUDRpL1dZChkBGVUe8DvUVLhHSbSL4EYQpcL3dCExiJQW6Xn","slot":83884328},{"blockTime":1625165325,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5UAANpLLkcTH6cyJc2UDddQFdn7cB6R7EH4vkoyPXCTJrpE84ngYoYx6jKcqByptdvWSaioRtXZvcUMdKa22ZbwS","slot":83883611},{"blockTime":1625165317,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5LLBikpbeVHtwiLsj3RJd1jUuGaTY9Bc16qoc5GjDsT8W5jMjT2yHGZtvozSmWLNEGaoYy34D4VZqgFnRh55wF4q","slot":83883599},{"blockTime":1625165301,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"o3mWtJuEmgAM21mLCvJ1sfuPkdPaKztyN2rR98hUczeuLmNDmvS7a1rHP9csv2kv9YhcFCG9pZBA79Ss832bGRU","slot":83883572},{"blockTime":1625165289,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2EokkZkWUMhHw826JkMdbtm8d62uwxoyzf9vETK4h4YkSXkCL6z7zF2Kx33dbr6TtW1jwbdthBa2fdBzkPjrJtT8","slot":83883551},{"blockTime":1625165007,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"35E2NF9tXaQdEJXABXBqa12feAU3cyeqjZ4oxdiXFcMCr7sYzBh3snjHmr3zYAhn81P4ZFTiW9PphnEqmwHwA2tL","slot":83883081},{"blockTime":1625165004,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"ccQhN6DpJUcXFnRQcNjnfuyb9JCBA2wpB71n5dWUPmngJBmao8JB6mFX9RvLbcRWRi5WXkVfVyr8GKLUG8Y3CMT","slot":83883077},{"blockTime":1625163103,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2JTbwbBgSjLxYLk9GHwSTk9zmBWtk8x1E4eqw2sB8B36gYKH4RHZhVA3R2fhA6caXjGKdm6gm5mjaKztGqmsFsp6","slot":83879909},{"blockTime":1625159988,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Rq3to6ycetWStzZTQJLRF3QvEzb93r85bbKc6NTEZBgpfz5YmiEUKzyCYHaSFAwc72AffWaDkYaQNjFWU532wVA","slot":83874718},{"blockTime":1625159838,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"rqeiciNVS6V4pGibzh3eYK8U7eJ5kDfZ5QDGyLrGHdLWZCLZD9oK3WmFezZ8fGSL5MwBH15B6x58ro8TaE25cPB","slot":83874466},{"blockTime":1625159376,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Y1bHo1xSZHEYXfJANogaWwWJf5xpN8Qbd1rxqEWfCD6oHnfYAcAZbRapjcNqMoW934VKNaDvjaLsBABXdznn5cJ","slot":83873697},{"blockTime":1625159376,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3zAH3SMJcAd1mLRVzmM5M1AuyhyxMbm9WaCM1eD74fBGVjiVoYAsykwTGEe4bJXJUjUxZBJzwLjoVesYUqB1khXY","slot":83873696},{"blockTime":1625155117,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Q6WH92UqCkwADw6r88mNKQ2KDWLDK2xUPEcwLwAJxyBwuEwg6p1ujMWjGF4ceWVC7mQ5z9JmUBTJ4bUdmBvUA9r","slot":83866600},{"blockTime":1625151457,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3ACFd6GQXzJ8hEEULusVr2UAzdy1zXRydDjbarYaFTdp4JN3VDpq4oJKJThEApEzLBTyJv8Pk5VEFnguWNFbx5FF","slot":83860500},{"blockTime":1625151193,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3K41kyPxtTGsRzZAjjmvgExZecSSs1A1mRqnDyVHSDWw5HAVtptX8iDGds5wp651avdLpW6ZMjvmQ9dQCaUpL7am","slot":83860060},{"blockTime":1625151124,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3wqgQ44RgJSojZvN7YZXKRWpJ61pm7vpCkWftZZPC4jMc3idqmNLdy1A1kRbbiBGZvhAZr7np9fdTMJcGYf58sU","slot":83859945},{"blockTime":1625148349,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3QAmAcmZQXMK5wY3kiEg8hXS5LGKX95U3GKtwZh8uku4tqvXWmWiyF9C5k6Cg7ciBUVkLz44h1Ns18gRW4zkV4av","slot":83855319},{"blockTime":1625147988,"confirmationStatus":"finalized","err":{"InstructionError":[3,{"Custom":16}]},"memo":null,"signature":"65efMaruKikKUwBjWJ41DMUShk6iKro49WkeP8eEKPBE1ZtVMdC79KX6zv2J8Qw8kfCheohA3hLYNSihcK51qzEE","slot":83854717},{"blockTime":1625147881,"confirmationStatus":"finalized","err":{"InstructionError":[3,{"Custom":16}]},"memo":null,"signature":"3MVMxHJp2NgSeZxu3ifWHkFsZGzt13zQNtRraqyiWmjtGiUt9UBbtmsLCAh1WcvPwmcs7EpsbWkLrw3J6wJwv8Wq","slot":83854540},{"blockTime":1625147860,"confirmationStatus":"finalized","err":{"InstructionError":[3,{"Custom":16}]},"memo":null,"signature":"5Gcx4N7A4ZdbhsFAQEZDyMEZuej5Tv4MMGD44m5nyXcLEwtiRyY7rR7H5rt9X7eCQ5F97ePQLK6gZSrBtMFQmAYh","slot":83854504},{"blockTime":1625147842,"confirmationStatus":"finalized","err":{"InstructionError":[3,{"Custom":16}]},"memo":null,"signature":"NJZ4k85pqWCpBZsp5kcvcwza9jSF1QQEEyxZ4eyJecQ69cVvLZG11GF677nmRYjW5DNqhowyPnxVQVMXvb75hkr","slot":83854474},{"blockTime":1625147799,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"NjSYmqrFwqCSMN8oyRUqVSTrB5nb65yvsDxSqkrUkiyeve52cBxx8PVKd6P65A7khMzBxEzKKBz13wLwebXyhQR","slot":83854401},{"blockTime":1625147755,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2LGazPnMMGVNrVLXUwQPoKMhoSCD1fTuh7pssnUnMCXbjwaMcmmxEGec2QSUjw7CHaXoNuXohw1ySjme1QWpPN9A","slot":83854329},{"blockTime":1625147707,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"TwEgBNtE96BxwFHgpmeayREXNkLJyViZGepEafPMV4vo715xWzdMuW6Dmcpnnj4NnBDYr9feBP5sc7bGHRCG1GS","slot":83854249},{"blockTime":1625147688,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3datmJvRtwzw3NGFmf5oY6Me4rAbaHXHtnX9pWQWrma6nFX4jGXLXCv29aR6DrVzoTpgQhNd6JkAtx1h2Go6M6XK","slot":83854216},{"blockTime":1625147676,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4s7y8vr7wWXZugNtaX9rEtQRXTyNNyai1UUuHqQvcGtsdgYxuZs7dFspAG3U3gYnSvYyToTCqxKgiJjyYynRgzpi","slot":83854196},{"blockTime":1625147671,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5eQiidozAqSqctwdjUPRb76zHxXdVcbfhXfFCJKwkEFii7rzc95Mc1CUg4hvtGjQBjUNie1A8UfD8mTHiwpfziab","slot":83854189},{"blockTime":1625147649,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5wmeGzEDqr3NaJ5J3HGsmGpDtbG5ntR29C4ZAQDBAbWewMzKpz5NXhrJk2kHEPn582jeQ88ssxXcPcZRatEfiEYB","slot":83854152},{"blockTime":1625147622,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"kcod9wstpp1pzT9PdvpKy3tjhBAYAV3sh56ueGtLEFkrpidCakWEheHHsaxPuEapNUrkm4YCwog1yH99NXWjtZz","slot":83854108},{"blockTime":1625147611,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4ox75LtFjeWKxgVaia2g5KqWVzjX4US4jZpV9tH2PLSrJcCMQJGwTrF3W6rakirGLtpERyBWuuFZ6wAH9C2FUjWL","slot":83854089},{"blockTime":1625146687,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"7dK9zjdi38HXCjwpMfspqedkwgZM3262w8cwbfjQ9h1pdewjePpmztCaUGWW65bVAKSjJ5EcuqsN9XYiLsHTFRE","slot":83852549},{"blockTime":1625146659,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"LX2iE9qwJaobFErvd1fKnwYCzNrG7Axhppp6W9ydsBKjKvEfdf4YSUBXVJvioNoAzKBXBetU6amcCT3zpPP6b1V","slot":83852502},{"blockTime":1625146284,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"YFzyHhojXL8Fu59FpdXsWkBYn99bXVHcpiDryxKiNnjqyFbU8WbtrBq3hsZBBJXpKDdXHJ1rRXPY343ik7tjKSU","slot":83851876},{"blockTime":1625146275,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"mgfUfRi7DRs5UvJMLq1qWNqiWqFz8YvMZNNZeouS4vvebgKND1uz1e6dGJ6hJgrsWkdUDL4gWFrRckZSRgu3fhj","slot":83851861},{"blockTime":1625146257,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4RMTvXAS65GUrNhZd1eXwyNYKgZP8fbCdEYVNCQ8KB7zETEFUcCr2N2R7RHtZ3Rhd5HRjKAExXzA7xJFT46HktB5","slot":83851832},{"blockTime":1625144371,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"DXo1rWoAjo1etiKzecHW7m9fUfgF38cgHhNQQEbE77aTYB2Rdk81FRXhm8K1u8yU5KK39MWD32Y44YLDsSotRMk","slot":83848689},{"blockTime":1625144308,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3jxiCjn32Y1TJUf2r1UgYTSwSvMbXaLF4tz97t7Y5dLHwSnRbmu32kWvjgJuHHSCuooyHXP6q9Ja6Pp5f8q7uTvc","slot":83848585},{"blockTime":1625144290,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5FY2P1jJCPjMEY6J4CcJSjfdBiDtJntfLmBM94ekLbZns5kPQwNab2URzsFU8S46EtUmjKtNGSQRPjgRWd5Z7EKx","slot":83848554},{"blockTime":1625139489,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"63rcGDkKPnSdfYAv7KvVpKfmgMkGjSi6RiiJRQAigC9kWwE3HPEkRLBS2UTQAbzVmAZmFmntFLf21drNBroS3ffn","slot":83840552},{"blockTime":1625137267,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"8mi9hvpnVXfV6PcJYitiEFQCT38A1sPvQguVJQRtKGJNkHZn2YWqU14ngvYsVmVdzpjFZMVWibwXSd4nVPRb9fz","slot":83836849},{"blockTime":1625137240,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3kcam3LeBq4ZHvTKZaStJ5mdWUKJkQESFeVejw7c97EfUQRLJamqbTm8acwcmXhoFvdNfVtM7mxS863LW51DjyV5","slot":83836804},{"blockTime":1625137200,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5wGQY6URx34tw7TS6oLP5sB2AJQDvTNVLgVkd6RkYnYPCUvdW1VJSnspfeSy3ZBnxkz1Zkozo2ubnBDEo5vXze54","slot":83836736},{"blockTime":1625137090,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"392DLjPfqdeFvAPD9BmFzTmcQ4uYYEDJDNVe2uBDoDdXbKfHbhFg9V7ufm9pLcWGPr2WBUf2MPRpMxSMFp1WzP35","slot":83836554},{"blockTime":1625137068,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2RnMj87PCGtSM7oFK9yi3X5yPMx7LQdd85Et5g8vSwJGJjJzoJHW2Ls94zdBkJrtu7YBck9SykJcoXwUiAv1hC77","slot":83836516},{"blockTime":1625136415,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5k5yuxgSPmMZq7wZNY8tt6ENM287w2ZkspCJwNwU88fXwvHyjs6b3Pr7HVWSWE3m39SMwkEPRffkr2WoYpU4Qgaq","slot":83835429},{"blockTime":1625135226,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2a5tfuiRB4oqcD1gXFbT3o6ejLiTyFpTwZeNLkf51FV7qaTZg72Vvu4XmggJHqATm4gsVXVXJpRLTER8NcavcG6d","slot":83833448},{"blockTime":1625135217,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4miVGeMcdkWbpKGyeHmzCLGWQZo7iRDvWkQah3METMTRksuE13E23ueS8qDEUiesfJFNPtfVBFcRE9EMBTEJBTYW","slot":83833432},{"blockTime":1625135209,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5PuUXiAy8Jy7DZrECTqefB4pbgjtvYhmPefNbK6RydWD3Ng1G5ga6skYMiYZuT1aKtttCcwifvRVAVxD86L1BwUW","slot":83833420},{"blockTime":1625135202,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"63AiTHZWTTT5pi29QhAn7xNroi9MEMbwt6YUxHi796UA7rV22kCh14kYx3a26gPun1HHe7X8ASyoLY9ij8b9yC6V","slot":83833406},{"blockTime":1625135193,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5ZaJxbPKPY1bkMKnNY4JYaCrGm48AbKHebwhi35KEJtbpopHif7kAZz5q4pftCg2xSqnx5PSDwrd9FKeNs2uBZTu","slot":83833392},{"blockTime":1625135187,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4dgS9BkGuT9b3CctATCtgyVSPrCEicmeLGemitQybZaZmPFWkrC7N4vUqnmLFoRA3dCmiMKv2b2wVYoEYWB1CafN","slot":83833381},{"blockTime":1625135181,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"21hZuoUvz9Nx3Jp6W3TW9oLG1R9TfmYy3yn3UjTgs9z2UAGiR6LJT26xKmYdCeULU1qaSEBF9aYLeRFvcKXTiS49","slot":83833372},{"blockTime":1625135176,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2QicQoBpQQNWs67S8kxAiESaZ8PdGm9arnqQWbb5RPRLBTJkampshmUKyPUdxgfD3LXEzoqSH1WqnUXT6jod7k5d","slot":83833364},{"blockTime":1625135167,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2cfoZS7jnM5E3rpD82aUXhfpcBC1pBhZwtJzveY5nSug75x1SBoupWLSU9HyvRkXKYECSce5TF65c1qc6cWEJDxE","slot":83833349},{"blockTime":1625135154,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"UApfigmN4SeHdMYqn1GcoQeq875fLh5u25wjQY4WXyXanYYMv3gQzmgG4AoBKR6uJYkT1B4QVcvE9a7VzqbJ2bo","slot":83833328},{"blockTime":1625135148,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"XvXyHmqcy2RUxJiMCpg8YRhzwZjHU48GLuJzYGPQ4SAi4aS4H6U3oniPEUqcTTt1dPoh68Apc6xFA8Dzs6Lcwf8","slot":83833317},{"blockTime":1625135136,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Dk6Armh6zaERYnZCgxc3o3xLNsZRTvuEDTHAkB66YsSMYKVcEhghjFKA3qxsEgwQjW8GAwAfsZ8zLuoGB7PSWDx","slot":83833298},{"blockTime":1625135130,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5ZqtC61dfmr3FQUgW1JJtmgnLb388bHYw9dkt1SyLF6Nbi8sejmAM84BSTYLFWq84UzQG4WuaeEvExnxSNziJHAn","slot":83833288},{"blockTime":1625135121,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3ChiKqhcEh5SJ3usXXE9yV1iDSZmUCC5AyLCoraB96zEt57knRtiFKgTPPiKT8MZo7AMjmz8ahu2BhcHeG6JAqVS","slot":83833272},{"blockTime":1625135107,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4usGTvtkM82xLnKWUJGxxPBAR65AeJHVhq2dDExdzdRQZRKaCqtk2vt3sv3GPXzReN1JhKrgBquZAW9pd3ydkuZ4","slot":83833250},{"blockTime":1625135101,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5VFXFBBFFgpsJyoU5fmhkpCkeqmWXzPmnMkdN3NSg78yHidMR6HUo3jgbMDX3ivau2aLKkTDNufsGmjf9REsHvuK","slot":83833240},{"blockTime":1625135092,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5vQa9c4ci6by4Q9pfiAfCh2xURNjoR8CWFCFNeYN2R57KT97xFDobfJD2RzxXYC4o4t534pXyvkgNMGGX3tFmGkF","slot":83833224},{"blockTime":1625135082,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3jczBJLAnoKyTWGsr5Nm7nRXEFi2U8Q3Wy1ynS12dySKGQYDnaLsWUdEngzDp7K754egwsNk6zFtnihNLKiBaunb","slot":83833208},{"blockTime":1625135073,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5fzhXaRbGUpDdjB1R92uD3o4h7qQi3hjGxk6bxxPciAUkmcaZtP99SzAyNrNz4SHyUDKdWnwaNpxkuGxrVWDpouG","slot":83833192},{"blockTime":1625135067,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4LvjUTQM5hM3FnRtMrYEyRmDrM23n4JVQVAUMHZyzCRCQVy2WbnbQRoA4Fq8EMHU34yZzfw8xENShEJm1MKv2DCY","slot":83833181},{"blockTime":1625135061,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2qBujfy3XcFzpYRtMJo4wHFxvceBrCQZXvbzhcpiUK1eayFNAsS2KNYrPzykEXvsSBfDk6VWatuSUT7UUZFQcpnE","slot":83833171},{"blockTime":1625135052,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2FW9ohtQ2KYjetDfMdSvff6KFzkQ1QtDLDxCNbSAGmghiUjPpcYCaVtwHGXpRmVseZnjTa36j8GzFTWfcjByMass","slot":83833157},{"blockTime":1625135046,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4jCdGp1RmNVfVizWUj95sKDfAyw245sFAm4SNMRUcWXMEknnijH5m1ZkDHy2nmAtFj2UyLMZjeahS6q5uX9vGgQp","slot":83833146},{"blockTime":1625135040,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4ywWKgeKYwVbg4GrJ8xJ7XQTMDA2MuqVmUYnwMyqmXEqxxvYdFnSC2bmsWpjAAt4xQXsut9oaAwdKUhot7qiDD51","slot":83833136},{"blockTime":1625135034,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5FpWdHGQbMztqGYFvpnnuMDcKdiF91zfH8Y2v1Dim52eUCgxppuASPf4YvN3o11EfboL7K2rgfc8zm9PRj3iVF2X","slot":83833128},{"blockTime":1625135028,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4n7WBUvf93PPsjV66JQHyqJfV79bvJP4R7yj8YefzNZbRJcNW2d1ShA2y4xtKMrHqWRtd2ukGwT5Sr8oi6S2jLV1","slot":83833116},{"blockTime":1625135016,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"386HXX5MsMembfYeGpXqQ5vTkqVowWWTVVgsixxftiEDbV2yp4NobzfMSbH75UKh3RxuiwW7vd8oGus6i5QPRiw4","slot":83833098},{"blockTime":1625135013,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"25HeTAj4LrhgxpGgvXjrHDWVbcGMEEMYRSqSHr2JmXFgyhVdrnUmBD9Sx8Fj7xoV2BcMctF2VgrLME2tFDK51YF6","slot":83833092},{"blockTime":1625135007,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Y6rh3EsyNs9DMe4DbtWWZ7e4a1jF1FecAhaxj4fE5XJcXEb6KUD6qpLTiCJoQzJx8ZuKUS9pdgeyVT3BDCHPmjB","slot":83833081},{"blockTime":1625135001,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"63DUcVQfbbmWVLLsXtGB5vC8J3RrjU3uWeRwRk2oifSoDHVdgumtWcjdc4ockq9KurBRyxHtfhZ9c49BUQR4N7C1","slot":83833071},{"blockTime":1625134987,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5LiDQbtvuih6LwCpSynFy7wvtXU9xLhGZiA4kKGvYE967kjfyTBHkUAyL1RDkHLorke36vzHQMExbR5RLvdAHJs6","slot":83833050},{"blockTime":1625134977,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5r8fcR2wqiJ6XNwWvfnJKqZMLgBrdXxtrRCgwvwQq1VsEsP1ZhWAc5jUP82JAmaFBZue8eQ5XB53Ejsnb9b5jV18","slot":83833032},{"blockTime":1625134972,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4rcWVLe1jpSHTPmXvnDqBAekps79ZBzpVo5w2YA5FN9Rbt5i8P1RUv3sddjJxoJCGwVD4HNczuM8Bpf9RFGi2GBB","slot":83833024},{"blockTime":1625134965,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3V9yVdnuRnmzgidVSrMTWr9EjDte4oGUvbGqAMC6oRpa1SCiuEyPbKqxizKHG5LEFQoMcSSLMYJgrUqanE2YrVmi","slot":83833013},{"blockTime":1625134956,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4p2ys2SDrHs9JTk8biERpHvBWvqrcKBME2CCMjgLyGqXiFHcaLiHJJ2u3Dkr417CwmM2FPKBDuXuKWUoKGvXnM7K","slot":83832996},{"blockTime":1625134947,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5vydfM3xWU3NzbyBVep6sRAjxLDRokbRXaC1GYNA9sknwZQnjA7DSwNVMJ68jj3KzHDESYMgEdGuUccTzxLY9AR2","slot":83832981},{"blockTime":1625134941,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2tSj5TjraGABStaVhRrhDg5qfRzZTa5zqtToDezTnChuV9i6EwHzioqTxsDkdm6irpQBcLpaeKbENNQyczY5RgsS","slot":83832972},{"blockTime":1625134933,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"49BnY1bsE7VC4WXmyYmv225vAJEd3AhzKffJoeAs3d1JuLfATk3cF6CgWayCUZMWiT8NaXpFHELDpaPEXryt3mzp","slot":83832960},{"blockTime":1625134926,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3jnGbbAnfhDzZFhiRVHZ1XxH6D9wCqrGqVqiFSponakt3zkQuGSE5eEFT942XpS6A8Kc5ANronLHGs2weyznfWEa","slot":83832948},{"blockTime":1625134914,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"65pqmMMj9X261BpZWgzihG5sP2KSsS8YPFb1CjEK4hxCvUSemvaJXBA8JCrTLKjeHz8FGDt3LGDTE6aepu3AjCes","slot":83832928},{"blockTime":1625134905,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3ezTa1nytbCxHGPorqAZBbkGX6BzfWssyhe4yHvRgn1tAdB45bLUeMhNT8BRAHynfX4bxrhbWZqjTCGcewrXFgba","slot":83832911},{"blockTime":1625134900,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"FvVM47ngArvrjbeddoTK1TLvmYQqwo2Ag3zzo144A5FyExgaGkcJRUAD654vnAUaPs5GcdF3tHWugSKX4vNa9M5","slot":83832904},{"blockTime":1625134893,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"22piiNXRJUQVyzRrxuZ17Jts2FwKiaEKbShjjY6Cuude5f5NjY2Qfjssg4Shx3SsiR5aBeEa4UMhquTnxvcuVrJB","slot":83832893},{"blockTime":1625134890,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"smzakmkS6oFaSPPp2GDdBP4kHj7nQwaSAwhgpENuvKozHAr2HB1at2du9Lq9buyayhhqaHwGecVZs9rsvMU5fMP","slot":83832887},{"blockTime":1625134881,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5a2LMf62zCZUa3d2UyX9Fc8rRC4Z8M6gsXLdovEHFJ2Y25g7ujXL4UdGT3jqVzdLHG9UcPbYR6iSMJ79XjNZd33B","slot":83832873},{"blockTime":1625134878,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5qLbT77fWTtQV6QYhBYGYyb8DbX5RR6zUctpnwiSSdwHStu7AMqPJeHPWr6Va6EQkRDAN4nvv5tfo1ErF65Np74X","slot":83832866},{"blockTime":1625134872,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5b4hHoZ8XpbPiBE65xxQE8FQz9hqErLfpHek5WkbaN4Jtvg7p1CUtJTPRuH7vrS5iq4zRtkqdYUrT6rNBsSp5sL6","slot":83832858},{"blockTime":1625134866,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2o85cHoDftHgtFRpCsLcBmXvDhAC5y6VkR8WqH2qU3zzfPCdMy7TCabBomzUk11X41DqVJkKcco2bjUfgvKmvnap","slot":83832847},{"blockTime":1625134864,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2C6xdwP1eLeu97petcaNpjiRvGLaZkbpvDtUJ5zCZWU2bSWyV5SRoHLiWh5QoA6tsU67gFCcXVCB5ZZUdjs6256U","slot":83832844},{"blockTime":1625134854,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2eySV1dEUPiSDVGzPyAbvAKKCWMc4eE8JWpCPSTgESyGjF4JFk6qicAZT8iJoG7AvHdL5yKUSgUjyRv3yRce5t1H","slot":83832827},{"blockTime":1625134843,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"23CotGZHJH6fLUF35bt23TWjJzUis7kY8Trm3y12rN7un1ghry9gYkzaELtXLATpMubhREe5biWdKDsT6YhmKzAk","slot":83832810},{"blockTime":1625134825,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3o134kuaQwBqpeaPTf54BbzGaVn5VozW1dsQF4oS9YfphiEy5NcJn9rtZ4obwS8R55PMihMiuURbhUMNHChNBnqu","slot":83832780},{"blockTime":1625134816,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5egui37EGUrNjRZTYZGc587f5JqZo87sYBSMSXDdPUbj8MuaaVsrWALe9DWNB7XGYi151WJjCAZvZHqaDLocWYoG","slot":83832764},{"blockTime":1625134809,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4jNA3hGGfqaQ7TqEAm4sccUfmJoXUbneDZva49DCPBnF538NdFcaha7SjbgTxG6bC5U33VFdoYgPY5JuHPbB9H5y","slot":83832752},{"blockTime":1625134800,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4aGoW7mjpansTTKCs13J8QhmLRA3KXpkhmKaid3xUT1RqhJKF8WWj9n9ZYMvXodnco5EoPEg5FceSzxnDzif9ovB","slot":83832736},{"blockTime":1625134791,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"28wrJJdrkAZu3qKhqidYWyb6vvf93wFP1RcLKFRnA5nbe1kE79d4sBCQkri1x6RbKVZGx1qzs537PpqQgNESq3N7","slot":83832723},{"blockTime":1625134783,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"fVKLZeBJv5VnoSPEEB3PDGzpM8GkakeAAibFap7cbYzjeET6cY9qWgjXm7NUxUGapvvbqJmQVp8y1JV2QguN6cx","slot":83832710},{"blockTime":1625134777,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Ao1bi78xKNZgC6b78PvBtQox9hs8wSbqTQK4VgZiES9zzuPKt1y4TXFDCvzFonuHXcGakEwWEBNsnCGLt2e6K7g","slot":83832700},{"blockTime":1625134771,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2denXrutfzVh2nEUPjWwxgsG4U5ZtHm9smarJmovsNrfDqCGuaPLkfH2S5MuBNsQMaRJGZiNjG8iR7g4uyxyKN2n","slot":83832689},{"blockTime":1625134753,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5sMF5mPT1wMJBmCi5ypStsgvRiTVo47US8QhNYe7DAGMa8rfTPMtoFUocRUFDuQ3fLHJdFsFrdcdpvpfxKnSn58f","slot":83832659},{"blockTime":1625134746,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4T4usuFPy3yYqBr58AwQdRuM97YWCAj1yQWDY11CnCoKx9XzkFPrTsFkyG9tjVWiYRe2xu5djt39QQ27Qqn5NETf","slot":83832648},{"blockTime":1625134737,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3PwQTTUd4F9ZfG8XvTwrQuPu3zM19XqgvJ9wWJFuoSDWQiuGa9dVLFeFxPmPi3DziRVLK3ZySgSSgLyWu3ujxh6K","slot":83832631},{"blockTime":1625134720,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"44ddT9cmy8ftiHQrJqWGy1hRqPbMhuD1soEsmrEeqGdnPaWYTzYhvGYSstpa3i2VDPkUFSJoqxy56SW7ZbqyQJcY","slot":83832604},{"blockTime":1625134713,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4nNW9sEYTpiNHaMxyJ5iznGjFWaQ2CMoi7RiYLssCato934nD9VwJYYqdBTWg7ycncitHgFmSjwXhZ83pAFh1CeV","slot":83832593},{"blockTime":1625134713,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"qfbKNJrAovkphefhCTDcwdPFTqFWh8tHK28peFDV4XY31VufYFcBN83aUoSw2nmTLkq4sWTJFHRdKrCwJqa4hH7","slot":83832592},{"blockTime":1625134707,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3waoeo4r293HnWHWBaRPxnQMATtcEAvgjx272v4uahrH6x6cWzUBLU8ptiBQbLv1rhVPSnVcU7sb5nQG59ETWg6Z","slot":83832581},{"blockTime":1625134701,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"vfcQAVxr9i5D1i9t7QXMHXHvSsA9RgomdtLUuzcuSB9wfFd6ot6CgF6hzdktkHzc1a6MF9sTfQ8xzDWkvj81pE8","slot":83832573},{"blockTime":1625134695,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3hNsEFW4pR3CA7mchXUqfyovHgRQheTBBDmKb9CxUAVQQFY8RcVRvhCvCB6CmJp6y7GAHN3HhQfg8T7LR9j8oXgT","slot":83832563},{"blockTime":1625134692,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5ax26CEb9jdxCM4GL7w3wTHDMHJ3bn6rpay2xPZHEvyCSZemuacNLPohNMtaLPfjvNSqcMUjwAFwCKys4BWitRuq","slot":83832557},{"blockTime":1625134680,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3GHzoPSw7fiK5Anm22e6xPHWDREVsnREciJgSftHAnQXvpnfrcvVL8pnR3Swy3fKkWjuFce3cyoNmSa1f19peWFK","slot":83832536},{"blockTime":1625134672,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"34k65XMpGHPUFP1MFf5PxE1KeZBkBpkgw7BQakHEDa3aDRk6rBMBQkJT6yCEWdkcjUStZQof7cWrp8ZmfcdXVWBq","slot":83832524},{"blockTime":1625134668,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"64r65Ztci5fDHPsC5JLxvGnAqQWtBaxRxWnjgBrVpMxiqB2HBpoG4CKLpMwvj4ZS65xP9sMcq2tsumrWBxLixemJ","slot":83832516},{"blockTime":1625134660,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3VZF155KTwndC1WkV4F9QFjM3SZnrpoMn3FmaDM77Do1xvw738K3RGFgAREqFEyfxErKjCiqVMh2iSDYwcVepZMP","slot":83832504},{"blockTime":1625134654,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5hdi3hx5y3Enxd1cjdmkfJzUFQrZisP9Q7u1Qgo4k4HuqJAZBE41hezKmJfEMYCYPjGu5Tzhkbz9wCECmU9ejFG7","slot":83832495},{"blockTime":1625134645,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2PCSJRKRTqRSA2BQNFwqA72XxRJt6Wp7A6pGHSGk6S9Ha1knnFzNbmiRdzmz5QWwCLCd59dzSyKJ5yiodrjCgVWy","slot":83832479},{"blockTime":1625134638,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"fp2D3h5yDQYnoc12kdQL4wwW2kejWENWiN1JxsA6PnfisegXxdEYxYChyKYCz5QDG62da38BrfCPQQxkHNPPgAL","slot":83832468},{"blockTime":1625134632,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"rtUaGBj388zAuocQD7FP1hg9q9xnUb8UctqdixPQtmQ9ofuhPUEN1aGmyAu9kMgyiRmJukjiptrVNEjmRLymXaQ","slot":83832456},{"blockTime":1625134626,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2J4LbA7w3ravCbB928jEy38nT8awUgAEnuvUwkSzGdi2H2fmV4e5HmwbzWgVoGRgkNRnJ2tKNLKm15UPWUjJZjRW","slot":83832447},{"blockTime":1625134620,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"rvn9tBmViEEi56SG35LBFaHfZkdHbXsVcC4HYNMqPQF4D2bXsMBhkcLZDdvfLNGkXLdGsbEpdbVs6QVX26EHafJ","slot":83832437},{"blockTime":1625134611,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4xN3k79n1uhpDUF4SHWjRxK4qanknJZSgX8Jusey6Cm7GGwrg2gcQoHE5C5UTqF4oUZer58a1ed3cSK2WUZ3m2Ro","slot":83832422},{"blockTime":1625134606,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2Wor55x3JPJL5FWviNPj5GbFHJ6hEeFX4vdJRDEfwyj1Q4a9TE8frfr8XCj1XV18SQD5aAPhSXxVJYm9KKMZUSdV","slot":83832414},{"blockTime":1625134596,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"qrZbBgq8iB2MsLc5uoxKT6Fmbw5HMUiRJkpE2gjaetxxoHjB8iBf81Q54Vm1v1Zv5Ls4Emc6Mo3W9VezxU4f9ar","slot":83832397},{"blockTime":1625134588,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3KsocAhbeSEmivB9bMTHFQLgV1QeLthhExdSGjNJfbGARinqUniw9A7G1LBB9y2QQe6r59QdDeVsxm2S4EYcbkiJ","slot":83832385},{"blockTime":1625134584,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5d9GRokeZciRzTWrsvf2YBAjiuS1r1pBXGNzzqwWBq3vY2PgELNLEWRVWqiEteuDV8VGJ4ok97pY1FHUaKMkeyAZ","slot":83832376},{"blockTime":1625134572,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4zCdB1ybgVacQKmtkehUQaGXqNuNx1YVDAoLmoQEhReg4keiHRtSLhp9Y5Ss7QHaRRotbw7xuJfHRkWo1vkn4moC","slot":83832356},{"blockTime":1625134564,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"42dzZzTQvCd9ozBAJYzyTLe9GbpgvSioe2CmL2V9khkuZJzKxMnKivF3vh7Ahk9L4BEVhiGneSprmd9Ps84FYMJD","slot":83832344},{"blockTime":1625134549,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5Sr7DGCkucgfRs7FfcEy3GqE3UHXahT1tspA7Sdga6FWoWHtfmMxGtBntL76dzC8NsnBMynnFG5zbEjivJF6m4Xf","slot":83832320},{"blockTime":1625134542,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2vvKSzfcbsqCFpD3BEdff7XZqu6cBrbDgNNPoGH5j4RYmLcyCpTK7afg48qDx7nkuHQ29TkasxfGqvGRGZK3RNVQ","slot":83832308},{"blockTime":1625134537,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3P2tVBuY5EyogufxP8UiisUYYVFwqjdDtAELzEKXf6KYWvnkhm2w9A1TomSguqHDEJpUKnA3pbBhUfVEpMAvKgnM","slot":83832300},{"blockTime":1625134536,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"cULvdVTHg2mZfKoGb5RA4MCRt1KiXRJdqksRP4bCMCkk7Fdp1dUg1pRfnzerKteH6bKg6gyyK9G9VJCBoKpjMgi","slot":83832296},{"blockTime":1625134518,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3YKSgX3iRdfXrTTyaAwaEjPkePqqfCAn5XC59f5CdMFR5fCecmVmorKt2jEUU1MR6LqhrTijQtPiFe2eQqNXrkfw","slot":83832266},{"blockTime":1625134506,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"29nuuN3Xqx4nUvcRqdkiUQrCRMyZPRU4NaCgqWXT2nbSwmpEsCFjrNaKoEMi62raLRGcmo9agBdyPDFBAcSeGXB1","slot":83832248},{"blockTime":1625134498,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"DKFt2w7vgFdLCedX7m9dN4eNt2oQJ7fz5hwdwXB9puojtNwUf1VHQuz5izqBzs8YH5FhummvPs31wRbPzgJ36QZ","slot":83832235},{"blockTime":1625134491,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3uF4LcUBT1ps6zY36UmJtBj8mSR4cKLMqGQziXzBtjPfnawj6Jihx6XAYeBFrWTKFrAmY7FGy9hx4iL6VEPjnvqC","slot":83832222},{"blockTime":1625134486,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5TYGhcQ3oB8uKm4fc85k1ACyKLtNWLmpnnXBKtgAonFMm6xWou6RhXimA8zGTHPVLmvR62eoB2cgmX5GS1EU8q4y","slot":83832214},{"blockTime":1625134477,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"JG5bjyjfnnyYrDv19ajucTRoh1Ew9hUB4T9Zmae2SDD6d4WpoDwRgc3VXHKedV2X55SPu6YxdTeTiQu1VQiAfBN","slot":83832200},{"blockTime":1625134467,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"9yRdoJzd5UZt9sjifcTD1UB16rSURkB3PgqYuXycj7gjcYkj8J9BU3Ygmua2VEtQBVf7dRd7re1LGGoZdMdw6Pb","slot":83832181},{"blockTime":1625134461,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4gfEV3yJKDLtm2R1tvazRpYSze9pr1uD5F55WEKKG9bTsHZJW4p8eg5jgEYJyDvphNhYQnzZqrVHfXJ4WtfSZU42","slot":83832171},{"blockTime":1625134455,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"38DEdgjqGfqk3A3N6e3kas52GyCfwsVv2wQc8gnTYuybhSU4Z8N7rWnSYtuWevUmH7EH5RgKMKiBnVPoGBDcqxh8","slot":83832162},{"blockTime":1625134449,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"54t5dyiaoHYxEsqrQZVX9KyDTDVTQzdmLpJB6Tw4LG21ExW1k7wx6mgheNJ7XxaVRm5WaWVPkDwh87E3LcLTmRcZ","slot":83832152},{"blockTime":1625134440,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5K9YcQ2Efz6vZ8Bo1iLTCe8M4NQ6AFAhW3jHNHkvq3dcmFB4492ZxkpHCetXsurR1xuzioAJKtceP9GqpARRsuWG","slot":83832138},{"blockTime":1625134435,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"T72VLArTVnmmjTz6BAneKY2mGYG53RrLyhzNDTYW4UFco2XcuPGxzrzdXxxSTXiRVHtQBcd3BjSRU7dEosWw1RX","slot":83832130},{"blockTime":1625134428,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4McGRBDYWm2n9BnpC6Z2t7MHaPv3sbWTu9LXBXoPrHggEmP5FkZJdge41e5M22LVwZjvKDXrwR5KrqUSpPJ9J9r7","slot":83832117},{"blockTime":1625134419,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"VzaPNR92R5z498UunMXLjQD91sbh8imva9fM3S6gheTBx2u8uG514nb463PbnbLMZLymdN2C49gcdNXyycgnhTn","slot":83832101},{"blockTime":1625134411,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5RaGLdFwy4myzQiJpiL2umzSZuX1Xw6ch5tWQkyCtfpBCYbndJFHSy6P1161Avck9E52ha9CUzuBpSA3673wrLqp","slot":83832089},{"blockTime":1625134405,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"Lab1XqFacN41Cx9uTAHS6YqWpJs8FntPk8FaXKxHK6RJZcX3JoB7fvSwN47Eqs3E73nv8jRLYrF2o1BcYFCB93o","slot":83832079},{"blockTime":1625134398,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"gfbDLXmEUN2YDuXqeYAjNqJRWPLMgQTgdR1D5o9dbAjRx7SPpndoAPeYLmvpY13DCKKE6t8u89Lhws9bJJM45q6","slot":83832068},{"blockTime":1625134392,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"xUcBXFKLcbSQgR2yoHCCcGYMej4w177BUhRpnFvdZxLTquFm9rrSSgTJ6pv29SPidA69UKYdBSv2DGyFv6KALVR","slot":83832058},{"blockTime":1625134389,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4raZnJshYsoLiCanuoRy2vyUmR55zbFPQGVQbXKx2g73E4WTs3dg4yChoS72PSdpYo8xpi8Jbw9qMsn9m1V36F3u","slot":83832052},{"blockTime":1625134381,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"KmXTAaDA7g7DDFQpFxZqj8zH3aEQ1cXcMwc3GkTqFCPQpZbfpjnPvDprsUD8wf39qtWEqtcbugtYvRVyDUbkH1F","slot":83832040},{"blockTime":1625134374,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2V6PhDdfqcVhUbniSnFGczf7h7pd5aGaS54sHZfkqHRTLbg5BreLyR23vwtA75DxUvJsGP6FJ7KcMJ17RZK6vfK","slot":83832028},{"blockTime":1625134365,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4X3itENNt6aPYGLKkMJDAB9fGC9hQHaT87iaPeaFzsY6jtn2odFZVCKKCsKvmECQ1qh3QsrE1hAKpQWX7ZyQDzy1","slot":83832013},{"blockTime":1625134357,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"276KwHPL3pMpb9Xy3X4FSrgrN2spa2ACGcvMDqSRwn1zEUA9X6PKKyBE582Zx2ywbj9Z4dRWNV7UwHy7zLGFytMP","slot":83832000},{"blockTime":1625134351,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"E29JGbxSXqdFfhWoTXvYir8KMHRXDBsMuPjgmomvMGREVAvC7FkhpijxEVR2RmuAEgVoRTSBeWWHF32Xe8rWDKA","slot":83831990},{"blockTime":1625134345,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5DwWppEVHXuktw84dv5VtVQZHgZQTbAihCX1JMTLf1ypfXtQu6RU35mXssA1t44LGevYPd9byF9zMmpqTsYzmCmu","slot":83831980},{"blockTime":1625134338,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"48TQSZQnzzkPNknY3VYz2bnuAjn8cccbePt9z1YQeqXgNkackHMTwuHdGFDMvhnoyWhVkBzYzwsVaBSxLvhTu1Lt","slot":83831967},{"blockTime":1625134324,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"pwU3gsHaSGkNrY1KovzwSnrAmGX8itjegKrsxW6kqgZUgB5VbrbRt8keHWH2Wo7SKvTF9WMwthVLUisA2hjSzfd","slot":83831945},{"blockTime":1625134320,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2PsUUPsX3Bvg4qsJEHThj51UN1gMf9pk22PuKCzoqLdEcLiezB2jzTSDv1UF4n1uYXjkfbDkDADw5qPiRsRrh6xL","slot":83831936},{"blockTime":1625134309,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4cuLdRmy8uqp2Fz6DZeR6TagxNYmgP2HQwqUtcNpxcv1qAEEHrNhig9A4V7dEbfPb3hV6Audptj5m2YAbHDgRRXs","slot":83831920},{"blockTime":1625134299,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3AKhQifHyRyr6wS7ogonAaeT2CXqUg3rEX6v6UgSbbrL8ybFpyPrwgwvwf2B7nRN3noNJSJyHBFzESq15gLiF65C","slot":83831901},{"blockTime":1625134293,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"tjrn4Cab5zsFq4QHHLWVRMz4YvDF2RmqmREfh3LBafotTvKZdS7vwrW85HndwxrEyoy1pWo8j4899aJ9grfdxc6","slot":83831892},{"blockTime":1625134237,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"623dCTk3z4rsfxev92WYs79LYXGZmg3VbxUo7ESjFUpGxaYYAFgtw11gRWiEKVFM4uXqMKfcNNHbTQGqYBPrgGm9","slot":83831800},{"blockTime":1625133996,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"56z3LwxkgGzaoa7839QsFskxTPAkPiZk3JBANkEKFmcgbdMkvGHSNeTTN3ShaE9N1EtRHJJgnKjtEupfXBhvpYxF","slot":83831396},{"blockTime":1625131182,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5xZoZzWyVesiyqZfZgVVQgTq2ks4sp71A4ZgPCnonvHxXxEKZtnSCtrcN2VJQpXuBEL5ihGRJQ6NjYWUAwVjpJYb","slot":83826708},{"blockTime":1625131162,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4tPaxBZZRnGFee3bvQxVTURgvUXRYwM9PrTdtS3eRATvTfLERrx7BjYic9StNfcXZ8YkyMicWD2EN3m9Eu2gGvHJ","slot":83826674},{"blockTime":1625131144,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4eK7f4z96AbcDzT9HNUL4DDwVvmk3FPKpWEoj9fXK5NfQCB9LL3RKaHenHadVegF5rGBZXQRhpccpsKWDXppyq45","slot":83826645},{"blockTime":1625130936,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5PEHwimZaPvmPq7GrztdSXELmnJaJDNujkYWxQe82ragQqzkg3JNd9E22mRJCKeKPkVuGQnz5KCaZZPZ8zv3oXGv","slot":83826296},{"blockTime":1625130903,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Cci1en3tAgJHzV6zVUJnwv4HDDaAoeuoc9Rouk3m6U6VD6m6uzS56UFFcPFST5PuibWvWmabpWcMjXQCQgCc2hE","slot":83826243},{"blockTime":1625130873,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3DzGZ3TgCQ3HLKAj5csafqk7SiT7y9yaesb6NaP9dbwFAuaTDGyzMtU4kKAqrgoGUA2ZF5zxQoAAf2wYEQDNftpt","slot":83826192},{"blockTime":1625128587,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5pzsxcnUS1AFr93sPT3qjN6AsE5ixukLmxGyk3tqaBkvFuUkjfdWSvJxSx2Qiee6xJaMirdfdFFhzwrtN3hYxGVC","slot":83822381},{"blockTime":1625127132,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4gD6DuXPNGNVKrQMQjMUfAJis6zpQsgKbmzeyM6cspbvhyrG8Gfm2FV6YjVkpzGgeKixRJH8eRRxp5BAETNrsRau","slot":83819956},{"blockTime":1625127112,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2rGZFZYiMW538ANnL6UM2n4N1dqQgZvBmSS3KNvAz31h8L4kKqtH5ej7KH5xyzyCaTAtG2NrhhiPefc5e7KjaXX4","slot":83819925},{"blockTime":1625127075,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5swYGq11974cErnCmtjJz2xqHFzJEBthpqWnihGqexZZArjErs8r1GVPyjAhjnhEp2JCNHu5epPEZkMbUtE2iLoY","slot":83819863},{"blockTime":1625125060,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4F9DfmHoYCZm9qFdqDk9fbyA2CaPJEEsbVacYvAEm7yGcRDcR1T3f44bZcipXS2iT7Sqp639DASpH1XXzbHM6WXT","slot":83816504},{"blockTime":1625125056,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5S2XZVkVjECEYGCsMg7kF1tNKZad9V3MbdzPwfYHD1HThGq9FbyL9JgV4M8h7UzRrRNjALPW2d1a7cmfYccaMNHb","slot":83816496},{"blockTime":1625125056,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3mQxcytui1PouR75Ag2U7Gi92kWqoi6FsEq6WZ8gg4WTRvkP9q4tJW3nEGK5A7hZgERi5QppGfUJnXrbPD4xeQte","slot":83816496},{"blockTime":1625124676,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"449mRdoKUbi9aVA4pHUvCTMqJTmqmyfCKhrpcMmdHuEhvvAzpymE8s7ZpEZvKsyigQee5buuedEUEg9GsEdbWmwx","slot":83815864},{"blockTime":1625123928,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"gDQenjdEnYVAJJpkxENrMNHGtK9gUcH2Xur355ofFooViAP9MAcqrojjPYTDaeiDs5zU43vkNWwVHbVZcd6PKis","slot":83814617},{"blockTime":1625123805,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2jFgiov88bBmCgPsru2WLcZURoBL3mnMbAjJhYPh4ZypVxjetXzp9WqfWDz2oX18TmRJGY6NBWDCPqwtsGdPna4C","slot":83814412},{"blockTime":1625123805,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2Hp5Vf5UvoGJtY6aKU9hC4pbnYxp3eZGbfWKM6CFdw6dDVHgLspjL3fkLiE5sYa7CaExS2Gwn4A9i1XkZ2f9Q4RM","slot":83814412},{"blockTime":1625117785,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5YV8mkJeJxRfsmGwvRPVU8CYBAKnC77URKZebKshKvjM5R23bUcPGYGutQnMR2wLH3p5dSaC4CL9c5GSPTNEL9a6","slot":83804380},{"blockTime":1625117169,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3dHEay3CJQN38gftyuhQLogFUzvzHw4vtr8AJs2C8N5oCLsfSH1P1ocTw91VBbcKSV3KM3fdAhB2N64V8DtU45E4","slot":83803353},{"blockTime":1625117109,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5FvwcEBAZhRoY3HVHT1WcT1ZXQMsWvUjJtQDSLLKaZJNGNUi4oga9LgByt5xBheRSQDHscU7cbVtWDjEbZjvA8vR","slot":83803252},{"blockTime":1625117100,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"DWZgbGfWrrrSs5woopFEG6LHJFEd1cjyr9HNnYhQQJnqkwz13qjdJJXqHNFNVYyGuvmR2pmvZ2LVxFY2ZyQYf55","slot":83803236},{"blockTime":1625114316,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2ibg7vaLsCqdvcDBE397iPkcpvzRNMTk45Lvmb6cuuDoTPKSX7z3BCG7CZR9YtAhtstVxSvceC5gKbehcJma91eS","slot":83798597},{"blockTime":1625114301,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3isX3NiL2KseBgRFGf6bXpe6S4FMgSURZc2hq7gqNqrZV2nEyT64F2HBmoVhQaHWBD9f2bdZsNZ1WSM5fPKRKkgF","slot":83798573},{"blockTime":1625114220,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5KjmRDhUq8owgYwJXDuiHGkMor876GQq7y7SvSzzjzRx96dhm4dSaiL1jzGu36RGrukMVs3MvAjN6gijKpXG1cpA","slot":83798436},{"blockTime":1625114217,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5qZwsaWHfuM4U2w7WJxry5tJjdk1bU62hHwueXsfTtpnm9o39AFGaDQFW4DdeetH3y2fhpU7QPgGpetbw8xtx7w5","slot":83798432},{"blockTime":1625114053,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3GDYWFzpeKjsqUHVvwgkrytnhzgNoCHodBS4Tx8uPEjuAw6QGJa18zoXP5wWyqoyLTnTJsiZYrqjvoehC2chuDmU","slot":83798160},{"blockTime":1625114029,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4HAYmeK2HEj9foNS91pzHQbRFz2n6yMn7E2Nic6rfknw3FadEtDn6aNpM1YChPA63vn9E53ww1RRa7bPav2Q8TBw","slot":83798120},{"blockTime":1625114014,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"w9zvJ7dJY2ZX2pfZKtXK6RaGEinRxdB6u8xZuZpZL4ZdStkzV9MMwXf1DibcjqyK1TC14Dzpc1N9JQ9zJBpeoXF","slot":83798095},{"blockTime":1625113845,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"38NNZYkVsXyEh3LrMDTy1jxCphs12SyTtLCxVaviyi6RK6qMSW4StwRuzhdisQ6LG2y2aRoBzFbZBXZwBwUjkqGn","slot":83797812},{"blockTime":1625113821,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4ZHYwzEDueMTdzqp4YxXGDWuXm4WVWuzgvRcuwAB7YS2BWL22HjKR45QtpAEepPBQoc2CMVRMEMzBKYZDVyZhiEL","slot":83797772},{"blockTime":1625113452,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5MgR3GDGrYcE5Ei7CYtc3DBQouq26aGDXBSmXG7TddtcMYKTqCWWiYsoEcgz7uAsZDxHPTyirLQkVgNAXNByCTfF","slot":83797156},{"blockTime":1625113417,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4YMsHrtzFaSYPdNLECJRq5UM2KtaT5uad5KZuscJ7jQiTv2MuuFf5C7i2jb1y4fZRRXeq2DaegZRCYdWthkh4HG3","slot":83797100},{"blockTime":1625113393,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"XJMDx1La3wPCpySmyBMYPDWyZFTHb29qUCXG6bCSjBBnJPhUkkBhCCLmRyHTedpL54yXRW7VBJ9q2CcDu19T5ua","slot":83797060},{"blockTime":1625113374,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2VgYTachSg8h67iKhuZnEGipnCzh7CEWgbACpyfNvFxs6QBq4pcFZTpnLiz69VUAfvrrFcKYp5QWJvCyQaqJxGRJ","slot":83797028},{"blockTime":1625113339,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3snzvEM6M8MgKgfMYHm7DdgxFW5BQZWYrvufgYz6BABbe7zoYJc2zf27uZgqyPGUjLpHVnQTobdsnBv1zgj2gLfU","slot":83796970},{"blockTime":1625113330,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"59LonPDGRCHWGPCyQ2QQyNy5wYqn83JG1xjCGHQPRQySWk2t5nKhUn95v9qhE6KhJJQGRT4yT1dSKPDGMT2cUHrd","slot":83796955},{"blockTime":1625113329,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2WZZXzoyTEE1F2XFLTyXDc3ZbC8YHhoLQ8K9AfcRUxFUaT3GGeaFxBdfpiNqj71AHYzXJfwsZiTy6qq1sdm11AQC","slot":83796952},{"blockTime":1625113270,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2dj2Js8VCEyZKkBRECE5kV7QzuAB6xhSJCckPKQpHEhsaTpphKKvPuhexLkctKTmB5PAo7Q1KHKYD7oVbyuxivhi","slot":83796855},{"blockTime":1625113266,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5p8ioss8CBjceSRaadkJBpgWXrjmgZjHXShTxB7GUZ6WEQ1ezmExTc3fywmRjvvdadSNRSHzDgr1zP32gpDkaedT","slot":83796847},{"blockTime":1625112997,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4eEZtPp3wdTN4ygHkctETZN2rdjJUVhokwwycixkaFuTXySoNbYoYaEGSibxfRG5Fwma1sqbo2k8G8Ljxc6Yupgn","slot":83796400},{"blockTime":1625112492,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3QGKGRMX6CQRPkb4VFrChpWhHKR2AhqbgC4yWr9cKJ7gPqkZgfXAfAgrPj8nFZ9swhE2PgTVcsVa6FvEWgqa76vD","slot":83795556},{"blockTime":1625112438,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"432FDWE2TRvQvCipJfEMPoRoF5jZxZwVpEoLwFM2UBZJcvYCovaBGkm9vQ83Upt1KHbpxcssQy2LTjwEYG1qvz7Z","slot":83795467},{"blockTime":1625112429,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2k7Z29wewiZUiVgYrg7SoQCk978PiRgUfMrgr4zWg6QFneN3SLJTuWT63h1foMQdcpabjGn26eNA7jrDDsCToMmB","slot":83795452},{"blockTime":1625112297,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3H8hE4W4pRQS42GUcTUVwASkcieNRL47vi5R235gNcvug9sgiUmJqegWKqRCuGogZg5Sui8PQmTEdc8heWM4EsoS","slot":83795232},{"blockTime":1625112270,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3x9WuvfUNdLvdR4rEQ9FXnee8Wja7Yu47XYhrQEAKkbdzn699wYYtYdf7fGkgTsXXK5VFt3juWd6uSNxG2yTRTQ9","slot":83795188},{"blockTime":1625112028,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"55r24oiRshLSUvkvyPs8mHhQ5QoNnqYC4Z1yUdMVTJYyWVjK9DA5rggFiQ1MjGnHD85LMYTYBnHsp3dWaHLphwuH","slot":83794784},{"blockTime":1625111877,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5cVewdahuwRbg3bamN3hUsDF24PmGBmkK5SkL6gb5eMEYW3ZgvLvC8KkfNFb9M2de71qJtZLAS4TuRhGmo8TEpgG","slot":83794532},{"blockTime":1625111826,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"31Zpr1vA135wbTgB36NZzkqB6ibTbBi4cp1AKnsx1dgMVtHB3Avgz953DrBGLDBTe7shKNozbpP2ANPy3B8dnjcA","slot":83794448},{"blockTime":1625111544,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3P7dCp1JozL9dYGxFsciXMT1KuENp9BNP859Hw1ufWdohiA54E4yhDq8poeeyXWA1XM3MAgPFVkHB4K5W453VWpA","slot":83793976},{"blockTime":1625111490,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3ZFp9TjaKBH6E5J65hEQaJ2CXz4zEdsmBhGyeq7jrh9ahvaXd5PUgqZu8NMGrQuvLDk8bhSPa5vbsVYb7zGdyiGx","slot":83793887},{"blockTime":1625111040,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"43KuA5sqwqnygLoqtUn7xtBgJaTS8F6R33hFiDX1y8J87yb77UpLj9N5ARNGmDHhW3js3JxHXzjDe8YTQHfUsC48","slot":83793136},{"blockTime":1625110863,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3iEnH95yHVKZzu5ZkJEZnGrMzuyYWEu7NvbD378mRKXTcDRDjubRnfTYWsbVdx4DLnbQnNEyNNBF95v7fmrUZ61B","slot":83792841},{"blockTime":1625110827,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"XH2HgrAkgxp69B58v3pCipQUSoWxksYDvTWPvVKMZXDuJvLhWY4sG3qNKm7BdxGHjr2AAo4SaDTd8q4Cm8gumZc","slot":83792783},{"blockTime":1625110818,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"mRFprjhnxy53hzsvypoMKUaQRVyM7PVkZAC9brxfUBqL3fYBpj2Jhowyck44wCrpMiWntCQqc2PM9NUbYe8aGGN","slot":83792767},{"blockTime":1625110581,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"3GTM95FtSR79VgdgR5Gi4r7TaGnqaY667KEQykwcSG1vQ94Cii9kAKKPNT5nz7Tf9f3ZUvT3cvUnWHREL2w9aj9f","slot":83792371},{"blockTime":1625110362,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"46W8ago6PkmFG3RUYMSEehdS98cz9KpRTQhTazJF8PXBcqk4h2RZnHstqzxYsMC3yGn7VSVSuRKfg665XtZ8sfWb","slot":83792007},{"blockTime":1625109573,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"iDUoLmNNcMgBD535T8By4hnjW3yhdzbsYptJPc3nFhaDipEvfso6E9kpgfiQaevZi9HZjjeroijU8cHzf3n4RAK","slot":83790693},{"blockTime":1625109531,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3D5bcvA7dD8FaiN75jj1eNFxLvxxsJcm5r4m9FsqG8yxi9oFGtrcSXSA3FGNTLsjvtqqMSQqBZwnu4fr1F7hiGfo","slot":83790623},{"blockTime":1625109385,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"YumDKFRs9Y4FCYcFmraixqzyAh8NEbHH9GDS4dN2rhW36rVpKGZbveDrc22b6co7AD68FkhL1unxacC972sJXXb","slot":83790380},{"blockTime":1625109273,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2kng2ipQaNeBpvgpxjt3aZ3UmEKQ8fmpvVK5WqDQRJEFcyvRsTH1X2bs1A2u5akazGudd2zhvssTgvG57AajYd89","slot":83790192},{"blockTime":1625109216,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4TiEb6y2HXrxMJJF9aXXjUdfnTCFkCGPgMa4EdKAifmBijgaox5jVvsFNRBgyyBFBp8Gat4GUncGZPdiwGEQWHXc","slot":83790098},{"blockTime":1625108790,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3oGfTMGGYrL328cWmoMpeoL9z6Qw7EZYZE5skeXMaDGSzUxKmic9TSy6pkigAZHfZaoc8WXsQvhVrysTDR9EeXdP","slot":83789388},{"blockTime":1625107108,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"RyBXXzGs7SWbfgzBvyWH6WKTUg57Ub839RP5VzK13BwYnt7qW1ZLL7fdhavHFh62XxFjyVfVLwWdPKoyoHBJEZh","slot":83786584},{"blockTime":1625106708,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3Z9eP5N7s84rftmXierPVpjv5xqAWEEz1BvqKoLkDevDZYUndJ1EFUWCisMCgLPmbYHBfmrVQizdcjBqLPAAnKUq","slot":83785917},{"blockTime":1625106705,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":2}]},"memo":null,"signature":"3isxtKq5unhu2khHawuyNZq2B8NTnHWEsvRjQ1i3jNGZSncP5HRybMYxpP75WLsfvu1FocG5okAygRBkmdhrTyY5","slot":83785913},{"blockTime":1625106705,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":2}]},"memo":null,"signature":"5XAr2hb1FtN8Z74DXZXAKyt89bFaJJP4WTr2T3szQNPi3K4seh7s3hPjJvhUzYmqkpf47ouYiheWjEBDX6sjnjgZ","slot":83785912},{"blockTime":1625106624,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"3WwVwwJ8XNaSbXmHoV7ocGTKzReMpECntcoBAgrSAqisMMm7z82VZgKz6w1WrjqYN3F4v37wq3yRFS3QMSQzqsL8","slot":83785777},{"blockTime":1625106385,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"5iU7CiVxRkvAMi2BTx5Di79HC5ywjuhCtZZULYJntaKxZnrNETUBrDdnFSmcBQr6KCdKoryY6TCU9URixYYgBVsc","slot":83785379},{"blockTime":1625106364,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"5MAdES8T6ggyKLCh491D4FUB8QSnWN37CUZYeYyjdM9DE1gSFtUPtHJAxsAxqscQziEUAXXtpi8ca6UvnmZFnwqm","slot":83785345},{"blockTime":1625106345,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"4rTinjXws3mi6NK8FwXJomu3Tn9MLV64QcxSEFjFSMMKtE9To7Tr7X38T4tHUHmtCBfgBhGUtAHN4Y567JAR4oHJ","slot":83785311},{"blockTime":1625106072,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"3vC4QyEUdXgCwxwxKwpb7fVdAKdBLuSzxn8hMz1KT35wffAwvLoKoDw21s9pS1oC7AGyrqxsh4YiPkdwbqjuwtT9","slot":83784856},{"blockTime":1625106001,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"5t9GpqopA1dQXYEuRK7dSPQbTP5dmrNkw7HQTNvsuunoRANdfHQYcpsGJCkt2qTuWdX91WPpKxv6kpTeTrq5frpM","slot":83784739},{"blockTime":1625105988,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"26s8TRV4jqTY9GitioAmepDvZsVW2LPrKn6V7KgAXcnJwv3ftMbtYTsRXoa5ydQYMi8STU4zZSx2mY1KPUs62Etk","slot":83784717},{"blockTime":1625105968,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"5afsY4DGNRFECe1ojgWPtNjMpfn7Tz4ykm5qMrG2ZVGFhD2ocsGvjrzmEsYUQCKmEFiMdLYyhyaemLPyAFnrCtXc","slot":83784684},{"blockTime":1625103759,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"z554QbETzMKzMq3BCRpkQ74u5eJ3e4rCxJjkEqf7DwmmPfWfdsPUXkkLacya9ykLBPyiFwkw4vbHJPAuufeTuab","slot":83781001},{"blockTime":1625103753,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"hQx1Qvc1UyY9WgA7fVU9hRV4CbAqzo5NDXGFiV1xJgTWY5j2a8uGr1LaPqWPhJ7kNYousQp3wGQSXtkso11oh4D","slot":83780993},{"blockTime":1625103706,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3cFmu98Jt67zaFt7T1L1t2yzAFp43dW5nSr91YDkQ4cGZk1AGJF51GiFHB3MbZK443DQfpte5Q9WPyxpcn9fJPcv","slot":83780914},{"blockTime":1625100342,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3uD6yQhhjNnkpvvgY3AQVaSDW7vC7y8rKvuZjvG68NmaurLDAMJTEfakt61LbNpBRkUrJ88ALsMcjEYxxVF6gZXM","slot":83775308},{"blockTime":1625100121,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5sfxTMc6HsgbtrtyfdF5iBBTmdVCEbLvfHfUUiXocokCVFmfPbcV1XipVHWEdjvsygdU1JyEu81FrGWVELgHkJ6e","slot":83774940},{"blockTime":1625100030,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5Sv2ir1Kr7ap5Mz7hdNFPLWs7Ruo31SCVbWHwvqSKRCPrBPyv7PwLi2JK5Zrav6kfyQAdseDEdq2B4zLNGCs5vCp","slot":83774788},{"blockTime":1625097301,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"aMZk9EW1iREhVRD1GktZ93cVFmTvzwr611GbjLyJzcbqmAwan2RvBFhRudxqEasMFXMnNW4FoLVyCtReUpTAH1R","slot":83770240},{"blockTime":1625097291,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5hcK47KMv9XecFCQB9AFpRKwk4SXQHnfbWAMEYdYHxjJQYBoFYcog8RUFp53A47TLEMpnPdAZWwzSacG5ywfyYDD","slot":83770221},{"blockTime":1625097244,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3LFvZk3AUEwCxUTHm8AST5D3BwGpMHjdKY1pBNuJZJvTZLWEr3frt2vyfiQ1ibdpMEqBwH6dth2GHRF9pksPHVtk","slot":83770145},{"blockTime":1625097169,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"JNfRwLDypLxi8CUHuANEY1knBQv8Ct9ubKFyh9gYY7aSU8ZZhZ14gJakxoSoX7HjUAQLtnNwCyBnrdH2SRm6Dcb","slot":83770020},{"blockTime":1625090365,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"A2vsY66Ui324BjjffCbVSBVBAVL2pZpBXdZSS8KGkwywoSb6NzQg4Ki3bjXLMEwtiLSUXm4MFeeNhV8FWKHRa1c","slot":83758680},{"blockTime":1625090334,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3vJBz5dDezibaSko4VFCDpoY9dRHg42eM8eScNR9WT34XuL829soBsrLtuQmr73iNVSRoh1rHTAfrMUn6pekipNz","slot":83758628},{"blockTime":1625090296,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3zRM3ojzJYyuGPMSSEudeppWuXBBDMRCCdXGdhHNW183YwGvwuwD5pfd8c6oHBpdTVogngTwytL4UrJqQ7mQnNzm","slot":83758564},{"blockTime":1625088585,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2Fk7tuqVqKJJxoxnd83kHMPDrdvgdNLGE54pB25h1yRifiqXMXjJc1NALxwsTyXbxJmL1Nh7izH3udoz8N3TpJMW","slot":83755713},{"blockTime":1625083842,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5G5rfMvmzmNkPRgba8jeNJqCTH8jfsWAAP8FJgWs8vrkSqaJRe3fXWknA94Humy34532X9qcETbRjyK6tMqMVwnZ","slot":83747808},{"blockTime":1625083803,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4F9YATTyM7pp8eKrjzMapM4HKRtgTu7HCZAia5xd6HK511GvdBb3ZMnJGqvUYKqqRhhRWmMvaELZz7KKQd56FWLc","slot":83747743},{"blockTime":1625083789,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5TFQCB1cEizjBwGPq1SUkwa6K9fZVwQa19nSBYFvRe44hFC5RHwuYsatxbqVnH8DibHdhYXc95szovuDzHfyXpJW","slot":83747720},{"blockTime":1625082450,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"4PRGr5LdoESkUZA49kKkoDiJQKSVpKBMhPFtsDuMwJs7etw3wfFb6PMjuiCPKUCS1jBdfe2KM6F3gfshWTuJqgis","slot":83745488},{"blockTime":1625082415,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"5zdWyZWy81EvxQDRyB2g9wUfX72aqe6ywai4bVnKTktk2wz58uTp8whVLbeKzcS3RaPQmiHUowBfc3pyoFW9P8RE","slot":83745430},{"blockTime":1625082378,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"4QMPtdBV38xnL5Tudv6hELNkJLwEVYa3rRn7Qx27UMCSGVVh4wegeJiqnFpErFAkwm4scMvrt5ARpdhh4TsbkdJ5","slot":83745367},{"blockTime":1625082352,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"5Y21bUMfsJz9qVCoxDRYdu7xvHpViv1yASS6t5rs3jRJY8ozfZgqkKBd3Dz59wuQUrtmKvvbzt5Jgb6N21CdL2Fc","slot":83745324},{"blockTime":1625082334,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3yC7tJUr92VrC8Xxtkxz9YcoEaU2cTTzPgPUop89g4vnzwegWEFAfTg773GjCnN8fpVp53EGkK2oEKrBHMCYGWHk","slot":83745295},{"blockTime":1625082324,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"3PAjXGGyKSvWEdDrmV9wpagx4NvDAcv4V6yvVKtjXc51JdEv35WvwFFytUu11se1nnHZZFYzJ32NoGUtu8Dyfjny","slot":83745276},{"blockTime":1625082307,"confirmationStatus":"finalized","err":{"InstructionError":[1,{"Custom":16}]},"memo":null,"signature":"3i388n2cpKQVmJmY2jEw6waHW6kgz4917ibTfh39bkhKoNSy7eHota6ZkJ3ySCP5HaUpso2L5nLopGfEvCpW3i7i","slot":83745250},{"blockTime":1625082282,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2HtnG7EBunoP6mqpeTxspm1vHBgyoi6yjUWd8zesWwSxdaruEgULar4kCWATZ9rHTRQEAf5FG13dPnkbzzFgG4QQ","slot":83745206},{"blockTime":1625082258,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2YqXbrsmgpJ1eqtp9EaYQZeTy5aCo6yzhy48wcSYpDjk1zwYv7GUuF5s9Dyi8923WoGLNj4EvfJdxoMZ9mtCQEcW","slot":83745168},{"blockTime":1625082231,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3k2gZ3zCecL45wvN9w3pct44av8tuBLoiqeYf83CFn7sLf7Y3kJpJ9RqTX3k5nZtj4dWWTRMW5nwtMfKMNqhuk1S","slot":83745121},{"blockTime":1625082219,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4nHDuNKsAtCASsq9i4R5BaVTrPz9aX9rt3Jf9eDyti7yJ3p25KsTJ3Bk5aGUXxXNFBgP8bjMWPjqSVcpxqKBVvDM","slot":83745102},{"blockTime":1625082198,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"24U9Zj92Fq2vsAo5RzzanPp26nrxiQBsf5APZSXPBWtZTD7XVbnefYtAGEKFxSfQ6YgSsYd7ztuJuEnaZMQvLBur","slot":83745066},{"blockTime":1625082060,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4gZnVi1V7K5J9U8SrApTRzkMYnBy5bWye8RFsPC6NYtvLuXRgTqGGp7yFXTtyEKSfmiPq5onSKnUzKsthMLGaVr8","slot":83744836},{"blockTime":1625081566,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4bowonddAK8pMSiKVRspzYpvYoF5gfGZR53GXBsni4mJnF9tFVse8bW9eEQ3mhuzBmPogxo1NXBHJisoH4PE8k3c","slot":83744014},{"blockTime":1625081130,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"Qv2gzJUJbvBquECu87Zj7KNPfaAPLCUW7UNnMmaB3pzg7LekrGijVtApXL4ipprvhGVnCMcjxxkpQWS9VLrrncs","slot":83743286},{"blockTime":1625081085,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2yZnS6ceZK764YdGTNNS33hNDko3UUj6YmZw3EJfdHGX4Yhve9Zm8VRht7kRfWzU25wdobTCi1BQrV8wXXwpL5i4","slot":83743212},{"blockTime":1625081043,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4tNoN5sHBCXDs8QfSRV1stBUprfEN89sHDGiNWh843PzZmCsNd6XMSNtAsBqHUj9P26Skhqk5RPWNxxAPH81orej","slot":83743142},{"blockTime":1625080938,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"23q63DeXspDFwcnHVtr8EjNNXzaej3Lgh3AzW4MTgsDwFUSZEGS4xaZoEhMrJQpQg1BpvLeALGeJwBSBZXmYyTy3","slot":83742967},{"blockTime":1625080909,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2NzECcGnz8r3BUJNBfJjy7aggpVRqUxqRt3W1oK55Tn3hdM3GD6kfbi9ZaJ3rBNw4KfipXfrcZoc1oir97GFSyx7","slot":83742920},{"blockTime":1625080890,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5cwFojjEA1NfCVKtaCP7RZ43VHAmsLFM6rbkst5fDgdLs74jaas9sjhmw7eKR1FT4v6s1sqU9A27DKPqjCoZTxyG","slot":83742886},{"blockTime":1625080818,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2F1ruC7MhdXUwQommGnGXc5CQc8D14AoEyru96CYwKJzdwbS2UJekdhiCZ7XxS2Wv1kGDnQ2DmTZz9RufJ9GEJwu","slot":83742768},{"blockTime":1625080752,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"XK1RtXnGvzJy5h3pDFSFZQJ2vGSVW3nf7pxVrKdCoQcX4d1Q1JEu626dZWVdTCX4Um9oC9ogPHoqdphRVXCxff5","slot":83742656},{"blockTime":1625080725,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4nft5Yeq8768ZMPEX7rQHwEpwc78H7ANq2knqR6cDm28MZ6Q4Jp1bJ4K58DMgjarfo9x9LYdroseDsv2opbhUc2g","slot":83742612},{"blockTime":1625080572,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"A91WL3BPgvdrbh13Q2HgceYb9XLvjWfpY1KXoaRr7oTmm6Nsqi5nbgHowgxt4zhnv8d2BJGr3toTiUHGYoNCoej","slot":83742358},{"blockTime":1625080569,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4hLpwbaVcPiAAD8ZvE8AjCZEb7WdodRme3i7nCG1hme7gMfYqxE2dJxDT4Lfac412uwV2jLaBt1bRcbV8xiTY3e3","slot":83742351},{"blockTime":1625080566,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3h6hSvG8qX5QiAUCzP3ZCwGpYvKgQryDz5UJy892KSN3iT3D1Z6Bh3mnGagiXX8FnbxMkTNuDFQWHEtbkJT1kR56","slot":83742346},{"blockTime":1625080563,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"ywnRQcnstKTEhrJ5GecSy4rUMtqkHjdo6sCBLeYLGArgb5LmpCVL1YHdfe14oM8vXuYBKLR1LA4awRbu1BGLTKa","slot":83742342},{"blockTime":1625080561,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5ffMhjrdMF2sRLtC4LXkQU44pWqSe2C9fv5orMaKtXjzotxmsAGdqce44vNFaFhrd49KAAKT6qQmbmGJrvJp32dN","slot":83742340},{"blockTime":1625080557,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"8fBNbdUmuthNVLsEhz5B9XXChWqJJw2xr5VpBif5m4gPJeQbXUcpynnT3R25CMXxG3i4aNkddYRBvf99MjFpqyS","slot":83742333},{"blockTime":1625080555,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4TdQDac79eFZp7SkzRtwmo193mkExS9ANJ4Fm85RTGHEP1CQmYssqxCQQHfr9qstPG8rLDE7ii5E7WBRpsMSMowg","slot":83742330},{"blockTime":1625080552,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4HM3VFTx5aVP2c9cpeBvU7oHdrp4Jserc2cz2cfYZ6JSXjEY1smhmaFCSd1t6TZuEei4nYjBRHF1djJQXWETgXD7","slot":83742324},{"blockTime":1625080549,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3XGGDhdjbi8HsXNiBebYq4RCq7QqMEm1eGZ69Wqjm5peo7rPVHKG8kDDyMkuqBMqjbBMDXZimrxAzpvXCQvWoWaN","slot":83742320},{"blockTime":1625080546,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4UbZySyaFQGvkBS9bREh3FAGrnrKBR2THBbvurQ3xjEyvJmTeR9dT27PeAdmu3sBJiAe8FxNhbJHgU9MSgWmcKug","slot":83742314},{"blockTime":1625080545,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3i9Waf4jTNG5ZjWuZhjP346duTVFk4TpANj4sthtyfQ2i8tBTqCznij2JyNdepVrzmbgXVJrmFtsYpCAoACBbnm","slot":83742311},{"blockTime":1625080542,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4RsMngsNSTkMZDnLoqCGmDLwDkxHj2ft6KB1iGMxVW93JnFFHGcAw727zWMb6eHTfNf7QC1ecXVj7FBiMHJJNg3Z","slot":83742307},{"blockTime":1625080540,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"26p7xwzMDF8VHZwHEY8jsJ8RXvWP6sgLE81jnECe1DSWcSuvrhCAWEJyMePr9nSuPoiNoVCtnrxLBTYsAMjqY8sb","slot":83742304},{"blockTime":1625080536,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"9NqRc3yH3UJsHz9yqvmJXr7BBBfmq8k2SKvAPToMMyXeSYqHGnwYnW5jZLCk7egB9vDXaWgbJ2NbA4eSGWGRbdG","slot":83742296},{"blockTime":1625080533,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3aiaagpawRNauAuH8x9K6RSgBaN5nRvfcEWCdEJ7KCC9HqgRjEz2wQwrnv7U9P497rhzHAnEsWFyXhwzsaw1K9YH","slot":83742293},{"blockTime":1625080533,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5h2hh3RUvjEbsLTwEXWLzCuMKK5NLCD1fXxhXQcmyov1uXkPWzwX5efZ6orJ8BvW9KuqscfsCEXVVJ3mndcsgGrs","slot":83742291},{"blockTime":1625080530,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5QBGEG1Qdqd8QFzgDVDaKHrao8GaEhVb6FSV4ngFDA2DNf7h4upJKhyykHirFHiQWexpSDi3gNcAWiTtrjRRHWz9","slot":83742286},{"blockTime":1625080525,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2JXKnuY81S9fW7Bn6kmBZvrP2X2bD7yHFui4QVbvjNnvNyateM3P9CLvcGr2kCrS7Yh11fvefrHca1mSYSrSX5ka","slot":83742280},{"blockTime":1625080524,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"29Hf86a6uAkDTr7xBtv5MyWDmdCgKcytEUce6DtY3ULciVPTswmUvmPKrwnMD76nuoKJ9anUktbGMieaTaWcDsjG","slot":83742278},{"blockTime":1625080522,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3mgmoCrFC5BAAfa4ptg1FwS1G17iwVVuNX7GRM2xLw7p3Uuu3Hvv6qdQPLYY6vzPF1yqmUNPV18z2wTZHgLKyDkW","slot":83742275},{"blockTime":1625080510,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"23jvRF5ueMPDw4vwUvA39ron8a4ypMa8mz1Zgj7Kum8eA8wctyNodsSRQi3w2wTTgWW1zDsNS7hNDy5JqcWkktaR","slot":83742254},{"blockTime":1625080509,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"66mJywxGRwnMy8SzvEcAAvxo1HuobPysRSRMNgqkLBAABVVJH1EyszW8zpTUgLZXuDj57fcU8knEFXWL5fDch1jS","slot":83742251},{"blockTime":1625080507,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"a6Qt68zLpua9fEmY6TsGoB4JZMmtv3UwFYDJvW2jdXyXx4XoUMPPp1r6vHcvGNrNzo5FfSv4oake1xBVypT9p6J","slot":83742249},{"blockTime":1625080506,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"28xBx8jMC6qQ2EhooDvb22jbZ5p3GECVmgdeBdU6NYLALDaAckuCsnLXFmGkUntHoetsY1NRc5SaAoCrSq6TDgwu","slot":83742247},{"blockTime":1625080504,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"t6ppd7EMwLBoiTnNJreJG3bH9AKvcMSfLQzHApLYnFKx1SrUPBPxebXndJhcrNKkkrAqmCw3akpAr5cZQoZLwJa","slot":83742244},{"blockTime":1625080498,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3yRumNAfQAZYRSJzDvbyeaLkT8EWJNh8Td33dMHYGCybCViRPQCDA4jGm8zGdsU5jZTrR4QocaLiJtbqWP6hDfYj","slot":83742235},{"blockTime":1625080495,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"oUyEcbHJyxRS8ZdKRAxgFDFjUGWyiSddwi7L9J9YxreHGtFxHUdYQEDALYgX9igJ8qJDwzoWM5JXtwfYQdgEeoh","slot":83742230},{"blockTime":1625080492,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2Y6ryJCWvHWn6yHFtgC3wtTyr3Q1gLmDFfnRLYHxZBE7msUnx8rhwkHp3JWChqEPRfMkt7Z8Enw1NHWhj5ZnR9ZH","slot":83742224},{"blockTime":1625080489,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"51ea1fSvh1wC5iBf2ZZuupJXiSkgSa9tGfxxeAkEMGMw8VSXjpxvHxCMkEQctd4xHm95zshqP2z8YAkJKffJmLYf","slot":83742220},{"blockTime":1625080488,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"56KZVESQBoup2TncS6i6KnsQsnVRuEVbaWZFHXeeC8tTK8dZKwJ3CbTZUpj57HN3wQAo6uYWX5fdLKxG4tF7e98m","slot":83742218},{"blockTime":1625080486,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4TsQrrNH3ArwMCAM8WVspCpjURAB78wPnnJDFafP7FAUuBC2Epfyn55YegZLNiLbjL5sPDVsfcLMn2uaAwLkXRQw","slot":83742214},{"blockTime":1625080485,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"22CraUwLFci4q25uM8YWqExkCEQFQ6jdpA7fsvVJsUbRbqSNLkHoch7dkhLCzsG3t7qR4jzrvTzDJn4kXWToLWjH","slot":83742212},{"blockTime":1625080482,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2cL8NPoLZHWkzRoSGoqVYf4iHZKSkVz76qfQ1rcGNuihynqxQ9mbD4XrhTFWRhutwhMuFbZdknDkTEfGPEMMKMFu","slot":83742207},{"blockTime":1625080480,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"Yxzhd9LqfZdYsEYMdi2zpYdzqjpK3GCVwaxirv7sprR9fb6Vme7J1JT4NJnyUdjY1wDEuxCCiwdxYmwGvMiw4Fi","slot":83742204},{"blockTime":1625080476,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"52qnUasupgtAnnTABtnbooacmmRTctXAMijF4rYkhbVptSxgFBptPgfbF9rmaLqJxzmAM3qkNF9jvQFHA5i7qKip","slot":83742196},{"blockTime":1625080470,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"LU4RYMnpk9zeBM42ScnXacM1MPFWp22uYZghHJdTKnszvLda7eCd5cbAsQUfyT5RgrAp2M1MnF3QWeZmQJV78Xw","slot":83742188},{"blockTime":1625080468,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"uiXbRZwJtHQ1FnsZKQqEe6D7fmwHL5Wh3jVTHeHEKwNp4kJVXwXZQYZRSVBZFBL2vvPYSSj1AAttCPZVTGfnJBM","slot":83742184},{"blockTime":1625080465,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"25s3tKEKFDHbffhyoXs8VLuLuKRFo8oxNQHfD1Bk8noJo54DrkfDZqgQ6wrQSKaFkrfRX7AaXSB7kAberMTmJMcT","slot":83742179},{"blockTime":1625080462,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2NT9pVA99VPGfZE9ppdMpjP2J5M9YkGh3ju9dvsp4aXWainMT7JVwN9HMtreD5X65fo5L5LP29Yr58BqQM7A1Zju","slot":83742174},{"blockTime":1625080458,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5b6MCYCPsbQMS3fHTJDg1ChKfvwzWEKCYkm51a1WUVhKPNuPHhPGkS9m1Y3btWMX8R8KpyRUsLYScTKzdGN3DLoH","slot":83742168},{"blockTime":1625080449,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3HmVF5UBdZFYggu9jH19Sf26K6eyK5vSJAz4E6ubYMPf8vvhDFysYaX9Bm2mfz5BE7jemVnQ1yfuLtd6iWdue6zb","slot":83742153},{"blockTime":1625080447,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5i1E4PJ1DdfH8vWBwqeRZeeu1r4oNz9twgb6tJjZ63Q8UXgvp2mqBLqs9P2jYxGn7mji9W9nhRaNj8zd7XX9Mv8W","slot":83742150},{"blockTime":1625080444,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5k9mhA8JmgwKbXBLqfazxgdKoSGu2Eny1M2bYLmKqkzGRSFmxXXMj1JgJFTqraos1sSUxn79oxkznvRYwHmynHN6","slot":83742145},{"blockTime":1625080443,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3MFoBBdm7twMjSUswtgQXtXeNLaa1kau8r2euLfQmSP5qHYoZPbGnqRCqRviuUUwjxgntZJ7nqyT8VmQmqZFTfSN","slot":83742141},{"blockTime":1625080432,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2CzRr6MjHfwaPSxH6YPjhwRTKhqxYrQrUui2XCFXvCcizfDUNkuWbrFCrspmhdiSXzv6xtDApZkzoa6uMGqW5DUF","slot":83742124},{"blockTime":1625078812,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5dUeCiEdUJkDRfvsVRSDLVD3Crd3ntoRxN3B5kJ9MNkFA3nXdi2kpu1LLJJVN84aqqkxZ9wLvXd9QwEqsfEktJng","slot":83739424},{"blockTime":1625077260,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Z8oQzvS517n27wAeG1zX7Utwpvf7NUBdRb2jVRFeCj3G7hzFmWJ75r6DmhVZ9tAcmwE6GfSPnEmWxvASAoiDGga","slot":83736838},{"blockTime":1625077260,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2vnnCW6oe6hGNE2e7tW3qzMZawroCKP6jSY4nxmQX4q9D5BaxMEyeMmSZMhAcoaC5sPUF2DBV5KgL32fMRLsazDV","slot":83736836},{"blockTime":1625077260,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"21VZZ3aBWCcq3M2i35niFDf2MFNxw9dV6WTsTfuD6pbsBJQfSw6QeHuGYcV7Wwth2b85ojKkqUtdoSLD3afvoMHL","slot":83736836},{"blockTime":1625077255,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5Et7W5WhdWevzYkmYup9tK3YyPCRjqQjaRDRKL2HsCcTEULVMmSGEA7EXrvi2scSyLbBxH1cca4t59Lo396oHpJD","slot":83736830},{"blockTime":1625077254,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"67SxZmi8fXUXrYxKmVWSjdVxusEjjKqwyy7bE5Xm76iwDDSCz2ADUxbs1iftcDDQ6cQwGzeab1Tmr3ACoZfBXRoi","slot":83736826},{"blockTime":1625077251,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"FLRY7QeB3p14pPTonjN42eBcQHBaWQKyknj8mHW1A5igqFj2CAYUXjTWBZJnTSt6daUZdvTrv7rZKDtvxQaPfVx","slot":83736822},{"blockTime":1625077248,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4ZyEzChUetYkK9odn4u41ARZfFXPNZpPKMzyTYej6koCXM3Q7ukQEVWbJueGt6z4SfZrZFDZbLteCjBZHJCGqvXs","slot":83736817},{"blockTime":1625077239,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4BjjGVBPwNdeSd7mSKJa1K1jNkQj9wsaDvf55FKp2CE2cnkEUAJuHRU8AZYFu3MRWmsQkurnrWe7jFwHPtRBSghn","slot":83736802},{"blockTime":1625077237,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3AwRoeWEjSBA3wK4CiTMr27RxRQhvSGymYXoYDUQcLR6GnrPhzqa9RQgBwzrokrWQ3q9wv8FJci3JwXJrowJxj8U","slot":83736799},{"blockTime":1625077236,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2wZbn8Ex1UYYfqLeoPWjLwYmWSiSfuCEywAtLFkKV2XcDW8qPhcC9yVaCLsFHtirjSLZtRhn4abRLZSuSYfRYG7s","slot":83736796},{"blockTime":1625077218,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3qDcSkUoFMjQa6xr1YMhQrfBCPm1pdi7scd6VHJZBQwJcaQoWEW48yfxHaxb7PtEboYMVgTr9HpHC8DD8ApkRB5f","slot":83736766},{"blockTime":1625077209,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"62pxS1a7u7L3PMrW7vbNKV5xHMbhoRMzkkBcXJAVxXd6Try7sTDYxqZL6gwwKPKsi1aWAuH1k4forn3bPPpJTxnN","slot":83736752},{"blockTime":1625076786,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2NSmkMj1XRQr6iHzRXE4F2otgVQj2N9QttfTP9wX2xryFPk794XK5AJwohUL74natHAt3R335ubazTieZ6nh7xeK","slot":83736048},{"blockTime":1625076784,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4uxNVxkXrnDtnpnPpEw4CR1BaNoH16Fk8LC6frcaQGXcHym129zyGb5eKJf317v5SXkgoqcVc171TgCVRm9qBoEN","slot":83736045},{"blockTime":1625076774,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3trijXRfh8Cb2BDUcP2wFyx2fYmtZAdDoAM4MMbeb2z5s9LXQLTeAVdNikqkCNgkAPrPTmRakmoRVyZKgwe8NKWL","slot":83736028},{"blockTime":1625076772,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"SK4z4KaiMoUVMkJMAa6z2kMziUEuB1t9jw83fcmy8ERx9rnMHM5mQYBwn1ANMu5LaZHaBLCLB5VhVQ4wCJXg6P2","slot":83736024},{"blockTime":1625076759,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3ChPfmvLbeG7hbWmmuHADwNrXsRN7NsKfzneqPUeBBQmpJZP26SNag95i49LKqvubdVzNfAAnThbzPyZXFFfj8yk","slot":83736001},{"blockTime":1625076756,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3svPmAri9wHTGbbXd9hcnztPELHYRDGvZdidQv3FCYi1hMCRV8VPxujLaCRreVfMEaKV1LVmmKeQ89XakViHv2oD","slot":83735998},{"blockTime":1625076753,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"TDf8yvHRbdZKgkiuLE89eEmx4c3BayFFKWp4EnCnFePCXJUyjpTS5xSvqNFP3cVa7nNAkeMhQHpEMBcspfbaqaE","slot":83735992},{"blockTime":1625076747,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"35LmRSyUkRXkz1TwX8hW37TFYVnN7adnz6scFdT43gp9XNJfx7xynUaQSAtA7YBtZckdG6osBwL7QKdSMoF5TgNs","slot":83735983},{"blockTime":1625076744,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5Gfyu5QdRhvUQ7frhZUHaDCCeP1z1GdSBKeF36pdaP7ZhgxqbtoBRyfTDuYY9V7uQCYH7b6cbpL2j6yVY3Es6Ngp","slot":83735978},{"blockTime":1625076537,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3ktJesFoFTsWBihjBUtCa6bpQxncpzMJbkkaVgsVkTSgy8QFS3DcaYhkp8i5Eu7FQ7HHejPhCdZ3cW55ppXYXnPT","slot":83735631},{"blockTime":1625076534,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"537d9T6SAMo3HqUoxKLuEtQ8VnAMDYswCzjDbGZBqkxU3znHhcG8LUZv6K2sGccStKyvwo8SazNox38B9zmrYP6b","slot":83735627},{"blockTime":1625076531,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3NqcoKpUH35QgVBPh1yycVxAYj4hxf5AYAQjpQsPCXy8DnqZtp9Js5Wd46pU4HNBBhVStTru3CBY4BHpYqWQjuAJ","slot":83735623},{"blockTime":1625076529,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4jPcNxgzrNMohYF4f7obDKw1TUkdFcq5MvQ9vK62DsFvkXunJBXVF6AKc8jabkAvDB7M26TtAV3zxPveFn5WAXps","slot":83735620},{"blockTime":1625076528,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4LyrcDjE43uVzF5uCYrQ8SAkNuYAccc8dEyJR4cdmTFnkSwhs81zHYTundmPKP8iqKZ1wd5udUPfxuiLQdkHtpzk","slot":83735618},{"blockTime":1625076526,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3UgoibtBrbbUUxXqHgtHqW8X63sSGQejEm1Eg8H7xuVD1YyMPdkcu23m1GSbNgD1FPqn4DcD2FaLN122zMt9MG51","slot":83735615},{"blockTime":1625076525,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Q1dXq1AhJqeJjoEsmJaY22Z8xNEBCKGE4umrYVbyKgATA6riEHhu2otWzDYZgibJeQGt1yP3vnydQ8kkiLvSPJy","slot":83735612},{"blockTime":1625076520,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5pCieFTqV61ti5jznoAAQsVdV958t6h5whQdxQVXtZYkQrjxf3zm5Fjx2VmQFLMzhiVK6NXZWJgQJmZ3A9ULcYxh","slot":83735605},{"blockTime":1625076517,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5NMhKqZNK1WcdnyKLv59iGXaS7ruGEes2h4NBo1cMneQoiN3CqtF4CiF5arGH5nuMmkrjCjVxRSnU5irJj97cKou","slot":83735600},{"blockTime":1625076427,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5xnRayHgj9RJ1beTdZ58yPQikTo9b88xKvSJq26dBGBNPBW9oLqoTXRvUvZ1YASJq5pHRYRTBfPzPWUYroiVQMYw","slot":83735449},{"blockTime":1625076420,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2QFHUoJTWeHciqZV2bT7pLXsAsHfGiUnocxtojVL6H96EQjgvetgPdEtkohwVgUMHu929Jnq9UigG7xNrNykcSMJ","slot":83735436},{"blockTime":1625076415,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"o2uvrLGMPmAhY3Y8uyHNZuj1m8jXb37eiLUFEHWPHy2MnKzy6BRr3QHgsH7yrdBjHYXWgujra3y1wumDDWFgWJ9","slot":83735430},{"blockTime":1625076414,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"54Uh53PVJMrEBDAWXaGZybAtJR5YpAen9TcvqQPgQoPizFaqEjTfvk5C4w6D4nLaTaukEDufQsFRxe5g6MAUMD2n","slot":83735426},{"blockTime":1625076408,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4SJJHpzvQhwPG4w4rHZEXFmDQszqDYJWKGhWBTGawWeKvKv3KiRMdfiELqewDJVugG2SbMSq4XRznpAhnSABskYe","slot":83735417},{"blockTime":1625076405,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4ps21mg8hd6QxdjEWcjn5p1pUaqKXJ6A2mJzmWkWHkLn3ubd1zydSfnVcNGJFBYpYbLeYpXHVYMWoj8JjaNWbF3W","slot":83735412},{"blockTime":1625076402,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"8fHSb2who8mVjxaTDacFDgGC7FWS9ptWxmMtSVtZva6hQbGnKNZEqgBTpLPCUUEXFwWs5nGjab6RsqVNLYqcGkM","slot":83735408},{"blockTime":1625076402,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"25qu8LAxKf9DAmxYiFDahE3r4dEouvmDtUcnret2q5D8EraJLQ1jpTivGqXfAnaRiHc4EBevj8vBM17QSLiHE1zw","slot":83735406},{"blockTime":1625076396,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2wogh59NbZkuNfekk1C3gX1JjnMwvmQQtehgvY6CdhqnApqotb4ApJHB2N5NY3Bfn3zReHUJTSJCCW8hPmTtb9mi","slot":83735398},{"blockTime":1625076345,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"mcx6w62cFL7AHEDmcM8XiKGxceUR8u5BU6eb6majm71hPKv39VVQugxpJxfUTgHLd7Ggap6tus72uuBoqJsoDeA","slot":83735312},{"blockTime":1625076342,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5GoCvMJzT9iWHuYdZy9tZwrbKzrhj1vV43A2NrjyKHM73KMimZ6fiqVZ2T5eGYqcds7gcRRmKaT7Bk9zFLFMMu4c","slot":83735307},{"blockTime":1625076340,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4TxUBUYMRVNXp9S5cvyQjH6jSV79jJE6z74KuPUPD2F1ev2JjwRk7rWNCBDnzxdNupYBnbbr5jWbCcUUEGPYYqLy","slot":83735304},{"blockTime":1625076334,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3YnVUk287ve3wusgMvPd8cuWYz3Y7PjpEDZEghjxdc2dcgDMASnrdZgEidn71urjUx6dzY9SBYcDD7SD3caBHhNY","slot":83735294},{"blockTime":1625076328,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"rimzuWjV28QXou3Hp7yUSmY4J7RXqQzjEtEELAKFZXi4FaXFAGPnqUFbqutbkSWAERqro4jCdu88boD9X6NT1v9","slot":83735284},{"blockTime":1625076327,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5xxAYW5T545m9uityzzH8tHCLftVMbA1CdSxHhM3zJ6zM3URnePnPsMqQLpaJvXTSDMYhTLKzWJUgRCPNmSXS5xT","slot":83735281},{"blockTime":1625076324,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"2veJJbWLer5LjoJowt8qPMc2BVczBksC8SMFT47daptjajXaNZH9MFCheqetxhzy3hr2uiAvj4mweDGyd2mr8cK","slot":83735277},{"blockTime":1625076321,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4MSMekrScr8T4L3CbDe8XeShUUojMLW69inws8ToBCU2DnnC2cUL4uaSXnGKX3MUCR7enuKUqFKo5xLAfMSM4FpN","slot":83735272},{"blockTime":1625076319,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4HyshAQoEfcCNcGQXG2ajLKFm1W16DgPbWxGZZoX6BtS1R6UBMWKb1BbR65Nf3WaFDWUW8eXLdwGHtots8NfeWDy","slot":83735269},{"blockTime":1625076001,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3TYpN4Ce7HeiJaWFvWXPe9WWFK16XJhrvosgLFPxu79zeC5JfrGYBkyuVQgC6NURueQh2bFBBzfgMagGtSqUL189","slot":83734740},{"blockTime":1625076000,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Yngcx43VayhKhHuQxCm6fY6xEMCSoCZTE8TVuzPSy4TPph7GzcxX5TSrm462DfBp1GzWvqXvDfxKi4rSB6Lnph5","slot":83734738},{"blockTime":1625075997,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5U9uqrGrTyG3GLXcMyipQG3UxYCjgPyYunEQCkJHtMhsUvjmnVRwdFvCoFoUk3wxmQhJ6jMVBBUbDXGBnzG7GyFd","slot":83734733},{"blockTime":1625075997,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3jnZ3CDMME59YSGXHZdc3GvRpHFiatzDDpADFx1Fr6rHEfsUcuFbTRLUMHmZPPEk8E81SXWAp5KTx8cJZfaEGr98","slot":83734731},{"blockTime":1625075994,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"AVHS4gMuaKdUQTPSCXXjJ4bXKYHxaYgxkPEA31GF1rV2RKvfgW6W4Ho2p4dRXjhvSogSUQwYoPnNRsbA7HWw5Q8","slot":83734726},{"blockTime":1625075992,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5u3HZDSPcQ83S2GFw2dxriaqCo1ZdwLpssAFa8TfkefPmbCqS3psQXNKg98pQVts5PpemeACR3ZKEZGFC8QAE28F","slot":83734724},{"blockTime":1625075989,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"hrvEYRU9hWzac51jMFEHeVAsgqZqkr6Mci39NYvcGKeWjvG9awRGWo6nS1yzgBxeou9VUBc21G8pfRbKwaX2owd","slot":83734719},{"blockTime":1625075988,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4AoESqGfx7V8hLbfTeXnYoFzMLXZjxFEnhShGWmd3CKoFpyVgDJ9keVnjZ8LYFK44rFNnMC3Atjv3CiMxsKESdv1","slot":83734716},{"blockTime":1625075983,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5PqbJFNVTLR1WHk1DPpmL89qDoTuRRPWx8FRmUzC4n9Fghua62tP9tHvSQABHANZFbe2f1YQh43zJKoZYvts5K63","slot":83734710},{"blockTime":1625075940,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"63jD1uq6iQJmjUxgo3cDB9AvJGd22WhXA8xjaoJyiuUSaABzbRQi1G7CVFbxoCZZDGA9v9Gy3166rApdSswV51Xb","slot":83734637}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	limit := 10
	minContextSlot := uint64(123456)

	before := solana.MustSignatureFromBase58("qN7RF6YSJT5QpVuhPNzjL8zNZ111NKVpJtBD53cxJHinorVW6AYLVE7bYtJnh42RjFfUTSKLrHhDBaG7AtEkymr")
	until := solana.MustSignatureFromBase58("zJTw3PHXJRqpmR2bqnTChcySGET1pZTCQZebCtJbxRp3966MHttJgCgA75jwrjHRPa7mqeuWYceqxqo2jgVAtZa")
	opts := GetSignaturesForAddressOpts{
		Limit:          &limit,
		Before:         before,
		Until:          until,
		Commitment:     CommitmentFinalized,
		MinContextSlot: &minContextSlot,
	}
	out, err := client.GetSignaturesForAddressWithOpts(
		context.Background(),
		pubKey,
		&opts,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getSignaturesForAddress",
			"params": []any{
				pubkeyString,
				map[string]any{
					"commitment":     string(CommitmentFinalized),
					"before":         before.String(),
					"until":          until.String(),
					"limit":          float64(limit),
					"minContextSlot": float64(minContextSlot),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetSignaturesForAddress_WithTransactionIndex(t *testing.T) {
	responseBody := `[{"blockTime":1234567890,"confirmationStatus":"finalized","err":null,"memo":"test memo","signature":"5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW","slot":123,"transactionIndex":42}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubKey := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")
	out, err := client.GetSignaturesForAddress(context.Background(), pubKey)
	require.NoError(t, err)

	require.Len(t, out, 1)
	sig := out[0]
	assert.Equal(t, uint64(123), sig.Slot)
	assert.Nil(t, sig.Err)
	require.NotNil(t, sig.Memo)
	assert.Equal(t, "test memo", *sig.Memo)
	require.NotNil(t, sig.TransactionIndex)
	assert.Equal(t, uint32(42), *sig.TransactionIndex)
	assert.Equal(t, ConfirmationStatusType("finalized"), sig.ConfirmationStatus)
}

func TestClient_GetSignaturesForAddress_WithoutTransactionIndex(t *testing.T) {
	responseBody := `[{"blockTime":1234567890,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW","slot":123}]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubKey := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")
	out, err := client.GetSignaturesForAddress(context.Background(), pubKey)
	require.NoError(t, err)

	require.Len(t, out, 1)
	assert.Nil(t, out[0].TransactionIndex)
}

func TestClient_GetSignatureStatuses(t *testing.T) {
	responseBody := `{"context":{"slot":83999323},"value":[{"confirmationStatus":"finalized","confirmations":null,"err":null,"slot":82233105,"status":{"Ok":null}},{"confirmationStatus":"finalized","confirmations":null,"err":null,"slot":82232349,"status":{"Ok":null}}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	sig1 := solana.MustSignatureFromBase58("APPAzLobMg62AW7tdot1s7qKjya4Htt7AqjvT4uMUje8FuFNKD6qnoSk3JvBrkBnBnUyknqXJUXpj9BXENSExSQ")
	sig2 := solana.MustSignatureFromBase58("eue8eTRd4puKR2aqsW9AigzyBsF9Em4uVKWKEkMeYUuT9XevvrYwk6Ps5ApCHKEdYDYxPsmE8tb9Gik6jZM1xHT")
	out, err := client.GetSignatureStatuses(
		context.Background(),
		true,
		sig1,
		sig2,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getSignatureStatuses",
			"params": []any{
				[]any{
					sig1.String(),
					sig2.String(),
				},
				map[string]any{
					"searchTransactionHistory": true,
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetSlot(t *testing.T) {
	responseBody := `83999325`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetSlot(
		context.Background(),
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getSlot",
			"params": []any{
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetSlotLeader(t *testing.T) {
	responseBody := `"Bdd4XhquueXBB7aZXVYUn1XBdJ18G7Wx3LUe6aKkmXEV"`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetSlotLeader(
		context.Background(),
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getSlotLeader",
			"params": []any{
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetSlotLeaders(t *testing.T) {
	responseBody := `["GDoZFWJNuiQdP3DMupgBeGr6mQJYCcWuUvcrnr7xhSqj","J1mnigj2PmzRCuLvjqBX3h6Lb5b6PoPt2Cvqu8g2wNG3","J1mnigj2PmzRCuLvjqBX3h6Lb5b6PoPt2Cvqu8g2wNG3","J1mnigj2PmzRCuLvjqBX3h6Lb5b6PoPt2Cvqu8g2wNG3","J1mnigj2PmzRCuLvjqBX3h6Lb5b6PoPt2Cvqu8g2wNG3","FcWgrc99RAix3y9th526GnzN23MQSkFmyWaeo9xJ6Jfo","FcWgrc99RAix3y9th526GnzN23MQSkFmyWaeo9xJ6Jfo","FcWgrc99RAix3y9th526GnzN23MQSkFmyWaeo9xJ6Jfo","FcWgrc99RAix3y9th526GnzN23MQSkFmyWaeo9xJ6Jfo","E9bcuniYQhMscfMjE8zaAXQ47TH56gsQoKuzvqXHxnAY"]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	start := uint64(83220831)
	limit := uint64(10)
	out, err := client.GetSlotLeaders(
		context.Background(),
		start,
		limit,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getSlotLeaders",
			"params": []any{
				float64(start),
				float64(limit),
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetSupply(t *testing.T) {
	responseBody := `{"context":{"slot":83999524},"value":{"circulating":1370901328666198300,"nonCirculating":154690270000000,"nonCirculatingAccounts":["Br3aeVGapRb2xTq17RU2pYZCoJpWA7bq6TKBCcYtMSmt","AzHQ8Bia1grVVbcGyci7wzueSWkgvu7YZVZ4B9rkL5P6","GpYnVDgB7dzvwSgsjQFeHznjG6Kt1DLBFYrKxjGU1LuD","6ii8XC6KrfRcCR63cvJVhE73iCB1G44ZEaLW4WFYzy61","CoqCEzUA7KpCUxkV8ihGn9oru6imf6oVnjYKpa6jY5TC","CqqiPBWPqr3qN4gjiBQjWNT52eRFys5xdGdbQ69ywHfX","CND6ZjRTzaCFVdX7pSSWgjTfHZuhxqFDoUBqWBJguNoA","2qXZP8ZUCpvEd3VPow2zobf9S1db1vTBG3oqLWUANVNm","3s7wyR22skqVwwYRLiboJ9BYaEMsKkKqgetGZw7xtkgc","5TXdcD9Sq8UE2h6wSQj6HC7TYHZNqTdXPvmVZWFMsDzp","DQQGPtj7pphPHCLzzBuEyDDQByUcKGrsJdsH7SP3hAug","EAJJD6nDqtXcZ4DnQb19F9XEz8y8bRDHxbWbahatZNbL","DrKzW5koKSZp4mg4BdHLwr72MMXscd2kTiWgckCvvPXz","BhvLngiqqKeZ8rpxch2uGjeCiC88zzewoWPRuoxpp1aS","CVgyXrbEd1ctEuvq11QdpnCQVnPit8NLdhyqXQHLprM2","4bDVNTq2xJKK4WjKQ214DaYBh1NE5s2H1PvcoRuPdnSf","3ZrsTmNM6AkMcqFfv3ryfhQ2jMfqP64RQbqVyAaxqhrQ","E6HM7ny8AAY28Q8Za9RyrX7x1MyEdDkaXYFGUwoy4kM2","AVYpwVou2BhdLivAwLxKPALZQsY7aZNkNmGbP2fZw7RU","H3Ni7vG1CsmJZdTvxF7RkAf9UM5qk4RsohJsmPvtZNnu","Ga7HnuewhNo3htQxy6mgs2oM6WxuZpA9hJCnBhP75J8o","AG3m2bAibcY8raMt4oXEGqRHwX4FWKPPJVjZxn1LySDX","CsUqV42gVQLJwQsKyjWHqGkfHarxn9hcY4YeSjgaaeTd","5XdtyEDREHJXXW1CTtCsVjJRjBapAwK78ZquzvnNVRrV","3jnknRabs7G2V9dKhxd2KP85pNWXKXiedYnYxtySnQMs","8W58E8JVJjH1jCy5CeHJQgvwFXTyAVyesuXRZGbcSUGG","3bTGcGB9F98XxnrBNftmmm48JGfPgi5sYxDEKiCjQYk3","JCwT5Ygmq3VeBEbDjL8s8E82Ra2rP9bq45QfZE7Xyaq7","Es13uD2p64UVPFpEWfDtd6SERdoNR2XVgqBQBZcZSLqW","C7C8odR8oashR5Feyrq2tJKaXL18id1dSj2zbkDGL2C2","GdnSyH3YtwcxFvQrVVJMm1JhTS4QVX7MFsX56uJLUfiZ","CuatS6njAcfkFHnvai7zXCs7syA9bykXWsDCJEWfhjHG","6nN69B4uZuESZYxr9nrLDjmKRtjDZQXrehwkfQTKw62U","Hm9JW7of5i9dnrboS8pCUCSeoQUPh7JsP1rkbJnW7An4","GvpCiTgq9dmEeojCDBivoLoZqc4AkbUDACpqPMwYLWKh","GK2zqSsXLA2rwVZk347RYhh6jJpRsCA69FjLW93ZGi3B","F9MWFw8cnYVwsRq8Am1PGfFL3cQUZV37mbGoxZftzLjN","63DtkW7zuARcd185EmHAkfF44bDcC2SiTSEj2spLP3iA","GEWSkfWgHkpiLbeKaAnwvqnECGdRNf49at5nFccVey7c","DbF5Cmc4A8gSVaLCxurLoRZE93K164xF4Mjcqqe1xsHk","HKJgYGTTYYR2ZkfJKHbn58w676fKueQXmvbtpyvrSM3N","3euMq5VfpURASdXrHComyoovnfQDPgBKV8Wa4omQ3Qpd","6zw7em7uQdmMpuS9fGz8Nq9TLHa5YQhEKKwPjo5PwDK4","3o6xgkJ9sTmDeQWyfj3sxwon18fXJB9PV5LDc8sfgR4a","9LJrasfs648fi2uzmFqNVSrcCtz6xQaYC5E1BeyPHTJM","8DE8fqPfv1fp9DHyGyDFFaMjpopMgDeXspzoi9jpBJjC","FgnjRCqdtAhdLxNmhMN2zGdUjm364QQhPR2Z9C5d9wut","GHzNBbsKr43UeJ2wQpkGdmNqowZsv1xnLpq1bPNqAiHn","5q54XjQ7vDx4y6KphPeE97LUNiYGtP55spjvXAWPGBuf","4sxwau4mdqZ8zEJsfryXq4QFYnMJSCp3HWuZQod8WU5k","Hz9nydgN1k15wnwffKX7CSmZp4VFTnTwLXAEdomFGNXy","CWeRmXme7LmbaUWTZWFLt6FMnpzLCHaQLuR2TdgFn4Lq","8CUUMKYNGxdgYio5CLHRHyzMEhhVRMcqefgE6dLqnVRK","DE1bawNcRJB9rVm3buyMVfr8mBEoyyu73NBovf2oXJsJ","xQadXQiUTCCFhfHjvQx1hyJK6KVWr1w2fD6DT3cdwj7","7Np41oeYqPefeNQEHSv1UDhYrehxin3NStELsSKCT4K2","BuCEvc9ze8UoAQwwsQLy8d447C8sA4zeVtVpc6m5wQeS","CUageMFi49kzoDqtdU8NvQ4Bq3sbtJygjKDAXJ45nmAi","14FUT96s9swbmH7ZjpDvfEDywnAYy9zaNhv4xvezySGu","H1rt8KvXkNhQExTRfkY8r9wjZbZ8yCih6J4wQ5Fz9HGP","9huDUZfxoJ7wGMTffUE7vh1xePqef7gyrLJu9NApncqA","BUnRE27mYXN9p8H1Ay24GXhJC88q2CuwLoNU2v2CrW4W","H3EP5q7LL6XfqPmxLp8yBvDwgUHfvhvQxKxrq644K8d5","FwfaykN7ACnsEUDHANzGHqTGQZMcGnUSsahAHUqbdPrz","Fg12tB1tz8w6zJSQ4ZAGotWoCztdMJF9hqK8R11pakog","8UVjvYyoqP6sqcctTso3xpCdCfgTMiv3VRh7vraC2eJk","GNiz4Mq886bTNDT3pijGsu2gbw6it7sqrwncro45USeB","7W8FhaRLM2Hr9sZMXFwWbe4QqphkCnVvPDvjv7YbRuDj","CQDYc4ET2mbFhVpgj41gXahL6Exn5ZoPcGAzSHuYxwmE","2WWb1gRzuXDd5viZLQF7pNRR6Y7UiyeaPpaL35X6j3ve","3epceuFZLxwjCKhMdiigxconx8GDGH9HVDQZ8eqazaHA","8rT45mqpuDBR1vcnDc9kwP9DrZAXDR4ZeuKWw3u1gTGa","GhsotwFMH6XUrRLJCxcx62h7748N2Uq8mf87hUGkmPhg","Fgyh8EeYGZtbW8sS33YmNQnzx54WXPrJ5KWNPkCfWPot","3itU5ME8L6FDqtMiRoUiT1F7PwbkTtHBbW51YWD5jtjm","7cvkjYAkUYs4W8XcXsca7cBrEGFeSUjeZmKoNBvEwyri","FiWYY85b58zEEcPtxe3PuqzWPjqBJXqdwgZeqSBmT9Cn","8vqrX3H2BYLaXVintse3gorPEM4TgTwTFZNN1Fm9TdYs","FbGeZS8LiPCZiFpFwdUUeF2yxXtSsdfJoHTsVMvM8STh","3ahQgaKYVhsKq5ybdxzHDD6nAgHCZNkxrNDfGo21ykUT","EziVYi3Sv5kJWxmU77PnbrT8jmkVuqwdiFLLzZpLVEn7","Ep5Y58PaSyALPrdFxDVAdfKtVdP55vApvsWjb3jSmXsG","9hknftBZAQL4f48tWfk3bUEV5YSLcYYtDRqNmpNnhCWG","6yKHERk8rsbmJxvMpPuwPs1ct3hRiP7xaJF2tvnGU6nK","8pNBEppa1VcFAsx4Hzq9CpdXUXZjUXbvQwLX2K7QsCwb","5D5NxsNVTgXHyVziwV7mDFwVDS6voaBsyyGxUbhQrhNW","FV8c2PQfsWqXUWBaiF7TSMMim5bZ5G53PCfh7eKbaz54","nGME7HgBT6tAJN1f6YuCCngpqT5cvSTndZUVLjQ4jwA","BUjkdqUuH5Lz9XzcMcR4DdEMnFG6r8QzUMBm16Rfau96","Mc5XB47H3DKJHym5RLa9mPzWv5snERsF3KNv5AauXK8","FR84wZQy3Y3j2gWz6pgETUiUoJtreMEuWfbg6573UCj9","7Y8smnoUrYKGGuDq2uaFKVxJYhojgg7DVixHyAtGTYEV","4NEb5MLmDDFCe4S9c3DacHLTHxfNwZrbk7Kojy41541h","3zFnorNhzsF3k446HB9bwb64CByzocBWaJ5JBqgN7Cez","BRz3NM1jouNETV6SBWW7Eg1EBLM2bB1vrRyMeur3cbGZ","GpxpMVhrBBBEYbEJxdR62w3daWz444V7m6dxYDZKH77D","HCV5dGFJXRrJ3jhDYA4DCeb9TEDTwGGYXtT3wHksu2Zr","8otuo6Jc7n9ceg5ESbMnsqzsk75yPwcNK7YiDz7e5Wb5","CzAHrrrHKx9Lxf6wdCMrsZkLvk74c7J2vGv8VYPUmY6v","HbZ5FfmKWNHC7uwk6TF1hVi6TCs7dtYfdjEcuPGgzFAg","Eyr9P5XsjK2NUKNCnfu39eqpGoiLFgVAv1LSQgMZCwiQ","7xJ9CLtEAcEShw9kW2gSoZkRWL566Dg12cvgzANJwbTr","1ddE4tL2WhjUE3iWBniF9HA7Yni8GWXNu5mFW7XabUC","5PLJZLJiRR9vf7d1JCCg7UuWjtyN9nkab9uok6TqSyuP","BivdSm1m8LtgfJRLS6QPdJ3oSys4DcNmstLiviB3ZVq1","6LHVCmk59bnpeNBobFkPR2GLneqVbQ4WyFuRuSAiJgMR","5khMKAcvmsFaAhoKkdg3u5abvKsmjUQNmhTNP624WB1F","5smrYwb1Hr2T8XMnvsqccTgXxuqQs14iuE8RbHFYf2Cf","5qC7uu1gHgJ4f2c6PixtYRxkzdZWR24DWcVGQR2BpBhj","8ndGYFjav6NDXvzYcxs449Aub3AxYv4vYpk89zRDwgj7","6o5v1HC7WhBnLfRHp8mQTtCP2khdXXjhuyGyYEoy2Suy","CHmdL15akDcJgBkY6BP3hzs98Dqr6wbdDC5p8odvtSbq","EMAY24PrS6rWfvpqffFCsTsFJypeeYYmtUc26wdh3Wup","6HUwuZs3PBup79UygZwyowozDNKydP33T1dt7ViFbQQr","AsrYX4FeLXnZcrjcZmrASY2Eq1jvEeQfwxtNTxS5zojA","GumSE5HsMV5HCwBTv2D2D81yy9x17aDkvobkqAfTRgmo","AzVV9ZZDxTgW4wWfJmsG6ytaHpQGSe1yz76Nyy84VbQF","CakcnaRDHka2gXyfbEd2d3xsvkJkqsLw2akB3zsN1D2S","DUS1KxwUhUyDKB4A81E8vdnTe3hSahd92Abtn9CXsEcj"],"total":1371056018936198100}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetSupply(context.Background(), CommitmentFinalized)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getSupply",
			"params": []any{
				map[string]any{
					"commitment":                        string(CommitmentFinalized),
					"excludeNonCirculatingAccountsList": false,
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetSupply_WithOpts(t *testing.T) {
	responseBody := `{"context":{"slot":83999524},"value":{"circulating":1370901328666198300,"nonCirculating":154690270000000,"nonCirculatingAccounts":["Br3aeVGapRb2xTq17RU2pYZCoJpWA7bq6TKBCcYtMSmt","AzHQ8Bia1grVVbcGyci7wzueSWkgvu7YZVZ4B9rkL5P6","GpYnVDgB7dzvwSgsjQFeHznjG6Kt1DLBFYrKxjGU1LuD","6ii8XC6KrfRcCR63cvJVhE73iCB1G44ZEaLW4WFYzy61","CoqCEzUA7KpCUxkV8ihGn9oru6imf6oVnjYKpa6jY5TC","CqqiPBWPqr3qN4gjiBQjWNT52eRFys5xdGdbQ69ywHfX","CND6ZjRTzaCFVdX7pSSWgjTfHZuhxqFDoUBqWBJguNoA","2qXZP8ZUCpvEd3VPow2zobf9S1db1vTBG3oqLWUANVNm","3s7wyR22skqVwwYRLiboJ9BYaEMsKkKqgetGZw7xtkgc","5TXdcD9Sq8UE2h6wSQj6HC7TYHZNqTdXPvmVZWFMsDzp","DQQGPtj7pphPHCLzzBuEyDDQByUcKGrsJdsH7SP3hAug","EAJJD6nDqtXcZ4DnQb19F9XEz8y8bRDHxbWbahatZNbL","DrKzW5koKSZp4mg4BdHLwr72MMXscd2kTiWgckCvvPXz","BhvLngiqqKeZ8rpxch2uGjeCiC88zzewoWPRuoxpp1aS","CVgyXrbEd1ctEuvq11QdpnCQVnPit8NLdhyqXQHLprM2","4bDVNTq2xJKK4WjKQ214DaYBh1NE5s2H1PvcoRuPdnSf","3ZrsTmNM6AkMcqFfv3ryfhQ2jMfqP64RQbqVyAaxqhrQ","E6HM7ny8AAY28Q8Za9RyrX7x1MyEdDkaXYFGUwoy4kM2","AVYpwVou2BhdLivAwLxKPALZQsY7aZNkNmGbP2fZw7RU","H3Ni7vG1CsmJZdTvxF7RkAf9UM5qk4RsohJsmPvtZNnu","Ga7HnuewhNo3htQxy6mgs2oM6WxuZpA9hJCnBhP75J8o","AG3m2bAibcY8raMt4oXEGqRHwX4FWKPPJVjZxn1LySDX","CsUqV42gVQLJwQsKyjWHqGkfHarxn9hcY4YeSjgaaeTd","5XdtyEDREHJXXW1CTtCsVjJRjBapAwK78ZquzvnNVRrV","3jnknRabs7G2V9dKhxd2KP85pNWXKXiedYnYxtySnQMs","8W58E8JVJjH1jCy5CeHJQgvwFXTyAVyesuXRZGbcSUGG","3bTGcGB9F98XxnrBNftmmm48JGfPgi5sYxDEKiCjQYk3","JCwT5Ygmq3VeBEbDjL8s8E82Ra2rP9bq45QfZE7Xyaq7","Es13uD2p64UVPFpEWfDtd6SERdoNR2XVgqBQBZcZSLqW","C7C8odR8oashR5Feyrq2tJKaXL18id1dSj2zbkDGL2C2","GdnSyH3YtwcxFvQrVVJMm1JhTS4QVX7MFsX56uJLUfiZ","CuatS6njAcfkFHnvai7zXCs7syA9bykXWsDCJEWfhjHG","6nN69B4uZuESZYxr9nrLDjmKRtjDZQXrehwkfQTKw62U","Hm9JW7of5i9dnrboS8pCUCSeoQUPh7JsP1rkbJnW7An4","GvpCiTgq9dmEeojCDBivoLoZqc4AkbUDACpqPMwYLWKh","GK2zqSsXLA2rwVZk347RYhh6jJpRsCA69FjLW93ZGi3B","F9MWFw8cnYVwsRq8Am1PGfFL3cQUZV37mbGoxZftzLjN","63DtkW7zuARcd185EmHAkfF44bDcC2SiTSEj2spLP3iA","GEWSkfWgHkpiLbeKaAnwvqnECGdRNf49at5nFccVey7c","DbF5Cmc4A8gSVaLCxurLoRZE93K164xF4Mjcqqe1xsHk","HKJgYGTTYYR2ZkfJKHbn58w676fKueQXmvbtpyvrSM3N","3euMq5VfpURASdXrHComyoovnfQDPgBKV8Wa4omQ3Qpd","6zw7em7uQdmMpuS9fGz8Nq9TLHa5YQhEKKwPjo5PwDK4","3o6xgkJ9sTmDeQWyfj3sxwon18fXJB9PV5LDc8sfgR4a","9LJrasfs648fi2uzmFqNVSrcCtz6xQaYC5E1BeyPHTJM","8DE8fqPfv1fp9DHyGyDFFaMjpopMgDeXspzoi9jpBJjC","FgnjRCqdtAhdLxNmhMN2zGdUjm364QQhPR2Z9C5d9wut","GHzNBbsKr43UeJ2wQpkGdmNqowZsv1xnLpq1bPNqAiHn","5q54XjQ7vDx4y6KphPeE97LUNiYGtP55spjvXAWPGBuf","4sxwau4mdqZ8zEJsfryXq4QFYnMJSCp3HWuZQod8WU5k","Hz9nydgN1k15wnwffKX7CSmZp4VFTnTwLXAEdomFGNXy","CWeRmXme7LmbaUWTZWFLt6FMnpzLCHaQLuR2TdgFn4Lq","8CUUMKYNGxdgYio5CLHRHyzMEhhVRMcqefgE6dLqnVRK","DE1bawNcRJB9rVm3buyMVfr8mBEoyyu73NBovf2oXJsJ","xQadXQiUTCCFhfHjvQx1hyJK6KVWr1w2fD6DT3cdwj7","7Np41oeYqPefeNQEHSv1UDhYrehxin3NStELsSKCT4K2","BuCEvc9ze8UoAQwwsQLy8d447C8sA4zeVtVpc6m5wQeS","CUageMFi49kzoDqtdU8NvQ4Bq3sbtJygjKDAXJ45nmAi","14FUT96s9swbmH7ZjpDvfEDywnAYy9zaNhv4xvezySGu","H1rt8KvXkNhQExTRfkY8r9wjZbZ8yCih6J4wQ5Fz9HGP","9huDUZfxoJ7wGMTffUE7vh1xePqef7gyrLJu9NApncqA","BUnRE27mYXN9p8H1Ay24GXhJC88q2CuwLoNU2v2CrW4W","H3EP5q7LL6XfqPmxLp8yBvDwgUHfvhvQxKxrq644K8d5","FwfaykN7ACnsEUDHANzGHqTGQZMcGnUSsahAHUqbdPrz","Fg12tB1tz8w6zJSQ4ZAGotWoCztdMJF9hqK8R11pakog","8UVjvYyoqP6sqcctTso3xpCdCfgTMiv3VRh7vraC2eJk","GNiz4Mq886bTNDT3pijGsu2gbw6it7sqrwncro45USeB","7W8FhaRLM2Hr9sZMXFwWbe4QqphkCnVvPDvjv7YbRuDj","CQDYc4ET2mbFhVpgj41gXahL6Exn5ZoPcGAzSHuYxwmE","2WWb1gRzuXDd5viZLQF7pNRR6Y7UiyeaPpaL35X6j3ve","3epceuFZLxwjCKhMdiigxconx8GDGH9HVDQZ8eqazaHA","8rT45mqpuDBR1vcnDc9kwP9DrZAXDR4ZeuKWw3u1gTGa","GhsotwFMH6XUrRLJCxcx62h7748N2Uq8mf87hUGkmPhg","Fgyh8EeYGZtbW8sS33YmNQnzx54WXPrJ5KWNPkCfWPot","3itU5ME8L6FDqtMiRoUiT1F7PwbkTtHBbW51YWD5jtjm","7cvkjYAkUYs4W8XcXsca7cBrEGFeSUjeZmKoNBvEwyri","FiWYY85b58zEEcPtxe3PuqzWPjqBJXqdwgZeqSBmT9Cn","8vqrX3H2BYLaXVintse3gorPEM4TgTwTFZNN1Fm9TdYs","FbGeZS8LiPCZiFpFwdUUeF2yxXtSsdfJoHTsVMvM8STh","3ahQgaKYVhsKq5ybdxzHDD6nAgHCZNkxrNDfGo21ykUT","EziVYi3Sv5kJWxmU77PnbrT8jmkVuqwdiFLLzZpLVEn7","Ep5Y58PaSyALPrdFxDVAdfKtVdP55vApvsWjb3jSmXsG","9hknftBZAQL4f48tWfk3bUEV5YSLcYYtDRqNmpNnhCWG","6yKHERk8rsbmJxvMpPuwPs1ct3hRiP7xaJF2tvnGU6nK","8pNBEppa1VcFAsx4Hzq9CpdXUXZjUXbvQwLX2K7QsCwb","5D5NxsNVTgXHyVziwV7mDFwVDS6voaBsyyGxUbhQrhNW","FV8c2PQfsWqXUWBaiF7TSMMim5bZ5G53PCfh7eKbaz54","nGME7HgBT6tAJN1f6YuCCngpqT5cvSTndZUVLjQ4jwA","BUjkdqUuH5Lz9XzcMcR4DdEMnFG6r8QzUMBm16Rfau96","Mc5XB47H3DKJHym5RLa9mPzWv5snERsF3KNv5AauXK8","FR84wZQy3Y3j2gWz6pgETUiUoJtreMEuWfbg6573UCj9","7Y8smnoUrYKGGuDq2uaFKVxJYhojgg7DVixHyAtGTYEV","4NEb5MLmDDFCe4S9c3DacHLTHxfNwZrbk7Kojy41541h","3zFnorNhzsF3k446HB9bwb64CByzocBWaJ5JBqgN7Cez","BRz3NM1jouNETV6SBWW7Eg1EBLM2bB1vrRyMeur3cbGZ","GpxpMVhrBBBEYbEJxdR62w3daWz444V7m6dxYDZKH77D","HCV5dGFJXRrJ3jhDYA4DCeb9TEDTwGGYXtT3wHksu2Zr","8otuo6Jc7n9ceg5ESbMnsqzsk75yPwcNK7YiDz7e5Wb5","CzAHrrrHKx9Lxf6wdCMrsZkLvk74c7J2vGv8VYPUmY6v","HbZ5FfmKWNHC7uwk6TF1hVi6TCs7dtYfdjEcuPGgzFAg","Eyr9P5XsjK2NUKNCnfu39eqpGoiLFgVAv1LSQgMZCwiQ","7xJ9CLtEAcEShw9kW2gSoZkRWL566Dg12cvgzANJwbTr","1ddE4tL2WhjUE3iWBniF9HA7Yni8GWXNu5mFW7XabUC","5PLJZLJiRR9vf7d1JCCg7UuWjtyN9nkab9uok6TqSyuP","BivdSm1m8LtgfJRLS6QPdJ3oSys4DcNmstLiviB3ZVq1","6LHVCmk59bnpeNBobFkPR2GLneqVbQ4WyFuRuSAiJgMR","5khMKAcvmsFaAhoKkdg3u5abvKsmjUQNmhTNP624WB1F","5smrYwb1Hr2T8XMnvsqccTgXxuqQs14iuE8RbHFYf2Cf","5qC7uu1gHgJ4f2c6PixtYRxkzdZWR24DWcVGQR2BpBhj","8ndGYFjav6NDXvzYcxs449Aub3AxYv4vYpk89zRDwgj7","6o5v1HC7WhBnLfRHp8mQTtCP2khdXXjhuyGyYEoy2Suy","CHmdL15akDcJgBkY6BP3hzs98Dqr6wbdDC5p8odvtSbq","EMAY24PrS6rWfvpqffFCsTsFJypeeYYmtUc26wdh3Wup","6HUwuZs3PBup79UygZwyowozDNKydP33T1dt7ViFbQQr","AsrYX4FeLXnZcrjcZmrASY2Eq1jvEeQfwxtNTxS5zojA","GumSE5HsMV5HCwBTv2D2D81yy9x17aDkvobkqAfTRgmo","AzVV9ZZDxTgW4wWfJmsG6ytaHpQGSe1yz76Nyy84VbQF","CakcnaRDHka2gXyfbEd2d3xsvkJkqsLw2akB3zsN1D2S","DUS1KxwUhUyDKB4A81E8vdnTe3hSahd92Abtn9CXsEcj"],"total":1371056018936198100}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetSupplyWithOpts(
		context.Background(),
		&GetSupplyOpts{
			Commitment:                        CommitmentFinalized,
			ExcludeNonCirculatingAccountsList: false,
		},
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getSupply",
			"params": []any{
				map[string]any{
					"commitment":                        string(CommitmentFinalized),
					"excludeNonCirculatingAccountsList": false,
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetSupply_ExcludeNonCirculatingAccounts(t *testing.T) {
	responseBody := `{"context":{"slot":83999524},"value":{"circulating":1370901328666198300,"nonCirculating":154690270000000,
"nonCirculatingAccounts":[],"total":1371056018936198100}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetSupplyWithOpts(
		context.Background(),
		&GetSupplyOpts{
			Commitment:                        CommitmentConfirmed,
			ExcludeNonCirculatingAccountsList: true,
		},
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getSupply",
			"params": []any{
				map[string]any{
					"commitment":                        string(CommitmentConfirmed),
					"excludeNonCirculatingAccountsList": true,
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetTokenLargestAccounts(t *testing.T) {
	responseBody := `{"context":{"slot":86069724},"value":[{"address":"7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932","amount":"100","decimals":0,"uiAmount":100,"uiAmountString":"100"},{"address":"H7YZoNkQq96FX6gwy1ZqVgunXhSm7hpSPtK7orjxgQDb","amount":"0","decimals":0,"uiAmount":0,"uiAmountString":"0"},{"address":"2UjQFRQRjqorKVBCsaYYSiRnRnydXpiwgbaykwKJFCjr","amount":"0","decimals":0,"uiAmount":0,"uiAmountString":"0"},{"address":"DSBUsy1rPjjLnhagcStNmBBicuVXjSRr7bBddMU37LEp","amount":"0","decimals":0,"uiAmount":0,"uiAmountString":"0"},{"address":"BZ3a2XdfAeWHscJNEMuBbq34n2MMtLeeb4PSPcKEvCjh","amount":"0","decimals":0,"uiAmount":0,"uiAmountString":"0"}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	out, err := client.GetTokenLargestAccounts(
		context.Background(),
		pubKey,
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTokenLargestAccounts",
			"params": []any{
				pubkeyString,
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetTokenSupply(t *testing.T) {
	responseBody := `{"context":{"slot":86069939},"value":{"amount":"100","decimals":0,"uiAmount":100,"uiAmountString":"100"}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	out, err := client.GetTokenSupply(
		context.Background(),
		pubKey,
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTokenSupply",
			"params": []any{
				pubkeyString,
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetTransaction(t *testing.T) {
	responseBody := `{"blockTime":1624821990,"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[199247210749,90459349430703,1,1,1],"postTokenBalances":[],"preBalances":[199247215749,90459349430703,1,1,1],"preTokenBalances":[],"rewards":[],"status":{"Ok":null}},"slot":83311386,"transaction":{"message":{"accountKeys":["2ZZkgKcBfp4tW8qCLj2yjxRYh9CuvEVJWb6e2KKS91Mj","53R9tmVrTQwJAgaUCWEA7SiVf7eWAbaQarZ159ixt2D9","SysvarS1otHashes111111111111111111111111111","SysvarC1ock11111111111111111111111111111111","Vote111111111111111111111111111111111111111"],"header":{"numReadonlySignedAccounts":0,"numReadonlyUnsignedAccounts":3,"numRequiredSignatures":1},"instructions":[{"accounts":[1,2,3,0],"data":"3yZe7d","programIdIndex":4}],"recentBlockhash":"6o9C27iJ5rPi7wEpvQu1cFbB1WnRudtsPnbY8GvFWrgR"},"signatures":["QPzWhnwHnCwk3nj1zVCcjz1VP7EcAKouPg9Joietje3GnQTVQ5XyWxyPC3zHby8K5ahSn9SbQupauDbVRvv5DuL"]}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	tx := "KBVcTWwgEhVzwywtunhAXRKjXYYEdPcSCpuEkg484tiE3dFGzHDu9LKKH23uBMdfYt3JCPHeaVeDTZWecboyTrd"

	maxSupportedTransactionVersion := uint64(0)
	opts := GetTransactionOpts{
		Encoding:                       solana.EncodingBase64,
		Commitment:                     CommitmentFinalized,
		MaxSupportedTransactionVersion: &maxSupportedTransactionVersion,
	}
	out, err := client.GetTransaction(
		context.Background(),
		solana.MustSignatureFromBase58(tx),
		&opts,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTransaction",
			"params": []any{
				tx,
				map[string]any{
					"encoding":                       string(solana.EncodingBase64),
					"commitment":                     string(CommitmentFinalized),
					"maxSupportedTransactionVersion": float64(maxSupportedTransactionVersion),
				},
			},
		},
		reqBody,
	)

	blockTimeSeconds := int64(1624821990)
	blockTime := solana.UnixTimeSeconds(blockTimeSeconds)
	expected := &GetTransactionResult{
		Slot:      83311386,
		BlockTime: &blockTime,
		Transaction: &TransactionResultEnvelope{
			asParsedTransaction: &solana.Transaction{
				Signatures: []solana.Signature{
					solana.MustSignatureFromBase58("QPzWhnwHnCwk3nj1zVCcjz1VP7EcAKouPg9Joietje3GnQTVQ5XyWxyPC3zHby8K5ahSn9SbQupauDbVRvv5DuL"),
				},
				Message: solana.Message{
					AccountKeys: []solana.PublicKey{
						solana.MustPublicKeyFromBase58("2ZZkgKcBfp4tW8qCLj2yjxRYh9CuvEVJWb6e2KKS91Mj"),
						solana.MustPublicKeyFromBase58("53R9tmVrTQwJAgaUCWEA7SiVf7eWAbaQarZ159ixt2D9"),
						solana.MustPublicKeyFromBase58("SysvarS1otHashes111111111111111111111111111"),
						solana.MustPublicKeyFromBase58("SysvarC1ock11111111111111111111111111111111"),
						solana.MustPublicKeyFromBase58("Vote111111111111111111111111111111111111111"),
					},
					RecentBlockhash: solana.MustHashFromBase58("6o9C27iJ5rPi7wEpvQu1cFbB1WnRudtsPnbY8GvFWrgR"),
					Instructions: []solana.CompiledInstruction{
						{
							Accounts: []uint16{
								1,
								2,
								3,
								0,
							},
							Data:           solana.Base58([]byte{0x74, 0x65, 0x73, 0x74}),
							ProgramIDIndex: 4,
						},
					},
					Header: solana.MessageHeader{
						NumRequiredSignatures:       1,
						NumReadonlySignedAccounts:   0,
						NumReadonlyUnsignedAccounts: 3,
					},
				},
			},
		},
		Meta: &TransactionMeta{
			Err: nil,
			Fee: 5000,
			PreBalances: []uint64{
				199247215749,
				90459349430703,
				1,
				1,
				1,
			},
			PostBalances: []uint64{
				199247210749,
				90459349430703,
				1,
				1,
				1,
			},
			InnerInstructions: []InnerInstruction{},
			PreTokenBalances:  []TokenBalance{},
			PostTokenBalances: []TokenBalance{},
			LogMessages: []string{
				"Program Vote111111111111111111111111111111111111111 invoke [1]",
				"Program Vote111111111111111111111111111111111111111 success",
			},
			Status: DeprecatedTransactionMetaStatus{
				"Ok": nil,
			},
			Rewards: []BlockReward{},
		},
	}

	assert.Equal(t, expected, out, "both deserialized values must be equal")
}

func TestClient_GetParsedTransaction(t *testing.T) {
	responseBody := `{"blockTime":1660570006,"meta":{"err":null,"fee":10000,"innerInstructions":[{"index":2,"instructions":[{"parsed":{"info":{"account":"BMnsyyG6S6zkaE3K5X3nbRMKdvBS5dT6HhcMozBVL7Ly","amount":"47444666","authority":"7oPa2PHQdZmjSPqvpZN7MQxnC7Dcf3uL4oLqknGLk2S3","mint":"E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1"},"type":"burn"},"program":"spl-token","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},{"parsed":{"info":{"destination":"9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy","lamports":100,"source":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo"},"type":"transfer"},"program":"system","programId":"11111111111111111111111111111111"},{"accounts":["2yVjuQwpsvdsrywzsJJVs9Ueh4zayyo5DYJbBNc3DDpn","3KEmPDRc6WEvhomG8awhfv2k33HgeqfGJmE1dptFmzhR"],"data":"2Af7uakYAFq8MGzDZQhLpcgRrAP9WHnAaA61z8nFafM8rFGNsKkksFcD6dDnAebHD6LCZBXqP6iyo8mX8XnteCsiEagZSqRLbe1QTRBpzZmwtFBVwY4SLyqBMxXKX35SM7zKVA7GYiTa2UDCaDvqQ3SQdHvRNaF5AED3HcJpYC1eFGhPpSjESVZHPN2rYYZXwma","programId":"worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth"}]}],"loadedAddresses":{"readonly":[],"writable":[]},"logMessages":["Program 11111111111111111111111111111111 invoke [1]","Program 11111111111111111111111111111111 success"],"postBalances":[72226420],"postTokenBalances":[{"accountIndex":4,"mint":"E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1","owner":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","uiTokenAmount":{"amount":"0","decimals":6,"uiAmount":null,"uiAmountString":"0"}}],"preBalances":[74714380],"preTokenBalances":[{"accountIndex":4,"mint":"E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1","owner":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","uiTokenAmount":{"amount":"47444666","decimals":6,"uiAmount":47.444666,"uiAmountString":"47.444666"}}],"rewards":[],"status":{"Ok":null}},"slot":146099091,"transaction":{"message":{"accountKeys":[{"pubkey":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","signer":true,"writable":true}],"addressTableLookups":null,"instructions":[{"parsed":{"info":{"destination":"9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy","lamports":100,"source":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo"},"type":"transfer"},"program":"system","programId":"11111111111111111111111111111111"},{"parsed":{"info":{"amount":"47444666","delegate":"7oPa2PHQdZmjSPqvpZN7MQxnC7Dcf3uL4oLqknGLk2S3","owner":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","source":"BMnsyyG6S6zkaE3K5X3nbRMKdvBS5dT6HhcMozBVL7Ly"},"type":"approve"},"program":"spl-token","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},{"accounts":["G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo"],"data":"2dmnzvSCNoP8bNbUnUtk7FTYod5czhUfk4E7LSPNMtK4V1FHgQVYeQ2GnsEtCKZCyLLHXvnkReP","programId":"wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb"}],"recentBlockhash":"9L8FEB81LfZ67ejxpMaaZmC9EmXBpV38dhNaiF9UbzZi"},"signatures":["2x1QBpfcEQetAx7zETLEmvVvjue9311s9AWroEvMAboFkqaHZVp1sUpTFXroc5Q6tkPmZK5pYfmPFteoZPVRLF89"]}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	tx := "KBVcTWwgEhVzwywtunhAXRKjXYYEdPcSCpuEkg484tiE3dFGzHDu9LKKH23uBMdfYt3JCPHeaVeDTZWecboyTrd"

	opts := GetParsedTransactionOpts{
		Commitment: CommitmentFinalized,
	}
	out, err := client.GetParsedTransaction(
		context.Background(),
		solana.MustSignatureFromBase58(tx),
		&opts,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTransaction",
			"params": []any{
				tx,
				map[string]any{
					"encoding":   string(solana.EncodingJSONParsed),
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	assert.Equal(t, uint64(2), out.Meta.InnerInstructions[0].Index)
	assert.Equal(t, &InstructionInfo{
		Info: map[string]any{
			"account":   "BMnsyyG6S6zkaE3K5X3nbRMKdvBS5dT6HhcMozBVL7Ly",
			"amount":    "47444666",
			"authority": "7oPa2PHQdZmjSPqvpZN7MQxnC7Dcf3uL4oLqknGLk2S3",
			"mint":      "E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1",
		},
		InstructionType: "burn",
	}, out.Meta.InnerInstructions[0].Instructions[0].Parsed.asInstructionInfo)
	assert.Equal(t, &InstructionInfo{
		Info: map[string]any{
			"destination": "9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy",
			"lamports":    float64(100),
			"source":      "G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo",
		},
		InstructionType: "transfer",
	}, out.Transaction.Message.Instructions[0].Parsed.asInstructionInfo)
}

func TestClient_GetTransactionCount(t *testing.T) {
	responseBody := `27293302873`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetTransactionCount(
		context.Background(),
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTransactionCount",
			"params": []any{
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetVersion(t *testing.T) {
	responseBody := `{"feature-set":743297851,"solana-core":"1.7.3"}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetVersion(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getVersion",
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetVoteAccounts(t *testing.T) {
	responseBody := `{"current":[],"delinquent":[{"activatedStake":4997717120,"commission":100,"epochCredits":[[127,1124979,892885],[128,1435333,1124979],[129,1603147,1435333],[131,1739262,1603147],[132,1895556,1739262]],"epochVoteAccount":true,"lastVote":51699331,"nodePubkey":"z3roU4WgvZvYkAEAYmUGK4LkPK6qFii6uzgMAswGYjb","rootSlot":51699288,"votePubkey":"vot33MHDqT6nSwubGzqtc6m16ChcUywxV7tNULF19Vu"}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	opts := &GetVoteAccountsOpts{
		VotePubkey: solana.MustPublicKeyFromBase58("vot33MHDqT6nSwubGzqtc6m16ChcUywxV7tNULF19Vu").ToPointer(),
		Commitment: CommitmentFinalized,
	}
	out, err := client.GetVoteAccounts(
		context.Background(),
		opts,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getVoteAccounts",
			"params": []any{
				map[string]any{
					"votePubkey": opts.VotePubkey.String(),
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetVoteAccounts_WithCommissionBps(t *testing.T) {
	responseBody := `{"current":[{"activatedStake":5000000000,"commission":7,"inflationRewardsCommissionBps":700,"epochCredits":[[127,1124979,892885]],"epochVoteAccount":true,"lastVote":51699331,"nodePubkey":"z3roU4WgvZvYkAEAYmUGK4LkPK6qFii6uzgMAswGYjb","rootSlot":51699288,"votePubkey":"vot33MHDqT6nSwubGzqtc6m16ChcUywxV7tNULF19Vu"}],"delinquent":[]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetVoteAccounts(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, out.Current, 1)
	va := out.Current[0]
	assert.Equal(t, uint8(7), va.Commission)
	require.NotNil(t, va.InflationRewardsCommissionBps)
	assert.Equal(t, uint16(700), *va.InflationRewardsCommissionBps)
}

func TestClient_GetVoteAccounts_WithoutCommissionBps(t *testing.T) {
	// Backward compatibility: pre-SIMD-0291 responses without inflationRewardsCommissionBps
	responseBody := `{"current":[{"activatedStake":5000000000,"commission":7,"epochCredits":[[127,1124979,892885]],"epochVoteAccount":true,"lastVote":51699331,"nodePubkey":"z3roU4WgvZvYkAEAYmUGK4LkPK6qFii6uzgMAswGYjb","rootSlot":51699288,"votePubkey":"vot33MHDqT6nSwubGzqtc6m16ChcUywxV7tNULF19Vu"}],"delinquent":[]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetVoteAccounts(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, out.Current, 1)
	assert.Nil(t, out.Current[0].InflationRewardsCommissionBps)
}

func TestClient_MinimumLedgerSlot(t *testing.T) {
	responseBody := `83686753`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.MinimumLedgerSlot(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "minimumLedgerSlot",
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_RequestAirdrop(t *testing.T) {
	responseBody := `"3ZmWDnFJ5REjxtmtQRrczmVDraVZs7BpUFo3NRfnoQs6wvTJ2kTkw9YyGod291UHjK5Qg6w63Hqn7t6nrGMLWhga"`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	lamports := uint64(10000000)
	out, err := client.RequestAirdrop(
		context.Background(),
		pubKey,
		lamports,
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "requestAirdrop",
			"params": []any{
				pubkeyString,
				float64(lamports),
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_RequestAirdropWithOpts(t *testing.T) {
	responseBody := `"3ZmWDnFJ5REjxtmtQRrczmVDraVZs7BpUFo3NRfnoQs6wvTJ2kTkw9YyGod291UHjK5Qg6w63Hqn7t6nrGMLWhga"`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	blockhash := solana.MustHashFromBase58("EkSnNWid2cvwEVnVx9aBqawnmiCNiDgp3gUdkDPTKN1N")
	lamports := uint64(10000000)
	out, err := client.RequestAirdropWithOpts(
		context.Background(),
		pubKey,
		lamports,
		&RequestAirdropOpts{
			Commitment:      CommitmentFinalized,
			RecentBlockhash: &blockhash,
		},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "requestAirdrop",
			"params": []any{
				pubkeyString,
				float64(lamports),
				map[string]any{
					"commitment":      string(CommitmentFinalized),
					"recentBlockhash": blockhash.String(),
				},
			},
		},
		reqBody,
	)

	assert.Equal(t,
		solana.MustSignatureFromBase58("3ZmWDnFJ5REjxtmtQRrczmVDraVZs7BpUFo3NRfnoQs6wvTJ2kTkw9YyGod291UHjK5Qg6w63Hqn7t6nrGMLWhga"),
		out,
	)
}

func TestClient_RequestAirdropWithOpts_OnlyCommitment(t *testing.T) {
	responseBody := `"3ZmWDnFJ5REjxtmtQRrczmVDraVZs7BpUFo3NRfnoQs6wvTJ2kTkw9YyGod291UHjK5Qg6w63Hqn7t6nrGMLWhga"`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	lamports := uint64(10000000)
	_, err := client.RequestAirdropWithOpts(
		context.Background(),
		pubKey,
		lamports,
		&RequestAirdropOpts{
			Commitment: CommitmentFinalized,
		},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "requestAirdrop",
			"params": []any{
				pubkeyString,
				float64(lamports),
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)
}

func TestClient_GetTokenAccountBalance(t *testing.T) {
	responseBody := `{"context":{"slot":1114},"value":{"amount":"9864","decimals":2,"uiAmount":98.64,"uiAmountString":"98.64"}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	out, err := client.GetTokenAccountBalance(
		context.Background(),
		pubKey,
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTokenAccountBalance",
			"params": []any{
				pubkeyString,
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetTokenAccountsByDelegate(t *testing.T) {
	responseBody := `{"context":{"slot":1114},"value":[{"account":{"data":{"program":"spl-token","parsed":{"accountType":"account","info":{"tokenAmount":{"amount":"1","decimals":1,"uiAmount":0.1,"uiAmountString":"0.1"},"delegate":"4Nd1mBQtrMJVYVfKf2PJy9NZUZdTAsp7D4xWLs4gDB4T","delegatedAmount":1,"isInitialized":true,"isNative":false,"mint":"3wyAj7Rt1TWVPZVteFJPLa26JmLvdb1CAKEFZm3NY75E","owner":"CnPoSPKXu7wJqxe59Fs72tkBeALovhsCxYeFwPCQH9TD"}}},"executable":false,"lamports":1726080,"owner":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","rentEpoch":4,"space":0},"pubkey":"CnPoSPKXu7wJqxe59Fs72tkBeALovhsCxYeFwPCQH9TD"}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	programIDString := "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	programID := solana.MustPublicKeyFromBase58(programIDString)

	out, err := client.GetTokenAccountsByDelegate(
		context.Background(),
		pubKey,
		&GetTokenAccountsConfig{
			ProgramId: &programID,
		},
		&GetTokenAccountsOpts{
			Commitment: CommitmentFinalized,
			Encoding:   solana.EncodingJSONParsed,
		},
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTokenAccountsByDelegate",
			"params": []any{
				pubkeyString,
				map[string]any{
					"programId": string(programIDString),
				},
				map[string]any{
					"commitment": string(CommitmentFinalized),
					"encoding":   string(solana.EncodingJSONParsed),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetTokenAccountsByOwner(t *testing.T) {
	responseBody := `{"context":{"slot":1114},"value":[{"account":{"data":{"program":"spl-token","parsed":{"accountType":"account","info":{"tokenAmount":{"amount":"1","decimals":1,"uiAmount":0.1,"uiAmountString":"0.1"},"delegate":null,"delegatedAmount":1,"isInitialized":true,"isNative":false,"mint":"3wyAj7Rt1TWVPZVteFJPLa26JmLvdb1CAKEFZm3NY75E","owner":"4Qkev8aNZcqFNSRhQzwyLMFSsi94jHqE8WNVTJzTP99F"}}},"executable":false,"lamports":1726080,"owner":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","rentEpoch":4,"space":0},"pubkey":"CnPoSPKXu7wJqxe59Fs72tkBeALovhsCxYeFwPCQH9TD"}]}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	programIDString := "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	programID := solana.MustPublicKeyFromBase58(programIDString)

	out, err := client.GetTokenAccountsByOwner(
		context.Background(),
		pubKey,
		&GetTokenAccountsConfig{
			ProgramId: &programID,
		},
		&GetTokenAccountsOpts{
			Commitment: CommitmentFinalized,
			Encoding:   solana.EncodingJSONParsed,
		},
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTokenAccountsByOwner",
			"params": []any{
				pubkeyString,
				map[string]any{
					"programId": string(programIDString),
				},
				map[string]any{
					"commitment": string(CommitmentFinalized),
					"encoding":   string(solana.EncodingJSONParsed),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

var (
	encodedTx         string = "AfjEs3XhTc3hrxEvlnMPkm/cocvAUbFNbCl00qKnrFue6J53AhEqIFmcJJlJW3EDP5RmcMz+cNTTcZHW/WJYwAcBAAEDO8hh4VddzfcO5jbCt95jryl6y8ff65UcgukHNLWH+UQGgxCGGpgyfQVQV02EQYqm4QwzUt2qf9f1gVLM7rI4hwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA6ANIF55zOZWROWRkeh+lExxZBnKFqbvIxZDLE7EijjoBAgIAAQwCAAAAOTAAAAAAAAA="
	txSignatureString string = "5yUSwqQqeZLEEYKxnG4JC4XhaaBpV3RS4nQbK8bQTyjLX5btVq9A1Ja5nuJzV7Z3Zq8G6EVKFvN4DKUL6PSAxmTk"
)

func TestClient_SendTransaction(t *testing.T) {
	responseBody := fmt.Sprintf(`"%s"`, txSignatureString)
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()

	data, err := base64.StdEncoding.DecodeString(encodedTx)
	require.NoError(t, err)

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(data))
	require.NoError(t, err)

	client := New(server.URL)

	out, err := client.SendTransaction(context.Background(), tx)
	require.NoError(t, err)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_SendEncodedTransaction(t *testing.T) {
	responseBody := fmt.Sprintf(`"%s"`, txSignatureString)
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()

	client := New(server.URL)

	out, err := client.SendEncodedTransaction(context.Background(), encodedTx)
	require.NoError(t, err)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_SendRawTransaction(t *testing.T) {
	responseBody := fmt.Sprintf(`"%s"`, txSignatureString)
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()

	client := New(server.URL)

	rawTx, err := base64.StdEncoding.DecodeString(encodedTx)
	require.NoError(t, err)

	out, err := client.SendRawTransaction(context.Background(), rawTx)
	require.NoError(t, err)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_IsBlockhashValid(t *testing.T) {
	responseBody := `{"context":{"slot":100688709},"value":true}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()

	client := New(server.URL)

	blockhashString := "dv4ACNkpYPcE3aKmYDqZm9G5EB3J4MRoeE7WNDRBVJB"
	blockhash := solana.MustHashFromBase58(blockhashString)
	out, err := client.IsBlockhashValid(
		context.Background(),
		blockhash,
		CommitmentFinalized,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "isBlockhashValid",
			"params": []any{
				blockhashString,
				map[string]any{
					"commitment": string(CommitmentFinalized),
				},
			},
		},
		reqBody,
	)

	assert.Equal(t,
		&IsValidBlockhashResult{
			RPCContext: RPCContext{
				Context{Slot: 100688709},
			},
			Value: true,
		}, out)
}

func TestClient_SimulateTransaction(t *testing.T) {
	responseBody := `{"context":{"slot":218},"value":{"accounts":null,"logs":["Program 83astBRguLMdt2h5U1Tpdq5tjFoJ6noeGwaY3mDLVcri invoke [1]","Program 83astBRguLMdt2h5U1Tpdq5tjFoJ6noeGwaY3mDLVcri consumed 2366 of 1400000 compute units","Program return: 83astBRguLMdt2h5U1Tpdq5tjFoJ6noeGwaY3mDLVcri KgAAAAAAAAA=","Program 83astBRguLMdt2h5U1Tpdq5tjFoJ6noeGwaY3mDLVcri success"],"unitsConsumed":2366}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	txData := []byte{1, 2, 3, 4} // dummy transaction data
	out, err := client.SimulateRawTransactionWithOpts(
		context.Background(),
		txData,
		nil,
	)
	require.NoError(t, err)

	assert.Nil(t, out.Value.Err)
	assert.Len(t, out.Value.Logs, 4)
	assert.Equal(t, uint64(2366), *out.Value.UnitsConsumed)

	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "simulateTransaction",
			"params": []any{
				base64.StdEncoding.EncodeToString(txData),
				map[string]any{
					"encoding": "base64",
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))
	got := mustJSONToInterface(mustAnyToJSON(out))
	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_SimulateTransactionWithOpts_AllOptions(t *testing.T) {
	responseBody := `{"context":{"slot":218},"value":{"err":null,"logs":["Program log: hello"],"accounts":null,"unitsConsumed":1000}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	txData := []byte{1, 2, 3, 4}
	minContextSlot := uint64(100)
	out, err := client.SimulateRawTransactionWithOpts(
		context.Background(),
		txData,
		&SimulateTransactionOpts{
			SigVerify:              true,
			Commitment:             CommitmentProcessed,
			ReplaceRecentBlockhash: true,
			InnerInstructions:      true,
			MinContextSlot:         &minContextSlot,
			Accounts: &SimulateTransactionAccountsOpts{
				Encoding:  solana.EncodingBase64,
				Addresses: []solana.PublicKey{solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")},
			},
		},
	)
	require.NoError(t, err)
	assert.NotNil(t, out)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "simulateTransaction",
			"params": []any{
				base64.StdEncoding.EncodeToString(txData),
				map[string]any{
					"encoding":               "base64",
					"sigVerify":              true,
					"commitment":             string(CommitmentProcessed),
					"replaceRecentBlockhash": true,
					"innerInstructions":      true,
					"minContextSlot":         float64(100),
					"accounts": map[string]any{
						"encoding":  string(solana.EncodingBase64),
						"addresses": []any{"7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"},
					},
				},
			},
		},
		reqBody,
	)
}

func TestClient_SimulateTransaction_InnerInstructions(t *testing.T) {
	responseBody := `{"context":{"slot":300},"value":{"logs":["Program log: invoke"],"accounts":null,"unitsConsumed":5000,"innerInstructions":[{"index":0,"instructions":[{"programIdIndex":2,"accounts":[0,1],"data":"3Bxs4ThwQbE4vyj5","stackHeight":2}]}],"returnData":{"programId":"83astBRguLMdt2h5U1Tpdq5tjFoJ6noeGwaY3mDLVcri","data":["KgAAAAAAAAA=","base64"]}}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	txData := []byte{1, 2, 3, 4}
	out, err := client.SimulateRawTransactionWithOpts(
		context.Background(),
		txData,
		&SimulateTransactionOpts{
			InnerInstructions: true,
		},
	)
	require.NoError(t, err)

	require.Len(t, out.Value.InnerInstructions, 1)
	assert.Equal(t, uint16(0), out.Value.InnerInstructions[0].Index)
	require.Len(t, out.Value.InnerInstructions[0].Instructions, 1)
	assert.Equal(t, uint16(2), out.Value.InnerInstructions[0].Instructions[0].ProgramIDIndex)
	assert.Equal(t, []uint16{0, 1}, out.Value.InnerInstructions[0].Instructions[0].Accounts)
	assert.Equal(t, uint16(2), out.Value.InnerInstructions[0].Instructions[0].StackHeight)

	require.NotNil(t, out.Value.ReturnData)
	assert.Equal(t, solana.MustPublicKeyFromBase58("83astBRguLMdt2h5U1Tpdq5tjFoJ6noeGwaY3mDLVcri"), out.Value.ReturnData.ProgramId)

	expected := mustJSONToInterface([]byte(responseBody))
	got := mustJSONToInterface(mustAnyToJSON(out))
	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_SimulateTransaction_FullResult(t *testing.T) {
	responseBody := `{"context":{"slot":400},"value":{"logs":["Program log: ok"],"accounts":null,"unitsConsumed":3000,"loadedAccountsDataSize":1024,"fee":5000,"preBalances":[10000000,0],"postBalances":[9995000,0],"loadedAddresses":{"readonly":["11111111111111111111111111111111"],"writable":[]},"replacementBlockhash":{"blockhash":"EETubP5AKHgjPAhzPkToc6S4eibc4FFqQGnHR1Sh9rAr","lastValidBlockHeight":500}}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	txData := []byte{1, 2, 3, 4}
	out, err := client.SimulateRawTransactionWithOpts(
		context.Background(),
		txData,
		&SimulateTransactionOpts{
			ReplaceRecentBlockhash: true,
			InnerInstructions:      true,
		},
	)
	require.NoError(t, err)

	assert.Nil(t, out.Value.Err)
	assert.Equal(t, uint64(3000), *out.Value.UnitsConsumed)
	assert.Equal(t, uint32(1024), *out.Value.LoadedAccountsDataSize)
	assert.Equal(t, uint64(5000), *out.Value.Fee)
	assert.Equal(t, []uint64{10000000, 0}, out.Value.PreBalances)
	assert.Equal(t, []uint64{9995000, 0}, out.Value.PostBalances)
	require.NotNil(t, out.Value.ReplacementBlockhash)
	assert.Equal(t, uint64(500), out.Value.ReplacementBlockhash.LastValidBlockHeight)
	require.NotNil(t, out.Value.LoadedAddresses)
	assert.Len(t, out.Value.LoadedAddresses.ReadOnly, 1)

	expected := mustJSONToInterface([]byte(responseBody))
	got := mustJSONToInterface(mustAnyToJSON(out))
	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetFeeForMessage(t *testing.T) {
	responseBody := `{"context":{"slot":5068},"value":5000}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetFeeForMessage(
		context.Background(),
		"AQABAgIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEBAQAA",
		CommitmentProcessed,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getFeeForMessage",
			"params": []any{
				"AQABAgIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEBAQAA",
				map[string]any{
					"commitment": string(CommitmentProcessed),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetHighestSnapshotSlot(t *testing.T) {
	responseBody := `{"full":100,"incremental":110}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetHighestSnapshotSlot(
		context.Background(),
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getHighestSnapshotSlot",
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetLatestBlockhash(t *testing.T) {
	responseBody := `{"context":{"slot":2792},"value":{"blockhash":"EkSnNWid2cvwEVnVx9aBqawnmiCNiDgp3gUdkDPTKN1N","lastValidBlockHeight":3090}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetLatestBlockhash(
		context.Background(),
		CommitmentProcessed,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getLatestBlockhash",
			"params": []any{
				map[string]any{
					"commitment": string(CommitmentProcessed),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}

func TestClient_GetRecentPrioritizationFees(t *testing.T) {
	responseBody := `[ { "slot": 348125, "prioritizationFee": 0 }, { "slot": 348126, "prioritizationFee": 1000 }, { "slot": 348127, "prioritizationFee": 500 } ]`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()

	client := New(server.URL)

	accounts := []solana.PublicKey{
		solana.MustPublicKeyFromBase58("41twqNJmPHv8a5AW32if2CcGRcPzaetwErXaNggGWu1q"),
		solana.MustPublicKeyFromBase58("5U3bH5b6XtG99aVWLqwVzYPVpQiFHytBD68Rz2eFPZd7"),
	}

	out, err := client.GetRecentPrioritizationFees(
		context.Background(),
		accounts,
	)
	require.NoError(t, err)

	// the ID is random, so we can't assert it; let's check that it is set, and then remove it
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getRecentPrioritizationFees",
			"params": []any{
				[]any{
					accounts[0].String(),
					accounts[1].String(),
				},
			},
		},
		reqBody,
	)

	expected := mustJSONToInterface([]byte(responseBody))

	got := mustJSONToInterface(mustAnyToJSON(out))

	assert.Equal(t, expected, got, "both deserialized values must be equal")
}
