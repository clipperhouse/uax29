package iterators_test

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/clipperhouse/uax29/graphemes"
	"github.com/clipperhouse/uax29/iterators"
	"github.com/clipperhouse/uax29/phrases"
	"github.com/clipperhouse/uax29/sentences"
	"github.com/clipperhouse/uax29/words"
)

var stringSplitFuncs = map[string]bufio.SplitFunc{
	"words":     words.SplitFunc,
	"sentences": sentences.SplitFunc,
	"graphemes": graphemes.SplitFunc,
	"phrases":   phrases.SplitFunc,
}

func TestStringSegmenterSameAsSegmenter(t *testing.T) {
	t.Parallel()

	text := make([]byte, 50000)

	for _, split := range stringSplitFuncs {
		for i := 0; i < 100; i++ {
			_, err := rand.Read(text)
			if err != nil {
				t.Fatal(err)
			}

			// Test with []byte segmenter
			seg := iterators.NewSegmenter(split)
			seg.SetText(text)

			// Test with string segmenter
			stringSeg := iterators.NewStringSegmenter(string(text), split)

			for seg.Next() && stringSeg.Next() {
				segBytes := seg.Bytes()
				stringBytes := []byte(stringSeg.Text())
				if !bytes.Equal(segBytes, stringBytes) {
					t.Fatalf(`
					StringSegmenter and Segmenter should give identical results
					Segmenter:       %q
					StringSegmenter: %q
					`, segBytes, stringBytes)
				}
			}
			if seg.Err() != nil {
				t.Fatal(seg.Err())
			}
			if stringSeg.Err() != nil {
				t.Fatal(stringSeg.Err())
			}
		}
	}
}

func TestStringSegmenterSameAsAll(t *testing.T) {
	t.Parallel()

	text := make([]byte, 50000)

	for _, split := range stringSplitFuncs {
		for i := 0; i < 100; i++ {
			_, err := rand.Read(text)
			if err != nil {
				t.Fatal(err)
			}

			var all [][]byte
			err = iterators.All(text, &all, split)
			if err != nil {
				t.Fatal(err)
			}

			stringSeg := iterators.NewStringSegmenter(string(text), split)

			for i := 0; stringSeg.Next(); i++ {
				expected := all[i]
				got := []byte(stringSeg.Text())
				if !bytes.Equal(expected, got) {
					t.Fatal("All and StringSegmenter should give identical results")
				}
			}
			if stringSeg.Err() != nil {
				t.Fatal(stringSeg.Err())
			}
		}
	}
}

func TestStringSegmenterStart(t *testing.T) {
	t.Parallel()

	text := "Hello world"

	{
		stringSeg := iterators.NewStringSegmenter(text, words.SplitFunc)
		expected := []int{0, 5, 6}
		var got []int
		for stringSeg.Next() {
			got = append(got, stringSeg.Start())
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("start failed for words.SplitFunc, expected %v, got %v", expected, got)
		}
	}

	{
		stringSeg := iterators.NewStringSegmenter(text, bufio.ScanWords)
		expected := []int{0, 6}
		var got []int
		for stringSeg.Next() {
			got = append(got, stringSeg.Start())
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("start failed for bufio.ScanWords, expected %v, got %v", expected, got)
		}
	}
}

func TestStringSegmenterEnd(t *testing.T) {
	t.Parallel()

	text := "Hello world"

	{
		stringSeg := iterators.NewStringSegmenter(text, words.SplitFunc)

		expected := []int{5, 6, len(text)}
		var got []int
		for stringSeg.Next() {
			got = append(got, stringSeg.End())
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("end failed for words.SplitFunc, expected %v, got %v", expected, got)
		}
	}

	{
		stringSeg := iterators.NewStringSegmenter(text, bufio.ScanWords)
		// bufio.ScanWords includes the space in the first token, so "Hello " ends at position 6
		expected := []int{6, len(text)}
		var got []int
		for stringSeg.Next() {
			got = append(got, stringSeg.End())
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("end failed for bufio.ScanWords, expected %v, got %v", expected, got)
		}
	}
}

func TestStringSegmenterSetText(t *testing.T) {
	t.Parallel()

	stringSeg := iterators.NewStringSegmenter("", graphemes.SplitFunc)

	// Test with first text
	stringSeg.SetText("Hello")
	var results1 []string
	for stringSeg.Next() {
		results1 = append(results1, stringSeg.Text())
	}

	// Test with second text
	stringSeg.SetText("世界")
	var results2 []string
	for stringSeg.Next() {
		results2 = append(results2, stringSeg.Text())
	}

	if len(results1) != 5 {
		t.Errorf("expected 5 graphemes for 'Hello', got %d", len(results1))
	}
	if len(results2) != 2 {
		t.Errorf("expected 2 graphemes for '世界', got %d", len(results2))
	}
}

func TestStringSegmenterWithGraphemes(t *testing.T) {
	t.Parallel()

	text := "Hello, 世界! 👍"
	stringSeg := iterators.NewStringSegmenter(text, graphemes.SplitFunc)

	var results []string
	for stringSeg.Next() {
		results = append(results, stringSeg.Text())
	}

	expected := []string{"H", "e", "l", "l", "o", ",", " ", "世", "界", "!", " ", "👍"}
	if len(results) != len(expected) {
		t.Fatalf("expected %d graphemes, got %d", len(expected), len(results))
	}

	for i, result := range results {
		if result != expected[i] {
			t.Errorf("grapheme %d: expected %q, got %q", i, expected[i], result)
		}
	}
}

func TestStringSegmenterWithWords(t *testing.T) {
	t.Parallel()

	text := "Hello world 世界"
	stringSeg := iterators.NewStringSegmenter(text, words.SplitFunc)

	var results []string
	for stringSeg.Next() {
		results = append(results, stringSeg.Text())
	}

	// words.SplitFunc includes spaces and punctuation as separate tokens
	expected := []string{"Hello", " ", "world", " ", "世", "界"}
	if len(results) != len(expected) {
		t.Fatalf("expected %d words, got %d: %v", len(expected), len(results), results)
	}

	for i, result := range results {
		if result != expected[i] {
			t.Errorf("word %d: expected %q, got %q", i, expected[i], result)
		}
	}
}

func TestStringSegmenterWithScanWords(t *testing.T) {
	t.Parallel()

	text := "Hello world 世界"
	stringSeg := iterators.NewStringSegmenter(text, bufio.ScanWords)

	var results []string
	for stringSeg.Next() {
		results = append(results, stringSeg.Text())
	}

	// bufio.ScanWords skips spaces and punctuation
	expected := []string{"Hello", "world", "世界"}
	if len(results) != len(expected) {
		t.Fatalf("expected %d words, got %d: %v", len(expected), len(results), results)
	}

	for i, result := range results {
		// Trim any trailing spaces that might be included
		trimmed := result
		if len(trimmed) > 0 && trimmed[len(trimmed)-1] == ' ' {
			trimmed = trimmed[:len(trimmed)-1]
		}
		if trimmed != expected[i] {
			t.Errorf("word %d: expected %q, got %q (trimmed from %q)", i, expected[i], trimmed, result)
		}
	}
}
