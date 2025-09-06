package phrases_test

import (
	"bytes"
	"os"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/internal/iterators"
	"github.com/clipperhouse/uax29/phrases"
)

// TestSegmenterRoundtrip tests that all input bytes are output after segmentation.
// De facto, it also tests that we don't get infinite loops, or ever return an error.
func TestSegmenterRoundtrip(t *testing.T) {
	t.Parallel()

	const runs = 2000

	seg := phrases.NewSegmenter(nil)

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

	sc := phrases.NewSegmenter(input)

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

var exists = struct{}{}

func segToSetTrimmed(seg *iterators.Segmenter) map[string]struct{} {
	founds := make(map[string]struct{})
	for seg.Next() {
		key := bytes.TrimSpace(seg.Bytes())
		founds[string(key)] = exists
	}
	return founds
}

func TestPhraseBoundaries(t *testing.T) {
	t.Parallel()

	input := []byte("This should break here. And then here. 世界. I think, perhaps you can understand that — aside 🏆 🐶 here — “a quote”.")
	seg := phrases.NewSegmenter(input)
	got := segToSetTrimmed(seg)
	expecteds := map[string]struct{}{
		"This should break here":          exists,
		"And then here":                   exists,
		"世":                               exists, // We don't have great logic for languages without spaces. Also true for words, see Notes: https://unicode.org/reports/tr29/#WB999
		"I think":                         exists,
		"perhaps you can understand that": exists,
		"aside 🏆 🐶 here":                  exists,
		"a quote":                         exists,
	}

	for phrase := range expecteds {
		_, found := got[phrase]
		if !found {
			t.Fatalf("phrase %q was expected, not found", phrase)
		}
	}
}

func BenchmarkSegmenter(b *testing.B) {
	file, err := os.ReadFile("../internal/testdata/sample.txt")

	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	bytes := len(file)
	b.SetBytes(int64(bytes))
	seg := phrases.NewSegmenter(file)

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

func BenchmarkSegmentAll(b *testing.B) {
	file, err := os.ReadFile("../internal/testdata/sample.txt")

	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	for i := 0; i < b.N; i++ {
		segs := phrases.SegmentAll(file)

		c := 0
		for range segs {
			c++
		}

		b.ReportMetric(float64(c), "tokens")
	}
}
