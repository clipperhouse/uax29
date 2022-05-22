package segmenter

import (
	"bufio"

	"github.com/clipperhouse/uax29/segmenter/filter"
)

// Segmenter is an iterator for byte arrays. See the New() and Next() funcs.
type Segmenter struct {
	split  bufio.SplitFunc
	filter filter.Func
	data   []byte
	token  []byte
	pos    int
	err    error
}

// New creates a new segmenter given a SplitFunc. To use the new segmenter,
// call SetText() and then iterate while Next() is true
func New(split bufio.SplitFunc) *Segmenter {
	return &Segmenter{
		split: split,
	}
}

// SetText sets the text for the segmenter to operate on, and resets
// all state
func (seg *Segmenter) SetText(data []byte) {
	seg.data = data
	seg.token = nil
	seg.pos = 0
	seg.err = nil
}

// Filter applies a filter to all tokens, only returning the token if filter(token) evaluates true
func (seg *Segmenter) Filter(f filter.Func) {
	seg.filter = f
}

// Next advances the Segmenter to the next segment. It returns false when there
// are no remaining segments, or an error occurred.
func (seg *Segmenter) Next() bool {
	for seg.pos < len(seg.data) {
		advance, token, err := seg.split(seg.data[seg.pos:], true)

		seg.pos += advance
		seg.token = token
		seg.err = err

		if advance == 0 {
			return false
		}

		if len(token) == 0 {
			return false
		}

		if seg.err != nil {
			return false
		}

		if seg.filter != nil && !seg.filter(seg.token) {
			continue
		}

		return true
	}

	return false
}

// Err indicates an error occured when calling Next() or Previous(). Next and
// Previous will return false when an error occurs.
func (seg *Segmenter) Err() error {
	return seg.err
}

// Bytes returns the current segment
func (seg *Segmenter) Bytes() []byte {
	return seg.token
}

// All will iterate through all tokens and collect them into a [][]byte. It is a
// convenience method -- if you will be allocating such a slice anyway, this
// will save you some code. The downside is that it allocates, and can do so
// unbounded -- O(n) on the number of tokens. Use Segmenter for more bounded
// memory usage.
func All(src []byte, dest *[][]byte, split bufio.SplitFunc) error {
	pos := 0

	for pos < len(src) {
		advance, token, err := split(src[pos:], true)
		if err != nil {
			return err
		}

		if advance == 0 {
			break
		}
		pos += advance

		if len(token) == 0 {
			break
		}
		*dest = append(*dest, token)
	}

	return nil
}
