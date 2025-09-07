package phrases

import (
	"github.com/clipperhouse/uax29/internal/iterators"
)

// BytesIterator is an iterator for phrases. Iterate while Next() is true,
// and access the phrase via Bytes().
type BytesIterator struct {
	*iterators.BytesIterator
}

// FromBytes returns an iterator for the phrases in the input bytes.
// Iterate while Next() is true, and access the phrase via Bytes().
func FromBytes(b []byte) *BytesIterator {
	iter := &BytesIterator{
		iterators.NewBytesIterator(SplitFunc),
	}
	iter.SetText(b)
	return iter
}
