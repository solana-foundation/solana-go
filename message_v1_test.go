package solana

import (
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ported from solana-sdk message/src/versions/v1/{config,message}.rs and versions/mod.rs.

// v1MessageBuilder mirrors the upstream test-only `v1::MessageBuilder`.
type v1MessageBuilder struct {
	msg Message
}

func newV1MessageBuilder() *v1MessageBuilder {
	b := &v1MessageBuilder{}
	b.msg.version = MessageVersionV1
	return b
}

func (b *v1MessageBuilder) requiredSignatures(n uint8) *v1MessageBuilder {
	b.msg.Header.NumRequiredSignatures = n
	return b
}

func (b *v1MessageBuilder) readonlySignedAccounts(n uint8) *v1MessageBuilder {
	b.msg.Header.NumReadonlySignedAccounts = n
	return b
}

func (b *v1MessageBuilder) readonlyUnsignedAccounts(n uint8) *v1MessageBuilder {
	b.msg.Header.NumReadonlyUnsignedAccounts = n
	return b
}

func (b *v1MessageBuilder) lifetimeSpecifier(h Hash) *v1MessageBuilder {
	b.msg.RecentBlockhash = h
	return b
}

func (b *v1MessageBuilder) priorityFee(v uint64) *v1MessageBuilder {
	b.msg.TransactionConfig = b.msg.TransactionConfig.WithPriorityFee(v)
	return b
}

func (b *v1MessageBuilder) computeUnitLimit(v uint32) *v1MessageBuilder {
	b.msg.TransactionConfig = b.msg.TransactionConfig.WithComputeUnitLimit(v)
	return b
}

func (b *v1MessageBuilder) loadedAccountsDataSizeLimit(v uint32) *v1MessageBuilder {
	b.msg.TransactionConfig = b.msg.TransactionConfig.WithLoadedAccountsDataSizeLimit(v)
	return b
}

func (b *v1MessageBuilder) heapSize(v uint32) *v1MessageBuilder {
	b.msg.TransactionConfig = b.msg.TransactionConfig.WithHeapSize(v)
	return b
}

func (b *v1MessageBuilder) accounts(keys ...PublicKey) *v1MessageBuilder {
	b.msg.AccountKeys = PublicKeySlice(keys)
	return b
}

func (b *v1MessageBuilder) instruction(ix CompiledInstruction) *v1MessageBuilder {
	b.msg.Instructions = append(b.msg.Instructions, ix)
	return b
}

func (b *v1MessageBuilder) instructions(ixs ...CompiledInstruction) *v1MessageBuilder {
	b.msg.Instructions = ixs
	return b
}

// build validates the message (like upstream `MessageBuilder::build`) and returns it.
func (b *v1MessageBuilder) build(t *testing.T) Message {
	t.Helper()
	require.NoError(t, b.msg.Sanitize())
	return b.msg
}

func uniqueHash() Hash {
	var h Hash
	copy(h[:], newUniqueKey().Bytes())
	return h
}

func uniqueKeys(n int) []PublicKey {
	out := make([]PublicKey, n)
	for i := range out {
		out[i] = newUniqueKey()
	}
	return out
}

// createTestV1Message mirrors upstream `create_test_message`.
func createTestV1Message(t *testing.T) Message {
	t.Helper()
	return newV1MessageBuilder().
		requiredSignatures(1).
		readonlyUnsignedAccounts(1).
		lifetimeSpecifier(uniqueHash()).
		accounts(
			newUniqueKey(), // fee payer
			newUniqueKey(), // program
			newUniqueKey(), // readonly account
		).
		computeUnitLimit(200_000).
		instruction(CompiledInstruction{
			ProgramIDIndex: 1,
			Accounts:       []uint16{0, 2},
			Data:           []byte{1, 2, 3, 4},
		}).
		build(t)
}

func mustMarshalV1(t *testing.T, msg *Message) []byte {
	t.Helper()
	out, err := msg.MarshalV1()
	require.NoError(t, err)
	return out
}

func mustUnmarshalV1(t *testing.T, data []byte) Message {
	t.Helper()
	var msg Message
	require.NoError(t, msg.UnmarshalWithDecoder(bin.NewBinDecoder(data)))
	return msg
}

// ---------------------------------------------------------------------------
// config.rs
// ---------------------------------------------------------------------------

func TestV1ConfigMask_HasUnknownBits(t *testing.T) {
	assert.False(t, TransactionConfigMask(0).HasUnknownBits())
	assert.False(t, TransactionConfigMask(0b11111).HasUnknownBits())
	assert.True(t, TransactionConfigMask(0b100000).HasUnknownBits())
	assert.True(t, TransactionConfigMask(0x80000000).HasUnknownBits())
	assert.True(t, TransactionConfigMask(0b111111).HasUnknownBits())
}

func TestV1ConfigMask_HasPriorityFeeRequiresBothBits(t *testing.T) {
	assert.False(t, TransactionConfigMask(0).HasPriorityFee())
	assert.False(t, TransactionConfigMask(0b01).HasPriorityFee())
	assert.False(t, TransactionConfigMask(0b10).HasPriorityFee())
	assert.True(t, TransactionConfigMask(0b11).HasPriorityFee())
}

func TestV1ConfigMask_HasInvalidPriorityFeeBits(t *testing.T) {
	assert.False(t, TransactionConfigMask(0).HasInvalidPriorityFeeBits())
	assert.True(t, TransactionConfigMask(0b01).HasInvalidPriorityFeeBits())
	assert.True(t, TransactionConfigMask(0b10).HasInvalidPriorityFeeBits())
	assert.False(t, TransactionConfigMask(0b11).HasInvalidPriorityFeeBits())
}

func TestV1ConfigMask_FieldMethods(t *testing.T) {
	mask := TransactionConfigMask(0b11100)
	assert.True(t, mask.HasComputeUnitLimit())
	assert.True(t, mask.HasLoadedAccountsDataSize())
	assert.True(t, mask.HasHeapSize())

	mask = TransactionConfigMask(0)
	assert.False(t, mask.HasComputeUnitLimit())
	assert.False(t, mask.HasLoadedAccountsDataSize())
	assert.False(t, mask.HasHeapSize())
}

func TestV1ConfigMask_SizeOfConfig(t *testing.T) {
	assert.Equal(t, 0, TransactionConfigMask(0).SizeOfConfig())
	assert.Equal(t, 8, TransactionConfigMask(0b11).SizeOfConfig())
	assert.Equal(t, 4, TransactionConfigMask(0b100).SizeOfConfig())
	assert.Equal(t, 20, TransactionConfigMask(0b11111).SizeOfConfig())
}

func TestV1ConfigMask_FromConfig(t *testing.T) {
	cfg := TransactionConfig{}.WithPriorityFee(1000).WithComputeUnitLimit(200_000)
	mask := cfg.Mask()
	assert.True(t, mask.HasPriorityFee())
	assert.True(t, mask.HasComputeUnitLimit())
	assert.False(t, mask.HasLoadedAccountsDataSize())
	assert.False(t, mask.HasHeapSize())
	assert.Equal(t, TransactionConfigMask(0b111), mask)
	assert.Equal(t, 12, cfg.Size())
	assert.Equal(t, mask.SizeOfConfig(), cfg.Size())
}

func TestV1ConfigMask_InvariantsForAllKnownBitPatterns(t *testing.T) {
	for raw := uint32(0); raw < 1<<5; raw++ {
		mask := TransactionConfigMask(raw)
		assert.False(t, mask.HasUnknownBits(), "raw=%b", raw)
		if mask.HasPriorityFee() {
			assert.False(t, mask.HasInvalidPriorityFeeBits())
		}
		expected := 0
		if mask.HasPriorityFee() {
			expected += 8
		}
		if mask.HasComputeUnitLimit() {
			expected += 4
		}
		if mask.HasLoadedAccountsDataSize() {
			expected += 4
		}
		if mask.HasHeapSize() {
			expected += 4
		}
		assert.Equal(t, expected, mask.SizeOfConfig(), "raw=%b", raw)
	}
}

func TestV1Config_BuilderSetsAllFields(t *testing.T) {
	cfg := TransactionConfig{}.
		WithPriorityFee(1000).
		WithComputeUnitLimit(200_000).
		WithLoadedAccountsDataSizeLimit(64 * 1024).
		WithHeapSize(64 * 1024)

	require.NotNil(t, cfg.PriorityFee)
	require.NotNil(t, cfg.ComputeUnitLimit)
	require.NotNil(t, cfg.LoadedAccountsDataSizeLimit)
	require.NotNil(t, cfg.HeapSize)
	assert.Equal(t, uint64(1000), *cfg.PriorityFee)
	assert.Equal(t, uint32(200_000), *cfg.ComputeUnitLimit)
	assert.Equal(t, uint32(64*1024), *cfg.LoadedAccountsDataSizeLimit)
	assert.Equal(t, uint32(64*1024), *cfg.HeapSize)
	assert.False(t, cfg.IsEmpty())
	assert.True(t, TransactionConfig{}.IsEmpty())
	assert.Equal(t, TransactionConfigMaskKnownBits, cfg.Mask())
}

func TestV1Config_WithDoesNotAlias(t *testing.T) {
	base := TransactionConfig{}.WithComputeUnitLimit(1)
	a := base.WithComputeUnitLimit(2)
	b := base.WithPriorityFee(3)
	assert.Equal(t, uint32(1), *base.ComputeUnitLimit)
	assert.Equal(t, uint32(2), *a.ComputeUnitLimit)
	assert.Nil(t, base.PriorityFee)
	assert.Equal(t, uint32(1), *b.ComputeUnitLimit)
	assert.Equal(t, uint64(3), *b.PriorityFee)
}

// ---------------------------------------------------------------------------
// message.rs
// ---------------------------------------------------------------------------

func TestV1_FeePayerReturnsFirstAccount(t *testing.T) {
	feePayer := newUniqueKey()
	msg := newV1MessageBuilder().
		requiredSignatures(1).
		lifetimeSpecifier(uniqueHash()).
		accounts(feePayer, newUniqueKey()).
		build(t)
	assert.Equal(t, feePayer, msg.AccountKeys[0])
	assert.Equal(t, PublicKeySlice{feePayer}, msg.Signers())
}

func TestV1_IsSignerChecksSignatureRequirement(t *testing.T) {
	msg := createTestV1Message(t)
	assert.True(t, msg.IsSigner(msg.AccountKeys[0]))  // fee payer is signer
	assert.False(t, msg.IsSigner(msg.AccountKeys[1])) // program is not signer
	assert.False(t, msg.IsSigner(msg.AccountKeys[2])) // readonly account is not signer
}

func TestV1_IsSignerWritableIdentifiesWritableSigners(t *testing.T) {
	msg := newV1MessageBuilder().
		requiredSignatures(3).
		readonlySignedAccounts(1). // last signer is readonly
		lifetimeSpecifier(uniqueHash()).
		accounts(
			newUniqueKey(), // 0: writable signer
			newUniqueKey(), // 1: writable signer
			newUniqueKey(), // 2: readonly signer
			newUniqueKey(), // 3: non-signer
		).
		build(t)

	isSignerWritable := func(i int) bool {
		key := msg.AccountKeys[i]
		w, err := msg.IsWritable(key)
		require.NoError(t, err)
		return msg.IsSigner(key) && w
	}
	assert.True(t, isSignerWritable(0))
	assert.True(t, isSignerWritable(1))
	assert.False(t, isSignerWritable(2))
	assert.False(t, isSignerWritable(3))
}

func TestV1_IsSignerWritableAllWritableWhenNoReadonly(t *testing.T) {
	msg := newV1MessageBuilder().
		requiredSignatures(2).
		readonlySignedAccounts(0).
		lifetimeSpecifier(uniqueHash()).
		accounts(newUniqueKey(), newUniqueKey(), newUniqueKey()).
		build(t)

	metas, err := msg.AccountMetaList()
	require.NoError(t, err)
	assert.True(t, metas[0].IsSigner && metas[0].IsWritable)
	assert.True(t, metas[1].IsSigner && metas[1].IsWritable)
	assert.False(t, metas[2].IsSigner)
	assert.True(t, metas[2].IsWritable) // writable unsigned (no readonly unsigned)
}

func TestV1_IsWritableIndexRespectsHeaderLayout(t *testing.T) {
	msg := createTestV1Message(t)
	// Account layout: [writable signer (fee payer), writable unsigned (program), readonly unsigned]
	metas, err := msg.AccountMetaList()
	require.NoError(t, err)
	assert.True(t, metas[0].IsWritable)  // fee payer is writable
	assert.True(t, metas[1].IsWritable)  // program position is writable unsigned
	assert.False(t, metas[2].IsWritable) // last account is readonly

	assert.True(t, msg.IsWritableStatic(msg.AccountKeys[0]))
	assert.True(t, msg.IsWritableStatic(msg.AccountKeys[1]))
	assert.False(t, msg.IsWritableStatic(msg.AccountKeys[2]))
}

func TestV1_IsWritableIndexHandlesMixedSignerPermissions(t *testing.T) {
	msg := createTestV1Message(t)
	// 2 signers: first writable, second readonly
	msg.Header.NumRequiredSignatures = 2
	msg.Header.NumReadonlySignedAccounts = 1
	msg.Header.NumReadonlyUnsignedAccounts = 1
	msg.AccountKeys = PublicKeySlice{
		newUniqueKey(), // writable signer
		newUniqueKey(), // readonly signer
		newUniqueKey(), // readonly unsigned
	}
	msg.Instructions[0].ProgramIDIndex = 2
	msg.Instructions[0].Accounts = []uint16{0, 1}

	require.NoError(t, msg.Sanitize())
	metas, err := msg.AccountMetaList()
	require.NoError(t, err)
	assert.True(t, metas[0].IsWritable)  // writable signer
	assert.False(t, metas[1].IsWritable) // readonly signer
	assert.False(t, metas[2].IsWritable) // readonly unsigned
	w, err := msg.IsWritable(newUniqueKey())
	require.NoError(t, err)
	assert.False(t, w) // unknown key
}

// -- sanitize --

func TestV1Sanitize_AcceptsValidMessage(t *testing.T) {
	msg := createTestV1Message(t)
	require.NoError(t, msg.Sanitize())
}

func TestV1Sanitize_RejectsZeroSigners(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Header.NumRequiredSignatures = 0
	err := msg.Sanitize()
	require.Error(t, err)
	assert.True(t, IsSanitizeError(err))
	assert.Contains(t, err.Error(), "no writable signer")
}

func TestV1Sanitize_RejectsOver12Signatures(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Header.NumRequiredSignatures = MaxSignaturesV1 + 1
	msg.AccountKeys = uniqueKeys(MaxSignaturesV1 + 1)
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many signatures")
}

func TestV1Sanitize_RejectsOver64Addresses(t *testing.T) {
	msg := createTestV1Message(t)
	msg.AccountKeys = uniqueKeys(65)
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many addresses")
}

func TestV1Sanitize_RejectsOver64Instructions(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Instructions = make([]CompiledInstruction, 65)
	for i := range msg.Instructions {
		msg.Instructions[i] = CompiledInstruction{ProgramIDIndex: 1, Accounts: []uint16{0}}
	}
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many instructions")
}

func TestV1Sanitize_RejectsInsufficientAccountsForHeader(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Header.NumReadonlyUnsignedAccounts = 10
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "more accounts than available")
}

func TestV1Sanitize_RejectsAllSignersReadonly(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Header.NumReadonlySignedAccounts = 1
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no writable signer")
}

func TestV1Sanitize_RejectsDuplicateAddresses(t *testing.T) {
	msg := createTestV1Message(t)
	msg.AccountKeys[1] = msg.AccountKeys[0]
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestV1Sanitize_RejectsUnalignedHeapSize(t *testing.T) {
	msg := createTestV1Message(t)
	msg.TransactionConfig = msg.TransactionConfig.WithHeapSize(MinHeapSizeV1 + 1)
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple of 1024")
}

func TestV1Sanitize_AcceptsAlignedHeapSize(t *testing.T) {
	msg := createTestV1Message(t)
	msg.TransactionConfig = msg.TransactionConfig.WithHeapSize(65536)
	require.NoError(t, msg.Sanitize())
}

// Upstream #699: heap size bounds.
func TestV1Sanitize_HeapSizeBounds(t *testing.T) {
	msg := createTestV1Message(t)

	msg.TransactionConfig = msg.TransactionConfig.WithHeapSize(MinHeapSizeV1 - 1024)
	require.Error(t, msg.Sanitize(), "below min")

	msg.TransactionConfig = msg.TransactionConfig.WithHeapSize(MaxHeapSizeV1 + 1024)
	require.Error(t, msg.Sanitize(), "above max")

	msg.TransactionConfig = msg.TransactionConfig.WithHeapSize(MinHeapSizeV1)
	require.NoError(t, msg.Sanitize(), "at min")

	msg.TransactionConfig = msg.TransactionConfig.WithHeapSize(MaxHeapSizeV1)
	require.NoError(t, msg.Sanitize(), "at max")

	msg.TransactionConfig = msg.TransactionConfig.WithHeapSize(0)
	require.Error(t, msg.Sanitize(), "zero (multiple of 1024 but below min)")
}

func TestV1Sanitize_RejectsInvalidProgramIDIndex(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Instructions[0].ProgramIDIndex = 99
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "program_id_index 99 out of bounds")
}

func TestV1Sanitize_RejectsFeePayerAsProgram(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Instructions[0].ProgramIDIndex = 0
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be 0")
}

func TestV1Sanitize_RejectsInstructionWithTooManyAccounts(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Instructions[0].Accounts = make([]uint16, 256)
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many accounts")
}

func TestV1Sanitize_RejectsInvalidInstructionAccountIndex(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Instructions[0].Accounts = []uint16{0, 99}
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account index 99 out of bounds")
}

func TestV1Sanitize_Accepts64Addresses(t *testing.T) {
	msg := createTestV1Message(t)
	msg.AccountKeys = uniqueKeys(MaxAddressesV1)
	msg.Header.NumRequiredSignatures = 1
	msg.Header.NumReadonlySignedAccounts = 0
	msg.Header.NumReadonlyUnsignedAccounts = 1
	msg.Instructions[0].ProgramIDIndex = 1
	msg.Instructions[0].Accounts = []uint16{0, 2}
	require.NoError(t, msg.Sanitize())
}

func TestV1Sanitize_Accepts64Instructions(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Instructions = make([]CompiledInstruction, MaxInstructionsV1)
	for i := range msg.Instructions {
		msg.Instructions[i] = CompiledInstruction{ProgramIDIndex: 1, Accounts: []uint16{0, 2}, Data: []byte{1, 2, 3}}
	}
	require.NoError(t, msg.Sanitize())
}

// V1 messages are never silently demoted to v0/legacy.
func TestV1_NoSilentDemotion(t *testing.T) {
	msg := createTestV1Message(t)

	msg.SetAddressTableLookups(nil)
	assert.Equal(t, MessageVersionV1, msg.GetVersion())
	msg.AddAddressTableLookup(MessageAddressTableLookup{AccountKey: newUniqueKey(), WritableIndexes: []uint8{0}})
	assert.Equal(t, MessageVersionV1, msg.GetVersion())
	_, err := msg.MarshalBinary()
	require.Error(t, err)
	msg.SetAddressTableLookups(nil)

	// Leaving V1 with a non-empty config is an error; clearing it first works.
	_, err = msg.SetVersion(MessageVersionLegacy)
	require.Error(t, err)
	_, err = msg.SetVersion(MessageVersionV0)
	require.Error(t, err)
	assert.Equal(t, MessageVersionV1, msg.GetVersion())
	msg.TransactionConfig = TransactionConfig{}
	_, err = msg.SetVersion(MessageVersionLegacy)
	require.NoError(t, err)
	assert.Equal(t, MessageVersionLegacy, msg.GetVersion())
}

// Reusing a Message across decodes must not leak a v1 config into legacy/v0.
func TestV1_DecodeResetsConfigOnReuse(t *testing.T) {
	v1 := createTestV1Message(t)
	var m Message
	require.NoError(t, m.UnmarshalWithDecoder(bin.NewBinDecoder(mustMarshalV1(t, &v1))))
	require.False(t, m.TransactionConfig.IsEmpty())

	legacy := Message{Header: MessageHeader{NumRequiredSignatures: 1}, AccountKeys: PublicKeySlice{newUniqueKey()}}
	raw, err := legacy.MarshalLegacy()
	require.NoError(t, err)
	require.NoError(t, m.UnmarshalWithDecoder(bin.NewBinDecoder(raw)))
	assert.Equal(t, MessageVersionLegacy, m.GetVersion())
	assert.True(t, m.TransactionConfig.IsEmpty())
	assert.Nil(t, m.AddressTableLookups)
}

func TestV1Sanitize_RejectsAddressTableLookups(t *testing.T) {
	msg := createTestV1Message(t)
	msg.AddressTableLookups = MessageAddressTableLookupSlice{{AccountKey: newUniqueKey(), WritableIndexes: []uint8{0}}}
	err := msg.Sanitize()
	require.Error(t, err)
	assert.True(t, IsSanitizeError(err))
	_, err = msg.MarshalBinary()
	require.Error(t, err)
}

// -- serialization --

func TestV1_SizeMatchesSerializedLength(t *testing.T) {
	cases := []Message{
		// Minimal message
		newV1MessageBuilder().
			requiredSignatures(1).
			lifetimeSpecifier(uniqueHash()).
			accounts(newUniqueKey()).
			build(t),
		// With config
		newV1MessageBuilder().
			requiredSignatures(1).
			lifetimeSpecifier(uniqueHash()).
			accounts(newUniqueKey(), newUniqueKey()).
			priorityFee(1000).
			computeUnitLimit(200_000).
			instruction(CompiledInstruction{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{1, 2, 3, 4}}).
			build(t),
		// Multiple instructions with varying data
		newV1MessageBuilder().
			requiredSignatures(2).
			readonlySignedAccounts(1).
			readonlyUnsignedAccounts(1).
			lifetimeSpecifier(uniqueHash()).
			accounts(newUniqueKey(), newUniqueKey(), newUniqueKey(), newUniqueKey()).
			heapSize(65536).
			instructions(
				CompiledInstruction{ProgramIDIndex: 2, Accounts: []uint16{0, 1}, Data: []byte{}},
				CompiledInstruction{ProgramIDIndex: 3, Accounts: []uint16{0, 1, 2}, Data: make([]byte, 100)},
			).
			build(t),
	}
	for i := range cases {
		out := mustMarshalV1(t, &cases[i])
		assert.Equal(t, cases[i].sizeV1(), len(out), "case %d", i)
	}
}

func TestV1_ByteLayoutWithoutConfig(t *testing.T) {
	feePayer := PublicKey{}
	program := PublicKey{}
	var blockhash Hash
	for i := range 32 {
		feePayer[i] = 1
		program[i] = 2
		blockhash[i] = 0xAB
	}

	msg := newV1MessageBuilder().
		requiredSignatures(1).
		lifetimeSpecifier(blockhash).
		accounts(feePayer, program).
		instruction(CompiledInstruction{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{0xDE, 0xAD}}).
		build(t)

	got := mustMarshalV1(t, &msg)

	// Build expected bytes manually per SIMD-0385.
	expected := []byte{0x81}                // version byte
	expected = append(expected, 1, 0, 0)    // header
	expected = append(expected, 0, 0, 0, 0) // ConfigMask = 0
	expected = append(expected, blockhash[:]...)
	expected = append(expected, 1) // NumInstructions
	expected = append(expected, 2) // NumAddresses
	expected = append(expected, feePayer[:]...)
	expected = append(expected, program[:]...)
	// ConfigValues: none
	expected = append(expected, 1)    // program_id_index
	expected = append(expected, 1)    // num_accounts
	expected = append(expected, 2, 0) // data_len u16 LE
	expected = append(expected, 0)    // account index 0
	expected = append(expected, 0xDE, 0xAD)
	assert.Equal(t, expected, got)

	// And it round-trips.
	back := mustUnmarshalV1(t, got)
	assert.Equal(t, MessageVersionV1, back.GetVersion())
	assert.Equal(t, msg.Header, back.Header)
	assert.Equal(t, msg.AccountKeys, back.AccountKeys)
	assert.Equal(t, msg.RecentBlockhash, back.RecentBlockhash)
	assert.Equal(t, msg.Instructions, back.Instructions)
	assert.True(t, back.TransactionConfig.IsEmpty())
}

func TestV1_ByteLayoutWithConfig(t *testing.T) {
	feePayer := PublicKey{}
	program := PublicKey{}
	var blockhash Hash
	for i := range 32 {
		feePayer[i] = 1
		program[i] = 2
		blockhash[i] = 0xBB
	}

	msg := newV1MessageBuilder().
		requiredSignatures(1).
		lifetimeSpecifier(blockhash).
		accounts(feePayer, program).
		priorityFee(0x0102030405060708).
		computeUnitLimit(0x11223344).
		instruction(CompiledInstruction{ProgramIDIndex: 1, Accounts: []uint16{}, Data: []byte{}}).
		build(t)

	got := mustMarshalV1(t, &msg)

	expected := []byte{0x81, 1, 0, 0}
	// ConfigMask: priority fee (bits 0,1) + CU limit (bit 2) = 0b111 = 7
	expected = append(expected, 7, 0, 0, 0)
	expected = append(expected, blockhash[:]...)
	expected = append(expected, 1, 2)
	expected = append(expected, feePayer[:]...)
	expected = append(expected, program[:]...)
	expected = binary.LittleEndian.AppendUint64(expected, 0x0102030405060708)
	expected = binary.LittleEndian.AppendUint32(expected, 0x11223344)
	expected = append(expected, 1)    // program_id_index
	expected = append(expected, 0)    // num_accounts
	expected = append(expected, 0, 0) // data_len
	assert.Equal(t, expected, got)
}

func TestV1_RoundtripPreservesAllConfigFields(t *testing.T) {
	msg := newV1MessageBuilder().
		requiredSignatures(1).
		lifetimeSpecifier(uniqueHash()).
		accounts(newUniqueKey(), newUniqueKey()).
		priorityFee(1000).
		computeUnitLimit(200_000).
		loadedAccountsDataSizeLimit(1_000_000).
		heapSize(65536).
		instruction(CompiledInstruction{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{}}).
		build(t)

	back := mustUnmarshalV1(t, mustMarshalV1(t, &msg))
	assert.Equal(t, msg.TransactionConfig, back.TransactionConfig)
	assert.Equal(t, uint64(1000), *back.TransactionConfig.PriorityFee)
	assert.Equal(t, uint32(200_000), *back.TransactionConfig.ComputeUnitLimit)
	assert.Equal(t, uint32(1_000_000), *back.TransactionConfig.LoadedAccountsDataSizeLimit)
	assert.Equal(t, uint32(65536), *back.TransactionConfig.HeapSize)
}

// Upstream #621/#622: config mask validation during deserialization.
func TestV1_DeserializeRejectsUnknownConfigMaskBits(t *testing.T) {
	msg := createTestV1Message(t)
	data := mustMarshalV1(t, &msg)
	// mask is at offset 4..8 (after version byte + 3-byte header)
	binary.LittleEndian.PutUint32(data[4:8], uint32(TransactionConfigMaskKnownBits)|(1<<5))
	var back Message
	err := back.UnmarshalWithDecoder(bin.NewBinDecoder(data))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transaction config mask")
}

func TestV1_DeserializeRejectsPartialPriorityFeeBits(t *testing.T) {
	msg := createTestV1Message(t)
	data := mustMarshalV1(t, &msg)
	for _, bad := range []uint32{0b01, 0b10} {
		binary.LittleEndian.PutUint32(data[4:8], bad)
		var back Message
		err := back.UnmarshalWithDecoder(bin.NewBinDecoder(data))
		require.Error(t, err, "mask %b", bad)
		assert.Contains(t, err.Error(), "invalid transaction config mask")
	}
}

func TestV1_DeserializeRejectsWrongPrefix(t *testing.T) {
	msg := createTestV1Message(t)
	data := mustMarshalV1(t, &msg)
	data[0] = 0x80
	var back Message
	err := back.UnmarshalV1(bin.NewBinDecoder(data))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid v1 message version prefix")
}

// Every strict prefix of a valid encoding must fail to decode (and never panic).
func TestV1_DeserializeBufferTooSmall(t *testing.T) {
	msg := newV1MessageBuilder().
		requiredSignatures(2).
		readonlySignedAccounts(1).
		readonlyUnsignedAccounts(1).
		lifetimeSpecifier(uniqueHash()).
		accounts(newUniqueKey(), newUniqueKey(), newUniqueKey(), newUniqueKey()).
		priorityFee(1).
		computeUnitLimit(2).
		loadedAccountsDataSizeLimit(3).
		heapSize(MinHeapSizeV1).
		instructions(
			CompiledInstruction{ProgramIDIndex: 2, Accounts: []uint16{0, 1}, Data: []byte{9, 9, 9}},
			CompiledInstruction{ProgramIDIndex: 3, Accounts: []uint16{}, Data: []byte{}},
			CompiledInstruction{ProgramIDIndex: 3, Accounts: []uint16{1}, Data: []byte{7}},
		).
		build(t)
	data := mustMarshalV1(t, &msg)
	for n := 0; n < len(data); n++ {
		var back Message
		err := back.UnmarshalWithDecoder(bin.NewBinDecoder(data[:n]))
		require.Error(t, err, "prefix of %d/%d bytes must fail", n, len(data))
	}
	// The full buffer decodes fine and matches.
	back := mustUnmarshalV1(t, data)
	assert.Equal(t, msg.Instructions, back.Instructions)
	assert.Equal(t, msg.TransactionConfig, back.TransactionConfig)
}

// ---------------------------------------------------------------------------
// versions/mod.rs
// ---------------------------------------------------------------------------

// Port of the proptest `test_v1_message_raw_bytes_roundtrip` using a seeded PRNG.
func TestV1_MessageRawBytesRoundtrip(t *testing.T) {
	rng := rand.New(rand.NewSource(0x5385))
	optU64 := func() *uint64 {
		if rng.Intn(2) == 0 {
			return nil
		}
		v := rng.Uint64()
		return &v
	}
	optU32 := func(max int) *uint32 {
		if rng.Intn(2) == 0 {
			return nil
		}
		v := uint32(rng.Intn(max + 1))
		return &v
	}

	for iter := 0; iter < 200; iter++ {
		numAccounts := 12 + rng.Intn(64-12+1)
		requiredSignatures := 1 + rng.Intn(12)
		keys := make(PublicKeySlice, numAccounts)
		for i := range keys {
			rng.Read(keys[i][:])
		}
		var lifetime Hash
		rng.Read(lifetime[:])

		cfg := TransactionConfig{
			PriorityFee:                 optU64(),
			ComputeUnitLimit:            optU32(1_400_000),
			LoadedAccountsDataSizeLimit: optU32(20_480),
		}
		if rng.Intn(2) == 1 {
			v := uint32(rng.Intn(33)) * 1024
			cfg.HeapSize = &v
		}

		programIDIndex := uint16(1 + rng.Intn(numAccounts-1))
		numIxAccounts := requiredSignatures + rng.Intn(numAccounts-requiredSignatures+1)
		ixAccounts := make([]uint16, numIxAccounts)
		for i := range ixAccounts {
			ixAccounts[i] = uint16(rng.Intn(numAccounts))
		}
		data := make([]byte, rng.Intn(2049))
		rng.Read(data)

		original := Message{
			version: MessageVersionV1,
			Header: MessageHeader{
				NumRequiredSignatures: uint8(requiredSignatures),
			},
			RecentBlockhash:   lifetime,
			AccountKeys:       keys,
			TransactionConfig: cfg,
			Instructions: []CompiledInstruction{{
				ProgramIDIndex: programIDIndex,
				Accounts:       ixAccounts,
				Data:           data,
			}},
		}

		bytes := mustMarshalV1(t, &original)
		require.Equal(t, byte(0x81), bytes[0])
		require.Equal(t, original.sizeV1(), len(bytes))

		parsed := mustUnmarshalV1(t, bytes)
		assert.Equal(t, MessageVersionV1, parsed.GetVersion())
		assert.Equal(t, original.Header, parsed.Header)
		assert.Equal(t, original.RecentBlockhash, parsed.RecentBlockhash)
		assert.Equal(t, original.AccountKeys, parsed.AccountKeys)
		assert.Equal(t, original.TransactionConfig, parsed.TransactionConfig)
		assert.Equal(t, original.Instructions, parsed.Instructions)

		// MarshalBinary on the V1 message is the same as MarshalV1.
		viaBinary, err := original.MarshalBinary()
		require.NoError(t, err)
		assert.Equal(t, bytes, viaBinary)

		// Re-encoding the parsed message is byte-identical.
		again, err := parsed.MarshalBinary()
		require.NoError(t, err)
		assert.Equal(t, bytes, again)
	}
}

func TestV1_VersionedMessageJSONRoundtrip(t *testing.T) {
	msg := newV1MessageBuilder().
		requiredSignatures(1).
		lifetimeSpecifier(uniqueHash()).
		accounts(newUniqueKey(), newUniqueKey()).
		priorityFee(1000).
		computeUnitLimit(200_000).
		instruction(CompiledInstruction{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{1, 2, 3, 4}}).
		build(t)

	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"transactionConfig":{"priorityFee":1000,"computeUnitLimit":200000,"loadedAccountsDataSizeLimit":null,"heapSize":null}`)
	assert.NotContains(t, string(data), "addressTableLookups")

	var back Message
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, MessageVersionV1, back.GetVersion())
	assert.Equal(t, msg.Header, back.Header)
	assert.Equal(t, msg.AccountKeys, back.AccountKeys)
	assert.Equal(t, msg.RecentBlockhash, back.RecentBlockhash)
	assert.Equal(t, msg.Instructions, back.Instructions)
	assert.Equal(t, msg.TransactionConfig, back.TransactionConfig)

	// Both keys present with non-empty lookups: error (v0 and v1 are exclusive).
	both := `{"accountKeys":[],"header":{"numRequiredSignatures":1,"numReadonlySignedAccounts":0,"numReadonlyUnsignedAccounts":0},"recentBlockhash":"11111111111111111111111111111111","instructions":[],"addressTableLookups":[{"accountKey":"11111111111111111111111111111111","writableIndexes":[0],"readonlyIndexes":[]}],"transactionConfig":{"priorityFee":null,"computeUnitLimit":5,"loadedAccountsDataSizeLimit":null,"heapSize":null}}`
	var m2 Message
	require.Error(t, json.Unmarshal([]byte(both), &m2))

	// Empty lookups array alongside a config is tolerated: V1.
	emptyLookups := `{"accountKeys":[],"header":{"numRequiredSignatures":1,"numReadonlySignedAccounts":0,"numReadonlyUnsignedAccounts":0},"recentBlockhash":"11111111111111111111111111111111","instructions":[],"addressTableLookups":[],"transactionConfig":{"priorityFee":null,"computeUnitLimit":5,"loadedAccountsDataSizeLimit":null,"heapSize":null}}`
	require.NoError(t, json.Unmarshal([]byte(emptyLookups), &m2))
	assert.Equal(t, MessageVersionV1, m2.GetVersion())
	assert.Nil(t, m2.AddressTableLookups)
	require.NotNil(t, m2.TransactionConfig.ComputeUnitLimit)
	assert.Equal(t, uint32(5), *m2.TransactionConfig.ComputeUnitLimit)

	// transactionConfig with all-null fields -> V1 with an empty config.
	emptyCfg := `{"accountKeys":[],"header":{"numRequiredSignatures":1,"numReadonlySignedAccounts":0,"numReadonlyUnsignedAccounts":0},"recentBlockhash":"11111111111111111111111111111111","instructions":[],"transactionConfig":{"priorityFee":null,"computeUnitLimit":null,"loadedAccountsDataSizeLimit":null,"heapSize":null}}`
	var m3 Message
	require.NoError(t, json.Unmarshal([]byte(emptyCfg), &m3))
	assert.Equal(t, MessageVersionV1, m3.GetVersion())
	assert.True(t, m3.TransactionConfig.IsEmpty())

	// A null transactionConfig reads as absent (the RPC omits it): legacy.
	nullCfg := `{"accountKeys":[],"header":{"numRequiredSignatures":1,"numReadonlySignedAccounts":0,"numReadonlyUnsignedAccounts":0},"recentBlockhash":"11111111111111111111111111111111","instructions":[],"transactionConfig":null}`
	var m4 Message
	require.NoError(t, json.Unmarshal([]byte(nullCfg), &m4))
	assert.Equal(t, MessageVersionLegacy, m4.GetVersion())
}

// Port of `test_v1_wincode_roundtrip`: fixed messages through the generic
// MarshalBinary/UnmarshalWithDecoder path.
func TestV1_WincodeRoundtrip(t *testing.T) {
	msgs := []Message{
		// Minimal message
		newV1MessageBuilder().
			requiredSignatures(1).
			lifetimeSpecifier(uniqueHash()).
			accounts(newUniqueKey(), newUniqueKey()).
			instruction(CompiledInstruction{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{}}).
			build(t),
		// With config
		newV1MessageBuilder().
			requiredSignatures(1).
			lifetimeSpecifier(uniqueHash()).
			accounts(newUniqueKey(), newUniqueKey()).
			priorityFee(1000).
			computeUnitLimit(200_000).
			instruction(CompiledInstruction{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{1, 2, 3, 4}}).
			build(t),
		// Multiple instructions
		newV1MessageBuilder().
			requiredSignatures(2).
			lifetimeSpecifier(uniqueHash()).
			accounts(newUniqueKey(), newUniqueKey(), newUniqueKey()).
			heapSize(65536).
			instructions(
				CompiledInstruction{ProgramIDIndex: 2, Accounts: []uint16{0, 1}, Data: []byte{0xAA, 0xBB}},
				CompiledInstruction{ProgramIDIndex: 2, Accounts: []uint16{1}, Data: []byte{0xCC}},
			).
			build(t),
	}
	for i := range msgs {
		data, err := msgs[i].MarshalBinary()
		require.NoError(t, err)
		back := mustUnmarshalV1(t, data)
		assert.Equal(t, MessageVersionV1, back.GetVersion())
		assert.Equal(t, msgs[i].Header, back.Header)
		assert.Equal(t, msgs[i].AccountKeys, back.AccountKeys)
		assert.Equal(t, msgs[i].RecentBlockhash, back.RecentBlockhash)
		assert.Equal(t, msgs[i].Instructions, back.Instructions)
		assert.Equal(t, msgs[i].TransactionConfig, back.TransactionConfig)

		// base64 helpers
		var viaB64 Message
		require.NoError(t, viaB64.UnmarshalBase64(msgs[i].ToBase64()))
		assert.Equal(t, MessageVersionV1, viaB64.GetVersion())
	}
}

// Golden vector from the Rust SDK (solana-message 4.4.0), case "a_empty_config".
func TestV1_GoldenMessageBytesFromRustSDK(t *testing.T) {
	const msgHex = "8101000200000000abababababababababababababababababababababababababababababababab01048a88e3dd7409f195fd52db2d3cba5d72ca6709bf1d94121bf3748801b40f6f5c02020202020202020202020202020202020202020202020202020202020202020303030303030303030303030303030303030303030303030303030303030303101010101010101010101010101010101010101010101010101010101010101003030400000102deadbeef"
	raw, err := hex.DecodeString(msgHex)
	require.NoError(t, err)

	msg := mustUnmarshalV1(t, raw)
	assert.Equal(t, MessageVersionV1, msg.GetVersion())
	assert.Equal(t, MessageHeader{NumRequiredSignatures: 1, NumReadonlySignedAccounts: 0, NumReadonlyUnsignedAccounts: 2}, msg.Header)
	require.Len(t, msg.AccountKeys, 4)
	assert.Equal(t, "AKnL4NNf3DGWZJS6cPknBuEGnVsV4A4m5tgebLHaRSZ9", msg.AccountKeys[0].String())
	assert.Equal(t, "8qbHbw2BbbTHBW1sbeqakYXVKRQM8Ne7pLK7m6CVfeR", msg.AccountKeys[1].String())
	assert.Equal(t, "CktRuQ2mttgRGkXJtyksdKHjUdc2C4TgDzyB98oEzy8", msg.AccountKeys[2].String())
	assert.Equal(t, "25hjHpTATmkdET17ynDhf1MCuYNDn1z7wXfVw5iaxLAK", msg.AccountKeys[3].String())
	assert.True(t, msg.TransactionConfig.IsEmpty())
	require.Len(t, msg.Instructions, 1)
	assert.Equal(t, CompiledInstruction{ProgramIDIndex: 3, Accounts: []uint16{0, 1, 2}, Data: []byte{0xDE, 0xAD, 0xBE, 0xEF}}, msg.Instructions[0])
	require.NoError(t, msg.Sanitize())

	again := mustMarshalV1(t, &msg)
	assert.Equal(t, raw, again)
}
