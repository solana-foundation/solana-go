package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_Vote_NoTimestamp(t *testing.T) {
	slots := []uint64{100, 101, 102}
	hash := hashOf(0xAB)

	inst := NewVoteInstruction(slots, hash, nil, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_Vote), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	v := decoded.Impl.(*Vote)
	require.Equal(t, slots, v.Slots)
	require.Equal(t, hash, v.Hash)
	require.Nil(t, v.Timestamp)
}

func TestRoundTrip_Vote_WithTimestamp(t *testing.T) {
	slots := []uint64{500}
	hash := hashOf(0xCD)
	ts := int64(1700000000)

	inst := NewVoteInstruction(slots, hash, &ts, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	v := decoded.Impl.(*Vote)
	require.Equal(t, slots, v.Slots)
	require.NotNil(t, v.Timestamp)
	require.Equal(t, ts, *v.Timestamp)
}
