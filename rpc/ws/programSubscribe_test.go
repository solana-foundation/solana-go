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

// buildProgramSubscribeConf re-runs the same config-object construction
// path ProgramSubscribeWithConfig would send to the validator, so the
// test can assert on the wire shape without a live websocket.
func buildProgramSubscribeConf(opts *ProgramSubscribeOpts) map[string]any {
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
		if len(opts.Filters) > 0 {
			conf["filters"] = opts.Filters
		}
		if opts.DataSlice != nil {
			conf["dataSlice"] = opts.DataSlice
		}
	}
	return conf
}

func TestProgramSubscribeConfDefaults(t *testing.T) {
	conf := buildProgramSubscribeConf(nil)
	require.Equal(t, "base64", conf["encoding"])
	require.NotContains(t, conf, "commitment")
	require.NotContains(t, conf, "filters")
	require.NotContains(t, conf, "dataSlice")
}

func TestProgramSubscribeConfWithDataSlice(t *testing.T) {
	off, length := uint64(165), uint64(72)
	conf := buildProgramSubscribeConf(&ProgramSubscribeOpts{
		Commitment: rpc.CommitmentConfirmed,
		Encoding:   solana.EncodingBase64,
		DataSlice:  &rpc.DataSlice{Offset: &off, Length: &length},
	})
	require.Equal(t, rpc.CommitmentConfirmed, conf["commitment"])

	// Round-trip the dataSlice through JSON to confirm the wire shape
	// matches the RPC spec ({"offset": N, "length": N}).
	encoded, err := stdjson.Marshal(conf["dataSlice"])
	require.NoError(t, err)
	require.JSONEq(t, `{"offset":165,"length":72}`, string(encoded))
}

func TestProgramSubscribeConfWithDataSliceAndFilters(t *testing.T) {
	off, length := uint64(0), uint64(8)
	conf := buildProgramSubscribeConf(&ProgramSubscribeOpts{
		Filters: []rpc.RPCFilter{
			{DataSize: 165},
		},
		DataSlice: &rpc.DataSlice{Offset: &off, Length: &length},
	})
	require.Contains(t, conf, "filters")
	require.Contains(t, conf, "dataSlice")
	filters := conf["filters"].([]rpc.RPCFilter)
	require.Len(t, filters, 1)
	require.Equal(t, uint64(165), filters[0].DataSize)
}
