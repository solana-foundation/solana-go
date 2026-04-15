package solana

import (
	"testing"
	"unsafe"

	bin "github.com/gagliardetto/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ported from solana-sdk/message/src/versions/mod.rs: test_legacy_message_serialization
func TestLegacyMessageSerializationRoundtrip(t *testing.T) {
	key0 := newUniqueKey()
	key1 := newUniqueKey()
	key2 := newUniqueKey()
	blockhash := Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys:     PublicKeySlice{key0, key1, key2},
		RecentBlockhash: blockhash,
		Instructions: []CompiledInstruction{
			{
				ProgramIDIndex: 2,
				Accounts:       []uint16{0, 1},
				Data:           []byte{0xAA, 0xBB},
			},
		},
	}
	msg.version = MessageVersionLegacy

	data, err := msg.MarshalBinary()
	require.NoError(t, err)

	// First byte should be numRequiredSignatures for legacy messages.
	assert.Equal(t, byte(2), data[0], "first byte should be numRequiredSignatures")

	var decoded Message
	err = decoded.UnmarshalWithDecoder(bin.NewBinDecoder(data))
	require.NoError(t, err)

	assert.Equal(t, MessageVersionLegacy, decoded.GetVersion())
	assert.Equal(t, msg.Header, decoded.Header)
	assert.Equal(t, msg.AccountKeys, decoded.AccountKeys)
	assert.Equal(t, msg.RecentBlockhash, decoded.RecentBlockhash)
	require.Equal(t, len(msg.Instructions), len(decoded.Instructions))
	assert.Equal(t, msg.Instructions[0].ProgramIDIndex, decoded.Instructions[0].ProgramIDIndex)
	assert.Equal(t, msg.Instructions[0].Accounts, decoded.Instructions[0].Accounts)
	assert.Equal(t, Base58(msg.Instructions[0].Data), decoded.Instructions[0].Data)
}

// Ported from solana-sdk/message/src/versions/mod.rs: test_versioned_message_serialization
func TestV0MessageSerializationRoundtrip(t *testing.T) {
	key0 := newUniqueKey()
	tableKey0 := newUniqueKey()
	tableKey1 := newUniqueKey()
	blockhash := Hash{9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 1, 2, 3, 4, 5, 6,
		7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22}

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 0,
		},
		AccountKeys:     PublicKeySlice{key0},
		RecentBlockhash: blockhash,
		Instructions: []CompiledInstruction{
			{
				ProgramIDIndex: 1,
				Accounts:       []uint16{0, 2, 3, 4},
				Data:           []byte{},
			},
		},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      tableKey0,
				WritableIndexes: []uint8{1},
				ReadonlyIndexes: []uint8{0},
			},
			{
				AccountKey:      tableKey1,
				WritableIndexes: []uint8{0},
				ReadonlyIndexes: []uint8{1},
			},
		},
	}
	msg.version = MessageVersionV0

	data, err := msg.MarshalBinary()
	require.NoError(t, err)

	// First byte must have high bit set for versioned messages (0x80).
	assert.Equal(t, byte(0x80), data[0], "first byte should be version prefix 0x80 for V0")

	var decoded Message
	err = decoded.UnmarshalWithDecoder(bin.NewBinDecoder(data))
	require.NoError(t, err)

	assert.Equal(t, MessageVersionV0, decoded.GetVersion())
	assert.Equal(t, msg.Header, decoded.Header)
	assert.Equal(t, msg.AccountKeys, decoded.AccountKeys)
	assert.Equal(t, msg.RecentBlockhash, decoded.RecentBlockhash)
	require.Equal(t, len(msg.Instructions), len(decoded.Instructions))
	assert.Equal(t, msg.Instructions[0].ProgramIDIndex, decoded.Instructions[0].ProgramIDIndex)
	assert.Equal(t, msg.Instructions[0].Accounts, decoded.Instructions[0].Accounts)

	require.Equal(t, 2, len(decoded.AddressTableLookups))
	assert.Equal(t, tableKey0, decoded.AddressTableLookups[0].AccountKey)
	assert.Equal(t, Uint8SliceAsNum{1}, decoded.AddressTableLookups[0].WritableIndexes)
	assert.Equal(t, Uint8SliceAsNum{0}, decoded.AddressTableLookups[0].ReadonlyIndexes)
	assert.Equal(t, tableKey1, decoded.AddressTableLookups[1].AccountKey)
	assert.Equal(t, Uint8SliceAsNum{0}, decoded.AddressTableLookups[1].WritableIndexes)
	assert.Equal(t, Uint8SliceAsNum{1}, decoded.AddressTableLookups[1].ReadonlyIndexes)
}

// Tests the version prefix detection logic ported from solana-sdk/message/src/versions/mod.rs.
// In Rust: MESSAGE_VERSION_PREFIX = 0x80; if first_byte & 0x80 != 0 → versioned.
// This specifically tests the bug fix where byte value 127 (0x7F) was incorrectly
// classified as versioned — it should be legacy (numRequiredSignatures = 127).
func TestVersionDetection_PrefixByte(t *testing.T) {
	tests := []struct {
		name            string
		firstByte       byte
		expectedVersion MessageVersion
	}{
		{"numRequiredSignatures=1 is legacy", 1, MessageVersionLegacy},
		{"numRequiredSignatures=64 is legacy", 64, MessageVersionLegacy},
		{"numRequiredSignatures=126 is legacy", 126, MessageVersionLegacy},
		{"numRequiredSignatures=127 is legacy", 127, MessageVersionLegacy}, // was buggy before fix
		{"0x80 is V0", 0x80, MessageVersionV0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a minimal valid message with the given first byte.
			// For legacy: first byte is numRequiredSignatures.
			// For V0: first byte is 0x80 | version.
			var buf []byte
			if tt.expectedVersion == MessageVersionLegacy {
				// legacy: header(3) + compact_len(1) + key(32) + blockhash(32) + compact_len(1)=0 instructions
				buf = make([]byte, 0, 69)
				buf = append(buf, tt.firstByte, 0, 0) // header
				buf = append(buf, 1)                   // 1 account key
				buf = append(buf, make([]byte, 32)...) // account key
				buf = append(buf, make([]byte, 32)...) // blockhash
				buf = append(buf, 0)                   // 0 instructions
			} else {
				// v0: prefix(1) + header(3) + compact_len(1) + key(32) + blockhash(32) + compact_len(1) + compact_len(1) lookups
				buf = make([]byte, 0, 71)
				buf = append(buf, tt.firstByte)        // version prefix
				buf = append(buf, 1, 0, 0)             // header
				buf = append(buf, 1)                    // 1 account key
				buf = append(buf, make([]byte, 32)...)  // account key
				buf = append(buf, make([]byte, 32)...)  // blockhash
				buf = append(buf, 0)                    // 0 instructions
				buf = append(buf, 0)                    // 0 address table lookups
			}

			var msg Message
			err := msg.UnmarshalWithDecoder(bin.NewBinDecoder(buf))
			require.NoError(t, err)
			assert.Equal(t, tt.expectedVersion, msg.GetVersion())
		})
	}
}

// Tests that unsupported version numbers (> 0) in versioned messages are rejected.
// Ported from solana-sdk/message/src/versions/v0/mod.rs version validation.
func TestVersionDetection_UnsupportedVersion(t *testing.T) {
	// 0x81 = messageVersionPrefix | 1 → version 1 (unsupported)
	buf := []byte{0x81, 1, 0, 0}
	buf = append(buf, 1)                   // 1 account key
	buf = append(buf, make([]byte, 32)...) // account key
	buf = append(buf, make([]byte, 32)...) // blockhash
	buf = append(buf, 0)                   // 0 instructions
	buf = append(buf, 0)                   // 0 lookups

	var msg Message
	err := msg.UnmarshalWithDecoder(bin.NewBinDecoder(buf))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported message version")
}

// Ported from solana-sdk/message/src/legacy.rs: test_is_writable_index_saturating_behavior.
// Tests edge cases where header values exceed the number of account keys.
func TestIsWritable_SaturatingBehavior(t *testing.T) {
	// Case 1: num_readonly_signed (2) > num_required_signatures (1)
	// Index 0 is signed but readonly count exceeds signature count → not writable.
	key0 := newUniqueKey()
	msg1 := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   2,
			NumReadonlyUnsignedAccounts: 0,
		},
		AccountKeys: PublicKeySlice{key0},
	}
	w, err := msg1.IsWritable(key0)
	require.NoError(t, err)
	assert.False(t, w, "case 1: readonly signed exceeds required signatures")

	// Case 2: num_readonly_unsigned (2) > num unsigned accounts (1)
	// Only 1 account, 0 signers, all are unsigned but readonly count exceeds → not writable.
	key1 := newUniqueKey()
	msg2 := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       0,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 2,
		},
		AccountKeys: PublicKeySlice{key1},
	}
	w, err = msg2.IsWritable(key1)
	require.NoError(t, err)
	assert.False(t, w, "case 2: readonly unsigned exceeds unsigned accounts")

	// Case 3: 1 signer, 0 readonly signed, 2 readonly unsigned but only 1 account.
	// Index 0 is a writable signer.
	msg3 := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 2,
		},
		AccountKeys: PublicKeySlice{key0},
	}
	w, err = msg3.IsWritable(key0)
	require.NoError(t, err)
	assert.True(t, w, "case 3: signer with no readonly signed is writable")

	// Case 4: 2 accounts, 1 signer, 0 readonly signed, 3 readonly unsigned.
	// Index 0: writable signer; index 1: unsigned but readonly exceeds → not writable.
	key2 := newUniqueKey()
	msg4 := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 3,
		},
		AccountKeys: PublicKeySlice{key0, key2},
	}
	w, err = msg4.IsWritable(key0)
	require.NoError(t, err)
	assert.True(t, w, "case 4: key0 writable signer")
	w, err = msg4.IsWritable(key2)
	require.NoError(t, err)
	assert.False(t, w, "case 4: key1 readonly unsigned")

	// Case 5: 2 accounts, 1 signer with 2 readonly signed, 3 readonly unsigned.
	// Both accounts should be readonly.
	msg5 := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   2,
			NumReadonlyUnsignedAccounts: 3,
		},
		AccountKeys: PublicKeySlice{key0, key2},
	}
	w, err = msg5.IsWritable(key0)
	require.NoError(t, err)
	assert.False(t, w, "case 5: key0 readonly (readonly_signed exceeds required)")
	w, err = msg5.IsWritable(key2)
	require.NoError(t, err)
	assert.False(t, w, "case 5: key1 readonly unsigned")
}

// Ported from solana-sdk/message/src/legacy.rs: test_is_maybe_writable.
// Tests the standard writability layout:
//
//	Header: 3 signers (2 readonly), 1 readonly unsigned → 6 accounts total.
//	idx 0: writable signer
//	idx 1: readonly signer
//	idx 2: readonly signer
//	idx 3: writable unsigned
//	idx 4: writable unsigned
//	idx 5: readonly unsigned
func TestIsWritable_StandardLayout(t *testing.T) {
	keys := [6]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       3,
			NumReadonlySignedAccounts:   2,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: PublicKeySlice{keys[0], keys[1], keys[2], keys[3], keys[4], keys[5]},
	}

	expected := []bool{true, false, false, true, true, false}
	for i, key := range keys {
		w, err := msg.IsWritable(key)
		require.NoError(t, err)
		assert.Equal(t, expected[i], w, "index %d", i)
	}
}

// Ported from solana-sdk/message/src/legacy.rs: test_message_signed_keys_len.
func TestIsSigner(t *testing.T) {
	key0 := newUniqueKey()
	key1 := newUniqueKey()
	programID := newUniqueKey()

	// No signers.
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures: 0,
		},
		AccountKeys: PublicKeySlice{key0, programID},
	}
	assert.False(t, msg.IsSigner(key0))

	// One signer.
	msg = Message{
		Header: MessageHeader{
			NumRequiredSignatures: 1,
		},
		AccountKeys: PublicKeySlice{key0, programID},
	}
	assert.True(t, msg.IsSigner(key0))
	assert.False(t, msg.IsSigner(programID))

	// Two signers.
	msg = Message{
		Header: MessageHeader{
			NumRequiredSignatures: 2,
		},
		AccountKeys: PublicKeySlice{key0, key1, programID},
	}
	assert.True(t, msg.IsSigner(key0))
	assert.True(t, msg.IsSigner(key1))
	assert.False(t, msg.IsSigner(programID))

	// Unknown key is not a signer.
	assert.False(t, msg.IsSigner(newUniqueKey()))
}

// Ported from solana-sdk/message/src/legacy.rs: test_program_ids.
func TestProgramIDs(t *testing.T) {
	key0 := newUniqueKey()
	key1 := newUniqueKey()
	programID := newUniqueKey()

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 1, // programID at index 2 is readonly unsigned
		},
		AccountKeys: PublicKeySlice{key0, key1, programID},
		Instructions: []CompiledInstruction{
			{
				ProgramIDIndex: 2,
				Accounts:       []uint16{0, 1},
				Data:           []byte{},
			},
		},
	}

	resolved, err := msg.Program(2)
	require.NoError(t, err)
	assert.Equal(t, programID, resolved)

	_, err = msg.Program(3) // out of range
	require.Error(t, err)
}

// Tests that IsWritableStatic only considers static accounts, ignoring lookups.
func TestIsWritableStatic_IgnoresLookups(t *testing.T) {
	keys := [4]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: PublicKeySlice{keys[0], keys[1], keys[2], keys[3]},
	}
	msg.version = MessageVersionV0

	assert.True(t, msg.IsWritableStatic(keys[0]), "writable signer")
	assert.False(t, msg.IsWritableStatic(keys[1]), "readonly signer")
	assert.True(t, msg.IsWritableStatic(keys[2]), "writable unsigned")
	assert.False(t, msg.IsWritableStatic(keys[3]), "readonly unsigned")
	assert.False(t, msg.IsWritableStatic(newUniqueKey()), "unknown key")
}

// Tests JSON serialization roundtrip for both legacy and V0 messages.
func TestMessageJSONRoundtrip(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		msg := Message{
			Header: MessageHeader{
				NumRequiredSignatures:       1,
				NumReadonlySignedAccounts:   0,
				NumReadonlyUnsignedAccounts: 1,
			},
			AccountKeys:     PublicKeySlice{newUniqueKey(), newUniqueKey()},
			RecentBlockhash: Hash{1, 2, 3},
			Instructions: []CompiledInstruction{
				{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{0xFF}},
			},
		}
		msg.version = MessageVersionLegacy

		data, err := msg.MarshalJSON()
		require.NoError(t, err)
		require.NotEmpty(t, data)

		// Should not contain addressTableLookups for legacy.
		assert.NotContains(t, string(data), "addressTableLookups")
	})

	t.Run("v0", func(t *testing.T) {
		msg := Message{
			Header: MessageHeader{
				NumRequiredSignatures:       1,
				NumReadonlySignedAccounts:   0,
				NumReadonlyUnsignedAccounts: 0,
			},
			AccountKeys:     PublicKeySlice{newUniqueKey()},
			RecentBlockhash: Hash{4, 5, 6},
			Instructions: []CompiledInstruction{
				{ProgramIDIndex: 0, Accounts: []uint16{0}, Data: []byte{0x01}},
			},
			AddressTableLookups: []MessageAddressTableLookup{
				{
					AccountKey:      newUniqueKey(),
					WritableIndexes: []uint8{0},
					ReadonlyIndexes: []uint8{1},
				},
			},
		}
		msg.version = MessageVersionV0

		data, err := msg.MarshalJSON()
		require.NoError(t, err)
		assert.Contains(t, string(data), "addressTableLookups")
	})
}

// Tests that the V0 prefix byte is exactly 0x80 for version 0,
// matching Rust's MESSAGE_VERSION_PREFIX | 0 = 0x80.
func TestMarshalV0_PrefixByte(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 0,
		},
		AccountKeys:     PublicKeySlice{newUniqueKey()},
		RecentBlockhash: Hash{},
		Instructions:    []CompiledInstruction{},
	}
	msg.version = MessageVersionV0

	data, err := msg.MarshalBinary()
	require.NoError(t, err)
	assert.Equal(t, byte(0x80), data[0])

	// Second byte should be numRequiredSignatures.
	assert.Equal(t, byte(1), data[1])
}

// Tests that legacy message does NOT have the 0x80 prefix.
func TestMarshalLegacy_NoPrefixByte(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       3,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys:     PublicKeySlice{newUniqueKey(), newUniqueKey(), newUniqueKey()},
		RecentBlockhash: Hash{},
		Instructions:    []CompiledInstruction{},
	}
	msg.version = MessageVersionLegacy

	data, err := msg.MarshalBinary()
	require.NoError(t, err)

	// First byte is numRequiredSignatures directly, not a version prefix.
	assert.Equal(t, byte(3), data[0])
	assert.Equal(t, byte(0), data[0]&0x80, "high bit should not be set for legacy")
}

// Tests Account() method for both static and resolved lookup accounts.
func TestAccount_StaticAndLookup(t *testing.T) {
	keys := [4]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}
	tableKey := newUniqueKey()
	lookupKey := newUniqueKey()

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 0,
		},
		AccountKeys: PublicKeySlice{keys[0], keys[1]},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      tableKey,
				WritableIndexes: []uint8{0},
				ReadonlyIndexes: []uint8{},
			},
		},
	}
	msg.version = MessageVersionV0
	err := msg.SetAddressTables(map[PublicKey]PublicKeySlice{
		tableKey: {lookupKey},
	})
	require.NoError(t, err)

	// Static account.
	acct, err := msg.Account(0)
	require.NoError(t, err)
	assert.Equal(t, keys[0], acct)

	acct, err = msg.Account(1)
	require.NoError(t, err)
	assert.Equal(t, keys[1], acct)

	// Lookup account (index 2 = first lookup).
	acct, err = msg.Account(2)
	require.NoError(t, err)
	assert.Equal(t, lookupKey, acct)

	// Out of range.
	_, err = msg.Account(3)
	require.Error(t, err)
}

// Tests HasAccount and GetAccountIndex.
func TestHasAccountAndGetIndex(t *testing.T) {
	keys := [3]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures: 1,
		},
		AccountKeys: PublicKeySlice{keys[0], keys[1], keys[2]},
	}

	for i, key := range keys {
		has, err := msg.HasAccount(key)
		require.NoError(t, err)
		assert.True(t, has)

		idx, err := msg.GetAccountIndex(key)
		require.NoError(t, err)
		assert.Equal(t, uint16(i), idx)
	}

	unknown := newUniqueKey()
	has, err := msg.HasAccount(unknown)
	require.NoError(t, err)
	assert.False(t, has)

	_, err = msg.GetAccountIndex(unknown)
	require.Error(t, err)
}

// Tests Signers() returns only the first numRequiredSignatures accounts.
func TestSigners(t *testing.T) {
	keys := [5]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       3,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 0,
		},
		AccountKeys: PublicKeySlice{keys[0], keys[1], keys[2], keys[3], keys[4]},
	}

	signers := msg.Signers()
	require.Equal(t, 3, len(signers))
	assert.Equal(t, keys[0], signers[0])
	assert.Equal(t, keys[1], signers[1])
	assert.Equal(t, keys[2], signers[2])
}

// Tests Writable() returns only writable accounts across static and lookup.
func TestWritable(t *testing.T) {
	keys := [4]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: PublicKeySlice{keys[0], keys[1], keys[2], keys[3]},
	}

	writable, err := msg.Writable()
	require.NoError(t, err)
	// idx 0: writable signer, idx 1: readonly signer, idx 2: writable unsigned, idx 3: readonly unsigned
	require.Equal(t, 2, len(writable))
	assert.Equal(t, keys[0], writable[0])
	assert.Equal(t, keys[2], writable[1])
}

// Tests that SetVersion validates input.
func TestSetVersion_Validation(t *testing.T) {
	msg := &Message{}

	_, err := msg.SetVersion(MessageVersionLegacy)
	require.NoError(t, err)
	assert.Equal(t, MessageVersionLegacy, msg.GetVersion())

	_, err = msg.SetVersion(MessageVersionV0)
	require.NoError(t, err)
	assert.Equal(t, MessageVersionV0, msg.GetVersion())

	_, err = msg.SetVersion(MessageVersion(99))
	require.Error(t, err)
}

// Tests IsVersioned().
func TestIsVersioned(t *testing.T) {
	msg := Message{}
	msg.version = MessageVersionLegacy
	assert.False(t, msg.IsVersioned())

	msg.version = MessageVersionV0
	assert.True(t, msg.IsVersioned())
}

// Tests that SetAddressTables can only be called once.
func TestSetAddressTables_OnlyOnce(t *testing.T) {
	msg := &Message{}
	msg.version = MessageVersionV0

	err := msg.SetAddressTables(map[PublicKey]PublicKeySlice{})
	require.NoError(t, err)

	err = msg.SetAddressTables(map[PublicKey]PublicKeySlice{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already set")
}

// Tests that checkPreconditions fails when address tables are needed but not set.
func TestCheckPreconditions_MissingTables(t *testing.T) {
	msg := Message{
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      newUniqueKey(),
				WritableIndexes: []uint8{0},
				ReadonlyIndexes: []uint8{},
			},
		},
	}
	msg.version = MessageVersionV0

	_, err := msg.AccountMetaList()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "without address tables")
}

// Tests base64 roundtrip.
func TestMarshalUnmarshalBase64(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys:     PublicKeySlice{newUniqueKey(), newUniqueKey()},
		RecentBlockhash: Hash{42},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{0xDE, 0xAD}},
		},
	}
	msg.version = MessageVersionLegacy

	b64 := msg.ToBase64()
	require.NotEmpty(t, b64)

	var decoded Message
	err := decoded.UnmarshalBase64(b64)
	require.NoError(t, err)
	assert.Equal(t, msg.Header, decoded.Header)
	assert.Equal(t, msg.AccountKeys, decoded.AccountKeys)
	assert.Equal(t, msg.RecentBlockhash, decoded.RecentBlockhash)
}

// --- Tests ported from anza-xyz/solana-sdk/message ---

// Ported from legacy.rs: test_message_header_len_constant
func TestMessageHeaderLenConstant(t *testing.T) {
	// MessageHeader is 3 × uint8 = 3 bytes, matching Rust's MESSAGE_HEADER_LENGTH.
	assert.Equal(t, uintptr(3), unsafe.Sizeof(MessageHeader{}))
}

// Ported from legacy.rs: test_program_ids (extended).
// Tests ProgramIDs() returns the unique set of program IDs.
func TestProgramIDs_Unique(t *testing.T) {
	key0 := newUniqueKey()
	key1 := newUniqueKey()
	loader := newUniqueKey()

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: PublicKeySlice{key0, key1, loader},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 2, Accounts: []uint16{0, 1}, Data: []byte{}},
		},
	}

	ids := msg.ProgramIDs()
	require.Equal(t, 1, len(ids))
	assert.Equal(t, loader, ids[0])
}

// Ported from legacy.rs: test_program_ids — multiple programs.
func TestProgramIDs_Multiple(t *testing.T) {
	key0 := newUniqueKey()
	prog0 := newUniqueKey()
	prog1 := newUniqueKey()

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 2,
		},
		AccountKeys: PublicKeySlice{key0, prog0, prog1},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{}},
			{ProgramIDIndex: 2, Accounts: []uint16{0}, Data: []byte{}},
			{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{}}, // duplicate
		},
	}

	ids := msg.ProgramIDs()
	require.Equal(t, 2, len(ids))
	assert.Equal(t, prog0, ids[0])
	assert.Equal(t, prog1, ids[1])
}

// Ported from legacy.rs: test_is_instruction_account.
func TestIsInstructionAccount(t *testing.T) {
	key0 := newUniqueKey()
	key1 := newUniqueKey()
	loader := newUniqueKey()

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: PublicKeySlice{key0, key1, loader},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 2, Accounts: []uint16{0, 1}, Data: []byte{}},
		},
	}

	assert.True(t, msg.IsInstructionAccount(0))
	assert.True(t, msg.IsInstructionAccount(1))
	assert.False(t, msg.IsInstructionAccount(2)) // program, not instruction account
}

// Ported from legacy.rs: test_program_position.
func TestProgramPosition(t *testing.T) {
	id := newUniqueKey()
	prog0 := newUniqueKey()
	prog1 := newUniqueKey()

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 2,
		},
		AccountKeys: PublicKeySlice{id, prog0, prog1},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{}},
			{ProgramIDIndex: 2, Accounts: []uint16{0}, Data: []byte{}},
		},
	}

	_, found := msg.ProgramPosition(0)
	assert.False(t, found, "id is not a program")

	pos, found := msg.ProgramPosition(1)
	assert.True(t, found)
	assert.Equal(t, 0, pos, "first program")

	pos, found = msg.ProgramPosition(2)
	assert.True(t, found)
	assert.Equal(t, 1, pos, "second program")
}

// Ported from sanitized.rs: test_num_readonly_accounts.
func TestNumReadonlyAccounts_Legacy(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: PublicKeySlice{
			newUniqueKey(), newUniqueKey(),
			newUniqueKey(), newUniqueKey(),
		},
	}

	assert.Equal(t, 2, msg.NumReadonlyAccounts())
}

// Ported from sanitized.rs: test_num_readonly_accounts (V0 variant).
// V0 messages also have readonly accounts from lookups.
func TestNumReadonlyAccounts_V0(t *testing.T) {
	tableKey := newUniqueKey()
	msg := Message{
		version: MessageVersionV0,
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: PublicKeySlice{
			newUniqueKey(), newUniqueKey(),
			newUniqueKey(), newUniqueKey(),
		},
		AddressTableLookups: MessageAddressTableLookupSlice{
			{
				AccountKey:      tableKey,
				WritableIndexes: []uint8{0},
				ReadonlyIndexes: []uint8{1},
			},
		},
	}

	// Static readonly: 1 signed + 1 unsigned = 2
	assert.Equal(t, 2, msg.NumReadonlyAccounts())
	// Total with lookups: 2 static readonly + 1 readonly lookup = 3
	numReadonlyLookups := msg.NumLookups() - msg.NumWritableLookups()
	assert.Equal(t, 3, msg.NumReadonlyAccounts()+numReadonlyLookups)
}

// Ported from sanitized.rs: test_get_ix_signers.
func TestGetIxSigners(t *testing.T) {
	signer0 := newUniqueKey()
	signer1 := newUniqueKey()
	nonSigner := newUniqueKey()
	loaderKey := newUniqueKey()

	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: PublicKeySlice{signer0, signer1, nonSigner, loaderKey},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 3, Accounts: []uint16{2, 0}, Data: []byte{}}, // ix 0: nonSigner, signer0
			{ProgramIDIndex: 3, Accounts: []uint16{0, 1}, Data: []byte{}}, // ix 1: signer0, signer1
			{ProgramIDIndex: 3, Accounts: []uint16{0, 0}, Data: []byte{}}, // ix 2: signer0, signer0 (dup)
		},
	}

	// ix 0: only signer0 is a signer (nonSigner at index 2 is not)
	signers0 := msg.GetIxSigners(0)
	require.Equal(t, 1, len(signers0))
	assert.Equal(t, signer0, signers0[0])

	// ix 1: both signer0 and signer1
	signers1 := msg.GetIxSigners(1)
	require.Equal(t, 2, len(signers1))
	assert.Contains(t, []PublicKey(signers1), signer0)
	assert.Contains(t, []PublicKey(signers1), signer1)

	// ix 2: signer0 referenced twice, but deduped
	signers2 := msg.GetIxSigners(2)
	require.Equal(t, 1, len(signers2))
	assert.Equal(t, signer0, signers2[0])

	// Out of range returns nil.
	assert.Nil(t, msg.GetIxSigners(3))
}

// Ported from sanitized.rs: test_static_account_keys.
func TestStaticAccountKeys(t *testing.T) {
	keys := PublicKeySlice{newUniqueKey(), newUniqueKey(), newUniqueKey()}

	t.Run("legacy", func(t *testing.T) {
		msg := Message{
			Header: MessageHeader{
				NumRequiredSignatures:       2,
				NumReadonlySignedAccounts:   1,
				NumReadonlyUnsignedAccounts: 1,
			},
			AccountKeys: keys,
		}
		assert.Equal(t, keys, msg.getStaticKeys())
	})

	t.Run("v0 no lookups", func(t *testing.T) {
		msg := Message{
			version: MessageVersionV0,
			Header: MessageHeader{
				NumRequiredSignatures:       2,
				NumReadonlySignedAccounts:   1,
				NumReadonlyUnsignedAccounts: 1,
			},
			AccountKeys: keys,
		}
		assert.Equal(t, keys, msg.getStaticKeys())
	})

	t.Run("v0 with lookups resolved", func(t *testing.T) {
		tableKey := newUniqueKey()
		extraWritable := newUniqueKey()
		extraReadonly := newUniqueKey()

		msg := Message{
			version: MessageVersionV0,
			Header: MessageHeader{
				NumRequiredSignatures:       2,
				NumReadonlySignedAccounts:   1,
				NumReadonlyUnsignedAccounts: 1,
			},
			AccountKeys: append(PublicKeySlice{}, keys...),
			AddressTableLookups: MessageAddressTableLookupSlice{
				{
					AccountKey:      tableKey,
					WritableIndexes: []uint8{0},
					ReadonlyIndexes: []uint8{1},
				},
			},
			addressTables: map[PublicKey]PublicKeySlice{
				tableKey: {extraWritable, extraReadonly},
			},
		}
		require.NoError(t, msg.ResolveLookups())

		// After resolution, AccountKeys has 5 entries, but static keys are only the first 3.
		assert.Equal(t, 5, len(msg.AccountKeys))
		staticKeys := msg.getStaticKeys()
		assert.Equal(t, keys, staticKeys)
	})
}

// Ported from v0/mod.rs: test_sanitize_with_max_table_loaded_keys.
// Tests the exact boundary: 256 total accounts (static + lookup) is valid.
func TestMessageSanitize_V0_MaxTableLoadedKeys(t *testing.T) {
	keys := make(PublicKeySlice, 2) // payer + program
	keys[0] = newUniqueKey()
	keys[1] = newUniqueKey()

	// 254 lookup accounts (to reach 256 total)
	writableIndexes := make([]uint8, 127)
	readonlyIndexes := make([]uint8, 127)
	for i := range writableIndexes {
		writableIndexes[i] = uint8(i)
	}
	for i := range readonlyIndexes {
		readonlyIndexes[i] = uint8(i + 127)
	}

	msg := Message{
		version: MessageVersionV0,
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: keys,
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{}},
		},
		AddressTableLookups: MessageAddressTableLookupSlice{
			{
				AccountKey:      newUniqueKey(),
				WritableIndexes: writableIndexes,
				ReadonlyIndexes: readonlyIndexes,
			},
		},
	}

	// 2 static + 254 lookup = 256 total → should pass
	err := msg.Sanitize()
	require.NoError(t, err)
}

// Ported from v0/loaded.rs: test_is_writable_index.
// Tests uncheckedAccountIndexIsWritable with a resolved V0 message.
func TestUncheckedAccountIndexIsWritable_WithLookups(t *testing.T) {
	tableKey := newUniqueKey()
	msg := Message{
		version: MessageVersionV0,
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: PublicKeySlice{
			newUniqueKey(), // 0: writable signer
			newUniqueKey(), // 1: readonly signer
			newUniqueKey(), // 2: writable unsigned
			newUniqueKey(), // 3: readonly unsigned
		},
		AddressTableLookups: MessageAddressTableLookupSlice{
			{
				AccountKey:      tableKey,
				WritableIndexes: []uint8{0},    // 4: writable from lookup
				ReadonlyIndexes: []uint8{1, 2}, // 5,6: readonly from lookup
			},
		},
		addressTables: map[PublicKey]PublicKeySlice{
			tableKey: {newUniqueKey(), newUniqueKey(), newUniqueKey()},
		},
	}
	require.NoError(t, msg.ResolveLookups())

	expected := []bool{
		true,  // 0: writable signer
		false, // 1: readonly signer
		true,  // 2: writable unsigned
		false, // 3: readonly unsigned
		true,  // 4: writable lookup
		false, // 5: readonly lookup
		false, // 6: readonly lookup
	}

	for i, want := range expected {
		got := msg.uncheckedAccountIndexIsWritable(i)
		assert.Equal(t, want, got, "index %d", i)
	}
}

// Ported from v0/loaded.rs: test_is_writable.
// Tests IsWritable by public key with a resolved V0 message.
func TestIsWritable_WithResolvedLookups(t *testing.T) {
	staticKeys := PublicKeySlice{
		newUniqueKey(), // 0: writable signer
		newUniqueKey(), // 1: readonly signer
		newUniqueKey(), // 2: writable unsigned
		newUniqueKey(), // 3: readonly unsigned
	}
	tableKey := newUniqueKey()
	lookupKeys := PublicKeySlice{
		newUniqueKey(), // table[0] → writable lookup (idx 4)
		newUniqueKey(), // table[1] → readonly lookup (idx 5)
		newUniqueKey(), // table[2] → readonly lookup (idx 6)
	}

	msg := Message{
		version: MessageVersionV0,
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: append(PublicKeySlice{}, staticKeys...),
		AddressTableLookups: MessageAddressTableLookupSlice{
			{
				AccountKey:      tableKey,
				WritableIndexes: []uint8{0},
				ReadonlyIndexes: []uint8{1, 2},
			},
		},
		addressTables: map[PublicKey]PublicKeySlice{
			tableKey: lookupKeys,
		},
	}
	require.NoError(t, msg.ResolveLookups())

	allKeys := append(staticKeys, lookupKeys...)
	expected := []bool{true, false, true, false, true, false, false}

	for i, key := range allKeys {
		w, err := msg.IsWritable(key)
		require.NoError(t, err)
		assert.Equal(t, expected[i], w, "key at index %d", i)
	}
}

// Ported from v0/mod.rs: test_is_maybe_writable.
// Tests writability including lookup accounts using header-based logic.
func TestIsWritable_V0_HeaderLayout(t *testing.T) {
	staticKeys := PublicKeySlice{
		newUniqueKey(), // 0: writable signer
		newUniqueKey(), // 1: readonly signer
		newUniqueKey(), // 2: writable unsigned
		newUniqueKey(), // 3: readonly unsigned
	}
	tableKey := newUniqueKey()
	lookupWritable := newUniqueKey()
	lookupReadonly := newUniqueKey()

	msg := Message{
		version: MessageVersionV0,
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: append(PublicKeySlice{}, staticKeys...),
		AddressTableLookups: MessageAddressTableLookupSlice{
			{
				AccountKey:      tableKey,
				WritableIndexes: []uint8{0},
				ReadonlyIndexes: []uint8{1},
			},
		},
		addressTables: map[PublicKey]PublicKeySlice{
			tableKey: {lookupWritable, lookupReadonly},
		},
	}
	require.NoError(t, msg.ResolveLookups())

	// Static accounts:
	w, err := msg.IsWritable(staticKeys[0])
	require.NoError(t, err)
	assert.True(t, w, "idx 0: writable signer")

	w, err = msg.IsWritable(staticKeys[1])
	require.NoError(t, err)
	assert.False(t, w, "idx 1: readonly signer")

	w, err = msg.IsWritable(staticKeys[2])
	require.NoError(t, err)
	assert.True(t, w, "idx 2: writable unsigned")

	w, err = msg.IsWritable(staticKeys[3])
	require.NoError(t, err)
	assert.False(t, w, "idx 3: readonly unsigned")

	// Lookup accounts:
	w, err = msg.IsWritable(lookupWritable)
	require.NoError(t, err)
	assert.True(t, w, "idx 4: writable lookup")

	w, err = msg.IsWritable(lookupReadonly)
	require.NoError(t, err)
	assert.False(t, w, "idx 5: readonly lookup")
}

// Ported from v0/loaded.rs: test_has_duplicates (already partially covered,
// but this adds explicit sub-cases from the Rust test).
func TestHasDuplicates_Loaded(t *testing.T) {
	tableKey := newUniqueKey()
	key0 := newUniqueKey()
	key1 := newUniqueKey()
	lookupW := newUniqueKey()
	lookupR := newUniqueKey()

	t.Run("no duplicates", func(t *testing.T) {
		msg := Message{
			version: MessageVersionV0,
			Header: MessageHeader{
				NumRequiredSignatures:       1,
				NumReadonlySignedAccounts:   0,
				NumReadonlyUnsignedAccounts: 0,
			},
			AccountKeys: PublicKeySlice{key0, key1},
			AddressTableLookups: MessageAddressTableLookupSlice{
				{
					AccountKey:      tableKey,
					WritableIndexes: []uint8{0},
					ReadonlyIndexes: []uint8{1},
				},
			},
			addressTables: map[PublicKey]PublicKeySlice{
				tableKey: {lookupW, lookupR},
			},
		}
		require.NoError(t, msg.ResolveLookups())

		allKeys, err := msg.GetAllKeys()
		require.NoError(t, err)
		assert.False(t, hasDuplicates(allKeys))
	})

	t.Run("with duplicate keys", func(t *testing.T) {
		// Static key duplicated in lookup → should detect duplicate.
		msg := Message{
			version: MessageVersionV0,
			Header: MessageHeader{
				NumRequiredSignatures:       1,
				NumReadonlySignedAccounts:   0,
				NumReadonlyUnsignedAccounts: 0,
			},
			AccountKeys: PublicKeySlice{key0, key1},
			AddressTableLookups: MessageAddressTableLookupSlice{
				{
					AccountKey:      tableKey,
					WritableIndexes: []uint8{0},
					ReadonlyIndexes: []uint8{},
				},
			},
			addressTables: map[PublicKey]PublicKeySlice{
				tableKey: {key0}, // key0 is both static and lookup
			},
		}
		require.NoError(t, msg.ResolveLookups())

		allKeys, err := msg.GetAllKeys()
		require.NoError(t, err)
		assert.True(t, hasDuplicates(allKeys))
	})
}

// hasDuplicates is a test helper matching Rust's has_duplicates check.
func hasDuplicates(keys PublicKeySlice) bool {
	seen := make(map[PublicKey]struct{}, len(keys))
	for _, k := range keys {
		if _, ok := seen[k]; ok {
			return true
		}
		seen[k] = struct{}{}
	}
	return false
}
