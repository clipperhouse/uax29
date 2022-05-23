package iterators_test

import (
	"bufio"
	"math/rand"
	"reflect"
	"testing"

	"github.com/clipperhouse/uax29/iterators"
)

func getRandomBytes() []byte {
	b := make([]byte, 5000)
	rand.Read(b)

	return b
}

func TestSegmenter(t *testing.T) {
	text := []byte("Hello. How are you?")

	seg := iterators.NewSegmenter(bufio.ScanWords)
	seg.SetText(text)

	for seg.Next() {
		t.Log(seg.Bytes())
	}
}

func TestAll(t *testing.T) {
	text := []byte("Hello. How are you?")

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
