package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_Withdraw(t *testing.T) {
	inst := NewWithdrawInstruction(1_000_000_000, pubkeyOf(1), pubkeyOf(2), pubkeyOf(3))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_Withdraw), data[:4])

	expected := concat(u32LE(Instruction_Withdraw), u64LE(1_000_000_000))
	require.Equal(t, expected, data)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	w := decoded.Impl.(*Withdraw)
	require.Equal(t, uint64(1_000_000_000), *w.Lamports)
}

// The on-chain Vote program's Withdraw account spec is:
//
//  0. [WRITE] vote account to withdraw from
//  1. [WRITE] recipient account
//  2. [SIGNER] withdraw authority
//
// i.e. the withdraw authority is a read-only signer: it authorizes the
// withdraw but is not itself mutated, so it must not carry the writable flag.
func TestAccountFlags_Withdraw(t *testing.T) {
	inst := NewWithdrawInstruction(1, pubkeyOf(1), pubkeyOf(2), pubkeyOf(3))

	vote := inst.GetVoteAccount()
	require.True(t, vote.IsWritable, "vote account must be writable")
	require.False(t, vote.IsSigner, "vote account must not be a signer")

	recipient := inst.GetRecipientAccount()
	require.True(t, recipient.IsWritable, "recipient must be writable")
	require.False(t, recipient.IsSigner, "recipient must not be a signer")

	withdrawer := inst.GetWithdrawAuthorityAccount()
	require.False(t, withdrawer.IsWritable, "withdraw authority must be read-only")
	require.True(t, withdrawer.IsSigner, "withdraw authority must be a signer")
}
