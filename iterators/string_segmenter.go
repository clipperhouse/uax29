package iterators

import (
	"bufio"
	"unsafe"

	"github.com/clipperhouse/uax29/iterators/filter"
	"golang.org/x/text/transform"
)

// StringSegmenter reuses the existing SplitFunc logic while achieving zero-copy behavior.
// It works by converting only the portion of the string needed for boundary detection
// to []byte, then extracting the result as a string slice.
type StringSegmenter struct {
	split       bufio.SplitFunc
	filter      filter.Func
	transformer transform.Transformer
	data        string
	pos         int
	start       int
	token       string
	err         error
}

// NewStringSegmenter creates a new StringSegmenter for the given string and SplitFunc.
func NewStringSegmenter(split bufio.SplitFunc) *StringSegmenter {
	return &StringSegmenter{
		split: split,
	}
}

// SetText sets the text for the segmenter to operate on, and resets all state.
func (seg *StringSegmenter) SetText(s string) {
	seg.data = s
	seg.pos = 0
	seg.start = 0
	seg.token = ""
	seg.err = nil
}

// Split sets the SplitFunc for the StringSegmenter.
func (seg *StringSegmenter) Split(split bufio.SplitFunc) {
	seg.split = split
}

// Filter applies a filter (predicate) to all tokens, returning only those
// where all filters evaluate true. Calling Filter will overwrite the previous
// filter.
func (seg *StringSegmenter) Filter(filter filter.Func) {
	seg.filter = filter
}

// Transform applies one or more transforms to all tokens. Calling Transform
// will overwrite previous transforms, so call it once
// (it's variadic, you can add multiple, which will be applied in order).
func (seg *StringSegmenter) Transform(transformers ...transform.Transformer) {
	seg.transformer = transform.Chain(transformers...)
}

// Next advances the segmenter to the next token. It returns false when there
// are no remaining tokens or an error occurred.
func (seg *StringSegmenter) Next() bool {
	for seg.pos < len(seg.data) {
		seg.start = seg.pos

		b := stringToBytes(seg.data[seg.pos:])

		advance, token, err := seg.split(b, true)
		if err != nil {
			seg.err = err
			return false
		}

		if advance <= 0 {
			return false
		}

		seg.pos += advance
		seg.token = bytesToString(token)

		// Apply transforms if any are set
		if seg.transformer != nil {
			transformed, _, err := transform.String(seg.transformer, seg.token)
			if err != nil {
				seg.err = err
				return false
			}
			seg.token = transformed
		}

		// Apply filter if any is set
		if seg.filter != nil && !seg.filter(token) {
			continue
		}

		return true
	}

	return false
}

// stringToBytes converts a string to []byte without allocation using unsafe.
// This is safe as long as the []byte is not modified and doesn't escape.
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// bytesToString converts a []byte to string without allocation using unsafe.
// This is safe as long as the []byte is not modified and doesn't escape.
func bytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// Text returns the current token as a string.
func (seg *StringSegmenter) Text() string {
	return seg.token
}

// Start returns the byte position of the current token in the original string.
func (seg *StringSegmenter) Start() int {
	return seg.start
}

// End returns the byte position after the current token in the original string.
func (seg *StringSegmenter) End() int {
	return seg.start + len(seg.token)
}

// Err returns any error that occurred during iteration.
func (seg *StringSegmenter) Err() error {
	return seg.err
}

// Reset resets the segmenter to the beginning of the string.
func (seg *StringSegmenter) Reset() {
	seg.pos = 0
	seg.start = 0
	seg.token = ""
	seg.err = nil
}
