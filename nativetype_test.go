// Copyright 2021 github.com/gagliardetto
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
	"strconv"
	"testing"

	"github.com/gagliardetto/solana-go/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustHashFromBase58(t *testing.T) {
	require.Panics(t, func() {
		MustHashFromBase58("toto")
	})
}

func TestHash_UnmarshalText(t *testing.T) {
	in := MustHashFromBase58("uoEAQCWCKjV9ecsBvngctJ7upNBZX7hpN4SfdR6TaUz")

	out, err := in.MarshalText()
	assert.NoError(t, err)

	var ha Hash
	err = ha.UnmarshalText(out)
	assert.NoError(t, err)
	assert.Equal(t, in, ha)
}

func TestHashFromBase58(t *testing.T) {
	in := "uoEAQCWCKjV9ecsBvngctJ7upNBZX7hpN4SfdR6TaUz"
	out, err := HashFromBase58(in)
	assert.NoError(t, err)
	assert.Equal(t, in, out.String())

	jsonOut, err := out.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, strconv.Quote(in), string(jsonOut))

	var ha Hash
	assert.True(t, ha.IsZero())
	err = ha.UnmarshalJSON(jsonOut)
	assert.NoError(t, err)
	assert.Equal(t, out, ha)
	assert.True(t, out.Equals(ha))
	assert.False(t, out.Equals(Hash{}))
	assert.False(t, out.IsZero())
}

func TestSignature_UnmarshalText(t *testing.T) {
	in := MustSignatureFromBase58("gD3jeeaPNiyuJvTKXNEv1gntazWEkvpocofEmrz2rL6Fi4prWSsBH6a9SrwyZEatAozyMsnK2fnk3APXNFxD2Mq")

	out, err := in.MarshalText()
	assert.NoError(t, err)

	var sig Signature
	err = sig.UnmarshalText(out)
	assert.NoError(t, err)
	assert.Equal(t, in, sig)
}

func TestSignatureFromBase58(t *testing.T) {
	in := "gD3jeeaPNiyuJvTKXNEv1gntazWEkvpocofEmrz2rL6Fi4prWSsBH6a9SrwyZEatAozyMsnK2fnk3APXNFxD2Mq"
	out, err := SignatureFromBase58(in)
	assert.NoError(t, err)
	assert.Equal(t, in, out.String())

	jsonOut, err := out.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, strconv.Quote(in), string(jsonOut))

	var sig Signature
	assert.True(t, sig.IsZero())
	err = sig.UnmarshalJSON(jsonOut)
	assert.NoError(t, err)
	assert.Equal(t, out, sig)
	assert.True(t, out.Equals(sig))
	assert.False(t, out.Equals(Signature{}))
	assert.False(t, out.IsZero())
}

func TestMustSignatureFromBase58(t *testing.T) {
	require.Panics(t, func() {
		MustSignatureFromBase58("toto")
	})
}

func TestBase58(t *testing.T) {
	in := "RYcCwZg97M2jet84ttG8"

	out, err := base58.Decode(in)
	assert.NoError(t, err)
	assert.Equal(t, in, Base58(out).String())

	jsonOut, err := Base58(out).MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, strconv.Quote(in), string(jsonOut))

	var b58 Base58
	err = b58.UnmarshalJSON(jsonOut)
	assert.NoError(t, err)
}

func TestData_base64(t *testing.T) {
	val := "dGVzdA=="
	in := `["` + val + `", "base64"]`

	var data Data
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		[]byte("test"),
		data.Content,
	)

	assert.Equal(t,
		EncodingBase64,
		data.Encoding,
	)

	assert.Equal(t,
		[]any{
			val,
			"base64",
		},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

func TestData_base64_empty(t *testing.T) {
	val := ""
	in := `["", "base64"]`

	var data Data
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		[]byte(""),
		data.Content,
	)

	assert.Equal(t,
		EncodingBase64,
		data.Encoding,
	)

	assert.Equal(t,
		[]any{
			val,
			"base64",
		},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

func TestData_base64_zstd(t *testing.T) {
	val := "KLUv/QQAWQAAaGVsbG8td29ybGTcLcaB"
	in := `["` + val + `", "base64+zstd"]`

	var data Data
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		[]byte("hello-world"),
		data.Content,
	)

	assert.Equal(t,
		EncodingBase64Zstd,
		data.Encoding,
	)

	assert.Equal(t,
		[]any{
			val,
			"base64+zstd",
		},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

func TestData_base64_zstd_empty(t *testing.T) {
	in := `["", "base64+zstd"]`

	var data Data
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		[]byte(""),
		data.Content,
	)

	assert.Equal(t,
		EncodingBase64Zstd,
		data.Encoding,
	)

	assert.Equal(t,
		[]any{
			"",
			"base64+zstd",
		},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

func TestData_base58(t *testing.T) {
	val := "3yZe7d"
	in := `["` + val + `", "base58"]`

	var data Data
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		[]byte("test"),
		data.Content,
	)

	assert.Equal(t,
		EncodingBase58,
		data.Encoding,
	)

	assert.Equal(t,
		[]any{
			val,
			"base58",
		},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

func TestData_base58_empty(t *testing.T) {
	val := ""
	in := `["", "base58"]`

	var data Data
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		[]byte(""),
		data.Content,
	)

	assert.Equal(t,
		EncodingBase58,
		data.Encoding,
	)

	assert.Equal(t,
		[]any{
			val,
			"base58",
		},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

// mustAnyToJSON marshals the provided variable
// to JSON bytes.
func mustAnyToJSON(raw any) []byte {
	out, err := json.Marshal(raw)
	if err != nil {
		panic(err)
	}
	return out
}

// mustJSONToInterface unmarshals the provided JSON bytes
// into an `interface{}` type variable, and returns it.
func mustJSONToInterface(rawJSON []byte) any {
	var out any
	err := json.Unmarshal(rawJSON, &out)
	if err != nil {
		panic(err)
	}
	return out
}

func TestMustPublicKeyFromBase58(t *testing.T) {
	require.Panics(t, func() {
		MustPublicKeyFromBase58("toto")
	})
}

func TestSignatureVerify(t *testing.T) {
	type signerSignature struct {
		Sig    Signature
		Signer PublicKey
	}
	type testCase struct {
		Message string
		Signers []signerSignature
	}

	testCases := []testCase{
		{
			Signers: []signerSignature{
				{
					Sig:    MustSignatureFromBase58("3APVUYY2Qcq9WwPSciQ1xicshQcsXJWhgD1x5YEMNnNnKtpaAJtRhDVMWcvePosamk3JfSKpBM8qt5ZkAQgRNSzo"),
					Signer: MustPublicKeyFromBase58("2V7t5NaKY7aGkwytCWQgvUYZfEr9XMwNChhJEakTExk6"),
				},
			},
			Message: "AQACBBYPusE6993YBdMXCj3gxr2XEmoeAsDSWdCobvgh1uXHoaZGX0wuvyRMMdgLyVwnNFo0JOQowt7zPs7Z6Q0/cBsGp9UXGMd0yShWY5hpHV62i164o5tLbVxzVVshAAAAANzl6+HknDufEUy1VExQqZ7A1pLWP1Z5WuAprIPZ6ovia//gAIGoVfMJgSuJp8/VsCUq8qvMdiHNGrMrQAxoS2QBAwMAAQIoAgAAAAcAAAABAAAAAAAAAM/ncrkCAAAACJ9fAQAAAAD7FSgHAAAAAA==",
		},
		{
			Signers: []signerSignature{
				{
					Sig:    MustSignatureFromBase58("5TPaoKhkhRck3TTjyEQe1TbgXHATggx9iyWaXqj7foCUSuXnVUh3iqMrKKndWZn31cV3KEiBDiqg9wM8ijk41eyQ"),
					Signer: MustPublicKeyFromBase58("G2fzpkX69kmaYtDtMfUkQHykSvgV24wze9ikb911FHT"),
				},
			},
			Message: "AQADBQPZmmOEu9u2f5Tkm9PJx3oVRbAt9pEyjDD4zyadddWyjNBZdQ4NOL1dsfZy9kwqkh5o0/MYOsFAM42WtfumAgUGp9UXGS8Kr8byZeP7d8x62oLFKdC+OxNuLQBVIAAAAAan1RcYx3TJKFZjmGkdXraLXrijm0ttXHNVWyEAAAAAB2FIHTV0dLt8TXYk69O9s9g1XnPREEP8DaNTgAAAAACQ0H7JStckBZBqg11runLYUTbmBpWo2DDpraEiQeoZjwEEBAECAwBNAgAAAAMAAAAAAAAA+RUoBwAAAAD6FSgHAAAAAPsVKAcAAAAAxAjCTj1+3u3zheR66vcYXXIdlKoYPL0DmMmTP+lffN4B0NIDYgAAAAA=",
		},
		{
			Signers: []signerSignature{
				{
					Sig:    MustSignatureFromBase58("T1BbonhT8wdLpRMPN8UAtshjukRwot8Z5ta7wSaHujSfrqKuFpQD56iwa5hjESMzsNKFpzPVETCVZ54puCiZh2y"),
					Signer: MustPublicKeyFromBase58("3jBp2LRCFgSmBS1Qemji1rHWS9ytSoN5pYpWzK5o9CDY"),
				},
			},
			Message: "AQACBSiF/dgUa5IBH6yxFwL17cOzoDz4R1PutgvAweUD6DIDUmyl4Hbm9LK+2Kdzb3a7Kekg6CeIHaMkWndgFfiWqY9HUbWFuDsRO/Bb0WU3qhxctjMv5tdv2nxiD0vIlJzBpQan1RcYx3TJKFZjmGkdXraLXrijm0ttXHNVWyEAAAAAv47fHJRrVrtD0l+YuHtr5REXMAphJhr5+flO7XpagIiuIXQDe/BrXhY6rSa8S18aVTSrLtZX3XBT12vynIOHuQEEBAEAAwInJjIkEAIhro89MaE0EkAoADDzq6A5ORdIUPwYMxJAQb1SliGONRJA",
		},
		{
			Signers: []signerSignature{
				{
					Sig:    MustSignatureFromBase58("2c5u22N6Yyjj7qQdGtEshp4n9r4akLyRLJijnXL9HnQAXqq7S9cfnGdi5mwnfQ5kJHQuRv62T366SKaztHiUUqnK"),
					Signer: MustPublicKeyFromBase58("AXUChvpRwUUPMJhA4d23WcoyAL7W8zgAeo7KoH57c75F"),
				},
			},
			Message: "AQACBo2HXSJ3fw0pdiiQj9fBoRA5Wicjyla4ARhrK90xeo0mD4K9wHSEzSq9lEFUOSPloeUrUL/2uV3S6+lTeGNtTMtTP2NRTSucECyiWtS2HzfxwLnbnxVbXNH0T0egDFCfVK5ESO7I2Pz56XFyxewO0gbya1rvqPPu0CGg/LC/Wl5BhQ8tbgKkevgk0Jq2ncQtcMsoy/okn7fuV7nSVsEnYu8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIvotbwkk/hzTtAxHIaRCSJjB4WdQUXrn+6aEapl6XgPAgQFAQIDAgIHAAMAAAABAAUCAAAMAgAAAOsNAAAAAAAA",
		},
		{
			Signers: []signerSignature{
				{
					Sig:    MustSignatureFromBase58("2amhinnUjRMef8aNMKKhGCf4odSR3q9Tq2SjYjvFePN6ZmStF5EbNmfdsqssXe4XjXn6Nidu1Sg4MKyo5UPVWQgf"),
					Signer: MustPublicKeyFromBase58("2vKtu3nW1TS6iPvJPK8R88B5QfDrwJDwwB11Uu1CN9o7"),
				},
				{
					Sig:    MustSignatureFromBase58("5qzhXaDMhcVZA2y6KDK218noKdX7cd7aRCC1XvFFUTEievDMt8EbCC1oCKyqGvGmZs5UzE34v9JR8846HtJVF6qe"),
					Signer: MustPublicKeyFromBase58("96i77Hg7RL7J4Lrc1d6MXFofCKwEFZe84DejFUMhUs19"),
				},
			},
			Message: "AgEBBByE1Y6EqCJKsr7iEupU6lsBHtBdtI4SK3yWMCFA0iEKeFPgnGmtp+1SIX1Ak+sN65iBaR7v4Iim5m1OEuFQTgi9N57UnhNpCNuUePaTt7HJaFBmyeZB3deXeKWVudpY3gAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWVECK/n3a7QR6OKWYR4DuAVjS6FXgZj82W0dJpSIPnEBAwQAAgEDDAIAAABAQg8AAAAAAA==",
		},
	}

	for _, tc := range testCases {
		msg, err := base64.StdEncoding.DecodeString(tc.Message)
		require.NoError(t, err)

		for _, tcs := range tc.Signers {
			require.True(t, tcs.Sig.Verify(tcs.Signer, msg))
			require.False(t, tcs.Sig.Verify(BPFLoaderDeprecatedProgramID, msg))
			require.True(t, tcs.Signer.Verify(msg, tcs.Sig))
		}
	}
}
