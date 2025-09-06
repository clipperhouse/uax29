//go:build go1.23
// +build go1.23

package iterators_test

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/clipperhouse/uax29/iterators"
)

func TestIterMatchesSegmenter(t *testing.T) {
	t.Parallel()

	file, err := os.ReadFile("../testdata/sample.txt")
	if err != nil {
		t.Fatal(err)
	}

	for _, splitFunc := range splitFuncs {
		seg1 := iterators.NewSegmenter(splitFunc)
		seg1.SetText(file)
		var expected [][]byte
		for seg1.Next() {
			expected = append(expected, seg1.Bytes())
		}

		seg2 := iterators.NewSegmenter(splitFunc)
		seg2.SetText(file)
		var got [][]byte
		for token := range seg2.Iter() {
			got = append(got, token.Value())
		}

		if len(got) == 0 || len(expected) != len(got) {
			t.Fatal("iter and segmenter returned different lengths")
		}

		if !reflect.DeepEqual(expected, got) {
			t.Fatal("iter and segmenter returned different results")
		}
	}
}

func TestIterMatchesStringSegmenter(t *testing.T) {
	t.Parallel()

	file, err := os.ReadFile("../testdata/sample.txt")
	if err != nil {
		t.Fatal(err)
	}

	s := string(file)

	for _, splitFunc := range splitFuncs {
		seg1 := iterators.NewStringSegmenter(splitFunc)
		seg1.SetText(s)
		var expected []string
		for seg1.Next() {
			expected = append(expected, seg1.Text())
		}

		seg2 := iterators.NewStringSegmenter(splitFunc)
		seg2.SetText(s)
		var got []string
		for token := range seg2.Iter() {
			got = append(got, token.Value())
		}

		if len(got) == 0 || len(expected) != len(got) {
			t.Fatal("iter and segmenter returned different lengths")
		}

		if !reflect.DeepEqual(expected, got) {
			t.Fatal("iter and segmenter returned different results")
		}
	}
}

func TestIterMatchesScanner(t *testing.T) {
	t.Parallel()

	for pkg, splitFunc := range splitFuncs {
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
			t.Logf("pkg: %s", pkg)
			t.Fatal(err)
		}
		file1.Close()

		file2, err := os.Open("../testdata/sample.txt")
		if err != nil {
			t.Fatal(err)
		}

		sc2 := iterators.NewScanner(file2, splitFunc)
		var got [][]byte
		for word, err := range sc2.Iter() {
			got = append(got, word.Value())
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

	for _, err := range sc.Iter() {
		if err == nil {
			t.Fatal("iter should have returned an error")
		}
		if err.Error() != e {
			t.Fatalf("iter should have returned error %q, got %q", e, err)
		}
	}
}
