package iterators_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/iterators"
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

	containsH := func(token []byte) bool {
		pos := 0
		for pos < len(token) {
			r, w := utf8.DecodeRune(token[pos:])
			if unicode.ToLower(r) == 'h' {
				return true
			}
			pos += w
		}

		return false
	}

	r := strings.NewReader(text)
	sc := iterators.NewScanner(r, bufio.ScanWords)
	sc.Filter(containsH)

	count := 0
	for sc.Scan() {
		if !containsH(sc.Bytes()) {
			t.Fatal("filter was not applied")
		}
		count++
	}

	if count != 3 {
		t.Fatalf("scanner filter should have found 3 results, got %d", count)
	}
}
