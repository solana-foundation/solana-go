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

package bin

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseFieldTag(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		expectValue *fieldTag
	}{
		{
			name: "no tags",
			tag:  "",
			expectValue: &fieldTag{
				Order: binary.LittleEndian,
			},
		},
		{
			name: "with a skip",
			tag:  `bin:"-"`,
			expectValue: &fieldTag{
				Order: binary.LittleEndian,
				Skip:  true,
			},
		},
		{
			name: "with a sizeof",
			tag:  `bin:"sizeof=Tokens"`,
			expectValue: &fieldTag{
				Order:  binary.LittleEndian,
				SizeOf: "Tokens",
			},
		},
		{
			name: "with a optional",
			tag:  `bin:"optional"`,
			expectValue: &fieldTag{
				Order:  binary.LittleEndian,
				Option: true,
			},
		},
		{
			name: "with a optional and size of",
			tag:  `bin:"optional sizeof=Nodes"`,
			expectValue: &fieldTag{
				Order:  binary.LittleEndian,
				Option: true,
				SizeOf: "Nodes",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectValue, parseFieldTag(reflect.StructTag(test.tag)))
		})
	}
}
