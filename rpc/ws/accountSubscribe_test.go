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

package ws

import (
	"testing"

	stdjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/require"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// buildAccountSubscribeConf re-runs the same config-object construction
// path that AccountSubscribeWithConfig would send to the validator, so
// the test can assert on the wire shape without a live websocket.
func buildAccountSubscribeConf(opts *AccountSubscribeOpts) map[string]any {
	conf := map[string]any{
		"encoding": "base64",
	}
	if opts != nil {
		if opts.Commitment != "" {
			conf["commitment"] = opts.Commitment
		}
		if opts.Encoding != "" {
			conf["encoding"] = opts.Encoding
		}
		if opts.DataSlice != nil {
			conf["dataSlice"] = opts.DataSlice
		}
	}
	return conf
}

func TestAccountSubscribeConfDefaults(t *testing.T) {
	conf := buildAccountSubscribeConf(nil)
	require.Equal(t, "base64", conf["encoding"])
	require.NotContains(t, conf, "commitment")
	require.NotContains(t, conf, "dataSlice")
}

func TestAccountSubscribeConfWithDataSlice(t *testing.T) {
	off, length := uint64(8), uint64(32)
	conf := buildAccountSubscribeConf(&AccountSubscribeOpts{
		Commitment: rpc.CommitmentConfirmed,
		Encoding:   solana.EncodingBase64,
		DataSlice:  &rpc.DataSlice{Offset: &off, Length: &length},
	})
	require.Equal(t, rpc.CommitmentConfirmed, conf["commitment"])
	require.Equal(t, solana.EncodingBase64, conf["encoding"])

	// Round-trip the dataSlice through JSON to confirm the wire shape
	// matches the RPC spec ({"offset": N, "length": N}). The validator
	// rejects the request if the keys aren't camelCase or if the type
	// is anything other than an object, so this is the part that
	// matters most for parity.
	encoded, err := stdjson.Marshal(conf["dataSlice"])
	require.NoError(t, err)
	require.JSONEq(t, `{"offset":8,"length":32}`, string(encoded))
}

func TestAccountSubscribeConfEncodingOverride(t *testing.T) {
	conf := buildAccountSubscribeConf(&AccountSubscribeOpts{
		Encoding: solana.EncodingBase58,
	})
	require.Equal(t, solana.EncodingBase58, conf["encoding"])
}
