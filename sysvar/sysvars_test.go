// Copyright 2021 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sysvar

// Tests ported from the upstream anza-xyz/solana-sdk sysvar crates to match the
// same functionality, one Go test per upstream test:
//   clock/src/lib.rs           -> TestClockSizeOf, TestClockClone
//   rent/src/lib.rs            -> TestRentSizeOf, TestRentClone, TestRentExemptionThreshold, TestRentMinimumBalance
//   epoch-schedule/src/lib.rs  -> TestEpochSchedule, TestEpochScheduleClone
//   slot-hashes/src/lib.rs     -> TestSlotHashesSizeOf, TestSlotHashes
//   slot-history/src/lib.rs    -> TestSlotHistory (slot_history_test1)
//   stake-history/src/lib.rs   -> TestStakeHistory
//   epoch-rewards/src/lib.rs   -> TestEpochRewardsSizeOf, TestEpochRewardsNew, TestEpochRewardsDistribute, TestEpochRewardsDistributePanic
//   last-restart-slot/src/lib.rs -> TestLastRestartSlotSizeOf
//   sysvar/src/fees.rs         -> TestFeesSizeOf, TestFeesClone
//   sysvar/src/recent_blockhashes.rs -> TestRecentBlockhashesSizeOf, TestRecentBlockhashesCanHoldAllActiveBlockhashes
// Plus Go-specific extras: live-mainnet golden vectors for the static sysvars
// (Rent, EpochSchedule) and encode/decode round-trips.

import (
	"encoding/hex"
	"math"
	"testing"

	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func filledHash(b byte) solana.Hash {
	var h solana.Hash
	for i := range h {
		h[i] = b
	}
	return h
}

// ============================ Clock ============================

func TestClockSizeOf(t *testing.T) {
	raw, err := Clock{}.MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, raw, ClockSize)
}

func TestClockClone(t *testing.T) {
	c := Clock{Slot: 1, EpochStartTimestamp: 2, Epoch: 3, LeaderScheduleEpoch: 4, UnixTimestamp: 5}
	raw, err := c.MarshalBinary()
	require.NoError(t, err)
	// 5 x 8-byte little-endian fields.
	assert.Equal(t,
		mustHex(t, "01000000000000000200000000000000030000000000000004000000000000000500000000000000"),
		raw)
	got, err := DecodeClock(raw)
	require.NoError(t, err)
	assert.Equal(t, c, *got)

	// Negative i64 timestamps round-trip.
	c2 := Clock{Slot: 100, EpochStartTimestamp: -7, Epoch: 9, UnixTimestamp: -1}
	raw2, err := c2.MarshalBinary()
	require.NoError(t, err)
	got2, err := DecodeClock(raw2)
	require.NoError(t, err)
	assert.Equal(t, c2, *got2)
}

// ============================ Rent ============================

func TestRentSizeOf(t *testing.T) {
	raw, err := DefaultRent().MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, raw, RentSize)
}

func TestRentClone(t *testing.T) {
	r := Rent{LamportsPerByte: 1, ExemptionThreshold: 2.2, BurnPercent: 3}
	raw, err := r.MarshalBinary()
	require.NoError(t, err)
	got, err := DecodeRent(raw)
	require.NoError(t, err)
	assert.Equal(t, r, *got)
}

// TestRentExemptionThreshold ports rent::test_exemption_threshold: the wire form
// of the threshold is the f64 little-endian byte pattern.
func TestRentExemptionThreshold(t *testing.T) {
	wireBytes := func(v float64) []byte {
		raw, err := Rent{ExemptionThreshold: v}.MarshalBinary()
		require.NoError(t, err)
		return raw[8:16]
	}
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0xf0, 0x3f}, wireBytes(1.0)) // SIMD-0194 default
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0x40}, wireBytes(2.0))    // legacy default
}

// TestRentMinimumBalance ports rent::test_minimum_balance: the SIMD-0194 default
// (6960 lamports/byte, threshold 1.0) yields the same minimum as the legacy
// parameters (3480 lamports/byte, threshold 2.0), equal to the float formula.
func TestRentMinimumBalance(t *testing.T) {
	def := DefaultRent()
	prev := Rent{LamportsPerByte: DefaultLamportsPerByte / 2, ExemptionThreshold: 2.0, BurnPercent: DefaultBurnPercent}
	for _, bytes := range []uint64{0, 1, 165, 1000, MaxPermittedDataLength} {
		defCalc := def.MinimumBalance(bytes)
		assert.Equal(t, defCalc, prev.MinimumBalance(bytes), "bytes=%d", bytes)
		floatCalc := uint64(float64((AccountStorageOverhead+bytes)*prev.LamportsPerByte) * 2.0)
		assert.Equal(t, floatCalc, defCalc, "bytes=%d", bytes)
	}
	// The famous 0-byte rent-exempt minimum.
	assert.Equal(t, uint64(890880), def.MinimumBalance(0))
	assert.True(t, def.IsExempt(890880, 0))
	assert.False(t, def.IsExempt(890879, 0))
}

func TestRentGoldenVector(t *testing.T) {
	// Live mainnet Rent sysvar account data (static; SIMD-0194 form).
	data := mustHex(t, "301b000000000000000000000000f03f32")
	require.Len(t, data, RentSize)
	r, err := DecodeRent(data)
	require.NoError(t, err)
	assert.Equal(t, uint64(6960), r.LamportsPerByte)
	assert.Equal(t, 1.0, r.ExemptionThreshold)
	assert.Equal(t, uint8(50), r.BurnPercent)
	assert.Equal(t, DefaultRent(), *r)
	raw, err := r.MarshalBinary()
	require.NoError(t, err)
	assert.Equal(t, data, raw)
}

// ============================ EpochSchedule ============================

// TestEpochSchedule ports epoch_schedule::test_epoch_schedule.
func TestEpochSchedule(t *testing.T) {
	for spe := MinimumSlotsPerEpoch; spe <= MinimumSlotsPerEpoch*16; spe++ {
		es := NewEpochScheduleCustom(spe, spe/2, true)

		require.Equal(t, uint64(0), es.GetFirstSlotInEpoch(0), "spe=%d", spe)
		require.Equal(t, MinimumSlotsPerEpoch-1, es.GetLastSlotInEpoch(0), "spe=%d", spe)

		var lastLeaderSchedule, lastEpoch, lastSlotsInEpoch uint64 = 0, 0, MinimumSlotsPerEpoch
		for slot := uint64(0); slot < 2*spe; slot++ {
			// leader_schedule_epoch is continuous over warmup into the first normal epoch.
			ls := es.GetLeaderScheduleEpoch(slot)
			if ls != lastLeaderSchedule {
				if ls != lastLeaderSchedule+1 {
					t.Fatalf("spe=%d slot=%d: leader schedule jumped %d->%d", spe, slot, lastLeaderSchedule, ls)
				}
				lastLeaderSchedule = ls
			}

			epoch, offset := es.GetEpochAndSlotIndex(slot)
			if epoch != lastEpoch {
				if epoch != lastEpoch+1 {
					t.Fatalf("spe=%d slot=%d: epoch jumped %d->%d", spe, slot, lastEpoch, epoch)
				}
				lastEpoch = epoch
				if got := es.GetFirstSlotInEpoch(epoch); got != slot {
					t.Fatalf("spe=%d: GetFirstSlotInEpoch(%d)=%d, want %d", spe, epoch, got, slot)
				}
				if got := es.GetLastSlotInEpoch(epoch - 1); got != slot-1 {
					t.Fatalf("spe=%d: GetLastSlotInEpoch(%d)=%d, want %d", spe, epoch-1, got, slot-1)
				}
				// slots per epoch double over warmup until they reach spe.
				slotsInEpoch := es.GetSlotsInEpoch(epoch)
				if slotsInEpoch != lastSlotsInEpoch && slotsInEpoch != spe && slotsInEpoch != lastSlotsInEpoch*2 {
					t.Fatalf("spe=%d epoch=%d: slots_in_epoch=%d, want %d or %d", spe, epoch, slotsInEpoch, lastSlotsInEpoch*2, spe)
				}
				lastSlotsInEpoch = slotsInEpoch
			}
			if offset >= lastSlotsInEpoch {
				t.Fatalf("spe=%d slot=%d: offset %d >= slots_in_epoch %d", spe, slot, offset, lastSlotsInEpoch)
			}
		}
		require.NotZero(t, lastLeaderSchedule, "spe=%d", spe)
		require.NotZero(t, lastEpoch, "spe=%d", spe)
		require.Equal(t, spe, lastSlotsInEpoch, "spe=%d: never reached normal mode", spe)
	}
}

func TestEpochScheduleClone(t *testing.T) {
	es := EpochSchedule{SlotsPerEpoch: 1, LeaderScheduleSlotOffset: 2, Warmup: true, FirstNormalEpoch: 4, FirstNormalSlot: 5}
	raw, err := es.MarshalBinary()
	require.NoError(t, err)
	got, err := DecodeEpochSchedule(raw)
	require.NoError(t, err)
	assert.Equal(t, es, *got)
}

func TestEpochScheduleGoldenVector(t *testing.T) {
	// Live mainnet EpochSchedule sysvar account data (static, no warmup).
	data := mustHex(t, "809706000000000080970600000000000000000000000000000000000000000000")
	require.Len(t, data, EpochScheduleSize)
	es, err := DecodeEpochSchedule(data)
	require.NoError(t, err)
	assert.Equal(t, EpochSchedule{SlotsPerEpoch: 432000, LeaderScheduleSlotOffset: 432000}, *es)
	assert.Equal(t, uint64(0), es.GetEpoch(431999))
	assert.Equal(t, uint64(1), es.GetEpoch(432000))
	assert.Equal(t, uint64(864000), es.GetFirstSlotInEpoch(2))
	raw, err := es.MarshalBinary()
	require.NoError(t, err)
	assert.Equal(t, data, raw)
}

// ============================ SlotHashes ============================

func TestSlotHashesSizeOf(t *testing.T) {
	sh := make(SlotHashes, SlotHashesMaxEntries)
	raw, err := sh.MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, raw, SlotHashesSize)
}

// TestSlotHashes ports slot_hashes::tests::test.
func TestSlotHashes(t *testing.T) {
	sh := NewSlotHashes(SlotHash{Slot: 1}, SlotHash{Slot: 3})
	sh.Add(2, solana.Hash{})
	assert.Equal(t, SlotHashes{{Slot: 3}, {Slot: 2}, {Slot: 1}}, sh)

	var sh2 SlotHashes
	for i := uint64(0); i < SlotHashesMaxEntries+1; i++ {
		sh2.Add(i, filledHash(byte(i)))
	}
	require.Len(t, sh2, SlotHashesMaxEntries)
	for i := 0; i < SlotHashesMaxEntries; i++ {
		require.Equal(t, uint64(SlotHashesMaxEntries-i), sh2[i].Slot, "i=%d", i)
	}

	// Round-trip.
	raw, err := sh2.MarshalBinary()
	require.NoError(t, err)
	got, err := DecodeSlotHashes(raw)
	require.NoError(t, err)
	assert.Equal(t, sh2, got)
}

// ============================ SlotHistory ============================

// checkRange asserts sh.Check(i) == want for all i in [lo, hi), failing fast.
func checkRange(t *testing.T, sh *SlotHistory, lo, hi uint64, want SlotHistoryCheck) {
	t.Helper()
	for i := lo; i < hi; i++ {
		if got := sh.Check(i); got != want {
			t.Fatalf("Check(%d) = %v, want %v", i, got, want)
		}
	}
}

// TestSlotHistory ports slot_history::slot_history_test1.
func TestSlotHistory(t *testing.T) {
	require.Zero(t, SlotHistoryMaxEntries%64) // clear logic works on 64-bit blocks

	sh := NewSlotHistory()
	sh.Add(2)
	assert.Equal(t, SlotHistoryFound, sh.Check(0))
	assert.Equal(t, SlotHistoryNotFound, sh.Check(1))
	checkRange(t, sh, 3, SlotHistoryMaxEntries, SlotHistoryFuture)

	sh.Add(20)
	sh.Add(SlotHistoryMaxEntries)
	assert.Equal(t, SlotHistoryTooOld, sh.Check(0))
	assert.Equal(t, SlotHistoryNotFound, sh.Check(1))
	for _, i := range []uint64{2, 20, SlotHistoryMaxEntries} {
		assert.Equal(t, SlotHistoryFound, sh.Check(i), "i=%d", i)
	}
	checkRange(t, sh, 3, 20, SlotHistoryNotFound)
	checkRange(t, sh, 21, SlotHistoryMaxEntries, SlotHistoryNotFound)
	assert.Equal(t, SlotHistoryFuture, sh.Check(SlotHistoryMaxEntries+1))

	slot := uint64(3*SlotHistoryMaxEntries + 3)
	sh.Add(slot)
	for _, i := range []uint64{0, 1, 2, 20, 21, SlotHistoryMaxEntries} {
		assert.Equal(t, SlotHistoryTooOld, sh.Check(i), "i=%d", i)
	}
	checkRange(t, sh, slot-SlotHistoryMaxEntries+1, slot, SlotHistoryNotFound)
	assert.Equal(t, SlotHistoryFound, sh.Check(slot))

	// Serialized account is the fixed sysvar size and round-trips.
	raw, err := sh.MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, raw, SlotHistorySize)
	got, err := DecodeSlotHistory(raw)
	require.NoError(t, err)
	assert.Equal(t, sh.NextSlot, got.NextSlot)
	assert.Equal(t, sh.Bits, got.Bits)
}

// ============================ StakeHistory ============================

// TestStakeHistory ports stake_history::test_stake_history.
func TestStakeHistory(t *testing.T) {
	var sh StakeHistory
	currentEpoch := uint64(StakeHistoryMaxEntries) + 1
	for i := uint64(0); i < currentEpoch; i++ {
		sh.Add(i, StakeHistoryEntry{Activating: i})
	}
	require.Len(t, sh, StakeHistoryMaxEntries)
	assert.Equal(t, uint64(1), sh[len(sh)-1].Epoch) // oldest retained epoch

	_, ok := sh.Get(0)
	assert.False(t, ok)
	for epoch := uint64(1); epoch < currentEpoch; epoch++ {
		e, ok := sh.Get(epoch)
		require.True(t, ok, "epoch=%d", epoch)
		assert.Equal(t, StakeHistoryEntry{Activating: epoch}, e)
	}
	_, ok = sh.Get(currentEpoch)
	assert.False(t, ok)

	// Round-trip a small history.
	small := StakeHistory{{Epoch: 9, Entry: StakeHistoryEntry{Effective: 100, Activating: 10, Deactivating: 1}}}
	raw, err := small.MarshalBinary()
	require.NoError(t, err)
	got, err := DecodeStakeHistory(raw)
	require.NoError(t, err)
	assert.Equal(t, small, got)
}

// ============================ EpochRewards ============================

func TestEpochRewardsSizeOf(t *testing.T) {
	raw, err := EpochRewards{}.MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, raw, EpochRewardsSize)
}

// TestEpochRewardsNew ports epoch_rewards::test_epoch_rewards_new.
func TestEpochRewardsNew(t *testing.T) {
	e := EpochRewards{TotalRewards: 100, DistributedRewards: 0, DistributionStartingBlockHeight: 64}
	assert.Equal(t, uint64(100), e.TotalRewards)
	assert.Equal(t, uint64(0), e.DistributedRewards)
	assert.Equal(t, uint64(64), e.DistributionStartingBlockHeight)
}

// TestEpochRewardsDistribute ports epoch_rewards::test_epoch_rewards_distribute.
func TestEpochRewardsDistribute(t *testing.T) {
	e := EpochRewards{TotalRewards: 100, DistributionStartingBlockHeight: 64}
	require.NoError(t, e.Distribute(100))
	assert.Equal(t, uint64(100), e.TotalRewards)
	assert.Equal(t, uint64(100), e.DistributedRewards)
}

// TestEpochRewardsDistributeError ports epoch_rewards::test_epoch_rewards_distribute_panic
// (the Go port returns an error rather than panicking).
func TestEpochRewardsDistributeError(t *testing.T) {
	e := EpochRewards{TotalRewards: 100, DistributionStartingBlockHeight: 64}
	assert.Error(t, e.Distribute(200))
	assert.Zero(t, e.DistributedRewards) // unchanged on error
}

func TestEpochRewardsRoundTrip(t *testing.T) {
	e := EpochRewards{
		DistributionStartingBlockHeight: 1000,
		NumPartitions:                   8,
		ParentBlockhash:                 filledHash(0x11),
		TotalPoints:                     bin.Uint128{Lo: 0xdeadbeef, Hi: 1},
		TotalRewards:                    5_000_000,
		DistributedRewards:              1_000_000,
		Active:                          true,
	}
	raw, err := e.MarshalBinary()
	require.NoError(t, err)
	got, err := DecodeEpochRewards(raw)
	require.NoError(t, err)
	assert.Equal(t, e.ParentBlockhash, got.ParentBlockhash)
	assert.Equal(t, e.TotalPoints.Lo, got.TotalPoints.Lo)
	assert.Equal(t, e.TotalPoints.Hi, got.TotalPoints.Hi)
	assert.Equal(t, e.TotalRewards, got.TotalRewards)
	assert.True(t, got.Active)
}

// ============================ LastRestartSlot ============================

func TestLastRestartSlotSizeOf(t *testing.T) {
	raw, err := LastRestartSlot{}.MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, raw, LastRestartSlotSize)

	// round-trip
	raw, err = LastRestartSlot{LastRestartSlot: 123456789}.MarshalBinary()
	require.NoError(t, err)
	got, err := DecodeLastRestartSlot(raw)
	require.NoError(t, err)
	assert.Equal(t, uint64(123456789), got.LastRestartSlot)
}

// ============================ Fees (deprecated) ============================

func TestFeesSizeOf(t *testing.T) {
	raw, err := Fees{}.MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, raw, FeesSize)
}

func TestFeesClone(t *testing.T) {
	f := Fees{FeeCalculator: solana.FeeCalculator{LamportsPerSignature: 1}}
	raw, err := f.MarshalBinary()
	require.NoError(t, err)
	got, err := DecodeFees(raw)
	require.NoError(t, err)
	assert.Equal(t, f, *got)
}

// ============================ RecentBlockhashes (deprecated) ============================

// TestRecentBlockhashesSizeOf ports recent_blockhashes::test_size_of.
func TestRecentBlockhashesSizeOf(t *testing.T) {
	rb := make(RecentBlockhashes, RecentBlockhashesMaxEntries)
	raw, err := rb.MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, raw, RecentBlockhashesSize)
}

// TestRecentBlockhashesCanHoldAllActiveBlockhashes ports
// recent_blockhashes::test_sysvar_can_hold_all_active_blockhashes.
func TestRecentBlockhashesCanHoldAllActiveBlockhashes(t *testing.T) {
	assert.True(t, MaxProcessingAge <= RecentBlockhashesMaxEntries)
}

func TestRecentBlockhashesRoundTrip(t *testing.T) {
	rb := RecentBlockhashes{
		{Blockhash: filledHash(0x01), FeeCalculator: solana.FeeCalculator{LamportsPerSignature: 5000}},
		{Blockhash: filledHash(0x02), FeeCalculator: solana.FeeCalculator{LamportsPerSignature: 4000}},
	}
	raw, err := rb.MarshalBinary()
	require.NoError(t, err)
	got, err := DecodeRecentBlockhashes(raw)
	require.NoError(t, err)
	assert.Equal(t, rb, got)
}

// ============================ size constants ============================

// FuzzDecodeSysvars ensures no sysvar decoder panics on arbitrary input
// (e.g. a hostile u64 length prefix must not blow up the allocator).
func FuzzDecodeSysvars(f *testing.F) {
	for _, s := range []string{
		"",
		"301b000000000000000000000000f03f32",
		"809706000000000080970600000000000000000000000000000000000000000000",
		"01000000000000000200000000000000030000000000000004000000000000000500000000000000",
		"ffffffffffffffff", // hostile Vec length prefix
	} {
		b, _ := hex.DecodeString(s)
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeClock(data)
		_, _ = DecodeRent(data)
		_, _ = DecodeEpochSchedule(data)
		_, _ = DecodeEpochRewards(data)
		_, _ = DecodeLastRestartSlot(data)
		_, _ = DecodeSlotHashes(data)
		_, _ = DecodeSlotHistory(data)
		_, _ = DecodeStakeHistory(data)
		_, _ = DecodeFees(data)
		_, _ = DecodeRecentBlockhashes(data)
	})
}

// ============================ regression tests ============================

// TestRentMinimumBalanceOverflow guards against the unchecked multiply wrapping
// and making an underfunded account look rent-exempt.
func TestRentMinimumBalanceOverflow(t *testing.T) {
	r := Rent{LamportsPerByte: math.MaxUint64, ExemptionThreshold: 1.0}
	assert.Equal(t, uint64(math.MaxUint64), r.MinimumBalance(0)) // saturates, never wraps low
	assert.False(t, r.IsExempt(1_000_000_000, 0))                // underfunded => not exempt
	_, ok := r.TryMinimumBalance(0)
	assert.False(t, ok)

	// The 2.0-threshold doubling overflow also saturates rather than wrapping.
	r2 := Rent{LamportsPerByte: math.MaxUint64 / 200, ExemptionThreshold: 2.0}
	assert.Equal(t, uint64(math.MaxUint64), r2.MinimumBalance(0))
	assert.False(t, r2.IsExempt(math.MaxUint64-1, 0))
}

// TestSlotHistoryAddOlderSlot guards against an older Add rewinding NextSlot and
// hiding newer recorded slots.
func TestSlotHistoryAddOlderSlot(t *testing.T) {
	sh := NewSlotHistory()
	sh.Add(100)
	sh.Add(101)
	sh.Add(102)
	require.Equal(t, uint64(103), sh.NextSlot)

	sh.Add(50) // older, still within the window
	assert.Equal(t, uint64(103), sh.NextSlot, "NextSlot must not rewind")
	assert.Equal(t, SlotHistoryFound, sh.Check(102))
	assert.Equal(t, SlotHistoryFound, sh.Check(101))
	assert.Equal(t, SlotHistoryFound, sh.Check(50)) // the older slot is now recorded
}

// TestDecodeSlotHashesNormalizesOrder guards against an out-of-order buffer
// breaking the binary-search lookups.
func TestDecodeSlotHashesNormalizesOrder(t *testing.T) {
	ascending := SlotHashes{{Slot: 1, Hash: filledHash(1)}, {Slot: 2, Hash: filledHash(2)}, {Slot: 3, Hash: filledHash(3)}}
	raw, err := ascending.MarshalBinary()
	require.NoError(t, err)
	got, err := DecodeSlotHashes(raw)
	require.NoError(t, err)
	assert.Equal(t, []uint64{3, 2, 1}, []uint64{got[0].Slot, got[1].Slot, got[2].Slot})
	h, ok := got.Get(2)
	require.True(t, ok)
	assert.Equal(t, filledHash(2), h)
}

// TestDecodeStakeHistoryNormalizesOrder is the StakeHistory analog.
func TestDecodeStakeHistoryNormalizesOrder(t *testing.T) {
	ascending := StakeHistory{
		{Epoch: 1, Entry: StakeHistoryEntry{Effective: 10}},
		{Epoch: 2, Entry: StakeHistoryEntry{Effective: 20}},
		{Epoch: 3, Entry: StakeHistoryEntry{Effective: 30}},
	}
	raw, err := ascending.MarshalBinary()
	require.NoError(t, err)
	got, err := DecodeStakeHistory(raw)
	require.NoError(t, err)
	assert.Equal(t, []uint64{3, 2, 1}, []uint64{got[0].Epoch, got[1].Epoch, got[2].Epoch})
	e, ok := got.Get(2)
	require.True(t, ok)
	assert.Equal(t, StakeHistoryEntry{Effective: 20}, e)
}

func TestSysvarSizeConstants(t *testing.T) {
	assert.Equal(t, 40, ClockSize)
	assert.Equal(t, 17, RentSize)
	assert.Equal(t, 33, EpochScheduleSize)
	assert.Equal(t, 81, EpochRewardsSize)
	assert.Equal(t, 8, LastRestartSlotSize)
	assert.Equal(t, 8, FeesSize)
	assert.Equal(t, 20488, SlotHashesSize)
	assert.Equal(t, 16392, StakeHistorySize)
	assert.Equal(t, 131097, SlotHistorySize)
	assert.Equal(t, 6008, RecentBlockhashesSize)
}
