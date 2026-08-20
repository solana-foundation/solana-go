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

// ClockSize is the serialized size, in bytes, of the Clock sysvar (5 x u64/i64).
const ClockSize = 40

// Cluster timing constants from the solana-sdk clock crate.
const (
	DefaultTicksPerSecond  uint64 = 160
	DefaultTicksPerSlot    uint64 = 64
	DefaultHashesPerSecond uint64 = 10_000_000
	SecondsPerDay          uint64 = 24 * 60 * 60
	TicksPerDay            uint64 = DefaultTicksPerSecond * SecondsPerDay
	// DefaultSlotsPerEpoch is 432000 slots (~2 days at the default 400ms
	// slot time; less as SIMD-0525 reductions activate).
	DefaultSlotsPerEpoch uint64 = 2 * TicksPerDay / DefaultTicksPerSlot
	// DefaultMsPerSlot is the default/target slot duration in milliseconds (400).
	// SIMD-0525 feature gates can reduce the actual duration (e.g. 350ms on
	// mainnet-beta starting at epoch 1020); use MsPerSlotAt for the effective value.
	DefaultMsPerSlot uint64 = 1000 * DefaultTicksPerSlot / DefaultTicksPerSecond
	// MaxRecentBlockhashes is the number of blockhashes kept in the
	// RecentBlockhashes sysvar (300).
	MaxRecentBlockhashes uint64 = 120 * DefaultTicksPerSecond / DefaultTicksPerSlot
	// MaxProcessingAge is the number of blocks a blockhash is valid for (150).
	MaxProcessingAge uint64 = MaxRecentBlockhashes / 2
)

// Clock is the data of the Clock sysvar (account solana.SysVarClockPubkey). It records
// the cluster's notion of time. Mirrors solana-sdk clock::Clock.
type Clock struct {
	// Slot is the current slot.
	Slot uint64
	// EpochStartTimestamp is the Unix timestamp (seconds) of the first slot in
	// the current epoch. Signed.
	EpochStartTimestamp int64
	// Epoch is the current epoch.
	Epoch uint64
	// LeaderScheduleEpoch is the most recent epoch for which the leader schedule
	// has been calculated.
	LeaderScheduleEpoch uint64
	// UnixTimestamp is the validator-estimated Unix timestamp (seconds) of the
	// current slot. Signed.
	UnixTimestamp int64
}

func (c Clock) MarshalWithEncoder(encoder *bin.Encoder) error {
	if err := encoder.WriteUint64(c.Slot, binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteUint64(uint64(c.EpochStartTimestamp), binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteUint64(c.Epoch, binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteUint64(c.LeaderScheduleEpoch, binary.LittleEndian); err != nil {
		return err
	}
	return encoder.WriteUint64(uint64(c.UnixTimestamp), binary.LittleEndian)
}

func (c *Clock) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	var err error
	if c.Slot, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	v, err := decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	c.EpochStartTimestamp = int64(v)
	if c.Epoch, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	if c.LeaderScheduleEpoch, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	if v, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	c.UnixTimestamp = int64(v)
	return nil
}

func (c Clock) MarshalBinary() ([]byte, error) { return encodeSysvar(c) }
func (c *Clock) UnmarshalBinary(data []byte) error {
	return c.UnmarshalWithDecoder(bin.NewBinDecoder(data))
}

// DecodeClock decodes Clock sysvar account data.
func DecodeClock(data []byte) (*Clock, error) {
	var c Clock
	if err := c.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return &c, nil
}
