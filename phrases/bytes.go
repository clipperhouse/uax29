package phrases

import (
	"github.com/clipperhouse/uax29/iterators"
)

// BytesIterator is an iterator for phrases. Iterate while Next() is true,
// and access the phrase via Bytes().
type BytesIterator = iterators.Segmenter

// FromBytes returns an iterator for the phrases in the input bytes.
// Iterate while Next() is true, and access the phrase via Bytes().
func FromBytes(data []byte) *BytesIterator {
	seg := iterators.NewSegmenter(SplitFunc)
	seg.SetText(data)
	return seg
}