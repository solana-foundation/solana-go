package token2022

import (
	"encoding/binary"
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

// TLV entry header for extensions in Token-2022 account data.
// Each extension is preceded by: [u16 extension_type] [u16 length] [data...]
const (
	// AccountTypeUninitialized is used before the account type is determined.
	AccountTypeUninitialized uint8 = 0
	// AccountTypeMint marks the account as a Mint.
	AccountTypeMint uint8 = 1
	// AccountTypeAccount marks the account as a Token Account.
	AccountTypeAccount uint8 = 2
)

// Base sizes of Token-2022 account types (without extensions).
// MINT_SIZE is also defined in rpc.go as 82.
const ACCOUNT_SIZE = 165

// Limits for decoding untrusted metadata to prevent OOM from malformed data.
const (
	maxMetadataStringLen = 10 * 1024 * 1024 // 10 MB per string field
	maxMetadataFields    = 1024             // max additional metadata key-value pairs
)

// OptionalPubkey represents an optional public key stored as 32 bytes where
// all zeros means None (used in extension state Pod types).
type OptionalPubkey struct {
	Key solana.PublicKey
}

func (o OptionalPubkey) IsNone() bool {
	return o.Key == solana.PublicKey{}
}

func (o OptionalPubkey) Get() *solana.PublicKey {
	if o.IsNone() {
		return nil
	}
	return &o.Key
}

func NewOptionalPubkey(key *solana.PublicKey) OptionalPubkey {
	if key == nil {
		return OptionalPubkey{}
	}
	return OptionalPubkey{Key: *key}
}

// --- TransferFee types ---

// TransferFee represents a transfer fee configuration epoch entry.
type TransferFee struct {
	// First epoch where the fee takes effect.
	Epoch uint64
	// Maximum fee assessed on transfers, in token amount.
	MaximumFee uint64
	// Amount of transfer collected as fees, expressed as basis points (1/100 of a percent).
	TransferFeeBasisPoints uint16
}

// TransferFeeConfigState is the extension state for ExtensionTransferFeeConfig (mint extension).
type TransferFeeConfigState struct {
	// Authority that can set the fee.
	TransferFeeConfigAuthority OptionalPubkey
	// Authority that can withdraw withheld fees.
	WithdrawWithheldAuthority OptionalPubkey
	// Withheld amount of fees in the mint.
	WithheldAmount uint64
	// Older transfer fee, used before newer_transfer_fee epoch.
	OlderTransferFee TransferFee
	// Newer transfer fee, used after its epoch.
	NewerTransferFee TransferFee
}

// TransferFeeAmountState is the extension state for ExtensionTransferFeeAmount (account extension).
type TransferFeeAmountState struct {
	// Amount of fees withheld on this account.
	WithheldAmount uint64
}

// --- MintCloseAuthority ---

// MintCloseAuthorityState is the extension state for ExtensionMintCloseAuthority.
type MintCloseAuthorityState struct {
	CloseAuthority OptionalPubkey
}

// --- ConfidentialTransferMint ---

// ConfidentialTransferMintState is the extension state for ExtensionConfidentialTransferMint.
type ConfidentialTransferMintState struct {
	// Authority to modify the ConfidentialTransferMint configuration.
	Authority OptionalPubkey
	// Determines if newly configured accounts must be approved before they can be used.
	AutoApproveNewAccounts bool
	// ElGamal pubkey used to encrypt data for the auditor.
	AuditorElGamalPubkey [32]byte
}

// ConfidentialTransferAccountState is the extension state for ExtensionConfidentialTransferAccount.
type ConfidentialTransferAccountState struct {
	// `true` if this account has been approved for use.
	Approved bool
	// The ElGamal public key associated with the account.
	ElGamalPubkey [32]byte
	// The pending balance (encrypted by ElGamal).
	PendingBalanceLo [64]byte
	PendingBalanceHi [64]byte
	// The available balance (encrypted by ElGamal).
	AvailableBalance [64]byte
	// The decryptable available balance.
	DecryptableAvailableBalance [36]byte
	// If `true`, the extended deposit side-car needs to be checked for pending deposits.
	AllowConfidentialCredits bool
	// If `true`, credits to this account must have an ElGamal-encrypted memo.
	AllowNonConfidentialCredits bool
	// The total number of `Deposit` and `Transfer` instructions applied.
	PendingBalanceCreditCounter uint64
	// The maximum number of `Deposit` and `Transfer` instructions that can be applied
	// before the `ApplyPendingBalance` instruction is used.
	MaximumPendingBalanceCreditCounter uint64
	// The `expected_pending_balance_credit_counter` value used in the last `ApplyPendingBalance`.
	ExpectedPendingBalanceCreditCounter uint64
	// The actual `pending_balance_credit_counter` when the last `ApplyPendingBalance` was applied.
	ActualPendingBalanceCreditCounter uint64
}

// --- DefaultAccountState ---

// DefaultAccountStateConfig is the extension state for ExtensionDefaultAccountState.
type DefaultAccountStateConfig struct {
	// The default state for new accounts.
	State AccountState
}

// --- ImmutableOwner ---
// ImmutableOwner is a marker extension with no state data.

// --- MemoTransfer ---

// MemoTransferState is the extension state for ExtensionMemoTransfer (account extension).
type MemoTransferState struct {
	// If true, incoming transfers must have a memo.
	RequireIncomingTransferMemos bool
}

// --- NonTransferable ---
// NonTransferable and NonTransferableAccount are marker extensions with no state data.

// --- InterestBearingConfig ---

// InterestBearingConfigState is the extension state for ExtensionInterestBearingConfig.
type InterestBearingConfigState struct {
	// Authority that can set the interest rate.
	RateAuthority OptionalPubkey
	// Timestamp of initialization, from which interest is accrued.
	InitializationTimestamp int64
	// Pre-update average rate, in basis points.
	PreUpdateAverageRate int16
	// Timestamp of the last update.
	LastUpdateTimestamp int64
	// Current rate, in basis points.
	CurrentRate int16
}

// --- CpiGuard ---

// CpiGuardState is the extension state for ExtensionCpiGuard (account extension).
type CpiGuardState struct {
	// If true, CPI is locked for privileged operations.
	LockCpi bool
}

// --- PermanentDelegate ---

// PermanentDelegateState is the extension state for ExtensionPermanentDelegate.
type PermanentDelegateState struct {
	// The permanent delegate.
	Delegate OptionalPubkey
}

// --- TransferHook ---

// TransferHookState is the extension state for ExtensionTransferHook (mint extension).
type TransferHookState struct {
	// Authority that can set the transfer hook program ID.
	Authority OptionalPubkey
	// The transfer hook program ID.
	ProgramID OptionalPubkey
}

// TransferHookAccountState is the extension state for ExtensionTransferHookAccount (account extension).
type TransferHookAccountState struct {
	// Whether the account is currently in the middle of a transfer.
	Transferring bool
}

// --- ConfidentialTransferFee ---

// ConfidentialTransferFeeConfigState is the extension state for ExtensionConfidentialTransferFeeConfig.
type ConfidentialTransferFeeConfigState struct {
	// Authority to set the withdraw withheld authority ElGamal key.
	Authority OptionalPubkey
	// ElGamal pubkey used to encrypt withheld fees.
	WithdrawWithheldAuthorityElGamalPubkey [32]byte
	// If true, harvest to mint is enabled.
	HarvestToMintEnabled bool
	// Withheld amount encrypted by the withdraw withheld authority.
	WithheldAmount [64]byte
}

// ConfidentialTransferFeeAmountState is the extension state for ExtensionConfidentialTransferFeeAmount.
type ConfidentialTransferFeeAmountState struct {
	// Encrypted withheld fees on this account.
	WithheldAmount [64]byte
}

// --- MetadataPointer ---

// MetadataPointerState is the extension state for ExtensionMetadataPointer.
type MetadataPointerState struct {
	// Authority that can set the metadata address.
	Authority OptionalPubkey
	// Account address that holds the metadata.
	MetadataAddress OptionalPubkey
}

// --- TokenMetadata ---

// TokenMetadataState is the extension state for ExtensionTokenMetadata.
// This is a variable-length extension.
type TokenMetadataState struct {
	// The authority that can update the metadata.
	UpdateAuthority OptionalPubkey
	// The associated mint.
	Mint solana.PublicKey
	// The name of the token.
	Name string
	// The symbol of the token.
	Symbol string
	// The URI of the token metadata.
	Uri string
	// Additional metadata as key-value pairs.
	AdditionalMetadata []MetadataField
}

// MetadataField is a key-value pair for additional token metadata.
type MetadataField struct {
	Key   string
	Value string
}

func (m *TokenMetadataState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	// UpdateAuthority: 32 bytes OptionalNonZeroPubkey
	{
		v, err := dec.ReadNBytes(32)
		if err != nil {
			return err
		}
		m.UpdateAuthority = OptionalPubkey{Key: solana.PublicKeyFromBytes(v)}
	}
	// Mint
	{
		v, err := dec.ReadNBytes(32)
		if err != nil {
			return err
		}
		m.Mint = solana.PublicKeyFromBytes(v)
	}
	// Name (borsh string: u32 length prefix + bytes)
	{
		length, err := dec.ReadUint32(binary.LittleEndian)
		if err != nil {
			return err
		}
		if length > maxMetadataStringLen {
			return fmt.Errorf("token metadata name too long: %d", length)
		}
		v, err := dec.ReadNBytes(int(length))
		if err != nil {
			return err
		}
		m.Name = string(v)
	}
	// Symbol
	{
		length, err := dec.ReadUint32(binary.LittleEndian)
		if err != nil {
			return err
		}
		if length > maxMetadataStringLen {
			return fmt.Errorf("token metadata symbol too long: %d", length)
		}
		v, err := dec.ReadNBytes(int(length))
		if err != nil {
			return err
		}
		m.Symbol = string(v)
	}
	// Uri
	{
		length, err := dec.ReadUint32(binary.LittleEndian)
		if err != nil {
			return err
		}
		if length > maxMetadataStringLen {
			return fmt.Errorf("token metadata uri too long: %d", length)
		}
		v, err := dec.ReadNBytes(int(length))
		if err != nil {
			return err
		}
		m.Uri = string(v)
	}
	// AdditionalMetadata: borsh Vec<(String, String)>
	{
		count, err := dec.ReadUint32(binary.LittleEndian)
		if err != nil {
			return err
		}
		if count > maxMetadataFields {
			return fmt.Errorf("token metadata additional fields count too large: %d", count)
		}
		m.AdditionalMetadata = make([]MetadataField, count)
		for i := uint32(0); i < count; i++ {
			// Key
			kLen, err := dec.ReadUint32(binary.LittleEndian)
			if err != nil {
				return err
			}
			if kLen > maxMetadataStringLen {
				return fmt.Errorf("token metadata field key too long: %d", kLen)
			}
			kBytes, err := dec.ReadNBytes(int(kLen))
			if err != nil {
				return err
			}
			// Value
			vLen, err := dec.ReadUint32(binary.LittleEndian)
			if err != nil {
				return err
			}
			if vLen > maxMetadataStringLen {
				return fmt.Errorf("token metadata field value too long: %d", vLen)
			}
			vBytes, err := dec.ReadNBytes(int(vLen))
			if err != nil {
				return err
			}
			m.AdditionalMetadata[i] = MetadataField{Key: string(kBytes), Value: string(vBytes)}
		}
	}
	return nil
}

func (m TokenMetadataState) MarshalWithEncoder(enc *bin.Encoder) error {
	// UpdateAuthority
	if err := enc.WriteBytes(m.UpdateAuthority.Key[:], false); err != nil {
		return err
	}
	// Mint
	if err := enc.WriteBytes(m.Mint[:], false); err != nil {
		return err
	}
	// Name
	if err := enc.WriteUint32(uint32(len(m.Name)), binary.LittleEndian); err != nil {
		return err
	}
	if err := enc.WriteBytes([]byte(m.Name), false); err != nil {
		return err
	}
	// Symbol
	if err := enc.WriteUint32(uint32(len(m.Symbol)), binary.LittleEndian); err != nil {
		return err
	}
	if err := enc.WriteBytes([]byte(m.Symbol), false); err != nil {
		return err
	}
	// Uri
	if err := enc.WriteUint32(uint32(len(m.Uri)), binary.LittleEndian); err != nil {
		return err
	}
	if err := enc.WriteBytes([]byte(m.Uri), false); err != nil {
		return err
	}
	// AdditionalMetadata
	if err := enc.WriteUint32(uint32(len(m.AdditionalMetadata)), binary.LittleEndian); err != nil {
		return err
	}
	for _, field := range m.AdditionalMetadata {
		if err := enc.WriteUint32(uint32(len(field.Key)), binary.LittleEndian); err != nil {
			return err
		}
		if err := enc.WriteBytes([]byte(field.Key), false); err != nil {
			return err
		}
		if err := enc.WriteUint32(uint32(len(field.Value)), binary.LittleEndian); err != nil {
			return err
		}
		if err := enc.WriteBytes([]byte(field.Value), false); err != nil {
			return err
		}
	}
	return nil
}

// --- GroupPointer ---

// GroupPointerState is the extension state for ExtensionGroupPointer.
type GroupPointerState struct {
	// Authority that can set the group address.
	Authority OptionalPubkey
	// Account address that holds the group.
	GroupAddress OptionalPubkey
}

// --- TokenGroup ---

// TokenGroupState is the extension state for ExtensionTokenGroup.
type TokenGroupState struct {
	// The authority that can update the group.
	UpdateAuthority OptionalPubkey
	// The associated mint.
	Mint solana.PublicKey
	// The current number of group members.
	Size uint32
	// The maximum number of group members.
	MaxSize uint32
}

// --- GroupMemberPointer ---

// GroupMemberPointerState is the extension state for ExtensionGroupMemberPointer.
type GroupMemberPointerState struct {
	// Authority that can set the member address.
	Authority OptionalPubkey
	// Account address that holds the member.
	MemberAddress OptionalPubkey
}

// --- TokenGroupMember ---

// TokenGroupMemberState is the extension state for ExtensionTokenGroupMember.
type TokenGroupMemberState struct {
	// The associated mint.
	Mint solana.PublicKey
	// The parent group.
	Group solana.PublicKey
	// The member number.
	MemberNumber uint32
}

// --- ConfidentialMintBurn ---

// ConfidentialMintBurnState is the extension state for ExtensionConfidentialMintBurn.
type ConfidentialMintBurnState struct {
	// Authority to modify the confidential mint burn configuration.
	ConfidentialSupply [64]byte
	// The decryptable supply.
	DecryptableSupply [36]byte
	// The ElGamal pubkey used for supply encryption.
	SupplyElGamalPubkey [32]byte
}

// --- ScaledUiAmount ---

// ScaledUiAmountState is the extension state for ExtensionScaledUiAmount.
type ScaledUiAmountState struct {
	// Authority that can set the multiplier.
	Authority OptionalPubkey
	// The current multiplier.
	Multiplier float64
	// The timestamp at which the new multiplier takes effect.
	NewMultiplierEffectiveTimestamp int64
	// The new multiplier.
	NewMultiplier float64
}

// --- Pausable ---

// PausableState is the extension state for ExtensionPausable.
type PausableState struct {
	// Authority that can pause/resume.
	Authority OptionalPubkey
	// If true, minting/burning/transferring is paused.
	Paused bool
}

// PausableAccountState is the extension state for ExtensionPausableAccount.
// This is a marker extension with no additional state.

// --- PermissionedBurn ---

// PermissionedBurnState is the extension state for ExtensionPermissionedBurn.
type PermissionedBurnState struct {
	// Authority that must approve burns.
	Authority OptionalPubkey
}

// --- Extension TLV Parsing ---

// ExtensionTLV represents a parsed extension from account data.
type ExtensionTLV struct {
	Type   ExtensionType
	Length uint16
	Data   []byte
}

// ParseExtensions parses extension TLV entries from raw account data.
// The data should start after the base account/mint data and the account type byte.
func ParseExtensions(data []byte) ([]ExtensionTLV, error) {
	var extensions []ExtensionTLV
	offset := 0
	for offset+4 <= len(data) {
		extType := ExtensionType(binary.LittleEndian.Uint16(data[offset:]))
		extLen := binary.LittleEndian.Uint16(data[offset+2:])
		offset += 4
		if offset+int(extLen) > len(data) {
			return extensions, fmt.Errorf("extension data truncated: need %d bytes, have %d", extLen, len(data)-offset)
		}
		extensions = append(extensions, ExtensionTLV{
			Type:   extType,
			Length: extLen,
			Data:   data[offset : offset+int(extLen)],
		})
		offset += int(extLen)
	}
	return extensions, nil
}

// ParseMintWithExtensions parses a Mint and its extensions from raw account data.
func ParseMintWithExtensions(data []byte) (*Mint, []ExtensionTLV, error) {
	if len(data) < MINT_SIZE {
		return nil, nil, fmt.Errorf("data too short for mint: %d < %d", len(data), MINT_SIZE)
	}
	mint := new(Mint)
	dec := bin.NewBinDecoder(data[:MINT_SIZE])
	if err := mint.UnmarshalWithDecoder(dec); err != nil {
		return nil, nil, fmt.Errorf("unable to decode mint: %w", err)
	}
	if len(data) <= MINT_SIZE {
		return mint, nil, nil
	}
	// Skip padding to multisig boundary and account type byte.
	extStart := ACCOUNT_SIZE + 1
	if extStart > len(data) {
		return mint, nil, nil
	}
	extensions, err := ParseExtensions(data[extStart:])
	if err != nil {
		return mint, extensions, err
	}
	return mint, extensions, nil
}

// ParseAccountWithExtensions parses a Token Account and its extensions from raw account data.
func ParseAccountWithExtensions(data []byte) (*Account, []ExtensionTLV, error) {
	if len(data) < ACCOUNT_SIZE {
		return nil, nil, fmt.Errorf("data too short for account: %d < %d", len(data), ACCOUNT_SIZE)
	}
	acct := new(Account)
	dec := bin.NewBinDecoder(data[:ACCOUNT_SIZE])
	if err := acct.UnmarshalWithDecoder(dec); err != nil {
		return nil, nil, fmt.Errorf("unable to decode account: %w", err)
	}
	if len(data) <= ACCOUNT_SIZE {
		return acct, nil, nil
	}
	// Account type byte is at offset ACCOUNT_SIZE.
	extStart := ACCOUNT_SIZE + 1
	if extStart > len(data) {
		return acct, nil, nil
	}
	extensions, err := ParseExtensions(data[extStart:])
	if err != nil {
		return acct, extensions, err
	}
	return acct, extensions, nil
}
