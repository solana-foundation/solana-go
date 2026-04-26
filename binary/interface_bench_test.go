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
	"crypto/rand"
	"io"
	mrand "math/rand"
	"testing"
)

func BenchmarkByteCount(b *testing.B) {
	nestedSmall := &benchNested{
		N1: &benchSubset1{F3: makeStringList(10), F4: makeUint64List(10)},
		N2: &benchSubset2{},
	}

	nestedLarge := &benchNested{
		N1: &benchSubset1{F3: makeStringList(200), F4: makeUint64List(200)},
		N2: &benchSubset2{},
	}

	benchmarks := []struct {
		name string
		v    interface{}
	}{
		{"flat", &benchFlat{}},
		{"nested/small list", nestedSmall},
		{"nested/large list", nestedLarge},
		{"deep/small list", &benchDeepNested{N1: nestedSmall, N2: nestedSmall, N3: nestedSmall}},
		{"deep/large list", &benchDeepNested{N1: nestedLarge, N2: nestedLarge, N3: nestedLarge}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			setupBench(b)
			for i := 0; i < b.N; i++ {
				BinByteCount(bm.v)
			}
		})
	}
}

type benchFlat struct {
	F1 string
	F2 int16
	F3 uint16
	F4 int32
	F5 uint32
	F6 int64
	F7 uint64
	F8 float32
	F9 float64
}

type benchNested struct {
	N1 *benchSubset1
	F1 string
	F2 int16
	F3 uint16
	F4 int32
	F5 uint32
	N2 *benchSubset2
	F6 int64
	F7 uint64
	F8 float32
	F9 float64
}

type benchDeepNested struct {
	N1 *benchNested
	F1 string
	F2 int16
	N2 *benchNested
	F4 int32
	F5 uint32
	N3 *benchNested
	F6 int64
	F7 uint64
	F8 float32
	F9 float64
}

type benchSubset1 struct {
	F1 int64
	F2 string
	F3 []string
	F4 []int64
}

type benchSubset2 struct {
	F7 uint64
	F8 float32
	F9 float64
}

func makeUint64List(itemCount int) (out []int64) {
	out = make([]int64, itemCount)
	for i := 0; i < itemCount; i++ {
		// get random int64:
		out[i] = mrand.Int63()
	}
	return
}

func makeStringList(itemCount int) (out []string) {
	out = make([]string, itemCount)
	for i := 0; i < itemCount; i++ {
		data := make([]byte, i>>1)
		io.ReadFull(rand.Reader, data)
		out[i] = string(data)
	}
	return
}
