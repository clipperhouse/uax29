package graphemes

import "github.com/clipperhouse/uax29/internal/iterators"

// StringIterator is an iterator for grapheme clusters. Iterate while Next() is
// true, and access the grapheme via Text().
type StringIterator struct {
	*iterators.StringIterator
}

// FromString returns an iterator for the grapheme clusters in the input
// string. Iterate while Next() is true, and access the grapheme via Text().
func FromString(s string) *StringIterator {
	iter := &StringIterator{
		iterators.NewStringIterator(SplitFunc),
	}
	iter.SetText(s)
	return iter
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
	seg := FromString(data)
	for seg.Next() {
		result = append(result, seg.Text())
	}

	return result
}
