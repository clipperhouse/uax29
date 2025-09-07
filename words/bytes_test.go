package words_test

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/internal/testdata"
	"github.com/clipperhouse/uax29/words"
)

func TestSegmenterUnicode(t *testing.T) {
	t.Parallel()

	// From the Unicode test suite; see the gen/ folder.
	var passed, failed int
	for _, test := range unicodeTests {
		test := test

		var segmented [][]byte
		segmenter := words.FromBytes(test.input)
		for segmenter.Next() {
			segmented = append(segmented, segmenter.Bytes())
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

	if len(unicodeTests) != passed+failed {
		t.Errorf("Incomplete %d tests: passed %d, failed %d", len(unicodeTests), passed, failed)
	}
}

// TestSegmenterRoundtrip tests that all input bytes are output after segmentation.
// De facto, it also tests that we don't get infinite loops, or ever return an error.
func TestSegmenterRoundtrip(t *testing.T) {
	t.Parallel()

	const runs = 2000

	seg := words.FromBytes(nil)

	for i := 0; i < runs; i++ {
		input := getRandomBytes()
		seg.SetText(input)

		var output []byte
		for seg.Next() {
			output = append(output, seg.Bytes()...)
		}

		if !bytes.Equal(output, input) {
			t.Fatal("input bytes are not the same as segmented bytes")
		}
	}
}

func segToSet(seg *words.BytesIterator) map[string]struct{} {
	founds := make(map[string]struct{})
	for seg.Next() {
		founds[string(seg.Bytes())] = struct{}{}
	}
	return founds
}

func TestSegmenterJoiners(t *testing.T) {
	seg1 := words.FromBytes(joinersInput)
	founds1 := segToSet(seg1)

	seg2 := words.FromBytes(joinersInput)
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

	seg := words.FromBytes(input)

	var output []byte
	for seg.Next() {
		output = append(output, seg.Bytes()...)
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
	seg := words.FromBytes(nil)
	benchSeg(b, seg)
}

func benchSeg(b *testing.B, seg *words.BytesIterator) {
	file, err := testdata.Sample()
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

	}

	elapsed := time.Since(start)
	n := float64(b.N)

	tokensPerOp := float64(c) / n
	nsPerOp := float64(elapsed.Nanoseconds()) / n

	b.ReportMetric(1e3*tokensPerOp/nsPerOp, "Mtok/s")
	b.ReportMetric(tokensPerOp, "tok/op")
	b.ReportMetric(float64(bytes)/tokensPerOp, "B/tok")
}

func BenchmarkSegmenterJoiners(b *testing.B) {
	var joiners = &words.Joiners{
		Middle:  []rune("@-/"),
		Leading: []rune("#."),
	}
	seg := words.FromBytes(nil)
	seg.Joiners(joiners)
	benchSeg(b, seg)
}

func BenchmarkUnicodeTests(b *testing.B) {
	var buf bytes.Buffer
	for _, test := range unicodeTests {
		buf.Write(test.input)
	}
	file := buf.Bytes()

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	seg := words.FromBytes(file)

	for i := 0; i < b.N; i++ {
		seg.SetText(file)

		c := 0
		for seg.Next() {
			c++
		}

		b.ReportMetric(float64(c), "tokens")
	}
}
