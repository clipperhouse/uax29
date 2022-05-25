package iterators_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/clipperhouse/uax29/iterators"
)

func TestScannerSameAsBufio(t *testing.T) {
	for i := 0; i < 100; i++ {
		text := getRandomBytes()

		r1 := bytes.NewReader(text)
		sc1 := iterators.NewScanner(r1, bufio.ScanWords)
		r2 := bytes.NewReader(text)
		sc2 := bufio.NewScanner(r2)
		sc2.Split(bufio.ScanWords)

		for sc1.Scan() && sc2.Scan() {
			if !bytes.Equal(sc1.Bytes(), sc2.Bytes()) {
				t.Fatal("Scanner and bufio.Scanner should give identical results")
			}
		}
	}
}
