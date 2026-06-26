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
	"math"
	"math/bits"

	bin "github.com/gagliardetto/binary"
)

// RentSize is the serialized size, in bytes, of the Rent sysvar
// (u64 + f64 + u8).
const RentSize = 17

// Rent parameters from the solana-sdk rent crate.
const (
	// AccountStorageOverhead is the number of bytes charged for an account with
	// no data, added to data length when computing the rent-exempt minimum.
	AccountStorageOverhead uint64 = 128
	// DefaultLamportsPerByte is the current default rental rate in
	// lamports/byte (SIMD-0194).
	DefaultLamportsPerByte uint64 = 6_960
	// DefaultExemptionThreshold is the current default exemption threshold
	// (SIMD-0194 reduced this from 2.0 to 1.0).
	DefaultExemptionThreshold float64 = 1.0
	// DefaultBurnPercent is the percentage of collected rent that is burned.
	DefaultBurnPercent uint8 = 50
	// MaxPermittedDataLength is the maximum permitted account data length
	// (10 MiB) when computing a rent-exempt minimum.
	MaxPermittedDataLength uint64 = 10 * 1024 * 1024
)

// Rent is the data of the Rent sysvar (account solana.SysVarRentPubkey): the cluster's
// rent parameters. Mirrors solana-sdk rent::Rent.
//
// Note: SIMD-0194 renamed the first field from lamports_per_byte_year to
// lamports_per_byte and changed the default exemption threshold to 1.0. The
// wire format is unchanged: u64 lamports_per_byte, then exemption_threshold as
// the 8-byte little-endian IEEE-754 bit pattern of an f64, then u8 burn_percent.
type Rent struct {
	// LamportsPerByte is the rental rate in lamports per byte.
	LamportsPerByte uint64
	// ExemptionThreshold multiplies the base rent to yield the rent-exempt
	// minimum balance.
	ExemptionThreshold float64
	// BurnPercent is the percentage of collected rent that is burned.
	BurnPercent uint8
}

// DefaultRent returns the current default rent configuration.
func DefaultRent() Rent {
	return Rent{
		LamportsPerByte:    DefaultLamportsPerByte,
		ExemptionThreshold: DefaultExemptionThreshold,
		BurnPercent:        DefaultBurnPercent,
	}
}

// MinimumBalance returns the minimum balance, in lamports, for an account of
// dataLen bytes to be rent-exempt:
// floor((ACCOUNT_STORAGE_OVERHEAD + dataLen) * LamportsPerByte * ExemptionThreshold).
//
// The two default thresholds (1.0 and 2.0) use exact integer math; other
// thresholds fall back to floating point. If the computation would overflow
// uint64 (e.g. a fixture or hostile account with a huge LamportsPerByte), it
// saturates to math.MaxUint64 so an underfunded account is never reported
// rent-exempt. Use TryMinimumBalance to detect the overflow explicitly.
func (r Rent) MinimumBalance(dataLen uint64) uint64 {
	v, overflow := r.computeMinimumBalance(dataLen)
	if overflow {
		return math.MaxUint64
	}
	return v
}

// TryMinimumBalance is MinimumBalance but reports ok=false when dataLen exceeds
// MaxPermittedDataLength or the computation would overflow uint64. Mirrors the
// guards in Rent::try_minimum_balance.
func (r Rent) TryMinimumBalance(dataLen uint64) (uint64, bool) {
	if dataLen > MaxPermittedDataLength {
		return 0, false
	}
	v, overflow := r.computeMinimumBalance(dataLen)
	return v, !overflow
}

// computeMinimumBalance returns the rent-exempt minimum and whether computing it
// overflowed uint64.
func (r Rent) computeMinimumBalance(dataLen uint64) (uint64, bool) {
	bytes := AccountStorageOverhead + dataLen
	if bytes < dataLen { // overflow in the byte count
		return 0, true
	}
	hi, base := bits.Mul64(bytes, r.LamportsPerByte)
	if hi != 0 { // overflow in base rent
		return 0, true
	}
	switch r.ExemptionThreshold {
	case 1.0:
		return base, false
	case 2.0:
		if base > math.MaxUint64/2 {
			return 0, true
		}
		return base * 2, false
	default:
		v := float64(base) * r.ExemptionThreshold
		if math.IsNaN(v) || v < 0 || v >= float64(math.MaxUint64) {
			return 0, true
		}
		return uint64(v), false
	}
}

// IsExempt reports whether balance is enough for an account of dataLen bytes to
// be rent-exempt. Because MinimumBalance saturates on overflow, IsExempt never
// reports an underfunded account as exempt.
func (r Rent) IsExempt(balance, dataLen uint64) bool {
	return balance >= r.MinimumBalance(dataLen)
}

func (r Rent) MarshalWithEncoder(encoder *bin.Encoder) error {
	if err := encoder.WriteUint64(r.LamportsPerByte, binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteFloat64(r.ExemptionThreshold, binary.LittleEndian); err != nil {
		return err
	}
	return encoder.WriteByte(r.BurnPercent)
}

func (r *Rent) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	var err error
	if r.LamportsPerByte, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	if r.ExemptionThreshold, err = decoder.ReadFloat64(binary.LittleEndian); err != nil {
		return err
	}
	r.BurnPercent, err = decoder.ReadByte()
	return err
}

func (r Rent) MarshalBinary() ([]byte, error) { return encodeSysvar(r) }
func (r *Rent) UnmarshalBinary(data []byte) error {
	return r.UnmarshalWithDecoder(bin.NewBinDecoder(data))
}

// DecodeRent decodes Rent sysvar account data.
func DecodeRent(data []byte) (*Rent, error) {
	var r Rent
	if err := r.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return &r, nil
}
