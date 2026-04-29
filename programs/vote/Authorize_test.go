package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_Authorize(t *testing.T) {
	newAuth := pubkeyOf(5)
	inst := NewAuthorizeInstruction(newAuth, VoteAuthorizeWithdrawer, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_Authorize), data[:4])

	// Data: 32 bytes pubkey + 4 bytes kind
	expected := concat(u32LE(Instruction_Authorize), newAuth[:], u32LE(uint32(VoteAuthorizeWithdrawer)))
	require.Equal(t, expected, data)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	a := decoded.Impl.(*Authorize)
	require.Equal(t, newAuth, *a.NewAuthority)
	require.Equal(t, VoteAuthorizeWithdrawer, *a.VoteAuthorize)
}
