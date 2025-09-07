package iterators_test

import (
	"bufio"
	"reflect"
	"testing"

	"github.com/clipperhouse/uax29/graphemes"
	"github.com/clipperhouse/uax29/internal/iterators"
	"github.com/clipperhouse/uax29/phrases"
	"github.com/clipperhouse/uax29/sentences"
	"github.com/clipperhouse/uax29/words"
)

var splitFuncs = map[string]bufio.SplitFunc{
	"words":     words.SplitFunc,
	"sentences": sentences.SplitFunc,
	"graphemes": graphemes.SplitFunc,
	"phrases":   phrases.SplitFunc,
}

func TestBytesStart(t *testing.T) {
	t.Parallel()

	text := []byte("Hello world")

	{
		tokens := words.FromBytes(text)
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

	{
		tokens := iterators.NewBytesIterator(bufio.ScanWords)
		tokens.SetText(text)
		expected := []int{0, 6}
		var got []int
		for tokens.Next() {
			got = append(got, tokens.Start())
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("start failed for bufio.ScanWords, expected %v, got %v", expected, got)
		}
	}
}

func TestBytesEnd(t *testing.T) {
	t.Parallel()

	text := []byte("Hello world")
	tokens := words.FromBytes(text)

	expected := []int{5, 6, len(text)}
	var got []int
	for tokens.Next() {
		got = append(got, tokens.End())
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("end failed for words.SplitFunc, expected %v, got %v", expected, got)
	}
}
