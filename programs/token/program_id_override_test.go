// Copyright 2026 github.com/gagliardetto
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

package token

import (
	"sync"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// TestInstruction_SetProgramID_DefaultMatchesPackage ensures that an
// instruction built without an override falls back to the package-level
// ProgramID, preserving prior behavior for the default code path.
func TestInstruction_SetProgramID_DefaultMatchesPackage(t *testing.T) {
	inst := NewTransferInstructionBuilder().
		SetAmount(1).
		SetSourceAccount(solana.MustPublicKeyFromBase58("11111111111111111111111111111112")).
		SetDestinationAccount(solana.MustPublicKeyFromBase58("11111111111111111111111111111113")).
		SetOwnerAccount(solana.MustPublicKeyFromBase58("11111111111111111111111111111114")).
		Build()

	require.Equal(t, ProgramID, inst.ProgramID())
}

// TestInstruction_SetProgramID_OverrideIsPerInstance verifies that
// SetProgramID only affects the receiver and does not leak into either
// the package-level ProgramID or a sibling instruction built off the
// same builder package.
func TestInstruction_SetProgramID_OverrideIsPerInstance(t *testing.T) {
	originalProgramID := ProgramID
	t.Cleanup(func() { ProgramID = originalProgramID })

	src := solana.MustPublicKeyFromBase58("11111111111111111111111111111112")
	dst := solana.MustPublicKeyFromBase58("11111111111111111111111111111113")
	owner := solana.MustPublicKeyFromBase58("11111111111111111111111111111114")

	defaultInst := NewTransferInstructionBuilder().
		SetAmount(1).
		SetSourceAccount(src).
		SetDestinationAccount(dst).
		SetOwnerAccount(owner).
		Build()

	overriddenInst := NewTransferInstructionBuilder().
		SetAmount(2).
		SetSourceAccount(src).
		SetDestinationAccount(dst).
		SetOwnerAccount(owner).
		Build().
		SetProgramID(solana.Token2022ProgramID)

	require.Equal(t, ProgramID, defaultInst.ProgramID(),
		"default instruction must read the package ProgramID")
	require.Equal(t, solana.Token2022ProgramID, overriddenInst.ProgramID(),
		"overridden instruction must read its own ProgramID")
	require.Equal(t, originalProgramID, ProgramID,
		"SetProgramID on an instruction must not mutate the package ProgramID")
}

// TestInstruction_SetProgramID_ConcurrentBuildersDoNotRace covers the
// motivating case from issue #254: building SPL Token and SPL Token-2022
// instructions in parallel must not require a process-wide mutex around
// the package-level ProgramID. Run under `-race`.
func TestInstruction_SetProgramID_ConcurrentBuildersDoNotRace(t *testing.T) {
	originalProgramID := ProgramID
	t.Cleanup(func() { ProgramID = originalProgramID })

	src := solana.MustPublicKeyFromBase58("11111111111111111111111111111112")
	dst := solana.MustPublicKeyFromBase58("11111111111111111111111111111113")
	owner := solana.MustPublicKeyFromBase58("11111111111111111111111111111114")

	const goroutines = 32
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()

			// Alternate between the default program ID and the Token-2022
			// override so both code paths are exercised under -race.
			inst := NewTransferInstructionBuilder().
				SetAmount(uint64(i + 1)).
				SetSourceAccount(src).
				SetDestinationAccount(dst).
				SetOwnerAccount(owner).
				Build()

			if i%2 == 0 {
				inst.SetProgramID(solana.Token2022ProgramID)
				if inst.ProgramID() != solana.Token2022ProgramID {
					t.Errorf("goroutine %d: expected Token2022ProgramID, got %s", i, inst.ProgramID())
				}
			} else {
				if inst.ProgramID() != ProgramID {
					t.Errorf("goroutine %d: expected package ProgramID, got %s", i, inst.ProgramID())
				}
			}
		}()
	}
	wg.Wait()

	require.Equal(t, originalProgramID, ProgramID,
		"package ProgramID must be unchanged after concurrent builders")
}
