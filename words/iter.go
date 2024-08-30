//go:build go1.23
// +build go1.23

package words

import (
	"io"
	"iter"

	"github.com/clipperhouse/uax29/iterators"
)

// Split is an iterator over words (tokens), for use with range
func Split(data []byte) iter.Seq[iterators.Token] {
	return NewSegmenter(data).All()
}

// Scan is an iterator over words (tokens), for use with range
func Scan(r io.Reader) iter.Seq2[iterators.Token, error] {
	return NewScanner(r).All()
}
