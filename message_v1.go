// Copyright 2026 github.com/gagliardetto
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

package solana

// V1 message format (SIMD-0385); ported from solana-sdk message/src/versions/v1.
//
// Signed message bytes:
//
//	0x81 | header(3) | mask(u32 LE) | lifetime(32) | num_ix(u8) | num_addr(u8)
//	| addresses | config values | ix headers (u8,u8,u16 LE) | ix payloads
//
// Transaction = message || fixed 64-byte signatures (no length prefix).

import (
	"encoding/binary"
	"fmt"

	bin "github.com/gagliardetto/binary"
)

// messageVersionV1Prefix is the first byte of a V1 message (0x81).
const messageVersionV1Prefix byte = messageVersionPrefix | 1

const (
	// MaxTransactionSizeV1 is the V1 size cap, signatures included (enforced by the cluster).
	MaxTransactionSizeV1 = 4096
	// MaxAddressesV1 is the maximum number of addresses in a V1 message.
	MaxAddressesV1 = 64
	// MaxInstructionsV1 is the maximum number of instructions in a V1 message.
	MaxInstructionsV1 = 64
	// MaxSignaturesV1 is the maximum number of signatures in a V1 transaction.
	MaxSignaturesV1 = 12
	// MinHeapSizeV1 is the minimum (and default) requested heap size.
	MinHeapSizeV1 uint32 = 32 * 1024
	// MaxHeapSizeV1 is the maximum requested heap size.
	MaxHeapSizeV1 uint32 = 256 * 1024
	// DefaultHeapSizeV1 is the heap size when none is requested.
	DefaultHeapSizeV1 = MinHeapSizeV1

	// fixedHeaderSizeV1: header(3)+mask(4)+lifetime(32)+num_ix(1)+num_addr(1), excludes prefix.
	fixedHeaderSizeV1 = 3 + 4 + 32 + 1 + 1
	// instructionHeaderSizeV1: program_id_index(1)+num_accounts(1)+data_len(2).
	instructionHeaderSizeV1 = 1 + 1 + 2
)

// TransactionConfig is the v1 inline compute budget; nil field = not requested.
// Unset ComputeUnitLimit means 0 CUs; PriorityFee is total lamports.
type TransactionConfig struct {
	// Total priority fee in lamports.
	PriorityFee *uint64 `json:"priorityFee"`
	// Maximum compute units; nil means 0.
	ComputeUnitLimit *uint32 `json:"computeUnitLimit"`
	// Maximum bytes of loaded account data; nil means 0.
	LoadedAccountsDataSizeLimit *uint32 `json:"loadedAccountsDataSizeLimit"`
	// Heap size: multiple of 1024 in [MinHeapSizeV1, MaxHeapSizeV1]; nil means default.
	HeapSize *uint32 `json:"heapSize"`
}

// WithPriorityFee returns a copy with the priority fee (total lamports) set.
func (c TransactionConfig) WithPriorityFee(lamports uint64) TransactionConfig {
	c.PriorityFee = &lamports
	return c
}

// WithComputeUnitLimit returns a copy with the compute unit limit set.
func (c TransactionConfig) WithComputeUnitLimit(units uint32) TransactionConfig {
	c.ComputeUnitLimit = &units
	return c
}

// WithLoadedAccountsDataSizeLimit returns a copy with the loaded accounts data size limit set.
func (c TransactionConfig) WithLoadedAccountsDataSizeLimit(bytes uint32) TransactionConfig {
	c.LoadedAccountsDataSizeLimit = &bytes
	return c
}

// WithHeapSize returns a copy with the heap size set (validated by Sanitize).
func (c TransactionConfig) WithHeapSize(bytes uint32) TransactionConfig {
	c.HeapSize = &bytes
	return c
}

// IsEmpty reports whether no config value is set.
func (c TransactionConfig) IsEmpty() bool {
	return c.Mask() == 0
}

// Size returns the encoded size of the config values in bytes.
func (c TransactionConfig) Size() int {
	return c.Mask().SizeOfConfig()
}

// Mask returns the TransactionConfigMask describing which values are set.
func (c TransactionConfig) Mask() TransactionConfigMask {
	var mask TransactionConfigMask
	if c.PriorityFee != nil {
		mask |= TransactionConfigMaskPriorityFee
	}
	if c.ComputeUnitLimit != nil {
		mask |= TransactionConfigMaskComputeUnitLimit
	}
	if c.LoadedAccountsDataSizeLimit != nil {
		mask |= TransactionConfigMaskLoadedAccountsDataSize
	}
	if c.HeapSize != nil {
		mask |= TransactionConfigMaskHeapSize
	}
	return mask
}

// TransactionConfigMask flags which config values a V1 message carries.
type TransactionConfigMask uint32

const (
	// TransactionConfigMaskPriorityFee: bits 0-1 (both required), u64 LE.
	TransactionConfigMaskPriorityFee TransactionConfigMask = 0b11
	// TransactionConfigMaskComputeUnitLimit: bit 2, u32 LE.
	TransactionConfigMaskComputeUnitLimit TransactionConfigMask = 0b100
	// TransactionConfigMaskLoadedAccountsDataSize: bit 3, u32 LE.
	TransactionConfigMaskLoadedAccountsDataSize TransactionConfigMask = 0b1000
	// TransactionConfigMaskHeapSize: bit 4, u32 LE.
	TransactionConfigMaskHeapSize TransactionConfigMask = 0b10000
	// TransactionConfigMaskKnownBits is the union of all supported bits.
	TransactionConfigMaskKnownBits = TransactionConfigMaskPriorityFee |
		TransactionConfigMaskComputeUnitLimit |
		TransactionConfigMaskLoadedAccountsDataSize |
		TransactionConfigMaskHeapSize
)

// HasUnknownBits reports whether any unsupported bit is set.
func (m TransactionConfigMask) HasUnknownBits() bool {
	return (m | TransactionConfigMaskKnownBits) != TransactionConfigMaskKnownBits
}

// HasPriorityFee reports whether both priority fee bits are set.
func (m TransactionConfigMask) HasPriorityFee() bool {
	return (m & TransactionConfigMaskPriorityFee) == TransactionConfigMaskPriorityFee
}

// HasInvalidPriorityFeeBits reports whether exactly one priority fee bit is set.
func (m TransactionConfigMask) HasInvalidPriorityFeeBits() bool {
	bits := m & TransactionConfigMaskPriorityFee
	return bits != 0 && bits != TransactionConfigMaskPriorityFee
}

// HasComputeUnitLimit reports whether the compute unit limit bit is set.
func (m TransactionConfigMask) HasComputeUnitLimit() bool {
	return (m & TransactionConfigMaskComputeUnitLimit) != 0
}

// HasLoadedAccountsDataSize reports whether the loaded accounts data size bit is set.
func (m TransactionConfigMask) HasLoadedAccountsDataSize() bool {
	return (m & TransactionConfigMaskLoadedAccountsDataSize) != 0
}

// HasHeapSize reports whether the heap size bit is set.
func (m TransactionConfigMask) HasHeapSize() bool {
	return (m & TransactionConfigMaskHeapSize) != 0
}

// SizeOfConfig returns the encoded size of the flagged config values.
func (m TransactionConfigMask) SizeOfConfig() int {
	size := 0
	if m.HasPriorityFee() {
		size += 8
	}
	if m.HasComputeUnitLimit() {
		size += 4
	}
	if m.HasLoadedAccountsDataSize() {
		size += 4
	}
	if m.HasHeapSize() {
		size += 4
	}
	return size
}

// sizeV1 returns the V1 encoded size, version byte included.
func (mx *Message) sizeV1() int {
	size := 1 + fixedHeaderSizeV1 +
		32*len(mx.AccountKeys) +
		mx.TransactionConfig.Size() +
		instructionHeaderSizeV1*len(mx.Instructions)
	for i := range mx.Instructions {
		size += len(mx.Instructions[i].Accounts) + len(mx.Instructions[i].Data)
	}
	return size
}

// MarshalV1 encodes the message as v1 (0x81 prefix included); never truncates.
func (mx *Message) MarshalV1() ([]byte, error) {
	if len(mx.AddressTableLookups) > 0 {
		return nil, fmt.Errorf("v1 messages do not support address table lookups (got %d)", len(mx.AddressTableLookups))
	}
	if len(mx.AccountKeys) > 255 {
		return nil, fmt.Errorf("v1 message: too many account keys (%d), max 255", len(mx.AccountKeys))
	}
	if len(mx.Instructions) > 255 {
		return nil, fmt.Errorf("v1 message: too many instructions (%d), max 255", len(mx.Instructions))
	}
	for i := range mx.Instructions {
		ix := &mx.Instructions[i]
		if ix.ProgramIDIndex > 255 {
			return nil, fmt.Errorf("v1 message: instruction %d: program_id_index %d does not fit in a byte", i, ix.ProgramIDIndex)
		}
		if len(ix.Accounts) > 255 {
			return nil, fmt.Errorf("v1 message: instruction %d: too many accounts (%d), max 255", i, len(ix.Accounts))
		}
		if len(ix.Data) > 65535 {
			return nil, fmt.Errorf("v1 message: instruction %d: data too large (%d bytes), max 65535", i, len(ix.Data))
		}
		for j, idx := range ix.Accounts {
			if idx > 255 {
				return nil, fmt.Errorf("v1 message: instruction %d: account index %d (%d) does not fit in a byte", i, j, idx)
			}
		}
	}

	buf := make([]byte, 0, mx.sizeV1())
	buf = append(buf, messageVersionV1Prefix,
		mx.Header.NumRequiredSignatures,
		mx.Header.NumReadonlySignedAccounts,
		mx.Header.NumReadonlyUnsignedAccounts,
	)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(mx.TransactionConfig.Mask()))
	buf = append(buf, mx.RecentBlockhash[:]...)
	buf = append(buf, byte(len(mx.Instructions)), byte(len(mx.AccountKeys)))
	for i := range mx.AccountKeys {
		buf = append(buf, mx.AccountKeys[i][:]...)
	}

	cfg := &mx.TransactionConfig
	if cfg.PriorityFee != nil {
		buf = binary.LittleEndian.AppendUint64(buf, *cfg.PriorityFee)
	}
	if cfg.ComputeUnitLimit != nil {
		buf = binary.LittleEndian.AppendUint32(buf, *cfg.ComputeUnitLimit)
	}
	if cfg.LoadedAccountsDataSizeLimit != nil {
		buf = binary.LittleEndian.AppendUint32(buf, *cfg.LoadedAccountsDataSizeLimit)
	}
	if cfg.HeapSize != nil {
		buf = binary.LittleEndian.AppendUint32(buf, *cfg.HeapSize)
	}

	for i := range mx.Instructions {
		ix := &mx.Instructions[i]
		buf = append(buf, byte(ix.ProgramIDIndex), byte(len(ix.Accounts)))
		buf = binary.LittleEndian.AppendUint16(buf, uint16(len(ix.Data)))
	}
	for i := range mx.Instructions {
		ix := &mx.Instructions[i]
		for _, idx := range ix.Accounts {
			buf = append(buf, byte(idx))
		}
		buf = append(buf, ix.Data...)
	}
	return buf, nil
}

// UnmarshalV1 decodes a v1 message from the 0x81 prefix; limits are checked by Sanitize.
func (mx *Message) UnmarshalV1(decoder *bin.Decoder) (err error) {
	prefix, err := decoder.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read message version prefix: %w", err)
	}
	if prefix != messageVersionV1Prefix {
		return fmt.Errorf("invalid v1 message version prefix: expected 0x%02x, got 0x%02x", messageVersionV1Prefix, prefix)
	}
	mx.version = MessageVersionV1
	mx.AddressTableLookups = nil
	mx.addressTables = nil
	mx.resolved = false
	mx.TransactionConfig = TransactionConfig{}

	if decoder.Remaining() < fixedHeaderSizeV1 {
		return fmt.Errorf("v1 message: buffer too small: need at least %d bytes for the fixed header, have %d", fixedHeaderSizeV1, decoder.Remaining())
	}
	if mx.Header.NumRequiredSignatures, err = decoder.ReadUint8(); err != nil {
		return fmt.Errorf("unable to decode mx.Header.NumRequiredSignatures: %w", err)
	}
	if mx.Header.NumReadonlySignedAccounts, err = decoder.ReadUint8(); err != nil {
		return fmt.Errorf("unable to decode mx.Header.NumReadonlySignedAccounts: %w", err)
	}
	if mx.Header.NumReadonlyUnsignedAccounts, err = decoder.ReadUint8(); err != nil {
		return fmt.Errorf("unable to decode mx.Header.NumReadonlyUnsignedAccounts: %w", err)
	}

	rawMask, err := decoder.ReadUint32(bin.LE)
	if err != nil {
		return fmt.Errorf("unable to decode v1 config mask: %w", err)
	}
	mask := TransactionConfigMask(rawMask)
	if mask.HasUnknownBits() || mask.HasInvalidPriorityFeeBits() {
		return fmt.Errorf("invalid transaction config mask: 0x%08x", rawMask)
	}

	if _, err := decoder.Read(mx.RecentBlockhash[:]); err != nil {
		return fmt.Errorf("unable to decode mx.RecentBlockhash: %w", err)
	}
	numInstructions, err := decoder.ReadUint8()
	if err != nil {
		return fmt.Errorf("unable to decode numInstructions: %w", err)
	}
	numAccountKeys, err := decoder.ReadUint8()
	if err != nil {
		return fmt.Errorf("unable to decode numAccountKeys: %w", err)
	}

	if int(numAccountKeys)*32 > decoder.Remaining() {
		return fmt.Errorf("numAccountKeys %d is too large for remaining bytes %d", numAccountKeys, decoder.Remaining())
	}
	mx.AccountKeys = make(PublicKeySlice, numAccountKeys)
	for i := range mx.AccountKeys {
		if _, err := decoder.Read(mx.AccountKeys[i][:]); err != nil {
			return fmt.Errorf("unable to decode mx.AccountKeys[%d]: %w", i, err)
		}
	}

	if mask.SizeOfConfig() > decoder.Remaining() {
		return fmt.Errorf("v1 message: buffer too small for config values: need %d bytes, have %d", mask.SizeOfConfig(), decoder.Remaining())
	}
	if mask.HasPriorityFee() {
		v, err := decoder.ReadUint64(bin.LE)
		if err != nil {
			return fmt.Errorf("unable to decode priority fee: %w", err)
		}
		mx.TransactionConfig.PriorityFee = &v
	}
	if mask.HasComputeUnitLimit() {
		v, err := decoder.ReadUint32(bin.LE)
		if err != nil {
			return fmt.Errorf("unable to decode compute unit limit: %w", err)
		}
		mx.TransactionConfig.ComputeUnitLimit = &v
	}
	if mask.HasLoadedAccountsDataSize() {
		v, err := decoder.ReadUint32(bin.LE)
		if err != nil {
			return fmt.Errorf("unable to decode loaded accounts data size limit: %w", err)
		}
		mx.TransactionConfig.LoadedAccountsDataSizeLimit = &v
	}
	if mask.HasHeapSize() {
		v, err := decoder.ReadUint32(bin.LE)
		if err != nil {
			return fmt.Errorf("unable to decode heap size: %w", err)
		}
		mx.TransactionConfig.HeapSize = &v
	}

	if int(numInstructions)*instructionHeaderSizeV1 > decoder.Remaining() {
		return fmt.Errorf("numInstructions %d is too large for remaining bytes %d", numInstructions, decoder.Remaining())
	}
	mx.Instructions = make([]CompiledInstruction, numInstructions)
	numAccounts := make([]uint8, numInstructions)
	dataLens := make([]uint16, numInstructions)
	for i := range mx.Instructions {
		programIDIndex, err := decoder.ReadUint8()
		if err != nil {
			return fmt.Errorf("unable to decode ix[%d].ProgramIDIndex: %w", i, err)
		}
		mx.Instructions[i].ProgramIDIndex = uint16(programIDIndex)
		if numAccounts[i], err = decoder.ReadUint8(); err != nil {
			return fmt.Errorf("unable to decode numAccounts for ix[%d]: %w", i, err)
		}
		if dataLens[i], err = decoder.ReadUint16(bin.LE); err != nil {
			return fmt.Errorf("unable to decode dataLen for ix[%d]: %w", i, err)
		}
	}

	for i := range mx.Instructions {
		needed := int(numAccounts[i]) + int(dataLens[i])
		if needed > decoder.Remaining() {
			return fmt.Errorf("ix[%d]: payload of %d bytes is greater than remaining bytes %d", i, needed, decoder.Remaining())
		}
		ix := &mx.Instructions[i]
		ix.Accounts = make([]uint16, numAccounts[i])
		for j := range ix.Accounts {
			idx, err := decoder.ReadUint8()
			if err != nil {
				return fmt.Errorf("unable to decode accountIndex for ix[%d].Accounts[%d]: %w", i, j, err)
			}
			ix.Accounts[j] = uint16(idx)
		}
		data, err := decoder.ReadBytes(int(dataLens[i]))
		if err != nil {
			return fmt.Errorf("unable to decode dataBytes for ix[%d]: %w", i, err)
		}
		ix.Data = Base58(data)
	}
	return nil
}
