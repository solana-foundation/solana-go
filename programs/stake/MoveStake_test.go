package stake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_MoveStake(t *testing.T) {
	inst := NewMoveStakeInstruction(2_000_000_000, pubkeyOf(1), pubkeyOf(2), pubkeyOf(3))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_MoveStake), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	ms := decoded.Impl.(*MoveStake)
	require.Equal(t, uint64(2_000_000_000), *ms.Lamports)
}
