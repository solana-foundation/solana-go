package stake

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip_InitializeChecked(t *testing.T) {
	inst := NewInitializeCheckedInstructionBuilder().
		SetStakeAccount(pubkeyOf(1)).
		SetRentSysvar(solana.SysVarRentPubkey).
		SetStakeAuthority(pubkeyOf(2)).
		SetWithdrawAuthority(pubkeyOf(3))

	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_InitializeChecked), data[:4])
	require.Len(t, data, 4)
}
