package sentences

import (
	"github.com/clipperhouse/uax29/v2/internal/iterators"
)

// BytesIterator is an iterator for sentences. Iterate while Next() is true,
// and access the sentence via Bytes().
type BytesIterator struct {
	*iterators.BytesIterator
}

// FromBytes returns an iterator for the sentences in the input bytes.
// Iterate while Next() is true, and access the sentence via Bytes().
func FromBytes(b []byte) *BytesIterator {
	iter := &BytesIterator{
		iterators.NewBytesIterator(SplitFunc),
	}
	iter.SetText(b)
	return iter
}
