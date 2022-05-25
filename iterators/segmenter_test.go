package iterators_test

import (
	"bufio"
	"bytes"
	"math/rand"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/iterators"
	"github.com/clipperhouse/uax29/words"
)

func getRandomBytes() []byte {
	b := make([]byte, 5000)
	rand.Read(b)

	return b
}

func TestSegmenterSameAsScanner(t *testing.T) {
	splits := []bufio.SplitFunc{words.SplitFunc, bufio.ScanWords}
	for _, split := range splits {
		for i := 0; i < 100; i++ {
			text := getRandomBytes()

			seg := iterators.NewSegmenter(split)
			seg.SetText(text)

			r := bytes.NewReader(text)
			sc := iterators.NewScanner(r, split)

			for seg.Next() && sc.Scan() {
				if !bytes.Equal(seg.Bytes(), sc.Bytes()) {
					t.Fatal("Scanner and Segmenter should give identical results")
				}
			}
			if seg.Err() != nil {
				t.Fatal(seg.Err())
			}
			if sc.Err() != nil {
				t.Fatal(sc.Err())
			}
		}
	}
}

func TestSegmenterSameAsAll(t *testing.T) {
	splits := []bufio.SplitFunc{words.SplitFunc, bufio.ScanWords}
	for _, split := range splits {
		for i := 0; i < 100; i++ {
			text := getRandomBytes()

			var all [][]byte
			err := iterators.All(text, &all, split)
			if err != nil {
				t.Fatal(err)
			}

			seg := iterators.NewSegmenter(split)
			seg.SetText(text)

			for i := 0; seg.Next(); i++ {
				if !bytes.Equal(seg.Bytes(), all[i]) {
					t.Fatal("All and Segmenter should give identical results")
				}
			}
			if seg.Err() != nil {
				t.Fatal(seg.Err())
			}
		}
	}
}

func TestSegmenterFilterIsApplied(t *testing.T) {
	text := "Hello, 世界, how are you? Nice dog aha! 👍🐶"

	containsH := func(token []byte) bool {
		pos := 0
		for pos < len(token) {
			r, w := utf8.DecodeRune(token[pos:])
			if unicode.ToLower(r) == 'h' {
				return true
			}
			pos += w
		}

		return false
	}

	seg := iterators.NewSegmenter(bufio.ScanWords)
	seg.SetText([]byte(text))
	seg.Filter(containsH)

	count := 0
	for seg.Next() {
		if !containsH(seg.Bytes()) {
			t.Fatal("filter was not applied")
		}
		count++
	}

	if count != 3 {
		t.Fatalf("segmenter filter should have found 3 results, got %d", count)
	}
}

func TestAllFilterIsApplied(t *testing.T) {
	text := "Hello, 世界, how are you? Nice dog aha! 👍🐶"

	containsH := func(token []byte) bool {
		pos := 0
		for pos < len(token) {
			r, w := utf8.DecodeRune(token[pos:])
			if unicode.ToLower(r) == 'h' {
				return true
			}
			pos += w
		}

		return false
	}

	var all [][]byte
	err := iterators.All([]byte(text), &all, bufio.ScanWords, containsH)
	if err != nil {
		t.Fatal(err)
	}

	for _, seg := range all {
		if !containsH(seg) {
			t.Fatal("filter was not applied")
		}
	}

	if len(all) != 3 {
		t.Fatalf("all filter should have found 3 results, got %d", len(all))
	}
}
