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

func TestGenericIteratorWithJoiners(t *testing.T) {
	t.Parallel()

	// Test with []byte and joiners
	text := []byte("Hello, 世界. Tell me about your super-cool .com. I'm .01% interested and 3/4 of a mile away. Email me at foo@example.biz. #winning")
	iter := words.FromBytes(text)

	joiners := &words.Joiners[[]byte]{
		Middle:  []rune("@-/"),
		Leading: []rune("#."),
	}
	iter.Joiners(joiners)

	var results []string
	for iter.Next() {
		results = append(results, string(iter.Value()))
	}

	// Check that some specific joined tokens exist
	expectedJoined := []string{"super-cool", ".com", ".01", "3/4", "foo@example.biz", "#winning"}
	for _, expected := range expectedJoined {
		found := false
		for _, result := range results {
			if result == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected joined token %q not found in results", expected)
		}
	}

	// Test with string and joiners
	textStr := string(text)
	iterStr := words.FromString(textStr)

	joinersStr := &words.Joiners[string]{
		Middle:  []rune("@-/"),
		Leading: []rune("#."),
	}
	iterStr.Joiners(joinersStr)

	var resultsStr []string
	for iterStr.Next() {
		resultsStr = append(resultsStr, iterStr.Value())
	}

	// Check that some specific joined tokens exist
	for _, expected := range expectedJoined {
		found := false
		for _, result := range resultsStr {
			if result == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected joined token %q not found in string results", expected)
		}
	}
}
