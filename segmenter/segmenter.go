package segmenter

import (
	"bufio"
	"unicode"

	"github.com/clipperhouse/uax29/segmenter/filter"
	"github.com/clipperhouse/uax29/segmenter/util"
)

// Segmenter is an iterator for byte arrays. See the New() and Next() funcs.
type Segmenter struct {
	split   bufio.SplitFunc
	filters []filter.Func
	data    []byte
	token   []byte
	pos     int
	err     error
}

// New creates a new segmenter given a SplitFunc. To use the new segmenter,
// call SetText() and then iterate while Next() is true.
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

// Filters applies one or more filters to all tokens, only returning those
// where all filters evaluate true.
func (seg *Segmenter) Filters(f ...filter.Func) {
	seg.filters = f
}

// Next advances the Segmenter to the next segment. It returns false when there
// are no remaining segments, or an error occurred.
func (seg *Segmenter) Next() bool {
outer:
	for seg.pos < len(seg.data) {
		advance, token, err := seg.split(seg.data[seg.pos:], true)
		seg.pos += advance
		seg.token = token
		seg.err = err

		if advance == 0 {
			return false
		}
		if len(seg.token) == 0 {
			return false
		}
		if seg.err != nil {
			return false
		}

		for _, f := range seg.filters {
			if !f(seg.token) {
				continue outer
			}
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

// Contains indicates that the current segment (token) contains one or more runes
// that are in one or more of the ranges.
func (seg *Segmenter) Contains(ranges ...*unicode.RangeTable) bool {
	return util.Contains(seg.token, ranges...)
}

// Entirely indicates that the current segment (token) consists entirely of
// runes that are in one or more of the ranges.
func (seg *Segmenter) Entirely(ranges ...*unicode.RangeTable) bool {
	return util.Entirely(seg.token, ranges...)
}

// Is indicates that the current segment (token) evaluates to true
// for all filters.
func (seg *Segmenter) Is(filters ...filter.Func) bool {
	for _, f := range filters {
		if !f(seg.token) {
			return false
		}
	}

	return true
}

// All will iterate through all tokens and collect them into a [][]byte. It is a
// convenience method -- if you will be allocating such a slice anyway, this
// will save you some code. The downside is that it allocates, and can do so
// unbounded -- O(n) on the number of tokens. Use Segmenter for more bounded
// memory usage.
func All(src []byte, dest *[][]byte, split bufio.SplitFunc, filters ...filter.Func) error {
outer:
	for pos := 0; pos < len(src); {
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

		for _, f := range filters {
			if !f(token) {
				continue outer
			}
		}

		*dest = append(*dest, token)
	}

	return nil
}
