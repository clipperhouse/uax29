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
			stringSeg := iterators.NewStringSegmenter(split)
			stringSeg.SetText(string(text))

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

			seg := iterators.NewStringSegmenter(split)
			seg.SetText(string(text))

			for i := 0; seg.Next(); i++ {
				expected := all[i]
				got := []byte(seg.Text())
				if !bytes.Equal(expected, got) {
					t.Fatal("All and StringSegmenter should give identical results")
				}
			}
			if seg.Err() != nil {
				t.Fatal(seg.Err())
			}
		}
	}
}

func TestStringSegmenterStart(t *testing.T) {
	t.Parallel()

	text := "Hello world"

	{
		seg := iterators.NewStringSegmenter(words.SplitFunc)
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
		seg := iterators.NewStringSegmenter(bufio.ScanWords)
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
}

func TestStringSegmenterEnd(t *testing.T) {
	t.Parallel()

	text := "Hello world"

	{
		seg := iterators.NewStringSegmenter(words.SplitFunc)
		seg.SetText(text)

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
		seg := iterators.NewStringSegmenter(bufio.ScanWords)
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
}
