package phrases_test

import (
	"bytes"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/v2/phrases"
	"github.com/clipperhouse/uax29/v2/testdata"
)

func TestStringRoundtrip(t *testing.T) {
	t.Parallel()

	const runs = 100

	tokens := phrases.FromString("")

	for i := 0; i < runs; i++ {
		input := string(getRandomBytes())
		tokens.SetText(input)

		var output string
		for tokens.Next() {
			output += tokens.Value()
		}

		if output != input {
			t.Fatal("input bytes are not the same as output bytes")
		}
	}
}

func TestStringInvalidUTF8(t *testing.T) {
	t.Parallel()

	// For background, see internal/testdata/UTF-8-test.txt, or:
	// https://www.cl.cam.ac.uk/~mgk25/ucs/examples/UTF-8-test.txt

	// Btw, don't edit UTF-8-test.txt: your editor might turn it into valid UTF-8!

	input, err := testdata.InvalidUTF8()

	if err != nil {
		t.Error(err)
	}

	if utf8.Valid(input) {
		t.Error("input file should not be valid utf8")
	}

	tokens := phrases.FromBytes(input)

	var output []byte
	for tokens.Next() {
		output = append(output, tokens.Value()...)
	}

	if !bytes.Equal(output, input) {
		t.Fatalf("input bytes are not the same as output bytes")
	}
}

func stringIterToSetTrimmed(tokens phrases.Iterator[string]) map[string]struct{} {
	founds := make(map[string]struct{})
	for tokens.Next() {
		key := strings.TrimSpace(tokens.Value())
		founds[key] = exists
	}
	return founds
}

func TestStringPhraseBoundaries(t *testing.T) {
	t.Parallel()

	input := []byte("This should break here. And then here. 世界. I think, perhaps you can understand that — aside 🏆 🐶 here — “a quote”.")
	tokens := phrases.FromString(string(input))
	got := stringIterToSetTrimmed(tokens)
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

func BenchmarkString(b *testing.B) {
	file, err := testdata.Sample()

	if err != nil {
		b.Error(err)
	}

	s := string(file)

	len := len(file)
	b.SetBytes(int64(len))

	b.ResetTimer()
	c := 0
	start := time.Now()

	for i := 0; i < b.N; i++ {
		tokens := phrases.FromString(s)

		for tokens.Next() {
			_ = tokens.Value()
			c++
		}
	}
	b.ReportMetric(float64(c), "tokens")

	elapsed := time.Since(start)
	n := float64(b.N)

	tokensPerOp := float64(c) / n
	nsPerOp := float64(elapsed.Nanoseconds()) / n

	b.ReportMetric(1e3*tokensPerOp/nsPerOp, "MMtokens/s")
	b.ReportMetric(tokensPerOp, "tokens/op")
	b.ReportMetric(float64(len)/tokensPerOp, "B/token")
}
