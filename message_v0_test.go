package solana

import (
	"testing"

	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUniqueKey() PublicKey {
	pk, err := NewRandomPrivateKey()
	if err != nil {
		panic(err)
	}
	return pk.PublicKey()
}

func TestGetAddressTableLookupAccounts_SingleTable(t *testing.T) {
	keys := [6]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}
	tableKey := newUniqueKey()

	msg := Message{}
	msg.version = MessageVersionV0
	msg.AccountKeys = PublicKeySlice{keys[0], keys[1], keys[2]}
	msg.AddressTableLookups = []MessageAddressTableLookup{
		{
			AccountKey:      tableKey,
			WritableIndexes: []uint8{0, 1},
			ReadonlyIndexes: []uint8{2},
		},
	}
	err := msg.SetAddressTables(map[PublicKey]PublicKeySlice{
		tableKey: {keys[3], keys[4], keys[5]},
	})
	require.NoError(t, err)

	resolved, err := msg.GetAddressTableLookupAccounts()
	require.NoError(t, err)

	require.Equal(t, PublicKeySlice{keys[3], keys[4], keys[5]}, resolved)

	allKeys, err := msg.GetAllKeys()
	require.NoError(t, err)
	require.Equal(t, PublicKeySlice{keys[0], keys[1], keys[2], keys[3], keys[4], keys[5]}, allKeys)
}

func TestGetAddressTableLookupAccounts_MultipleTables(t *testing.T) {
	keys := [8]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}
	tableKeyA := newUniqueKey()
	tableKeyB := newUniqueKey()

	msg := Message{}
	msg.version = MessageVersionV0
	msg.AccountKeys = PublicKeySlice{keys[0], keys[1]}
	msg.AddressTableLookups = []MessageAddressTableLookup{
		{
			AccountKey:      tableKeyA,
			WritableIndexes: []uint8{0, 1},
			ReadonlyIndexes: []uint8{2},
		},
		{
			AccountKey:      tableKeyB,
			WritableIndexes: []uint8{0},
			ReadonlyIndexes: []uint8{1, 2},
		},
	}
	err := msg.SetAddressTables(map[PublicKey]PublicKeySlice{
		tableKeyA: {keys[2], keys[3], keys[4]},
		tableKeyB: {keys[5], keys[6], keys[7]},
	})
	require.NoError(t, err)

	resolved, err := msg.GetAddressTableLookupAccounts()
	require.NoError(t, err)

	expected := PublicKeySlice{keys[2], keys[3], keys[5], keys[4], keys[6], keys[7]}
	require.Equal(t, expected, resolved)

	allKeys, err := msg.GetAllKeys()
	require.NoError(t, err)
	require.Equal(t, PublicKeySlice{keys[0], keys[1], keys[2], keys[3], keys[5], keys[4], keys[6], keys[7]}, allKeys)
}

func TestGetAddressTableLookupAccounts_MissingTable(t *testing.T) {
	msg := Message{}
	msg.version = MessageVersionV0
	msg.AccountKeys = PublicKeySlice{newUniqueKey()}
	msg.AddressTableLookups = []MessageAddressTableLookup{
		{
			AccountKey:      newUniqueKey(),
			WritableIndexes: []uint8{0},
			ReadonlyIndexes: []uint8{},
		},
	}
	err := msg.SetAddressTables(map[PublicKey]PublicKeySlice{})
	require.NoError(t, err)

	_, err = msg.GetAddressTableLookupAccounts()
	require.Error(t, err)
}

func TestGetAddressTableLookupAccounts_IndexOutOfRange(t *testing.T) {
	tableKey := newUniqueKey()

	msg := Message{}
	msg.version = MessageVersionV0
	msg.AccountKeys = PublicKeySlice{newUniqueKey()}
	msg.AddressTableLookups = []MessageAddressTableLookup{
		{
			AccountKey:      tableKey,
			WritableIndexes: []uint8{5}, // out of range
			ReadonlyIndexes: []uint8{},
		},
	}
	err := msg.SetAddressTables(map[PublicKey]PublicKeySlice{
		tableKey: {newUniqueKey(), newUniqueKey()}, // only 2 entries
	})
	require.NoError(t, err)

	_, err = msg.GetAddressTableLookupAccounts()
	require.Error(t, err)
	require.Contains(t, err.Error(), "index out of range")
}

func TestResolveLookups_AppendsToAccountKeys(t *testing.T) {
	keys := [6]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}
	tableKey := newUniqueKey()

	msg := Message{}
	msg.version = MessageVersionV0
	msg.AccountKeys = PublicKeySlice{keys[0], keys[1], keys[2], keys[3]}
	msg.Header = MessageHeader{
		NumRequiredSignatures:       2,
		NumReadonlySignedAccounts:   1,
		NumReadonlyUnsignedAccounts: 1,
	}
	msg.AddressTableLookups = []MessageAddressTableLookup{
		{
			AccountKey:      tableKey,
			WritableIndexes: []uint8{0},
			ReadonlyIndexes: []uint8{1},
		},
	}
	err := msg.SetAddressTables(map[PublicKey]PublicKeySlice{
		tableKey: {keys[4], keys[5]},
	})
	require.NoError(t, err)

	err = msg.ResolveLookups()
	require.NoError(t, err)
	require.True(t, msg.IsResolved())

	require.Equal(t,
		PublicKeySlice{keys[0], keys[1], keys[2], keys[3], keys[4], keys[5]},
		msg.AccountKeys,
	)
}

func TestResolveLookups_Idempotent(t *testing.T) {
	tableKey := newUniqueKey()

	msg := Message{}
	msg.version = MessageVersionV0
	msg.AccountKeys = PublicKeySlice{newUniqueKey()}
	msg.AddressTableLookups = []MessageAddressTableLookup{
		{
			AccountKey:      tableKey,
			WritableIndexes: []uint8{0},
			ReadonlyIndexes: []uint8{},
		},
	}
	err := msg.SetAddressTables(map[PublicKey]PublicKeySlice{
		tableKey: {newUniqueKey()},
	})
	require.NoError(t, err)

	err = msg.ResolveLookups()
	require.NoError(t, err)
	count := len(msg.AccountKeys)

	// calling again should not append more
	err = msg.ResolveLookups()
	require.NoError(t, err)
	require.Equal(t, count, len(msg.AccountKeys))
}

func TestResolveLookupsWith_MultipleTables(t *testing.T) {
	staticKeys := PublicKeySlice{newUniqueKey(), newUniqueKey()}
	writable := PublicKeySlice{newUniqueKey(), newUniqueKey()}
	readonly := PublicKeySlice{newUniqueKey()}

	msg := Message{}
	msg.version = MessageVersionV0
	msg.AccountKeys = staticKeys

	err := msg.ResolveLookupsWith(writable, readonly)
	require.NoError(t, err)
	require.True(t, msg.IsResolved())

	expected := append(append(PublicKeySlice{}, staticKeys...), append(writable, readonly...)...)
	require.Equal(t, expected, msg.AccountKeys)

	// calling again should error
	err = msg.ResolveLookupsWith(writable, readonly)
	require.ErrorIs(t, err, ErrAlreadyResolved)
}

// Mirrors the Rust loaded.rs test_is_writable_index test.
// Setup: 4 static keys (2 signers: 1 writable + 1 readonly, 2 unsigned: 1 writable + 1 readonly),
// 1 writable lookup, 1 readonly lookup.
// Expected writability by index:
//
//	0: writable signer     -> true
//	1: readonly signer     -> false
//	2: writable unsigned   -> true
//	3: readonly unsigned   -> false
//	4: writable lookup     -> true
//	5: readonly lookup     -> false
func TestIsWritableIndex_WithLookups(t *testing.T) {
	keys := [6]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}
	tableKey := newUniqueKey()

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: PublicKeySlice{keys[0], keys[1], keys[2], keys[3]},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      tableKey,
				WritableIndexes: []uint8{0},
				ReadonlyIndexes: []uint8{1},
			},
		},
	}
	msg.version = MessageVersionV0
	err := msg.SetAddressTables(map[PublicKey]PublicKeySlice{
		tableKey: {keys[4], keys[5]},
	})
	require.NoError(t, err)

	err = msg.ResolveLookups()
	require.NoError(t, err)

	writable0, err := msg.IsWritable(keys[0])
	require.NoError(t, err)
	assert.True(t, writable0, "key0: writable signer")

	writable1, err := msg.IsWritable(keys[1])
	require.NoError(t, err)
	assert.False(t, writable1, "key1: readonly signer")

	writable2, err := msg.IsWritable(keys[2])
	require.NoError(t, err)
	assert.True(t, writable2, "key2: writable unsigned")

	writable3, err := msg.IsWritable(keys[3])
	require.NoError(t, err)
	assert.False(t, writable3, "key3: readonly unsigned")

	writable4, err := msg.IsWritable(keys[4])
	require.NoError(t, err)
	assert.True(t, writable4, "key4: writable lookup")

	writable5, err := msg.IsWritable(keys[5])
	require.NoError(t, err)
	assert.False(t, writable5, "key5: readonly lookup")
}

// Mirrors the Rust mod.rs test_is_maybe_writable test.
// Setup: 6 static keys, header: 3 signers (2 readonly), 1 readonly unsigned,
// 1 writable lookup, 1 readonly lookup.
// Expected:
//
//	idx 0 (key0): writable signer   -> true
//	idx 1 (key1): readonly signer   -> false
//	idx 2 (key2): readonly signer   -> false
//	idx 3 (key3): writable unsigned -> true
//	idx 4 (key4): writable unsigned -> true
//	idx 5 (key5): readonly unsigned -> false
//	idx 6: writable lookup          -> true
//	idx 7: readonly lookup          -> false
func TestIsWritable_WithLookupsStaticHeader(t *testing.T) {
	keys := [6]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}
	tableKey := newUniqueKey()
	lookupWritable := newUniqueKey()
	lookupReadonly := newUniqueKey()

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       3,
			NumReadonlySignedAccounts:   2,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: PublicKeySlice{keys[0], keys[1], keys[2], keys[3], keys[4], keys[5]},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      tableKey,
				WritableIndexes: []uint8{0},
				ReadonlyIndexes: []uint8{1},
			},
		},
	}
	msg.version = MessageVersionV0
	err := msg.SetAddressTables(map[PublicKey]PublicKeySlice{
		tableKey: {lookupWritable, lookupReadonly},
	})
	require.NoError(t, err)

	err = msg.ResolveLookups()
	require.NoError(t, err)

	tests := []struct {
		key      PublicKey
		name     string
		writable bool
	}{
		{keys[0], "writable signer", true},
		{keys[1], "readonly signer 1", false},
		{keys[2], "readonly signer 2", false},
		{keys[3], "writable unsigned 1", true},
		{keys[4], "writable unsigned 2", true},
		{keys[5], "readonly unsigned", false},
		{lookupWritable, "writable lookup", true},
		{lookupReadonly, "readonly lookup", false},
	}

	for _, tt := range tests {
		w, err := msg.IsWritable(tt.key)
		require.NoError(t, err)
		assert.Equal(t, tt.writable, w, tt.name)
	}
}

// Mirrors Rust test with multiple tables to verify writable boundary across tables.
func TestIsWritable_MultipleTableLookups(t *testing.T) {
	staticKey := newUniqueKey()
	tableKeyA := newUniqueKey()
	tableKeyB := newUniqueKey()

	wA0 := newUniqueKey()
	wA1 := newUniqueKey()
	rA0 := newUniqueKey()
	wB0 := newUniqueKey()
	rB0 := newUniqueKey()
	rB1 := newUniqueKey()

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 0,
		},
		AccountKeys: PublicKeySlice{staticKey},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      tableKeyA,
				WritableIndexes: []uint8{0, 1},
				ReadonlyIndexes: []uint8{2},
			},
			{
				AccountKey:      tableKeyB,
				WritableIndexes: []uint8{0},
				ReadonlyIndexes: []uint8{1, 2},
			},
		},
	}
	msg.version = MessageVersionV0
	err := msg.SetAddressTables(map[PublicKey]PublicKeySlice{
		tableKeyA: {wA0, wA1, rA0},
		tableKeyB: {wB0, rB0, rB1},
	})
	require.NoError(t, err)

	err = msg.ResolveLookups()
	require.NoError(t, err)

	// Resolved order: static, then writable(A)+writable(B), then readonly(A)+readonly(B)
	// Total NumWritableLookups = 2 + 1 = 3
	require.Equal(t,
		PublicKeySlice{staticKey, wA0, wA1, wB0, rA0, rB0, rB1},
		msg.AccountKeys,
	)

	tests := []struct {
		key      PublicKey
		name     string
		writable bool
	}{
		{staticKey, "static signer", true},
		{wA0, "writable from table A [0]", true},
		{wA1, "writable from table A [1]", true},
		{wB0, "writable from table B [0]", true},
		{rA0, "readonly from table A", false},
		{rB0, "readonly from table B [0]", false},
		{rB1, "readonly from table B [1]", false},
	}
	for _, tt := range tests {
		w, err := msg.IsWritable(tt.key)
		require.NoError(t, err)
		assert.Equal(t, tt.writable, w, tt.name)
	}
}

func TestNumStaticAccounts_BeforeAndAfterResolve(t *testing.T) {
	tableKey := newUniqueKey()

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 0,
		},
		AccountKeys: PublicKeySlice{newUniqueKey(), newUniqueKey(), newUniqueKey()},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      tableKey,
				WritableIndexes: []uint8{0, 1},
				ReadonlyIndexes: []uint8{2},
			},
		},
	}
	msg.version = MessageVersionV0
	err := msg.SetAddressTables(map[PublicKey]PublicKeySlice{
		tableKey: {newUniqueKey(), newUniqueKey(), newUniqueKey()},
	})
	require.NoError(t, err)

	require.Equal(t, 3, msg.numStaticAccounts())
	require.Equal(t, 3, msg.NumLookups())
	require.Equal(t, 2, msg.NumWritableLookups())

	err = msg.ResolveLookups()
	require.NoError(t, err)

	require.Equal(t, 6, len(msg.AccountKeys))
	require.Equal(t, 3, msg.numStaticAccounts())
}

func TestMarshalUnmarshalV0_MultipleLookupTables(t *testing.T) {
	keys := [4]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}
	tableKeyA := newUniqueKey()
	tableKeyB := newUniqueKey()

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 0,
		},
		RecentBlockhash: Hash{1, 2, 3},
		AccountKeys:     PublicKeySlice{keys[0]},
		Instructions: []CompiledInstruction{
			{
				ProgramIDIndex: 0,
				Accounts:       []uint16{0},
				Data:           []byte{0xAA},
			},
		},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      tableKeyA,
				WritableIndexes: []uint8{0, 1},
				ReadonlyIndexes: []uint8{2},
			},
			{
				AccountKey:      tableKeyB,
				WritableIndexes: []uint8{0},
				ReadonlyIndexes: []uint8{1, 2},
			},
		},
	}
	msg.version = MessageVersionV0

	data, err := msg.MarshalBinary()
	require.NoError(t, err)

	var decoded Message
	err = decoded.UnmarshalWithDecoder(bin.NewBinDecoder(data))
	require.NoError(t, err)

	require.Equal(t, MessageVersionV0, decoded.GetVersion())
	require.Equal(t, msg.Header, decoded.Header)
	require.Equal(t, msg.AccountKeys, decoded.AccountKeys)
	require.Equal(t, msg.RecentBlockhash, decoded.RecentBlockhash)
	require.Equal(t, len(msg.Instructions), len(decoded.Instructions))
	require.Equal(t, msg.Instructions[0].ProgramIDIndex, decoded.Instructions[0].ProgramIDIndex)
	require.Equal(t, msg.Instructions[0].Data, decoded.Instructions[0].Data)

	require.Equal(t, 2, len(decoded.AddressTableLookups))
	require.Equal(t, tableKeyA, decoded.AddressTableLookups[0].AccountKey)
	require.Equal(t, Uint8SliceAsNum{0, 1}, decoded.AddressTableLookups[0].WritableIndexes)
	require.Equal(t, Uint8SliceAsNum{2}, decoded.AddressTableLookups[0].ReadonlyIndexes)
	require.Equal(t, tableKeyB, decoded.AddressTableLookups[1].AccountKey)
	require.Equal(t, Uint8SliceAsNum{0}, decoded.AddressTableLookups[1].WritableIndexes)
	require.Equal(t, Uint8SliceAsNum{1, 2}, decoded.AddressTableLookups[1].ReadonlyIndexes)
}

func TestNewTransaction_DeterministicWithMultipleLookupTables(t *testing.T) {
	payer := newUniqueKey()
	programID := newUniqueKey()

	acctA := newUniqueKey()
	acctB := newUniqueKey()
	acctC := newUniqueKey()
	acctD := newUniqueKey()

	tableKeyA := newUniqueKey()
	tableKeyB := newUniqueKey()

	tables := map[PublicKey]PublicKeySlice{
		tableKeyA: {acctA, acctB},
		tableKeyB: {acctC, acctD},
	}

	instructions := []Instruction{
		NewInstruction(
			programID,
			AccountMetaSlice{
				Meta(acctA).WRITE(),
				Meta(acctB),
				Meta(acctC).WRITE(),
				Meta(acctD),
			},
			[]byte{0x01},
		),
	}

	var firstBytes []byte
	for i := 0; i < 20; i++ {
		tx, err := NewTransaction(
			instructions,
			Hash{},
			TransactionPayer(payer),
			TransactionAddressTables(tables),
		)
		require.NoError(t, err)

		data, err := tx.Message.MarshalBinary()
		require.NoError(t, err)

		if i == 0 {
			firstBytes = data
		} else {
			require.Equal(t, firstBytes, data, "transaction bytes differ on iteration %d", i)
		}
	}
}

func TestNewTransaction_MultipleLookupTables_Resolution(t *testing.T) {
	payer := newUniqueKey()
	programID := newUniqueKey()

	acctA := newUniqueKey()
	acctB := newUniqueKey()

	tableKeyA := newUniqueKey()
	tableKeyB := newUniqueKey()

	tables := map[PublicKey]PublicKeySlice{
		tableKeyA: {acctA},
		tableKeyB: {acctB},
	}

	instructions := []Instruction{
		NewInstruction(
			programID,
			AccountMetaSlice{
				Meta(acctA).WRITE(),
				Meta(acctB),
			},
			[]byte{0x01},
		),
	}

	tx, err := NewTransaction(
		instructions,
		Hash{},
		TransactionPayer(payer),
		TransactionAddressTables(tables),
	)
	require.NoError(t, err)

	require.True(t, tx.Message.IsVersioned())

	lookups := tx.Message.GetAddressTableLookups()
	require.Equal(t, 2, len(lookups))

	err = tx.Message.ResolveLookups()
	require.NoError(t, err)

	allKeys, err := tx.Message.GetAllKeys()
	require.NoError(t, err)

	found := map[PublicKey]bool{}
	for _, k := range allKeys {
		found[k] = true
	}
	assert.True(t, found[acctA], "acctA should be in resolved keys")
	assert.True(t, found[acctB], "acctB should be in resolved keys")
	assert.True(t, found[payer], "payer should be in resolved keys")
	assert.True(t, found[programID], "programID should be in resolved keys")

	require.Equal(t, 1, len(tx.Message.Instructions))
	ix := tx.Message.Instructions[0]
	require.Equal(t, 2, len(ix.Accounts))

	resolvedAcct0 := allKeys[ix.Accounts[0]]
	resolvedAcct1 := allKeys[ix.Accounts[1]]
	assert.Equal(t, acctA, resolvedAcct0, "first ix account should be acctA")
	assert.Equal(t, acctB, resolvedAcct1, "second ix account should be acctB")
}
