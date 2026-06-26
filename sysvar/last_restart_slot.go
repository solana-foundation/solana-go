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
)

// LastRestartSlotSize is the serialized size, in bytes, of the LastRestartSlot
// sysvar.
const LastRestartSlotSize = 8

// LastRestartSlot is the data of the LastRestartSlot sysvar (account
// solana.SysVarLastRestartSlotPubkey): the slot of the last cluster restart, or 0 if
// none. Mirrors solana-sdk last_restart_slot::LastRestartSlot.
type LastRestartSlot struct {
	LastRestartSlot uint64
}

func (l LastRestartSlot) MarshalWithEncoder(encoder *bin.Encoder) error {
	return encoder.WriteUint64(l.LastRestartSlot, binary.LittleEndian)
}

func (l *LastRestartSlot) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	var err error
	l.LastRestartSlot, err = decoder.ReadUint64(binary.LittleEndian)
	return err
}

func (l LastRestartSlot) MarshalBinary() ([]byte, error) { return encodeSysvar(l) }
func (l *LastRestartSlot) UnmarshalBinary(data []byte) error {
	return l.UnmarshalWithDecoder(bin.NewBinDecoder(data))
}

// DecodeLastRestartSlot decodes LastRestartSlot sysvar account data.
func DecodeLastRestartSlot(data []byte) (*LastRestartSlot, error) {
	var l LastRestartSlot
	if err := l.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return &l, nil
}
