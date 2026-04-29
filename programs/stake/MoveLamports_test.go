package stake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_MoveLamports(t *testing.T) {
	inst := NewMoveLamportsInstruction(3_000_000_000, pubkeyOf(1), pubkeyOf(2), pubkeyOf(3))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_MoveLamports), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	ml := decoded.Impl.(*MoveLamports)
	require.Equal(t, uint64(3_000_000_000), *ml.Lamports)
}
