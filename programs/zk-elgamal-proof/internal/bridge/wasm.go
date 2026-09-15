package bridge

import (
	"context"
	cryptorand "crypto/rand"
	_ "embed"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	zk "github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof"
)

// Solana-zk-sdk Rust prover compiled to wasm32.
//
//go:embed solana_zk_sdk_wasm.wasm
var bridgeWasm []byte

const ALLOC_FUNC = "zk_alloc"
const FREE_FUNC = "zk_free"
const MEMORY_PAGE_LIMIT = 256

var (
	initOnce     sync.Once
	initErr      error
	wasmRuntime  wazero.Runtime
	wasmCompiled wazero.CompiledModule
	// instancePool holds zk.MaxConcurrency() wasm instances, each initialized to nil.
	instancePool chan api.Module
)

func initWasmRuntime() {
	ctx := context.Background()
	rt := wazero.NewRuntimeWithConfig(ctx,
		wazero.NewRuntimeConfig().WithMemoryLimitPages(MEMORY_PAGE_LIMIT))
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

	// Initialize pool of zk.MaxConcurrency() instances.
	size := zk.MaxConcurrency()
	instancePool = make(chan api.Module, size)
	for i := 0; i < size; i++ {
		instancePool <- nil
	}
}

// getInstance takes slot from pool, instantiating slot if nil.
func getInstance() (api.Module, error) {
	initOnce.Do(initWasmRuntime)
	if initErr != nil {
		return nil, initErr
	}
	if inst := <-instancePool; inst != nil {
		// Use instantiated instance from pool.
		return inst, nil
	}
	// Empty slot: build a new instance.
	inst, err := initInstance()
	if err != nil {
		instancePool <- nil
		return nil, err
	}
	return inst, nil
}

// returnInstance hands the slot back to the pool.
func returnInstance(inst api.Module) {
	instancePool <- inst
}

func initInstance() (mod api.Module, err error) {
	ctx := context.Background()

	// Configure as anonymous library module.
	cfg := wazero.NewModuleConfig().
		WithName("").
		WithStartFunctions().             // WASI reactor: suppress _start
		WithRandSource(cryptorand.Reader) // Randomness source must be set explicitly, wazero's default uses fixed seed
	mod, err = wasmRuntime.InstantiateModule(ctx, wasmCompiled, cfg)
	if err != nil {
		return nil, fmt.Errorf("zk: instantiating bridge module: %w", err)
	}
	// Any validation failure must close the instantiated module.
	defer func() {
		if err != nil {
			mod.Close(ctx)
			mod = nil
		}
	}()

	// Make sure that expected memory functions are exported by the WASM ABI
	for _, name := range []string{ALLOC_FUNC, FREE_FUNC} {
		if mod.ExportedFunction(name) == nil {
			return mod, fmt.Errorf("zk: required export %q not found", name)
		}
	}
	return mod, nil
}
