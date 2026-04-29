package stake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstructionIDValues(t *testing.T) {
	require.Equal(t, uint32(0), Instruction_Initialize)
	require.Equal(t, uint32(1), Instruction_Authorize)
	require.Equal(t, uint32(2), Instruction_DelegateStake)
	require.Equal(t, uint32(3), Instruction_Split)
	require.Equal(t, uint32(4), Instruction_Withdraw)
	require.Equal(t, uint32(5), Instruction_Deactivate)
	require.Equal(t, uint32(6), Instruction_SetLockup)
	require.Equal(t, uint32(7), Instruction_Merge)
	require.Equal(t, uint32(8), Instruction_AuthorizeWithSeed)
	require.Equal(t, uint32(9), Instruction_InitializeChecked)
	require.Equal(t, uint32(10), Instruction_AuthorizeChecked)
	require.Equal(t, uint32(11), Instruction_AuthorizeCheckedWithSeed)
	require.Equal(t, uint32(12), Instruction_SetLockupChecked)
	require.Equal(t, uint32(13), Instruction_GetMinimumDelegation)
	require.Equal(t, uint32(14), Instruction_DeactivateDelinquent)
	require.Equal(t, uint32(15), Instruction_Redelegate)
	require.Equal(t, uint32(16), Instruction_MoveStake)
	require.Equal(t, uint32(17), Instruction_MoveLamports)
}
