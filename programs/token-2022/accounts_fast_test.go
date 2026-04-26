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

package token2022

import (
	"testing"
	"unsafe"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
)

func TestSPLLayoutSizes(t *testing.T) {
	if got := unsafe.Sizeof(MintRaw{}); got != MintAccountSize {
		t.Errorf("sizeof(MintRaw) = %d, want %d", got, MintAccountSize)
	}
	if got := unsafe.Sizeof(AccountRaw{}); got != TokenAccountSize {
		t.Errorf("sizeof(AccountRaw) = %d, want %d", got, TokenAccountSize)
	}
	if got := unsafe.Sizeof(MultisigRaw{}); got != MultisigAccountSize {
		t.Errorf("sizeof(MultisigRaw) = %d, want %d", got, MultisigAccountSize)
	}
}

func TestAssertPODPasses(t *testing.T) {
	if err := bin.AssertPOD[MintRaw](); err != nil {
		t.Error(err)
	}
	if err := bin.AssertPOD[AccountRaw](); err != nil {
		t.Error(err)
	}
	if err := bin.AssertPOD[MultisigRaw](); err != nil {
		t.Error(err)
	}
}

func buildMintWire(t *testing.T, mint *Mint) []byte {
	t.Helper()
	var buf [MintAccountSize]byte
	enc := bin.NewBinEncoderInto(buf[:])
	if err := mint.MarshalWithEncoder(enc); err != nil {
		t.Fatalf("encode mint: %v", err)
	}
	if enc.Written() != MintAccountSize {
		t.Fatalf("encoded mint is %d bytes, want %d", enc.Written(), MintAccountSize)
	}
	return buf[:]
}

func buildAccountWire(t *testing.T, acc *Account) []byte {
	t.Helper()
	var buf [TokenAccountSize]byte
	enc := bin.NewBinEncoderInto(buf[:])
	if err := acc.MarshalWithEncoder(enc); err != nil {
		t.Fatalf("encode account: %v", err)
	}
	if enc.Written() != TokenAccountSize {
		t.Fatalf("encoded account is %d bytes, want %d", enc.Written(), TokenAccountSize)
	}
	return buf[:]
}

func newTestKey(b byte) solana.PublicKey {
	var k solana.PublicKey
	for i := range k {
		k[i] = b + byte(i)
	}
	return k
}

// ---- Parity: decoded fast path matches decoded traditional path ----

func TestDecodeMintFast_ParityWithTraditional_AllSet(t *testing.T) {
	mintAuth := newTestKey(0x10)
	freezeAuth := newTestKey(0x20)
	orig := &Mint{
		MintAuthority:   &mintAuth,
		Supply:          1_000_000_000,
		Decimals:        9,
		IsInitialized:   true,
		FreezeAuthority: &freezeAuth,
	}
	wire := buildMintWire(t, orig)

	var trad Mint
	if err := bin.NewBinDecoder(wire).Decode(&trad); err != nil {
		t.Fatal(err)
	}
	fast, err := DecodeMintFast(wire)
	if err != nil {
		t.Fatal(err)
	}
	assertMintEqual(t, &trad, &fast)
}

func TestDecodeMintFast_ParityWithTraditional_NoOptionals(t *testing.T) {
	orig := &Mint{
		Supply:        42,
		Decimals:      6,
		IsInitialized: true,
	}
	wire := buildMintWire(t, orig)

	var trad Mint
	if err := bin.NewBinDecoder(wire).Decode(&trad); err != nil {
		t.Fatal(err)
	}
	fast, err := DecodeMintFast(wire)
	if err != nil {
		t.Fatal(err)
	}
	assertMintEqual(t, &trad, &fast)
}

func TestDecodeAccountFast_ParityWithTraditional_AllSet(t *testing.T) {
	mint := newTestKey(0x30)
	owner := newTestKey(0x40)
	delegate := newTestKey(0x50)
	closeAuth := newTestKey(0x60)
	native := uint64(2_039_280)
	orig := &Account{
		Mint:            mint,
		Owner:           owner,
		Amount:          999_999,
		Delegate:        &delegate,
		State:           AccountStateInitialized,
		IsNative:        &native,
		DelegatedAmount: 10_000,
		CloseAuthority:  &closeAuth,
	}
	wire := buildAccountWire(t, orig)

	var trad Account
	if err := bin.NewBinDecoder(wire).Decode(&trad); err != nil {
		t.Fatal(err)
	}
	fast, err := DecodeAccountFast(wire)
	if err != nil {
		t.Fatal(err)
	}
	assertAccountEqual(t, &trad, &fast)
}

func TestDecodeAccountFast_ParityWithTraditional_NoOptionals(t *testing.T) {
	mint := newTestKey(0x70)
	owner := newTestKey(0x80)
	orig := &Account{
		Mint:   mint,
		Owner:  owner,
		Amount: 1,
		State:  AccountStateInitialized,
	}
	wire := buildAccountWire(t, orig)

	var trad Account
	if err := bin.NewBinDecoder(wire).Decode(&trad); err != nil {
		t.Fatal(err)
	}
	fast, err := DecodeAccountFast(wire)
	if err != nil {
		t.Fatal(err)
	}
	assertAccountEqual(t, &trad, &fast)
}

// ---- TLV-extended buffer: base decode must still work ----

// Token-2022 accounts often carry trailing TLV extensions. The POD
// path must ignore those bytes — bin.UnmarshalPOD copies only sizeof(T),
// bin.ViewAs aliases a prefix.
func TestDecodeAccountFast_IgnoresTrailingTLV(t *testing.T) {
	mint := newTestKey(0x91)
	owner := newTestKey(0x92)
	orig := &Account{Mint: mint, Owner: owner, Amount: 7, State: AccountStateInitialized}
	base := buildAccountWire(t, orig)

	// Append arbitrary TLV-shaped bytes.
	extended := append([]byte(nil), base...)
	extended = append(extended, 0x01, 0x02, 0x03, 0x04, 0x05)

	fast, err := DecodeAccountFast(extended)
	if err != nil {
		t.Fatal(err)
	}
	if fast.Mint != mint || fast.Owner != owner || fast.Amount != 7 {
		t.Errorf("base fields mis-decoded from TLV-extended buffer")
	}

	view, err := ViewAccountRaw(extended)
	if err != nil {
		t.Fatal(err)
	}
	if view.AmountU64() != 7 {
		t.Errorf("view on TLV-extended buffer misread amount")
	}
}

// ---- View semantics: aliasing + mutation-through ----

func TestViewMintRaw_AliasesBuffer(t *testing.T) {
	mintAuth := newTestKey(0x11)
	orig := &Mint{MintAuthority: &mintAuth, Supply: 100, Decimals: 2, IsInitialized: true}
	wire := buildMintWire(t, orig)

	view, err := ViewMintRaw(wire)
	if err != nil {
		t.Fatal(err)
	}

	if view.SupplyU64() != 100 {
		t.Errorf("SupplyU64 = %d, want 100", view.SupplyU64())
	}
	wire[36] = 0xff
	if view.SupplyU64()&0xff != 0xff {
		t.Errorf("view did not observe mutation through buffer")
	}
}

func TestViewAccountRaw_AccessorsMatchFields(t *testing.T) {
	mint := newTestKey(0x12)
	owner := newTestKey(0x22)
	orig := &Account{
		Mint:   mint,
		Owner:  owner,
		Amount: 777,
		State:  AccountStateInitialized,
	}
	wire := buildAccountWire(t, orig)

	view, err := ViewAccountRaw(wire)
	if err != nil {
		t.Fatal(err)
	}
	if view.MintKey() != mint {
		t.Errorf("MintKey mismatch")
	}
	if view.OwnerKey() != owner {
		t.Errorf("OwnerKey mismatch")
	}
	if view.AmountU64() != 777 {
		t.Errorf("AmountU64 = %d, want 777", view.AmountU64())
	}
	if view.AccountState() != AccountStateInitialized {
		t.Errorf("AccountState = %v, want AccountStateInitialized", view.AccountState())
	}
	if view.HasDelegate() {
		t.Error("unexpected delegate")
	}
	if _, native := view.IsNativeU64(); native {
		t.Error("unexpected native flag")
	}
}

// ---- Short buffer error ----

func TestViewMintRaw_ShortBuffer(t *testing.T) {
	_, err := ViewMintRaw(make([]byte, MintAccountSize-1))
	if err == nil {
		t.Fatal("expected error on short mint buffer")
	}
}

func TestDecodeAccountRaw_ShortBuffer(t *testing.T) {
	var a AccountRaw
	if err := DecodeAccountRaw(&a, make([]byte, TokenAccountSize-1)); err == nil {
		t.Fatal("expected error on short account buffer")
	}
}

// ---- Round-trip through MarshalPOD ----

func TestMarshalPOD_RoundTripAccount(t *testing.T) {
	src := &Account{
		Mint:   newTestKey(1),
		Owner:  newTestKey(2),
		Amount: 12345,
		State:  AccountStateInitialized,
	}
	wire := buildAccountWire(t, src)

	var raw AccountRaw
	if err := DecodeAccountRaw(&raw, wire); err != nil {
		t.Fatal(err)
	}

	var reencoded [TokenAccountSize]byte
	n, err := bin.MarshalPOD(&raw, reencoded[:])
	if err != nil {
		t.Fatal(err)
	}
	if n != TokenAccountSize {
		t.Errorf("MarshalPOD wrote %d bytes, want %d", n, TokenAccountSize)
	}
	if string(reencoded[:]) != string(wire) {
		t.Errorf("round-trip mismatch:\n got:  %x\n want: %x", reencoded[:], wire)
	}
}

// ---- Helpers ----

func assertMintEqual(t *testing.T, a, b *Mint) {
	t.Helper()
	if a.Supply != b.Supply {
		t.Errorf("Supply: %d vs %d", a.Supply, b.Supply)
	}
	if a.Decimals != b.Decimals {
		t.Errorf("Decimals: %d vs %d", a.Decimals, b.Decimals)
	}
	if a.IsInitialized != b.IsInitialized {
		t.Errorf("IsInitialized: %v vs %v", a.IsInitialized, b.IsInitialized)
	}
	if (a.MintAuthority == nil) != (b.MintAuthority == nil) {
		t.Errorf("MintAuthority presence mismatch: %v vs %v", a.MintAuthority, b.MintAuthority)
	} else if a.MintAuthority != nil && *a.MintAuthority != *b.MintAuthority {
		t.Errorf("MintAuthority: %v vs %v", *a.MintAuthority, *b.MintAuthority)
	}
	if (a.FreezeAuthority == nil) != (b.FreezeAuthority == nil) {
		t.Errorf("FreezeAuthority presence mismatch")
	} else if a.FreezeAuthority != nil && *a.FreezeAuthority != *b.FreezeAuthority {
		t.Errorf("FreezeAuthority: %v vs %v", *a.FreezeAuthority, *b.FreezeAuthority)
	}
}

func assertAccountEqual(t *testing.T, a, b *Account) {
	t.Helper()
	if a.Mint != b.Mint {
		t.Errorf("Mint: %v vs %v", a.Mint, b.Mint)
	}
	if a.Owner != b.Owner {
		t.Errorf("Owner: %v vs %v", a.Owner, b.Owner)
	}
	if a.Amount != b.Amount {
		t.Errorf("Amount: %d vs %d", a.Amount, b.Amount)
	}
	if a.State != b.State {
		t.Errorf("State: %v vs %v", a.State, b.State)
	}
	if a.DelegatedAmount != b.DelegatedAmount {
		t.Errorf("DelegatedAmount: %d vs %d", a.DelegatedAmount, b.DelegatedAmount)
	}
	if (a.Delegate == nil) != (b.Delegate == nil) {
		t.Errorf("Delegate presence mismatch")
	} else if a.Delegate != nil && *a.Delegate != *b.Delegate {
		t.Errorf("Delegate: %v vs %v", *a.Delegate, *b.Delegate)
	}
	if (a.IsNative == nil) != (b.IsNative == nil) {
		t.Errorf("IsNative presence mismatch")
	} else if a.IsNative != nil && *a.IsNative != *b.IsNative {
		t.Errorf("IsNative: %v vs %v", *a.IsNative, *b.IsNative)
	}
	if (a.CloseAuthority == nil) != (b.CloseAuthority == nil) {
		t.Errorf("CloseAuthority presence mismatch")
	} else if a.CloseAuthority != nil && *a.CloseAuthority != *b.CloseAuthority {
		t.Errorf("CloseAuthority: %v vs %v", *a.CloseAuthority, *b.CloseAuthority)
	}
}
