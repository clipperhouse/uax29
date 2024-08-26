//go:build go1.23
// +build go1.23

package iterators

import (
	"iter"
)

// All returns an iterator that yields the all of the tokens in the segmenter
func (seg *Segmenter) All() iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		for seg.Next() {
			if !yield(seg.Bytes()) {
				return
			}
		}
	}
}

// All returns an iterator that yields the all of the tokens in the scanner
func (sc *Scanner) All() iter.Seq2[[]byte, error] {
	return func(yield func([]byte, error) bool) {
		for sc.Scan() {
			if !yield(sc.Bytes(), nil) {
				return
			}
		}
		if err := sc.Err(); err != nil {
			yield(nil, err)
		}
	}
}
