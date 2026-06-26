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
	"sort"

	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
)

const (
	// SlotHashesMaxEntries is the maximum number of (slot, hash) pairs the
	// SlotHashes sysvar holds (the most recent 512 slots).
	SlotHashesMaxEntries = 512
	// SlotHashSerializedSize is the serialized size of one SlotHash (u64 + Hash).
	SlotHashSerializedSize = 8 + 32
	// SlotHashesSize is the fixed on-chain account size of the SlotHashes sysvar
	// (u64 length prefix + MAX_ENTRIES * SlotHashSerializedSize).
	SlotHashesSize = 8 + SlotHashesMaxEntries*SlotHashSerializedSize
)

// SlotHash is one (slot, hash) entry of the SlotHashes sysvar.
type SlotHash struct {
	Slot uint64
	Hash solana.Hash
}

// SlotHashes is the data of the SlotHashes sysvar (account solana.SysVarSlotHashesPubkey):
// the hashes of the most recent slots. Entries are kept sorted by slot
// DESCENDING (index 0 is the newest slot). Mirrors solana-sdk slot_hashes::SlotHashes.
type SlotHashes []SlotHash

// NewSlotHashes builds a SlotHashes from entries, sorting them by slot
// descending. Mirrors SlotHashes::new.
func NewSlotHashes(entries ...SlotHash) SlotHashes {
	sh := make(SlotHashes, len(entries))
	copy(sh, entries)
	sort.SliceStable(sh, func(i, j int) bool { return sh[i].Slot > sh[j].Slot })
	return sh
}

// search returns the index of slot (or its descending-order insertion point) and
// whether an exact match was found.
func (sh SlotHashes) search(slot uint64) (int, bool) {
	i := sort.Search(len(sh), func(i int) bool { return sh[i].Slot <= slot })
	if i < len(sh) && sh[i].Slot == slot {
		return i, true
	}
	return i, false
}

// Get returns the hash recorded for slot, if present.
func (sh SlotHashes) Get(slot uint64) (solana.Hash, bool) {
	if i, ok := sh.search(slot); ok {
		return sh[i].Hash, true
	}
	return solana.Hash{}, false
}

// Position returns the index of slot within the entries, if present.
func (sh SlotHashes) Position(slot uint64) (int, bool) {
	return sh.search(slot)
}

// Add inserts or updates the hash for slot, keeping entries sorted by slot
// descending and capped at SlotHashesMaxEntries. Mirrors SlotHashes::add.
func (sh *SlotHashes) Add(slot uint64, hash solana.Hash) {
	i, found := sh.search(slot)
	if found {
		(*sh)[i].Hash = hash
	} else {
		*sh = append(*sh, SlotHash{})
		copy((*sh)[i+1:], (*sh)[i:])
		(*sh)[i] = SlotHash{Slot: slot, Hash: hash}
	}
	if len(*sh) > SlotHashesMaxEntries {
		*sh = (*sh)[:SlotHashesMaxEntries]
	}
}

func (sh SlotHashes) MarshalWithEncoder(encoder *bin.Encoder) error {
	if err := encoder.WriteUint64(uint64(len(sh)), binary.LittleEndian); err != nil {
		return err
	}
	for _, e := range sh {
		if err := encoder.WriteUint64(e.Slot, binary.LittleEndian); err != nil {
			return err
		}
		if err := encoder.WriteBytes(e.Hash[:], false); err != nil {
			return err
		}
	}
	return nil
}

func (sh *SlotHashes) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	n, err := decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	out := make(SlotHashes, 0, min(n, uint64(decoder.Remaining()/SlotHashSerializedSize)))
	for range n {
		slot, err := decoder.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
		buf, err := decoder.ReadNBytes(32)
		if err != nil {
			return err
		}
		var h solana.Hash
		copy(h[:], buf)
		out = append(out, SlotHash{Slot: slot, Hash: h})
	}
	// Normalize to the descending-by-slot invariant that Get/Position/Add rely
	// on, so lookups stay correct even for an out-of-order (forked or hostile)
	// buffer. A well-formed account is already sorted, so this is a no-op.
	sort.SliceStable(out, func(i, j int) bool { return out[i].Slot > out[j].Slot })
	*sh = out
	return nil
}

func (sh SlotHashes) MarshalBinary() ([]byte, error) { return encodeSysvar(sh) }
func (sh *SlotHashes) UnmarshalBinary(data []byte) error {
	return sh.UnmarshalWithDecoder(bin.NewBinDecoder(data))
}

// DecodeSlotHashes decodes SlotHashes sysvar account data.
func DecodeSlotHashes(data []byte) (SlotHashes, error) {
	var sh SlotHashes
	if err := sh.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return sh, nil
}
