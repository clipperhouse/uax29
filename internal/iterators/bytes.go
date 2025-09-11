// Package iterators is a support (base types) package for other packages in UAX29.
package iterators

import (
	"bufio"
	"errors"
)

// BytesIterator is an iterator for bytes. Iterate while Next() is true,
// It is only intended for use within this package. It will do the
// wrong thing with a SplitFunc that does not return all bytes.
type BytesIterator struct {
	split bufio.SplitFunc
	data  []byte
	token []byte
	start int
	pos   int
}

func NewBytesIterator(split bufio.SplitFunc) *BytesIterator {
	return &BytesIterator{
		split: split,
	}
}

// SetText sets the text for the BytesIterator to operate on, and resets
// all state.
func (iter *BytesIterator) SetText(data []byte) {
	iter.data = data
	iter.token = nil
	iter.pos = 0
}

// Split sets the SplitFunc for the BytesIterator
func (iter *BytesIterator) Split(split bufio.SplitFunc) {
	iter.split = split
}

var errAdvanceIllegal = errors.New("SplitFunc returned a zero or negative advance, this is likely a bug in the SplitFunc")
var errAdvanceTooFar = errors.New("SplitFunc advanced beyond the end of the data, this is likely a bug in the SplitFunc")

// Next advances BytesIterator to the next token. It returns false
// when there are no remaining tokens.
func (iter *BytesIterator) Next() bool {
	if iter.pos == len(iter.data) {
		return false
	}
	if iter.pos > len(iter.data) {
		panic(errAdvanceTooFar)
	}

	iter.start = iter.pos

	advance, token, err := iter.split(iter.data[iter.pos:], true)
	if err != nil {
		panic(err)
	}
	if advance <= 0 {
		panic(errAdvanceIllegal)
	}

	iter.pos += advance
	if iter.pos > len(iter.data) {
		panic(errAdvanceTooFar)
	}

	iter.token = token

	return true
}

// Bytes returns the current token.
func (iter *BytesIterator) Bytes() []byte {
	return iter.token
}

// Text returns the current token as a newly-allocated string.
func (iter *BytesIterator) Text() string {
	return string(iter.token)
}

// Start returns the position (byte index) of the current token in the original text.
func (iter *BytesIterator) Start() int {
	return iter.start
}

// End returns the position (byte index) of the first byte after the current token,
// in the original text.
func (iter *BytesIterator) End() int {
	return iter.pos
}
