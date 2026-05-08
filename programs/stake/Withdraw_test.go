package stake

import (
	"testing"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip_Withdraw(t *testing.T) {
	inst := NewWithdrawInstructionBuilder().
		SetStakeAccount(pubkeyOf(1)).
		SetRecipientAccount(pubkeyOf(2)).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetStakeHistorySysvar(solana.SysVarStakeHistoryPubkey).
		SetWithdrawAuthority(pubkeyOf(3)).
		SetLamports(500_000)

	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_Withdraw), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	w := decoded.Impl.(*Withdraw)
	require.Equal(t, uint64(500_000), *w.Lamports)
}
