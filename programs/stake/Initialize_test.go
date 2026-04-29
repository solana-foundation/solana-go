package stake

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip_Initialize(t *testing.T) {
	staker := pubkeyOf(1)
	withdrawer := pubkeyOf(2)
	custodian := pubkeyOf(3)

	inst := NewInitializeInstructionBuilder().
		SetStakeAccount(pubkeyOf(10)).
		SetRentSysvarAccount(solana.SysVarRentPubkey).
		SetStaker(staker).
		SetWithdrawer(withdrawer).
		SetLockupTimestamp(1000).
		SetLockupEpoch(42).
		SetCustodian(custodian)

	data, err := encodeInst(inst)
	require.NoError(t, err)

	// First 4 bytes: instruction ID (u32 LE)
	require.Equal(t, u32LE(Instruction_Initialize), data[:4])

	// Decode back
	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	init := decoded.Impl.(*Initialize)
	require.Equal(t, staker, *init.Authorized.Staker)
	require.Equal(t, withdrawer, *init.Authorized.Withdrawer)
	require.Equal(t, int64(1000), *init.Lockup.UnixTimestamp)
	require.Equal(t, uint64(42), *init.Lockup.Epoch)
	require.Equal(t, custodian, *init.Lockup.Custodian)
}
