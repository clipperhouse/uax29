//go:build go1.23
// +build go1.23

package iterators

import (
	"iter"
)

// All is an iterator that yields the all of the tokens in the segmenter, for use with range
func (seg *Segmenter) All() iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		for seg.Next() {
			yield(seg.Bytes())
		}
	}
}

// All is an iterator that yields the all of the tokens in the scanner, for use with range
func (sc *Scanner) All() iter.Seq2[[]byte, error] {
	return func(yield func([]byte, error) bool) {
		for sc.Scan() {
			yield(sc.Bytes(), sc.Err()) // err should be nil here but yield anyway
		}
		if sc.Err() != nil {
			yield(sc.Bytes(), sc.Err()) // bytes should be irrelevant here but yield anyway
		}
	}
}