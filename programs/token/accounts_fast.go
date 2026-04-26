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

package token

import (
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
)

// SPL-Token on-wire account sizes. These are stable-by-consensus (any
// change would be a breaking protocol fork). Using fixed constants
// lets the POD path bounds-check at a single compare-and-branch.
const (
	MintAccountSize     = 82
	TokenAccountSize    = 165
	MultisigAccountSize = 3 + MAX_SIGNERS*32
)

// MintRaw is the POD-safe wire layout of an SPL-Token Mint account
// (82 bytes). All multi-byte integer fields are declared as [N]byte
// so Go does not insert alignment padding — Go size == wire size.
// This is what makes the type safe for binary.ViewAs / MarshalPOD.
//
// Accessors (SupplyU64, MintAuthorityKey, ...) provide typed views
// over the byte fields. Each one is a single binary.LittleEndian.Uint*
// call and inlines to a straight load on amd64/arm64.
//
// Use ViewMintRaw to alias an incoming RPC / Geyser buffer (zero copy)
// or DecodeMintRaw to copy into a caller-owned MintRaw (detached).
type MintRaw struct {
	MintAuthorityTag   [4]byte  // COption tag: 0 = None, 1 = Some
	MintAuthority      [32]byte // only valid when MintAuthorityTag == 1
	Supply             [8]byte  // u64 LE
	Decimals           uint8
	IsInitialized      uint8 // 1 = true, 0 = false
	FreezeAuthorityTag [4]byte
	FreezeAuthority    [32]byte
}

// AccountRaw is the POD-safe wire layout of an SPL-Token Account
// (165 bytes). Same design rationale as MintRaw.
type AccountRaw struct {
	Mint              [32]byte
	Owner             [32]byte
	Amount            [8]byte // u64 LE
	DelegateTag       [4]byte
	Delegate          [32]byte
	State             uint8 // AccountState
	IsNativeTag       [4]byte
	IsNative          [8]byte // u64 LE, rent-exempt reserve when tag == 1
	DelegatedAmount   [8]byte // u64 LE
	CloseAuthorityTag [4]byte
	CloseAuthority    [32]byte
}

// MultisigRaw is the POD-safe wire layout of an SPL-Token Multisig
// account (3 + MAX_SIGNERS*32 = 355 bytes with MAX_SIGNERS=11).
type MultisigRaw struct {
	M             uint8 // number of signatures required
	N             uint8 // number of valid signers
	IsInitialized uint8
	Signers       [MAX_SIGNERS][32]byte
}

func init() {
	// Asserts that the Go memory layout of each Raw type matches the
	// SPL wire layout with no compiler-inserted padding. Runs once at
	// program start; fires a panic that points to the offending field
	// if any regression lands.
	bin.MustAssertPOD[MintRaw]()
	bin.MustAssertPOD[AccountRaw]()
	bin.MustAssertPOD[MultisigRaw]()
}

// ---- MintRaw accessors ----

func (m *MintRaw) SupplyU64() uint64 { return binary.LittleEndian.Uint64(m.Supply[:]) }

func (m *MintRaw) HasMintAuthority() bool {
	return binary.LittleEndian.Uint32(m.MintAuthorityTag[:]) != 0
}

func (m *MintRaw) MintAuthorityKey() solana.PublicKey { return solana.PublicKey(m.MintAuthority) }

func (m *MintRaw) HasFreezeAuthority() bool {
	return binary.LittleEndian.Uint32(m.FreezeAuthorityTag[:]) != 0
}

func (m *MintRaw) FreezeAuthorityKey() solana.PublicKey { return solana.PublicKey(m.FreezeAuthority) }

// ToMint materializes the pointered Mint struct (the existing public
// API type). Allocates one *PublicKey per present COption.
func (m *MintRaw) ToMint() Mint {
	out := Mint{
		Supply:        m.SupplyU64(),
		Decimals:      m.Decimals,
		IsInitialized: m.IsInitialized != 0,
	}
	if m.HasMintAuthority() {
		k := m.MintAuthorityKey()
		out.MintAuthority = &k
	}
	if m.HasFreezeAuthority() {
		k := m.FreezeAuthorityKey()
		out.FreezeAuthority = &k
	}
	return out
}

// ---- AccountRaw accessors ----

func (a *AccountRaw) MintKey() solana.PublicKey  { return solana.PublicKey(a.Mint) }
func (a *AccountRaw) OwnerKey() solana.PublicKey { return solana.PublicKey(a.Owner) }
func (a *AccountRaw) AmountU64() uint64          { return binary.LittleEndian.Uint64(a.Amount[:]) }

func (a *AccountRaw) HasDelegate() bool {
	return binary.LittleEndian.Uint32(a.DelegateTag[:]) != 0
}

func (a *AccountRaw) DelegateKey() solana.PublicKey { return solana.PublicKey(a.Delegate) }

func (a *AccountRaw) AccountState() AccountState { return AccountState(a.State) }

// IsNativeU64 returns the rent-exempt reserve and whether the account
// is a wrapped-SOL native account.
func (a *AccountRaw) IsNativeU64() (reserve uint64, isNative bool) {
	isNative = binary.LittleEndian.Uint32(a.IsNativeTag[:]) != 0
	reserve = binary.LittleEndian.Uint64(a.IsNative[:])
	return
}

func (a *AccountRaw) DelegatedAmountU64() uint64 {
	return binary.LittleEndian.Uint64(a.DelegatedAmount[:])
}

func (a *AccountRaw) HasCloseAuthority() bool {
	return binary.LittleEndian.Uint32(a.CloseAuthorityTag[:]) != 0
}

func (a *AccountRaw) CloseAuthorityKey() solana.PublicKey {
	return solana.PublicKey(a.CloseAuthority)
}

// ToAccount materializes the pointered Account struct. Allocates one
// *PublicKey or *uint64 per present COption.
func (a *AccountRaw) ToAccount() Account {
	out := Account{
		Mint:            a.MintKey(),
		Owner:           a.OwnerKey(),
		Amount:          a.AmountU64(),
		State:           a.AccountState(),
		DelegatedAmount: a.DelegatedAmountU64(),
	}
	if a.HasDelegate() {
		k := a.DelegateKey()
		out.Delegate = &k
	}
	if reserve, native := a.IsNativeU64(); native {
		v := reserve
		out.IsNative = &v
	}
	if a.HasCloseAuthority() {
		k := a.CloseAuthorityKey()
		out.CloseAuthority = &k
	}
	return out
}

// ---- View / Decode helpers ----

// ViewMintRaw reinterprets data as a *MintRaw. The returned pointer
// aliases data's backing array; mutations write through. data may be
// longer than MintAccountSize — trailing bytes are ignored.
//
// See binary.ViewAs for the alignment/lifetime contract. Safe on
// amd64/arm64; requires len(data) >= MintAccountSize.
func ViewMintRaw(data []byte) (*MintRaw, error) {
	if len(data) < MintAccountSize {
		return nil, fmt.Errorf("token: mint data too short: have %d, need %d", len(data), MintAccountSize)
	}
	return bin.ViewAs[MintRaw](data)
}

// ViewAccountRaw reinterprets data as a *AccountRaw. See ViewMintRaw.
func ViewAccountRaw(data []byte) (*AccountRaw, error) {
	if len(data) < TokenAccountSize {
		return nil, fmt.Errorf("token: account data too short: have %d, need %d", len(data), TokenAccountSize)
	}
	return bin.ViewAs[AccountRaw](data)
}

// ViewMultisigRaw reinterprets data as a *MultisigRaw. See ViewMintRaw.
func ViewMultisigRaw(data []byte) (*MultisigRaw, error) {
	if len(data) < MultisigAccountSize {
		return nil, fmt.Errorf("token: multisig data too short: have %d, need %d", len(data), MultisigAccountSize)
	}
	return bin.ViewAs[MultisigRaw](data)
}

// DecodeMintRaw copies the mint wire bytes into m. Detached from data
// (safe to mutate data afterwards). Inverse: binary.MarshalPOD(m, buf).
func DecodeMintRaw(m *MintRaw, data []byte) error { return bin.UnmarshalPOD(m, data) }

// DecodeAccountRaw copies the account wire bytes into a.
func DecodeAccountRaw(a *AccountRaw, data []byte) error { return bin.UnmarshalPOD(a, data) }

// DecodeMultisigRaw copies the multisig wire bytes into m.
func DecodeMultisigRaw(m *MultisigRaw, data []byte) error { return bin.UnmarshalPOD(m, data) }

// DecodeMintFast decodes an SPL Mint account into the pointered Mint
// struct via the POD fast path. Drop-in replacement for the
// traditional `NewBinDecoder(data).Decode(&mint)` flow.
//
// One memcpy (82 bytes) + up to two *PublicKey allocations for present
// COption fields. ~10-50x faster than the hand-written decoder.
func DecodeMintFast(data []byte) (Mint, error) {
	var raw MintRaw
	if err := DecodeMintRaw(&raw, data); err != nil {
		return Mint{}, err
	}
	return raw.ToMint(), nil
}

// DecodeAccountFast decodes an SPL Token Account via the POD fast
// path. Drop-in replacement for the traditional flow.
func DecodeAccountFast(data []byte) (Account, error) {
	var raw AccountRaw
	if err := DecodeAccountRaw(&raw, data); err != nil {
		return Account{}, err
	}
	return raw.ToAccount(), nil
}
