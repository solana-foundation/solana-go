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
	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
)

// FeesSize is the serialized size, in bytes, of the Fees sysvar (one
// FeeCalculator).
const FeesSize = 8

// Fees is the data of the (deprecated) Fees sysvar (account solana.SysVarFeesPubkey):
// the fee calculator for the current slot.
//
// Deprecated: the Fees sysvar is deprecated; use the getFeeForMessage RPC
// (rpc.GetFeeForMessage) to estimate transaction fees instead.
type Fees struct {
	FeeCalculator solana.FeeCalculator
}

func (f Fees) MarshalWithEncoder(encoder *bin.Encoder) error {
	return f.FeeCalculator.MarshalWithEncoder(encoder)
}

func (f *Fees) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	return f.FeeCalculator.UnmarshalWithDecoder(decoder)
}

func (f Fees) MarshalBinary() ([]byte, error) { return encodeSysvar(f) }
func (f *Fees) UnmarshalBinary(data []byte) error {
	return f.UnmarshalWithDecoder(bin.NewBinDecoder(data))
}

// DecodeFees decodes Fees sysvar account data.
//
// Deprecated: see Fees.
func DecodeFees(data []byte) (*Fees, error) {
	var f Fees
	if err := f.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return &f, nil
}
