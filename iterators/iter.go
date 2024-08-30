//go:build go1.23
// +build go1.23

package iterators

import (
	"iter"
)

type Token struct {
	value []byte
}

func (t Token) Value() []byte {
	return t.value
}

// All is an iterator that yields the all of the tokens in the segmenter, for use with range
func (seg *Segmenter) All() iter.Seq[Token] {
	return func(yield func(Token) bool) {
		for seg.Next() {
			t := Token{seg.Bytes()}
			yield(t)
		}
	}
}

// All is an iterator that yields the all of the tokens in the scanner, for use with range
func (sc *Scanner) All() iter.Seq2[Token, error] {
	return func(yield func(Token, error) bool) {
		for sc.Scan() {
			yield(Token{sc.Bytes()}, sc.Err()) // err should be nil here but yield anyway
		}
		if sc.Err() != nil {
			yield(Token{sc.Bytes()}, sc.Err()) // bytes should be irrelevant here but yield anyway
		}
	}
}
