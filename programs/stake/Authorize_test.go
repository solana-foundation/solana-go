package stake

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip_Authorize(t *testing.T) {
	newAuth := pubkeyOf(5)
	inst := NewAuthorizeInstructionBuilder().
		SetStakeAccount(pubkeyOf(1)).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetAuthority(pubkeyOf(2)).
		SetNewAuthorized(newAuth).
		SetStakeAuthorize(StakeAuthorizeWithdrawer)

	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_Authorize), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	auth := decoded.Impl.(*Authorize)
	require.Equal(t, newAuth, *auth.NewAuthorized)
	require.Equal(t, StakeAuthorizeWithdrawer, *auth.StakeAuthorize)
}
