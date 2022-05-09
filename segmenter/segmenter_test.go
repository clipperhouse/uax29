package segmenter

import (
	"bufio"
	"math/rand"
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
