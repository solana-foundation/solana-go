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
	"errors"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/introspection"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/require"
)

func newKey() solana.PublicKey {
	return solana.NewWallet().PublicKey()
}

func newKeys(n int) solana.PublicKeySlice {
	keys := make(solana.PublicKeySlice, n)
	for i := range keys {
		keys[i] = newKey()
	}
	return keys
}

// v0Fixture is a v0 message with 4 static accounts (writable signer,
// readonly signer, writable, readonly program), 2 address lookup tables that
// load 2 writable and 2 readonly accounts, and instructions that use all of
// them. Every key is random and every data byte is non-zero, so no assertion
// can pass by accident on zero values.
//
// Account list: 0-3 static, 4-5 loaded writable, 6-7 loaded readonly.
type v0Fixture struct {
	static     solana.PublicKeySlice
	tableA     solana.PublicKey
	tableB     solana.PublicKey
	tableAKeys solana.PublicKeySlice
	tableBKeys solana.PublicKeySlice
	loaded     rpc.LoadedAddresses
}

func newV0Fixture() v0Fixture {
	f := v0Fixture{
		static:     newKeys(4),
		tableA:     newKey(),
		tableB:     newKey(),
		tableAKeys: newKeys(2),
		tableBKeys: newKeys(2),
	}
	f.loaded = rpc.LoadedAddresses{
		Writable: solana.PublicKeySlice{f.tableAKeys[0], f.tableBKeys[1]},
		ReadOnly: solana.PublicKeySlice{f.tableAKeys[1], f.tableBKeys[0]},
	}
	return f
}

// message returns a fresh, unresolved message for the fixture.
func (f v0Fixture) message() *solana.Message {
	msg := &solana.Message{
		Header: solana.MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: append(solana.PublicKeySlice{}, f.static...),
		Instructions: []solana.CompiledInstruction{
			{ProgramIDIndex: 3, Accounts: []uint16{0, 4, 6}, Data: solana.Base58{0xAA, 0x01}},
			{ProgramIDIndex: 3, Accounts: []uint16{5, 7, 2, 2}, Data: solana.Base58{0xAA, 0x02}},
			{ProgramIDIndex: 3},
		},
	}
	msg.SetAddressTableLookups([]solana.MessageAddressTableLookup{
		{AccountKey: f.tableA, WritableIndexes: []uint8{0}, ReadonlyIndexes: []uint8{1}},
		{AccountKey: f.tableB, WritableIndexes: []uint8{1}, ReadonlyIndexes: []uint8{0}},
	})
	return msg
}

func (f v0Fixture) tables() map[solana.PublicKey]solana.PublicKeySlice {
	return map[solana.PublicKey]solana.PublicKeySlice{
		f.tableA: append(solana.PublicKeySlice{}, f.tableAKeys...),
		f.tableB: append(solana.PublicKeySlice{}, f.tableBKeys...),
	}
}

// meta returns a fresh meta with the loaded addresses and two groups of
// inner instructions, for the first and second top-level instructions.
func (f v0Fixture) meta() *rpc.TransactionMeta {
	return &rpc.TransactionMeta{
		LoadedAddresses: rpc.LoadedAddresses{
			Writable: append(solana.PublicKeySlice{}, f.loaded.Writable...),
			ReadOnly: append(solana.PublicKeySlice{}, f.loaded.ReadOnly...),
		},
		InnerInstructions: []rpc.InnerInstruction{
			{Index: 0, Instructions: []rpc.CompiledInstruction{
				{ProgramIDIndex: 3, Accounts: []uint16{4, 1}, Data: solana.Base58{0xBB}, StackHeight: 2},
				{ProgramIDIndex: 3, Accounts: []uint16{7}, StackHeight: 3},
			}},
			{Index: 1, Instructions: []rpc.CompiledInstruction{
				{ProgramIDIndex: 3, Accounts: []uint16{6}, Data: solana.Base58{0xBC}, StackHeight: 2},
			}},
		},
	}
}

// metas returns the account list of the fixture: 0-3 static, 4-5 loaded
// writable, 6-7 loaded readonly.
func (f v0Fixture) metas() []solana.AccountMeta {
	return []solana.AccountMeta{
		{PublicKey: f.static[0], IsSigner: true, IsWritable: true},
		{PublicKey: f.static[1], IsSigner: true},
		{PublicKey: f.static[2], IsWritable: true},
		{PublicKey: f.static[3]},
		{PublicKey: f.loaded.Writable[0], IsWritable: true},
		{PublicKey: f.loaded.Writable[1], IsWritable: true},
		{PublicKey: f.loaded.ReadOnly[0]},
		{PublicKey: f.loaded.ReadOnly[1]},
	}
}

func (f v0Fixture) wantMetas() solana.AccountMetaSlice {
	return toMetaSlice(f.metas())
}

func toMetaSlice(values []solana.AccountMeta) solana.AccountMetaSlice {
	out := make(solana.AccountMetaSlice, len(values))
	for i := range values {
		out[i] = &values[i]
	}
	return out
}

// instruction builds the expected resolved instruction.
func (f v0Fixture) instruction(trace introspection.Trace, accounts []int, data []byte) *introspection.Instruction {
	metas := f.metas()
	ix := &introspection.Instruction{Trace: trace}
	ix.ProgID = f.static[3]
	for _, i := range accounts {
		meta := metas[i]
		ix.AccountValues = append(ix.AccountValues, &meta)
	}
	if len(data) > 0 {
		ix.DataBytes = data
	}
	return ix
}

func outerTrace(i int) introspection.Trace {
	return introspection.Trace{OuterIndex: i}
}

func innerTrace(outer, inner int, stackHeight uint16) introspection.Trace {
	return introspection.Trace{Inner: true, OuterIndex: outer, InnerIndex: inner, StackHeight: stackHeight}
}

// wantWalk is the expected walk of the fixture message and meta.
func (f v0Fixture) wantWalk() []*introspection.Instruction {
	return []*introspection.Instruction{
		f.instruction(outerTrace(0), []int{0, 4, 6}, []byte{0xAA, 0x01}),
		f.instruction(innerTrace(0, 0, 2), []int{4, 1}, []byte{0xBB}),
		f.instruction(innerTrace(0, 1, 3), []int{7}, nil),
		f.instruction(outerTrace(1), []int{5, 7, 2, 2}, []byte{0xAA, 0x02}),
		f.instruction(innerTrace(1, 0, 2), []int{6}, []byte{0xBC}),
		f.instruction(outerTrace(2), nil, nil),
	}
}

func TestAccountMetas_HeaderRoles(t *testing.T) {
	keys := newKeys(6)
	msg := &solana.Message{
		Header: solana.MessageHeader{
			NumRequiredSignatures:       3,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 2,
		},
		AccountKeys: keys,
	}

	got, err := introspection.AccountMetas(msg, nil)
	require.NoError(t, err)
	want := toMetaSlice([]solana.AccountMeta{
		{PublicKey: keys[0], IsSigner: true, IsWritable: true},
		{PublicKey: keys[1], IsSigner: true, IsWritable: true},
		{PublicKey: keys[2], IsSigner: true},
		{PublicKey: keys[3], IsWritable: true},
		{PublicKey: keys[4]},
		{PublicKey: keys[5]},
	})
	require.Equal(t, want, got)
}

func TestAccountMetas_LoadedAccountsFollowStaticAccounts(t *testing.T) {
	f := newV0Fixture()

	got, err := introspection.AccountMetas(f.message(), &f.loaded)
	require.NoError(t, err)
	require.Equal(t, f.wantMetas(), got)
}

func TestAccountMetas_LoadedAccountSources(t *testing.T) {
	f := newV0Fixture()

	cases := []struct {
		name   string
		msg    func(t *testing.T) *solana.Message
		loaded *rpc.LoadedAddresses
		// listErr is the error of Message.AccountMetaList, which needs the
		// lookups resolved or the address tables set.
		listErr error
	}{
		{
			name:    "meta loaded addresses",
			msg:     func(t *testing.T) *solana.Message { return f.message() },
			loaded:  &f.loaded,
			listErr: solana.ErrAddressTablesNotSet,
		},
		{
			name: "ResolveLookupsWith",
			msg: func(t *testing.T) *solana.Message {
				msg := f.message()
				require.NoError(t, msg.ResolveLookupsWith(f.loaded.Writable, f.loaded.ReadOnly))
				return msg
			},
		},
		{
			name: "ResolveLookupsWith and the same loaded addresses",
			msg: func(t *testing.T) *solana.Message {
				msg := f.message()
				require.NoError(t, msg.ResolveLookupsWith(f.loaded.Writable, f.loaded.ReadOnly))
				return msg
			},
			loaded: &f.loaded,
		},
		{
			name: "ResolveLookupsWith and empty loaded addresses",
			msg: func(t *testing.T) *solana.Message {
				msg := f.message()
				require.NoError(t, msg.ResolveLookupsWith(f.loaded.Writable, f.loaded.ReadOnly))
				return msg
			},
			loaded: &rpc.LoadedAddresses{},
		},
		{
			name: "ResolveLookups",
			msg: func(t *testing.T) *solana.Message {
				msg := f.message()
				require.NoError(t, msg.SetAddressTables(f.tables()))
				require.NoError(t, msg.ResolveLookups())
				return msg
			},
		},
		{
			name: "SetAddressTables",
			msg: func(t *testing.T) *solana.Message {
				msg := f.message()
				require.NoError(t, msg.SetAddressTables(f.tables()))
				return msg
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := tc.msg(t)
			resolved := msg.IsResolved()

			got, err := introspection.AccountMetas(msg, tc.loaded)
			require.NoError(t, err)
			require.Equal(t, f.wantMetas(), got)
			require.Equal(t, resolved, msg.IsResolved(), "AccountMetas must not resolve the message")

			// The same accounts as the existing API, where it works.
			want, err := msg.AccountMetaList()
			if tc.listErr != nil {
				require.ErrorIs(t, err, tc.listErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}

func TestAccountMetas_Errors(t *testing.T) {
	f := newV0Fixture()
	legacy := &solana.Message{
		Header:      solana.MessageHeader{NumRequiredSignatures: 1},
		AccountKeys: newKeys(2),
	}
	resolved := f.message()
	require.NoError(t, resolved.ResolveLookupsWith(f.loaded.Writable, f.loaded.ReadOnly))

	cases := []struct {
		name   string
		msg    *solana.Message
		loaded *rpc.LoadedAddresses
		want   error
	}{
		{"lookups without loaded addresses", f.message(), nil, solana.ErrAddressTablesNotSet},
		{"lookups with empty loaded addresses", f.message(), &rpc.LoadedAddresses{}, solana.ErrAddressTablesNotSet},
		{
			"too few loaded addresses", f.message(),
			&rpc.LoadedAddresses{Writable: f.loaded.Writable[:1], ReadOnly: f.loaded.ReadOnly},
			introspection.ErrLoadedAddressesMismatch,
		},
		{
			"right total, wrong split", f.message(),
			&rpc.LoadedAddresses{
				Writable: append(append(solana.PublicKeySlice{}, f.loaded.Writable...), f.loaded.ReadOnly[0]),
				ReadOnly: f.loaded.ReadOnly[1:],
			},
			introspection.ErrLoadedAddressesMismatch,
		},
		{"loaded addresses for a legacy message", legacy, &f.loaded, introspection.ErrLoadedAddressesMismatch},
		{
			"message resolved with other accounts", resolved,
			&rpc.LoadedAddresses{Writable: newKeys(2), ReadOnly: newKeys(2)},
			introspection.ErrLoadedAddressesMismatch,
		},
		{"nil message", nil, nil, introspection.ErrUnrecognizedTransaction},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := introspection.AccountMetas(tc.msg, tc.loaded)
			require.ErrorIs(t, err, tc.want)

			// Every entry point resolves accounts the same way.
			_, err = introspection.OuterInstructions(tc.msg, tc.loaded)
			require.ErrorIs(t, err, tc.want)
			var meta *rpc.TransactionMeta
			if tc.loaded != nil {
				meta = &rpc.TransactionMeta{LoadedAddresses: *tc.loaded}
			}
			_, err = introspection.Walk(tc.msg, meta)
			require.ErrorIs(t, err, tc.want)
		})
	}
}

func TestAccountMetas_DoesNotShareMemory(t *testing.T) {
	f := newV0Fixture()
	msg := f.message()
	require.NoError(t, msg.ResolveLookupsWith(f.loaded.Writable, f.loaded.ReadOnly))
	keys := append(solana.PublicKeySlice{}, msg.AccountKeys...)

	first, err := introspection.AccountMetas(msg, nil)
	require.NoError(t, err)
	for _, meta := range first {
		meta.PublicKey = newKey()
		meta.IsSigner = !meta.IsSigner
	}

	require.Equal(t, keys, msg.AccountKeys)
	second, err := introspection.AccountMetas(msg, nil)
	require.NoError(t, err)
	require.Equal(t, f.wantMetas(), second)
}

func TestOuterInstructions(t *testing.T) {
	f := newV0Fixture()

	got, err := introspection.OuterInstructions(f.message(), &f.loaded)
	require.NoError(t, err)
	want := []*introspection.Instruction{
		f.instruction(outerTrace(0), []int{0, 4, 6}, []byte{0xAA, 0x01}),
		f.instruction(outerTrace(1), []int{5, 7, 2, 2}, []byte{0xAA, 0x02}),
		f.instruction(outerTrace(2), nil, nil),
	}
	require.Equal(t, want, got)

	// Like kit, empty accounts and data are omitted.
	require.Nil(t, got[2].AccountValues)
	require.Nil(t, got[2].DataBytes)
}

func TestOuterInstructions_IndexOutOfRange(t *testing.T) {
	f := newV0Fixture()

	cases := []struct {
		name string
		ix   solana.CompiledInstruction
		want introspection.IndexError
	}{
		{
			"program index",
			solana.CompiledInstruction{ProgramIDIndex: 8, Accounts: []uint16{0}},
			introspection.IndexError{Trace: outerTrace(1), Position: -1, Index: 8, NumAccounts: 8},
		},
		{
			"account index",
			solana.CompiledInstruction{ProgramIDIndex: 3, Accounts: []uint16{0, 1, 9}},
			introspection.IndexError{Trace: outerTrace(1), Position: 2, Index: 9, NumAccounts: 8},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := f.message()
			msg.Instructions[1] = tc.ix

			got, err := introspection.OuterInstructions(msg, &f.loaded)
			require.Nil(t, got)
			require.ErrorIs(t, err, introspection.ErrIndexOutOfRange)
			var indexErr *introspection.IndexError
			require.ErrorAs(t, err, &indexErr)
			require.Equal(t, tc.want, *indexErr)
		})
	}
}

func TestInnerInstructions(t *testing.T) {
	f := newV0Fixture()
	metas := f.wantMetas()

	t.Run("no inner instructions", func(t *testing.T) {
		got, err := introspection.InnerInstructions(nil, metas)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("resolves accounts and data in meta order", func(t *testing.T) {
		got, err := introspection.InnerInstructions(f.meta().InnerInstructions, metas)
		require.NoError(t, err)
		want := []*introspection.Instruction{
			f.instruction(innerTrace(0, 0, 2), []int{4, 1}, []byte{0xBB}),
			f.instruction(innerTrace(0, 1, 3), []int{7}, nil),
			f.instruction(innerTrace(1, 0, 2), []int{6}, []byte{0xBC}),
		}
		require.Equal(t, want, got)
	})

	t.Run("no stack height reported", func(t *testing.T) {
		inner := []rpc.InnerInstruction{{Index: 0, Instructions: []rpc.CompiledInstruction{
			{ProgramIDIndex: 3, Accounts: []uint16{0}, Data: solana.Base58{0xBB}},
		}}}
		got, err := introspection.InnerInstructions(inner, metas)
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.Equal(t, innerTrace(0, 0, 0), got[0].Trace)
	})

	t.Run("empty accounts and data", func(t *testing.T) {
		inner := []rpc.InnerInstruction{{Index: 0, Instructions: []rpc.CompiledInstruction{{ProgramIDIndex: 3}}}}
		got, err := introspection.InnerInstructions(inner, metas)
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.Nil(t, got[0].AccountValues)
		require.Nil(t, got[0].DataBytes)
	})

	t.Run("program index out of range", func(t *testing.T) {
		inner := []rpc.InnerInstruction{{Index: 1, Instructions: []rpc.CompiledInstruction{
			{ProgramIDIndex: 3},
			{ProgramIDIndex: 8},
		}}}
		_, err := introspection.InnerInstructions(inner, metas)
		var indexErr *introspection.IndexError
		require.ErrorAs(t, err, &indexErr)
		require.Equal(t, introspection.IndexError{Trace: innerTrace(1, 1, 0), Position: -1, Index: 8, NumAccounts: 8}, *indexErr)
	})

	t.Run("account index out of range", func(t *testing.T) {
		inner := []rpc.InnerInstruction{{Index: 0, Instructions: []rpc.CompiledInstruction{
			{ProgramIDIndex: 3, Accounts: []uint16{1, 200}, StackHeight: 2},
		}}}
		_, err := introspection.InnerInstructions(inner, metas)
		var indexErr *introspection.IndexError
		require.ErrorAs(t, err, &indexErr)
		require.Equal(t, introspection.IndexError{Trace: innerTrace(0, 0, 2), Position: 1, Index: 200, NumAccounts: 8}, *indexErr)
	})

	t.Run("nil account meta", func(t *testing.T) {
		withNil := append(solana.AccountMetaSlice{}, metas...)
		withNil[2] = nil
		_, err := introspection.InnerInstructions(f.meta().InnerInstructions, withNil)
		require.Error(t, err)
	})
}

func TestWalk(t *testing.T) {
	f := newV0Fixture()

	got, err := introspection.Walk(f.message(), f.meta())
	require.NoError(t, err)
	require.Equal(t, f.wantWalk(), got)
}

func TestWalk_NilMetaReturnsOuterInstructions(t *testing.T) {
	keys := newKeys(3)
	msg := &solana.Message{
		Header:      solana.MessageHeader{NumRequiredSignatures: 1, NumReadonlyUnsignedAccounts: 1},
		AccountKeys: keys,
		Instructions: []solana.CompiledInstruction{
			{ProgramIDIndex: 2, Accounts: []uint16{0, 1}, Data: solana.Base58{0xAA}},
			{ProgramIDIndex: 2, Accounts: []uint16{1}, Data: solana.Base58{0xAB}},
		},
	}

	got, err := introspection.Walk(msg, nil)
	require.NoError(t, err)
	require.Len(t, got, 2)
	for i, ix := range got {
		require.Equal(t, outerTrace(i), ix.Trace)
		require.Equal(t, keys[2], ix.ProgID)
	}
	require.Equal(t, []byte{0xAB}, got[1].DataBytes)
}

func TestWalk_UnmatchedInnerGroups(t *testing.T) {
	f := newV0Fixture()
	meta := f.meta()
	group := func(index uint16, data byte) rpc.InnerInstruction {
		return rpc.InnerInstruction{Index: index, Instructions: []rpc.CompiledInstruction{
			{ProgramIDIndex: 3, Accounts: []uint16{0}, Data: solana.Base58{data}, StackHeight: 2},
		}}
	}
	// Groups for missing top-level instructions 5 and 7, and a second group
	// for instruction 0.
	meta.InnerInstructions = []rpc.InnerInstruction{
		group(5, 0xC1), group(0, 0xC2), group(7, 0xC3), group(0, 0xC4), group(5, 0xC5),
	}

	got, err := introspection.Walk(f.message(), meta)
	require.NoError(t, err)

	var traces []string
	var data []byte
	for _, ix := range got {
		traces = append(traces, ix.Trace.String())
		if ix.Trace.Inner {
			data = append(data, ix.DataBytes...)
		}
	}
	// Both groups of instruction 0 follow it in meta order. The groups of
	// missing instructions go last, in the order they first appear, so that
	// no instruction is lost.
	require.Equal(t, []string{
		"outer[0]", "inner[0/0]", "inner[0/0]", "outer[1]", "outer[2]", "inner[5/0]", "inner[5/0]", "inner[7/0]",
	}, traces)
	require.Equal(t, []byte{0xC2, 0xC4, 0xC1, 0xC5, 0xC3}, data)
}

func TestWalk_DecodesWithProgramDecoders(t *testing.T) {
	from, to := newKey(), newKey()
	transfer, err := system.NewTransferInstruction(4242, from, to).ValidateAndBuild()
	require.NoError(t, err)
	data, err := transfer.Data()
	require.NoError(t, err)

	msg := &solana.Message{
		Header:      solana.MessageHeader{NumRequiredSignatures: 1, NumReadonlyUnsignedAccounts: 1},
		AccountKeys: solana.PublicKeySlice{from, to, solana.SystemProgramID},
		Instructions: []solana.CompiledInstruction{
			{ProgramIDIndex: 2, Accounts: []uint16{0, 1}, Data: data},
		},
	}
	got, err := introspection.Walk(msg, nil)
	require.NoError(t, err)
	require.Len(t, got, 1)
	ix := got[0]

	// *Instruction is a solana.Instruction, and its program can be checked
	// directly.
	var generic solana.Instruction = ix
	require.Equal(t, solana.SystemProgramID, generic.ProgramID())
	require.True(t, ix.ProgID.Equals(system.ProgramID))

	// Through the program package.
	inst, err := system.DecodeInstruction(ix.AccountValues, ix.DataBytes)
	require.NoError(t, err)
	require.Equal(t, uint64(4242), *inst.Impl.(*system.Transfer).Lamports)

	// Through the instruction decoder registry.
	decoded, err := ix.Decode()
	require.NoError(t, err)
	viaRegistry, ok := decoded.(*system.Instruction)
	require.True(t, ok, "got %T", decoded)
	transferImpl, ok := viaRegistry.Impl.(*system.Transfer)
	require.True(t, ok, "got %T", viaRegistry.Impl)
	require.Equal(t, uint64(4242), *transferImpl.Lamports)

	// Decode works on copies: changing the decoded accounts leaves ix as is.
	transferImpl.GetFundingAccount().PublicKey = newKey()
	transferImpl.GetFundingAccount().IsSigner = false
	require.Equal(t, solana.AccountMeta{PublicKey: from, IsSigner: true, IsWritable: true}, *ix.AccountValues[0])
}

// Each returned instruction owns its program, accounts and data: changing or
// appending to them must leave the message, the meta, the other instructions
// and later walks unchanged.
func TestWalk_InstructionsOwnTheirMemory(t *testing.T) {
	f := newV0Fixture()
	msg, meta := f.message(), f.meta()
	wantMsg, wantMeta := f.message(), f.meta()

	got, err := introspection.Walk(msg, meta)
	require.NoError(t, err)
	require.Equal(t, f.wantWalk(), got)
	first, second := got[0], got[3] // outer[0] and outer[1] are neighbors in memory

	// Duplicate account references are distinct metas.
	require.NotSame(t, second.AccountValues[2], second.AccountValues[3])
	second.AccountValues[2].PublicKey = newKey()
	second.AccountValues[2].IsWritable = false

	// Appending must not write into the neighbor's accounts or data.
	first.AccountValues = append(first.AccountValues, &solana.AccountMeta{PublicKey: newKey()})
	first.DataBytes = append(first.DataBytes, 0xEE, 0xEE)
	first.DataBytes[0] = 0xEF
	first.ProgID = newKey()
	first.AccountValues[1].IsSigner = true

	want := f.wantWalk()
	require.Equal(t, want[1:3], got[1:3])
	require.Equal(t, want[4:], got[4:])
	require.Equal(t, f.static[2], second.AccountValues[3].PublicKey)
	require.Equal(t, want[3].AccountValues[:2], second.AccountValues[:2])
	require.Equal(t, want[3].DataBytes, second.DataBytes)

	// The inputs are unchanged, and a new walk gives the original result.
	require.Equal(t, wantMsg, msg)
	require.Equal(t, wantMeta, meta)
	again, err := introspection.Walk(msg, meta)
	require.NoError(t, err)
	require.Equal(t, f.wantWalk(), again)
}

func TestTrace_String(t *testing.T) {
	require.Equal(t, "outer[2]", outerTrace(2).String())
	require.Equal(t, "inner[2/0]", innerTrace(2, 0, 3).String())
}

func TestIndexError(t *testing.T) {
	program := &introspection.IndexError{Trace: outerTrace(1), Position: -1, Index: 9, NumAccounts: 8}
	account := &introspection.IndexError{Trace: innerTrace(2, 0, 2), Position: 3, Index: 9, NumAccounts: 8}

	require.Equal(t, "introspection: outer[1]: program index 9 out of range (8 accounts)", program.Error())
	require.Equal(t, "introspection: inner[2/0]: account 3: index 9 out of range (8 accounts)", account.Error())
	require.True(t, errors.Is(account, introspection.ErrIndexOutOfRange))
}
