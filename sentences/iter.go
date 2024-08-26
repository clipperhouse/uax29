//go:build go1.23
// +build go1.23

package sentences

import (
	"io"
	"iter"
)

// Split is an iterator over sentences (tokens), for use with range
func Split(data []byte) iter.Seq[[]byte] {
	return NewSegmenter(data).All()
}

// Scan is an iterator over sentences (tokens), for use with range
func Scan(r io.Reader) iter.Seq2[[]byte, error] {
	return NewScanner(r).All()
}
