// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
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

package solana

import (
	"encoding/base64"
	"testing"

	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// newTestInstruction is a shorthand for creating a testTransactionInstructions.
func newTestInstruction(programID PublicKey, accounts []*AccountMeta, data []byte) *testTransactionInstructions {
	return &testTransactionInstructions{
		accounts:  accounts,
		data:      data,
		programID: programID,
	}
}

type testTransactionInstructions struct {
	accounts  []*AccountMeta
	data      []byte
	programID PublicKey
}

func (t *testTransactionInstructions) Accounts() []*AccountMeta {
	return t.accounts
}

func (t *testTransactionInstructions) ProgramID() PublicKey {
	return t.programID
}

func (t *testTransactionInstructions) Data() ([]byte, error) {
	return t.data, nil
}

func TestNewTransaction(t *testing.T) {
	debugNewTransaction = true

	instructions := []Instruction{
		&testTransactionInstructions{
			accounts: []*AccountMeta{
				{PublicKey: MustPublicKeyFromBase58("A9QnpgfhCkmiBSjgBuWk76Wo3HxzxvDopUq9x6UUMmjn"), IsSigner: true, IsWritable: false},
				{PublicKey: MustPublicKeyFromBase58("9hFtYBYmBJCVguRYs9pBTWKYAFoKfjYR7zBPpEkVsmD"), IsSigner: true, IsWritable: true},
			},
			data:      []byte{0xaa, 0xbb},
			programID: SystemProgramID,
		},
		&testTransactionInstructions{
			accounts: []*AccountMeta{
				{PublicKey: MustPublicKeyFromBase58("SysvarC1ock11111111111111111111111111111111"), IsSigner: false, IsWritable: false},
				{PublicKey: MustPublicKeyFromBase58("SysvarS1otHashes111111111111111111111111111"), IsSigner: false, IsWritable: true},
				{PublicKey: MustPublicKeyFromBase58("9hFtYBYmBJCVguRYs9pBTWKYAFoKfjYR7zBPpEkVsmD"), IsSigner: false, IsWritable: true},
				{PublicKey: MustPublicKeyFromBase58("6FzXPEhCJoBx7Zw3SN9qhekHemd6E2b8kVguitmVAngW"), IsSigner: true, IsWritable: false},
			},
			data:      []byte{0xcc, 0xdd},
			programID: MustPublicKeyFromBase58("Vote111111111111111111111111111111111111111"),
		},
	}

	blockhash, err := HashFromBase58("A9QnpgfhCkmiBSjgBuWk76Wo3HxzxvDopUq9x6UUMmjn")
	require.NoError(t, err)

	trx, err := NewTransaction(instructions, blockhash)
	require.NoError(t, err)

	assert.Equal(t, trx.Message.Header, MessageHeader{
		NumRequiredSignatures:       3,
		NumReadonlySignedAccounts:   1,
		NumReadonlyUnsignedAccounts: 3,
	})

	assert.Equal(t, trx.Message.RecentBlockhash, blockhash)

	assert.Equal(t, trx.Message.AccountKeys, PublicKeySlice{
		MustPublicKeyFromBase58("A9QnpgfhCkmiBSjgBuWk76Wo3HxzxvDopUq9x6UUMmjn"),
		MustPublicKeyFromBase58("9hFtYBYmBJCVguRYs9pBTWKYAFoKfjYR7zBPpEkVsmD"),
		MustPublicKeyFromBase58("6FzXPEhCJoBx7Zw3SN9qhekHemd6E2b8kVguitmVAngW"),
		MustPublicKeyFromBase58("SysvarS1otHashes111111111111111111111111111"),
		SystemProgramID,
		MustPublicKeyFromBase58("SysvarC1ock11111111111111111111111111111111"),
		MustPublicKeyFromBase58("Vote111111111111111111111111111111111111111"),
	})

	assert.Equal(t, trx.Message.Instructions, []CompiledInstruction{
		{
			ProgramIDIndex: 4,
			Accounts:       []uint16{0, 0o1},
			Data:           []byte{0xaa, 0xbb},
		},
		{
			ProgramIDIndex: 6,
			Accounts:       []uint16{5, 3, 1, 2},
			Data:           []byte{0xcc, 0xdd},
		},
	})
}

func TestPartialSignTransaction(t *testing.T) {
	signers := []PrivateKey{
		NewWallet().PrivateKey,
		NewWallet().PrivateKey,
		NewWallet().PrivateKey,
	}
	instructions := []Instruction{
		&testTransactionInstructions{
			accounts: []*AccountMeta{
				{PublicKey: signers[0].PublicKey(), IsSigner: true, IsWritable: false},
				{PublicKey: signers[1].PublicKey(), IsSigner: true, IsWritable: true},
				{PublicKey: signers[2].PublicKey(), IsSigner: true, IsWritable: false},
			},
			data:      []byte{0xaa, 0xbb},
			programID: SystemProgramID,
		},
	}

	blockhash, err := HashFromBase58("A9QnpgfhCkmiBSjgBuWk76Wo3HxzxvDopUq9x6UUMmjn")
	require.NoError(t, err)

	trx, err := NewTransaction(instructions, blockhash)
	require.NoError(t, err)

	assert.Equal(t, trx.Message.Header.NumRequiredSignatures, uint8(3))

	// Test various signing orders
	signingOrders := [][]int{
		{0, 1, 2}, // ABC
		{0, 2, 1}, // ACB
		{1, 0, 2}, // BAC
		{1, 2, 0}, // BCA
		{2, 0, 1}, // CAB
		{2, 1, 0}, // CBA
	}

	for _, order := range signingOrders {
		// Reset the transaction signatures before each test
		trx.Signatures = make([]Signature, len(signers))

		// Sign the transaction in the specified order
		for _, idx := range order {
			signer := signers[idx]
			signatures, err := trx.PartialSign(func(key PublicKey) *PrivateKey {
				if key.Equals(signer.PublicKey()) {
					return &signer
				}
				return nil
			})
			require.NoError(t, err)
			assert.Equal(t, len(signatures), 3)
		}
		// Verify Signatures
		require.NoError(t, trx.VerifySignatures())
	}
}

func TestSignTransaction(t *testing.T) {
	signers := []PrivateKey{
		NewWallet().PrivateKey,
		NewWallet().PrivateKey,
	}
	instructions := []Instruction{
		&testTransactionInstructions{
			accounts: []*AccountMeta{
				{PublicKey: signers[0].PublicKey(), IsSigner: true, IsWritable: false},
				{PublicKey: signers[1].PublicKey(), IsSigner: true, IsWritable: true},
			},
			data:      []byte{0xaa, 0xbb},
			programID: SystemProgramID,
		},
	}

	blockhash, err := HashFromBase58("A9QnpgfhCkmiBSjgBuWk76Wo3HxzxvDopUq9x6UUMmjn")
	require.NoError(t, err)

	trx, err := NewTransaction(instructions, blockhash)
	require.NoError(t, err)

	assert.Equal(t, trx.Message.Header.NumRequiredSignatures, uint8(2))

	t.Run("should reject missing signer(s)", func(t *testing.T) {
		_, err := trx.Sign(func(key PublicKey) *PrivateKey {
			if key.Equals(signers[0].PublicKey()) {
				return &signers[0]
			}
			return nil
		})
		require.Error(t, err)
	})

	t.Run("should sign with signer(s)", func(t *testing.T) {
		signatures, err := trx.Sign(func(key PublicKey) *PrivateKey {
			for _, signer := range signers {
				if key.Equals(signer.PublicKey()) {
					return &signer
				}
			}
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, len(signatures), 2)
	})
}

func FuzzTransaction(f *testing.F) {
	encoded := "AfjEs3XhTc3hrxEvlnMPkm/cocvAUbFNbCl00qKnrFue6J53AhEqIFmcJJlJW3EDP5RmcMz+cNTTcZHW/WJYwAcBAAEDO8hh4VddzfcO5jbCt95jryl6y8ff65UcgukHNLWH+UQGgxCGGpgyfQVQV02EQYqm4QwzUt2qf9f1gVLM7rI4hwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA6ANIF55zOZWROWRkeh+lExxZBnKFqbvIxZDLE7EijjoBAgIAAQwCAAAAOTAAAAAAAAA="
	data, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(f, err)
	f.Add(data)

	f.Fuzz(func(t *testing.T, data []byte) {
		require.NotPanics(t, func() {
			TransactionFromDecoder(bin.NewBinDecoder(data))
		})
	})
}

func TestTransactionDecode(t *testing.T) {
	encoded := "AfjEs3XhTc3hrxEvlnMPkm/cocvAUbFNbCl00qKnrFue6J53AhEqIFmcJJlJW3EDP5RmcMz+cNTTcZHW/WJYwAcBAAEDO8hh4VddzfcO5jbCt95jryl6y8ff65UcgukHNLWH+UQGgxCGGpgyfQVQV02EQYqm4QwzUt2qf9f1gVLM7rI4hwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA6ANIF55zOZWROWRkeh+lExxZBnKFqbvIxZDLE7EijjoBAgIAAQwCAAAAOTAAAAAAAAA="
	data, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)

	tx, err := TransactionFromDecoder(bin.NewBinDecoder(data))
	require.NoError(t, err)
	require.NotNil(t, tx)

	require.Len(t, tx.Signatures, 1)
	require.Equal(t,
		MustSignatureFromBase58("5yUSwqQqeZLEEYKxnG4JC4XhaaBpV3RS4nQbK8bQTyjLX5btVq9A1Ja5nuJzV7Z3Zq8G6EVKFvN4DKUL6PSAxmTk"),
		tx.Signatures[0],
	)

	require.Equal(t,
		PublicKeySlice{
			MustPublicKeyFromBase58("52NGrUqh6tSGhr59ajGxsH3VnAaoRdSdTbAaV9G3UW35"),
			MustPublicKeyFromBase58("SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt"),
			SystemProgramID,
		},
		tx.Message.AccountKeys,
	)

	require.Equal(t,
		MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 1,
		},
		tx.Message.Header,
	)

	require.Equal(t,
		MustHashFromBase58("GcgVK9buRA7YepZh3zXuS399GJAESCisLnLDBCmR5Aoj"),
		tx.Message.RecentBlockhash,
	)

	decodedData, err := base58.Decode("3Bxs4ART6LMJ13T5")
	require.NoError(t, err)
	require.Equal(t, 12, len(decodedData))
	require.Equal(t, []byte{2, 0, 0, 0, 57, 48, 0, 0, 0, 0, 0, 0}, decodedData)
	require.Equal(t,
		[]CompiledInstruction{
			{
				ProgramIDIndex: 2,
				Accounts: []uint16{
					0,
					1,
				},
				Data: Base58(decodedData),
			},
		},
		tx.Message.Instructions,
	)
}

func TestTransactionVerifySignatures(t *testing.T) {
	type testCase struct {
		Transaction string
	}

	testCases := []testCase{
		{
			Transaction: "AVBFwRrn4wroV9+NVQfgg/GbjFtQFodLnNI5oTpDMQiQ4HfZNyFzcFamHSSFW4p5wc3efeEKvykbmk8jzf2LCQwBAAIGjYddInd/DSl2KJCP18GhEDlaJyPKVrgBGGsr3TF6jSYPgr3AdITNKr2UQVQ5I+Wh5StQv/a5XdLr6VN4Y21My1M/Y1FNK5wQLKJa1LYfN/HAudufFVtc0fRPR6AMUJ9UrkRI7sjY/PnpcXLF7A7SBvJrWu+o8+7QIaD8sL9aXkGFDy1uAqR6+CTQmradxC1wyyjL+iSft+5XudJWwSdi7wAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAi+i1vCST+HNO0DEchpEJImMHhZ1BReuf7poRqmXpeA8CBAUBAgMCAgcAAwAAAAEABQIAAAwCAAAA6w0AAAAAAAA=",
		},
		{
			Transaction: "AWwhMTxKhl9yZOlidY0u3gYmy+J/6V3kFSXU7GgK5zwN+SwljR2dOlHgKtUDRX8uee2HtfeyL3t4lB3n749L4QQBAAIEFg+6wTr33dgF0xcKPeDGvZcSah4CwNJZ0Khu+CHW5cehpkZfTC6/JEwx2AvJXCc0WjQk5CjC3vM+ztnpDT9wGwan1RcYx3TJKFZjmGkdXraLXrijm0ttXHNVWyEAAAAA3OXr4eScO58RTLVUTFCpnsDWktY/Vnla4Cmsg9nqi+Jr/+AAgahV8wmBK4mnz9WwJSryq8x2Ic0asytADGhLZAEDAwABAigCAAAABwAAAAEAAAAAAAAAz+dyuQIAAAAIn18BAAAAAPsVKAcAAAAA",
		},
		{
			Transaction: "ARZsk8+AvvT9onUT8FU1VRaiC8Sp+FKveOwhdPoigWHA+MGNcIOqbow6mwSILEYvvyOB/fi3UQ/xKQCjEtxBRgIBAAIFKIX92BRrkgEfrLEXAvXtw7OgPPhHU+62C8DB5QPoMgNSbKXgdub0sr7Yp3Nvdrsp6SDoJ4gdoyRad2AV+Japj0dRtYW4OxE78FvRZTeqHFy2My/m12/afGIPS8iUnMGlBqfVFxjHdMkoVmOYaR1etoteuKObS21cc1VbIQAAAAC/jt8clGtWu0PSX5i4e2vlERcwCmEmGvn5+U7telqAiK4hdAN78GteFjqtJrxLXxpVNKsu1lfdcFPXa/Kcg4e5AQQEAQADAicmMiQQAiGujz0xoTQSQCgAMPOroDk5F0hQ/BgzEkBBvVKWIY41EkA=",
		},
		{
			Transaction: "Ad7TPpYTvSpO//KNA5YTZVojVwz4NlH4gH9ktl+rTObJcgo8QkqmHK4t6DQr9dD58B/A/5/N7v9K+0j6y1TVCAsBAAMFA9maY4S727Z/lOSb08nHehVFsC32kTKMMPjPJp111bKM0Fl1Dg04vV2x9nL2TCqSHmjT8xg6wUAzjZa1+6YCBQan1RcZLwqvxvJl4/t3zHragsUp0L47E24tAFUgAAAABqfVFxjHdMkoVmOYaR1etoteuKObS21cc1VbIQAAAAAHYUgdNXR0u3xNdiTr072z2DVec9EQQ/wNo1OAAAAAAJDQfslK1yQFkGqDXWu6cthRNuYGlajYMOmtoSJB6hmPAQQEAQIDAE0CAAAAAwAAAAAAAAD5FSgHAAAAAPoVKAcAAAAA+xUoBwAAAADECMJOPX7e7fOF5Hrq9xhdch2Uqhg8vQOYyZM/6V983gHQ0gNiAAAAAA==",
		},
		{
			Transaction: "Ak8jvC3ch5hq3lhOHPkACoFepIUON2zEN4KRcw4lDS6GBsQfnSdzNGPETm/yi0hPKk75/i2VXFj0FLUWnGR64ADyUbqnirFjFtaSNgcGi02+Tm7siT4CPpcaTq0jxfYQK/h9FdxXXPnLry74J+RE8yji/BtJ/Cjxbx+TIHigeIYJAgEBBByE1Y6EqCJKsr7iEupU6lsBHtBdtI4SK3yWMCFA0iEKeFPgnGmtp+1SIX1Ak+sN65iBaR7v4Iim5m1OEuFQTgi9N57UnhNpCNuUePaTt7HJaFBmyeZB3deXeKWVudpY3gAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWVECK/n3a7QR6OKWYR4DuAVjS6FXgZj82W0dJpSIPnEBAwQAAgEDDAIAAABAQg8AAAAAAA==",
		},
	}

	for _, tc := range testCases {
		txBin, err := base64.StdEncoding.DecodeString(tc.Transaction)
		require.NoError(t, err)
		tx, err := TransactionFromDecoder(bin.NewBinDecoder(txBin))
		require.NoError(t, err)
		require.NoError(t, tx.VerifySignatures())
		require.Equal(t, len(tx.Signatures), len(tx.Message.Signers()))
	}
}

func TestTransactionSerializeExisting(t *testing.T) {
	// random pump amm swap transaction (3HWKcTbnAMXt3TZDi8LitZCAT5ht7tYqXCroQgNkEnRuvjWCAhmx4UAFnKTWzxS2JXxhbfTiKEdXeU3VHWKzEkNY)
	trueEncoded := "AXJEirR5ePYXWIemCsSrRB3kTOxBQvJ8pZ4Of+vs75/Lw4NgNf0jr+eyI+2CAZZcwEQ54v/tcIh0p5qisd6ZfQWAAQAHDuIP6DQ7XgVvZzx4ZyZS5BjxS0s5JIGa893M/2foLj/N9tJulKNTEJ+CcURWZpXddGLJc0niyvBs8fADaZXWBStZSYulGfGwKCcEVhAem2vYiJnZnDQHW8EbXuli2pgFPcs4OJAtX+MJFvqNDoxBApNOXVmhOPi+s+Xj5G6dipCzZIG9Det/kYMpmklt7LaKvckpmIhAC+cygPT/H7L6uDUm47unlLqsWvR0JhT3lhSFUnSWfDsT92IHGjplDB07Bwa4Mppngp8gB0rx3o6jx3q5zZqJ7jaF//YA5f9pM0rWAwZGb+UhFzL/7K26csOb57yM5bvF9xJrLEObOkAAAACMlyWPTiSJ8bs9ECkUjg2DC1oTmdr/EIQEjnvY2+n4WQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABt324ddloZPZy+FGzut5rBy0he1fWzeROoz1hX7/AKkMFN78gl7GdpQlCBi7ZUBl9CmNMVbVcbTU+AkMGOmoY2CQL4wWkL0iw2m16tD4VGZ2ZSCn9Rr5uDY8UTDVNj2KSwXRBHtnMBytk20Zd1LAlKbsLUEcQ8RupHeXS0aTLV9S0zqFMcANe7dmXMUOgO/qYKWak7hGKVjQNpnUxyaEPQgHAAUCkHwBAAcACQNkjocBAAAAAAgGAAEAEAkKAQELEQwADg0QAgEDBBEPCgoJCBILGDPmhaQBf4OtOLwvMTcAAABscskIAAAAAAoDAQAAAQkKAwIAAAEJCQIABQwCAAAAQEtMAAAAAAAJAgAGDAIAAABAQg8AAAAAAAEcQpjY9205kUMXxtRgI8+U286AVT/u2xYmbclxYfAs+QJDZQMS0kc="

	tx, err := TransactionFromBase64(trueEncoded)
	require.NoError(t, err)

	encoded, err := tx.ToBase64()
	require.NoError(t, err)
	require.Equal(t, trueEncoded, encoded)
}

func TestTransactionSerializePumpFunSwap(t *testing.T) {
	expectedEncoded := "AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAcMC8/60geEHarnEMtcmE3t7lADNe2/gxa7NAhBe5Ufe5mFrdnKgSF1vS6tlbOFzHIhVOQc18kOVGRpppm6a1S4mJ3oue9Q5rbnoxPRibQwb/LpSq3Y9feKOsv28L15NoYrrRHmpPwpRKT6glG++BVCbhv7KMa2ZGZ3YHxq2fVmpkb4+4AeAMe2pCpZXV9h1tGabHA5XBtMYEhq5uKj2C46VwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAVbg9pNmWs9E2xVovxdbqlGJy5f10v87ZV0rtv1tGLAGp9UXGSxcUSGMyUw9SvF/WNruCJuh/UTj29mKAAAAAAbd9uHXZaGT2cvhRs7reawctIXtX1s3kTqM9YV+/wCpOoZeae4PVIDKvPZjV+TcLxjVjUXB6nSJ+zcj2Xk8cqas8TbrAfwcTog9I8i1hEq1mjf2at1XxemsO1PgWdNcZOnOJ75MojvVd5YEoaM5xPpH75XeQqEuKjDCqZ1DhK2P0WlzNBoKArNIcmjqJI0o+XoQGnMjSEA/HZqyYrSJgLkBBgwJAwsEAQIABQgHCgYYZgY9EgHa6+pMR9bEaRkAAHg1HjQAAAAA"
	instruction := &testTransactionInstructions{
		programID: MPK("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"),
		accounts: []*AccountMeta{
			{PublicKey: MPK("4wTV1YmiEkRvAtNtsSGPtUrqRYQMe5SKy2uB4Jjaxnjf"), IsSigner: false, IsWritable: false},
			{PublicKey: MPK("CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM"), IsSigner: false, IsWritable: true},
			{PublicKey: MPK("GjgKTqtzDei5E3uZyA2CN29KQgugF564K1hoc1jHpump"), IsSigner: false, IsWritable: false},
			{PublicKey: MPK("HkvYAZV1Mg6kt5KMaA5YBQazZECg21zaZdQEMUiLrjKc"), IsSigner: false, IsWritable: true},
			{PublicKey: MPK("9zpyjwrYdRWNMyqicoiuL3gUcrbvrkd5Kq9nxui1znw1"), IsSigner: false, IsWritable: true},
			{PublicKey: MPK("BdQqJnuqqFhNZUNYGEEsuhBidpf8qHqfjDQvcjDN3nti"), IsSigner: false, IsWritable: true},
			{PublicKey: MPK("o7RY6P2vQMuGSu1TrLM81weuzgDjaCRTXYRaXJwWcvc"), IsSigner: true, IsWritable: true},
			{PublicKey: SystemProgramID, IsSigner: false, IsWritable: false},
			{PublicKey: MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"), IsSigner: false, IsWritable: false},
			{PublicKey: MPK("SysvarRent111111111111111111111111111111111"), IsSigner: false, IsWritable: false},
			{PublicKey: MPK("Ce6TQqeHC9p8KetsN6JsjHK7UTZk7nasjjnr7XxXp9F1"), IsSigner: false, IsWritable: false},
			{PublicKey: MPK("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"), IsSigner: false, IsWritable: false},
		},
		data: []byte{102, 6, 61, 18, 1, 218, 235, 234, 76, 71, 214, 196, 105, 25, 0, 0, 120, 53, 30, 52, 0, 0, 0, 0},
	}

	tx, err := NewTransactionBuilder().
		AddInstruction(instruction).
		SetFeePayer(MPK("o7RY6P2vQMuGSu1TrLM81weuzgDjaCRTXYRaXJwWcvc")).
		SetRecentBlockHash(MustHashFromBase58("F6TUDvYPMwDLP1MW4BUWTNm6S94XR1UZ2nGVyubqo6oi")).
		Build()
	require.NoError(t, err)
	require.NotNil(t, tx)

	encoded, err := tx.ToBase64()
	require.NoError(t, err)

	zlog.Debug("encoded", zap.String("encoded", encoded))
	require.Equal(t, expectedEncoded, encoded)
}

func BenchmarkTransactionFromDecoder(b *testing.B) {
	txString := "Ak8jvC3ch5hq3lhOHPkACoFepIUON2zEN4KRcw4lDS6GBsQfnSdzNGPETm/yi0hPKk75/i2VXFj0FLUWnGR64ADyUbqnirFjFtaSNgcGi02+Tm7siT4CPpcaTq0jxfYQK/h9FdxXXPnLry74J+RE8yji/BtJ/Cjxbx+TIHigeIYJAgEBBByE1Y6EqCJKsr7iEupU6lsBHtBdtI4SK3yWMCFA0iEKeFPgnGmtp+1SIX1Ak+sN65iBaR7v4Iim5m1OEuFQTgi9N57UnhNpCNuUePaTt7HJaFBmyeZB3deXeKWVudpY3gAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWVECK/n3a7QR6OKWYR4DuAVjS6FXgZj82W0dJpSIPnEBAwQAAgEDDAIAAABAQg8AAAAAAA=="
	txBin, err := base64.StdEncoding.DecodeString(txString)
	if err != nil {
		b.Error(err)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := TransactionFromDecoder(bin.NewBinDecoder(txBin))
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkTransactionVerifySignatures(b *testing.B) {
	txString := "Ak8jvC3ch5hq3lhOHPkACoFepIUON2zEN4KRcw4lDS6GBsQfnSdzNGPETm/yi0hPKk75/i2VXFj0FLUWnGR64ADyUbqnirFjFtaSNgcGi02+Tm7siT4CPpcaTq0jxfYQK/h9FdxXXPnLry74J+RE8yji/BtJ/Cjxbx+TIHigeIYJAgEBBByE1Y6EqCJKsr7iEupU6lsBHtBdtI4SK3yWMCFA0iEKeFPgnGmtp+1SIX1Ak+sN65iBaR7v4Iim5m1OEuFQTgi9N57UnhNpCNuUePaTt7HJaFBmyeZB3deXeKWVudpY3gAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWVECK/n3a7QR6OKWYR4DuAVjS6FXgZj82W0dJpSIPnEBAwQAAgEDDAIAAABAQg8AAAAAAA=="
	txBin, err := base64.StdEncoding.DecodeString(txString)
	if err != nil {
		b.Error(err)
	}

	tx, err := TransactionFromDecoder(bin.NewBinDecoder(txBin))
	if err != nil {
		b.Error(err)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tx.VerifySignatures()
	}
}

// Ported from solana-sdk/transaction/src/lib.rs: test_transaction_serialize.
// Tests that a transaction survives binary serialization roundtrip.
func TestTransactionSerializationRoundtrip(t *testing.T) {
	signers := []PrivateKey{
		NewWallet().PrivateKey,
		NewWallet().PrivateKey,
	}
	instructions := []Instruction{
		newTestInstruction(
			SystemProgramID,
			[]*AccountMeta{
				{PublicKey: signers[0].PublicKey(), IsSigner: true, IsWritable: true},
				{PublicKey: signers[1].PublicKey(), IsSigner: true, IsWritable: false},
			},
			[]byte{0x01, 0x02, 0x03},
		),
	}

	blockhash, err := HashFromBase58("A9QnpgfhCkmiBSjgBuWk76Wo3HxzxvDopUq9x6UUMmjn")
	require.NoError(t, err)

	tx, err := NewTransaction(instructions, blockhash)
	require.NoError(t, err)

	_, err = tx.Sign(func(key PublicKey) *PrivateKey {
		for _, signer := range signers {
			if key.Equals(signer.PublicKey()) {
				return &signer
			}
		}
		return nil
	})
	require.NoError(t, err)

	// Marshal to binary.
	data, err := tx.MarshalBinary()
	require.NoError(t, err)

	// Unmarshal back.
	var decoded Transaction
	err = decoded.UnmarshalWithDecoder(bin.NewBinDecoder(data))
	require.NoError(t, err)

	// Compare.
	assert.Equal(t, tx.Signatures, decoded.Signatures)
	assert.Equal(t, tx.Message.Header, decoded.Message.Header)
	assert.Equal(t, tx.Message.AccountKeys, decoded.Message.AccountKeys)
	assert.Equal(t, tx.Message.RecentBlockhash, decoded.Message.RecentBlockhash)
	require.Equal(t, len(tx.Message.Instructions), len(decoded.Message.Instructions))
	for i := range tx.Message.Instructions {
		assert.Equal(t, tx.Message.Instructions[i].ProgramIDIndex, decoded.Message.Instructions[i].ProgramIDIndex)
		assert.Equal(t, tx.Message.Instructions[i].Accounts, decoded.Message.Instructions[i].Accounts)
	}

	// Verify signatures still valid after roundtrip.
	require.NoError(t, decoded.VerifySignatures())
}

// Ported from solana-sdk/transaction/src/lib.rs: test_sanitize_txs.
// Tests that signature count must match num_required_signatures.
func TestVerifySignatures_SignatureCountMismatch(t *testing.T) {
	signer := NewWallet().PrivateKey

	tx, err := NewTransaction([]Instruction{
		newTestInstruction(
			SystemProgramID,
			[]*AccountMeta{{PublicKey: signer.PublicKey(), IsSigner: true, IsWritable: true}},
			[]byte{0x01},
		),
	}, Hash{1, 2, 3})
	require.NoError(t, err)

	_, err = tx.Sign(func(key PublicKey) *PrivateKey {
		if key.Equals(signer.PublicKey()) {
			return &signer
		}
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, tx.VerifySignatures())

	// Too many signatures.
	tx.Signatures = append(tx.Signatures, tx.Signatures[0])
	err = tx.VerifySignatures()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signers")

	// Too few signatures.
	tx.Signatures = nil
	err = tx.VerifySignatures()
	require.Error(t, err)
}

// Ported from solana-sdk/transaction/src/lib.rs: test_transaction_instruction_with_duplicate_keys.
// Tests that duplicate account keys in instructions are deduplicated.
func TestNewTransaction_DuplicateAccountKeys(t *testing.T) {
	key := newUniqueKey()

	tx, err := NewTransaction([]Instruction{
		newTestInstruction(
			SystemProgramID,
			[]*AccountMeta{
				{PublicKey: key, IsSigner: true, IsWritable: true},
				{PublicKey: key, IsSigner: true, IsWritable: true}, // duplicate
			},
			[]byte{0x01},
		),
	}, Hash{})
	require.NoError(t, err)

	// The duplicate should be deduplicated in AccountKeys.
	// Should have: key (payer/signer) + system program = 2 keys.
	assert.Equal(t, 2, len(tx.Message.AccountKeys))
	assert.Equal(t, uint8(1), tx.Message.Header.NumRequiredSignatures)
}

// Ported from solana-sdk/transaction/src/lib.rs: test_transaction_correct_key.
// Tests that signing with the correct key produces verifiable signatures.
func TestTransaction_SignAndVerify(t *testing.T) {
	signer := NewWallet().PrivateKey

	tx, err := NewTransaction([]Instruction{
		newTestInstruction(
			SystemProgramID,
			[]*AccountMeta{{PublicKey: signer.PublicKey(), IsSigner: true, IsWritable: true}},
			[]byte{0xDE, 0xAD},
		),
	}, Hash{42})
	require.NoError(t, err)

	_, err = tx.Sign(func(key PublicKey) *PrivateKey {
		if key.Equals(signer.PublicKey()) {
			return &signer
		}
		return nil
	})
	require.NoError(t, err)

	// Valid signature.
	require.NoError(t, tx.VerifySignatures())

	// Tamper with the signature -> should fail.
	tx.Signatures[0][0] ^= 0xFF
	err = tx.VerifySignatures()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature")
}

// Tests that NewTransaction requires at least one instruction.
func TestNewTransaction_EmptyInstructions(t *testing.T) {
	_, err := NewTransaction(nil, Hash{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires at-least one instruction")

	_, err = NewTransaction([]Instruction{}, Hash{})
	require.Error(t, err)
}

// Tests TransactionBuilder.
func TestTransactionBuilder(t *testing.T) {
	signer := NewWallet().PrivateKey
	programID := SystemProgramID
	blockhash := Hash{1, 2, 3}

	tx, err := NewTransactionBuilder().
		AddInstruction(newTestInstruction(
			programID,
			[]*AccountMeta{{PublicKey: signer.PublicKey(), IsSigner: true, IsWritable: true}},
			[]byte{0x01},
		)).
		SetRecentBlockHash(blockhash).
		SetFeePayer(signer.PublicKey()).
		Build()
	require.NoError(t, err)

	assert.Equal(t, blockhash, tx.Message.RecentBlockhash)
	assert.Equal(t, uint8(1), tx.Message.Header.NumRequiredSignatures)
	assert.Equal(t, signer.PublicKey(), tx.Message.AccountKeys[0])
}

// Tests NumWriteableAccounts, NumReadonlyAccounts, NumSigners for legacy transactions.
// Ported from solana-sdk/transaction/src/lib.rs header validation tests.
func TestTransaction_AccountCounts_Legacy(t *testing.T) {
	keys := [5]PublicKey{}
	for i := range keys {
		keys[i] = newUniqueKey()
	}

	// 2 signers (1 writable, 1 readonly), 3 unsigned (1 writable, 2 readonly)
	// -> header: 2 required, 1 readonly_signed, 2 readonly_unsigned
	tx, err := NewTransaction([]Instruction{
		newTestInstruction(
			keys[4],
			[]*AccountMeta{
				{PublicKey: keys[0], IsSigner: true, IsWritable: true},
				{PublicKey: keys[1], IsSigner: true, IsWritable: false},
				{PublicKey: keys[2], IsSigner: false, IsWritable: true},
				{PublicKey: keys[3], IsSigner: false, IsWritable: false},
			},
			[]byte{0x01},
		),
	}, Hash{}, TransactionPayer(keys[0]))
	require.NoError(t, err)

	assert.Equal(t, uint8(2), tx.Message.Header.NumRequiredSignatures)
	assert.Equal(t, uint8(1), tx.Message.Header.NumReadonlySignedAccounts)
	// keys[3] (readonly unsigned) + keys[4] (program, readonly unsigned) = 2
	assert.Equal(t, uint8(2), tx.Message.Header.NumReadonlyUnsignedAccounts)

	assert.Equal(t, 2, tx.NumSigners())
	assert.Equal(t, 3, tx.NumReadonlyAccounts())  // 1 readonly signed + 2 readonly unsigned
	assert.Equal(t, 2, tx.NumWriteableAccounts()) // keys[0] writable signer + keys[2] writable unsigned
}

// Tests GetProgramIDs.
func TestTransaction_GetProgramIDs(t *testing.T) {
	prog1 := newUniqueKey()
	prog2 := newUniqueKey()
	signer := newUniqueKey()

	tx, err := NewTransaction([]Instruction{
		newTestInstruction(prog1, []*AccountMeta{{PublicKey: signer, IsSigner: true, IsWritable: true}}, []byte{0x01}),
		newTestInstruction(prog2, []*AccountMeta{{PublicKey: signer, IsSigner: false, IsWritable: false}}, []byte{0x02}),
	}, Hash{}, TransactionPayer(signer))
	require.NoError(t, err)

	programIDs, err := tx.GetProgramIDs()
	require.NoError(t, err)
	require.Equal(t, 2, len(programIDs))

	// Verify both programs are present.
	found := map[PublicKey]bool{}
	for _, pid := range programIDs {
		found[pid] = true
	}
	assert.True(t, found[prog1], "prog1 should be in program IDs")
	assert.True(t, found[prog2], "prog2 should be in program IDs")
}

// Tests IsVote.
func TestTransaction_IsVote(t *testing.T) {
	signer := newUniqueKey()
	t.Run("non-vote transaction", func(t *testing.T) {
		tx, err := NewTransaction([]Instruction{
			newTestInstruction(SystemProgramID, []*AccountMeta{{PublicKey: signer, IsSigner: true, IsWritable: true}}, []byte{0x01}),
		}, Hash{}, TransactionPayer(signer))
		require.NoError(t, err)
		assert.False(t, tx.IsVote())
	})

	t.Run("vote transaction", func(t *testing.T) {
		tx, err := NewTransaction([]Instruction{
			newTestInstruction(VoteProgramID, []*AccountMeta{{PublicKey: signer, IsSigner: true, IsWritable: true}}, []byte{0x01}),
		}, Hash{}, TransactionPayer(signer))
		require.NoError(t, err)
		assert.True(t, tx.IsVote())
	})
}

// Tests HasAccount, IsSigner, IsWritable on Transaction.
func TestTransaction_AccountQueries(t *testing.T) {
	signer := newUniqueKey()
	writable := newUniqueKey()
	readonly := newUniqueKey()
	programID := SystemProgramID

	tx, err := NewTransaction([]Instruction{
		newTestInstruction(programID, []*AccountMeta{
			{PublicKey: signer, IsSigner: true, IsWritable: true},
			{PublicKey: writable, IsSigner: false, IsWritable: true},
			{PublicKey: readonly, IsSigner: false, IsWritable: false},
		}, []byte{0x01}),
	}, Hash{}, TransactionPayer(signer))
	require.NoError(t, err)

	// HasAccount.
	has, err := tx.HasAccount(signer)
	require.NoError(t, err)
	assert.True(t, has)

	has, err = tx.HasAccount(newUniqueKey())
	require.NoError(t, err)
	assert.False(t, has)

	// IsSigner.
	assert.True(t, tx.IsSigner(signer))
	assert.False(t, tx.IsSigner(writable))
	assert.False(t, tx.IsSigner(readonly))

	// IsWritable.
	w, err := tx.IsWritable(signer)
	require.NoError(t, err)
	assert.True(t, w, "signer should be writable")

	w, err = tx.IsWritable(writable)
	require.NoError(t, err)
	assert.True(t, w, "writable account should be writable")

	w, err = tx.IsWritable(readonly)
	require.NoError(t, err)
	assert.False(t, w, "readonly account should not be writable")
}

// Tests that MarshalBinary pads missing signatures with zeroes.
// Ported from solana-web3.js reference in the Go code comment.
func TestMarshalBinary_PadsMissingSignatures(t *testing.T) {
	signer := newUniqueKey()

	tx := &Transaction{
		Message: Message{
			Header: MessageHeader{
				NumRequiredSignatures:       2,
				NumReadonlySignedAccounts:   0,
				NumReadonlyUnsignedAccounts: 1,
			},
			AccountKeys: PublicKeySlice{
				signer,
				newUniqueKey(),
				SystemProgramID,
			},
			RecentBlockhash: Hash{1},
			Instructions: []CompiledInstruction{
				{ProgramIDIndex: 2, Accounts: []uint16{0, 1}, Data: []byte{0x01}},
			},
		},
		// Only 0 signatures provided, but 2 required.
		Signatures: nil,
	}

	data, err := tx.MarshalBinary()
	require.NoError(t, err)

	// Decode and check that 2 dummy signatures were added.
	var decoded Transaction
	err = decoded.UnmarshalWithDecoder(bin.NewBinDecoder(data))
	require.NoError(t, err)
	assert.Equal(t, 2, len(decoded.Signatures))
	// Both should be zero-filled.
	assert.Equal(t, Signature{}, decoded.Signatures[0])
	assert.Equal(t, Signature{}, decoded.Signatures[1])
}

// Tests base64 roundtrip on Transaction.
func TestTransaction_Base64Roundtrip(t *testing.T) {
	signer := NewWallet().PrivateKey

	tx, err := NewTransaction([]Instruction{
		newTestInstruction(
			SystemProgramID,
			[]*AccountMeta{{PublicKey: signer.PublicKey(), IsSigner: true, IsWritable: true}},
			[]byte{0xCA, 0xFE},
		),
	}, Hash{99}, TransactionPayer(signer.PublicKey()))
	require.NoError(t, err)

	_, err = tx.Sign(func(key PublicKey) *PrivateKey {
		if key.Equals(signer.PublicKey()) {
			return &signer
		}
		return nil
	})
	require.NoError(t, err)

	b64, err := tx.ToBase64()
	require.NoError(t, err)
	require.NotEmpty(t, b64)

	decoded, err := TransactionFromBase64(b64)
	require.NoError(t, err)
	require.NoError(t, decoded.VerifySignatures())

	assert.Equal(t, tx.Signatures, decoded.Signatures)
	assert.Equal(t, tx.Message.Header, decoded.Message.Header)
	assert.Equal(t, tx.Message.AccountKeys, decoded.Message.AccountKeys)
}

// Tests TransactionFromBytes and TransactionFromBase58.
func TestTransaction_FromVariousFormats(t *testing.T) {
	// Use a known valid transaction.
	b64 := "AfjEs3XhTc3hrxEvlnMPkm/cocvAUbFNbCl00qKnrFue6J53AhEqIFmcJJlJW3EDP5RmcMz+cNTTcZHW/WJYwAcBAAEDO8hh4VddzfcO5jbCt95jryl6y8ff65UcgukHNLWH+UQGgxCGGpgyfQVQV02EQYqm4QwzUt2qf9f1gVLM7rI4hwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA6ANIF55zOZWROWRkeh+lExxZBnKFqbvIxZDLE7EijjoBAgIAAQwCAAAAOTAAAAAAAAA="
	data, err := base64.StdEncoding.DecodeString(b64)
	require.NoError(t, err)

	// FromBytes.
	tx1, err := TransactionFromBytes(data)
	require.NoError(t, err)
	require.NotNil(t, tx1)
	require.NoError(t, tx1.VerifySignatures())

	// FromBase64.
	tx2, err := TransactionFromBase64(b64)
	require.NoError(t, err)
	require.NotNil(t, tx2)
	assert.Equal(t, tx1.Signatures, tx2.Signatures)
	assert.Equal(t, tx1.Message.Header, tx2.Message.Header)
}

// Tests Transaction.String() doesn't panic (EncodeToTree coverage).
func TestTransaction_String_NoPanic(t *testing.T) {
	signer := NewWallet().PrivateKey

	tx, err := NewTransaction([]Instruction{
		newTestInstruction(
			SystemProgramID,
			[]*AccountMeta{{PublicKey: signer.PublicKey(), IsSigner: true, IsWritable: true}},
			[]byte{0x01},
		),
	}, Hash{}, TransactionPayer(signer.PublicKey()))
	require.NoError(t, err)

	_, err = tx.Sign(func(key PublicKey) *PrivateKey {
		if key.Equals(signer.PublicKey()) {
			return &signer
		}
		return nil
	})
	require.NoError(t, err)

	require.NotPanics(t, func() {
		s := tx.String()
		assert.NotEmpty(t, s)
	})
}

// Tests that fee payer is always first in account keys.
// Ported from solana-sdk/transaction/src/lib.rs: test_message_payer_first.
func TestNewTransaction_FeePayerFirst(t *testing.T) {
	payer := newUniqueKey()
	other := newUniqueKey()
	programID := SystemProgramID

	tx, err := NewTransaction([]Instruction{
		newTestInstruction(programID, []*AccountMeta{{PublicKey: other, IsSigner: true, IsWritable: true}}, []byte{0x01}),
	}, Hash{}, TransactionPayer(payer))
	require.NoError(t, err)

	// Fee payer must be first.
	assert.Equal(t, payer, tx.Message.AccountKeys[0])
	// Fee payer is always a writable signer.
	assert.True(t, tx.IsSigner(payer))
	w, err := tx.IsWritable(payer)
	require.NoError(t, err)
	assert.True(t, w)
}

// Tests NumWriteableAccounts for V0 transactions with address table lookups.
func TestTransaction_NumWriteableAccounts_V0(t *testing.T) {
	payer := newUniqueKey()
	programID := newUniqueKey()
	acctA := newUniqueKey()
	acctB := newUniqueKey()
	tableKey := newUniqueKey()

	tables := map[PublicKey]PublicKeySlice{
		tableKey: {acctA, acctB},
	}

	tx, err := NewTransaction(
		[]Instruction{
			newTestInstruction(programID, []*AccountMeta{
				{PublicKey: acctA, IsSigner: false, IsWritable: true},
				{PublicKey: acctB, IsSigner: false, IsWritable: false},
			}, []byte{0x01}),
		},
		Hash{},
		TransactionPayer(payer),
		TransactionAddressTables(tables),
	)
	require.NoError(t, err)
	require.True(t, tx.Message.IsVersioned())

	// payer (writable signer) + acctA (writable lookup) = 2 writable
	assert.Equal(t, 2, tx.NumWriteableAccounts())
}

// Tests PartialSign with no signer provided doesn't corrupt state.
func TestPartialSign_NoSignerProvided(t *testing.T) {
	signer := NewWallet().PrivateKey

	tx, err := NewTransaction([]Instruction{
		newTestInstruction(
			SystemProgramID,
			[]*AccountMeta{{PublicKey: signer.PublicKey(), IsSigner: true, IsWritable: true}},
			[]byte{0x01},
		),
	}, Hash{}, TransactionPayer(signer.PublicKey()))
	require.NoError(t, err)

	// PartialSign with no matching key — should succeed but leave signature as zero.
	sigs, err := tx.PartialSign(func(key PublicKey) *PrivateKey {
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(sigs))
	assert.Equal(t, Signature{}, sigs[0])
}

// Tests that multiple instructions with the same program ID are handled correctly.
func TestNewTransaction_MultipleInstructionsSameProgram(t *testing.T) {
	signer := newUniqueKey()
	target := newUniqueKey()
	programID := SystemProgramID

	tx, err := NewTransaction([]Instruction{
		newTestInstruction(programID, []*AccountMeta{
			{PublicKey: signer, IsSigner: true, IsWritable: true},
			{PublicKey: target, IsSigner: false, IsWritable: true},
		}, []byte{0x01}),
		newTestInstruction(programID, []*AccountMeta{
			{PublicKey: signer, IsSigner: true, IsWritable: true},
			{PublicKey: target, IsSigner: false, IsWritable: true},
		}, []byte{0x02}),
	}, Hash{}, TransactionPayer(signer))
	require.NoError(t, err)

	// Program should appear only once in AccountKeys despite being used in 2 instructions.
	programCount := 0
	for _, key := range tx.Message.AccountKeys {
		if key.Equals(programID) {
			programCount++
		}
	}
	assert.Equal(t, 1, programCount, "program should appear only once")

	// Both instructions should reference the same program index.
	assert.Equal(t, tx.Message.Instructions[0].ProgramIDIndex, tx.Message.Instructions[1].ProgramIDIndex)
	assert.Equal(t, 2, len(tx.Message.Instructions))
}

// Tests nil transaction edge cases for count functions.
func TestTransaction_NilCounts(t *testing.T) {
	assert.Equal(t, -1, countSigners(nil))
	assert.Equal(t, -1, countReadonlyAccounts(nil))
	assert.Equal(t, -1, countWriteableAccounts(nil))
}
