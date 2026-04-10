package stake

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip_AuthorizeChecked(t *testing.T) {
	inst := NewAuthorizeCheckedInstructionBuilder().
		SetStakeAccount(pubkeyOf(1)).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetCurrentAuthority(pubkeyOf(2)).
		SetNewAuthority(pubkeyOf(3)).
		SetStakeAuthorize(StakeAuthorizeStaker)

	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_AuthorizeChecked), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	ac := decoded.Impl.(*AuthorizeChecked)
	require.Equal(t, StakeAuthorizeStaker, *ac.StakeAuthorize)
}
