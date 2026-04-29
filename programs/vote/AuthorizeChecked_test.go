package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_AuthorizeChecked(t *testing.T) {
	inst := NewAuthorizeCheckedInstruction(VoteAuthorizeVoter, pubkeyOf(1), pubkeyOf(2), pubkeyOf(3))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_AuthorizeChecked), data[:4])

	expected := concat(u32LE(Instruction_AuthorizeChecked), u32LE(uint32(VoteAuthorizeVoter)))
	require.Equal(t, expected, data)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	ac := decoded.Impl.(*AuthorizeChecked)
	require.Equal(t, VoteAuthorizeVoter, *ac.VoteAuthorize)
}
