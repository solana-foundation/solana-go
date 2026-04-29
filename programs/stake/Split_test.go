package stake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_Split(t *testing.T) {
	inst := NewSplitInstruction(1_000_000_000, pubkeyOf(1), pubkeyOf(2), pubkeyOf(3))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_Split), data[:4])

	expected := concat(u32LE(3), u64LE(1_000_000_000))
	require.Equal(t, expected, data)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	split := decoded.Impl.(*Split)
	require.Equal(t, uint64(1_000_000_000), *split.Lamports)
}
