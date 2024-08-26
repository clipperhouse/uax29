//go:build go1.23
// +build go1.23

package phrases

import (
	"io"
	"iter"
)

func Split(data []byte) iter.Seq[[]byte] {
	return NewSegmenter(data).All()
}

func Scan(r io.Reader) iter.Seq2[[]byte, error] {
	return NewScanner(r).All()
}
