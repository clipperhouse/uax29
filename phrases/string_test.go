package phrases_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/internal/testdata"
	"github.com/clipperhouse/uax29/phrases"
)

// TestSegmenterRoundtrip tests that all input bytes are output after segmentation.
// De facto, it also tests that we don't get infinite loops, or ever return an error.
func TestStringSegmenterRoundtrip(t *testing.T) {
	t.Parallel()

	const runs = 2000

	seg := phrases.FromString("")

	for i := 0; i < runs; i++ {
		input := string(getRandomBytes())
		seg.SetText(input)

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

	sc := phrases.FromBytes(input)

	var output []byte
	for sc.Next() {
		output = append(output, sc.Bytes()...)
	}

	if !bytes.Equal(output, input) {
		t.Fatalf("input bytes are not the same as segmented bytes")
	}
}

func stringSegToSetTrimmed(seg *phrases.StringIterator) map[string]struct{} {
	founds := make(map[string]struct{})
	for seg.Next() {
		key := strings.TrimSpace(seg.Text())
		founds[key] = exists
	}
	return founds
}

func TestStringPhraseBoundaries(t *testing.T) {
	t.Parallel()

	input := []byte("This should break here. And then here. 世界. I think, perhaps you can understand that — aside 🏆 🐶 here — “a quote”.")
	seg := phrases.FromString(string(input))
	got := stringSegToSetTrimmed(seg)
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

func BenchmarkStringSegmenter(b *testing.B) {
	file, err := testdata.Sample()

	if err != nil {
		b.Error(err)
	}

	s := string(file)

	len := len(file)
	b.SetBytes(int64(len))
	seg := phrases.FromString(s)

	b.ResetTimer()
	c := 0
	start := time.Now()

	for i := 0; i < b.N; i++ {
		seg.SetText(s)

		for seg.Next() {
			c++
		}
	}

	elapsed := time.Since(start)
	n := float64(b.N)

	tokensPerOp := float64(c) / n
	nsPerOp := float64(elapsed.Nanoseconds()) / n

	b.ReportMetric(1e3*tokensPerOp/nsPerOp, "MMtokens/s")
	b.ReportMetric(tokensPerOp, "tokens/op")
	b.ReportMetric(float64(len)/tokensPerOp, "B/token")
}

func BenchmarkStringSegmentAll(b *testing.B) {
	file, err := testdata.Sample()

	if err != nil {
		b.Error(err)
	}

	s := string(file)
	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	for i := 0; i < b.N; i++ {
		phrases := phrases.SegmentAllString(s)

		c := 0
		for range phrases {
			c++
		}

		b.ReportMetric(float64(c), "tokens")
	}
}
