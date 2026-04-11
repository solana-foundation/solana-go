package tokenregistry

import (
	"bytes"
	"testing"

	bin "github.com/gagliardetto/solana-go/binary"

	"github.com/gagliardetto/solana-go"
)

// makeBenchTokenMeta builds a fully-populated TokenMeta. It is the largest
// reflect-marshaled struct in the repo (9 fields, two foreign-package pointer
// types, four nested fixed-size byte arrays), so it stresses the typePlan
// cache, the indirect() walker, and the array-write fast path simultaneously.
func makeBenchTokenMeta() TokenMeta {
	mint := solana.PublicKey{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	auth := solana.PublicKey{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	logo, _ := LogoFromString("https://example.com/token-logo.png")
	name, _ := NameFromString("Example Token")
	site, _ := WebsiteFromString("https://example.com")
	sym, _ := SymbolFromString("EXMPL")
	return TokenMeta{
		IsInitialized:         true,
		Reg:                   [3]byte{1, 2, 3},
		DataType:              7,
		MintAddress:           &mint,
		RegistrationAuthority: &auth,
		Logo:                  logo,
		Name:                  name,
		Website:               site,
		Symbol:                sym,
	}
}

// BenchmarkEncode_TokenMeta exercises the reflect/typePlan path on the
// largest struct in the repo. Uses MarshalBin (the high-level helper) so
// the benchmark is portable across the upstream gagliardetto/binary
// v0.8.0 module and the vendored copy.
func BenchmarkEncode_TokenMeta(b *testing.B) {
	tm := makeBenchTokenMeta()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err := bin.MarshalBin(&tm)
		if err != nil {
			b.Fatal(err)
		}
		_ = buf
	}
}

func BenchmarkDecode_TokenMeta(b *testing.B) {
	tm := makeBenchTokenMeta()
	data, err := bin.MarshalBin(&tm)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out TokenMeta
		dec := bin.NewBinDecoder(data)
		if err := dec.Decode(&out); err != nil {
			b.Fatal(err)
		}
	}
}

// Reused-buffer encoder variant — writes into a single bytes.Buffer that's
// reset between iterations, isolating the reflect-walk cost from
// MarshalBin's allocation of a fresh buffer per call.
func BenchmarkEncode_TokenMeta_Reused(b *testing.B) {
	tm := makeBenchTokenMeta()
	var buf bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc := bin.NewBinEncoder(&buf)
		if err := enc.Encode(&tm); err != nil {
			b.Fatal(err)
		}
	}
}
