package iterators_test

import (
	"testing"

	"github.com/clipperhouse/uax29/v2/words"
)

func TestGenericIterator(t *testing.T) {
	t.Parallel()

	// Test with []byte
	text := []byte("Hello, 世界. Nice dog! 👍🐶")
	iter := words.FromBytes(text)

	var results []string
	for iter.Next() {
		results = append(results, string(iter.Value()))
	}

	expected := []string{"Hello", ",", " ", "世", "界", ".", " ", "Nice", " ", "dog", "!", " ", "👍", "🐶"}
	if len(results) != len(expected) {
		t.Logf("Results: %v", results)
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(results))
	}

	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Token %d: expected %q, got %q", i, expected[i], result)
		}
	}

	// Test with string
	textStr := string(text)
	iterStr := words.FromString(textStr)

	var resultsStr []string
	for iterStr.Next() {
		resultsStr = append(resultsStr, iterStr.Value())
	}

	if len(resultsStr) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(resultsStr))
	}

	for i, result := range resultsStr {
		if result != expected[i] {
			t.Errorf("Token %d: expected %q, got %q", i, expected[i], result)
		}
	}
}
