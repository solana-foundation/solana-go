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
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// TestBlockSubscribeRejectsJSONParsedEncoding verifies that BlockSubscribe
// fails fast when the caller asks for EncodingJSONParsed. The response struct
// (BlockResult) models the non-parsed transaction layout; letting the request
// go out anyway only surfaces as a confusing decode error on the first
// notification (issue #291). The rejection should also point the caller at
// ParsedBlockSubscribe so the fix is obvious.
func TestBlockSubscribeRejectsJSONParsedEncoding(t *testing.T) {
	cl := &Client{}
	_, err := cl.BlockSubscribe(NewBlockSubscribeFilterAll(), &BlockSubscribeOpts{
		Encoding: solana.EncodingJSONParsed,
	})
	require.Error(t, err)
	if !strings.Contains(err.Error(), "ParsedBlockSubscribe") {
		t.Fatalf("error should mention ParsedBlockSubscribe, got: %v", err)
	}
}

// TestBlockSubscribeRejectsUnsupportedEncoding verifies the existing
// "unsupported encoding" rejection path is unchanged for arbitrary values.
func TestBlockSubscribeRejectsUnsupportedEncoding(t *testing.T) {
	cl := &Client{}
	_, err := cl.BlockSubscribe(NewBlockSubscribeFilterAll(), &BlockSubscribeOpts{
		Encoding: solana.EncodingType("not-a-real-encoding"),
	})
	require.Error(t, err)
	if !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("error should describe an unsupported encoding, got: %v", err)
	}
}
