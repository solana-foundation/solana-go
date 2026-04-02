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

package diff

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiff(t *testing.T) {
	type pair struct {
		left  any
		right any
	}

	tests := []struct {
		name      string
		in        pair
		expecteds []string
	}{
		// Plain
		{"plain - left nil, right nil",
			pair{nil, nil},
			nil,
		},
		{"plain - left nil, right set",
			pair{nil, &plainStruct{}},
			[]string{"<nil> => &{0} (*diff.plainStruct) (added)"},
		},
		{"plain - left set, right nil",
			pair{&plainStruct{}, nil},
			[]string{"&{0} (*diff.plainStruct) => <nil> (removed)"},
		},
		{"plain - equal",
			pair{&plainStruct{}, &plainStruct{}},
			nil,
		},
		{"plain - diff",
			pair{&plainStruct{Field: 1}, &plainStruct{Field: 2}},
			[]string{"1 (int) => 2 (int) (changed @ Field)"},
		},

		// Slice
		{"slice - equal both nil",
			pair{[]string(nil), []string(nil)},
			nil,
		},
		{"slice - equal both length 0",
			pair{[]string{}, []string{}},
			nil,
		},
		{"slice - diff both length 1",
			pair{[]string{"a"}, []string{"b"}},
			[]string{
				"a (string) => <nil> (removed @ [0])",
				"<nil> => b (string) (added @ [0])",
			},
		},
		{"slice - diff both length 2 re-ordered",
			pair{[]string{"a", "b"}, []string{"b", "a"}},
			[]string{
				"a (string) => <nil> (removed @ [0])",
				"<nil> => b (string) (added @ [0])",
				"b (string) => <nil> (removed @ [1])",
				"<nil> => a (string) (added @ [1])",
			},
		},
		{"slice - diff left is longer than right, all different",
			pair{[]string{"a", "b"}, []string{"c"}},
			[]string{
				"a (string) => <nil> (removed @ [0])",
				"<nil> => c (string) (added @ [0])",
				"b (string) => <nil> (removed @ [1->?])",
			},
		},
		{"slice - diff left is longer than right, some equals",
			pair{[]string{"a", "b"}, []string{"a"}},
			[]string{
				"b (string) => <nil> (removed @ [1->?])",
			},
		},
		{"slice - diff left is smaller than right, all different",
			pair{[]string{"a"}, []string{"b", "c"}},
			[]string{
				"a (string) => <nil> (removed @ [0])",
				"<nil> => b (string) (added @ [0])",
				"<nil> => c (string) (added @ [?->1])",
			},
		},
		{"slice - diff left is smaller than right, some equals",
			pair{[]string{"a"}, []string{"a", "b"}},
			[]string{
				"<nil> => b (string) (added @ [?->1])",
			},
		},

		// Full
		{"full - everything diff",
			pair{
				&topStruct{
					Literal: "x",
					Pointer: &plainStruct{Field: 1},
					Array:   []string{"a", "b"},
					Child:   &childStruct{Literal: "1", Pointer: &plainStruct{Field: 10}, Array: []string{"1", "2"}},
				},
				&topStruct{
					Literal: "y",
					Pointer: &plainStruct{Field: 2},
					Array:   []string{"b", "c"},
					Child:   &childStruct{Literal: "2", Pointer: &plainStruct{Field: 20}, Array: []string{"2"}},
				},
			},
			[]string{
				"x (string) => y (string) (changed @ Literal)",
				"1 (int) => 2 (int) (changed @ Pointer.Field)",
				"a (string) => <nil> (removed @ Array[0])",
				"<nil> => b (string) (added @ Array[0])",
				"b (string) => <nil> (removed @ Array[1])",
				"<nil> => c (string) (added @ Array[1])",
				"1 (string) => 2 (string) (changed @ Child.Literal)",
				"10 (int) => 20 (int) (changed @ Child.Pointer.Field)",
				"1 (string) => <nil> (removed @ Child.Array[0->?])",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actuals := accumulateDiffStrings(test.in.left, test.in.right)
			assert.Equal(t, test.expecteds, actuals)
		})
	}
}

func eventToString(event *Event) string {
	path := ""
	if len(event.Path) > 1 {
		path = " @ " + event.Path.String()
	}

	return fmt.Sprintf("%s => %s (%s%s)", reflectValueToString(event.Old), reflectValueToString(event.New), event.Kind, path)
}

func TestDiff_EventMatch(t *testing.T) {
	tests := []struct {
		name           string
		left           any
		right          any
		pattern        string
		expectedMatch  bool
		expectedGroups []string
	}{
		{
			"deep array added one",
			&topStruct{Child: &childStruct{Array: []string{"1"}}},
			&topStruct{Child: &childStruct{Array: []string{"1", "2"}}},
			"Child.Array[#]",
			true,
			[]string{"?->1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events := accumulateDiff(test.left, test.right)
			require.Len(t, events, 1)

			match, groups := events[0].Match(test.pattern)
			if test.expectedMatch {
				assert.True(t, match, "Expected pattern %q to match diff path %q", test.pattern, events[0].Path)
				assert.Equal(t, test.expectedGroups, groups)
			} else {
				assert.False(t, match, "Expected pattern %q to NOT match diff path %q", test.pattern, events[0].Path)
			}
		})
	}
}

// There is something inherently broken with this, the accumulation seems to broke leading to incorrect
// results. I assume it's a Golang thing related to slice and struct as value versus pointers. It works
// only single event but starts to act weirdly when there > 1, like the Event's Path is all wrong. It's
// better to try to avoid it when possible.
func accumulateDiff(left, right any) (out []Event) {
	Diff(left, right, OnEvent(func(event Event) {
		out = append(out, event)
	}))
	return
}

func accumulateDiffStrings(left, right any) (out []string) {
	Diff(left, right, OnEvent(func(event Event) {
		out = append(out, eventToString(&event))
	}))
	return
}

type topStruct struct {
	Literal string
	Pointer *plainStruct
	Array   []string
	Child   *childStruct
}

type childStruct struct {
	Literal string
	Pointer *plainStruct
	Array   []string
}

type plainStruct struct {
	Field int
}
