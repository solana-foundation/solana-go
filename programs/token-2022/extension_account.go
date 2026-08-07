package token2022

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"math"

	bin "github.com/gagliardetto/binary"
)

// MintWithExtensions is a token-2022 mint together with its typed extensions.
// Extension pointer fields are nil when the extension is absent; boolean
// fields represent zero-length marker extensions. Unrecognized extension
// types are preserved verbatim in Unknown so that data written by newer
// program versions survives a decode/encode round-trip.
type MintWithExtensions struct {
	Mint Mint

	TransferFeeConfig             *TransferFeeConfigState
	MintCloseAuthority            *MintCloseAuthorityState
	ConfidentialTransferMint      *ConfidentialTransferMintState
	DefaultAccountState           *DefaultAccountStateConfig
	NonTransferable               bool
	InterestBearingConfig         *InterestBearingConfigState
	PermanentDelegate             *PermanentDelegateState
	TransferHook                  *TransferHookState
	ConfidentialTransferFeeConfig *ConfidentialTransferFeeConfigState
	MetadataPointer               *MetadataPointerState
	TokenMetadata                 *TokenMetadataState
	GroupPointer                  *GroupPointerState
	TokenGroup                    *TokenGroup
	GroupMemberPointer            *GroupMemberPointerState
	TokenGroupMember              *TokenGroupMember
	ConfidentialMintBurn          *ConfidentialMintBurnState
	ScaledUiAmount                *ScaledUiAmountState
	Pausable                      *PausableState
	PermissionedBurn              *PermissionedBurnState

	// Unknown holds TLV entries with unrecognized extension types, in their
	// original order.
	Unknown []ExtensionTLV

	// tlvOrder is the original TLV entry order captured at decode time, used
	// to re-encode byte-exactly.
	tlvOrder []ExtensionType
}

// AccountWithExtensions is a token-2022 token account together with its typed
// extensions. See MintWithExtensions for the field conventions.
type AccountWithExtensions struct {
	Account Account

	TransferFeeAmount             *TransferFeeAmountState
	ConfidentialTransferAccount   *ConfidentialTransferAccountState
	ImmutableOwner                bool
	MemoTransfer                  *MemoTransferState
	CpiGuard                      *CpiGuardState
	NonTransferableAccount        bool
	TransferHookAccount           *TransferHookAccountState
	ConfidentialTransferFeeAmount *ConfidentialTransferFeeAmountState
	PausableAccount               bool

	// Unknown holds TLV entries with unrecognized extension types, in their
	// original order.
	Unknown []ExtensionTLV

	// tlvOrder is the original TLV entry order captured at decode time, used
	// to re-encode byte-exactly.
	tlvOrder []ExtensionType
}

var (
	_ encoding.BinaryMarshaler   = (*MintWithExtensions)(nil)
	_ encoding.BinaryUnmarshaler = (*MintWithExtensions)(nil)
	_ encoding.BinaryMarshaler   = (*AccountWithExtensions)(nil)
	_ encoding.BinaryUnmarshaler = (*AccountWithExtensions)(nil)
)

// DecodeMintWithExtensions decodes a token-2022 mint account, including all
// extensions, from raw account data.
func DecodeMintWithExtensions(data []byte) (*MintWithExtensions, error) {
	m := new(MintWithExtensions)
	if err := m.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return m, nil
}

// DecodeAccountWithExtensions decodes a token-2022 token account, including
// all extensions, from raw account data.
func DecodeAccountWithExtensions(data []byte) (*AccountWithExtensions, error) {
	a := new(AccountWithExtensions)
	if err := a.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return a, nil
}

// splitExtensionRegion validates the layout of extended account data and
// returns the TLV region (nil when the data is exactly the base state).
//
// Rules, mirroring the token-2022 program: data shorter than the base state
// is invalid; data of exactly the base length carries no extensions; data of
// exactly 355 bytes is never valid (that length is reserved for multisig
// accounts); extended data must reach past offset 166, have zero padding
// between the base state and offset 165 (relevant for mints), and carry the
// expected account type byte at offset 165.
func splitExtensionRegion(data []byte, baseSize int, wantAccountType uint8) ([]byte, error) {
	if len(data) < baseSize {
		return nil, fmt.Errorf("%w: %d bytes is shorter than the base state (%d)", ErrInvalidAccountData, len(data), baseSize)
	}
	if len(data) == baseSize {
		return nil, nil
	}
	if len(data) == multisigSize {
		return nil, fmt.Errorf("%w: %d bytes is the multisig length, not valid extended state", ErrInvalidAccountData, len(data))
	}
	if len(data) < tlvStartOffset {
		return nil, fmt.Errorf("%w: %d bytes is too short for extended state (need at least %d)", ErrInvalidAccountData, len(data), tlvStartOffset)
	}
	for i := baseSize; i < accountTypeOffset; i++ {
		if data[i] != 0 {
			return nil, fmt.Errorf("%w: nonzero padding byte at offset %d", ErrInvalidAccountData, i)
		}
	}
	if data[accountTypeOffset] != wantAccountType {
		return nil, fmt.Errorf("%w: account type byte is %d, want %d", ErrInvalidAccountData, data[accountTypeOffset], wantAccountType)
	}
	return data[tlvStartOffset:], nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler. It decodes full mint
// account data, validating the extension layout strictly (see
// splitExtensionRegion) and rejecting uninitialized mints, account-side
// extensions, duplicate extensions, and length mismatches.
func (m *MintWithExtensions) UnmarshalBinary(data []byte) error {
	*m = MintWithExtensions{}
	tlvRegion, err := splitExtensionRegion(data, MINT_SIZE, AccountTypeMint)
	if err != nil {
		return err
	}
	if err := m.Mint.UnmarshalWithDecoder(bin.NewBinDecoder(data)); err != nil {
		return fmt.Errorf("unable to decode mint: %w", err)
	}
	if !m.Mint.IsInitialized {
		return fmt.Errorf("%w: mint is_initialized is false", ErrUninitializedAccount)
	}
	if tlvRegion == nil {
		return nil
	}
	tlvs, err := ParseExtensions(tlvRegion)
	if err != nil {
		return err
	}
	for _, tlv := range tlvs {
		if err := m.applyExtension(tlv); err != nil {
			return err
		}
		m.tlvOrder = append(m.tlvOrder, tlv.Type)
	}
	return nil
}

func (m *MintWithExtensions) applyExtension(tlv ExtensionTLV) error {
	info, known := extensionInfos[tlv.Type]
	if !known {
		m.Unknown = append(m.Unknown, tlv)
		return nil
	}
	if !info.mint {
		return fmt.Errorf("%w: %s is not a mint extension", ErrExtensionTypeMismatch, tlv.Type)
	}
	var err error
	switch tlv.Type {
	case ExtensionTransferFeeConfig:
		err = decodeUniqueExtension(tlv, &m.TransferFeeConfig, DecodeTransferFeeConfigState)
	case ExtensionMintCloseAuthority:
		err = decodeUniqueExtension(tlv, &m.MintCloseAuthority, DecodeMintCloseAuthorityState)
	case ExtensionConfidentialTransferMint:
		err = decodeUniqueExtension(tlv, &m.ConfidentialTransferMint, DecodeConfidentialTransferMintState)
	case ExtensionDefaultAccountState:
		err = decodeUniqueExtension(tlv, &m.DefaultAccountState, DecodeDefaultAccountStateConfig)
	case ExtensionNonTransferable:
		err = applyMarkerExtension(tlv, &m.NonTransferable)
	case ExtensionInterestBearingConfig:
		err = decodeUniqueExtension(tlv, &m.InterestBearingConfig, DecodeInterestBearingConfigState)
	case ExtensionPermanentDelegate:
		err = decodeUniqueExtension(tlv, &m.PermanentDelegate, DecodePermanentDelegateState)
	case ExtensionTransferHook:
		err = decodeUniqueExtension(tlv, &m.TransferHook, DecodeTransferHookState)
	case ExtensionConfidentialTransferFeeConfig:
		err = decodeUniqueExtension(tlv, &m.ConfidentialTransferFeeConfig, DecodeConfidentialTransferFeeConfigState)
	case ExtensionMetadataPointer:
		err = decodeUniqueExtension(tlv, &m.MetadataPointer, DecodeMetadataPointerState)
	case ExtensionTokenMetadata:
		err = decodeUniqueExtension(tlv, &m.TokenMetadata, DecodeTokenMetadataState)
	case ExtensionGroupPointer:
		err = decodeUniqueExtension(tlv, &m.GroupPointer, DecodeGroupPointerState)
	case ExtensionTokenGroup:
		err = decodeUniqueExtension(tlv, &m.TokenGroup, DecodeTokenGroup)
	case ExtensionGroupMemberPointer:
		err = decodeUniqueExtension(tlv, &m.GroupMemberPointer, DecodeGroupMemberPointerState)
	case ExtensionTokenGroupMember:
		err = decodeUniqueExtension(tlv, &m.TokenGroupMember, DecodeTokenGroupMember)
	case ExtensionConfidentialMintBurn:
		err = decodeUniqueExtension(tlv, &m.ConfidentialMintBurn, DecodeConfidentialMintBurnState)
	case ExtensionScaledUiAmount:
		err = decodeUniqueExtension(tlv, &m.ScaledUiAmount, DecodeScaledUiAmountState)
	case ExtensionPausable:
		err = decodeUniqueExtension(tlv, &m.Pausable, DecodePausableState)
	case ExtensionPermissionedBurn:
		err = decodeUniqueExtension(tlv, &m.PermissionedBurn, DecodePermissionedBurnState)
	default:
		// Known but not a mint extension: unreachable (filtered above).
		err = fmt.Errorf("%w: unsupported mint extension %s", ErrInvalidAccountData, tlv.Type)
	}
	return err
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler. It decodes full
// token account data with the same strictness as its mint counterpart
// (accounts whose state is neither Initialized nor Frozen are rejected).
func (a *AccountWithExtensions) UnmarshalBinary(data []byte) error {
	*a = AccountWithExtensions{}
	tlvRegion, err := splitExtensionRegion(data, ACCOUNT_SIZE, AccountTypeAccount)
	if err != nil {
		return err
	}
	if err := a.Account.UnmarshalWithDecoder(bin.NewBinDecoder(data)); err != nil {
		return fmt.Errorf("unable to decode account: %w", err)
	}
	// Mirrors PodAccount::is_initialized: only Initialized and Frozen count.
	if a.Account.State != AccountStateInitialized && a.Account.State != AccountStateFrozen {
		return fmt.Errorf("%w: account state is %d", ErrUninitializedAccount, a.Account.State)
	}
	if tlvRegion == nil {
		return nil
	}
	tlvs, err := ParseExtensions(tlvRegion)
	if err != nil {
		return err
	}
	for _, tlv := range tlvs {
		if err := a.applyExtension(tlv); err != nil {
			return err
		}
		a.tlvOrder = append(a.tlvOrder, tlv.Type)
	}
	return nil
}

func (a *AccountWithExtensions) applyExtension(tlv ExtensionTLV) error {
	info, known := extensionInfos[tlv.Type]
	if !known {
		a.Unknown = append(a.Unknown, tlv)
		return nil
	}
	if !info.account {
		return fmt.Errorf("%w: %s is not a token-account extension", ErrExtensionTypeMismatch, tlv.Type)
	}
	var err error
	switch tlv.Type {
	case ExtensionTransferFeeAmount:
		err = decodeUniqueExtension(tlv, &a.TransferFeeAmount, DecodeTransferFeeAmountState)
	case ExtensionConfidentialTransferAccount:
		err = decodeUniqueExtension(tlv, &a.ConfidentialTransferAccount, DecodeConfidentialTransferAccountState)
	case ExtensionImmutableOwner:
		err = applyMarkerExtension(tlv, &a.ImmutableOwner)
	case ExtensionMemoTransfer:
		err = decodeUniqueExtension(tlv, &a.MemoTransfer, DecodeMemoTransferState)
	case ExtensionCpiGuard:
		err = decodeUniqueExtension(tlv, &a.CpiGuard, DecodeCpiGuardState)
	case ExtensionNonTransferableAccount:
		err = applyMarkerExtension(tlv, &a.NonTransferableAccount)
	case ExtensionTransferHookAccount:
		err = decodeUniqueExtension(tlv, &a.TransferHookAccount, DecodeTransferHookAccountState)
	case ExtensionConfidentialTransferFeeAmount:
		err = decodeUniqueExtension(tlv, &a.ConfidentialTransferFeeAmount, DecodeConfidentialTransferFeeAmountState)
	case ExtensionPausableAccount:
		err = applyMarkerExtension(tlv, &a.PausableAccount)
	default:
		err = fmt.Errorf("%w: unsupported token-account extension %s", ErrInvalidAccountData, tlv.Type)
	}
	return err
}

// decodeUniqueExtension decodes tlv.Data into *dst, rejecting duplicates.
func decodeUniqueExtension[T any](tlv ExtensionTLV, dst **T, decode func([]byte) (*T, error)) error {
	if *dst != nil {
		return fmt.Errorf("%w: %s", ErrDuplicateExtension, tlv.Type)
	}
	v, err := decode(tlv.Data)
	if err != nil {
		return err
	}
	*dst = v
	return nil
}

// applyMarkerExtension records a zero-length marker extension, rejecting
// duplicates and nonzero lengths.
func applyMarkerExtension(tlv ExtensionTLV, dst *bool) error {
	if *dst {
		return fmt.Errorf("%w: %s", ErrDuplicateExtension, tlv.Type)
	}
	if tlv.Length != 0 {
		return fmt.Errorf("%w: marker extension %s has length %d, want 0", ErrInvalidExtensionLength, tlv.Type, tlv.Length)
	}
	*dst = true
	return nil
}

// HasExtensions reports whether any extension (including unknown ones) is set.
func (m *MintWithExtensions) HasExtensions() bool {
	return m.TransferFeeConfig != nil ||
		m.MintCloseAuthority != nil ||
		m.ConfidentialTransferMint != nil ||
		m.DefaultAccountState != nil ||
		m.NonTransferable ||
		m.InterestBearingConfig != nil ||
		m.PermanentDelegate != nil ||
		m.TransferHook != nil ||
		m.ConfidentialTransferFeeConfig != nil ||
		m.MetadataPointer != nil ||
		m.TokenMetadata != nil ||
		m.GroupPointer != nil ||
		m.TokenGroup != nil ||
		m.GroupMemberPointer != nil ||
		m.TokenGroupMember != nil ||
		m.ConfidentialMintBurn != nil ||
		m.ScaledUiAmount != nil ||
		m.Pausable != nil ||
		m.PermissionedBurn != nil ||
		len(m.Unknown) > 0
}

// HasExtensions reports whether any extension (including unknown ones) is set.
func (a *AccountWithExtensions) HasExtensions() bool {
	return a.TransferFeeAmount != nil ||
		a.ConfidentialTransferAccount != nil ||
		a.ImmutableOwner ||
		a.MemoTransfer != nil ||
		a.CpiGuard != nil ||
		a.NonTransferableAccount ||
		a.TransferHookAccount != nil ||
		a.ConfidentialTransferFeeAmount != nil ||
		a.PausableAccount ||
		len(a.Unknown) > 0
}

// presentExtensionTypes returns the set typed extensions in ascending
// ExtensionType order.
func (m *MintWithExtensions) presentExtensionTypes() []ExtensionType {
	var out []ExtensionType
	add := func(t ExtensionType, set bool) {
		if set {
			out = append(out, t)
		}
	}
	add(ExtensionTransferFeeConfig, m.TransferFeeConfig != nil)
	add(ExtensionMintCloseAuthority, m.MintCloseAuthority != nil)
	add(ExtensionConfidentialTransferMint, m.ConfidentialTransferMint != nil)
	add(ExtensionDefaultAccountState, m.DefaultAccountState != nil)
	add(ExtensionNonTransferable, m.NonTransferable)
	add(ExtensionInterestBearingConfig, m.InterestBearingConfig != nil)
	add(ExtensionPermanentDelegate, m.PermanentDelegate != nil)
	add(ExtensionTransferHook, m.TransferHook != nil)
	add(ExtensionConfidentialTransferFeeConfig, m.ConfidentialTransferFeeConfig != nil)
	add(ExtensionMetadataPointer, m.MetadataPointer != nil)
	add(ExtensionTokenMetadata, m.TokenMetadata != nil)
	add(ExtensionGroupPointer, m.GroupPointer != nil)
	add(ExtensionTokenGroup, m.TokenGroup != nil)
	add(ExtensionGroupMemberPointer, m.GroupMemberPointer != nil)
	add(ExtensionTokenGroupMember, m.TokenGroupMember != nil)
	add(ExtensionConfidentialMintBurn, m.ConfidentialMintBurn != nil)
	add(ExtensionScaledUiAmount, m.ScaledUiAmount != nil)
	add(ExtensionPausable, m.Pausable != nil)
	add(ExtensionPermissionedBurn, m.PermissionedBurn != nil)
	return out
}

// presentExtensionTypes returns the set typed extensions in ascending
// ExtensionType order.
func (a *AccountWithExtensions) presentExtensionTypes() []ExtensionType {
	var out []ExtensionType
	add := func(t ExtensionType, set bool) {
		if set {
			out = append(out, t)
		}
	}
	add(ExtensionTransferFeeAmount, a.TransferFeeAmount != nil)
	add(ExtensionConfidentialTransferAccount, a.ConfidentialTransferAccount != nil)
	add(ExtensionImmutableOwner, a.ImmutableOwner)
	add(ExtensionMemoTransfer, a.MemoTransfer != nil)
	add(ExtensionCpiGuard, a.CpiGuard != nil)
	add(ExtensionNonTransferableAccount, a.NonTransferableAccount)
	add(ExtensionTransferHookAccount, a.TransferHookAccount != nil)
	add(ExtensionConfidentialTransferFeeAmount, a.ConfidentialTransferFeeAmount != nil)
	add(ExtensionPausableAccount, a.PausableAccount)
	return out
}

// extensionPayload marshals the typed extension t into its wire bytes.
// present is false when the extension is not set. Marker extensions yield an
// empty payload.
func (m *MintWithExtensions) extensionPayload(t ExtensionType) (payload []byte, present bool, err error) {
	switch t {
	case ExtensionTransferFeeConfig:
		return marshalOptionalExtension(m.TransferFeeConfig)
	case ExtensionMintCloseAuthority:
		return marshalOptionalExtension(m.MintCloseAuthority)
	case ExtensionConfidentialTransferMint:
		return marshalOptionalExtension(m.ConfidentialTransferMint)
	case ExtensionDefaultAccountState:
		return marshalOptionalExtension(m.DefaultAccountState)
	case ExtensionNonTransferable:
		return []byte{}, m.NonTransferable, nil
	case ExtensionInterestBearingConfig:
		return marshalOptionalExtension(m.InterestBearingConfig)
	case ExtensionPermanentDelegate:
		return marshalOptionalExtension(m.PermanentDelegate)
	case ExtensionTransferHook:
		return marshalOptionalExtension(m.TransferHook)
	case ExtensionConfidentialTransferFeeConfig:
		return marshalOptionalExtension(m.ConfidentialTransferFeeConfig)
	case ExtensionMetadataPointer:
		return marshalOptionalExtension(m.MetadataPointer)
	case ExtensionTokenMetadata:
		return marshalOptionalExtension(m.TokenMetadata)
	case ExtensionGroupPointer:
		return marshalOptionalExtension(m.GroupPointer)
	case ExtensionTokenGroup:
		return marshalOptionalExtension(m.TokenGroup)
	case ExtensionGroupMemberPointer:
		return marshalOptionalExtension(m.GroupMemberPointer)
	case ExtensionTokenGroupMember:
		return marshalOptionalExtension(m.TokenGroupMember)
	case ExtensionConfidentialMintBurn:
		return marshalOptionalExtension(m.ConfidentialMintBurn)
	case ExtensionScaledUiAmount:
		return marshalOptionalExtension(m.ScaledUiAmount)
	case ExtensionPausable:
		return marshalOptionalExtension(m.Pausable)
	case ExtensionPermissionedBurn:
		return marshalOptionalExtension(m.PermissionedBurn)
	default:
		return nil, false, nil
	}
}

// extensionPayload marshals the typed extension t into its wire bytes.
func (a *AccountWithExtensions) extensionPayload(t ExtensionType) (payload []byte, present bool, err error) {
	switch t {
	case ExtensionTransferFeeAmount:
		return marshalOptionalExtension(a.TransferFeeAmount)
	case ExtensionConfidentialTransferAccount:
		return marshalOptionalExtension(a.ConfidentialTransferAccount)
	case ExtensionImmutableOwner:
		return []byte{}, a.ImmutableOwner, nil
	case ExtensionMemoTransfer:
		return marshalOptionalExtension(a.MemoTransfer)
	case ExtensionCpiGuard:
		return marshalOptionalExtension(a.CpiGuard)
	case ExtensionNonTransferableAccount:
		return []byte{}, a.NonTransferableAccount, nil
	case ExtensionTransferHookAccount:
		return marshalOptionalExtension(a.TransferHookAccount)
	case ExtensionConfidentialTransferFeeAmount:
		return marshalOptionalExtension(a.ConfidentialTransferFeeAmount)
	case ExtensionPausableAccount:
		return []byte{}, a.PausableAccount, nil
	default:
		return nil, false, nil
	}
}

// marshalOptionalExtension marshals a possibly-nil extension state pointer.
func marshalOptionalExtension[T any, PT interface {
	*T
	bin.BinaryMarshaler
}](s *T) ([]byte, bool, error) {
	if s == nil {
		return nil, false, nil
	}
	buf := new(bytes.Buffer)
	if err := PT(s).MarshalWithEncoder(bin.NewBinEncoder(buf)); err != nil {
		return nil, false, err
	}
	return buf.Bytes(), true, nil
}

// MarshalBinary implements encoding.BinaryMarshaler. It produces byte-exact
// token-2022 mint account data: the 82-byte base state alone when no
// extensions are set, otherwise base state, zero padding to offset 165, the
// account type byte, and the TLV entries. Extensions decoded from existing
// data keep their original TLV order; extensions added afterwards (or set on
// a hand-built value) are appended in ascending ExtensionType order. If the
// resulting data would be exactly 355 bytes long (the multisig length), two
// zero bytes are appended, as done by the token-2022 program.
func (m *MintWithExtensions) MarshalBinary() ([]byte, error) {
	base := new(bytes.Buffer)
	if err := m.Mint.MarshalWithEncoder(bin.NewBinEncoder(base)); err != nil {
		return nil, fmt.Errorf("unable to encode mint: %w", err)
	}
	return encodeWithExtensions(base.Bytes(), MINT_SIZE, AccountTypeMint, m.tlvOrder, m.presentExtensionTypes(), m.Unknown, m.extensionPayload)
}

// MarshalBinary implements encoding.BinaryMarshaler. See the documentation
// on MintWithExtensions.MarshalBinary for the exact semantics.
func (a *AccountWithExtensions) MarshalBinary() ([]byte, error) {
	base := new(bytes.Buffer)
	if err := a.Account.MarshalWithEncoder(bin.NewBinEncoder(base)); err != nil {
		return nil, fmt.Errorf("unable to encode account: %w", err)
	}
	return encodeWithExtensions(base.Bytes(), ACCOUNT_SIZE, AccountTypeAccount, a.tlvOrder, a.presentExtensionTypes(), a.Unknown, a.extensionPayload)
}

// encodeWithExtensions assembles full account data from the base state bytes
// and the extension payload source. Entries listed in tlvOrder are written
// first (in that order, skipping ones no longer set); remaining present
// extensions follow in ascending type order; remaining unknown TLVs are
// written last in their stored order.
func encodeWithExtensions(
	base []byte,
	baseSize int,
	accountType uint8,
	tlvOrder []ExtensionType,
	present []ExtensionType,
	unknown []ExtensionTLV,
	payloadOf func(ExtensionType) ([]byte, bool, error),
) ([]byte, error) {
	if len(base) != baseSize {
		return nil, fmt.Errorf("%w: base state encoded to %d bytes, want %d", ErrInvalidAccountData, len(base), baseSize)
	}
	if len(present) == 0 && len(unknown) == 0 {
		return base, nil
	}

	out := new(bytes.Buffer)
	out.Write(base)
	out.Write(make([]byte, accountTypeOffset-baseSize))
	out.WriteByte(accountType)

	writeEntry := func(t ExtensionType, payload []byte) error {
		if len(payload) > math.MaxUint16 {
			return fmt.Errorf("%w: %s payload is %d bytes, exceeds the u16 TLV length", ErrInvalidExtensionLength, t, len(payload))
		}
		var header [4]byte
		binary.LittleEndian.PutUint16(header[0:], uint16(t))
		binary.LittleEndian.PutUint16(header[2:], uint16(len(payload)))
		out.Write(header[:])
		out.Write(payload)
		return nil
	}

	written := make(map[ExtensionType]struct{}, len(present))
	unknownIdx := 0

	// First pass: original TLV order. Unknown entries are consumed from the
	// Unknown slice in order; each entry is written with its own type and
	// data so the pair stays consistent even if the caller edited Unknown
	// after decoding.
	for _, t := range tlvOrder {
		if _, known := extensionInfos[t]; !known {
			if unknownIdx < len(unknown) {
				if err := writeEntry(unknown[unknownIdx].Type, unknown[unknownIdx].Data); err != nil {
					return nil, err
				}
				unknownIdx++
			}
			continue
		}
		if _, done := written[t]; done {
			continue
		}
		payload, ok, err := payloadOf(t)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue // extension was cleared after decoding
		}
		if err := writeEntry(t, payload); err != nil {
			return nil, err
		}
		written[t] = struct{}{}
	}

	// Second pass: newly set extensions, ascending type order.
	for _, t := range present {
		if _, done := written[t]; done {
			continue
		}
		payload, ok, err := payloadOf(t)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if err := writeEntry(t, payload); err != nil {
			return nil, err
		}
		written[t] = struct{}{}
	}

	// Remaining unknown TLVs not covered by tlvOrder.
	for ; unknownIdx < len(unknown); unknownIdx++ {
		if err := writeEntry(unknown[unknownIdx].Type, unknown[unknownIdx].Data); err != nil {
			return nil, err
		}
	}

	if out.Len() == multisigSize {
		out.Write([]byte{0, 0})
	}
	return out.Bytes(), nil
}
