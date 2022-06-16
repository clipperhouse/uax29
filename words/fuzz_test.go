//go:build go1.18

package words_test

import (
	"bytes"
	"io/ioutil"
	"testing"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/words"
)

func FuzzWords(f *testing.F) {

	// start with the unicode test suite
	for _, test := range unicodeTests {
		f.Add(test.input)
	}

	// add multi-lingual text
	file, err := ioutil.ReadFile("testdata/sample.txt")
	if err != nil {
		f.Error(err)
	}

	// add as a large one
	f.Add(file)

	// add a bunch of small ones
	lines := bytes.Split(file, []byte("\n"))
	for _, line := range lines {
		f.Add(line)
	}

	// add some random
	f.Add(getRandomBytes())

	// known invalid utf-8
	badUTF8, err := ioutil.ReadFile("testdata/UTF-8-test.txt")
	if err != nil {
		f.Error(err)
	}
	f.Add(badUTF8)

	f.Fuzz(func(t *testing.T, original []byte) {
		var segs [][]byte
		valid1 := utf8.Valid(original)
		seg := words.NewSegmenter(original)
		for seg.Next() {
			segs = append(segs, seg.Bytes())
		}
		if seg.Err() != nil {
			t.Error(seg.Err())
		}

		roundtrip := make([]byte, 0, len(original))
		for _, s := range segs {
			roundtrip = append(roundtrip, s...)
		}

		if !bytes.Equal(roundtrip, original) {
			t.Error("bytes did not roundtrip")
		}

		valid2 := utf8.Valid(roundtrip)

		if valid1 != valid2 {
			t.Error("utf8 validity of original did not match roundtrip")
		}
	})
}
