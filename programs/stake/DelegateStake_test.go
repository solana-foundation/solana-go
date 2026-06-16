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

// The on-chain Stake program's DelegateStake account spec is:
//
//  0. [WRITE]  stake account to be delegated
//  5. [SIGNER] stake authority
//
// i.e. the stake account is writable but not a signer, and the stake
// authority signs but is not writable. This matches the flags the other
// authority-based stake instructions (Authorize, Deactivate, Split) use.
func TestAccountFlags_DelegateStake(t *testing.T) {
	inst := NewDelegateStakeInstructionBuilder().
		SetStakeAccount(pubkeyOf(1)).
		SetStakeAuthority(pubkeyOf(4))

	stakeAccount := inst.GetStakeAccount()
	require.True(t, stakeAccount.IsWritable, "stake account must be writable")
	require.False(t, stakeAccount.IsSigner, "stake account must not be a signer")

	stakeAuthority := inst.GetStakeAuthority()
	require.True(t, stakeAuthority.IsSigner, "stake authority must be a signer")
	require.False(t, stakeAuthority.IsWritable, "stake authority must not be writable")
}
