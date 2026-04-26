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
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
)

// Constants from the official vote program.
const (
	MAX_LOCKOUT_HISTORY                     = 31
	INITIAL_LOCKOUT                         = 2
	MAX_EPOCH_CREDITS_HISTORY               = 64
	VOTE_CREDITS_GRACE_SLOTS                = 2
	VOTE_CREDITS_MAXIMUM_PER_SLOT           = 16
	CircBufMaxItems                         = 32
	BLS_PUBLIC_KEY_COMPRESSED_SIZE          = 48
	BLS_PROOF_OF_POSSESSION_COMPRESSED_SIZE = 96
	VoteStateV1_14_11Size                   = 3731
	VoteStateV3Size                         = 3762
	VoteStateV4Size                         = 3762
)

// Instruction IDs for VoteInstruction variants.
const (
	Instruction_InitializeAccount uint32 = iota
	Instruction_Authorize
	Instruction_Vote
	Instruction_Withdraw
	Instruction_UpdateValidatorIdentity
	Instruction_UpdateCommission
	Instruction_VoteSwitch
	Instruction_AuthorizeChecked
	Instruction_UpdateVoteState
	Instruction_UpdateVoteStateSwitch
	Instruction_AuthorizeWithSeed
	Instruction_AuthorizeCheckedWithSeed
	Instruction_CompactUpdateVoteState
	Instruction_CompactUpdateVoteStateSwitch
	Instruction_TowerSync
	Instruction_TowerSyncSwitch
	Instruction_InitializeAccountV2
	Instruction_UpdateCommissionCollector
	Instruction_UpdateCommissionBps
	Instruction_DepositDelegatorRewards
)

// InstructionIDToName returns the name of the instruction given its ID.
func InstructionIDToName(id uint32) string {
	switch id {
	case Instruction_InitializeAccount:
		return "InitializeAccount"
	case Instruction_Authorize:
		return "Authorize"
	case Instruction_Vote:
		return "Vote"
	case Instruction_Withdraw:
		return "Withdraw"
	case Instruction_UpdateValidatorIdentity:
		return "UpdateValidatorIdentity"
	case Instruction_UpdateCommission:
		return "UpdateCommission"
	case Instruction_VoteSwitch:
		return "VoteSwitch"
	case Instruction_AuthorizeChecked:
		return "AuthorizeChecked"
	case Instruction_UpdateVoteState:
		return "UpdateVoteState"
	case Instruction_UpdateVoteStateSwitch:
		return "UpdateVoteStateSwitch"
	case Instruction_AuthorizeWithSeed:
		return "AuthorizeWithSeed"
	case Instruction_AuthorizeCheckedWithSeed:
		return "AuthorizeCheckedWithSeed"
	case Instruction_CompactUpdateVoteState:
		return "CompactUpdateVoteState"
	case Instruction_CompactUpdateVoteStateSwitch:
		return "CompactUpdateVoteStateSwitch"
	case Instruction_TowerSync:
		return "TowerSync"
	case Instruction_TowerSyncSwitch:
		return "TowerSyncSwitch"
	case Instruction_InitializeAccountV2:
		return "InitializeAccountV2"
	case Instruction_UpdateCommissionCollector:
		return "UpdateCommissionCollector"
	case Instruction_UpdateCommissionBps:
		return "UpdateCommissionBps"
	case Instruction_DepositDelegatorRewards:
		return "DepositDelegatorRewards"
	default:
		return ""
	}
}

// VoteAuthorizeKind identifies the type of authorization being set.
type VoteAuthorizeKind uint32

const (
	VoteAuthorizeVoter        VoteAuthorizeKind = 0
	VoteAuthorizeWithdrawer   VoteAuthorizeKind = 1
	VoteAuthorizeVoterWithBLS VoteAuthorizeKind = 2
)

// VoteAuthorize represents the bincode-encoded VoteAuthorize enum.
// For Voter and Withdrawer variants, only Kind is set.
// For VoterWithBLS, BLSPubkey and BLSProofOfPossession are also set.
//
// Serialization is handled by custom MarshalWithEncoder/UnmarshalWithDecoder
// methods — struct tags are not used by this type.
type VoteAuthorize struct {
	Kind                 VoteAuthorizeKind
	BLSPubkey            *[BLS_PUBLIC_KEY_COMPRESSED_SIZE]byte
	BLSProofOfPossession *[BLS_PROOF_OF_POSSESSION_COMPRESSED_SIZE]byte
}

func (va *VoteAuthorize) UnmarshalWithDecoder(dec *bin.Decoder) error {
	raw, err := dec.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	va.Kind = VoteAuthorizeKind(raw)
	if va.Kind == VoteAuthorizeVoterWithBLS {
		pk := new([BLS_PUBLIC_KEY_COMPRESSED_SIZE]byte)
		v, err := dec.ReadNBytes(BLS_PUBLIC_KEY_COMPRESSED_SIZE)
		if err != nil {
			return err
		}
		copy(pk[:], v)
		va.BLSPubkey = pk

		proof := new([BLS_PROOF_OF_POSSESSION_COMPRESSED_SIZE]byte)
		v, err = dec.ReadNBytes(BLS_PROOF_OF_POSSESSION_COMPRESSED_SIZE)
		if err != nil {
			return err
		}
		copy(proof[:], v)
		va.BLSProofOfPossession = proof
	}
	return nil
}

func (va VoteAuthorize) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteUint32(uint32(va.Kind), binary.LittleEndian); err != nil {
		return err
	}
	if va.Kind == VoteAuthorizeVoterWithBLS {
		if va.BLSPubkey == nil || va.BLSProofOfPossession == nil {
			return errors.New("VoterWithBLS requires BLSPubkey and BLSProofOfPossession")
		}
		if err := enc.WriteBytes(va.BLSPubkey[:], false); err != nil {
			return err
		}
		if err := enc.WriteBytes(va.BLSProofOfPossession[:], false); err != nil {
			return err
		}
	}
	return nil
}

// CommissionKind identifies which commission bucket an operation targets.
type CommissionKind uint8

const (
	CommissionKindInflationRewards CommissionKind = 0
	CommissionKindBlockRevenue     CommissionKind = 1
)

// VoteInit is the data for the InitializeAccount instruction.
type VoteInit struct {
	NodePubkey           solana.PublicKey
	AuthorizedVoter      solana.PublicKey
	AuthorizedWithdrawer solana.PublicKey
	Commission           uint8
}

func (v *VoteInit) UnmarshalWithDecoder(dec *bin.Decoder) error {
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	v.NodePubkey = solana.PublicKeyFromBytes(b)
	b, err = dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	v.AuthorizedVoter = solana.PublicKeyFromBytes(b)
	b, err = dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	v.AuthorizedWithdrawer = solana.PublicKeyFromBytes(b)
	v.Commission, err = dec.ReadUint8()
	return err
}

func (v VoteInit) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteBytes(v.NodePubkey[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(v.AuthorizedVoter[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(v.AuthorizedWithdrawer[:], false); err != nil {
		return err
	}
	return enc.WriteUint8(v.Commission)
}

// VoteInitV2 is the data for the InitializeAccountV2 instruction (Alpenglow).
type VoteInitV2 struct {
	NodePubkey                          solana.PublicKey
	AuthorizedVoter                     solana.PublicKey
	AuthorizedVoterBLSPubkey            [BLS_PUBLIC_KEY_COMPRESSED_SIZE]byte
	AuthorizedVoterBLSProofOfPossession [BLS_PROOF_OF_POSSESSION_COMPRESSED_SIZE]byte
	AuthorizedWithdrawer                solana.PublicKey
	InflationRewardsCommissionBps       uint16
	BlockRevenueCommissionBps           uint16
}

func (v *VoteInitV2) UnmarshalWithDecoder(dec *bin.Decoder) error {
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	v.NodePubkey = solana.PublicKeyFromBytes(b)
	b, err = dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	v.AuthorizedVoter = solana.PublicKeyFromBytes(b)
	b, err = dec.ReadNBytes(BLS_PUBLIC_KEY_COMPRESSED_SIZE)
	if err != nil {
		return err
	}
	copy(v.AuthorizedVoterBLSPubkey[:], b)
	b, err = dec.ReadNBytes(BLS_PROOF_OF_POSSESSION_COMPRESSED_SIZE)
	if err != nil {
		return err
	}
	copy(v.AuthorizedVoterBLSProofOfPossession[:], b)
	b, err = dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	v.AuthorizedWithdrawer = solana.PublicKeyFromBytes(b)
	v.InflationRewardsCommissionBps, err = dec.ReadUint16(binary.LittleEndian)
	if err != nil {
		return err
	}
	v.BlockRevenueCommissionBps, err = dec.ReadUint16(binary.LittleEndian)
	return err
}

func (v VoteInitV2) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteBytes(v.NodePubkey[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(v.AuthorizedVoter[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(v.AuthorizedVoterBLSPubkey[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(v.AuthorizedVoterBLSProofOfPossession[:], false); err != nil {
		return err
	}
	if err := enc.WriteBytes(v.AuthorizedWithdrawer[:], false); err != nil {
		return err
	}
	if err := enc.WriteUint16(v.InflationRewardsCommissionBps, binary.LittleEndian); err != nil {
		return err
	}
	return enc.WriteUint16(v.BlockRevenueCommissionBps, binary.LittleEndian)
}

// Lockout represents a vote lockout with slot and confirmation count.
type Lockout struct {
	Slot              uint64
	ConfirmationCount uint32
}

func (l *Lockout) UnmarshalWithDecoder(dec *bin.Decoder) error {
	var err error
	l.Slot, err = dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	l.ConfirmationCount, err = dec.ReadUint32(binary.LittleEndian)
	return err
}

func (l Lockout) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteUint64(l.Slot, binary.LittleEndian); err != nil {
		return err
	}
	return enc.WriteUint32(l.ConfirmationCount, binary.LittleEndian)
}

// LandedVote combines vote latency with a Lockout (used in VoteStateV3 and V4).
type LandedVote struct {
	Latency uint8
	Lockout Lockout
}

func (lv *LandedVote) UnmarshalWithDecoder(dec *bin.Decoder) error {
	var err error
	lv.Latency, err = dec.ReadUint8()
	if err != nil {
		return err
	}
	return lv.Lockout.UnmarshalWithDecoder(dec)
}

func (lv LandedVote) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteUint8(lv.Latency); err != nil {
		return err
	}
	return lv.Lockout.MarshalWithEncoder(enc)
}

// VoteStateUpdate is the data for the UpdateVoteState and related instructions.
type VoteStateUpdate struct {
	Lockouts  []Lockout
	Root      *uint64
	Hash      solana.Hash
	Timestamp *int64
}

func (u *VoteStateUpdate) UnmarshalWithDecoder(dec *bin.Decoder) error {
	count, err := dec.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	u.Lockouts = make([]Lockout, count)
	for i := range count {
		if err := u.Lockouts[i].UnmarshalWithDecoder(dec); err != nil {
			return err
		}
	}
	// Root: Option<u64>
	hasRoot, err := dec.ReadUint8()
	if err != nil {
		return err
	}
	if hasRoot == 1 {
		root, err := dec.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
		u.Root = &root
	} else if hasRoot != 0 {
		return fmt.Errorf("invalid Option<u64> discriminant: %d", hasRoot)
	}
	// Hash
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(u.Hash[:], b)
	// Timestamp: Option<i64>
	hasTs, err := dec.ReadUint8()
	if err != nil {
		return err
	}
	if hasTs == 1 {
		ts, err := dec.ReadInt64(binary.LittleEndian)
		if err != nil {
			return err
		}
		u.Timestamp = &ts
	} else if hasTs != 0 {
		return fmt.Errorf("invalid Option<i64> discriminant: %d", hasTs)
	}
	return nil
}

func (u VoteStateUpdate) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := enc.WriteUint64(uint64(len(u.Lockouts)), binary.LittleEndian); err != nil {
		return err
	}
	for _, l := range u.Lockouts {
		if err := l.MarshalWithEncoder(enc); err != nil {
			return err
		}
	}
	if u.Root == nil {
		if err := enc.WriteUint8(0); err != nil {
			return err
		}
	} else {
		if err := enc.WriteUint8(1); err != nil {
			return err
		}
		if err := enc.WriteUint64(*u.Root, binary.LittleEndian); err != nil {
			return err
		}
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

// TowerSyncUpdate is the data for the TowerSync and TowerSyncSwitch instructions.
// It represents a tower sync update (the current consensus mechanism) and
// adds a BlockID field to VoteStateUpdate.
//
// The wire format is the compact serde format (short_vec + delta-encoded
// lockout offsets + varint + trailing block_id) — see compact.go.
// Lockouts are stored here as absolute slots; delta encoding happens at
// marshal time.
type TowerSyncUpdate struct {
	Lockouts  []Lockout
	Root      *uint64
	Hash      solana.Hash
	Timestamp *int64
	BlockID   solana.Hash
}

func (t *TowerSyncUpdate) UnmarshalWithDecoder(dec *bin.Decoder) error {
	upd := VoteStateUpdate{}
	if err := unmarshalCompactVoteStateUpdate(dec, &upd); err != nil {
		return err
	}
	t.Lockouts = upd.Lockouts
	t.Root = upd.Root
	t.Hash = upd.Hash
	t.Timestamp = upd.Timestamp
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(t.BlockID[:], b)
	return nil
}

func (t TowerSyncUpdate) MarshalWithEncoder(enc *bin.Encoder) error {
	upd := VoteStateUpdate{
		Lockouts:  t.Lockouts,
		Root:      t.Root,
		Hash:      t.Hash,
		Timestamp: t.Timestamp,
	}
	if err := marshalCompactVoteStateUpdate(enc, &upd); err != nil {
		return err
	}
	return enc.WriteBytes(t.BlockID[:], false)
}

// VoteAuthorizeWithSeedArgs is the data for AuthorizeWithSeed.
type VoteAuthorizeWithSeedArgs struct {
	AuthorizationType               VoteAuthorize
	CurrentAuthorityDerivedKeyOwner solana.PublicKey
	CurrentAuthorityDerivedKeySeed  string
	NewAuthority                    solana.PublicKey
}

func (a *VoteAuthorizeWithSeedArgs) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := a.AuthorizationType.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	a.CurrentAuthorityDerivedKeyOwner = solana.PublicKeyFromBytes(b)
	seed, err := dec.ReadRustString()
	if err != nil {
		return err
	}
	a.CurrentAuthorityDerivedKeySeed = seed
	b, err = dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	a.NewAuthority = solana.PublicKeyFromBytes(b)
	return nil
}

func (a VoteAuthorizeWithSeedArgs) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := a.AuthorizationType.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := enc.WriteBytes(a.CurrentAuthorityDerivedKeyOwner[:], false); err != nil {
		return err
	}
	if err := enc.WriteRustString(a.CurrentAuthorityDerivedKeySeed); err != nil {
		return err
	}
	return enc.WriteBytes(a.NewAuthority[:], false)
}

// VoteAuthorizeCheckedWithSeedArgs is the data for AuthorizeCheckedWithSeed.
type VoteAuthorizeCheckedWithSeedArgs struct {
	AuthorizationType               VoteAuthorize
	CurrentAuthorityDerivedKeyOwner solana.PublicKey
	CurrentAuthorityDerivedKeySeed  string
}

func (a *VoteAuthorizeCheckedWithSeedArgs) UnmarshalWithDecoder(dec *bin.Decoder) error {
	if err := a.AuthorizationType.UnmarshalWithDecoder(dec); err != nil {
		return err
	}
	b, err := dec.ReadNBytes(32)
	if err != nil {
		return err
	}
	a.CurrentAuthorityDerivedKeyOwner = solana.PublicKeyFromBytes(b)
	seed, err := dec.ReadRustString()
	if err != nil {
		return err
	}
	a.CurrentAuthorityDerivedKeySeed = seed
	return nil
}

func (a VoteAuthorizeCheckedWithSeedArgs) MarshalWithEncoder(enc *bin.Encoder) error {
	if err := a.AuthorizationType.MarshalWithEncoder(enc); err != nil {
		return err
	}
	if err := enc.WriteBytes(a.CurrentAuthorityDerivedKeyOwner[:], false); err != nil {
		return err
	}
	return enc.WriteRustString(a.CurrentAuthorityDerivedKeySeed)
}
