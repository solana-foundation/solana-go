package zk

import (
	"context"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Solana-zk-sdk Rust prover compiled to wasm32.
//
//go:embed solana_zk_sdk_wasm.wasm
var bridgeWasm []byte

var poolSize = runtime.NumCPU()

const ALLOC_FUNC = "zk_alloc"
const FREE_FUNC = "zk_free"

var (
	initOnce     sync.Once
	initErr      error
	wasmRuntime  wazero.Runtime
	wasmCompiled wazero.CompiledModule
	instancePool = make(chan api.Module, poolSize)
)

func initialize() {
	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		rt.Close(ctx)
		initErr = fmt.Errorf("zk: instantiating WASI host module: %w", err)
		return
	}
	compiled, err := rt.CompileModule(ctx, bridgeWasm)
	if err != nil {
		rt.Close(ctx)
		initErr = fmt.Errorf("zk: compiling bridge module: %w", err)
		return
	}
	wasmRuntime = rt
	wasmCompiled = compiled
}

func acquireInstance() (api.Module, error) {
	initOnce.Do(initialize)
	if initErr != nil {
		return nil, initErr
	}
	select {
	case inst := <-instancePool:
		return inst, nil
	default:
	}
	ctx := context.Background()

	// Configure as anonymous library module
	cfg := wazero.NewModuleConfig().WithName("").WithStartFunctions() // WASI reactor: suppress _start
	mod, err := wasmRuntime.InstantiateModule(ctx, wasmCompiled, cfg)
	if err != nil {
		return nil, fmt.Errorf("zk: instantiating bridge module: %w", err)
	}
	if initFn := mod.ExportedFunction("_initialize"); initFn != nil {
		if _, err := initFn.Call(ctx); err != nil {
			mod.Close(ctx)
			return nil, fmt.Errorf("zk: running module _initialize: %w", err)
		}
	}
	// Make sure that expected memory functions are exported by the WASM ABI
	for _, name := range []string{ALLOC_FUNC, FREE_FUNC} {
		if mod.ExportedFunction(name) == nil {
			mod.Close(ctx)
			return nil, fmt.Errorf("zk: required export %q not found", name)
		}
	}
	return mod, nil
}

func releaseInstance(inst api.Module) {
	select {
	case instancePool <- inst:
	default:
		inst.Close(context.Background())
	}
}

// span is one guest allocation owned by a frame.
type span struct {
	ptr, size uint32
}

// frame controls the memory allocations of a single bridge call.
type frame struct {
	inst     api.Module
	allocs   []span
	poisoned bool
}

// acquire binds the frame to a pooled instance, instantiating a fresh one if
// the pool is empty. Every acquired frame must be released.
func (f *frame) acquire() error {
	inst, err := acquireInstance()
	if err != nil {
		return err
	}
	f.inst = inst
	return nil
}

// release scrubs and frees the call's guest allocations and returns the
// instance to the pool, or closes the instance if poisoned.
func (f *frame) release() {
	if !f.poisoned {
		free := f.inst.ExportedFunction(FREE_FUNC)
		ctx := context.Background()
		var zeros []byte
		for _, a := range f.allocs {
			if int(a.size) > len(zeros) {
				zeros = make([]byte, a.size)
			}
			f.inst.Memory().Write(a.ptr, zeros[:a.size])
			if _, err := free.Call(ctx, uint64(a.ptr)); err != nil {
				f.poisoned = true
				break
			}
		}
	}
	if f.poisoned {
		f.inst.Close(context.Background())
		return
	}
	releaseInstance(f.inst)
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
		return 0, Error(OOM)
	}
	f.allocs = append(f.allocs, span{ptr, uint32(len(b))})
	if !f.inst.Memory().Write(ptr, b) {
		return 0, errors.New("zk: guest memory write out of range")
	}
	return uint64(ptr), nil
}

// invokeWith borrows an instance, marshals parts into export arguments, calls the named export,
// and copies out its result.
func invokeWith(name string, parts ...any) ([]byte, error) {
	f := &frame{}
	if err := f.acquire(); err != nil {
		return nil, err
	}
	defer f.release()

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

	packed := int64(res[0])
	if packed < 0 {
		return nil, Error(int32(packed))
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

// buildArgs writes each []byte part into guest memory (appending its pointer
// to the argument list) and passes uint64 parts through as-is, preserving
// order.
func buildArgs(f *frame, parts ...any) ([]uint64, error) {
	args := make([]uint64, 0, len(parts))
	for _, part := range parts {
		switch v := part.(type) {
		case []byte:
			ptr, err := f.write(v)
			if err != nil {
				return nil, err
			}
			args = append(args, ptr)
		case uint64:
			args = append(args, v)
		default:
			return nil, fmt.Errorf("zk: unsupported argument type %T", part)
		}
	}
	return args, nil
}

// invokeStatus is invokeWith for exports that return only a status code.
func invokeStatus(name string, parts ...any) error {
	_, err := invokeWith(name, parts...)
	return err
}

// copyOut copies a guest result into dst, rejecting results whose length does
// not match exactly.
func copyOut(dst, out []byte, err error) error {
	if err != nil {
		return err
	}
	if len(out) != len(dst) {
		return fmt.Errorf("zk: guest returned %d bytes, want %d", len(out), len(dst))
	}
	copy(dst, out)
	return nil
}

// toAmount decodes a guest result holding a little-endian u64 amount.
func toAmount(out []byte, err error) (uint64, error) {
	if err != nil {
		return 0, err
	}
	if len(out) != 8 {
		return 0, fmt.Errorf("zk: guest returned %d bytes, want 8", len(out))
	}
	return binary.LittleEndian.Uint64(out), nil
}
