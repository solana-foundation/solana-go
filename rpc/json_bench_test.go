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

package rpc

import (
	"bytes"
	"context"
	stdjson "encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

// These benchmarks compare encoding/json (stdlib), json-iterator/go (jsoniter),
// and goccy/go-json on real JSON-RPC responses that use `encoding: jsonParsed`.
//
// Fixtures are fetched once from a public Solana RPC endpoint and cached under
// rpc/testdata/bench/. Subsequent runs reuse the cache. To refresh, delete the
// cached files or set BENCH_REFRESH=1.
//
// Run with:
//   go test -bench=BenchmarkJSONParsed -benchmem -benchtime=2s ./rpc/
//
// Optional env:
//   BENCH_RPC=<url>       public RPC endpoint (default: mainnet-beta)
//   BENCH_REFRESH=1       force re-fetch of cached fixtures
//   BENCH_SKIP_FETCH=1    skip benchmarks if fixture missing instead of fetching

var jsoniterCompat = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	benchFixturesDir = "testdata/bench"
	defaultBenchRPC  = "https://api.mainnet-beta.solana.com"

	// Well-known public accounts/signatures stable enough for a benchmark fixture.
	// USDC mint (SPL token mint, parsed by spl-token parser).
	usdcMint = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	// Serum vault wallet with a large USDC balance (stable, long-standing account).
	serumVaultUSDC = "8BnEgHoWFysVcuFFX7QztDmzuH8r5ZFvyP3sYwn1XTh6"
	// Marinade state account (complex program-parsed account if jsonParsed falls
	// back to base64 this is still realistic binary data we need to decode).
	marinadeState = "8szGkuLTAux9XMgZ2vtY39jVSowEcpBfFfD8hXSEqdGC"
)

type benchFixture struct {
	name  string
	build func() ([]byte, error) // builds a JSON-RPC request body
}

func benchFixtures() []benchFixture {
	return []benchFixture{
		{
			name: "getAccountInfo_usdc_mint_jsonParsed",
			build: func() ([]byte, error) {
				return gojson.Marshal(map[string]any{
					"jsonrpc": "2.0",
					"id":      1,
					"method":  "getAccountInfo",
					"params": []any{
						usdcMint,
						map[string]any{"encoding": "jsonParsed"},
					},
				})
			},
		},
		{
			name: "getAccountInfo_token_account_jsonParsed",
			build: func() ([]byte, error) {
				return gojson.Marshal(map[string]any{
					"jsonrpc": "2.0",
					"id":      1,
					"method":  "getAccountInfo",
					"params": []any{
						serumVaultUSDC,
						map[string]any{"encoding": "jsonParsed"},
					},
				})
			},
		},
		{
			name: "getTokenAccountsByOwner_jsonParsed",
			build: func() ([]byte, error) {
				return gojson.Marshal(map[string]any{
					"jsonrpc": "2.0",
					"id":      1,
					"method":  "getTokenAccountsByOwner",
					"params": []any{
						marinadeState,
						map[string]any{"programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},
						map[string]any{"encoding": "jsonParsed"},
					},
				})
			},
		},
		{
			name: "getBlock_full_jsonParsed",
			build: func() ([]byte, error) {
				// Use a slot that is well past finalization (recent slot - 200)
				// when the fixture is populated; cached afterwards.
				slot, err := fetchRecentFinalizedSlot()
				if err != nil {
					return nil, err
				}
				return gojson.Marshal(map[string]any{
					"jsonrpc": "2.0",
					"id":      1,
					"method":  "getBlock",
					"params": []any{
						slot,
						map[string]any{
							"encoding":                       "jsonParsed",
							"maxSupportedTransactionVersion": 0,
							"rewards":                        false,
							"transactionDetails":             "full",
						},
					},
				})
			},
		},
	}
}

func fetchRecentFinalizedSlot() (uint64, error) {
	body, _ := gojson.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getSlot",
		"params":  []any{map[string]any{"commitment": "finalized"}},
	})
	raw, err := doRPC(body)
	if err != nil {
		return 0, err
	}
	var env struct {
		Result uint64 `json:"result"`
	}
	if err := gojson.Unmarshal(raw, &env); err != nil {
		return 0, err
	}
	// Back off a couple hundred slots so the block is always available on a
	// vanilla RPC node.
	return env.Result - 200, nil
}

func rpcEndpoint() string {
	if v := os.Getenv("BENCH_RPC"); v != "" {
		return v
	}
	return defaultBenchRPC
}

func doRPC(body []byte) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", rpcEndpoint(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// loadFixture returns the raw JSON-RPC response body for the given fixture.
// Cached on disk under testdata/bench/<name>.json.
func loadFixture(tb testing.TB, fx benchFixture) []byte {
	tb.Helper()
	path := filepath.Join(benchFixturesDir, fx.name+".json")

	refresh := os.Getenv("BENCH_REFRESH") == "1"
	if !refresh {
		if data, err := os.ReadFile(path); err == nil {
			return data
		}
	}
	if os.Getenv("BENCH_SKIP_FETCH") == "1" {
		tb.Skipf("fixture %s missing and BENCH_SKIP_FETCH=1", path)
	}

	reqBody, err := fx.build()
	if err != nil {
		tb.Fatalf("build request %s: %v", fx.name, err)
	}
	raw, err := doRPC(reqBody)
	if err != nil {
		tb.Skipf("fetch fixture %s: %v (set BENCH_RPC to a working endpoint)", fx.name, err)
	}
	// Validate the payload has a `result` field; some public endpoints
	// occasionally rate-limit and return an error envelope.
	var probe struct {
		Result stdjson.RawMessage `json:"result"`
		Error  stdjson.RawMessage `json:"error"`
	}
	if err := gojson.Unmarshal(raw, &probe); err != nil {
		tb.Skipf("fixture %s: unparsable RPC response: %v", fx.name, err)
	}
	if len(probe.Result) == 0 && len(probe.Error) > 0 {
		tb.Skipf("fixture %s: RPC returned error: %s", fx.name, string(probe.Error))
	}

	if err := os.MkdirAll(benchFixturesDir, 0o755); err != nil {
		tb.Fatalf("mkdir %s: %v", benchFixturesDir, err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		tb.Fatalf("write fixture %s: %v", path, err)
	}
	return raw
}

// extractResult pulls the `result` field out of a JSON-RPC envelope so we can
// benchmark decoding into typed RPC result shapes rather than a generic any.
func extractResult(tb testing.TB, raw []byte) []byte {
	tb.Helper()
	var env struct {
		Result stdjson.RawMessage `json:"result"`
	}
	if err := gojson.Unmarshal(raw, &env); err != nil {
		tb.Fatalf("extract result: %v", err)
	}
	return env.Result
}

// -----------------------------------------------------------------------------
// Generic Unmarshal benchmarks: decode the full RPC envelope into map[string]any.
// This measures the pure JSON parser throughput on realistic payloads without
// introducing type-specific decode overhead.
// -----------------------------------------------------------------------------

func benchUnmarshalGeneric(b *testing.B, data []byte, unmarshal func([]byte, any) error) {
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out map[string]any
		if err := unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONParsed_Generic_Unmarshal(b *testing.B) {
	for _, fx := range benchFixtures() {
		data := loadFixture(b, fx)
		b.Run(fx.name+"/stdlib", func(b *testing.B) {
			benchUnmarshalGeneric(b, data, stdjson.Unmarshal)
		})
		b.Run(fx.name+"/jsoniter", func(b *testing.B) {
			benchUnmarshalGeneric(b, data, jsoniterCompat.Unmarshal)
		})
		b.Run(fx.name+"/goccy", func(b *testing.B) {
			benchUnmarshalGeneric(b, data, gojson.Unmarshal)
		})
	}
}

// -----------------------------------------------------------------------------
// Typed Unmarshal benchmarks: decode the RPC `result` into the solana-go
// result structs (GetAccountInfoResult, GetBlockResult, etc.). This is closer
// to how real client code uses the library.
// -----------------------------------------------------------------------------

type typedTarget struct {
	name string
	new  func() any
}

func typedTargets() map[string]typedTarget {
	return map[string]typedTarget{
		"getAccountInfo_usdc_mint_jsonParsed": {
			name: "GetAccountInfoResult",
			new:  func() any { return &GetAccountInfoResult{} },
		},
		"getAccountInfo_token_account_jsonParsed": {
			name: "GetAccountInfoResult",
			new:  func() any { return &GetAccountInfoResult{} },
		},
		"getTokenAccountsByOwner_jsonParsed": {
			name: "GetTokenAccountsResult",
			new:  func() any { return &GetTokenAccountsResult{} },
		},
		// getBlock_full_jsonParsed is intentionally omitted from the typed
		// decode: the rpc.GetBlockResult struct models compiled instructions
		// (`accounts []uint16`), whereas jsonParsed blocks inline pubkey
		// strings, so the decode fails for all libraries. The block is still
		// covered by the Generic Unmarshal/Marshal benchmarks above.
	}
}

func benchUnmarshalTyped(b *testing.B, data []byte, newTarget func() any, unmarshal func([]byte, any) error) {
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := newTarget()
		if err := unmarshal(data, out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONParsed_Typed_Unmarshal(b *testing.B) {
	tgts := typedTargets()
	for _, fx := range benchFixtures() {
		tgt, ok := tgts[fx.name]
		if !ok {
			continue
		}
		raw := loadFixture(b, fx)
		result := extractResult(b, raw)
		name := fmt.Sprintf("%s->%s", fx.name, tgt.name)
		b.Run(name+"/stdlib", func(b *testing.B) {
			benchUnmarshalTyped(b, result, tgt.new, stdjson.Unmarshal)
		})
		b.Run(name+"/jsoniter", func(b *testing.B) {
			benchUnmarshalTyped(b, result, tgt.new, jsoniterCompat.Unmarshal)
		})
		b.Run(name+"/goccy", func(b *testing.B) {
			benchUnmarshalTyped(b, result, tgt.new, gojson.Unmarshal)
		})
	}
}

// -----------------------------------------------------------------------------
// Marshal benchmarks: re-encode a decoded generic JSON tree. Measures encode
// throughput for the same realistic shapes.
// -----------------------------------------------------------------------------

func benchMarshal(b *testing.B, in any, marshal func(any) ([]byte, error)) {
	out, err := marshal(in)
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(out)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := marshal(in); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONParsed_Generic_Marshal(b *testing.B) {
	for _, fx := range benchFixtures() {
		data := loadFixture(b, fx)
		var decoded map[string]any
		if err := gojson.Unmarshal(data, &decoded); err != nil {
			b.Fatalf("prep decode %s: %v", fx.name, err)
		}
		b.Run(fx.name+"/stdlib", func(b *testing.B) {
			benchMarshal(b, decoded, stdjson.Marshal)
		})
		b.Run(fx.name+"/jsoniter", func(b *testing.B) {
			benchMarshal(b, decoded, jsoniterCompat.Marshal)
		})
		b.Run(fx.name+"/goccy", func(b *testing.B) {
			benchMarshal(b, decoded, gojson.Marshal)
		})
	}
}
