package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_UpdateValidatorIdentity(t *testing.T) {
	inst := NewUpdateValidatorIdentityInstruction(pubkeyOf(1), pubkeyOf(2), pubkeyOf(3))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_UpdateValidatorIdentity), data[:4])
	require.Len(t, data, 4)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	require.IsType(t, &UpdateValidatorIdentity{}, decoded.Impl)
}
