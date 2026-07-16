package vault

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassphraseBoxer_SealOpen(t *testing.T) {
	boxer := NewPassphraseBoxer("secret")

	sealed, err := boxer.Seal([]byte("hello"))
	require.NoError(t, err)

	opened, err := boxer.Open(sealed)
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), opened)
}

func TestPassphraseBoxer_OpenWrongPassphrase(t *testing.T) {
	sealed, err := NewPassphraseBoxer("secret").Seal([]byte("hello"))
	require.NoError(t, err)

	_, err = NewPassphraseBoxer("wrong").Open(sealed)
	require.Error(t, err)
}

func TestPassphraseBoxer_OpenShortCiphertext(t *testing.T) {
	boxer := NewPassphraseBoxer("secret")

	// Anything shorter than the salt and nonce prefix cannot be a sealed
	// payload, and must be reported rather than sliced past the buffer.
	for _, size := range []int{0, 1, saltLength, saltLength + nonceLength - 1} {
		in := base64.RawStdEncoding.EncodeToString(make([]byte, size))

		_, err := boxer.Open(in)
		require.Error(t, err, "expected error for %d-byte ciphertext", size)
	}
}

func TestPassphraseBoxer_OpenInvalidBase64(t *testing.T) {
	_, err := NewPassphraseBoxer("secret").Open("!!!not-base64!!!")
	require.Error(t, err)
}
