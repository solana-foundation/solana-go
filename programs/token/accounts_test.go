package token

import (
	"bytes"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/stretchr/testify/require"
)

func TestMint(t *testing.T) {
	accountBytes := []byte{
		1, 0, 0, 0,
		5, 234, 156, 241, 108, 228, 17, 152, 241, 164, 153, 55, 200, 140, 55, 10, 148, 212, 175, 255, 137, 181, 186, 203, 142, 244, 94, 99, 36, 187, 120, 247,
		9, 169, 49, 235, 241, 182, 6, 0,
		6,
		1,
		1, 0, 0, 0,
		5, 234, 156, 241, 108, 228, 17, 152, 241, 164, 153, 55, 200, 140, 55, 10, 148, 212, 175, 255, 137, 181, 186, 203, 142, 244, 94, 99, 36, 187, 120, 247,
	}
	{
		dec := bin.NewBinDecoder(accountBytes)
		mint := Mint{}

		err := dec.Decode(&mint)
		require.NoError(t, err, spew.Sdump(mint))

		require.Equal(t,
			&Mint{
				MintAuthority:   solana.MustPublicKeyFromBase58("Q6XprfkF8RQQKoQVG33xT88H7wi8Uk1B1CC7YAs69Gi").ToPointer(),
				Supply:          1890000009537801,
				Decimals:        6,
				IsInitialized:   true,
				FreezeAuthority: solana.MustPublicKeyFromBase58("Q6XprfkF8RQQKoQVG33xT88H7wi8Uk1B1CC7YAs69Gi").ToPointer(),
			},
			&mint,
		)

		{
			buf := new(bytes.Buffer)
			err := bin.NewBinEncoder(buf).Encode(mint)
			require.NoError(t, err)
			require.Equal(t, accountBytes, buf.Bytes(), bin.FormatByteSlice(buf.Bytes()))
		}
	}
}

func TestAccount(t *testing.T) {
	accountBytes := []byte{
		6, 155, 136, 87, 254, 171, 129, 132, 251, 104, 127, 99, 70, 24, 192, 53, 218, 196, 57, 220, 26, 235, 59, 85, 152, 160, 240, 0, 0, 0, 0, 1,
		93, 100, 62, 133, 31, 102, 235, 161, 170, 152, 161, 7, 39, 223, 9, 180, 1, 224, 134, 204, 54, 241, 9, 195, 240, 147, 219, 146, 35, 92, 26, 224,
		42, 34, 176, 1, 0, 0, 0, 0,

		0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,

		1,
		1, 0, 0, 0,
		240, 29, 31, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,

		0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	{
		dec := bin.NewBinDecoder(accountBytes)
		account := Account{}

		err := dec.Decode(&account)
		require.NoError(t, err, spew.Sdump(account))

		balance := uint64(2039280)
		require.Equal(t,
			&Account{
				Mint:            solana.WrappedSol,
				Owner:           solana.MustPublicKeyFromBase58("7HZaCWazgTuuFuajxaaxGYbGnyVKwxvsJKue1W4Nvyro"),
				Amount:          (uint64)(28320298),
				Delegate:        (*solana.PublicKey)(nil),
				State:           (AccountState)(1),
				IsNative:        (*uint64)(&balance),
				DelegatedAmount: (uint64)(0),
				CloseAuthority:  (*solana.PublicKey)(nil),
			},
			&account,
		)

		{
			buf := bin.NewWriteByWrite("")
			err := bin.NewBinEncoder(buf).Encode(account)
			require.NoError(t, err)
			require.Equal(t, accountBytes, buf.Bytes(), bin.FormatByteSlice(buf.Bytes()))
		}
	}
}
