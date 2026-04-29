package stake

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip_DelegateStake(t *testing.T) {
	inst := NewDelegateStakeInstructionBuilder().
		SetStakeAccount(pubkeyOf(1)).
		SetVoteAccount(pubkeyOf(2)).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetStakeHistorySysvar(solana.SysVarStakeHistoryPubkey).
		SetConfigAccount(pubkeyOf(3)).
		SetStakeAuthority(pubkeyOf(4))

	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_DelegateStake), data[:4])
	require.Len(t, data, 4)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	require.IsType(t, &DelegateStake{}, decoded.Impl)
}
