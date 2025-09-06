package iterators_test

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/clipperhouse/uax29/graphemes"
	"github.com/clipperhouse/uax29/internal/iterators"
	"github.com/clipperhouse/uax29/phrases"
	"github.com/clipperhouse/uax29/sentences"
	"github.com/clipperhouse/uax29/words"
)

var splitFuncs = map[string]bufio.SplitFunc{
	"words":     words.SplitFunc,
	"sentences": sentences.SplitFunc,
	"graphemes": graphemes.SplitFunc,
	"phrases":   phrases.SplitFunc,
}

func TestSegmenterSameAsAll(t *testing.T) {
	t.Parallel()

	text := make([]byte, 50000)

	for _, split := range splitFuncs {
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
}
