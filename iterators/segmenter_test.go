package iterators_test

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"reflect"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/iterators"
	"github.com/clipperhouse/uax29/iterators/filter"
	"github.com/clipperhouse/uax29/iterators/transformer"
	"github.com/clipperhouse/uax29/words"
)

func TestSegmenterSameAsScanner(t *testing.T) {
	t.Parallel()

	text := make([]byte, 50000)

	for _, split := range splits {
		for i := 0; i < 100; i++ {
			_, err := rand.Read(text)
			if err != nil {
				t.Fatal(err)
			}

			seg := iterators.NewSegmenter(split)
			seg.SetText(text)

			r := bytes.NewReader(text)
			sc := iterators.NewScanner(r, split)

			for seg.Next() && sc.Scan() {
				if !bytes.Equal(seg.Bytes(), sc.Bytes()) {

					t.Fatalf(`
					Scanner and Segmenter should give identical results
					Scanner:   %q
					Segmenter: %q
					`, sc.Bytes(), seg.Bytes())
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
	t.Parallel()

	text := make([]byte, 50000)

	for _, split := range splits {
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

func TestSegmenterFilterIsApplied(t *testing.T) {
	t.Parallel()

	text := "Hello, 世界, how are you? Nice dog aha! 👍🐶"

	seg := iterators.NewSegmenter(words.SplitFunc)
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
	t.Parallel()

	text := "Hello, 世界, I am enjoying cups of Açaí in Örebro."

	seg := iterators.NewSegmenter(bufio.ScanWords)
	seg.SetText([]byte(text))
	seg.Transform(transformer.Lower, transformer.Diacritics)

	var tokens [][]byte
	for seg.Next() {
		tokens = append(tokens, seg.Bytes())
	}

	{
		got := tokens[7]
		expected := []byte("acai")
		if !bytes.Equal(expected, got) {
			t.Fatalf("transforms of lower case or diacritics were not applied, expected %q, got %q", expected, got)
		}
	}
}

func TestSegmenterStart(t *testing.T) {
	t.Parallel()

	text := []byte("Hello world")

	{
		seg := words.NewSegmenter(text)
		seg.SetText(text)
		expected := []int{0, 5, 6}
		var got []int
		for seg.Next() {
			got = append(got, seg.Start())
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("start failed for words.SplitFunc, expected %v, got %v", expected, got)
		}
	}

	{
		seg := iterators.NewSegmenter(bufio.ScanWords)
		seg.SetText(text)
		expected := []int{0, 6}
		var got []int
		for seg.Next() {
			got = append(got, seg.Start())
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("start failed for bufio.ScanWords, expected %v, got %v", expected, got)
		}
	}

	{
		seg := words.NewSegmenter(text)
		seg.Filter(filter.AlphaNumeric)
		expected := []int{0, 6}
		var got []int
		for seg.Next() {
			got = append(got, seg.Start())
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("start failed for filter.AlphaNumeric, expected %v, got %v", expected, got)
		}
	}
}

func TestSegmenterEnd(t *testing.T) {
	t.Parallel()

	text := []byte("Hello world")

	{
		seg := words.NewSegmenter(text)

		expected := []int{5, 6, len(text)}
		var got []int
		for seg.Next() {
			got = append(got, seg.End())
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("end failed for words.SplitFunc, expected %v, got %v", expected, got)
		}
	}

	{
		seg := iterators.NewSegmenter(bufio.ScanWords)
		seg.SetText(text)
		expected := []int{5, len(text)}
		var got []int
		for seg.Next() {
			got = append(got, seg.End())
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("end failed for bufio.ScanWords, expected %v, got %v", expected, got)
		}
	}
	{
		seg := words.NewSegmenter(text)
		seg.Filter(filter.AlphaNumeric)

		expected := []int{5, len(text)}
		var got []int
		for seg.Next() {
			got = append(got, seg.End())
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("end failed for filter.AlphaNumeric, expected %v, got %v", expected, got)
		}
	}
}
