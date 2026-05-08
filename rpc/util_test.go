package rpc

import (
	"bytes"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// Layout references for the test data below:
//
// SPL Token (solana-program/token):
//
//	Mint::LEN    = 82
//	Account::LEN = 165
//
// Token-2022 (solana-program/token-2022):
//
//	Extended records place a 1-byte AccountType discriminator at offset
//	Account::LEN (= 165). Mint base (82 bytes) is padded with 83 zeros so
//	Mint and Account share the discriminator offset.
//
//	AccountType::Uninitialized = 0
//	AccountType::Mint          = 1
//	AccountType::Account       = 2
//
//	Extensions follow as TLV: [u16 LE type][u16 LE length][value...].
const (
	testAccountTypeUninitialized uint8 = 0
	testAccountTypeMint          uint8 = 1
	testAccountTypeAccount       uint8 = 2

	// Real Token-2022 extension type numbers used in the realistic TLV
	// fixtures below.
	testExtTypeMintCloseAuthority uint16 = 3
	testExtTypeImmutableOwner     uint16 = 7
)

func mkAccount(owner solana.PublicKey, data []byte) *Account {
	return &Account{
		Owner: owner,
		Data:  DataBytesOrJSONFromBytes(data),
	}
}

// mkToken2022ExtensionData builds a synthetic Token-2022 extended account
// record: base (padded to 165 bytes) + 1-byte AccountType discriminator +
// one TLV entry.
func mkToken2022ExtensionData(accountType uint8, extType uint16, extValue []byte) []byte {
	data := make([]byte, 0, token2022AccountBaseSize+1+4+len(extValue))
	data = append(data, bytes.Repeat([]byte{0}, token2022AccountBaseSize)...)
	data = append(data, accountType)
	data = append(data,
		byte(extType), byte(extType>>8),
		byte(len(extValue)), byte(len(extValue)>>8),
	)
	data = append(data, extValue...)
	return data
}

func TestIsTokenMint(t *testing.T) {
	// A 32-byte pubkey payload used as an extension value.
	pubkeyValue := bytes.Repeat([]byte{0xAB}, 32)

	tests := []struct {
		name string
		acc  *Account
		want bool
	}{
		// --- nil / empty handling ---
		{
			name: "nil account",
			acc:  nil,
			want: false,
		},
		{
			name: "nil Data field",
			acc:  &Account{Owner: solana.TokenProgramID},
			want: false,
		},
		{
			name: "empty data, Token owner",
			acc:  mkAccount(solana.TokenProgramID, nil),
			want: false,
		},
		{
			name: "empty data, Token-2022 owner",
			acc:  mkAccount(solana.Token2022ProgramID, nil),
			want: false,
		},

		// --- wrong owner ---
		{
			name: "SystemProgram owner, mint-sized data",
			acc:  mkAccount(solana.SystemProgramID, make([]byte, 82)),
			want: false,
		},
		{
			name: "random owner, mint-sized data",
			acc:  mkAccount(solana.MustPublicKeyFromBase58("11111111111111111111111111111112"), make([]byte, 82)),
			want: false,
		},

		// --- SPL Token ---
		{
			name: "SPL Token: exact Mint::LEN = 82",
			acc:  mkAccount(solana.TokenProgramID, make([]byte, 82)),
			want: true,
		},
		{
			name: "SPL Token: 81 bytes (one short of Mint::LEN)",
			acc:  mkAccount(solana.TokenProgramID, make([]byte, 81)),
			want: false,
		},
		{
			name: "SPL Token: 83 bytes (one past Mint::LEN)",
			acc:  mkAccount(solana.TokenProgramID, make([]byte, 83)),
			want: false,
		},
		{
			name: "SPL Token: Account::LEN = 165 (a token account, not a mint)",
			acc:  mkAccount(solana.TokenProgramID, make([]byte, 165)),
			want: false,
		},
		{
			name: "SPL Token: extension-shaped data is not valid for classic Token",
			acc: mkAccount(solana.TokenProgramID,
				mkToken2022ExtensionData(testAccountTypeMint, testExtTypeMintCloseAuthority, pubkeyValue)),
			want: false,
		},

		// --- Token-2022 bare (no extensions) ---
		{
			name: "Token-2022: bare Mint (82 bytes)",
			acc:  mkAccount(solana.Token2022ProgramID, make([]byte, 82)),
			want: true,
		},
		{
			name: "Token-2022: bare Account (165 bytes, no discriminator)",
			acc:  mkAccount(solana.Token2022ProgramID, make([]byte, 165)),
			want: false,
		},

		// --- Token-2022 invalid intermediate sizes ---
		{
			name: "Token-2022: 83 bytes (between Mint and Account sizes)",
			acc:  mkAccount(solana.Token2022ProgramID, make([]byte, 83)),
			want: false,
		},
		{
			name: "Token-2022: 164 bytes (one short of Account::LEN)",
			acc:  mkAccount(solana.Token2022ProgramID, make([]byte, 164)),
			want: false,
		},

		// --- Token-2022 extended record discriminator cases ---
		{
			name: "Token-2022: extended with AccountType=Mint (1) at offset 165",
			acc: mkAccount(solana.Token2022ProgramID,
				mkToken2022ExtensionData(testAccountTypeMint, testExtTypeMintCloseAuthority, pubkeyValue)),
			want: true,
		},
		{
			name: "Token-2022: extended with AccountType=Account (2) at offset 165",
			acc: mkAccount(solana.Token2022ProgramID,
				mkToken2022ExtensionData(testAccountTypeAccount, testExtTypeImmutableOwner, nil)),
			want: false,
		},
		{
			name: "Token-2022: extended with AccountType=Uninitialized (0) at offset 165",
			acc: mkAccount(solana.Token2022ProgramID,
				mkToken2022ExtensionData(testAccountTypeUninitialized, 0, nil)),
			want: false,
		},
		{
			name: "Token-2022: extended with unknown discriminator (3)",
			acc: mkAccount(solana.Token2022ProgramID,
				mkToken2022ExtensionData(3, 0, nil)),
			want: false,
		},
		{
			name: "Token-2022: 166 bytes, discriminator byte = Mint, no TLV payload",
			acc: mkAccount(solana.Token2022ProgramID,
				append(make([]byte, 165), testAccountTypeMint)),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, IsTokenMint(tt.acc))
		})
	}
}

// TestIsTokenMint_RealisticExtendedMintSize anchors the Token-2022 mint-
// with-extensions layout at the same byte length the Token-2022 sub-
// package already exercises in extension_test.go (a MintCloseAuthority
// extension produces a 202-byte record).
func TestIsTokenMint_RealisticExtendedMintSize(t *testing.T) {
	data := mkToken2022ExtensionData(
		testAccountTypeMint,
		testExtTypeMintCloseAuthority,
		bytes.Repeat([]byte{1}, 32),
	)
	require.Equal(t, 202, len(data), "MintCloseAuthority TLV should produce a 202-byte record")
	require.True(t, IsTokenMint(mkAccount(solana.Token2022ProgramID, data)))
}
