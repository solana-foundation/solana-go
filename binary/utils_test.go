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
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_twosComplement(t *testing.T) {
	tests := []struct {
		name   string
		in     []byte
		expect []byte
	}{
		{
			name:   "empty array",
			in:     []byte{},
			expect: []byte{0x1},
		},
		{
			name:   "one element",
			in:     []byte{0x01},
			expect: []byte{0xff},
		},
		{
			name:   "basic test",
			in:     []byte{0xaa, 0xbb, 0xcc},
			expect: []byte{0x55, 0x44, 0x34},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, twosComplement(test.in))
		})
	}

}
