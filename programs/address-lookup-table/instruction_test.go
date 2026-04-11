package addresslookuptable

import (
	"encoding/binary"
	"encoding/hex"
	"testing"

	solana "github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodingInstruction(t *testing.T) {
	tests := []struct {
		name              string
		hexData           string
		expectInstruction *Instruction
	}{
		{
			name:    "CreateLookupTable",
			hexData: "00000000" + "0a00000000000000" + "ff",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint32(Instruction_CreateLookupTable, binary.LittleEndian),
					Impl: &CreateLookupTable{
						RecentSlot: ptrUint64(10),
						BumpSeed:   ptrUint8(255),
					},
				},
			},
		},
		{
			name:    "FreezeLookupTable",
			hexData: "01000000",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint32(Instruction_FreezeLookupTable, binary.LittleEndian),
					Impl:   &FreezeLookupTable{},
				},
			},
		},
		{
			name:    "DeactivateLookupTable",
			hexData: "03000000",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint32(Instruction_DeactivateLookupTable, binary.LittleEndian),
					Impl:   &DeactivateLookupTable{},
				},
			},
		},
		{
			name:    "CloseLookupTable",
			hexData: "04000000",
			expectInstruction: &Instruction{
				BaseVariant: bin.BaseVariant{
					TypeID: bin.TypeIDFromUint32(Instruction_CloseLookupTable, binary.LittleEndian),
					Impl:   &CloseLookupTable{},
				},
			},
		},
	}

	t.Run("should encode", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				data, err := test.expectInstruction.Data()
				require.NoError(t, err)
				require.Equal(t, test.hexData, hex.EncodeToString(data))
			})
		}
	})

	t.Run("should decode", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				data, err := hex.DecodeString(test.hexData)
				require.NoError(t, err)
				var instruction *Instruction
				err = bin.NewBinDecoder(data).Decode(&instruction)
				require.NoError(t, err)
				assert.Equal(t, test.expectInstruction, instruction)
			})
		}
	})
}

func TestEncodeExtendLookupTable(t *testing.T) {
	addr1 := solana.SystemProgramID
	addr2 := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")

	lookupTable := solana.NewWallet().PublicKey()
	authority := solana.NewWallet().PublicKey()
	payer := solana.NewWallet().PublicKey()

	inst := NewExtendLookupTableInstruction(lookupTable, authority, payer, []solana.PublicKey{addr1, addr2})
	ix := inst.Build()

	data, err := ix.Data()
	require.NoError(t, err)

	require.Equal(t, 4+8+64, len(data))

	assert.Equal(t, "02000000", hex.EncodeToString(data[:4]))

	assert.Equal(t, "0200000000000000", hex.EncodeToString(data[4:12]))

	decoded, err := DecodeInstruction(ix.Accounts(), data)
	require.NoError(t, err)

	ext, ok := decoded.Impl.(*ExtendLookupTable)
	require.True(t, ok)
	require.Len(t, ext.Addresses, 2)
	assert.Equal(t, addr1, ext.Addresses[0])
	assert.Equal(t, addr2, ext.Addresses[1])
}

func TestCreateLookupTableBuild(t *testing.T) {
	authority := solana.NewWallet().PublicKey()
	payer := solana.NewWallet().PublicKey()
	recentSlot := uint64(42)

	inst, tableAddr, err := NewCreateLookupTableInstruction(authority, payer, recentSlot)
	require.NoError(t, err)
	require.False(t, tableAddr.IsZero())

	expectedAddr, expectedBump, err := DeriveLookupTableAddress(authority, recentSlot)
	require.NoError(t, err)
	assert.Equal(t, expectedAddr, tableAddr)
	assert.Equal(t, expectedBump, *inst.BumpSeed)

	ix := inst.Build()
	accounts := ix.Accounts()
	require.Len(t, accounts, 4)

	assert.Equal(t, tableAddr, accounts[0].PublicKey)
	assert.True(t, accounts[0].IsWritable)

	assert.Equal(t, authority, accounts[1].PublicKey)
	assert.True(t, accounts[1].IsSigner)

	assert.Equal(t, payer, accounts[2].PublicKey)
	assert.True(t, accounts[2].IsWritable)
	assert.True(t, accounts[2].IsSigner)

	assert.Equal(t, solana.SystemProgramID, accounts[3].PublicKey)
}

func TestDecodeSetsAccounts(t *testing.T) {
	t.Run("FreezeLookupTable", func(t *testing.T) {
		lookupTable := solana.NewWallet().PublicKey()
		authority := solana.NewWallet().PublicKey()

		inst := NewFreezeLookupTableInstruction(lookupTable, authority)
		ix := inst.Build()

		data, err := ix.Data()
		require.NoError(t, err)

		decoded, err := DecodeInstruction(ix.Accounts(), data)
		require.NoError(t, err)

		freeze, ok := decoded.Impl.(*FreezeLookupTable)
		require.True(t, ok)

		require.NotNil(t, freeze.GetLookupTableAccount())
		assert.Equal(t, lookupTable, freeze.GetLookupTableAccount().PublicKey)
		assert.True(t, freeze.GetLookupTableAccount().IsWritable)

		require.NotNil(t, freeze.GetAuthorityAccount())
		assert.Equal(t, authority, freeze.GetAuthorityAccount().PublicKey)
		assert.True(t, freeze.GetAuthorityAccount().IsSigner)
	})

	t.Run("DeactivateLookupTable", func(t *testing.T) {
		lookupTable := solana.NewWallet().PublicKey()
		authority := solana.NewWallet().PublicKey()

		inst := NewDeactivateLookupTableInstruction(lookupTable, authority)
		ix := inst.Build()

		data, err := ix.Data()
		require.NoError(t, err)

		decoded, err := DecodeInstruction(ix.Accounts(), data)
		require.NoError(t, err)

		deactivate, ok := decoded.Impl.(*DeactivateLookupTable)
		require.True(t, ok)

		require.NotNil(t, deactivate.GetLookupTableAccount())
		assert.Equal(t, lookupTable, deactivate.GetLookupTableAccount().PublicKey)

		require.NotNil(t, deactivate.GetAuthorityAccount())
		assert.Equal(t, authority, deactivate.GetAuthorityAccount().PublicKey)
	})

	t.Run("CloseLookupTable", func(t *testing.T) {
		lookupTable := solana.NewWallet().PublicKey()
		authority := solana.NewWallet().PublicKey()
		recipient := solana.NewWallet().PublicKey()

		inst := NewCloseLookupTableInstruction(lookupTable, authority, recipient)
		ix := inst.Build()

		data, err := ix.Data()
		require.NoError(t, err)

		decoded, err := DecodeInstruction(ix.Accounts(), data)
		require.NoError(t, err)

		close, ok := decoded.Impl.(*CloseLookupTable)
		require.True(t, ok)

		require.NotNil(t, close.GetLookupTableAccount())
		assert.Equal(t, lookupTable, close.GetLookupTableAccount().PublicKey)

		require.NotNil(t, close.GetAuthorityAccount())
		assert.Equal(t, authority, close.GetAuthorityAccount().PublicKey)

		require.NotNil(t, close.GetRecipientAccount())
		assert.Equal(t, recipient, close.GetRecipientAccount().PublicKey)
		assert.True(t, close.GetRecipientAccount().IsWritable)
	})

	t.Run("ExtendLookupTable", func(t *testing.T) {
		lookupTable := solana.NewWallet().PublicKey()
		authority := solana.NewWallet().PublicKey()
		payer := solana.NewWallet().PublicKey()
		addr1 := solana.NewWallet().PublicKey()
		addr2 := solana.NewWallet().PublicKey()

		inst := NewExtendLookupTableInstruction(lookupTable, authority, payer, []solana.PublicKey{addr1, addr2})
		ix := inst.Build()

		accounts := ix.Accounts()
		data, err := ix.Data()
		require.NoError(t, err)

		decoded, err := DecodeInstruction(accounts, data)
		require.NoError(t, err)

		ext, ok := decoded.Impl.(*ExtendLookupTable)
		require.True(t, ok)

		require.NotNil(t, ext.GetLookupTableAccount())
		assert.Equal(t, lookupTable, ext.GetLookupTableAccount().PublicKey)

		require.NotNil(t, ext.GetAuthorityAccount())
		assert.Equal(t, authority, ext.GetAuthorityAccount().PublicKey)

		require.NotNil(t, ext.GetPayerAccount())
		assert.Equal(t, payer, ext.GetPayerAccount().PublicKey)

		require.Len(t, ext.Addresses, 2)
		assert.Equal(t, addr1, ext.Addresses[0])
		assert.Equal(t, addr2, ext.Addresses[1])
	})
}

func ptrUint64(v uint64) *uint64 { return &v }
func ptrUint8(v uint8) *uint8    { return &v }
