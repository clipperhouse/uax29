package iterators_test

import (
	"testing"

	"github.com/clipperhouse/uax29/v2/internal/iterators"
)

// simpleSpaceSplitString is a lossless SplitFunc that splits on spaces for strings
// It treats contiguous spaces and contiguous non-spaces as separate tokens
func simpleSpaceSplitString(data string, atEOF bool) (int, string, error) {
	if len(data) == 0 {
		return 0, data, nil
	}

	// Determine if we're starting with a space or non-space
	isSpace := data[0] == ' '

	// Find the end of the current token (same type as start)
	i := 1
	for i < len(data) && (data[i] == ' ') == isSpace {
		i++
	}

	// Return the token and advance by its length
	token := data[:i]
	return len(token), token, nil
}

// simpleSpaceSplitBytes is a lossless SplitFunc that splits on spaces for []byte
// It treats contiguous spaces and contiguous non-spaces as separate tokens
func simpleSpaceSplitBytes(data []byte, atEOF bool) (int, []byte, error) {
	if len(data) == 0 {
		return 0, data, nil
	}

	// Determine if we're starting with a space or non-space
	isSpace := data[0] == ' '

	// Find the end of the current token (same type as start)
	i := 1
	for i < len(data) && (data[i] == ' ') == isSpace {
		i++
	}

	// Return the token and advance by its length
	token := data[:i]
	return len(token), token, nil
}

func TestIterator_String(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single word",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "two words",
			input:    "hello world",
			expected: []string{"hello", " ", "world"},
		},
		{
			name:     "multiple words",
			input:    "hello world test",
			expected: []string{"hello", " ", "world", " ", "test"},
		},
		{
			name:     "words with multiple spaces",
			input:    "hello  world   test",
			expected: []string{"hello", "  ", "world", "   ", "test"},
		},
		{
			name:     "leading and trailing spaces",
			input:    " hello world ",
			expected: []string{" ", "hello", " ", "world", " "},
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: []string{"   "},
		},
		{
			name:     "unicode characters",
			input:    "café naïve",
			expected: []string{"café", " ", "naïve"},
		},
		{
			name:     "emoji and unicode",
			input:    "hello 🌍 world",
			expected: []string{"hello", " ", "🌍", " ", "world"},
		},
		{
			name:     "chinese characters",
			input:    "你好 世界",
			expected: []string{"你好", " ", "世界"},
		},
		{
			name:     "mixed unicode and spaces",
			input:    "  café  naïve  ",
			expected: []string{"  ", "café", "  ", "naïve", "  "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iter := iterators.New(simpleSpaceSplitString, tt.input)
			var got []string

			for iter.Next() {
				got = append(got, iter.Value())
			}

			if len(got) != len(tt.expected) {
				t.Errorf("expected %d tokens, got %d", len(tt.expected), len(got))
				return
			}

			for i, expected := range tt.expected {
				if got[i] != expected {
					t.Errorf("token %d: expected %q, got %q", i, expected, got[i])
				}
			}
		})
	}
}

func TestIterator_Bytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected [][]byte
	}{
		{
			name:     "empty bytes",
			input:    []byte{},
			expected: [][]byte{},
		},
		{
			name:     "single word",
			input:    []byte("hello"),
			expected: [][]byte{[]byte("hello")},
		},
		{
			name:     "multiple words",
			input:    []byte("hello world test"),
			expected: [][]byte{[]byte("hello"), []byte(" "), []byte("world"), []byte(" "), []byte("test")},
		},
		{
			name:     "unicode characters",
			input:    []byte("café naïve"),
			expected: [][]byte{[]byte("café"), []byte(" "), []byte("naïve")},
		},
		{
			name:     "emoji and unicode",
			input:    []byte("hello 🌍 world"),
			expected: [][]byte{[]byte("hello"), []byte(" "), []byte("🌍"), []byte(" "), []byte("world")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iter := iterators.New(simpleSpaceSplitBytes, tt.input)
			var got [][]byte

			for iter.Next() {
				got = append(got, iter.Value())
			}

			if len(got) != len(tt.expected) {
				t.Errorf("expected %d tokens, got %d", len(tt.expected), len(got))
				return
			}

			for i, expected := range tt.expected {
				if string(got[i]) != string(expected) {
					t.Errorf("token %d: expected %q, got %q", i, string(expected), string(got[i]))
				}
			}
		})
	}
}

func TestIterator_Positions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			start, end int
			value      string
		}
	}{
		{
			name:  "ascii",
			input: "hello world test",
			expected: []struct {
				start, end int
				value      string
			}{
				{0, 5, "hello"},
				{5, 6, " "},
				{6, 11, "world"},
				{11, 12, " "},
				{12, 16, "test"},
			},
		},
		{
			name:  "unicode",
			input: "café naïve",
			expected: []struct {
				start, end int
				value      string
			}{
				{0, 5, "café"},   // "café" is 5 bytes in UTF-8
				{5, 6, " "},      // space is 1 byte
				{6, 12, "naïve"}, // "naïve" is 6 bytes in UTF-8
			},
		},
		{
			name:  "emoji",
			input: "hello 🌍 world",
			expected: []struct {
				start, end int
				value      string
			}{
				{0, 5, "hello"},   // "hello" is 5 bytes
				{5, 6, " "},       // space is 1 byte
				{6, 10, "🌍"},      // "🌍" is 4 bytes in UTF-8
				{10, 11, " "},     // space is 1 byte
				{11, 16, "world"}, // "world" is 5 bytes
			},
		},
		{
			name:  "chinese",
			input: "你好 世界",
			expected: []struct {
				start, end int
				value      string
			}{
				{0, 6, "你好"},  // "你好" is 6 bytes in UTF-8
				{6, 7, " "},   // space is 1 byte
				{7, 13, "世界"}, // "世界" is 6 bytes in UTF-8
			},
		},
		{
			name:  "mixed_spaces",
			input: "  café  naïve  ",
			expected: []struct {
				start, end int
				value      string
			}{
				{0, 2, "  "},     // leading spaces
				{2, 7, "café"},   // "café" is 5 bytes
				{7, 9, "  "},     // middle spaces
				{9, 15, "naïve"}, // "naïve" is 6 bytes
				{15, 17, "  "},   // trailing spaces
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iter := iterators.New(simpleSpaceSplitString, tt.input)

			for i, expected := range tt.expected {
				if !iter.Next() {
					t.Fatalf("expected token %d but Next() returned false", i)
				}

				if iter.Start() != expected.start {
					t.Errorf("token %d: expected start %d, got %d", i, expected.start, iter.Start())
				}
				if iter.End() != expected.end {
					t.Errorf("token %d: expected end %d, got %d", i, expected.end, iter.End())
				}
				if iter.Value() != expected.value {
					t.Errorf("token %d: expected value %q, got %q", i, expected.value, iter.Value())
				}
			}

			if iter.Next() {
				t.Error("expected Next() to return false after all tokens")
			}
		})
	}
}

func TestIterator_Reset(t *testing.T) {
	input := "hello world"
	iter := iterators.New(simpleSpaceSplitString, input)

	// First iteration
	var first []string
	for iter.Next() {
		first = append(first, iter.Value())
	}

	// Reset and iterate again
	iter.Reset()
	var second []string
	for iter.Next() {
		second = append(second, iter.Value())
	}

	if len(first) != len(second) {
		t.Errorf("expected same number of tokens after reset, got %d vs %d", len(first), len(second))
	}

	for i := range first {
		if first[i] != second[i] {
			t.Errorf("token %d: expected %q, got %q", i, first[i], second[i])
		}
	}
}

func TestIterator_SetText(t *testing.T) {
	iter := iterators.New(simpleSpaceSplitString, "hello world")

	// First iteration
	var first []string
	for iter.Next() {
		first = append(first, iter.Value())
	}

	// Set new text
	iter.SetText("foo bar baz")
	var second []string
	for iter.Next() {
		second = append(second, iter.Value())
	}

	expectedFirst := []string{"hello", " ", "world"}
	expectedSecond := []string{"foo", " ", "bar", " ", "baz"}

	if len(first) != len(expectedFirst) {
		t.Errorf("first iteration: expected %d tokens, got %d", len(expectedFirst), len(first))
	}
	if len(second) != len(expectedSecond) {
		t.Errorf("second iteration: expected %d tokens, got %d", len(expectedSecond), len(second))
	}

	for i, expected := range expectedFirst {
		if first[i] != expected {
			t.Errorf("first iteration token %d: expected %q, got %q", i, expected, first[i])
		}
	}
	for i, expected := range expectedSecond {
		if second[i] != expected {
			t.Errorf("second iteration token %d: expected %q, got %q", i, expected, second[i])
		}
	}
}

func TestIterator_EmptyTokens(t *testing.T) {
	// Test that empty tokens are handled correctly
	input := "a b c"
	iter := iterators.New(simpleSpaceSplitString, input)

	var tokens []string
	for iter.Next() {
		tokens = append(tokens, iter.Value())
	}

	expected := []string{"a", " ", "b", " ", "c"}
	if len(tokens) != len(expected) {
		t.Errorf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, expected := range expected {
		if tokens[i] != expected {
			t.Errorf("token %d: expected %q, got %q", i, expected, tokens[i])
		}
	}
}

func TestIterator_ValueBeforeNext(t *testing.T) {
	input := "hello world"
	iter := iterators.New(simpleSpaceSplitString, input)

	// Value() before Next() should return zero value
	var zero string
	if iter.Value() != zero {
		t.Errorf("expected zero value before Next(), got %q", iter.Value())
	}

	// After Next(), should return the actual value
	if !iter.Next() {
		t.Fatal("expected Next() to return true")
	}
	if iter.Value() != "hello" {
		t.Errorf("expected %q, got %q", "hello", iter.Value())
	}
}

func TestIterator_StartEndBeforeNext(t *testing.T) {
	input := "hello world"
	iter := iterators.New(simpleSpaceSplitString, input)

	// Start() and End() before Next() should return 0
	if iter.Start() != 0 {
		t.Errorf("expected Start() to return 0 before Next(), got %d", iter.Start())
	}
	if iter.End() != 0 {
		t.Errorf("expected End() to return 0 before Next(), got %d", iter.End())
	}

	// After Next(), should return actual positions
	if !iter.Next() {
		t.Fatal("expected Next() to return true")
	}
	if iter.Start() != 0 {
		t.Errorf("expected Start() to return 0, got %d", iter.Start())
	}
	if iter.End() != 5 {
		t.Errorf("expected End() to return 5, got %d", iter.End())
	}
}
