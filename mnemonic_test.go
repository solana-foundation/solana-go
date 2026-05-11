// Copyright 2026 github.com/gagliardetto
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
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSlip10Vectors covers the official SLIP-0010 ed25519 test vectors from
// https://github.com/satoshilabs/slips/blob/master/slip-0010.md, exercising
// the master + hardened child derivation in isolation from BIP-39.
func TestSlip10Vectors(t *testing.T) {
	t.Run("vector 1", func(t *testing.T) {
		seed, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
		cases := []struct {
			path string
			key  string
		}{
			{"m", "2b4be7f19ee27bbf30c667b642d5f4aa69fd169872f8fc3059c08ebae2eb19e7"},
			{"m/0'", "68e0fe46dfb67e368c75379acec591dad19df3cde26e63b93a8e704f1dade7a3"},
			{"m/0'/1'", "b1d0bad404bf35da785a64ca1ac54b2617211d2777696fbffaf208f746ae84f2"},
			{"m/0'/1'/2'", "92a5b23c0b8a99e37d07df3fb9966917f5d06e02ddbd909c7e184371463e9fc9"},
			{"m/0'/1'/2'/2'", "30d1dc7e5fc04c31219ab25a27ae00b50f6fd66622f6e9c913253d6511d1e662"},
			{"m/0'/1'/2'/2'/1000000000'", "8f94d394a8e8fd6b1bc2f3f49f5c47e385281d5c17e65324b0f62483e37e8793"},
		}
		for _, tc := range cases {
			t.Run(tc.path, func(t *testing.T) {
				pk, err := PrivateKeyFromSeedAtPath(seed, tc.path)
				require.NoError(t, err)
				// PrivateKey is seed(32) || pubkey(32); compare the seed half.
				assert.Equal(t, tc.key, hex.EncodeToString(pk[:32]))
			})
		}
	})

	t.Run("vector 2", func(t *testing.T) {
		seed, _ := hex.DecodeString("fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542")
		cases := []struct {
			path string
			key  string
		}{
			{"m", "171cb88b1b3c1db25add599712e36245d75bc65a1a5c9e18d76f9f2b1eab4012"},
			{"m/0'", "1559eb2bbec5790b0c65d8693e4d0875b1747f4970ae8b650486ed7470845635"},
		}
		for _, tc := range cases {
			t.Run(tc.path, func(t *testing.T) {
				pk, err := PrivateKeyFromSeedAtPath(seed, tc.path)
				require.NoError(t, err)
				assert.Equal(t, tc.key, hex.EncodeToString(pk[:32]))
			})
		}
	})
}

func TestParseDerivationPath(t *testing.T) {
	t.Run("empty path returns no indices", func(t *testing.T) {
		for _, p := range []string{"", "m", "/"} {
			indices, err := parseDerivationPath(p)
			require.NoError(t, err)
			assert.Empty(t, indices)
		}
	})

	t.Run("solana default path", func(t *testing.T) {
		indices, err := parseDerivationPath(SolanaDerivationPath)
		require.NoError(t, err)
		require.Len(t, indices, 4)
		assert.Equal(t, slip10HardenedOffset+44, indices[0])
		assert.Equal(t, slip10HardenedOffset+501, indices[1])
		assert.Equal(t, slip10HardenedOffset+0, indices[2])
		assert.Equal(t, slip10HardenedOffset+0, indices[3])
	})

	t.Run("h and H suffixes are accepted", func(t *testing.T) {
		a, err := parseDerivationPath("m/44'/501'/0'/0'")
		require.NoError(t, err)
		b, err := parseDerivationPath("m/44h/501h/0h/0h")
		require.NoError(t, err)
		c, err := parseDerivationPath("m/44H/501H/0H/0H")
		require.NoError(t, err)
		assert.Equal(t, a, b)
		assert.Equal(t, a, c)
	})

	t.Run("non-hardened segments are rejected", func(t *testing.T) {
		_, err := parseDerivationPath("m/44'/501'/0'/0")
		assert.Error(t, err)
	})

	t.Run("malformed paths are rejected", func(t *testing.T) {
		for _, p := range []string{"m//0'", "m/abc'", "m/4294967296'"} {
			_, err := parseDerivationPath(p)
			assert.Errorf(t, err, "expected error parsing %q", p)
		}
	})
}

func TestPrivateKeyFromMnemonic(t *testing.T) {
	t.Run("invalid mnemonic returns error", func(t *testing.T) {
		_, err := PrivateKeyFromMnemonic("not a real mnemonic", "")
		require.Error(t, err)
	})

	t.Run("valid mnemonic derives a usable signing key", func(t *testing.T) {
		mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
		pk, err := PrivateKeyFromMnemonic(mnemonic, "")
		require.NoError(t, err)
		require.NoError(t, pk.Validate())

		// Same inputs must always produce the same key.
		pk2, err := PrivateKeyFromMnemonic(mnemonic, "")
		require.NoError(t, err)
		assert.Equal(t, pk, pk2)

		// A different passphrase must produce a different key.
		pk3, err := PrivateKeyFromMnemonic(mnemonic, "different")
		require.NoError(t, err)
		assert.NotEqual(t, pk, pk3)

		// A different path must produce a different key.
		pk4, err := PrivateKeyFromMnemonicAtPath(mnemonic, "", "m/44'/501'/1'/0'")
		require.NoError(t, err)
		assert.NotEqual(t, pk, pk4)

		// The derived key must produce verifiable signatures.
		sig, err := pk.Sign([]byte("test payload"))
		require.NoError(t, err)
		assert.True(t, pk.PublicKey().Verify([]byte("test payload"), sig))
	})
}

func TestNewWalletFromMnemonic(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	w, err := NewWalletFromMnemonic(mnemonic, "")
	require.NoError(t, err)
	require.NotNil(t, w)

	pk, err := PrivateKeyFromMnemonic(mnemonic, "")
	require.NoError(t, err)
	assert.Equal(t, pk, w.PrivateKey)
	assert.Equal(t, pk.PublicKey(), w.PublicKey())
}
