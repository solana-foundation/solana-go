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

package rpc

import (
	stdjson "encoding/json"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
)

func TestData_base64_zstd(t *testing.T) {
	val := "KLUv/QQAWQAAaGVsbG8td29ybGTcLcaB"
	in := `["` + val + `", "base64+zstd"]`

	var data DataBytesOrJSON
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		[]byte("hello-world"),
		data.GetBinary(),
	)
	assert.Equal(t,
		solana.EncodingBase64Zstd,
		data.asDecodedBinary.Encoding,
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

	var data DataBytesOrJSON
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		[]byte(""),
		data.GetBinary(),
	)
	assert.Equal(t,
		solana.EncodingBase64Zstd,
		data.asDecodedBinary.Encoding,
	)
	assert.Equal(t,
		[]any{
			"",
			"base64+zstd",
		},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

func TestData_jsonParsed(t *testing.T) {
	in := `{"hello":"world"}`

	var data DataBytesOrJSON
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		stdjson.RawMessage(in),
		data.GetRawJSON(),
	)
	assert.Equal(t,
		map[string]any{
			"hello": "world",
		},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

func TestData_jsonParsed_empty(t *testing.T) {
	in := `{}`

	var data DataBytesOrJSON
	err := data.UnmarshalJSON([]byte(in))
	assert.NoError(t, err)

	assert.Equal(t,
		stdjson.RawMessage(in),
		data.GetRawJSON(),
	)
	assert.Equal(t,
		map[string]any{},
		mustJSONToInterface(mustAnyToJSON(data)),
	)
}

func TestData_DataBytesOrJSONFromBytes(t *testing.T) {
	in := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	dataBytesOrJSON := DataBytesOrJSONFromBytes(in)
	out := dataBytesOrJSON.GetBinary()
	assert.Equal(t, in, out)
}
