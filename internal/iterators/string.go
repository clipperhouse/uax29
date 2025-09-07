package iterators

import (
	"bufio"
	"unsafe"
)

// StringIterator reuses the existing SplitFunc logic while achieving zero-copy behavior.
// It works by converting only the portion of the string needed for boundary detection
// to []byte, then extracting the result as a string slice.
type StringIterator struct {
	split bufio.SplitFunc
	data  string
	pos   int
	start int
	token string
	err   error
}

// NewStringIterator creates a new StringIterator for the given string and SplitFunc.
func NewStringIterator(split bufio.SplitFunc) *StringIterator {
	return &StringIterator{
		split: split,
	}
}

// SetText sets the text for the iterator to operate on, and resets all state.
func (iter *StringIterator) SetText(s string) {
	iter.data = s
	iter.pos = 0
	iter.start = 0
	iter.token = ""
	iter.err = nil
}

// Split sets the SplitFunc for the StringIterator.
func (iter *StringIterator) Split(split bufio.SplitFunc) {
	iter.split = split
}

// Next advances the iterator to the next token. It returns false when there
// are no remaining tokens or an error occurred.
func (iter *StringIterator) Next() bool {
	if iter.pos >= len(iter.data) {
		return false
	}

	iter.start = iter.pos

	b := stringToBytes(iter.data[iter.pos:])

	advance, token, err := iter.split(b, true)
	if err != nil {
		iter.err = err
		return false
	}

	if advance <= 0 {
		return false
	}

	iter.pos += advance
	iter.token = bytesToString(token)

	return true
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
func (iter *StringIterator) Text() string {
	return iter.token
}

// Start returns the byte position of the current token in the original string.
func (iter *StringIterator) Start() int {
	return iter.start
}

// End returns the byte position after the current token in the original string.
func (iter *StringIterator) End() int {
	return iter.pos
}

// Err returns any error that occurred during iteration.
func (iter *StringIterator) Err() error {
	return iter.err
}

// Reset resets the segmenter to the beginning of the string.
func (iter *StringIterator) Reset() {
	iter.pos = 0
	iter.start = 0
	iter.token = ""
	iter.err = nil
}
