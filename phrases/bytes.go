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
func FromBytes(data []byte) *BytesIterator {
	iter := &BytesIterator{
		iterators.NewBytesIterator(SplitFunc),
	}
	iter.SetText(data)
	return iter
}

// SegmentAll will iterate through all phrases and collect them into a [][]byte.
// This is a convenience method -- if you will be allocating such a slice anyway,
// this will save you some code.
//
// The downside is that this allocation is unbounded -- O(n) on the number of
// phrases. Use Segmenter for more bounded memory usage.
func SegmentAll(data []byte) [][]byte {
	// Optimization: guesstimate that the average phrase is 3 bytes,
	// allocate a large enough array to avoid resizing
	result := make([][]byte, 0, len(data)/3)

	_ = iterators.All(data, &result, SplitFunc) // can elide the error, see tests
	return result
}
