package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_UpdateVoteState(t *testing.T) {
	root := uint64(99)
	ts := int64(1700000000)
	update := VoteStateUpdate{
		Lockouts: []Lockout{
			{Slot: 100, ConfirmationCount: 3},
			{Slot: 101, ConfirmationCount: 2},
			{Slot: 102, ConfirmationCount: 1},
		},
		Root:      &root,
		Hash:      hashOf(0xAB),
		Timestamp: &ts,
	}

	inst := NewUpdateVoteStateInstruction(update, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_UpdateVoteState), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	u := decoded.Impl.(*UpdateVoteState)
	require.Len(t, u.Update.Lockouts, 3)
	require.Equal(t, uint64(100), u.Update.Lockouts[0].Slot)
	require.Equal(t, uint32(3), u.Update.Lockouts[0].ConfirmationCount)
	require.NotNil(t, u.Update.Root)
	require.Equal(t, uint64(99), *u.Update.Root)
	require.Equal(t, hashOf(0xAB), u.Update.Hash)
	require.NotNil(t, u.Update.Timestamp)
	require.Equal(t, ts, *u.Update.Timestamp)
}

func TestRoundTrip_UpdateVoteState_NoRoot_NoTs(t *testing.T) {
	update := VoteStateUpdate{
		Lockouts: []Lockout{{Slot: 1, ConfirmationCount: 1}},
		Hash:     hashOf(0xEE),
	}
	inst := NewUpdateVoteStateInstruction(update, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	u := decoded.Impl.(*UpdateVoteState)
	require.Nil(t, u.Update.Root)
	require.Nil(t, u.Update.Timestamp)
}

func TestRoundTrip_UpdateVoteStateSwitch(t *testing.T) {
	update := VoteStateUpdate{
		Lockouts: []Lockout{{Slot: 500, ConfirmationCount: 1}},
		Hash:     hashOf(0x11),
	}
	proof := hashOf(0x22)
	inst := NewUpdateVoteStateSwitchInstruction(update, proof, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_UpdateVoteStateSwitch), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	us := decoded.Impl.(*UpdateVoteStateSwitch)
	require.Equal(t, proof, us.Hash)
}

func TestRoundTrip_CompactUpdateVoteState(t *testing.T) {
	update := VoteStateUpdate{
		Lockouts: []Lockout{{Slot: 1, ConfirmationCount: 1}},
		Hash:     hashOf(0x33),
	}
	inst := NewCompactUpdateVoteStateInstruction(update, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_CompactUpdateVoteState), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	require.IsType(t, &CompactUpdateVoteState{}, decoded.Impl)
}
