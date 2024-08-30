//go:build go1.23
// +build go1.23

package iterators_test

import (
	"errors"
	"io"
	"iter"
	"os"
	"reflect"
	"testing"

	"github.com/clipperhouse/uax29/graphemes"
	"github.com/clipperhouse/uax29/iterators"
	"github.com/clipperhouse/uax29/phrases"
	"github.com/clipperhouse/uax29/sentences"
	"github.com/clipperhouse/uax29/words"
)

type iterSplitFunc func(data []byte) iter.Seq[[]byte]

var iterSplitFuncs = []iterSplitFunc{words.Split, sentences.Split, graphemes.Split, phrases.Split}

func TestIterMatchesSegmenter(t *testing.T) {
	t.Parallel()

	if len(splitFuncs) != len(iterSplitFuncs) {
		t.Fatal("need equal number of splitFunc and iterSplitFunc")
	}

	file, err := os.ReadFile("../testdata/sample.txt")
	if err != nil {
		t.Fatal(err)
	}

	for i := range iterSplitFuncs {
		iterFunc := iterSplitFuncs[i]
		splitFunc := splitFuncs[i]

		seg1 := iterators.NewSegmenter(splitFunc)
		seg1.SetText(file)
		var expected [][]byte
		for seg1.Next() {
			expected = append(expected, seg1.Bytes())
		}

		iter := iterFunc(file)
		var got [][]byte
		for word := range iter {
			got = append(got, word)
		}

		if len(got) == 0 || len(expected) != len(got) {
			t.Fatal("iter and segmenter return different lengths")
		}

		if !reflect.DeepEqual(expected, got) {
			t.Fatal("iter and segmenter return different results")
		}
	}
}

type iterScanFunc func(io.Reader) iter.Seq2[[]byte, error]

var iterScanFuncs = []iterScanFunc{words.Scan, sentences.Scan, graphemes.Scan, phrases.Scan}

func TestIterMatchesScanner(t *testing.T) {
	t.Parallel()

	if len(splitFuncs) != len(iterSplitFuncs) {
		t.Fatal("need equal number of splitFunc and iterScanFunc")
	}

	for i := range iterSplitFuncs {
		iterFunc := iterScanFuncs[i]
		splitFunc := splitFuncs[i]

		file1, err := os.Open("../testdata/sample.txt")
		if err != nil {
			t.Fatal(err)
		}

		sc1 := iterators.NewScanner(file1, splitFunc)
		var expected [][]byte
		for sc1.Scan() {
			expected = append(expected, sc1.Bytes())
		}
		if err := sc1.Err(); err != nil {
			t.Fatal(err)
		}
		file1.Close()

		file2, err := os.Open("../testdata/sample.txt")
		if err != nil {
			t.Fatal(err)
		}

		iter := iterFunc(file2)

		var got [][]byte
		for word, err := range iter {
			got = append(got, word)
			if err != nil {
				t.Fatal(err)
			}
		}
		file2.Close()

		if len(got) == 0 || len(expected) != len(got) {
			t.Fatal("iter and scanner return different results")
		}

		if !reflect.DeepEqual(expected, got) {
			t.Fatal("iter and scanner return different results")
		}
	}
}

func TestScannerIterErr(t *testing.T) {
	file1, err := os.Open("../testdata/sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file1.Close()

	e := "hello error"
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		return 1, nil, errors.New(e)
	}

	sc := iterators.NewScanner(file1, split)

	for _, err := range sc.All() {
		if err == nil {
			t.Fatal("iter should have returned an error")
		}
		if err.Error() != e {
			t.Fatalf("iter should have returned %q, got %q", e, err)
		}
	}
}
