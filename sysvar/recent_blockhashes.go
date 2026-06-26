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
	solana "github.com/gagliardetto/solana-go"
)

const (
	// RecentBlockhashesMaxEntries is the maximum number of entries in the
	// RecentBlockhashes sysvar.
	RecentBlockhashesMaxEntries = 150
	// RecentBlockhashesEntrySerializedSize is the serialized size of one entry
	// (Hash + FeeCalculator).
	RecentBlockhashesEntrySerializedSize = 32 + 8
	// RecentBlockhashesSize is the fixed on-chain account size (u64 length
	// prefix + MAX_ENTRIES * entry size).
	RecentBlockhashesSize = 8 + RecentBlockhashesMaxEntries*RecentBlockhashesEntrySerializedSize
)

// RecentBlockhashesEntry is one (blockhash, fee_calculator) entry of the
// RecentBlockhashes sysvar.
//
// Deprecated: see RecentBlockhashes.
type RecentBlockhashesEntry struct {
	Blockhash     solana.Hash
	FeeCalculator solana.FeeCalculator
}

// RecentBlockhashes is the data of the (deprecated) RecentBlockhashes sysvar
// (account solana.SysVarRecentBlockHashesPubkey): the most recent blockhashes,
// ordered newest first.
//
// Deprecated: the RecentBlockhashes sysvar is deprecated; fetch a recent
// blockhash via the getLatestBlockhash RPC (rpc.GetLatestBlockhash) instead.
type RecentBlockhashes []RecentBlockhashesEntry

func (rb RecentBlockhashes) MarshalWithEncoder(encoder *bin.Encoder) error {
	if err := encoder.WriteUint64(uint64(len(rb)), binary.LittleEndian); err != nil {
		return err
	}
	for _, e := range rb {
		if err := encoder.WriteBytes(e.Blockhash[:], false); err != nil {
			return err
		}
		if err := e.FeeCalculator.MarshalWithEncoder(encoder); err != nil {
			return err
		}
	}
	return nil
}

func (rb *RecentBlockhashes) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	n, err := decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	out := make(RecentBlockhashes, 0, min(n, uint64(decoder.Remaining()/RecentBlockhashesEntrySerializedSize)))
	for range n {
		buf, err := decoder.ReadNBytes(32)
		if err != nil {
			return err
		}
		var e RecentBlockhashesEntry
		copy(e.Blockhash[:], buf)
		if err := e.FeeCalculator.UnmarshalWithDecoder(decoder); err != nil {
			return err
		}
		out = append(out, e)
	}
	*rb = out
	return nil
}

func (rb RecentBlockhashes) MarshalBinary() ([]byte, error) { return encodeSysvar(rb) }
func (rb *RecentBlockhashes) UnmarshalBinary(data []byte) error {
	return rb.UnmarshalWithDecoder(bin.NewBinDecoder(data))
}

// DecodeRecentBlockhashes decodes RecentBlockhashes sysvar account data.
//
// Deprecated: see RecentBlockhashes.
func DecodeRecentBlockhashes(data []byte) (RecentBlockhashes, error) {
	var rb RecentBlockhashes
	if err := rb.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return rb, nil
}
