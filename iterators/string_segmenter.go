package iterators

import (
	"bufio"
	"unsafe"
)

// StringSegmenter reuses the existing SplitFunc logic while achieving zero-copy behavior.
// It works by converting only the portion of the string needed for boundary detection
// to []byte, then extracting the result as a string slice.
type StringSegmenter struct {
	split bufio.SplitFunc
	data  string
	pos   int
	start int
	token string
	err   error
}

// NewStringSegmenter creates a new StringSegmenter for the given string and SplitFunc.
func NewStringSegmenter(s string, split bufio.SplitFunc) *StringSegmenter {
	return &StringSegmenter{
		split: split,
		data:  s,
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

// Next advances the segmenter to the next token. It returns false when there
// are no remaining tokens or an error occurred.
func (seg *StringSegmenter) Next() bool {
	if seg.pos >= len(seg.data) {
		return false
	}

	seg.start = seg.pos

	// Convert only the remaining portion to []byte for SplitFunc using unsafe
	// This avoids the allocation that []byte(string) would cause
	remaining := seg.data[seg.pos:]
	remainingBytes := stringToBytes(remaining)

	advance, _, err := seg.split(remainingBytes, true)
	if err != nil {
		seg.err = err
		return false
	}

	if advance <= 0 {
		return false
	}

	seg.pos += advance

	// Extract the token as a string slice of the original string
	// This is zero-copy since we're just slicing the original string
	seg.token = seg.data[seg.start:seg.pos]

	return true
}

// stringToBytes converts a string to []byte without allocation using unsafe.
// This is safe as long as the []byte is not modified and doesn't escape.
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
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
