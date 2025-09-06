package phrases_test

import (
	"bytes"
	"crypto/rand"
	mathrand "math/rand"
	"os"
	"testing"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/internal/testdata"
	"github.com/clipperhouse/uax29/phrases"
)

// TestScannerRoundtrip tests that all input bytes are output after segmentation.
// De facto, it also tests that we don't get infinite loops, or ever return an error.
func TestScannerRoundtrip(t *testing.T) {
	t.Parallel()

	const runs = 2000

	for i := 0; i < runs; i++ {

		input := getRandomBytes()

		r := bytes.NewReader(input)
		sc := phrases.NewScanner(r)

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

	r := bytes.NewReader(input)
	sc := phrases.NewScanner(r)

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
	t.Parallel()

	// SplitFunc should never return advance = 0 when atEOF. This test is redundant
	// with the roundtrip test above, but nice to call out this invariant.

	const runs = 100

	for i := 0; i < runs; i++ {
		input := getRandomBytes()

		advance, _, _ := phrases.SplitFunc(input, true)

		if advance == 0 {
			t.Error("advance should never be zero when atEOF is true")
		}
	}
}

func TestNeverErr(t *testing.T) {
	t.Parallel()

	// SplitFunc should never return an error. This test is redundant
	// with the roundtrip test above, but nice to call out this invariant.

	const runs = 100
	atEOFs := []bool{true, false}

	for i := 0; i < runs; i++ {
		for _, atEOF := range atEOFs {
			input := getRandomBytes()

			_, _, err := phrases.SplitFunc(input, atEOF)

			if err != nil {
				t.Errorf("SplitFunc should never error (atEOF %t)", atEOF)
			}
		}
	}
}

func getRandomBytes() []byte {
	const max = 10000
	const min = 1

	len := mathrand.Intn(max-min) + min
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return b
}

func BenchmarkScanner(b *testing.B) {
	file, err := testdata.Sample()

	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	r := bytes.NewReader(file)

	for i := 0; i < b.N; i++ {
		r.Reset(file)
		sc := phrases.NewScanner(r)

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
