package token2022

import (
	"bytes"
	"math"
	"testing"

	ag_binary "github.com/gagliardetto/binary"
	ag_require "github.com/stretchr/testify/require"
)

func i64LE(v int64) []byte {
	return u64LE(uint64(v))
}

func i16LE(v int16) []byte {
	return u16LE(uint16(v))
}

func f64LE(v float64) []byte {
	return u64LE(math.Float64bits(v))
}

func marshalStateBytes(t *testing.T, v ag_binary.BinaryMarshaler) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	err := v.MarshalWithEncoder(ag_binary.NewBinEncoder(buf))
	ag_require.NoError(t, err)
	return buf.Bytes()
}

// TestRoundTrip_FixedSizeExtensionStates builds a byte fixture for every
// fixed-size extension state at the exact offsets of the Rust
// spl-token-2022-interface layout, decodes it, asserts every field, and
// requires a byte-exact re-encode. It also checks that data one byte shorter
// or longer is rejected.
func TestRoundTrip_FixedSizeExtensionStates(t *testing.T) {
	cases := []struct {
		name    string
		fixture []byte
		decode  func([]byte) (ag_binary.BinaryMarshaler, error)
		check   func(t *testing.T, got any)
	}{
		{
			name: "TransferFeeConfig",
			fixture: concat(
				repeatByte(0x11, 32),             // transfer_fee_config_authority
				repeatByte(0x22, 32),             // withdraw_withheld_authority
				u64LE(0x1122334455667788),        // withheld_amount
				u64LE(5), u64LE(1000), u16LE(50), // older_transfer_fee
				u64LE(6), u64LE(2000), u16LE(75), // newer_transfer_fee
			),
			decode: func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeTransferFeeConfigState(d) },
			check: func(t *testing.T, got any) {
				s := got.(*TransferFeeConfigState)
				ag_require.Equal(t, pubkeyOf(0x11), s.TransferFeeConfigAuthority.Key)
				ag_require.Equal(t, pubkeyOf(0x22), s.WithdrawWithheldAuthority.Key)
				ag_require.Equal(t, uint64(0x1122334455667788), s.WithheldAmount)
				ag_require.Equal(t, TransferFee{Epoch: 5, MaximumFee: 1000, TransferFeeBasisPoints: 50}, s.OlderTransferFee)
				ag_require.Equal(t, TransferFee{Epoch: 6, MaximumFee: 2000, TransferFeeBasisPoints: 75}, s.NewerTransferFee)
			},
		},
		{
			name:    "TransferFeeAmount",
			fixture: u64LE(0xAABBCCDDEEFF0011),
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeTransferFeeAmountState(d) },
			check: func(t *testing.T, got any) {
				ag_require.Equal(t, uint64(0xAABBCCDDEEFF0011), got.(*TransferFeeAmountState).WithheldAmount)
			},
		},
		{
			name:    "MintCloseAuthority",
			fixture: repeatByte(0x33, 32),
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeMintCloseAuthorityState(d) },
			check: func(t *testing.T, got any) {
				ag_require.Equal(t, pubkeyOf(0x33), got.(*MintCloseAuthorityState).CloseAuthority.Key)
			},
		},
		{
			name: "ConfidentialTransferMint",
			fixture: concat(
				repeatByte(0x44, 32), // authority
				[]byte{1},            // auto_approve_new_accounts
				repeatByte(0x55, 32), // auditor_elgamal_pubkey
			),
			decode: func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeConfidentialTransferMintState(d) },
			check: func(t *testing.T, got any) {
				s := got.(*ConfidentialTransferMintState)
				ag_require.Equal(t, pubkeyOf(0x44), s.Authority.Key)
				ag_require.True(t, s.AutoApproveNewAccounts)
				ag_require.Equal(t, [32]byte(pubkeyOf(0x55)), s.AuditorElGamalPubkey)
			},
		},
		{
			name: "ConfidentialTransferAccount",
			fixture: concat(
				[]byte{1},                                  // approved
				repeatByte(0x61, 32),                       // elgamal_pubkey
				repeatByte(0x62, 64),                       // pending_balance_lo
				repeatByte(0x63, 64),                       // pending_balance_hi
				repeatByte(0x64, 64),                       // available_balance
				repeatByte(0x65, 36),                       // decryptable_available_balance
				[]byte{1},                                  // allow_confidential_credits
				[]byte{0},                                  // allow_non_confidential_credits
				u64LE(11), u64LE(12), u64LE(13), u64LE(14), // credit counters
			),
			decode: func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeConfidentialTransferAccountState(d) },
			check: func(t *testing.T, got any) {
				s := got.(*ConfidentialTransferAccountState)
				ag_require.True(t, s.Approved)
				ag_require.Equal(t, [32]byte(pubkeyOf(0x61)), s.ElGamalPubkey)
				ag_require.Equal(t, [64]byte(repeatByte(0x62, 64)), s.PendingBalanceLo)
				ag_require.Equal(t, [64]byte(repeatByte(0x63, 64)), s.PendingBalanceHi)
				ag_require.Equal(t, [64]byte(repeatByte(0x64, 64)), s.AvailableBalance)
				ag_require.Equal(t, [36]byte(repeatByte(0x65, 36)), s.DecryptableAvailableBalance)
				ag_require.True(t, s.AllowConfidentialCredits)
				ag_require.False(t, s.AllowNonConfidentialCredits)
				ag_require.Equal(t, uint64(11), s.PendingBalanceCreditCounter)
				ag_require.Equal(t, uint64(12), s.MaximumPendingBalanceCreditCounter)
				ag_require.Equal(t, uint64(13), s.ExpectedPendingBalanceCreditCounter)
				ag_require.Equal(t, uint64(14), s.ActualPendingBalanceCreditCounter)
			},
		},
		{
			name:    "DefaultAccountState",
			fixture: []byte{2},
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeDefaultAccountStateConfig(d) },
			check: func(t *testing.T, got any) {
				ag_require.Equal(t, AccountStateFrozen, got.(*DefaultAccountStateConfig).State)
			},
		},
		{
			name:    "MemoTransfer",
			fixture: []byte{1},
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeMemoTransferState(d) },
			check: func(t *testing.T, got any) {
				ag_require.True(t, got.(*MemoTransferState).RequireIncomingTransferMemos)
			},
		},
		{
			name: "InterestBearingConfig",
			fixture: concat(
				repeatByte(0x77, 32), // rate_authority
				i64LE(1700000001),    // initialization_timestamp
				i16LE(-1234),         // pre_update_average_rate
				i64LE(1700000002),    // last_update_timestamp
				i16LE(567),           // current_rate
			),
			decode: func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeInterestBearingConfigState(d) },
			check: func(t *testing.T, got any) {
				s := got.(*InterestBearingConfigState)
				ag_require.Equal(t, pubkeyOf(0x77), s.RateAuthority.Key)
				ag_require.Equal(t, int64(1700000001), s.InitializationTimestamp)
				ag_require.Equal(t, int16(-1234), s.PreUpdateAverageRate)
				ag_require.Equal(t, int64(1700000002), s.LastUpdateTimestamp)
				ag_require.Equal(t, int16(567), s.CurrentRate)
			},
		},
		{
			name:    "CpiGuard",
			fixture: []byte{1},
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeCpiGuardState(d) },
			check: func(t *testing.T, got any) {
				ag_require.True(t, got.(*CpiGuardState).LockCpi)
			},
		},
		{
			name:    "PermanentDelegate",
			fixture: repeatByte(0x88, 32),
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodePermanentDelegateState(d) },
			check: func(t *testing.T, got any) {
				ag_require.Equal(t, pubkeyOf(0x88), got.(*PermanentDelegateState).Delegate.Key)
			},
		},
		{
			name:    "TransferHook",
			fixture: concat(repeatByte(0x91, 32), repeatByte(0x92, 32)),
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeTransferHookState(d) },
			check: func(t *testing.T, got any) {
				s := got.(*TransferHookState)
				ag_require.Equal(t, pubkeyOf(0x91), s.Authority.Key)
				ag_require.Equal(t, pubkeyOf(0x92), s.ProgramID.Key)
			},
		},
		{
			name:    "TransferHookAccount",
			fixture: []byte{1},
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeTransferHookAccountState(d) },
			check: func(t *testing.T, got any) {
				ag_require.True(t, got.(*TransferHookAccountState).Transferring)
			},
		},
		{
			name: "ConfidentialTransferFeeConfig",
			fixture: concat(
				repeatByte(0xA1, 32), // authority
				repeatByte(0xA2, 32), // withdraw_withheld_authority_elgamal_pubkey
				[]byte{1},            // harvest_to_mint_enabled
				repeatByte(0xA3, 64), // withheld_amount
			),
			decode: func(d []byte) (ag_binary.BinaryMarshaler, error) {
				return DecodeConfidentialTransferFeeConfigState(d)
			},
			check: func(t *testing.T, got any) {
				s := got.(*ConfidentialTransferFeeConfigState)
				ag_require.Equal(t, pubkeyOf(0xA1), s.Authority.Key)
				ag_require.Equal(t, [32]byte(pubkeyOf(0xA2)), s.WithdrawWithheldAuthorityElGamalPubkey)
				ag_require.True(t, s.HarvestToMintEnabled)
				ag_require.Equal(t, [64]byte(repeatByte(0xA3, 64)), s.WithheldAmount)
			},
		},
		{
			name:    "ConfidentialTransferFeeAmount",
			fixture: repeatByte(0xB1, 64),
			decode: func(d []byte) (ag_binary.BinaryMarshaler, error) {
				return DecodeConfidentialTransferFeeAmountState(d)
			},
			check: func(t *testing.T, got any) {
				ag_require.Equal(t, [64]byte(repeatByte(0xB1, 64)), got.(*ConfidentialTransferFeeAmountState).WithheldAmount)
			},
		},
		{
			name:    "MetadataPointer",
			fixture: concat(repeatByte(0xC1, 32), repeatByte(0xC2, 32)),
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeMetadataPointerState(d) },
			check: func(t *testing.T, got any) {
				s := got.(*MetadataPointerState)
				ag_require.Equal(t, pubkeyOf(0xC1), s.Authority.Key)
				ag_require.Equal(t, pubkeyOf(0xC2), s.MetadataAddress.Key)
			},
		},
		{
			name:    "GroupPointer",
			fixture: concat(repeatByte(0xC3, 32), repeatByte(0xC4, 32)),
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeGroupPointerState(d) },
			check: func(t *testing.T, got any) {
				s := got.(*GroupPointerState)
				ag_require.Equal(t, pubkeyOf(0xC3), s.Authority.Key)
				ag_require.Equal(t, pubkeyOf(0xC4), s.GroupAddress.Key)
			},
		},
		{
			name: "TokenGroup",
			fixture: concat(
				repeatByte(0xD1, 32),      // update_authority
				repeatByte(0xD2, 32),      // mint
				u64LE(0x1_0000_0001),      // size > MaxUint32: proves u64 width
				u64LE(0xFFFF_FFFF_FFFF_0), // max_size
			),
			decode: func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeTokenGroup(d) },
			check: func(t *testing.T, got any) {
				s := got.(*TokenGroup)
				ag_require.Equal(t, pubkeyOf(0xD1), s.UpdateAuthority.Key)
				ag_require.Equal(t, pubkeyOf(0xD2), s.Mint)
				ag_require.Equal(t, uint64(0x1_0000_0001), s.Size)
				ag_require.Equal(t, uint64(0xFFFF_FFFF_FFFF_0), s.MaxSize)
			},
		},
		{
			name:    "GroupMemberPointer",
			fixture: concat(repeatByte(0xD3, 32), repeatByte(0xD4, 32)),
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeGroupMemberPointerState(d) },
			check: func(t *testing.T, got any) {
				s := got.(*GroupMemberPointerState)
				ag_require.Equal(t, pubkeyOf(0xD3), s.Authority.Key)
				ag_require.Equal(t, pubkeyOf(0xD4), s.MemberAddress.Key)
			},
		},
		{
			name: "TokenGroupMember",
			fixture: concat(
				repeatByte(0xD5, 32), // mint
				repeatByte(0xD6, 32), // group
				u64LE(0x2_0000_0007), // member_number > MaxUint32: proves u64 width
			),
			decode: func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeTokenGroupMember(d) },
			check: func(t *testing.T, got any) {
				s := got.(*TokenGroupMember)
				ag_require.Equal(t, pubkeyOf(0xD5), s.Mint)
				ag_require.Equal(t, pubkeyOf(0xD6), s.Group)
				ag_require.Equal(t, uint64(0x2_0000_0007), s.MemberNumber)
			},
		},
		{
			name: "ConfidentialMintBurn",
			fixture: concat(
				repeatByte(0xE1, 64), // confidential_supply
				repeatByte(0xE2, 36), // decryptable_supply
				repeatByte(0xE3, 32), // supply_elgamal_pubkey
				repeatByte(0xE4, 64), // pending_burn
			),
			decode: func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeConfidentialMintBurnState(d) },
			check: func(t *testing.T, got any) {
				s := got.(*ConfidentialMintBurnState)
				ag_require.Equal(t, [64]byte(repeatByte(0xE1, 64)), s.ConfidentialSupply)
				ag_require.Equal(t, [36]byte(repeatByte(0xE2, 36)), s.DecryptableSupply)
				ag_require.Equal(t, [32]byte(pubkeyOf(0xE3)), s.SupplyElGamalPubkey)
				ag_require.Equal(t, [64]byte(repeatByte(0xE4, 64)), s.PendingBurn)
			},
		},
		{
			name: "ScaledUiAmount",
			fixture: concat(
				repeatByte(0xF1, 32), // authority
				f64LE(1.5),           // multiplier
				i64LE(1750000000),    // new_multiplier_effective_timestamp
				f64LE(2.25),          // new_multiplier
			),
			decode: func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodeScaledUiAmountState(d) },
			check: func(t *testing.T, got any) {
				s := got.(*ScaledUiAmountState)
				ag_require.Equal(t, pubkeyOf(0xF1), s.Authority.Key)
				ag_require.Equal(t, 1.5, s.Multiplier)
				ag_require.Equal(t, int64(1750000000), s.NewMultiplierEffectiveTimestamp)
				ag_require.Equal(t, 2.25, s.NewMultiplier)
			},
		},
		{
			name:    "Pausable",
			fixture: concat(repeatByte(0xF2, 32), []byte{1}),
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodePausableState(d) },
			check: func(t *testing.T, got any) {
				s := got.(*PausableState)
				ag_require.Equal(t, pubkeyOf(0xF2), s.Authority.Key)
				ag_require.True(t, s.Paused)
			},
		},
		{
			name:    "PermissionedBurn",
			fixture: repeatByte(0xF3, 32),
			decode:  func(d []byte) (ag_binary.BinaryMarshaler, error) { return DecodePermissionedBurnState(d) },
			check: func(t *testing.T, got any) {
				ag_require.Equal(t, pubkeyOf(0xF3), got.(*PermissionedBurnState).Authority.Key)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The fixture length must agree with the extension type table.
			var typeOf ExtensionType
			found := false
			for et, info := range extensionInfos {
				if info.name == tc.name {
					typeOf = et
					found = true
				}
			}
			ag_require.True(t, found, "case name %q not present in extensionInfos", tc.name)
			n, sized := typeOf.TypeLen()
			ag_require.True(t, sized)
			ag_require.Equal(t, n, len(tc.fixture), "extensionInfos length disagrees with fixture")

			got, err := tc.decode(tc.fixture)
			ag_require.NoError(t, err)
			tc.check(t, got)

			// Byte-exact re-encode.
			ag_require.Equal(t, tc.fixture, marshalStateBytes(t, got))

			// Wrong lengths are rejected.
			_, err = tc.decode(tc.fixture[:len(tc.fixture)-1])
			ag_require.ErrorIs(t, err, ErrInvalidExtensionLength)
			_, err = tc.decode(concat(tc.fixture, []byte{0}))
			ag_require.ErrorIs(t, err, ErrInvalidExtensionLength)
		})
	}
}

func TestOptionalPubkey_NoneIsAllZero(t *testing.T) {
	s, err := DecodeMintCloseAuthorityState(repeatByte(0, 32))
	ag_require.NoError(t, err)
	ag_require.True(t, s.CloseAuthority.IsNone())
	ag_require.Nil(t, s.CloseAuthority.Get())
	ag_require.Equal(t, repeatByte(0, 32), marshalStateBytes(t, s))
}

func TestPodBool_NonzeroIsTrue(t *testing.T) {
	// The on-chain PodBool treats any nonzero byte as true; byte 2 must not
	// be rejected. Re-encoding normalizes to 1.
	s, err := DecodeCpiGuardState([]byte{2})
	ag_require.NoError(t, err)
	ag_require.True(t, s.LockCpi)
	ag_require.Equal(t, []byte{1}, marshalStateBytes(t, s))
}

func TestDecodeTokenMetadataState_TrailingBytes(t *testing.T) {
	authority := pubkeyOf(0x0F)
	meta := TokenMetadataState{
		UpdateAuthority: NewOptionalPubkey(&authority),
		Mint:            pubkeyOf(10),
		Name:            "N",
		Symbol:          "S",
		Uri:             "U",
	}
	payload := marshalStateBytes(t, meta)

	got, err := DecodeTokenMetadataState(payload)
	ag_require.NoError(t, err)
	ag_require.Equal(t, meta.Name, got.Name)

	_, err = DecodeTokenMetadataState(concat(payload, []byte{0}))
	ag_require.ErrorIs(t, err, ErrInvalidExtensionLength)
}

func TestAccountState_String(t *testing.T) {
	ag_require.Equal(t, "Uninitialized", AccountStateUninitialized.String())
	ag_require.Equal(t, "Initialized", AccountStateInitialized.String())
	ag_require.Equal(t, "Frozen", AccountStateFrozen.String())
	ag_require.Equal(t, "AccountState(5)", AccountState(5).String())
}

func TestExtensionType_String(t *testing.T) {
	ag_require.Equal(t, "TransferFeeConfig", ExtensionTransferFeeConfig.String())
	ag_require.Equal(t, "TokenMetadata", ExtensionTokenMetadata.String())
	ag_require.Equal(t, "Uninitialized", ExtensionUninitialized.String())
	ag_require.Equal(t, "ExtensionType(999)", ExtensionType(999).String())
}

func TestExtensionType_Sides(t *testing.T) {
	mintSide := []ExtensionType{
		ExtensionTransferFeeConfig, ExtensionMintCloseAuthority, ExtensionConfidentialTransferMint,
		ExtensionDefaultAccountState, ExtensionNonTransferable, ExtensionInterestBearingConfig,
		ExtensionPermanentDelegate, ExtensionTransferHook, ExtensionConfidentialTransferFeeConfig,
		ExtensionMetadataPointer, ExtensionTokenMetadata, ExtensionGroupPointer, ExtensionTokenGroup,
		ExtensionGroupMemberPointer, ExtensionTokenGroupMember, ExtensionConfidentialMintBurn,
		ExtensionScaledUiAmount, ExtensionPausable, ExtensionPermissionedBurn,
	}
	accountSide := []ExtensionType{
		ExtensionTransferFeeAmount, ExtensionConfidentialTransferAccount, ExtensionImmutableOwner,
		ExtensionMemoTransfer, ExtensionCpiGuard, ExtensionNonTransferableAccount,
		ExtensionTransferHookAccount, ExtensionConfidentialTransferFeeAmount, ExtensionPausableAccount,
	}
	for _, e := range mintSide {
		ag_require.True(t, e.IsMintExtension(), e.String())
		ag_require.False(t, e.IsAccountExtension(), e.String())
	}
	for _, e := range accountSide {
		ag_require.True(t, e.IsAccountExtension(), e.String())
		ag_require.False(t, e.IsMintExtension(), e.String())
	}
	ag_require.False(t, ExtensionUninitialized.IsMintExtension())
	ag_require.False(t, ExtensionUninitialized.IsAccountExtension())
	ag_require.False(t, ExtensionType(999).IsMintExtension())
	ag_require.False(t, ExtensionType(999).IsAccountExtension())
}

func TestCalculateLen(t *testing.T) {
	// No extensions: base sizes.
	n, err := CalculateMintLen(nil)
	ag_require.NoError(t, err)
	ag_require.Equal(t, MINT_SIZE, n)
	n, err = CalculateTokenAccountLen(nil)
	ag_require.NoError(t, err)
	ag_require.Equal(t, ACCOUNT_SIZE, n)

	// Worked example from the Rust interface: TransferFeeConfig +
	// MetadataPointer = 166 + (4+108) + (4+64) = 346.
	n, err = CalculateMintLen([]ExtensionType{ExtensionTransferFeeConfig, ExtensionMetadataPointer})
	ag_require.NoError(t, err)
	ag_require.Equal(t, 346, n)

	// Duplicates count once.
	n, err = CalculateMintLen([]ExtensionType{
		ExtensionTransferFeeConfig, ExtensionTransferFeeConfig, ExtensionMetadataPointer,
	})
	ag_require.NoError(t, err)
	ag_require.Equal(t, 346, n)

	// A total of exactly 355 (multisig length) is bumped to 357:
	// 166 + (4+80) + (4+64) + (4+33) = 355 -> 357.
	n, err = CalculateMintLen([]ExtensionType{ExtensionTokenGroup, ExtensionMetadataPointer, ExtensionPausable})
	ag_require.NoError(t, err)
	ag_require.Equal(t, 357, n)

	// Zero-length markers still cost the 4-byte header.
	n, err = CalculateTokenAccountLen([]ExtensionType{ExtensionImmutableOwner})
	ag_require.NoError(t, err)
	ag_require.Equal(t, 170, n)

	// Variable-length and unknown types cannot be sized.
	_, err = CalculateMintLen([]ExtensionType{ExtensionTokenMetadata})
	ag_require.ErrorIs(t, err, ErrUnsizedExtension)
	_, err = CalculateMintLen([]ExtensionType{ExtensionType(999)})
	ag_require.ErrorIs(t, err, ErrUnsizedExtension)
}

func TestParseExtensions_TerminatorAndTruncation(t *testing.T) {
	entry := concat(u16LE(uint16(ExtensionMintCloseAuthority)), u16LE(32), repeatByte(0x11, 32))

	// A zeroed tail is the normal terminator: nothing after type 0 is parsed.
	tlvs, err := ParseExtensions(concat(entry, repeatByte(0, 40)))
	ag_require.NoError(t, err)
	ag_require.Len(t, tlvs, 1)
	ag_require.Equal(t, ExtensionMintCloseAuthority, tlvs[0].Type)

	// A single trailing byte is legal.
	tlvs, err = ParseExtensions(concat(entry, []byte{7}))
	ag_require.NoError(t, err)
	ag_require.Len(t, tlvs, 1)

	// A nonzero type with a truncated length field is an error.
	_, err = ParseExtensions(concat(entry, u16LE(uint16(ExtensionCpiGuard)), []byte{1}))
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)

	// A value overrunning the buffer is an error; parsed entries are kept.
	tlvs, err = ParseExtensions(concat(entry, u16LE(uint16(ExtensionMintCloseAuthority)), u16LE(32), repeatByte(0x22, 10)))
	ag_require.ErrorIs(t, err, ErrInvalidAccountData)
	ag_require.Len(t, tlvs, 1)

	// Empty data parses to no extensions.
	tlvs, err = ParseExtensions(nil)
	ag_require.NoError(t, err)
	ag_require.Empty(t, tlvs)
}

func TestGetExtensionData(t *testing.T) {
	tlvs := []ExtensionTLV{
		{Type: ExtensionCpiGuard, Length: 1, Data: []byte{1}},
		{Type: ExtensionMemoTransfer, Length: 1, Data: []byte{0}},
	}
	d, ok := GetExtensionData(tlvs, ExtensionMemoTransfer)
	ag_require.True(t, ok)
	ag_require.Equal(t, []byte{0}, d)
	_, ok = GetExtensionData(tlvs, ExtensionTransferFeeAmount)
	ag_require.False(t, ok)
}

func TestMintDecode_PopulatesReceiver(t *testing.T) {
	// Regression: Mint.Decode used to decode into a discarded local copy.
	var mint Mint
	err := mint.Decode(testMintSlice)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(42), mint.Supply)
	ag_require.Equal(t, uint8(7), mint.Decimals)
	ag_require.NotNil(t, mint.MintAuthority)
	ag_require.Equal(t, pubkeyOf(1), *mint.MintAuthority)
}

func TestExtensionInfoTableConsistency(t *testing.T) {
	// Every declared extension type is present in the table exactly as
	// declared, and no side flag overlaps.
	for et := ExtensionUninitialized; et <= ExtensionPermissionedBurn; et++ {
		info, ok := extensionInfos[et]
		ag_require.True(t, ok, "missing table entry for %d", uint16(et))
		if et == ExtensionUninitialized {
			continue
		}
		ag_require.False(t, info.mint && info.account, "%s is flagged for both sides", et)
		ag_require.True(t, info.mint || info.account, "%s has no side", et)
		if info.variable {
			ag_require.Equal(t, 0, info.length)
		}
	}
	ag_require.Len(t, extensionInfos, 29)
}
