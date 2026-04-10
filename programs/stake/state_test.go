package stake

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// Official test vector: borsh_deserialization_live_data
// A real on-chain Initialized stake account (200 bytes).
var liveInitializedStakeAccount = []byte{
	1, 0, 0, 0, // discriminator = 1 (Initialized)
	128, 213, 34, 0, 0, 0, 0, 0, // rent_exempt_reserve = 2282880
	133, 0, 79, 231, 141, 29, 73, 61, // authorized.staker (32 bytes)
	232, 35, 119, 124, 168, 12, 120, 216,
	195, 29, 12, 166, 139, 28, 36, 182,
	186, 154, 246, 149, 224, 109, 52, 100,
	133, 0, 79, 231, 141, 29, 73, 61, // authorized.withdrawer (32 bytes, same as staker)
	232, 35, 119, 124, 168, 12, 120, 216,
	195, 29, 12, 166, 139, 28, 36, 182,
	186, 154, 246, 149, 224, 109, 52, 100,
	0, 0, 0, 0, 0, 0, 0, 0, // lockup.unix_timestamp = 0
	0, 0, 0, 0, 0, 0, 0, 0, // lockup.epoch = 0
	0, 0, 0, 0, 0, 0, 0, 0, // lockup.custodian (32 bytes, all zeros)
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	// remaining 76 bytes are zeros (padding to 200 total)
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
}

func TestOfficialVector_InitializedStakeAccount(t *testing.T) {
	require.Len(t, liveInitializedStakeAccount, 200)

	state, err := DecodeStakeAccount(liveInitializedStakeAccount)
	require.NoError(t, err)

	require.True(t, state.IsInitialized())
	require.False(t, state.IsStake())
	require.False(t, state.IsUninitialized())
	require.False(t, state.IsRewardsPool())

	require.NotNil(t, state.Meta)
	require.Equal(t, uint64(2282880), state.Meta.RentExemptReserve)

	expectedAuth := solana.PublicKey{
		133, 0, 79, 231, 141, 29, 73, 61,
		232, 35, 119, 124, 168, 12, 120, 216,
		195, 29, 12, 166, 139, 28, 36, 182,
		186, 154, 246, 149, 224, 109, 52, 100,
	}
	require.Equal(t, expectedAuth, state.Meta.Authorized.Staker)
	require.Equal(t, expectedAuth, state.Meta.Authorized.Withdrawer)

	require.Equal(t, int64(0), state.Meta.Lockup.UnixTimestamp)
	require.Equal(t, uint64(0), state.Meta.Lockup.Epoch)
	require.Equal(t, solana.PublicKey{}, state.Meta.Lockup.Custodian)

	require.Nil(t, state.Stake)
	require.Nil(t, state.Flags)
}

func TestDecodeStakeState_Uninitialized(t *testing.T) {
	data := make([]byte, 200)

	state, err := DecodeStakeAccount(data)
	require.NoError(t, err)
	require.True(t, state.IsUninitialized())
	require.Nil(t, state.Meta)
	require.Nil(t, state.Stake)
}

func TestDecodeStakeState_RewardsPool(t *testing.T) {
	data := make([]byte, 200)
	binary.LittleEndian.PutUint32(data, 3)

	state, err := DecodeStakeAccount(data)
	require.NoError(t, err)
	require.True(t, state.IsRewardsPool())
	require.Nil(t, state.Meta)
	require.Nil(t, state.Stake)
}

func TestDecodeStakeState_Stake(t *testing.T) {
	data := make([]byte, 200)
	offset := 0

	binary.LittleEndian.PutUint32(data[offset:], 2)
	offset += 4

	binary.LittleEndian.PutUint64(data[offset:], 2282880)
	offset += 8

	staker := pubkeyOf(0xAA)
	copy(data[offset:], staker[:])
	offset += 32

	withdrawer := pubkeyOf(0xBB)
	copy(data[offset:], withdrawer[:])
	offset += 32

	binary.LittleEndian.PutUint64(data[offset:], uint64(1234567890))
	offset += 8

	binary.LittleEndian.PutUint64(data[offset:], 100)
	offset += 8

	custodian := pubkeyOf(0xCC)
	copy(data[offset:], custodian[:])
	offset += 32

	voter := pubkeyOf(0xDD)
	copy(data[offset:], voter[:])
	offset += 32

	binary.LittleEndian.PutUint64(data[offset:], 5_000_000_000)
	offset += 8

	binary.LittleEndian.PutUint64(data[offset:], 50)
	offset += 8

	binary.LittleEndian.PutUint64(data[offset:], ^uint64(0))
	offset += 8

	binary.LittleEndian.PutUint64(data[offset:], math.Float64bits(0.25))
	offset += 8

	binary.LittleEndian.PutUint64(data[offset:], 999)
	offset += 8

	data[offset] = 1
	offset += 1

	require.Equal(t, 197, offset)

	state, err := DecodeStakeAccount(data)
	require.NoError(t, err)
	require.True(t, state.IsStake())

	require.Equal(t, uint64(2282880), state.Meta.RentExemptReserve)
	require.Equal(t, staker, state.Meta.Authorized.Staker)
	require.Equal(t, withdrawer, state.Meta.Authorized.Withdrawer)
	require.Equal(t, int64(1234567890), state.Meta.Lockup.UnixTimestamp)
	require.Equal(t, uint64(100), state.Meta.Lockup.Epoch)
	require.Equal(t, custodian, state.Meta.Lockup.Custodian)

	require.NotNil(t, state.Stake)
	require.Equal(t, voter, state.Stake.Delegation.VoterPubkey)
	require.Equal(t, uint64(5_000_000_000), state.Stake.Delegation.Stake)
	require.Equal(t, uint64(50), state.Stake.Delegation.ActivationEpoch)
	require.Equal(t, ^uint64(0), state.Stake.Delegation.DeactivationEpoch)
	require.Equal(t, 0.25, state.Stake.Delegation.WarmupCooldownRate)
	require.Equal(t, uint64(999), state.Stake.CreditsObserved)

	require.NotNil(t, state.Flags)
	require.Equal(t, uint8(1), state.Flags.Bits)
}

func TestStakeFlagsOffset(t *testing.T) {
	const flagOffset = 196

	data := make([]byte, 200)
	binary.LittleEndian.PutUint32(data, 2)
	data[flagOffset] = 1

	state, err := DecodeStakeAccount(data)
	require.NoError(t, err)
	require.True(t, state.IsStake())
	require.NotNil(t, state.Flags)
	require.Equal(t, StakeFlagsMustFullyActivateBeforeDeactivationIsPermitted.Bits, state.Flags.Bits)
}

func TestStakeStateV2_RoundTrip_Initialized(t *testing.T) {
	state := &StakeStateV2{
		Type: StakeStateInitialized,
		Meta: &Meta{
			RentExemptReserve: 2282880,
			Authorized: StateAuthorized{
				Staker:     pubkeyOf(1),
				Withdrawer: pubkeyOf(2),
			},
			Lockup: StateLockup{
				UnixTimestamp: 1000,
				Epoch:         42,
				Custodian:     pubkeyOf(3),
			},
		},
	}

	buf := new(bytes.Buffer)
	err := state.MarshalWithEncoder(bin.NewBinEncoder(buf))
	require.NoError(t, err)

	padded := make([]byte, 200)
	copy(padded, buf.Bytes())

	decoded, err := DecodeStakeAccount(padded)
	require.NoError(t, err)
	require.True(t, decoded.IsInitialized())
	require.Equal(t, state.Meta.RentExemptReserve, decoded.Meta.RentExemptReserve)
	require.Equal(t, state.Meta.Authorized.Staker, decoded.Meta.Authorized.Staker)
	require.Equal(t, state.Meta.Authorized.Withdrawer, decoded.Meta.Authorized.Withdrawer)
	require.Equal(t, state.Meta.Lockup.UnixTimestamp, decoded.Meta.Lockup.UnixTimestamp)
	require.Equal(t, state.Meta.Lockup.Epoch, decoded.Meta.Lockup.Epoch)
	require.Equal(t, state.Meta.Lockup.Custodian, decoded.Meta.Lockup.Custodian)
}

func TestStakeStateV2_RoundTrip_Stake(t *testing.T) {
	state := &StakeStateV2{
		Type: StakeStateStake,
		Meta: &Meta{
			RentExemptReserve: 2282880,
			Authorized: StateAuthorized{
				Staker:     pubkeyOf(1),
				Withdrawer: pubkeyOf(2),
			},
			Lockup: StateLockup{},
		},
		Stake: &StakeInfo{
			Delegation: Delegation{
				VoterPubkey:        pubkeyOf(5),
				Stake:              1_000_000_000,
				ActivationEpoch:    100,
				DeactivationEpoch:  ^uint64(0),
				WarmupCooldownRate: 0.25,
			},
			CreditsObserved: 500,
		},
		Flags: &StakeFlags{Bits: 0},
	}

	buf := new(bytes.Buffer)
	err := state.MarshalWithEncoder(bin.NewBinEncoder(buf))
	require.NoError(t, err)
	require.Equal(t, 197, buf.Len())

	padded := make([]byte, 200)
	copy(padded, buf.Bytes())

	decoded, err := DecodeStakeAccount(padded)
	require.NoError(t, err)
	require.True(t, decoded.IsStake())
	require.Equal(t, pubkeyOf(5), decoded.Stake.Delegation.VoterPubkey)
	require.Equal(t, uint64(1_000_000_000), decoded.Stake.Delegation.Stake)
	require.Equal(t, uint64(100), decoded.Stake.Delegation.ActivationEpoch)
	require.Equal(t, ^uint64(0), decoded.Stake.Delegation.DeactivationEpoch)
	require.Equal(t, 0.25, decoded.Stake.Delegation.WarmupCooldownRate)
	require.Equal(t, uint64(500), decoded.Stake.CreditsObserved)
	require.Equal(t, uint8(0), decoded.Flags.Bits)
}

func TestStakeAccountSize(t *testing.T) {
	require.Equal(t, 200, StakeAccountSize)

	state := &StakeStateV2{
		Type: StakeStateInitialized,
		Meta: &Meta{},
	}
	buf := new(bytes.Buffer)
	err := state.MarshalWithEncoder(bin.NewBinEncoder(buf))
	require.NoError(t, err)
	require.Equal(t, 124, buf.Len())

	state = &StakeStateV2{
		Type:  StakeStateStake,
		Meta:  &Meta{},
		Stake: &StakeInfo{},
		Flags: &StakeFlags{},
	}
	buf = new(bytes.Buffer)
	err = state.MarshalWithEncoder(bin.NewBinEncoder(buf))
	require.NoError(t, err)
	require.Equal(t, 197, buf.Len())
}

func TestDecodeStakeAccount_TooShort(t *testing.T) {
	data := make([]byte, 100)
	_, err := DecodeStakeAccount(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "too short")
}

func TestDecodeStakeAccount_InvalidDiscriminator(t *testing.T) {
	data := make([]byte, 200)
	binary.LittleEndian.PutUint32(data, 99)
	_, err := DecodeStakeAccount(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown stake state discriminator")
}
