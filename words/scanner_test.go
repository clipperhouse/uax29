package words_test

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/kevwang/uax29/words"
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
	for _, test := range unicodeTests {

		var got [][]byte
		sc := words.NewScanner(bytes.NewReader(test.input))

		for sc.Scan() {
			got = append(got, sc.Bytes())
		}

		if err := sc.Err(); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(got, test.expected) {
			failed++
			t.Errorf(`
for input %v
expected  %v
got       %v
spec      %s`, test.input, test.expected, got, test.comment)
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

	if !bytes.Equal(output, input) {
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

	if !bytes.Equal(output, input) {
		t.Fatalf("input bytes are not the same as scanned bytes")
	}
}

var seed = time.Now().UnixNano()
var rnd = rand.New(rand.NewSource(seed))

const max = 10000
const min = 1

func getRandomBytes() []byte {
	len := rnd.Intn(max-min) + min
	b := make([]byte, len)
	rand.Read(b)

	return b
}

func TestRandomBytes(t *testing.T) {
	// Increase this number to do pseudo-fuzzing
	const runs = 2_000
	const workers = 4
	var ran int32

	var wg sync.WaitGroup
	for j := 0; j < workers; j++ {
		wg.Add(1)
		go func() {
			for i := 0; i < runs/workers; i++ {
				input := getRandomBytes()

				// Randomize buffer size, too
				s := rnd.Intn(max-min) + min
				r := bufio.NewReaderSize(bytes.NewReader(input), s)

				sc := words.NewScanner(r)

				var output []byte
				for sc.Scan() {
					output = append(output, sc.Bytes()...)
				}
				if err := sc.Err(); err != nil {
					t.Error(err)
				}

				if !bytes.Equal(output, input) {
					t.Errorf("input bytes are not the same as scanned bytes; rand seed is %d", seed)
				}
				atomic.AddInt32(&ran, 1)
			}
			wg.Done()
		}()
	}

	t.Logf("started: %d runs on %d goroutines, with random seed %d", runs, workers, seed)

	ticker := time.NewTicker(5 * time.Second)
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				t.Logf("%d runs of %d completed (%v%%)", ran, runs, math.Round(float64(ran)/runs*100))
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()

	wg.Wait()

	stop <- struct{}{}

	t.Logf("finished: %d runs of %d completed (%v%%)", ran, runs, math.Round(float64(ran)/runs*100))

	if int(ran) != runs {
		t.Errorf("expected %d runs, got %d", runs, ran)
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
	for _, test := range unicodeTests {
		buf.Write(test.input)
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

func TestScannerMaxTokenLen(t *testing.T) {
	str := ""
	for i := 0; i < 505; i++ {
		if i == 500 {
			str += "."
		}
		str += "a"
	}

	scanner := words.NewScanner(strings.NewReader(str))

	expected := []string{str[:words.TOKEN_BREAK_LEN], ".", str[words.TOKEN_BREAK_LEN + 1:]}
	i := 0

	for scanner.Scan() {
		token := scanner.Text()
		if token != expected[i] {
			t.Errorf("expected %q to be %q", token, expected[i])
		}
		i++
	}
	if i != len(expected) {
		t.Errorf("found %d tokens but expected %d", i, len(expected))
	}
	if err := scanner.Err(); err != nil {
		t.Error(err)
	}
}
