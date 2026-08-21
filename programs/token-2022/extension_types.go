package token2022

import (
	"errors"
	"fmt"
)

// Sentinel errors for token-2022 extension state decoding and encoding.
// They are wrapped with additional context; test with errors.Is.
var (
	// ErrInvalidAccountData indicates account data that does not match the
	// token-2022 layout (bad length, nonzero padding, bad account type byte,
	// or malformed TLV data).
	ErrInvalidAccountData = errors.New("token2022: invalid account data")
	// ErrExtensionNotFound indicates the requested extension is not present.
	ErrExtensionNotFound = errors.New("token2022: extension not found")
	// ErrExtensionTypeMismatch indicates an extension found on the wrong kind
	// of account (a mint extension in token-account data or vice versa).
	ErrExtensionTypeMismatch = errors.New("token2022: extension type mismatch")
	// ErrInvalidExtensionLength indicates a TLV entry whose length does not
	// match the fixed size of its extension type.
	ErrInvalidExtensionLength = errors.New("token2022: invalid extension length")
	// ErrDuplicateExtension indicates the same extension type appears more
	// than once in the TLV data.
	ErrDuplicateExtension = errors.New("token2022: duplicate extension")
	// ErrUnsizedExtension indicates an extension type whose data length is
	// not fixed (TokenMetadata) or not known (unrecognized types), so its
	// size cannot be calculated from the type alone.
	ErrUnsizedExtension = errors.New("token2022: extension type has no fixed size")
	// ErrUninitializedAccount indicates mint or token-account data whose
	// base state is not initialized (mirrors the unpack checks of the
	// token-2022 program).
	ErrUninitializedAccount = errors.New("token2022: account not initialized")
	// ErrCalculationOverflow indicates that a fee or rate calculation
	// overflowed (a case where the Rust implementation returns None).
	ErrCalculationOverflow = errors.New("token2022: calculation overflow")
	// ErrInvalidUiAmount indicates a UI amount string that cannot be
	// converted to a raw token amount.
	ErrInvalidUiAmount = errors.New("token2022: invalid ui amount")
)

// Absolute offsets shared by extended mints and token accounts: the account
// type byte sits at offset 165 (mints are zero-padded from 82 up to 165), and
// TLV entries start at offset 166.
const (
	accountTypeOffset = ACCOUNT_SIZE     // 165
	tlvStartOffset    = ACCOUNT_SIZE + 1 // 166
	// multisigSize is the length of a Multisig account. Extended mint or
	// token-account data must never be exactly this long, so encoded data
	// landing on it is padded by two zero bytes.
	multisigSize = 355
)

// extensionInfo describes the wire format of a known extension type.
type extensionInfo struct {
	name     string
	length   int  // fixed data length in bytes; 0 for markers and variable-length types
	variable bool // true if the data length is instance-dependent (TokenMetadata)
	mint     bool // valid on mint accounts
	account  bool // valid on token accounts
}

var extensionInfos = map[ExtensionType]extensionInfo{
	ExtensionUninitialized:                 {name: "Uninitialized"},
	ExtensionTransferFeeConfig:             {name: "TransferFeeConfig", length: 108, mint: true},
	ExtensionTransferFeeAmount:             {name: "TransferFeeAmount", length: 8, account: true},
	ExtensionMintCloseAuthority:            {name: "MintCloseAuthority", length: 32, mint: true},
	ExtensionConfidentialTransferMint:      {name: "ConfidentialTransferMint", length: 65, mint: true},
	ExtensionConfidentialTransferAccount:   {name: "ConfidentialTransferAccount", length: 295, account: true},
	ExtensionDefaultAccountState:           {name: "DefaultAccountState", length: 1, mint: true},
	ExtensionImmutableOwner:                {name: "ImmutableOwner", length: 0, account: true},
	ExtensionMemoTransfer:                  {name: "MemoTransfer", length: 1, account: true},
	ExtensionNonTransferable:               {name: "NonTransferable", length: 0, mint: true},
	ExtensionInterestBearingConfig:         {name: "InterestBearingConfig", length: 52, mint: true},
	ExtensionCpiGuard:                      {name: "CpiGuard", length: 1, account: true},
	ExtensionPermanentDelegate:             {name: "PermanentDelegate", length: 32, mint: true},
	ExtensionNonTransferableAccount:        {name: "NonTransferableAccount", length: 0, account: true},
	ExtensionTransferHook:                  {name: "TransferHook", length: 64, mint: true},
	ExtensionTransferHookAccount:           {name: "TransferHookAccount", length: 1, account: true},
	ExtensionConfidentialTransferFeeConfig: {name: "ConfidentialTransferFeeConfig", length: 129, mint: true},
	ExtensionConfidentialTransferFeeAmount: {name: "ConfidentialTransferFeeAmount", length: 64, account: true},
	ExtensionMetadataPointer:               {name: "MetadataPointer", length: 64, mint: true},
	ExtensionTokenMetadata:                 {name: "TokenMetadata", variable: true, mint: true},
	ExtensionGroupPointer:                  {name: "GroupPointer", length: 64, mint: true},
	ExtensionTokenGroup:                    {name: "TokenGroup", length: 80, mint: true},
	ExtensionGroupMemberPointer:            {name: "GroupMemberPointer", length: 64, mint: true},
	ExtensionTokenGroupMember:              {name: "TokenGroupMember", length: 72, mint: true},
	ExtensionConfidentialMintBurn:          {name: "ConfidentialMintBurn", length: 196, mint: true},
	ExtensionScaledUiAmount:                {name: "ScaledUiAmount", length: 56, mint: true},
	ExtensionPausable:                      {name: "Pausable", length: 33, mint: true},
	ExtensionPausableAccount:               {name: "PausableAccount", length: 0, account: true},
	ExtensionPermissionedBurn:              {name: "PermissionedBurn", length: 32, mint: true},
}

// String returns the account state name as used by the token-2022 program,
// or "AccountState(N)" for unknown values.
func (s AccountState) String() string {
	switch s {
	case AccountStateUninitialized:
		return "Uninitialized"
	case AccountStateInitialized:
		return "Initialized"
	case AccountStateFrozen:
		return "Frozen"
	default:
		return fmt.Sprintf("AccountState(%d)", uint8(s))
	}
}

// String returns the extension name as used by the token-2022 program, or
// "ExtensionType(N)" for unknown values.
func (t ExtensionType) String() string {
	if info, ok := extensionInfos[t]; ok {
		return info.name
	}
	return fmt.Sprintf("ExtensionType(%d)", uint16(t))
}

// TypeLen returns the fixed data length of the extension type in bytes.
// sized is false for variable-length extensions (TokenMetadata) and for
// unknown types; zero-length marker extensions return (0, true).
func (t ExtensionType) TypeLen() (n int, sized bool) {
	info, ok := extensionInfos[t]
	if !ok || info.variable {
		return 0, false
	}
	return info.length, true
}

// IsMintExtension reports whether the extension type applies to mint accounts.
func (t ExtensionType) IsMintExtension() bool {
	return extensionInfos[t].mint
}

// IsAccountExtension reports whether the extension type applies to token accounts.
func (t ExtensionType) IsAccountExtension() bool {
	return extensionInfos[t].account
}

// CalculateMintLen returns the byte length of a mint account with the given
// extension types, mirroring ExtensionType::try_calculate_account_len in the
// token-2022 program: duplicates are counted once, no extensions yields
// MINT_SIZE, and a total of exactly 355 bytes (the multisig length) is bumped
// to 357. Variable-length and unknown extension types yield an error.
func CalculateMintLen(types []ExtensionType) (int, error) {
	return calculateAccountLen(types, MINT_SIZE)
}

// CalculateTokenAccountLen returns the byte length of a token account with
// the given extension types. See CalculateMintLen for the exact semantics.
func CalculateTokenAccountLen(types []ExtensionType) (int, error) {
	return calculateAccountLen(types, ACCOUNT_SIZE)
}

func calculateAccountLen(types []ExtensionType, baseSize int) (int, error) {
	if len(types) == 0 {
		return baseSize, nil
	}
	total := tlvStartOffset
	seen := make(map[ExtensionType]struct{}, len(types))
	for _, t := range types {
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		n, sized := t.TypeLen()
		if !sized {
			return 0, fmt.Errorf("%w: %s", ErrUnsizedExtension, t)
		}
		total += 4 + n
	}
	if total == multisigSize {
		total += 2
	}
	return total, nil
}
