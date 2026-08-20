package sysvar

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
)

// Mainnet-beta scenario: reduce_slot_time_to_350ms activated at slot
// 440208000 (first slot of epoch 1019) -> effective at epoch 1020.
func TestMsPerSlotAt_Mainnet350ms(t *testing.T) {
	// Mainnet-beta runs without warmup epochs (epoch = slot / 432000).
	es := NewEpochScheduleCustom(DefaultSlotsPerEpoch, DefaultSlotsPerEpoch, false)
	features := solana.FeatureSet{
		solana.FeatureReduceSlotTimeTo350ms: 440208000,
	}

	epoch1020Start := es.GetFirstSlotInEpoch(1020)
	assert.Equal(t, uint64(440640000), epoch1020Start)

	assert.Equal(t, DefaultMsPerSlot, MsPerSlotAt(440208000, es, features)) // activation slot: still 400
	assert.Equal(t, DefaultMsPerSlot, MsPerSlotAt(440464041, es, features)) // mid epoch 1019
	assert.Equal(t, DefaultMsPerSlot, MsPerSlotAt(epoch1020Start-1, es, features))
	assert.Equal(t, MsPerSlot350, MsPerSlotAt(epoch1020Start, es, features)) // effective
	assert.Equal(t, MsPerSlot350, MsPerSlotAt(epoch1020Start+1_000_000, es, features))

	assert.Equal(t, DefaultMsPerSlot, MsPerSlotInEpoch(1019, es, features))
	assert.Equal(t, MsPerSlot350, MsPerSlotInEpoch(1020, es, features))
	assert.Equal(t, MsPerSlot350, MsPerSlotInEpoch(1021, es, features))
}

func TestMsPerSlotAt_NoGates(t *testing.T) {
	es := NewEpochSchedule(DefaultSlotsPerEpoch)
	assert.Equal(t, DefaultMsPerSlot, MsPerSlotAt(0, es, nil))
	assert.Equal(t, DefaultMsPerSlot, MsPerSlotAt(1<<40, es, nil))
	assert.Equal(t, DefaultMsPerSlot, MsPerSlotInEpoch(100, es, nil))
}

// Out-of-order activations must never increase the slot duration (agave parity).
func TestMsPerSlotAt_OutOfOrderNeverIncreases(t *testing.T) {
	es := NewEpochSchedule(DefaultSlotsPerEpoch)
	// 300ms activates in epoch 10, 350ms later in epoch 20.
	features := solana.FeatureSet{
		solana.FeatureReduceSlotTimeTo300ms: es.GetFirstSlotInEpoch(10),
		solana.FeatureReduceSlotTimeTo350ms: es.GetFirstSlotInEpoch(20),
	}

	assert.Equal(t, DefaultMsPerSlot, MsPerSlotInEpoch(10, es, features))
	assert.Equal(t, MsPerSlot300, MsPerSlotInEpoch(11, es, features))
	// The later, longer gate must not bump it back up to 350.
	assert.Equal(t, MsPerSlot300, MsPerSlotInEpoch(21, es, features))

	prev := DefaultMsPerSlot
	for epoch := uint64(0); epoch < 30; epoch++ {
		ms := MsPerSlotInEpoch(epoch, es, features)
		assert.LessOrEqual(t, ms, prev, "epoch %d", epoch)
		prev = ms
	}
}

// Degenerate inputs must fail safe (no reduction claimed).
func TestMsPerSlotAt_DegenerateInputs(t *testing.T) {
	features := solana.FeatureSet{
		solana.FeatureReduceSlotTimeTo200ms: 1,
	}

	// Zero-value schedule: gates must not be reported effective from slot 0.
	assert.Equal(t, DefaultMsPerSlot, MsPerSlotAt(1<<40, EpochSchedule{}, features))

	// Activation epoch near the maximum: the epoch+1 hop saturates instead
	// of wrapping to epoch 0 (which would make the gate always effective).
	es := NewEpochSchedule(DefaultSlotsPerEpoch)
	huge := solana.FeatureSet{
		solana.FeatureReduceSlotTimeTo200ms: ^uint64(0),
	}
	assert.Equal(t, DefaultMsPerSlot, MsPerSlotAt(0, es, huge))
	assert.Equal(t, DefaultMsPerSlot, MsPerSlotAt(1<<40, es, huge))
}

func TestSlotTimeGates_Order(t *testing.T) {
	gates := SlotTimeGates()
	assert.Len(t, gates, 4)
	assert.Equal(t, MsPerSlot350, gates[0].MsPerSlot)
	assert.Equal(t, MsPerSlot200, gates[3].MsPerSlot)
	for i := 1; i < len(gates); i++ {
		assert.Less(t, gates[i].MsPerSlot, gates[i-1].MsPerSlot)
	}

	// The returned slice is a copy; callers cannot corrupt the gate table.
	gates[0].MsPerSlot = 1
	assert.Equal(t, MsPerSlot350, SlotTimeGates()[0].MsPerSlot)
}
