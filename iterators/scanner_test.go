package iterators_test

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"strings"
	"testing"

	"github.com/clipperhouse/uax29/graphemes"
	"github.com/clipperhouse/uax29/iterators"
	"github.com/clipperhouse/uax29/phrases"
	"github.com/clipperhouse/uax29/sentences"
	"github.com/clipperhouse/uax29/words"
)

var splitFuncs = map[string]bufio.SplitFunc{
	"words":     words.SplitFunc,
	"sentences": sentences.SplitFunc,
	"graphemes": graphemes.SplitFunc,
	"phrases":   phrases.SplitFunc,
}

func TestScannerSameAsBufio(t *testing.T) {
	t.Parallel()

	text := make([]byte, 50000)

	for _, split := range splitFuncs {
		for i := 0; i < 100; i++ {
			_, err := rand.Read(text)
			if err != nil {
				t.Fatal(err)
			}

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
	t.Parallel()

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
}
