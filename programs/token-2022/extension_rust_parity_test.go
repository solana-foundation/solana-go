package token2022

import (
	"encoding/base64"
	"testing"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_require "github.com/stretchr/testify/require"
)

// The tests in this file are ported from the test module of
// interface/src/extension/mod.rs in the Rust token-2022 repository, using the
// same byte fixtures. Tests that exercise the program-runtime mutable API
// (alloc, realloc, init_extension, set_account_type) are not applicable to
// this client-side port and are represented by their decode-visible effects.

// mintWithAccountType mirrors the Rust MINT_WITH_ACCOUNT_TYPE fixture:
// base mint, zero padding to offset 165, AccountType::Mint.
func mintWithAccountType() []byte {
	return concat(testMintSlice, repeatByte(0, 83), []byte{1})
}

// mintWithExtension mirrors the Rust MINT_WITH_EXTENSION fixture:
// mintWithAccountType plus a MintCloseAuthority TLV entry of [1; 32].
func mintWithExtension() []byte {
	return concat(mintWithAccountType(), []byte{3, 0, 32, 0}, repeatByte(1, 32))
}

// accountWithExtension mirrors the Rust ACCOUNT_WITH_EXTENSION fixture:
// base account, AccountType::Account, and a TransferHookAccount TLV entry.
func accountWithExtension() []byte {
	return concat(testAccountSlice, []byte{2}, []byte{15, 0, 1, 0, 1})
}

// Ported from mod.rs unpack_opaque_buffer.
func TestRustParity_UnpackOpaqueBuffer(t *testing.T) {
	// Mint with account type byte but no TLV entries.
	m, err := DecodeMintWithExtensions(mintWithAccountType())
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(42), m.Mint.Supply)
	ag_require.False(t, m.HasExtensions())

	// Mint with a MintCloseAuthority extension.
	m, err = DecodeMintWithExtensions(mintWithExtension())
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(42), m.Mint.Supply)
	ag_require.NotNil(t, m.MintCloseAuthority)
	ag_require.Equal(t, pubkeyOf(1), m.MintCloseAuthority.CloseAuthority.Key)
	// TransferFeeConfig is not present (Rust: get_extension errors).
	ag_require.Nil(t, m.TransferFeeConfig)

	// Unpacking mint data as an account fails (Rust: UninitializedAccount).
	_, err = DecodeAccountWithExtensions(mintWithExtension())
	ag_require.Error(t, err)

	// Plain base mint slice decodes with no extensions.
	m, err = DecodeMintWithExtensions(testMintSlice)
	ag_require.NoError(t, err)
	ag_require.False(t, m.HasExtensions())

	// Account with a TransferHookAccount extension.
	a, err := DecodeAccountWithExtensions(accountWithExtension())
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(3), a.Account.Amount)
	ag_require.NotNil(t, a.TransferHookAccount)
	ag_require.True(t, a.TransferHookAccount.Transferring)

	// Unpacking account data as a mint fails (Rust: InvalidAccountData).
	_, err = DecodeMintWithExtensions(accountWithExtension())
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// Plain base account slice decodes with no extensions.
	a, err = DecodeAccountWithExtensions(testAccountSlice)
	ag_require.NoError(t, err)
	ag_require.False(t, a.HasExtensions())
}

// Ported from mod.rs mint_fail_unpack_opaque_buffer.
func TestRustParity_MintFailUnpackOpaqueBuffer(t *testing.T) {
	// input buffer too small
	_, err := DecodeMintWithExtensions([]byte{0, 3})
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// tweak the account type
	buffer := mintWithExtension()
	buffer[ACCOUNT_SIZE] = 3
	_, err = DecodeMintWithExtensions(buffer)
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// clear the mint initialized byte
	buffer = mintWithExtension()
	buffer[45] = 0
	_, err = DecodeMintWithExtensions(buffer)
	ag_require.ErrorIs(t, err, ErrUninitializedAccount)

	// tweak the padding
	buffer = mintWithExtension()
	buffer[MINT_SIZE] = 100
	_, err = DecodeMintWithExtensions(buffer)
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// tweak the extension type to an account extension (TransferFeeAmount).
	// Rust defers the error to get_extension; our decoder rejects eagerly.
	buffer = mintWithExtension()
	buffer[ACCOUNT_SIZE+1] = 2
	_, err = DecodeMintWithExtensions(buffer)
	ag_require.ErrorIs(t, err, ErrExtensionTypeMismatch)

	// tweak the length, too big: the value overruns the buffer
	buffer = mintWithExtension()
	buffer[ACCOUNT_SIZE+3] = 100
	_, err = DecodeMintWithExtensions(buffer)
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// tweak the length, too small: the remaining bytes parse as garbage TLV
	buffer = mintWithExtension()
	buffer[ACCOUNT_SIZE+3] = 10
	_, err = DecodeMintWithExtensions(buffer)
	ag_require.Error(t, err)

	// data buffer is too small
	buffer = mintWithExtension()
	_, err = DecodeMintWithExtensions(buffer[:len(buffer)-1])
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)
}

// Ported from mod.rs account_fail_unpack_opaque_buffer.
func TestRustParity_AccountFailUnpackOpaqueBuffer(t *testing.T) {
	// input buffer too small
	_, err := DecodeAccountWithExtensions([]byte{0, 3})
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// all 5's: not a valid AccountState
	_, err = DecodeAccountWithExtensions(repeatByte(5, ACCOUNT_SIZE))
	ag_require.ErrorIs(t, err, ErrUninitializedAccount)

	// tweak the account type
	buffer := accountWithExtension()
	buffer[ACCOUNT_SIZE] = 3
	_, err = DecodeAccountWithExtensions(buffer)
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// clear the state byte
	buffer = accountWithExtension()
	buffer[108] = 0
	_, err = DecodeAccountWithExtensions(buffer)
	ag_require.ErrorIs(t, err, ErrUninitializedAccount)

	// tweak the extension type to a mint extension (PermanentDelegate)
	buffer = accountWithExtension()
	buffer[ACCOUNT_SIZE+1] = 12
	_, err = DecodeAccountWithExtensions(buffer)
	ag_require.ErrorIs(t, err, ErrExtensionTypeMismatch)

	// tweak the length, too big
	buffer = accountWithExtension()
	buffer[ACCOUNT_SIZE+3] = 100
	_, err = DecodeAccountWithExtensions(buffer)
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// tweak the length, larger than the remaining value bytes
	buffer = accountWithExtension()
	buffer[ACCOUNT_SIZE+3] = 10
	_, err = DecodeAccountWithExtensions(buffer)
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// data buffer is too small
	buffer = accountWithExtension()
	_, err = DecodeAccountWithExtensions(buffer[:len(buffer)-1])
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)
}

// Ported from mod.rs get_extension_types_with_opaque_buffer, against the raw
// TLV walker.
func TestRustParity_GetExtensionTypesWithOpaqueBuffer(t *testing.T) {
	// incorrect due to the length
	_, err := ParseExtensions([]byte{1, 0, 1, 1})
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// huge enum number: Rust rejects unknown extension types; this port
	// deliberately tolerates them for forward compatibility.
	tlvs, err := ParseExtensions([]byte{0, 1, 0, 0})
	ag_require.NoError(t, err)
	ag_require.Len(t, tlvs, 1)
	ag_require.Equal(t, ExtensionType(256), tlvs[0].Type)

	// correct: good enum number and zero length
	tlvs, err = ParseExtensions([]byte{1, 0, 0, 0})
	ag_require.NoError(t, err)
	ag_require.Len(t, tlvs, 1)
	ag_require.Equal(t, ExtensionTransferFeeConfig, tlvs[0].Type)
	ag_require.Equal(t, uint16(0), tlvs[0].Length)

	// correct: just uninitialized data at the end
	tlvs, err = ParseExtensions([]byte{0, 0})
	ag_require.NoError(t, err)
	ag_require.Empty(t, tlvs)
}

// Ported from mod.rs mint_extension_any_order.
func TestRustParity_MintExtensionAnyOrder(t *testing.T) {
	closeAuthorityEntry := tlvEntry(ExtensionMintCloseAuthority, repeatByte(1, 32))
	feeConfigEntry := tlvEntry(ExtensionTransferFeeConfig, marshalStateBytes(t, testTransferFeeConfig()))

	buffer := extendedMintData(closeAuthorityEntry, feeConfigEntry)
	otherBuffer := extendedMintData(feeConfigEntry, closeAuthorityEntry)

	// buffers are NOT the same because written in a different order
	ag_require.NotEqual(t, buffer, otherBuffer)

	state, err := DecodeMintWithExtensions(buffer)
	ag_require.NoError(t, err)
	otherState, err := DecodeMintWithExtensions(otherBuffer)
	ag_require.NoError(t, err)

	// BUT mint and extensions are the same
	ag_require.Equal(t, state.TransferFeeConfig, otherState.TransferFeeConfig)
	ag_require.Equal(t, state.MintCloseAuthority, otherState.MintCloseAuthority)
	ag_require.Equal(t, state.Mint, otherState.Mint)

	// and each re-encodes byte-exactly in its own order
	out, err := state.MarshalBinary()
	ag_require.NoError(t, err)
	ag_require.Equal(t, buffer, out)
	out, err = otherState.MarshalBinary()
	ag_require.NoError(t, err)
	ag_require.Equal(t, otherBuffer, out)
}

// Ported from mod.rs mint_with_multisig_len / account_with_multisig_len.
// The Rust test uses the test-only MintPaddingTest (65535) and
// AccountPaddingTest (65534) extensions, 185 bytes each, chosen so the total
// lands exactly on the 355-byte multisig length and gets padded to 357.
func TestRustParity_MultisigLen(t *testing.T) {
	// A buffer of exactly Multisig::LEN is never valid extended state.
	_, err := DecodeMintWithExtensions(make([]byte, multisigSize))
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)
	_, err = DecodeAccountWithExtensions(make([]byte, multisigSize))
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// Mint: the expected raw buffer from the Rust test, terminator included.
	expect := concat(
		testMintSlice,
		repeatByte(0, 83),
		[]byte{1},
		u16LE(65535), u16LE(185), repeatByte(1, 185),
		u16LE(0),
	)
	ag_require.Equal(t, multisigSize+2, len(expect))

	m, err := DecodeMintWithExtensions(expect)
	ag_require.NoError(t, err)
	ag_require.Len(t, m.Unknown, 1)
	ag_require.Equal(t, ExtensionType(65535), m.Unknown[0].Type)
	ag_require.Equal(t, repeatByte(1, 185), m.Unknown[0].Data)

	out, err := m.MarshalBinary()
	ag_require.NoError(t, err)
	ag_require.Equal(t, expect, out)

	// Account: same shape with AccountPaddingTest (65534).
	expect = concat(
		testAccountSlice,
		[]byte{2},
		u16LE(65534), u16LE(185), repeatByte(1, 185),
		u16LE(0),
	)
	ag_require.Equal(t, multisigSize+2, len(expect))

	a, err := DecodeAccountWithExtensions(expect)
	ag_require.NoError(t, err)
	ag_require.Len(t, a.Unknown, 1)
	ag_require.Equal(t, ExtensionType(65534), a.Unknown[0].Type)

	out, err = a.MarshalBinary()
	ag_require.NoError(t, err)
	ag_require.Equal(t, expect, out)
}

// Ported from mod.rs mint_without_extensions.
func TestRustParity_MintWithoutExtensions(t *testing.T) {
	space, err := CalculateMintLen(nil)
	ag_require.NoError(t, err)
	ag_require.Equal(t, MINT_SIZE, space)

	// unpacking base mint data as an account fails
	_, err = DecodeAccountWithExtensions(testMintSlice)
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// a mint without extensions round-trips to exactly the base slice
	m, err := DecodeMintWithExtensions(testMintSlice)
	ag_require.NoError(t, err)
	out, err := m.MarshalBinary()
	ag_require.NoError(t, err)
	ag_require.Equal(t, testMintSlice, out)
}

// Ported from mod.rs test_extension_with_no_data.
func TestRustParity_ExtensionWithNoData(t *testing.T) {
	accountSize, err := CalculateTokenAccountLen([]ExtensionType{ExtensionImmutableOwner})
	ag_require.NoError(t, err)
	ag_require.Equal(t, ACCOUNT_SIZE+1+4, accountSize)

	data := extendedAccountData(tlvEntry(ExtensionImmutableOwner, nil))
	ag_require.Equal(t, accountSize, len(data))

	a, err := DecodeAccountWithExtensions(data)
	ag_require.NoError(t, err)
	ag_require.True(t, a.ImmutableOwner)

	tlvs, err := ParseExtensions(data[tlvStartOffset:])
	ag_require.NoError(t, err)
	ag_require.Len(t, tlvs, 1)
	ag_require.Equal(t, ExtensionImmutableOwner, tlvs[0].Type)
	ag_require.Equal(t, uint16(0), tlvs[0].Length)
}

// Ported from mod.rs fail_account_len_with_metadata.
func TestRustParity_FailAccountLenWithMetadata(t *testing.T) {
	_, err := CalculateMintLen([]ExtensionType{
		ExtensionMintCloseAuthority,
		ExtensionTokenMetadata,
		ExtensionTransferFeeConfig,
	})
	ag_require.ErrorIs(t, err, ErrUnsizedExtension)
}

// mainnetXStockMintB64 is the on-chain account data of the token-2022 mint
// Xs3oZwbHvqis4NYcf4YKWmEia2eC84wSiVrcYcTqpH8 (SpaceX xStock), fetched from
// mainnet at slot 437873054. 679 bytes, 8 extensions.
const mainnetXStockMintB64 = "AQAAAGVqQkIv6okUBqQZ0dHeCPQqhHlBtaGulevOYZrDFyk0XOODfu4yAAAIAQEAAAD/3+wbzSzTg5PITaoIyRzA041nf/jQq3tdAz8A9zLMMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARIAQABD+fHuLje4B+pFy3TmAJcivsAJKWAm5ORVi0NeJKXpxgfoA5u0rUz3nDg+Q8DFHdGFN0BRK3nQ+UISkjRR0AODDAAgAEP58e4uN7gH6kXLdOYAlyK+wAkpYCbk5FWLQ14kpenGBgABAAEZADgABm9ZIlHMR3R4JaWa0UIupDVz9SjaXe4q94ErMU+ZReMAAAAAAADwPwAAAAAAAAAAAAAAAAAA8D8aACEA/9/sG80s04OTyE2qCMkcwNONZ3/40Kt7XQM/APcyzDAABABBAEP58e4uN7gH6kXLdOYAlyK+wAkpYCbk5FWLQ14kpenGAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADgBAAEP58e4uN7gH6kXLdOYAlyK+wAkpYCbk5FWLQ14kpenGAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAATAKYAQ/nx7i43uAfqRct05gCXIr7ACSlgJuTkVYtDXiSl6cYH6AObtK1M95w4PkPAxR3RhTdAUSt50PlCEpI0UdADgw0AAABTcGFjZVggeFN0b2NrBQAAAFNQQ1h4RAAAAGh0dHBzOi8veHN0b2Nrcy1tZXRhZGF0YS5iYWNrZWQuZmkvdG9rZW5zL1NvbGFuYS9TUENYeC9tZXRhZGF0YS5qc29uAAAAAA=="

// TestMainnetVector_XStockMint decodes a real mainnet token-2022 mint with
// eight extensions and re-encodes it byte-exactly.
func TestMainnetVector_XStockMint(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString(mainnetXStockMintB64)
	ag_require.NoError(t, err)
	ag_require.Equal(t, 679, len(data))

	m, err := DecodeMintWithExtensions(data)
	ag_require.NoError(t, err)

	// Base mint.
	backedAuthority := ag_solanago.MustPublicKeyFromBase58("7pt9tkctJPK7PPNQJ77GKg8ZffSF6QxoMiCFYHxrtaCj")
	freezeAuthority := ag_solanago.MustPublicKeyFromBase58("JDq14BWvqCRFNu1krb12bcRpbGtJZ1FLEakMw6FdxJNs")
	ag_require.NotNil(t, m.Mint.MintAuthority)
	ag_require.Equal(t, backedAuthority, *m.Mint.MintAuthority)
	ag_require.NotNil(t, m.Mint.FreezeAuthority)
	ag_require.Equal(t, freezeAuthority, *m.Mint.FreezeAuthority)
	ag_require.Equal(t, uint64(55999906177884), m.Mint.Supply)
	ag_require.Equal(t, uint8(8), m.Mint.Decimals)
	ag_require.True(t, m.Mint.IsInitialized)

	mintAddress := ag_solanago.MustPublicKeyFromBase58("Xs3oZwbHvqis4NYcf4YKWmEia2eC84wSiVrcYcTqpH8")
	adminAuthority := ag_solanago.MustPublicKeyFromBase58("5aMNNLQJwAEeoemTEMkv5NVjqKwvvefRYCQ5Z67HFvEq")

	// MetadataPointer: points at the mint itself.
	ag_require.NotNil(t, m.MetadataPointer)
	ag_require.Equal(t, adminAuthority, m.MetadataPointer.Authority.Key)
	ag_require.Equal(t, mintAddress, m.MetadataPointer.MetadataAddress.Key)

	// PermanentDelegate.
	ag_require.NotNil(t, m.PermanentDelegate)
	ag_require.Equal(t, adminAuthority, m.PermanentDelegate.Delegate.Key)

	// DefaultAccountState: Initialized.
	ag_require.NotNil(t, m.DefaultAccountState)
	ag_require.Equal(t, AccountStateInitialized, m.DefaultAccountState.State)

	// ScaledUiAmount: multiplier 1.0, effective from timestamp 0.
	ag_require.NotNil(t, m.ScaledUiAmount)
	ag_require.Equal(t, 1.0, m.ScaledUiAmount.Multiplier)
	ag_require.Equal(t, int64(0), m.ScaledUiAmount.NewMultiplierEffectiveTimestamp)
	ag_require.Equal(t, 1.0, m.ScaledUiAmount.NewMultiplier)

	// Pausable: not paused.
	ag_require.NotNil(t, m.Pausable)
	ag_require.Equal(t, freezeAuthority, m.Pausable.Authority.Key)
	ag_require.False(t, m.Pausable.Paused)

	// ConfidentialTransferMint: manual approval, no auditor.
	ag_require.NotNil(t, m.ConfidentialTransferMint)
	ag_require.Equal(t, adminAuthority, m.ConfidentialTransferMint.Authority.Key)
	ag_require.False(t, m.ConfidentialTransferMint.AutoApproveNewAccounts)
	ag_require.Equal(t, [32]byte{}, m.ConfidentialTransferMint.AuditorElGamalPubkey)

	// TransferHook: authority set, no program.
	ag_require.NotNil(t, m.TransferHook)
	ag_require.Equal(t, adminAuthority, m.TransferHook.Authority.Key)
	ag_require.True(t, m.TransferHook.ProgramID.IsNone())

	// TokenMetadata.
	ag_require.NotNil(t, m.TokenMetadata)
	ag_require.Equal(t, adminAuthority, m.TokenMetadata.UpdateAuthority.Key)
	ag_require.Equal(t, mintAddress, m.TokenMetadata.Mint)
	ag_require.Equal(t, "SpaceX xStock", m.TokenMetadata.Name)
	ag_require.Equal(t, "SPCXx", m.TokenMetadata.Symbol)
	ag_require.Equal(t, "https://xstocks-metadata.backed.fi/tokens/Solana/SPCXx/metadata.json", m.TokenMetadata.Uri)
	ag_require.Empty(t, m.TokenMetadata.AdditionalMetadata)

	// No unknown extensions, and the whole account re-encodes byte-exactly.
	ag_require.Empty(t, m.Unknown)
	out, err := m.MarshalBinary()
	ag_require.NoError(t, err)
	ag_require.Equal(t, data, out)
}
