package graphemes

import (
	"github.com/clipperhouse/stringish"
	"github.com/clipperhouse/uax29/v2/internal/iterators"
)

// Iterator is a generic iterator for grapheme clusters in []byte data.
type Iterator[T stringish.Interface] struct {
	*iterators.Iterator[T]
}

var (
	splitFuncString = splitFunc[string]
	splitFuncBytes  = splitFunc[[]byte]
)

// StringIterator is an iterator for grapheme clusters in strings,
// with an ASCII hot path optimization.
type StringIterator struct {
	str   string
	pos   int
	start int
}

// FromString returns an iterator for the grapheme clusters in the input string.
// Iterate while Next() is true, and access the grapheme via Value().
func FromString(s string) *StringIterator {
	return &StringIterator{
		str: s,
	}
}

// Next advances the iterator to the next grapheme cluster.
// Returns false when there are no more grapheme clusters.
func (iter *StringIterator) Next() bool {
	if iter.pos >= len(iter.str) {
		return false
	}
	iter.start = iter.pos

	// ASCII hot path: if current byte is printable ASCII and
	// next byte is also ASCII (or end of string), return single byte
	b := iter.str[iter.pos]
	if b >= 0x20 && b < 0x7F {
		// Check next byte - if non-ASCII, it could be a combining mark
		if iter.pos+1 >= len(iter.str) || iter.str[iter.pos+1] < 0x80 {
			iter.pos++
			return true
		}
	}

	// Fall back to splitFunc
	remaining := iter.str[iter.pos:]
	advance, _, err := splitFuncString(remaining, true)
	if err != nil {
		panic(err)
	}
	if advance <= 0 {
		panic("splitFunc returned a zero or negative advance")
	}
	iter.pos += advance
	return true
}

// Value returns the current grapheme cluster.
func (iter *StringIterator) Value() string {
	return iter.str[iter.start:iter.pos]
}

// Start returns the byte position of the current grapheme in the original string.
func (iter *StringIterator) Start() int {
	return iter.start
}

// End returns the byte position after the current grapheme in the original string.
func (iter *StringIterator) End() int {
	return iter.pos
}

// Reset resets the iterator to the beginning of the string.
func (iter *StringIterator) Reset() {
	iter.start = 0
	iter.pos = 0
}

// SetText sets the text for the iterator to operate on, and resets all state.
func (iter *StringIterator) SetText(s string) {
	iter.str = s
	iter.start = 0
	iter.pos = 0
}

// FromBytes returns an iterator for the grapheme clusters in the input bytes.
// Iterate while Next() is true, and access the grapheme via Value().
func FromBytes(b []byte) Iterator[[]byte] {
	return Iterator[[]byte]{
		iterators.New(splitFuncBytes, b),
	}
}
