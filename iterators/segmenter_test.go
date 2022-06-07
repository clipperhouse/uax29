package iterators_test

import (
	"bufio"
	"bytes"
	"math/rand"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/iterators"
	"github.com/clipperhouse/uax29/iterators/transformer"
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

var startsWithH = func(token []byte) bool {
	r, _ := utf8.DecodeRune(token)
	return unicode.ToLower(r) == 'h'
}

var endsWithW = func(token []byte) bool {
	r, _ := utf8.DecodeLastRune(token)
	return unicode.ToLower(r) == 'w'
}

func TestSegmenterFilterIsApplied(t *testing.T) {
	text := "Hello, 世界, how are you? Nice dog aha! 👍🐶"

	seg := iterators.NewSegmenter(bufio.ScanWords)
	seg.SetText([]byte(text))
	seg.Filter(startsWithH)

	count := 0
	for seg.Next() {
		if !startsWithH(seg.Bytes()) {
			t.Fatal("segmenter filter was not applied")
		}
		count++
	}

	if count != 2 {
		t.Fatalf("segmenter filter should have found 2 results, got %d", count)
	}
}

func TestSegmenterTransformIsApplied(t *testing.T) {
	text := "Hello, 世界, I enjoy Açaí in Örebro."

	seg := iterators.NewSegmenter(bufio.ScanWords)
	seg.SetText([]byte(text))
	seg.Transform(transformer.Lower, transformer.Diacritics)

	var tokens [][]byte
	for seg.Next() {
		tokens = append(tokens, seg.Bytes())
	}

	{
		got := tokens[4]
		expected := []byte("acai")
		if !bytes.Equal(expected, got) {
			t.Fatalf("transforms of lower case or diacritics were not applied, expected %q, got %q", expected, got)
		}
	}

	{
		got := tokens[6]
		expected := []byte("orebro.")
		if !bytes.Equal(expected, got) {
			t.Fatalf("transforms of lower case or diacritics were not applied, expected %q, got %q", expected, got)
		}
	}
}
