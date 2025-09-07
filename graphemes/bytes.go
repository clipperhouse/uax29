package graphemes

import (
	"github.com/clipperhouse/uax29/internal/iterators"
)

// BytesIterator is an iterator for grapheme clusters. Iterate while Next() is
// true, and access the grapheme via Bytes().
type BytesIterator struct {
	*iterators.BytesIterator
}

// FromBytes returns an iterator for the grapheme clusters in the input bytes.
// Iterate while Next() is true, and access the grapheme cluster via Bytes().
func FromBytes(data []byte) *BytesIterator {
	iter := &BytesIterator{
		iterators.NewBytesIterator(SplitFunc),
	}
	iter.SetText(data)
	return iter
}
