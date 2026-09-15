package bridge

import (
	"context"
	_ "embed"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/tetratelabs/wazero/api"

	zk "github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof"
)

// decoder is a pointer to T that can decode result bytes.
type decoder[T any] interface {
	*T
	encoding.BinaryUnmarshaler
}

// InvokeWith calls the named export and decodes its result into a T.
// UnmarshalBinary must copy: the result buffer is scrubbed after decoding.
func InvokeWith[T any, PT decoder[T]](name string, parts ...Arg) (T, error) {
	var v T
	out, err := invoke(name, parts...)
	if err != nil {
		return v, err
	}
	defer zeroize(out)
	if err := PT(&v).UnmarshalBinary(out); err != nil {
		return v, err
	}
	return v, nil
}

// InvokeStatus is invoke for exports that return only a status code.
func InvokeStatus(name string, parts ...Arg) error {
	_, err := invoke(name, parts...)
	return err
}

// invoke borrows an instance, marshals parts into export arguments, calls the named export,
// and copies out its result.
func invoke(name string, parts ...Arg) ([]byte, error) {
	f := &frame{}
	if err := f.acquireInstance(); err != nil {
		return nil, err
	}
	defer f.releaseInstance()

	args, err := buildArgs(f, parts...)
	if err != nil {
		return nil, err
	}

	fn := f.inst.ExportedFunction(name)
	if fn == nil {
		return nil, fmt.Errorf("zk: bridge export %q not found", name)
	}
	res, err := fn.Call(context.Background(), args...)
	if err != nil {
		f.poisoned = true
		return nil, fmt.Errorf("zk: calling %s: %w", name, err)
	}
	if len(res) != 1 {
		return nil, fmt.Errorf("zk: %s returned %d results, want 1", name, len(res))
	}
	return f.decodeResult(res[0])
}

// zeroize clears a transient buffer holding secret material.
func zeroize(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// span is one guest allocation owned by a frame.
type span struct {
	ptr, size uint32
}

// frame controls the memory allocations of a single wasm call.
type frame struct {
	inst     api.Module
	allocs   []span
	poisoned bool
}

// acquireInstance binds the frame to a pooled instance, blocking until a free-slot is available.
func (f *frame) acquireInstance() error {
	inst, err := getInstance()
	if err != nil {
		return err
	}
	f.inst = inst
	return nil
}

// releaseInstance releases the call's module instance back to the pool.
func (f *frame) releaseInstance() {
	f.scrubAndFreeMemory()
	if f.poisoned {
		// Close the instance, feed nil slot to pool.
		f.inst.Close(context.Background())
		f.inst = nil
	}
	returnInstance(f.inst)
}

func (f *frame) scrubAndFreeMemory() {
	mem := f.inst.Memory()
	if mem == nil {
		return
	}
	free := f.inst.ExportedFunction(FREE_FUNC)
	ctx := context.Background()
	var zeros []byte
	for _, a := range f.allocs {
		if int(a.size) > len(zeros) {
			zeros = make([]byte, a.size)
		}

		mem.Write(a.ptr, zeros[:a.size])
		if f.poisoned {
			continue
		}
		if _, err := free.Call(ctx, uint64(a.ptr)); err != nil {
			f.poisoned = true
		}
	}
}

// write copies b into guest memory and returns the guest pointer.
func (f *frame) write(b []byte) (uint64, error) {
	if len(b) == 0 {
		return 0, errors.New("zk: attempted to write empty buffer")
	}
	res, err := f.inst.ExportedFunction(ALLOC_FUNC).Call(context.Background(), uint64(len(b)))
	if err != nil {
		f.poisoned = true
		return 0, fmt.Errorf("zk: guest alloc: %w", err)
	}
	if len(res) != 1 {
		return 0, fmt.Errorf("zk: zk_alloc returned %d results, want 1", len(res))
	}
	// zk_alloc returns 0 to signal allocation failure
	ptr := uint32(res[0])
	if ptr == 0 {
		return 0, zk.Error(zk.OOM)
	}
	f.allocs = append(f.allocs, span{ptr, uint32(len(b))})
	if !f.inst.Memory().Write(ptr, b) {
		return 0, errors.New("zk: guest memory write out of range")
	}
	return uint64(ptr), nil
}

// decodeResult interprets an export's return value
func (f *frame) decodeResult(raw uint64) ([]byte, error) {
	packed := int64(raw)
	if packed < 0 {
		return nil, zk.Error(int32(packed))
	}
	if packed == 0 {
		return nil, nil
	}
	ptr, length := uint32(packed), uint32(packed>>32)
	f.allocs = append(f.allocs, span{ptr, length})
	view, ok := f.inst.Memory().Read(ptr, length)
	if !ok {
		return nil, errors.New("zk: guest memory read out of range")
	}
	out := make([]byte, length)
	copy(out, view)
	return out, nil
}

// buildArgs marshals each wasm call argument into guest memory.
func buildArgs(f *frame, parts ...Arg) ([]uint64, error) {
	args := make([]uint64, 0, len(parts))
	for _, part := range parts {
		if s, ok := part.(Scalar); ok {
			args = append(args, uint64(s))
			continue
		}
		b, err := part.MarshalBinary()
		if err != nil {
			return nil, err
		}
		ptr, err := f.write(b)
		zeroize(b)
		if err != nil {
			return nil, err
		}
		args = append(args, ptr)
	}
	return args, nil
}

// Arg is a value that can cross into a wasm call
type Arg interface {
	MarshalBinary() ([]byte, error)
}

// Scalar is an integer passed directly as a wasm argument, and the decoded
// form of an 8-byte little-endian result. Use U64s for a u64 sequence that
// crosses as a buffer.
type Scalar uint64

// MarshalBinary makes Scalar an Arg. A Scalar is never marshaled: it is
// passed as an integer argument, so reaching here is a misuse.
func (s Scalar) MarshalBinary() ([]byte, error) {
	return nil, errors.New("zk: Scalar crosses as an integer argument, not a buffer; use U64s for a u64 buffer")
}

func (s *Scalar) UnmarshalBinary(b []byte) error {
	if len(b) != 8 {
		return fmt.Errorf("zk: guest returned %d bytes, want 8", len(b))
	}
	*s = Scalar(binary.LittleEndian.Uint64(b))
	return nil
}

type U64s []uint64

func (s U64s) MarshalBinary() ([]byte, error) {
	out := make([]byte, len(s)*8)
	for i, v := range s {
		binary.LittleEndian.PutUint64(out[i*8:], v)
	}
	return out, nil
}

// Bytes marshals an already-serialized buffer. MarshalBinary copies, so the
// scrub after the guest write cannot reach the caller's slice.
type Bytes []byte

func (b Bytes) MarshalBinary() ([]byte, error) { return append([]byte(nil), b...), nil }

// Slice marshals as the concatenation of its elements' marshalings.
type Slice[E encoding.BinaryMarshaler] []E

func (s Slice[E]) MarshalBinary() ([]byte, error) {
	var out []byte
	for i := range s {
		b, err := s[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
		out = append(out, b...)
		zeroize(b)
	}
	return out, nil
}
