package segmenter

import (
	"bufio"
	"math/rand"
	"reflect"
	"testing"
)

func getRandomBytes() []byte {
	b := make([]byte, 5000)
	rand.Read(b)

	return b
}

func TestSegmenter(t *testing.T) {
	text := []byte("Hello. How are you?")

	seg := New(bufio.ScanWords)
	seg.SetText(text)

	for seg.Next() {
		t.Log(seg.Bytes())
	}
}

func TestAll(t *testing.T) {
	text := []byte("Hello. How are you?")

	split := bufio.ScanWords

	seg := New(split)
	seg.SetText(text)

	var result1 [][]byte
	for seg.Next() {
		result1 = append(result1, seg.Bytes())
	}

	if seg.Err() != nil {
		t.Fatal(seg.Err())
	}

	result2 := All(text, split)

	if !reflect.DeepEqual(result1, result2) {
		t.Fatal("All and Segmenter should be identical")
	}

	t.Log(result1, result2)
	t.Log(result2)
}
