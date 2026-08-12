package base58_test

import (
	"errors"
	"testing"

	fluxbase58 "github.com/fluxrpc/base58"
	base58 "github.com/gagliardetto/solana-go/base58"
	"github.com/stretchr/testify/require"
)

func TestCompatibilityAPI(t *testing.T) {
	var value32 [32]byte
	for i := range value32 {
		value32[i] = byte(i)
	}

	encoded32 := base58.Encode32(&value32)
	require.Equal(t, fluxbase58.Encode32(&value32), encoded32)
	require.Equal(t, encoded32, base58.Encode(value32[:]))
	require.Equal(t, "prefix:"+encoded32, string(base58.AppendEncode32([]byte("prefix:"), &value32)))

	decoded, err := base58.Decode(encoded32)
	require.NoError(t, err)
	require.Equal(t, value32[:], decoded)

	var decoded32 [32]byte
	require.NoError(t, base58.Decode32(encoded32, &decoded32))
	require.Equal(t, value32, decoded32)

	var value64 [64]byte
	for i := range value64 {
		value64[i] = byte(i)
	}

	encoded64 := base58.Encode64(&value64)
	require.Equal(t, fluxbase58.Encode64(&value64), encoded64)
	require.Equal(t, encoded64, base58.Encode(value64[:]))
	require.Equal(t, "prefix:"+encoded64, string(base58.AppendEncode64([]byte("prefix:"), &value64)))

	var decoded64 [64]byte
	require.NoError(t, base58.Decode64(encoded64, &decoded64))
	require.Equal(t, value64, decoded64)
}

func TestCompatibilityConstantsAndErrors(t *testing.T) {
	require.Equal(t, fluxbase58.EncodedMaxLen32, base58.EncodedMaxLen32)
	require.Equal(t, fluxbase58.EncodedMaxLen64, base58.EncodedMaxLen64)

	require.True(t, base58.ErrInvalidChar == fluxbase58.ErrInvalidChar)
	require.True(t, base58.ErrInvalidLength == fluxbase58.ErrInvalidLength)
	require.True(t, base58.ErrValueTooLarge == fluxbase58.ErrValueTooLarge)
	require.True(t, base58.ErrLeadingZeros == fluxbase58.ErrLeadingZeros)

	_, err := base58.Decode("0")
	require.True(t, errors.Is(err, base58.ErrInvalidChar))

	var dst [32]byte
	err = base58.Decode32("", &dst)
	require.True(t, errors.Is(err, base58.ErrInvalidLength))
}
