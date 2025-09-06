package words_test

import (
	"bytes"
	"os"
	"reflect"
	"testing"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/words"
)

func TestStringSegmenterUnicode(t *testing.T) {
	t.Parallel()

	// From the Unicode test suite; see the gen/ folder.
	var passed, failed int
	for _, test := range unicodeTests {
		test := test

		var segmented []string
		segmenter := words.NewStringSegmenter(string(test.input))
		for segmenter.Next() {
			segmented = append(segmented, segmenter.Text())
		}

		if err := segmenter.Err(); err != nil {
			t.Fatal(err)
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

		// Test SegmentAll while we're here
		all := words.SegmentAllString(string(test.input))
		if !reflect.DeepEqual(all, segmented) {
			t.Error("calling SegmentAll should be identical to iterating Segmenter")
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
		seg := words.NewStringSegmenter(input)

		var output string
		for seg.Next() {
			output += seg.Text()
		}

		if err := seg.Err(); err != nil {
			t.Fatal(err)
		}

		if output != input {
			t.Fatal("input bytes are not the same as segmented bytes")
		}
	}
}

func stringSegToSet(seg *words.StringSegmenter) map[string]struct{} {
	founds := make(map[string]struct{})
	for seg.Next() {
		founds[seg.Text()] = struct{}{}
	}
	return founds
}

func TestStringSegmenterJoiners(t *testing.T) {
	s := string(joinersInput)
	seg1 := words.NewStringSegmenter(s)
	founds1 := stringSegToSet(seg1)

	seg2 := words.NewStringSegmenter(s)
	seg2.Joiners(joiners)
	founds2 := stringSegToSet(seg2)

	for _, test := range joinersTests {
		_, found1 := founds1[test.input]
		if found1 != test.found1 {
			t.Fatalf("For %q, expected %t for found in non-config segmenter, but got %t", test.input, test.found1, found1)
		}
		_, found2 := founds2[test.input]
		if found2 != test.found2 {
			t.Fatalf("For %q, expected %t for found in segmenter with joiners, but got %t", test.input, test.found2, found2)
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

	sc := words.NewStringSegmenter(string(input))

	var output string
	for sc.Next() {
		output += sc.Text()
	}
	if err := sc.Err(); err != nil {
		t.Error(err)
	}

	if output != string(input) {
		t.Fatalf("input bytes are not the same as segmented bytes")
	}
}

func BenchmarkStringSegmenter(b *testing.B) {
	file, err := os.ReadFile("../internal/testdata/sample.txt")
	if err != nil {
		b.Error(err)
	}

	s := string(file)

	b.ResetTimer()
	b.SetBytes(int64(len(file)))
	seg := words.NewStringSegmenter(s)

	for i := 0; i < b.N; i++ {
		seg.SetText(s)

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

func BenchmarkStringSegmentAll(b *testing.B) {
	file, err := os.ReadFile("../internal/testdata/sample.txt")
	if err != nil {
		b.Error(err)
	}

	b.SetBytes(int64(len(file)))
	s := string(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = words.SegmentAllString(s)
	}

	c := len(words.SegmentAllString(s))
	b.ReportMetric(float64(c), "tokens")
	b.Logf("tokens %d, len %d, avg %d", c, len(file), len(file)/c)
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

	seg := words.NewStringSegmenter(s)

	for i := 0; i < b.N; i++ {
		seg.SetText(s)

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
