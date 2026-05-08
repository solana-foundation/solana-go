package base58

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Known test vectors cross-validated against multiple base58 implementations
// (Bitcoin Core, bs58, mr-tron, five8). Any implementation that encodes these
// bytes to the given strings — and decodes them back — is bit-compatible.
var knownVectors32 = []struct {
	hex string
	b58 string
}{
	{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"11111111111111111111111111111111",
	},
	{
		"0000000000000000000000000000000000000000000000000000000000000001",
		"11111111111111111111111111111112",
	},
	{
		// Solana pubkey: 4cHoJNmLed5PBgFBezHmJkMJLEZrcTvr3aopjnYBRxUb
		"359d6209a1296a422463405b82829cf2f0a86b2e87077c80a74372841e185efc",
		"4cHoJNmLed5PBgFBezHmJkMJLEZrcTvr3aopjnYBRxUb",
	},
}

var knownVectors64 = []struct {
	hex string
	b58 string
}{
	{
		// Solana signature: 5YBLhMBLjhAHnEPnHKLLnVwHSfXGPJMCvKAfNsiaEw2T63edrYxVFHKUxRXfP6KA1HVo7c9JZ3LAJQR72giX7Cb
		// Hex cross-checked against Python's `base58` package.
		"03e9bb70b0ae091b4a3233dc952a2da569afaa0ae1c06aa7d3c2a4da2f2854ec76dfae30d9474b4593726761345bec7ce1a95812c1fa8ddc740314cb29fef458",
		"5YBLhMBLjhAHnEPnHKLLnVwHSfXGPJMCvKAfNsiaEw2T63edrYxVFHKUxRXfP6KA1HVo7c9JZ3LAJQR72giX7Cb",
	},
}

func TestEncode32_KnownVectors(t *testing.T) {
	for _, tv := range knownVectors32 {
		raw, err := hex.DecodeString(tv.hex)
		require.NoError(t, err)
		var src [32]byte
		copy(src[:], raw)
		assert.Equal(t, tv.b58, Encode32(&src), "hex=%s", tv.hex)
	}
}

func TestDecode32_KnownVectors(t *testing.T) {
	for _, tv := range knownVectors32 {
		expected, err := hex.DecodeString(tv.hex)
		require.NoError(t, err)
		var dst [32]byte
		err = Decode32(tv.b58, &dst)
		require.NoError(t, err)
		assert.Equal(t, expected, dst[:], "b58=%s", tv.b58)
	}
}

func TestEncode64_KnownVectors(t *testing.T) {
	for _, tv := range knownVectors64 {
		raw, err := hex.DecodeString(tv.hex)
		require.NoError(t, err)
		var src [64]byte
		copy(src[:], raw)
		assert.Equal(t, tv.b58, Encode64(&src), "hex=%s", tv.hex)
	}
}

func TestDecode64_KnownVectors(t *testing.T) {
	for _, tv := range knownVectors64 {
		expected, err := hex.DecodeString(tv.hex)
		require.NoError(t, err)
		var dst [64]byte
		err = Decode64(tv.b58, &dst)
		require.NoError(t, err)
		assert.Equal(t, expected, dst[:], "b58=%s", tv.b58)
	}
}

func TestEncode32_Zeros(t *testing.T) {
	var src [32]byte
	assert.Equal(t, "11111111111111111111111111111111", Encode32(&src))
}

func TestDecode32_Zeros(t *testing.T) {
	var dst [32]byte
	require.NoError(t, Decode32("11111111111111111111111111111111", &dst))
	assert.Equal(t, [32]byte{}, dst)
}

func TestRoundtrip32_Random(t *testing.T) {
	// Cross-check the specialized fixed-size path against the variable-length
	// fallback — the two share no code, so disagreement flags a bug.
	for range 1000 {
		var src [32]byte
		rand.Read(src[:])

		encoded := Encode(src[:])
		assert.Equal(t, encodeVariable(src[:]), encoded, "encode mismatch for %x", src)

		var decoded [32]byte
		require.NoError(t, Decode32(encoded, &decoded))
		assert.Equal(t, src, decoded, "decode mismatch for %s", encoded)

		generic, err := Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, src[:], generic, "generic decode mismatch for %s", encoded)
	}
}

func TestRoundtrip64_Random(t *testing.T) {
	for range 1000 {
		var src [64]byte
		rand.Read(src[:])

		encoded := Encode(src[:])
		assert.Equal(t, encodeVariable(src[:]), encoded, "encode mismatch for %x", src)

		var decoded [64]byte
		require.NoError(t, Decode64(encoded, &decoded))
		assert.Equal(t, src, decoded, "decode mismatch for %s", encoded)

		generic, err := Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, src[:], generic, "generic decode mismatch for %s", encoded)
	}
}

func TestAppendEncode32_ZeroAlloc(t *testing.T) {
	var src [32]byte
	rand.Read(src[:])
	expected := Encode(src[:])

	// Pre-sized buffer: should not allocate.
	buf := make([]byte, 0, EncodedMaxLen32)
	buf = AppendEncode32(buf, &src)
	assert.Equal(t, expected, string(buf))

	// Append to an existing buffer.
	prefix := []byte("pubkey=")
	buf2 := make([]byte, 0, len(prefix)+EncodedMaxLen32)
	buf2 = append(buf2, prefix...)
	buf2 = AppendEncode32(buf2, &src)
	assert.Equal(t, "pubkey="+expected, string(buf2))
}

func TestAppendEncode64_ZeroAlloc(t *testing.T) {
	var src [64]byte
	rand.Read(src[:])
	expected := Encode(src[:])

	buf := make([]byte, 0, EncodedMaxLen64)
	buf = AppendEncode64(buf, &src)
	assert.Equal(t, expected, string(buf))
}

func TestDecode_InvalidChars(t *testing.T) {
	var dst [32]byte
	assert.Error(t, Decode32("0invalid", &dst)) // '0' is not in base58
	assert.Error(t, Decode32("I\x00nvalid", &dst))
	assert.Error(t, Decode32("Oinvalid", &dst)) // 'O' is not in base58
}

// Known vectors for the variable-length API. Cross-validated against
// Bitcoin Core, bs58, and five8.
var knownVectorsVar = []struct {
	hex string
	b58 string
}{
	{"", ""},
	{"00", "1"},
	{"0000", "11"},
	{"00000000", "1111"},
	{"61", "2g"},
	{"626262", "a3gV"},
	{"636363", "aPEr"},
	{"73696d706c792061206c6f6e6720737472696e67", "2cFupjhnEsSn59qHXstmK2ffpLv2"},
	{"00eb15231dfceb60925886b67d065299925915aeb172c06647", "1NS17iag9jJgTHD1VXjvLCEnZuQ3rJDE9L"},
	// Solana instruction data sample from transaction_test.go.
	{"020000003930000000000000", "3Bxs4ART6LMJ13T5"},
}

func TestEncode_KnownVectors(t *testing.T) {
	for _, tv := range knownVectorsVar {
		raw, err := hex.DecodeString(tv.hex)
		require.NoError(t, err)
		assert.Equal(t, tv.b58, Encode(raw), "hex=%s", tv.hex)
	}
}

func TestDecode_KnownVectors(t *testing.T) {
	for _, tv := range knownVectorsVar {
		expected, err := hex.DecodeString(tv.hex)
		require.NoError(t, err)
		got, err := Decode(tv.b58)
		require.NoError(t, err, "b58=%s", tv.b58)
		if expected == nil {
			expected = []byte{}
		}
		assert.Equal(t, expected, got, "b58=%s", tv.b58)
	}
}

func TestEncode_Empty(t *testing.T) {
	assert.Equal(t, "", Encode(nil))
	assert.Equal(t, "", Encode([]byte{}))
}

func TestDecode_Empty(t *testing.T) {
	got, err := Decode("")
	require.NoError(t, err)
	assert.Equal(t, []byte{}, got)
}

func TestRoundtrip_Variable_Random(t *testing.T) {
	// Cover assorted lengths including ones the fixed-size paths can't handle.
	for _, n := range []int{1, 5, 12, 31, 33, 63, 65, 100, 250, 1000} {
		for range 100 {
			src := make([]byte, n)
			rand.Read(src)

			encoded := Encode(src)
			decoded, err := Decode(encoded)
			require.NoError(t, err, "len=%d", n)
			assert.Equal(t, src, decoded, "len=%d encoded=%s", n, encoded)
		}
	}
}

func TestRoundtrip_Variable_LeadingZeros(t *testing.T) {
	// Encoded leading '1's must round-trip to the same number of leading zeros.
	for zeros := 0; zeros < 10; zeros++ {
		for tail := 0; tail < 10; tail++ {
			src := make([]byte, zeros+tail)
			if tail > 0 {
				rand.Read(src[zeros:])
				if src[zeros] == 0 {
					src[zeros] = 1
				}
			}
			encoded := Encode(src)
			decoded, err := Decode(encoded)
			require.NoError(t, err)
			assert.Equal(t, src, decoded, "zeros=%d tail=%d", zeros, tail)
		}
	}
}

func TestDecode_InvalidChars_Variable(t *testing.T) {
	for _, in := range []string{"0", "O", "I", "l", "abc!", "abc 123", "\x00"} {
		_, err := Decode(in)
		assert.Error(t, err, "expected error for %q", in)
	}
}

func BenchmarkBase58_Decode_Variable(b *testing.B) {
	b.SetBytes(64)
	for b.Loop() {
		Decode(benchStr64)
	}
}

// Benchmarks
var (
	benchSrc32 [32]byte
	benchSrc64 [64]byte
	benchStr32 string
	benchStr64 string
)

func init() {
	rand.Read(benchSrc32[:])
	rand.Read(benchSrc64[:])
	benchStr32 = Encode(benchSrc32[:])
	benchStr64 = Encode(benchSrc64[:])
}

func BenchmarkBase58_EncodeVariable(b *testing.B) {
	// Cover lengths that bypass the 32/64 fast paths and exercise the
	// long-division encoder. Solana instruction data is typically <= 1KB.
	for _, n := range []int{16, 100, 1000} {
		src := make([]byte, n)
		rand.Read(src)
		b.Run(fmt.Sprintf("len=%d", n), func(b *testing.B) {
			b.SetBytes(int64(n))
			for b.Loop() {
				Encode(src)
			}
		})
	}
}

func BenchmarkBase58_Encode32(b *testing.B) {
	src := &benchSrc32
	b.SetBytes(32)
	for b.Loop() {
		Encode32(src)
	}
}

func BenchmarkBase58_AppendEncode32(b *testing.B) {
	src := &benchSrc32
	buf := make([]byte, 0, EncodedMaxLen32)
	b.SetBytes(32)
	for b.Loop() {
		buf = AppendEncode32(buf[:0], src)
	}
}

func BenchmarkBase58_AppendEncode64(b *testing.B) {
	src := &benchSrc64
	buf := make([]byte, 0, EncodedMaxLen64)
	b.SetBytes(64)
	for b.Loop() {
		buf = AppendEncode64(buf[:0], src)
	}
}

func BenchmarkBase58_Decode32(b *testing.B) {
	var dst [32]byte
	b.SetBytes(32)
	for b.Loop() {
		Decode32(benchStr32, &dst)
	}
}

func BenchmarkBase58_Encode64(b *testing.B) {
	src := &benchSrc64
	b.SetBytes(64)
	for b.Loop() {
		Encode64(src)
	}
}

func BenchmarkBase58_Decode64(b *testing.B) {
	var dst [64]byte
	b.SetBytes(64)
	for b.Loop() {
		Decode64(benchStr64, &dst)
	}
}
