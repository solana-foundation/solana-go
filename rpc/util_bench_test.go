package rpc

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/solana-foundation/solana-go/v2"
)

var isTokenMintBenchSink bool

// isTokenMintPR396 is the classifier as proposed in PR #396. Kept in a test
// file so benchmarks can measure the current implementation against it
// without polluting the public API. Intentionally preserved verbatim —
// including the missing nil-guard on acc — so the comparison reflects the
// PR as submitted.
func isTokenMintPR396(acc *Account) bool {
	data := acc.Data.GetBinary()
	n := len(data)

	switch acc.Owner {
	case solana.TokenProgramID:
		return n == 82
	case solana.Token2022ProgramID:
		if n == 82 {
			return true //Normal Mint
		}
		if n <= 165 {
			return false //Normal Token Account
		}
		return data[165] == 1 // Mint Extensions
	}

	return false
}

// buildMixedAccounts returns a set of account shapes approximating an
// unfiltered getProgramAccounts result: bare mints, extended mints,
// accounts, and non-token junk. Indexing by i prevents the compiler from
// constant-folding benchmark inputs.
func buildMixedAccounts() []*Account {
	extMint := append(make([]byte, 165), 1)
	extMint = append(extMint, bytes.Repeat([]byte{3, 0, 32, 0}, 1)...)
	extAcc := append(make([]byte, 165), 2)
	return []*Account{
		{Owner: solana.TokenProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 82))},
		{Owner: solana.TokenProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 165))},
		{Owner: solana.Token2022ProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 82))},
		{Owner: solana.Token2022ProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 165))},
		{Owner: solana.Token2022ProgramID, Data: DataBytesOrJSONFromBytes(extMint)},
		{Owner: solana.Token2022ProgramID, Data: DataBytesOrJSONFromBytes(extAcc)},
		{Owner: solana.SystemProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 0))},
		{Owner: solana.SystemProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 200))},
	}
}

func runIsTokenMintBench(b *testing.B, fn func(*Account) bool, accounts []*Account) {
	var acc bool
	var i int
	for b.Loop() {
		acc = acc != fn(accounts[i%len(accounts)])
		i++
	}
	isTokenMintBenchSink = acc
	runtime.KeepAlive(&isTokenMintBenchSink)
}

// --- PR #396 version (baseline) ---

func BenchmarkIsTokenMint_PR396_Mixed(b *testing.B) {
	runIsTokenMintBench(b, isTokenMintPR396, buildMixedAccounts())
}

func BenchmarkIsTokenMint_PR396_HotMint(b *testing.B) {
	runIsTokenMintBench(b, isTokenMintPR396, []*Account{
		{Owner: solana.TokenProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 82))},
		{Owner: solana.TokenProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 82))},
	})
}

func BenchmarkIsTokenMint_PR396_WrongOwner(b *testing.B) {
	runIsTokenMintBench(b, isTokenMintPR396, []*Account{
		{Owner: solana.SystemProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 200))},
		{Owner: solana.SystemProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 500))},
	})
}

// --- Current (length-first) version ---

func BenchmarkIsTokenMint_Current_Mixed(b *testing.B) {
	runIsTokenMintBench(b, IsTokenMint, buildMixedAccounts())
}

func BenchmarkIsTokenMint_Current_HotMint(b *testing.B) {
	runIsTokenMintBench(b, IsTokenMint, []*Account{
		{Owner: solana.TokenProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 82))},
		{Owner: solana.TokenProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 82))},
	})
}

func BenchmarkIsTokenMint_Current_WrongOwner(b *testing.B) {
	runIsTokenMintBench(b, IsTokenMint, []*Account{
		{Owner: solana.SystemProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 200))},
		{Owner: solana.SystemProgramID, Data: DataBytesOrJSONFromBytes(make([]byte, 500))},
	})
}
