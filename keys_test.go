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
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublicKeyFromBytes(t *testing.T) {
	tests := []struct {
		name     string
		inHex    string
		isErr    bool
		expected PublicKey
	}{
		{
			"empty",
			"",
			true,
			PublicKey{},
		},
		{
			"smaller than required",
			"010203040506",
			true,
			PublicKey{},
		},
		{
			"equal to 32 bytes",
			"0102030405060102030405060102030405060102030405060102030405060101",
			false,
			MustPublicKeyFromBase58("4wBqpZM9msxygzsdeLPq6Zw3LoiAxJk3GjtKPpqkcsi"),
		},
		{
			"longer than required",
			"0102030405060102030405060102030405060102030405060102030405060101FFFFFFFFFF",
			true,
			PublicKey{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bytes, err := hex.DecodeString(test.inHex)
			require.NoError(t, err)

			if test.isErr {
				require.Panics(t, func() {
					actual := PublicKeyFromBytes(bytes)
					require.Equal(t, zeroPublicKey, actual)
				})
			} else {
				actual := PublicKeyFromBytes(bytes)
				require.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestPublicKeyFromBase58(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		expected    PublicKey
		expectedErr error
	}{
		{
			"hand crafted",
			"SerumkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
			MustPublicKeyFromBase58("SerumkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
			nil,
		},
		{
			"hand crafted error",
			"SerkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
			zeroPublicKey,
			errors.New("invalid length, expected 32, got 30"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := PublicKeyFromBase58(test.in)
			if test.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, test.expected, actual)
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func TestPrivateKeyFromSolanaKeygenFile(t *testing.T) {
	tests := []struct {
		inFile      string
		expected    PrivateKey
		expectedPub PublicKey
		expectedErr error
	}{
		{
			"testdata/standard.solana-keygen.json",
			MustPrivateKeyFromBase58("66cDvko73yAf8LYvFMM3r8vF5vJtkk7JKMgEKwkmBC86oHdq41C7i1a2vS3zE1yCcdLLk6VUatUb32ZzVjSBXtRs"),
			MustPublicKeyFromBase58("F8UvVsKnzWyp2nF8aDcqvQ2GVcRpqT91WDsAtvBKCMt9"),
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.inFile, func(t *testing.T) {
			actual, err := PrivateKeyFromSolanaKeygenFile(test.inFile)
			if test.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, test.expected, actual)
				assert.Equal(t, test.expectedPub, actual.PublicKey(), "%s != %s", test.expectedPub, actual.PublicKey())

			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func TestPrivateKeyFromBase58RejectsMismatchedSeedAndPublicKey(t *testing.T) {
	original := MustPrivateKeyFromBase58("66cDvko73yAf8LYvFMM3r8vF5vJtkk7JKMgEKwkmBC86oHdq41C7i1a2vS3zE1yCcdLLk6VUatUb32ZzVjSBXtRs")
	require.Len(t, original, PrivateKeyLength)

	tampered := append([]byte(nil), original...)
	tampered[0] ^= 0xFF

	valid, err := ValidatePrivateKey(tampered)
	require.False(t, valid)
	require.EqualError(t, err, "invalid private key: seed/public key mismatch")

	_, err = PrivateKeyFromBase58(PrivateKey(tampered).String())
	require.EqualError(t, err, "invalid private key: seed/public key mismatch")
}

func TestPrivateKeyFromBase58ReturnsDiagnosticForOffCurvePublicKey(t *testing.T) {
	original := MustPrivateKeyFromBase58("66cDvko73yAf8LYvFMM3r8vF5vJtkk7JKMgEKwkmBC86oHdq41C7i1a2vS3zE1yCcdLLk6VUatUb32ZzVjSBXtRs")
	require.Len(t, original, PrivateKeyLength)

	tampered := append([]byte(nil), original...)
	offCurve := make([]byte, ed25519.PublicKeySize)
	found := false
	for i := uint32(0); i < 100_000; i++ {
		binary.LittleEndian.PutUint32(offCurve[:4], i)
		if !IsOnCurve(offCurve) {
			found = true
			break
		}
	}
	require.True(t, found, "expected to find an off-curve public key test vector")
	copy(tampered[ed25519.SeedSize:], offCurve)

	valid, err := ValidatePrivateKey(tampered)
	require.False(t, valid)
	require.EqualError(t, err, "invalid private key: seed/public key mismatch (provided public key is NOT on the ed25519 curve)")
}

func TestPublicKey_MarshalText(t *testing.T) {
	keyString := "4wBqpZM9k69W87zdYXT2bRtLViWqTiJV3i2Kn9q7S6j"
	keyParsed := MustPublicKeyFromBase58(keyString)

	var key PublicKey
	err := key.UnmarshalText([]byte(keyString))
	require.NoError(t, err)

	assert.True(t, keyParsed.Equals(key))

	keyText, err := key.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, []byte(keyString), keyText)

	type IdentityToSlotsBlocks map[PublicKey][2]int64

	var payload IdentityToSlotsBlocks
	data := `{"` + keyString + `":[3,4]}`
	err = json.Unmarshal([]byte(data), &payload)
	require.NoError(t, err)

	assert.Equal(t,
		IdentityToSlotsBlocks{
			keyParsed: [2]int64{3, 4},
		},
		payload,
	)
}

func TestPublicKey_Flag(t *testing.T) {
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	var key PublicKey
	flagSet.Var(&key, "account", "Public key")
	err := flagSet.Parse([]string{"--account", "7cVfgArCheMR6Cs4t6vz5rfnqd56vZq4ndaBrY5xkxXy"})
	require.NoError(t, err)
	assert.Equal(t, PublicKey{
		0x62, 0x3d, 0xdd, 0x11, 0x7e, 0x7c, 0xc5, 0x62,
		0xf6, 0x63, 0x15, 0x05, 0x25, 0x8c, 0xd1, 0xdc,
		0xee, 0x81, 0x94, 0x9f, 0x8a, 0xfd, 0x1e, 0xa2,
		0x94, 0xdc, 0x47, 0xbe, 0x6e, 0xcf, 0xf3, 0xa8,
	}, key)
}

func TestPublicKeySlice(t *testing.T) {
	{
		slice := make(PublicKeySlice, 0)
		require.False(t, slice.Has(BPFLoaderProgramID))

		slice.Append(BPFLoaderProgramID)
		require.True(t, slice.Has(BPFLoaderProgramID))
		require.Len(t, slice, 1)

		slice.UniqueAppend(BPFLoaderProgramID)
		require.Len(t, slice, 1)
		slice.Append(ConfigProgramID)
		require.Len(t, slice, 2)
		require.True(t, slice.Has(ConfigProgramID))
	}

	{
		slice := make(PublicKeySlice, 0)
		{
			require.Equal(t, []PublicKeySlice{}, slice.Split(1))
		}
		slice.Append(
			SysVarRentPubkey,
			SysVarRewardsPubkey,
		)
		{
			require.Equal(t,
				[]PublicKeySlice{},
				slice.Split(0),
			)
			require.Equal(t,
				[]PublicKeySlice{},
				slice.Split(-333),
			)
		}
		{
			require.Equal(t,
				[]PublicKeySlice{
					{SysVarRentPubkey},
					{SysVarRewardsPubkey},
				},
				slice.Split(1),
			)
		}
		slice.Append(
			BPFLoaderProgramID,
			BPFLoaderDeprecatedProgramID,
			FeatureProgramID,
			ConfigProgramID,
			StakeProgramID,
			VoteProgramID,
			SystemProgramID,
		)
		{
			require.Equal(t,
				[]PublicKeySlice{
					{SysVarRentPubkey},
					{SysVarRewardsPubkey},
					{BPFLoaderProgramID},
					{BPFLoaderDeprecatedProgramID},
					{FeatureProgramID},
					{ConfigProgramID},
					{StakeProgramID},
					{VoteProgramID},
					{SystemProgramID},
				},
				slice.Split(1),
			)
		}
		{
			require.Equal(t,
				[]PublicKeySlice{
					{SysVarRentPubkey, SysVarRewardsPubkey},
					{BPFLoaderProgramID, BPFLoaderDeprecatedProgramID},
					{FeatureProgramID, ConfigProgramID},
					{StakeProgramID, VoteProgramID},
					{SystemProgramID},
				},
				slice.Split(2),
			)
		}
	}
}

func TestGetAddedRemovedPubkeys(t *testing.T) {
	{
		previous := PublicKeySlice{}
		next := PublicKeySlice{BPFLoaderProgramID}

		added, removed := GetAddedRemovedPubkeys(previous, next)
		require.Equal(t,
			PublicKeySlice{BPFLoaderProgramID},
			added,
		)
		require.Equal(t,
			PublicKeySlice{},
			removed,
		)
	}
	{
		previous := PublicKeySlice{
			SysVarClockPubkey,
			SysVarEpochSchedulePubkey,
			SysVarFeesPubkey,
			SysVarInstructionsPubkey,
			SysVarRecentBlockHashesPubkey,
		}
		next := PublicKeySlice{
			SysVarClockPubkey,
			SysVarEpochSchedulePubkey,
			SysVarFeesPubkey,
			SysVarInstructionsPubkey,
			SysVarRecentBlockHashesPubkey,
		}

		added, removed := GetAddedRemovedPubkeys(previous, next)
		require.Equal(t,
			PublicKeySlice{},
			added,
		)
		require.Equal(t,
			PublicKeySlice{},
			removed,
		)
	}
	{
		previous := PublicKeySlice{
			SysVarClockPubkey,
			SysVarEpochSchedulePubkey,
			SysVarFeesPubkey,
			SysVarInstructionsPubkey,
			SysVarRecentBlockHashesPubkey,
		}
		next := PublicKeySlice{
			SysVarEpochSchedulePubkey,
			SysVarFeesPubkey,
			SysVarInstructionsPubkey,
			SysVarRecentBlockHashesPubkey,
			ConfigProgramID,
		}

		added, removed := GetAddedRemovedPubkeys(previous, next)
		require.Equal(t,
			PublicKeySlice{ConfigProgramID},
			added,
		)
		require.Equal(t,
			PublicKeySlice{SysVarClockPubkey},
			removed,
		)
	}
}

func TestGetAddedRemoved(t *testing.T) {
	{
		previous := PublicKeySlice{}
		next := PublicKeySlice{BPFLoaderProgramID}

		added, removed := previous.GetAddedRemoved(next)
		require.Equal(t,
			PublicKeySlice{BPFLoaderProgramID},
			added,
		)
		require.Equal(t,
			PublicKeySlice{},
			removed,
		)
	}
	{
		previous := PublicKeySlice{
			SysVarClockPubkey,
			SysVarEpochSchedulePubkey,
			SysVarFeesPubkey,
			SysVarInstructionsPubkey,
			SysVarRecentBlockHashesPubkey,
		}
		next := PublicKeySlice{
			SysVarClockPubkey,
			SysVarEpochSchedulePubkey,
			SysVarFeesPubkey,
			SysVarInstructionsPubkey,
			SysVarRecentBlockHashesPubkey,
		}

		added, removed := previous.GetAddedRemoved(next)
		require.Equal(t,
			PublicKeySlice{},
			added,
		)
		require.Equal(t,
			PublicKeySlice{},
			removed,
		)
	}
	{
		previous := PublicKeySlice{
			SysVarClockPubkey,
			SysVarEpochSchedulePubkey,
			SysVarFeesPubkey,
			SysVarInstructionsPubkey,
			SysVarRecentBlockHashesPubkey,
		}
		next := PublicKeySlice{
			SysVarEpochSchedulePubkey,
			SysVarFeesPubkey,
			SysVarInstructionsPubkey,
			SysVarRecentBlockHashesPubkey,
			ConfigProgramID,
		}

		added, removed := previous.GetAddedRemoved(next)
		require.Equal(t,
			PublicKeySlice{ConfigProgramID},
			added,
		)
		require.Equal(t,
			PublicKeySlice{SysVarClockPubkey},
			removed,
		)
	}
}

func TestIsNativeProgramID(t *testing.T) {
	require.True(t, isNativeProgramID(ConfigProgramID))
}

func TestCreateWithSeed(t *testing.T) {
	{
		got, err := CreateWithSeed(PublicKey{}, "limber chicken: 4/45", PublicKey{})
		require.NoError(t, err)
		require.True(t, got.Equals(MustPublicKeyFromBase58("9h1HyLCW5dZnBVap8C5egQ9Z6pHyjsh5MNy83iPqqRuq")))
	}
}

func TestCreateProgramAddressFromRust(t *testing.T) {
	// Ported from https://github.com/solana-labs/solana/blob/f32216588dfdbc7a7160c26331ce657a90f95ae7/sdk/program/src/pubkey.rs#L636
	program_id := MustPublicKeyFromBase58("BPFLoaderUpgradeab1e11111111111111111111111")
	public_key := MustPublicKeyFromBase58("SeedPubey1111111111111111111111111111111111")

	{
		got, err := CreateProgramAddress([][]byte{
			{},
			{1},
		},
			program_id,
		)
		require.NoError(t, err)
		require.True(t, got.Equals(MustPublicKeyFromBase58("BwqrghZA2htAcqq8dzP1WDAhTXYTYWj7CHxF5j7TDBAe")))
	}

	{
		got, err := CreateProgramAddress([][]byte{
			[]byte("☉"),
			{0},
		},
			program_id,
		)
		require.NoError(t, err)
		require.True(t, got.Equals(MustPublicKeyFromBase58("13yWmRpaTR4r5nAktwLqMpRNr28tnVUZw26rTvPSSB19")))
	}

	{
		got, err := CreateProgramAddress([][]byte{
			[]byte("Talking"),
			[]byte("Squirrels"),
		},
			program_id,
		)
		require.NoError(t, err)
		require.True(t, got.Equals(MustPublicKeyFromBase58("2fnQrngrQT4SeLcdToJAD96phoEjNL2man2kfRLCASVk")))
	}

	{
		got, err := CreateProgramAddress([][]byte{
			public_key[:],
			{1},
		},
			program_id,
		)
		require.NoError(t, err)
		require.True(t, got.Equals(MustPublicKeyFromBase58("976ymqVnfE32QFe6NfGDctSvVa36LWnvYxhU6G2232YL")))
	}
}

func TestCreateProgramAddressFromTypescript(t *testing.T) {
	t.Run(
		"createProgramAddress",
		// Ported from https://github.com/solana-labs/solana-web3.js/blob/168d5e088edd48f9f0c1a877e888592ca4cfdf38/test/publickey.test.ts#L113
		func(t *testing.T) {
			program_id := MustPublicKeyFromBase58("BPFLoader1111111111111111111111111111111111")
			public_key := MustPublicKeyFromBase58("SeedPubey1111111111111111111111111111111111")

			{
				programAddress, err := CreateProgramAddress([][]byte{
					[]byte(""),
					{1},
				},
					program_id,
				)
				require.NoError(t, err)
				require.True(t, programAddress.Equals(MustPublicKeyFromBase58("3gF2KMe9KiC6FNVBmfg9i267aMPvK37FewCip4eGBFcT")))
			}
			{
				programAddress, err := CreateProgramAddress([][]byte{
					[]byte("☉"),
				},
					program_id,
				)
				require.NoError(t, err)
				require.True(t, programAddress.Equals(MustPublicKeyFromBase58("7ytmC1nT1xY4RfxCV2ZgyA7UakC93do5ZdyhdF3EtPj7")))
			}
			{
				programAddress, err := CreateProgramAddress([][]byte{
					[]byte("Talking"),
					[]byte("Squirrels"),
				},
					program_id,
				)
				require.NoError(t, err)
				require.True(t, programAddress.Equals(MustPublicKeyFromBase58("HwRVBufQ4haG5XSgpspwKtNd3PC9GM9m1196uJW36vds")))
			}
			{
				programAddress, err := CreateProgramAddress([][]byte{
					public_key[:],
				},
					program_id,
				)
				require.NoError(t, err)
				require.True(t, programAddress.Equals(MustPublicKeyFromBase58("GUs5qLUfsEHkcMB9T38vjr18ypEhRuNWiePW2LoK4E3K")))

				{
					programAddress2, err := CreateProgramAddress([][]byte{
						[]byte("Talking"),
					},
						program_id,
					)
					require.NoError(t, err)
					require.False(t, programAddress.Equals(programAddress2))
				}
			}
			{
				_, err := CreateProgramAddress([][]byte{
					make([]byte, MaxSeedLength+1),
				},
					program_id,
				)
				require.EqualError(t, err, ErrMaxSeedLengthExceeded.Error())
			}
			{
				bn := make([]byte, 8)
				binary.LittleEndian.PutUint64(bn, 2)
				programAddress, err := CreateProgramAddress([][]byte{
					MustPublicKeyFromBase58("H4snTKK9adiU15gP22ErfZYtro3aqR9BTMXiH3AwiUTQ").Bytes(),
					bn,
				},
					MustPublicKeyFromBase58("4ckmDgGdxQoPDLUkDT3vHgSAkzA3QRdNq5ywwY4sUSJn"),
				)
				require.NoError(t, err)
				require.True(t, programAddress.Equals(MustPublicKeyFromBase58("12rqwuEgBYiGhBrDJStCiqEtzQpTTiZbh7teNVLuYcFA")))
			}
		},
	)

	t.Run(
		"findProgramAddress",
		// Ported from https://github.com/solana-labs/solana-web3.js/blob/168d5e088edd48f9f0c1a877e888592ca4cfdf38/test/publickey.test.ts#L194
		func(t *testing.T) {
			programId := MustPublicKeyFromBase58("BPFLoader1111111111111111111111111111111111")

			programAddress, nonce, err := FindProgramAddress(
				[][]byte{
					[]byte(""),
				},
				programId,
			)
			require.NoError(t, err)

			{
				got, err := CreateProgramAddress([][]byte{
					[]byte(""),
					{nonce},
				},
					programId,
				)
				require.NoError(t, err)
				require.True(t, programAddress.Equals(got))
			}
		},
	)

	t.Run(
		"isOnCurve",
		// Ported from https://github.com/solana-labs/solana-web3.js/blob/168d5e088edd48f9f0c1a877e888592ca4cfdf38/test/publickey.test.ts#L212
		func(t *testing.T) {
			onCurve := NewWallet().PublicKey()
			require.True(t, onCurve.IsOnCurve())

			// A program address, yanked from one of the above tests. This is a pretty
			// poor test vector since it was created by the same code it is testing.
			// Unfortunately, I've been unable to find a golden negative example input
			// for curve25519 point decompression :/
			offCurve := MustPublicKeyFromBase58("12rqwuEgBYiGhBrDJStCiqEtzQpTTiZbh7teNVLuYcFA")
			require.False(t, offCurve.IsOnCurve())
		},
	)
}

// https://github.com/solana-labs/solana/blob/216983c50e0a618facc39aa07472ba6d23f1b33a/sdk/program/src/pubkey.rs#L590
func TestFindProgramAddress(t *testing.T) {
	for i := 0; i < 1_000; i++ {

		program_id := NewWallet().PrivateKey.PublicKey()
		address, bump_seed, err := FindProgramAddress(
			[][]byte{
				[]byte("Lil'"),
				[]byte("Bits"),
			},
			program_id,
		)
		require.NoError(t, err)

		got, err := CreateProgramAddress(
			[][]byte{
				[]byte("Lil'"),
				[]byte("Bits"),
				{bump_seed},
			},
			program_id,
		)
		require.NoError(t, err)
		require.Equal(t, address, got)
	}
}

func TestFindTokenMetadataAddress(t *testing.T) {
	// Zuuper Grapes (TOILET)
	// https://solscan.io/token/77K8mr457qxUSSNSfi4sSj5euP8DyuJJWHAUQVW8QCp3
	mint := MustPublicKeyFromBase58("77K8mr457qxUSSNSfi4sSj5euP8DyuJJWHAUQVW8QCp3")
	metadataPDA, bumpSeed, err := FindTokenMetadataAddress(mint)
	require.NoError(t, err)
	// https://solscan.io/account/GfihrEYCPrvUyrMyMQPdhGEStxa9nKEK2Wfn9iK4AZq2
	assert.Equal(t, metadataPDA, MustPublicKeyFromBase58("GfihrEYCPrvUyrMyMQPdhGEStxa9nKEK2Wfn9iK4AZq2"))
	assert.Equal(t, bumpSeed, uint8(0xfd))
}
