package iterators_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/clipperhouse/uax29/iterators"
	"github.com/clipperhouse/uax29/iterators/filter"
	"github.com/clipperhouse/uax29/iterators/transformer"
	"github.com/clipperhouse/uax29/words"
)

func TestScannerSameAsBufio(t *testing.T) {
	splits := []bufio.SplitFunc{words.SplitFunc, bufio.ScanWords}
	for _, split := range splits {
		for i := 0; i < 100; i++ {
			text := getRandomBytes()

			r1 := bytes.NewReader(text)
			sc1 := iterators.NewScanner(r1, split)
			r2 := bytes.NewReader(text)
			sc2 := bufio.NewScanner(r2)
			sc2.Split(split)

			for sc1.Scan() && sc2.Scan() {
				if !bytes.Equal(sc1.Bytes(), sc2.Bytes()) {
					t.Fatal("Scanner and bufio.Scanner should give identical results")
				}
			}
		}
	}
}

func TestScannerFilterIsApplied(t *testing.T) {
	text := "Hello, 世界, how are you? Nice dog aha! 👍🐶"

	{
		r := strings.NewReader(text)
		sc := iterators.NewScanner(r, bufio.ScanWords)
		sc.Filter(startsWithH)

		count := 0
		for sc.Scan() {
			if !startsWithH(sc.Bytes()) {
				t.Fatal("filter was not applied")
			}
			count++
		}

		if count != 2 {
			t.Fatalf("scanner filter should have found 2 results, got %d", count)
		}
	}

	{
		// variadic
		r := strings.NewReader(text)
		sc := iterators.NewScanner(r, bufio.ScanWords)
		sc.Filter(startsWithH, endsWithW)

		count := 0
		for sc.Scan() {
			if !(startsWithH(sc.Bytes()) && endsWithW(sc.Bytes())) {
				t.Fatal("variadic scanner filter was not applied")
			}
			count++
		}

		if count != 1 {
			t.Fatalf("variadic scanner filter should have found 1 result, got %d", count)
		}
	}
}

func TestScannerTransformIsApplied(t *testing.T) {
	text := "HelloÖ, 世界, how are you at the façade café? Nice dog aha! 👍🐶"
	r := strings.NewReader(text)
	sc := iterators.NewScanner(r, words.SplitFunc)
	sc.Filter(filter.Wordlike)
	sc.Transform(transformer.Diacritics, transformer.Lower)

	for sc.Scan() {
		t.Logf("%s\n", sc.Text())
	}
}
