package solana

import (
	"testing"
)

// helpers to build realistic messages for benchmarking.

func buildLegacyMessage() Message {
	keys := make(PublicKeySlice, 10)
	for i := range keys {
		keys[i] = newUniqueKey()
	}
	return Message{
		version: MessageVersionLegacy,
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 3,
		},
		AccountKeys:     keys,
		RecentBlockhash: Hash{1, 2, 3},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 9, Accounts: []uint16{0, 1, 2}, Data: []byte{0xAA}},
			{ProgramIDIndex: 8, Accounts: []uint16{3, 4, 5}, Data: []byte{0xBB}},
		},
	}
}

func buildV0Message() Message {
	staticKeys := make(PublicKeySlice, 6)
	for i := range staticKeys {
		staticKeys[i] = newUniqueKey()
	}

	// Build lookup tables with writable and readonly entries.
	tableKey1 := newUniqueKey()
	tableKey2 := newUniqueKey()

	tableAccounts1 := make(PublicKeySlice, 10)
	for i := range tableAccounts1 {
		tableAccounts1[i] = newUniqueKey()
	}
	tableAccounts2 := make(PublicKeySlice, 8)
	for i := range tableAccounts2 {
		tableAccounts2[i] = newUniqueKey()
	}

	msg := Message{
		version: MessageVersionV0,
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   1,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys:     staticKeys,
		RecentBlockhash: Hash{1, 2, 3},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 5, Accounts: []uint16{0, 1, 6, 7}, Data: []byte{0xCC}},
			{ProgramIDIndex: 4, Accounts: []uint16{2, 3, 8, 9}, Data: []byte{0xDD}},
		},
		AddressTableLookups: MessageAddressTableLookupSlice{
			{
				AccountKey:      tableKey1,
				WritableIndexes: []uint8{0, 1, 2},
				ReadonlyIndexes: []uint8{3, 4},
			},
			{
				AccountKey:      tableKey2,
				WritableIndexes: []uint8{0, 1},
				ReadonlyIndexes: []uint8{2, 3, 4},
			},
		},
		addressTables: map[PublicKey]PublicKeySlice{
			tableKey1: tableAccounts1,
			tableKey2: tableAccounts2,
		},
	}

	return msg
}

func buildV0MessageResolved() Message {
	msg := buildV0Message()
	_ = msg.ResolveLookups()
	return msg
}

// --- Benchmarks ---

func BenchmarkMessage_AccountMetaList_Legacy(b *testing.B) {
	msg := buildLegacyMessage()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.AccountMetaList()
	}
}

func BenchmarkMessage_AccountMetaList_V0(b *testing.B) {
	msg := buildV0Message()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.AccountMetaList()
	}
}

func BenchmarkMessage_AccountMetaList_V0Resolved(b *testing.B) {
	msg := buildV0MessageResolved()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.AccountMetaList()
	}
}

func BenchmarkMessage_IsWritable_Legacy(b *testing.B) {
	msg := buildLegacyMessage()
	account := msg.AccountKeys[3] // an unsigned writable account
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.IsWritable(account)
	}
}

func BenchmarkMessage_IsWritable_V0(b *testing.B) {
	msg := buildV0Message()
	account := msg.AccountKeys[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.IsWritable(account)
	}
}

func BenchmarkMessage_IsSigner(b *testing.B) {
	msg := buildLegacyMessage()
	account := msg.AccountKeys[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.IsSigner(account)
	}
}

func BenchmarkMessage_Signers(b *testing.B) {
	msg := buildLegacyMessage()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.Signers()
	}
}

func BenchmarkMessage_Writable_Legacy(b *testing.B) {
	msg := buildLegacyMessage()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.Writable()
	}
}

func BenchmarkMessage_Writable_V0(b *testing.B) {
	msg := buildV0Message()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.Writable()
	}
}

func BenchmarkMessage_GetAllKeys_V0(b *testing.B) {
	msg := buildV0Message()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.GetAllKeys()
	}
}

func BenchmarkMessage_NumLookups(b *testing.B) {
	msg := buildV0Message()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.NumLookups()
	}
}

func BenchmarkMessage_NumWritableLookups(b *testing.B) {
	msg := buildV0Message()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.NumWritableLookups()
	}
}

func BenchmarkMessage_MarshalBinary_Legacy(b *testing.B) {
	msg := buildLegacyMessage()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.MarshalBinary()
	}
}

func BenchmarkMessage_MarshalBinary_V0(b *testing.B) {
	msg := buildV0Message()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.MarshalBinary()
	}
}

func BenchmarkMessage_numStaticAccounts_V0Resolved(b *testing.B) {
	msg := buildV0MessageResolved()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.numStaticAccounts()
	}
}

func BenchmarkMessage_uncheckedAccountIndexIsWritable_V0Resolved(b *testing.B) {
	msg := buildV0MessageResolved()
	numStatic := msg.numStaticAccounts()
	idx := numStatic + 1 // a lookup index
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.uncheckedAccountIndexIsWritable(idx)
	}
}
