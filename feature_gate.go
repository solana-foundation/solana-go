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

// Ported from solana-sdk feature-gate-interface (state.rs).

import (
	"fmt"

	bin "github.com/gagliardetto/binary"
)

// SIMD-0525 slot-time reduction feature gates, in activation order.
// Each takes effect one epoch after its activation epoch (see sysvar.MsPerSlotAt).
var (
	FeatureReduceSlotTimeTo350ms = MustPublicKeyFromBase58("iBRL5RuWhw4yqaAZu96RUULHckHTZAoe2b77qaV38JZ")
	FeatureReduceSlotTimeTo300ms = MustPublicKeyFromBase58("iBRLL3k18HST852F1Mf3Lv83waTNQmmqvKDxvYGwQFL")
	FeatureReduceSlotTimeTo250ms = MustPublicKeyFromBase58("iBRLMc81UjRa8fn8A6eE8bJTnRbgQoPTynM51akENCV")
	FeatureReduceSlotTimeTo200ms = MustPublicKeyFromBase58("iBRLjhJnkmDZgNoZRDMW11d8ZV7HvsL3vAyRjZB5npW")
)

// FeatureSizeOf is the serialized size of a Feature account (1-byte option tag + u64).
const FeatureSizeOf = 9

// Feature is the state of a feature-gate account (owner FeatureProgramID).
type Feature struct {
	// Slot at which the feature was activated; nil while pending.
	ActivatedAt *uint64
}

// IsActive reports whether the feature has been activated.
func (f Feature) IsActive() bool {
	return f.ActivatedAt != nil
}

func (f Feature) MarshalWithEncoder(encoder *bin.Encoder) error {
	if f.ActivatedAt == nil {
		return encoder.WriteByte(0)
	}
	if err := encoder.WriteByte(1); err != nil {
		return err
	}
	return encoder.WriteUint64(*f.ActivatedAt, bin.LE)
}

func (f *Feature) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	// Bincode Option<u64>: strict 0/1 tag, lenient about trailing bytes.
	tag, err := decoder.ReadByte()
	if err != nil {
		return fmt.Errorf("unable to read feature option tag: %w", err)
	}
	switch tag {
	case 0:
		f.ActivatedAt = nil
		return nil
	case 1:
		v, err := decoder.ReadUint64(bin.LE)
		if err != nil {
			return fmt.Errorf("unable to read feature activation slot: %w", err)
		}
		f.ActivatedAt = &v
		return nil
	default:
		return fmt.Errorf("invalid feature option tag: %d", tag)
	}
}

// DecodeFeature decodes a bincode-encoded Feature (Option<u64>).
func DecodeFeature(data []byte) (Feature, error) {
	var f Feature
	err := f.UnmarshalWithDecoder(bin.NewBinDecoder(data))
	return f, err
}

// FeatureFromAccount decodes a Feature after checking the account owner,
// mirroring feature-gate-interface from_account.
func FeatureFromAccount(owner PublicKey, data []byte) (Feature, error) {
	if !owner.Equals(FeatureProgramID) {
		return Feature{}, fmt.Errorf("invalid feature account owner: %s", owner)
	}
	// Upstream from_account enforces the minimum account size even for
	// pending features, beyond what the lenient bincode decode requires.
	if len(data) < FeatureSizeOf {
		return Feature{}, fmt.Errorf("feature account data too short: %d bytes, need %d", len(data), FeatureSizeOf)
	}
	return DecodeFeature(data)
}

// FeatureSet maps feature gate IDs to their activation slot,
// a minimal counterpart of the Rust FeatureSet's active map.
// The zero value (nil) is a valid, empty set.
type FeatureSet map[PublicKey]uint64

// ActivatedSlot returns the activation slot of the feature, if activated.
func (fs FeatureSet) ActivatedSlot(id PublicKey) (uint64, bool) {
	slot, ok := fs[id]
	return slot, ok
}

// IsActive reports whether the feature is activated.
func (fs FeatureSet) IsActive(id PublicKey) bool {
	_, ok := fs[id]
	return ok
}
