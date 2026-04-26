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

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
)

func benchNonceKey(x byte) solana.PublicKey {
	var k solana.PublicKey
	for i := range k {
		k[i] = x + byte(i)
	}
	return k
}

func benchNonceWire(b *testing.B) []byte {
	b.Helper()
	n := &NonceAccount{
		Version:          1,
		State:            1,
		AuthorizedPubkey: benchNonceKey(0x11),
		Nonce:            benchNonceKey(0x22),
		FeeCalculator:    FeeCalculator{LamportsPerSignature: 5000},
	}
	buf := make([]byte, NonceAccountSize)
	enc := bin.NewBinEncoderInto(buf)
	if err := n.MarshalWithEncoder(enc); err != nil {
		b.Fatal(err)
	}
	return buf
}

// Traditional: bin.NewBinDecoder + UnmarshalWithDecoder, per-field
// ReadUint32 / ReadNBytes calls.
func BenchmarkDecodeNonceAccount_Traditional(b *testing.B) {
	wire := benchNonceWire(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var n NonceAccount
		if err := bin.NewBinDecoder(wire).Decode(&n); err != nil {
			b.Fatal(err)
		}
	}
}

// Fast: bin.UnmarshalPOD — single 80-byte memcpy.
func BenchmarkDecodeNonceAccount_Fast(b *testing.B) {
	wire := benchNonceWire(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DecodeNonceAccountFast(wire); err != nil {
			b.Fatal(err)
		}
	}
}

// View: zero-copy pointer alias; touch one field so the compiler does
// not dead-code eliminate the call.
func BenchmarkDecodeNonceAccount_View(b *testing.B) {
	wire := benchNonceWire(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v, err := ViewNonceAccount(wire)
		if err != nil {
			b.Fatal(err)
		}
		_ = v.FeeCalculator.LamportsPerSignature
	}
}
