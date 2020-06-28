package words_test

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/clipperhouse/segment"
	"github.com/clipperhouse/uax29/words"
)

func TestScanner(t *testing.T) {
	original := `Hi.    
	node.js, first_last, my.name@domain.com
	123.456, 789, .234, 1,000, a16z, 3G and $200.13.
	wishy-washy and C++ and F# and .net
	Let’s Let's possessive' possessive’
	ש״ח
	א"ב
	ב'
	🇦🇺🇦🇶
	"אא"בב"abc
	Then ウィキペディア and 象形.`
	original += "crlf is \r\n"

	scanner := words.NewScanner(strings.NewReader(original))
	got := map[string]bool{}

	for scanner.Scan() {
		token := scanner.Text()
		got[token] = true
	}
	if err := scanner.Err(); err != nil {
		t.Error(err)
	}

	type test struct {
		value string
		found bool
	}

	expecteds := []test{
		{"Hi", true},
		{".", true},
		{"Hi.", false},

		{"node.js", true},
		{"node", false},
		{"js", false},

		{"first_last", true},
		{"first", false},
		{"_", false},
		{"last", false},

		{"my.name", true},
		{"my.name@", false},
		{"@", true},
		{"domain.com", true},
		{"@domain.com", false},

		{"123.456", true},
		{"123,", false},
		{"456", false},
		{"123.456,", false},

		{"789", true},
		{"789,", false},

		{".234", false},
		{"234", true},

		{"1,000", true},
		{"1,000,", false},

		{"wishy-washy", false},
		{"wishy", true},
		{"-", true},
		{"washy", true},

		{"C++", false},
		{"C", true},
		{"+", true},

		{"F#", false},
		{"F", true},
		{"#", true},

		{".net", false},
		{"net", true},

		{"Let's", true},
		{"Let’s", true},
		{"Let", false},
		{"s", false},

		{"possessive", true},
		{"'", true},
		{"’", true},
		{"possessive'", false},
		{"possessive’", false},

		{"a16z", true},

		{"3G", true},

		{"$", true},
		{"200.13", true},

		{"ש״ח", true},
		{`א"ב`, true},
		{"ב'", true},
		{"אא\"בב", true},
		{"abc", true},

		{"ウィキペディア", true},
		{"ウ", false},

		{"象", true},
		{"形", true},
		{"象形", false},

		{"\r\n", true},
		{"\r", false},

		{"🇦🇺", true},
		{"🇦🇶", true},
	}

	for _, expected := range expecteds {
		if got[expected.value] != expected.found {
			t.Errorf("expected %q to be %t", expected.value, expected.found)
		}
	}
}

func TestUnicodeSegments(t *testing.T) {
	var passed, failed int
	for _, test := range segment.UnicodeWordTests {
		rv := make([][]byte, 0)
		scanner := words.NewScanner(bytes.NewReader(test.Input))
		for scanner.Scan() {
			rv = append(rv, scanner.Bytes())
		}
		if err := scanner.Err(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(rv, test.Output) {
			failed++
			t.Fatalf("expected:\n%#v\ngot:\n%#v\nfor: '%s' comment: %s", test.Output, rv, test.Input, test.Comment)
		} else {
			passed++
		}
	}
	t.Logf("passed %d, failed %d", passed, failed)
}

func TestRoundtrip(t *testing.T) {
	input, err := ioutil.ReadFile("testdata/sample.txt")
	inlen := len(input)

	if err != nil {
		t.Error(err)
	}

	if !utf8.Valid(input) {
		t.Error("input file is not valid utf8")
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
	outlen := len(output)

	if inlen != outlen {
		t.Fatalf("input: %d bytes, output: %d bytes", inlen, outlen)
	}

	if !reflect.DeepEqual(output, input) {
		t.Fatalf("input bytes are not the same as scanned bytes")
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
	sc := words.NewScanner(r)

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
	max := 100

	// rand is deliberately not seeded, to keep tests dterministic

	len := rand.Intn(max-min) + min
	b := make([]byte, len)
	rand.Read(b)

	return b
}

func TestRandomBytes(t *testing.T) {
	runs := 50

	for i := 0; i < runs; i++ {
		input := getRandomBytes()

		sc := words.NewScanner(bytes.NewReader(input))

		var output []byte
		for sc.Scan() {
			output = append(output, sc.Bytes()...)
		}
		if err := sc.Err(); err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(output, input) {
			t.Error("input bytes are not the same as scanned bytes")
			t.Errorf("input:\n%#v", input)
			t.Fatalf("output:\n%#v", output)
		}
	}
}

func BenchmarkScanner(b *testing.B) {
	file, err := ioutil.ReadFile("testdata/sample.txt")

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
	for _, test := range segment.UnicodeWordTests {
		buf.Write(test.Input)
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
