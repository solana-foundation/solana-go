package solana

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ported from solana-sdk transaction/src/versioned/mod.rs and sanitized.rs (v1 parts).

// Golden v1 transactions from the Rust SDK; signer seeds [1;32] and [2;32].
const (
	// (a) empty config, 1 signer, 1 instruction
	v1GoldenTxEmptyConfig = "8101000200000000abababababababababababababababababababababababababababababababab01048a88e3dd7409" +
		"f195fd52db2d3cba5d72ca6709bf1d94121bf3748801b40f6f5c02020202020202020202020202020202020202020202" +
		"020202020202020202020303030303030303030303030303030303030303030303030303030303030303101010101010" +
		"101010101010101010101010101010101010101010101010101003030400000102deadbeef4e72ba5d7f993321339fd3" +
		"511742c1c21f5b80ca529337413f766a32485b86ddc124f38ebe5f3a907ed3ede52331e746c7e55fddefae148c5aea7a" +
		"780af4b604"

	// (b) all four config fields, 2 signers (one readonly), 3 instructions (one with empty data)
	v1GoldenTxFullConfig = "810201031f000000bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb03068a88e3dd7409" +
		"f195fd52db2d3cba5d72ca6709bf1d94121bf3748801b40f6f5c8139770ea87d175f56a35466c34c7ecccb8d8a91b4ee" +
		"37a25df60f5b8fc9b3940404040404040404040404040404040404040404040404040404040404040404050505050505" +
		"050505050505050505050505050505050505050505050505050510101010101010101010101010101010101010101010" +
		"101010101010101010102020202020202020202020202020202020202020202020202020202020202020881300000000" +
		"0000400d03000000010000000100040303000501000004002c0100010201020303000102030405060708090a0b0c0d0e" +
		"0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e" +
		"3f404142434445464748494a4b4c4d4e4f505152535455565758595a5b5c5d5e5f606162636465666768696a6b6c6d6e" +
		"6f707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e" +
		"9fa0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c8c9cacbcccdce" +
		"cfd0d1d2d3d4d5d6d7d8d9dadbdcdddedfe0e1e2e3e4e5e6e7e8e9eaebecedeeeff0f1f2f3f4f5f6f7f8f9fafbfcfdfe" +
		"ff000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2bf24e03" +
		"afc6235e9892eefba1956ccc9fb2b88cbd8dc8bd8680ca858655c0400e9059af3a76b3c2aab2f2c52c4169eb2c21db5a" +
		"4f3e31fb9fc056290bc97162079c370ae9e55f2f6a0c1ef654558900295e253bbec40644a6a1c5e49c9b66eb7d3e8b51" +
		"596f5059c82e1cfa680ffcc7572e962b0113dfde0dcd2eaefbdf86c003"

	// (c) priority fee + heap size only, empty instruction data
	v1GoldenTxFeeHeap = "8101000113000000070707070707070707070707070707070707070707070707070707070707070701028a88e3dd7409" +
		"f195fd52db2d3cba5d72ca6709bf1d94121bf3748801b40f6f5c10101010101010101010101010101010101010101010" +
		"10101010101010101010010000000000000000800000010100000070083fb1bf5a9e2c90c6903b7328529cdf59e04e2f" +
		"cf29c718bea8c76fd325f4b67e18feb1f95f9aab171140b81b8c8238c9ffea33d4133bae92af68c6453800"
)

var (
	v1TestPayerKey   = PrivateKey(ed25519.NewKeyFromSeed(bytesRepeat(1, 32)))
	v1TestSigner2Key = PrivateKey(ed25519.NewKeyFromSeed(bytesRepeat(2, 32)))
)

func bytesRepeat(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}

func pubkeyRepeat(b byte) PublicKey {
	var pk PublicKey
	copy(pk[:], bytesRepeat(b, 32))
	return pk
}

func hashRepeat(b byte) Hash {
	var h Hash
	copy(h[:], bytesRepeat(b, 32))
	return h
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	out, err := hex.DecodeString(s)
	require.NoError(t, err)
	return out
}

func v1TestSigner(t *testing.T, keys ...PrivateKey) privateKeyGetter {
	t.Helper()
	return func(pub PublicKey) *PrivateKey {
		for i := range keys {
			if keys[i].PublicKey().Equals(pub) {
				return &keys[i]
			}
		}
		return nil
	}
}

func TestV1_TestKeysMatchRustGenerator(t *testing.T) {
	assert.Equal(t, "AKnL4NNf3DGWZJS6cPknBuEGnVsV4A4m5tgebLHaRSZ9", v1TestPayerKey.PublicKey().String())
	assert.Equal(t, "9hSR6S7WPtxmTojgo6GG3k4yDPecgJY292j7xrsUGWBu", v1TestSigner2Key.PublicKey().String())
}

// The Rust-produced transactions must decode, verify, sanitize and re-encode
// byte-for-byte.
func TestV1_GoldenTransactionsFromRustSDK(t *testing.T) {
	type expect struct {
		name      string
		hex       string
		header    MessageHeader
		numKeys   int
		numIx     int
		config    TransactionConfig
		signers   []PrivateKey
		blockhash Hash
	}
	cases := []expect{
		{
			name:      "a_empty_config",
			hex:       v1GoldenTxEmptyConfig,
			header:    MessageHeader{NumRequiredSignatures: 1, NumReadonlySignedAccounts: 0, NumReadonlyUnsignedAccounts: 2},
			numKeys:   4,
			numIx:     1,
			config:    TransactionConfig{},
			signers:   []PrivateKey{v1TestPayerKey},
			blockhash: hashRepeat(0xAB),
		},
		{
			name:    "b_full_config",
			hex:     v1GoldenTxFullConfig,
			header:  MessageHeader{NumRequiredSignatures: 2, NumReadonlySignedAccounts: 1, NumReadonlyUnsignedAccounts: 3},
			numKeys: 6,
			numIx:   3,
			config: TransactionConfig{}.
				WithPriorityFee(5000).
				WithComputeUnitLimit(200_000).
				WithLoadedAccountsDataSizeLimit(65_536).
				WithHeapSize(65_536),
			signers:   []PrivateKey{v1TestPayerKey, v1TestSigner2Key},
			blockhash: hashRepeat(0xBB),
		},
		{
			name:      "c_fee_heap",
			hex:       v1GoldenTxFeeHeap,
			header:    MessageHeader{NumRequiredSignatures: 1, NumReadonlySignedAccounts: 0, NumReadonlyUnsignedAccounts: 1},
			numKeys:   2,
			numIx:     1,
			config:    TransactionConfig{}.WithPriorityFee(1).WithHeapSize(32_768),
			signers:   []PrivateKey{v1TestPayerKey},
			blockhash: hashRepeat(7),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := mustHex(t, tc.hex)
			require.LessOrEqual(t, len(raw), MaxTransactionSizeV1)

			tx, err := TransactionFromBytes(raw)
			require.NoError(t, err)
			assert.Equal(t, MessageVersionV1, tx.Message.GetVersion())
			assert.Equal(t, tc.header, tx.Message.Header)
			assert.Len(t, tx.Message.AccountKeys, tc.numKeys)
			assert.Len(t, tx.Message.Instructions, tc.numIx)
			assert.Equal(t, tc.config, tx.Message.TransactionConfig)
			assert.Equal(t, tc.blockhash, tx.Message.RecentBlockhash)
			assert.Nil(t, tx.Message.AddressTableLookups)
			require.Len(t, tx.Signatures, int(tc.header.NumRequiredSignatures))
			for i, k := range tc.signers {
				assert.Equal(t, k.PublicKey(), tx.Message.AccountKeys[i], "signer %d", i)
			}

			// Rust signatures verify against the Go message bytes (0x81 included).
			require.NoError(t, tx.VerifySignatures())
			require.NoError(t, tx.Sanitize())
			results, err := tx.VerifyWithResults()
			require.NoError(t, err)
			for _, ok := range results {
				assert.True(t, ok)
			}

			// Byte-exact re-encode.
			again, err := tx.MarshalBinary()
			require.NoError(t, err)
			assert.Equal(t, raw, again)

			// Deterministic ed25519: re-signing reproduces the Rust signatures.
			resigned := *tx
			resigned.Signatures = nil
			_, err = resigned.Sign(v1TestSigner(t, tc.signers...))
			require.NoError(t, err)
			assert.Equal(t, tx.Signatures, resigned.Signatures)

			// Base64 / base58 helpers.
			b64, err := tx.ToBase64()
			require.NoError(t, err)
			fromB64, err := TransactionFromBase64(b64)
			require.NoError(t, err)
			assert.Equal(t, MessageVersionV1, fromB64.Message.GetVersion())
			assert.Equal(t, tx.Signatures, fromB64.Signatures)

			// The tree renderer must not panic on V1.
			assert.NotEmpty(t, tx.String())
		})
	}
}

// NewTransaction must reproduce Rust try_compile_with_config byte-for-byte.
func TestV1_NewTransactionMatchesRustSDK(t *testing.T) {
	payer := v1TestPayerKey.PublicKey()
	s2 := v1TestSigner2Key.PublicKey()
	p1 := pubkeyRepeat(0x10)
	p2 := pubkeyRepeat(0x20)

	t.Run("a_empty_config", func(t *testing.T) {
		ix := NewInstruction(p1, AccountMetaSlice{
			Meta(payer).WRITE().SIGNER(),
			Meta(pubkeyRepeat(0x02)).WRITE(),
			Meta(pubkeyRepeat(0x03)),
		}, []byte{0xDE, 0xAD, 0xBE, 0xEF})
		tx, err := NewTransaction([]Instruction{ix}, hashRepeat(0xAB),
			TransactionPayer(payer),
			TransactionMessageVersion(MessageVersionV1),
		)
		require.NoError(t, err)
		_, err = tx.Sign(v1TestSigner(t, v1TestPayerKey))
		require.NoError(t, err)
		got, err := tx.MarshalBinary()
		require.NoError(t, err)
		assert.Equal(t, mustHex(t, v1GoldenTxEmptyConfig), got)
	})

	t.Run("b_full_config", func(t *testing.T) {
		ix1 := NewInstruction(p1, AccountMetaSlice{
			Meta(payer).WRITE().SIGNER(),
			Meta(s2).SIGNER(),
			Meta(pubkeyRepeat(0x04)).WRITE(),
		}, []byte{1, 2, 3})
		ix2 := NewInstruction(p2, AccountMetaSlice{Meta(pubkeyRepeat(0x05))}, []byte{})
		data3 := make([]byte, 300)
		for i := range data3 {
			data3[i] = byte(i % 256)
		}
		ix3 := NewInstruction(p1, AccountMetaSlice{}, data3)
		cfg := TransactionConfig{}.
			WithPriorityFee(5000).
			WithComputeUnitLimit(200_000).
			WithLoadedAccountsDataSizeLimit(65_536).
			WithHeapSize(65_536)
		tx, err := NewTransactionBuilder().
			AddInstruction(ix1).
			AddInstruction(ix2).
			AddInstruction(ix3).
			SetRecentBlockHash(hashRepeat(0xBB)).
			SetFeePayer(payer).
			SetTransactionConfig(cfg).
			Build()
		require.NoError(t, err)
		assert.Equal(t, MessageVersionV1, tx.Message.GetVersion())
		_, err = tx.Sign(v1TestSigner(t, v1TestPayerKey, v1TestSigner2Key))
		require.NoError(t, err)
		got, err := tx.MarshalBinary()
		require.NoError(t, err)
		assert.Equal(t, mustHex(t, v1GoldenTxFullConfig), got)
	})

	t.Run("c_fee_heap", func(t *testing.T) {
		ix := NewInstruction(p1, AccountMetaSlice{Meta(payer).WRITE().SIGNER()}, []byte{})
		tx, err := NewTransaction([]Instruction{ix}, hashRepeat(7),
			TransactionPayer(payer),
			TransactionV1Config(TransactionConfig{}.WithPriorityFee(1).WithHeapSize(32_768)),
		)
		require.NoError(t, err)
		_, err = tx.Sign(v1TestSigner(t, v1TestPayerKey))
		require.NoError(t, err)
		got, err := tx.MarshalBinary()
		require.NoError(t, err)
		assert.Equal(t, mustHex(t, v1GoldenTxFeeHeap), got)
	})
}

// Port of `v1_transaction_serialization` (test_case 0 = at max size, 1 = over by one).
func TestV1_TransactionSerializationAtMaxSize(t *testing.T) {
	const (
		numSignatures          = 1
		numAddresses           = 2
		numInstructionAccounts = 1
	)
	overhead := 1 + // version byte
		numSignatures*64 +
		fixedHeaderSizeV1 +
		numAddresses*32 +
		instructionHeaderSizeV1 +
		numInstructionAccounts

	for _, delta := range []int{0, 1} {
		maxDataSize := MaxTransactionSizeV1 - overhead + delta
		msg := Message{
			version: MessageVersionV1,
			Header: MessageHeader{
				NumRequiredSignatures:       numSignatures,
				NumReadonlySignedAccounts:   0,
				NumReadonlyUnsignedAccounts: 0,
			},
			AccountKeys:     PublicKeySlice{newUniqueKey(), newUniqueKey()},
			RecentBlockhash: uniqueHash(),
			Instructions: []CompiledInstruction{{
				ProgramIDIndex: 1,
				Accounts:       []uint16{0},
				Data:           make([]byte, maxDataSize),
			}},
		}
		tx := &Transaction{Message: msg, Signatures: []Signature{{}}}

		serialized, err := tx.MarshalBinary()
		require.NoError(t, err)
		if delta == 0 {
			assert.Equal(t, MaxTransactionSizeV1, len(serialized), "transaction should be exactly at max size")
		} else {
			assert.Equal(t, MaxTransactionSizeV1+delta, len(serialized), "transaction should be over by %d byte(s)", delta)
		}

		back, err := TransactionFromBytes(serialized)
		require.NoError(t, err)
		assert.Equal(t, tx.Signatures, back.Signatures)
		assert.Equal(t, tx.Message.Header, back.Message.Header)
		assert.Equal(t, tx.Message.AccountKeys, back.Message.AccountKeys)
		assert.Equal(t, tx.Message.Instructions, back.Message.Instructions)
		assert.Equal(t, MessageVersionV1, back.Message.GetVersion())
	}
}

// Port of `test_v1_message_in_legacy_transaction`: a V1 message inside a
// legacy/v0-shaped transaction envelope must be rejected.
func TestV1_MessageInLegacyTransactionRejected(t *testing.T) {
	malformed := []byte{
		0x00, // 0 signatures via compact-u16 -> takes the legacy/v0 path
		0x81, // V1 message prefix
		// V1 LegacyHeader (3 bytes)
		0x01, 0x00, 0x00,
		// TransactionConfigMask (4 bytes, little-endian)
		0x00, 0x00, 0x00, 0x00,
	}
	malformed = append(malformed, make([]byte, 32)...) // lifetime specifier
	malformed = append(malformed, 0x00)                // NumInstructions
	malformed = append(malformed, 0x01)                // NumAddresses
	malformed = append(malformed, make([]byte, 32)...) // 1 address

	_, err := TransactionFromBytes(malformed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid message version")
}

// A legacy body behind a 0x81 prefix must fail to decode.
func TestV1_LegacyBodyBehindV1PrefixRejected(t *testing.T) {
	legacy := Message{
		Header:          MessageHeader{NumRequiredSignatures: 1},
		AccountKeys:     PublicKeySlice{newUniqueKey(), newUniqueKey()},
		RecentBlockhash: uniqueHash(),
		Instructions:    []CompiledInstruction{{ProgramIDIndex: 1, Accounts: []uint16{0}, Data: []byte{1, 2, 3}}},
	}
	body, err := legacy.MarshalLegacy()
	require.NoError(t, err)
	raw := append([]byte{0x81}, body...)
	_, err = TransactionFromBytes(raw)
	require.Error(t, err)
}

func TestV1_TransactionTruncationNeverPanics(t *testing.T) {
	raw := mustHex(t, v1GoldenTxFullConfig)
	for n := 0; n < len(raw); n++ {
		_, err := TransactionFromBytes(raw[:n])
		require.Error(t, err, "prefix of %d/%d bytes must fail", n, len(raw))
	}
	// One byte fewer than needed for the last signature.
	_, err := TransactionFromBytes(raw[:len(raw)-1])
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signatures")
}

// Any other high-bit first byte is not a valid transaction (upstream parity).
func TestV1_InvalidTransactionDiscriminator(t *testing.T) {
	for _, b := range []byte{0x80, 0x82, 0xFF} {
		_, err := TransactionFromBytes(append([]byte{b}, make([]byte, 200)...))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid transaction discriminator")
	}
}

// SIMD-0385: no trailing data after the signatures. Slice-level decoding is
// strict for v1; the streaming decoder stays lenient (wincode parity), and
// legacy/v0 slices keep their historical lenient behavior.
func TestV1_TrailingBytesRejected(t *testing.T) {
	raw := mustHex(t, v1GoldenTxFeeHeap)
	padded := append(append([]byte{}, raw...), 0x00)

	_, err := TransactionFromBytes(padded)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "trailing bytes")
	_, err = TransactionFromBase64(base64.StdEncoding.EncodeToString(padded))
	require.Error(t, err)

	// Exact bytes still decode.
	_, err = TransactionFromBytes(raw)
	require.NoError(t, err)

	// Streaming decode is lenient and leaves the extra byte unread.
	decoder := bin.NewBinDecoder(padded)
	tx, err := TransactionFromDecoder(decoder)
	require.NoError(t, err)
	assert.Equal(t, MessageVersionV1, tx.Message.GetVersion())
	assert.Equal(t, 1, decoder.Remaining())

	// Legacy transactions with trailing bytes remain accepted.
	legacyTx, err := NewTransaction(
		[]Instruction{NewInstruction(pubkeyRepeat(0x10), AccountMetaSlice{Meta(newUniqueKey()).WRITE().SIGNER()}, []byte{1})},
		uniqueHash(),
	)
	require.NoError(t, err)
	legacyRaw, err := legacyTx.MarshalBinary()
	require.NoError(t, err)
	_, err = TransactionFromBytes(append(legacyRaw, 0x00))
	require.NoError(t, err)
}

// A decoded header cannot require more signatures than there are keys.
func TestV1_DecodeRejectsMoreSignaturesThanKeys(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Header.NumRequiredSignatures = 5 // 3 keys
	msgBytes, err := msg.MarshalV1()
	require.NoError(t, err)
	_, err = TransactionFromBytes(append(msgBytes, make([]byte, 5*64)...))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only 3 account keys")

	// signerKeys is clamped, so signing a hostile message never panics.
	tx := &Transaction{Message: msg}
	_, err = tx.PartialSign(func(PublicKey) *PrivateKey { return nil })
	require.NoError(t, err)
	assert.Len(t, tx.Signatures, 3)
}

// ComputeBudget instructions are no-ops in v1; NewTransaction rejects them.
func TestV1_NewTransactionRejectsComputeBudgetInstructions(t *testing.T) {
	payer := newUniqueKey()
	transfer := NewInstruction(pubkeyRepeat(0x10), AccountMetaSlice{Meta(payer).WRITE().SIGNER()}, []byte{1})
	cb := NewInstruction(ComputeBudget, AccountMetaSlice{}, []byte{2, 0, 0, 0, 0})
	_, err := NewTransaction([]Instruction{cb, transfer}, uniqueHash(),
		TransactionPayer(payer),
		TransactionV1Config(TransactionConfig{}.WithComputeUnitLimit(1)),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ComputeBudget")

	// Fine as legacy.
	_, err = NewTransaction([]Instruction{cb, transfer}, uniqueHash(), TransactionPayer(payer))
	require.NoError(t, err)
}

func TestV1_TransactionRequiresSignaturesToFitHeader(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Header.NumRequiredSignatures = 3
	msg.Header.NumReadonlyUnsignedAccounts = 0
	msgBytes, err := msg.MarshalV1()
	require.NoError(t, err)
	// Only two signatures follow, header requires three.
	raw := append(msgBytes, make([]byte, 2*64)...)
	_, err = TransactionFromBytes(raw)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires 3 signatures")
}

// Port of transaction/src/sanitized.rs `test_verify_v1_message_data_includes_prefix`
// (#715): the signed bytes are `0x81 || message body`.
func TestV1_VerifySignedBytesIncludePrefix(t *testing.T) {
	payer := v1TestPayerKey.PublicKey()
	ix := NewInstruction(pubkeyRepeat(0x10), AccountMetaSlice{Meta(payer).WRITE().SIGNER()}, []byte{1})
	tx, err := NewTransaction([]Instruction{ix}, uniqueHash(),
		TransactionPayer(payer),
		TransactionV1Config(TransactionConfig{}.WithComputeUnitLimit(1000)),
	)
	require.NoError(t, err)
	_, err = tx.Sign(v1TestSigner(t, v1TestPayerKey))
	require.NoError(t, err)
	require.NoError(t, tx.VerifySignatures())

	msgBytes, err := tx.Message.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, byte(0x81), msgBytes[0])
	assert.True(t, tx.Signatures[0].Verify(payer, msgBytes))
	// The signature does NOT verify over the body without the version byte.
	assert.False(t, tx.Signatures[0].Verify(payer, msgBytes[1:]))
}

func TestV1_NewTransactionRejectsAddressTables(t *testing.T) {
	payer := newUniqueKey()
	table := newUniqueKey()
	extra := newUniqueKey()
	ix := NewInstruction(pubkeyRepeat(0x10), AccountMetaSlice{Meta(payer).WRITE().SIGNER(), Meta(extra).WRITE()}, []byte{1})

	_, err := NewTransaction([]Instruction{ix}, uniqueHash(),
		TransactionPayer(payer),
		TransactionAddressTables(map[PublicKey]PublicKeySlice{table: {extra}}),
		TransactionMessageVersion(MessageVersionV1),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "address lookup tables")

	_, err = NewTransaction([]Instruction{ix}, uniqueHash(),
		TransactionPayer(payer),
		TransactionMessageVersion(MessageVersion(42)),
	)
	require.Error(t, err)
}

func TestV1_NewTransactionVersionOption(t *testing.T) {
	payer := newUniqueKey()
	ix := NewInstruction(pubkeyRepeat(0x10), AccountMetaSlice{Meta(payer).WRITE().SIGNER()}, []byte{1})
	bh := uniqueHash()

	// Explicit v0 without address tables is honored.
	tx, err := NewTransaction([]Instruction{ix}, bh, TransactionPayer(payer), TransactionMessageVersion(MessageVersionV0))
	require.NoError(t, err)
	assert.Equal(t, MessageVersionV0, tx.Message.GetVersion())
	raw, err := tx.MarshalBinary()
	require.NoError(t, err)
	assert.Equal(t, byte(0x80), raw[1+64]) // after compact-u16 len + 1 padded signature

	// Default is legacy.
	tx, err = NewTransaction([]Instruction{ix}, bh, TransactionPayer(payer))
	require.NoError(t, err)
	assert.Equal(t, MessageVersionLegacy, tx.Message.GetVersion())

	// Config with a non-V1 version is an error, regardless of option order.
	_, err = NewTransaction([]Instruction{ix}, bh, TransactionPayer(payer),
		TransactionV1Config(TransactionConfig{}.WithComputeUnitLimit(1)),
		TransactionMessageVersion(MessageVersionLegacy),
	)
	require.Error(t, err)
	_, err = NewTransaction([]Instruction{ix}, bh, TransactionPayer(payer),
		TransactionMessageVersion(MessageVersionV0),
		TransactionV1Config(TransactionConfig{}),
	)
	require.NoError(t, err, "empty config after V0: last version option wins")

	// Builder helpers.
	tx, err = NewTransactionBuilder().AddInstruction(ix).SetRecentBlockHash(bh).SetFeePayer(payer).SetVersion(MessageVersionV1).Build()
	require.NoError(t, err)
	assert.Equal(t, MessageVersionV1, tx.Message.GetVersion())
	assert.True(t, tx.Message.TransactionConfig.IsEmpty())
}

// Malformed-but-well-formed-on-the-wire messages decode and are caught by Sanitize.
func TestV1_DecodedMessageFailsSanitize(t *testing.T) {
	msg := createTestV1Message(t)
	base := mustMarshalV1(t, &msg)
	// 0x81 | hdr(3) | mask(4) | lifetime(32) | counts(2) | 3 addrs | cu(4) | ix hdr(4) | payload
	ixHeaderOff := 1 + 3 + 4 + 32 + 1 + 1 + 3*32 + 4
	payloadOff := ixHeaderOff + 4

	// program_id_index out of range
	data := append([]byte(nil), base...)
	data[ixHeaderOff] = 7
	decoded := mustUnmarshalV1(t, data)
	assert.Equal(t, uint16(7), decoded.Instructions[0].ProgramIDIndex)
	require.Error(t, decoded.Sanitize())

	// account index out of range
	data = append([]byte(nil), base...)
	data[payloadOff] = 9
	decoded = mustUnmarshalV1(t, data)
	assert.Equal(t, uint16(9), decoded.Instructions[0].Accounts[0])
	require.Error(t, decoded.Sanitize())

	// duplicate address (copy addr[0] over addr[1])
	data = append([]byte(nil), base...)
	addrOff := 1 + 3 + 4 + 32 + 1 + 1
	copy(data[addrOff+32:addrOff+64], data[addrOff:addrOff+32])
	decoded = mustUnmarshalV1(t, data)
	require.Error(t, decoded.Sanitize())

	// heap size present but out of bounds: set bit 4 and append the value
	data = append([]byte(nil), base[:ixHeaderOff]...)
	data[4] |= byte(TransactionConfigMaskHeapSize)
	data = append(data, 0, 4, 0, 0) // heap = 1024 (< 32 KiB)
	data = append(data, base[ixHeaderOff:]...)
	decoded = mustUnmarshalV1(t, data)
	require.NotNil(t, decoded.TransactionConfig.HeapSize)
	assert.Equal(t, uint32(1024), *decoded.TransactionConfig.HeapSize)
	require.Error(t, decoded.Sanitize())
}

func TestV1_NewTransactionSameLayoutAsLegacy(t *testing.T) {
	payer := newUniqueKey()
	a := newUniqueKey()
	b := newUniqueKey()
	prog := newUniqueKey()
	ix := NewInstruction(prog, AccountMetaSlice{Meta(payer).WRITE().SIGNER(), Meta(a).WRITE(), Meta(b)}, []byte{9, 8, 7})
	bh := uniqueHash()

	legacyTx, err := NewTransaction([]Instruction{ix}, bh, TransactionPayer(payer))
	require.NoError(t, err)
	v1Tx, err := NewTransaction([]Instruction{ix}, bh, TransactionPayer(payer), TransactionV1Config(TransactionConfig{}.WithComputeUnitLimit(50_000)))
	require.NoError(t, err)

	assert.Equal(t, MessageVersionLegacy, legacyTx.Message.GetVersion())
	assert.Equal(t, MessageVersionV1, v1Tx.Message.GetVersion())
	assert.Equal(t, legacyTx.Message.Header, v1Tx.Message.Header)
	assert.Equal(t, legacyTx.Message.AccountKeys, v1Tx.Message.AccountKeys)
	assert.Equal(t, legacyTx.Message.Instructions, v1Tx.Message.Instructions)
	assert.Equal(t, uint32(50_000), *v1Tx.Message.TransactionConfig.ComputeUnitLimit)
	require.NoError(t, v1Tx.Message.Sanitize())

	// V1 with too many accounts fails early in NewTransaction.
	metas := AccountMetaSlice{Meta(payer).WRITE().SIGNER()}
	for range 64 {
		metas = append(metas, Meta(newUniqueKey()))
	}
	big := NewInstruction(prog, metas, nil)
	_, err = NewTransaction([]Instruction{big}, bh, TransactionPayer(payer), TransactionMessageVersion(MessageVersionV1))
	require.Error(t, err)
	assert.True(t, IsSanitizeError(err))
	// ...but is fine as legacy.
	_, err = NewTransaction([]Instruction{big}, bh, TransactionPayer(payer))
	require.NoError(t, err)
}

func TestV1_MarshalBinaryPadsMissingAndRejectsExtraSignatures(t *testing.T) {
	msg := createTestV1Message(t)
	tx := &Transaction{Message: msg}

	// No signatures: padded with one zero signature.
	raw, err := tx.MarshalBinary()
	require.NoError(t, err)
	msgBytes := mustMarshalV1(t, &msg)
	assert.Equal(t, len(msgBytes)+64, len(raw))
	assert.Equal(t, make([]byte, 64), raw[len(msgBytes):])

	// Too many signatures: error (no length prefix on the wire).
	tx.Signatures = []Signature{{}, {}}
	_, err = tx.MarshalBinary()
	require.Error(t, err)
}

func TestV1_TransactionSanitize(t *testing.T) {
	msg := createTestV1Message(t)
	tx := &Transaction{Message: msg, Signatures: []Signature{{}}}
	require.NoError(t, tx.Sanitize())

	tx.Signatures = nil
	err := tx.Sanitize()
	require.Error(t, err)
	assert.True(t, IsSanitizeError(err))

	tx.Signatures = []Signature{{}}
	tx.Message.TransactionConfig = tx.Message.TransactionConfig.WithHeapSize(MaxHeapSizeV1 + 1024)
	err = tx.Sanitize()
	require.Error(t, err)
	assert.True(t, IsSanitizeError(err))
}

func TestV1_TransactionJSONRoundtrip(t *testing.T) {
	tx, err := TransactionFromBytes(mustHex(t, v1GoldenTxFullConfig))
	require.NoError(t, err)

	data, err := json.Marshal(tx)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"transactionConfig":{"priorityFee":5000,"computeUnitLimit":200000,"loadedAccountsDataSizeLimit":65536,"heapSize":65536}`)

	var back Transaction
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, MessageVersionV1, back.Message.GetVersion())
	assert.Equal(t, tx.Signatures, back.Signatures)
	assert.Equal(t, tx.Message.Header, back.Message.Header)
	assert.Equal(t, tx.Message.AccountKeys, back.Message.AccountKeys)
	assert.Equal(t, tx.Message.Instructions, back.Message.Instructions)
	assert.Equal(t, tx.Message.TransactionConfig, back.Message.TransactionConfig)

	// The JSON-decoded transaction re-encodes to the original wire bytes.
	again, err := back.MarshalBinary()
	require.NoError(t, err)
	assert.Equal(t, mustHex(t, v1GoldenTxFullConfig), again)
}

func TestV1_UsesDurableNonceUnaffected(t *testing.T) {
	tx, err := TransactionFromBytes(mustHex(t, v1GoldenTxEmptyConfig))
	require.NoError(t, err)
	assert.False(t, tx.UsesDurableNonce())
	assert.Equal(t, 1, tx.NumSigners())
	assert.Equal(t, 2, tx.NumReadonlyAccounts())
	assert.Equal(t, 2, tx.NumWriteableAccounts()) // payer + [0x02;32]
}

func TestV1_DecoderIsGenericPath(t *testing.T) {
	// TransactionFromDecoder / bin.Decoder.Decode also work for V1.
	raw := mustHex(t, v1GoldenTxFeeHeap)
	tx, err := TransactionFromDecoder(bin.NewBinDecoder(raw))
	require.NoError(t, err)
	assert.Equal(t, MessageVersionV1, tx.Message.GetVersion())
	assert.Equal(t, uint64(1), *tx.Message.TransactionConfig.PriorityFee)
	assert.Equal(t, uint32(32_768), *tx.Message.TransactionConfig.HeapSize)
}
