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

package sysvar

// SIMD-0525 slot-time reduction; ported from agave runtime/src/slot_params.rs.
// Assumes the default 400ms genesis slot time (true for mainnet-beta, testnet
// and devnet); clusters with a custom genesis baseline are not modeled.

import (
	"github.com/gagliardetto/solana-go"
)

// Slot durations selectable by the SIMD-0525 feature gates
// (the baseline is DefaultMsPerSlot).
const (
	MsPerSlot350 uint64 = 350
	MsPerSlot300 uint64 = 300
	MsPerSlot250 uint64 = 250
	MsPerSlot200 uint64 = 200
)

// SlotTimeGate pairs a slot-time reduction feature gate with its slot duration.
type SlotTimeGate struct {
	FeatureID solana.PublicKey
	MsPerSlot uint64
}

var slotTimeGates = [4]SlotTimeGate{
	{solana.FeatureReduceSlotTimeTo350ms, MsPerSlot350},
	{solana.FeatureReduceSlotTimeTo300ms, MsPerSlot300},
	{solana.FeatureReduceSlotTimeTo250ms, MsPerSlot250},
	{solana.FeatureReduceSlotTimeTo200ms, MsPerSlot200},
}

// SlotTimeGates returns the reduction gates in intended activation order.
func SlotTimeGates() []SlotTimeGate {
	gates := slotTimeGates
	return gates[:]
}

// MsPerSlotAt returns the slot duration in effect at slot, given the set of
// activated features (fetch the gate accounts and decode them with
// solana.FeatureFromAccount; a nil set means no reductions).
//
// A gate activated in epoch E takes effect at the first slot of epoch E+1.
// The result is the minimum duration among effective gates, so it never
// increases as slots advance regardless of activation order (agave parity).
func MsPerSlotAt(slot uint64, es EpochSchedule, features solana.FeatureSet) uint64 {
	if es.SlotsPerEpoch == 0 {
		// Not a valid schedule (e.g. a zero value): claim no reduction
		// rather than treating every gate as effective at slot 0.
		return DefaultMsPerSlot
	}
	ms := DefaultMsPerSlot
	for _, gate := range slotTimeGates {
		activationSlot, ok := features.ActivatedSlot(gate.FeatureID)
		if !ok {
			continue
		}
		effectiveSlot := es.GetFirstSlotInEpoch(satAddU64(es.GetEpoch(activationSlot), 1))
		if effectiveSlot <= slot && gate.MsPerSlot < ms {
			ms = gate.MsPerSlot
		}
	}
	return ms
}

// MsPerSlotInEpoch returns the slot duration in effect during epoch.
// Gates only take effect at epoch boundaries, so the duration is constant
// within an epoch.
func MsPerSlotInEpoch(epoch uint64, es EpochSchedule, features solana.FeatureSet) uint64 {
	return MsPerSlotAt(es.GetFirstSlotInEpoch(epoch), es, features)
}
