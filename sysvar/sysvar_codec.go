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

// Shared helpers for the sysvar account-data types (clock.go, rent.go,
// epoch_schedule.go, slot_hashes.go, slot_history.go, stake_history.go,
// epoch_rewards.go, last_restart_slot.go, fees.go, recent_blockhashes.go).
//
// All sysvar account data uses standard bincode: little-endian integers and a
// u64 little-endian length prefix for Vec<T> (NOT the compact-u16/short-vec
// encoding used in transaction wire formats).

import (
	"bytes"
	"math"
	"math/bits"

	bin "github.com/gagliardetto/binary"
)

// encodeSysvar bincode-serializes a sysvar value via its MarshalWithEncoder.
func encodeSysvar(m interface{ MarshalWithEncoder(*bin.Encoder) error }) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.MarshalWithEncoder(bin.NewBinEncoder(buf)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Saturating unsigned arithmetic mirroring Rust's saturating_add/sub/mul. Used
// by the EpochSchedule slot/epoch math so results match the runtime exactly.

func satAddU64(a, b uint64) uint64 {
	s := a + b
	if s < a {
		return math.MaxUint64
	}
	return s
}

func satSubU64(a, b uint64) uint64 {
	if a < b {
		return 0
	}
	return a - b
}

func satMulU64(a, b uint64) uint64 {
	if a == 0 || b == 0 {
		return 0
	}
	c := a * b
	if c/a != b {
		return math.MaxUint64
	}
	return c
}

func satAddU32(a, b uint32) uint32 {
	s := a + b
	if s < a {
		return math.MaxUint32
	}
	return s
}

func satSubU32(a, b uint32) uint32 {
	if a < b {
		return 0
	}
	return a - b
}

// nextPowerOfTwo returns the smallest power of two >= n (n<=1 -> 1), matching
// Rust's u64::next_power_of_two.
func nextPowerOfTwo(n uint64) uint64 {
	if n <= 1 {
		return 1
	}
	return uint64(1) << bits.Len64(n-1)
}

// satPow2 returns 2^exp, saturating at math.MaxUint64, matching
// 2u64.saturating_pow(exp).
func satPow2(exp uint32) uint64 {
	if exp >= 64 {
		return math.MaxUint64
	}
	return uint64(1) << exp
}

// numTrailingZeros64 mirrors u64::trailing_zeros.
func numTrailingZeros64(n uint64) int {
	return bits.TrailingZeros64(n)
}
