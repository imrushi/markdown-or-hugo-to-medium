package helper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"dot.slash/backslash\\underscore_pound#plus+hyphen-", "dot.slash/backslash\\underscore_pound#plus+hyphen-"},
		{"abcXYZ0123456789", "abcXYZ0123456789"},
		{"%20 %2", "%20-2"},
		{"foo- bar", "foo-bar"},
		{"  Foo bar  ", "Foo-bar"},
		{"Foo.Bar/foo_Bar-Foo", "Foo.Bar/foo_Bar-Foo"},
		{"fOO,bar:foobAR", "fOObarfoobAR"},
		{"FOo/BaR.html", "FOo/BaR.html"},
		{"трям/трям", "трям/трям"},
		{"은행", "은행"},
		{"Банковский кассир", "Банковский-кассир"},
		{"संस्कृत", "संस्कृत"},
		{"a%C3%B1ame", "a%C3%B1ame"},
		{"this+is+a+test", "this+is+a+test"},
		{"~foo", "~foo"},
		{"foo--bar", "foo--bar"},
		{"foo@bar", "foo@bar"},
	}

	for _, test := range tests {
		p := NewPathSpec("https://ruship.dev/post/")
		output := p.MakePath(test.input)
		assert.Equal(t, test.expected, output)
	}
}

func TestMakePathSanitized(t *testing.T) {
	p := NewPathSpec("https://ruship.dev/post/")

	tests := []struct {
		input    string
		expected string
	}{
		{"  FOO bar  ", "foo-bar"},
		{"Foo.Bar/fOO_bAr-Foo", "foo.bar/foo_bar-foo"},
		{"FOO,bar:FooBar", "foobarfoobar"},
		{"foo/BAR.HTML", "foo/bar.html"},
		{"трям/трям", "трям/трям"},
		{"은행", "은행"},
	}

	for _, test := range tests {
		output := p.MakePathSanitized(test.input)
		assert.Equal(t, test.expected, output)
	}
}

func TestURLize(t *testing.T) {
	p := NewPathSpec("https://ruship.dev/post/")

	fmt.Println(p.URLize("https://ruship.dev/post/Foo.Bar/foo_Bar-Foo"))
}
