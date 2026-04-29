package stake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_GetMinimumDelegation(t *testing.T) {
	inst := NewGetMinimumDelegationInstruction()
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_GetMinimumDelegation), data[:4])
	require.Len(t, data, 4)
}
