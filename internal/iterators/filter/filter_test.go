package filter_test

import (
	"testing"
	"unicode"

	"github.com/clipperhouse/uax29/internal/iterators/filter"
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

	f := filter.Contains(unicode.Latin, unicode.Ideographic)

	for _, test := range tests {
		got := f([]byte(test.input))

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
		{"Hello世界", true},
		{"Hello ", false},
		{"Hello,世界", false},
	}

	f := filter.Entirely(unicode.Latin, unicode.Ideographic)

	for _, test := range tests {
		got := f([]byte(test.input))

		if got != test.expected {
			t.Error(test.expected)
		}
	}
}
