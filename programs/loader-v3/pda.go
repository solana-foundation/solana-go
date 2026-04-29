// Copyright 2021 github.com/gagliardetto
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

package loaderv3

import (
	ag_solanago "github.com/gagliardetto/solana-go"
)

// GetProgramDataAddress derives the PDA that stores a program's data for a
// given program account, matching the upstream helper:
//
//	Pubkey::find_program_address(&[program.as_ref()], &bpf_loader_upgradeable::id())
func GetProgramDataAddress(programAddress ag_solanago.PublicKey) (ag_solanago.PublicKey, uint8, error) {
	return ag_solanago.FindProgramAddress(
		[][]byte{programAddress[:]},
		ProgramID,
	)
}

// MustGetProgramDataAddress panics on error. Convenience for callers that
// trust the inputs (e.g. tests, hard-coded program ids).
func MustGetProgramDataAddress(programAddress ag_solanago.PublicKey) ag_solanago.PublicKey {
	addr, _, err := GetProgramDataAddress(programAddress)
	if err != nil {
		panic(err)
	}
	return addr
}
