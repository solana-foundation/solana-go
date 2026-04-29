package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_InitializeAccount(t *testing.T) {
	node := pubkeyOf(1)
	voter := pubkeyOf(2)
	withdrawer := pubkeyOf(3)

	inst := NewInitializeAccountInstruction(node, voter, withdrawer, 42, pubkeyOf(10))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_InitializeAccount), data[:4])

	// Data payload: node(32) + voter(32) + withdrawer(32) + commission(1) = 97 bytes
	require.Len(t, data, 4+97)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	ia := decoded.Impl.(*InitializeAccount)
	require.Equal(t, node, ia.VoteInit.NodePubkey)
	require.Equal(t, voter, ia.VoteInit.AuthorizedVoter)
	require.Equal(t, withdrawer, ia.VoteInit.AuthorizedWithdrawer)
	require.Equal(t, uint8(42), ia.VoteInit.Commission)
}
