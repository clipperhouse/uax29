package words

import "github.com/clipperhouse/uax29/internal/iterators"

type StringSegmenter struct {
	// made a words.Segmenter so we can attach the Joiners method just for words.
	*iterators.StringSegmenter
}

// NewStringSegmenter returns a StringSegmenter, which is an iterator over the
// source text. Iterate while Next() is true, and access the word via Text().
func NewStringSegmenter(data string) *StringSegmenter {
	seg := &StringSegmenter{
		iterators.NewStringSegmenter(SplitFunc),
	}
	seg.SetText(data)
	return seg
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
	seg := NewStringSegmenter(data)
	for seg.Next() {
		result = append(result, seg.Text())
	}

	return result
}
