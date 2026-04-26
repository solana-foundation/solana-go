package stake

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
)

// StakeAccountSize is the fixed size of a stake account (200 bytes).
const StakeAccountSize = 200

// StakeStateType identifies the variant of a StakeStateV2 account.
type StakeStateType uint32

// StakeStateV2 discriminator values (u32 LE).
const (
	StakeStateUninitialized StakeStateType = 0
	StakeStateInitialized   StakeStateType = 1
	StakeStateStake         StakeStateType = 2
	StakeStateRewardsPool   StakeStateType = 3
)

// StakeFlags is a bitflags wrapper for stake account flags.
type StakeFlags struct {
	Bits uint8
}

var (
	StakeFlagsEmpty = StakeFlags{Bits: 0}
	// MustFullyActivateBeforeDeactivationIsPermitted is deprecated (was for redelegate).
	StakeFlagsMustFullyActivateBeforeDeactivationIsPermitted = StakeFlags{Bits: 0b0000_0001}
)

// Delegation contains the delegation information for a stake account.
type Delegation struct {
	// Voter public key to which the stake is delegated.
	VoterPubkey solana.PublicKey
	// Activated stake amount in lamports.
	Stake uint64
	// Epoch at which the stake was activated.
	ActivationEpoch uint64
	// Epoch at which the stake was deactivated (u64::MAX means not deactivated).
	DeactivationEpoch uint64
	// Warmup/cooldown rate (deprecated since 1.16.7).
	WarmupCooldownRate float64
}

func (d *Delegation) UnmarshalWithDecoder(dec *bin.Decoder) error {
	v, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	d.VoterPubkey = solana.PublicKeyFromBytes(v)

	d.Stake, err = dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	d.ActivationEpoch, err = dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	d.DeactivationEpoch, err = dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	bits, err := dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	d.WarmupCooldownRate = math.Float64frombits(bits)
	return nil
}

func (d Delegation) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteBytes(d.VoterPubkey[:], false); err != nil {
		return err
	}
	if err := enc.WriteUint64(d.Stake, binary.LittleEndian); err != nil {
		return err
	}
	if err := enc.WriteUint64(d.ActivationEpoch, binary.LittleEndian); err != nil {
		return err
	}
	if err := enc.WriteUint64(d.DeactivationEpoch, binary.LittleEndian); err != nil {
		return err
	}
	if err := enc.WriteUint64(math.Float64bits(d.WarmupCooldownRate), binary.LittleEndian); err != nil {
		return err
	}
	return nil
}

// StakeInfo contains the staking information (Delegation + credits_observed).
type StakeInfo struct {
	Delegation      Delegation
	CreditsObserved uint64
}

func (s *StakeInfo) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := s.Delegation.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	var err error
	s.CreditsObserved, err = dec.ReadUint64(binary.LittleEndian)
	return err
}

func (s StakeInfo) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := s.Delegation.MarshalWithEncoder(enc); err != nil {
		return err
	}
	return enc.WriteUint64(s.CreditsObserved, binary.LittleEndian)
}

// Meta contains the metadata for a stake account.
type Meta struct {
	// Rent exempt reserve in lamports (deprecated since 3.0.1).
	RentExemptReserve uint64
	// Authorization settings.
	Authorized StateAuthorized
	// Lockup settings.
	Lockup StateLockup
}

// StateAuthorized is the on-chain Authorized struct (non-optional pubkeys).
// This differs from the instruction-level Authorized struct which uses pointer fields.
type StateAuthorized struct {
	Staker     solana.PublicKey
	Withdrawer solana.PublicKey
}

// StateLockup is the on-chain Lockup struct (non-optional fields).
// This differs from the instruction-level Lockup struct which uses pointer fields.
type StateLockup struct {
	UnixTimestamp int64
	Epoch         uint64
	Custodian     solana.PublicKey
}

func (m *Meta) UnmarshalWithDecoder(dec *bin.Decoder) error {
	var err error
	m.RentExemptReserve, err = dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	// Authorized.Staker
	v, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	m.Authorized.Staker = solana.PublicKeyFromBytes(v)
	// Authorized.Withdrawer
	v, err = dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	m.Authorized.Withdrawer = solana.PublicKeyFromBytes(v)
	// Lockup.UnixTimestamp
	m.Lockup.UnixTimestamp, err = dec.ReadInt64(binary.LittleEndian)
	if err != nil {
		return err
	}
	// Lockup.Epoch
	m.Lockup.Epoch, err = dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	// Lockup.Custodian
	v, err = dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	m.Lockup.Custodian = solana.PublicKeyFromBytes(v)
	return nil
}

func (m Meta) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteUint64(m.RentExemptReserve, binary.LittleEndian); err != nil {
		return err
	}
	if err := enc.WriteBytes(m.Authorized.Staker[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(m.Authorized.Withdrawer[:], false); err != nil {
		return err
	}
	if err := enc.WriteInt64(m.Lockup.UnixTimestamp, binary.LittleEndian); err != nil {
		return err
	}
	if err := enc.WriteUint64(m.Lockup.Epoch, binary.LittleEndian); err != nil {
		return err
	}
	if err := enc.WriteBytes(m.Lockup.Custodian[:], false); err != nil {
		return err
	}
	return nil
}

// StakeStateV2 represents the on-chain state of a stake account.
// The account is always 200 bytes. The discriminator (u32 LE) determines the variant.
type StakeStateV2 struct {
	// Discriminator: 0=Uninitialized, 1=Initialized, 2=Stake, 3=RewardsPool
	Type StakeStateType
	// Meta is present for Initialized and Stake variants (Type 1 or 2).
	Meta *Meta
	// Stake is present only for the Stake variant (Type 2).
	Stake *StakeInfo
	// Flags is present only for the Stake variant (Type 2).
	Flags *StakeFlags
}

// IsUninitialized returns true if the stake account is uninitialized.
func (s *StakeStateV2) IsUninitialized() bool { return s.Type == StakeStateUninitialized }

// IsInitialized returns true if the stake account is initialized but not staked.
func (s *StakeStateV2) IsInitialized() bool { return s.Type == StakeStateInitialized }

// IsStake returns true if the stake account is staked (delegated).
func (s *StakeStateV2) IsStake() bool { return s.Type == StakeStateStake }

// IsRewardsPool returns true if the stake account is a rewards pool.
func (s *StakeStateV2) IsRewardsPool() bool { return s.Type == StakeStateRewardsPool }

func (s *StakeStateV2) UnmarshalWithDecoder(dec *bin.Decoder) error {
	raw, err := dec.ReadUint32(binary.LittleEndian)
	s.Type = StakeStateType(raw)
	if err != nil {
		return err
	}
	switch s.Type {
	case StakeStateUninitialized, StakeStateRewardsPool:
		// No additional data.
	case StakeStateInitialized:
		s.Meta = new(Meta)
		if err := s.Meta.UnmarshalWithDecoder(dec); err != nil {
			return err
		}
	case StakeStateStake:
		s.Meta = new(Meta)
		if err := s.Meta.UnmarshalWithDecoder(dec); err != nil {
			return err
		}
		s.Stake = new(StakeInfo)
		if err := s.Stake.UnmarshalWithDecoder(dec); err != nil {
			return err
		}
		flagByte, err := dec.ReadUint8()
		if err != nil {
			return err
		}
		s.Flags = &StakeFlags{Bits: flagByte}
	default:
		return fmt.Errorf("unknown stake state discriminator: %d", s.Type)
	}
	return nil
}

// MarshalWithEncoder encodes the state WITHOUT padding to 200 bytes.
// On-chain accounts are always 200 bytes with trailing zero padding.
func (s StakeStateV2) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteUint32(uint32(s.Type), binary.LittleEndian); err != nil {
		return err
	}
	switch s.Type {
	case StakeStateUninitialized, StakeStateRewardsPool:
		// No additional data.
	case StakeStateInitialized:
		if s.Meta != nil {
			if err := s.Meta.MarshalWithEncoder(enc); err != nil {
				return err
			}
		}
	case StakeStateStake:
		if s.Meta != nil {
			if err := s.Meta.MarshalWithEncoder(enc); err != nil {
				return err
			}
		}
		if s.Stake != nil {
			if err := s.Stake.MarshalWithEncoder(enc); err != nil {
				return err
			}
		}
		flags := uint8(0)
		if s.Flags != nil {
			flags = s.Flags.Bits
		}
		if err := enc.WriteUint8(flags); err != nil {
			return err
		}
	}
	return nil
}

// DecodeStakeAccount decodes a StakeStateV2 from raw account data (200 bytes).
func DecodeStakeAccount(data []byte) (*StakeStateV2, error) {
	if len(data) < StakeAccountSize {
		return nil, fmt.Errorf("stake account data too short: %d < %d", len(data), StakeAccountSize)
	}
	state := new(StakeStateV2)
	dec := bin.NewBinDecoder(data)
	if err := state.UnmarshalWithDecoder(dec); err != nil {
		return nil, fmt.Errorf("unable to decode stake account: %w", err)
	}
	return state, nil
}
