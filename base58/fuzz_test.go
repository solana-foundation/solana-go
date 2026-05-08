package base58

import (
	"bytes"
	"testing"
)

// --- Encode fuzz: fixed-size paths must agree with the variable-length path ---
//
// Encode32/Encode64 use a tuned matrix-multiply that shares no code with
// encodeVariable. Disagreement between them flags a bug in either path.

func FuzzEncode32_MatchesVariable(f *testing.F) {
	f.Add(make([]byte, 32))                       // all zeros
	f.Add(bytes.Repeat([]byte{0xff}, 32))         // all 0xFF
	f.Add(append([]byte{1}, make([]byte, 31)...)) // single leading byte
	f.Add(append(make([]byte, 31), 1))            // trailing 1

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) != 32 {
			t.Skip()
		}
		var src [32]byte
		copy(src[:], data)

		fast := Encode32(&src)
		generic := encodeVariable(src[:])
		if fast != generic {
			t.Fatalf("Encode32 mismatch for %x:\n  fast:    %s\n  generic: %s", src, fast, generic)
		}
	})
}

func FuzzEncode64_MatchesVariable(f *testing.F) {
	f.Add(make([]byte, 64))
	f.Add(bytes.Repeat([]byte{0xff}, 64))
	f.Add(append([]byte{1}, make([]byte, 63)...))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) != 64 {
			t.Skip()
		}
		var src [64]byte
		copy(src[:], data)

		fast := Encode64(&src)
		generic := encodeVariable(src[:])
		if fast != generic {
			t.Fatalf("Encode64 mismatch for %x:\n  fast:    %s\n  generic: %s", src, fast, generic)
		}
	})
}

// --- Decode fuzz: fixed-size and variable-length paths must agree ---

func FuzzDecode32_MatchesVariable(f *testing.F) {
	f.Add("11111111111111111111111111111111")
	f.Add("11111111111111111111111111111112")
	f.Add("4cHoJNmLed5PBgFBezHmJkMJLEZrcTvr3aopjnYBRxUb")

	f.Fuzz(func(t *testing.T, encoded string) {
		var dst [32]byte
		err := Decode32(encoded, &dst)
		if err != nil {
			// We're stricter than the variable-length decoder (fixed size,
			// leading-zero validation). Just verify we don't panic.
			return
		}

		generic, gerr := Decode(encoded)
		if gerr != nil {
			t.Fatalf("Decode32 accepted %q but Decode rejected it: %v", encoded, gerr)
		}

		// Decode strips leading zeros; pad to compare.
		padded := make([]byte, 32)
		copy(padded[32-len(generic):], generic)
		if !bytes.Equal(dst[:], padded) {
			t.Fatalf("decode mismatch for %q:\n  fixed:   %x\n  generic: %x", encoded, dst, padded)
		}

		// Re-encode must produce the original string.
		reEncoded := Encode(dst[:])
		if reEncoded != encoded {
			t.Fatalf("round-trip mismatch: %q -> %x -> %q", encoded, dst, reEncoded)
		}
	})
}

func FuzzDecode64_MatchesVariable(f *testing.F) {
	f.Add("5YBLhMBLjhAHnEPnHKLLnVwHSfXGPJMCvKAfNsiaEw2T63edrYxVFHKUxRXfP6KA1HVo7c9JZ3LAJQR72giX7Cb")

	f.Fuzz(func(t *testing.T, encoded string) {
		var dst [64]byte
		err := Decode64(encoded, &dst)
		if err != nil {
			return
		}

		generic, gerr := Decode(encoded)
		if gerr != nil {
			t.Fatalf("Decode64 accepted %q but Decode rejected it: %v", encoded, gerr)
		}

		padded := make([]byte, 64)
		copy(padded[64-len(generic):], generic)
		if !bytes.Equal(dst[:], padded) {
			t.Fatalf("decode mismatch for %q:\n  fixed:   %x\n  generic: %x", encoded, dst, padded)
		}

		reEncoded := Encode(dst[:])
		if reEncoded != encoded {
			t.Fatalf("round-trip mismatch: %q -> %x -> %q", encoded, dst, reEncoded)
		}
	})
}

// --- Variable-length round-trip fuzz ---

func FuzzEncodeDecode_RoundTrip(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0})
	f.Add([]byte{0, 0, 0})
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add(bytes.Repeat([]byte{0xff}, 100))

	f.Fuzz(func(t *testing.T, data []byte) {
		encoded := Encode(data)
		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("Decode rejected our own Encode output %q: %v", encoded, err)
		}
		if !bytes.Equal(data, decoded) {
			t.Fatalf("round-trip mismatch for %x:\n  encoded: %s\n  decoded: %x", data, encoded, decoded)
		}
	})
}

// --- Invalid input fuzz: verify we never panic ---

func FuzzDecode32_NoPanic(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("0"))                   // invalid char
	f.Add([]byte("O"))                   // invalid char
	f.Add([]byte("I"))                   // invalid char
	f.Add([]byte("l"))                   // invalid char
	f.Add([]byte("\x00"))                // null byte
	f.Add([]byte("\xff"))                // high byte
	f.Add(bytes.Repeat([]byte("z"), 45)) // too long
	f.Add(bytes.Repeat([]byte("1"), 50)) // way too long

	f.Fuzz(func(t *testing.T, data []byte) {
		var dst [32]byte
		// Must not panic regardless of input.
		Decode32(string(data), &dst)
	})
}

func FuzzDecode64_NoPanic(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("0"))
	f.Add([]byte("\x00"))
	f.Add([]byte("\xff"))
	f.Add(bytes.Repeat([]byte("z"), 91))

	f.Fuzz(func(t *testing.T, data []byte) {
		var dst [64]byte
		Decode64(string(data), &dst)
	})
}

func FuzzDecode_NoPanic(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("0"))
	f.Add([]byte("\x00"))
	f.Add([]byte("\xff"))
	f.Add(bytes.Repeat([]byte("z"), 200))

	f.Fuzz(func(t *testing.T, data []byte) {
		Decode(string(data))
	})
}
