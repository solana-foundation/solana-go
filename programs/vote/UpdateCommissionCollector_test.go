package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_UpdateCommissionCollector(t *testing.T) {
	inst := NewUpdateCommissionCollectorInstruction(CommissionKindBlockRevenue, pubkeyOf(1), pubkeyOf(2), pubkeyOf(3))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_UpdateCommissionCollector), data[:4])

	expected := concat(u32LE(Instruction_UpdateCommissionCollector), []byte{uint8(CommissionKindBlockRevenue)})
	require.Equal(t, expected, data)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	u := decoded.Impl.(*UpdateCommissionCollector)
	require.Equal(t, CommissionKindBlockRevenue, *u.Kind)
}

func TestRoundTrip_UpdateCommissionBps(t *testing.T) {
	inst := NewUpdateCommissionBpsInstruction(500, CommissionKindInflationRewards, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_UpdateCommissionBps), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	u := decoded.Impl.(*UpdateCommissionBps)
	require.Equal(t, uint16(500), *u.CommissionBps)
	require.Equal(t, CommissionKindInflationRewards, *u.Kind)
}

func TestRoundTrip_DepositDelegatorRewards(t *testing.T) {
	inst := NewDepositDelegatorRewardsInstruction(5_000_000, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_DepositDelegatorRewards), data[:4])

	expected := concat(u32LE(Instruction_DepositDelegatorRewards), u64LE(5_000_000))
	require.Equal(t, expected, data)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	d := decoded.Impl.(*DepositDelegatorRewards)
	require.Equal(t, uint64(5_000_000), *d.Deposit)
}
