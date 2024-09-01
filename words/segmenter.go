package words

import (
	"github.com/clipperhouse/uax29/iterators"
)

// Segmenter is an iterator for byte slices, which are segmented into tokens (segments).
// To use it, you will define a SplitFunc, SetText with the bytes you wish to tokenize,
// loop over Next until false, call Bytes to retrieve the current token, and check Err
// after the loop.
type Segmenter struct {
	// made a words.Segmenter so we can attach the Joiners method just for words.
	*iterators.Segmenter
}

// NewSegmenter retuns a Segmenter, which is an iterator over the source text.
// Iterate while Next() is true, and access the segmented words via Bytes().
func NewSegmenter(data []byte) *Segmenter {
	seg := &Segmenter{
		iterators.NewSegmenter(SplitFunc),
	}
	seg.SetText(data)
	return seg
}

// SegmentAll will iterate through all tokens and collect them into a [][]byte.
// This is a convenience method -- if you will be allocating such a slice anyway,
// this will save you some code. The downside is that this allocation is
// unbounded -- O(n) on the number of tokens. Use Segmenter for more bounded
// memory usage.
func SegmentAll(data []byte) [][]byte {
	// Optimization: guesstimate that the average word is 3 bytes,
	// allocate a large enough array to avoid resizing
	result := make([][]byte, 0, len(data)/3)
	j := Joiners{}

	_ = iterators.All(data, &result, j.splitFunc) // can elide the error, see tests
	return result
}
