package iterators_test

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"

	"github.com/clipperhouse/uax29/iterators"
)

func TestScannerSameAsBufio(t *testing.T) {
	for i := 0; i < 100; i++ {
		text := getRandomBytes()

		r1 := bytes.NewReader(text)
		sc1 := iterators.NewScanner(r1, bufio.ScanWords)
		var all1 [][]byte
		for sc1.Scan() {
			all1 = append(all1, sc1.Bytes())
		}

		r2 := bytes.NewReader(text)
		sc2 := bufio.NewScanner(r2)
		sc2.Split(bufio.ScanWords)
		var all2 [][]byte
		for sc2.Scan() {
			all2 = append(all2, sc2.Bytes())
		}

		if !reflect.DeepEqual(all1, all2) {
			t.Error("iterators.Scanner and bufio.Scanner give different results")
		}
	}
}
