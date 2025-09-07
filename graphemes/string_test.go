package graphemes_test

import (
	"bytes"
	"os"
	"reflect"
	"testing"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/graphemes"
	"github.com/clipperhouse/uax29/internal/testdata"
)

func TestStringSegmenterUnicode(t *testing.T) {
	t.Parallel()

	// From the Unicode test suite; see the gen/ folder.
	var passed, failed int
	for _, test := range unicodeTests {
		test := test

		var segmented []string
		segmenter := graphemes.FromString(string(test.input))
		for segmenter.Next() {
			segmented = append(segmented, segmenter.Text())
		}

		expected := make([]string, len(test.expected))
		for i, v := range test.expected {
			expected[i] = string(v)
		}

		if !reflect.DeepEqual(segmented, expected) {
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

	if len(unicodeTests) != passed+failed {
		t.Errorf("Incomplete %d tests: passed %d, failed %d", len(unicodeTests), passed, failed)
	}
}

// TestSegmenterRoundtrip tests that all input bytes are output after segmentation.
// De facto, it also tests that we don't get infinite loops, or ever return an error.
func TestStringSegmenterRoundtrip(t *testing.T) {
	t.Parallel()

	const runs = 2000

	for i := 0; i < runs; i++ {
		input := string(getRandomBytes())
		seg := graphemes.FromString(input)

		var output string
		for seg.Next() {
			output += seg.Text()
		}

		if output != input {
			t.Fatal("input bytes are not the same as segmented bytes")
		}
	}
}

func TestStringSegmenterInvalidUTF8(t *testing.T) {
	t.Parallel()

	// For background, see internal/testdata/UTF-8-test.txt, or:
	// https://www.cl.cam.ac.uk/~mgk25/ucs/examples/UTF-8-test.txt

	// Btw, don't edit UTF-8-test.txt: your editor might turn it into valid UTF-8!

	input, err := os.ReadFile("../internal/testdata/UTF-8-test.txt")

	if err != nil {
		t.Error(err)
	}

	if utf8.Valid(input) {
		t.Error("input file should not be valid utf8")
	}

	sc := graphemes.FromString(string(input))

	var output string
	for sc.Next() {
		output += sc.Text()
	}

	if output != string(input) {
		t.Fatalf("input bytes are not the same as segmented bytes")
	}
}

func BenchmarkStringSegmenter(b *testing.B) {
	file, err := testdata.Sample()
	if err != nil {
		b.Error(err)
	}

	s := string(file)

	b.ResetTimer()
	b.SetBytes(int64(len(file)))
	seg := graphemes.FromString(s)

	for i := 0; i < b.N; i++ {
		seg.SetText(s)

		c := 0
		for seg.Next() {
			c++
		}

		b.ReportMetric(float64(c), "tokens")
	}
}

func BenchmarkStringUnicodeTests(b *testing.B) {
	var buf bytes.Buffer
	for _, test := range unicodeTests {
		buf.Write(test.input)
	}
	file := buf.Bytes()
	s := string(file)

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	seg := graphemes.FromString(s)

	for i := 0; i < b.N; i++ {
		seg.SetText(s)

		c := 0
		for seg.Next() {
			c++
		}

		b.ReportMetric(float64(c), "tokens")
	}
}
