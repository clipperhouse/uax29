package words_test

import (
	"bytes"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/words"
)

func TestScannerUnicode(t *testing.T) {
	// From the Unicode test suite; see the gen/ folder.
	var passed, failed int
	for _, test := range unicodeTests {
		var scanned [][]byte
		scanner := words.NewScanner(bytes.NewReader(test.input))
		for scanner.Scan() {
			scanned = append(scanned, scanner.Bytes())
		}

		if err := scanner.Err(); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(scanned, test.expected) {
			failed++
			t.Errorf(`
	for input %v
	expected  %v
	got       %v
	spec      %s`, test.input, test.expected, scanned, test.comment)
		} else {
			passed++
		}
	}
	t.Logf("%d tests: passed %d, failed %d", len(unicodeTests), passed, failed)
}

// TestScannerRoundtrip tests that all input bytes are output after segmentation.
// De facto, it also tests that we don't get infinite loops, or ever return an error.
func TestScannerRoundtrip(t *testing.T) {
	const runs = 2000

	for i := 0; i < runs; i++ {

		input := getRandomBytes()

		r := bytes.NewReader(input)
		sc := words.NewScanner(r)

		var output []byte
		for sc.Scan() {
			output = append(output, sc.Bytes()...)
		}
		if err := sc.Err(); err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(output, input) {
			t.Fatal("input bytes are not the same as scanned bytes")
		}
	}
}

func TestInvalidUTF8(t *testing.T) {
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

	r := bytes.NewReader(input)
	sc := words.NewScanner(r)

	var output []byte
	for sc.Scan() {
		output = append(output, sc.Bytes()...)
	}

	if err := sc.Err(); err != nil {
		t.Error(err)
	}

	if !bytes.Equal(output, input) {
		t.Fatalf("input bytes are not the same as scanned bytes")
	}
}

func TestNeverZeroAtEOF(t *testing.T) {
	// SplitFunc should never return advance = 0 when atEOF. This test is redundant
	// with the roundtrip test above, but nice to call out this invariant.

	const runs = 100

	for i := 0; i < runs; i++ {
		input := getRandomBytes()

		advance, _, _ := words.SplitFunc(input, true)

		if advance == 0 {
			t.Error("advance should never be zero when atEOF is true")
		}
	}
}

func TestNeverErr(t *testing.T) {
	// SplitFunc should never return an error. This test is redundant
	// with the roundtrip test above, but nice to call out this invariant.

	const runs = 100
	atEOFs := []bool{true, false}

	for i := 0; i < runs; i++ {
		for _, atEOF := range atEOFs {
			input := getRandomBytes()

			_, _, err := words.SplitFunc(input, atEOF)

			if err != nil {
				t.Errorf("SplitFunc should never error (atEOF %t)", atEOF)
			}
		}
	}
}

var seed = time.Now().UnixNano()
var rnd = rand.New(rand.NewSource(seed))

const max = 10000
const min = 1

func getRandomBytes() []byte {
	len := rnd.Intn(max-min) + min
	b := make([]byte, len)
	rand.Read(b)

	return b
}

func BenchmarkScanner(b *testing.B) {
	file, err := os.ReadFile("../testdata/sample.txt")

	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	r := bytes.NewReader(file)

	for i := 0; i < b.N; i++ {
		r.Reset(file)
		sc := words.NewScanner(r)

		c := 0
		for sc.Scan() {
			c++
		}
		if err := sc.Err(); err != nil {
			b.Error(err)
		}

		b.ReportMetric(float64(c), "tokens")
	}
}

func BenchmarkUnicodeSegments(b *testing.B) {
	var buf bytes.Buffer
	for _, test := range unicodeTests {
		buf.Write(test.input)
	}
	file := buf.Bytes()

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	r := bytes.NewReader(file)

	for i := 0; i < b.N; i++ {
		r.Reset(file)
		sc := words.NewScanner(r)

		c := 0
		for sc.Scan() {
			c++
		}
		if err := sc.Err(); err != nil {
			b.Error(err)
		}

		b.ReportMetric(float64(c), "tokens")
	}
}
