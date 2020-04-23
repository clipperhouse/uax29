package tests

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

	// First, test roundtrip
	scanner := sentences.NewScanner(strings.NewReader(original))
	roundtrip := ""
	for scanner.Scan() {
		roundtrip += scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		t.Error(err)
	}
	if roundtrip != original {
		t.Error("expected roundtrip to equal original")
	}

	var got []string
	scanner = sentences.NewScanner(strings.NewReader(original))
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
			rv = append(rv, []byte(scanner.Text()))
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

func BenchmarkScanner(b *testing.B) {
	file, err := ioutil.ReadFile("wikipedia.txt")

	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	count := 0
	for i := 0; i < b.N; i++ {
		var bb bytes.Buffer

		r := bytes.NewReader(file)
		sc := sentences.NewScanner(r)

		c := 0
		for sc.Scan() {
			c++
			bb.WriteString(sc.Text())
		}
		if err := sc.Err(); err != nil {
			b.Error(err)
		}

		count = c
		bb.Reset()
	}
	b.Logf("%d tokens\n", count)
}