package iterators_test

import (
	"bufio"
	"bytes"
	"math/rand"
	"reflect"
	"testing"

	"github.com/clipperhouse/uax29/iterators"
	"github.com/clipperhouse/uax29/words"
)

func getRandomBytes() []byte {
	b := make([]byte, 5000)
	rand.Read(b)

	return b
}

func TestSegmenterSameAsScanner(t *testing.T) {
	for i := 0; i < 100; i++ {
		text := getRandomBytes()
		split := words.SplitFunc

		seg := iterators.NewSegmenter(split)
		seg.SetText(text)

		r := bytes.NewReader(text)
		sc := iterators.NewScanner(r, split)

		for seg.Next() && sc.Scan() {
			if !bytes.Equal(seg.Bytes(), sc.Bytes()) {
				t.Fatal("Scanner and Segmenter should give identical results")
			}
		}
		if seg.Err() != nil {
			t.Fatal(seg.Err())
		}
		if sc.Err() != nil {
			t.Fatal(sc.Err())
		}
	}
}

func TestSegmenterSameAsAll(t *testing.T) {
	for i := 0; i < 100; i++ {
		text := getRandomBytes()

		split := bufio.ScanWords

		seg := iterators.NewSegmenter(split)
		seg.SetText(text)

		var segResult [][]byte
		for seg.Next() {
			segResult = append(segResult, seg.Bytes())
		}
		if seg.Err() != nil {
			t.Fatal(seg.Err())
		}

		var allResult [][]byte
		err := iterators.All(text, &allResult, split)
		if seg.Err() != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(segResult, allResult) {
			t.Logf("Segmenter result: %q", segResult)
			t.Logf("All result: %q", allResult)
			t.Fatal("All and Segmenter should give identical results")
		}
	}
}
