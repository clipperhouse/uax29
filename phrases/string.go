package phrases

import "github.com/clipperhouse/uax29/internal/iterators"

// StringIterator is an iterator for phrases. Iterate while Next() is
// true, and access the phrase via Text().
type StringIterator struct {
	*iterators.StringIterator
}

// FromString returns an iterator for the phrases in the input string.
// Iterate while Next() is true, and access the phrase via Text().
func FromString(s string) *StringIterator {
	iter := &StringIterator{
		iterators.NewStringIterator(SplitFunc),
	}
	iter.SetText(s)
	return iter
}

// SegmentAllString will iterate through all phrases and collect them into a
// []string. This is a convenience method -- if you will be allocating such a
// slice anyway, this will save you some code.
//
// The downside is that this allocation is unbounded -- O(n) on the number of
// phrases. Use StringSegmenter for more bounded memory usage.
func SegmentAllString(data string) []string {
	// Optimization: guesstimate that the average phrase is 20 bytes,
	// allocate a large enough array to avoid resizing
	result := make([]string, 0, len(data)/20)
	seg := FromString(data)
	for seg.Next() {
		result = append(result, seg.Text())
	}

	return result
}
