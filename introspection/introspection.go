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

package introspection

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	// ErrIndexOutOfRange is wrapped by *IndexError, returned when an
	// instruction refers to an account index the transaction does not have.
	ErrIndexOutOfRange = errors.New("introspection: account index out of range")

	// ErrLoadedAddressesMismatch is returned when the loaded addresses do not
	// match the address table lookups of the message.
	ErrLoadedAddressesMismatch = errors.New("introspection: loaded addresses do not match the message lookups")

	// ErrJSONParsedUnsupported is returned for transactions fetched with the
	// jsonParsed encoding: their instructions come without their raw data.
	ErrJSONParsedUnsupported = errors.New("introspection: jsonParsed transactions are not supported, fetch them with the json, base58 or base64 encoding")

	// ErrUnrecognizedTransaction is returned when the input holds no
	// transaction message to inspect.
	ErrUnrecognizedTransaction = errors.New("introspection: no transaction message to inspect")

	// ErrUnsupportedVersion is returned for transaction versions other than
	// legacy, 0 and 1.
	ErrUnsupportedVersion = errors.New("introspection: unsupported transaction version")
)

// Trace records where an instruction sits in a transaction.
type Trace struct {
	// Inner is false for a top-level instruction and true for an instruction
	// emitted by a cross-program invocation (CPI).
	Inner bool

	// OuterIndex is the position of a top-level instruction in the message.
	// For an inner instruction, it is the index of the top-level instruction
	// whose CPIs emitted it, as reported by the RPC.
	OuterIndex int

	// InnerIndex is the position of an inner instruction in its group of
	// inner instructions. It is 0 for top-level instructions.
	InnerIndex int

	// StackHeight is the CPI depth of an inner instruction: 2 when a
	// top-level instruction invoked it, 3 one level deeper, and so on. It is
	// 0 for top-level instructions, and when the RPC did not report it.
	StackHeight uint16
}

// String returns "outer[i]" for a top-level instruction and
// "inner[outer/inner]" for an inner instruction.
func (t Trace) String() string {
	if t.Inner {
		return fmt.Sprintf("inner[%d/%d]", t.OuterIndex, t.InnerIndex)
	}
	return fmt.Sprintf("outer[%d]", t.OuterIndex)
}

// Instruction is an instruction of a confirmed transaction, with its program,
// accounts and data resolved, and its position in the transaction.
//
// *Instruction implements solana.Instruction through the embedded
// GenericInstruction. ProgID, AccountValues and DataBytes belong to this
// value alone: they share no memory with the message, the meta or the other
// instructions, and empty accounts or data are nil.
type Instruction struct {
	solana.GenericInstruction

	// Trace is the position of the instruction in the transaction.
	Trace Trace
}

var _ solana.Instruction = (*Instruction)(nil)

// Decode decodes the instruction with the decoder registered for its program
// (see solana.RegisterInstructionDecoder). The package of the program must be
// imported for its decoder to be registered. The decoder gets copies of the
// accounts and data, so the decoded value shares no memory with ix.
func (ix *Instruction) Decode() (any, error) {
	accounts := make(solana.AccountMetaSlice, len(ix.AccountValues))
	values := make([]solana.AccountMeta, len(ix.AccountValues))
	for i, meta := range ix.AccountValues {
		if meta != nil {
			values[i] = *meta
		}
		accounts[i] = &values[i]
	}
	return solana.DecodeInstruction(ix.ProgID, accounts, bytes.Clone(ix.DataBytes))
}

// IndexError reports an instruction that refers to an account index the
// transaction does not have. It wraps ErrIndexOutOfRange.
type IndexError struct {
	// Trace locates the instruction.
	Trace Trace

	// Position is the position of the bad index in the account list of the
	// instruction, or -1 when the program index is out of range.
	Position int

	// Index is the out-of-range account index.
	Index uint16

	// NumAccounts is the number of accounts of the transaction.
	NumAccounts int
}

func (e *IndexError) Error() string {
	if e.Position < 0 {
		return fmt.Sprintf("introspection: %s: program index %d out of range (%d accounts)",
			e.Trace, e.Index, e.NumAccounts)
	}
	return fmt.Sprintf("introspection: %s: account %d: index %d out of range (%d accounts)",
		e.Trace, e.Position, e.Index, e.NumAccounts)
}

// Unwrap returns ErrIndexOutOfRange.
func (e *IndexError) Unwrap() error {
	return ErrIndexOutOfRange
}

// AccountMetas returns the accounts of msg in the order instruction account
// indexes refer to them: the static account keys, then the accounts loaded
// from address lookup tables, writable ones first.
//
// The roles of static accounts come from the message header. Loaded accounts
// are never signers, and only the loaded writable ones are writable. As with
// Message.AccountMetaList, the runtime demotion of writable accounts (for
// example, of invoked programs) is not applied.
//
// The loaded accounts come from the first of:
//   - loaded, when it is not empty, typically the LoadedAddresses of the
//     transaction meta. Its counts must match the lookups of the message,
//     otherwise the error wraps ErrLoadedAddressesMismatch;
//   - the accounts that Message.ResolveLookups or Message.ResolveLookupsWith
//     appended to the message;
//   - the address tables set with Message.SetAddressTables.
//
// For a message with lookups and none of the above, the error wraps
// solana.ErrAddressTablesNotSet.
//
// msg is not modified, and the result shares no memory with it.
func AccountMetas(msg *solana.Message, loaded *rpc.LoadedAddresses) (solana.AccountMetaSlice, error) {
	accounts, err := accountList(msg, loaded)
	if err != nil {
		return nil, err
	}
	out := make(solana.AccountMetaSlice, len(accounts))
	for i := range accounts {
		out[i] = &accounts[i]
	}
	return out, nil
}

// OuterInstructions returns the top-level instructions of msg, resolved
// against its accounts (see AccountMetas). The first instruction that refers
// to a missing account makes it return an *IndexError.
func OuterInstructions(msg *solana.Message, loaded *rpc.LoadedAddresses) ([]*Instruction, error) {
	accounts, err := accountList(msg, loaded)
	if err != nil {
		return nil, err
	}
	var size sizes
	size.addOuter(msg)
	return newBuilder(accounts, size).outer(msg)
}

// InnerInstructions returns the inner instructions of a transaction meta,
// resolved against accounts, the account list of the transaction as returned
// by AccountMetas. The result follows the order of the meta, and each
// instruction has an inner Trace. The first instruction that refers to a
// missing account makes it return an *IndexError.
func InnerInstructions(inner []rpc.InnerInstruction, accounts solana.AccountMetaSlice) ([]*Instruction, error) {
	list := make([]solana.AccountMeta, len(accounts))
	for i, meta := range accounts {
		if meta == nil {
			return nil, fmt.Errorf("introspection: nil account meta at index %d", i)
		}
		list[i] = *meta
	}
	var size sizes
	size.addInner(inner)
	return newBuilder(list, size).inner(inner)
}

// Walk returns every instruction of a confirmed transaction in the order an
// explorer shows them: each top-level instruction, followed by the inner
// instructions that its CPIs emitted.
//
// meta is the RPC meta of the transaction. Its LoadedAddresses resolve the
// accounts loaded from address lookup tables (see AccountMetas), and its
// InnerInstructions are interleaved with the top-level instructions. Inner
// instruction groups whose index matches no top-level instruction can only
// come from malformed input: they are appended at the end, so that no
// instruction is lost. With a nil meta, Walk returns the top-level
// instructions only.
//
// The first instruction that refers to a missing account makes Walk return an
// *IndexError. Neither msg nor meta is modified, so concurrent walks of the
// same transaction are safe.
func Walk(msg *solana.Message, meta *rpc.TransactionMeta) ([]*Instruction, error) {
	var (
		loaded *rpc.LoadedAddresses
		groups []rpc.InnerInstruction
	)
	if meta != nil {
		loaded = &meta.LoadedAddresses
		groups = meta.InnerInstructions
	}
	accounts, err := accountList(msg, loaded)
	if err != nil {
		return nil, err
	}

	var size sizes
	size.addOuter(msg)
	size.addInner(groups)
	b := newBuilder(accounts, size)
	outer, err := b.outer(msg)
	if err != nil {
		return nil, err
	}
	inner, err := b.inner(groups)
	if err != nil {
		return nil, err
	}
	return interleave(outer, inner), nil
}

// interleave puts each group of inner instructions right after the top-level
// instruction it belongs to, keeping the order of the meta. Groups that match
// no top-level instruction go last, in the order they first appear.
func interleave(outer, inner []*Instruction) []*Instruction {
	if len(inner) == 0 {
		return outer
	}
	byOuter := make(map[int][]*Instruction)
	var orphans []int
	for _, ix := range inner {
		index := ix.Trace.OuterIndex
		if _, seen := byOuter[index]; !seen && index >= len(outer) {
			orphans = append(orphans, index)
		}
		byOuter[index] = append(byOuter[index], ix)
	}

	out := make([]*Instruction, 0, len(outer)+len(inner))
	for i, ix := range outer {
		out = append(out, ix)
		out = append(out, byOuter[i]...)
	}
	for _, index := range orphans {
		out = append(out, byOuter[index]...)
	}
	return out
}

// accountList returns the flat account list of msg, as described in
// AccountMetas. The list is new on every call.
func accountList(msg *solana.Message, loaded *rpc.LoadedAddresses) ([]solana.AccountMeta, error) {
	if msg == nil {
		return nil, fmt.Errorf("%w: nil message", ErrUnrecognizedTransaction)
	}
	switch version := msg.GetVersion(); version {
	case solana.MessageVersionLegacy, solana.MessageVersionV0, solana.MessageVersionV1:
	default:
		return nil, fmt.Errorf("%w: message version %d", ErrUnsupportedVersion, version)
	}

	numStatic := len(msg.AccountKeys)
	if msg.IsResolved() {
		// ResolveLookups and ResolveLookupsWith append the looked-up
		// accounts to AccountKeys.
		numStatic -= msg.NumLookups()
		if numStatic < 0 {
			return nil, fmt.Errorf("introspection: resolved message has %d account keys, fewer than its %d lookups",
				len(msg.AccountKeys), msg.NumLookups())
		}
	}
	writable, readonly, err := loadedAccounts(msg, loaded, numStatic)
	if err != nil {
		return nil, err
	}

	// Same roles as Message.AccountMetaList, with saturating arithmetic for
	// headers that claim more accounts than the message has.
	h := msg.Header
	numSigners := int(h.NumRequiredSignatures)
	numWritableSigners := max(numSigners-int(h.NumReadonlySignedAccounts), 0)
	numWritableUnsigned := max(numStatic-numSigners-int(h.NumReadonlyUnsignedAccounts), 0)

	accounts := make([]solana.AccountMeta, 0, numStatic+len(writable)+len(readonly))
	for i, key := range msg.AccountKeys[:numStatic] {
		meta := solana.AccountMeta{PublicKey: key, IsSigner: i < numSigners}
		if i < numSigners {
			meta.IsWritable = i < numWritableSigners
		} else {
			meta.IsWritable = i-numSigners < numWritableUnsigned
		}
		accounts = append(accounts, meta)
	}
	for _, key := range writable {
		accounts = append(accounts, solana.AccountMeta{PublicKey: key, IsWritable: true})
	}
	for _, key := range readonly {
		accounts = append(accounts, solana.AccountMeta{PublicKey: key})
	}
	return accounts, nil
}

// loadedAccounts returns the accounts that msg loads from address lookup
// tables, from the sources listed in AccountMetas. The result may share
// memory with msg or loaded: it is only read.
func loadedAccounts(
	msg *solana.Message,
	loaded *rpc.LoadedAddresses,
	numStatic int,
) (writable, readonly solana.PublicKeySlice, err error) {
	numWritable := msg.NumWritableLookups()
	numReadonly := msg.NumLookups() - numWritable

	if loaded != nil && len(loaded.Writable)+len(loaded.ReadOnly) > 0 {
		if len(loaded.Writable) != numWritable || len(loaded.ReadOnly) != numReadonly {
			return nil, nil, fmt.Errorf(
				"%w: got %d writable and %d readonly addresses, the lookups need %d writable and %d readonly",
				ErrLoadedAddressesMismatch, len(loaded.Writable), len(loaded.ReadOnly), numWritable, numReadonly,
			)
		}
		if msg.IsResolved() {
			appended := msg.AccountKeys[numStatic:]
			if !equalKeys(appended[:numWritable], loaded.Writable) || !equalKeys(appended[numWritable:], loaded.ReadOnly) {
				return nil, nil, fmt.Errorf("%w: the message was resolved with other accounts", ErrLoadedAddressesMismatch)
			}
		}
		return loaded.Writable, loaded.ReadOnly, nil
	}

	switch {
	case msg.NumLookups() == 0:
		return nil, nil, nil
	case msg.IsResolved():
		appended := msg.AccountKeys[numStatic:]
		return appended[:numWritable], appended[numWritable:], nil
	case len(msg.GetAddressTables()) > 0:
		// GetAllKeys does not modify an unresolved message.
		all, err := msg.GetAllKeys()
		if err != nil {
			return nil, nil, fmt.Errorf("introspection: %w", err)
		}
		lookedUp := all[numStatic:]
		return lookedUp[:numWritable], lookedUp[numWritable:], nil
	default:
		return nil, nil, fmt.Errorf("introspection: pass the LoadedAddresses of the transaction meta: %w",
			solana.ErrAddressTablesNotSet)
	}
}

func equalKeys(a, b solana.PublicKeySlice) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// sizes counts what the instructions of one call need, so that builder can
// allocate everything up front.
type sizes struct {
	instructions int
	accounts     int
	data         int
}

func (s *sizes) addOuter(msg *solana.Message) {
	for i := range msg.Instructions {
		s.instructions++
		s.accounts += len(msg.Instructions[i].Accounts)
		s.data += len(msg.Instructions[i].Data)
	}
}

func (s *sizes) addInner(groups []rpc.InnerInstruction) {
	for g := range groups {
		for i := range groups[g].Instructions {
			s.instructions++
			s.accounts += len(groups[g].Instructions[i].Accounts)
			s.data += len(groups[g].Instructions[i].Data)
		}
	}
}

// builder resolves compiled instructions against an account list. It carves
// the instructions, account metas and data of one call out of a few
// preallocated slabs. Each instruction gets its own capacity-capped
// sub-slices, so appending to the slices of one instruction never writes
// into another's.
type builder struct {
	accounts     []solana.AccountMeta
	instructions []Instruction
	refs         []*solana.AccountMeta
	values       []solana.AccountMeta
	data         []byte
}

func newBuilder(accounts []solana.AccountMeta, size sizes) *builder {
	return &builder{
		accounts:     accounts,
		instructions: make([]Instruction, 0, size.instructions),
		refs:         make([]*solana.AccountMeta, 0, size.accounts),
		values:       make([]solana.AccountMeta, 0, size.accounts),
		data:         make([]byte, 0, size.data),
	}
}

func (b *builder) outer(msg *solana.Message) ([]*Instruction, error) {
	out := make([]*Instruction, 0, len(msg.Instructions))
	for i := range msg.Instructions {
		ci := &msg.Instructions[i]
		ix, err := b.add(ci.ProgramIDIndex, ci.Accounts, ci.Data, Trace{OuterIndex: i})
		if err != nil {
			return nil, err
		}
		out = append(out, ix)
	}
	return out, nil
}

func (b *builder) inner(groups []rpc.InnerInstruction) ([]*Instruction, error) {
	var out []*Instruction
	for g := range groups {
		group := &groups[g]
		for i := range group.Instructions {
			ci := &group.Instructions[i]
			trace := Trace{
				Inner:       true,
				OuterIndex:  int(group.Index),
				InnerIndex:  i,
				StackHeight: ci.StackHeight,
			}
			ix, err := b.add(ci.ProgramIDIndex, ci.Accounts, ci.Data, trace)
			if err != nil {
				return nil, err
			}
			out = append(out, ix)
		}
	}
	return out, nil
}

func (b *builder) add(programIndex uint16, accountIndexes []uint16, data []byte, trace Trace) (*Instruction, error) {
	if int(programIndex) >= len(b.accounts) {
		return nil, &IndexError{Trace: trace, Position: -1, Index: programIndex, NumAccounts: len(b.accounts)}
	}
	for position, index := range accountIndexes {
		if int(index) >= len(b.accounts) {
			return nil, &IndexError{Trace: trace, Position: position, Index: index, NumAccounts: len(b.accounts)}
		}
	}

	b.instructions = append(b.instructions, Instruction{Trace: trace})
	ix := &b.instructions[len(b.instructions)-1]
	ix.ProgID = b.accounts[programIndex].PublicKey

	if n := len(accountIndexes); n > 0 {
		start := len(b.values)
		for _, index := range accountIndexes {
			b.values = append(b.values, b.accounts[index])
		}
		// Every account reference gets its own meta, duplicates included,
		// so changing one never changes another.
		for i := start; i < start+n; i++ {
			b.refs = append(b.refs, &b.values[i])
		}
		end := len(b.refs)
		ix.AccountValues = b.refs[end-n : end : end]
	}
	if n := len(data); n > 0 {
		start := len(b.data)
		b.data = append(b.data, data...)
		ix.DataBytes = b.data[start : start+n : start+n]
	}
	return ix, nil
}
