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
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/introspection"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/require"
)

// fuzzKeys returns n distinct, non-zero keys.
func fuzzKeys(tag byte, n int) solana.PublicKeySlice {
	keys := make(solana.PublicKeySlice, n)
	for i := range keys {
		keys[i][0] = tag
		keys[i][1] = byte(i)
		keys[i][31] = 0xFF
	}
	return keys
}

// appendInnerSpec encodes an inner instruction for parseInnerSpec:
// [group, program, numAccounts, accounts..., dataLen, data..., stackHeight].
func appendInnerSpec(spec []byte, group uint16, ix rpc.CompiledInstruction) []byte {
	accounts := ix.Accounts[:min(len(ix.Accounts), 7)]
	data := ix.Data[:min(len(ix.Data), 7)]
	spec = append(spec, byte(group), byte(ix.ProgramIDIndex), byte(len(accounts)))
	for _, account := range accounts {
		spec = append(spec, byte(account))
	}
	spec = append(spec, byte(len(data)))
	spec = append(spec, data...)
	return append(spec, byte(ix.StackHeight))
}

// parseInnerSpec decodes the inner instructions encoded by appendInnerSpec,
// tolerating any input.
func parseInnerSpec(spec []byte) []rpc.InnerInstruction {
	next := func() (byte, bool) {
		if len(spec) == 0 {
			return 0, false
		}
		b := spec[0]
		spec = spec[1:]
		return b, true
	}
	var groups []rpc.InnerInstruction
	for {
		group, ok1 := next()
		program, ok2 := next()
		numAccounts, ok3 := next()
		if !ok1 || !ok2 || !ok3 {
			return groups
		}
		ix := rpc.CompiledInstruction{ProgramIDIndex: uint16(program)}
		for range numAccounts % 8 {
			if account, ok := next(); ok {
				ix.Accounts = append(ix.Accounts, uint16(account))
			}
		}
		dataLen, _ := next()
		for range dataLen % 8 {
			if b, ok := next(); ok {
				ix.Data = append(ix.Data, b)
			}
		}
		height, _ := next()
		ix.StackHeight = uint16(height % 5)

		index := uint16(group % 8)
		if n := len(groups); n > 0 && groups[n-1].Index == index {
			groups[n-1].Instructions = append(groups[n-1].Instructions, ix)
		} else {
			groups = append(groups, rpc.InnerInstruction{Index: index, Instructions: []rpc.CompiledInstruction{ix}})
		}
	}
}

// FuzzWalk walks arbitrary wire transactions with arbitrary inner
// instructions and loaded addresses. Walk must never panic, and every
// instruction it returns must point at accounts of the transaction.
func FuzzWalk(f *testing.F) {
	for _, fx := range fixtures {
		res := getTransactionResult(f, fx.name, "base64")
		var spec []byte
		for _, group := range res.Meta.InnerInstructions {
			for _, ix := range group.Instructions {
				spec = appendInnerSpec(spec, group.Index, ix)
			}
		}
		loaded := res.Meta.LoadedAddresses
		f.Add(res.Transaction.GetBinary(), spec, uint8(len(loaded.Writable)), uint8(len(loaded.ReadOnly)))
	}

	f.Fuzz(func(t *testing.T, wire, spec []byte, numWritable, numReadonly uint8) {
		tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(wire))
		if err != nil {
			return
		}
		meta := &rpc.TransactionMeta{
			LoadedAddresses: rpc.LoadedAddresses{
				Writable: fuzzKeys(1, int(numWritable)),
				ReadOnly: fuzzKeys(2, int(numReadonly)),
			},
			InnerInstructions: parseInnerSpec(spec),
		}

		ixs, err := introspection.Walk(&tx.Message, meta)
		if err != nil {
			return
		}

		accounts, err := introspection.AccountMetas(&tx.Message, &meta.LoadedAddresses)
		require.NoError(t, err)
		known := make(map[solana.PublicKey]bool, len(accounts))
		for _, account := range accounts {
			known[account.PublicKey] = true
		}
		numInner := 0
		for _, group := range meta.InnerInstructions {
			numInner += len(group.Instructions)
		}
		require.Len(t, ixs, len(tx.Message.Instructions)+numInner)
		for _, ix := range ixs {
			require.True(t, known[ix.ProgID], "%s: unknown program", ix.Trace)
			for _, account := range ix.AccountValues {
				require.NotNil(t, account)
				require.True(t, known[account.PublicKey], "%s: unknown account", ix.Trace)
			}
		}
	})
}
