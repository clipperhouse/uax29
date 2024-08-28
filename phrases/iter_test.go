//go:build go1.23
// +build go1.23

package phrases_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/clipperhouse/uax29/phrases"
)

func TestIterMatchesSegmenter(t *testing.T) {
	file, err := os.ReadFile("../testdata/sample.txt")
	if err != nil {
		t.Fatal(err)
	}

	seg1 := phrases.NewSegmenter(file)
	var expected [][]byte
	for seg1.Next() {
		expected = append(expected, seg1.Bytes())
	}

	iter := phrases.Split(file)
	var got [][]byte
	for word := range iter {
		got = append(got, word)
	}

	if len(got) == 0 || len(expected) != len(got) {
		t.Fatal("phrases iter and segmenter return different lengths")
	}

	if !reflect.DeepEqual(expected, got) {
		t.Fatal("phrases iter and segmenter return different results")
	}
}

func TestIterMatchesScanner(t *testing.T) {
	file1, err := os.Open("../testdata/sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file1.Close()

	sc1 := phrases.NewScanner(file1)
	var expected [][]byte
	for sc1.Scan() {
		expected = append(expected, sc1.Bytes())
	}
	if err := sc1.Err(); err != nil {
		t.Fatal(err)
	}

	file2, err := os.Open("../testdata/sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file2.Close()

	iter := phrases.Scan(file2)

	var got [][]byte
	for word, err := range iter {
		got = append(got, word)
		if err != nil {
			t.Fatal(err)
		}
	}

	if len(got) == 0 || len(expected) != len(got) {
		t.Fatal("phrases iter and scanner return different results")
	}

	if !reflect.DeepEqual(expected, got) {
		t.Fatal("phrases iter and scanner return different results")
	}
}
