package vote

import (
	"bytes"
	"testing"

	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/stretchr/testify/require"
)

func TestVoteStateV3_RoundTrip(t *testing.T) {
	root := uint64(100)
	state := VoteStateVersions{
		Version: VoteStateVersionV3,
		V3: &VoteStateV3{
			NodePubkey:           pubkeyOf(1),
			AuthorizedWithdrawer: pubkeyOf(2),
			Commission:           10,
			Votes: []LandedVote{
				{Latency: 2, Lockout: Lockout{Slot: 100, ConfirmationCount: 3}},
				{Latency: 1, Lockout: Lockout{Slot: 101, ConfirmationCount: 2}},
			},
			RootSlot: &root,
			AuthorizedVoters: AuthorizedVoters{
				Voters: []AuthorizedVoter{
					{Epoch: 42, Pubkey: pubkeyOf(3)},
				},
			},
			PriorVoters: PriorVotersCircBuf{IsEmpty: true},
			EpochCredits: []EpochCredit{
				{Epoch: 1, Credits: 1000, PrevCredits: 0},
				{Epoch: 2, Credits: 2000, PrevCredits: 1000},
			},
			LastTimestamp: BlockTimestamp{Slot: 500, Timestamp: 1700000000},
		},
	}

	buf := new(bytes.Buffer)
	require.NoError(t, state.MarshalWithEncoder(bin.NewBinEncoder(buf)))

	decoded, err := DecodeVoteAccount(buf.Bytes())
	require.NoError(t, err)
	require.Equal(t, VoteStateVersionV3, decoded.Version)
	require.NotNil(t, decoded.V3)
	require.Equal(t, pubkeyOf(1), decoded.V3.NodePubkey)
	require.Equal(t, pubkeyOf(2), decoded.V3.AuthorizedWithdrawer)
	require.Equal(t, uint8(10), decoded.V3.Commission)
	require.Len(t, decoded.V3.Votes, 2)
	require.Equal(t, uint8(2), decoded.V3.Votes[0].Latency)
	require.Equal(t, uint64(100), decoded.V3.Votes[0].Lockout.Slot)
	require.NotNil(t, decoded.V3.RootSlot)
	require.Equal(t, uint64(100), *decoded.V3.RootSlot)
	require.Len(t, decoded.V3.AuthorizedVoters.Voters, 1)
	require.Equal(t, uint64(42), decoded.V3.AuthorizedVoters.Voters[0].Epoch)
	require.Len(t, decoded.V3.EpochCredits, 2)
	require.Equal(t, uint64(1000), decoded.V3.EpochCredits[0].Credits)
	require.Equal(t, uint64(500), decoded.V3.LastTimestamp.Slot)
	require.Equal(t, int64(1700000000), decoded.V3.LastTimestamp.Timestamp)
}

func TestVoteStateV1_14_11_RoundTrip(t *testing.T) {
	state := VoteStateVersions{
		Version: VoteStateVersionV1_14_11,
		V1_14_11: &VoteState1_14_11{
			NodePubkey:           pubkeyOf(1),
			AuthorizedWithdrawer: pubkeyOf(2),
			Commission:           5,
			Votes: []Lockout{
				{Slot: 100, ConfirmationCount: 3},
			},
			AuthorizedVoters: AuthorizedVoters{Voters: []AuthorizedVoter{}},
			PriorVoters:      PriorVotersCircBuf{IsEmpty: true},
			EpochCredits:     []EpochCredit{},
			LastTimestamp:    BlockTimestamp{},
		},
	}

	buf := new(bytes.Buffer)
	require.NoError(t, state.MarshalWithEncoder(bin.NewBinEncoder(buf)))

	decoded, err := DecodeVoteAccount(buf.Bytes())
	require.NoError(t, err)
	require.Equal(t, VoteStateVersionV1_14_11, decoded.Version)
	require.NotNil(t, decoded.V1_14_11)
	require.Equal(t, uint8(5), decoded.V1_14_11.Commission)
	require.Len(t, decoded.V1_14_11.Votes, 1)
	require.Equal(t, uint64(100), decoded.V1_14_11.Votes[0].Slot)
}

func TestVoteStateV4_RoundTrip(t *testing.T) {
	blsPk := [BLS_PUBLIC_KEY_COMPRESSED_SIZE]byte{}
	for i := range blsPk {
		blsPk[i] = 0xAB
	}
	state := VoteStateVersions{
		Version: VoteStateVersionV4,
		V4: &VoteStateV4{
			NodePubkey:                    pubkeyOf(1),
			AuthorizedWithdrawer:          pubkeyOf(2),
			InflationRewardsCollector:     pubkeyOf(3),
			BlockRevenueCollector:         pubkeyOf(4),
			InflationRewardsCommissionBps: 500,
			BlockRevenueCommissionBps:     1000,
			PendingDelegatorRewards:       123456,
			BLSPubkeyCompressed:           &blsPk,
			Votes:                         []LandedVote{},
			AuthorizedVoters:              AuthorizedVoters{Voters: []AuthorizedVoter{}},
			EpochCredits:                  []EpochCredit{},
			LastTimestamp:                 BlockTimestamp{},
		},
	}

	buf := new(bytes.Buffer)
	require.NoError(t, state.MarshalWithEncoder(bin.NewBinEncoder(buf)))

	decoded, err := DecodeVoteAccount(buf.Bytes())
	require.NoError(t, err)
	require.Equal(t, VoteStateVersionV4, decoded.Version)
	require.NotNil(t, decoded.V4)
	require.Equal(t, pubkeyOf(3), decoded.V4.InflationRewardsCollector)
	require.Equal(t, pubkeyOf(4), decoded.V4.BlockRevenueCollector)
	require.Equal(t, uint16(500), decoded.V4.InflationRewardsCommissionBps)
	require.Equal(t, uint16(1000), decoded.V4.BlockRevenueCommissionBps)
	require.Equal(t, uint64(123456), decoded.V4.PendingDelegatorRewards)
	require.NotNil(t, decoded.V4.BLSPubkeyCompressed)
	require.Equal(t, blsPk, *decoded.V4.BLSPubkeyCompressed)
}

func TestVoteStateV4_NoBLSKey(t *testing.T) {
	state := VoteStateVersions{
		Version: VoteStateVersionV4,
		V4: &VoteStateV4{
			NodePubkey:           pubkeyOf(1),
			AuthorizedWithdrawer: pubkeyOf(2),
			Votes:                []LandedVote{},
			AuthorizedVoters:     AuthorizedVoters{Voters: []AuthorizedVoter{}},
			EpochCredits:         []EpochCredit{},
			LastTimestamp:        BlockTimestamp{},
		},
	}
	buf := new(bytes.Buffer)
	require.NoError(t, state.MarshalWithEncoder(bin.NewBinEncoder(buf)))

	decoded, err := DecodeVoteAccount(buf.Bytes())
	require.NoError(t, err)
	require.Nil(t, decoded.V4.BLSPubkeyCompressed)
}

// Ported from vote_state_v3.rs::test_invalid_option_bool_discriminants:
// In V3, root_slot Option starts at byte offset 77.
func TestVoteStateV3_RootSlotOffset(t *testing.T) {
	const rootSlotOffset = 77
	state := VoteStateVersions{
		Version: VoteStateVersionV3,
		V3: &VoteStateV3{
			NodePubkey:           pubkeyOf(0),
			AuthorizedWithdrawer: pubkeyOf(0),
			Votes:                []LandedVote{},
			AuthorizedVoters:     AuthorizedVoters{Voters: []AuthorizedVoter{}},
			PriorVoters:          PriorVotersCircBuf{IsEmpty: true},
			EpochCredits:         []EpochCredit{},
			LastTimestamp:        BlockTimestamp{},
		},
	}
	buf := new(bytes.Buffer)
	require.NoError(t, state.MarshalWithEncoder(bin.NewBinEncoder(buf)))
	data := buf.Bytes()

	// At offset 77 we should have the Option<u64> discriminator for root_slot.
	// With RootSlot == nil, discriminator should be 0.
	require.GreaterOrEqual(t, len(data), rootSlotOffset+1)
	require.Equal(t, uint8(0), data[rootSlotOffset])
}

// Ported from vote_state_v4.rs::test_invalid_option_discriminants:
// In V4, BLS Option discriminant is at byte offset 144.
func TestVoteStateV4_BLSOffset(t *testing.T) {
	const blsOffset = 144
	state := VoteStateVersions{
		Version: VoteStateVersionV4,
		V4: &VoteStateV4{
			NodePubkey:           pubkeyOf(0),
			AuthorizedWithdrawer: pubkeyOf(0),
			Votes:                []LandedVote{},
			AuthorizedVoters:     AuthorizedVoters{Voters: []AuthorizedVoter{}},
			EpochCredits:         []EpochCredit{},
			LastTimestamp:        BlockTimestamp{},
		},
	}
	buf := new(bytes.Buffer)
	require.NoError(t, state.MarshalWithEncoder(bin.NewBinEncoder(buf)))
	data := buf.Bytes()

	require.GreaterOrEqual(t, len(data), blsOffset+1)
	// BLSPubkeyCompressed is nil => discriminator = 0
	require.Equal(t, uint8(0), data[blsOffset])
}

func TestDecodeVoteAccount_Uninitialized_Rejected(t *testing.T) {
	data := make([]byte, 4) // tag = 0 (V0_23_5)
	_, err := DecodeVoteAccount(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "V0_23_5")
}

func TestDecodeVoteAccount_UnknownVersion(t *testing.T) {
	data := []byte{99, 0, 0, 0}
	_, err := DecodeVoteAccount(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown vote state version")
}

func TestDecodeVoteAccount_TooShort(t *testing.T) {
	_, err := DecodeVoteAccount([]byte{1, 2})
	require.Error(t, err)
}
