package iterators

import (
	"bufio"

	"github.com/clipperhouse/uax29/iterators/filter"
)

// Segmenter is an iterator for byte slices, which are segmented into tokens (segments).
// To use it, you will define a SplitFunc, SetText with the bytes you wish to tokenize,
// loop over Next until false, call Bytes to retrieve the current token, and check Err
// after the loop.
type Segmenter struct {
	split      bufio.SplitFunc
	predicates []filter.Predicate
	data       []byte
	token      []byte
	pos        int
	err        error
}

// NewSegmenter creates a new segmenter given a SplitFunc. To use the new segmenter,
// call SetText() and then iterate while Next() is true.
func NewSegmenter(split bufio.SplitFunc) *Segmenter {
	return &Segmenter{
		split: split,
	}
}

// SetText sets the text for the segmenter to operate on, and resets
// all state.
func (seg *Segmenter) SetText(data []byte) {
	seg.data = data
	seg.token = nil
	seg.pos = 0
	seg.err = nil
}

// Filter applies one or more filters (predicates) to all tokens (segments), only returning those
// where all predicates evaluate true.
func (seg *Segmenter) Filter(predicates ...filter.Predicate) {
	seg.predicates = predicates
}

// Next advances Segmenter to the next token (segment). It returns false when there
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

		for _, f := range seg.predicates {
			if !f(seg.token) {
				continue outer
			}
		}

		return true
	}

	return false
}

// Err indicates an error occured when calling Next; Next will return false
// when an error occurs.
func (seg *Segmenter) Err() error {
	return seg.err
}

// Bytes returns the current token (segment).
func (seg *Segmenter) Bytes() []byte {
	return seg.token
}

// Text returns the current token (segment) as a newly-allocated string.
func (seg *Segmenter) Text() string {
	return string(seg.token)
}

// All will iterate through all tokens and collect them into a [][]byte. It is a
// convenience method -- if you will be allocating such a slice anyway, this
// will save you some code. The downside is that it allocates, and can do so
// unbounded -- O(n) on the number of tokens. Use Segmenter for more bounded
// memory usage.
//
// The predicates parameter is optional; when predicates is specified, All will
// only return tokens (segments) for which all predicates evaluate to true.
func All(src []byte, dest *[][]byte, split bufio.SplitFunc, predicates ...filter.Predicate) error {
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

		for _, f := range predicates {
			if !f(token) {
				continue outer
			}
		}

		*dest = append(*dest, token)
	}

	return nil
}
