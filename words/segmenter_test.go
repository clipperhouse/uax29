package words_test

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/iterators/filter"
	"github.com/clipperhouse/uax29/words"
)

func TestSegmenterUnicode(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	text := []byte("Hello, 世界. Nice dog! 👍🐶")
	seg := words.NewSegmenter(text)
	seg.Filter(filter.Entirely(unicode.Punct))

	for seg.Next() {
		t.Logf("%q\n", seg.Bytes())
	}
}

func segToSet(seg *words.Segmenter) map[string]struct{} {
	founds := make(map[string]struct{})
	for seg.Next() {
		founds[string(seg.Bytes())] = struct{}{}
	}
	return founds
}

func TestSegmenterJoiners(t *testing.T) {
	seg1 := words.NewSegmenter(joinersInput)
	founds1 := segToSet(seg1)

	seg2 := words.NewSegmenter(joinersInput)
	seg2.Joiners(joiners)
	founds2 := segToSet(seg2)

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

func TestSegmenterInvalidUTF8(t *testing.T) {
	t.Parallel()

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

var example = ", ;👍-@/.#"
var searches = []rune{' ', 'x', '@', '👍'}

func BenchmarkArray(b *testing.B) {
	var array []rune
	for _, r := range example {
		array = append(array, r)
	}

	var found rune
	for i := 0; i < b.N; i++ {
		for _, s := range searches {
			for j := range array {
				if array[j] == s {
					found = s
				}
			}
		}
	}

	fmt.Println(found)
}

func BenchmarkMap(b *testing.B) {
	m := make(map[rune]struct{})
	for _, r := range example {
		m[r] = struct{}{}
	}

	var found rune
	for i := 0; i < b.N; i++ {
		for _, s := range searches {
			if _, ok := m[s]; ok {
				found = s
			}
		}
	}

	fmt.Println(found)
}

func BenchmarkSegmenter(b *testing.B) {
	seg := words.NewSegmenter(nil)
	benchSeg(b, seg)
}

func benchSeg(b *testing.B, seg *words.Segmenter) {
	file, err := os.ReadFile("../testdata/sample.txt")
	if err != nil {
		b.Error(err)
	}

	bytes := len(file)
	b.SetBytes(int64(bytes))

	c := 0
	b.ResetTimer()
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

	b.ReportMetric(1e3*tokensPerOp/nsPerOp, "Mtok/s")
	b.ReportMetric(tokensPerOp, "tok/op")
	b.ReportMetric(float64(bytes)/tokensPerOp, "B/tok")
}

func BenchmarkSegmenterFilter(b *testing.B) {
	seg := words.NewSegmenter(nil)
	seg.Filter(filter.Wordlike)
	benchSeg(b, seg)
}

func BenchmarkSegmenterJoiners(b *testing.B) {
	var joiners = &words.Joiners{
		Middle:  []rune("@-/"),
		Leading: []rune("#."),
	}
	seg := words.NewSegmenter(nil)
	seg.Joiners(joiners)
	benchSeg(b, seg)
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
