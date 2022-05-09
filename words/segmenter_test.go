package words_test

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/words"
)

func TestSegmenterUnicode(t *testing.T) {
	// From the Unicode test suite; see the gen/ folder.
	var passed, failed int
	for _, test := range unicodeTests {
		test := test

		var segmented [][]byte
		segmenter := words.NewSegmenter(test.input)
		for segmenter.Next() {
			segmented = append(segmented, segmenter.Bytes())
		}

		if err := segmenter.Err(); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(segmented, test.expected) {
			failed++
			t.Errorf(`
	for input %v
	expected  %v
	got       %v
	spec      %s`, test.input, test.expected, segmented, test.comment)
		} else {
			passed++
		}
	}
	t.Logf("passed %d, failed %d", passed, failed)
}

// TestSegmenterRoundtrip tests that all input bytes are output after segmentation.
// De facto, it also tests that we don't get infinite loops, or ever return an error.
func TestSegmenterRoundtrip(t *testing.T) {
	const runs = 2000

	seg := words.NewSegmenter(nil)

	for i := 0; i < runs; i++ {
		input := getRandomBytes()
		seg.SetText(input)

		var output []byte
		for seg.Next() {
			output = append(output, seg.Bytes()...)
		}

		if err := seg.Err(); err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(output, input) {
			t.Fatal("input bytes are not the same as segmented bytes")
		}
	}
}

func TestSegmenterInvalidUTF8(t *testing.T) {
	// For background, see testdata/UTF-8-test.txt, or:
	// https://www.cl.cam.ac.uk/~mgk25/ucs/examples/UTF-8-test.txt

	// Btw, don't edit UTF-8-test.txt: your editor might turn it into valid UTF-8!

	input, err := ioutil.ReadFile("testdata/UTF-8-test.txt")

	if err != nil {
		t.Error(err)
	}

	if utf8.Valid(input) {
		t.Error("input file should not be valid utf8")
	}

	sc := words.NewSegmenter(input)

	var output []byte
	for sc.Next() {
		output = append(output, sc.Bytes()...)
	}
	if err := sc.Err(); err != nil {
		t.Error(err)
	}

	if !bytes.Equal(output, input) {
		t.Fatalf("input bytes are not the same as segmented bytes")
	}
}

func BenchmarkSegmenter(b *testing.B) {
	file, err := ioutil.ReadFile("testdata/sample.txt")

	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	for i := 0; i < b.N; i++ {
		seg := words.NewSegmenter(file)

		c := 0
		for seg.Next() {
			c++
		}

		if err := seg.Err(); err != nil {
			b.Error(err)
		}

		b.ReportMetric(float64(c), "tokens")
	}
}

func BenchmarkUnicodeSegmenter(b *testing.B) {
	var buf bytes.Buffer
	for _, test := range unicodeTests {
		buf.Write(test.input)
	}
	file := buf.Bytes()

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	for i := 0; i < b.N; i++ {
		sc := words.NewSegmenter(file)

		c := 0
		for sc.Next() {
			c++
		}
		if err := sc.Err(); err != nil {
			b.Error(err)
		}

		b.ReportMetric(float64(c), "tokens")
	}
}
