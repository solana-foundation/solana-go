package token2022

import (
	"testing"

	ag_require "github.com/stretchr/testify/require"
)

// tlvEntry builds one TLV entry: u16 LE type, u16 LE length, payload.
func tlvEntry(t ExtensionType, payload []byte) []byte {
	return concat(u16LE(uint16(t)), u16LE(uint16(len(payload))), payload)
}

// extendedMintData assembles full mint account data: base mint, zero padding
// to offset 165, the mint account type byte, and the given TLV entries.
func extendedMintData(entries ...[]byte) []byte {
	data := concat(testMintSlice, repeatByte(0, ACCOUNT_SIZE-MINT_SIZE), []byte{AccountTypeMint})
	return concat(append([][]byte{data}, entries...)...)
}

// extendedAccountData assembles full token account data: base account, the
// account type byte, and the given TLV entries.
func extendedAccountData(entries ...[]byte) []byte {
	data := concat(testAccountSlice, []byte{AccountTypeAccount})
	return concat(append([][]byte{data}, entries...)...)
}

func TestDecodeMintWithExtensions_BaseOnly(t *testing.T) {
	m, err := DecodeMintWithExtensions(testMintSlice)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(42), m.Mint.Supply)
	ag_require.False(t, m.HasExtensions())

	out, err := m.MarshalBinary()
	ag_require.NoError(t, err)
	ag_require.Equal(t, testMintSlice, out)
}

func TestDecodeAccountWithExtensions_BaseOnly(t *testing.T) {
	a, err := DecodeAccountWithExtensions(testAccountSlice)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(3), a.Account.Amount)
	ag_require.False(t, a.HasExtensions())

	out, err := a.MarshalBinary()
	ag_require.NoError(t, err)
	ag_require.Equal(t, testAccountSlice, out)
}

func TestRoundTrip_MintWithExtensions(t *testing.T) {
	authority := pubkeyOf(0x0A)
	metadata := TokenMetadataState{
		UpdateAuthority: NewOptionalPubkey(&authority),
		Mint:            pubkeyOf(0x0B),
		Name:            "Example",
		Symbol:          "EXM",
		Uri:             "https://example.com/meta.json",
		AdditionalMetadata: []MetadataField{
			{Key: "kind", Value: "test"},
		},
	}
	metadataPayload := marshalStateBytes(t, metadata)

	// Deliberately non-ascending TLV order, with an unknown extension type
	// in the middle, followed by a zeroed tail (realloc slack).
	data := extendedMintData(
		tlvEntry(ExtensionMetadataPointer, concat(repeatByte(0xC1, 32), repeatByte(0xC2, 32))),
		tlvEntry(ExtensionType(999), []byte{9, 9, 9, 9, 9}),
		tlvEntry(ExtensionTransferFeeConfig, concat(
			repeatByte(0x11, 32), repeatByte(0x22, 32), u64LE(777),
			u64LE(5), u64LE(1000), u16LE(50),
			u64LE(6), u64LE(2000), u16LE(75),
		)),
		tlvEntry(ExtensionNonTransferable, nil),
		tlvEntry(ExtensionTokenMetadata, metadataPayload),
		repeatByte(0, 6),
	)

	m, err := DecodeMintWithExtensions(data)
	ag_require.NoError(t, err)
	ag_require.True(t, m.HasExtensions())

	ag_require.NotNil(t, m.MetadataPointer)
	ag_require.Equal(t, pubkeyOf(0xC1), m.MetadataPointer.Authority.Key)
	ag_require.Equal(t, pubkeyOf(0xC2), m.MetadataPointer.MetadataAddress.Key)

	ag_require.NotNil(t, m.TransferFeeConfig)
	ag_require.Equal(t, uint64(777), m.TransferFeeConfig.WithheldAmount)
	ag_require.Equal(t, uint16(75), m.TransferFeeConfig.NewerTransferFee.TransferFeeBasisPoints)

	ag_require.True(t, m.NonTransferable)

	ag_require.NotNil(t, m.TokenMetadata)
	ag_require.Equal(t, "Example", m.TokenMetadata.Name)
	ag_require.Equal(t, metadata.AdditionalMetadata, m.TokenMetadata.AdditionalMetadata)

	ag_require.Len(t, m.Unknown, 1)
	ag_require.Equal(t, ExtensionType(999), m.Unknown[0].Type)
	ag_require.Equal(t, []byte{9, 9, 9, 9, 9}, m.Unknown[0].Data)

	// Byte-exact re-encode preserves the original non-ascending TLV order
	// and the unknown entry. The zeroed tail is not reproduced (it is
	// decode-tolerated slack, not TLV content).
	out, err := m.MarshalBinary()
	ag_require.NoError(t, err)
	ag_require.Equal(t, data[:len(data)-6], out)

	// Decoding the re-encoded bytes yields the same result.
	m2, err := DecodeMintWithExtensions(out)
	ag_require.NoError(t, err)
	ag_require.Equal(t, m, m2)
}

func TestRoundTrip_AccountWithExtensions(t *testing.T) {
	data := extendedAccountData(
		tlvEntry(ExtensionCpiGuard, []byte{1}),
		tlvEntry(ExtensionImmutableOwner, nil),
		tlvEntry(ExtensionTransferFeeAmount, u64LE(123456789)),
	)

	a, err := DecodeAccountWithExtensions(data)
	ag_require.NoError(t, err)
	ag_require.True(t, a.HasExtensions())
	ag_require.NotNil(t, a.CpiGuard)
	ag_require.True(t, a.CpiGuard.LockCpi)
	ag_require.True(t, a.ImmutableOwner)
	ag_require.NotNil(t, a.TransferFeeAmount)
	ag_require.Equal(t, uint64(123456789), a.TransferFeeAmount.WithheldAmount)
	ag_require.False(t, a.NonTransferableAccount)
	ag_require.False(t, a.PausableAccount)

	out, err := a.MarshalBinary()
	ag_require.NoError(t, err)
	ag_require.Equal(t, data, out)
}

func TestDecodeMintWithExtensions_Strictness(t *testing.T) {
	valid := extendedMintData(tlvEntry(ExtensionMintCloseAuthority, repeatByte(0x11, 32)))

	t.Run("too short for base", func(t *testing.T) {
		_, err := DecodeMintWithExtensions(testMintSlice[:50])
		ag_require.ErrorIs(t, err, ErrInvalidAccountData)
	})
	t.Run("between base and TLV start", func(t *testing.T) {
		_, err := DecodeMintWithExtensions(concat(testMintSlice, repeatByte(0, 18)))
		ag_require.ErrorIs(t, err, ErrInvalidAccountData)
	})
	t.Run("multisig length rejected", func(t *testing.T) {
		data := concat(valid, repeatByte(0, multisigSize-len(valid)))
		ag_require.Equal(t, multisigSize, len(data))
		_, err := DecodeMintWithExtensions(data)
		ag_require.ErrorIs(t, err, ErrInvalidAccountData)
	})
	t.Run("nonzero padding", func(t *testing.T) {
		data := append([]byte(nil), valid...)
		data[100] = 1
		_, err := DecodeMintWithExtensions(data)
		ag_require.ErrorIs(t, err, ErrInvalidAccountData)
	})
	t.Run("wrong account type byte", func(t *testing.T) {
		data := append([]byte(nil), valid...)
		data[ACCOUNT_SIZE] = AccountTypeAccount
		_, err := DecodeMintWithExtensions(data)
		ag_require.ErrorIs(t, err, ErrInvalidAccountData)
	})
	t.Run("account extension on mint", func(t *testing.T) {
		data := extendedMintData(tlvEntry(ExtensionTransferFeeAmount, u64LE(1)))
		_, err := DecodeMintWithExtensions(data)
		ag_require.ErrorIs(t, err, ErrExtensionTypeMismatch)
	})
	t.Run("mint extension on account", func(t *testing.T) {
		data := extendedAccountData(tlvEntry(ExtensionMintCloseAuthority, repeatByte(0x11, 32)))
		_, err := DecodeAccountWithExtensions(data)
		ag_require.ErrorIs(t, err, ErrExtensionTypeMismatch)
	})
	t.Run("duplicate extension", func(t *testing.T) {
		data := extendedMintData(
			tlvEntry(ExtensionMintCloseAuthority, repeatByte(0x11, 32)),
			tlvEntry(ExtensionMintCloseAuthority, repeatByte(0x22, 32)),
		)
		_, err := DecodeMintWithExtensions(data)
		ag_require.ErrorIs(t, err, ErrDuplicateExtension)
	})
	t.Run("wrong extension length", func(t *testing.T) {
		data := extendedMintData(tlvEntry(ExtensionMintCloseAuthority, repeatByte(0x11, 31)))
		_, err := DecodeMintWithExtensions(data)
		ag_require.ErrorIs(t, err, ErrInvalidExtensionLength)
	})
	t.Run("marker with nonzero length", func(t *testing.T) {
		data := extendedMintData(tlvEntry(ExtensionNonTransferable, []byte{1}))
		_, err := DecodeMintWithExtensions(data)
		ag_require.ErrorIs(t, err, ErrInvalidExtensionLength)
	})
	t.Run("valid baseline decodes", func(t *testing.T) {
		m, err := DecodeMintWithExtensions(valid)
		ag_require.NoError(t, err)
		ag_require.NotNil(t, m.MintCloseAuthority)
		ag_require.Equal(t, pubkeyOf(0x11), m.MintCloseAuthority.CloseAuthority.Key)
	})
}

func TestMarshalBinary_HandBuiltMint_MultisigBump(t *testing.T) {
	// Hand-built mints encode extensions in ascending type order:
	// MetadataPointer(18), TokenGroup(21), Pausable(26).
	// 166 + (4+64) + (4+80) + (4+33) = 355, which collides with the multisig
	// length and must be padded to 357.
	groupAuthority := pubkeyOf(0x51)
	pauseAuthority := pubkeyOf(0x52)
	ptrAuthority := pubkeyOf(0x53)

	m := &MintWithExtensions{
		Mint: Mint{Supply: 9, Decimals: 6, IsInitialized: true},
		TokenGroup: &TokenGroup{
			UpdateAuthority: NewOptionalPubkey(&groupAuthority),
			Mint:            pubkeyOf(0x54),
			Size:            7,
			MaxSize:         1 << 40,
		},
		Pausable: &PausableState{Authority: NewOptionalPubkey(&pauseAuthority), Paused: true},
		MetadataPointer: &MetadataPointerState{
			Authority:       NewOptionalPubkey(&ptrAuthority),
			MetadataAddress: NewOptionalPubkey(&ptrAuthority),
		},
	}

	want, err := CalculateMintLen([]ExtensionType{ExtensionTokenGroup, ExtensionPausable, ExtensionMetadataPointer})
	ag_require.NoError(t, err)
	ag_require.Equal(t, 357, want)

	out, err := m.MarshalBinary()
	ag_require.NoError(t, err)
	ag_require.Equal(t, want, len(out))
	ag_require.Equal(t, []byte{0, 0}, out[355:])

	// Extensions appear in ascending type order.
	ag_require.Equal(t, u16LE(uint16(ExtensionMetadataPointer)), out[166:168])
	ag_require.Equal(t, u16LE(uint16(ExtensionTokenGroup)), out[234:236])
	ag_require.Equal(t, u16LE(uint16(ExtensionPausable)), out[318:320])

	// The trailing zero bytes decode as a terminator, so the data round-trips.
	m2, err := DecodeMintWithExtensions(out)
	ag_require.NoError(t, err)
	ag_require.Equal(t, m.Mint, m2.Mint)
	ag_require.Equal(t, m.TokenGroup, m2.TokenGroup)
	ag_require.Equal(t, m.Pausable, m2.Pausable)
	ag_require.Equal(t, m.MetadataPointer, m2.MetadataPointer)
}

func TestMarshalBinary_ClearedExtensionIsSkipped(t *testing.T) {
	data := extendedAccountData(
		tlvEntry(ExtensionCpiGuard, []byte{1}),
		tlvEntry(ExtensionTransferFeeAmount, u64LE(42)),
	)
	a, err := DecodeAccountWithExtensions(data)
	ag_require.NoError(t, err)

	a.CpiGuard = nil
	out, err := a.MarshalBinary()
	ag_require.NoError(t, err)

	a2, err := DecodeAccountWithExtensions(out)
	ag_require.NoError(t, err)
	ag_require.Nil(t, a2.CpiGuard)
	ag_require.NotNil(t, a2.TransferFeeAmount)
	ag_require.Equal(t, uint64(42), a2.TransferFeeAmount.WithheldAmount)
}

func TestMarshalBinary_EditedUnknownStaysConsistent(t *testing.T) {
	// Regression: unknown entries are re-encoded with their own type+data
	// pair. Removing one from Unknown after decode must not shift another
	// entry's data under the removed entry's type.
	data := extendedMintData(
		tlvEntry(ExtensionType(999), []byte{9, 9, 9}),
		tlvEntry(ExtensionType(998), []byte{8, 8}),
	)
	m, err := DecodeMintWithExtensions(data)
	ag_require.NoError(t, err)
	ag_require.Len(t, m.Unknown, 2)

	m.Unknown = m.Unknown[1:] // drop the 999 entry
	out, err := m.MarshalBinary()
	ag_require.NoError(t, err)

	want := extendedMintData(tlvEntry(ExtensionType(998), []byte{8, 8}))
	ag_require.Equal(t, want, out)
}

func TestParseExtensions_DataAppendDoesNotClobber(t *testing.T) {
	// Regression: ExtensionTLV.Data capacity is capped, so appending to one
	// entry's Data must allocate instead of overwriting the next entry.
	buf := concat(
		tlvEntry(ExtensionCpiGuard, []byte{1}),
		tlvEntry(ExtensionMemoTransfer, []byte{1}),
	)
	tlvs, err := ParseExtensions(buf)
	ag_require.NoError(t, err)
	ag_require.Len(t, tlvs, 2)

	_ = append(tlvs[0].Data, 0xFF, 0xFF, 0xFF)
	ag_require.Equal(t, []byte{1}, tlvs[1].Data)
	ag_require.Equal(t, concat(
		tlvEntry(ExtensionCpiGuard, []byte{1}),
		tlvEntry(ExtensionMemoTransfer, []byte{1}),
	), buf)
}

func TestMarshalBinary_AddedExtensionAppended(t *testing.T) {
	data := extendedAccountData(tlvEntry(ExtensionCpiGuard, []byte{1}))
	a, err := DecodeAccountWithExtensions(data)
	ag_require.NoError(t, err)

	// Adding a lower-numbered extension after decode appends it after the
	// original entries rather than reordering them.
	a.TransferFeeAmount = &TransferFeeAmountState{WithheldAmount: 5}
	out, err := a.MarshalBinary()
	ag_require.NoError(t, err)

	wantPrefix := extendedAccountData(
		tlvEntry(ExtensionCpiGuard, []byte{1}),
		tlvEntry(ExtensionTransferFeeAmount, u64LE(5)),
	)
	ag_require.Equal(t, wantPrefix, out)
}
