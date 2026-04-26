// Copyright 2024 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package vote

import (
	"encoding/binary"
	"fmt"

	bin "github.com/gagliardetto/solana-go/binary"
)

// writeCompactU16Len writes a Solana compact-u16 (short_vec) length prefix.
// 1 byte if n < 0x80, 2 bytes if n < 0x4000, 3 bytes otherwise (max 2^16-1).
func writeCompactU16Len(enc *bin.Encoder, n int) error {
	if n < 0 || n > 0xffff {
		return fmt.Errorf("compact-u16 length out of range: %d", n)
	}
	rem := uint16(n)
	for {
		elem := uint8(rem & 0x7f)
		rem >>= 7
		if rem == 0 {
			return enc.WriteUint8(elem)
		}
		if err := enc.WriteUint8(elem | 0x80); err != nil {
			return err
		}
	}
}

// writeVarintU64 writes a LEB128-encoded u64 (as used by serde_varint).
func writeVarintU64(enc *bin.Encoder, v uint64) error {
	for {
		b := uint8(v & 0x7f)
		v >>= 7
		if v == 0 {
			return enc.WriteUint8(b)
		}
		if err := enc.WriteUint8(b | 0x80); err != nil {
			return err
		}
	}
}

// readVarintU64 reads a LEB128-encoded u64.
func readVarintU64(dec *bin.Decoder) (uint64, error) {
	var v uint64
	var shift uint
	for i := 0; i < 10; i++ {
		b, err := dec.ReadUint8()
		if err != nil {
			return 0, err
		}
		v |= uint64(b&0x7f) << shift
		if b&0x80 == 0 {
			return v, nil
		}
		shift += 7
	}
	return 0, fmt.Errorf("varint u64 overflow")
}

// marshalCompactVoteStateUpdate encodes a VoteStateUpdate using the
// serde_compact_vote_state_update wire format:
//
//	root:            u64 (u64::MAX means None)
//	lockout_offsets: short_vec of { varint offset, u8 confirmation_count }
//	hash:            [u8; 32]
//	timestamp:       Option<i64>
//
// Offsets are delta-encoded: each offset is saturating_sub(slot, previous_slot)
// where the initial previous_slot is the root (or u64::MAX if None).
// Lockouts are assumed to be sorted by slot (ascending).
func marshalCompactVoteStateUpdate(enc *bin.Encoder, u *VoteStateUpdate) error {
	root := ^uint64(0)
	if u.Root != nil {
		root = *u.Root
	}
	if err := enc.WriteUint64(root, binary.LittleEndian); err != nil {
		return err
	}
	if err := writeCompactU16Len(enc, len(u.Lockouts)); err != nil {
		return err
	}
	prev := root
	for _, l := range u.Lockouts {
		var offset uint64
		if l.Slot > prev {
			offset = l.Slot - prev
		} // else saturating_sub -> 0
		if err := writeVarintU64(enc, offset); err != nil {
			return err
		}
		if l.ConfirmationCount > 0xff {
			return fmt.Errorf("confirmation_count %d exceeds u8 max", l.ConfirmationCount)
		}
		if err := enc.WriteUint8(uint8(l.ConfirmationCount)); err != nil {
			return err
		}
		prev = l.Slot
	}
	if err := enc.WriteBytes(u.Hash[:], false); err != nil {
		return err
	}
	if u.Timestamp == nil {
		return enc.WriteUint8(0)
	}
	if err := enc.WriteUint8(1); err != nil {
		return err
	}
	return enc.WriteInt64(*u.Timestamp, binary.LittleEndian)
}

// unmarshalCompactVoteStateUpdate decodes a VoteStateUpdate from the
// serde_compact_vote_state_update wire format. Absolute lockout slots are
// reconstructed by cumulative sum over the delta offsets.
func unmarshalCompactVoteStateUpdate(dec *bin.Decoder, u *VoteStateUpdate) error {
	root, err := dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	if root == ^uint64(0) {
		u.Root = nil
	} else {
		r := root
		u.Root = &r
	}
	count, err := dec.ReadCompactU16Length()
	if err != nil {
		return err
	}
	u.Lockouts = make([]Lockout, count)
	prev := root
	for i := range count {
		offset, err := readVarintU64(dec)
		if err != nil {
			return err
		}
		cc, err := dec.ReadUint8()
		if err != nil {
			return err
		}
		// checked_add semantics: overflow produces a wrapped slot, but the
		// on-chain program rejects such values at runtime. We preserve the
		// bits so callers can detect the problem.
		prev += offset
		u.Lockouts[i] = Lockout{Slot: prev, ConfirmationCount: uint32(cc)}
	}
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(u.Hash[:], b)
	hasTs, err := dec.ReadUint8()
	if err != nil {
		return err
	}
	switch hasTs {
	case 0:
	case 1:
		ts, err := dec.ReadInt64(binary.LittleEndian)
		if err != nil {
			return err
		}
		u.Timestamp = &ts
	default:
		return fmt.Errorf("invalid Option<i64> discriminant: %d", hasTs)
	}
	return nil
}
