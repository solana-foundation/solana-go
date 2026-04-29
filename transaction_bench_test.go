package solana

import (
	"testing"
)

// buildBenchInstructions creates a realistic set of instructions for benchmarking
// NewTransaction. Each instruction references a distinct program ID and mixes
// signer/writable/readonly accounts so the builder exercises all header-counting
// and indexing paths.
//
// numInstructions: how many instructions in the transaction.
// accountsPerIx:   how many AccountMeta entries each instruction references.
//
//	(first is always a signer; second is always writable; rest readonly)
func buildBenchInstructions(numInstructions, accountsPerIx int) ([]Instruction, Hash) {
	// Pre-generate a pool of unique accounts so instructions share some accounts
	// (realistic — the same fee payer / writable state account appears in many ixs)
	// while still producing a meaningful total.
	poolSize := numInstructions * accountsPerIx / 2
	if poolSize < accountsPerIx {
		poolSize = accountsPerIx
	}
	pool := make([]PublicKey, poolSize)
	for i := range pool {
		pool[i] = newUniqueKey()
	}

	feePayer := pool[0]

	instructions := make([]Instruction, numInstructions)
	for i := 0; i < numInstructions; i++ {
		accounts := make(AccountMetaSlice, accountsPerIx)
		for j := 0; j < accountsPerIx; j++ {
			pk := pool[(i*accountsPerIx+j)%len(pool)]
			var isSigner, isWritable bool
			switch j {
			case 0:
				// fee payer — signer, writable
				pk = feePayer
				isSigner = true
				isWritable = true
			case 1:
				isWritable = true
			}
			accounts[j] = &AccountMeta{
				PublicKey:  pk,
				IsSigner:   isSigner,
				IsWritable: isWritable,
			}
		}
		instructions[i] = NewInstruction(
			newUniqueKey(), // distinct program ID per instruction
			accounts,
			[]byte{byte(i), 0xAA, 0xBB, 0xCC},
		)
	}

	return instructions, Hash{1, 2, 3, 4, 5}
}

// buildBenchInstructionsWithLookups is like buildBenchInstructions but also
// prepares an address-lookup-table so most of the accounts get compiled into
// ATL lookups instead of the static key list. This stresses the path where
// lookupsWritableKeys / lookupsReadOnlyKeys are non-empty.
func buildBenchInstructionsWithLookups(numInstructions, accountsPerIx int) ([]Instruction, Hash, map[PublicKey]PublicKeySlice) {
	instructions, blockhash := buildBenchInstructions(numInstructions, accountsPerIx)

	// Collect all non-signer, non-program accounts into a single ATL.
	seen := make(map[PublicKey]struct{})
	var tableAccounts PublicKeySlice
	for _, ix := range instructions {
		for _, am := range ix.Accounts() {
			if am.IsSigner {
				continue
			}
			if _, ok := seen[am.PublicKey]; ok {
				continue
			}
			seen[am.PublicKey] = struct{}{}
			tableAccounts = append(tableAccounts, am.PublicKey)
			if len(tableAccounts) == 256 {
				break
			}
		}
		if len(tableAccounts) == 256 {
			break
		}
	}

	tableKey := newUniqueKey()
	tables := map[PublicKey]PublicKeySlice{
		tableKey: tableAccounts,
	}
	return instructions, blockhash, tables
}

// Realistic transaction shapes for solana-go benchmarks.
// Upper bound is ~10 instructions / ~30 accounts per ix — beyond that
// becomes a synthetic stress shape that doesn't represent real traffic.
var benchTxShapes = []struct {
	name            string
	numInstructions int
	accountsPerIx   int
}{
	{"small_2ix_5accts", 2, 5},
	{"medium_5ix_15accts", 5, 15},
	{"large_10ix_30accts", 10, 30},
}

func BenchmarkNewTransaction(b *testing.B) {
	for _, tc := range benchTxShapes {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			instructions, blockhash := buildBenchInstructions(tc.numInstructions, tc.accountsPerIx)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tx, err := NewTransaction(instructions, blockhash)
				if err != nil {
					b.Fatal(err)
				}
				_ = tx
			}
		})
	}
}

func BenchmarkNewTransaction_WithLookupTable(b *testing.B) {
	for _, tc := range benchTxShapes {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			instructions, blockhash, tables := buildBenchInstructionsWithLookups(tc.numInstructions, tc.accountsPerIx)
			opt := TransactionAddressTables(tables)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tx, err := NewTransaction(instructions, blockhash, opt)
				if err != nil {
					b.Fatal(err)
				}
				_ = tx
			}
		})
	}
}

func BenchmarkTransaction_NumWriteableAccounts(b *testing.B) {
	// Legacy (non-versioned) takes the AccountMetaList path.
	b.Run("legacy_large", func(b *testing.B) {
		instructions, blockhash := buildBenchInstructions(10, 30)
		tx, err := NewTransaction(instructions, blockhash)
		if err != nil {
			b.Fatal(err)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.NumWriteableAccounts()
		}
	})

	// V0 with ATL takes the static-key-scan path (the O(n^2) candidate).
	b.Run("v0_large", func(b *testing.B) {
		instructions, blockhash, tables := buildBenchInstructionsWithLookups(10, 30)
		tx, err := NewTransaction(instructions, blockhash, TransactionAddressTables(tables))
		if err != nil {
			b.Fatal(err)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.NumWriteableAccounts()
		}
	})
}
