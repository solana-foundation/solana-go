package tokenregistry

import (
	"bytes"
	"testing"
	"time"

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

// BenchmarkDecode_TokenMeta_UnmarshalBin exercises the pooled
// UnmarshalBin convenience helper, which draws a Decoder from a
// sync.Pool, resets it with the input bytes, decodes, and returns it
// for reuse. Compare to BenchmarkDecode_TokenMeta which constructs a
// fresh Decoder every iteration.
func BenchmarkDecode_TokenMeta_UnmarshalBin(b *testing.B) {
	tm := makeBenchTokenMeta()
	data, err := bin.MarshalBin(&tm)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out TokenMeta
		if err := bin.UnmarshalBin(&out, data); err != nil {
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

// coldStruct is a copy of TokenMeta intentionally never referenced
// outside this file. Its typePlan has not been prewarmed by init(), so
// the very first call to MarshalBin pays the cold-path cost (reflect
// walk + plan construction + sync.Map.LoadOrStore).
type coldStruct struct {
	IsInitialized         bool
	Reg                   [3]byte
	DataType              byte
	MintAddress           *solana.PublicKey
	RegistrationAuthority *solana.PublicKey
	Logo                  Logo
	Name                  Name
	Website               Website
	Symbol                Symbol
}

// runtimeWarmup is a throwaway type used to pre-load the encoder's
// reflect-driven code paths, CPU caches, and any lazy-init globals
// BEFORE we measure typePlan-build cost. Without this, the very first
// MarshalBin call in a process pays a fixed ~100-200µs runtime cold-
// start cost that dwarfs the actual typePlan construction (~5-10µs)
// and gives nonsense readings.
type runtimeWarmup struct{ X uint64 }

// TestFirstCallCost measures, after a runtime warmup, the cost of
// encoding TokenMeta (whose typePlan has been prewarmed by init via
// PrewarmTypes) versus coldStruct (whose typePlan has NEVER been
// touched, so the very first encode pays the reflect-walk cost). The
// delta is the cold-path overhead that PrewarmTypes eliminates.
func TestFirstCallCost(t *testing.T) {
	// Throwaway call to warm the runtime, the encoder, and the CPU
	// caches. After this, both subsequent measurements start from the
	// same hot state and only the typePlan lookup differs.
	if _, err := bin.MarshalBin(&runtimeWarmup{X: 42}); err != nil {
		t.Fatal(err)
	}

	// Warm path: TokenMeta. Prewarmed in init() — the typePlan is
	// already in the sync.Map and the first encode hits the load fast
	// path.
	tm := makeBenchTokenMeta()
	startWarm := time.Now()
	if _, err := bin.MarshalBin(&tm); err != nil {
		t.Fatal(err)
	}
	warmDur := time.Since(startWarm)

	// Cold path: coldStruct. Cache miss, typePlan built inline.
	cs := coldStruct{IsInitialized: true, Reg: [3]byte{1, 2, 3}, DataType: 7}
	mint := solana.PublicKey{1, 2, 3}
	auth := solana.PublicKey{4, 5, 6}
	cs.MintAddress = &mint
	cs.RegistrationAuthority = &auth
	logo, _ := LogoFromString("logo")
	name, _ := NameFromString("name")
	site, _ := WebsiteFromString("site")
	sym, _ := SymbolFromString("sym")
	cs.Logo = logo
	cs.Name = name
	cs.Website = site
	cs.Symbol = sym

	startCold := time.Now()
	if _, err := bin.MarshalBin(&cs); err != nil {
		t.Fatal(err)
	}
	coldDur := time.Since(startCold)

	t.Logf("warm (TokenMeta, typePlan prewarmed via init): %v", warmDur)
	t.Logf("cold (coldStruct, typePlan miss + build):      %v", coldDur)
	t.Logf("cold-path overhead saved by prewarm:           %v", coldDur-warmDur)
}
