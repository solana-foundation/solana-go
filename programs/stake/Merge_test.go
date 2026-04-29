package stake

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip_Merge(t *testing.T) {
	inst := NewMergeInstructionBuilder().
		SetDestinationStakeAccount(pubkeyOf(1)).
		SetSourceStakeAccount(pubkeyOf(2)).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetStakeHistorySysvar(solana.SysVarStakeHistoryPubkey).
		SetStakeAuthority(pubkeyOf(3))

	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_Merge), data[:4])
	require.Len(t, data, 4)
}
