package stake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_DeactivateDelinquent(t *testing.T) {
	inst := NewDeactivateDelinquentInstruction(pubkeyOf(1), pubkeyOf(2), pubkeyOf(3))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_DeactivateDelinquent), data[:4])
	require.Len(t, data, 4)
}
