// Copyright 2024 github.com/gagliardetto
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

package vote

import (
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
)

// VoteStateVersion identifies which version of VoteState is stored.
type VoteStateVersion uint32

const (
	VoteStateVersionV0_23_5  VoteStateVersion = 0 // rejected on deserialize
	VoteStateVersionV1_14_11 VoteStateVersion = 1
	VoteStateVersionV3       VoteStateVersion = 2
	VoteStateVersionV4       VoteStateVersion = 3
)

// BlockTimestamp pairs a slot with a unix timestamp.
type BlockTimestamp struct {
	Slot      uint64
	Timestamp int64
}

func (b *BlockTimestamp) UnmarshalWithDecoder(dec *bin.Decoder) error {
	var err error
	b.Slot, err = dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	b.Timestamp, err = dec.ReadInt64(binary.LittleEndian)
	return err
}

func (b BlockTimestamp) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteUint64(b.Slot, binary.LittleEndian); err != nil {
		return err
	}
	return enc.WriteInt64(b.Timestamp, binary.LittleEndian)
}

// AuthorizedVoters is a map of epoch to authorized voter public key.
// In the on-chain representation this is a BTreeMap<Epoch, Pubkey> with u64 length prefix.
type AuthorizedVoters struct {
	Voters []AuthorizedVoter
}

// AuthorizedVoter is a single epoch -> pubkey mapping.
type AuthorizedVoter struct {
	Epoch  uint64
	Pubkey solana.PublicKey
}

func (av *AuthorizedVoters) UnmarshalWithDecoder(dec *bin.Decoder) error {
	count, err := dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	av.Voters = make([]AuthorizedVoter, count)
	for i := range count {
		av.Voters[i].Epoch, err = dec.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
		b, err := dec.ReadNBytes(32)
		if err != nil {
			return err
		}
		av.Voters[i].Pubkey = solana.PublicKeyFromBytes(b)
	}
	return nil
}

func (av AuthorizedVoters) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteUint64(uint64(len(av.Voters)), binary.LittleEndian); err != nil {
		return err
	}
	for _, v := range av.Voters {
		if err := enc.WriteUint64(v.Epoch, binary.LittleEndian); err != nil {
			return err
		}
		if err := enc.WriteBytes(v.Pubkey[:], false); err != nil {
			return err
		}
	}
	return nil
}

// PriorVoter is a single entry in the prior voters circular buffer.
type PriorVoter struct {
	Pubkey     solana.PublicKey
	StartEpoch uint64
	EndEpoch   uint64
}

// PriorVotersCircBuf is a circular buffer of 32 prior voters.
//
// The zero value is NOT a valid empty buffer — use NewPriorVotersCircBuf()
// to construct a correctly-defaulted instance (Idx = MAX_ITEMS - 1, IsEmpty = true).
// This matches the Rust CircBuf::default() behavior.
type PriorVotersCircBuf struct {
	Buf     [CircBufMaxItems]PriorVoter
	Idx     uint64
	IsEmpty bool
}

// NewPriorVotersCircBuf returns an empty PriorVotersCircBuf with the same
// default values as the Rust CircBuf::default(): idx = MAX_ITEMS - 1,
// is_empty = true.
func NewPriorVotersCircBuf() PriorVotersCircBuf {
	return PriorVotersCircBuf{
		Idx:     CircBufMaxItems - 1,
		IsEmpty: true,
	}
}

func (c *PriorVotersCircBuf) UnmarshalWithDecoder(dec *bin.Decoder) error {
	for i := range CircBufMaxItems {
		b, err := dec.ReadNBytes(32)
		if err != nil {
			return err
		}
		c.Buf[i].Pubkey = solana.PublicKeyFromBytes(b)
		c.Buf[i].StartEpoch, err = dec.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
		c.Buf[i].EndEpoch, err = dec.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	var err error
	c.Idx, err = dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	c.IsEmpty, err = dec.ReadBool()
	return err
}

func (c PriorVotersCircBuf) MarshalWithEncoder(enc *bin.Encoder) error {
	for i := range CircBufMaxItems {
		if err := enc.WriteBytes(c.Buf[i].Pubkey[:], false); err != nil {
			return err
		}
		if err := enc.WriteUint64(c.Buf[i].StartEpoch, binary.LittleEndian); err != nil {
			return err
		}
		if err := enc.WriteUint64(c.Buf[i].EndEpoch, binary.LittleEndian); err != nil {
			return err
		}
	}
	if err := enc.WriteUint64(c.Idx, binary.LittleEndian); err != nil {
		return err
	}
	return enc.WriteBool(c.IsEmpty)
}

// EpochCredit is a single (epoch, credits, prev_credits) tuple.
type EpochCredit struct {
	Epoch       uint64
	Credits     uint64
	PrevCredits uint64
}

// decodeEpochCredits decodes a Vec<(Epoch, u64, u64)> with a u64 length prefix.
func decodeEpochCredits(dec *bin.Decoder) ([]EpochCredit, error) {
	count, err := dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	out := make([]EpochCredit, count)
	for i := range count {
		out[i].Epoch, err = dec.ReadUint64(binary.LittleEndian)
		if err != nil {
			return nil, err
		}
		out[i].Credits, err = dec.ReadUint64(binary.LittleEndian)
		if err != nil {
			return nil, err
		}
		out[i].PrevCredits, err = dec.ReadUint64(binary.LittleEndian)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func encodeEpochCredits(enc *bin.Encoder, credits []EpochCredit) error {
	if err := enc.WriteUint64(uint64(len(credits)), binary.LittleEndian); err != nil {
		return err
	}
	for _, c := range credits {
		if err := enc.WriteUint64(c.Epoch, binary.LittleEndian); err != nil {
			return err
		}
		if err := enc.WriteUint64(c.Credits, binary.LittleEndian); err != nil {
			return err
		}
		if err := enc.WriteUint64(c.PrevCredits, binary.LittleEndian); err != nil {
			return err
		}
	}
	return nil
}

// VoteState1_14_11 is the legacy vote state (tag 1, size 3731).
// Uses VecDeque<Lockout> without latency bytes.
type VoteState1_14_11 struct {
	NodePubkey           solana.PublicKey
	AuthorizedWithdrawer solana.PublicKey
	Commission           uint8
	Votes                []Lockout
	RootSlot             *uint64
	AuthorizedVoters     AuthorizedVoters
	PriorVoters          PriorVotersCircBuf
	EpochCredits         []EpochCredit
	LastTimestamp        BlockTimestamp
}

// VoteStateV3 is the current production vote state (tag 2, size 3762).
// Uses VecDeque<LandedVote>.
type VoteStateV3 struct {
	NodePubkey           solana.PublicKey
	AuthorizedWithdrawer solana.PublicKey
	Commission           uint8
	Votes                []LandedVote
	RootSlot             *uint64
	AuthorizedVoters     AuthorizedVoters
	PriorVoters          PriorVotersCircBuf
	EpochCredits         []EpochCredit
	LastTimestamp        BlockTimestamp
}

// VoteStateV4 is the newest vote state (tag 3, size 3762).
// Adds BLS pubkey, commission basis points, collector accounts, removes prior voters.
type VoteStateV4 struct {
	NodePubkey                    solana.PublicKey
	AuthorizedWithdrawer          solana.PublicKey
	InflationRewardsCollector     solana.PublicKey
	BlockRevenueCollector         solana.PublicKey
	InflationRewardsCommissionBps uint16
	BlockRevenueCommissionBps     uint16
	PendingDelegatorRewards       uint64
	BLSPubkeyCompressed           *[BLS_PUBLIC_KEY_COMPRESSED_SIZE]byte
	Votes                         []LandedVote
	RootSlot                      *uint64
	AuthorizedVoters              AuthorizedVoters
	EpochCredits                  []EpochCredit
	LastTimestamp                 BlockTimestamp
}

// VoteStateVersions is a tagged union holding one of the vote state versions.
type VoteStateVersions struct {
	Version  VoteStateVersion
	V1_14_11 *VoteState1_14_11
	V3       *VoteStateV3
	V4       *VoteStateV4
}

func decodeOptionU64(dec *bin.Decoder) (*uint64, error) {
	has, err := dec.ReadUint8()
	if err != nil {
		return nil, err
	}
	if has == 0 {
		return nil, nil
	}
	if has != 1 {
		return nil, fmt.Errorf("invalid Option<u64> discriminant: %d", has)
	}
	v, err := dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func encodeOptionU64(enc *bin.Encoder, v *uint64) error {
	if v == nil {
		return enc.WriteUint8(0)
	}
	if err := enc.WriteUint8(1); err != nil {
		return err
	}
	return enc.WriteUint64(*v, binary.LittleEndian)
}

func (s *VoteState1_14_11) unmarshalBody(dec *bin.Decoder) error {
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	s.NodePubkey = solana.PublicKeyFromBytes(b)
	b, err = dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	s.AuthorizedWithdrawer = solana.PublicKeyFromBytes(b)
	s.Commission, err = dec.ReadUint8()
	if err != nil {
		return err
	}
	count, err := dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	s.Votes = make([]Lockout, count)
	for i := range count {
		if err := s.Votes[i].UnmarshalWithDecoder(dec); err != nil {
			return err
		}
	}
	s.RootSlot, err = decodeOptionU64(dec)
	if err != nil {
		return err
	}
	if err := s.AuthorizedVoters.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	if err := s.PriorVoters.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	s.EpochCredits, err = decodeEpochCredits(dec)
	if err != nil {
		return err
	}
	return s.LastTimestamp.UnmarshalWithDecoder(dec)
}

func (s VoteState1_14_11) marshalBody(enc *bin.Encoder) error {
	if err := enc.WriteBytes(s.NodePubkey[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.AuthorizedWithdrawer[:], false); err != nil {
		return err
	}
	if err := enc.WriteUint8(s.Commission); err != nil {
		return err
	}
	if err := enc.WriteUint64(uint64(len(s.Votes)), binary.LittleEndian); err != nil {
		return err
	}
	for _, v := range s.Votes {
		if err := v.MarshalWithEncoder(enc); err != nil {
			return err
		}
	}
	if err := encodeOptionU64(enc, s.RootSlot); err != nil {
		return err
	}
	if err := s.AuthorizedVoters.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := s.PriorVoters.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := encodeEpochCredits(enc, s.EpochCredits); err != nil {
		return err
	}
	return s.LastTimestamp.MarshalWithEncoder(enc)
}

func (s *VoteStateV3) unmarshalBody(dec *bin.Decoder) error {
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	s.NodePubkey = solana.PublicKeyFromBytes(b)
	b, err = dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	s.AuthorizedWithdrawer = solana.PublicKeyFromBytes(b)
	s.Commission, err = dec.ReadUint8()
	if err != nil {
		return err
	}
	count, err := dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	s.Votes = make([]LandedVote, count)
	for i := range count {
		if err := s.Votes[i].UnmarshalWithDecoder(dec); err != nil {
			return err
		}
	}
	s.RootSlot, err = decodeOptionU64(dec)
	if err != nil {
		return err
	}
	if err := s.AuthorizedVoters.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	if err := s.PriorVoters.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	s.EpochCredits, err = decodeEpochCredits(dec)
	if err != nil {
		return err
	}
	return s.LastTimestamp.UnmarshalWithDecoder(dec)
}

func (s VoteStateV3) marshalBody(enc *bin.Encoder) error {
	if err := enc.WriteBytes(s.NodePubkey[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.AuthorizedWithdrawer[:], false); err != nil {
		return err
	}
	if err := enc.WriteUint8(s.Commission); err != nil {
		return err
	}
	if err := enc.WriteUint64(uint64(len(s.Votes)), binary.LittleEndian); err != nil {
		return err
	}
	for _, v := range s.Votes {
		if err := v.MarshalWithEncoder(enc); err != nil {
			return err
		}
	}
	if err := encodeOptionU64(enc, s.RootSlot); err != nil {
		return err
	}
	if err := s.AuthorizedVoters.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := s.PriorVoters.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := encodeEpochCredits(enc, s.EpochCredits); err != nil {
		return err
	}
	return s.LastTimestamp.MarshalWithEncoder(enc)
}

func (s *VoteStateV4) unmarshalBody(dec *bin.Decoder) error {
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	s.NodePubkey = solana.PublicKeyFromBytes(b)
	b, err = dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	s.AuthorizedWithdrawer = solana.PublicKeyFromBytes(b)
	b, err = dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	s.InflationRewardsCollector = solana.PublicKeyFromBytes(b)
	b, err = dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	s.BlockRevenueCollector = solana.PublicKeyFromBytes(b)
	s.InflationRewardsCommissionBps, err = dec.ReadUint16(binary.LittleEndian)
	if err != nil {
		return err
	}
	s.BlockRevenueCommissionBps, err = dec.ReadUint16(binary.LittleEndian)
	if err != nil {
		return err
	}
	s.PendingDelegatorRewards, err = dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	// Option<[u8; 48]>
	hasBls, err := dec.ReadUint8()
	if err != nil {
		return err
	}
	if hasBls == 1 {
		pk := new([BLS_PUBLIC_KEY_COMPRESSED_SIZE]byte)
		b, err := dec.ReadNBytes(BLS_PUBLIC_KEY_COMPRESSED_SIZE)
		if err != nil {
			return err
		}
		copy(pk[:], b)
		s.BLSPubkeyCompressed = pk
	} else if hasBls != 0 {
		return fmt.Errorf("invalid Option<BLSPubkey> discriminant: %d", hasBls)
	}
	count, err := dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	s.Votes = make([]LandedVote, count)
	for i := range count {
		if err := s.Votes[i].UnmarshalWithDecoder(dec); err != nil {
			return err
		}
	}
	s.RootSlot, err = decodeOptionU64(dec)
	if err != nil {
		return err
	}
	if err := s.AuthorizedVoters.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	s.EpochCredits, err = decodeEpochCredits(dec)
	if err != nil {
		return err
	}
	return s.LastTimestamp.UnmarshalWithDecoder(dec)
}

func (s VoteStateV4) marshalBody(enc *bin.Encoder) error {
	if err := enc.WriteBytes(s.NodePubkey[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.AuthorizedWithdrawer[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.InflationRewardsCollector[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(s.BlockRevenueCollector[:], false); err != nil {
		return err
	}
	if err := enc.WriteUint16(s.InflationRewardsCommissionBps, binary.LittleEndian); err != nil {
		return err
	}
	if err := enc.WriteUint16(s.BlockRevenueCommissionBps, binary.LittleEndian); err != nil {
		return err
	}
	if err := enc.WriteUint64(s.PendingDelegatorRewards, binary.LittleEndian); err != nil {
		return err
	}
	if s.BLSPubkeyCompressed == nil {
		if err := enc.WriteUint8(0); err != nil {
			return err
		}
	} else {
		if err := enc.WriteUint8(1); err != nil {
			return err
		}
		if err := enc.WriteBytes(s.BLSPubkeyCompressed[:], false); err != nil {
			return err
		}
	}
	if err := enc.WriteUint64(uint64(len(s.Votes)), binary.LittleEndian); err != nil {
		return err
	}
	for _, v := range s.Votes {
		if err := v.MarshalWithEncoder(enc); err != nil {
			return err
		}
	}
	if err := encodeOptionU64(enc, s.RootSlot); err != nil {
		return err
	}
	if err := s.AuthorizedVoters.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := encodeEpochCredits(enc, s.EpochCredits); err != nil {
		return err
	}
	return s.LastTimestamp.MarshalWithEncoder(enc)
}

// UnmarshalWithDecoder reads the version discriminator and dispatches to the appropriate variant.
func (v *VoteStateVersions) UnmarshalWithDecoder(dec *bin.Decoder) error {
	raw, err := dec.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	v.Version = VoteStateVersion(raw)
	switch v.Version {
	case VoteStateVersionV1_14_11:
		v.V1_14_11 = new(VoteState1_14_11)
		return v.V1_14_11.unmarshalBody(dec)
	case VoteStateVersionV3:
		v.V3 = new(VoteStateV3)
		return v.V3.unmarshalBody(dec)
	case VoteStateVersionV4:
		v.V4 = new(VoteStateV4)
		return v.V4.unmarshalBody(dec)
	case VoteStateVersionV0_23_5:
		return fmt.Errorf("vote state version V0_23_5 is not supported")
	default:
		return fmt.Errorf("unknown vote state version: %d", v.Version)
	}
}

// MarshalWithEncoder writes the version discriminator and the selected variant.
func (v VoteStateVersions) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteUint32(uint32(v.Version), binary.LittleEndian); err != nil {
		return err
	}
	switch v.Version {
	case VoteStateVersionV1_14_11:
		if v.V1_14_11 == nil {
			return fmt.Errorf("V1_14_11 field is nil")
		}
		return v.V1_14_11.marshalBody(enc)
	case VoteStateVersionV3:
		if v.V3 == nil {
			return fmt.Errorf("V3 field is nil")
		}
		return v.V3.marshalBody(enc)
	case VoteStateVersionV4:
		if v.V4 == nil {
			return fmt.Errorf("V4 field is nil")
		}
		return v.V4.marshalBody(enc)
	default:
		return fmt.Errorf("cannot marshal vote state version: %d", v.Version)
	}
}

// DecodeVoteAccount decodes a vote account's raw data into a VoteStateVersions.
func DecodeVoteAccount(data []byte) (*VoteStateVersions, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("vote account data too short: %d", len(data))
	}
	state := new(VoteStateVersions)
	if err := state.UnmarshalWithDecoder(bin.NewBinDecoder(data)); err != nil {
		return nil, fmt.Errorf("unable to decode vote account: %w", err)
	}
	return state, nil
}
