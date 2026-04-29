package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_VoteSwitch(t *testing.T) {
	slots := []uint64{100, 101}
	voteHash := hashOf(0xAB)
	proofHash := hashOf(0xCD)

	inst := NewVoteSwitchInstruction(slots, voteHash, nil, proofHash, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_VoteSwitch), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	vs := decoded.Impl.(*VoteSwitch)
	require.Equal(t, slots, vs.Vote.Slots)
	require.Equal(t, voteHash, vs.Vote.Hash)
	require.Equal(t, proofHash, vs.Hash)
}
