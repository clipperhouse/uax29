//go:build go1.23
// +build go1.23

package iterators

import (
	"iter"
)

// stringish is a type constraint that allows []byte, string, or named types backed by those
type stringish interface {
	~[]byte | ~string
}

type Token[T stringish] struct {
	value T
}

func (t Token[T]) Value() T {
	return t.value
}

// Iter is an iterator that yields the all of the tokens in the segmenter, for use with range
func (seg *Segmenter) Iter() iter.Seq[Token[[]byte]] {
	return func(yield func(Token[[]byte]) bool) {
		for seg.Next() {
			if !yield(Token[[]byte]{seg.Bytes()}) {
				return
			}
		}
	}
}

// Iter is an iterator that yields the all of the tokens in the scanner, for use with range
func (sc *Scanner) Iter() iter.Seq2[Token[[]byte], error] {
	return func(yield func(Token[[]byte], error) bool) {
		for sc.Scan() {
			if !yield(Token[[]byte]{sc.Bytes()}, sc.Err()) { // err should be nil here but yield anyway
				return
			}
		}
		if sc.Err() != nil {
			yield(Token[[]byte]{sc.Bytes()}, sc.Err()) // bytes should be irrelevant here but yield anyway
		}
	}
}

// Iter is an iterator that yields the all of the tokens in the segmenter, for use with range
func (seg *StringSegmenter) Iter() iter.Seq[Token[string]] {
	return func(yield func(Token[string]) bool) {
		for seg.Next() {
			if !yield(Token[string]{seg.Text()}) {
				return
			}
		}
	}
}
