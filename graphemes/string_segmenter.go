package graphemes

import "github.com/clipperhouse/uax29/iterators"

// NewSegmenter retuns a Segmenter, which is an iterator over the source text.
// Iterate while Next() is true, and access the segmented graphemes via Bytes().
func NewStringSegmenter(data string) *iterators.StringSegmenter {
	seg := iterators.NewStringSegmenter(SplitFunc)
	seg.SetText(data)
	return seg
}

// SegmentAll will iterate through all tokens and collect them into a [][]byte.
// This is a convenience method -- if you will be allocating such a slice anyway,
// this will save you some code. The downside is that this allocation is
// unbounded -- O(n) on the number of tokens. Use Segmenter for more bounded
// memory usage.
func SegmentAllString(data string) []string {
	// Optimization: guesstimate that the average grapheme is 1 bytes,
	// allocate a large enough array to avoid resizing
	result := make([]string, 0, len(data))
	seg := NewStringSegmenter(data)
	for seg.Next() {
		result = append(result, seg.Text())
	}

	return result
}
