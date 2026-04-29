package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstructionIDValues(t *testing.T) {
	require.Equal(t, uint32(0), Instruction_InitializeAccount)
	require.Equal(t, uint32(1), Instruction_Authorize)
	require.Equal(t, uint32(2), Instruction_Vote)
	require.Equal(t, uint32(3), Instruction_Withdraw)
	require.Equal(t, uint32(4), Instruction_UpdateValidatorIdentity)
	require.Equal(t, uint32(5), Instruction_UpdateCommission)
	require.Equal(t, uint32(6), Instruction_VoteSwitch)
	require.Equal(t, uint32(7), Instruction_AuthorizeChecked)
	require.Equal(t, uint32(8), Instruction_UpdateVoteState)
	require.Equal(t, uint32(9), Instruction_UpdateVoteStateSwitch)
	require.Equal(t, uint32(10), Instruction_AuthorizeWithSeed)
	require.Equal(t, uint32(11), Instruction_AuthorizeCheckedWithSeed)
	require.Equal(t, uint32(12), Instruction_CompactUpdateVoteState)
	require.Equal(t, uint32(13), Instruction_CompactUpdateVoteStateSwitch)
	require.Equal(t, uint32(14), Instruction_TowerSync)
	require.Equal(t, uint32(15), Instruction_TowerSyncSwitch)
	require.Equal(t, uint32(16), Instruction_InitializeAccountV2)
	require.Equal(t, uint32(17), Instruction_UpdateCommissionCollector)
	require.Equal(t, uint32(18), Instruction_UpdateCommissionBps)
	require.Equal(t, uint32(19), Instruction_DepositDelegatorRewards)
}

func TestInstructionIDToName(t *testing.T) {
	require.Equal(t, "InitializeAccount", InstructionIDToName(0))
	require.Equal(t, "TowerSync", InstructionIDToName(14))
	require.Equal(t, "DepositDelegatorRewards", InstructionIDToName(19))
	require.Equal(t, "", InstructionIDToName(99))
}
