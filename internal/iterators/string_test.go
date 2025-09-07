package iterators_test

import (
	"bytes"
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/clipperhouse/uax29/internal/iterators"
	"github.com/clipperhouse/uax29/words"
)

func TestStringSameAsBytes(t *testing.T) {
	t.Parallel()

	text := make([]byte, 50000)

	for _, split := range splitFuncs {
		for i := 0; i < 100; i++ {
			_, err := rand.Read(text)
			if err != nil {
				t.Fatal(err)
			}

			b := iterators.NewBytesIterator(split)
			b.SetText(text)

			s := iterators.NewStringIterator(split)
			s.SetText(string(text))

			for b.Next() && s.Next() {
				bbytes := b.Bytes()
				sbytes := []byte(s.Text())
				if !bytes.Equal(bbytes, sbytes) {
					t.Fatalf(`
					StringIterator and BytesIterator should give identical results
					BytesIterator:  %q
					StringIterator: %q
					`, bbytes, sbytes)
				}
			}
		}
	}
}

func TestStringStart(t *testing.T) {
	t.Parallel()

	text := "Hello world"

	tokens := iterators.NewStringIterator(words.SplitFunc)
	tokens.SetText(text)
	expected := []int{0, 5, 6}
	var got []int
	for tokens.Next() {
		got = append(got, tokens.Start())
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("start failed for words.SplitFunc, expected %v, got %v", expected, got)
	}

}

func TestStringEnd(t *testing.T) {
	t.Parallel()

	text := "Hello world"

	tokens := iterators.NewStringIterator(words.SplitFunc)
	tokens.SetText(text)

	expected := []int{5, 6, len(text)}
	var got []int
	for tokens.Next() {
		got = append(got, tokens.End())
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("end failed for words.SplitFunc, expected %v, got %v", expected, got)
	}

}
