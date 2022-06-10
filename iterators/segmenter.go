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
	filter      filter.Func
	transformer transform.Transformer
	data        []byte
	token       []byte
	pos         int
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
	seg.pos = 0
	seg.err = nil
}

// Filter applies a filter (predicate) to all tokens, returning only those
// where all filters evaluate true. Calling Filter will overwrite the previous
// filter.
func (seg *Segmenter) Filter(filter filter.Func) {
	seg.filter = filter
}

// Transform applies one or more transforms to all tokens. Calling Transform
// will overwrite previous transforms, so call it once
// (it's variadic, you can add multiple, which will be applied in order).
func (seg *Segmenter) Transform(transformers ...transform.Transformer) {
	seg.transformer = transform.Chain(transformers...)
}

var ErrAdvanceNegative = errors.New("SplitFunc returned a negative advance, this is likely a bug in the SplitFunc")
var ErrAdvanceTooFar = errors.New("SplitFunc advanced beyond the end of the data, this is likely a bug in the SplitFunc")

// Next advances Segmenter to the next token (segment). It returns false when there
// are no remaining segments, or an error occurred.
func (seg *Segmenter) Next() bool {
next:
	for seg.pos < len(seg.data) {
		advance, token, err := seg.split(seg.data[seg.pos:], true)
		seg.pos += advance
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
		if seg.pos > len(seg.data) {
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
			seg.token, _, seg.err = transform.Bytes(seg.transformer, seg.token)
			if seg.err != nil {
				return false
			}
		}

		if seg.filter != nil && !seg.filter(seg.token) {
			continue next
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
	return seg.pos - len(seg.token)
}

// End returns the position (byte index) of the first byte after the current token,
// in the original text; segmenter.Bytes() == original[segmenter.Start():segmenter.End()]
func (seg *Segmenter) End() int {
	return seg.pos
}

// All will iterate through all tokens and collect them into a [][]byte. It is a
// convenience method -- if you will be allocating such a slice anyway, this
// will save you some code. The downside is that it allocates, and can do so
// unbounded -- O(n) on the number of tokens. Use Segmenter for more bounded
// memory usage.
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
