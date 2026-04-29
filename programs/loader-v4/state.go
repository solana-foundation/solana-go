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

package loaderv4

import (
	"encoding/binary"
	"fmt"

	ag_solanago "github.com/gagliardetto/solana-go"
)

// LoaderV4Status mirrors `#[repr(u64)]` upstream. Serialized as u64 LE.
type LoaderV4Status uint64

const (
	StatusRetracted LoaderV4Status = 0
	StatusDeployed  LoaderV4Status = 1
	StatusFinalized LoaderV4Status = 2
)

func (s LoaderV4Status) String() string {
	switch s {
	case StatusRetracted:
		return "Retracted"
	case StatusDeployed:
		return "Deployed"
	case StatusFinalized:
		return "Finalized"
	default:
		return fmt.Sprintf("Unknown(%d)", uint64(s))
	}
}

// ProgramDataOffset is the fixed byte offset at which program bytes begin in
// a loader-v4 account. It equals the in-memory size of LoaderV4State.
const ProgramDataOffset = 48

// LoaderV4State is read directly (zero-copy) from the start of a loader-v4
// account. Unlike loader-v3 state, this struct is NOT bincode-encoded —
// upstream uses `#[repr(C)]` and casts the account data pointer. The layout
// is u64 slot | [32]u8 authority | u64 status.
type LoaderV4State struct {
	Slot                          uint64
	AuthorityAddressOrNextVersion ag_solanago.PublicKey
	Status                        LoaderV4Status
}

// UnpackLoaderV4State parses the 48-byte header at the start of a loader-v4
// account's data. Extra bytes after the header are ignored (they are the
// program ELF payload).
func UnpackLoaderV4State(data []byte) (*LoaderV4State, error) {
	if len(data) < ProgramDataOffset {
		return nil, fmt.Errorf("loader-v4 state too short: %d < %d", len(data), ProgramDataOffset)
	}
	s := &LoaderV4State{
		Slot:   binary.LittleEndian.Uint64(data[0:8]),
		Status: LoaderV4Status(binary.LittleEndian.Uint64(data[40:48])),
	}
	copy(s.AuthorityAddressOrNextVersion[:], data[8:40])
	return s, nil
}

// Pack emits the 48-byte header in upstream's `#[repr(C)]` layout. Callers
// that also need program bytes should append them to the returned slice.
func (s *LoaderV4State) Pack() []byte {
	out := make([]byte, ProgramDataOffset)
	binary.LittleEndian.PutUint64(out[0:8], s.Slot)
	copy(out[8:40], s.AuthorityAddressOrNextVersion[:])
	binary.LittleEndian.PutUint64(out[40:48], uint64(s.Status))
	return out
}
