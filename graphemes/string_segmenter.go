package graphemes

import "github.com/clipperhouse/uax29/iterators"

// NewStringSegmenter returns a StringSegmenter, which is an iterator over the
// source text. Iterate while Next() is true, and access the grapheme via
// Text().
func NewStringSegmenter(data string) *iterators.StringSegmenter {
	seg := iterators.NewStringSegmenter(SplitFunc)
	seg.SetText(data)
	return seg
}

// SegmentAllString will iterate through all graphemes and collect them into a
// []string. This is a convenience method -- if you will be allocating such a
// slice anyway, this will save you some code.
//
// The downside is that this allocation is unbounded -- O(n) on the number of
// graphemes. Use StringSegmenter for more bounded memory usage.
func SegmentAllString(data string) []string {
	// Optimization: guesstimate that the average grapheme is 1 byte,
	// allocate a large enough array to avoid resizing
	result := make([]string, 0, len(data))
	seg := NewStringSegmenter(data)
	for seg.Next() {
		result = append(result, seg.Text())
	}

	return result
}
