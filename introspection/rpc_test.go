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

package introspection_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/introspection"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixtures are getTransaction results captured from mainnet, see
// testdata/README.md.
var fixtures = []struct {
	name      string
	version   solana.MessageVersion
	encodings []string // without jsonParsed
	outer     int
	innerOf   int // the top-level instruction with inner instructions, or -1
	inner     int
}{
	{"sync_native", solana.MessageVersionLegacy, []string{"json", "base64", "base58"}, 2, -1, 0},
	{"swap_v0", solana.MessageVersionV0, []string{"json", "base64", "base58"}, 4, 2, 8},
	{"failed_v0", solana.MessageVersionV0, []string{"json", "base64", "base58"}, 6, 3, 8},
	// The RPC serves no base58 for v1 transactions.
	{"swap_v1", solana.MessageVersionV1, []string{"json", "base64"}, 2, 1, 4},
}

func fixturePath(name, encoding string) string {
	return filepath.Join("testdata", name+"."+encoding+".json")
}

func readFixture(t testing.TB, name, encoding string) []byte {
	t.Helper()
	data, err := os.ReadFile(fixturePath(name, encoding))
	require.NoError(t, err)
	return data
}

// readResult reads a saved getTransaction result. The examples use it too.
func readResult(path string) (*rpc.GetTransactionResult, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var res rpc.GetTransactionResult
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &res, nil
}

func getTransactionResult(t testing.TB, name, encoding string) *rpc.GetTransactionResult {
	t.Helper()
	res, err := readResult(fixturePath(name, encoding))
	require.NoError(t, err)
	return res
}

func traces(ixs []*introspection.Instruction) []string {
	out := make([]string, len(ixs))
	for i, ix := range ixs {
		out[i] = ix.Trace.String()
	}
	return out
}

func TestWalkGetTransactionResult_Fixtures(t *testing.T) {
	for _, fx := range fixtures {
		t.Run(fx.name, func(t *testing.T) {
			// Explorer order: the inner instructions follow the top-level
			// instruction that emitted them.
			var wantTraces []string
			for i := range fx.outer {
				wantTraces = append(wantTraces, fmt.Sprintf("outer[%d]", i))
				if i == fx.innerOf {
					for j := range fx.inner {
						wantTraces = append(wantTraces, fmt.Sprintf("inner[%d/%d]", i, j))
					}
				}
			}

			var reference []*introspection.Instruction
			for _, encoding := range fx.encodings {
				res := getTransactionResult(t, fx.name, encoding)
				tx, err := res.Transaction.GetTransaction()
				require.NoError(t, err, encoding)
				require.Equal(t, fx.version, tx.Message.GetVersion(), encoding)

				got, err := introspection.WalkGetTransactionResult(res)
				require.NoError(t, err, encoding)
				require.Equal(t, wantTraces, traces(got), encoding)
				if reference == nil {
					reference = got
					continue
				}
				// Every encoding gives the same walk.
				require.Equal(t, reference, got, encoding)
			}
		})
	}
}

func TestWalkGetTransactionResult_FailedTransaction(t *testing.T) {
	res := getTransactionResult(t, "failed_v0", "base64")
	require.NotNil(t, res.Meta.Err)

	got, err := introspection.WalkGetTransactionResult(res)
	require.NoError(t, err)
	require.Len(t, got, 6+8, "the walk keeps the inner instructions of the failed instruction")
}

func TestWalkGetTransactionResult_V1Config(t *testing.T) {
	res := getTransactionResult(t, "swap_v1", "json")
	tx, err := res.Transaction.GetTransaction()
	require.NoError(t, err)

	cfg := tx.Message.TransactionConfig
	require.NotNil(t, cfg.ComputeUnitLimit)
	require.Equal(t, uint32(108354), *cfg.ComputeUnitLimit)
	require.NotNil(t, cfg.LoadedAccountsDataSizeLimit)
	require.Equal(t, uint32(7966555), *cfg.LoadedAccountsDataSizeLimit)
	require.Nil(t, cfg.PriorityFee)
	require.Nil(t, cfg.HeapSize)

	var heights []uint16
	ixs, err := introspection.WalkGetTransactionResult(res)
	require.NoError(t, err)
	for _, ix := range ixs {
		if ix.Trace.Inner {
			heights = append(heights, ix.Trace.StackHeight)
		}
	}
	require.Contains(t, heights, uint16(2))
	require.Contains(t, heights, uint16(3))
}

// getBlock, blockSubscribe and getTransactionsForAddress items carry the same
// transaction, meta and version as a getTransaction result.
func TestWalkBlockAndAddressItems_Fixtures(t *testing.T) {
	for _, fx := range fixtures {
		t.Run(fx.name, func(t *testing.T) {
			want, err := introspection.WalkGetTransactionResult(getTransactionResult(t, fx.name, "json"))
			require.NoError(t, err)

			for _, encoding := range fx.encodings {
				raw := readFixture(t, fx.name, encoding)

				var block rpc.TransactionWithMeta
				require.NoError(t, json.Unmarshal(raw, &block))
				got, err := introspection.WalkTransactionWithMeta(&block)
				require.NoError(t, err, encoding)
				require.Equal(t, want, got, encoding)

				// getTransactionsForAddress items have no version.
				var item rpc.TransactionForAddress
				require.NoError(t, json.Unmarshal(raw, &item))
				got, err = introspection.WalkTransactionForAddress(&item)
				require.NoError(t, err, encoding)
				require.Equal(t, want, got, encoding)
			}
		})
	}
}

// jsonParsedTransaction returns a block or address item that holds the
// transaction of a jsonParsed fixture. The meta is left out: rpc cannot
// unmarshal jsonParsed inner instructions.
func jsonParsedTransaction(t *testing.T, name string) []byte {
	var parsed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(readFixture(t, name, "jsonParsed"), &parsed))
	return []byte(`{"transaction":` + string(parsed["transaction"]) + `,"version":` + string(parsed["version"]) + `}`)
}

func TestWalk_JSONParsedIsRejected(t *testing.T) {
	items := []struct {
		name string
		raw  []byte
	}{
		{"parsed instructions", jsonParsedTransaction(t, "sync_native")},
		{"parsed and partially decoded instructions", jsonParsedTransaction(t, "swap_v0")},
		{"no instructions", []byte(`{"transaction":{"signatures":[],"message":{` +
			`"accountKeys":[],"recentBlockhash":"11111111111111111111111111111111","instructions":[]}}}`)},
	}
	for _, tc := range items {
		raw := tc.raw
		t.Run(tc.name, func(t *testing.T) {
			var block rpc.TransactionWithMeta
			require.NoError(t, json.Unmarshal(raw, &block))
			_, err := introspection.WalkTransactionWithMeta(&block)
			require.ErrorIs(t, err, introspection.ErrJSONParsedUnsupported)

			var item rpc.TransactionForAddress
			require.NoError(t, json.Unmarshal(raw, &item))
			_, err = introspection.WalkTransactionForAddress(&item)
			require.ErrorIs(t, err, introspection.ErrJSONParsedUnsupported)
		})
	}

	// rpc already fails to unmarshal a jsonParsed getTransaction result, whose
	// account keys are objects.
	for _, fx := range fixtures {
		var res rpc.GetTransactionResult
		require.Error(t, json.Unmarshal(readFixture(t, fx.name, "jsonParsed"), &res), fx.name)
	}
}

func TestWalk_UnrecognizedTransactions(t *testing.T) {
	_, err := introspection.WalkGetTransactionResult(nil)
	require.ErrorIs(t, err, introspection.ErrUnrecognizedTransaction)

	var noTransaction rpc.GetTransactionResult
	require.NoError(t, json.Unmarshal([]byte(`{"slot":1,"transaction":null,"meta":null}`), &noTransaction))
	_, err = introspection.WalkGetTransactionResult(&noTransaction)
	require.ErrorIs(t, err, introspection.ErrUnrecognizedTransaction)

	_, err = introspection.WalkTransactionWithMeta(nil)
	require.ErrorIs(t, err, introspection.ErrUnrecognizedTransaction)
	_, err = introspection.WalkTransactionForAddress(nil)
	require.ErrorIs(t, err, introspection.ErrUnrecognizedTransaction)

	items := []struct {
		name string
		raw  string
	}{
		{"no transaction", `{"slot":1,"transaction":null}`},
		// transactionDetails "accounts"
		{"accounts details", `{"transaction":{"signatures":["5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW"],` +
			`"accountKeys":[{"pubkey":"11111111111111111111111111111111","signer":false,"source":"transaction","writable":false}]}}`},
		// getTransactionsForAddress with transactionDetails "signatures"
		{"signatures details", `{"slot":1,"transactionIndex":2,"signature":"5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW"}`},
	}
	for _, tc := range items {
		raw := tc.raw
		t.Run(tc.name, func(t *testing.T) {
			var block rpc.TransactionWithMeta
			require.NoError(t, json.Unmarshal([]byte(raw), &block))
			_, err := introspection.WalkTransactionWithMeta(&block)
			require.ErrorIs(t, err, introspection.ErrUnrecognizedTransaction)

			var item rpc.TransactionForAddress
			require.NoError(t, json.Unmarshal([]byte(raw), &item))
			_, err = introspection.WalkTransactionForAddress(&item)
			require.ErrorIs(t, err, introspection.ErrUnrecognizedTransaction)
		})
	}
}

func TestWalk_UnsupportedVersion(t *testing.T) {
	res := getTransactionResult(t, "swap_v1", "base64")
	res.Version = 2
	_, err := introspection.WalkGetTransactionResult(res)
	require.ErrorIs(t, err, introspection.ErrUnsupportedVersion)

	var block rpc.TransactionWithMeta
	require.NoError(t, json.Unmarshal(readFixture(t, "swap_v1", "base64"), &block))
	block.Version = 2
	_, err = introspection.WalkTransactionWithMeta(&block)
	require.ErrorIs(t, err, introspection.ErrUnsupportedVersion)
}

// Unlike kit, which falls back to the static accounts, a v0 transaction with
// lookups needs its loaded addresses.
func TestWalk_V0WithoutLoadedAddresses(t *testing.T) {
	res := getTransactionResult(t, "swap_v0", "base64")
	tx, err := res.Transaction.GetTransaction()
	require.NoError(t, err)

	_, err = introspection.Walk(&tx.Message, nil)
	require.ErrorIs(t, err, solana.ErrAddressTablesNotSet)

	meta := *res.Meta
	meta.LoadedAddresses = rpc.LoadedAddresses{}
	_, err = introspection.Walk(&tx.Message, &meta)
	require.ErrorIs(t, err, solana.ErrAddressTablesNotSet)
}

func TestWalk_ResolvedMessage(t *testing.T) {
	res := getTransactionResult(t, "swap_v0", "base64")
	tx, err := res.Transaction.GetTransaction()
	require.NoError(t, err)
	want, err := introspection.Walk(&tx.Message, res.Meta)
	require.NoError(t, err)

	resolved, err := res.Transaction.GetTransaction()
	require.NoError(t, err)
	loaded := res.Meta.LoadedAddresses
	require.NoError(t, resolved.Message.ResolveLookupsWith(loaded.Writable, loaded.ReadOnly))

	got, err := introspection.Walk(&resolved.Message, res.Meta)
	require.NoError(t, err)
	require.Equal(t, want, got)

	// The resolved message alone is enough.
	meta := *res.Meta
	meta.LoadedAddresses = rpc.LoadedAddresses{}
	got, err = introspection.Walk(&resolved.Message, &meta)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

// For json results, rpc returns the same cached *solana.Transaction on every
// call. Walking it from many goroutines must be safe (run with -race).
func TestWalk_ConcurrentWalksOfOneResult(t *testing.T) {
	res := getTransactionResult(t, "swap_v0", "json")
	want, err := introspection.WalkGetTransactionResult(res)
	require.NoError(t, err)

	const workers = 8
	results := make([][]*introspection.Instruction, workers)
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[w], errs[w] = introspection.WalkGetTransactionResult(res)
		}()
	}
	wg.Wait()
	for w := range workers {
		require.NoError(t, errs[w])
		require.Equal(t, want, results[w])
	}
}

// parsedTransaction is the part of a jsonParsed getTransaction result that
// agave resolves itself: account keys with their roles, and the program,
// accounts and data of every instruction.
type parsedTransaction struct {
	Transaction struct {
		Message struct {
			AccountKeys []struct {
				Pubkey   solana.PublicKey `json:"pubkey"`
				Signer   bool             `json:"signer"`
				Writable bool             `json:"writable"`
				Source   string           `json:"source"`
			} `json:"accountKeys"`
			Instructions []parsedInstruction `json:"instructions"`
		} `json:"message"`
	} `json:"transaction"`
	Meta struct {
		InnerInstructions []struct {
			Index        int                 `json:"index"`
			Instructions []parsedInstruction `json:"instructions"`
		} `json:"innerInstructions"`
	} `json:"meta"`
}

// parsedInstruction has accounts and data only when agave could not parse
// the instruction ("partially decoded").
type parsedInstruction struct {
	ProgramID   solana.PublicKey   `json:"programId"`
	Accounts    []solana.PublicKey `json:"accounts"`
	Data        *string            `json:"data"`
	Parsed      json.RawMessage    `json:"parsed"`
	StackHeight *uint16            `json:"stackHeight"`
}

// inExplorerOrder flattens the instructions of a jsonParsed result in the
// order of Walk.
func (p *parsedTransaction) inExplorerOrder() []parsedInstruction {
	var out []parsedInstruction
	for i, ix := range p.Transaction.Message.Instructions {
		out = append(out, ix)
		for _, group := range p.Meta.InnerInstructions {
			if group.Index == i {
				out = append(out, group.Instructions...)
			}
		}
	}
	return out
}

// TestWalk_MatchesAgaveJSONParsed checks the resolution against agave's own,
// from the jsonParsed result of the same transaction.
func TestWalk_MatchesAgaveJSONParsed(t *testing.T) {
	for _, fx := range fixtures {
		t.Run(fx.name, func(t *testing.T) {
			res := getTransactionResult(t, fx.name, "base64")
			tx, err := res.Transaction.GetTransaction()
			require.NoError(t, err)
			var parsed parsedTransaction
			require.NoError(t, json.Unmarshal(readFixture(t, fx.name, "jsonParsed"), &parsed))

			want := parsed.inExplorerOrder()
			invoked := make(map[solana.PublicKey]bool)
			for _, ix := range want {
				invoked[ix.ProgramID] = true
			}

			// Accounts: same keys, order, signers, and static/loaded split.
			metas, err := introspection.AccountMetas(&tx.Message, &res.Meta.LoadedAddresses)
			require.NoError(t, err)
			keys := parsed.Transaction.Message.AccountKeys
			require.Len(t, metas, len(keys))
			for i, key := range keys {
				meta := metas[i]
				require.Equal(t, key.Pubkey, meta.PublicKey, "key %d", i)
				require.Equal(t, key.Signer, meta.IsSigner, "signer %d", i)
				wantSource := "transaction"
				if i >= len(tx.Message.AccountKeys) {
					wantSource = "lookupTable"
				}
				require.Equal(t, wantSource, key.Source, "source %d", i)
				if key.Writable {
					require.True(t, meta.IsWritable, "agave has key %d writable", i)
				} else if meta.IsWritable {
					// Agave demotes invoked programs and reserved accounts
					// such as sysvars; the header does not.
					require.True(t, invoked[meta.PublicKey] || strings.HasPrefix(meta.PublicKey.String(), "Sysvar"),
						"key %d (%s) is writable here and readonly in agave", i, meta.PublicKey)
				}
			}

			// Instructions: same programs and stack heights in walk order,
			// and the same accounts and data where agave keeps them raw.
			got, err := introspection.Walk(&tx.Message, res.Meta)
			require.NoError(t, err)
			require.Len(t, got, len(want))
			var raw int
			for i, ix := range got {
				w := want[i]
				require.Equal(t, w.ProgramID, ix.ProgID, "%s program", ix.Trace)
				if ix.Trace.Inner {
					var height uint16
					if w.StackHeight != nil {
						height = *w.StackHeight
					}
					require.Equal(t, height, ix.Trace.StackHeight, "%s stack height", ix.Trace)
				}
				if w.Parsed != nil {
					continue
				}
				raw++
				accounts := make([]solana.PublicKey, len(ix.AccountValues))
				for j, meta := range ix.AccountValues {
					accounts[j] = meta.PublicKey
				}
				require.Equal(t, len(w.Accounts), len(accounts), "%s accounts", ix.Trace)
				for j := range w.Accounts {
					assert.Equal(t, w.Accounts[j], accounts[j], "%s account %d", ix.Trace, j)
				}
				require.NotNil(t, w.Data, "%s data", ix.Trace)
				require.Equal(t, *w.Data, solana.Base58(ix.DataBytes).String(), "%s data", ix.Trace)
			}
			t.Logf("%d instructions, %d compared with raw accounts and data", len(got), raw)
		})
	}
}

func BenchmarkWalk(b *testing.B) {
	for _, fx := range fixtures {
		res := getTransactionResult(b, fx.name, "base64")
		tx, err := res.Transaction.GetTransaction()
		require.NoError(b, err)
		b.Run(fx.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := introspection.Walk(&tx.Message, res.Meta); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
