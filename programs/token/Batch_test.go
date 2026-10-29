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
	"testing"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_require "github.com/stretchr/testify/require"
)

// fixtureSubInstruction is a test-only sub-instruction Impl that lets us control
// exactly how many accounts and data bytes a batch sub-instruction encodes to,
// so we can exercise the per-sub-instruction 255 account / 255 byte limits.
type fixtureSubInstruction struct {
	accounts []*ag_solanago.AccountMeta
	data     []byte
}

func (f fixtureSubInstruction) GetAccounts() []*ag_solanago.AccountMeta { return f.accounts }

func (f fixtureSubInstruction) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	_, err := encoder.Write(f.data)
	return err
}

func newFixtureInstruction(numAccounts, numDataBytes int) *Instruction {
	accounts := make([]*ag_solanago.AccountMeta, numAccounts)
	for i := range accounts {
		accounts[i] = ag_solanago.Meta(ag_solanago.NewWallet().PublicKey())
	}
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   fixtureSubInstruction{accounts: accounts, data: make([]byte, numDataBytes)},
		TypeID: ag_binary.TypeIDFromUint8(Instruction_Transfer),
	}}
}

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

func TestBatch_MarshalWithEncoder_Limits(t *testing.T) {
	t.Run("too many accounts returns error instead of truncating", func(t *testing.T) {
		batch := NewBatchInstruction(newFixtureInstruction(256, 1))
		_, err := batch.Build().Data()
		ag_require.Error(t, err)
		ag_require.Contains(t, err.Error(), "256 accounts")
	})

	t.Run("too much data returns error instead of truncating", func(t *testing.T) {
		// Instruction.Data() prepends a 1-byte discriminator, so 255 fixture
		// bytes become 256 bytes of sub-instruction data.
		batch := NewBatchInstruction(newFixtureInstruction(1, 255))
		_, err := batch.Build().Data()
		ag_require.Error(t, err)
		ag_require.Contains(t, err.Error(), "256 bytes")
	})

	t.Run("at the 255 limit encodes successfully", func(t *testing.T) {
		// 255 accounts is allowed; data is 1 discriminator byte + 254 = 255 bytes.
		batch := NewBatchInstruction(newFixtureInstruction(255, 254))
		data, err := batch.Build().Data()
		ag_require.NoError(t, err)
		ag_require.NotEmpty(t, data)
	})
}

func TestBatch_DecodeDistributesAccountsToSubInstructions(t *testing.T) {
	src := ag_solanago.NewWallet().PublicKey()
	dst := ag_solanago.NewWallet().PublicKey()
	owner := ag_solanago.NewWallet().PublicKey()

	batch, err := NewBatchInstruction(
		NewTransferInstruction(100, src, dst, owner, nil).Build(),
		NewTransferInstruction(200, dst, src, owner, nil).Build(),
	).ValidateAndBuild()
	ag_require.NoError(t, err)

	data, err := batch.Data()
	ag_require.NoError(t, err)

	decoded, err := DecodeInstruction(batch.Accounts(), data)
	ag_require.NoError(t, err)

	got, ok := decoded.Impl.(*Batch)
	ag_require.True(t, ok)
	ag_require.Len(t, got.Instructions, 2)

	// Each sub-instruction must get its own slice of the flat account list
	// back, not an empty one.
	first, ok := got.Instructions[0].Impl.(*Transfer)
	ag_require.True(t, ok)
	ag_require.Equal(t, uint64(100), *first.Amount)
	ag_require.Equal(t, src, first.GetSourceAccount().PublicKey)
	ag_require.Equal(t, dst, first.GetDestinationAccount().PublicKey)
	ag_require.Equal(t, owner, first.GetOwnerAccount().PublicKey)

	second, ok := got.Instructions[1].Impl.(*Transfer)
	ag_require.True(t, ok)
	ag_require.Equal(t, uint64(200), *second.Amount)
	ag_require.Equal(t, dst, second.GetSourceAccount().PublicKey)
	ag_require.Equal(t, src, second.GetDestinationAccount().PublicKey)
	ag_require.Equal(t, owner, second.GetOwnerAccount().PublicKey)
}

func TestBatch_RejectsNestedBatch(t *testing.T) {
	src := ag_solanago.NewWallet().PublicKey()
	dst := ag_solanago.NewWallet().PublicKey()
	owner := ag_solanago.NewWallet().PublicKey()

	inner, err := NewBatchInstruction(
		NewTransferInstruction(1, src, dst, owner, nil).Build(),
	).ValidateAndBuild()
	ag_require.NoError(t, err)

	_, err = NewBatchInstruction(inner).ValidateAndBuild()
	ag_require.Error(t, err)
	ag_require.Contains(t, err.Error(), "nested batches are not supported")
}

func TestBatch_RejectsEmptySubInstructionData(t *testing.T) {
	// header says 1 account, 0 data bytes — the program rejects this.
	_, err := DecodeInstruction(nil, []byte{Instruction_Batch, 0x01, 0x00})
	ag_require.Error(t, err)
	ag_require.Contains(t, err.Error(), "no instruction data")
}

func TestBatch_DecodeRejectsTruncatedAccountList(t *testing.T) {
	src := ag_solanago.NewWallet().PublicKey()
	dst := ag_solanago.NewWallet().PublicKey()
	owner := ag_solanago.NewWallet().PublicKey()

	batch, err := NewBatchInstruction(
		NewTransferInstruction(100, src, dst, owner, nil).Build(),
	).ValidateAndBuild()
	ag_require.NoError(t, err)
	data, err := batch.Data()
	ag_require.NoError(t, err)

	// The header claims 3 accounts, but only 2 are supplied.
	_, err = DecodeInstruction(batch.Accounts()[:2], data)
	ag_require.Error(t, err)
	ag_require.Contains(t, err.Error(), "expects 3 accounts")
}
