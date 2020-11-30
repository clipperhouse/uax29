package graphemes_test

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"math"
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/graphemes"
)

func TestUnicodeSegments(t *testing.T) {
	var passed, failed int
	for _, test := range unicodeTests {

		var got [][]byte
		sc := graphemes.NewScanner(bytes.NewReader(test.input))

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
	file, err := ioutil.ReadFile("testdata/wikipedia.txt")

	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(file)
	sc := graphemes.NewScanner(r)

	var result []byte
	for sc.Scan() {
		result = append(result, sc.Bytes()...)
	}
	if err := sc.Err(); err != nil {
		t.Error(err)
	}

	if !bytes.Equal(result, file) {
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
	sc := graphemes.NewScanner(r)

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

				sc := graphemes.NewScanner(r)

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
	file, err := ioutil.ReadFile("testdata/wikipedia.txt")

	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	r := bytes.NewReader(file)

	for i := 0; i < b.N; i++ {
		r.Reset(file)
		sc := graphemes.NewScanner(r)

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
		sc := graphemes.NewScanner(r)

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
