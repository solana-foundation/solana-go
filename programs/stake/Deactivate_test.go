package stake

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip_Deactivate(t *testing.T) {
	inst := NewDeactivateInstructionBuilder().
		SetStakeAccount(pubkeyOf(1)).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetStakeAuthority(pubkeyOf(2))

	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_Deactivate), data[:4])
	require.Len(t, data, 4)
}
