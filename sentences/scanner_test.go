package sentences_test

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

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

func BenchmarkScanner(b *testing.B) {
	file, err := ioutil.ReadFile("testdata/wikipedia.txt")

	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	count := 0
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

		count = c
	}
	b.Logf("%d tokens\n", count)
}
