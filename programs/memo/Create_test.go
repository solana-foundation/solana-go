package memo

import (
	"bytes"
	"testing"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ported from https://github.com/solana-program/memo/blob/main/program/src/processor.rs

func TestUTF8Memo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message []byte
		wantErr string
	}{
		{
			name:    "valid ASCII",
			message: []byte("letters and such"),
		},
		{
			name:    "valid emoji",
			message: []byte("🐆"),
		},
		{
			name:    "emoji bytes match expected encoding",
			message: []byte{0xF0, 0x9F, 0x90, 0x86}, // 🐆 in UTF-8
		},
		{
			name:    "invalid UTF-8",
			message: []byte{0xF0, 0x9F, 0x90, 0xFF},
			wantErr: "message is not valid UTF-8",
		},
		{
			name:    "empty message",
			message: []byte{},
			wantErr: "message not set",
		},
		{
			name:    "nil message",
			message: nil,
			wantErr: "message not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			inst := NewMemoInstructionBuilder().SetMessage(tt.message)
			err := inst.Validate()
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSigners(t *testing.T) {
	t.Parallel()

	pubkey0 := ag_solanago.NewWallet().PublicKey()
	pubkey1 := ag_solanago.NewWallet().PublicKey()
	pubkey2 := ag_solanago.NewWallet().PublicKey()
	memo := []byte("🐆")

	t.Run("all signed", func(t *testing.T) {
		t.Parallel()
		inst := NewMemoInstruction(memo, pubkey0, pubkey1, pubkey2)
		require.NoError(t, inst.Validate())
		require.Len(t, inst.AccountMetaSlice, 3)
		for _, acc := range inst.AccountMetaSlice {
			require.True(t, acc.IsSigner, "all accounts should be signers")
		}
	})

	t.Run("no signers (unsigned memo)", func(t *testing.T) {
		t.Parallel()
		inst := NewMemoInstruction(memo)
		require.NoError(t, inst.Validate())
		assert.Empty(t, inst.AccountMetaSlice)
	})

	t.Run("single signer", func(t *testing.T) {
		t.Parallel()
		inst := NewMemoInstruction(memo, pubkey0)
		require.NoError(t, inst.Validate())
		require.Len(t, inst.AccountMetaSlice, 1)
		assert.Equal(t, pubkey0, inst.AccountMetaSlice[0].PublicKey)
		assert.True(t, inst.AccountMetaSlice[0].IsSigner)
	})
}

func TestSetSignerReplaces(t *testing.T) {
	t.Parallel()

	pubkey0 := ag_solanago.NewWallet().PublicKey()
	pubkey1 := ag_solanago.NewWallet().PublicKey()

	inst := NewMemoInstructionBuilder().
		SetMessage([]byte("hello")).
		AddSigner(pubkey0).
		AddSigner(pubkey1)
	require.Len(t, inst.AccountMetaSlice, 2)

	// SetSigner replaces all signers with a single one
	pubkey2 := ag_solanago.NewWallet().PublicKey()
	inst.SetSigner(pubkey2)
	require.Len(t, inst.AccountMetaSlice, 1)
	assert.Equal(t, pubkey2, inst.AccountMetaSlice[0].PublicKey)
}

func TestGetSignerEmpty(t *testing.T) {
	t.Parallel()

	inst := NewMemoInstructionBuilder()
	assert.Nil(t, inst.GetSigner())

	pubkey := ag_solanago.NewWallet().PublicKey()
	inst.AddSigner(pubkey)
	require.NotNil(t, inst.GetSigner())
	assert.Equal(t, pubkey, inst.GetSigner().PublicKey)
}

func TestEncodeDecode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message []byte
	}{
		{name: "ascii", message: []byte("hello world")},
		{name: "emoji", message: []byte("🐆🦀🎉")},
		{name: "long message", message: bytes.Repeat([]byte("a"), 566)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			original := NewMemoInstructionBuilder().SetMessage(tt.message)

			buf := new(bytes.Buffer)
			err := ag_binary.NewBinEncoder(buf).Encode(original)
			require.NoError(t, err)

			decoded := new(Create)
			err = ag_binary.NewBinDecoder(buf.Bytes()).Decode(decoded)
			require.NoError(t, err)

			assert.Equal(t, original.Message, decoded.Message)
		})
	}
}

func TestEncodeDataIsRawBytes(t *testing.T) {
	t.Parallel()

	// The on-chain program reads instruction_data directly as the memo string —
	// no length prefix, no discriminator.
	message := []byte("hello memo")
	inst := NewMemoInstruction(message).Build()

	data, err := inst.Data()
	require.NoError(t, err)
	assert.Equal(t, message, data)
}

func FuzzUnmarshalMemo(f *testing.F) {
	f.Add([]byte("hello"))
	f.Add([]byte("🐆"))
	f.Add([]byte{0xFF, 0xFE})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		inst := new(Create)
		_ = inst.UnmarshalWithDecoder(ag_binary.NewBinDecoder(data))
	})
}

func TestValidateAndBuild(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		inst := NewMemoInstruction([]byte("test"))
		built, err := inst.ValidateAndBuild()
		require.NoError(t, err)
		require.NotNil(t, built)
	})

	t.Run("invalid empty message", func(t *testing.T) {
		t.Parallel()
		inst := NewMemoInstructionBuilder()
		_, err := inst.ValidateAndBuild()
		require.Error(t, err)
	})

	t.Run("invalid UTF-8", func(t *testing.T) {
		t.Parallel()
		inst := NewMemoInstructionBuilder().SetMessage([]byte{0xFF, 0xFE})
		_, err := inst.ValidateAndBuild()
		require.EqualError(t, err, "message is not valid UTF-8")
	})
}

func TestDecodeInstruction(t *testing.T) {
	t.Parallel()

	pubkey := ag_solanago.NewWallet().PublicKey()
	message := []byte("decode test")

	inst := NewMemoInstruction(message, pubkey).Build()
	data, err := inst.Data()
	require.NoError(t, err)

	accounts := inst.Accounts()
	decoded, err := DecodeInstruction(accounts, data)
	require.NoError(t, err)

	create, ok := decoded.Impl.(*Create)
	require.True(t, ok)
	assert.Equal(t, message, create.Message)
	require.Len(t, create.AccountMetaSlice, 1)
	assert.Equal(t, pubkey, create.AccountMetaSlice[0].PublicKey)
}

func TestProgramID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ag_solanago.MemoProgramID, ProgramID)

	inst := NewMemoInstruction([]byte("test")).Build()
	assert.Equal(t, ag_solanago.MemoProgramID, inst.ProgramID())
}
