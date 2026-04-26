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

package system

import (
	"fmt"

	bin "github.com/gagliardetto/solana-go/binary"
)

// Durable-nonce wire layout (stable-by-consensus, any change would be
// a breaking fork):
//
//	u32 LE  Version
//	u32 LE  State
//	[32]    AuthorizedPubkey
//	[32]    Nonce
//	u64 LE  FeeCalculator.LamportsPerSignature
//	------- 80 bytes total
//
// The existing NonceAccount struct already matches this layout exactly
// — no COption fields, no variable-length data, no Go-compiler-inserted
// padding. So we can reuse it as the POD type directly; there is no
// separate "Raw" struct like token/ has.
const NonceAccountSize = 80

func init() {
	// Catch future field reorders that would introduce padding.
	bin.MustAssertPOD[NonceAccount]()
}

// ViewNonceAccount reinterprets data as a *NonceAccount aliasing the
// buffer's backing array. Zero copy; mutations write through.
//
// data may be longer than NonceAccountSize — trailing bytes are ignored.
// See binary.ViewAs for the alignment/lifetime contract. Safe on
// amd64/arm64; requires len(data) >= NonceAccountSize.
func ViewNonceAccount(data []byte) (*NonceAccount, error) {
	if len(data) < NonceAccountSize {
		return nil, fmt.Errorf("system: nonce account data too short: have %d, need %d", len(data), NonceAccountSize)
	}
	return bin.ViewAs[NonceAccount](data)
}

// DecodeNonceAccountFast decodes a durable-nonce account into a
// detached NonceAccount via a single memcpy. Drop-in replacement for
// `bin.NewBinDecoder(data).Decode(&nonce)` with no allocations.
//
// Returned value is fully detached — mutating data afterwards does
// not affect it.
func DecodeNonceAccountFast(data []byte) (NonceAccount, error) {
	var out NonceAccount
	if err := bin.UnmarshalPOD(&out, data); err != nil {
		return NonceAccount{}, err
	}
	return out, nil
}

// DecodeNonceAccountRaw is DecodeNonceAccountFast into a caller-owned
// destination. Zero allocations, zero return-value copy.
func DecodeNonceAccountRaw(n *NonceAccount, data []byte) error {
	return bin.UnmarshalPOD(n, data)
}
