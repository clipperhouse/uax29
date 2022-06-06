package iterators

import (
	"bufio"
	"errors"

	"github.com/clipperhouse/uax29/iterators/filter"
	"golang.org/x/text/transform"
)

// Segmenter is an iterator for byte slices, which are segmented into tokens (segments).
// To use it, you will define a SplitFunc, SetText with the bytes you wish to tokenize,
// loop over Next until false, call Bytes to retrieve the current token, and check Err
// after the loop.
type Segmenter struct {
	split       bufio.SplitFunc
	filters     []filter.Func
	transformer transform.Transformer
	data        []byte
	token       []byte
	start       int
	err         error
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
	seg.start = 0
	seg.err = nil
}

// Filter applies one or more filters (predicates) to all tokens, returning only those
// where all predicates evaluate true. Calling Filter will overwrite previous filters, so call it
// once (it's variadic, you can add multiple).
func (seg *Segmenter) Filter(filters ...filter.Func) {
	seg.filters = filters
}

// Transform applies one or more transforms to all tokens. Calling Transform will overwrite
// previous transforms, so call it once (it's variadic, you can add multiple).
func (seg *Segmenter) Transform(transformers ...transform.Transformer) {
	seg.transformer = transform.Chain(transformers...)
}

var ErrAdvanceNegative = errors.New("SplitFunc returned a negative advance")
var ErrAdvanceTooFar = errors.New("SplitFunc advanced beyond the end of the data")

// Next advances Segmenter to the next token (segment). It returns false when there
// are no remaining segments, or an error occurred.
func (seg *Segmenter) Next() bool {
next:
	for seg.start < len(seg.data) {
		advance, token, err := seg.split(seg.data[seg.start:], true)
		seg.start += advance
		seg.token = token
		seg.err = err

		if seg.err != nil {
			return false
		}

		// Guardrails
		if advance < 0 {
			seg.err = ErrAdvanceNegative
			return false
		}
		if seg.start > len(seg.data) {
			seg.err = ErrAdvanceTooFar
			return false
		}

		// Interpret as EOF
		if advance == 0 {
			return false
		}

		// Interpret as EOF
		if len(seg.token) == 0 {
			return false
		}

		if seg.transformer != nil {
			seg.token, _, err = transform.Bytes(seg.transformer, seg.token)
			if err != nil {
				seg.err = err
				return false
			}
		}

		for _, f := range seg.filters {
			if !f(seg.token) {
				continue next
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

// Bytes returns the current token.
func (seg *Segmenter) Bytes() []byte {
	return seg.token
}

// Text returns the current token as a newly-allocated string.
func (seg *Segmenter) Text() string {
	return string(seg.token)
}

// Start returns the position (byte index) of the current token in the original text.
func (seg *Segmenter) Start() int {
	return seg.start
}

// All will iterate through all tokens and collect them into a [][]byte. It is a
// convenience method -- if you will be allocating such a slice anyway, this
// will save you some code. The downside is that it allocates, and can do so
// unbounded -- O(n) on the number of tokens. Use Segmenter for more bounded
// memory usage.
//
// The predicates parameter is optional; when predicates is specified, All will
// only return tokens (segments) for which all predicates evaluate to true.
func All(src []byte, dest *[][]byte, split bufio.SplitFunc) error {
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

		*dest = append(*dest, token)
	}

	return nil
}
