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

package ws

import (
	"io"

	gojson "github.com/goccy/go-json"
)

var json = struct {
	Marshal       func(v any) ([]byte, error)
	MarshalIndent func(v any, prefix, indent string) ([]byte, error)
	Unmarshal     func(data []byte, v any) error
	NewDecoder    func(r io.Reader) *gojson.Decoder
	NewEncoder    func(w io.Writer) *gojson.Encoder
}{
	Marshal:       gojson.Marshal,
	MarshalIndent: gojson.MarshalIndent,
	Unmarshal:     gojson.Unmarshal,
	NewDecoder:    gojson.NewDecoder,
	NewEncoder:    gojson.NewEncoder,
}
