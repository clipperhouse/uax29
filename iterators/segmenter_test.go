package iterators_test

import (
	"bytes"
	"math/rand"
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
		split := words.SplitFunc

		var all [][]byte
		err := iterators.All(text, &all, split)
		if err != nil {
			t.Fatal(err)
		}

		seg := iterators.NewSegmenter(split)
		seg.SetText(text)

		for i := 0; seg.Next(); i++ {
			if !bytes.Equal(seg.Bytes(), all[i]) {
				t.Fatal("All and Segmenter should give identical results")
			}
		}
		if seg.Err() != nil {
			t.Fatal(seg.Err())
		}
	}
}
