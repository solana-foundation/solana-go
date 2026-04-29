package base58

import (
	"bytes"
	"testing"

	mrtronbase58 "github.com/mr-tron/base58"
)

// --- Encode fuzz: our output must match mr-tron for every input ---

func FuzzEncode32_MatchesMrTron(f *testing.F) {
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

		ours := Encode32(&src)
		theirs := mrtronbase58.Encode(src[:])
		if ours != theirs {
			t.Fatalf("Encode32 mismatch for %x:\n  ours:   %s\n  theirs: %s", src, ours, theirs)
		}
	})
}

func FuzzEncode64_MatchesMrTron(f *testing.F) {
	f.Add(make([]byte, 64))
	f.Add(bytes.Repeat([]byte{0xff}, 64))
	f.Add(append([]byte{1}, make([]byte, 63)...))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) != 64 {
			t.Skip()
		}
		var src [64]byte
		copy(src[:], data)

		ours := Encode64(&src)
		theirs := mrtronbase58.Encode(src[:])
		if ours != theirs {
			t.Fatalf("Encode64 mismatch for %x:\n  ours:   %s\n  theirs: %s", src, ours, theirs)
		}
	})
}

// --- Decode fuzz: round-trip through mr-tron must agree ---

func FuzzDecode32_MatchesMrTron(f *testing.F) {
	f.Add("11111111111111111111111111111111")
	f.Add("11111111111111111111111111111112")
	f.Add("4cHoJNmLed5PBgFBezHmJkMJLEZrcTvr3aopjnYBRxUb")

	f.Fuzz(func(t *testing.T, encoded string) {
		var dst [32]byte
		err := Decode32(encoded, &dst)
		if err != nil {
			// We're stricter than mr-tron (fixed size, leading-zero
			// validation). Just verify we don't panic.
			return
		}

		// If we accepted it, mr-tron must decode to the same bytes.
		theirBytes, theirErr := mrtronbase58.Decode(encoded)
		if theirErr != nil {
			t.Fatalf("we accepted %q but mr-tron rejected it: %v", encoded, theirErr)
		}

		// mr-tron strips leading zeros; pad to compare.
		padded := make([]byte, 32)
		copy(padded[32-len(theirBytes):], theirBytes)
		if !bytes.Equal(dst[:], padded) {
			t.Fatalf("decode mismatch for %q:\n  ours:   %x\n  theirs: %x", encoded, dst, padded)
		}

		// Re-encode must produce the original string.
		reEncoded := Encode32(&dst)
		if reEncoded != encoded {
			t.Fatalf("round-trip mismatch: %q -> %x -> %q", encoded, dst, reEncoded)
		}
	})
}

func FuzzDecode64_MatchesMrTron(f *testing.F) {
	f.Add("5YBLhMBLjhAHnEPnHKLLnVwHSfXGPJMCvKAfNsiaEw2T63edrYxVFHKUxRXfP6KA1HVo7c9JZ3LAJQR72giX7Cb")

	f.Fuzz(func(t *testing.T, encoded string) {
		var dst [64]byte
		err := Decode64(encoded, &dst)
		if err != nil {
			return
		}

		theirBytes, theirErr := mrtronbase58.Decode(encoded)
		if theirErr != nil {
			t.Fatalf("we accepted %q but mr-tron rejected it: %v", encoded, theirErr)
		}

		padded := make([]byte, 64)
		copy(padded[64-len(theirBytes):], theirBytes)
		if !bytes.Equal(dst[:], padded) {
			t.Fatalf("decode mismatch for %q:\n  ours:   %x\n  theirs: %x", encoded, dst, padded)
		}

		reEncoded := Encode64(&dst)
		if reEncoded != encoded {
			t.Fatalf("round-trip mismatch: %q -> %x -> %q", encoded, dst, reEncoded)
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
