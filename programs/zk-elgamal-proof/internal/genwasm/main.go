// Genwasm builds the embedded wasm bridge.
//
// Usage:
//
//	go run ./internal/genwasm          # rebuild and install the blob
//	go run ./internal/genwasm -check   # clean rebuild and byte-compare (CI)
//
// Requires rustup.
package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const wasmName = "solana_zk_sdk_wasm.wasm"

func main() {
	check := flag.Bool("check", false, "clean rebuild and byte-compare against the checked-in blob")
	flag.Parse()
	if err := run(*check); err != nil {
		fmt.Fprintln(os.Stderr, "genwasm:", err)
		os.Exit(1)
	}
}

func run(check bool) error {
	pkgDir, err := findPackageDir()
	if err != nil {
		return err
	}
	crate := filepath.Join(pkgDir, "rust")
	if err := checkToolchain(crate); err != nil {
		return err
	}
	if check { // clean slate so stale artifacts can't mask an irreproducible build
		if err := cargo(crate, nil, "clean"); err != nil {
			return err
		}
	}
	cargoHome := os.Getenv("CARGO_HOME")
	if cargoHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		cargoHome = filepath.Join(home, ".cargo")
	}
	// Remap machine-specific paths out of anything the compiler embeds; the
	// \x1f-separated encoded form survives paths containing spaces (plain
	// RUSTFLAGS is whitespace-split).
	env := []string{"CARGO_ENCODED_RUSTFLAGS=" +
		"--remap-path-prefix=" + cargoHome + "=/cargo\x1f" +
		"--remap-path-prefix=" + crate + "=/build"}
	if err := cargo(crate, env, "build", "--target", "wasm32-wasip1", "--release", "--locked"); err != nil {
		return err
	}
	built, err := os.ReadFile(filepath.Join(crate, "target", "wasm32-wasip1", "release", wasmName))
	if err != nil {
		return fmt.Errorf("reading build output: %w", err)
	}
	sum := sha256.Sum256(built)
	installed := filepath.Join(pkgDir, "internal", "bridge", wasmName)
	if check {
		current, err := os.ReadFile(installed)
		if err != nil {
			return fmt.Errorf("reading checked-in blob: %w", err)
		}
		if !bytes.Equal(built, current) {
			return fmt.Errorf("checked-in %s (sha256 %x) does not match rebuild from pinned sources (sha256 %x); run `go generate ./programs/zk-elgamal-proof` and commit the result",
				wasmName, sha256.Sum256(current), sum)
		}
		fmt.Printf("genwasm: %s matches pinned sources (sha256 %x)\n", wasmName, sum)
		return nil
	}
	if err := os.WriteFile(installed, built, 0o644); err != nil {
		return err
	}
	fmt.Printf("genwasm: installed %s (%d bytes, sha256 %x)\n", wasmName, len(built), sum)
	return nil
}

func cargo(dir string, env []string, args ...string) error {
	cmd := exec.Command("cargo", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cargo %s failed: %w", args[0], err)
	}
	return nil
}

// checkToolchain fails fast when the cargo in PATH won't honor the pin in rust-toolchain.toml
func checkToolchain(crate string) error {
	if _, err := exec.LookPath("rustup"); err != nil {
		return fmt.Errorf("rustup not found in PATH; a non-rustup cargo ignores rust-toolchain.toml, so the build would not be reproducible (install from https://rustup.rs)")
	}
	toml, err := os.ReadFile(filepath.Join(crate, "rust-toolchain.toml"))
	if err != nil {
		return fmt.Errorf("reading toolchain pin: %w", err)
	}
	m := regexp.MustCompile(`(?m)^\s*channel\s*=\s*"([^"]+)"`).FindSubmatch(toml)
	if m == nil {
		return fmt.Errorf("no channel pin found in rust-toolchain.toml")
	}
	// Run in the crate dir so the rustup shim applies the pin.
	cmd := exec.Command("cargo", "--version")
	cmd.Dir = crate
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cargo --version failed: %w", err)
	}
	if f := strings.Fields(string(out)); len(f) < 2 || f[1] != string(m[1]) {
		return fmt.Errorf("active toolchain is %q but rust-toolchain.toml pins %s; the cargo in PATH is likely not rustup's shim (rebuilds would not be reproducible)",
			strings.TrimSpace(string(out)), m[1])
	}
	return nil
}

// findPackageDir returns the zk package directory.
func findPackageDir() (string, error) {
	for _, dir := range []string{".", "programs/zk-elgamal-proof"} {
		if _, err := os.Stat(filepath.Join(dir, "rust", "Cargo.toml")); err == nil {
			return filepath.Abs(dir)
		}
	}
	return "", fmt.Errorf("cannot locate the zk package (run from the repo root or programs/zk-elgamal-proof/)")
}
