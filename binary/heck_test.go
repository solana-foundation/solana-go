package bin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCamelCase(t *testing.T) {
	type item struct {
		input string
		want  string
	}
	tests := []item{
		// TODO: find out if need to fix, and if yes, then fix.
		// {"1hello", "1Hello"},            // actual: `1hello`
		// {"1Hello", "1Hello"},            // actual: `1hello`
		// {"hello1world", "Hello1World"},  // actual: `Hello1world`
		// {"mGridCol6@md", "MGridCol6md"}, // actual: `MGridCol6Md`
		// {"A::a", "Aa"},                  // actual: `AA`
		// {"foìBar-baz", "FoìBarBaz"},
		//
		{"hello1World", "Hello1World"},
		{"Hello1World", "Hello1World"},
		{"foo", "Foo"},
		{"foo-bar", "FooBar"},
		{"foo-bar-baz", "FooBarBaz"},
		{"foo--bar", "FooBar"},
		{"--foo-bar", "FooBar"},
		{"--foo--bar", "FooBar"},
		{"FOO-BAR", "FooBar"},
		{"FOÈ-BAR", "FoèBar"},
		{"-foo-bar-", "FooBar"},
		{"--foo--bar--", "FooBar"},
		{"foo-1", "Foo1"},
		{"foo.bar", "FooBar"},
		{"foo..bar", "FooBar"},
		{"..foo..bar..", "FooBar"},
		{"foo_bar", "FooBar"},
		{"__foo__bar__", "FooBar"},
		{"__foo__bar__", "FooBar"},
		{"foo bar", "FooBar"},
		{"  foo  bar  ", "FooBar"},
		{"-", ""},
		{" - ", ""},
		{"fooBar", "FooBar"},
		{"fooBar-baz", "FooBarBaz"},
		{"fooBarBaz-bazzy", "FooBarBazBazzy"},
		{"FBBazzy", "FbBazzy"},
		{"F", "F"},
		{"FooBar", "FooBar"},
		{"Foo", "Foo"},
		{"FOO", "Foo"},
		{"--", ""},
		{"", ""},
		{"--__--_--_", ""},
		{"foo bar?", "FooBar"},
		{"foo bar!", "FooBar"},
		{"foo bar$", "FooBar"},
		{"foo-bar#", "FooBar"},
		{"XMLHttpRequest", "XmlHttpRequest"},
		{"AjaxXMLHttpRequest", "AjaxXmlHttpRequest"},
		{"Ajax-XMLHttpRequest", "AjaxXmlHttpRequest"},
		{"Hello11World", "Hello11World"},
		{"hello1", "Hello1"},
		{"Hello1", "Hello1"},
		{"h1W", "H1W"},
		// TODO: add support to non-alphanumeric characters (non-latin, non-ascii).
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, ToPascalCase(test.input))
		})
	}
}
