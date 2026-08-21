//go:build amd64 && !purego

package base58

import (
	"encoding/hex"
	"math/rand"
	"testing"
)

func requireAVX2(t *testing.T) {
	if !useAVX2 {
		t.Skip("AVX2 not available")
	}
}

func checkEncode32(t *testing.T, src *[32]byte) {
	t.Helper()
	want, wantApp := encode32Generic(src), string(appendEncode32Generic([]byte("x"), src))
	if got := Encode32(src); got != want {
		t.Fatalf("Encode32(%x) = %q, want %q", src[:], got, want)
	}
	if got := string(AppendEncode32([]byte("x"), src)); got != wantApp {
		t.Fatalf("AppendEncode32(%x) = %q, want %q", src[:], got, wantApp)
	}
	var back [32]byte
	if err := Decode32(want, &back); err != nil || back != *src {
		t.Fatalf("Decode32(%q) = %x, %v; want %x", want, back[:], err, src[:])
	}
}

func checkEncode64(t *testing.T, src *[64]byte) {
	t.Helper()
	want, wantApp := encode64Generic(src), string(appendEncode64Generic([]byte("x"), src))
	if got := Encode64(src); got != want {
		t.Fatalf("Encode64(%x) = %q, want %q", src[:], got, want)
	}
	if got := string(AppendEncode64([]byte("x"), src)); got != wantApp {
		t.Fatalf("AppendEncode64(%x) = %q, want %q", src[:], got, wantApp)
	}
	var back [64]byte
	if err := Decode64(want, &back); err != nil || back != *src {
		t.Fatalf("Decode64(%q) = %x, %v; want %x", want, back[:], err, src[:])
	}
}

// Inputs that take the ripple slow path (about one random input in 1e8).
var rippleInputs32 = []string{
	"6df6d5331ee396d5c9fefd53ec237a5375defb6fc0bc0e05625812a69bd3ce4e",
	"f61687c75556f5c368d8a755e40d99aafe2073418c32a2b5004fa141fad3d4c4",
}

var rippleInputs64 = []string{
	"f69e3013e2ea5dee375824057fbc5c07e7859a4521d47bd1f200e89b0ef6c893db202e8127300dd220ee8e752edf0d19042d30f945221125db6e31a5e542fcea",
	"4fa18e9148ba81e73dbb56a9664595f78f02d456e387b78a5e6d828027e96f76c81b2022c9a2411b664a0aa510510da4cdc5bb2f3ce73a8d058fbb8c0c359762",
	"3daf83075065e87b6278aef6a3364095701a2a00b14885b7f2cc819cfe381cc6741b1f7cf74487293b3d707831c150578847965f0b69a805f0908f0d51745487",
}

func TestAVX2_EncodeEdgeCases(t *testing.T) {
	requireAVX2(t)
	rng := rand.New(rand.NewSource(7))
	var s32 [32]byte
	var s64 [64]byte
	for z := 0; z <= 32; z++ {
		for _, fill := range []byte{0xFF, 0x01, 0x80} {
			for i := range s32 {
				s32[i] = fill
			}
			rng.Read(s32[z:])
			if fill != 0x01 {
				for i := z; i < 32; i++ {
					s32[i] = fill
				}
			}
			for i := 0; i < z; i++ {
				s32[i] = 0
			}
			checkEncode32(t, &s32)
		}
	}
	for z := 0; z <= 64; z++ {
		for _, fill := range []byte{0xFF, 0x01, 0x80} {
			rng.Read(s64[:])
			if fill != 0x01 {
				for i := z; i < 64; i++ {
					s64[i] = fill
				}
			}
			for i := 0; i < z; i++ {
				s64[i] = 0
			}
			checkEncode64(t, &s64)
		}
	}
	for _, h := range rippleInputs32 {
		b, _ := hex.DecodeString(h)
		copy(s32[:], b)
		checkEncode32(t, &s32)
	}
	for _, h := range rippleInputs64 {
		b, _ := hex.DecodeString(h)
		copy(s64[:], b)
		checkEncode64(t, &s64)
	}
	for i := 0; i < 1000; i++ {
		rng.Read(s32[:])
		checkEncode32(t, &s32)
		rng.Read(s64[:])
		checkEncode64(t, &s64)
	}
}

func TestAVX2_DecodeMatchesScalar(t *testing.T) {
	requireAVX2(t)
	rng := rand.New(rand.NewSource(11))
	check32 := func(s string) {
		var got, want [32]byte
		wantErr := decode32Generic(s, &want)
		gotErr := Decode32(s, &got)
		if gotErr != wantErr || (wantErr == nil && got != want) {
			t.Fatalf("Decode32(%q) = %x, %v; want %x, %v", s, got[:], gotErr, want[:], wantErr)
		}
	}
	check64 := func(s string) {
		var got, want [64]byte
		wantErr := decode64Generic(s, &want)
		gotErr := Decode64(s, &got)
		if gotErr != wantErr || (wantErr == nil && got != want) {
			t.Fatalf("Decode64(%q) = %x, %v; want %x, %v", s, got[:], gotErr, want[:], wantErr)
		}
	}
	for n := 1; n <= 45; n++ {
		for ones := 0; ones <= n && ones <= 33; ones++ {
			b := make([]byte, n)
			for i := range b {
				b[i] = base58Chars[rng.Intn(58)]
			}
			for i := 0; i < ones; i++ {
				b[i] = '1'
			}
			check32(string(b))
			for i := range b {
				b[i] = 'z'
			}
			check32(string(b))
		}
	}
	for n := 1; n <= 90; n++ {
		for ones := 0; ones <= n && ones <= 65; ones += 3 {
			b := make([]byte, n)
			for i := range b {
				b[i] = base58Chars[rng.Intn(58)]
			}
			for i := 0; i < ones; i++ {
				b[i] = '1'
			}
			check64(string(b))
			for i := range b {
				b[i] = 'z'
			}
			check64(string(b))
		}
	}
	var s32 [32]byte
	var s64 [64]byte
	rng.Read(s32[:])
	rng.Read(s64[:])
	base32 := []byte(Encode32(&s32))
	base64 := []byte(Encode64(&s64))
	for pos := 0; pos < len(base32); pos++ {
		for c := 0; c < 256; c++ {
			b := append([]byte(nil), base32...)
			b[pos] = byte(c)
			check32(string(b))
		}
	}
	for pos := 0; pos < len(base64); pos += 5 {
		for c := 0; c < 256; c++ {
			b := append([]byte(nil), base64...)
			b[pos] = byte(c)
			check64(string(b))
		}
	}
	for i := 0; i < 1000; i++ {
		rng.Read(s32[:])
		z := rng.Intn(33)
		for j := 0; j < z; j++ {
			s32[j] = 0
		}
		check32(Encode32(&s32))
		rng.Read(s64[:])
		z = rng.Intn(65)
		for j := 0; j < z; j++ {
			s64[j] = 0
		}
		check64(Encode64(&s64))
	}
}
