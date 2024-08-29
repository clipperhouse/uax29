package util_test

import (
	"testing"
	"unicode"

	"github.com/clipperhouse/uax29/iterators/util"
	"golang.org/x/text/unicode/rangetable"
)

func TestContains(t *testing.T) {
	t.Parallel()

	type test struct {
		input    string
		expected bool
	}

	tests := []test{
		{"", false},
		{"👍🐶", false},
		{"Hello", true},
		{"Hello, 世界.", true},
		{"世界", true},
	}

	ranges := rangetable.Merge(unicode.Latin, unicode.Ideographic)

	for _, test := range tests {
		got := util.Contains([]byte(test.input), ranges)

		if got != test.expected {
			t.Error(test.expected)
		}
	}
}

func TestEntirely(t *testing.T) {
	t.Parallel()

	type test struct {
		input    string
		expected bool
	}

	tests := []test{
		{"", false},
		{"👍🐶", false},
		{"Hello", true},
		{"Hello世界", false},
		{"Hello ", false},
		{"Hello,世界", false},
		{"世界", false},
	}

	for _, test := range tests {
		got := util.Entirely([]byte(test.input), unicode.Latin)

		if got != test.expected {
			t.Error(test.expected)
		}
	}
}
