package words

import (
	"github.com/clipperhouse/uax29/internal/iterators"
)

// BytesIterator is an iterator for words. Iterate while Next() is true,
// and access the word via Bytes().
type BytesIterator struct {
	*iterators.BytesIterator
}

// FromBytes returns an iterator for the words in the input bytes.
// Iterate while Next() is true, and access the word via Bytes().
func FromBytes(data []byte) *BytesIterator {
	iter := &BytesIterator{
		iterators.NewBytesIterator(SplitFunc),
	}
	iter.SetText(data)
	return iter
}
