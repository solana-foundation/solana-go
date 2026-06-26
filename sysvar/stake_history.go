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
)

const (
	// StakeHistoryMaxEntries is the maximum number of epochs the StakeHistory
	// sysvar holds.
	StakeHistoryMaxEntries = 512
	// StakeHistoryEntrySerializedSize is the serialized size of one
	// (epoch, entry) pair (4 x u64).
	StakeHistoryEntrySerializedSize = 32
	// StakeHistorySize is the fixed on-chain account size of the StakeHistory
	// sysvar (u64 length prefix + MAX_ENTRIES * StakeHistoryEntrySerializedSize).
	StakeHistorySize = 8 + StakeHistoryMaxEntries*StakeHistoryEntrySerializedSize
)

// StakeHistoryEntry is the cluster-wide stake activation state for one epoch.
type StakeHistoryEntry struct {
	// Effective is the effective stake at this epoch.
	Effective uint64
	// Activating is the sum of stake activating at this epoch.
	Activating uint64
	// Deactivating is the sum of stake deactivating at this epoch.
	Deactivating uint64
}

// StakeHistoryItem pairs an epoch with its StakeHistoryEntry.
type StakeHistoryItem struct {
	Epoch uint64
	Entry StakeHistoryEntry
}

// StakeHistory is the data of the StakeHistory sysvar (account
// solana.SysVarStakeHistoryPubkey): per-epoch stake activation history, kept sorted by
// epoch DESCENDING (index 0 is the newest epoch). Mirrors solana-sdk
// stake_history::StakeHistory.
type StakeHistory []StakeHistoryItem

// Get returns the StakeHistoryEntry for epoch, if present.
func (sh StakeHistory) Get(epoch uint64) (StakeHistoryEntry, bool) {
	i := sort.Search(len(sh), func(i int) bool { return sh[i].Epoch <= epoch })
	if i < len(sh) && sh[i].Epoch == epoch {
		return sh[i].Entry, true
	}
	return StakeHistoryEntry{}, false
}

// Add inserts or updates the entry for epoch, keeping entries sorted by epoch
// descending and capped at StakeHistoryMaxEntries (dropping the oldest).
// Mirrors StakeHistory::add.
func (sh *StakeHistory) Add(epoch uint64, entry StakeHistoryEntry) {
	i := sort.Search(len(*sh), func(i int) bool { return (*sh)[i].Epoch <= epoch })
	if i < len(*sh) && (*sh)[i].Epoch == epoch {
		(*sh)[i].Entry = entry
	} else {
		*sh = append(*sh, StakeHistoryItem{})
		copy((*sh)[i+1:], (*sh)[i:])
		(*sh)[i] = StakeHistoryItem{Epoch: epoch, Entry: entry}
	}
	if len(*sh) > StakeHistoryMaxEntries {
		*sh = (*sh)[:StakeHistoryMaxEntries]
	}
}

func (sh StakeHistory) MarshalWithEncoder(encoder *bin.Encoder) error {
	if err := encoder.WriteUint64(uint64(len(sh)), binary.LittleEndian); err != nil {
		return err
	}
	for _, it := range sh {
		for _, v := range [4]uint64{it.Epoch, it.Entry.Effective, it.Entry.Activating, it.Entry.Deactivating} {
			if err := encoder.WriteUint64(v, binary.LittleEndian); err != nil {
				return err
			}
		}
	}
	return nil
}

func (sh *StakeHistory) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	n, err := decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	out := make(StakeHistory, 0, min(n, uint64(decoder.Remaining()/StakeHistoryEntrySerializedSize)))
	for range n {
		var vals [4]uint64
		for j := range vals {
			if vals[j], err = decoder.ReadUint64(binary.LittleEndian); err != nil {
				return err
			}
		}
		out = append(out, StakeHistoryItem{
			Epoch: vals[0],
			Entry: StakeHistoryEntry{Effective: vals[1], Activating: vals[2], Deactivating: vals[3]},
		})
	}
	// Normalize to the descending-by-epoch invariant that Get/Add rely on, so
	// lookups stay correct even for an out-of-order (forked or hostile) buffer.
	// A well-formed account is already sorted, so this is a no-op.
	sort.SliceStable(out, func(i, j int) bool { return out[i].Epoch > out[j].Epoch })
	*sh = out
	return nil
}

func (sh StakeHistory) MarshalBinary() ([]byte, error) { return encodeSysvar(sh) }
func (sh *StakeHistory) UnmarshalBinary(data []byte) error {
	return sh.UnmarshalWithDecoder(bin.NewBinDecoder(data))
}

// DecodeStakeHistory decodes StakeHistory sysvar account data.
func DecodeStakeHistory(data []byte) (StakeHistory, error) {
	var sh StakeHistory
	if err := sh.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return sh, nil
}
