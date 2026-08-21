package token2022

import (
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

// This file implements byte-exact codecs for the fixed-size token-2022
// extension states, mirroring the Pod layouts of the Rust
// spl-token-2022-interface crate: all fields are little-endian with align 1,
// so structs serialize as the plain concatenation of their fields.
//
// Booleans follow PodBool semantics: any nonzero byte decodes as true, and
// true encodes as 1. Optional pubkeys are raw 32 bytes where all zeros means
// None (no tag byte).

func (o *OptionalPubkey) UnmarshalWithDecoder(dec *bin.Decoder) error {
	v, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	o.Key = solana.PublicKeyFromBytes(v)
	return nil
}

func (o OptionalPubkey) MarshalWithEncoder(enc *bin.Encoder) error {
	return enc.WriteBytes(o.Key[:], false)
}

// readFixedBytes fills dst from the decoder.
func readFixedBytes(dec *bin.Decoder, dst []byte) error {
	v, err := dec.ReadNBytes(len(dst))
	if err != nil {
		return err
	}
	copy(dst, v)
	return nil
}

// decodeFixedState decodes an extension state with a fixed wire size,
// rejecting data of any other length.
func decodeFixedState[T any, PT interface {
	*T
	bin.BinaryUnmarshaler
}](data []byte, want int, name string) (*T, error) {
	if len(data) != want {
		return nil, fmt.Errorf("%w: %s requires %d bytes, got %d", ErrInvalidExtensionLength, name, want, len(data))
	}
	out := PT(new(T))
	if err := out.UnmarshalWithDecoder(bin.NewBinDecoder(data)); err != nil {
		return nil, fmt.Errorf("unable to decode %s: %w", name, err)
	}
	return out, nil
}

// --- TransferFee ---

func (f *TransferFee) UnmarshalWithDecoder(dec *bin.Decoder) error {
	var err error
	if f.Epoch, err = dec.ReadUint64(bin.LE); err != nil {
		return err
	}
	if f.MaximumFee, err = dec.ReadUint64(bin.LE); err != nil {
		return err
	}
	if f.TransferFeeBasisPoints, err = dec.ReadUint16(bin.LE); err != nil {
		return err
	}
	return nil
}

func (f TransferFee) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteUint64(f.Epoch, bin.LE); err != nil {
		return err
	}
	if err := enc.WriteUint64(f.MaximumFee, bin.LE); err != nil {
		return err
	}
	return enc.WriteUint16(f.TransferFeeBasisPoints, bin.LE)
}

// --- TransferFeeConfigState (108 bytes) ---

func (s *TransferFeeConfigState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := s.TransferFeeConfigAuthority.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	if err := s.WithdrawWithheldAuthority.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	var err error
	if s.WithheldAmount, err = dec.ReadUint64(bin.LE); err != nil {
		return err
	}
	if err := s.OlderTransferFee.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	return s.NewerTransferFee.UnmarshalWithDecoder(dec)
}

func (s TransferFeeConfigState) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := s.TransferFeeConfigAuthority.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := s.WithdrawWithheldAuthority.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := enc.WriteUint64(s.WithheldAmount, bin.LE); err != nil {
		return err
	}
	if err := s.OlderTransferFee.MarshalWithEncoder(enc); err != nil {
		return err
	}
	return s.NewerTransferFee.MarshalWithEncoder(enc)
}

// DecodeTransferFeeConfigState decodes a TransferFeeConfigState from extension data.
func DecodeTransferFeeConfigState(data []byte) (*TransferFeeConfigState, error) {
	return decodeFixedState[TransferFeeConfigState](data, 108, "TransferFeeConfig")
}

// --- TransferFeeAmountState (8 bytes) ---

func (s *TransferFeeAmountState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	var err error
	s.WithheldAmount, err = dec.ReadUint64(bin.LE)
	return err
}

func (s TransferFeeAmountState) MarshalWithEncoder(enc *bin.Encoder) error {
	return enc.WriteUint64(s.WithheldAmount, bin.LE)
}

// DecodeTransferFeeAmountState decodes a TransferFeeAmountState from extension data.
func DecodeTransferFeeAmountState(data []byte) (*TransferFeeAmountState, error) {
	return decodeFixedState[TransferFeeAmountState](data, 8, "TransferFeeAmount")
}

// --- MintCloseAuthorityState (32 bytes) ---

func (s *MintCloseAuthorityState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	return s.CloseAuthority.UnmarshalWithDecoder(dec)
}

func (s MintCloseAuthorityState) MarshalWithEncoder(enc *bin.Encoder) error {
	return s.CloseAuthority.MarshalWithEncoder(enc)
}

// DecodeMintCloseAuthorityState decodes a MintCloseAuthorityState from extension data.
func DecodeMintCloseAuthorityState(data []byte) (*MintCloseAuthorityState, error) {
	return decodeFixedState[MintCloseAuthorityState](data, 32, "MintCloseAuthority")
}

// --- ConfidentialTransferMintState (65 bytes) ---

func (s *ConfidentialTransferMintState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := s.Authority.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	var err error
	if s.AutoApproveNewAccounts, err = dec.ReadBool(); err != nil {
		return err
	}
	return readFixedBytes(dec, s.AuditorElGamalPubkey[:])
}

func (s ConfidentialTransferMintState) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := s.Authority.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := enc.WriteBool(s.AutoApproveNewAccounts); err != nil {
		return err
	}
	return enc.WriteBytes(s.AuditorElGamalPubkey[:], false)
}

// DecodeConfidentialTransferMintState decodes a ConfidentialTransferMintState from extension data.
func DecodeConfidentialTransferMintState(data []byte) (*ConfidentialTransferMintState, error) {
	return decodeFixedState[ConfidentialTransferMintState](data, 65, "ConfidentialTransferMint")
}

// --- ConfidentialTransferAccountState (295 bytes) ---

func (s *ConfidentialTransferAccountState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	var err error
	if s.Approved, err = dec.ReadBool(); err != nil {
		return err
	}
	if err = readFixedBytes(dec, s.ElGamalPubkey[:]); err != nil {
		return err
	}
	if err = readFixedBytes(dec, s.PendingBalanceLo[:]); err != nil {
		return err
	}
	if err = readFixedBytes(dec, s.PendingBalanceHi[:]); err != nil {
		return err
	}
	if err = readFixedBytes(dec, s.AvailableBalance[:]); err != nil {
		return err
	}
	if err = readFixedBytes(dec, s.DecryptableAvailableBalance[:]); err != nil {
		return err
	}
	if s.AllowConfidentialCredits, err = dec.ReadBool(); err != nil {
		return err
	}
	if s.AllowNonConfidentialCredits, err = dec.ReadBool(); err != nil {
		return err
	}
	if s.PendingBalanceCreditCounter, err = dec.ReadUint64(bin.LE); err != nil {
		return err
	}
	if s.MaximumPendingBalanceCreditCounter, err = dec.ReadUint64(bin.LE); err != nil {
		return err
	}
	if s.ExpectedPendingBalanceCreditCounter, err = dec.ReadUint64(bin.LE); err != nil {
		return err
	}
	s.ActualPendingBalanceCreditCounter, err = dec.ReadUint64(bin.LE)
	return err
}

func (s ConfidentialTransferAccountState) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteBool(s.Approved); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.ElGamalPubkey[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.PendingBalanceLo[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.PendingBalanceHi[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.AvailableBalance[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.DecryptableAvailableBalance[:], false); err != nil {
		return err
	}
	if err := enc.WriteBool(s.AllowConfidentialCredits); err != nil {
		return err
	}
	if err := enc.WriteBool(s.AllowNonConfidentialCredits); err != nil {
		return err
	}
	if err := enc.WriteUint64(s.PendingBalanceCreditCounter, bin.LE); err != nil {
		return err
	}
	if err := enc.WriteUint64(s.MaximumPendingBalanceCreditCounter, bin.LE); err != nil {
		return err
	}
	if err := enc.WriteUint64(s.ExpectedPendingBalanceCreditCounter, bin.LE); err != nil {
		return err
	}
	return enc.WriteUint64(s.ActualPendingBalanceCreditCounter, bin.LE)
}

// DecodeConfidentialTransferAccountState decodes a ConfidentialTransferAccountState from extension data.
func DecodeConfidentialTransferAccountState(data []byte) (*ConfidentialTransferAccountState, error) {
	return decodeFixedState[ConfidentialTransferAccountState](data, 295, "ConfidentialTransferAccount")
}

// --- DefaultAccountStateConfig (1 byte) ---

func (s *DefaultAccountStateConfig) UnmarshalWithDecoder(dec *bin.Decoder) error {
	v, err := dec.ReadUint8()
	if err != nil {
		return err
	}
	s.State = AccountState(v)
	return nil
}

func (s DefaultAccountStateConfig) MarshalWithEncoder(enc *bin.Encoder) error {
	return enc.WriteUint8(uint8(s.State))
}

// DecodeDefaultAccountStateConfig decodes a DefaultAccountStateConfig from extension data.
func DecodeDefaultAccountStateConfig(data []byte) (*DefaultAccountStateConfig, error) {
	return decodeFixedState[DefaultAccountStateConfig](data, 1, "DefaultAccountState")
}

// --- MemoTransferState (1 byte) ---

func (s *MemoTransferState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	var err error
	s.RequireIncomingTransferMemos, err = dec.ReadBool()
	return err
}

func (s MemoTransferState) MarshalWithEncoder(enc *bin.Encoder) error {
	return enc.WriteBool(s.RequireIncomingTransferMemos)
}

// DecodeMemoTransferState decodes a MemoTransferState from extension data.
func DecodeMemoTransferState(data []byte) (*MemoTransferState, error) {
	return decodeFixedState[MemoTransferState](data, 1, "MemoTransfer")
}

// --- InterestBearingConfigState (52 bytes) ---
// Note the interleaved field order: authority, i64 timestamp, i16 rate,
// i64 timestamp, i16 rate. This matches the on-chain layout exactly.

func (s *InterestBearingConfigState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := s.RateAuthority.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	var err error
	if s.InitializationTimestamp, err = dec.ReadInt64(bin.LE); err != nil {
		return err
	}
	if s.PreUpdateAverageRate, err = dec.ReadInt16(bin.LE); err != nil {
		return err
	}
	if s.LastUpdateTimestamp, err = dec.ReadInt64(bin.LE); err != nil {
		return err
	}
	s.CurrentRate, err = dec.ReadInt16(bin.LE)
	return err
}

func (s InterestBearingConfigState) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := s.RateAuthority.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := enc.WriteInt64(s.InitializationTimestamp, bin.LE); err != nil {
		return err
	}
	if err := enc.WriteInt16(s.PreUpdateAverageRate, bin.LE); err != nil {
		return err
	}
	if err := enc.WriteInt64(s.LastUpdateTimestamp, bin.LE); err != nil {
		return err
	}
	return enc.WriteInt16(s.CurrentRate, bin.LE)
}

// DecodeInterestBearingConfigState decodes an InterestBearingConfigState from extension data.
func DecodeInterestBearingConfigState(data []byte) (*InterestBearingConfigState, error) {
	return decodeFixedState[InterestBearingConfigState](data, 52, "InterestBearingConfig")
}

// --- CpiGuardState (1 byte) ---

func (s *CpiGuardState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	var err error
	s.LockCpi, err = dec.ReadBool()
	return err
}

func (s CpiGuardState) MarshalWithEncoder(enc *bin.Encoder) error {
	return enc.WriteBool(s.LockCpi)
}

// DecodeCpiGuardState decodes a CpiGuardState from extension data.
func DecodeCpiGuardState(data []byte) (*CpiGuardState, error) {
	return decodeFixedState[CpiGuardState](data, 1, "CpiGuard")
}

// --- PermanentDelegateState (32 bytes) ---

func (s *PermanentDelegateState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	return s.Delegate.UnmarshalWithDecoder(dec)
}

func (s PermanentDelegateState) MarshalWithEncoder(enc *bin.Encoder) error {
	return s.Delegate.MarshalWithEncoder(enc)
}

// DecodePermanentDelegateState decodes a PermanentDelegateState from extension data.
func DecodePermanentDelegateState(data []byte) (*PermanentDelegateState, error) {
	return decodeFixedState[PermanentDelegateState](data, 32, "PermanentDelegate")
}

// --- TransferHookState (64 bytes) ---

func (s *TransferHookState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := s.Authority.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	return s.ProgramID.UnmarshalWithDecoder(dec)
}

func (s TransferHookState) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := s.Authority.MarshalWithEncoder(enc); err != nil {
		return err
	}
	return s.ProgramID.MarshalWithEncoder(enc)
}

// DecodeTransferHookState decodes a TransferHookState from extension data.
func DecodeTransferHookState(data []byte) (*TransferHookState, error) {
	return decodeFixedState[TransferHookState](data, 64, "TransferHook")
}

// --- TransferHookAccountState (1 byte) ---

func (s *TransferHookAccountState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	var err error
	s.Transferring, err = dec.ReadBool()
	return err
}

func (s TransferHookAccountState) MarshalWithEncoder(enc *bin.Encoder) error {
	return enc.WriteBool(s.Transferring)
}

// DecodeTransferHookAccountState decodes a TransferHookAccountState from extension data.
func DecodeTransferHookAccountState(data []byte) (*TransferHookAccountState, error) {
	return decodeFixedState[TransferHookAccountState](data, 1, "TransferHookAccount")
}

// --- ConfidentialTransferFeeConfigState (129 bytes) ---

func (s *ConfidentialTransferFeeConfigState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := s.Authority.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	if err := readFixedBytes(dec, s.WithdrawWithheldAuthorityElGamalPubkey[:]); err != nil {
		return err
	}
	var err error
	if s.HarvestToMintEnabled, err = dec.ReadBool(); err != nil {
		return err
	}
	return readFixedBytes(dec, s.WithheldAmount[:])
}

func (s ConfidentialTransferFeeConfigState) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := s.Authority.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.WithdrawWithheldAuthorityElGamalPubkey[:], false); err != nil {
		return err
	}
	if err := enc.WriteBool(s.HarvestToMintEnabled); err != nil {
		return err
	}
	return enc.WriteBytes(s.WithheldAmount[:], false)
}

// DecodeConfidentialTransferFeeConfigState decodes a ConfidentialTransferFeeConfigState from extension data.
func DecodeConfidentialTransferFeeConfigState(data []byte) (*ConfidentialTransferFeeConfigState, error) {
	return decodeFixedState[ConfidentialTransferFeeConfigState](data, 129, "ConfidentialTransferFeeConfig")
}

// --- ConfidentialTransferFeeAmountState (64 bytes) ---

func (s *ConfidentialTransferFeeAmountState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	return readFixedBytes(dec, s.WithheldAmount[:])
}

func (s ConfidentialTransferFeeAmountState) MarshalWithEncoder(enc *bin.Encoder) error {
	return enc.WriteBytes(s.WithheldAmount[:], false)
}

// DecodeConfidentialTransferFeeAmountState decodes a ConfidentialTransferFeeAmountState from extension data.
func DecodeConfidentialTransferFeeAmountState(data []byte) (*ConfidentialTransferFeeAmountState, error) {
	return decodeFixedState[ConfidentialTransferFeeAmountState](data, 64, "ConfidentialTransferFeeAmount")
}

// --- MetadataPointerState (64 bytes) ---

func (s *MetadataPointerState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := s.Authority.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	return s.MetadataAddress.UnmarshalWithDecoder(dec)
}

func (s MetadataPointerState) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := s.Authority.MarshalWithEncoder(enc); err != nil {
		return err
	}
	return s.MetadataAddress.MarshalWithEncoder(enc)
}

// DecodeMetadataPointerState decodes a MetadataPointerState from extension data.
func DecodeMetadataPointerState(data []byte) (*MetadataPointerState, error) {
	return decodeFixedState[MetadataPointerState](data, 64, "MetadataPointer")
}

// --- TokenMetadataState (variable length) ---

// DecodeTokenMetadataState decodes a TokenMetadataState from extension data,
// requiring that all bytes are consumed.
func DecodeTokenMetadataState(data []byte) (*TokenMetadataState, error) {
	out := new(TokenMetadataState)
	dec := bin.NewBinDecoder(data)
	if err := out.UnmarshalWithDecoder(dec); err != nil {
		return nil, fmt.Errorf("unable to decode TokenMetadata: %w", err)
	}
	if dec.Remaining() != 0 {
		return nil, fmt.Errorf("%w: TokenMetadata has %d trailing bytes", ErrInvalidExtensionLength, dec.Remaining())
	}
	return out, nil
}

// --- GroupPointerState (64 bytes) ---

func (s *GroupPointerState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := s.Authority.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	return s.GroupAddress.UnmarshalWithDecoder(dec)
}

func (s GroupPointerState) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := s.Authority.MarshalWithEncoder(enc); err != nil {
		return err
	}
	return s.GroupAddress.MarshalWithEncoder(enc)
}

// DecodeGroupPointerState decodes a GroupPointerState from extension data.
func DecodeGroupPointerState(data []byte) (*GroupPointerState, error) {
	return decodeFixedState[GroupPointerState](data, 64, "GroupPointer")
}

// --- TokenGroup (80 bytes) ---

func (s *TokenGroup) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := s.UpdateAuthority.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	v, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	s.Mint = solana.PublicKeyFromBytes(v)
	if s.Size, err = dec.ReadUint64(bin.LE); err != nil {
		return err
	}
	s.MaxSize, err = dec.ReadUint64(bin.LE)
	return err
}

func (s TokenGroup) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := s.UpdateAuthority.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.Mint[:], false); err != nil {
		return err
	}
	if err := enc.WriteUint64(s.Size, bin.LE); err != nil {
		return err
	}
	return enc.WriteUint64(s.MaxSize, bin.LE)
}

// DecodeTokenGroup decodes a TokenGroup from extension data.
func DecodeTokenGroup(data []byte) (*TokenGroup, error) {
	return decodeFixedState[TokenGroup](data, 80, "TokenGroup")
}

// --- GroupMemberPointerState (64 bytes) ---

func (s *GroupMemberPointerState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := s.Authority.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	return s.MemberAddress.UnmarshalWithDecoder(dec)
}

func (s GroupMemberPointerState) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := s.Authority.MarshalWithEncoder(enc); err != nil {
		return err
	}
	return s.MemberAddress.MarshalWithEncoder(enc)
}

// DecodeGroupMemberPointerState decodes a GroupMemberPointerState from extension data.
func DecodeGroupMemberPointerState(data []byte) (*GroupMemberPointerState, error) {
	return decodeFixedState[GroupMemberPointerState](data, 64, "GroupMemberPointer")
}

// --- TokenGroupMember (72 bytes) ---

func (s *TokenGroupMember) UnmarshalWithDecoder(dec *bin.Decoder) error {
	v, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	s.Mint = solana.PublicKeyFromBytes(v)
	if v, err = dec.ReadNBytes(32); err != nil {
		return err
	}
	s.Group = solana.PublicKeyFromBytes(v)
	s.MemberNumber, err = dec.ReadUint64(bin.LE)
	return err
}

func (s TokenGroupMember) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteBytes(s.Mint[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.Group[:], false); err != nil {
		return err
	}
	return enc.WriteUint64(s.MemberNumber, bin.LE)
}

// DecodeTokenGroupMember decodes a TokenGroupMember from extension data.
func DecodeTokenGroupMember(data []byte) (*TokenGroupMember, error) {
	return decodeFixedState[TokenGroupMember](data, 72, "TokenGroupMember")
}

// --- ConfidentialMintBurnState (196 bytes) ---

func (s *ConfidentialMintBurnState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := readFixedBytes(dec, s.ConfidentialSupply[:]); err != nil {
		return err
	}
	if err := readFixedBytes(dec, s.DecryptableSupply[:]); err != nil {
		return err
	}
	if err := readFixedBytes(dec, s.SupplyElGamalPubkey[:]); err != nil {
		return err
	}
	return readFixedBytes(dec, s.PendingBurn[:])
}

func (s ConfidentialMintBurnState) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteBytes(s.ConfidentialSupply[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.DecryptableSupply[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.SupplyElGamalPubkey[:], false); err != nil {
		return err
	}
	return enc.WriteBytes(s.PendingBurn[:], false)
}

// DecodeConfidentialMintBurnState decodes a ConfidentialMintBurnState from extension data.
func DecodeConfidentialMintBurnState(data []byte) (*ConfidentialMintBurnState, error) {
	return decodeFixedState[ConfidentialMintBurnState](data, 196, "ConfidentialMintBurn")
}

// --- ScaledUiAmountState (56 bytes) ---

func (s *ScaledUiAmountState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := s.Authority.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	var err error
	if s.Multiplier, err = dec.ReadFloat64(bin.LE); err != nil {
		return err
	}
	if s.NewMultiplierEffectiveTimestamp, err = dec.ReadInt64(bin.LE); err != nil {
		return err
	}
	s.NewMultiplier, err = dec.ReadFloat64(bin.LE)
	return err
}

func (s ScaledUiAmountState) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := s.Authority.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := enc.WriteFloat64(s.Multiplier, bin.LE); err != nil {
		return err
	}
	if err := enc.WriteInt64(s.NewMultiplierEffectiveTimestamp, bin.LE); err != nil {
		return err
	}
	return enc.WriteFloat64(s.NewMultiplier, bin.LE)
}

// DecodeScaledUiAmountState decodes a ScaledUiAmountState from extension data.
func DecodeScaledUiAmountState(data []byte) (*ScaledUiAmountState, error) {
	return decodeFixedState[ScaledUiAmountState](data, 56, "ScaledUiAmount")
}

// --- PausableState (33 bytes) ---

func (s *PausableState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := s.Authority.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	var err error
	s.Paused, err = dec.ReadBool()
	return err
}

func (s PausableState) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := s.Authority.MarshalWithEncoder(enc); err != nil {
		return err
	}
	return enc.WriteBool(s.Paused)
}

// DecodePausableState decodes a PausableState from extension data.
func DecodePausableState(data []byte) (*PausableState, error) {
	return decodeFixedState[PausableState](data, 33, "Pausable")
}

// --- PermissionedBurnState (32 bytes) ---

func (s *PermissionedBurnState) UnmarshalWithDecoder(dec *bin.Decoder) error {
	return s.Authority.UnmarshalWithDecoder(dec)
}

func (s PermissionedBurnState) MarshalWithEncoder(enc *bin.Encoder) error {
	return s.Authority.MarshalWithEncoder(enc)
}

// DecodePermissionedBurnState decodes a PermissionedBurnState from extension data.
func DecodePermissionedBurnState(data []byte) (*PermissionedBurnState, error) {
	return decodeFixedState[PermissionedBurnState](data, 32, "PermissionedBurn")
}
