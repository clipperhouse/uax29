package tests

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/clipperhouse/segment"
	"github.com/clipperhouse/uax29/sentences"
)

func TestUnicodeSegments(t *testing.T) {
	var passed, failed int
	for _, test := range segment.UnicodeSentenceTests {
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
			t.Fatalf("expected:\n%#v\ngot:\n%#v\nfor: '%s' comment: %s", test.Output, rv, test.Input, test.Comment)
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
