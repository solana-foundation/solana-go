package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_TowerSync(t *testing.T) {
	root := uint64(50)
	sync := TowerSyncUpdate{
		Lockouts: []Lockout{
			{Slot: 100, ConfirmationCount: 3},
			{Slot: 101, ConfirmationCount: 2},
		},
		Root:    &root,
		Hash:    hashOf(0xAA),
		BlockID: hashOf(0xBB),
	}
	inst := NewTowerSyncInstruction(sync, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_TowerSync), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	ts := decoded.Impl.(*TowerSync)
	require.Len(t, ts.Sync.Lockouts, 2)
	require.Equal(t, hashOf(0xAA), ts.Sync.Hash)
	require.Equal(t, hashOf(0xBB), ts.Sync.BlockID)
	require.NotNil(t, ts.Sync.Root)
	require.Equal(t, uint64(50), *ts.Sync.Root)
}

func TestRoundTrip_TowerSyncSwitch(t *testing.T) {
	sync := TowerSyncUpdate{
		Lockouts: []Lockout{{Slot: 1, ConfirmationCount: 1}},
		Hash:     hashOf(0x11),
		BlockID:  hashOf(0x22),
	}
	proof := hashOf(0x33)
	inst := NewTowerSyncSwitchInstruction(sync, proof, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_TowerSyncSwitch), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	tss := decoded.Impl.(*TowerSyncSwitch)
	require.Equal(t, proof, tss.Hash)
	require.Equal(t, hashOf(0x22), tss.Sync.BlockID)
}
