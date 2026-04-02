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

package rpc

import (
	stdjson "encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockJSONRPCServer struct {
	*httptest.Server
	body []byte
}

func mockJSONRPC(t *testing.T, response any) (mock *mockJSONRPCServer, close func()) {
	mock = &mockJSONRPCServer{
		Server: httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			var err error
			mock.body, err = io.ReadAll(req.Body)
			require.NoError(t, err)

			var responseBody []byte
			if v, ok := response.(stdjson.RawMessage); ok {
				responseBody = v
			} else {
				responseBody, err = json.Marshal(response)
				require.NoError(t, err)
			}

			rw.Write(responseBody)
		})),
	}

	return mock, func() { mock.Close() }
}

func (s *mockJSONRPCServer) RequestBodyAsJSON(t *testing.T) (out string) {
	return string(s.body)
}

func (s *mockJSONRPCServer) RequestBody(t *testing.T) (out map[string]any) {
	err := json.Unmarshal(s.body, &out)
	require.NoError(t, err)

	return out
}
