package token2022

import (
	"bytes"
	"encoding/base64"
	"testing"
)

// The fuzz targets check two invariants of the strict decoders on arbitrary
// input: they never panic, and any data that decodes successfully re-encodes
// deterministically (decode -> encode -> decode -> encode is a fixed point;
// the first encode may legally differ from the input by dropping trailing
// TLV slack).

func fuzzSeeds(f *testing.F) {
	f.Helper()
	mainnet, err := base64.StdEncoding.DecodeString(mainnetXStockMintB64)
	if err != nil {
		f.Fatal(err)
	}
	f.Add([]byte{})
	f.Add(testMintSlice)
	f.Add(testAccountSlice)
	f.Add(mintWithAccountType())
	f.Add(mintWithExtension())
	f.Add(accountWithExtension())
	f.Add(mainnet)
}

func FuzzDecodeMintWithExtensions(f *testing.F) {
	fuzzSeeds(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		m, err := DecodeMintWithExtensions(data)
		if err != nil {
			return
		}
		out, err := m.MarshalBinary()
		if err != nil {
			t.Fatalf("decoded mint failed to re-encode: %v", err)
		}
		m2, err := DecodeMintWithExtensions(out)
		if err != nil {
			t.Fatalf("re-encoded mint failed to decode: %v", err)
		}
		out2, err := m2.MarshalBinary()
		if err != nil {
			t.Fatalf("second re-encode failed: %v", err)
		}
		if !bytes.Equal(out, out2) {
			t.Fatalf("encode is not a fixed point:\nfirst:  %x\nsecond: %x", out, out2)
		}
	})
}

func FuzzDecodeAccountWithExtensions(f *testing.F) {
	fuzzSeeds(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		a, err := DecodeAccountWithExtensions(data)
		if err != nil {
			return
		}
		out, err := a.MarshalBinary()
		if err != nil {
			t.Fatalf("decoded account failed to re-encode: %v", err)
		}
		a2, err := DecodeAccountWithExtensions(out)
		if err != nil {
			t.Fatalf("re-encoded account failed to decode: %v", err)
		}
		out2, err := a2.MarshalBinary()
		if err != nil {
			t.Fatalf("second re-encode failed: %v", err)
		}
		if !bytes.Equal(out, out2) {
			t.Fatalf("encode is not a fixed point:\nfirst:  %x\nsecond: %x", out, out2)
		}
	})
}
