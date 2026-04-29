package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_Withdraw(t *testing.T) {
	inst := NewWithdrawInstruction(1_000_000_000, pubkeyOf(1), pubkeyOf(2), pubkeyOf(3))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_Withdraw), data[:4])

	expected := concat(u32LE(Instruction_Withdraw), u64LE(1_000_000_000))
	require.Equal(t, expected, data)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	w := decoded.Impl.(*Withdraw)
	require.Equal(t, uint64(1_000_000_000), *w.Lamports)
}
