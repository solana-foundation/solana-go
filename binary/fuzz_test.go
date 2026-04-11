package bin

import (
	"encoding/binary"
	"testing"
)

// The binary package is used to decode on-chain Solana data, which is
// attacker-controlled. These fuzz targets exist to shake out panics,
// over-allocations, and infinite loops in the three top-level decoders when
// they're handed arbitrary bytes. A "pass" means: no panic, no OOM, no hang.
// Returning an error is fine — silently accepting malformed input is what
// we're looking for.

type fuzzStruct struct {
	A uint64
	B uint32
	C int16
	D bool
	E string
	F []byte
	G [4]uint64
}

// seedForFuzz returns a buffer produced by the encoder for a known-valid
// value. Seeding the fuzzer with valid-looking bytes gives coverage-guided
// fuzzing a starting point close to the decoder's happy path.
func seedForFuzz(t testing.TB, enc Encoding) []byte {
	t.Helper()
	v := fuzzStruct{
		A: 0xdeadbeefcafebabe,
		B: 0x01020304,
		C: -1234,
		D: true,
		E: "hello",
		F: []byte{1, 2, 3, 4, 5},
		G: [4]uint64{1, 2, 3, 4},
	}
	var buf []byte
	var err error
	switch enc {
	case EncodingBin:
		buf, err = MarshalBin(&v)
	case EncodingBorsh:
		buf, err = MarshalBorsh(&v)
	case EncodingCompactU16:
		buf, err = MarshalCompactU16(&v)
	}
	if err != nil {
		t.Fatalf("seed encode: %v", err)
	}
	return buf
}

func FuzzDecodeBin(f *testing.F) {
	f.Add(seedForFuzz(f, EncodingBin))
	// Extra small seeds to exercise length-prefix edge cases.
	f.Add([]byte{})
	f.Add([]byte{0xff})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		var v fuzzStruct
		_ = NewBinDecoder(data).Decode(&v)
	})
}

func FuzzDecodeBorsh(f *testing.F) {
	f.Add(seedForFuzz(f, EncodingBorsh))
	f.Add([]byte{})
	f.Add([]byte{0xff})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		var v fuzzStruct
		_ = NewBorshDecoder(data).Decode(&v)
	})
}

func FuzzDecodeCompactU16(f *testing.F) {
	f.Add(seedForFuzz(f, EncodingCompactU16))
	f.Add([]byte{})
	f.Add([]byte{0xff})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		var v fuzzStruct
		_ = NewCompactU16Decoder(data).Decode(&v)
	})
}

// FuzzCompactU16Length targets the three compact-u16 length decoders for
// non-canonical encodings, overflow, and continuation-bit handling.
func FuzzCompactU16Length(f *testing.F) {
	f.Add([]byte{0x00})
	f.Add([]byte{0x7f})
	f.Add([]byte{0x80, 0x01})
	f.Add([]byte{0xff, 0xff, 0x03})
	f.Add([]byte{0xff, 0xff, 0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _, _ = DecodeCompactU16(data)
	})
}

// FuzzUint128JSON exercises the big.Int / hex JSON paths, which do their
// own length checks and base-conversion parsing.
func FuzzUint128JSON(f *testing.F) {
	f.Add([]byte(`"0"`))
	f.Add([]byte(`"0x00000000000000000000000000000001"`))
	f.Add([]byte(`"123456789012345678901234567890"`))
	f.Add([]byte(`null`))
	f.Fuzz(func(t *testing.T, data []byte) {
		var v Uint128
		v.Endianness = binary.LittleEndian
		_ = v.UnmarshalJSON(data)
	})
}
