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
	"errors"
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

// ErrNotPOD is returned by AssertPOD when T is not safe for zero-copy
// reinterpretation. See the function doc for the full set of checks.
var ErrNotPOD = errors.New("binary: type is not POD for zero-copy view")

// ViewAs reinterprets buf as *T for in-place read/write access. The
// returned pointer aliases buf's backing array: mutations through *T
// write straight into the caller's buffer and reads reflect the
// buffer's current contents. No allocation, no copy.
//
// This is the Go analog of wincode's deserialize_mut. Typical use:
// parse a Solana transaction / compute-budget instruction / durable
// nonce advance, patch a single field (priority fee, blockhash,
// nonce), then send the same buffer on the wire — no re-serialize.
//
// CONSTRAINTS. For the cast to be sound, T must satisfy all of these.
// Verify at program start with AssertPOD[T] rather than debugging at
// first mutation.
//
//  1. Fixed-size POD: no pointers, slices, maps, strings, interfaces,
//     channels, or functions. Only bool, fixed-width integers, floats,
//     fixed-width arrays of POD, and structs composed of POD.
//
//  2. No Go-compiler-inserted padding. Declare fields in descending
//     alignment order, or reorder so offsets are contiguous. AssertPOD
//     catches violations at run time.
//
//  3. Platform-dependent types (int, uint, uintptr) are rejected —
//     their wire-size differs between 32- and 64-bit builds.
//
//  4. Wire byte order must match host byte order. On amd64 and arm64
//     (the common Solana targets) this is little-endian, so borsh /
//     Solana compact-u16 / bincode integer fields are compatible.
//     Big-endian hosts require byte-swaps ViewAs cannot do.
//
//  5. buf[0] must be aligned to alignof(T). Slices from make() are
//     8- or 16-byte aligned, but subslicing (buf[3:]) can produce
//     misaligned pointers. Misaligned access is legal on amd64/arm64
//     (with a small cost) and UB on strict-alignment architectures.
//
// Returns io.ErrShortBuffer when len(buf) < unsafe.Sizeof(T{}).
//
// The returned *T is only valid while buf is alive and its backing
// array is not reallocated. Retaining it past buf's lifetime is a
// use-after-free; retaining it across an append that grows the backing
// array is a silent read of stale memory.
func ViewAs[T any](buf []byte) (*T, error) {
	var zero T
	sz := unsafe.Sizeof(zero)
	if uintptr(len(buf)) < sz {
		return nil, io.ErrShortBuffer
	}
	return (*T)(unsafe.Pointer(&buf[0])), nil
}

// MustViewAs is ViewAs without the error return. Panics with the same
// io.ErrShortBuffer if buf is too small. Intended for call sites where
// the caller has statically verified the buffer size.
func MustViewAs[T any](buf []byte) *T {
	p, err := ViewAs[T](buf)
	if err != nil {
		panic(err)
	}
	return p
}

// ViewSliceAs reinterprets buf as []T of length n. Same constraints as
// ViewAs apply to T. The returned slice aliases buf's backing array:
// mutations write through and reads reflect the buffer.
//
// Typical use: a Solana account-keys array (N * 32-byte pubkeys) or a
// signatures array (N * 64-byte signatures) exposed as []Pubkey /
// []Signature for O(1) field access.
//
// Returns io.ErrShortBuffer when len(buf) < n * unsafe.Sizeof(T{}).
func ViewSliceAs[T any](buf []byte, n int) ([]T, error) {
	if n < 0 {
		return nil, fmt.Errorf("binary: negative slice length %d", n)
	}
	if n == 0 {
		return nil, nil
	}
	var zero T
	sz := unsafe.Sizeof(zero)
	total := uintptr(n) * sz
	if uintptr(len(buf)) < total {
		return nil, io.ErrShortBuffer
	}
	return unsafe.Slice((*T)(unsafe.Pointer(&buf[0])), n), nil
}

// MustViewSliceAs is ViewSliceAs without the error return. Panics on
// short buffer or negative n.
func MustViewSliceAs[T any](buf []byte, n int) []T {
	s, err := ViewSliceAs[T](buf, n)
	if err != nil {
		panic(err)
	}
	return s
}

// AssertPOD verifies that T satisfies the zero-copy view constraints
// (see ViewAs). On violation it returns an error naming the offending
// field path.
//
// Call from init() or a test to validate at program start:
//
//	func init() {
//	    if err := AssertPOD[MessageHeader](); err != nil {
//	        panic(err)
//	    }
//	}
//
// AssertPOD allocates temporarily via reflect; it is not zero-alloc.
// Do not call from hot paths.
func AssertPOD[T any]() error {
	var zero T
	rt := reflect.TypeOf(zero)
	if rt == nil {
		return fmt.Errorf("%w: nil type", ErrNotPOD)
	}
	return assertPODType(rt, rt.String())
}

// MustAssertPOD is AssertPOD that panics on violation. Suitable for
// init() blocks where failure is fatal.
func MustAssertPOD[T any]() {
	if err := AssertPOD[T](); err != nil {
		panic(err)
	}
}

func assertPODType(rt reflect.Type, path string) error {
	switch rt.Kind() {
	case reflect.Bool,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return nil
	case reflect.Array:
		return assertPODType(rt.Elem(), path+"[]")
	case reflect.Struct:
		var expected uintptr
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			sub := path + "." + f.Name
			if err := assertPODType(f.Type, sub); err != nil {
				return err
			}
			if f.Offset != expected {
				return fmt.Errorf("%w: padding before %s (offset=%d, expected=%d)",
					ErrNotPOD, sub, f.Offset, expected)
			}
			expected += f.Type.Size()
		}
		if expected != rt.Size() {
			return fmt.Errorf("%w: tail padding in %s (size=%d, expected=%d)",
				ErrNotPOD, path, rt.Size(), expected)
		}
		return nil
	case reflect.Int, reflect.Uint, reflect.Uintptr:
		return fmt.Errorf("%w: %s has platform-dependent size (%s)", ErrNotPOD, path, rt.Kind())
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.String:
		return fmt.Errorf("%w: %s has non-POD kind %s", ErrNotPOD, path, rt.Kind())
	default:
		return fmt.Errorf("%w: %s has unsupported kind %s", ErrNotPOD, path, rt.Kind())
	}
}
