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
	"fmt"

	bin "github.com/gagliardetto/binary"
)

const (
	// SlotHistoryMaxEntries is the number of slots tracked by the SlotHistory
	// sysvar bitvector (1,048,576 slots, ~5 days).
	SlotHistoryMaxEntries = 1024 * 1024
	// slotHistoryBlocks is the number of u64 blocks backing the bitvector.
	slotHistoryBlocks = SlotHistoryMaxEntries / 64 // 16384
	// SlotHistorySize is the fixed on-chain account size of the SlotHistory
	// sysvar: 1 (Option tag) + 8 (slice len) + blocks*8 + 8 (bit len) + 8 (next_slot).
	SlotHistorySize = 1 + 8 + slotHistoryBlocks*8 + 8 + 8 // 131097
)

// SlotHistoryCheck is the result of SlotHistory.Check.
type SlotHistoryCheck int

const (
	// SlotHistoryFuture means the slot is newer than anything recorded.
	SlotHistoryFuture SlotHistoryCheck = iota
	// SlotHistoryTooOld means the slot has aged out of the window.
	SlotHistoryTooOld
	// SlotHistoryFound means the slot is recorded as present.
	SlotHistoryFound
	// SlotHistoryNotFound means the slot is within the window but absent.
	SlotHistoryNotFound
)

func (c SlotHistoryCheck) String() string {
	switch c {
	case SlotHistoryFuture:
		return "Future"
	case SlotHistoryTooOld:
		return "TooOld"
	case SlotHistoryFound:
		return "Found"
	case SlotHistoryNotFound:
		return "NotFound"
	default:
		return fmt.Sprintf("SlotHistoryCheck(%d)", int(c))
	}
}

// SlotHistory is the data of the SlotHistory sysvar (account
// solana.SysVarSlotHistoryPubkey): a bitvector recording which of the most recent
// SlotHistoryMaxEntries slots were rooted. Mirrors solana-sdk
// slot_history::SlotHistory.
type SlotHistory struct {
	// Bits is the backing bitvector, one bit per slot modulo SlotHistoryMaxEntries
	// (bit i lives at Bits[i/64] bit i%64). A full SlotHistory has
	// slotHistoryBlocks (16384) blocks.
	Bits []uint64
	// NextSlot is one past the newest recorded slot.
	NextSlot uint64
}

// NewSlotHistory returns a SlotHistory in its genesis state: an all-false
// bitvector with slot 0 marked present and NextSlot == 1. Mirrors the upstream
// SlotHistory::default.
func NewSlotHistory() *SlotHistory {
	sh := &SlotHistory{Bits: make([]uint64, slotHistoryBlocks), NextSlot: 1}
	sh.setBit(0, true)
	return sh
}

func (sh *SlotHistory) getBit(i uint64) bool {
	block := i / 64
	if block >= uint64(len(sh.Bits)) {
		return false
	}
	return (sh.Bits[block]>>(i%64))&1 == 1
}

func (sh *SlotHistory) setBit(i uint64, v bool) {
	block := i / 64
	if block >= uint64(len(sh.Bits)) {
		return
	}
	mask := uint64(1) << (i % 64)
	if v {
		sh.Bits[block] |= mask
	} else {
		sh.Bits[block] &^= mask
	}
}

// Newest returns the most recently recorded slot. Mirrors SlotHistory::newest.
// It assumes NextSlot >= 1, which holds for any SlotHistory from NewSlotHistory
// or decoded from a real account.
func (sh *SlotHistory) Newest() uint64 { return sh.NextSlot - 1 }

// Oldest returns the oldest slot still within the window. Mirrors
// SlotHistory::oldest.
func (sh *SlotHistory) Oldest() uint64 { return satSubU64(sh.NextSlot, SlotHistoryMaxEntries) }

// Check reports the status of slot relative to the recorded window. Mirrors
// SlotHistory::check.
func (sh *SlotHistory) Check(slot uint64) SlotHistoryCheck {
	switch {
	case slot > sh.Newest():
		return SlotHistoryFuture
	case slot < sh.Oldest():
		return SlotHistoryTooOld
	case sh.getBit(slot % SlotHistoryMaxEntries):
		return SlotHistoryFound
	default:
		return SlotHistoryNotFound
	}
}

// Add records slot as present. For a slot at or beyond NextSlot it advances the
// window, clearing any skipped slots (mirroring the runtime's SlotHistory::add,
// which assumes monotonically increasing slots).
//
// Unlike the runtime, adding an OLDER slot does not rewind NextSlot: the slot is
// simply marked present if it is still within the window, and ignored if it has
// already aged out. This keeps Check correct for newer slots that were already
// recorded.
func (sh *SlotHistory) Add(slot uint64) {
	if slot < sh.NextSlot {
		if slot >= sh.Oldest() {
			sh.setBit(slot%SlotHistoryMaxEntries, true)
		}
		return
	}
	if slot-sh.NextSlot >= SlotHistoryMaxEntries {
		// Wrapped all the way around: clear everything.
		for i := range sh.Bits {
			sh.Bits[i] = 0
		}
	} else {
		for skipped := sh.NextSlot; skipped < slot; skipped++ {
			sh.setBit(skipped%SlotHistoryMaxEntries, false)
		}
	}
	sh.setBit(slot%SlotHistoryMaxEntries, true)
	sh.NextSlot = slot + 1
}

func (sh SlotHistory) MarshalWithEncoder(encoder *bin.Encoder) error {
	// bits: bv::BitVec<u64> serializes as Option<Box<[u64]>> then the u64 bit length.
	if err := encoder.WriteByte(0x01); err != nil { // Some
		return err
	}
	if err := encoder.WriteUint64(uint64(len(sh.Bits)), binary.LittleEndian); err != nil {
		return err
	}
	for _, block := range sh.Bits {
		if err := encoder.WriteUint64(block, binary.LittleEndian); err != nil {
			return err
		}
	}
	// BitVec bit length: blocks * 64 (== SlotHistoryMaxEntries for a full,
	// well-formed SlotHistory).
	if err := encoder.WriteUint64(uint64(len(sh.Bits))*64, binary.LittleEndian); err != nil {
		return err
	}
	return encoder.WriteUint64(sh.NextSlot, binary.LittleEndian)
}

func (sh *SlotHistory) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	tag, err := decoder.ReadByte()
	if err != nil {
		return err
	}
	switch tag {
	case 0x00:
		sh.Bits = nil
	case 0x01:
		n, err := decoder.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
		blocks := make([]uint64, 0, min(n, uint64(decoder.Remaining()/8)))
		for range n {
			block, err := decoder.ReadUint64(binary.LittleEndian)
			if err != nil {
				return err
			}
			blocks = append(blocks, block)
		}
		sh.Bits = blocks
	default:
		return fmt.Errorf("invalid slot history bitvec option tag: %d", tag)
	}
	// bit length (always SlotHistoryMaxEntries); read and discard.
	if _, err := decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	sh.NextSlot, err = decoder.ReadUint64(binary.LittleEndian)
	return err
}

func (sh SlotHistory) MarshalBinary() ([]byte, error) { return encodeSysvar(sh) }
func (sh *SlotHistory) UnmarshalBinary(data []byte) error {
	return sh.UnmarshalWithDecoder(bin.NewBinDecoder(data))
}

// DecodeSlotHistory decodes SlotHistory sysvar account data.
func DecodeSlotHistory(data []byte) (*SlotHistory, error) {
	var sh SlotHistory
	if err := sh.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return &sh, nil
}
