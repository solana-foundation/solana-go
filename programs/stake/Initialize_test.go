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

// TestInitialize_AccountFlags pins the canonical Solana account spec for the
// stake Initialize instruction: the stake account is writable but NOT a signer
// (its keypair signs the preceding SystemProgram::CreateAccount, not this
// instruction), and the rent sysvar is read-only.
func TestInitialize_AccountFlags(t *testing.T) {
	inst := NewInitializeInstruction(pubkeyOf(1), pubkeyOf(2), pubkeyOf(10))

	stakeAccount := inst.GetStakeAccount()
	require.True(t, stakeAccount.IsWritable, "stake account must be writable")
	require.False(t, stakeAccount.IsSigner, "stake account must not be a signer")

	rent := inst.GetRentSysvarAccount()
	require.False(t, rent.IsWritable, "rent sysvar must be read-only")
	require.False(t, rent.IsSigner, "rent sysvar must not be a signer")
}
