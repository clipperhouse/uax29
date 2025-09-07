package iterators_test

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/clipperhouse/uax29/internal/iterators"
	"github.com/clipperhouse/uax29/words"
)

func TestStringSegmenterSameAsSegmenter(t *testing.T) {
	t.Parallel()

	text := make([]byte, 50000)

	for _, split := range splitFuncs {
		for i := 0; i < 100; i++ {
			_, err := rand.Read(text)
			if err != nil {
				t.Fatal(err)
			}

			// Test with []byte segmenter
			seg := iterators.NewBytesIterator(split)
			seg.SetText(text)

			// Test with string segmenter
			stringSeg := iterators.NewStringIterator(split)
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
		}
	}
}

func TestStringSegmenterStart(t *testing.T) {
	t.Parallel()

	text := "Hello world"

	{
		seg := iterators.NewStringIterator(words.SplitFunc)
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
		seg := iterators.NewStringIterator(bufio.ScanWords)
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
		seg := iterators.NewStringIterator(words.SplitFunc)
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
}
