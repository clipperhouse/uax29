package iterators_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/clipperhouse/stemmer"
	"github.com/clipperhouse/uax29/iterators"
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
}

func TestScannerTransformIsApplied(t *testing.T) {
	text := "Hello, 世界, I am enjoying cups of Açaí in Örebro."
	r := strings.NewReader(text)
	sc := iterators.NewScanner(r, bufio.ScanWords)
	sc.Transform(transformer.Lower, transformer.Diacritics, stemmer.English)

	var tokens [][]byte
	for sc.Scan() {
		tokens = append(tokens, sc.Bytes())
	}

	if sc.Err() != nil {
		t.Fatal(sc.Err())
	}

	{
		got := tokens[4]
		expected := []byte("enjoy")
		if !bytes.Equal(expected, got) {
			t.Fatalf("stemmer was not applied, expected %q, got %q", expected, got)
		}
	}

	{
		got := tokens[7]
		expected := []byte("acai")
		if !bytes.Equal(expected, got) {
			t.Fatalf("transforms of lower case or diacritics were not applied, expected %q, got %q", expected, got)
		}
	}

}
