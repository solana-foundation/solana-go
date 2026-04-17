package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Message Sanitize Tests ---
// Ported from solana-sdk/message/src/legacy.rs and versions/v0/mod.rs

func TestMessageSanitize_LegacyValid(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 1,
		},
		AccountKeys: PublicKeySlice{newUniqueKey(), newUniqueKey()},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{}},
		},
	}
	msg.version = MessageVersionLegacy
	require.NoError(t, msg.Sanitize())
}

// Ported from legacy.rs: test_sanitize_txs (signing area + readonly overlap).
func TestMessageSanitize_Legacy_HeaderOverflow(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       2,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 2,
		},
		// Only 3 keys, but header needs 2+2=4.
		AccountKeys: PublicKeySlice{newUniqueKey(), newUniqueKey(), newUniqueKey()},
	}
	msg.version = MessageVersionLegacy
	require.Error(t, msg.Sanitize())
}

// Ported from v0/mod.rs: test_sanitize_without_writable_signer.
func TestMessageSanitize_NoWritableSigner(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   1, // all signers are readonly
			NumReadonlyUnsignedAccounts: 0,
		},
		AccountKeys: PublicKeySlice{newUniqueKey()},
	}
	msg.version = MessageVersionLegacy
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no writable signer")
}

// Ported from v0/mod.rs: test_sanitize_without_signer.
func TestMessageSanitize_NoSigner(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures: 0,
		},
		AccountKeys: PublicKeySlice{newUniqueKey()},
	}
	msg.version = MessageVersionLegacy
	err := msg.Sanitize()
	require.Error(t, err)
}

// Ported from legacy.rs: program_id_index out of bounds.
func TestMessageSanitize_Legacy_InvalidProgramID(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 0,
		},
		AccountKeys: PublicKeySlice{newUniqueKey()},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 5, Accounts: []uint16{0}, Data: []byte{}}, // out of bounds
		},
	}
	msg.version = MessageVersionLegacy
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "program_id_index")
}

// Ported from v0/mod.rs: test_sanitize_with_instruction — program at index 0 (payer) is invalid.
func TestMessageSanitize_ProgramIsPayer(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 0,
		},
		AccountKeys: PublicKeySlice{newUniqueKey(), newUniqueKey()},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 0, Accounts: []uint16{1}, Data: []byte{}},
		},
	}
	msg.version = MessageVersionLegacy
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fee payer")
}

// Ported from legacy.rs: account index out of bounds.
func TestMessageSanitize_Legacy_InvalidAccountIndex(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 0,
		},
		AccountKeys: PublicKeySlice{newUniqueKey(), newUniqueKey()},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 1, Accounts: []uint16{5}, Data: []byte{}}, // out of bounds
		},
	}
	msg.version = MessageVersionLegacy
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account index")
}

// --- V0 Sanitize Tests ---

// Ported from v0/mod.rs: test_sanitize.
func TestMessageSanitize_V0_Valid(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 0,
		},
		AccountKeys: PublicKeySlice{newUniqueKey()},
	}
	msg.version = MessageVersionV0
	require.NoError(t, msg.Sanitize())
}

// Ported from v0/mod.rs: test_sanitize_with_table_lookup.
func TestMessageSanitize_V0_WithTableLookup(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures: 1,
		},
		AccountKeys: PublicKeySlice{newUniqueKey()},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      newUniqueKey(),
				WritableIndexes: []uint8{1, 2, 3},
				ReadonlyIndexes: []uint8{0},
			},
		},
	}
	msg.version = MessageVersionV0
	require.NoError(t, msg.Sanitize())
}

// Ported from v0/mod.rs: test_sanitize_with_empty_table_lookup.
func TestMessageSanitize_V0_EmptyTableLookup(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures: 1,
		},
		AccountKeys: PublicKeySlice{newUniqueKey()},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      newUniqueKey(),
				WritableIndexes: []uint8{},
				ReadonlyIndexes: []uint8{},
			},
		},
	}
	msg.version = MessageVersionV0
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loads no accounts")
}

// Ported from v0/mod.rs: test_sanitize_with_max_account_keys (256 = ok).
func TestMessageSanitize_V0_MaxAccountKeys(t *testing.T) {
	keys := make(PublicKeySlice, 256)
	for i := range keys {
		keys[i] = newUniqueKey()
	}
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures: 1,
		},
		AccountKeys: keys,
	}
	msg.version = MessageVersionV0
	require.NoError(t, msg.Sanitize())
}

// Ported from v0/mod.rs: test_sanitize_with_too_many_account_keys (257 = error).
func TestMessageSanitize_V0_TooManyAccountKeys(t *testing.T) {
	keys := make(PublicKeySlice, 257)
	for i := range keys {
		keys[i] = newUniqueKey()
	}
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures: 1,
		},
		AccountKeys: keys,
	}
	msg.version = MessageVersionV0
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum")
}

// Ported from v0/mod.rs: test_sanitize_with_table_lookup_and_ix_with_dynamic_program_id.
// Program IDs loaded from lookup tables should be rejected.
func TestMessageSanitize_V0_DynamicProgramID(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures: 1,
		},
		AccountKeys: PublicKeySlice{newUniqueKey()},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      newUniqueKey(),
				WritableIndexes: []uint8{1, 2, 3},
				ReadonlyIndexes: []uint8{0},
			},
		},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 4, Accounts: []uint16{0, 1, 2, 3}, Data: []byte{}}, // index 4 is in lookup table
		},
	}
	msg.version = MessageVersionV0
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "static keys")
}

// Ported from v0/mod.rs: test_sanitize_with_table_lookup_and_ix_with_static_program_id.
func TestMessageSanitize_V0_StaticProgramID(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures: 1,
		},
		AccountKeys: PublicKeySlice{newUniqueKey(), newUniqueKey()},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      newUniqueKey(),
				WritableIndexes: []uint8{1, 2, 3},
				ReadonlyIndexes: []uint8{0},
			},
		},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 1, Accounts: []uint16{2, 3, 4, 5}, Data: []byte{}},
		},
	}
	msg.version = MessageVersionV0
	require.NoError(t, msg.Sanitize())
}

// Ported from v0/mod.rs: test_sanitize_with_invalid_ix_account.
func TestMessageSanitize_V0_InvalidIxAccount(t *testing.T) {
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures: 1,
		},
		AccountKeys: PublicKeySlice{newUniqueKey(), newUniqueKey()},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      newUniqueKey(),
				WritableIndexes: []uint8{},
				ReadonlyIndexes: []uint8{0},
			},
		},
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 1, Accounts: []uint16{3}, Data: []byte{}}, // index 3 out of bounds (2 static + 1 lookup = 3 total, max index = 2)
		},
	}
	msg.version = MessageVersionV0
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account index")
}

// Ported from v0/mod.rs: test_sanitize_with_too_many_table_loaded_keys.
func TestMessageSanitize_V0_TooManyDynamicKeys(t *testing.T) {
	writable := make([]uint8, 128)
	readonly := make([]uint8, 128)
	for i := range writable {
		writable[i] = uint8(i * 2)
		readonly[i] = uint8(i*2 + 1)
	}
	msg := Message{
		Header: MessageHeader{
			NumRequiredSignatures: 1,
		},
		AccountKeys: PublicKeySlice{newUniqueKey()},
		AddressTableLookups: []MessageAddressTableLookup{
			{
				AccountKey:      newUniqueKey(),
				WritableIndexes: writable,
				ReadonlyIndexes: readonly,
			},
		},
	}
	msg.version = MessageVersionV0
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum")
}

// --- HasDuplicates Tests ---

func TestMessageHasDuplicates(t *testing.T) {
	key := newUniqueKey()

	t.Run("no duplicates", func(t *testing.T) {
		msg := Message{AccountKeys: PublicKeySlice{newUniqueKey(), newUniqueKey(), newUniqueKey()}}
		assert.False(t, msg.HasDuplicates())
	})

	t.Run("with duplicates", func(t *testing.T) {
		msg := Message{AccountKeys: PublicKeySlice{key, newUniqueKey(), key}}
		assert.True(t, msg.HasDuplicates())
	})

	t.Run("adjacent duplicates", func(t *testing.T) {
		msg := Message{AccountKeys: PublicKeySlice{key, key}}
		assert.True(t, msg.HasDuplicates())
	})

	t.Run("single key", func(t *testing.T) {
		msg := Message{AccountKeys: PublicKeySlice{key}}
		assert.False(t, msg.HasDuplicates())
	})

	t.Run("empty", func(t *testing.T) {
		msg := Message{AccountKeys: PublicKeySlice{}}
		assert.False(t, msg.HasDuplicates())
	})
}

// --- Transaction Sanitize Tests ---
// Ported from solana-sdk/transaction/src/lib.rs and versioned/mod.rs

// Ported from versioned/mod.rs: test_sanitize_signatures_inner — exact match.
func TestTransactionSanitize_Valid(t *testing.T) {
	signer := NewWallet().PrivateKey

	tx, err := NewTransaction([]Instruction{
		&testTransactionInstructions{
			accounts:  []*AccountMeta{{PublicKey: signer.PublicKey(), IsSigner: true, IsWritable: true}},
			data:      []byte{0x01},
			programID: SystemProgramID,
		},
	}, Hash{1, 2, 3})
	require.NoError(t, err)

	_, err = tx.Sign(func(key PublicKey) *PrivateKey {
		if key.Equals(signer.PublicKey()) {
			return &signer
		}
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, tx.Sanitize())
}

// Ported from lib.rs: test_sanitize_txs — not enough signatures.
func TestTransactionSanitize_NotEnoughSignatures(t *testing.T) {
	tx := &Transaction{
		Message: Message{
			Header: MessageHeader{
				NumRequiredSignatures:       2,
				NumReadonlySignedAccounts:   0,
				NumReadonlyUnsignedAccounts: 1,
			},
			AccountKeys: PublicKeySlice{newUniqueKey(), newUniqueKey(), newUniqueKey()},
		},
		Signatures: []Signature{{}}, // only 1, need 2
	}
	err := tx.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not enough signatures")
}

// Ported from versioned/mod.rs: test_sanitize_signatures_inner — too many signatures.
func TestTransactionSanitize_TooManySignatures(t *testing.T) {
	tx := &Transaction{
		Message: Message{
			Header: MessageHeader{
				NumRequiredSignatures:       1,
				NumReadonlySignedAccounts:   0,
				NumReadonlyUnsignedAccounts: 0,
			},
			AccountKeys: PublicKeySlice{newUniqueKey(), newUniqueKey()},
		},
		Signatures: []Signature{{}, {}}, // 2, but only 1 required
	}
	err := tx.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many signatures")
}

// Ported from versioned/mod.rs: signatures exceed static account keys.
func TestTransactionSanitize_SignaturesExceedKeys(t *testing.T) {
	tx := &Transaction{
		Message: Message{
			Header: MessageHeader{
				NumRequiredSignatures:       3,
				NumReadonlySignedAccounts:   0,
				NumReadonlyUnsignedAccounts: 0,
			},
			AccountKeys: PublicKeySlice{newUniqueKey(), newUniqueKey()}, // only 2 keys
		},
		Signatures: []Signature{{}, {}, {}}, // 3 signatures
	}
	err := tx.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "static account keys")
}

// --- VerifyWithResults Tests ---

func TestVerifyWithResults_AllValid(t *testing.T) {
	signers := []PrivateKey{
		NewWallet().PrivateKey,
		NewWallet().PrivateKey,
	}

	tx, err := NewTransaction([]Instruction{
		&testTransactionInstructions{
			accounts: []*AccountMeta{
				{PublicKey: signers[0].PublicKey(), IsSigner: true, IsWritable: true},
				{PublicKey: signers[1].PublicKey(), IsSigner: true, IsWritable: false},
			},
			data:      []byte{0x01},
			programID: SystemProgramID,
		},
	}, Hash{42})
	require.NoError(t, err)

	_, err = tx.Sign(func(key PublicKey) *PrivateKey {
		for _, s := range signers {
			if key.Equals(s.PublicKey()) {
				return &s
			}
		}
		return nil
	})
	require.NoError(t, err)

	results, err := tx.VerifyWithResults()
	require.NoError(t, err)
	require.Equal(t, 2, len(results))
	assert.True(t, results[0], "first signature should be valid")
	assert.True(t, results[1], "second signature should be valid")
}

func TestVerifyWithResults_OneBad(t *testing.T) {
	signers := []PrivateKey{
		NewWallet().PrivateKey,
		NewWallet().PrivateKey,
	}

	tx, err := NewTransaction([]Instruction{
		&testTransactionInstructions{
			accounts: []*AccountMeta{
				{PublicKey: signers[0].PublicKey(), IsSigner: true, IsWritable: true},
				{PublicKey: signers[1].PublicKey(), IsSigner: true, IsWritable: false},
			},
			data:      []byte{0x01},
			programID: SystemProgramID,
		},
	}, Hash{42})
	require.NoError(t, err)

	_, err = tx.Sign(func(key PublicKey) *PrivateKey {
		for _, s := range signers {
			if key.Equals(s.PublicKey()) {
				return &s
			}
		}
		return nil
	})
	require.NoError(t, err)

	// Corrupt the second signature.
	tx.Signatures[1][0] ^= 0xFF

	results, err := tx.VerifyWithResults()
	require.NoError(t, err)
	assert.True(t, results[0], "first signature should still be valid")
	assert.False(t, results[1], "second signature should be invalid")
}

// --- UsesDurableNonce Tests ---
// Ported from solana-sdk/transaction: tx_uses_nonce_* tests.

func makeNonceAdvanceIxData() []byte {
	// System instruction discriminant for AdvanceNonceAccount = 4 (LE u32).
	return []byte{4, 0, 0, 0}
}

// Ported from lib.rs: tx_uses_nonce_ok.
func TestUsesDurableNonce_Valid(t *testing.T) {
	nonceAccount := newUniqueKey()
	nonceAuthority := newUniqueKey()
	recentBlockhashes := newUniqueKey()

	tx := &Transaction{
		Message: Message{
			Header: MessageHeader{
				NumRequiredSignatures:       1,
				NumReadonlySignedAccounts:   0,
				NumReadonlyUnsignedAccounts: 2,
			},
			AccountKeys: PublicKeySlice{
				nonceAuthority,    // 0: signer
				nonceAccount,      // 1: writable
				recentBlockhashes, // 2: readonly
				SystemProgramID,   // 3: system program
			},
			Instructions: []CompiledInstruction{
				{
					ProgramIDIndex: 3,                 // system program
					Accounts:       []uint16{1, 2, 0}, // nonce account, recent blockhashes, authority
					Data:           makeNonceAdvanceIxData(),
				},
			},
		},
		Signatures: []Signature{{}},
	}

	assert.True(t, tx.UsesDurableNonce())

	account, ok := tx.GetNonceAccount()
	require.True(t, ok)
	assert.Equal(t, nonceAccount, account)
}

// Ported from lib.rs: tx_uses_nonce_empty_ix_fail.
func TestUsesDurableNonce_EmptyInstructions(t *testing.T) {
	tx := &Transaction{
		Message: Message{
			Header: MessageHeader{NumRequiredSignatures: 1},
			AccountKeys: PublicKeySlice{
				newUniqueKey(),
				SystemProgramID,
			},
			Instructions: []CompiledInstruction{},
		},
		Signatures: []Signature{{}},
	}
	assert.False(t, tx.UsesDurableNonce())
}

// Ported from lib.rs: tx_uses_nonce_bad_prog_id_idx_fail.
func TestUsesDurableNonce_BadProgramIDIndex(t *testing.T) {
	tx := &Transaction{
		Message: Message{
			Header:      MessageHeader{NumRequiredSignatures: 1},
			AccountKeys: PublicKeySlice{newUniqueKey()},
			Instructions: []CompiledInstruction{
				{
					ProgramIDIndex: 255, // out of bounds
					Accounts:       []uint16{0},
					Data:           makeNonceAdvanceIxData(),
				},
			},
		},
		Signatures: []Signature{{}},
	}
	assert.False(t, tx.UsesDurableNonce())
}

// Ported from lib.rs: tx_uses_nonce_first_prog_id_not_nonce_fail.
func TestUsesDurableNonce_WrongProgram(t *testing.T) {
	tx := &Transaction{
		Message: Message{
			Header: MessageHeader{
				NumRequiredSignatures:       1,
				NumReadonlyUnsignedAccounts: 1,
			},
			AccountKeys: PublicKeySlice{
				newUniqueKey(),
				newUniqueKey(), // not system program
			},
			Instructions: []CompiledInstruction{
				{
					ProgramIDIndex: 1,
					Accounts:       []uint16{0},
					Data:           makeNonceAdvanceIxData(),
				},
			},
		},
		Signatures: []Signature{{}},
	}
	assert.False(t, tx.UsesDurableNonce())
}

// Ported from lib.rs: tx_uses_nonce_wrong_first_nonce_ix_fail.
func TestUsesDurableNonce_WrongInstruction(t *testing.T) {
	tx := &Transaction{
		Message: Message{
			Header: MessageHeader{
				NumRequiredSignatures:       1,
				NumReadonlyUnsignedAccounts: 1,
			},
			AccountKeys: PublicKeySlice{
				newUniqueKey(),
				SystemProgramID,
			},
			Instructions: []CompiledInstruction{
				{
					ProgramIDIndex: 1,
					Accounts:       []uint16{0},
					Data:           []byte{2, 0, 0, 0}, // Transfer, not AdvanceNonce
				},
			},
		},
		Signatures: []Signature{{}},
	}
	assert.False(t, tx.UsesDurableNonce())
}

// Tests GetNonceAccount returns false for non-nonce transactions.
func TestGetNonceAccount_NotNonce(t *testing.T) {
	tx := &Transaction{
		Message: Message{
			Header:      MessageHeader{NumRequiredSignatures: 1},
			AccountKeys: PublicKeySlice{newUniqueKey()},
			Instructions: []CompiledInstruction{
				{ProgramIDIndex: 0, Accounts: []uint16{0}, Data: []byte{0x01}},
			},
		},
		Signatures: []Signature{{}},
	}
	_, ok := tx.GetNonceAccount()
	assert.False(t, ok)
}

// Tests that nonce instruction with short data is not detected.
func TestUsesDurableNonce_ShortData(t *testing.T) {
	tx := &Transaction{
		Message: Message{
			Header: MessageHeader{
				NumRequiredSignatures:       1,
				NumReadonlyUnsignedAccounts: 1,
			},
			AccountKeys: PublicKeySlice{
				newUniqueKey(),
				SystemProgramID,
			},
			Instructions: []CompiledInstruction{
				{
					ProgramIDIndex: 1,
					Accounts:       []uint16{0},
					Data:           []byte{4, 0}, // too short — need 4 bytes for u32
				},
			},
		},
		Signatures: []Signature{{}},
	}
	assert.False(t, tx.UsesDurableNonce())
}
