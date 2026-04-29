package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_UpdateCommission(t *testing.T) {
	inst := NewUpdateCommissionInstruction(50, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_UpdateCommission), data[:4])

	expected := concat(u32LE(Instruction_UpdateCommission), []byte{50})
	require.Equal(t, expected, data)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	u := decoded.Impl.(*UpdateCommission)
	require.Equal(t, uint8(50), *u.Commission)
}
