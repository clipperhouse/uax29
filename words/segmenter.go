package words

import (
	"bufio"

	"github.com/clipperhouse/uax29/iterators"
)

// NewSegmenter retuns a Segmenter, which is an iterator over the source text.
// Iterate while Next() is true, and access the segmented words via Bytes().
func NewSegmenter(data []byte) *iterators.Segmenter {
	seg := iterators.NewSegmenter(SplitFunc)
	seg.SetText(data)
	return seg
}

// NewSegmenterWeb returns a Segmenter, which is an iterator over the source text.
// It joins tokens on 'web' characters such as '@' (for email addresses and handles),
// and '#' (for hashtags), and many characters for URLs (such as '/', '?', etc).
// The basic segmenter would treat these as separate tokens; this web segmenter will
// join them into a single token.
//
// It is fairly naive, in that it makes no attempt to validate (for example) email
// addresses or URLs. It simply treats the above-mentioned characters as alphanumeric.
//
// Iterate while Next() is true, and access the segmented words via Bytes().
func NewSegmenterWeb(data []byte) *iterators.Segmenter {
	opts := options{Web: true}
	var split bufio.SplitFunc = func(data []byte, atEOF bool) (int, []byte, error) {
		return splitFuncOpts(data, atEOF, opts)
	}

	seg := iterators.NewSegmenter(split)
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

	_ = iterators.All(data, &result, SplitFunc) // can elide the error, see tests
	return result
}
