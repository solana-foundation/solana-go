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

package token

import (
	ag_require "github.com/stretchr/testify/require"
	"testing"
)

func TestEncodeDecode_Batch(t *testing.T) {
	t.Run("Batch_InstructionIDToName", func(t *testing.T) {
		ag_require.Equal(t, "Batch", InstructionIDToName(Instruction_Batch))
		ag_require.Equal(t, "WithdrawExcessLamports", InstructionIDToName(Instruction_WithdrawExcessLamports))
		ag_require.Equal(t, "UnwrapLamports", InstructionIDToName(Instruction_UnwrapLamports))
	})

	t.Run("Batch_InstructionIDs", func(t *testing.T) {
		ag_require.Equal(t, uint8(21), Instruction_GetAccountDataSize)
		ag_require.Equal(t, uint8(22), Instruction_InitializeImmutableOwner)
		ag_require.Equal(t, uint8(23), Instruction_AmountToUiAmount)
		ag_require.Equal(t, uint8(24), Instruction_UiAmountToAmount)
		ag_require.Equal(t, uint8(38), Instruction_WithdrawExcessLamports)
		ag_require.Equal(t, uint8(45), Instruction_UnwrapLamports)
		ag_require.Equal(t, uint8(255), Instruction_Batch)
	})
}
