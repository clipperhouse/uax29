package words_test

import (
	"bytes"
	"os"
	"reflect"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/iterators"
	"github.com/clipperhouse/uax29/iterators/filter"
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

		// Test SegmentAll while we're here
		all := words.SegmentAll(test.input)
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

func TestSegmenterWordlike(t *testing.T) {
	text := []byte("Hello, 世界. Nice dog! 👍🐶")
	seg := words.NewSegmenter(text)
	seg.Filter(filter.Entirely(unicode.Punct))

	for seg.Next() {
		t.Logf("%q\n", seg.Bytes())
	}
}

func TestSegmenterJoiners(t *testing.T) {
	var config = words.NewConfig().JoinMiddleChars("@-/").JoinLeadingChars("#.")

	set := func(seg *iterators.Segmenter) map[string]struct{} {
		founds := make(map[string]struct{})
		for seg.Next() {
			founds[string(seg.Bytes())] = struct{}{}
		}
		return founds
	}

	text := []byte("Hello, 世界. Tell me about your super-cool .com. I'm .01% interested and 3/4 of a mile away. Email me at foo@example.biz. #winning")

	seg1 := words.NewSegmenter(text)
	founds1 := set(seg1)

	seg2 := words.NewSegmenterConfig(text, config)
	founds2 := set(seg2)

	type test struct {
		input string
		// word should be found in standard, no-config segmenter
		found1 bool
		// word should be found in segmenter configured with joiners
		found2 bool
	}

	tests := []test{
		{"Hello", true, true},
		{"世", true, true},
		{"super", true, false},
		{"-", true, false},
		{"cool", true, false},
		{"super-cool", false, true},
		{"com", true, false}, // ".com" should usually be split, but joined with config
		{".com", false, true},
		{"01", true, false},
		{".01", false, true},
		{"3", true, false},
		{"3/4", false, true},
		{"foo", true, false},
		{"@", true, false},
		{"example.biz", true, false},
		{"foo@example.biz", false, true},
		{"#", true, false},
		{"winning", true, false},
		{"#winning", false, true},
	}

	for _, test := range tests {
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

func TestSegmenterInvalidUTF8(t *testing.T) {
	// For background, see testdata/UTF-8-test.txt, or:
	// https://www.cl.cam.ac.uk/~mgk25/ucs/examples/UTF-8-test.txt

	// Btw, don't edit UTF-8-test.txt: your editor might turn it into valid UTF-8!

	input, err := os.ReadFile("../testdata/UTF-8-test.txt")

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
	file, err := os.ReadFile("../testdata/sample.txt")

	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	bytes := len(file)
	b.SetBytes(int64(bytes))
	seg := words.NewSegmenter(file)

	c := 0
	start := time.Now()

	for i := 0; i < b.N; i++ {
		seg.SetText(file)

		for seg.Next() {
			c++
		}

		if err := seg.Err(); err != nil {
			b.Error(err)
		}
	}

	elapsed := time.Since(start)
	n := float64(b.N)

	tokensPerOp := float64(c) / n
	nsPerOp := float64(elapsed.Nanoseconds()) / n

	b.ReportMetric(1e3*tokensPerOp/nsPerOp, "MMtokens/s")
	b.ReportMetric(tokensPerOp, "tokens/op")
	b.ReportMetric(float64(bytes)/tokensPerOp, "B/token")
}

func BenchmarkSegmenterFilter(b *testing.B) {
	file, err := os.ReadFile("../testdata/sample.txt")

	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	b.SetBytes(int64(len(file)))
	seg := words.NewSegmenter(file)
	seg.Filter(filter.Wordlike)

	for i := 0; i < b.N; i++ {
		seg.SetText(file)

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

func BenchmarkSegmentAll(b *testing.B) {
	file, err := os.ReadFile("../testdata/sample.txt")

	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	for i := 0; i < b.N; i++ {
		segs := words.SegmentAll(file)

		c := 0
		for range segs {
			c++
		}

		b.ReportMetric(float64(c), "tokens")
	}
}

func BenchmarkUnicodeTests(b *testing.B) {
	var buf bytes.Buffer
	for _, test := range unicodeTests {
		buf.Write(test.input)
	}
	file := buf.Bytes()

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	seg := words.NewSegmenter(file)

	for i := 0; i < b.N; i++ {
		seg.SetText(file)

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
