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

import (
	"encoding/binary"

	bin "github.com/gagliardetto/binary"
)

// EpochScheduleSize is the serialized size, in bytes, of the EpochSchedule
// sysvar (u64 + u64 + bool + u64 + u64).
const EpochScheduleSize = 33

// MinimumSlotsPerEpoch is the minimum number of slots per epoch.
const MinimumSlotsPerEpoch uint64 = 32

// minSlotsPerEpochTrailingZeros == bits.TrailingZeros64(MinimumSlotsPerEpoch).
const minSlotsPerEpochTrailingZeros uint32 = 5

// EpochSchedule is the data of the EpochSchedule sysvar (account
// solana.SysVarEpochSchedulePubkey): the cluster's epoch length configuration and the
// warmup schedule. Mirrors solana-sdk epoch_schedule::EpochSchedule.
type EpochSchedule struct {
	// SlotsPerEpoch is the number of slots in each normal epoch.
	SlotsPerEpoch uint64
	// LeaderScheduleSlotOffset is how many slots before an epoch its leader
	// schedule is calculated.
	LeaderScheduleSlotOffset uint64
	// Warmup reports whether epochs start short and grow to SlotsPerEpoch.
	Warmup bool
	// FirstNormalEpoch is the first epoch with SlotsPerEpoch slots (post-warmup).
	FirstNormalEpoch uint64
	// FirstNormalSlot is the first slot of FirstNormalEpoch.
	FirstNormalSlot uint64
}

// NewEpochScheduleCustom builds an EpochSchedule, computing the warmup
// boundaries. Mirrors EpochSchedule::custom (slotsPerEpoch must be >=
// MinimumSlotsPerEpoch).
func NewEpochScheduleCustom(slotsPerEpoch, leaderScheduleSlotOffset uint64, warmup bool) EpochSchedule {
	var firstNormalEpoch, firstNormalSlot uint64
	if warmup {
		npot := nextPowerOfTwo(slotsPerEpoch)
		log2 := satSubU32(uint32(numTrailingZeros64(npot)), minSlotsPerEpochTrailingZeros)
		firstNormalEpoch = uint64(log2)
		firstNormalSlot = satSubU64(npot, MinimumSlotsPerEpoch)
	}
	return EpochSchedule{
		SlotsPerEpoch:            slotsPerEpoch,
		LeaderScheduleSlotOffset: leaderScheduleSlotOffset,
		Warmup:                   warmup,
		FirstNormalEpoch:         firstNormalEpoch,
		FirstNormalSlot:          firstNormalSlot,
	}
}

// NewEpochSchedule builds an EpochSchedule with warmup enabled and the leader
// schedule offset equal to slotsPerEpoch. Mirrors EpochSchedule::new.
func NewEpochSchedule(slotsPerEpoch uint64) EpochSchedule {
	return NewEpochScheduleCustom(slotsPerEpoch, slotsPerEpoch, true)
}

// GetEpoch returns the epoch containing the given slot.
func (es EpochSchedule) GetEpoch(slot uint64) uint64 {
	epoch, _ := es.GetEpochAndSlotIndex(slot)
	return epoch
}

// GetEpochAndSlotIndex returns the epoch containing slot and the slot's index
// within that epoch. Mirrors EpochSchedule::get_epoch_and_slot_index.
func (es EpochSchedule) GetEpochAndSlotIndex(slot uint64) (epoch, slotIndex uint64) {
	if slot < es.FirstNormalSlot {
		npot := nextPowerOfTwo(satAddU64(satAddU64(slot, MinimumSlotsPerEpoch), 1))
		tz := uint32(numTrailingZeros64(npot))
		ep := satSubU32(satSubU32(tz, minSlotsPerEpochTrailingZeros), 1)
		epochLen := satPow2(satAddU32(ep, minSlotsPerEpochTrailingZeros))
		return uint64(ep), satSubU64(slot, satSubU64(epochLen, MinimumSlotsPerEpoch))
	}
	normalSlotIndex := satSubU64(slot, es.FirstNormalSlot)
	var normalEpochIndex uint64
	if es.SlotsPerEpoch != 0 {
		normalEpochIndex = normalSlotIndex / es.SlotsPerEpoch
		slotIndex = normalSlotIndex % es.SlotsPerEpoch
	}
	return satAddU64(es.FirstNormalEpoch, normalEpochIndex), slotIndex
}

// GetSlotsInEpoch returns the number of slots in the given epoch (accounting for
// warmup). Mirrors EpochSchedule::get_slots_in_epoch.
func (es EpochSchedule) GetSlotsInEpoch(epoch uint64) uint64 {
	if epoch < es.FirstNormalEpoch {
		return satPow2(satAddU32(uint32(epoch), minSlotsPerEpochTrailingZeros))
	}
	return es.SlotsPerEpoch
}

// GetFirstSlotInEpoch returns the first slot of the given epoch.
// Mirrors EpochSchedule::get_first_slot_in_epoch.
func (es EpochSchedule) GetFirstSlotInEpoch(epoch uint64) uint64 {
	if epoch <= es.FirstNormalEpoch {
		return satMulU64(satSubU64(satPow2(uint32(epoch)), 1), MinimumSlotsPerEpoch)
	}
	return satAddU64(satMulU64(satSubU64(epoch, es.FirstNormalEpoch), es.SlotsPerEpoch), es.FirstNormalSlot)
}

// GetLastSlotInEpoch returns the last slot of the given epoch.
func (es EpochSchedule) GetLastSlotInEpoch(epoch uint64) uint64 {
	return satSubU64(satAddU64(es.GetFirstSlotInEpoch(epoch), es.GetSlotsInEpoch(epoch)), 1)
}

// GetLeaderScheduleEpoch returns the epoch whose leader schedule should be
// computed at the given slot. Mirrors EpochSchedule::get_leader_schedule_epoch.
func (es EpochSchedule) GetLeaderScheduleEpoch(slot uint64) uint64 {
	if slot < es.FirstNormalSlot {
		epoch, _ := es.GetEpochAndSlotIndex(slot)
		return satAddU64(epoch, 1)
	}
	newSlots := satSubU64(slot, es.FirstNormalSlot)
	newFirst := satAddU64(newSlots, es.LeaderScheduleSlotOffset)
	var newEpochs uint64
	if es.SlotsPerEpoch != 0 {
		newEpochs = newFirst / es.SlotsPerEpoch
	}
	return satAddU64(es.FirstNormalEpoch, newEpochs)
}

func (es EpochSchedule) MarshalWithEncoder(encoder *bin.Encoder) error {
	if err := encoder.WriteUint64(es.SlotsPerEpoch, binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteUint64(es.LeaderScheduleSlotOffset, binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteBool(es.Warmup); err != nil {
		return err
	}
	if err := encoder.WriteUint64(es.FirstNormalEpoch, binary.LittleEndian); err != nil {
		return err
	}
	return encoder.WriteUint64(es.FirstNormalSlot, binary.LittleEndian)
}

func (es *EpochSchedule) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	var err error
	if es.SlotsPerEpoch, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	if es.LeaderScheduleSlotOffset, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	if es.Warmup, err = decoder.ReadBool(); err != nil {
		return err
	}
	if es.FirstNormalEpoch, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	es.FirstNormalSlot, err = decoder.ReadUint64(binary.LittleEndian)
	return err
}

func (es EpochSchedule) MarshalBinary() ([]byte, error) { return encodeSysvar(es) }
func (es *EpochSchedule) UnmarshalBinary(data []byte) error {
	return es.UnmarshalWithDecoder(bin.NewBinDecoder(data))
}

// DecodeEpochSchedule decodes EpochSchedule sysvar account data.
func DecodeEpochSchedule(data []byte) (*EpochSchedule, error) {
	var es EpochSchedule
	if err := es.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return &es, nil
}
