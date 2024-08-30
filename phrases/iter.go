//go:build go1.23
// +build go1.23

package phrases

import (
	"io"
	"iter"

	"github.com/clipperhouse/uax29/iterators"
)

// Split is an iterator over phrases (tokens), for use with range
func Split(data []byte) iter.Seq[iterators.Token] {
	return NewSegmenter(data).All()
}

// Scan is an iterator over phrases (tokens), for use with range
func Scan(r io.Reader) iter.Seq2[iterators.Token, error] {
	return NewScanner(r).All()
}
