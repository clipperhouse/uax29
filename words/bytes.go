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

// SegmentAll will iterate through all words and collect them into a [][]byte.
// This is a convenience method -- if you will be allocating such a slice anyway,
// this will save you some code.
//
// The downside is that this allocation is unbounded -- O(n) on the number of
// words. Use Segmenter for more bounded memory usage.
func SegmentAll(data []byte) [][]byte {
	// Optimization: guesstimate that the average word is 3 bytes,
	// allocate a large enough array to avoid resizing
	result := make([][]byte, 0, len(data)/3)
	j := Joiners{}

	_ = iterators.All(data, &result, j.splitFunc) // can elide the error, see tests
	return result
}
