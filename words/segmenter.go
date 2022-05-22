package words

import (
	"github.com/clipperhouse/uax29/segmenter"
	"github.com/clipperhouse/uax29/segmenter/filter"
)

// NewSegmenter retuns a Segmenter, which is an iterator over the source text.
// Iterate while Next() is true, and access the segmented words via Bytes().
func NewSegmenter(data []byte) *segmenter.Segmenter {
	seg := segmenter.New(SplitFunc)
	seg.SetText(data)
	return seg
}

// SegmentAll will iterate through all tokens and collect them into a [][]byte.
// This is a convenience method -- if you will be allocating such a slice anyway,
// this will save you some code. The downside is that this allocation is
// unbounded -- O(n) on the number of tokens. Use Segmenter for more bounded
// memory usage.
func SegmentAll(data []byte, filter ...filter.Func) [][]byte {
	// Optimization: guesstimate that the average word is 3 bytes,
	// allocate a large enough array to avoid resizing
	result := make([][]byte, 0, len(data)/3)

	_ = segmenter.All(data, &result, SplitFunc, filter...) // can elide the error, see tests
	return result
}
