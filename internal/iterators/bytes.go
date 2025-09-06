// Package iterators is a support (base types) package for other packages in UAX29.
package iterators

import (
	"bufio"
	"errors"
)

type BytesIterator struct {
	split bufio.SplitFunc
	data  []byte
	token []byte
	start int
	pos   int
	err   error
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
	iter.err = nil
}

// Split sets the SplitFunc for the BytesIterator
func (iter *BytesIterator) Split(split bufio.SplitFunc) {
	iter.split = split
}

var ErrAdvanceNegative = errors.New("SplitFunc returned a negative advance, this is likely a bug in the SplitFunc")
var ErrAdvanceTooFar = errors.New("SplitFunc advanced beyond the end of the data, this is likely a bug in the SplitFunc")

// Next advances BytesIterator to the next token (segment). It returns false
// when there are no remaining segments, or an error occurred.
//
// Always check Err() after Next() returns false.
func (iter *BytesIterator) Next() bool {
	if iter.pos >= len(iter.data) {
		return false
	}

	iter.start = iter.pos

	advance, token, err := iter.split(iter.data[iter.pos:], true)
	iter.pos += advance
	iter.token = token
	iter.err = err

	if iter.err != nil {
		return false
	}

	// Guardrails
	if advance < 0 {
		iter.err = ErrAdvanceNegative
		return false
	}
	if iter.pos > len(iter.data) {
		iter.err = ErrAdvanceTooFar
		return false
	}

	// Interpret as EOF
	if advance == 0 {
		return false
	}

	// Interpret as EOF
	if len(iter.token) == 0 {
		return false
	}

	return true
}

// Err indicates an error occured when calling Next; Next() will return false
// when an error occurs.
func (iter *BytesIterator) Err() error {
	return iter.err
}

// Bytes returns the current token.
func (iter *BytesIterator) Bytes() []byte {
	return iter.token
}

// Text returns the current token as a newly-allocated string.
func (iter *BytesIterator) Text() string {
	return string(iter.token)
}

// These extensive comments are here because someone is gonna be surprised by
// some custom SplitFunc, and it will be an annoying bug, so let's spell it all out.

// If you're just using a BytesIterator from the words, sentences, or graphemes
// sub-packages, what follows is irrelevant, carry on.

// For Start and End, we are taking some assumptions below. The SplitFunc interface
// allows ambiguity -- it doesn't return an explicit start or end. The SplitFunc
// could skip bytes before or after a token, and we won't know. We've found that
// skipping bytes at the beginning is unconventional, so we make that assumption.

// The SplitFuncs in the words, sentences, and graphemes packages adhere to this
// assumption, and in fact skip no bytes at all. This BytesIterator is designed for
// use with those, otherwise caveat emptor.

// If a SplitFunc skips bytes before *and* after a token, then there is unlikely to
// be a knowable right answer. Maybe the imprecision is OK for a given application.

// For future work, we might consider implementing a SegmentFunc interface,
// to make start and end explicit.

// Start returns the position (byte index) of the current token in the original text.
func (iter *BytesIterator) Start() int {
	return iter.start
}

// End returns the position (byte index) of the first byte after the current token,
// in the original text.
//
// In other words, segmenter.Bytes() == original[segmenter.Start():segmenter.End()]
func (iter *BytesIterator) End() int {
	return iter.start + len(iter.token)
}

// All iterates through all tokens and collect them into a [][]byte. It is a
// convenience method. The downside is that it allocates, and can do so unbounded:
// O(n) on the number of tokens (24 bytes per token). Prefer BytesIterator for constant
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
