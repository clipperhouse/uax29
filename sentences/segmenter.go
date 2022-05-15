package sentences

import "github.com/clipperhouse/uax29/segmenter"

type s = segmenter.Segmenter

// Segmenter is an iterator over the segmented sentences. Use Next().
type Segmenter struct {
	*s
}

// NewSegmenter retuns a Segmenter, which is an iterator over the source text.
// Call Next() until it's false.
func NewSegmenter(data []byte) *Segmenter {
	seg := &Segmenter{
		s: segmenter.New(SplitFunc),
	}
	seg.SetText(data)
	return seg
}

// SegmentAll will iterate through all tokens and collect them into a [][]byte.
// It is a convenience method -- if you will be allocating such a slice anyway,
// this will save you some code. The downside is that this allocation is
// unbounded -- O(n) on the number of tokens. Use Segmenter for more bounded
// memory usage.
func SegmentAll(data []byte) [][]byte {
	// Optimization: guesstimate that the average sentence is 100 bytes,
	// allocate a large enough array to avoid resizing
	result := make([][]byte, 0, len(data)/100)

	_ = segmenter.All(data, &result, SplitFunc) // can elide the error, see tests
	return result
}
