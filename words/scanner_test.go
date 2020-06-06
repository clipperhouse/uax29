package words_test

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

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
	file, err := ioutil.ReadFile("testdata/wikipedia.txt")

	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(file)
	sc := words.NewScanner(r)

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
