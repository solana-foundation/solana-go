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
	"testing"
	"unsafe"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
)

func TestNonceAccountLayoutSize(t *testing.T) {
	if got := unsafe.Sizeof(NonceAccount{}); got != NonceAccountSize {
		t.Errorf("sizeof(NonceAccount) = %d, want %d", got, NonceAccountSize)
	}
}

func TestNonceAccountAssertPOD(t *testing.T) {
	if err := bin.AssertPOD[NonceAccount](); err != nil {
		t.Error(err)
	}
}

func newNonceTestKey(b byte) solana.PublicKey {
	var k solana.PublicKey
	for i := range k {
		k[i] = b + byte(i)
	}
	return k
}

func buildNonceWire(t *testing.T, n *NonceAccount) []byte {
	t.Helper()
	var buf [NonceAccountSize]byte
	enc := bin.NewBinEncoderInto(buf[:])
	if err := n.MarshalWithEncoder(enc); err != nil {
		t.Fatalf("encode nonce: %v", err)
	}
	if enc.Written() != NonceAccountSize {
		t.Fatalf("encoded nonce is %d bytes, want %d", enc.Written(), NonceAccountSize)
	}
	return buf[:]
}

func TestDecodeNonceAccountFast_ParityWithTraditional(t *testing.T) {
	orig := &NonceAccount{
		Version:          1,
		State:            1,
		AuthorizedPubkey: newNonceTestKey(0x10),
		Nonce:            newNonceTestKey(0x20),
		FeeCalculator:    FeeCalculator{LamportsPerSignature: 5000},
	}
	wire := buildNonceWire(t, orig)

	var trad NonceAccount
	if err := bin.NewBinDecoder(wire).Decode(&trad); err != nil {
		t.Fatal(err)
	}
	fast, err := DecodeNonceAccountFast(wire)
	if err != nil {
		t.Fatal(err)
	}
	if trad != fast {
		t.Errorf("fast decode mismatch:\n trad: %+v\n fast: %+v", trad, fast)
	}
	if fast != *orig {
		t.Errorf("fast decode lost data:\n got:  %+v\n want: %+v", fast, *orig)
	}
}

func TestViewNonceAccount_AliasesBuffer(t *testing.T) {
	orig := &NonceAccount{
		Version:          1,
		State:            1,
		AuthorizedPubkey: newNonceTestKey(0x30),
		Nonce:            newNonceTestKey(0x40),
		FeeCalculator:    FeeCalculator{LamportsPerSignature: 1234},
	}
	wire := buildNonceWire(t, orig)

	view, err := ViewNonceAccount(wire)
	if err != nil {
		t.Fatal(err)
	}
	if view.FeeCalculator.LamportsPerSignature != 1234 {
		t.Errorf("LamportsPerSignature = %d, want 1234", view.FeeCalculator.LamportsPerSignature)
	}

	// Mutating the buffer is visible through the view.
	wire[72] = 0xff // low byte of LamportsPerSignature
	if view.FeeCalculator.LamportsPerSignature&0xff != 0xff {
		t.Errorf("view did not observe mutation through buffer")
	}
}

func TestDecodeNonceAccountFast_ShortBuffer(t *testing.T) {
	if _, err := DecodeNonceAccountFast(make([]byte, NonceAccountSize-1)); err == nil {
		t.Fatal("expected error on short nonce buffer")
	}
}

func TestViewNonceAccount_ShortBuffer(t *testing.T) {
	if _, err := ViewNonceAccount(make([]byte, NonceAccountSize-1)); err == nil {
		t.Fatal("expected error on short nonce buffer")
	}
}

func TestMarshalPOD_RoundTripNonceAccount(t *testing.T) {
	src := &NonceAccount{
		Version:          42,
		State:            1,
		AuthorizedPubkey: newNonceTestKey(0x50),
		Nonce:            newNonceTestKey(0x60),
		FeeCalculator:    FeeCalculator{LamportsPerSignature: 7777},
	}
	wire := buildNonceWire(t, src)

	var got NonceAccount
	if err := DecodeNonceAccountRaw(&got, wire); err != nil {
		t.Fatal(err)
	}

	var reencoded [NonceAccountSize]byte
	n, err := bin.MarshalPOD(&got, reencoded[:])
	if err != nil {
		t.Fatal(err)
	}
	if n != NonceAccountSize {
		t.Errorf("MarshalPOD wrote %d bytes, want %d", n, NonceAccountSize)
	}
	if string(reencoded[:]) != string(wire) {
		t.Errorf("round-trip mismatch:\n got:  %x\n want: %x", reencoded[:], wire)
	}
}
