package filter_test

import (
	"testing"
	"unicode"

	"github.com/clipperhouse/uax29/iterators/filter"
	"github.com/clipperhouse/uax29/iterators/util"
)

func TestContains(t *testing.T) {
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

	f := filter.Contains(unicode.Latin, unicode.Ideographic)

	for _, test := range tests {
		got := f([]byte(test.input))

		if got != test.expected {
			t.Error(test.expected)
		}
	}
}

func TestEntirely(t *testing.T) {
	type test struct {
		input    string
		expected bool
	}

	tests := []test{
		{"", false},
		{"👍🐶", false},
		{"Hello", true},
		{"Hello世界", true},
		{"Hello ", false},
		{"Hello,世界", false},
	}

	ranges := []*unicode.RangeTable{
		unicode.Latin, unicode.Ideographic,
	}

	for _, test := range tests {
		got := util.Entirely([]byte(test.input), ranges...)

		if got != test.expected {
			t.Error(test.expected)
		}
	}
}
