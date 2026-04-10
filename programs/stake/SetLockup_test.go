package stake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_SetLockup(t *testing.T) {
	ts := int64(999)
	epoch := uint64(100)
	custodian := pubkeyOf(7)
	inst := NewSetLockupInstructionBuilder().
		SetStakeAccount(pubkeyOf(1)).
		SetAuthority(pubkeyOf(2)).
		SetLockupTimestamp(ts).
		SetLockupEpoch(epoch).
		SetCustodian(custodian)

	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_SetLockup), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	sl := decoded.Impl.(*SetLockup)
	require.Equal(t, ts, *sl.LockupArgs.UnixTimestamp)
	require.Equal(t, epoch, *sl.LockupArgs.Epoch)
	require.Equal(t, custodian, *sl.LockupArgs.Custodian)
}
