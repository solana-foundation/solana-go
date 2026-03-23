package token

import (
	"bytes"
	"encoding/binary"
	"testing"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_require "github.com/stretchr/testify/require"
)

func TestInstructionIDs_MatchOnChain(t *testing.T) {
	ag_require.Equal(t, uint8(0), Instruction_InitializeMint)
	ag_require.Equal(t, uint8(1), Instruction_InitializeAccount)
	ag_require.Equal(t, uint8(2), Instruction_InitializeMultisig)
	ag_require.Equal(t, uint8(3), Instruction_Transfer)
	ag_require.Equal(t, uint8(4), Instruction_Approve)
	ag_require.Equal(t, uint8(5), Instruction_Revoke)
	ag_require.Equal(t, uint8(6), Instruction_SetAuthority)
	ag_require.Equal(t, uint8(7), Instruction_MintTo)
	ag_require.Equal(t, uint8(8), Instruction_Burn)
	ag_require.Equal(t, uint8(9), Instruction_CloseAccount)
	ag_require.Equal(t, uint8(10), Instruction_FreezeAccount)
	ag_require.Equal(t, uint8(11), Instruction_ThawAccount)
	ag_require.Equal(t, uint8(12), Instruction_TransferChecked)
	ag_require.Equal(t, uint8(13), Instruction_ApproveChecked)
	ag_require.Equal(t, uint8(14), Instruction_MintToChecked)
	ag_require.Equal(t, uint8(15), Instruction_BurnChecked)
	ag_require.Equal(t, uint8(16), Instruction_InitializeAccount2)
	ag_require.Equal(t, uint8(17), Instruction_SyncNative)
	ag_require.Equal(t, uint8(18), Instruction_InitializeAccount3)
	ag_require.Equal(t, uint8(19), Instruction_InitializeMultisig2)
	ag_require.Equal(t, uint8(20), Instruction_InitializeMint2)
	ag_require.Equal(t, uint8(21), Instruction_GetAccountDataSize)
	ag_require.Equal(t, uint8(22), Instruction_InitializeImmutableOwner)
	ag_require.Equal(t, uint8(23), Instruction_AmountToUiAmount)
	ag_require.Equal(t, uint8(24), Instruction_UiAmountToAmount)
	ag_require.Equal(t, uint8(38), Instruction_WithdrawExcessLamports)
	ag_require.Equal(t, uint8(45), Instruction_UnwrapLamports)
	ag_require.Equal(t, uint8(255), Instruction_Batch)
}

func TestGetAccountDataSize_ByteFormat(t *testing.T) {
	mint := ag_solanago.NewWallet().PublicKey()
	ix := NewGetAccountDataSizeInstruction(mint)
	built := ix.Build()

	data, err := built.Data()
	ag_require.NoError(t, err)
	ag_require.Equal(t, []byte{21}, data)

	accounts := built.Accounts()
	ag_require.Len(t, accounts, 1)
	ag_require.Equal(t, mint, accounts[0].PublicKey)
	ag_require.False(t, accounts[0].IsWritable)
	ag_require.False(t, accounts[0].IsSigner)
}

func TestInitializeImmutableOwner_ByteFormat(t *testing.T) {
	account := ag_solanago.NewWallet().PublicKey()
	ix := NewInitializeImmutableOwnerInstruction(account)
	built := ix.Build()

	data, err := built.Data()
	ag_require.NoError(t, err)
	ag_require.Equal(t, []byte{22}, data)

	accounts := built.Accounts()
	ag_require.Len(t, accounts, 1)
	ag_require.Equal(t, account, accounts[0].PublicKey)
	ag_require.True(t, accounts[0].IsWritable)
}

func TestAmountToUiAmount_ByteFormat(t *testing.T) {
	mint := ag_solanago.NewWallet().PublicKey()
	amount := uint64(1_000_000_000)
	ix := NewAmountToUiAmountInstruction(amount, mint)
	built := ix.Build()

	data, err := built.Data()
	ag_require.NoError(t, err)

	ag_require.Equal(t, byte(23), data[0])
	ag_require.Len(t, data, 9)
	gotAmount := binary.LittleEndian.Uint64(data[1:9])
	ag_require.Equal(t, amount, gotAmount)
}

func TestUiAmountToAmount_ByteFormat(t *testing.T) {
	mint := ag_solanago.NewWallet().PublicKey()
	uiAmount := "1.5"
	ix := NewUiAmountToAmountInstruction(uiAmount, mint)
	built := ix.Build()

	data, err := built.Data()
	ag_require.NoError(t, err)

	ag_require.Equal(t, byte(24), data[0])
	ag_require.Equal(t, "1.5", string(data[1:]))
}

func TestWithdrawExcessLamports_ByteFormat(t *testing.T) {
	source := ag_solanago.NewWallet().PublicKey()
	dest := ag_solanago.NewWallet().PublicKey()
	authority := ag_solanago.NewWallet().PublicKey()

	ix := NewWithdrawExcessLamportsInstruction(source, dest, authority, nil)
	built := ix.Build()

	data, err := built.Data()
	ag_require.NoError(t, err)
	ag_require.Equal(t, []byte{38}, data)

	accounts := built.Accounts()
	ag_require.Len(t, accounts, 3)
	ag_require.True(t, accounts[0].IsWritable)
	ag_require.True(t, accounts[1].IsWritable)
	ag_require.True(t, accounts[2].IsSigner)
}

func TestWithdrawExcessLamports_Multisig(t *testing.T) {
	source := ag_solanago.NewWallet().PublicKey()
	dest := ag_solanago.NewWallet().PublicKey()
	authority := ag_solanago.NewWallet().PublicKey()
	signer1 := ag_solanago.NewWallet().PublicKey()
	signer2 := ag_solanago.NewWallet().PublicKey()

	ix := NewWithdrawExcessLamportsInstruction(source, dest, authority, []ag_solanago.PublicKey{signer1, signer2})
	built := ix.Build()

	accounts := built.Accounts()
	ag_require.Len(t, accounts, 5)
	ag_require.False(t, accounts[2].IsSigner)
	ag_require.True(t, accounts[3].IsSigner)
	ag_require.True(t, accounts[4].IsSigner)
}

func TestUnwrapLamports_ByteFormat_NoAmount(t *testing.T) {
	source := ag_solanago.NewWallet().PublicKey()
	dest := ag_solanago.NewWallet().PublicKey()
	authority := ag_solanago.NewWallet().PublicKey()

	ix := NewUnwrapLamportsInstruction(source, dest, authority, nil)
	built := ix.Build()

	data, err := built.Data()
	ag_require.NoError(t, err)
	ag_require.Equal(t, []byte{45, 0}, data)
}

func TestUnwrapLamports_ByteFormat_WithAmount(t *testing.T) {
	source := ag_solanago.NewWallet().PublicKey()
	dest := ag_solanago.NewWallet().PublicKey()
	authority := ag_solanago.NewWallet().PublicKey()
	amount := uint64(500_000)

	ix := NewUnwrapLamportsWithAmountInstruction(amount, source, dest, authority, nil)
	built := ix.Build()

	data, err := built.Data()
	ag_require.NoError(t, err)

	ag_require.Equal(t, byte(45), data[0])
	ag_require.Equal(t, byte(1), data[1])
	ag_require.Len(t, data, 10)
	gotAmount := binary.LittleEndian.Uint64(data[2:10])
	ag_require.Equal(t, amount, gotAmount)
}

func TestDecodeInstruction_PTokenInstructions(t *testing.T) {
	t.Run("DecodeWithdrawExcessLamports", func(t *testing.T) {
		source := ag_solanago.NewWallet().PublicKey()
		dest := ag_solanago.NewWallet().PublicKey()
		authority := ag_solanago.NewWallet().PublicKey()

		accounts := []*ag_solanago.AccountMeta{
			ag_solanago.Meta(source).WRITE(),
			ag_solanago.Meta(dest).WRITE(),
			ag_solanago.Meta(authority).SIGNER(),
		}

		data := []byte{38}
		inst, err := DecodeInstruction(accounts, data)
		ag_require.NoError(t, err)
		ag_require.Equal(t, uint8(38), inst.TypeID.Uint8())

		_, ok := inst.Impl.(*WithdrawExcessLamports)
		ag_require.True(t, ok)
	})

	t.Run("DecodeUnwrapLamports_NoAmount", func(t *testing.T) {
		accounts := []*ag_solanago.AccountMeta{
			ag_solanago.Meta(ag_solanago.NewWallet().PublicKey()).WRITE(),
			ag_solanago.Meta(ag_solanago.NewWallet().PublicKey()).WRITE(),
			ag_solanago.Meta(ag_solanago.NewWallet().PublicKey()).SIGNER(),
		}

		data := []byte{45, 0}
		inst, err := DecodeInstruction(accounts, data)
		ag_require.NoError(t, err)
		ag_require.Equal(t, uint8(45), inst.TypeID.Uint8())

		unwrap, ok := inst.Impl.(*UnwrapLamports)
		ag_require.True(t, ok)
		ag_require.Nil(t, unwrap.Amount)
	})

	t.Run("DecodeUnwrapLamports_WithAmount", func(t *testing.T) {
		accounts := []*ag_solanago.AccountMeta{
			ag_solanago.Meta(ag_solanago.NewWallet().PublicKey()).WRITE(),
			ag_solanago.Meta(ag_solanago.NewWallet().PublicKey()).WRITE(),
			ag_solanago.Meta(ag_solanago.NewWallet().PublicKey()).SIGNER(),
		}

		amount := uint64(999)
		buf := new(bytes.Buffer)
		buf.WriteByte(45)
		buf.WriteByte(1)
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, amount)
		buf.Write(b)

		inst, err := DecodeInstruction(accounts, buf.Bytes())
		ag_require.NoError(t, err)

		unwrap, ok := inst.Impl.(*UnwrapLamports)
		ag_require.True(t, ok)
		ag_require.NotNil(t, unwrap.Amount)
		ag_require.Equal(t, amount, *unwrap.Amount)
	})
}

func TestDecodeInstruction_LegacyInstructions(t *testing.T) {
	t.Run("DecodeGetAccountDataSize", func(t *testing.T) {
		mint := ag_solanago.NewWallet().PublicKey()
		accounts := []*ag_solanago.AccountMeta{
			ag_solanago.Meta(mint),
		}
		data := []byte{21}
		inst, err := DecodeInstruction(accounts, data)
		ag_require.NoError(t, err)
		ag_require.Equal(t, uint8(21), inst.TypeID.Uint8())
		_, ok := inst.Impl.(*GetAccountDataSize)
		ag_require.True(t, ok)
	})

	t.Run("DecodeInitializeImmutableOwner", func(t *testing.T) {
		account := ag_solanago.NewWallet().PublicKey()
		accounts := []*ag_solanago.AccountMeta{
			ag_solanago.Meta(account).WRITE(),
		}
		data := []byte{22}
		inst, err := DecodeInstruction(accounts, data)
		ag_require.NoError(t, err)
		ag_require.Equal(t, uint8(22), inst.TypeID.Uint8())
		_, ok := inst.Impl.(*InitializeImmutableOwner)
		ag_require.True(t, ok)
	})

	t.Run("DecodeAmountToUiAmount", func(t *testing.T) {
		mint := ag_solanago.NewWallet().PublicKey()
		accounts := []*ag_solanago.AccountMeta{
			ag_solanago.Meta(mint),
		}
		amount := uint64(1_000_000)
		buf := new(bytes.Buffer)
		buf.WriteByte(23)
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, amount)
		buf.Write(b)

		inst, err := DecodeInstruction(accounts, buf.Bytes())
		ag_require.NoError(t, err)
		ag_require.Equal(t, uint8(23), inst.TypeID.Uint8())
		a2u, ok := inst.Impl.(*AmountToUiAmount)
		ag_require.True(t, ok)
		ag_require.Equal(t, amount, *a2u.Amount)
	})

	t.Run("DecodeUiAmountToAmount", func(t *testing.T) {
		mint := ag_solanago.NewWallet().PublicKey()
		accounts := []*ag_solanago.AccountMeta{
			ag_solanago.Meta(mint),
		}
		buf := new(bytes.Buffer)
		buf.WriteByte(24)
		buf.WriteString("1.5")

		inst, err := DecodeInstruction(accounts, buf.Bytes())
		ag_require.NoError(t, err)
		ag_require.Equal(t, uint8(24), inst.TypeID.Uint8())
		u2a, ok := inst.Impl.(*UiAmountToAmount)
		ag_require.True(t, ok)
		ag_require.Equal(t, "1.5", *u2a.UiAmount)
	})
}

func TestBuildAndEncode_RoundTrip(t *testing.T) {
	t.Run("Transfer_RoundTrip", func(t *testing.T) {
		src := ag_solanago.NewWallet().PublicKey()
		dst := ag_solanago.NewWallet().PublicKey()
		owner := ag_solanago.NewWallet().PublicKey()

		ix := NewTransferInstruction(100, src, dst, owner, nil)
		built := ix.Build()

		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(built.Accounts(), data)
		ag_require.NoError(t, err)

		transfer, ok := decoded.Impl.(*Transfer)
		ag_require.True(t, ok)
		ag_require.Equal(t, uint64(100), *transfer.Amount)
	})

	t.Run("AmountToUiAmount_RoundTrip", func(t *testing.T) {
		mint := ag_solanago.NewWallet().PublicKey()
		ix := NewAmountToUiAmountInstruction(999, mint)
		built := ix.Build()

		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(built.Accounts(), data)
		ag_require.NoError(t, err)

		a2u, ok := decoded.Impl.(*AmountToUiAmount)
		ag_require.True(t, ok)
		ag_require.Equal(t, uint64(999), *a2u.Amount)
	})

	t.Run("WithdrawExcessLamports_RoundTrip", func(t *testing.T) {
		source := ag_solanago.NewWallet().PublicKey()
		dest := ag_solanago.NewWallet().PublicKey()
		auth := ag_solanago.NewWallet().PublicKey()

		ix := NewWithdrawExcessLamportsInstruction(source, dest, auth, nil)
		built := ix.Build()

		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(built.Accounts(), data)
		ag_require.NoError(t, err)

		_, ok := decoded.Impl.(*WithdrawExcessLamports)
		ag_require.True(t, ok)
		ag_require.Equal(t, uint8(38), decoded.TypeID.Uint8())
	})

	t.Run("UnwrapLamports_RoundTrip_WithAmount", func(t *testing.T) {
		source := ag_solanago.NewWallet().PublicKey()
		dest := ag_solanago.NewWallet().PublicKey()
		auth := ag_solanago.NewWallet().PublicKey()

		ix := NewUnwrapLamportsWithAmountInstruction(12345, source, dest, auth, nil)
		built := ix.Build()

		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(built.Accounts(), data)
		ag_require.NoError(t, err)

		unwrap, ok := decoded.Impl.(*UnwrapLamports)
		ag_require.True(t, ok)
		ag_require.NotNil(t, unwrap.Amount)
		ag_require.Equal(t, uint64(12345), *unwrap.Amount)
	})
}

func TestInstructionIDToName_AllNew(t *testing.T) {
	tests := []struct {
		id   uint8
		name string
	}{
		{21, "GetAccountDataSize"},
		{22, "InitializeImmutableOwner"},
		{23, "AmountToUiAmount"},
		{24, "UiAmountToAmount"},
		{38, "WithdrawExcessLamports"},
		{45, "UnwrapLamports"},
		{255, "Batch"},
	}
	for _, tt := range tests {
		ag_require.Equal(t, tt.name, InstructionIDToName(tt.id))
	}
}

func TestValidation(t *testing.T) {
	t.Run("GetAccountDataSize_MissingMint", func(t *testing.T) {
		ix := NewGetAccountDataSizeInstructionBuilder()
		err := ix.Validate()
		ag_require.Error(t, err)
	})

	t.Run("AmountToUiAmount_MissingAmount", func(t *testing.T) {
		ix := NewAmountToUiAmountInstructionBuilder()
		ix.SetMintAccount(ag_solanago.NewWallet().PublicKey())
		err := ix.Validate()
		ag_require.Error(t, err)
	})

	t.Run("WithdrawExcessLamports_MissingAccounts", func(t *testing.T) {
		ix := NewWithdrawExcessLamportsInstructionBuilder()
		err := ix.Validate()
		ag_require.Error(t, err)
	})

	t.Run("UnwrapLamports_MissingAccounts", func(t *testing.T) {
		ix := NewUnwrapLamportsInstructionBuilder()
		err := ix.Validate()
		ag_require.Error(t, err)
	})

	t.Run("Batch_Empty", func(t *testing.T) {
		ix := NewBatchInstructionBuilder()
		err := ix.Validate()
		ag_require.Error(t, err)
	})

	t.Run("WithdrawExcessLamports_TooManySigners", func(t *testing.T) {
		ix := NewWithdrawExcessLamportsInstructionBuilder().
			SetSourceAccount(ag_solanago.NewWallet().PublicKey()).
			SetDestinationAccount(ag_solanago.NewWallet().PublicKey())

		signers := make([]ag_solanago.PublicKey, 12)
		for i := range signers {
			signers[i] = ag_solanago.NewWallet().PublicKey()
		}
		ix.SetAuthorityAccount(ag_solanago.NewWallet().PublicKey(), signers...)
		err := ix.Validate()
		ag_require.Error(t, err)
		ag_require.Contains(t, err.Error(), "too many signers")
	})
}

func TestMarshalWithEncoder_PTokenInstructions(t *testing.T) {
	t.Run("WithdrawExcessLamports_MarshalEncode", func(t *testing.T) {
		obj := WithdrawExcessLamports{
			Accounts: make(ag_solanago.AccountMetaSlice, 3),
			Signers:  make(ag_solanago.AccountMetaSlice, 0),
		}
		buf := new(bytes.Buffer)
		enc := ag_binary.NewBinEncoder(buf)
		err := obj.MarshalWithEncoder(enc)
		ag_require.NoError(t, err)
		ag_require.Equal(t, 0, buf.Len())
	})

	t.Run("UnwrapLamports_MarshalEncode_Nil", func(t *testing.T) {
		obj := UnwrapLamports{
			Amount:   nil,
			Accounts: make(ag_solanago.AccountMetaSlice, 3),
			Signers:  make(ag_solanago.AccountMetaSlice, 0),
		}
		buf := new(bytes.Buffer)
		enc := ag_binary.NewBinEncoder(buf)
		err := obj.MarshalWithEncoder(enc)
		ag_require.NoError(t, err)
		ag_require.Equal(t, []byte{0}, buf.Bytes())
	})

	t.Run("UnwrapLamports_MarshalEncode_WithAmount", func(t *testing.T) {
		amount := uint64(42)
		obj := UnwrapLamports{
			Amount:   &amount,
			Accounts: make(ag_solanago.AccountMetaSlice, 3),
			Signers:  make(ag_solanago.AccountMetaSlice, 0),
		}
		buf := new(bytes.Buffer)
		enc := ag_binary.NewBinEncoder(buf)
		err := obj.MarshalWithEncoder(enc)
		ag_require.NoError(t, err)

		expected := make([]byte, 9)
		expected[0] = 1
		binary.LittleEndian.PutUint64(expected[1:], 42)
		ag_require.Equal(t, expected, buf.Bytes())
	})
}

func TestBatch_ByteFormat(t *testing.T) {
	src := ag_solanago.NewWallet().PublicKey()
	dst := ag_solanago.NewWallet().PublicKey()
	owner := ag_solanago.NewWallet().PublicKey()

	transfer1 := NewTransferInstruction(100, src, dst, owner, nil).Build()
	transfer2 := NewTransferInstruction(200, src, dst, owner, nil).Build()

	batch := NewBatchInstruction(transfer1, transfer2)
	built := batch.Build()

	data, err := built.Data()
	ag_require.NoError(t, err)

	// byte 0: discriminator 255
	ag_require.Equal(t, byte(255), data[0])

	// Sub-instruction 1 header
	// transfer has 3 accounts, data is [3 (discriminator) + 8 bytes (u64 amount)] = 9 bytes
	ag_require.Equal(t, byte(3), data[1])  // num_accounts
	ag_require.Equal(t, byte(9), data[2])  // data_len
	ag_require.Equal(t, byte(3), data[3])  // Transfer discriminator (ID=3)
	amount1 := binary.LittleEndian.Uint64(data[4:12])
	ag_require.Equal(t, uint64(100), amount1)

	// Sub-instruction 2 header
	ag_require.Equal(t, byte(3), data[12]) // num_accounts
	ag_require.Equal(t, byte(9), data[13]) // data_len
	ag_require.Equal(t, byte(3), data[14]) // Transfer discriminator
	amount2 := binary.LittleEndian.Uint64(data[15:23])
	ag_require.Equal(t, uint64(200), amount2)

	ag_require.Len(t, data, 23)
}

func TestBatch_Accounts(t *testing.T) {
	src := ag_solanago.NewWallet().PublicKey()
	dst := ag_solanago.NewWallet().PublicKey()
	owner := ag_solanago.NewWallet().PublicKey()
	mint := ag_solanago.NewWallet().PublicKey()

	transfer := NewTransferInstruction(100, src, dst, owner, nil).Build()
	getSize := NewGetAccountDataSizeInstruction(mint).Build()

	batch := NewBatchInstruction(transfer, getSize)
	built := batch.Build()

	accounts := built.Accounts()
	ag_require.Len(t, accounts, 4) // 3 from transfer + 1 from getSize
}

func TestBatch_BuildBatchData(t *testing.T) {
	src := ag_solanago.NewWallet().PublicKey()
	dst := ag_solanago.NewWallet().PublicKey()
	owner := ag_solanago.NewWallet().PublicKey()

	transfer := NewTransferInstruction(50, src, dst, owner, nil).Build()

	batchData, err := BuildBatchData([]*Instruction{transfer})
	ag_require.NoError(t, err)

	ag_require.Equal(t, byte(255), batchData[0])
	ag_require.Equal(t, byte(3), batchData[1])  // num_accounts
	ag_require.Equal(t, byte(9), batchData[2])  // data_len
	ag_require.Equal(t, byte(3), batchData[3])  // Transfer discriminator
	amount := binary.LittleEndian.Uint64(batchData[4:12])
	ag_require.Equal(t, uint64(50), amount)
}

func TestBatch_DecodeRoundTrip(t *testing.T) {
	src := ag_solanago.NewWallet().PublicKey()
	dst := ag_solanago.NewWallet().PublicKey()
	owner := ag_solanago.NewWallet().PublicKey()

	transfer1 := NewTransferInstruction(100, src, dst, owner, nil).Build()
	transfer2 := NewTransferInstruction(200, src, dst, owner, nil).Build()

	batch := NewBatchInstruction(transfer1, transfer2)
	built := batch.Build()

	data, err := built.Data()
	ag_require.NoError(t, err)

	accounts := built.Accounts()
	decoded, err := DecodeInstruction(accounts, data)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint8(255), decoded.TypeID.Uint8())

	decodedBatch, ok := decoded.Impl.(*Batch)
	ag_require.True(t, ok)
	ag_require.Len(t, decodedBatch.Instructions, 2)

	t1, ok := decodedBatch.Instructions[0].Impl.(*Transfer)
	ag_require.True(t, ok)
	ag_require.Equal(t, uint64(100), *t1.Amount)

	t2, ok := decodedBatch.Instructions[1].Impl.(*Transfer)
	ag_require.True(t, ok)
	ag_require.Equal(t, uint64(200), *t2.Amount)
}

func TestBatch_MixedInstructions(t *testing.T) {
	src := ag_solanago.NewWallet().PublicKey()
	dst := ag_solanago.NewWallet().PublicKey()
	owner := ag_solanago.NewWallet().PublicKey()
	mint := ag_solanago.NewWallet().PublicKey()

	transfer := NewTransferInstruction(500, src, dst, owner, nil).Build()
	getSize := NewGetAccountDataSizeInstruction(mint).Build()

	batch := NewBatchInstruction(transfer, getSize)
	built := batch.Build()

	data, err := built.Data()
	ag_require.NoError(t, err)

	ag_require.Equal(t, byte(255), data[0])

	// Sub-instruction 1: Transfer (3 accounts, 9 bytes data)
	ag_require.Equal(t, byte(3), data[1])
	ag_require.Equal(t, byte(9), data[2])
	ag_require.Equal(t, byte(3), data[3]) // Transfer ID

	// Sub-instruction 2: GetAccountDataSize (1 account, 1 byte data)
	offset := 3 + 9
	ag_require.Equal(t, byte(1), data[offset])   // num_accounts
	ag_require.Equal(t, byte(1), data[offset+1]) // data_len
	ag_require.Equal(t, byte(21), data[offset+2]) // GetAccountDataSize ID
}
