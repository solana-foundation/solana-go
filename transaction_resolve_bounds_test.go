package solana

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// A decoded transaction whose instruction references an account index past
// the message's key list must produce an error, not an index-out-of-range
// panic, from ResolveInstructionAccounts.
func TestResolveInstructionAccounts_OutOfRangeIndexIsAnError(t *testing.T) {
	msg := Message{
		AccountKeys: PublicKeySlice{{1}, {2}},
		Header:      MessageHeader{NumRequiredSignatures: 1},
	}
	ci := CompiledInstruction{ProgramIDIndex: 1, Accounts: []uint16{0, 48}}
	require.NotPanics(t, func() {
		_, err := ci.ResolveInstructionAccounts(&msg)
		require.Error(t, err)
	})
}
