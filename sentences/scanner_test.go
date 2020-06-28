package sentences_test

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/clipperhouse/segment"
	"github.com/clipperhouse/uax29/sentences"
)

func TestSentences(t *testing.T) {
	original := "This is a test. “Is it?”, he wondered."

	scanner := sentences.NewScanner(strings.NewReader(original))
	var got []string
	for scanner.Scan() {
		got = append(got, scanner.Text())
		t.Log(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Error(err)
	}

	expected := []string{
		"This is a test. ",
		"“Is it?”, he wondered.",
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestUnicodeSegments(t *testing.T) {
	var passed, failed int
	for i, test := range segment.UnicodeSentenceTests {
		rv := make([][]byte, 0)
		scanner := sentences.NewScanner(bytes.NewReader(test.Input))
		for scanner.Scan() {
			rv = append(rv, scanner.Bytes())
		}
		if err := scanner.Err(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(rv, test.Output) {
			failed++
			t.Fatalf("test %d, expected:\n%#v\ngot:\n%#v\nfor: '%s'\ncomment: %s", i, test.Output, rv, test.Input, test.Comment)
		} else {
			passed++
		}
	}
	t.Logf("passed %d, failed %d", passed, failed)
}

func TestRoundtrip(t *testing.T) {
	file, err := ioutil.ReadFile("testdata/wikipedia.txt")

	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(file)
	sc := sentences.NewScanner(r)

	var result []byte
	for sc.Scan() {
		result = append(result, sc.Bytes()...)
	}
	if err := sc.Err(); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(result, file) {
		t.Error("input bytes are not the same as scanned bytes")
	}
}

func TestInvalidUTF8(t *testing.T) {
	// This tests that we don't get into an infinite loop or otherwise blow up
	// on invalid UTF-8. Bad UTF-8 is undefined behavior for our purposes;
	// our goal is merely to be non-pathological.

	// The SplitFunc seems to just pass on the bad bytes verbatim,
	// as their own segments, though it's not specified to do so.

	// For background, see testdata/UTF-8-test.txt, or:
	// https://www.cl.cam.ac.uk/~mgk25/ucs/examples/UTF-8-test.txt

	// Btw, don't edit UTF-8-test.txt: your editor might turn it into valid UTF-8!

	input, err := ioutil.ReadFile("testdata/UTF-8-test.txt")
	inlen := len(input)

	if err != nil {
		t.Error(err)
	}

	if utf8.Valid(input) {
		t.Error("input file should not be valid utf8")
	}

	r := bytes.NewReader(input)
	sc := sentences.NewScanner(r)

	var output []byte
	for sc.Scan() {
		output = append(output, sc.Bytes()...)
	}
	if err := sc.Err(); err != nil {
		t.Error(err)
	}
	outlen := len(output)

	if inlen != outlen {
		t.Fatalf("input: %d bytes, output: %d bytes", inlen, outlen)
	}

	if !reflect.DeepEqual(output, input) {
		t.Fatalf("input bytes are not the same as scanned bytes")
	}
}

func getRandomBytes() []byte {
	min := 1
	max := 5000

	// rand is deliberately not seeded, to keep tests deterministic

	len := rand.Intn(max-min) + min
	b := make([]byte, len)
	rand.Read(b)

	return b
}

func TestRandomBytes(t *testing.T) {
	runs := 100

	for i := 0; i < runs; i++ {
		input := getRandomBytes()

		sc := sentences.NewScanner(bytes.NewReader(input))

		var output []byte
		for sc.Scan() {
			output = append(output, sc.Bytes()...)
		}
		if err := sc.Err(); err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(output, input) {
			t.Log("input bytes are not the same as scanned bytes")
			t.Logf("input:\n%#v", input)
			t.Fatalf("output:\n%#v", output)
		}
	}
}

func BenchmarkScanner(b *testing.B) {
	file, err := ioutil.ReadFile("testdata/wikipedia.txt")

	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	r := bytes.NewReader(file)

	for i := 0; i < b.N; i++ {
		r.Reset(file)
		sc := sentences.NewScanner(r)

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
	for _, test := range segment.UnicodeSentenceTests {
		buf.Write(test.Input)
	}
	file := buf.Bytes()

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	r := bytes.NewReader(file)

	for i := 0; i < b.N; i++ {
		r.Reset(file)
		sc := sentences.NewScanner(r)

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
