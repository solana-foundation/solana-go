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

package bin

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"unsafe"
)

// Solana-shaped POD types used in tests. Field order is chosen so there
// is no Go-compiler-inserted padding.

// messageHeaderView: three 1-byte counters. Wire size = Go size = 3.
type messageHeaderView struct {
	NumRequiredSignatures       uint8
	NumReadonlySignedAccounts   uint8
	NumReadonlyUnsignedAccounts uint8
}

// pubkey / signature / blockhash are fixed-width byte arrays — the
// simplest POD shape. No alignment caveats, no padding, Go size =
// wire size.
type pubkeyView [32]byte
type signatureView [64]byte
type blockhashView [32]byte

// homogeneousFeeView: a u64-only struct. All fields have the same
// alignment so Go does not insert padding. Wire size = Go size = 16.
// (Contrast with a naive {u8, u64} layout where Go pads the u8 to 8
// bytes, breaking wire compatibility — AssertPOD catches that.)
type homogeneousFeeView struct {
	A uint64
	B uint64
}

// ---------- ViewAs basics ----------

func TestViewAs_ReadThroughBuffer(t *testing.T) {
	buf := []byte{1, 2, 3}
	hdr, err := ViewAs[messageHeaderView](buf)
	if err != nil {
		t.Fatalf("ViewAs failed: %v", err)
	}
	if hdr.NumRequiredSignatures != 1 ||
		hdr.NumReadonlySignedAccounts != 2 ||
		hdr.NumReadonlyUnsignedAccounts != 3 {
		t.Fatalf("read mismatch: %+v", hdr)
	}
}

func TestViewAs_WriteThroughBuffer(t *testing.T) {
	buf := make([]byte, 3)
	hdr, err := ViewAs[messageHeaderView](buf)
	if err != nil {
		t.Fatal(err)
	}
	hdr.NumRequiredSignatures = 7
	hdr.NumReadonlySignedAccounts = 11
	hdr.NumReadonlyUnsignedAccounts = 13

	if buf[0] != 7 || buf[1] != 11 || buf[2] != 13 {
		t.Fatalf("buffer not updated: %v", buf)
	}
}

func TestViewAs_ShortBuffer(t *testing.T) {
	// messageHeaderView is 3 bytes; feed 2.
	_, err := ViewAs[messageHeaderView](make([]byte, 2))
	if !errors.Is(err, io.ErrShortBuffer) {
		t.Fatalf("expected io.ErrShortBuffer, got %v", err)
	}
}

func TestMustViewAs_PanicsOnShortBuffer(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = MustViewAs[messageHeaderView](make([]byte, 2))
}

// ---------- ViewAs with Solana-shaped structs ----------

// TestViewAs_BlockhashPatch demonstrates the wincode deserialize_mut
// use case on a realistic Solana shape: patching the recent blockhash
// in a pre-built transaction message without re-serializing.
//
// The blockhash is a fixed 32-byte field at a known offset — exactly
// the kind of layout ViewAs handles safely. (Contrast with a priority-
// fee patch of {u8 disc, u64 microlamports}: that shape is 9 bytes on
// the wire but Go structs add 7 bytes of padding, so ViewAs is *not*
// sound there. For mixed-width packed formats, use byte-level
// patching via Cursor or binary.LittleEndian.PutUint64 instead.)
func TestViewAs_BlockhashPatch(t *testing.T) {
	// Pretend the wire holds: 3 header bytes, then a 32-byte
	// blockhash, then more payload. Fill with sentinel bytes so we
	// can detect corruption of unrelated regions.
	const blockhashOffset = 3
	wire := make([]byte, 100)
	for i := range wire {
		wire[i] = 0x5a
	}

	// Zero the blockhash region for a clean starting state.
	for i := blockhashOffset; i < blockhashOffset+32; i++ {
		wire[i] = 0
	}

	// Patch the blockhash in place via ViewAs on the subslice.
	newHash := blockhashView{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7,
		0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7,
		0xc0, 0xc1, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7,
		0xd0, 0xd1, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7}
	v, err := ViewAs[blockhashView](wire[blockhashOffset:])
	if err != nil {
		t.Fatal(err)
	}
	*v = newHash

	// Verify the patch took effect in the backing buffer.
	if !bytes.Equal(wire[blockhashOffset:blockhashOffset+32], newHash[:]) {
		t.Fatalf("blockhash region mismatch:\n got: %x\nwant: %x",
			wire[blockhashOffset:blockhashOffset+32], newHash[:])
	}
	// Verify surrounding bytes are untouched.
	for i := 0; i < blockhashOffset; i++ {
		if wire[i] != 0x5a {
			t.Errorf("header byte %d corrupted: got %x", i, wire[i])
		}
	}
	for i := blockhashOffset + 32; i < len(wire); i++ {
		if wire[i] != 0x5a {
			t.Errorf("tail byte %d corrupted: got %x", i, wire[i])
		}
	}
}

// ---------- ViewSliceAs ----------

func TestViewSliceAs_PubkeyArray(t *testing.T) {
	// 4 pubkeys = 128 bytes. Fill with distinct bytes so we can detect
	// ordering issues.
	buf := make([]byte, 128)
	for i := range buf {
		buf[i] = byte(i)
	}

	keys, err := ViewSliceAs[pubkeyView](buf, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 4 {
		t.Fatalf("expected 4 keys, got %d", len(keys))
	}
	// Spot-check boundaries.
	if keys[0][0] != 0 || keys[0][31] != 31 {
		t.Errorf("key[0] boundaries wrong: %v", keys[0])
	}
	if keys[1][0] != 32 {
		t.Errorf("key[1] boundaries wrong: %v", keys[1])
	}
	if keys[3][31] != 127 {
		t.Errorf("key[3] boundaries wrong: %v", keys[3])
	}

	// Mutate through the slice; verify it reaches the backing buffer.
	keys[2][0] = 0xaa
	if buf[64] != 0xaa {
		t.Errorf("mutation did not write through: buf[64]=%x", buf[64])
	}
}

func TestViewSliceAs_Empty(t *testing.T) {
	s, err := ViewSliceAs[pubkeyView](nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if s != nil {
		t.Fatalf("expected nil slice, got %v", s)
	}
}

func TestViewSliceAs_ShortBuffer(t *testing.T) {
	_, err := ViewSliceAs[pubkeyView](make([]byte, 31), 1)
	if !errors.Is(err, io.ErrShortBuffer) {
		t.Fatalf("expected io.ErrShortBuffer, got %v", err)
	}
}

func TestViewSliceAs_NegativeLen(t *testing.T) {
	_, err := ViewSliceAs[pubkeyView](make([]byte, 32), -1)
	if err == nil {
		t.Fatal("expected error for negative length")
	}
}

// ---------- AssertPOD ----------

func TestAssertPOD_AcceptsSafeTypes(t *testing.T) {
	cases := []func() error{
		AssertPOD[uint8],
		AssertPOD[uint64],
		AssertPOD[pubkeyView],
		AssertPOD[signatureView],
		AssertPOD[blockhashView],
		AssertPOD[messageHeaderView],
		AssertPOD[homogeneousFeeView],
		AssertPOD[[8]uint32],
		AssertPOD[struct {
			A uint64
			B [32]byte
			C uint32
			D uint32 // paired with C to avoid tail padding
		}],
	}
	for i, fn := range cases {
		if err := fn(); err != nil {
			t.Errorf("case %d: AssertPOD rejected safe type: %v", i, err)
		}
	}
}

func TestAssertPOD_RejectsImplicitPadding(t *testing.T) {
	// Go inserts 7 bytes of padding after A to align B to 8.
	type padded struct {
		A uint8
		B uint64
	}
	err := AssertPOD[padded]()
	if err == nil {
		t.Fatal("expected padding rejection")
	}
	if !errors.Is(err, ErrNotPOD) {
		t.Fatalf("expected ErrNotPOD, got %v", err)
	}
}

func TestAssertPOD_RejectsTailPadding(t *testing.T) {
	// 9 bytes of fields + 7 tail-padding bytes = 16 total.
	type tailPadded struct {
		A uint64
		B uint8
	}
	err := AssertPOD[tailPadded]()
	if err == nil {
		t.Fatal("expected tail padding rejection")
	}
	if !errors.Is(err, ErrNotPOD) {
		t.Fatalf("expected ErrNotPOD, got %v", err)
	}
}

func TestAssertPOD_RejectsNonPODKinds(t *testing.T) {
	cases := []struct {
		name string
		run  func() error
	}{
		{"pointer", AssertPOD[*uint64]},
		{"slice", AssertPOD[[]uint64]},
		{"map", AssertPOD[map[uint64]uint64]},
		{"string", AssertPOD[string]},
		{"interface", AssertPOD[any]},
		{"plain-int", AssertPOD[int]},
		{"plain-uint", AssertPOD[uint]},
		{"uintptr", AssertPOD[uintptr]},
		{"struct-with-slice", AssertPOD[struct {
			A uint64
			B []byte
		}]},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.run()
			if err == nil {
				t.Fatal("expected rejection")
			}
			if !errors.Is(err, ErrNotPOD) {
				t.Errorf("expected ErrNotPOD, got %v", err)
			}
		})
	}
}

func TestMustAssertPOD_PanicsOnUnsafe(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	MustAssertPOD[map[uint64]uint64]()
}

// ---------- ViewAs alias invariant ----------

// Confirm the pointer returned by ViewAs aliases the caller's buffer —
// there is no copy.
func TestViewAs_NoCopy(t *testing.T) {
	buf := make([]byte, 3)
	hdr, _ := ViewAs[messageHeaderView](buf)
	if unsafe.Pointer(hdr) != unsafe.Pointer(&buf[0]) {
		t.Fatal("ViewAs returned a copy, not an alias")
	}
}

// ---------- ViewAs zero-allocation ----------

func TestViewAs_ZeroAllocs(t *testing.T) {
	if testing.Short() {
		t.Skip("allocation test disabled under -short")
	}
	buf := make([]byte, 128)
	allocs := testing.AllocsPerRun(100, func() {
		v, err := ViewAs[homogeneousFeeView](buf)
		if err != nil {
			t.Fatal(err)
		}
		v.A++
	})
	if allocs != 0 {
		t.Fatalf("ViewAs allocs = %g, want 0", allocs)
	}
}
