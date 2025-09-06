package words

import "github.com/clipperhouse/uax29/internal/iterators"

// StringIterator is an iterator for words. Iterate while Next() is
// true, and access the word via Text().
type StringIterator struct {
	*iterators.StringIterator
}

// FromString returns an iterator for the words in the input string.
// Iterate while Next() is true, and access the word via Text().
func FromString(s string) *StringIterator {
	iter := &StringIterator{
		iterators.NewStringIterator(SplitFunc),
	}
	iter.SetText(s)
	return iter
}

// SegmentAllString will iterate through all words and collect them into a
// []string. This is a convenience method -- if you will be allocating such a
// slice anyway, this will save you some code.
//
// The downside is that this allocation is unbounded -- O(n) on the number of
// words. Use StringSegmenter for more bounded memory usage.
func SegmentAllString(data string) []string {
	// Optimization: guesstimate that the average word is 3 bytes,
	// allocate a large enough array to avoid resizing
	result := make([]string, 0, len(data)/3)
	seg := FromString(data)
	for seg.Next() {
		result = append(result, seg.Text())
	}

	return result
}
